package attestation

import (
	"testing"
	"time"

	"github.com/mrled/suns/symval/internal/groupid"
	"github.com/mrled/suns/symval/internal/symgroup"
)

func TestFilterDomainRecord(t *testing.T) {
	hostname := "example.com"
	owner := "testowner"
	typeA := symgroup.Palindrome
	validateTime := time.Now()

	// Generate valid group IDs for testing
	groupID1, err := groupid.CalculateV1(owner, string(typeA), []string{hostname})
	if err != nil {
		t.Fatalf("failed to calculate group ID: %v", err)
	}

	groupID2, err := groupid.CalculateV1("differentowner", string(typeA), []string{hostname})
	if err != nil {
		t.Fatalf("failed to calculate group ID: %v", err)
	}

	groupID3, err := groupid.CalculateV1(owner, string(symgroup.Flip180), []string{hostname})
	if err != nil {
		t.Fatalf("failed to calculate group ID: %v", err)
	}

	tests := []struct {
		name        string
		hostname    string
		records     []string
		criteria    FilterCriteria
		wantCount   int
		wantOwner   string
		wantType    symgroup.SymmetryType
		wantGroupID string
	}{
		{
			name:        "no filters - all valid records returned",
			hostname:    hostname,
			records:     []string{groupID1, groupID2, groupID3},
			criteria:    FilterCriteria{},
			wantCount:   3,
			wantOwner:   "",
			wantType:    typeA,
			wantGroupID: groupID1,
		},
		{
			name:        "filter by owner - matching record",
			hostname:    hostname,
			records:     []string{groupID1, groupID2, groupID3},
			criteria:    FilterCriteria{Owner: &owner},
			wantCount:   2, // groupID1 and groupID3 match owner
			wantOwner:   owner,
			wantType:    typeA,
			wantGroupID: groupID1,
		},
		{
			name:        "filter by type - matching record",
			hostname:    hostname,
			records:     []string{groupID1, groupID2, groupID3},
			criteria:    FilterCriteria{Type: &typeA},
			wantCount:   2, // groupID1 and groupID2 match type
			wantOwner:   "",
			wantType:    typeA,
			wantGroupID: groupID1,
		},
		{
			name:        "filter by groupID - exact match",
			hostname:    hostname,
			records:     []string{groupID1, groupID2, groupID3},
			criteria:    FilterCriteria{GroupID: &groupID1},
			wantCount:   1,
			wantOwner:   "",
			wantType:    typeA,
			wantGroupID: groupID1,
		},
		{
			name:        "filter by owner and type - matching record",
			hostname:    hostname,
			records:     []string{groupID1, groupID2, groupID3},
			criteria:    FilterCriteria{Owner: &owner, Type: &typeA},
			wantCount:   1, // only groupID1 matches both
			wantOwner:   owner,
			wantType:    typeA,
			wantGroupID: groupID1,
		},
		{
			name:      "filter by owner - no matching records",
			hostname:  hostname,
			records:   []string{groupID2}, // only differentowner
			criteria:  FilterCriteria{Owner: &owner},
			wantCount: 0,
		},
		{
			name:      "empty records list",
			hostname:  hostname,
			records:   []string{},
			criteria:  FilterCriteria{Owner: &owner},
			wantCount: 0,
		},
		{
			name:        "invalid records are skipped",
			hostname:    hostname,
			records:     []string{"invalid:record", groupID1, "also:invalid:format"},
			criteria:    FilterCriteria{Owner: &owner},
			wantCount:   1,
			wantOwner:   owner,
			wantType:    typeA,
			wantGroupID: groupID1,
		},
		{
			name:      "all invalid records",
			hostname:  hostname,
			records:   []string{"invalid:record", "also:invalid:format"},
			criteria:  FilterCriteria{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filterDomainRecords(tt.hostname, tt.records, tt.criteria, validateTime)
			if err != nil {
				t.Fatalf("filterDomainRecords returned unexpected error: %v", err)
			}

			if len(result) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(result), tt.wantCount)
			}

			if tt.wantCount > 0 {
				first := result[0]

				if first.Hostname != tt.hostname {
					t.Errorf("got hostname %q, want %q", first.Hostname, tt.hostname)
				}

				if tt.wantOwner != "" && first.Owner != tt.wantOwner {
					t.Errorf("got owner %q, want %q", first.Owner, tt.wantOwner)
				}

				if first.Type != tt.wantType {
					t.Errorf("got type %q, want %q", first.Type, tt.wantType)
				}

				if tt.wantGroupID != "" && first.GroupID != tt.wantGroupID {
					t.Errorf("got groupID %q, want %q", first.GroupID, tt.wantGroupID)
				}

				if !first.ValidateTime.Equal(validateTime) {
					t.Errorf("got validateTime %v, want %v", first.ValidateTime, validateTime)
				}
			}
		})
	}
}

