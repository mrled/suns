package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrled/suns/symval/internal/model"
	"github.com/mrled/suns/symval/internal/symgroup"
	"github.com/mrled/suns/symval/internal/validation"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <owner> <type> <groupid> <hostname1> [hostname2] [hostname3...]",
	Short: "Validate a domain group",
	Long: `Validate a domain group with specified owner, type, group ID, and one or more hostnames.

Arguments:
  owner      Owner of the domain
  type       Type of validation (one of: ` + getAvailableTypes() + `)
  groupid    Group ID for the domain
  hostname1  First hostname to validate
  hostname2+ Additional hostnames (optional)`,
	Args: cobra.MinimumNArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := args[0]
		typeName := strings.ToLower(args[1])
		groupID := args[2]
		hostnames := args[3:]

		// Convert type name to code
		typeCode, ok := symgroup.TypeNameToCode[typeName]
		if !ok {
			return fmt.Errorf("invalid type %q, must be one of: %s", args[1], getAvailableTypes())
		}

		// Create DomainData structs from arguments
		dataList := make([]*model.DomainData, 0, len(hostnames))
		validateTime := time.Now()

		for _, hostname := range hostnames {
			data := &model.DomainData{
				Owner:        owner,
				Type:         symgroup.SymmetryType(typeCode),
				Hostname:     hostname,
				GroupID:      groupID,
				ValidateTime: validateTime,
			}
			dataList = append(dataList, data)
		}

		// Validate
		valid, err := validation.Validate(dataList)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}

		// Echo the input values
		fmt.Printf("Owner: %s\n", owner)
		fmt.Printf("Type: %s (%s)\n", typeName, typeCode)
		fmt.Printf("Group ID: %s\n", groupID)
		fmt.Printf("Hostnames: %v\n", hostnames)

		fmt.Printf("Valid: %t\n", valid)

		return nil
	},
}
