package model

import (
	"testing"
	"time"

	"github.com/mrled/suns/symval/internal/symgroup"
)

func TestSortRecords_ByOwner(t *testing.T) {
	now := time.Now()
	records := []*DomainRecord{
		{Owner: "charlie@example.com", Hostname: "example.com", ValidateTime: now},
		{Owner: "alice@example.com", Hostname: "test.com", ValidateTime: now},
		{Owner: "bob@example.com", Hostname: "demo.com", ValidateTime: now},
	}

	SortRecords(records, "owner")

	if records[0].Owner != "alice@example.com" {
		t.Errorf("Expected alice first, got %s", records[0].Owner)
	}
	if records[1].Owner != "bob@example.com" {
		t.Errorf("Expected bob second, got %s", records[1].Owner)
	}
	if records[2].Owner != "charlie@example.com" {
		t.Errorf("Expected charlie third, got %s", records[2].Owner)
	}
}

func TestSortRecords_ByDomain(t *testing.T) {
	now := time.Now()
	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "zebra.com", ValidateTime: now},
		{Owner: "alice@example.com", Hostname: "apple.com", ValidateTime: now},
		{Owner: "alice@example.com", Hostname: "middle.com", ValidateTime: now},
	}

	SortRecords(records, "domain")

	if records[0].Hostname != "apple.com" {
		t.Errorf("Expected apple.com first, got %s", records[0].Hostname)
	}
	if records[1].Hostname != "middle.com" {
		t.Errorf("Expected middle.com second, got %s", records[1].Hostname)
	}
	if records[2].Hostname != "zebra.com" {
		t.Errorf("Expected zebra.com third, got %s", records[2].Hostname)
	}
}

func TestSortRecords_ByGroupID(t *testing.T) {
	now := time.Now()
	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "example.com", GroupID: "g3", ValidateTime: now},
		{Owner: "alice@example.com", Hostname: "test.com", GroupID: "g1", ValidateTime: now},
		{Owner: "alice@example.com", Hostname: "demo.com", GroupID: "g2", ValidateTime: now},
	}

	SortRecords(records, "group")

	if records[0].GroupID != "g1" {
		t.Errorf("Expected g1 first, got %s", records[0].GroupID)
	}
	if records[1].GroupID != "g2" {
		t.Errorf("Expected g2 second, got %s", records[1].GroupID)
	}
	if records[2].GroupID != "g3" {
		t.Errorf("Expected g3 third, got %s", records[2].GroupID)
	}
}

func TestSortRecords_ByValidateTime(t *testing.T) {
	now := time.Now()
	older := now.Add(-1 * time.Hour)
	oldest := now.Add(-2 * time.Hour)

	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "example.com", ValidateTime: oldest},
		{Owner: "alice@example.com", Hostname: "test.com", ValidateTime: now},
		{Owner: "alice@example.com", Hostname: "demo.com", ValidateTime: older},
	}

	SortRecords(records, "validate-time")

	// Should be sorted newest first (After comparison)
	if !records[0].ValidateTime.Equal(now) {
		t.Errorf("Expected newest first, got %v", records[0].ValidateTime)
	}
	if !records[1].ValidateTime.Equal(older) {
		t.Errorf("Expected older second, got %v", records[1].ValidateTime)
	}
	if !records[2].ValidateTime.Equal(oldest) {
		t.Errorf("Expected oldest third, got %v", records[2].ValidateTime)
	}
}

func TestSortRecords_ByType(t *testing.T) {
	now := time.Now()
	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "example.com", Type: symgroup.MirrorText, ValidateTime: now},
		{Owner: "alice@example.com", Hostname: "test.com", Type: symgroup.Palindrome, ValidateTime: now},
		{Owner: "alice@example.com", Hostname: "demo.com", Type: symgroup.Flip180, ValidateTime: now},
	}

	SortRecords(records, "type")

	// Types are single character codes, so should sort: a (palindrome), b (flip180), d (mirrortext)
	if records[0].Type != symgroup.Palindrome {
		t.Errorf("Expected palindrome (a) first, got %s", records[0].Type)
	}
	if records[1].Type != symgroup.Flip180 {
		t.Errorf("Expected flip180 (b) second, got %s", records[1].Type)
	}
	if records[2].Type != symgroup.MirrorText {
		t.Errorf("Expected mirrortext (d) third, got %s", records[2].Type)
	}
}

func TestSortRecords_Default(t *testing.T) {
	now := time.Now()
	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "zebra.com", GroupID: "g2", ValidateTime: now},
		{Owner: "alice@example.com", Hostname: "apple.com", GroupID: "g2", ValidateTime: now},
		{Owner: "alice@example.com", Hostname: "example.com", GroupID: "g1", ValidateTime: now},
	}

	SortRecords(records, "") // Empty string should use default sort

	// Default sort: by GroupID first, then by Hostname
	if records[0].GroupID != "g1" {
		t.Errorf("Expected g1 first, got %s", records[0].GroupID)
	}
	if records[1].GroupID != "g2" || records[1].Hostname != "apple.com" {
		t.Errorf("Expected g2/apple.com second, got %s/%s", records[1].GroupID, records[1].Hostname)
	}
	if records[2].GroupID != "g2" || records[2].Hostname != "zebra.com" {
		t.Errorf("Expected g2/zebra.com third, got %s/%s", records[2].GroupID, records[2].Hostname)
	}
}

func TestSortRecords_UnrecognizedFallsBackToDefault(t *testing.T) {
	now := time.Now()
	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "zebra.com", GroupID: "g2", ValidateTime: now},
		{Owner: "alice@example.com", Hostname: "apple.com", GroupID: "g1", ValidateTime: now},
	}

	SortRecords(records, "invalid-sort-field")

	// Should fall back to default sort (by GroupID, then Hostname)
	if records[0].GroupID != "g1" {
		t.Errorf("Expected default sort behavior, got %s first", records[0].GroupID)
	}
}

func TestSortRecords_EmptySlice(t *testing.T) {
	records := []*DomainRecord{}

	// Should not panic
	SortRecords(records, "owner")

	if len(records) != 0 {
		t.Errorf("Expected empty slice to remain empty")
	}
}

func TestSortRecords_SingleRecord(t *testing.T) {
	now := time.Now()
	records := []*DomainRecord{
		{Owner: "alice@example.com", Hostname: "example.com", ValidateTime: now},
	}

	SortRecords(records, "owner")

	if len(records) != 1 {
		t.Errorf("Expected single record to remain")
	}
}
