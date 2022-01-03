package cmd

import (
	"strings"

	govClient "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/spf13/cobra"
)

var (
	flagDepositor = "depostior"
	flagVoter     = "voter"
	flagStatus    = "status"
)

func getGovernanceProposalsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposals",
		Short: "query things about a chain's proposals",
		Long: strings.TrimSpace(
			`Query for paginated proposals that match optional filters:
Example:
$ lens query gov proposals --depositor cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
$ lens query gov proposals --voter cosmos1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk
$ lens query gov proposals --status (DepositPeriod|VotingPeriod|Passed|Rejected)
$ lens query gov proposals --page=2 --limit=100`,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()

			bechDepositorAddr, _ := cmd.Flags().GetString(flagDepositor)
			bechVoterAddr, _ := cmd.Flags().GetString(flagVoter)
			strProposalStatus, _ := cmd.Flags().GetString(flagStatus)

			var proposalStatus types.ProposalStatus

			if len(bechDepositorAddr) != 0 {
				_, err := cl.DecodeBech32ValAddr(bechDepositorAddr)
				if err != nil {
					return err
				}
			}

			if len(bechVoterAddr) != 0 {
				_, err := cl.DecodeBech32ValAddr(bechVoterAddr)
				if err != nil {
					return err
				}
			}

			if len(strProposalStatus) != 0 {
				proposalStatus1, err := types.ProposalStatusFromString(govClient.NormalizeProposalStatus(strProposalStatus))
				proposalStatus = proposalStatus1
				if err != nil {
					return err
				}
			}

			pageReq, err := ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			proposalsResponse, err := cl.QueryGovernanceProposals(proposalStatus, bechDepositorAddr, bechVoterAddr, pageReq)
			if err != nil {
				return err
			}
			return cl.PrintObject(proposalsResponse)
		},
	}
	cmd.Flags().String(flagDepositor, "", "(optional) filter by proposals deposited on by depositor")
	cmd.Flags().String(flagVoter, "", "(optional) filter by proposals voted on by voted")
	cmd.Flags().String(flagStatus, "", "(optional) filter proposals by proposal status, status: deposit_period/voting_period/passed/rejected")
	return paginationFlags(cmd)
}
