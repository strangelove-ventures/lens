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
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	zaplogfmt "github.com/jsternberg/zap-logfmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	provtypes "github.com/cometbft/cometbft/light/provider"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"
)

const appName = "lens"

// ClientOverrides specifies an RPCClient and LightProvider
// to use for a specific chain.
//
// This should only be set during tests.
type ClientOverrides struct {
	RPCClient     rpcclient.Client
	LightProvider provtypes.Provider
}

// NewRootCmd returns the root command for relayer.
//
// o is used to override rpc clients and light providers for test.
// If o is nil, reasonable default values are used.
func NewRootCmd(log *zap.Logger, atom zap.AtomicLevel, o map[string]ClientOverrides) *cobra.Command {
	// Use a local app state instance scoped to the new root command,
	// so that tests don't concurrently access the state.
	a := &appState{
		Log: log,

		Viper: viper.New(),
	}

	defaultHome := os.ExpandEnv("$HOME/.lens")

	// RootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   appName,
		Short: "This is my lens, there are many like it, but this one is mine.",
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		// Inside persistent pre-run because this takes effect after flags are parsed.
		if a.Viper.GetBool("debug") {
			atom.SetLevel(zapcore.DebugLevel)
		}

		// reads `homeDir/config.yaml` into `var config *Config` before each command
		if err := initConfig(rootCmd, a, o); err != nil {
			return err
		}

		return nil
	}

	// --home flag
	rootCmd.PersistentFlags().StringVar(&a.HomePath, flags.FlagHome, defaultHome, "set home directory")
	if err := a.Viper.BindPFlag(flags.FlagHome, rootCmd.PersistentFlags().Lookup(flags.FlagHome)); err != nil {
		panic(err)
	}

	// --debug flag
	rootCmd.PersistentFlags().BoolVarP(&a.Debug, "debug", "d", false, "debug output")
	if err := a.Viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")); err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().StringP("output", "o", "json", "output format (json, indent, yaml)")
	if err := a.Viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")); err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().StringVar(&a.OverriddenChain, "chain", "", "override default chain")
	if err := a.Viper.BindPFlag("chain", rootCmd.PersistentFlags().Lookup("chain")); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(
		chainsCmd(a),
		keysCmd(a),
		queryCmd(a),
		tendermintCmd(a),
		crosschainCmd(a),
		txCmd(a),
		versionCmd(),
		airdropCmd(a),
		dynamicCmd(a),
	)

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.EnableCommandSorting = false

	log, atom := rootLogger()
	defer log.Sync()

	rootCmd := NewRootCmd(log, atom, nil)
	rootCmd.SilenceUsage = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	if err := rootCmd.Execute(); err != nil {
		log.Sync()
		os.Exit(1)
	}
}

func rootLogger() (*zap.Logger, zap.AtomicLevel) {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(ts.UTC().Format("2006-01-02T15:04:05.000000Z07:00"))
	}
	config.LevelKey = "lvl"

	var enc zapcore.Encoder
	if term.IsTerminal(int(os.Stderr.Fd())) {
		// When a user runs relayer in the foreground, use easier to read output.
		enc = zapcore.NewConsoleEncoder(config)
	} else {
		// Otherwise, use consistent logfmt format for simplistic machine processing.
		enc = zaplogfmt.NewEncoder(config)
	}

	atom := zap.NewAtomicLevel()
	return zap.New(zapcore.NewCore(
		enc,
		os.Stderr,
		atom,
	)), atom
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

// withUsage wraps a PositionalArgs to display usage only when the PositionalArgs
// variant is violated.
func withUsage(inner cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := inner(cmd, args); err != nil {
			cmd.Root().SilenceUsage = false
			cmd.SilenceUsage = false
			return err
		}

		return nil
	}
}
