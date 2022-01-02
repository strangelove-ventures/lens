package cmd

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/spf13/cobra"
)

var (
	FlagCommission = "commission"
	FlagAll        = "all"
)

// TODO: should this be [from] [validator-address]?
// if so then we should make the first arg manditory and further args be []sdk.ValAddr
// and make the []sdk.ValAddr optional. This way we don't need any of the flags except
// commission
func distributionWithdrawRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-rewards [validator-addr] [from]",
		Short: "Withdraw rewards from a given delegation address, and optionally withdraw validator commission if the delegation address given is a validator operator",
		Long: strings.TrimSpace(
			`Withdraw rewards from a given delegation address,
and optionally withdraw validator commission if the delegation address given is a validator operator.
Example:
$ lens tx withdraw-rewards cosmosvaloper1uyccnks6gn6g62fqmahf8eafkedq6xq400rjxr default
$ lens tx withdraw-rewards cosmosvaloper1uyccnks6gn6g62fqmahf8eafkedq6xq400rjxr default --commission
$ lens tx withdraw-rewards --from mykey --all
`,
		),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			key := ""
			if len(args) == 1 {
				key = cl.Config.Key
			} else {
				key = args[1]
			}

			delAddr, err := cl.AccountFromKeyOrAddress(key)
			if err != nil {
				return err
			}

			msgs := []sdk.Msg{}

			if all, _ := cmd.Flags().GetBool(FlagAll); all {

				validators, err := cl.QueryDelegatorValidators(delAddr)
				if err != nil {
					return err
				}

				// build multi-message transaction
				for _, valAddr := range validators {
					val, err := cl.DecodeBech32ValAddr(valAddr)
					if err != nil {
						return err
					}
					msg := types.NewMsgWithdrawDelegatorReward(delAddr, sdk.ValAddress(val))
					msgs = append(msgs, msg)
				}

			} else if len(args) == 1 {
				valAddr, err := cl.DecodeBech32ValAddr(args[0])
				if err != nil {
					return err
				}
				msgs = append(msgs, types.NewMsgWithdrawDelegatorReward(delAddr, sdk.ValAddress(valAddr)))
			}

			if commission, _ := cmd.Flags().GetBool(FlagCommission); commission {
				valAddr, err := cl.DecodeBech32ValAddr(args[0])
				if err != nil {
					return err
				}
				msgs = append(msgs, types.NewMsgWithdrawValidatorCommission(sdk.ValAddress(valAddr)))
			}

			return cl.HandleAndPrintMsgSend(cl.SendMsgs(cmd.Context(), msgs))
		},
	}
	cmd.Flags().BoolP(FlagCommission, "c", false, "withdraw commission from a validator")
	cmd.Flags().BoolP(FlagAll, "a", false, "withdraw all rewards of a delegator")
	AddTxFlagsToCmd(cmd)
	return cmd
}

func distributionParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "query things about a chain's distribution params",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()

			params, err := cl.QueryDistributionParams()
			if err != nil {
				return err
			}

			return cl.PrintObject(params)
		},
	}

	return cmd
}

func distributionCommunityPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "community-pool",
		Short: "query things about a chain's community pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()

			pool, err := cl.QueryDistributionCommunityPool()
			if err != nil {
				return err
			}

			return cl.PrintObject(pool)
		},
	}

	return cmd
}

func distributionCommissionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commission [validator-address]",
		Args:  cobra.ExactArgs(1),
		Short: "query a specific validator's commission",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			address, err := cl.DecodeBech32ValAddr(args[0])
			if err != nil {
				return err
			}
			commission, err := cl.QueryDistributionCommission(address)
			if err != nil {
				return err
			}
			return cl.PrintObject(commission)
		},
	}

	return cmd
}

func distributionRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards [key-or-delegator-address] [validator-address]",
		Short: "query things about a delegator's rewards",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			delAddr, err := cl.AccountFromKeyOrAddress(args[0])
			if err != nil {
				return err
			}

			valAddr, err := cl.DecodeBech32ValAddr(args[1])
			if err != nil {
				return err
			}

			rewards, err := cl.QueryDistributionRewards(delAddr, valAddr)
			if err != nil {
				return err
			}

			return cl.PrintObject(rewards)
		},
	}

	return cmd
}

func distributionSlashesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slashes [validator-address] [start-height] [end-height]",
		Short: "query things about a validator's slashes on a chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			address, err := cl.DecodeBech32ValAddr(args[0])
			if err != nil {
				return err
			}

			startHeight, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			endHeight, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			slashes, err := cl.QueryDistributionSlashes(address, startHeight, endHeight, pageReq)
			if err != nil {
				return err
			}

			return cl.PrintObject(slashes)
		},
	}

	return paginationFlags(cmd)
}

func distributionValidatorRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-outstanding-rewards [address]",
		Short: "query things about a validator's (and all their delegators) outstanding rewards on a chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()

			address, err := cl.DecodeBech32ValAddr(args[0])
			if err != nil {
				return err
			}

			rewards, err := cl.QueryDistributionValidatorRewards(address)
			if err != nil {
				return err
			}

			return cl.PrintObject(rewards)
		},
	}
	return cmd
}
