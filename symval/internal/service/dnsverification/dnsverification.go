package dnsverification

import (
	"errors"
	"fmt"
	"net"
)

const (
	// RecordName is the TXT record label used for SUNS lookups
	RecordName = "_suns"
)

// ErrRecordNotFound is returned when the TXT record does not exist
// after checking both the direct lookup and one CNAME hop
var ErrRecordNotFound = errors.New("TXT record not found")

// Resolver is an interface for DNS lookups, allowing dependency injection
// for testing with mock implementations
type Resolver interface {
	// LookupTXT returns the TXT records for the given domain
	LookupTXT(domain string) ([]string, error)

	// LookupCNAME returns the CNAME record for the given domain
	LookupCNAME(domain string) (string, error)
}

// DefaultResolver wraps the standard library's net package
type DefaultResolver struct{}

// LookupTXT implements Resolver.LookupTXT using net.LookupTXT
func (r *DefaultResolver) LookupTXT(domain string) ([]string, error) {
	return net.LookupTXT(domain)
}

// LookupCNAME implements Resolver.LookupCNAME using net.LookupCNAME
func (r *DefaultResolver) LookupCNAME(domain string) (string, error) {
	return net.LookupCNAME(domain)
}

// Service handles TXT record lookups for SUNS
type Service struct {
	resolver Resolver
}

// NewService creates a new TXT lookup service with the default resolver
func NewService() *Service {
	return &Service{
		resolver: &DefaultResolver{},
	}
}

// NewServiceWithResolver creates a new TXT lookup service with a custom resolver
// This is useful for testing with mock resolvers
func NewServiceWithResolver(resolver Resolver) *Service {
	return &Service{
		resolver: resolver,
	}
}

// Lookup performs a TXT record lookup for the SUNS verification records of the given domain.
// It computes the label as "_suns.domain" and attempts to fetch all TXT records at that label.
// If no TXT records are found, it checks for a CNAME record at that label and performs one
// CNAME hop to re-check for TXT records at the target.
//
// Multiple TXT records are supported - all verification records found will be returned.
// This allows users to publish multiple SUNS verification records for different purposes.
//
// The single CNAME hop allows users to delegate control to another zone while keeping
// verification deterministic by limiting to one hop.
//
// Returns:
//   - All TXT record values as a slice of strings (may contain multiple verification records)
//   - ErrRecordNotFound if no records exist after checking CNAME
//   - Other errors for DNS lookup failures
func (s *Service) Lookup(domain string) ([]string, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	// Compute the label: _suns.INPUT
	label := fmt.Sprintf("%s.%s", RecordName, domain)

	// First attempt: try to fetch TXT records directly
	txtRecords, err := s.resolver.LookupTXT(label)
	if err == nil && len(txtRecords) > 0 {
		return txtRecords, nil
	}

	// Store the original error to determine if it's a "not found" case
	originalErr := err

	// Second attempt: check for CNAME and follow one hop
	cname, cnameErr := s.resolver.LookupCNAME(label)
	if cnameErr != nil {
		// No CNAME found, return the appropriate error
		if isNotFoundError(originalErr) {
			return nil, ErrRecordNotFound
		}
		return nil, fmt.Errorf("failed to lookup TXT or CNAME for %s: %w", label, originalErr)
	}

	// If CNAME exists and points to a different domain, try TXT lookup there
	if cname != "" && cname != label && cname != label+"." {
		txtRecords, err = s.resolver.LookupTXT(cname)
		if err == nil && len(txtRecords) > 0 {
			return txtRecords, nil
		}
	}

	// After CNAME hop, still no TXT record found
	if isNotFoundError(err) || isNotFoundError(originalErr) {
		return nil, ErrRecordNotFound
	}

	return nil, fmt.Errorf("failed to lookup TXT after CNAME hop: %w", err)
}

// isNotFoundError checks if the error indicates a DNS record was not found
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for standard DNS errors that indicate "not found"
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.IsNotFound
	}

	return false
}
