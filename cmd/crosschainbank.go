package cmd

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

// crosschainBankCmd represents the command to get balances across chains
func crosschainBankCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "crosschainbank",
		Aliases: []string{"cb"},
		Short:   "query about balances across chains",
	}

	cmd.AddCommand(
		getEnabledChainbalancesCmd(),
	)

	return cmd
}

func getEnabledChainbalancesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "balances",
		Args:  cobra.MinimumNArgs(1),
		Short: "get balances across chains",
		RunE: func(cmd *cobra.Command, args []string) error {
			enabledChains := []string{}
			for chain := range config.Chains {
				enabledChains = append(enabledChains, chain)
			}
			// alphabetically sort the chains - this is to make the output more readable/consistent
			sort.StringSlice(enabledChains).Sort()

			// copied from bank.go
			cl := config.GetDefaultClient()
			var (
				keyNameOrAddress = ""
				address          sdk.AccAddress
				err              error
			)
			if len(args) == 0 {
				keyNameOrAddress = cl.Config.Key
			} else {
				keyNameOrAddress = args[0]
			}
			if cl.KeyExists(keyNameOrAddress) {
				cl.Config.Key = keyNameOrAddress
				address, err = cl.GetKeyAddress()
			} else {
				address, err = cl.DecodeBech32AccAddr(keyNameOrAddress)
			}
			if err != nil {
				return err
			}
			denomBalanceMap := make(map[string]sdk.Int)
			// end: copied from bank.go
			for _, chain := range enabledChains {
				// fmt.Printf("%s: %s\n", chain, config.Chains[chain].ChainID)
				cl := config.GetClient(chain)
				chainAddress, err := cl.EncodeBech32AccAddr(address)
				if err != nil {
					return err
				}
				balance, err := cl.QueryBalance(address, false)
				if err != nil {
					return err
				}
				fmt.Printf("Chain: %s\n", chain)
				fmt.Printf("Address: %s\n", chainAddress)
				for _, coin := range balance {
					denom := coin.Denom
					if strings.HasPrefix(denom, "transfer/") {
						items := strings.Split(denom, "/")
						denom = items[len(items)-1]
					}
					if _, ok := denomBalanceMap[denom]; !ok {
						denomBalanceMap[denom] = sdk.ZeroInt()
					}
					denomBalanceMap[denom] = denomBalanceMap[denom].Add(coin.Amount)
				}
				cl.PrintObject(balance)

			}
			for denom, balance := range denomBalanceMap {
				fmt.Printf("%s: %s\n", denom, balance.String())
			}
			return nil
		},
	}
}
