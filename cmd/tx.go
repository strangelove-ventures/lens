package cmd

import "github.com/spf13/cobra"

// TxCommand regesters a new tx command.
func txCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "query things about a chain",
	}

	cmd.AddCommand(bankTxCmd())
	cmd.AddCommand(stakingTxCmd())
	cmd.AddCommand(distributionTxCmd())

	return cmd
}

func bankTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bank",
		Aliases: []string{"b"},
		Short:   "bank things",
	}

	cmd.AddCommand(bankSendCmd())

	return cmd
}

func stakingTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "staking",
		Aliases: []string{"stake", "stk"},
		Short:   "staking things",
	}

	cmd.AddCommand(stakingDelegateCmd())
	cmd.AddCommand(stakingRedelegateCmd())

	return cmd
}

func distributionTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "distribution",
		Aliases: []string{"dist", "distr", "d"},
		Short:   "distribution things",
	}

	cmd.AddCommand(distributionWithdrawRewardsCmd())

	return cmd
}
