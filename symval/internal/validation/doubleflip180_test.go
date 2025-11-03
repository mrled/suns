package validation

import (
	"testing"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

func TestValidateDoubleFlip180(t *testing.T) {
	tests := []struct {
		name        string
		hostname1   string
		hostname2   string
		expectValid bool
	}{
		{"zq.su and ns.bz", "zq.su", "ns.bz", true},
		{"ns.bz and zq.su", "ns.bz", "zq.su", true},
		{"pods and spod", "pods", "spod", true},
		{"spod and pods", "spod", "pods", true},
		{"Not flips", "example.com", "test.org", false},
		{"One flippable", "pods", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []*model.DomainRecord{
				{
					Owner:    "test@example.com",
					Type:     symgroup.DoubleFlip180,
					Hostname: tt.hostname1,
					GroupID:  "test-group-id",
				},
				{
					Owner:    "test@example.com",
					Type:     symgroup.DoubleFlip180,
					Hostname: tt.hostname2,
					GroupID:  "test-group-id",
				},
			}

			valid, err := validateDoubleFlip180(data)
			if tt.expectValid {
				if err != nil {
					t.Errorf("Expected no error for hostnames %q and %q, got: %v", tt.hostname1, tt.hostname2, err)
				}
				if !valid {
					t.Errorf("Expected valid=true for hostnames %q and %q, got false", tt.hostname1, tt.hostname2)
				}
			} else {
				if valid {
					t.Errorf("Expected valid=false for hostnames %q and %q, got true", tt.hostname1, tt.hostname2)
				}
			}
		})
	}
}
