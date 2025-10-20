package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
)

func TestMemoryRepository_JSONPersistence(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-domains-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	ctx := context.Background()

	// Create first repository and add data
	repo1, err := NewMemoryRepositoryWithPersistence(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	testData := &model.DomainRecord{
		Owner:        "test-owner",
		Type:         symgroup.Palindrome,
		Hostname:     "example.com",
		GroupID:      "group-123",
		ValidateTime: time.Now(),
	}

	if err := repo1.Store(ctx, testData); err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Create second repository with same file path
	repo2, err := NewMemoryRepositoryWithPersistence(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create second repository: %v", err)
	}

	// Verify data was loaded from file
	retrieved, err := repo2.Get(ctx, "group-123", "example.com")
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if retrieved.Hostname != testData.Hostname {
		t.Errorf("Expected hostname %s, got %s", testData.Hostname, retrieved.Hostname)
	}
	if retrieved.Owner != testData.Owner {
		t.Errorf("Expected owner %s, got %s", testData.Owner, retrieved.Owner)
	}
	if retrieved.Type != testData.Type {
		t.Errorf("Expected type %s, got %s", testData.Type, retrieved.Type)
	}
	if retrieved.GroupID != testData.GroupID {
		t.Errorf("Expected groupID %s, got %s", testData.GroupID, retrieved.GroupID)
	}
}

func TestMemoryRepository_DeletePersistence(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-domains-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	ctx := context.Background()

	repo, err := NewMemoryRepositoryWithPersistence(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	testData := &model.DomainRecord{
		Owner:        "test-owner",
		Type:         symgroup.MirrorText,
		Hostname:     "example.com",
		GroupID:      "group-456",
		ValidateTime: time.Now(),
	}

	if err := repo.Store(ctx, testData); err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	if err := repo.Delete(ctx, "group-456", "example.com"); err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	// Create new repository to verify deletion was persisted
	repo2, err := NewMemoryRepositoryWithPersistence(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create second repository: %v", err)
	}

	_, err = repo2.Get(ctx, "group-456", "example.com")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestMemoryRepository_NonPersistent(t *testing.T) {
	ctx := context.Background()

	// Create repository in memory-only mode
	repo := NewMemoryRepository()

	testData := &model.DomainRecord{
		Owner:        "test-owner",
		Type:         symgroup.Flip180,
		Hostname:     "example.com",
		GroupID:      "group-789",
		ValidateTime: time.Now(),
	}

	if err := repo.Store(ctx, testData); err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Verify data is in memory
	retrieved, err := repo.Get(ctx, "group-789", "example.com")
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if retrieved.Hostname != testData.Hostname {
		t.Errorf("Expected hostname %s, got %s", testData.Hostname, retrieved.Hostname)
	}

	// Verify List works
	allData, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list data: %v", err)
	}

	if len(allData) != 1 {
		t.Errorf("Expected 1 item, got %d", len(allData))
	}

	// Verify Delete works
	if err := repo.Delete(ctx, "group-789", "example.com"); err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	_, err = repo.Get(ctx, "group-789", "example.com")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound after delete, got %v", err)
	}
}

func TestMemoryRepository_FromJsonString(t *testing.T) {
	ctx := context.Background()

	// Create test data as JSON string
	// Note: Field names match the Go struct field names (capitalized) as Go's default JSON marshaling is used
	// Type values are single characters as defined in symgroup package
	jsonString := `[
		{
			"Owner": "alice",
			"Type": "a",
			"Hostname": "example.com",
			"GroupID": "group-123",
			"ValidateTime": "2025-01-15T10:00:00Z"
		},
		{
			"Owner": "bob",
			"Type": "d",
			"Hostname": "test.org",
			"GroupID": "group-456",
			"ValidateTime": "2025-01-16T12:00:00Z"
		}
	]`

	// Create repository from JSON string
	repo, err := NewMemoryRepositoryFromJsonString(jsonString)
	if err != nil {
		t.Fatalf("Failed to create repository from JSON string: %v", err)
	}

	// Verify first record
	record1, err := repo.Get(ctx, "group-123", "example.com")
	if err != nil {
		t.Fatalf("Failed to get example.com: %v", err)
	}
	if record1.Owner != "alice" {
		t.Errorf("Expected owner alice, got %s", record1.Owner)
	}
	if record1.Type != symgroup.Palindrome {
		t.Errorf("Expected type palindrome, got %s", record1.Type)
	}
	if record1.GroupID != "group-123" {
		t.Errorf("Expected groupID group-123, got %s", record1.GroupID)
	}

	// Verify second record
	record2, err := repo.Get(ctx, "group-456", "test.org")
	if err != nil {
		t.Fatalf("Failed to get test.org: %v", err)
	}
	if record2.Owner != "bob" {
		t.Errorf("Expected owner bob, got %s", record2.Owner)
	}
	if record2.Type != symgroup.MirrorText {
		t.Errorf("Expected type mirror_text, got %s", record2.Type)
	}

	// Verify List contains both records
	allData, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list data: %v", err)
	}
	if len(allData) != 2 {
		t.Errorf("Expected 2 items, got %d", len(allData))
	}

	// Verify repository is not persistent (no file path)
	testData := &model.DomainRecord{
		Owner:        "charlie",
		Type:         symgroup.Flip180,
		Hostname:     "new.com",
		GroupID:      "group-789",
		ValidateTime: time.Now(),
	}
	if err := repo.Store(ctx, testData); err != nil {
		t.Fatalf("Failed to store new data: %v", err)
	}

	// Verify new data is stored in memory
	retrieved, err := repo.Get(ctx, "group-789", "new.com")
	if err != nil {
		t.Fatalf("Failed to get new.com: %v", err)
	}
	if retrieved.Owner != "charlie" {
		t.Errorf("Expected owner charlie, got %s", retrieved.Owner)
	}
}

func TestMemoryRepository_FromJsonString_Empty(t *testing.T) {
	// Test with empty array
	repo, err := NewMemoryRepositoryFromJsonString("[]")
	if err != nil {
		t.Fatalf("Failed to create repository from empty JSON: %v", err)
	}

	ctx := context.Background()
	allData, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list data: %v", err)
	}
	if len(allData) != 0 {
		t.Errorf("Expected 0 items, got %d", len(allData))
	}
}

func TestMemoryRepository_FromJsonString_Invalid(t *testing.T) {
	// Test with invalid JSON
	_, err := NewMemoryRepositoryFromJsonString("invalid json")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
