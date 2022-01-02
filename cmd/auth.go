package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
)

func getAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account [address]",
		Aliases: []string{},
		Short:   "query an account for its number and sequence or pass no arguement to query default account",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			var (
				keyNameOrAddress = ""
				address          sdk.AccAddress
				err              error
			)
			if len(args) == 0 {
				keyNameOrAddress = cl.Config.Key
			} else {
				keyNameOrAddress = args[0]
			}
			if cl.KeyExists(keyNameOrAddress) {
				cl.Config.Key = keyNameOrAddress
				address, err = cl.GetKeyAddress()
			} else {
				address, err = cl.DecodeBech32AccAddr(keyNameOrAddress)
			}
			if err != nil {
				return err
			}
			addr, err := cl.EncodeBech32AccAddr(address)
			if err != nil {
				return err
			}
			res, err := authtypes.NewQueryClient(cl).Account(cmd.Context(), &authtypes.QueryAccountRequest{Address: addr})
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
		Short:   "query all accounts on a given chain w/ pagination",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			pr, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			res, err := authtypes.NewQueryClient(cl).Accounts(cmd.Context(), &authtypes.QueryAccountsRequest{Pagination: pr})
			if err != nil {
				return err
			}
			return cl.PrintObject(res)
		},
	}
	return paginationFlags(cmd)
}

func getParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "parameters",
		Aliases: []string{"param", "params", "p"},
		Short:   "query the current auth parameters",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			res, err := authtypes.NewQueryClient(cl).Params(cmd.Context(), &authtypes.QueryParamsRequest{})
			if err != nil {
				return err
			}
			return cl.PrintObject(res)
		},
	}
	return cmd
}
