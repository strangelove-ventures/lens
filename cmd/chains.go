package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
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
	)

	return cmd
}

func cmdChainsRegistryList() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "registry-list",
		Args:    cobra.NoArgs,
		Aliases: []string{"rl"},
		Short:   "list chains available for configuration from the regitry",
		RunE: func(cmd *cobra.Command, args []string) error {
			chains, err := fetchRegistryChains(cmd.Context())
			if err != nil {
				return err
			}
			bz, err := json.Marshal(chains)
			if err != nil {
				return err
			}
			fmt.Println(string(bz))
			return nil
		},
	}
	return cmd
}

func fetchRegistryChains(ctx context.Context) ([]string, error) {
	chains := []string{}
	tree, res, err := github.NewClient(nil).Git.GetTree(ctx, "cosmos", "chain-registry", "master", true)
	if err != nil || res.StatusCode != 200 {
		return chains, err
	}
	for _, entry := range tree.Entries {
		if *entry.Type == "tree" && !strings.Contains(*entry.Path, ".github") {
			chains = append(chains, *entry.Path)
		}
	}
	return chains, nil
}

func cmdChainsAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [[chain-name]]",
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"a"},
		Short:   "add configraion for a chain or a number of chains from the chain registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: input validation, maybe fetch the registry chains and validate the chain name
			debug, _ := cmd.Flags().GetBool("debug")
			home, _ := cmd.Flags().GetString("home")
			for _, chain := range args {
				ch, err := getChainConfigFromRegistry(chain, path.Join(home, "keys"), debug)
				if err != nil {
					return err
				}
				config.Chains[chain] = ch
			}
			return overwriteConfig(home, config)
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
			home, _ := cmd.Flags().GetString("home")
			for _, arg := range args {
				delete(config.Chains, arg)
			}
			return overwriteConfig(home, config)
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
			home, _ := cmd.Flags().GetString("home")
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
			return overwriteConfig(home, config)
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
			bz, err := json.Marshal(config.Chains)
			if err != nil {
				return err
			}
			fmt.Println(string(bz))
			return nil
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
			ch, ok := config.Chains[args[0]]
			if !ok {
				return fmt.Errorf("chain %s not found", args[0])
			}
			bz, err := json.Marshal(ch)
			if err != nil {
				return err
			}
			fmt.Println(string(bz))
			return nil
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
			home, _ := cmd.Flags().GetString("home")
			_, ok := config.Chains[args[0]]
			if !ok {
				return fmt.Errorf("chain %s not found", args[0])
			}
			config.DefaultChain = args[0]
			return overwriteConfig(home, config)
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
			fmt.Println(config.DefaultChain)
			return nil
		},
	}
	return cmd
}
