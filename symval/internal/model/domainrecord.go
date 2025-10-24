package model

import (
	"time"

	"github.com/mrled/suns/symval/internal/symgroup"
)

// DomainRecord represents domain validation information
type DomainRecord struct {
	Owner        string
	Type         symgroup.SymmetryType
	Hostname     string
	GroupID      string
	ValidateTime time.Time
}

// GroupByGroupID groups domain records by their GroupID
func GroupByGroupID(records []*DomainRecord) map[string][]*DomainRecord {
	grouped := make(map[string][]*DomainRecord)

	for _, record := range records {
		grouped[record.GroupID] = append(grouped[record.GroupID], record)
	}

	return grouped
}
