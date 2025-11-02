package validation

import (
	"fmt"
	"strings"

	"github.com/mrled/suns/symval/internal/model"
)

// flip180Mapping maps ASCII characters to their 180-degree rotated equivalents
// Only characters that have a meaningful visual rotation are included
var flip180Mapping = map[rune]rune{
	// Lowercase letters
	'b': 'q',
	'd': 'p',
	'l': 'l',
	'n': 'u',
	'o': 'o',
	'p': 'd',
	'q': 'b',
	's': 's',
	'u': 'n',
	'x': 'x',
	'z': 'z',

	// Uppercase letters
	// 'H': 'H',
	// 'I': 'I',
	// 'M': 'W',
	// 'N': 'N',
	// 'O': 'O',
	// 'S': 'S',
	// 'W': 'M',
	// 'X': 'X',
	// 'Z': 'Z',

	// Numbers
	'0': '0',
	'1': '1',
	'6': '9',
	'8': '8',
	'9': '6',

	// Special characters
	'.': '.', // Treat the period as itself
	'-': '-',
}

// Flip180String returns the 180-degree rotated version of a string
// Exported for use in doubleflip180 validation
func Flip180String(s string) (string, error) {
	// Convert to lowercase for comparison
	s = strings.ToLower(s)
	runes := []rune(s)
	result := make([]rune, len(runes))

	// Process string in reverse order (180 rotation also reverses the string)
	for i := 0; i < len(runes); i++ {
		char := runes[len(runes)-1-i]
		if flipped, ok := flip180Mapping[char]; ok {
			result[i] = flipped
		} else {
			// Character cannot be flipped
			return "", fmt.Errorf("character '%c' cannot be rotated 180 degrees", char)
		}
	}

	return string(result), nil
}

// isFlip180 checks if a string is identical to its 180-degree rotated version
func isFlip180(s string) bool {
	flipped, err := Flip180String(s)
	if err != nil {
		return false
	}
	return strings.ToLower(flipped) == strings.ToLower(s)
}

// validateFlip180 validates 180-degree flip symmetry
func validateFlip180(data []*model.DomainRecord) (bool, error) {
	if len(data) != 1 {
		return false, fmt.Errorf("flip180 validation expects exactly one domain, got %d", len(data))
	}

	hostname := data[0].Hostname
	if !isFlip180(hostname) {
		return false, fmt.Errorf("hostname %q does not have 180-degree flip symmetry", hostname)
	}

	return true, nil
}
