package cmd

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/lens/client"
)

func bankSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [from] [to] [amount]",
		Short: "send coins from one address to another",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
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
				return err
			}

			req := &banktypes.MsgSend{
				FromAddress: cl.MustEncodeAccAddr(fromAddr),
				ToAddress:   cl.MustEncodeAccAddr(toAddr),
				Amount:      coins,
			}

			res, err := cl.SendMsg(cmd.Context(), req)
			if err != nil {
				if res != nil {
					return fmt.Errorf("failed to send coins: code(%d) msg(%s)", res.Code, res.Logs)
				}
				return fmt.Errorf("failed to send coins: err(%w)", err)
			}
			return cl.PrintTxResponse(res)

		},
	}
	return cmd
}

// ========== Query Functions ==========

func bankBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "balance [key-or-address]",
		Aliases: []string{"bal", "b"},
		Short:   "query the account balance for a key or address, if none is passed will query the balance of the default account",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
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
			balance, err := cl.QueryBalanceWithDenomTraces(cmd.Context(), address, client.DefaultPageRequest())
			if err != nil {
				return err
			}
			return cl.PrintObject(balance)
		},
	}
	return cmd
}

func bankTotalSupplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "total-supply [denom]?",
		Aliases: []string{"totalsupply", "tot", "ts", "totsupplys"},
		Short:   "query the total supply of coins",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			denom := ""
			if len(args) == 1 {
				denom = args[0]
			}
			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			totalSupply, err := cl.QueryTotalSupply(cmd.Context(), pageReq)
			if err != nil {
				return err
			}
			return cl.PrintObject(totalSupply)
		},
	}
	return cmd
}
func bankDenomsMetadataCmd() *cobra.Command {

}
