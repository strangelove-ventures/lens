package cmd

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/lens/client"
)

// crosschainCmd represents the command to get balances across chains
func crosschainCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "crosschain",
		Aliases: []string{"cc", "kriskross", "cchain", "coolchain"},
		Short:   "query about balances across chains",
	}

	cmd.AddCommand(
		crosschainBankQueryCmd(a),
	)

	cmd.PersistentFlags().Bool("combined", false, "combine balances from all chains")

	return cmd
}

// crosschainBankQueryCmd returns the transaction commands for this module
func crosschainBankQueryCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bank",
		Aliases: []string{"b"},
		Short:   "Querying commands for the auth module",
	}

	cmd.AddCommand(
		getEnabledChainbalancesCmd(a),
	)

	return cmd
}

func getEnabledChainbalancesCmd(a *appState) *cobra.Command {
	return &cobra.Command{
		Use:   "balances [key-or-address]",
		Args:  cobra.RangeArgs(0, 1),
		Short: "get balances across all configured chains",
		Long:  "if no key or addresss is passed, the key on default chain is used to enumerate through all configured chains",
		RunE: func(cmd *cobra.Command, args []string) error {
			combineBalances, err := cmd.Flags().GetBool("combined")
			if err != nil {
				return err
			}
			enabledChains := []string{}
			for chain := range a.Config.Chains {
				enabledChains = append(enabledChains, chain)
			}
			// alphabetically sort the chains - this is to make the output more readable/consistent
			sort.StringSlice(enabledChains).Sort()

			// copied from bank.go
			cl := a.Config.GetDefaultClient()
			var (
				keyNameOrAddress = ""
				address          sdk.AccAddress
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
			denomBalanceMap := make(map[string]sdk.Coins)
			// end: copied from bank.go
			for _, chain := range enabledChains {
				cl := a.Config.GetClient(chain)
				balance, err := cl.QueryBalanceWithDenomTraces(cmd.Context(), address, client.DefaultPageRequest())
				if err != nil {
					return err
				}
				denomBalanceMap[chain] = balance
			}
			if combineBalances {
				combinedBalanceMap := make(map[string]sdk.Int)
				for _, coins := range denomBalanceMap {
					for _, coin := range coins {
						denom := coin.Denom
						if strings.HasPrefix(denom, "transfer/") {
							items := strings.Split(denom, "/")
							denom = items[len(items)-1]
						}
						if _, ok := combinedBalanceMap[denom]; !ok {
							combinedBalanceMap[denom] = sdk.ZeroInt()
						}
						combinedBalanceMap[denom] = combinedBalanceMap[denom].Add(coin.Amount)
					}
				}
				for denom, balance := range combinedBalanceMap {
					fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", denom, balance.String())
				}
			} else {
				for _, chain := range enabledChains {
					cl := a.Config.GetClient(chain)
					chainAddress, err := cl.EncodeBech32AccAddr(address)
					if err != nil {
						return err
					}
					fmt.Fprintln(cmd.OutOrStdout(), "==============================================================")
					fmt.Fprintf(cmd.OutOrStdout(), "Chain: %s, Address: %s\n", chain, chainAddress)
					for _, balance := range denomBalanceMap[chain] {
						fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", balance.Denom, balance.Amount.String())
					}
				}
			}
			return nil
		},
	}
}
