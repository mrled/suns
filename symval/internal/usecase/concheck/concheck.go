package concheck

import (
	"fmt"

	"github.com/mrled/suns/symval/internal/groupid"
	"github.com/mrled/suns/symval/internal/service/dnsclaims"
)

// ConsistencyCheckUseCase orchestrates the DNS verification and group ID validation process
type ConsistencyCheckUseCase struct {
	dnsService *dnsclaims.Service
}

// NewConsistencyCheckUseCase creates a new verify use case with the given DNS service
func NewConsistencyCheckUseCase(dnsService *dnsclaims.Service) *ConsistencyCheckUseCase {
	return &ConsistencyCheckUseCase{
		dnsService: dnsService,
	}
}

// CheckGroupIdConsistency checks consistency of group IDs.
// 1. Verify that all of the same owner hash
// 2. ... more checks can be added here ...
// Returns an error if the group IDs are inconsistent or if any group ID cannot be parsed.
func CheckGroupIdConsistency(groupIDs []groupid.GroupIDV1) error {
	if len(groupIDs) == 0 {
		return nil
	}

	// Use the first group ID's owner hash as the reference
	referenceOwnerHash := groupIDs[0].OwnerHash

	// Check that all subsequent group IDs have the same owner hash
	for i, gid := range groupIDs {
		if gid.OwnerHash != referenceOwnerHash {
			return fmt.Errorf("group ID at index %d has different owner hash: expected %s, got %s",
				i, referenceOwnerHash, gid.OwnerHash)
		}
	}

	return nil
}

// CheckDomainClaimRecordsConsistency looks up the TXT records for a domain and checks their consistency.
// It returns the parsed group IDs if verification passes,
// an empty slice with no error if no records exist, or an empty slice with an error
// if verification fails or parsing fails.
func (uc *ConsistencyCheckUseCase) CheckDomainClaimRecordsConsistency(domain string) ([]groupid.GroupIDV1, error) {
	// Lookup TXT records
	records, err := uc.dnsService.Lookup(domain)
	if err != nil {
		return nil, err
	}

	// If no records found, return empty slice with no error
	if len(records) == 0 {
		return []groupid.GroupIDV1{}, nil
	}

	groupIDs, err := groupid.ParseGroupIDv1Slice(records)
	if err != nil {
		return nil, fmt.Errorf("failed to parse group IDs: %w", err)
	}

	// Verify consistency
	if err := CheckGroupIdConsistency(groupIDs); err != nil {
		return nil, fmt.Errorf("verification failed: %w", err)
	}

	return groupIDs, nil
}
