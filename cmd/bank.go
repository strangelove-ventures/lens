package cmd

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"
	query "github.com/strangelove-ventures/lens/client/query"
)

func bankSendCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [from] [to] [amount]",
		Short: "send coins from one address to another",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			fromAddr, err := cl.AccountFromKeyOrAddress(args[0])
			if err != nil {
				return err
			}

			toAddr, err := cl.DecodeBech32AccAddr(args[1])
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return fmt.Errorf("parsing coin string (i.e. 20000uatom): %s", err)
			}

			req := &banktypes.MsgSend{
				FromAddress: cl.MustEncodeAccAddr(fromAddr),
				ToAddress:   cl.MustEncodeAccAddr(toAddr),
				Amount:      coins,
			}

			memo, err := cmd.Flags().GetString(flagMemo)
			if err != nil {
				return err
			}

			res, err := cl.SendMsg(cmd.Context(), req, memo)
			if err != nil {
				if res != nil {
					return fmt.Errorf("failed to send coins: code(%d) msg(%s)", res.Code, res.Logs)
				}
				return fmt.Errorf("failed to send coins: err(%w)", err)
			}
			return cl.PrintTxResponse(res)

		},
	}
	memoFlag(a.Viper, cmd)
	return cmd
}

// ========== Querier Functions ==========

func bankBalanceCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "balances [key-or-address]",
		Aliases: []string{"bal", "b"},
		Short:   "query the account balance for a key or address (if none is specified, the balance of the default account is returned)",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			pr, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			height, err := ReadHeight(cmd.Flags())
			if err != nil {
				return err
			}
			keyNameOrAddress := ""
			if len(args) == 0 {
				keyNameOrAddress = cl.Config.Key
			} else {
				keyNameOrAddress = args[0]
			}
			address, err := cl.AccountFromKeyOrAddress(keyNameOrAddress)
			if err != nil {
				return err
			}
			encodedAddr := cl.MustEncodeAccAddr(address)
			options := query.QueryOptions{Pagination: pr, Height: height}
			query := query.Query{Client: cl, Options: &options}
			balance, err := query.Bank_Balances(encodedAddr)
			if err != nil {
				return err
			}
			return cl.PrintObject(balance)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "balance")
	return cmd
}

func bankTotalSupplyCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "total-supply",
		Aliases: []string{"totalsupply", "tot", "ts", "totsupplys"},
		Short:   "query the total supply of coins in the chain",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			pr, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			height, err := ReadHeight(cmd.Flags())
			if err != nil {
				return err
			}
			options := query.QueryOptions{Pagination: pr, Height: height}
			query := query.Query{Client: cl, Options: &options}
			totalSupply, err := query.Bank_TotalSupply()
			if err != nil {
				return err
			}
			return cl.PrintObject(totalSupply)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "total-supply")
	return cmd
}

func bankDenomsMetadataCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "denoms-metadata",
		Aliases: []string{"denoms", "d"},
		Short:   "query the denoms metadata",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			pr, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			height, err := ReadHeight(cmd.Flags())
			if err != nil {
				return err
			}
			options := query.QueryOptions{Pagination: pr, Height: height}
			query := query.Query{Client: cl, Options: &options}
			denoms, err := query.Bank_DenomsMetadata()
			if err != nil {
				return err
			}
			return cl.PrintObject(denoms)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "denoms-metadata")
	return cmd
}
