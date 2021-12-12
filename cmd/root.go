/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var (
	homePath    string
	debug       bool
	config      *Config
	defaultHome = os.ExpandEnv("$HOME/.lens")
	appName     = "lens"
)

// NewRootCmd returns the root command for relayer.
func NewRootCmd() *cobra.Command {
	// RootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   appName,
		Short: "This is my lens, there are many like it, but this one is mine.",
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		// reads `homeDir/config.yaml` into `var config *Config` before each command
		if err := initConfig(rootCmd); err != nil {
			return err
		}

		return nil
	}

	// --home flag
	rootCmd.PersistentFlags().StringVar(&homePath, flags.FlagHome, defaultHome, "set home directory")
	if err := viper.BindPFlag(flags.FlagHome, rootCmd.PersistentFlags().Lookup(flags.FlagHome)); err != nil {
		panic(err)
	}

	// --debug flag
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug output")
	if err := viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(
		configCmd(),
		tendermintCmd(),
		keysCmd(),
	)

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.EnableCommandSorting = false

	rootCmd := NewRootCmd()
	rootCmd.SilenceUsage = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
