package cmd

import (
	"fmt"
	dbg "runtime/debug"
	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "show version information for lens, sdk, and tendermint",
		RunE: func(cmd *cobra.Command, args []string) error {
			bi, _ := dbg.ReadBuildInfo()

			dependencyVersions := map[string]string{}

			for _, dep := range bi.Deps {
				dependencyVersions[dep.Path] = dep.Version
			}

			fmt.Printf(`Lens: %s
Cosmos SDK: %s
Tendermint: %s
`, 
				bi.Main.Version,
				dependencyVersions["github.com/cosmos/cosmos-sdk"],
				dependencyVersions["github.com/tendermint/tendermint"],
			)

			return nil
		},
	}

	return cmd
}