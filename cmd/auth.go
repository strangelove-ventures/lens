package cmd

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
)

// getAuthQueryCmd returns the transaction commands for this module
func getAuthQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Querying commands for the auth module",
	}

	cmd.AddCommand(
		getAccountCmd(),
		getAccountsCmd(),
		queryParamsCmd(),
	)

	return cmd
}

func getAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account [address]",
		Aliases: []string{},
		Short:   "query an account for its number and sequence",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			res, err := authtypes.NewQueryClient(cl).Account(cmd.Context(), &authtypes.QueryAccountRequest{Address: args[0]})
			if err != nil {
				return err
			}
			return cl.PrintObject(res)
		},
	}
	return cmd
}
func getAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "accounts",
		Aliases: []string{},
		Short:   "",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}
func queryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "parameters",
		Aliases: []string{"param", "params", "p"},
		Short:   "query the current auth parameters",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}
