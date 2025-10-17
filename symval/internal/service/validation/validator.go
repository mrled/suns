package validation

import (
	"context"
	"fmt"

	"github.com/callista/symval/internal/model"
	"github.com/callista/symval/internal/service/groupid"
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

// Validate checks that all DomainData structs have consistent owner, type, and groupid,
// and that the groupid matches the calculated groupid for the given hostnames
func (s *Service) Validate(ctx context.Context, data []*model.DomainData) (bool, error) {
	if len(data) == 0 {
		return false, fmt.Errorf("no domain data provided")
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
			return false, fmt.Errorf("owner mismatch: expected %s, got %s", owner, d.Owner)
		}
		if d.Type != symmetryType {
			return false, fmt.Errorf("type mismatch: expected %s, got %s", symmetryType, d.Type)
		}
		if d.GroupID != groupID {
			return false, fmt.Errorf("groupID mismatch: expected %s, got %s", groupID, d.GroupID)
		}
		hostnames = append(hostnames, d.Hostname)
	}

	// Calculate the expected groupID
	expectedGroupID, err := s.groupIDService.CalculateV1(owner, string(symmetryType), hostnames)
	if err != nil {
		return false, fmt.Errorf("failed to calculate group ID: %w", err)
	}

	// Compare the provided groupID with the calculated one
	if groupID != expectedGroupID {
		return false, fmt.Errorf("groupID validation failed: expected %s, got %s", expectedGroupID, groupID)
	}

	return true, nil
}
