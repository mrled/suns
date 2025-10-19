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

	// Get retrieves domain data by domain name
	Get(ctx context.Context, domain string) (*model.DomainRecord, error)

	// List retrieves all domain data
	List(ctx context.Context) ([]*model.DomainRecord, error)

	// Delete removes domain data by domain name
	Delete(ctx context.Context, domain string) error
}
