package cmd

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"
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

			res, ok, err := cl.SendMsg(cmd.Context(), req)
			if err != nil || !ok {
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
			balance, err := cl.QueryBalance(address, false)
			if err != nil {
				return err
			}
			return cl.PrintObject(balance)
		},
	}
	return cmd
}
