package validation

import (
	"context"
	"fmt"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/service/groupid"
)

// Validator defines the interface for domain validation
type Validator interface {
	Validate(ctx context.Context, data []*model.DomainData) (bool, error)
}

// Service implements the Validator interface
type Service struct {
	groupIDService *groupid.Service
}

// NewService creates a new validation service
func NewService() *Service {
	return &Service{
		groupIDService: groupid.NewService(),
	}
}

// ValidateBase checks that all DomainData structs have consistent owner, type, and groupid,
// and that the groupid matches the calculated groupid for the given hostnames.
// Returns the common owner, groupID, and type if validation succeeds.
func (s *Service) ValidateBase(ctx context.Context, data []*model.DomainData) (string, string, model.SymmetryType, error) {
	if len(data) == 0 {
		return "", "", "", fmt.Errorf("no domain data provided")
	}

	// Use the first entry as the reference
	reference := data[0]
	owner := reference.Owner
	symmetryType := reference.Type
	groupID := reference.GroupID

	// Collect all hostnames and validate consistency
	hostnames := make([]string, 0, len(data))
	for _, d := range data {
		if d.Owner != owner {
			return "", "", "", fmt.Errorf("owner mismatch: expected %s, got %s", owner, d.Owner)
		}
		if d.Type != symmetryType {
			return "", "", "", fmt.Errorf("type mismatch: expected %s, got %s", symmetryType, d.Type)
		}
		if d.GroupID != groupID {
			return "", "", "", fmt.Errorf("groupID mismatch: expected %s, got %s", groupID, d.GroupID)
		}
		hostnames = append(hostnames, d.Hostname)
	}

	// Calculate the expected groupID
	expectedGroupID, err := s.groupIDService.CalculateV1(owner, string(symmetryType), hostnames)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to calculate group ID: %w", err)
	}

	// Compare the provided groupID with the calculated one
	if groupID != expectedGroupID {
		return "", "", "", fmt.Errorf("groupID validation failed: expected %s, got %s", expectedGroupID, groupID)
	}

	return owner, groupID, symmetryType, nil
}

// validatePalindrome validates palindrome symmetry
func (s *Service) validatePalindrome(ctx context.Context, data []*model.DomainData) (bool, error) {
	// Stub implementation - always returns true
	return true, nil
}

// validateFlip180 validates 180-degree flip symmetry
func (s *Service) validateFlip180(ctx context.Context, data []*model.DomainData) (bool, error) {
	// Stub implementation - always returns true
	return true, nil
}

// validateDoubleFlip180 validates double 180-degree flip symmetry
func (s *Service) validateDoubleFlip180(ctx context.Context, data []*model.DomainData) (bool, error) {
	// Stub implementation - always returns true
	return true, nil
}

// validateMirrorText validates mirror text symmetry
func (s *Service) validateMirrorText(ctx context.Context, data []*model.DomainData) (bool, error) {
	// Stub implementation - always returns true
	return true, nil
}

// validateMirrorNames validates mirror names symmetry
func (s *Service) validateMirrorNames(ctx context.Context, data []*model.DomainData) (bool, error) {
	// Stub implementation - always returns true
	return true, nil
}

// validateAntonymNames validates antonym names symmetry
func (s *Service) validateAntonymNames(ctx context.Context, data []*model.DomainData) (bool, error) {
	// Stub implementation - always returns true
	return true, nil
}

// Validate performs base validation and then calls the appropriate type-specific validator
func (s *Service) Validate(ctx context.Context, data []*model.DomainData) (bool, error) {
	// Perform base validation
	_, _, symmetryType, err := s.ValidateBase(ctx, data)
	if err != nil {
		return false, err
	}

	// Call type-specific validation
	switch symmetryType {
	case model.Palindrome:
		return s.validatePalindrome(ctx, data)
	case model.Flip180:
		return s.validateFlip180(ctx, data)
	case model.DoubleFlip180:
		return s.validateDoubleFlip180(ctx, data)
	case model.MirrorText:
		return s.validateMirrorText(ctx, data)
	case model.MirrorNames:
		return s.validateMirrorNames(ctx, data)
	case model.AntonymNames:
		return s.validateAntonymNames(ctx, data)
	default:
		return false, fmt.Errorf("unknown symmetry type: %s", symmetryType)
	}
}
