package usecase

import (
	"net"
	"testing"

	"github.com/mrled/suns/symval/internal/service/dnsverification"
	"github.com/mrled/suns/symval/internal/service/groupid"
)

// MockResolver is a mock implementation of the Resolver interface for testing
type MockResolver struct {
	// TXTRecords maps domain names to their TXT records
	TXTRecords map[string][]string
	// CNAMERecords maps domain names to their CNAME targets
	CNAMERecords map[string]string
	// TXTError is returned when LookupTXT is called (if set)
	TXTError error
	// CNAMEError is returned when LookupCNAME is called (if set)
	CNAMEError error
}

// LookupTXT returns mocked TXT records
func (m *MockResolver) LookupTXT(domain string) ([]string, error) {
	if m.TXTError != nil {
		return nil, m.TXTError
	}
	if records, ok := m.TXTRecords[domain]; ok {
		return records, nil
	}
	// Return a "not found" DNS error
	return nil, &net.DNSError{
		Err:        "no such host",
		Name:       domain,
		IsNotFound: true,
	}
}

// LookupCNAME returns mocked CNAME records
func (m *MockResolver) LookupCNAME(domain string) (string, error) {
	if m.CNAMEError != nil {
		return "", m.CNAMEError
	}
	if cname, ok := m.CNAMERecords[domain]; ok {
		return cname, nil
	}
	// Return a "not found" DNS error
	return "", &net.DNSError{
		Err:        "no such host",
		Name:       domain,
		IsNotFound: true,
	}
}

func TestVerify(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		err := Verify([]groupid.GroupIDV1{})
		if err != nil {
			t.Errorf("expected no error for empty slice, got %v", err)
		}
	})

	t.Run("single group ID", func(t *testing.T) {
		gid := groupid.GroupIDV1{
			Version:     "v1",
			TypeCode:    "type1",
			OwnerHash:   "hash123",
			DomainsHash: "domainhash1",
			Raw:         "v1:type1:hash123:domainhash1",
		}
		err := Verify([]groupid.GroupIDV1{gid})
		if err != nil {
			t.Errorf("expected no error for single group ID, got %v", err)
		}
	})

	t.Run("multiple group IDs with same owner hash", func(t *testing.T) {
		gids := []groupid.GroupIDV1{
			{
				Version:     "v1",
				TypeCode:    "type1",
				OwnerHash:   "sameowner",
				DomainsHash: "domains1",
				Raw:         "v1:type1:sameowner:domains1",
			},
			{
				Version:     "v1",
				TypeCode:    "type2",
				OwnerHash:   "sameowner",
				DomainsHash: "domains2",
				Raw:         "v1:type2:sameowner:domains2",
			},
			{
				Version:     "v1",
				TypeCode:    "type3",
				OwnerHash:   "sameowner",
				DomainsHash: "domains3",
				Raw:         "v1:type3:sameowner:domains3",
			},
		}
		err := Verify(gids)
		if err != nil {
			t.Errorf("expected no error for consistent owner hashes, got %v", err)
		}
	})

	t.Run("multiple group IDs with different owner hashes", func(t *testing.T) {
		gids := []groupid.GroupIDV1{
			{
				Version:     "v1",
				TypeCode:    "type1",
				OwnerHash:   "owner1",
				DomainsHash: "domains1",
				Raw:         "v1:type1:owner1:domains1",
			},
			{
				Version:     "v1",
				TypeCode:    "type2",
				OwnerHash:   "owner2",
				DomainsHash: "domains2",
				Raw:         "v1:type2:owner2:domains2",
			},
		}
		err := Verify(gids)
		if err == nil {
			t.Fatal("expected error for inconsistent owner hashes")
		}
	})

	t.Run("third group ID has different owner", func(t *testing.T) {
		gids := []groupid.GroupIDV1{
			{
				Version:     "v1",
				TypeCode:    "type1",
				OwnerHash:   "sameowner",
				DomainsHash: "domains1",
				Raw:         "v1:type1:sameowner:domains1",
			},
			{
				Version:     "v1",
				TypeCode:    "type2",
				OwnerHash:   "sameowner",
				DomainsHash: "domains2",
				Raw:         "v1:type2:sameowner:domains2",
			},
			{
				Version:     "v1",
				TypeCode:    "type3",
				OwnerHash:   "differentowner",
				DomainsHash: "domains3",
				Raw:         "v1:type3:differentowner:domains3",
			},
		}
		err := Verify(gids)
		if err == nil {
			t.Fatal("expected error when third group ID has different owner")
		}
	})
}

