package groupid

import (
	"testing"
)

func TestCalculateV1(t *testing.T) {
	service := NewService()

	t.Run("basic calculation", func(t *testing.T) {
		groupID, err := service.CalculateV1("owner1", "type1", []string{"host1.example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check format: v1:type:base64hash
		if len(groupID) == 0 {
			t.Fatal("groupID should not be empty")
		}

		// Verify it starts with v1:type1:
		expected := "v1:type1:"
		if len(groupID) < len(expected) || groupID[:len(expected)] != expected {
			t.Errorf("groupID should start with %q, got %q", expected, groupID)
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
			want:      "v1:mytype:gFEOj0UEJk8EI+oV0HCaKMfpeCwL6EFDbqAUQPFchSU=",
		},
		{
			name:      "multiple hostnames",
			owner:     "myowner",
			gtype:     "mytype",
			hostnames: []string{"host1.example.com", "host2.example.com", "host3.example.com"},
			want:      "v1:mytype:j1QFHMJKG8Gj/VbzAC5NXW30J2esEIHeWVelAskdKro=",
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
