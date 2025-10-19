package attestation

import (
	"time"

	"github.com/mrled/suns/symval/internal/groupid"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

// FilterCriteria contains optional filtering criteria for domain data
type FilterCriteria struct {
	Owner   *string
	Type    *symgroup.SymmetryType
	GroupID *string
}

// filterDomainRecords filters DNS records based on the provided criteria
// and returns matching DomainRecord structs. Records are parsed into GroupIDV1
// and filtered by optional Owner, Type, and GroupID values.
func filterDomainRecords(hostname string, records []string, criteria FilterCriteria, validateTime time.Time) ([]*model.DomainRecord, error) {
	var filtered []*model.DomainRecord

	for _, record := range records {
		// Parse the record
		gid, err := groupid.ParseGroupIDv1(record)
		if err != nil {
			// Skip invalid records
			continue
		}

		// Apply filters if specified
		if criteria.GroupID != nil && record != *criteria.GroupID {
			continue
		}

		if criteria.Type != nil && gid.TypeCode != string(*criteria.Type) {
			continue
		}

		// For owner filtering, we need to calculate the expected owner hash
		if criteria.Owner != nil {
			// Since GroupIDV1 only contains OwnerHash, we need to calculate
			// the expected hash from the provided owner
			expectedGroupID, err := groupid.CalculateV1(*criteria.Owner, gid.TypeCode, []string{hostname})
			if err != nil {
				continue
			}
			expectedGID, err := groupid.ParseGroupIDv1(expectedGroupID)
			if err != nil {
				continue
			}
			if gid.OwnerHash != expectedGID.OwnerHash {
				continue
			}
		}

		// Create DomainRecord for this matching record
		var ownerValue string
		if criteria.Owner != nil {
			ownerValue = *criteria.Owner
		}

		var typeValue symgroup.SymmetryType
		if criteria.Type != nil {
			typeValue = *criteria.Type
		} else {
			typeValue = symgroup.SymmetryType(gid.TypeCode)
		}

		filtered = append(filtered, &model.DomainRecord{
			Owner:        ownerValue,
			Type:         typeValue,
			Hostname:     hostname,
			GroupID:      record,
			ValidateTime: validateTime,
		})
	}

	return filtered, nil
}
