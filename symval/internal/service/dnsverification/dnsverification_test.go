package dnsverification

import (
	"net"
	"testing"

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

func TestLookup_DirectTXTRecord(t *testing.T) {
	mock := &MockResolver{
		TXTRecords: map[string][]string{
			"_suns.example.com": {"v1:example-data"},
		},
	}

	service := NewServiceWithResolver(mock)
	records, err := service.Lookup("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0] != "v1:example-data" {
		t.Errorf("expected 'v1:example-data', got '%s'", records[0])
	}
}

func TestLookup_MultipleTXTRecords(t *testing.T) {
	mock := &MockResolver{
		TXTRecords: map[string][]string{
			"_suns.example.com": {"record1", "record2", "record3"},
		},
	}

	service := NewServiceWithResolver(mock)
	records, err := service.Lookup("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}

	// Verify all records are present
	expected := map[string]bool{"record1": true, "record2": true, "record3": true}
	for _, record := range records {
		if !expected[record] {
			t.Errorf("unexpected record: %s", record)
		}
	}
}

func TestLookup_CNAMEHop(t *testing.T) {
	mock := &MockResolver{
		TXTRecords: map[string][]string{
			"delegation.example.net": {"v1:delegated-data"},
		},
		CNAMERecords: map[string]string{
			"_suns.example.com": "delegation.example.net",
		},
	}

	service := NewServiceWithResolver(mock)
	records, err := service.Lookup("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0] != "v1:delegated-data" {
		t.Errorf("expected 'v1:delegated-data', got '%s'", records[0])
	}
}

func TestLookup_CNAMEHopWithMultipleRecords(t *testing.T) {
	// Test that multiple verification records are returned when following a CNAME
	mock := &MockResolver{
		TXTRecords: map[string][]string{
			"verification.example.net": {
				"v1:groupid1:hash1",
				"v1:groupid2:hash2",
				"v1:groupid3:hash3",
			},
		},
		CNAMERecords: map[string]string{
			"_suns.example.com": "verification.example.net",
		},
	}

	service := NewServiceWithResolver(mock)
	records, err := service.Lookup("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("expected 3 records via CNAME, got %d", len(records))
	}

	// Verify all records are present
	expected := map[string]bool{
		"v1:groupid1:hash1": true,
		"v1:groupid2:hash2": true,
		"v1:groupid3:hash3": true,
	}
	for _, record := range records {
		if !expected[record] {
			t.Errorf("unexpected record: %s", record)
		}
	}
}

func TestLookup_CNAMEWithTrailingDot(t *testing.T) {
	// Test that CNAME with trailing dot doesn't cause infinite recursion
	mock := &MockResolver{
		TXTRecords: map[string][]string{},
		CNAMERecords: map[string]string{
			"_suns.example.com": "_suns.example.com.",
		},
	}

	service := NewServiceWithResolver(mock)
	records, err := service.Lookup("example.com")

	// Should get empty list, not hang or error differently
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected empty list, got %d records", len(records))
	}
}

func TestLookup_RecordNotFound(t *testing.T) {
	mock := &MockResolver{
		TXTRecords:   map[string][]string{},
		CNAMERecords: map[string]string{},
	}

	service := NewServiceWithResolver(mock)
	records, err := service.Lookup("nonexistent.example.com")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected empty list, got %d records", len(records))
	}
}

func TestLookup_EmptyDomain(t *testing.T) {
	mock := &MockResolver{
		TXTRecords: map[string][]string{},
	}

	service := NewServiceWithResolver(mock)
	_, err := service.Lookup("")

	if err == nil {
		t.Fatal("expected error for empty domain")
	}
}

func TestLookup_LabelConstruction(t *testing.T) {
	// Verify that the service correctly constructs _suns.domain
	mock := &MockResolver{
		TXTRecords: map[string][]string{
			"_suns.subdomain.example.com": {"found"},
		},
	}

	service := NewServiceWithResolver(mock)
	records, err := service.Lookup("subdomain.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 1 || records[0] != "found" {
		t.Errorf("label construction failed")
	}
}

func TestLookup_PreferDirectOverCNAME(t *testing.T) {
	// If both TXT and CNAME exist, TXT should be preferred
	mock := &MockResolver{
		TXTRecords: map[string][]string{
			"_suns.example.com":  {"direct-record"},
			"other.example.net": {"cname-record"},
		},
		CNAMERecords: map[string]string{
			"_suns.example.com": "other.example.net",
		},
	}

	service := NewServiceWithResolver(mock)
	records, err := service.Lookup("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	if records[0] != "direct-record" {
		t.Errorf("expected direct TXT record to be preferred, got '%s'", records[0])
	}
}

func TestLookup_CNAMEPointsToSelf(t *testing.T) {
	// CNAME pointing to itself should not cause issues
	mock := &MockResolver{
		TXTRecords: map[string][]string{},
		CNAMERecords: map[string]string{
			"_suns.example.com": "_suns.example.com",
		},
	}

	service := NewServiceWithResolver(mock)
	records, err := service.Lookup("example.com")

	if err != nil {
		t.Errorf("expected no error for self-referencing CNAME, got %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected empty list, got %d records", len(records))
	}
}

