package model

import (
	"testing"

	"github.com/mrled/suns/symval/internal/symgroup"
)

func TestFilterRecords_EmptyFilter(t *testing.T) {
	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "example.com", GroupID: "g1"},
		{Owner: "bob@example.com", Hostname: "test.com", GroupID: "g2"},
	}

	filter := RecordFilter{}
	result := FilterRecords(records, filter)

	if len(result) != 2 {
		t.Errorf("Expected 2 records with empty filter, got %d", len(result))
	}
}

func TestFilterRecords_SingleOwner(t *testing.T) {
	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "example.com", GroupID: "g1"},
		{Owner: "bob@example.com", Hostname: "test.com", GroupID: "g2"},
		{Owner: "alice@example.com", Hostname: "demo.com", GroupID: "g3"},
	}

	filter := RecordFilter{Owners: []string{"alice@example.com"}}
	result := FilterRecords(records, filter)

	if len(result) != 2 {
		t.Errorf("Expected 2 records for alice, got %d", len(result))
	}

	for _, record := range result {
		if record.Owner != "alice@example.com" {
			t.Errorf("Expected only alice records, got %s", record.Owner)
		}
	}
}

func TestFilterRecords_CaseInsensitiveOwner(t *testing.T) {
	records := []*DomainRecord{
		{Owner: "Alice@Example.com", Hostname: "example.com", GroupID: "g1"},
		{Owner: "bob@example.com", Hostname: "test.com", GroupID: "g2"},
	}

	filter := RecordFilter{Owners: []string{"alice@example.com"}}
	result := FilterRecords(records, filter)

	if len(result) != 1 {
		t.Errorf("Expected case-insensitive match, got %d records", len(result))
	}
}

func TestFilterRecords_MultipleOwners(t *testing.T) {
	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "example.com", GroupID: "g1"},
		{Owner: "bob@example.com", Hostname: "test.com", GroupID: "g2"},
		{Owner: "charlie@example.com", Hostname: "demo.com", GroupID: "g3"},
	}

	filter := RecordFilter{Owners: []string{"alice@example.com", "bob@example.com"}}
	result := FilterRecords(records, filter)

	if len(result) != 2 {
		t.Errorf("Expected 2 records (alice or bob), got %d", len(result))
	}
}

func TestFilterRecords_CombinedFilters(t *testing.T) {
	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "example.com", GroupID: "g1", Type: "a"},
		{Owner: "alice@example.com", Hostname: "test.com", GroupID: "g2", Type: "b"},
		{Owner: "bob@example.com", Hostname: "example.com", GroupID: "g3", Type: "a"},
	}

	// Filter for alice AND example.com
	filter := RecordFilter{
		Owners:  []string{"alice@example.com"},
		Domains: []string{"example.com"},
	}
	result := FilterRecords(records, filter)

	if len(result) != 1 {
		t.Errorf("Expected 1 record matching both filters, got %d", len(result))
	}

	if result[0].Owner != "alice@example.com" || result[0].Hostname != "example.com" {
		t.Errorf("Got wrong record: %v", result[0])
	}
}

func TestFilterRecords_GroupIDExactMatch(t *testing.T) {
	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "example.com", GroupID: "v1:a:hash1:hash2"},
		{Owner: "bob@example.com", Hostname: "test.com", GroupID: "v1:b:hash3:hash4"},
	}

	filter := RecordFilter{GroupIDs: []string{"v1:a:hash1:hash2"}}
	result := FilterRecords(records, filter)

	if len(result) != 1 {
		t.Errorf("Expected 1 record with exact group ID match, got %d", len(result))
	}

	if result[0].GroupID != "v1:a:hash1:hash2" {
		t.Errorf("Got wrong group ID: %s", result[0].GroupID)
	}
}

func TestFilterRecords_NoMatches(t *testing.T) {
	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "example.com", GroupID: "g1"},
		{Owner: "bob@example.com", Hostname: "test.com", GroupID: "g2"},
	}

	filter := RecordFilter{Owners: []string{"charlie@example.com"}}
	result := FilterRecords(records, filter)

	if len(result) != 0 {
		t.Errorf("Expected 0 records with no matches, got %d", len(result))
	}
}

func TestFilterRecords_TypeFilter(t *testing.T) {
	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "example.com", Type: symgroup.Palindrome},
		{Owner: "bob@example.com", Hostname: "test.com", Type: symgroup.Flip180},
		{Owner: "charlie@example.com", Hostname: "demo.com", Type: symgroup.Palindrome},
	}

	filter := RecordFilter{Types: []string{string(symgroup.Palindrome)}}
	result := FilterRecords(records, filter)

	if len(result) != 2 {
		t.Errorf("Expected 2 palindrome records, got %d", len(result))
	}

	for _, record := range result {
		if record.Type != symgroup.Palindrome {
			t.Errorf("Expected only palindrome type, got %s", record.Type)
		}
	}
}
