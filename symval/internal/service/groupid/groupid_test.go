package groupid

import (
	"strings"
	"testing"
)

func TestCalculateV1(t *testing.T) {
	service := NewService()

	t.Run("basic calculation", func(t *testing.T) {
		groupID, err := service.CalculateV1("owner1", "type1", []string{"host1.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check format: v1:type:base64hash:base64hash
		if len(groupID) == 0 {
			t.Fatal("groupID should not be empty")
		}

		// Verify it starts with v1:type1:
		expected := "v1:type1:"
		if len(groupID) < len(expected) || groupID[:len(expected)] != expected {
			t.Errorf("groupID should start with %q, got %q", expected, groupID)
		}

		// Verify it has 4 colon-separated parts
		parts := len(groupID) - len(strings.ReplaceAll(groupID, ":", ""))
		if parts != 3 {
			t.Errorf("groupID should have 4 colon-separated parts (3 colons), got %d colons", parts)
		}
	})

	t.Run("multiple hostnames", func(t *testing.T) {
		groupID, err := service.CalculateV1("owner1", "type1", []string{"host1.example.com", "host2.example.com", "host3.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(groupID) == 0 {
			t.Fatal("groupID should not be empty")
		}
	})

	t.Run("hostname order independence", func(t *testing.T) {
		// Calculate with hostnames in one order
		groupID1, err := service.CalculateV1("owner1", "type1", []string{"host1.example.com", "host2.example.com", "host3.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Calculate with hostnames in different order
		groupID2, err := service.CalculateV1("owner1", "type1", []string{"host3.example.com", "host1.example.com", "host2.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should produce the same result
		if groupID1 != groupID2 {
			t.Errorf("groupIDs should be identical regardless of hostname order\ngot:  %s\nwant: %s", groupID2, groupID1)
		}
	})

	t.Run("different owners produce different IDs", func(t *testing.T) {
		groupID1, err := service.CalculateV1("owner1", "type1", []string{"host1.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		groupID2, err := service.CalculateV1("owner2", "type1", []string{"host1.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if groupID1 == groupID2 {
			t.Errorf("different owners should produce different group IDs")
		}
	})

	t.Run("different types produce different IDs", func(t *testing.T) {
		groupID1, err := service.CalculateV1("owner1", "type1", []string{"host1.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		groupID2, err := service.CalculateV1("owner1", "type2", []string{"host1.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// The type is in the prefix, so they should be different
		if groupID1 == groupID2 {
			t.Errorf("different types should produce different group IDs\ngot:  %s\nwant: different from %s", groupID2, groupID1)
		}
	})

	t.Run("different hostnames produce different IDs", func(t *testing.T) {
		groupID1, err := service.CalculateV1("owner1", "type1", []string{"host1.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		groupID2, err := service.CalculateV1("owner1", "type1", []string{"host2.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if groupID1 == groupID2 {
			t.Errorf("different hostnames should produce different group IDs")
		}
	})

	t.Run("deterministic results", func(t *testing.T) {
		// Calculate the same group ID multiple times
		groupID1, err := service.CalculateV1("owner1", "type1", []string{"host1.example.com", "host2.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		groupID2, err := service.CalculateV1("owner1", "type1", []string{"host1.example.com", "host2.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if groupID1 != groupID2 {
			t.Errorf("same inputs should produce identical results\ngot:  %s\nwant: %s", groupID2, groupID1)
		}
	})
}

func TestCalculateV1_Errors(t *testing.T) {
	service := NewService()

	t.Run("empty owner", func(t *testing.T) {
		_, err := service.CalculateV1("", "type1", []string{"host1.example.com"})
		if err == nil {
			t.Fatal("expected error for empty owner")
		}
	})

	t.Run("empty type", func(t *testing.T) {
		_, err := service.CalculateV1("owner1", "", []string{"host1.example.com"})
		if err == nil {
			t.Fatal("expected error for empty type")
		}
	})

	t.Run("no hostnames", func(t *testing.T) {
		_, err := service.CalculateV1("owner1", "type1", []string{})
		if err == nil {
			t.Fatal("expected error for empty hostnames")
		}
	})

	t.Run("nil hostnames", func(t *testing.T) {
		_, err := service.CalculateV1("owner1", "type1", nil)
		if err == nil {
			t.Fatal("expected error for nil hostnames")
		}
	})
}

func TestCalculateV1_KnownValues(t *testing.T) {
	service := NewService()

	tests := []struct {
		name      string
		owner     string
		gtype     string
		hostnames []string
		want      string
	}{
		{
			name:      "single hostname",
			owner:     "myowner",
			gtype:     "mytype",
			hostnames: []string{"host1.example.com"},
			want:      "v1:mytype:ONhEevmGtSryy82u9a14bIzvtB3rpWzExC0atTB5ATI=:Gwha2Fxaavzv1ZfiQ+kOkXkprDhIaHnHDjRcd/RRZqM=",
		},
		{
			name:      "multiple hostnames",
			owner:     "myowner",
			gtype:     "mytype",
			hostnames: []string{"host1.example.com", "host2.example.com", "host3.example.com"},
			want:      "v1:mytype:ONhEevmGtSryy82u9a14bIzvtB3rpWzExC0atTB5ATI=:i+9aaIps5MPVnI4+JCWDQJvOYiKx4iRQ9PLp+vAxqdE=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.CalculateV1(tt.owner, tt.gtype, tt.hostnames)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("CalculateV1() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseGroupIDv1(t *testing.T) {
	t.Run("valid group ID", func(t *testing.T) {
		raw := "v1:mytype:ONhEevmGtSryy82u9a14bIzvtB3rpWzExC0atTB5ATI=:Gwha2Fxaavzv1ZfiQ+kOkXkprDhIaHnHDjRcd/RRZqM="
		gid, err := ParseGroupIDv1(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if gid.Version != "v1" {
			t.Errorf("expected version 'v1', got '%s'", gid.Version)
		}
		if gid.TypeCode != "mytype" {
			t.Errorf("expected type 'mytype', got '%s'", gid.TypeCode)
		}
		if gid.OwnerHash != "ONhEevmGtSryy82u9a14bIzvtB3rpWzExC0atTB5ATI=" {
			t.Errorf("unexpected owner hash: %s", gid.OwnerHash)
		}
		if gid.DomainsHash != "Gwha2Fxaavzv1ZfiQ+kOkXkprDhIaHnHDjRcd/RRZqM=" {
			t.Errorf("unexpected domains hash: %s", gid.DomainsHash)
		}
		if gid.Raw != raw {
			t.Errorf("expected raw to be '%s', got '%s'", raw, gid.Raw)
		}
		if gid.String() != raw {
			t.Errorf("String() should return raw value")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := ParseGroupIDv1("")
		if err == nil {
			t.Fatal("expected error for empty string")
		}
	})

	t.Run("invalid format - too few parts", func(t *testing.T) {
		_, err := ParseGroupIDv1("v1:type:hash")
		if err == nil {
			t.Fatal("expected error for too few parts")
		}
	})

	t.Run("invalid format - too many parts", func(t *testing.T) {
		_, err := ParseGroupIDv1("v1:type:hash1:hash2:extra")
		if err == nil {
			t.Fatal("expected error for too many parts")
		}
	})

	t.Run("unsupported version", func(t *testing.T) {
		_, err := ParseGroupIDv1("v2:type:hash1:hash2")
		if err == nil {
			t.Fatal("expected error for unsupported version")
		}
	})

	t.Run("roundtrip with CalculateV1", func(t *testing.T) {
		service := NewService()
		groupID, err := service.CalculateV1("owner1", "type1", []string{"host1.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		parsed, err := ParseGroupIDv1(groupID)
		if err != nil {
			t.Fatalf("failed to parse generated group ID: %v", err)
		}

		if parsed.Version != "v1" {
			t.Errorf("expected version v1, got %s", parsed.Version)
		}
		if parsed.TypeCode != "type1" {
			t.Errorf("expected type 'type1', got %s", parsed.TypeCode)
		}
		if parsed.String() != groupID {
			t.Errorf("String() should return original group ID")
		}
	})
}

func TestCalculateV1_TXTRecordValidity(t *testing.T) {
	service := NewService()

	// WARNING: DO NOT CHANGE THIS TEST WITHOUT CONSIDERING DNS TXT RECORD LIMITATIONS.
	// A single string in a DNS TXT record can hold a maximum of 255 bytes.
	// Currently, we require the generated group ID to fit within a single string.
	// If we need to expand it in the future, we can consider splitting across multiple strings.
	t.Run("group ID fits in DNS TXT string (255 bytes)", func(t *testing.T) {
		// Test with a long owner URL and type
		longOwner := "https://very-long-domain-name-example.com/path/to/resource/that/is/quite/long"
		longType := "verylongsymmetrytypename"
		hostnames := []string{
			"subdomain1.example.com",
			"subdomain2.example.com",
			"subdomain3.example.com",
		}

		groupID, err := service.CalculateV1(longOwner, longType, hostnames)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// len() in Go returns byte length, not logical Unicode character length, which is what we want
		if len(groupID) > 255 {
			t.Errorf("group ID byte length %d exceeds DNS TXT string size limit of 255 bytes: %s", len(groupID), groupID)
		}
	})
}
