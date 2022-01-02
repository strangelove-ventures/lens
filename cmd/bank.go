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
			var (
				fromAddress sdk.AccAddress
				err         error
			)
			if cl.KeyExists(args[0]) {
				fromAddress, err = cl.GetKeyAddress()
			} else {
				fromAddress, err = cl.DecodeBech32AccAddr(args[0])
			}
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

			from, _ := cl.EncodeBech32AccAddr(fromAddress)
			to, _ := cl.EncodeBech32AccAddr(toAddr)

			res, ok, err := cl.SendMsg(cmd.Context(), &banktypes.MsgSend{FromAddress: from, ToAddress: to, Amount: coins})
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

func getBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "balance [key-or-address]",
		Aliases: []string{"bal", "b"},
		Short:   "query the account balance for a key or address, if none is passed will query the balance of the default account",
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
			balance, err := cl.QueryBalance(address, false)
			if err != nil {
				return err
			}
			return cl.PrintObject(balance)
		},
	}
	return cmd
}
