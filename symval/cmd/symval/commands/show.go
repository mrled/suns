package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/presenter"
	"github.com/mrled/suns/symval/internal/repository"
	"github.com/spf13/cobra"
)

var showFlags struct {
	PersistenceFlags
	Owner   string
	GroupID string
	Domain  string
	Format  string
	SortBy  string
}

var showCmd = &cobra.Command{
	Use:           "show",
	Short:         "Show records from the data store",
	GroupID:       "attestation",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `Display records from the data store filtered by owner, group ID, or domain.

This command allows you to view the stored attestation records with various filters.
If no filters are specified, all records are displayed.

Examples:
  # Show all records
  symval show --file ./data.json

  # Show records for a specific owner
  symval show --file ./data.json --owner alice@example.com

  # Show records for a specific group
  symval show --file ./data.json --group-id abc123

  # Show records for a specific domain
  symval show --file ./data.json --domain example.com

  # Show records with multiple filters (AND operation)
  symval show --file ./data.json --owner alice@example.com --group-id abc123

  # Show records sorted by validation time
  symval show --file ./data.json --sort validate-time

  # Show records in compact format
  symval show --file ./data.json --format compact`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Create repository based on persistence flags
		repo, err := repository.NewRepository(ctx, repository.RepositoryConfig{
			FilePath:       showFlags.FilePath,
			DynamoTable:    showFlags.DynamoTable,
			DynamoEndpoint: showFlags.DynamoEndpoint,
		})
		if err != nil {
			return err
		}

		// Get all records from repository
		allRecords, err := repo.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list records: %w", err)
		}

		// Filter records based on flags
		filter := model.RecordFilter{}
		if showFlags.Owner != "" {
			filter.Owners = []string{showFlags.Owner}
		}
		if showFlags.GroupID != "" {
			filter.GroupIDs = []string{showFlags.GroupID}
		}
		if showFlags.Domain != "" {
			filter.Domains = []string{showFlags.Domain}
		}
		filteredRecords := model.FilterRecords(allRecords, filter)

		// Sort records
		model.SortRecords(filteredRecords, showFlags.SortBy)

		// Display results
		if len(filteredRecords) == 0 {
			fmt.Println("\nNo records found matching the specified criteria.")
			return nil
		}

		// Display based on format
		switch showFlags.Format {
		case "compact":
			displayRecordsCompact(filteredRecords)
		default: // "detailed" or empty
			displayRecordsDetailed(filteredRecords)
		}

		// Print summary
		fmt.Printf("\nTotal records: %d\n", len(filteredRecords))

		// If filtering was applied, show filter summary
		if showFlags.Owner != "" || showFlags.GroupID != "" || showFlags.Domain != "" {
			fmt.Printf("Filters applied:\n")
			if showFlags.Owner != "" {
				fmt.Printf("  Owner: %s\n", showFlags.Owner)
			}
			if showFlags.GroupID != "" {
				fmt.Printf("  Group ID: %s\n", showFlags.GroupID)
			}
			if showFlags.Domain != "" {
				fmt.Printf("  Domain: %s\n", showFlags.Domain)
			}
		}

		return nil
	},
}

// displayRecordsDetailed displays records in detailed format
func displayRecordsDetailed(records []*model.DomainRecord) {
	fmt.Println("\n=== Domain Records ===")

	// Group records by GroupID for better display
	grouped := model.GroupByGroupID(records)

	for groupID, groupRecords := range grouped {
		fmt.Printf("\nGroup ID: %s\n", groupID)
		fmt.Printf("Type: %s\n", groupRecords[0].Type)
		fmt.Printf("Owner: %s\n", groupRecords[0].Owner)
		fmt.Printf("Domains (%d):\n", len(groupRecords))

		for _, record := range groupRecords {
			timeStr := presenter.FormatTimeSince(record.ValidateTime)

			fmt.Printf("  - %s (validated: %s, rev: %d)\n",
				record.Hostname,
				timeStr,
				record.Rev)
		}
	}
}

// displayRecordsCompact displays records in compact format
func displayRecordsCompact(records []*model.DomainRecord) {
	fmt.Println("\n=== Domain Records (Compact) ===")
	fmt.Printf("%-40s %-30s %-20s %-15s %s\n", "Domain", "Owner", "Type", "Group ID", "Last Validated")
	fmt.Println(strings.Repeat("-", 120))

	for _, record := range records {
		timeStr := presenter.FormatTimeSinceCompact(record.ValidateTime)

		// Truncate long fields for compact display
		domain := truncateString(record.Hostname, 38)
		owner := truncateString(record.Owner, 28)
		groupID := truncateString(record.GroupID, 13)

		fmt.Printf("%-40s %-30s %-20s %-15s %s\n",
			domain,
			owner,
			record.Type,
			groupID,
			timeStr)
	}
}

// truncateString truncates a string to the specified length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func init() {
	// Add persistence flags
	showCmd.Flags().StringVarP(&showFlags.FilePath, "file", "f", "", "Path to JSON file for persistence")
	showCmd.Flags().StringVarP(&showFlags.DynamoTable, "dynamodb-table", "t", "", "DynamoDB table name for persistence")
	showCmd.Flags().StringVarP(&showFlags.DynamoEndpoint, "dynamodb-endpoint", "e", "", "DynamoDB endpoint URL (optional, uses AWS SDK default if not specified)")

	// Add filter flags
	showCmd.Flags().StringVarP(&showFlags.Owner, "owner", "o", "", "Filter by owner email")
	showCmd.Flags().StringVarP(&showFlags.GroupID, "group-id", "g", "", "Filter by group ID")
	showCmd.Flags().StringVarP(&showFlags.Domain, "domain", "d", "", "Filter by domain name")

	// Add format and sort flags
	showCmd.Flags().StringVar(&showFlags.Format, "format", "detailed", "Output format: detailed or compact")
	showCmd.Flags().StringVar(&showFlags.SortBy, "sort", "", "Sort by: owner, domain, group, validate-time, or type")
}
