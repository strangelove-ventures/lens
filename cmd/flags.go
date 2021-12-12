package cmd

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func heightFlag(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Int64(flags.FlagHeight, 0, "Height of headers to fetch")
	if err := viper.BindPFlag(flags.FlagHeight, cmd.Flags().Lookup(flags.FlagHeight)); err != nil {
		panic(err)
	}
	return cmd
}

func peersFlag(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Bool("peers", false, "Comma-delimited list of peers to connect to for syncing")
	if err := viper.BindPFlag("peers", cmd.Flags().Lookup("peers")); err != nil {
		panic(err)
	}
	return cmd
}

func logFlag(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Bool("log", false, "show just the transaction logs")
	if err := viper.BindPFlag("log", cmd.Flags().Lookup("log")); err != nil {
		panic(err)
	}
	return cmd
}

func proveFlag(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Bool("prove", false, "return the proof as well as the result")
	if err := viper.BindPFlag("prove", cmd.Flags().Lookup("prove")); err != nil {
		panic(err)
	}
	return cmd
}

func limitFlag(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().Int("limit", 100, "limit the number of things to fetch")
	if err := viper.BindPFlag("limit", cmd.Flags().Lookup("limit")); err != nil {
		panic(err)
	}
	return cmd
}

func skipConfirm(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().BoolP("skip", "y", false, "output using yaml")
	if err := viper.BindPFlag("skip", cmd.Flags().Lookup("skip")); err != nil {
		panic(err)
	}
	return cmd
}
