package attestation

import (
	"fmt"
	"time"

	"github.com/mrled/suns/symval/internal/groupid"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/service/dnsclaims"
	"github.com/mrled/suns/symval/internal/symgroup"
	"github.com/mrled/suns/symval/internal/usecase/concheck"
	"github.com/mrled/suns/symval/internal/validation"
)

// AttestationUseCase handles attestation of domain groups
type AttestationUseCase struct {
	dnsService *dnsclaims.Service
}

// NewAttestationUseCase creates a new attestation use case
func NewAttestationUseCase(dnsService *dnsclaims.Service) *AttestationUseCase {
	return &AttestationUseCase{
		dnsService: dnsService,
	}
}

// AttestResult contains the result of an attestation check
type AttestResult struct {
	IsValid      bool
	ExpectedID   string
	GroupIDs     []groupid.GroupIDV1
	DomainData   []*model.DomainData
	ErrorMessage string
}

// Attest verifies a group of domains for consistency and validity
// It calculates the expected group ID, looks up DNS records for all domains,
// checks for consistency, validates the group, and returns the validity result
func (uc *AttestationUseCase) Attest(owner string, symmetryType symgroup.SymmetryType, domains []string) (*AttestResult, error) {
	result := &AttestResult{}

	// Calculate the expected group ID
	expectedID, err := groupid.CalculateV1(owner, string(symmetryType), domains)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate group ID: %w", err)
	}
	result.ExpectedID = expectedID

	// Look up DNS records for all domains
	var allRecords []string
	domainDataMap := make(map[string]*model.DomainData)

	for _, domain := range domains {
		records, err := uc.dnsService.Lookup(domain)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup DNS records for %s: %w", domain, err)
		}

		// Collect all records for parsing
		allRecords = append(allRecords, records...)

		// Create DomainData entry for this domain using the first record
		if len(records) > 0 {
			if _, exists := domainDataMap[domain]; !exists {
				domainDataMap[domain] = &model.DomainData{
					Owner:        owner,
					Type:         symmetryType,
					Hostname:     domain,
					GroupID:      records[0],
					ValidateTime: time.Now(),
				}
			}
		}
	}

	// Parse all records at once using ParseGroupIDv1Slice
	allGroupIDs, err := groupid.ParseGroupIDv1Slice(allRecords)
	if err != nil {
		// If any record fails to parse, return error
		return nil, fmt.Errorf("failed to parse DNS records: %w", err)
	}

	result.GroupIDs = allGroupIDs

	// Check if we found any group IDs
	if len(allGroupIDs) == 0 {
		result.IsValid = false
		result.ErrorMessage = "no valid group IDs found in DNS records"
		return result, nil
	}

	// Check for consistency across all group IDs
	if err := concheck.CheckGroupIdConsistency(allGroupIDs); err != nil {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("consistency check failed: %v", err)
		return result, nil
	}

	// Convert domainDataMap to slice for validation
	domainDataSlice := make([]*model.DomainData, 0, len(domainDataMap))
	for _, dd := range domainDataMap {
		domainDataSlice = append(domainDataSlice, dd)
	}
	result.DomainData = domainDataSlice

	// Validate the group
	isValid, err := validation.Validate(domainDataSlice)
	if err != nil {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("validation failed: %v", err)
		return result, nil
	}

	result.IsValid = isValid
	return result, nil
}
