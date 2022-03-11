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
	"encoding/json"
	"io"
	"os"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

const appName = "lens"

type lensConfig struct {
	homePath       string
	overridenChain string
	debug          bool
	config         Config
}

// NewRootCmd returns the root command for relayer.
func NewRootCmd() *cobra.Command {
	// Use a local viper instance scoped to a root command,
	// so that tests don't concurrently access a single viper instance.
	v := viper.New()

	defaultHome := os.ExpandEnv("$HOME/.lens")

	var lc lensConfig

	// RootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   appName,
		Short: "This is my lens, there are many like it, but this one is mine.",
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		// reads `homeDir/config.yaml` into `var config *Config` before each command
		if err := initConfig(rootCmd, v, &lc); err != nil {
			return err
		}

		return nil
	}

	// --home flag
	rootCmd.PersistentFlags().StringVar(&lc.homePath, flags.FlagHome, defaultHome, "set home directory")
	if err := v.BindPFlag(flags.FlagHome, rootCmd.PersistentFlags().Lookup(flags.FlagHome)); err != nil {
		panic(err)
	}

	// --debug flag
	rootCmd.PersistentFlags().BoolVarP(&lc.debug, "debug", "d", false, "debug output")
	if err := v.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")); err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().StringP("output", "o", "json", "output format (json, indent, yaml)")
	if err := v.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")); err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().StringVar(&lc.overridenChain, "chain", "", "override default chain")
	if err := v.BindPFlag("chain", rootCmd.PersistentFlags().Lookup("chain")); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(
		chainsCmd(v, &lc),
		keysCmd(v, &lc),
		queryCmd(v, &lc),
		tendermintCmd(v, &lc),
		crosschainCmd(&lc),
		txCmd(&lc),
		versionCmd(),
		airdropCmd(&lc),
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

// writeJSON encodes the given object to the given writer.
func writeJSON(w io.Writer, obj interface{}) error {
	// Although simple, this is just subtle enough
	// and used in enough places to justify its own function.

	// Using an encoder is slightly preferable over json.Marshal
	// as the encoder will write directly to the io.Writer
	// instead of copying to a temporary buffer.
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(obj)
}
