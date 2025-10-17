package validation

import (
	"context"

	"github.com/mrled/suns/symval/internal/model"
)

// validateAntonymNames validates antonym names symmetry
func (s *Service) validateAntonymNames(ctx context.Context, data []*model.DomainData) (bool, error) {
	// Stub implementation - always returns true
	return true, nil
}
