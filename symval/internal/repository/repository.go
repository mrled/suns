package repository

import (
	"context"
	"errors"

	"github.com/mrled/suns/symval/internal/model"
)

var (
	ErrNotFound      = errors.New("domain data not found")
	ErrAlreadyExists = errors.New("domain data already exists")
)

// DomainRepository defines the interface for storing and retrieving domain data
type DomainRepository interface {
	// Store saves domain data
	Store(ctx context.Context, data *model.DomainRecord) error

	// Get retrieves domain data by group ID and domain name (the composite key)
	Get(ctx context.Context, groupID, domain string) (*model.DomainRecord, error)

	// List retrieves all domain data
	List(ctx context.Context) ([]*model.DomainRecord, error)

	// Delete removes domain data by group ID and domain name (the composite key)
	Delete(ctx context.Context, groupID, domain string) error
}
