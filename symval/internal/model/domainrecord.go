package model

import (
	"context"
	"errors"
	"time"

	"github.com/mrled/suns/symval/internal/symgroup"
)

var (
	ErrNotFound      = errors.New("domain data not found")
	ErrAlreadyExists = errors.New("domain data already exists")
)

// DomainRepository defines the interface for storing and retrieving domain data
type DomainRepository interface {
	// Store saves domain data
	Store(ctx context.Context, data *DomainRecord) error

	// Get retrieves domain data by group ID and domain name (the composite key)
	Get(ctx context.Context, groupID, domain string) (*DomainRecord, error)

	// List retrieves all domain data
	List(ctx context.Context) ([]*DomainRecord, error)

	// Delete removes domain data by group ID and domain name (the composite key)
	Delete(ctx context.Context, groupID, domain string) error
}

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
