package model

import "sort"

// SortBy specifies the field and order for sorting domain records
type SortBy string

const (
	SortByOwner        SortBy = "owner"
	SortByDomain       SortBy = "domain"
	SortByGroup        SortBy = "group"
	SortByValidateTime SortBy = "validate-time"
	SortByType         SortBy = "type"
	SortByDefault      SortBy = "" // Default sort: group ID, then hostname
)

// SortRecords sorts a slice of domain records in place based on the specified field.
// The sortBy parameter should be one of: "owner", "domain", "group", "validate-time", "type".
// If sortBy is empty or unrecognized, records are sorted by group ID, then by hostname.
func SortRecords(records []*DomainRecord, sortBy string) {
	switch SortBy(sortBy) {
	case SortByOwner:
		sort.Slice(records, func(i, j int) bool {
			return records[i].Owner < records[j].Owner
		})
	case SortByDomain:
		sort.Slice(records, func(i, j int) bool {
			return records[i].Hostname < records[j].Hostname
		})
	case SortByGroup:
		sort.Slice(records, func(i, j int) bool {
			return records[i].GroupID < records[j].GroupID
		})
	case SortByValidateTime:
		sort.Slice(records, func(i, j int) bool {
			return records[i].ValidateTime.After(records[j].ValidateTime)
		})
	case SortByType:
		sort.Slice(records, func(i, j int) bool {
			return records[i].Type < records[j].Type
		})
	default:
		// Default sort by group ID, then by hostname
		sort.Slice(records, func(i, j int) bool {
			if records[i].GroupID != records[j].GroupID {
				return records[i].GroupID < records[j].GroupID
			}
			return records[i].Hostname < records[j].Hostname
		})
	}
}
