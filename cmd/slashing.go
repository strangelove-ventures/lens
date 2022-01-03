package cmd

import (
	"strings"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func slashingParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current slashing parameters",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(`Query genesis parameters for the slashing module:

$ <appd> query slashing params
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := config.GetDefaultClient()
			result, err := client.QuerySlashingParams()
			if err != nil {
				return err
			}
			return client.PrintObject(result)
		},
	}
	return cmd
}

func slashingSigningInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signing-info [validator-conspub]",
		Short: "Query a validator's signing information",
		Long: strings.TrimSpace(`Use a validators' consensus public key to find the signing-info for that validator:

$ <appd> query slashing signing-info '{"@type":"/cosmos.crypto.ed25519.PubKey","key":"OauFcTKbN5Lx3fJL689cikXBqe+hcp6Y+x0rYUdR9Jk="}'
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := config.GetDefaultClient()
			pageReq, err := sdkclient.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			result, err := client.QuerySlashingSigningInfo(args[0], pageReq)
			if err != nil {
				return err
			}

			return client.PrintObject(result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func slashingSigningInfosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signing-infos",
		Short: "Query signing information of all validators",
		Long: strings.TrimSpace(`signing infos of validators:

$ <appd> query slashing signing-infos
`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := config.GetDefaultClient()
			pageReq, err := sdkclient.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			result, err := client.QuerySlashingSigningInfos(pageReq)
			if err != nil {
				return err
			}

			return client.PrintObject(result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "signing infos")

	return cmd
}
