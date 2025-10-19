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

// GroupIDV1 represents a parsed v1 group ID
type GroupIDV1 struct {
	Version     string
	TypeCode    string
	OwnerHash   string
	DomainsHash string
	Raw         string
}

// String returns the raw group ID string
func (g GroupIDV1) String() string {
	return g.Raw
}

// ParseGroupIDv1 parses a raw group ID string into a GroupIDV1 struct.
// The expected format is: v1:typecode:ownerhash:domainshash
// Returns an error if the format is invalid or the version is not v1.
func ParseGroupIDv1(raw string) (GroupIDV1, error) {
	if raw == "" {
		return GroupIDV1{}, fmt.Errorf("group ID cannot be empty")
	}

	parts := strings.Split(raw, ":")
	if len(parts) != 4 {
		return GroupIDV1{}, fmt.Errorf("invalid group ID format: expected 4 colon-separated parts, got %d", len(parts))
	}

	version := parts[0]
	if version != "v1" {
		return GroupIDV1{}, fmt.Errorf("unsupported group ID version: %s (expected v1)", version)
	}

	return GroupIDV1{
		Version:     parts[0],
		TypeCode:    parts[1],
		OwnerHash:   parts[2],
		DomainsHash: parts[3],
		Raw:         raw,
	}, nil
}

// ParseGroupIDv1Slice parses a slice of raw group ID strings into a slice of GroupIDV1 structs.
func ParseGroupIDv1Slice(records []string) ([]GroupIDV1, error) {
	groupIDs := make([]GroupIDV1, 0, len(records))
	for i, record := range records {
		gid, err := ParseGroupIDv1(record)
		if err != nil {
			return nil, fmt.Errorf("failed to parse record at index %d: %w", i, err)
		}
		groupIDs = append(groupIDs, gid)
	}
	return groupIDs, nil
}

// CalculateV1 generates a group ID by hashing owner and hostnames separately
// The result is formatted as: idversion:type:base64(sha256(owner)):base64(sha256(sort(hostnames))).
func CalculateV1(owner, gtype string, hostnames []string) (string, error) {
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
