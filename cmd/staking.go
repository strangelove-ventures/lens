package cmd

import (
	"strings"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/lens/client/staking"
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

func stakingDelegationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegations [delegator-addr]",
		Short: "Query all delegations made by one delegator",
		Long: strings.TrimSpace(`Query delegations for an individual delegator on all validators.

Example:
$ lens query staking delegations cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			pageReq, err := sdkclient.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			response, err := staking.QueryDelegations(cl, args[0], pageReq)
			if err != nil {
				return err
			}
			return cl.PrintObject(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "delegations")

	return cmd
}

func stakingDelegationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegation [delegator-addr] [validator-addr]",
		Short: "Query a delegation based on address and validator address",
		Long: strings.TrimSpace(`Query delegations for an individual delegator on an individual validator.

Example:
$ lens query staking delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			response, err := staking.QueryDelegation(cl, args[0], args[1])
			if err != nil {
				return err
			}
			return cl.PrintObject(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
