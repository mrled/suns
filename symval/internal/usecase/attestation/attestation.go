package attestation

import (
	"context"
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
	repository model.DomainRepository
}

// NewAttestationUseCase creates a new attestation use case
// If repository is nil, attestation results will not be persisted
func NewAttestationUseCase(dnsService *dnsclaims.Service, repo model.DomainRepository) *AttestationUseCase {
	return &AttestationUseCase{
		dnsService: dnsService,
		repository: repo,
	}
}

// AttestResult contains the result of an attestation check
type AttestResult struct {
	IsValid       bool
	ExpectedID    string
	GroupIDs      []groupid.GroupIDV1
	DomainRecords []*model.DomainRecord
	ErrorMessage  string
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

	// Look up DNS records for all domains and filter them
	var allRawRecords []string
	var allDomainRecords []*model.DomainRecord
	validateTime := time.Now()

	// Set up filter criteria using the provided owner and type
	criteria := FilterCriteria{
		Owner: &owner,
		Type:  &symmetryType,
	}

	for _, domain := range domains {
		records, err := uc.dnsService.Lookup(domain)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup DNS records for %s: %w", domain, err)
		}

		// Filter the records for this domain
		filteredData, err := filterDomainRecords(domain, records, criteria, validateTime)
		if err != nil {
			return nil, fmt.Errorf("failed to filter records for %s: %w", domain, err)
		}

		// Fail attestation if no matching records found for this domain
		if len(filteredData) == 0 {
			result.IsValid = false
			result.ErrorMessage = fmt.Sprintf("no matching records found for domain %s", domain)
			return result, nil
		}

		// Use the first matching record for this domain
		allDomainRecords = append(allDomainRecords, filteredData[0])

		// Collect the group ID for consistency checking
		allRawRecords = append(allRawRecords, filteredData[0].GroupID)
	}

	// Parse all records at once using ParseGroupIDv1Slice
	allGroupIDs, err := groupid.ParseGroupIDv1Slice(allRawRecords)
	if err != nil {
		// If any record fails to parse, return error
		return nil, fmt.Errorf("failed to parse DNS records: %w", err)
	}

	result.GroupIDs = allGroupIDs
	result.DomainRecords = allDomainRecords

	// Check for consistency across all group IDs
	if err := concheck.CheckGroupIdConsistency(allGroupIDs); err != nil {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("consistency check failed: %v", err)
		return result, nil
	}

	// Validate the group
	isValid, err := validation.Validate(allDomainRecords)
	if err != nil {
		result.IsValid = false
		result.ErrorMessage = fmt.Sprintf("validation failed: %v", err)
		return result, nil
	}

	result.IsValid = isValid

	// If attestation is successful and repository is configured, persist the records
	if result.IsValid && uc.repository != nil {
		ctx := context.Background()
		for _, record := range allDomainRecords {
			if err := uc.repository.Store(ctx, record); err != nil {
				// Log and exit with error
				fmt.Printf("Warning: failed to store record for %s: %v\n", record.Hostname, err)
				return nil, fmt.Errorf("failed to store record for %s: %w", record.Hostname, err)
			}
		}
	}

	return result, nil
}
