package commands

import (
	"context"
	"fmt"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository"
	"github.com/mrled/suns/symval/internal/usecase/revalidate"
	"github.com/spf13/cobra"
)

var (
	revalidateFilePath   string
	revalidateDynamoName string
	revalidateOwners     []string
	revalidateDomains    []string
	revalidateGroupIDs   []string
	revalidateDryRun     bool
)

var revalidateCmd = &cobra.Command{
	Use:   "revalidate",
	Short: "Revalidate domain records in the data store",
	Long: `Revalidate checks all records in the data store for consistency.

It does not query DNS - it only validates existing records. For each record,
it ensures the record is part of a valid group using the validation rules.

You can filter which records to check using the following flags:
  --owner, -o    : Filter by owner(s)
  --domain, -n   : Filter by domain name(s)
  --groupid, -g  : Filter by group ID(s)

When filtering by domain, the check expands to include all records in any
group that the specified domain(s) belong to.

By default, invalid records are dropped from the data store. Use --dry-run
to see what would be removed without actually removing anything.

Examples:
  # Check all records and drop invalid ones
  symval revalidate --file ./data.json

  # Dry run to see what would be removed
  symval revalidate --file ./data.json --dry-run

  # Check records for specific owner
  symval revalidate --file ./data.json --owner alice@example.com

  # Check multiple owners
  symval revalidate --file ./data.json -o alice@example.com -o bob@example.com

  # Check specific domains (expands to full groups)
  symval revalidate --file ./data.json --domain test.com --domain example.org

  # Check specific group IDs
  symval revalidate --file ./data.json -g "v1:a:hash1:hash2"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create repository based on persistence flags
		var repo repository.DomainRepository
		if revalidateDynamoName != "" {
			return fmt.Errorf("--dynamo flag is not yet implemented")
		} else if revalidateFilePath != "" {
			// Use JSON file persistence
			memRepo, err := repository.NewMemoryRepositoryWithPersistence(revalidateFilePath)
			if err != nil {
				return fmt.Errorf("failed to create repository: %w", err)
			}
			repo = memRepo
			fmt.Printf("Using JSON persistence: %s\n", revalidateFilePath)
		} else {
			// Use in-memory only (no persistence)
			repo = repository.NewMemoryRepository()
			fmt.Println("Using in-memory storage (no persistence)")
		}

		// Create revalidate use case
		revalidateUC := revalidate.NewRevalidateUseCase(repo)
		ctx := context.Background()

		// Build filter options
		filters := revalidate.FilterOptions{
			Owners:   revalidateOwners,
			Domains:  revalidateDomains,
			GroupIDs: revalidateGroupIDs,
		}

		// Print filter information
		if len(revalidateOwners) > 0 {
			fmt.Printf("Filtering by owner(s): %v\n", revalidateOwners)
		}
		if len(revalidateDomains) > 0 {
			fmt.Printf("Filtering by domain(s): %v\n", revalidateDomains)
		}
		if len(revalidateGroupIDs) > 0 {
			fmt.Printf("Filtering by group ID(s): %v\n", revalidateGroupIDs)
		}
		if len(revalidateOwners) == 0 && len(revalidateDomains) == 0 && len(revalidateGroupIDs) == 0 {
			fmt.Println("No filters specified - checking all records")
		}

		// Get all records that match the filters (before validation)
		allRecords, err := repo.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list records: %w", err)
		}

		// Apply filters to get the records we're checking
		candidateRecords := filterRecordsForDisplay(allRecords, filters)

		if len(candidateRecords) == 0 {
			fmt.Println("\nNo records to check.")
			return nil
		}

		// Perform revalidation
		var invalidRecords []revalidate.InvalidRecordInfo

		if revalidateDryRun {
			fmt.Println("\n--- DRY RUN MODE (no changes will be made) ---")
			invalidRecords, err = revalidateUC.FindInvalid(ctx, filters)
			if err != nil {
				return fmt.Errorf("revalidation failed: %w", err)
			}
		} else {
			invalidRecords, err = revalidateUC.FindInvalidAndDrop(ctx, filters)
			if err != nil {
				return fmt.Errorf("revalidation failed: %w", err)
			}
		}

		// Create a map of invalid records for quick lookup with reasons
		invalidMap := make(map[string]string)
		for _, info := range invalidRecords {
			invalidMap[info.Record.Hostname] = info.Reason
		}

		// Print status of all records
		fmt.Printf("\nChecked %d record(s) of %d total record(s):\n\n", len(candidateRecords), len(allRecords))

		validCount := 0
		invalidCount := 0

		for i, record := range candidateRecords {
			status := "✓ VALID"
			reason, isInvalid := invalidMap[record.Hostname]
			if isInvalid {
				status = "✗ INVALID"
				invalidCount++
			} else {
				validCount++
			}

			fmt.Printf("%d. [%s] %s\n", i+1, status, record.Hostname)
			fmt.Printf("   Owner: %s\n", record.Owner)
			fmt.Printf("   Type: %s\n", record.Type)
			fmt.Printf("   GroupID: %s\n", record.GroupID)
			fmt.Printf("   ValidateTime: %s\n", record.ValidateTime.Format("2006-01-02 15:04:05"))
			if isInvalid {
				fmt.Printf("   Reason: %s\n", reason)
			}
			fmt.Println()
		}

		// Print summary
		fmt.Printf("Summary: %d valid, %d invalid\n", validCount, invalidCount)

		if invalidCount > 0 {
			if !revalidateDryRun {
				fmt.Printf("✓ Removed %d invalid record(s)\n", invalidCount)
				if revalidateFilePath != "" {
					fmt.Printf("Changes persisted to: %s\n", revalidateFilePath)
				}
			} else {
				fmt.Printf("(No changes made - dry run)\n")
			}
		}

		return nil
	},
}

// filterRecordsForDisplay applies basic filtering to determine which records to display
// This mirrors the logic in the revalidate use case but is used for display purposes
func filterRecordsForDisplay(records []*model.DomainRecord, filters revalidate.FilterOptions) []*model.DomainRecord {
	// If no filters specified, return all records
	if len(filters.Owners) == 0 && len(filters.Domains) == 0 && len(filters.GroupIDs) == 0 {
		return records
	}

	// Create lookup maps for efficient filtering
	ownerMap := make(map[string]bool)
	for _, owner := range filters.Owners {
		ownerMap[owner] = true
	}

	domainMap := make(map[string]bool)
	for _, domain := range filters.Domains {
		domainMap[domain] = true
	}

	groupIDMap := make(map[string]bool)
	for _, groupID := range filters.GroupIDs {
		groupIDMap[groupID] = true
	}

	var filtered []*model.DomainRecord

	for _, record := range records {
		// Apply owner filter
		if len(filters.Owners) > 0 && !ownerMap[record.Owner] {
			continue
		}

		// Apply domain filter - note: the use case expands this to include whole groups,
		// but for display we just show records that match initially
		if len(filters.Domains) > 0 && !domainMap[record.Hostname] {
			continue
		}

		// Apply groupID filter
		if len(filters.GroupIDs) > 0 && !groupIDMap[record.GroupID] {
			continue
		}

		filtered = append(filtered, record)
	}

	return filtered
}

func init() {
	revalidateCmd.Flags().StringVarP(&revalidateFilePath, "file", "f", "", "Path to JSON file for persistence")
	revalidateCmd.Flags().StringVarP(&revalidateDynamoName, "dynamo", "d", "", "DynamoDB table name for persistence")
	revalidateCmd.Flags().StringSliceVarP(&revalidateOwners, "owner", "o", []string{}, "Filter by owner (can be repeated)")
	revalidateCmd.Flags().StringSliceVarP(&revalidateDomains, "domain", "n", []string{}, "Filter by domain name (can be repeated)")
	revalidateCmd.Flags().StringSliceVarP(&revalidateGroupIDs, "groupid", "g", []string{}, "Filter by group ID (can be repeated)")
	revalidateCmd.Flags().BoolVarP(&revalidateDryRun, "dry-run", "r", false, "Show what would be removed without making changes")
}
