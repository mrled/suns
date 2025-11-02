package reattest

import (
	"context"
	"fmt"
	"time"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/service/dnsclaims"
	"github.com/mrled/suns/symval/internal/symgroup"
	"github.com/mrled/suns/symval/internal/usecase/attestation"
)

// ReattestUseCase handles re-attestation of all groups in the data store
type ReattestUseCase struct {
	dnsService       *dnsclaims.Service
	repository       model.DomainRepository
	dynamoRepo       model.DomainRepository // Optional: for updating validation timestamps
	gracePeriodHours int
}

// NewReattestUseCase creates a new reattest use case
func NewReattestUseCase(dnsService *dnsclaims.Service, repo model.DomainRepository) *ReattestUseCase {
	return &ReattestUseCase{
		dnsService:       dnsService,
		repository:       repo,
		gracePeriodHours: 72, // Default grace period
	}
}

// NewReattestUseCaseWithDynamo creates a new reattest use case with DynamoDB support for updates
func NewReattestUseCaseWithDynamo(dnsService *dnsclaims.Service, repo model.DomainRepository, dynamoRepo model.DomainRepository) *ReattestUseCase {
	return &ReattestUseCase{
		dnsService:       dnsService,
		repository:       repo,
		dynamoRepo:       dynamoRepo,
		gracePeriodHours: 72, // Default grace period
	}
}

// SetGracePeriod sets the grace period in hours for dropping invalid groups
func (uc *ReattestUseCase) SetGracePeriod(hours int) {
	uc.gracePeriodHours = hours
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

// UpdateStats tracks statistics for ReattestAllAndUpdate operations
type UpdateStats struct {
	GroupsProcessed int
	RecordsUpdated  int
	RecordsDeleted  int
	RecordsSkipped  int
	Errors          int
}

// ReattestAllAndUpdate loads all groups from the datastore, re-attests them,
// updates validation timestamps for valid groups, and removes records for
// invalid groups that have exceeded the grace period.
func (uc *ReattestUseCase) ReattestAllAndUpdate(ctx context.Context) ([]GroupAttestResult, UpdateStats, error) {
	stats := UpdateStats{}

	// If no dynamoRepo is set, fall back to using the regular repository
	updateRepo := uc.dynamoRepo
	if updateRepo == nil {
		updateRepo = uc.repository
	}

	// Perform re-attestation
	results, err := uc.ReattestAll(ctx)
	if err != nil {
		return nil, stats, fmt.Errorf("failed to re-attest groups: %w", err)
	}

	stats.GroupsProcessed = len(results)

	// Process each attestation result
	for _, result := range results {
		if result.IsValid {
			// Attestation succeeded - update all records in the group with current timestamp
			for _, record := range result.Records {
				// Keep the snapshot revision for conditional update
				snapshotRev := record.Rev
				record.ValidateTime = time.Now()
				if _, err := updateRepo.SetValidationIfUnchanged(ctx, record, snapshotRev); err != nil {
					if err == model.ErrRevConflict {
						// Record changed during validation, skip
						stats.RecordsSkipped++
					} else {
						// Other error
						stats.Errors++
					}
				} else {
					stats.RecordsUpdated++
				}
			}
		} else {
			// Attestation failed - check grace period
			// Get the oldest validation time from the group
			var oldestValidation time.Time
			for _, record := range result.Records {
				if oldestValidation.IsZero() || record.ValidateTime.Before(oldestValidation) {
					oldestValidation = record.ValidateTime
				}
			}

			hoursSinceValidation := time.Since(oldestValidation).Hours()

			if hoursSinceValidation > float64(uc.gracePeriodHours) {
				// Grace period exceeded - delete all records in the group
				for _, record := range result.Records {
					if err := updateRepo.DeleteIfUnchanged(ctx, result.GroupID, record.Hostname, record.Rev); err != nil {
						if err == model.ErrRevConflict {
							// Record changed during deletion, skip
							stats.RecordsSkipped++
						} else {
							// Other error
							stats.Errors++
						}
					} else {
						stats.RecordsDeleted++
					}
				}
			} else {
				// Within grace period - skip all records
				stats.RecordsSkipped += len(result.Records)
			}
		}
	}

	return results, stats, nil
}
