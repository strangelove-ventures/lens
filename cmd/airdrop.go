package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
)

func airdropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "airdrop [airdrop.json] [denom] [key]?",
		Short: "Airdrop coins to a specified address",
		Long:  "The airdrop file consists of map[string]float64 where the key is the address on the target chain and the value is the amount of coins to be airdropped to that address/1e6 (i.e. atom instead of uatom). The airdrop command 1. checks the addresses in the file to ensure that they are valid for the given chain l",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			// keyNameOrAddress := ""
			// if len(args) == 1 {
			// 	keyNameOrAddress = cl.Config.Key
			// } else {
			// 	keyNameOrAddress = args[1]
			// }
			// address, err := cl.AccountFromKeyOrAddress(keyNameOrAddress)
			// if err != nil {
			// 	return err
			// }

			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			bz, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}
			var airdrop airdropFile
			if err := json.Unmarshal(bz, &airdrop); err != nil {
				return err
			}

			sum := 0.0
			for k, v := range airdrop {
				_, err := cl.DecodeBech32AccAddr(k)
				if err != nil {
					return err
				}
				sum += v * 1e6
				// fmt.Println(k, fmt.Sprintf("%fusomm", v*1000000))
			}
			fmt.Printf("TOTAL: %f\n", sum)
			return nil
		},
	}
	cmd.Flags().Int("max-msgs", 20, "max number of msgs per tx to send")
	return cmd
}

type airdropFile map[string]float64
