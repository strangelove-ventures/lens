package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/strangelove-ventures/lens/client/chain_registry"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func chainsCmd(v *viper.Viper, lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "chains",
		Aliases: []string{"ch", "c"},
		Short:   "manage local chain configurations",
	}

	cmd.AddCommand(
		cmdChainsAdd(v, lc),
		cmdChainsDelete(v, lc),
		cmdChainsEdit(v, lc),
		cmdChainsList(lc),
		cmdChainsShow(lc),
		cmdChainsSetDefault(v, lc),
		cmdChainsRegistryList(lc),
		cmdChainsShowDefault(lc),
		cmdChainsEditorDefault(),
	)

	return cmd
}

func cmdChainsRegistryList(lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "registry-list",
		Args:    cobra.NoArgs,
		Aliases: []string{"rl"},
		Short:   "list chains available for configuration from the registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			chains, err := chain_registry.DefaultChainRegistry().ListChains()
			if err != nil {
				return err
			}
			return lc.config.GetDefaultClient().PrintObject(chains)
		},
	}
	return cmd
}

func cmdChainsAdd(v *viper.Viper, lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [[chain-name]]",
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"a"},
		Short:   "add configuration for a chain or a number of chains from the chain registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			registry := chain_registry.DefaultChainRegistry()
			allChains, err := registry.ListChains()
			if err != nil {
				return err
			}

			for _, chain := range args {
				found := false
				for _, possibleChain := range allChains {
					if chain == possibleChain {
						found = true
					}
				}

				if !found {
					log.Printf("unable to find chain %s in %s", chain, registry.SourceLink())
					continue
				}

				chainInfo, err := registry.GetChain(chain)
				if err != nil {
					log.Printf("error getting chain: %s", err)
					continue
				}

				chainConfig, err := chainInfo.GetChainConfig()
				if err != nil {
					log.Printf("error generating chain config: %s", err)
					continue
				}

				lc.config.Chains[chain] = chainConfig
			}

			return overwriteConfig(v, &lc.config)
		},
	}
	return cmd
}

func cmdChainsDelete(v *viper.Viper, lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [[chain-name]]",
		Aliases: []string{"d"},
		Short:   "delete a chain from the configuration",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			originalChainCount := len(lc.config.Chains)
			for _, arg := range args {
				if lc.config.DefaultChain == arg {
					fmt.Fprintf(cmd.ErrOrStderr(), "Ignoring delete request for %s, unable to delete default chain.\n", arg)
					continue
				}
				delete(lc.config.Chains, arg)
			}

			// If nothing was removed, there's no need to update the configuration file.
			if len(lc.config.Chains) == originalChainCount {
				return nil
			}

			return overwriteConfig(v, &lc.config)
		},
	}
	return cmd
}

func cmdChainsEdit(v *viper.Viper, lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit [chain-name] [key] [value]",
		Aliases: []string{"e"},
		Short:   "edit a chain configuration value",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := lc.config.Chains[args[0]]; !ok {
				return fmt.Errorf("chain %s not found in configuration", args[0])
			}
			switch args[1] {
			case "key":
				lc.config.Chains[args[0]].Key = args[2]
			case "chain-id":
				lc.config.Chains[args[0]].ChainID = args[2]
			case "rpc-addr":
				lc.config.Chains[args[0]].RPCAddr = args[2]
			case "grpc-addr":
				lc.config.Chains[args[0]].GRPCAddr = args[2]
			case "account-prefix":
				lc.config.Chains[args[0]].AccountPrefix = args[2]
			case "gas-adjustment":
				fl, err := strconv.ParseFloat(args[2], 64)
				if err != nil {
					return err
				}
				lc.config.Chains[args[0]].GasAdjustment = fl
			case "gas-prices":
				lc.config.Chains[args[0]].GasPrices = args[2]
			case "debug":
				b, err := strconv.ParseBool(args[2])
				if err != nil {
					return err
				}
				lc.config.Chains[args[0]].Debug = b
			case "timeout":
				lc.config.Chains[args[0]].Timeout = args[2]
			default:
				return fmt.Errorf("unknown key %s, try 'key', 'chain-id', 'rpc-addr', 'grpc-addr', 'account-prefix', 'gas-adjustment', 'gas-prices', 'debug', or 'timeout'", args[1])
			}
			return overwriteConfig(v, &lc.config)
		},
	}
	return cmd
}

func cmdChainsList(lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List all chains in the configuration",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return lc.config.GetDefaultClient().PrintObject(lc.config.Chains)
		},
	}
	return cmd
}

func cmdChainsShow(lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [chain-name]",
		Aliases: []string{"s"},
		Short:   "show an individual chain configuration",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Return a helpful error so the user knows which chain names are available.
				names := make([]string, 0, len(lc.config.Chains))
				for name := range lc.config.Chains {
					names = append(names, name)
				}
				sort.Strings(names)
				return fmt.Errorf("no chain-name provided; available names are: %s", strings.Join(names, ", "))
			}

			if ch, ok := lc.config.Chains[args[0]]; ok {
				return lc.config.GetDefaultClient().PrintObject(ch)
			}
			return fmt.Errorf("chain %s not found", args[0])
		},
	}
	return cmd
}

func cmdChainsSetDefault(v *viper.Viper, lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-default [chain-name]",
		Aliases: []string{"sd"},
		Short:   "set the default chain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := lc.config.Chains[args[0]]; ok {
				lc.config.DefaultChain = args[0]
				return overwriteConfig(v, &lc.config)
			}
			return fmt.Errorf("chain %s not found", args[0])
		},
	}
	return cmd
}

func cmdChainsShowDefault(lc *lensConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-default",
		Aliases: []string{"d", "default"},
		Short:   "show the configured default chain",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), lc.config.DefaultChain)
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
