package cmd

import (
	"github.com/spf13/cobra"
)

// queryCmd represents the query command tree.
func queryCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "query things about a chain",
	}

	cmd.AddCommand(
		authQueryCmd(a),
		authzQueryCmd(a),
		bankQueryCmd(a),
		distributionQueryCmd(a),
		stakingQueryCmd(a),
	)

	if false {
		// TODO: enable these when commands are available
		cmd.AddCommand(
			feegrantQueryCmd(),
			govQueryCmd(),
			slashingQueryCmd(),
		)
	}

	return cmd
}

// authQueryCmd returns the transaction commands for this module
func authQueryCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth",
		Aliases: []string{"a"},
		Short:   "Querying commands for the auth module",
	}

	cmd.AddCommand(
		authAccountCmd(a),
		authAccountsCmd(a),
		authParamsCmd(a),
	)

	return cmd
}

// authzQueryCmd returns the authz query commands for this module
func authzQueryCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "authz",
		Aliases: []string{"authz"},
		Short:   "Querying commands for the authz module",
	}

	cmd.AddCommand(
		authzGrantsCmd(a),
	)

	return cmd
}

// bankQueryCmd  returns the transaction commands for this module
func bankQueryCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bank",
		Aliases: []string{"b"},
		Short:   "Querying commands for the bank module",
	}

	cmd.AddCommand(
		bankBalanceCmd(a),
		bankTotalSupplyCmd(a),
		bankDenomsMetadataCmd(a),
	)

	return cmd
}

// distributionQueryCmd returns the distribution query commands for this module
func distributionQueryCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "distribution",
		Aliases: []string{"dist", "distr", "d"},
		Short:   "Querying commands for the distribution module",
	}

	cmd.AddCommand(
		distributionParamsCmd(a),
		distributionValidatorRewardsCmd(a),
		distributionCommissionCmd(a),
		distributionCommunityPoolCmd(a),
		distributionRewardsCmd(a),
		distributionSlashesCmd(a),
		distributionDelegatorValidatorsCmd(a),
	)

	return cmd
}

// feegrantQueryCmd returns the fee grant query commands for this module
func feegrantQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "feegrant",
		Aliases: []string{"feegrant"},
		Short:   "Querying commands for the feegrant module",
	}

	cmd.AddCommand(
	// feegrantGrantsCmd(),
	// feegrantFeeGrantsCmd(),
	)

	return cmd
}

// govQueryCmd returns the gov query commands for this module
func govQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "governance",
		Aliases: []string{"gov", "g"},
		Short:   "Querying commands for the gov module",
	}

	cmd.AddCommand(
	// govProposalCmd(),
	// govProposalsCmd(),
	// govVoteCmd(),
	// govVotesCmd(),
	// govParamCmd(),
	// govParamsCmd(),
	// govProposerCmd(),
	// govDepositCmd(),
	// govDepositsCmd(),
	// govTallyCmd(),
	)

	return cmd
}

// slashingQueryCmd returns the slashing query commands for this module
func slashingQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "slashing",
		Aliases: []string{"sl", "slash"},
		Short:   "Querying commands for the slashing module",
	}

	cmd.AddCommand(
	// slashingSigningInfoCmd(),
	// slashingParamsCmd(),
	// slashingSigningInfosCmd(),
	)

	return cmd
}

// stakingQueryCmd returns the staking query commands for this module
func stakingQueryCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "staking",
		Aliases: []string{"stake", "s"},
		Short:   "Querying commands for the staking module",
	}

	cmd.AddCommand(
		stakingDelegationCmd(a),
		stakingDelegationsCmd(a),
		stakingUnbondingDelegationCmd(a),
		stakingUnbondingDelegationsCmd(a),
		// stakingRedelegationCmd(),
		// stakingRedelegationsCmd(),
		stakingValidatorCmd(a),
		stakingValidatorsCmd(a),
		stakingValidatorDelegationsCmd(a),
		// stakingValidatorUnbondingDelegationsCmd(),
		// stakingValidatorRedelegationsCmd(),
		// stakingHistoricalInfoCmd(),
		stakingParamsCmd(a),
		stakingPoolCmd(a),
	)

	return cmd
}
