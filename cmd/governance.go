package cmd

import (
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/spf13/cobra"
)

func getGovernanceProposalsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposals",
		Short: "query things about a chain's proposals",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			proposalsResponse, err := cl.QueryGovernanceProposals(types.StatusVotingPeriod, "", "", pageReq)
			if err != nil {
				return err
			}
			return cl.PrintObject(proposalsResponse)
		},
	}

	return paginationFlags(cmd)
}
