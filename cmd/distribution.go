package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/spf13/cobra"
)

// FlagCommission to withdraw validator's commission.
var (
	FlagCommission = "commission"
	FlagAll        = "all"
)

func distributionWithdrawRewardsCmd() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "withdraw-rewards [validator-addr] [mykey]",
		Short: "Withdraw rewards from a given delegation address, and optionally withdraw validator commission if the delegation address given is a validator operator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw rewards from a given delegation address,
and optionally withdraw validator commission if the delegation address given is a validator operator.
Example:
$ %s tx distribution withdraw-rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj mykey
$ %s tx distribution withdraw-rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj mykey --commission
`,
				version.AppName, bech32PrefixValAddr, version.AppName, bech32PrefixValAddr,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				delAddr sdk.AccAddress
				err     error
			)

			cl := config.GetDefaultClient()

			if args[2] != cl.Config.Key {
				cl.Config.Key = args[2]
			}

			if cl.KeyExists(args[2]) {
				delAddr, err = cl.GetKeyByName(args[2])
			} else {
				delAddr, err = cl.DecodeBech32AccAddr(args[2])
			}
			if err != nil {
				return err
			}

			valAddr, err := cl.DecodeBech32ValAddr(args[0])
			if err != nil {
				return err
			}

			valperAddr := sdk.ValAddress(valAddr)

			msgs := []sdk.Msg{}
			if All, _ := cmd.Flags().GetBool(FlagAll); All {

				// delValsRes, err := queryClient.DelegatorValidators(cmd.Context(), &types.QueryDelegatorValidatorsRequest{DelegatorAddress: delAddr.String()})
				// if err != nil {
				// 	return err
				// }

				// validators := delValsRes.Validators
				// // build multi-message transaction
				// msgs := make([]sdk.Msg, 0, len(validators))
				// for _, valAddr := range validators {
				// 	val, err := sdk.ValAddressFromBech32(valAddr)
				// 	if err != nil {
				// 		return err
				// 	}

				// 	msg := types.NewMsgWithdrawDelegatorReward(delAddr, val)
				// 	msgs = append(msgs, msg)
				// }
			} else {
				msgs = append(msgs, types.NewMsgWithdrawDelegatorReward(delAddr, valperAddr))
			}

			if commission, _ := cmd.Flags().GetBool(FlagCommission); commission {
				msgs = append(msgs, types.NewMsgWithdrawValidatorCommission(valperAddr))
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

	return cmd
}
