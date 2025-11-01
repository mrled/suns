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
	ErrRevConflict   = errors.New("revision conflict")
)

// DomainRepository defines the interface for storing and retrieving domain data
type DomainRepository interface {
	// UnconditionalStore saves domain data (new name for existing Store method). Returns new rev.
	UnconditionalStore(ctx context.Context, data *DomainRecord) (int64, error)

	// Upsert uses UpdateItem + rev = if_not_exists(rev,0)+1. This is called by the webhook/API in response to user requests. Returns new rev.
	Upsert(ctx context.Context, data *DomainRecord) (int64, error)

	// SetValidationIfUnchanged uses ConditionExpression rev = :snapshotRev to ensure that it only updates items that have not changed. Should only set a new validation date. Returns new rev.
	SetValidationIfUnchanged(ctx context.Context, data *DomainRecord, snapshotRev int64) (int64, error)

	// Get retrieves domain data by group ID and domain name (the composite key)
	Get(ctx context.Context, groupID, domain string) (*DomainRecord, error)

	// List retrieves all domain data
	List(ctx context.Context) ([]*DomainRecord, error)

	// UnconditionalDelete removes domain data by group ID and domain name (the composite key) (new name for existing Delete method)
	UnconditionalDelete(ctx context.Context, groupID, domain string) error

	// DeleteIfUnchanged does the same as SetValidationIfUnchanged, but for deletions.
	DeleteIfUnchanged(ctx context.Context, groupID, domain string, snapshotRev int64) error
}

// DomainRecord represents domain validation information
type DomainRecord struct {
	Owner        string
	Type         symgroup.SymmetryType
	Hostname     string
	GroupID      string
	ValidateTime time.Time
	Rev          int64 // Monotonically increasing revision number
}

// GroupByGroupID groups domain records by their GroupID
func GroupByGroupID(records []*DomainRecord) map[string][]*DomainRecord {
	grouped := make(map[string][]*DomainRecord)

	for _, record := range records {
		grouped[record.GroupID] = append(grouped[record.GroupID], record)
	}

	return grouped
}
