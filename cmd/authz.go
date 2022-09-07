package cmd

import (
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/spf13/cobra"
)

func authzGrantsCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "grants [grantor] [grantee] [msg_type]?",
		Aliases: []string{"grants"},
		Short:   "query the grants for an account",
		Args:    cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: allow the user to specify the grantor client
			// and the grantee client. This will allow for use of a
			// ledger as the grantor (i.e. cosmoshub-ledger in the config)
			// and test keyringbacked for the grantee (i.e. cosmoshub)
			cl := a.Config.GetDefaultClient()
			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			orAddr, err := cl.AccountFromKeyOrAddress(args[0])
			if err != nil {
				return err
			}

			eeAddr, err := cl.AccountFromKeyOrAddress(args[1])
			if err != nil {
				return err
			}

			var MsgTypeUrl string
			if len(args) == 3 {
				// TODO: input validation for msg_type
				MsgTypeUrl = args[2]
			}
			req := &authz.QueryGrantsRequest{
				Granter:    cl.MustEncodeAccAddr(orAddr),
				Grantee:    cl.MustEncodeAccAddr(eeAddr),
				MsgTypeUrl: MsgTypeUrl,
				Pagination: pageReq,
			}

			res, err := authz.NewQueryClient(cl).Grants(cmd.Context(), req)
			if err != nil {
				return err
			}

			return cl.PrintObject(res)
		},
	}
	return paginationFlags(cmd, a.Viper)
}

// TODO: rethink UX here. We should break this up into a number
// of smaller commands. For example, we should have a grantSend,
// grantStake, grantWithdraw, grantVote, grantValidator, etc...
// authzGrantAuthorizationCmd returns the authz grant authorization command for this module
func authzGrantAuthorizationCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "grant [grantor] [grantee] [role]",
		Aliases: []string{"grant"},
		Short:   "grant an authorization",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: allow the user to specify the grantor client
			// and the grantee client. This will allow for use of a
			// ledger as the grantor (i.e. cosmoshub-ledger in the config)
			// and test keyringbacked for the grantee (i.e. cosmoshub)

			// cl := config.GetDefaultClient()
			// var authorization authz.Authorization
			// msg := &authz.MsgGrant{
			// 	Granter: "cosmos1y54exmx84cqtasvjnskf9f63djuuj68p7hqf47",
			// 	Grantee: "cosmos1y54exmx84cqtasvjnskf9f63djuuj68p7hqf47",
			// 	Grant: authz.Grant{
			// 		Authorization: authorization,
			// 		Expiration:    0,
			// 	},
			// }
			// return cl.HandleAndPrintMsgSend(cl.SendMsg(cmd.Context(), msg))
			return nil
		},
	}
	return cmd
}

// authzRevokeAuthorizationCmd returns the authz revoke authorization command for this module
func authzRevokeAuthorizationCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "revoke [grantee] [msg_type] [grantor]?",
		Aliases: []string{"r"},
		Short:   "revoke an authorization, if grantor is not provided, the grantor will be the default key",
		Args:    cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: allow the user to specify the grantor client
			// and the grantee client. This will allow for use of a
			// ledger as the grantor (i.e. cosmoshub-ledger in the config)
			// and test keyringbacked for the grantee (i.e. cosmoshub)
			cl := a.Config.GetDefaultClient()
			var key string
			if len(args) == 3 {
				key = args[2]
			}
			orAddr, err := cl.AccountFromKeyOrAddress(key)
			if err != nil {
				return err
			}
			eeAddr, err := cl.DecodeBech32AccAddr(args[0])
			if err != nil {
				return err
			}
			memo, err := cmd.Flags().GetString(flagMemo)
			if err != nil {
				return err
			}
			// TODO: query the grants to see if there is one
			// to revoke and if so, ensure that revoke msg will
			// pass (i.e. is right msg_type) before sending it.
			msg := &authz.MsgRevoke{
				Granter:    cl.MustEncodeAccAddr(orAddr),
				Grantee:    cl.MustEncodeAccAddr(eeAddr),
				MsgTypeUrl: args[1],
			}
			return cl.HandleAndPrintMsgSend(cl.SendMsg(cmd.Context(), msg, memo))
		},
	}
	memoFlag(a.Viper, cmd)
	return cmd
}

// authzExecAuthorizationCmd returns the authz exec authorization command for this module
func authzExecAuthorizationCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "exec [msg_tx_json_file] [grantee]?",
		Aliases: []string{"exec"},
		Short:   "execute an authorization",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: should we expose this to the user?
			// is there a way to check in the other transaction commands
			// if the sender is a grantee and then wrap their messages in
			// execute messages?

			// cl := config.GetDefaultClient()
			// var key string
			// if len(args) == 2 {
			// 	key = args[1]
			// }
			// sender, err := cl.AccountFromKeyOrAddress(key)
			// if err != nil {
			// 	return err
			// }
			// msg := &authz.MsgExec{
			// 	Grantee: cl.MustEncodeAccAddr(sender),
			// 	Msgs:    []authz.Msg{},
			// }
			return nil
		},
	}
	return cmd
}
