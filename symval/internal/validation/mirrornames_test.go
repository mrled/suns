package validation

import (
	"testing"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

// Test isMirrorPair function with various inputs
func TestIsMirrorPair(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected bool
	}{
		// Odd number of segments
		{"odd single segment", "com", "com", true},
		{"odd three segments", "a.b.com", "com.b.a", true},
		{"odd five segments", "a.b.c.d.e", "e.d.c.b.a", true},
		{"odd seven segments", "one.two.three.four.five.six.seven", "seven.six.five.four.three.two.one", true},

		// Even number of segments
		{"even two segments", "example.com", "com.example", true},
		{"even four segments", "a.b.c.d", "d.c.b.a", true},
		{"even six segments", "one.two.three.four.five.six", "six.five.four.three.two.one", true},

		// Not mirror pairs
		{"not mirror same segments", "a.b.com", "a.b.com", false},
		{"not mirror different order", "a.b.com", "b.a.com", false},
		{"not mirror partial match", "a.b.com", "com.a.b", false},

		// Different number of segments
		{"different length 1 vs 2", "com", "example.com", false},
		{"different length 2 vs 3", "a.com", "a.b.com", false},
		{"different length 3 vs 4", "a.b.com", "a.b.c.com", false},
		{"different length 4 vs 5", "one.two.three.four", "one.two.three.four.five", false},

		// Edge cases
		{"empty strings", "", "", true},
		{"single character segments", "a.b.c", "c.b.a", true},
		{"identical segments odd", "a.a.a", "a.a.a", true},
		{"identical segments even", "b.b.b.b", "b.b.b.b", true},
		{"unicode segments", "café.über.naïve", "naïve.über.café", true},
		{"numbers", "1.2.3", "3.2.1", true},
		{"mixed alphanumeric", "test1.test2.test3", "test3.test2.test1", true},

		// Real-world-like examples
		{"subdomain example", "api.example.com", "com.example.api", true},
		{"deep subdomain", "a.b.c.d.example.com", "com.example.d.c.b.a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMirrorPair(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("isMirrorPair(%q, %q) = %v, expected %v", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}

// Test validateMirrorNames with valid mirror pairs
func TestValidateMirrorNames_Success(t *testing.T) {
	tests := []struct {
		name      string
		hostname1 string
		hostname2 string
	}{
		{"two segments", "example.com", "com.example"},
		{"three segments odd", "a.b.com", "com.b.a"},
		{"four segments even", "a.b.c.d", "d.c.b.a"},
		{"five segments odd", "one.two.three.four.five", "five.four.three.two.one"},
		{"single segment", "com", "com"},
		{"unicode segments", "café.com", "com.café"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []*model.DomainRecord{
				{
					Owner:    "alice@example.com",
					Type:     symgroup.MirrorNames,
					Hostname: tt.hostname1,
					GroupID:  "test-group-id",
				},
				{
					Owner:    "alice@example.com",
					Type:     symgroup.MirrorNames,
					Hostname: tt.hostname2,
					GroupID:  "test-group-id",
				},
			}

			valid, err := validateMirrorNames(data)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if !valid {
				t.Errorf("Expected valid=true for mirror pair %q and %q", tt.hostname1, tt.hostname2)
			}
		})
	}
}

// Test validateMirrorNames with non-mirror pairs
func TestValidateMirrorNames_NotMirrorPairs(t *testing.T) {
	tests := []struct {
		name      string
		hostname1 string
		hostname2 string
	}{
		{"same hostname", "example.com", "example.com"},
		{"different order", "a.b.com", "b.a.com"},
		{"partial match", "a.b.com", "com.a.b"},
		{"different segments", "foo.bar.com", "baz.qux.com"},
		{"different length", "a.com", "a.b.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []*model.DomainRecord{
				{
					Owner:    "alice@example.com",
					Type:     symgroup.MirrorNames,
					Hostname: tt.hostname1,
					GroupID:  "test-group-id",
				},
				{
					Owner:    "alice@example.com",
					Type:     symgroup.MirrorNames,
					Hostname: tt.hostname2,
					GroupID:  "test-group-id",
				},
			}

			valid, err := validateMirrorNames(data)
			if err == nil {
				t.Errorf("Expected error for non-mirror pair %q and %q, got nil", tt.hostname1, tt.hostname2)
			}
			if valid {
				t.Errorf("Expected valid=false for non-mirror pair %q and %q", tt.hostname1, tt.hostname2)
			}
		})
	}
}

// Test validateMirrorNames expects exactly two domains
func TestValidateMirrorNames_WrongNumberOfDomains(t *testing.T) {
	tests := []struct {
		name    string
		records []*model.DomainRecord
	}{
		{
			"zero domains",
			[]*model.DomainRecord{},
		},
		{
			"one domain",
			[]*model.DomainRecord{
				{
					Owner:    "alice@example.com",
					Type:     symgroup.MirrorNames,
					Hostname: "example.com",
					GroupID:  "test-group-id",
				},
			},
		},
		{
			"three domains",
			[]*model.DomainRecord{
				{
					Owner:    "alice@example.com",
					Type:     symgroup.MirrorNames,
					Hostname: "a.com",
					GroupID:  "test-group-id",
				},
				{
					Owner:    "alice@example.com",
					Type:     symgroup.MirrorNames,
					Hostname: "com.a",
					GroupID:  "test-group-id",
				},
				{
					Owner:    "alice@example.com",
					Type:     symgroup.MirrorNames,
					Hostname: "b.com",
					GroupID:  "test-group-id",
				},
			},
		},
		{
			"four domains",
			[]*model.DomainRecord{
				{
					Owner:    "alice@example.com",
					Type:     symgroup.MirrorNames,
					Hostname: "a.b.com",
					GroupID:  "test-group-id",
				},
				{
					Owner:    "alice@example.com",
					Type:     symgroup.MirrorNames,
					Hostname: "com.b.a",
					GroupID:  "test-group-id",
				},
				{
					Owner:    "alice@example.com",
					Type:     symgroup.MirrorNames,
					Hostname: "x.y.com",
					GroupID:  "test-group-id",
				},
				{
					Owner:    "alice@example.com",
					Type:     symgroup.MirrorNames,
					Hostname: "com.y.x",
					GroupID:  "test-group-id",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateMirrorNames(tt.records)
			if err == nil {
				t.Errorf("Expected error for %d domains, got nil", len(tt.records))
			}
			if valid {
				t.Errorf("Expected valid=false for wrong number of domains")
			}
		})
	}
}
