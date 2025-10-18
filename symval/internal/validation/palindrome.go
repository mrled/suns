package validation

import (
	"fmt"

	"github.com/mrled/suns/symval/internal/model"
)

// isPalindrome checks if a string is a palindrome.
// It compares characters from both ends, working towards the middle.
// Works with both ASCII and Unicode characters.
func isPalindrome(s string) bool {
	// Convert string to rune slice to properly handle Unicode characters
	runes := []rune(s)
	length := len(runes)

	// Compare characters from both ends
	for i := 0; i < length/2; i++ {
		if runes[i] != runes[length-1-i] {
			return false
		}
	}

	return true
}

// validatePalindrome validates palindrome symmetry
func validatePalindrome(data []*model.DomainData) (bool, error) {
	if len(data) != 1 {
		return false, fmt.Errorf("palindrome validation expects exactly one domain, got %d", len(data))
	}

	hostname := data[0].Hostname
	if !isPalindrome(hostname) {
		return false, fmt.Errorf("hostname %q is not a palindrome", hostname)
	}

	return true, nil
}
