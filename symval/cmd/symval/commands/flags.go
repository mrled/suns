package commands

import (
	"github.com/spf13/cobra"
)

// PersistenceFlags holds flags related to persistence and data storage options
type PersistenceFlags struct {
	FilePath       string
	DynamoTable    string
	DynamoEndpoint string
	DryRun         bool
}

// addPersistenceFlags adds common persistence-related flags to a command
func addPersistenceFlags(cmd *cobra.Command, flags *PersistenceFlags) {
	cmd.Flags().StringVarP(&flags.FilePath, "file", "f", "", "Path to JSON file for persistence")
	cmd.Flags().StringVarP(&flags.DynamoTable, "dynamodb-table", "t", "", "DynamoDB table name for persistence")
	cmd.Flags().StringVarP(&flags.DynamoEndpoint, "dynamodb-endpoint", "e", "", "DynamoDB endpoint URL (optional, uses AWS SDK default if not specified)")
	cmd.Flags().BoolVarP(&flags.DryRun, "dry-run", "r", false, "Show what would be changed without making changes")
}
