package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mrled/suns/symval/internal/groupid"
	"github.com/mrled/suns/symval/internal/symgroup"
	"github.com/spf13/cobra"
)

var groupidCmd = &cobra.Command{
	Use:     "groupid <owner> <type> <hostname1> [hostname2] [hostname3...]",
	Short:   "Calculate a group ID",
	GroupID: "attestation",
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
		typeCode, ok := symgroup.TypeNameToCode[typeName]
		if !ok {
			return fmt.Errorf("invalid type %q, must be one of: %s", args[1], getAvailableTypes())
		}

		// Calculate group ID
		groupID, err := groupid.CalculateV1(owner, typeCode, hostnames)
		if err != nil {
			return fmt.Errorf("failed to calculate group ID: %w", err)
		}

		fmt.Println(groupID)

		return nil
	},
}

// getAvailableTypes returns a comma-separated list of available type names
func getAvailableTypes() string {
	types := make([]string, 0, len(symgroup.TypeNameToCode))
	for name := range symgroup.TypeNameToCode {
		types = append(types, name)
	}
	sort.Strings(types)
	return strings.Join(types, ", ")
}
