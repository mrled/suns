package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/mrled/suns/symval/internal/repository"
	"github.com/mrled/suns/symval/internal/repository/memrepo"
	"github.com/mrled/suns/symval/internal/service/dnsclaims"
	"github.com/mrled/suns/symval/internal/symgroup"
	"github.com/mrled/suns/symval/internal/usecase/attestation"
	"github.com/spf13/cobra"
)

var attestFlags PersistenceFlags

var attestCmd = &cobra.Command{
	Use:           "attest <owner> <type> <domain1> [domain2]...",
	Short:         "Attest a group of domains for consistency and validity",
	GroupID:       "attestation",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `Attest verifies that a group of domains forms a valid symmetric group.

It performs the following checks:
  1. Calculates the expected group ID based on owner and domains
  2. Looks up DNS TXT records (_suns.<domain>) for all domains
  3. Checks that all group IDs are consistent (same owner hash)
  4. Validates the group according to its symmetry type

The type can be specified as either a name or code:
  palindrome (a)    - Domain names that read the same forwards and backwards
  flip180 (b)       - Domain names that look the same when rotated 180 degrees
  doubleflip180 (c) - Two domains that flip180 relative to each other
  mirrortext (d)    - Domain names that mirror each other visually
  mirrornames (e)   - Domain names with parts that mirror each other
  antonymnames (f)  - Domain names with antonym parts

Example:
  symval attest myowner palindrome example.com test.com
  symval attest myowner a example.com test.com
  symval attest owner123 mirrortext domain1.com domain2.com domain3.com`,
	Args: cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		owner := args[0]
		typeName := strings.ToLower(args[1])
		domains := args[2:]

		// Convert type name to type code
		typeCode, ok := symgroup.TypeNameToCode[typeName]
		if !ok {
			// Check if it's already a valid type code
			if _, codeExists := symgroup.TypeCodeToName[typeName]; codeExists {
				typeCode = typeName
			} else {
				cmd.SilenceUsage = false
				validTypesMsg := "Valid types: palindrome (a), flip180 (b), doubleflip180 (c), mirrortext (d), mirrornames (e), antonymnames (f)"
				return UsageError{fmt.Errorf("invalid symmetry type: %s\n%s", typeName, validTypesMsg)}
			}
		}

		symmetryType := symgroup.SymmetryType(typeCode)

		// Create repository based on persistence flags
		var repo attestation.DomainRepository
		if attestFlags.DynamoTable != "" || attestFlags.FilePath != "" {
			// Use persistent repository (file or DynamoDB)
			r, err := repository.NewRepository(ctx, repository.RepositoryConfig{
				FilePath:       attestFlags.FilePath,
				DynamoTable:    attestFlags.DynamoTable,
				DynamoEndpoint: attestFlags.DynamoEndpoint,
			})
			if err != nil {
				return err
			}
			repo = r
		} else {
			// Use in-memory only (no persistence)
			repo = memrepo.NewMemoryRepository()
		}

		// Create DNS service and attestation use case
		dnsService := dnsclaims.NewService()
		attestUseCase := attestation.NewAttestationUseCase(dnsService, repo)

		// Perform attestation
		result, err := attestUseCase.Attest(owner, symmetryType, domains)
		if err != nil {
			return ExitWithCode(1, fmt.Errorf("attestation failed: %w", err))
		}

		// Print results
		fmt.Printf("Expected Group ID: %s\n", result.ExpectedID)
		fmt.Printf("Found %d group ID(s) in DNS records\n", len(result.GroupIDs))

		if result.IsValid {
			fmt.Println("\n✓ Attestation PASSED")
			fmt.Println("The domains form a valid symmetric group.")
			if attestFlags.DynamoTable != "" {
				fmt.Printf("Results persisted to DynamoDB table: %s\n", attestFlags.DynamoTable)
			} else if attestFlags.FilePath != "" {
				fmt.Printf("Results persisted to: %s\n", attestFlags.FilePath)
			}
		} else {
			fmt.Println("\n✗ Attestation FAILED")
			if result.ErrorMessage != "" {
				fmt.Printf("Reason: %s\n", result.ErrorMessage)
			}
		}

		return nil
	},
}

func init() {
	addPersistenceFlags(attestCmd, &attestFlags)
}
