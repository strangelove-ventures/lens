package cmd

import (
	"encoding/json"
	"fmt"
	"path"
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
			tree, res, err := github.NewClient(nil).Git.GetTree(cmd.Context(), "cosmos", "chain-registry", "master", true)
			if err != nil || res.StatusCode != 200 {
				return err
			}
			chains := []string{}
			for _, entry := range tree.Entries {
				if *entry.Type == "tree" && !strings.Contains(*entry.Path, ".github") {
					chains = append(chains, *entry.Path)
				}
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

func cmdChainsAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add [[chain-name]]",
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"a"},
		Short:   "add configraion for a chain or a number of chains from the chain registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: input validation
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
		Use:     "delete ",
		Aliases: []string{},
		Short:   "",
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}
	return cmd
}

func cmdChainsEdit() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Edit",
		Aliases: []string{},
		Short:   "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func cmdChainsList() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "List",
		Aliases: []string{},
		Short:   "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func cmdChainsShow() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Show",
		Aliases: []string{},
		Short:   "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func cmdChainsSetDefault() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "SetDefault",
		Aliases: []string{},
		Short:   "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}
