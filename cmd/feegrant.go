package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	query "github.com/strangelove-ventures/lens/client/query"
	tx "github.com/strangelove-ventures/lens/client/tx"
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
	var update bool
	cmd := &cobra.Command{
		Use:   "basicallowance [chain-name] [granter] --grantees [int] --update-granter",
		Short: "feegrants for the given chain and granter (if granter is unspecified, use the default key)",
		Long:  "feegrants for the given chain. 10 grantees by default, all with an unrestricted BasicAllowance.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chain := args[0]
			cl := a.Config.GetClient(chain)
			granterKeyOrAddr := ""

			if len(args) > 1 {
				granterKeyOrAddr = args[1]
			} else if cl.Config.FeeGrants != nil {
				granterKeyOrAddr = cl.Config.FeeGrants.GranterKey
			} else {
				granterKeyOrAddr = cl.Config.Key
			}

			granterKey, err := cl.KeyFromKeyOrAddress(granterKeyOrAddr)
			if err != nil {
				return fmt.Errorf("could not get granter key from '%s'", granterKeyOrAddr)
			}

			if cl.Config.FeeGrants != nil && granterKey != cl.Config.FeeGrants.GranterKey && !update {
				return fmt.Errorf("you specified granter '%s' which is different than configured feegranter '%s', but you did not specify the --update flag", granterKeyOrAddr, cl.Config.FeeGrants.GranterKey)
			} else if cl.Config.FeeGrants != nil && granterKey != cl.Config.FeeGrants.GranterKey && update {
				cl.Config.FeeGrants.GranterKey = granterKey
				cfgErr := a.OverwriteConfig(a.Config)
				cobra.CheckErr(cfgErr)
			}

			if cl.Config.FeeGrants == nil {
				feegrantErr := cl.ConfigureFeegrants(numGrantees, granterKey)
				if feegrantErr != nil {
					return feegrantErr
				}

				cfgErr := a.OverwriteConfig(a.Config)
				cobra.CheckErr(cfgErr)
			}

			memo, err := cmd.Flags().GetString(flagMemo)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			_, err = tx.EnsureBasicGrants(ctx, memo, cl)
			cobra.CheckErr(err)

			//Get latest height from the chain, mark feegrant configuration as verified up to that height.
			//This means we've verified feegranting is enabled on-chain and TXs can be sent with a feegranter.
			if cl.Config.FeeGrants != nil {
				h, err := cl.QueryLatestHeight(ctx)
				cobra.CheckErr(err)
				cl.Config.FeeGrants.BlockHeightVerified = h
				cfgErr := a.OverwriteConfig(a.Config)
				cobra.CheckErr(cfgErr)
			}

			return nil
		},
	}
	cmd.Flags().BoolVar(&update, "update-granter", false, "if a granter is configured and you want to change the granter key")
	cmd.Flags().IntVar(&numGrantees, "grantees", 10, "number of grantees that will be feegranted with basic allowances")
	memoFlag(a.Viper, cmd)
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

			res, err := query.Feegrant_GrantsByGranterRPC(q, granterAddr)
			if err != nil {
				return err
			}

			for _, grant := range res {
				allowance, e := cl.MarshalProto(grant.Allowance)
				cobra.CheckErr(e)
				fmt.Printf("Granter: %s, Grantee: %s, Allowance: %s\n", grant.Granter, grant.Grantee, allowance)
			}

			return nil
		},
	}
	return paginationFlags(cmd, a.Viper)
}
