package cmd

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/lens/client/query"
)

func stakingDelegateCmd(a *appState) *cobra.Command {
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

			cl := a.Config.GetDefaultClient()

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

func stakingRedelegateCmd(a *appState) *cobra.Command {
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
			cl := a.Config.GetDefaultClient()
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

func stakingParamsCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "parameters",
		Aliases: []string{"params"},
		Short:   "query things about a chain's staking params",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			opts, err := queryOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}
			query := query.Query{Client: cl, Options: opts}
			params, err := query.Staking_Params()
			if err != nil {
				return err
			}
			return cl.PrintObject(params.Params)
		},
	}

	return cmd
}

func stakingPoolCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool",
		Short: "query things about a chain's staking pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			opts, err := queryOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}
			query := query.Query{Client: cl, Options: opts}
			pool, err := query.Staking_Pool()
			if err != nil {
				return err
			}
			return cl.PrintObject(pool.Pool)
		},
	}

	return cmd
}

func stakingDelegationsCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delegations [delegator-addr]",
		Aliases: []string{"dels"},
		Short:   "query all delegations for a delegator address",
		Long: strings.TrimSpace(`query delegations for an individual delegator on all validators.

Example:
$ lens query staking delegations cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			opts, err := queryOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}
			query := query.Query{Client: cl, Options: opts}
			response, err := query.Staking_DelegatorDelegations(args[0])
			if err != nil {
				return err
			}
			return cl.PrintObject(response.DelegationResponses)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "delegations")
	return cmd
}

func stakingDelegationCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delegation [delegator-addr] [validator-addr]",
		Aliases: []string{"del"},
		Short:   "query a delegation based on a delegator address and validator address",
		Long: strings.TrimSpace(`query delegations for an individual delegator on an individual validator.

Example:
$ lens query staking delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			opts, err := queryOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}
			query := query.Query{Client: cl, Options: opts}
			delegator := args[0]
			validator := args[1]
			response, err := query.Staking_Delegation(delegator, validator)
			if err != nil {
				return err
			}
			return cl.PrintObject(response.DelegationResponse)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func stakingUnbondingDelegationCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unbonding-delegation [delegator-addr] [validator-addr]",
		Aliases: []string{"unbonding", "ubd"},
		Short:   "query an unbonding delegation based on a delegator address and validator address",
		Long: strings.TrimSpace(`query unbonding delegations for an individual delegator on an individual validator.

Example:
$ lens query staking unbonding-delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			opts, err := queryOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}
			query := query.Query{Client: cl, Options: opts}
			delegator := args[0]
			validator := args[1]
			response, err := query.Staking_UnbondingDelegation(delegator, validator)
			if err != nil {
				return err
			}
			return cl.PrintObject(response.Unbond)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func stakingUnbondingDelegationsCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unbonding-delegations [delegator-addr]",
		Aliases: []string{"unbondings", "ubds"},
		Short:   "query all unbonding delegations for a delegator address",
		Long: strings.TrimSpace(`query delegations for an individual delegator on all validators.

Example:
$ lens query staking unbonding-delegations cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			opts, err := queryOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}
			query := query.Query{Client: cl, Options: opts}
			response, err := query.Staking_DelegatorUnbondingDelegations(args[0])
			if err != nil {
				return err
			}
			return cl.PrintObject(response.UnbondingResponses)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "unbonding-delegations")
	return cmd
}

func stakingValidatorDelegationsCmd(a *appState) *cobra.Command {
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
			cl := a.Config.GetDefaultClient()
			opts, err := queryOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}
			query := query.Query{Client: cl, Options: opts}
			response, err := query.Staking_ValidatorDelegations(args[0])
			if err != nil {
				return err
			}
			return cl.PrintObject(response.DelegationResponses)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "validator-delegations")
	return cmd
}

func stakingValidatorsCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validators [status]",
		Aliases: []string{"vals", "vs"},
		Short:   "query all validators for a status",
		Long: strings.TrimSpace(`query validators.

Example:
$ lens query staking validators <bonded|unbonded|unbonding>
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			opts, err := queryOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}
			query := query.Query{Client: cl, Options: opts}
			var status string
			switch args[0] {
			case "bonded":
				status = "BOND_STATUS_BONDED"
			case "unbonded":
				status = "BOND_STATUS_UNBONDED"
			case "unbonding":
				status = "BOND_STATUS_UNBONDING"
			default:
				status = "BOND_STATUS_BONDED"
			}
			response, err := query.Staking_Validators(status)
			if err != nil {
				return err
			}
			return cl.PrintObject(response)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "validators")
	return cmd
}

func stakingValidatorCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validator [address]",
		Aliases: []string{"val", "v"},
		Short:   "query validator for an address",
		Long: strings.TrimSpace(`query validator.

Example:
$ lens query staking validator [valoper_address]
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			opts, err := queryOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}
			query := query.Query{Client: cl, Options: opts}

			response, err := query.Staking_Validator(args[0])
			if err != nil {
				return err
			}
			return cl.PrintObject(response)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
