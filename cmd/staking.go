package cmd

import (
	"strings"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/lens/client/query"
)

func stakingDelegateCmd(lc *lensConfig) *cobra.Command {
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

			cl := lc.config.GetDefaultClient()

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

func stakingRedelegateCmd(lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegate [from] [src-validator-addr] [dst-validator-addr] [amount]",
		Short: "redelegate tokens from one validator to another",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			`Redelegate an amount of illiquid staking tokens from one validator to another.
Example:
$ lens tx staking redelegate cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0 cosmosvaloper1a3yjj7d3qnx4spgvjcwjq9cw9snrrrhu5h6jll 100stake --from mykey
`,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := lc.config.GetDefaultClient()
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

func stakingDelegationsCmd(lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegations [delegator-addr]",
		Short: "query all delegations for a delegator address",
		Long: strings.TrimSpace(`query delegations for an individual delegator on all validators.

Example:
$ lens query staking delegations cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := lc.config.GetDefaultClient()
			pr, err := sdkclient.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			height, err := ReadHeight(cmd.Flags())
			if err != nil {
				return err
			}
			options := query.QueryOptions{Pagination: pr, Height: height}
			query := query.Query{cl, &options}
			response, err := query.Delegations(args[0])
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

func stakingDelegationCmd(lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegation [delegator-addr] [validator-addr]",
		Short: "query a delegation based on a delegator address and validator address",
		Long: strings.TrimSpace(`query delegations for an individual delegator on an individual validator.

Example:
$ lens query staking delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := lc.config.GetDefaultClient()
			cq := query.Query{Client: cl, Options: query.DefaultOptions()}
			delegator := args[0]
			validator := args[1]
			response, err := cq.Delegation(delegator, validator)
			if err != nil {
				return err
			}
			return cl.PrintObject(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func stakingValidatorDelegationsCmd(lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validator-delegations [validator-addr]",
		Aliases: []string{"valdel", "vd"},
		Short:   "query all delegations for a validator address",
		Long: strings.TrimSpace(`query delegations for an individual validator.

Example:
$ lens query staking validator-delegations [validator address (valoper)]
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := lc.config.GetDefaultClient()
			pr, err := sdkclient.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			height, err := ReadHeight(cmd.Flags())
			if err != nil {
				return err
			}
			validator := args[0]
			options := query.QueryOptions{Pagination: pr, Height: height}
			query := query.Query{cl, &options}
			response, err := query.ValidatorDelegations(validator)
			if err != nil {
				return err
			}
			return cl.PrintObject(response)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "validator-delegations")
	return cmd
}
