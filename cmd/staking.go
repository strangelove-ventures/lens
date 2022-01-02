package cmd

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
)

func stakingDelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate [validator-addr] [amount] ",
		Args:  cobra.ExactArgs(2),
		Short: "Delegate liquid tokens to a validator",
		Long: strings.TrimSpace(
			`Delegate an amount of liquid coins to a validator from your wallet.
Example:
$ lens tx staking delegate cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0 1000stake --from mykey`,
		),

		RunE: func(cmd *cobra.Command, args []string) error {

			var (
				delAddr sdk.AccAddress
				err     error
			)

			cl := config.GetDefaultClient()

			if args[2] != cl.Config.Key {
				cl.Config.Key = args[2]
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			key, _ := cmd.Flags().GetString(FlagFrom)
			if key != "" {
				if key != cl.Config.Key {
					cl.Config.Key = key
				}
			}

			if cl.KeyExists(cl.Config.Key) {
				delAddr, err = cl.GetKeyAddress()
			} else {
				delAddr, err = cl.DecodeBech32AccAddr(key)
			}
			if err != nil {
				return err
			}

			valAddr, err := cl.DecodeBech32ValAddr(args[0])
			if err != nil {
				return err
			}

			// msg := types.NewMsgDelegate(delAddr, valAddr, amount)
			msg := &types.MsgDelegate{
				DelegatorAddress: cl.MustEncodeAccAddr(delAddr),
				ValidatorAddress: cl.MustEncodeValAddr(valAddr),
				Amount:           amount,
			}
			return cl.HandleAndPrintMsgSend(cl.SendMsg(cmd.Context(), msg))

		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func stakingRedelegateCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "redelegate [from] [src-validator-addr] [dst-validator-addr] [amount]",
		Short: "Redelegate illiquid tokens from one validator to another",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			`Redelegate an amount of illiquid staking tokens from one validator to another.
Example:
$ lens tx staking redelegate cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0 cosmosvaloper1a3yjj7d3qnx4spgvjcwjq9cw9snrrrhu5h6jll 100stake --from mykey
`,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			key, _ := cmd.Flags().GetString(FlagFrom)
			delAddr, err := cl.AccountFromKeyOrAddress(key)
			if err != nil {
				return err
			}

			valSrcAddr, err := cl.DecodeBech32ValAddr(args[0])
			if err != nil {
				return err
			}

			valDstAddr, err := cl.DecodeBech32ValAddr(args[1])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := &types.MsgBeginRedelegate{
				DelegatorAddress:    cl.MustEncodeAccAddr(delAddr),
				ValidatorSrcAddress: cl.MustEncodeValAddr(sdk.ValAddress(valSrcAddr)),
				ValidatorDstAddress: cl.MustEncodeValAddr(sdk.ValAddress(valDstAddr)),
				Amount:              amount,
			}

			return cl.HandleAndPrintMsgSend(cl.SendMsg(cmd.Context(), msg))
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
