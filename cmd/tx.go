package cmd

import "github.com/spf13/cobra"

// TxCommand regesters a new tx command.
func txCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "query things about a chain",
	}

	cmd.AddCommand(bankSendCmd())
	cmd.AddCommand(stakingDelegateCmd())
	cmd.AddCommand(stakingRedelegateCmd())

	return cmd
}
