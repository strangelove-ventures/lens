package cmd

import (
	"github.com/spf13/cobra"
)

func getDistributionQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "distribution",
		Short: "Query things about a chain's distribution module",
	}

	cmd.AddCommand(
		queryDistributionParamsCmd(),
		queryDistributionCommunityPoolCmd(),
		queryDistributionRewardsCmd(),
		queryDistributionSlashesCmd(),
		queryDistributionValidatorRewardsCmd(),
		queryDistributionCommissionCmd(),
	)

	return cmd
}

func queryDistributionParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "query things about a chain's distribution params",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			params, err := cl.QueryDistributionParams()
			if err != nil {
				return err
			}
			cl.PrintObject(params)

			return nil
		},
	}

	return cmd
}

func queryDistributionCommunityPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "community-pool",
		Short: "query things about a chain's community pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			pool, err := cl.QueryDistributionCommunityPool()
			if err != nil {
				return err
			}
			cl.PrintObject(pool)

			return nil
		},
	}

	return cmd
}

func queryDistributionCommissionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commission [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "query a specific validator's comission",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			address := args[0]
			commission, err := cl.QueryDistributionCommission(address)

			if err != nil {
				return err
			}

			cl.PrintObject(commission)
			return nil
		},
	}

	return cmd
}

func queryDistributionRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards",
		Short: "query things about a delegator's rewards",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			delegatorAddress := args[0]
			validatorAddress := args[1]

			pool, err := cl.QueryDistributionRewards(delegatorAddress, validatorAddress)
			if err != nil {
				return err
			}
			cl.PrintObject(pool)

			return nil
		},
	}

	return cmd
}

func queryDistributionSlashesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slashes",
		Short: "query things about a validator's slashes on a chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			address := args[0]

			slashes, err := cl.QueryDistributionSlashes(address)
			if err != nil {
				return err
			}
			cl.PrintObject(slashes)

			return nil
		},
	}

	return cmd
}

func queryDistributionValidatorRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator-outstanding-rewards [address]",
		Short: "query things about a validator's (and all their delegators) outstanding rewards on a chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			address := args[0]

			rewards, err := cl.QueryDistributionValidatorRewards(address)
			if err != nil {
				return err
			}
			cl.PrintObject(rewards)

			return nil
		},
	}

	return cmd
}
