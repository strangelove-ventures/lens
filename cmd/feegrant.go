package cmd

import (
	"github.com/spf13/cobra"
	query "github.com/strangelove-ventures/lens/client/query"
)

func feegrantGrantsCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "grants [grantor] [grantee] [msg_type]?",
		Aliases: []string{"grants"},
		Short:   "query the grants for an account",
		Args:    cobra.RangeArgs(2, 3),
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

			granterAcc, err := cl.AccountFromKeyOrAddress(args[0])
			if err != nil {
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
