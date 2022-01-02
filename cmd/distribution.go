package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/spf13/cobra"
)

var (
	FlagCommission = "commission"
	FlagAll        = "all"
)

func distributionWithdrawRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-rewards [validator-addr] [from]",
		Short: "Withdraw rewards from a given delegation address, and optionally withdraw validator commission if the delegation address given is a validator operator",
		Long: strings.TrimSpace(
			`Withdraw rewards from a given delegation address,
and optionally withdraw validator commission if the delegation address given is a validator operator.
Example:
$ lens tx withdraw-rewards cosmosvaloper1uyccnks6gn6g62fqmahf8eafkedq6xq400rjxr --from mykey
$ lens tx withdraw-rewards cosmosvaloper1uyccnks6gn6g62fqmahf8eafkedq6xq400rjxr --from mykey --commission
$ lens tx withdraw-rewards --from mykey --all
`,
		),
		Args: cobra.MaximumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				delAddr sdk.AccAddress
				err     error
			)

			cl := config.GetDefaultClient()

			key, _ := cmd.Flags().GetString(FlagFrom)
			if key != "" {
				if key != cl.Config.Key {
					cl.Config.Key = key
				}
			}

			if cl.KeyExists(key) {
				delAddr, err = cl.GetDefaultAddress()
			} else {
				delAddr, err = cl.DecodeBech32AccAddr(key)
			}
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

			res, ok, err := cl.SendMsgs(cmd.Context(), msgs)
			if err != nil || !ok {
				if res != nil {
					return fmt.Errorf("failed to withdraw rewards: code(%d) msg(%s)", res.Code, res.Logs)
				}
				return fmt.Errorf("failed to withdraw rewards: err(%w)", err)
			}

			bz, err := cl.Codec.Marshaler.MarshalJSON(res)
			if err != nil {
				return err
			}

			var out = bytes.NewBuffer([]byte{})
			if err := json.Indent(out, bz, "", "  "); err != nil {
				return err
			}
			fmt.Println(out.String())
      			return nil
		},
	}
  cmd.Flags().Bool(FlagCommission, false, "Withdraw commission from a validator")
	cmd.Flags().Bool(FlagAll, false, "Withdraw All rewards of a delegator")
	AddTxFlagsToCmd(cmd)
	return cmd
}

func getDistributionQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "distribution",
		Short: "Query things about a chain's distribution module",
	}

	cmd.AddCommand(
		queryDistributionCommissionCmd(),
		queryDistributionCommunityPoolCmd(),
		queryDistributionParamsCmd(),
		queryDistributionRewardsCmd(),
		queryDistributionSlashesCmd(),
		queryDistributionValidatorRewardsCmd(),
	)

	return cmd
}

func queryDistributionParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "query things about a chain's distribution params",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			params, err := cl.QueryDistributionParams()
			if err != nil {
				return err
			}
			cl.PrintObject(params)

			return nil
		},
	}

	return cmd
}

func queryDistributionCommunityPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "community-pool",
		Short: "query things about a chain's community pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			pool, err := cl.QueryDistributionCommunityPool()
			if err != nil {
				return err
			}
			cl.PrintObject(pool)

			return nil
		},
	}

	return cmd
}

func queryDistributionCommissionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commission [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "query a specific validator's commission",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			address := args[0]
			commission, err := cl.QueryDistributionCommission(address)

			if err != nil {
				return err
			}

			cl.PrintObject(commission)
			return nil
		},
	}

	return cmd
}

func queryDistributionRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards",
		Short: "query things about a delegator's rewards",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			delegatorAddress := args[0]
			validatorAddress := args[1]

			pool, err := cl.QueryDistributionRewards(delegatorAddress, validatorAddress)
			if err != nil {
				return err
			}
			cl.PrintObject(pool)

			return nil
		},
	}

	return cmd
}

func queryDistributionSlashesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slashes",
		Short: "query things about a validator's slashes on a chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			address := args[0]

			slashes, err := cl.QueryDistributionSlashes(address)
			if err != nil {
				return err
			}
			cl.PrintObject(slashes)

			return nil
		},
	}

	return cmd
}

func queryDistributionValidatorRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-outstanding-rewards [address]",
		Short: "query things about a validator's (and all their delegators) outstanding rewards on a chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			address := args[0]

			rewards, err := cl.QueryDistributionValidatorRewards(address)
			if err != nil {
				return err
			}
			cl.PrintObject(rewards)

			return nil
		},
	}