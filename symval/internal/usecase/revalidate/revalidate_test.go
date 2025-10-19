package revalidate

import (
	"context"
	"testing"
	"time"

	"github.com/mrled/suns/symval/internal/groupid"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository"
	"github.com/mrled/suns/symval/internal/symgroup"
)

// setupTestRepo creates a memory repository with test data
func setupTestRepo(t *testing.T, records []*model.DomainRecord) repository.DomainRepository {
	repo := repository.NewMemoryRepository()
	ctx := context.Background()

	for _, record := range records {
		// Use a modified Store that allows updates for testing
		// Since Store returns ErrAlreadyExists, we'll directly manipulate the repo
		if err := repo.Store(ctx, record); err != nil {
			t.Fatalf("failed to setup test data: %v", err)
		}
	}

	return repo
}

// createValidPalindromeGroup creates a valid palindrome group for testing
func createValidPalindromeGroup(t *testing.T, owner string) []*model.DomainRecord {
	// A palindrome group has exactly one domain, and that domain must be a palindrome
	hostname := "noon"
	domains := []string{hostname}
	groupID, err := groupid.CalculateV1(owner, string(symgroup.Palindrome), domains)
	if err != nil {
		t.Fatalf("failed to calculate group ID: %v", err)
	}

	return []*model.DomainRecord{
		{
			Owner:        owner,
			Type:         symgroup.Palindrome,
			Hostname:     hostname,
			GroupID:      groupID,
			ValidateTime: time.Now(),
		},
	}
}

// createInvalidGroup creates an invalid group (mismatched GroupID)
func createInvalidGroup(t *testing.T, owner string, domains []string) []*model.DomainRecord {
	// Calculate a valid group ID but then use a different one
	validGroupID, err := groupid.CalculateV1(owner, string(symgroup.Palindrome), domains)
	if err != nil {
		t.Fatalf("failed to calculate group ID: %v", err)
	}

	// Create records with a modified (invalid) GroupID
	invalidGroupID := validGroupID + "invalid"
	records := make([]*model.DomainRecord, len(domains))
	for i, domain := range domains {
		records[i] = &model.DomainRecord{
			Owner:        owner,
			Type:         symgroup.Palindrome,
			Hostname:     domain,
			GroupID:      invalidGroupID,
			ValidateTime: time.Now(),
		}
	}

	return records
}

