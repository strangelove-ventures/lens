package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/lens/client/chain_registry"
	"go.uber.org/zap"
)

func chainsCmd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "chains",
		Aliases: []string{"ch", "c"},
		Short:   "manage local chain configurations",
	}

	cmd.AddCommand(
		cmdChainsAdd(a),
		cmdChainsDelete(a),
		cmdChainsEdit(a),
		cmdChainsList(a),
		cmdChainsShow(a),
		cmdChainsSetDefault(a),
		cmdChainsRegistryList(a),
		cmdChainsShowDefault(a),
		cmdChainsEditorDefault(),
	)

	return cmd
}

func cmdChainsRegistryList(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "registry-list",
		Args:    cobra.NoArgs,
		Aliases: []string{"rl"},
		Short:   "list chains available for configuration from the registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			chains, err := chain_registry.DefaultChainRegistry(a.Log).ListChains(cmd.Context())
			if err != nil {
				return err
			}
			return a.Config.GetDefaultClient().PrintObject(chains)
		},
	}
	return cmd
}

func cmdChainsAdd(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [[chain-name]]",
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"a"},
		Short:   "add configuration for a chain or a number of chains from the chain registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := chain_registry.DefaultChainRegistry(a.Log)
			overwriteConfig := false

			for _, chain := range args {
				chainInfo, err := registry.GetChain(cmd.Context(), chain)
				if err != nil {
					a.Log.Info(
						"Failed to get chain",
						zap.String("name", chain),
						zap.Error(err),
					)
					continue
				}

				chainConfig, err := chainInfo.GetChainConfig(cmd.Context())
				if err != nil {
					a.Log.Info(
						"Failed to generate chain config",
						zap.String("name", chain),
						zap.Error(err),
					)
					continue
				}
				overwriteConfig = true
				a.Config.Chains[chain] = chainConfig
			}
			if overwriteConfig {
				return a.OverwriteConfig(a.Config)
			} else {
				return nil
			}
		},
	}
	return cmd
}

func cmdChainsDelete(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [[chain-name]]",
		Aliases: []string{"d"},
		Short:   "delete a chain from the configuration",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			originalChainCount := len(a.Config.Chains)
			for _, arg := range args {
				if a.Config.DefaultChain == arg {
					fmt.Fprintf(cmd.ErrOrStderr(), "Ignoring delete request for %s, unable to delete default chain.\n", arg)
					continue
				}
				delete(a.Config.Chains, arg)
			}

			// If nothing was removed, there's no need to update the configuration file.
			if len(a.Config.Chains) == originalChainCount {
				return nil
			}

			return a.OverwriteConfig(a.Config)
		},
	}
	return cmd
}

func cmdChainsEdit(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit [chain-name] [key] [value]",
		Aliases: []string{"e"},
		Short:   "edit a chain configuration value",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := a.Config.Chains[args[0]]; !ok {
				return fmt.Errorf("chain %s not found in configuration", args[0])
			}
			switch args[1] {
			case "key":
				a.Config.Chains[args[0]].Key = args[2]
			case "chain-id":
				a.Config.Chains[args[0]].ChainID = args[2]
			case "rpc-addr":
				a.Config.Chains[args[0]].RPCAddr = args[2]
			case "grpc-addr":
				a.Config.Chains[args[0]].GRPCAddr = args[2]
			case "account-prefix":
				a.Config.Chains[args[0]].AccountPrefix = args[2]
			case "gas-adjustment":
				fl, err := strconv.ParseFloat(args[2], 64)
				if err != nil {
					return err
				}
				a.Config.Chains[args[0]].GasAdjustment = fl
			case "gas-prices":
				a.Config.Chains[args[0]].GasPrices = args[2]
			case "min-gas-amount":
				ga, err := strconv.ParseUint(args[2], 10, 64)
				if err != nil {
					return err
				}
				a.Config.Chains[args[0]].MinGasAmount = ga
			case "debug":
				b, err := strconv.ParseBool(args[2])
				if err != nil {
					return err
				}
				a.Config.Chains[args[0]].Debug = b
			case "timeout":
				a.Config.Chains[args[0]].Timeout = args[2]
			default:
				return fmt.Errorf("unknown key %s, try 'key', 'chain-id', 'rpc-addr', 'grpc-addr', 'account-prefix', 'gas-adjustment', 'gas-prices', 'min-gas-amount', 'debug', or 'timeout'", args[1])
			}
			return a.OverwriteConfig(a.Config)
		},
	}
	return cmd
}

func cmdChainsList(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List all chains in the configuration",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.Config.GetDefaultClient().PrintObject(a.Config.Chains)
		},
	}
	return cmd
}

func cmdChainsShow(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [chain-name]",
		Aliases: []string{"s"},
		Short:   "show an individual chain configuration",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Return a helpful error so the user knows which chain names are available.
				names := make([]string, 0, len(a.Config.Chains))
				for name := range a.Config.Chains {
					names = append(names, name)
				}
				sort.Strings(names)
				return fmt.Errorf("no chain-name provided; available names are: %s", strings.Join(names, ", "))
			}

			if ch, ok := a.Config.Chains[args[0]]; ok {
				return a.Config.GetDefaultClient().PrintObject(ch)
			}
			return fmt.Errorf("chain %s not found", args[0])
		},
	}
	return cmd
}

func cmdChainsSetDefault(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-default [chain-name]",
		Aliases: []string{"sd"},
		Short:   "set the default chain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := a.Config.Chains[args[0]]; ok {
				a.Config.DefaultChain = args[0]
				return a.OverwriteConfig(a.Config)
			}
			return fmt.Errorf("chain %s not found", args[0])
		},
	}
	return cmd
}

func cmdChainsShowDefault(a *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-default",
		Aliases: []string{"d", "default"},
		Short:   "show the configured default chain",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), a.Config.DefaultChain)
			return nil
		},
	}
	return cmd
}

func cmdChainsEditorDefault() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "editor",
		Short: "Open Lens configuration in an editor",
		Long: `Open Lens configuration in an editor. By default, command will spawn a vim window. You can 
override the editor using the environment variable LENS_EDITOR. Please ensure $LENS_EDITOR points to 
an editor in your path that can be called using $LENS_EDITOR <file-path>.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString("home")
			if err != nil {
				return err
			}

			editor := os.Getenv("LENS_EDITOR")
			if editor == "" {
				editor = os.Getenv("EDITOR") // Should hold system default
				if editor == "" {
					editor = "vi"
				}
			}

			c := exec.Command(editor, path.Join(home, "config.yaml"))
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			return c.Run()
		},
	}
	return cmd
}
