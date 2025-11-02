package reattest

import (
	"context"
	"fmt"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/service/dnsclaims"
	"github.com/mrled/suns/symval/internal/symgroup"
	"github.com/mrled/suns/symval/internal/usecase/attestation"
)

// ReattestUseCase handles re-attestation of all groups in the data store
type ReattestUseCase struct {
	dnsService *dnsclaims.Service
	repository model.DomainRepository
}

// NewReattestUseCase creates a new reattest use case
func NewReattestUseCase(dnsService *dnsclaims.Service, repo model.DomainRepository) *ReattestUseCase {
	return &ReattestUseCase{
		dnsService: dnsService,
		repository: repo,
	}
}

// GroupAttestResult contains the result of re-attesting a group
type GroupAttestResult struct {
	GroupID      string
	Owner        string
	Type         string
	Domains      []string
	Records      []*model.DomainRecord // Include full records with revision info
	IsValid      bool
	ErrorMessage string
}

// ReattestAll loads all groups from the datastore and re-attests them by querying DNS.
// Returns a list of results for each group, indicating which groups are valid or invalid.
func (uc *ReattestUseCase) ReattestAll(ctx context.Context) ([]GroupAttestResult, error) {
	// Get all records from repository
	allRecords, err := uc.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	// If no records, return empty list
	if len(allRecords) == 0 {
		return []GroupAttestResult{}, nil
	}

	// Group records by GroupID
	groupedRecords := model.GroupByGroupID(allRecords)

	// Create attestation use case for performing attestations
	attestUC := attestation.NewAttestationUseCase(uc.dnsService, nil)

	// Re-attest each group
	var results []GroupAttestResult

	for groupID, groupRecords := range groupedRecords {
		// Get first record to extract owner and type
		firstRecord := groupRecords[0]
		owner := firstRecord.Owner
		symmetryType := firstRecord.Type

		// Extract all domains in this group
		domains := make([]string, 0, len(groupRecords))
		for _, record := range groupRecords {
			domains = append(domains, record.Hostname)
		}

		// Perform attestation
		attestResult, err := attestUC.Attest(owner, symgroup.SymmetryType(symmetryType), domains)
		if err != nil {
			// If there's an error performing attestation, mark as invalid
			result := GroupAttestResult{
				GroupID:      groupID,
				Owner:        owner,
				Type:         string(symmetryType),
				Domains:      domains,
				Records:      groupRecords,
				IsValid:      false,
				ErrorMessage: fmt.Sprintf("attestation error: %v", err),
			}
			results = append(results, result)
			continue
		}

		// Create result
		result := GroupAttestResult{
			GroupID:      groupID,
			Owner:        owner,
			Type:         string(symmetryType),
			Domains:      domains,
			Records:      groupRecords,
			IsValid:      attestResult.IsValid,
			ErrorMessage: attestResult.ErrorMessage,
		}

		results = append(results, result)
	}

	return results, nil
}

// ReattestAllAndDrop loads all groups from the datastore, re-attests them,
// and removes records for invalid groups from the repository.
func (uc *ReattestUseCase) ReattestAllAndDrop(ctx context.Context) ([]GroupAttestResult, error) {
	// Perform re-attestation
	results, err := uc.ReattestAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to re-attest groups: %w", err)
	}

	// Delete records for invalid groups
	for _, result := range results {
		if !result.IsValid {
			// Delete all records in this group using their snapshot revisions
			for _, record := range result.Records {
				if err := uc.repository.DeleteIfUnchanged(ctx, record.GroupID, record.Hostname, record.Rev); err != nil {
					if err == model.ErrRevConflict {
						// Record was modified, skip deletion
						continue
					}
					// If delete fails for other reasons, return what we've processed so far with an error
					return results, fmt.Errorf("failed to delete record %s (group %s): %w", record.Hostname, result.GroupID, err)
				}
			}
		}
	}

	return results, nil
}
