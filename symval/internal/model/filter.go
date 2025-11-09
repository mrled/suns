package model

import "strings"

// RecordFilter contains criteria for filtering domain records with multiple values per field.
// All criteria are optional; only non-empty slices are applied.
// Within each field, values are combined with OR logic (any value matches).
// Between fields, criteria are combined with AND logic (all fields must match).
type RecordFilter struct {
	// Owners filters by owner emails (case-insensitive, OR within list)
	Owners []string

	// GroupIDs filters by exact group ID matches (OR within list)
	GroupIDs []string

	// Domains filters by hostnames (case-insensitive, OR within list)
	Domains []string

	// Types filters by symmetry types (OR within list)
	Types []string
}

// FilterRecords filters a slice of domain records based on the provided criteria.
// Returns a new slice containing only records that match the filter.
// Empty filter slices are ignored (treated as "match all").
func FilterRecords(records []*DomainRecord, filter RecordFilter) []*DomainRecord {
	// If no filters specified, return all records
	if len(filter.Owners) == 0 && len(filter.Domains) == 0 && len(filter.GroupIDs) == 0 && len(filter.Types) == 0 {
		return records
	}

	// Create lookup maps for efficient filtering
	ownerMap := make(map[string]bool)
	for _, owner := range filter.Owners {
		ownerMap[strings.ToLower(owner)] = true
	}

	domainMap := make(map[string]bool)
	for _, domain := range filter.Domains {
		domainMap[strings.ToLower(domain)] = true
	}

	groupIDMap := make(map[string]bool)
	for _, groupID := range filter.GroupIDs {
		groupIDMap[groupID] = true
	}

	typeMap := make(map[string]bool)
	for _, t := range filter.Types {
		typeMap[t] = true
	}

	var filtered []*DomainRecord

	for _, record := range records {
		// Apply owner filter (case-insensitive)
		if len(filter.Owners) > 0 && !ownerMap[strings.ToLower(record.Owner)] {
			continue
		}

		// Apply domain filter (case-insensitive)
		if len(filter.Domains) > 0 && !domainMap[strings.ToLower(record.Hostname)] {
			continue
		}

		// Apply groupID filter (exact match)
		if len(filter.GroupIDs) > 0 && !groupIDMap[record.GroupID] {
			continue
		}

		// Apply type filter (exact match)
		if len(filter.Types) > 0 && !typeMap[string(record.Type)] {
			continue
		}

		filtered = append(filtered, record)
	}

	return filtered
}
