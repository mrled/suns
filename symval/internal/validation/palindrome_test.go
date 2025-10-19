package validation

import (
	"testing"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

// Test isPalindrome function with various inputs
func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Latin characters - odd length
		{"odd length palindrome", "racecar", true},
		{"odd length palindrome uppercase", "RACECAR", true},
		{"odd length single char", "a", true},
		{"odd length three chars", "aba", true},

		// Latin characters - even length
		{"even length palindrome", "noon", true},
		{"even length two chars", "aa", true},
		{"even length four chars", "abba", true},

		// Non-palindromes
		{"not a palindrome", "hello", false},
		{"almost palindrome", "raceca", false},
		{"reversed not same", "abcd", false},

		// Non-Latin characters
		{"unicode palindrome odd", "אבא", true},
		{"unicode palindrome even", "הללה", true},
		{"emoji palindrome", "🎉🎈🎉", true},
		{"mixed unicode palindrome", "καιακ", true},
		{"japanese palindrome", "たけやぶやけた", true},
		{"unicode not palindrome", "αβγ", false},

		// Dots and special characters
		{"with dots palindrome", "a.b.a", true},
		{"with dots not palindrome", "a.b.c", false},
		{"dots only palindrome", "...", true},
		{"dots only even", "..", true},
		{"mixed special chars palindrome", "a-b-a", true},
		{"domain-like palindrome", "a.a", true},

		// Edge cases
		{"empty string", "", true},
		{"whitespace palindrome", " ", true},
		{"number palindrome", "12321", true},
		{"alphanumeric palindrome", "a1b1a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPalindrome(tt.input)
			if result != tt.expected {
				t.Errorf("isPalindrome(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Test validatePalindrome with valid single domain
func TestValidatePalindrome_Success(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
	}{
		{"latin odd length", "racecar"},
		{"latin even length", "noon"},
		{"with dots", "a.b.a"},
		{"unicode", "αβα"},
		{"single character", "a"},
		{"two characters", "aa"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []*model.DomainRecord{
				{
					Owner:    "alice@example.com",
					Type:     symgroup.Palindrome,
					Hostname: tt.hostname,
					GroupID:  "test-group-id",
				},
			}

			valid, err := validatePalindrome(data)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
			if !valid {
				t.Errorf("Expected valid=true for palindrome %q", tt.hostname)
			}
		})
	}
}

// Test validatePalindrome with non-palindrome hostnames
func TestValidatePalindrome_NotPalindrome(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
	}{
		{"not palindrome", "example.com"},
		{"reversed not same", "abcd.org"},
		{"unicode not palindrome", "αβγ.example"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []*model.DomainRecord{
				{
					Owner:    "alice@example.com",
					Type:     symgroup.Palindrome,
					Hostname: tt.hostname,
					GroupID:  "test-group-id",
				},
			}

			valid, err := validatePalindrome(data)
			if err == nil {
				t.Errorf("Expected error for non-palindrome %q, got nil", tt.hostname)
			}
			if valid {
				t.Errorf("Expected valid=false for non-palindrome %q", tt.hostname)
			}
		})
	}
}

// Test validatePalindrome expects exactly one domain
func TestValidatePalindrome_WrongNumberOfDomains(t *testing.T) {
	tests := []struct {
		name    string
		records []*model.DomainRecord
	}{
		{
			"zero domains",
			[]*model.DomainRecord{},
		},
		{
			"two domains",
			[]*model.DomainRecord{
				{
					Owner:    "alice@example.com",
					Type:     symgroup.Palindrome,
					Hostname: "racecar.com",
					GroupID:  "test-group-id",
				},
				{
					Owner:    "alice@example.com",
					Type:     symgroup.Palindrome,
					Hostname: "noon.com",
					GroupID:  "test-group-id",
				},
			},
		},
		{
			"three domains",
			[]*model.DomainRecord{
				{
					Owner:    "alice@example.com",
					Type:     symgroup.Palindrome,
					Hostname: "aba.com",
					GroupID:  "test-group-id",
				},
				{
					Owner:    "alice@example.com",
					Type:     symgroup.Palindrome,
					Hostname: "noon.com",
					GroupID:  "test-group-id",
				},
				{
					Owner:    "alice@example.com",
					Type:     symgroup.Palindrome,
					Hostname: "racecar.com",
					GroupID:  "test-group-id",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validatePalindrome(tt.records)
			if err == nil {
				t.Errorf("Expected error for %d domains, got nil", len(tt.records))
			}
			if valid {
				t.Errorf("Expected valid=false for wrong number of domains")
			}
		})
	}
}
