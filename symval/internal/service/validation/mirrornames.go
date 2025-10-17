package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/mrled/suns/symval/internal/model"
)

// isMirrorPair checks if two strings are mirror pairs.
// It divides each string by dots and checks that the first segment of s1
// equals the last segment of s2, the second segment of s1 equals the
// penultimate segment of s2, and so on.
// Returns false if the number of segments don't match.
func isMirrorPair(s1, s2 string) bool {
	// Split by dots
	segments1 := strings.Split(s1, ".")
	segments2 := strings.Split(s2, ".")

	// Check if number of segments match
	if len(segments1) != len(segments2) {
		return false
	}

	// Check if segments mirror each other
	length := len(segments1)
	for i := 0; i < length; i++ {
		if segments1[i] != segments2[length-1-i] {
			return false
		}
	}

	return true
}

// validateMirrorNames validates mirror names symmetry
func (s *Service) validateMirrorNames(ctx context.Context, data []*model.DomainData) (bool, error) {
	if len(data) != 2 {
		return false, fmt.Errorf("mirrornames validation expects exactly two domains, got %d", len(data))
	}

	hostname1 := data[0].Hostname
	hostname2 := data[1].Hostname

	if !isMirrorPair(hostname1, hostname2) {
		return false, fmt.Errorf("hostnames %q and %q are not mirror pairs", hostname1, hostname2)
	}

	return true, nil
}
