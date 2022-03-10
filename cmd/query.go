package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// queryCmd represents the query command tree.
func queryCmd(v *viper.Viper, lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "query things about a chain",
	}

	cmd.AddCommand(
		authQueryCmd(v, lc),
		authzQueryCmd(v, lc),
		bankQueryCmd(lc),
		distributionQueryCmd(v, lc),
		stakingQueryCmd(lc),
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
func authQueryCmd(v *viper.Viper, lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "auth",
		Aliases: []string{"a"},
		Short:   "Querying commands for the auth module",
	}

	cmd.AddCommand(
		authAccountCmd(lc),
		authAccountsCmd(v, lc),
		authParamsCmd(lc),
	)

	return cmd
}

// authzQueryCmd returns the authz query commands for this module
func authzQueryCmd(v *viper.Viper, lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "authz",
		Aliases: []string{"authz"},
		Short:   "Querying commands for the authz module",
	}

	cmd.AddCommand(
		authzGrantsCmd(v, lc),
	)

	return cmd
}

// bankQueryCmd  returns the transaction commands for this module
func bankQueryCmd(lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bank",
		Aliases: []string{"b"},
		Short:   "Querying commands for the bank module",
	}

	cmd.AddCommand(
		bankBalanceCmd(lc),
		bankTotalSupplyCmd(lc),
		bankDenomsMetadataCmd(lc),
	)

	return cmd
}

// distributionQueryCmd returns the distribution query commands for this module
func distributionQueryCmd(v *viper.Viper, lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "distribution",
		Aliases: []string{"dist", "distr", "d"},
		Short:   "Querying commands for the distribution module",
	}

	cmd.AddCommand(
		distributionParamsCmd(lc),
		distributionValidatorRewardsCmd(lc),
		distributionCommissionCmd(lc),
		distributionCommunityPoolCmd(lc),
		distributionRewardsCmd(lc),
		distributionSlashesCmd(v, lc),
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
func stakingQueryCmd(lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "staking",
		Aliases: []string{"stake", "s"},
		Short:   "Querying commands for the staking module",
	}

	cmd.AddCommand(
		stakingDelegationCmd(lc),
		stakingDelegationsCmd(lc),
		// stakingUnbondingDelegationCmd(),
		// stakingUnbondingDelegationsCmd(),
		// stakingRedelegationCmd(),
		// stakingRedelegationsCmd(),
		// stakingValidatorCmd(),
		// stakingValidatorsCmd(),
		// stakingValidatorDelegationsCmd(),
		// stakingValidatorUnbondingDelegationsCmd(),
		// stakingValidatorRedelegationsCmd(),
		// stakingHistoricalInfoCmd(),
		// stakingParamsCmd(),
		// stakingPoolCmd(),
	)

	return cmd
}
