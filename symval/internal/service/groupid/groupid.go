package groupid

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
)

const (
	// IDVersion is the current version of the group ID algorithm
	IDVersion = "v1"
)

// Service handles group ID calculation
type Service struct{}

// NewService creates a new group ID service
func NewService() *Service {
	return &Service{}
}

// CalculateV1 generates a group ID by hashing owner and hostnames separately
// The result is formatted as: idversion:type:base64(sha256(owner)):base64(sha256(sort(hostnames))).
func (s *Service) CalculateV1(owner, gtype string, hostnames []string) (string, error) {
	if owner == "" {
		return "", fmt.Errorf("owner cannot be empty")
	}
	if gtype == "" {
		return "", fmt.Errorf("type cannot be empty")
	}
	if len(hostnames) == 0 {
		return "", fmt.Errorf("at least one hostname is required")
	}

	// Sort hostnames for consistent hashing
	sorted := make([]string, len(hostnames))
	copy(sorted, hostnames)
	sort.Strings(sorted)

	// Hash the owner
	ownerHash := sha256.Sum256([]byte(owner))
	ownerEncoded := base64.StdEncoding.EncodeToString(ownerHash[:])

	// Build the string to hash: all sorted hostnames
	var builder strings.Builder
	for _, hostname := range sorted {
		builder.WriteString(hostname)
	}

	// Hash the hostnames
	hostnamesHash := sha256.Sum256([]byte(builder.String()))
	hostnamesEncoded := base64.StdEncoding.EncodeToString(hostnamesHash[:])

	// Format: idversion:type:base64(sha256(owner)):base64(sha256(sort(hostnames)))
	groupID := fmt.Sprintf("%s:%s:%s:%s", IDVersion, gtype, ownerEncoded, hostnamesEncoded)

	return groupID, nil
}