func TestVerifyDomain(t *testing.T) {
	t.Run("no records found", func(t *testing.T) {
		mock := &MockResolver{
			TXTRecords: map[string][]string{},
		}
		dnsService := dnsverification.NewServiceWithResolver(mock)
		verifyUC := NewVerifyUseCase(dnsService)

		gids, err := verifyUC.VerifyDomain("example.com")
		if err != nil {
			t.Fatalf("expected no error for no records, got %v", err)
		}
		if len(gids) != 0 {
			t.Errorf("expected empty slice, got %d group IDs", len(gids))
		}
	})

	t.Run("single valid group ID", func(t *testing.T) {
		mock := &MockResolver{
			TXTRecords: map[string][]string{
				"_suns.example.com": {"v1:type1:ownerhash123:domainshash456"},
			},
		}
		dnsService := dnsverification.NewServiceWithResolver(mock)
		verifyUC := NewVerifyUseCase(dnsService)

		gids, err := verifyUC.VerifyDomain("example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(gids) != 1 {
			t.Fatalf("expected 1 group ID, got %d", len(gids))
		}
		if gids[0].TypeCode != "type1" {
			t.Errorf("expected type 'type1', got '%s'", gids[0].TypeCode)
		}
		if gids[0].OwnerHash != "ownerhash123" {
			t.Errorf("expected owner hash 'ownerhash123', got '%s'", gids[0].OwnerHash)
		}
	})

	t.Run("multiple group IDs with consistent owner", func(t *testing.T) {
		mock := &MockResolver{
			TXTRecords: map[string][]string{
				"_suns.example.com": {
					"v1:type1:sameowner:domains1",
					"v1:type2:sameowner:domains2",
					"v1:type3:sameowner:domains3",
				},
			},
		}
		dnsService := dnsverification.NewServiceWithResolver(mock)
		verifyUC := NewVerifyUseCase(dnsService)

		gids, err := verifyUC.VerifyDomain("example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(gids) != 3 {
			t.Fatalf("expected 3 group IDs, got %d", len(gids))
		}

		// Verify all have the same owner hash
		for i, gid := range gids {
			if gid.OwnerHash != "sameowner" {
				t.Errorf("group ID %d has unexpected owner hash: %s", i, gid.OwnerHash)
			}
		}
	})

	t.Run("multiple group IDs with inconsistent owner", func(t *testing.T) {
		mock := &MockResolver{
			TXTRecords: map[string][]string{
				"_suns.example.com": {
					"v1:type1:owner1:domains1",
					"v1:type2:owner2:domains2",
				},
			},
		}
		dnsService := dnsverification.NewServiceWithResolver(mock)
		verifyUC := NewVerifyUseCase(dnsService)

		_, err := verifyUC.VerifyDomain("example.com")
		if err == nil {
			t.Fatal("expected error for inconsistent owner hashes")
		}
	})

	t.Run("invalid group ID format", func(t *testing.T) {
		mock := &MockResolver{
			TXTRecords: map[string][]string{
				"_suns.example.com": {"invalid-format"},
			},
		}
		dnsService := dnsverification.NewServiceWithResolver(mock)
		verifyUC := NewVerifyUseCase(dnsService)

		_, err := verifyUC.VerifyDomain("example.com")
		if err == nil {
			t.Fatal("expected error for invalid group ID format")
		}
	})

	t.Run("mixed valid and invalid formats", func(t *testing.T) {
		mock := &MockResolver{
			TXTRecords: map[string][]string{
				"_suns.example.com": {
					"v1:type1:owner:domains",
					"invalid",
				},
			},
		}
		dnsService := dnsverification.NewServiceWithResolver(mock)
		verifyUC := NewVerifyUseCase(dnsService)

		_, err := verifyUC.VerifyDomain("example.com")
		if err == nil {
			t.Fatal("expected error when parsing mixed valid/invalid formats")
		}
	})

	t.Run("DNS lookup error", func(t *testing.T) {
		mock := &MockResolver{
			TXTError: &net.DNSError{
				Err:         "temporary failure",
				Name:        "_suns.example.com",
				IsTemporary: true,
			},
		}
		dnsService := dnsverification.NewServiceWithResolver(mock)
		verifyUC := NewVerifyUseCase(dnsService)

		_, err := verifyUC.VerifyDomain("example.com")
		if err == nil {
			t.Fatal("expected error for DNS lookup failure")
		}
	})
}
