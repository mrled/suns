package revalidate

import (
	"context"
	"fmt"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/validation"
)

// RevalidateUseCase handles revalidation of domain records in the data store
type RevalidateUseCase struct {
	repository model.DomainRepository
}

// NewRevalidateUseCase creates a new revalidate use case
func NewRevalidateUseCase(repo model.DomainRepository) *RevalidateUseCase {
	return &RevalidateUseCase{
		repository: repo,
	}
}

// FilterOptions contains optional filtering criteria for revalidation
type FilterOptions struct {
	Owners   []string
	Domains  []string
	GroupIDs []string
}

// InvalidRecordInfo contains an invalid record along with the reason it's invalid
type InvalidRecordInfo struct {
	Record *model.DomainRecord
	Reason string
}

// FindInvalid checks all records in the data store for consistency.
// It does not query DNS - it only validates existing records.
// For each record, it ensures the record is part of a valid group using Validate.
// If filters are provided:
//   - owners: checks records for those owners
//   - domains: checks the record for those domains AND all records in any group that those domains are part of
//   - groupIDs: checks records for those groups
//
// Returns a list of invalid records with their validation failure reasons.
func (uc *RevalidateUseCase) FindInvalid(ctx context.Context, filters FilterOptions) ([]InvalidRecordInfo, error) {
	// Get all records from repository
	allRecords, err := uc.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	// If no records, return empty list
	if len(allRecords) == 0 {
		return []InvalidRecordInfo{}, nil
	}

	// Apply initial filtering to get candidate records
	candidateRecords := filterRecords(allRecords, filters)

	// If domain filter is specified, we need to expand to include all records
	// in any group that the domains are part of
	if len(filters.Domains) > 0 {
		candidateRecords = expandForDomainFilter(allRecords, candidateRecords, filters.Domains)
	}

	// Group records by GroupID
	groupedRecords := model.GroupByGroupID(candidateRecords)

	// Validate each group and collect invalid records with reasons
	var invalidRecords []InvalidRecordInfo

	for _, groupRecords := range groupedRecords {
		// Validate the group
		_, err := validation.Validate(groupRecords)
		if err != nil {
			// If validation fails, add all records in this group to invalid list with the error reason
			reason := err.Error()
			for _, record := range groupRecords {
				invalidRecords = append(invalidRecords, InvalidRecordInfo{
					Record: record,
					Reason: reason,
				})
			}
		}
	}

	return invalidRecords, nil
}

// FindInvalidAndDrop finds invalid records and removes them from the repository
func (uc *RevalidateUseCase) FindInvalidAndDrop(ctx context.Context, filters FilterOptions) ([]InvalidRecordInfo, error) {
	// Find invalid records
	invalidRecords, err := uc.FindInvalid(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to find invalid records: %w", err)
	}

	// Delete each invalid record
	for _, info := range invalidRecords {
		record := info.Record
		if err := uc.repository.UnconditionalDelete(ctx, record.GroupID, record.Hostname); err != nil {
			// If delete fails, return what we've found so far with an error
			return invalidRecords, fmt.Errorf("failed to delete record %s (group %s): %w", record.Hostname, record.GroupID, err)
		}
	}

	return invalidRecords, nil
}

// filterRecords applies basic filtering based on owners, domains, and groupIDs
func filterRecords(records []*model.DomainRecord, filters FilterOptions) []*model.DomainRecord {
	// If no filters specified, return all records
	if len(filters.Owners) == 0 && len(filters.Domains) == 0 && len(filters.GroupIDs) == 0 {
		return records
	}

	// Create lookup maps for efficient filtering
	ownerMap := make(map[string]bool)
	for _, owner := range filters.Owners {
		ownerMap[owner] = true
	}

	domainMap := make(map[string]bool)
	for _, domain := range filters.Domains {
		domainMap[domain] = true
	}

	groupIDMap := make(map[string]bool)
	for _, groupID := range filters.GroupIDs {
		groupIDMap[groupID] = true
	}

	var filtered []*model.DomainRecord

	for _, record := range records {
		// Apply owner filter
		if len(filters.Owners) > 0 && !ownerMap[record.Owner] {
			continue
		}

		// Apply domain filter (will be expanded later)
		if len(filters.Domains) > 0 && !domainMap[record.Hostname] {
			continue
		}

		// Apply groupID filter
		if len(filters.GroupIDs) > 0 && !groupIDMap[record.GroupID] {
			continue
		}

		filtered = append(filtered, record)
	}

	return filtered
}

// expandForDomainFilter expands the candidate records to include all records
// in any group that the specified domains are part of
func expandForDomainFilter(allRecords []*model.DomainRecord, candidateRecords []*model.DomainRecord, domains []string) []*model.DomainRecord {
	// Create a lookup map for target domains
	domainMap := make(map[string]bool)
	for _, domain := range domains {
		domainMap[domain] = true
	}

	// Find all groups that any of the domains are part of
	groupIDs := make(map[string]bool)
	for _, record := range candidateRecords {
		if domainMap[record.Hostname] {
			groupIDs[record.GroupID] = true
		}
	}

	// If no groups found for the domains, return the original candidates
	if len(groupIDs) == 0 {
		return candidateRecords
	}

	// Create a map of existing candidates for quick lookup
	existingRecords := make(map[string]bool)
	for _, record := range candidateRecords {
		existingRecords[record.Hostname] = true
	}

	// Add all records from the identified groups
	var expanded []*model.DomainRecord
	expanded = append(expanded, candidateRecords...)

	for _, record := range allRecords {
		if groupIDs[record.GroupID] && !existingRecords[record.Hostname] {
			expanded = append(expanded, record)
			existingRecords[record.Hostname] = true
		}
	}

	return expanded
}
