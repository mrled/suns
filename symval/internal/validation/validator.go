package validation

import (
	"fmt"

	"github.com/mrled/suns/symval/internal/groupid"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

// ValidateBase checks that all DomainRecord structs have consistent owner, type, and groupid,
// and that the groupid matches the calculated groupid for the given hostnames.
// Returns the common owner, groupID, and type if validation succeeds.
func ValidateBase(data []*model.DomainRecord) (string, string, symgroup.SymmetryType, error) {
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
	expectedGroupID, err := groupid.CalculateV1(owner, string(symmetryType), hostnames)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to calculate group ID: %w", err)
	}

	// Compare the provided groupID with the calculated one
	if groupID != expectedGroupID {
		return "", "", "", fmt.Errorf("groupID validation failed: expected %s, got %s", expectedGroupID, groupID)
	}

	return owner, groupID, symmetryType, nil
}

// Validate performs base validation and then calls the appropriate type-specific validator
func Validate(data []*model.DomainRecord) (bool, error) {
	// Perform base validation
	_, _, symmetryType, err := ValidateBase(data)
	if err != nil {
		return false, err
	}

	// Call type-specific validation
	switch symmetryType {
	case symgroup.Palindrome:
		return validatePalindrome(data)
	case symgroup.Flip180:
		return validateFlip180(data)
	case symgroup.DoubleFlip180:
		return validateDoubleFlip180(data)
	case symgroup.MirrorText:
		return validateMirrorText(data)
	case symgroup.MirrorNames:
		return validateMirrorNames(data)
	default:
		return false, fmt.Errorf("unknown symmetry type: %s", symmetryType)
	}
}
