package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/avast/retry-go/v4"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"
)

func airdropCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "airdrop [airdrop.json] [denom] [exclude] [key]?",
		Short: "Airdrop coins to a specified address",
		Long:  "The airdrop file consists of map[string]float64 where the key is the address on the target chain and the value is the amount of coins to be airdropped to that address/1e6 (i.e. atom instead of uatom). The airdrop command 1. checks the addresses in the file to ensure that they are valid for the given chain l",
		Args:  cobra.RangeArgs(3, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			keyNameOrAddress := ""
			if len(args) == 3 {
				keyNameOrAddress = cl.Config.Key
			} else {
				keyNameOrAddress = args[3]
			}
			address, err := cl.AccountFromKeyOrAddress(keyNameOrAddress)
			if err != nil {
				return err
			}

			bz, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			var airdrop airdropFile
			if err := json.Unmarshal(bz, &airdrop); err != nil {
				return err
			}

			denom := args[1]

			exclude, err := os.Open(args[2])
			if err != nil {
				return err
			}
			defer exclude.Close()

			// scanner := bufio.NewScanner(exclude)

			// for scanner.Scan() {
			// 	delete(airdrop, scanner.Text())
			// }

			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				return err
			}
			if dryRun {
				var dropTotal float64
				var dropAddress int64
				for _, v := range airdrop {

					_, err := cl.DecodeBech32AccAddr(v.Address)
					if err != nil {
						return err
					}
					dropTotal += float64(v.Amount)
					dropAddress++

				}
				fmt.Fprintf(cmd.OutOrStdout(), "Airdrop total: %f %s\n", dropTotal, denom)
				fmt.Fprintf(cmd.OutOrStdout(), "Airdrop address count: %d\n", dropAddress)
				return nil
			}

			maxSends, err := cmd.Flags().GetInt("max-sends")
			if err != nil {
				return err
			}

			multiMsg := &banktypes.MsgMultiSend{
				Inputs:  []banktypes.Input{},
				Outputs: []banktypes.Output{},
			}
			amount := sdk.Coin{Denom: denom, Amount: sdk.NewInt(0)}
			var sent int
			for _, v := range airdrop {
				to, err := cl.DecodeBech32AccAddr(v.Address)
				if err != nil {
					return err
				}
				toSendCoin := sdk.NewCoin(denom, sdk.NewInt(int64(v.Amount)))
				toSend := sdk.NewCoins(toSendCoin)
				amount = amount.Add(toSendCoin)
				multiMsg.Outputs = append(multiMsg.Outputs, banktypes.Output{cl.MustEncodeAccAddr(to), toSend})
				sent += 1

				fmt.Println(v.Address)

				if len(multiMsg.Outputs) > maxSends-1 {
					completion := float64(sent) / float64(len(airdrop))
					fmt.Fprintf(cmd.OutOrStdout(), "(%f) sending %s to %d addresses\n", completion, amount.String(), len(multiMsg.Outputs))
					multiMsg.Inputs = append(multiMsg.Inputs, banktypes.Input{cl.MustEncodeAccAddr(address), sdk.NewCoins(amount)})
					retry.Do(func() error {
						fmt.Fprintf(cmd.OutOrStdout(), "sending tx\n")
						res, err := cl.SendMsgs(cmd.Context(), []sdk.Msg{multiMsg})
						if err != nil || res.Code != 0 {
							if err != nil {
								fmt.Fprintf(cmd.OutOrStdout(), "failed to send airdrop: %s\n", err)
								fmt.Println(err.Error())
								return err
							}
							fmt.Fprintf(cmd.OutOrStdout(), "failed to send airdrop: %s\n", res.RawLog)
							err = fmt.Errorf("failed to send airdrop")
							time.Sleep(time.Second * 10)
							return err
						}
						return nil
					}, retry.Context(cmd.Context()))
					fmt.Println("send successful")
					multiMsg.Inputs = []banktypes.Input{}
					multiMsg.Outputs = []banktypes.Output{}
					amount = sdk.Coin{Denom: denom, Amount: sdk.NewInt(0)}
				}
			}

			fmt.Println("remaining sends")

			fmt.Println(multiMsg)
			return nil
		},
	}
	cmd.Flags().Int("max-sends", 200, "max number of msgs per tx to send")
	cmd.Flags().Bool("dry-run", false, "read the aidrop file and print metrics")

	return cmd
}

// type airdropFile []map[string]float64
type airdropFile []airdropItem

type airdropItem struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
}
