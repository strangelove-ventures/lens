package cmd

import (
	"github.com/spf13/cobra"
)

// queryCmd represents the keys command
func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "query things about a chain",
	}

	cmd.AddCommand(
		authQueryCmd(),
		bankQueryCmd(),
		distributionQueryCmd(),
		governanceQueryCmd(),
	)

	return cmd
}

// authQueryCmd returns the transaction commands for this module
func authQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth",
		Aliases: []string{"a"},
		Short:   "Querying commands for the auth module",
	}

	cmd.AddCommand(
		getAccountCmd(),
		getAccountsCmd(),
		getParamsCmd(),
	)

	return cmd
}

// bankQueryCmd  returns the transaction commands for this module
func bankQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bank",
		Aliases: []string{"b"},
		Short:   "Querying commands for the auth module",
	}

	cmd.AddCommand(
		getBalanceCmd(),
	)

	return cmd
}

func distributionQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "distribution",
		Aliases: []string{"dist", "distr", "d"},
		Short:   "Query things about a chain's distribution module",
	}

	cmd.AddCommand(
		getDistributionCommissionCmd(),
		getDistributionCommunityPoolCmd(),
		getDistributionParamsCmd(),
		getDistributionRewardsCmd(),
		getDistributionSlashesCmd(),
		getDistributionValidatorRewardsCmd(),
	)

	return cmd
}

func governanceQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "governance",
		Aliases: []string{"gov", "govern", "g"},
		Short:   "Query things about a chain's governance module",
	}

	cmd.AddCommand(
		getGovernanceProposalsCmd(),
	)

	return cmd
}
