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
	"github.com/spf13/cobra"
)

// tendermintCmd represents the tendermint command
var tendermintCmd = &cobra.Command{
	Use:     "tendermint",
	Aliases: []string{"tm"},
	Short:   "all tendermint query commands",
}

func init() {
	rootCmd.AddCommand(tendermintCmd)
	tendermintCmd.AddCommand(
		abciInfoCmd(),
		abciQueryCmd(),
		blockCmd(),
		blockByHashCmd(),
		blockResultsCmd(),
		blockSearchCmd(),
		consensusParamsCmd(),
		consensusStateCmd(),
		dumpConsensusStateCmd(),
		healthCmd(),
		netInfoCmd(),
		numUnconfirmedTxs(),
		statusCmd(),
		queryTxCmd(),
	)
}

func abciInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "abci-info",
		Aliases: []string{"abcii"},
		Short:   "queries for block height, app name and app hash",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func abciQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "abci-query [path] [data] [height]",
		Aliases: []string{"qabci"},
		Short:   "query the abci interface for tendermint directly",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	// TODO: add prove flag
	return cmd
}

func blockCmd() *cobra.Command {
	cmd := &cobra.Command{
		// TODO: make this use a height flag and make height arg optional
		Use:     "block [height]",
		Aliases: []string{"bl"},
		Short:   "query tendermint data for a block at given height",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func blockByHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "block-by-hash [hash]",
		Aliases: []string{"blhash", "blh"},
		Short:   "query tendermint for a given block by hash",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func blockResultsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "block-results [height]",
		Aliases: []string{"blres"},
		Short:   "query tendermint tx results for a given block by height",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func blockSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "block-search [query] [page] [per-page]",
		Aliases: []string{"bls", "bs", "blsearch"},
		Short:   "search blocks with given query",
		// TODO: long explaination and example should include example queries
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	// TODO: order by flag
	return cmd
}

func consensusParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		// TODO: make this use a height flag and make height arg optional
		Use:     "consensus-params [height]",
		Aliases: []string{"csparams", "cs-params"},
		Short:   "query tendermint consensus params at a given height",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func consensusStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		// TODO: add special flag to this for network startup
		// that runs query on timer and shows a progress bar
		// _{*extraCredit*}_
		Use:     "consensus-state",
		Aliases: []string{"csstate", "cs-state"},
		Short:   "query current tendermint consensus state",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func dumpConsensusStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dump-consensus-state",
		Aliases: []string{"dump-cs", "csdump", "cs-dump", "dumpcs"},
		Short:   "query detailed version of current tendermint consensus state",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func healthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "health",
		Aliases: []string{"h", "ok"},
		Short:   "query to see if node server is online",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func netInfoCmd() *cobra.Command {
	// TODO: add flag for pulling out comma seperated list of peers
	// and also filter out private IPs and other ill formed peers
	// _{*extraCredit*}_
	cmd := &cobra.Command{
		Use:     "net-info",
		Aliases: []string{"ni", "net", "netinfo", "peers"},
		Short:   "query for p2p network connection information",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func numUnconfirmedTxs() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "num-unconfirmed-txs",
		Aliases: []string{"count-unconf", "unconf-count"},
		Short:   "query for number of unconfirmed txs",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"stat", "s"},
		Short:   "query status of the node",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}

func queryTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx [hash]",
		Short: "query for a transaction by hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	// TODO: add prove flag
	return cmd
}
