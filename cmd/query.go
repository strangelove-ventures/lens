package cmd

import (
	"github.com/spf13/cobra"
)

// queryCmd represents the keys command
func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "query things about a chain",
	}

	cmd.AddCommand(
		queryBalanceCmd(),
		queryAccountCmd(),
	)

	return cmd
}
