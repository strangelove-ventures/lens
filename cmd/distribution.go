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

func distributionWithdrawRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-rewards [validator-addr] [from]",
		Short: "Withdraw rewards from a given delegation address, and optionally withdraw validator commission if the delegation address given is a validator operator",
		Long: strings.TrimSpace(
			`Withdraw rewards from a given delegation address,
and optionally withdraw validator commission if the delegation address given is a validator operator.
Example:
$ lens tx withdraw-rewards cosmosvaloper1uyccnks6gn6g62fqmahf8eafkedq6xq400rjxr mykey
$ lens tx withdraw-rewards cosmosvaloper1uyccnks6gn6g62fqmahf8eafkedq6xq400rjxr mykey commission
$ lens tx withdraw-rewards mykey all
`,
		),
		Args: cobra.MaximumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				delAddr sdk.AccAddress
				err     error
			)

			cl := config.GetDefaultClient()

			msgs := []sdk.Msg{}

			if args[1] == "all" {
				if args[0] != cl.Config.Key {
					cl.Config.Key = args[0]
				}

				if cl.KeyExists(args[0]) {
					delAddr, err = cl.GetDefaultAddress()
				} else {
					delAddr, err = cl.DecodeBech32AccAddr(args[1])
				}
				if err != nil {
					return err
				}

				validators, err := cl.QueryDelegatorValidators(delAddr)
				if err != nil {
					return err
				}

				// build multi-message transaction
				msgs := make([]sdk.Msg, 0, len(validators))
				for _, valAddr := range validators {
					val, err := cl.DecodeBech32ValAddr(valAddr)
					if err != nil {
						return err
					}

					msg := types.NewMsgWithdrawDelegatorReward(delAddr, sdk.ValAddress(val))
					msgs = append(msgs, msg)
				}
			} else {
				if args[1] != cl.Config.Key {
					cl.Config.Key = args[1]
				}

				if cl.KeyExists(args[1]) {
					delAddr, err = cl.GetDefaultAddress()
				} else {
					delAddr, err = cl.DecodeBech32AccAddr(args[1])
				}
				if err != nil {
					return err
				}

				valAddr, err := cl.DecodeBech32ValAddr(args[0])
				if err != nil {
					return err
				}

				msgs = append(msgs, types.NewMsgWithdrawDelegatorReward(delAddr, sdk.ValAddress(valAddr)))
			}

			if len(args) == 3 {
				if args[2] == "commission" {
					valAddr, err := cl.DecodeBech32ValAddr(args[0])
					if err != nil {
						return err
					}
					msgs = append(msgs, types.NewMsgWithdrawValidatorCommission(sdk.ValAddress(valAddr)))
				}
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

	return cmd
}
