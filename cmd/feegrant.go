package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	query "github.com/strangelove-ventures/lens/client/query"
)

// feegrantConfigureCmd returns the fee grant configuration commands for this module
func feegrantConfigureBaseCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feegrant",
		Short: "Configure the client to use round-robin feegranted accounts when sending TXs",
		Long:  "Use round-robin feegranted accounts when sending TXs. Useful for relayers and applications where sequencing is important",
	}

	cmd.AddCommand(
		feegrantConfigureBasicCmd(a),
	)

	return cmd
}

func feegrantConfigureBasicCmd(a *appState) *cobra.Command {
	var numGrantees int
	cmd := &cobra.Command{
		Use:   "basicallowance [chain-name] [granter] --grantees [int]",
		Short: "feegrants for the given chain (if none specified, uses the default account)",
		Long:  "feegrants for the given chain. 10 grantees by default, all with an unrestricted BasicAllowance. Fails if already configured.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chain := args[0]
			cl := a.Config.GetClient(chain)

			if cl.Config.FeeGrants != nil {
				return fmt.Errorf("feegrants are already configured for chain '%s'", chain)
			}

			granterKeyOrAddr := ""
			if len(args) > 1 {
				granterKeyOrAddr = args[1]
			} else {
				granterKeyOrAddr = cl.Config.Key
			}

			granterKey, err := cl.KeyFromKeyOrAddress(granterKeyOrAddr)
			if err != nil {
				return fmt.Errorf("could not get granter key from '%s'", granterKeyOrAddr)
			}
			feegrantErr := cl.ConfigureFeegrants(numGrantees, granterKey)
			if feegrantErr != nil {
				return feegrantErr
			}

			return nil
			//return cl.PrintObject(res)
		},
	}
	cmd.Flags().IntVar(&numGrantees, "grantees", 10, "number of grantees that will be feegranted with basic allowances")
	return cmd
}

func feegrantBasicGrantsCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "basic [granter]",
		Short: "query the grants for an account (if none is specified, the default account is returned)",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := a.Config.GetDefaultClient()
			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			height, err := ReadHeight(cmd.Flags())
			if err != nil {
				return err
			}

			keyNameOrAddress := ""
			if len(args) == 0 {
				keyNameOrAddress = cl.Config.Key
			} else {
				keyNameOrAddress = args[0]
			}

			granterAcc, err := cl.AccountFromKeyOrAddress(keyNameOrAddress)
			if err != nil {
				fmt.Printf("Error retrieving account from key '%s'\n", keyNameOrAddress)
				return err
			}
			granterAddr := cl.MustEncodeAccAddr(granterAcc)
			options := query.QueryOptions{Pagination: pageReq, Height: height}
			q := &query.Query{Client: cl, Options: &options}

			res, err := query.Feegrant_GrantsRPC(q, granterAddr)
			if err != nil {
				return err
			}

			return cl.PrintObject(res)
		},
	}
	return paginationFlags(cmd, a.Viper)
}
