package validation

import (
	"context"

	"github.com/callista/symval/internal/model"
)

// Validator defines the interface for domain validation
type Validator interface {
	Validate(ctx context.Context, data *model.DomainData) (bool, error)
}

// Service is a stub implementation of the Validator interface
type Service struct{}

// NewService creates a new validation service
func NewService() *Service {
	return &Service{}
}

// Validate is a stub that always returns true
func (s *Service) Validate(ctx context.Context, data *model.DomainData) (bool, error) {
	// Stub implementation - always returns true
	return true, nil
}
