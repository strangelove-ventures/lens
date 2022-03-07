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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/strangelove-ventures/lens/client"
	"github.com/strangelove-ventures/lens/client/query"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// tendermintCmd represents the tendermint command
func tendermintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tendermint",
		Aliases: []string{"tm"},
		Short:   "all tendermint query commands",
	}
	cmd.AddCommand(
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
	return cmd
}

func abciInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "abci-info",
		Aliases: []string{"abcii"},
		Short:   "queries for block height, app name and app hash",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			query := query.Query{Client: cl, Options: query.DefaultOptions()}

			res, err := query.ABCIInfo()
			if err != nil {
				return err
			}
			return cl.PrintObject(res)
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
			cl := config.GetDefaultClient()
			path := args[0]
			data := args[1]
			prove := false // Hookup to a flag
			height, err := ReadHeight(cmd.Flags())
			if err != nil {
				return err
			}
			options := query.QueryOptions{Pagination: client.DefaultPageRequest(), Height: height}
			query := query.Query{Client: cl, Options: &options}

			res, err := query.ABCIQuery(path, data, prove)
			if err != nil {
				return err
			}
			return cl.PrintObject(res)
		},
	}
	// TODO: add prove flag
	return cmd
}

func blockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "block",
		Aliases: []string{"bl", "blk"},
		Short:   "query tendermint data for a block at a given height",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			height, err := ReadHeight(cmd.Flags())
			if err != nil {
				return err
			}
			options := query.QueryOptions{Pagination: client.DefaultPageRequest(), Height: height}
			query := query.Query{Client: cl, Options: &options}

			block, err := query.Block()
			if err != nil {
				return err
			}
			return cl.PrintObject(block)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func blockByHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "block-by-hash [hash]",
		Aliases: []string{"blhash", "blh", "bbh"},
		Short:   "query tendermint for a given block by hash",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			query := query.Query{Client: cl, Options: query.DefaultOptions()}
			hash := args[0]
			res, err := query.BlockByHash(hash)
			if err != nil {
				return err
			}
			return cl.PrintObject(res)
		},
	}
	return cmd
}

func blockResultsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "block-results",
		Aliases: []string{"blres"},
		Short:   "query tendermint tx results for a given block by height",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			height, err := ReadHeight(cmd.Flags())
			if err != nil {
				return err
			}
			options := query.QueryOptions{Pagination: client.DefaultPageRequest(), Height: height}
			query := query.Query{Client: cl, Options: &options}

			block, err := query.BlockResults()
			if err != nil {
				return err
			}
			return cl.PrintObject(block)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
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
			fmt.Println("TODO")
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
			cl := config.GetDefaultClient()
			height, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			block, err := cl.RPCClient.ConsensusParams(cmd.Context(), &height)
			if err != nil {
				return err
			}
			// TODO: figure out how to fix the base64 output here
			bz, err := json.MarshalIndent(block, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(bz))
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
			cl := config.GetDefaultClient()
			block, err := cl.RPCClient.ConsensusState(cmd.Context())
			if err != nil {
				return err
			}
			bz, err := json.MarshalIndent(block, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(bz))
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
			cl := config.GetDefaultClient()
			block, err := cl.RPCClient.DumpConsensusState(cmd.Context())
			if err != nil {
				return err
			}
			bz, err := json.MarshalIndent(block, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(bz))
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
			cl := config.GetDefaultClient()
			block, err := cl.RPCClient.Health(cmd.Context())
			if err != nil {
				return err
			}
			bz, err := json.MarshalIndent(block, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(bz))
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
			cl := config.GetDefaultClient()
			peers, err := cmd.Flags().GetBool("peers")
			if err != nil {
				return err
			}
			block, err := cl.RPCClient.NetInfo(cmd.Context())
			if err != nil {
				return err
			}
			if !peers {
				bz, err := json.MarshalIndent(block, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(bz))
				return nil
			}
			peersList := ""
			for _, peer := range block.Peers {
				url, err := url.Parse(peer.NodeInfo.ListenAddr)
				if err != nil {
					continue
				}
				peersList += fmt.Sprintf("%s@%s:%s,", peer.NodeInfo.ID(), peer.RemoteIP, url.Port())
			}
			fmt.Println(strings.TrimSuffix(peersList, ","))
			return nil
		},
	}
	return peersFlag(cmd)
}

func numUnconfirmedTxs() *cobra.Command {
	// TODO: add example for parsing these txs
	// _{*extraCredit*}_
	cmd := &cobra.Command{
		Use:     "mempool",
		Aliases: []string{"unconfirmed", "mem"},
		Short:   "query for number of unconfirmed txs",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			limit, err := cmd.Flags().GetInt("limit")
			if err != nil {
				return err
			}
			block, err := cl.RPCClient.UnconfirmedTxs(cmd.Context(), &limit)
			if err != nil {
				return err
			}
			// for _, txbz := range block.Txs {
			// 	fmt.Printf("%X\n", tmtypes.Tx(txbz).Hash())
			// }
			bz, err := json.MarshalIndent(block, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(bz))
			return nil
		},
	}
	return limitFlag(cmd)
}

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"stat", "s"},
		Short:   "query the status of a node",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl := config.GetDefaultClient()
			query := query.Query{Client: cl, Options: query.DefaultOptions()}

			status, err := query.Status()
			if err != nil {
				return err
			}
			return cl.PrintObject(status)
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
			cl := config.GetDefaultClient()
			prove, err := cmd.Flags().GetBool("prove")
			if err != nil {
				return err
			}
			h, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}
			block, err := cl.RPCClient.Tx(cmd.Context(), h, prove)
			if err != nil {
				return err
			}
			return cl.PrintObject(block)

		},
	}
	return proveFlag(cmd)
}