func TestLookup_DNSErrorHandling(t *testing.T) {
	t.Run("temporary DNS error", func(t *testing.T) {
		mock := &MockResolver{
			TXTError: &net.DNSError{
				Err:         "temporary failure",
				Name:        "_suns.example.com",
				IsTemporary: true,
			},
		}

		service := NewServiceWithResolver(mock)
		_, err := service.Lookup("example.com")

		if err == nil {
			t.Fatal("expected error for temporary DNS failure")
		}
	})

	t.Run("timeout DNS error", func(t *testing.T) {
		mock := &MockResolver{
			TXTError: &net.DNSError{
				Err:       "timeout",
				Name:      "_suns.example.com",
				IsTimeout: true,
			},
		}

		service := NewServiceWithResolver(mock)
		_, err := service.Lookup("example.com")

		if err == nil {
			t.Fatal("expected error for DNS timeout")
		}
	})
}

func TestLookup_OnlyOneCNAMEHop(t *testing.T) {
	// Verify that only one CNAME hop is performed
	// We can't directly test this without observing resolver calls,
	// but we can verify the behavior: second-level CNAME is not followed
	mock := &MockResolver{
		TXTRecords: map[string][]string{
			"final.example.org": {"should-not-reach"},
		},
		CNAMERecords: map[string]string{
			"_suns.example.com":    "middle.example.net",
			"middle.example.net":   "final.example.org",
		},
	}

	service := NewServiceWithResolver(mock)
	records, err := service.Lookup("example.com")

	// Since we only do one CNAME hop, and middle.example.net has no TXT,
	// we should get empty list
	if err != nil {
		t.Errorf("expected no error since only one hop is allowed, got %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected empty list, got %d records", len(records))
	}
}

func TestNewService(t *testing.T) {
	// Test that NewService creates a service with default resolver
	service := NewService()
	if service == nil {
		t.Fatal("NewService returned nil")
	}
	if service.resolver == nil {
		t.Fatal("NewService created service with nil resolver")
	}
}

func TestRecordNameConstant(t *testing.T) {
	// Verify the constant value
	if RecordName != "_suns" {
		t.Errorf("RecordName should be '_suns', got '%s'", RecordName)
	}
}

func TestLookup_MultipleVerificationRecordsRealistic(t *testing.T) {
	// Test a realistic scenario with multiple SUNS verification records
	// representing different group memberships
	mock := &MockResolver{
		TXTRecords: map[string][]string{
			"_suns.myapp.example.com": {
				"v1:team-alpha:YWxwaGEtdGVhbS12ZXJpZmljYXRpb24taGFzaA==",
				"v1:team-beta:YmV0YS10ZWFtLXZlcmlmaWNhdGlvbi1oYXNo",
				"v1:team-gamma:Z2FtbWEtdGVhbS12ZXJpZmljYXRpb24taGFzaA==",
				"v1:org-wide:b3JnLXdpZGUtdmVyaWZpY2F0aW9uLWhhc2g=",
			},
		},
	}

	service := NewServiceWithResolver(mock)
	records, err := service.Lookup("myapp.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 4 {
		t.Fatalf("expected 4 verification records, got %d", len(records))
	}

	// Verify specific records exist
	hasTeamAlpha := false
	hasTeamBeta := false
	hasTeamGamma := false
	hasOrgWide := false

	for _, record := range records {
		switch record {
		case "v1:team-alpha:YWxwaGEtdGVhbS12ZXJpZmljYXRpb24taGFzaA==":
			hasTeamAlpha = true
		case "v1:team-beta:YmV0YS10ZWFtLXZlcmlmaWNhdGlvbi1oYXNo":
			hasTeamBeta = true
		case "v1:team-gamma:Z2FtbWEtdGVhbS12ZXJpZmljYXRpb24taGFzaA==":
			hasTeamGamma = true
		case "v1:org-wide:b3JnLXdpZGUtdmVyaWZpY2F0aW9uLWhhc2g=":
			hasOrgWide = true
		}
	}

	if !hasTeamAlpha {
		t.Error("missing team-alpha verification record")
	}
	if !hasTeamBeta {
		t.Error("missing team-beta verification record")
	}
	if !hasTeamGamma {
		t.Error("missing team-gamma verification record")
	}
	if !hasOrgWide {
		t.Error("missing org-wide verification record")
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
		service := NewServiceWithResolver(mock)

		gids, err := service.VerifyDomain("example.com")
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
		service := NewServiceWithResolver(mock)

		gids, err := service.VerifyDomain("example.com")
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
		service := NewServiceWithResolver(mock)

		gids, err := service.VerifyDomain("example.com")
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
		service := NewServiceWithResolver(mock)

		_, err := service.VerifyDomain("example.com")
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
		service := NewServiceWithResolver(mock)

		_, err := service.VerifyDomain("example.com")
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
		service := NewServiceWithResolver(mock)

		_, err := service.VerifyDomain("example.com")
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
		service := NewServiceWithResolver(mock)

		_, err := service.VerifyDomain("example.com")
		if err == nil {
			t.Fatal("expected error for DNS lookup failure")
		}
	})
}