func TestFindInvalid_NoFilters(t *testing.T) {
	t.Run("empty repository", func(t *testing.T) {
		repo := repository.NewMemoryRepository()
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		invalid, err := uc.FindInvalid(ctx, FilterOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(invalid) != 0 {
			t.Errorf("expected 0 invalid records, got %d", len(invalid))
		}
	})

	t.Run("all valid records", func(t *testing.T) {
		validRecords := createValidPalindromeGroup(t, "owner1")
		repo := setupTestRepo(t, validRecords)
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		invalid, err := uc.FindInvalid(ctx, FilterOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(invalid) != 0 {
			t.Errorf("expected 0 invalid records, got %d", len(invalid))
		}
	})

	t.Run("all invalid records", func(t *testing.T) {
		invalidRecords := createInvalidGroup(t, "owner1", []string{"test.com"})
		repo := setupTestRepo(t, invalidRecords)
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		invalid, err := uc.FindInvalid(ctx, FilterOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(invalid) != 1 {
			t.Errorf("expected 1 invalid record, got %d", len(invalid))
		}
	})

	t.Run("mixed valid and invalid records", func(t *testing.T) {
		validRecords := createValidPalindromeGroup(t, "owner1")
		invalidRecords := createInvalidGroup(t, "owner2", []string{"bad.com"})
		allRecords := append(validRecords, invalidRecords...)

		repo := setupTestRepo(t, allRecords)
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		invalid, err := uc.FindInvalid(ctx, FilterOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(invalid) != 1 {
			t.Errorf("expected 1 invalid record, got %d", len(invalid))
		}
		if len(invalid) > 0 && invalid[0].Hostname != "bad.com" {
			t.Errorf("expected invalid record to be bad.com, got %s", invalid[0].Hostname)
		}
	})
}

func TestFindInvalid_OwnerFilter(t *testing.T) {
	t.Run("filter by owner with invalid records", func(t *testing.T) {
		owner1 := "owner1"
		owner2 := "owner2"

		validRecords := createValidPalindromeGroup(t, owner1)
		invalidRecords := createInvalidGroup(t, owner2, []string{"bad.com"})
		allRecords := append(validRecords, invalidRecords...)

		repo := setupTestRepo(t, allRecords)
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		// Filter for owner2 (has invalid records)
		invalid, err := uc.FindInvalid(ctx, FilterOptions{Owners: []string{owner2}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(invalid) != 1 {
			t.Errorf("expected 1 invalid record, got %d", len(invalid))
		}

		// Filter for owner1 (has valid records)
		invalid, err = uc.FindInvalid(ctx, FilterOptions{Owners: []string{owner1}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(invalid) != 0 {
			t.Errorf("expected 0 invalid records for owner1, got %d", len(invalid))
		}
	})

	t.Run("filter by multiple owners", func(t *testing.T) {
		owner1 := "owner1"
		owner2 := "owner2"
		owner3 := "owner3"

		valid1 := createValidPalindromeGroup(t, owner1)
		invalid1 := createInvalidGroup(t, owner2, []string{"bad1.com"})
		invalid2 := createInvalidGroup(t, owner3, []string{"bad2.com"})
		allRecords := append(valid1, append(invalid1, invalid2...)...)

		repo := setupTestRepo(t, allRecords)
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		// Filter for owner2 and owner3 (both have invalid records)
		invalid, err := uc.FindInvalid(ctx, FilterOptions{Owners: []string{owner2, owner3}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(invalid) != 2 {
			t.Errorf("expected 2 invalid records, got %d", len(invalid))
		}
	})
}

func TestFindInvalid_DomainFilter(t *testing.T) {
	t.Run("filter by domain expands to group", func(t *testing.T) {
		owner := "owner1"
		domains := []string{"test1.com", "test2.com"}

		// Create an invalid group with multiple domains
		invalidRecords := createInvalidGroup(t, owner, domains)

		repo := setupTestRepo(t, invalidRecords)
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		// Filter for just one domain - should expand to whole group
		invalid, err := uc.FindInvalid(ctx, FilterOptions{Domains: []string{"test1.com"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return both records from the group
		if len(invalid) != 2 {
			t.Errorf("expected 2 invalid records (whole group), got %d", len(invalid))
		}
	})

	t.Run("filter by multiple domains", func(t *testing.T) {
		owner := "owner1"
		group1Domains := []string{"test1.com", "test2.com"}
		group2Domains := []string{"test3.com", "test4.com"}

		// Create two invalid groups
		invalid1 := createInvalidGroup(t, owner, group1Domains)
		invalid2 := createInvalidGroup(t, owner, group2Domains)
		allRecords := append(invalid1, invalid2...)

		repo := setupTestRepo(t, allRecords)
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		// Filter for one domain from each group
		invalid, err := uc.FindInvalid(ctx, FilterOptions{Domains: []string{"test1.com", "test3.com"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return all 4 records (both complete groups)
		if len(invalid) != 4 {
			t.Errorf("expected 4 invalid records (both groups), got %d", len(invalid))
		}
	})

	t.Run("filter by domain not in repository", func(t *testing.T) {
		validRecords := createValidPalindromeGroup(t, "owner1")
		repo := setupTestRepo(t, validRecords)
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		invalid, err := uc.FindInvalid(ctx, FilterOptions{Domains: []string{"nonexistent.com"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(invalid) != 0 {
			t.Errorf("expected 0 invalid records, got %d", len(invalid))
		}
	})
}

func TestFindInvalid_GroupIDFilter(t *testing.T) {
	t.Run("filter by groupID", func(t *testing.T) {
		owner := "owner1"
		invalidRecords := createInvalidGroup(t, owner, []string{"bad.com"})
		validRecords := createValidPalindromeGroup(t, "owner2")
		allRecords := append(invalidRecords, validRecords...)

		repo := setupTestRepo(t, allRecords)
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		// Filter for the invalid group ID
		groupID := invalidRecords[0].GroupID
		invalid, err := uc.FindInvalid(ctx, FilterOptions{GroupIDs: []string{groupID}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(invalid) != 1 {
			t.Errorf("expected 1 invalid record, got %d", len(invalid))
		}

		// Filter for the valid group ID
		validGroupID := validRecords[0].GroupID
		invalid, err = uc.FindInvalid(ctx, FilterOptions{GroupIDs: []string{validGroupID}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(invalid) != 0 {
			t.Errorf("expected 0 invalid records for valid group, got %d", len(invalid))
		}
	})

	t.Run("filter by multiple groupIDs", func(t *testing.T) {
		owner := "owner1"
		invalid1 := createInvalidGroup(t, owner, []string{"bad1.com"})
		invalid2 := createInvalidGroup(t, owner, []string{"bad2.com"})
		validRecords := createValidPalindromeGroup(t, "owner2")
		allRecords := append(invalid1, append(invalid2, validRecords...)...)

		repo := setupTestRepo(t, allRecords)
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		// Filter for both invalid group IDs
		groupID1 := invalid1[0].GroupID
		groupID2 := invalid2[0].GroupID
		invalid, err := uc.FindInvalid(ctx, FilterOptions{GroupIDs: []string{groupID1, groupID2}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(invalid) != 2 {
			t.Errorf("expected 2 invalid records, got %d", len(invalid))
		}
	})
}

func TestFindInvalidAndDrop(t *testing.T) {
	t.Run("drop invalid records", func(t *testing.T) {
		owner := "owner1"
		invalidRecords := createInvalidGroup(t, owner, []string{"bad.com"})
		validRecords := createValidPalindromeGroup(t, "owner2")
		allRecords := append(invalidRecords, validRecords...)

		repo := setupTestRepo(t, allRecords)
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		// Verify initial count
		initialRecords, _ := repo.List(ctx)
		if len(initialRecords) != 2 {
			t.Fatalf("expected 2 initial records, got %d", len(initialRecords))
		}

		// Drop invalid records
		dropped, err := uc.FindInvalidAndDrop(ctx, FilterOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dropped) != 1 {
			t.Errorf("expected 1 dropped record, got %d", len(dropped))
		}

		// Verify remaining count
		remainingRecords, _ := repo.List(ctx)
		if len(remainingRecords) != 1 {
			t.Errorf("expected 1 remaining record, got %d", len(remainingRecords))
		}

		// Verify the remaining record is valid
		if remainingRecords[0].Hostname != "noon" {
			t.Errorf("expected remaining record to be noon, got %s", remainingRecords[0].Hostname)
		}
	})

	t.Run("drop with owner filter", func(t *testing.T) {
		owner1 := "owner1"
		owner2 := "owner2"

		invalid1 := createInvalidGroup(t, owner1, []string{"bad1.com"})
		invalid2 := createInvalidGroup(t, owner2, []string{"bad2.com"})
		allRecords := append(invalid1, invalid2...)

		repo := setupTestRepo(t, allRecords)
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		// Drop only owner1's invalid records
		dropped, err := uc.FindInvalidAndDrop(ctx, FilterOptions{Owners: []string{owner1}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dropped) != 1 {
			t.Errorf("expected 1 dropped record, got %d", len(dropped))
		}

		// Verify owner2's record remains
		remainingRecords, _ := repo.List(ctx)
		if len(remainingRecords) != 1 {
			t.Errorf("expected 1 remaining record, got %d", len(remainingRecords))
		}
		if remainingRecords[0].Owner != owner2 {
			t.Errorf("expected remaining record to belong to owner2, got %s", remainingRecords[0].Owner)
		}
	})

	t.Run("no invalid records to drop", func(t *testing.T) {
		validRecords := createValidPalindromeGroup(t, "owner1")
		repo := setupTestRepo(t, validRecords)
		uc := NewRevalidateUseCase(repo)
		ctx := context.Background()

		dropped, err := uc.FindInvalidAndDrop(ctx, FilterOptions{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(dropped) != 0 {
			t.Errorf("expected 0 dropped records, got %d", len(dropped))
		}

		// Verify record still exists
		remainingRecords, _ := repo.List(ctx)
		if len(remainingRecords) != 1 {
			t.Errorf("expected 1 remaining record, got %d", len(remainingRecords))
		}
	})
}

func TestFilterHelpers(t *testing.T) {
	t.Run("groupByGroupID", func(t *testing.T) {
		records := []*model.DomainRecord{
			{GroupID: "group1", Hostname: "a.com"},
			{GroupID: "group1", Hostname: "b.com"},
			{GroupID: "group2", Hostname: "c.com"},
		}

		grouped := groupByGroupID(records)
		if len(grouped) != 2 {
			t.Errorf("expected 2 groups, got %d", len(grouped))
		}
		if len(grouped["group1"]) != 2 {
			t.Errorf("expected 2 records in group1, got %d", len(grouped["group1"]))
		}
		if len(grouped["group2"]) != 1 {
			t.Errorf("expected 1 record in group2, got %d", len(grouped["group2"]))
		}
	})
}
