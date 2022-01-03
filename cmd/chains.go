package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/lens/internal/chain_registry"
)

func chainsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "chains",
		Aliases: []string{"ch", "c"},
		Short:   "manage local chain configurations",
	}

	cmd.AddCommand(
		cmdChainsAdd(),
		cmdChainsDelete(),
		cmdChainsEdit(),
		cmdChainsList(),
		cmdChainsShow(),
		cmdChainsSetDefault(),
		cmdChainsRegistryList(),
		cmdChainsShowDefault(),
		cmdChainsEditorDefault(),
	)

	return cmd
}

func cmdChainsRegistryList() *cobra.Command {
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
			return config.GetDefaultClient().PrintObject(chains)
		},
	}
	return cmd
}

func cmdChainsAdd() *cobra.Command {
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
					return err
				}

				chainConfig, err := chainInfo.GetChainConfig()
				if err != nil {
					return err
				}

				config.Chains[chain] = chainConfig
			}

			return overwriteConfig(config)
		},
	}
	return cmd
}

func cmdChainsDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [[chain-name]]",
		Aliases: []string{"d"},
		Short:   "delete a chain from the configuration",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				delete(config.Chains, arg)
			}
			return overwriteConfig(config)
		},
	}
	return cmd
}

func cmdChainsEdit() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit [chain-name] [key] [value]",
		Aliases: []string{"e"},
		Short:   "edit a chain configuration value",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := config.Chains[args[0]]; !ok {
				return fmt.Errorf("chain %s not found in configuration", args[0])
			}
			switch args[1] {
			case "key":
				config.Chains[args[0]].Key = args[2]
			case "chain-id":
				config.Chains[args[0]].ChainID = args[2]
			case "rpc-addr":
				config.Chains[args[0]].RPCAddr = args[2]
			case "grpc-addr":
				config.Chains[args[0]].GRPCAddr = args[2]
			case "account-prefix":
				config.Chains[args[0]].AccountPrefix = args[2]
			case "gas-adjustment":
				fl, err := strconv.ParseFloat(args[2], 64)
				if err != nil {
					return err
				}
				config.Chains[args[0]].GasAdjustment = fl
			case "gas-prices":
				config.Chains[args[0]].GasPrices = args[2]
			case "debug":
				b, err := strconv.ParseBool(args[2])
				if err != nil {
					return err
				}
				config.Chains[args[0]].Debug = b
			case "timeout":
				config.Chains[args[0]].Timeout = args[2]
			default:
				return fmt.Errorf("unknown key %s, try 'key', 'chain-id', 'rpc-addr', 'grpc-addr', 'account-prefix', 'gas-adjustment', 'gas-prices', 'debug', or 'timeout'", args[1])
			}
			return overwriteConfig(config)
		},
	}
	return cmd
}

func cmdChainsList() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l"},
		Short:   "List all chains in the configuration",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.GetDefaultClient().PrintObject(config.Chains)
		},
	}
	return cmd
}

func cmdChainsShow() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [chain-name]",
		Aliases: []string{"s"},
		Short:   "show an individual chain configuration",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ch, ok := config.Chains[args[0]]; ok {
				return config.GetDefaultClient().PrintObject(ch)

			}
			return fmt.Errorf("chain %s not found", args[0])
		},
	}
	return cmd
}

func cmdChainsSetDefault() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-default [chain-name]",
		Aliases: []string{"sd"},
		Short:   "set the default chain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := config.Chains[args[0]]; ok {
				config.DefaultChain = args[0]
				return overwriteConfig(config)
			}
			return fmt.Errorf("chain %s not found", args[0])
		},
	}
	return cmd
}

func cmdChainsShowDefault() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-default",
		Aliases: []string{"d", "default"},
		Short:   "show the configured default chain",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.GetDefaultClient().PrintObject(config.DefaultChain)
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

			if editor := os.Getenv("LENS_EDITOR"); editor == "" {
				// 'EDITOR' should hold the system default
				if editor = os.Getenv("EDITOR"); editor == "" {
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
