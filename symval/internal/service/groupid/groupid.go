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

// CalculateV1 generates a group ID by hashing owner + all hostnames
// The result is formatted as: idversion:type:base64(sha256(owner+sort(hostnames))).
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

	// Build the string to hash: owner + all hostnames
	var builder strings.Builder
	builder.WriteString(owner)
	for _, hostname := range sorted {
		builder.WriteString(hostname)
	}

	// Hash the combined string
	hash := sha256.Sum256([]byte(builder.String()))

	// Base64 encode the hash
	encoded := base64.StdEncoding.EncodeToString(hash[:])

	// Format: idversion:type:base64hash
	groupID := fmt.Sprintf("%s:%s:%s", IDVersion, gtype, encoded)

	return groupID, nil
}
