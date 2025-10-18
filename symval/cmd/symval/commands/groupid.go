package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/service/groupid"
	"github.com/spf13/cobra"
)

var groupidCmd = &cobra.Command{
	Use:   "groupid <owner> <type> <hostname1> [hostname2] [hostname3...]",
	Short: "Calculate a group ID",
	Long: `Calculate a group ID by hashing owner and all hostnames, prepending type and version.

Arguments:
  owner      Owner of the group
  type       Type of the group (one of: ` + getAvailableTypes() + `)
  hostname   One or more hostnames (at least one required)`,
	Args: cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := args[0]
		typeName := strings.ToLower(args[1])
		hostnames := args[2:]

		// Convert type name to code
		typeCode, ok := model.TypeNameToCode[typeName]
		if !ok {
			return fmt.Errorf("invalid type %q, must be one of: %s", args[1], getAvailableTypes())
		}

		// Create service and calculate group ID
		service := groupid.NewService()
		groupID, err := service.CalculateV1(owner, typeCode, hostnames)
		if err != nil {
			return fmt.Errorf("failed to calculate group ID: %w", err)
		}

		fmt.Println(groupID)

		return nil
	},
}

// getAvailableTypes returns a comma-separated list of available type names
func getAvailableTypes() string {
	types := make([]string, 0, len(model.TypeNameToCode))
	for name := range model.TypeNameToCode {
		types = append(types, name)
	}
	sort.Strings(types)
	return strings.Join(types, ", ")
}
