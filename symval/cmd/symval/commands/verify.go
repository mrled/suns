package commands

import (
	"fmt"

	"github.com/mrled/suns/symval/internal/service/dnsverification"
	"github.com/spf13/cobra"
)

var (
	resolverAddr string
)

var lookupCmd = &cobra.Command{
	Use:   "lookup <domain> [domain...]",
	Short: "Lookup DNS records for one or more domains",
	Long: `Query DNS TXT records at _suns.<domain> and display verification records.

Arguments:
  domain     One or more domain names to verify (at least one required)

For each domain, this command will:
  - Look up TXT records at _suns.<domain>
  - Display all found records, or indicate if no records were found
  - Follow CNAME records if the TXT record is not found directly`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		domains := args

		// Create DNS verification service with custom resolver
		resolver := dnsverification.NewCustomResolver(resolverAddr)
		service := dnsverification.NewServiceWithResolver(resolver)

		// Process each domain
		for _, domain := range domains {
			fmt.Printf("Domain: %s\n", domain)

			groupIDs, err := service.VerifyDomain(domain)
			if err != nil {
				fmt.Printf("  Error: %v\n", err)
			} else if len(groupIDs) == 0 {
				fmt.Println("  No _suns records found")
			} else {
				fmt.Printf("  Found %d record(s) (verified consistent):\n", len(groupIDs))
				for _, gid := range groupIDs {
					fmt.Printf("    %s\n", gid.String())
				}
			}

			// Add blank line between domains for readability (except after last one)
			if domain != domains[len(domains)-1] {
				fmt.Println()
			}
		}

		return nil
	},
}

func init() {
	lookupCmd.Flags().StringVarP(&resolverAddr, "resolver", "r", "1.1.1.1:53", "DNS resolver address (host:port)")
}
