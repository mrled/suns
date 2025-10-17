package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mrled/suns/symval/internal/model"
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

	testData := &model.DomainData{
		Owner:        "test-owner",
		Type:         model.Palindrome,
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
	retrieved, err := repo2.Get(ctx, "example.com")
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

	testData := &model.DomainData{
		Owner:        "test-owner",
		Type:         model.MirrorText,
		Hostname:     "example.com",
		GroupID:      "group-456",
		ValidateTime: time.Now(),
	}

	if err := repo.Store(ctx, testData); err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	if err := repo.Delete(ctx, "example.com"); err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	// Create new repository to verify deletion was persisted
	repo2, err := NewMemoryRepositoryWithPersistence(tmpPath)
	if err != nil {
		t.Fatalf("Failed to create second repository: %v", err)
	}

	_, err = repo2.Get(ctx, "example.com")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestMemoryRepository_NonPersistent(t *testing.T) {
	ctx := context.Background()

	// Create repository in memory-only mode
	repo := NewMemoryRepository()

	testData := &model.DomainData{
		Owner:        "test-owner",
		Type:         model.Flip180,
		Hostname:     "example.com",
		GroupID:      "group-789",
		ValidateTime: time.Now(),
	}

	if err := repo.Store(ctx, testData); err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	// Verify data is in memory
	retrieved, err := repo.Get(ctx, "example.com")
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
	if err := repo.Delete(ctx, "example.com"); err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	_, err = repo.Get(ctx, "example.com")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound after delete, got %v", err)
	}
}
