package commands

import (
	"context"
	"fmt"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository/memrepo"
	"github.com/mrled/suns/symval/internal/service/dnsclaims"
	"github.com/mrled/suns/symval/internal/usecase/reattest"
	"github.com/spf13/cobra"
)

var reattestFlags PersistenceFlags

var reattestCmd = &cobra.Command{
	Use:     "reattest",
	Short:   "Re-attest all groups in the data store",
	GroupID: "attestation",
	Long: `Re-attest loads all groups from the data store and re-attests them by querying DNS.

This command performs a fresh attestation for each group in the datastore, checking
that the DNS records are still valid and consistent. Unlike 'revalidate' which only
checks stored records without querying DNS, 'reattest' performs a full DNS lookup
and validation for each group.

By default, invalid groups are removed from the data store. Use --dry-run to see
what would be removed without actually removing anything.

Invalid groups are always printed in both regular and dry-run modes.

Examples:
  # Re-attest all groups and remove invalid ones
  symval reattest --file ./data.json

  # Dry run to see what would be removed
  symval reattest --file ./data.json --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create repository based on persistence flags
		var repo model.DomainRepository
		if reattestFlags.DynamoTable != "" {
			return fmt.Errorf("--dynamodb-table flag is not yet implemented")
		} else if reattestFlags.FilePath != "" {
			// Use JSON file persistence
			memRepo, err := memrepo.NewMemoryRepositoryWithPersistence(reattestFlags.FilePath)
			if err != nil {
				return fmt.Errorf("failed to create repository: %w", err)
			}
			repo = memRepo
			fmt.Printf("Using JSON persistence: %s\n", reattestFlags.FilePath)
		} else {
			// Use in-memory only (no persistence)
			repo = memrepo.NewMemoryRepository()
			fmt.Println("Using in-memory storage (no persistence)")
		}

		// Create DNS service
		dnsService := dnsclaims.NewService()

		// Create reattest use case
		reattestUC := reattest.NewReattestUseCase(dnsService, repo)
		ctx := context.Background()

		// Perform re-attestation
		var results []reattest.GroupAttestResult
		var err error

		if reattestFlags.DryRun {
			fmt.Println("\n--- DRY RUN MODE (no changes will be made) ---")
			results, err = reattestUC.ReattestAll(ctx)
			if err != nil {
				return fmt.Errorf("re-attestation failed: %w", err)
			}
		} else {
			results, err = reattestUC.ReattestAllAndDrop(ctx)
			if err != nil {
				return fmt.Errorf("re-attestation failed: %w", err)
			}
		}

		if len(results) == 0 {
			fmt.Println("\nNo groups to re-attest.")
			return nil
		}

		// Print results
		fmt.Printf("\nRe-attested %d group(s):\n\n", len(results))

		validCount := 0
		invalidCount := 0

		for i, result := range results {
			status := "✓ VALID"
			if !result.IsValid {
				status = "✗ INVALID"
				invalidCount++
			} else {
				validCount++
			}

			fmt.Printf("%d. [%s] Group\n", i+1, status)
			fmt.Printf("   Owner: %s\n", result.Owner)
			fmt.Printf("   Type: %s\n", result.Type)
			fmt.Printf("   GroupID: %s\n", result.GroupID)
			fmt.Printf("   Domains: %v\n", result.Domains)
			if !result.IsValid {
				fmt.Printf("   Error: %s\n", result.ErrorMessage)
			}
			fmt.Println()
		}

		// Print summary
		fmt.Printf("Summary: %d valid, %d invalid\n", validCount, invalidCount)

		if invalidCount > 0 {
			if !reattestFlags.DryRun {
				fmt.Printf("✓ Removed %d invalid group(s)\n", invalidCount)
				if reattestFlags.FilePath != "" {
					fmt.Printf("Changes persisted to: %s\n", reattestFlags.FilePath)
				}
			} else {
				fmt.Printf("(No changes made - dry run)\n")
			}
		}

		return nil
	},
}

func init() {
	addPersistenceFlags(reattestCmd, &reattestFlags)
}
