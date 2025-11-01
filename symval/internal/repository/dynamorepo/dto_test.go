package dynamorepo

import (
	"testing"
	"time"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

func TestFromDomain(t *testing.T) {
	testTime := time.Date(2025, 10, 17, 12, 0, 0, 0, time.UTC)

	record := &model.DomainRecord{
		Owner:        "alice@example.com",
		Type:         symgroup.Palindrome,
		Hostname:     "example.com",
		GroupID:      "abc123",
		ValidateTime: testTime,
	}

	dto := FromDomain(record)

	// Check that PK and SK are correctly mapped
	if dto.PK != "abc123" {
		t.Errorf("Expected PK to be 'abc123', got '%s'", dto.PK)
	}
	if dto.SK != "example.com" {
		t.Errorf("Expected SK to be 'example.com', got '%s'", dto.SK)
	}

	// Check that all other fields are preserved
	if dto.Owner != record.Owner {
		t.Errorf("Expected Owner to be '%s', got '%s'", record.Owner, dto.Owner)
	}
	if dto.Type != record.Type {
		t.Errorf("Expected Type to be '%s', got '%s'", record.Type, dto.Type)
	}
	if !dto.ValidateTime.Equal(record.ValidateTime) {
		t.Errorf("Expected ValidateTime to be '%s', got '%s'", record.ValidateTime, dto.ValidateTime)
	}
}

func TestToDomain(t *testing.T) {
	testTime := time.Date(2025, 10, 17, 12, 0, 0, 0, time.UTC)

	dto := &DynamoDTO{
		PK:           "group456",
		SK:           "test.org",
		Owner:        "bob@example.com",
		Type:         symgroup.Flip180,
		ValidateTime: testTime,
	}

	record := dto.ToDomain()

	// Check that all fields are correctly mapped to the domain model
	if record.Owner != dto.Owner {
		t.Errorf("Expected Owner to be '%s', got '%s'", dto.Owner, record.Owner)
	}
	if record.Type != dto.Type {
		t.Errorf("Expected Type to be '%s', got '%s'", dto.Type, record.Type)
	}
	// Check that Hostname is correctly reconstructed from SK
	if record.Hostname != dto.SK {
		t.Errorf("Expected Hostname to be '%s' (from SK), got '%s'", dto.SK, record.Hostname)
	}
	// Check that GroupID is correctly reconstructed from PK
	if record.GroupID != dto.PK {
		t.Errorf("Expected GroupID to be '%s' (from PK), got '%s'", dto.PK, record.GroupID)
	}
	if !record.ValidateTime.Equal(dto.ValidateTime) {
		t.Errorf("Expected ValidateTime to be '%s', got '%s'", dto.ValidateTime, record.ValidateTime)
	}
}

func TestRoundTripConversion(t *testing.T) {
	testTime := time.Date(2025, 10, 17, 12, 0, 0, 0, time.UTC)

	originalRecord := &model.DomainRecord{
		Owner:        "charlie@example.com",
		Type:         symgroup.MirrorText,
		Hostname:     "mirror.example.com",
		GroupID:      "mirror-group",
		ValidateTime: testTime,
	}

	// Convert to DTO and back
	dto := FromDomain(originalRecord)
	reconstructedRecord := dto.ToDomain()

	// Verify all fields match
	if reconstructedRecord.Owner != originalRecord.Owner {
		t.Errorf("Owner mismatch: expected '%s', got '%s'", originalRecord.Owner, reconstructedRecord.Owner)
	}
	if reconstructedRecord.Type != originalRecord.Type {
		t.Errorf("Type mismatch: expected '%s', got '%s'", originalRecord.Type, reconstructedRecord.Type)
	}
	if reconstructedRecord.Hostname != originalRecord.Hostname {
		t.Errorf("Hostname mismatch: expected '%s', got '%s'", originalRecord.Hostname, reconstructedRecord.Hostname)
	}
	if reconstructedRecord.GroupID != originalRecord.GroupID {
		t.Errorf("GroupID mismatch: expected '%s', got '%s'", originalRecord.GroupID, reconstructedRecord.GroupID)
	}
	if !reconstructedRecord.ValidateTime.Equal(originalRecord.ValidateTime) {
		t.Errorf("ValidateTime mismatch: expected '%s', got '%s'", originalRecord.ValidateTime, reconstructedRecord.ValidateTime)
	}
}

func TestFromDomainList(t *testing.T) {
	testTime := time.Date(2025, 10, 17, 12, 0, 0, 0, time.UTC)

	records := []*model.DomainRecord{
		{
			Owner:        "alice@example.com",
			Type:         symgroup.Palindrome,
			Hostname:     "example1.com",
			GroupID:      "group1",
			ValidateTime: testTime,
		},
		{
			Owner:        "bob@example.com",
			Type:         symgroup.Flip180,
			Hostname:     "example2.com",
			GroupID:      "group2",
			ValidateTime: testTime.Add(time.Hour),
		},
		{
			Owner:        "charlie@example.com",
			Type:         symgroup.DoubleFlip180,
			Hostname:     "example3.com",
			GroupID:      "group3",
			ValidateTime: testTime.Add(2 * time.Hour),
		},
	}

	dtos := FromDomainList(records)

	if len(dtos) != len(records) {
		t.Fatalf("Expected %d DTOs, got %d", len(records), len(dtos))
	}

	for i, dto := range dtos {
		record := records[i]

		// Check PK and SK mapping
		if dto.PK != record.GroupID {
			t.Errorf("Record %d: Expected PK to be '%s', got '%s'", i, record.GroupID, dto.PK)
		}
		if dto.SK != record.Hostname {
			t.Errorf("Record %d: Expected SK to be '%s', got '%s'", i, record.Hostname, dto.SK)
		}

		// Check other fields
		if dto.Owner != record.Owner {
			t.Errorf("Record %d: Owner mismatch", i)
		}
		if dto.Type != record.Type {
			t.Errorf("Record %d: Type mismatch", i)
		}
		if !dto.ValidateTime.Equal(record.ValidateTime) {
			t.Errorf("Record %d: ValidateTime mismatch", i)
		}
	}
}

func TestToDomainList(t *testing.T) {
	testTime := time.Date(2025, 10, 17, 12, 0, 0, 0, time.UTC)

	dtos := []*DynamoDTO{
		{
			PK:           "group1",
			SK:           "example1.com",
			Owner:        "alice@example.com",
			Type:         symgroup.Palindrome,
			ValidateTime: testTime,
		},
		{
			PK:           "group2",
			SK:           "example2.com",
			Owner:        "bob@example.com",
			Type:         symgroup.Flip180,
			ValidateTime: testTime.Add(time.Hour),
		},
		{
			PK:           "group3",
			SK:           "example3.com",
			Owner:        "charlie@example.com",
			Type:         symgroup.MirrorNames,
			ValidateTime: testTime.Add(2 * time.Hour),
		},
	}

	records := ToDomainList(dtos)

	if len(records) != len(dtos) {
		t.Fatalf("Expected %d records, got %d", len(dtos), len(records))
	}

	for i, record := range records {
		dto := dtos[i]

		// Check all fields are correctly mapped
		if record.Owner != dto.Owner {
			t.Errorf("Record %d: Owner mismatch", i)
		}
		if record.Type != dto.Type {
			t.Errorf("Record %d: Type mismatch", i)
		}
		// Verify Hostname is reconstructed from SK
		if record.Hostname != dto.SK {
			t.Errorf("Record %d: Expected Hostname '%s' (from SK), got '%s'", i, dto.SK, record.Hostname)
		}
		// Verify GroupID is reconstructed from PK
		if record.GroupID != dto.PK {
			t.Errorf("Record %d: Expected GroupID '%s' (from PK), got '%s'", i, dto.PK, record.GroupID)
		}
		if !record.ValidateTime.Equal(dto.ValidateTime) {
			t.Errorf("Record %d: ValidateTime mismatch", i)
		}
	}
}

func TestEmptyListConversions(t *testing.T) {
	// Test empty list conversions
	emptyRecords := []*model.DomainRecord{}
	emptyDTOs := []*DynamoDTO{}

	convertedDTOs := FromDomainList(emptyRecords)
	if len(convertedDTOs) != 0 {
		t.Errorf("Expected empty DTO list, got %d items", len(convertedDTOs))
	}

	convertedRecords := ToDomainList(emptyDTOs)
	if len(convertedRecords) != 0 {
		t.Errorf("Expected empty record list, got %d items", len(convertedRecords))
	}
}