package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/repository/dynamorepo"
	"github.com/mrled/suns/symval/internal/repository/memrepo"
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
		var repo model.DomainRepository
		if showFlags.DynamoTable != "" {
			// Use DynamoDB persistence
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				return fmt.Errorf("failed to load AWS config: %w", err)
			}

			// Create DynamoDB client
			var client *dynamodb.Client
			if showFlags.DynamoEndpoint != "" {
				// Use custom endpoint if specified
				client = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
					o.BaseEndpoint = &showFlags.DynamoEndpoint
				})
				fmt.Printf("Using DynamoDB endpoint: %s\n", showFlags.DynamoEndpoint)
			} else {
				// Use default endpoint discovery
				client = dynamodb.NewFromConfig(cfg)
			}

			repo = dynamorepo.NewDynamoRepository(client, showFlags.DynamoTable)
			fmt.Printf("Using DynamoDB table: %s\n", showFlags.DynamoTable)
		} else if showFlags.FilePath != "" {
			// Use JSON file persistence
			memRepo, err := memrepo.NewMemoryRepositoryWithPersistence(showFlags.FilePath)
			if err != nil {
				return fmt.Errorf("failed to create repository: %w", err)
			}
			repo = memRepo
			fmt.Printf("Using JSON persistence: %s\n", showFlags.FilePath)
		} else {
			return fmt.Errorf("must specify either --file or --dynamodb-table")
		}

		// Get all records from repository
		allRecords, err := repo.List(ctx)
		if err != nil {
			return fmt.Errorf("failed to list records: %w", err)
		}

		// Filter records based on flags
		filteredRecords := filterRecords(allRecords, showFlags.Owner, showFlags.GroupID, showFlags.Domain)

		// Sort records
		sortRecords(filteredRecords, showFlags.SortBy)

		// Display results
		if len(filteredRecords) == 0 {
			fmt.Println("\nNo records found matching the specified criteria.")
			return nil
		}

		// Display based on format
		switch showFlags.Format {
		case "compact":
			displayRecordsCompact(filteredRecords)
		case "json":
			displayRecordsJSON(filteredRecords)
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

// filterRecords filters records based on the provided criteria
func filterRecords(records []*model.DomainRecord, owner, groupID, domain string) []*model.DomainRecord {
	var filtered []*model.DomainRecord

	for _, record := range records {
		// Check if record matches all specified filters (AND operation)
		matches := true

		if owner != "" && !strings.EqualFold(record.Owner, owner) {
			matches = false
		}

		if groupID != "" && record.GroupID != groupID {
			matches = false
		}

		if domain != "" && !strings.EqualFold(record.Hostname, domain) {
			matches = false
		}

		if matches {
			filtered = append(filtered, record)
		}
	}

	return filtered
}

// sortRecords sorts records based on the specified field
func sortRecords(records []*model.DomainRecord, sortBy string) {
	switch sortBy {
	case "owner":
		sort.Slice(records, func(i, j int) bool {
			return records[i].Owner < records[j].Owner
		})
	case "domain":
		sort.Slice(records, func(i, j int) bool {
			return records[i].Hostname < records[j].Hostname
		})
	case "group":
		sort.Slice(records, func(i, j int) bool {
			return records[i].GroupID < records[j].GroupID
		})
	case "validate-time":
		sort.Slice(records, func(i, j int) bool {
			return records[i].ValidateTime.After(records[j].ValidateTime)
		})
	case "type":
		sort.Slice(records, func(i, j int) bool {
			return records[i].Type < records[j].Type
		})
	default:
		// Default sort by group ID, then by hostname
		sort.Slice(records, func(i, j int) bool {
			if records[i].GroupID != records[j].GroupID {
				return records[i].GroupID < records[j].GroupID
			}
			return records[i].Hostname < records[j].Hostname
		})
	}
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
			timeSince := time.Since(record.ValidateTime)
			var timeStr string
			if timeSince < time.Hour {
				timeStr = fmt.Sprintf("%.0f minutes ago", timeSince.Minutes())
			} else if timeSince < 24*time.Hour {
				timeStr = fmt.Sprintf("%.1f hours ago", timeSince.Hours())
			} else {
				timeStr = fmt.Sprintf("%.0f days ago", timeSince.Hours()/24)
			}

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
		timeSince := time.Since(record.ValidateTime)
		var timeStr string
		if timeSince < time.Hour {
			timeStr = fmt.Sprintf("%.0fm ago", timeSince.Minutes())
		} else if timeSince < 24*time.Hour {
			timeStr = fmt.Sprintf("%.1fh ago", timeSince.Hours())
		} else {
			timeStr = fmt.Sprintf("%.0fd ago", timeSince.Hours()/24)
		}

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

// displayRecordsJSON displays records in JSON format
func displayRecordsJSON(records []*model.DomainRecord) {
	// Simple JSON-like output without external dependencies
	fmt.Println("[")
	for i, record := range records {
		fmt.Printf("  {\n")
		fmt.Printf("    \"hostname\": \"%s\",\n", record.Hostname)
		fmt.Printf("    \"owner\": \"%s\",\n", record.Owner)
		fmt.Printf("    \"type\": \"%s\",\n", record.Type)
		fmt.Printf("    \"groupId\": \"%s\",\n", record.GroupID)
		fmt.Printf("    \"validateTime\": \"%s\",\n", record.ValidateTime.Format(time.RFC3339))
		fmt.Printf("    \"revision\": %d\n", record.Rev)
		if i < len(records)-1 {
			fmt.Printf("  },\n")
		} else {
			fmt.Printf("  }\n")
		}
	}
	fmt.Println("]")
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
	showCmd.Flags().StringVar(&showFlags.Format, "format", "detailed", "Output format: detailed, compact, or json")
	showCmd.Flags().StringVar(&showFlags.SortBy, "sort", "", "Sort by: owner, domain, group, validate-time, or type")
}