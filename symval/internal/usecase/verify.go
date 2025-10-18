package usecase

import (
	"fmt"

	"github.com/mrled/suns/symval/internal/service/dnsclaims"
	"github.com/mrled/suns/symval/internal/service/groupid"
)

// VerifyUseCase orchestrates the DNS verification and group ID validation process
type VerifyUseCase struct {
	dnsService *dnsclaims.Service
}

// NewVerifyUseCase creates a new verify use case with the given DNS service
func NewVerifyUseCase(dnsService *dnsclaims.Service) *VerifyUseCase {
	return &VerifyUseCase{
		dnsService: dnsService,
	}
}

// Verify checks consistency of group IDs by verifying that all group IDs
// have the same owner hash. This function validates that all provided group IDs
// belong to the same owner. Additional consistency checks may be added in the future.
// Returns an error if the group IDs are inconsistent or if any group ID cannot be parsed.
func Verify(groupIDs []groupid.GroupIDV1) error {
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

// VerifyDomain looks up the TXT records for a domain, parses them as group IDs,
// and verifies their consistency. It returns the parsed group IDs if verification passes,
// an empty slice with no error if no records exist, or an empty slice with an error
// if verification fails or parsing fails.
func (uc *VerifyUseCase) VerifyDomain(domain string) ([]groupid.GroupIDV1, error) {
	// Lookup TXT records
	records, err := uc.dnsService.Lookup(domain)
	if err != nil {
		return nil, err
	}

	// If no records found, return empty slice with no error
	if len(records) == 0 {
		return []groupid.GroupIDV1{}, nil
	}

	// Parse all records as group IDs
	groupIDs := make([]groupid.GroupIDV1, 0, len(records))
	for i, record := range records {
		gid, err := groupid.ParseGroupIDv1(record)
		if err != nil {
			return nil, fmt.Errorf("failed to parse record at index %d: %w", i, err)
		}
		groupIDs = append(groupIDs, gid)
	}

	// Verify consistency
	if err := Verify(groupIDs); err != nil {
		return nil, fmt.Errorf("verification failed: %w", err)
	}

	return groupIDs, nil
}
