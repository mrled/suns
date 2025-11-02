package validation

import (
	"fmt"
	"strings"

	"github.com/mrled/suns/symval/internal/model"
)

// validateDoubleFlip180 validates double 180-degree flip symmetry
// This checks if two domains are 180-degree flips of each other
// For example: "zq.su" and "ns.bz" - when you flip one 180 degrees, you get the other
func validateDoubleFlip180(data []*model.DomainRecord) (bool, error) {
	if len(data) != 2 {
		return false, fmt.Errorf("doubleflip180 validation expects exactly two domains, got %d", len(data))
	}

	hostname1 := data[0].Hostname
	hostname2 := data[1].Hostname

	flipped1, err1 := Flip180String(hostname1)
	if err1 == nil && strings.EqualFold(flipped1, hostname2) {
		// Also verify the reverse: hostname2 flipped should equal hostname1
		flipped2, err2 := Flip180String(hostname2)
		if err2 == nil && strings.EqualFold(flipped2, hostname1) {
			return true, nil
		}
	}

	if !strings.EqualFold(flipped1, hostname2) {
		return false, fmt.Errorf("hostnames %q and %q are not 180-degree flips of each other", hostname1, hostname2)
	}

	// Verify the reverse transformation
	flipped2, err2 := Flip180String(hostname2)
	if err2 != nil {
		return false, fmt.Errorf("cannot flip hostname %q: %v", hostname2, err2)
	}

	if !strings.EqualFold(flipped2, hostname1) {
		return false, fmt.Errorf("reverse flip validation failed: %q does not flip to %q", hostname2, hostname1)
	}

	return true, nil
}
