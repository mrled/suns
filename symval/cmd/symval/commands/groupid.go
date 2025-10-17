package commands

import (
	"fmt"

	"github.com/mrled/suns/symval/internal/service/groupid"
	"github.com/spf13/cobra"
)

var groupidCmd = &cobra.Command{
	Use:   "groupid <owner> <type> <hostname1> [hostname2] [hostname3...]",
	Short: "Calculate a group ID",
	Long: `Calculate a group ID by hashing owner and all hostnames, prepending type and version.

Arguments:
  owner      Owner of the group
  type       Type of the group
  hostname   One or more hostnames (at least one required)`,
	Args: cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := args[0]
		gtype := args[1]
		hostnames := args[2:]

		// Create service and calculate group ID
		service := groupid.NewService()
		groupID, err := service.CalculateV1(owner, gtype, hostnames)
		if err != nil {
			return fmt.Errorf("failed to calculate group ID: %w", err)
		}

		fmt.Println(groupID)

		return nil
	},
}
