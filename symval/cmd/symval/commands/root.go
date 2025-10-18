package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "symval",
	Short: "Symval is a tool for validating symmetric domains",
	Long:  `A command-line tool for validating and managing symmetric domain names.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(groupidCmd)
	rootCmd.AddCommand(lookupCmd)
	rootCmd.AddCommand(validateCmd)
}
