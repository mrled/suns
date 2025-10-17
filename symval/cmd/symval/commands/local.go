package commands

import (
	"github.com/spf13/cobra"
)

var (
	persistenceFile string
)

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Manage local domain validations",
	Long:  `Perform domain validation operations using local storage.`,
}

func init() {
	localCmd.PersistentFlags().StringVarP(&persistenceFile, "file", "f", "", "JSON file for persistence (optional)")
	localCmd.AddCommand(validateCmd)
	localCmd.AddCommand(groupidCmd)
	localCmd.AddCommand(verifyCmd)
}