func TestFilterDomainRecords_TypeInference(t *testing.T) {
	// Test that when Type filter is not specified, the type is inferred from the record
	hostname := "example.com"
	owner := "testowner"
	validateTime := time.Now()

	groupIDFlip180, err := groupid.CalculateV1(owner, string(symgroup.Flip180), []string{hostname})
	if err != nil {
		t.Fatalf("failed to calculate group ID: %v", err)
	}

	records := []string{groupIDFlip180}
	criteria := FilterCriteria{} // No type specified

	result, err := filterDomainRecords(hostname, records, criteria, validateTime)
	if err != nil {
		t.Fatalf("filterDomainRecords returned unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("got %d results, want 1", len(result))
	}

	if result[0].Type != symgroup.Flip180 {
		t.Errorf("got type %q, want %q (inferred from record)", result[0].Type, symgroup.Flip180)
	}
}

func TestFilterDomainRecords_MultipleMatchingRecords(t *testing.T) {
	// Test that multiple records matching the criteria are all returned
	hostname := "example.com"
	owner := "testowner"
	typeA := symgroup.Palindrome
	validateTime := time.Now()

	// Create multiple group IDs with same owner and type but different domains
	groupID1, err := groupid.CalculateV1(owner, string(typeA), []string{hostname})
	if err != nil {
		t.Fatalf("failed to calculate group ID: %v", err)
	}

	groupID2, err := groupid.CalculateV1(owner, string(typeA), []string{hostname, "other.com"})
	if err != nil {
		t.Fatalf("failed to calculate group ID: %v", err)
	}

	records := []string{groupID1, groupID2}
	criteria := FilterCriteria{Owner: &owner, Type: &typeA}

	result, err := filterDomainRecords(hostname, records, criteria, validateTime)
	if err != nil {
		t.Fatalf("filterDomainRecords returned unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("got %d results, want 2 (both records should match)", len(result))
	}

	// Verify both results have correct properties
	for i, res := range result {
		if res.Owner != owner {
			t.Errorf("result[%d]: got owner %q, want %q", i, res.Owner, owner)
		}
		if res.Type != typeA {
			t.Errorf("result[%d]: got type %q, want %q", i, res.Type, typeA)
		}
		if res.Hostname != hostname {
			t.Errorf("result[%d]: got hostname %q, want %q", i, res.Hostname, hostname)
		}
	}
}

func TestFilterDomainRecords_OwnerHashComparison(t *testing.T) {
	// Test that owner filtering works correctly by comparing owner hashes
	hostname := "example.com"
	owner1 := "owner1"
	owner2 := "owner2"
	typeA := symgroup.Palindrome
	validateTime := time.Now()

	groupID1, err := groupid.CalculateV1(owner1, string(typeA), []string{hostname})
	if err != nil {
		t.Fatalf("failed to calculate group ID: %v", err)
	}

	groupID2, err := groupid.CalculateV1(owner2, string(typeA), []string{hostname})
	if err != nil {
		t.Fatalf("failed to calculate group ID: %v", err)
	}

	records := []string{groupID1, groupID2}

	// Filter for owner1
	criteria := FilterCriteria{Owner: &owner1}
	result, err := filterDomainRecords(hostname, records, criteria, validateTime)
	if err != nil {
		t.Fatalf("filterDomainRecords returned unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("got %d results, want 1", len(result))
	}

	if result[0].GroupID != groupID1 {
		t.Errorf("got groupID %q, want %q (should match owner1)", result[0].GroupID, groupID1)
	}

	// Filter for owner2
	criteria2 := FilterCriteria{Owner: &owner2}
	result2, err := filterDomainRecords(hostname, records, criteria2, validateTime)
	if err != nil {
		t.Fatalf("filterDomainRecords returned unexpected error: %v", err)
	}

	if len(result2) != 1 {
		t.Fatalf("got %d results, want 1", len(result2))
	}

	if result2[0].GroupID != groupID2 {
		t.Errorf("got groupID %q, want %q (should match owner2)", result2[0].GroupID, groupID2)
	}
}
