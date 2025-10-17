package validation

import (
	"context"

	"github.com/mrled/suns/symval/internal/model"
)

// validateMirrorNames validates mirror names symmetry
func (s *Service) validateMirrorNames(ctx context.Context, data []*model.DomainData) (bool, error) {
	// Stub implementation - always returns true
	return true, nil
}
