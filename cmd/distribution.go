package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/spf13/cobra"
)

// FlagCommission to withdraw validator's commission.
var (
	FlagCommission = "commission"
	FlagAll        = "all"
	FlagFrom       = "from"
)

func distributionWithdrawRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-rewards [validator-addr]",
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
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				delAddr sdk.AccAddress
				err     error
			)

			cl := config.GetDefaultClient()

			key, _ := cmd.Flags().GetString(FlagFrom)
			if key != "" {
				return errors.New("From key unset")
			}

			if key != cl.Config.Key {
				cl.Config.Key = key
			}

			if cl.KeyExists(key) {
				delAddr, err = cl.GetKeyByName(key)
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
				msgs := make([]sdk.Msg, 0, len(validators))
				for _, valAddr := range validators {
					val, err := sdk.ValAddressFromBech32(valAddr)
					if err != nil {
						return err
					}

					msg := types.NewMsgWithdrawDelegatorReward(delAddr, val)
					msgs = append(msgs, msg)
				}
			} else {
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

	cmd.Flags().Bool(FlagCommission, false, "Withdraw the validator's commission in addition to the rewards")
	cmd.Flags().Bool(FlagAll, false, "Withdraw all your accounts rewards")
	cmd.Flags().String(FlagFrom, "", "Withdraw rewards from the given address")

	return cmd
}
