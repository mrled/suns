package validation

import (
	"strings"
	"testing"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

func TestFlip180String(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{"palindrome flip", "pods", "spod", false},
		{"symmetric flip", "sos", "sos", false},
		{"with dots", "no.on", "uo.ou", false},
		{"numbers", "69", "69", false}, // 6->9, 9->6, reversed: "96" -> "69"
		{"unmappable char", "abc", "", true},

		// Test specific examples
		{"zq flips to bz", "zq", "bz", false},
		{"bz flips to zq", "bz", "zq", false},
		{"su flips to ns", "su", "ns", false},
		{"ns flips to su", "ns", "su", false},

		// Self-symmetric strings
		{"Symmetric sos", "sos", "sos", false},
		{"Symmetric swims", "swims", "", true}, // has unmapped chars (w, i, m)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Flip180String(tt.input)
			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("For input %q, expected %q, got %q", tt.input, tt.expected, result)
				}
			}
		})
	}
}

func TestIsFlip180(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Self-symmetric strings (flip to themselves)
		{"sos", "sos", true},
		{"SOS", "SOS", true},
		{"o", "o", true},
		{"88", "88", true},
		{"8008", "8008", true},

		// Non-symmetric strings
		{"hello", "hello", false},
		{"test", "test", false},
		{"abc", "abc", false},
		{"pods", "pods", false}, // flips to "spod", not itself
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isFlip180(tt.input)
			if result != tt.expected {
				t.Errorf("For input %q, expected %v, got %v", tt.input, tt.expected, result)
			}
		})
	}
}

func TestValidateFlip180(t *testing.T) {
	tests := []struct {
		name        string
		hostname    string
		expectValid bool
	}{
		{"Self-symmetric sos", "sos", true},
		{"Self-symmetric SOS", "SOS", true},
		{"Self-symmetric with dots", "s.o.s", true},
		{"Single o", "o", true},
		{"Numbers 88", "88", true},
		{"Not symmetric", "example.com", false},
		{"Not symmetric pods", "pods", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []*model.DomainRecord{
				{
					Owner:    "test@example.com",
					Type:     symgroup.Flip180,
					Hostname: tt.hostname,
					GroupID:  "test-group-id",
				},
			}

			valid, err := validateFlip180(data)
			if tt.expectValid {
				if err != nil {
					t.Errorf("Expected no error for hostname %q, got: %v", tt.hostname, err)
				}
				if !valid {
					t.Errorf("Expected valid=true for hostname %q, got false", tt.hostname)
				}
			} else {
				if valid {
					t.Errorf("Expected valid=false for hostname %q, got true", tt.hostname)
				}
			}
		})
	}
}

func TestFlip180MappingCompleteness(t *testing.T) {
	for char, flipped := range flip180Mapping {
		charStr := string(char)
		flippedStr := string(flipped)

		// Skip self-symmetric characters
		if char == flipped {
			continue
		}

		// Check reverse mapping exists
		if reverseFlipped, ok := flip180Mapping[flipped]; !ok {
			t.Errorf("Missing reverse mapping for %q -> %q", charStr, flippedStr)
		} else if reverseFlipped != char {
			t.Errorf("Inconsistent mapping: %q -> %q, but %q -> %q",
				charStr, flippedStr, flippedStr, string(reverseFlipped))
		}
	}
}

func TestFlip180Examples(t *testing.T) {
	// Test the specific example from the documentation
	// "zq.suns.bz" where zq.su + ns.bz flip to get each other

	tests := []struct {
		name  string
		part1 string
		part2 string
	}{
		{"Example from docs", "zq.su", "ns.bz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test flipping part1 gives part2
			flipped1, err := Flip180String(tt.part1)
			if err != nil {
				t.Errorf("Error flipping %q: %v", tt.part1, err)
			} else if !strings.EqualFold(flipped1, tt.part2) {
				t.Errorf("Flipping %q gave %q, expected %q", tt.part1, flipped1, tt.part2)
			}

			// Test flipping part2 gives part1
			flipped2, err := Flip180String(tt.part2)
			if err != nil {
				t.Errorf("Error flipping %q: %v", tt.part2, err)
			} else if !strings.EqualFold(flipped2, tt.part1) {
				t.Errorf("Flipping %q gave %q, expected %q", tt.part2, flipped2, tt.part1)
			}
		})
	}
}
