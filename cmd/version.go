package cmd

import (
	"encoding/json"
	"fmt"
	dbg "runtime/debug"

	"github.com/spf13/cobra"
)

var (
	Version string
	Commit  string
)

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "show version information for lens, sdk, and tendermint",
		RunE: func(cmd *cobra.Command, args []string) error {
			bi, ok := dbg.ReadBuildInfo()
			if !ok {
				return fmt.Errorf("didn't return build info")
			}

			dependencyVersions := map[string]string{}

			for _, dep := range bi.Deps {
				dependencyVersions[dep.Path] = dep.Version
			}

			v := version{
				Version:    Version,
				Commit:     Commit,
				CosmosSDK:  dependencyVersions["github.com/cosmos/cosmos-sdk"],
				Tendermint: dependencyVersions["github.com/tendermint/tendermint"],
			}

			bz, _ := json.MarshalIndent(v, "", "  ")
			fmt.Println(string(bz))
			return nil
		},
	}

	return cmd
}

type version struct {
	Version    string `json:"version"`
	Commit     string `json:"commit"`
	CosmosSDK  string `json:"cosmos_sdk"`
	Tendermint string `json:"tendermint"`
}
