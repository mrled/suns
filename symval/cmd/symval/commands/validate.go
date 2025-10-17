package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/callista/symval/internal/model"
	"github.com/callista/symval/internal/service/validation"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <owner> <type> <domain> [flip]",
	Short: "Validate a domain",
	Long: `Validate a domain with specified owner, type, domain, and optional flipped domain.

Arguments:
  owner   Owner of the domain
  type    Type of validation
  domain  Domain to validate
  flip    Flipped domain (optional)`,
	Args: cobra.RangeArgs(3, 4),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := args[0]
		vtype := args[1]
		domain := args[2]
		var flipPtr *string
		if len(args) == 4 {
			flip := args[3]
			flipPtr = &flip
		}

		// Create DomainData from arguments
		data := &model.DomainData{
			ValidateTime: time.Now(),
			Owner:        owner,
			Domain:       domain,
			Flip:         flipPtr,
			Type:         model.SymmetryType(vtype),
		}

		// Create validator and validate
		validator := validation.NewService()
		ctx := context.Background()
		valid, err := validator.Validate(ctx, data)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}

		// Echo the input values
		fmt.Printf("Owner: %s\n", owner)
		fmt.Printf("Type: %s\n", vtype)
		fmt.Printf("Domain: %s\n", domain)
		if flipPtr != nil {
			fmt.Printf("Flip: %s\n", *flipPtr)
		} else {
			fmt.Println("Flip: (none)")
		}

		if persistenceFile != "" {
			fmt.Printf("Persistence file: %s\n", persistenceFile)
		} else {
			fmt.Println("Persistence: in-memory only")
		}

		fmt.Printf("Valid: %t\n", valid)

		return nil
	},
}
