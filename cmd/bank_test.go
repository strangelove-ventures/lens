package cmd_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/google/go-cmp/cmp"
	"github.com/strangelove-ventures/lens/cmd"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/rpc/client/mocks"
)

func TestQueryBankBalances_SpecificAddress(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	balanceResp := banktypes.QueryAllBalancesResponse{
		Balances: types.NewCoins(types.Coin{
			Denom:  "atom",
			Amount: types.NewInt(2),
		}),
	}
	mc := new(mocks.Client)
	MockQueryAllBalances(
		t, mc,
		&banktypes.QueryAllBalancesRequest{
			Address:    ZeroCosmosAddr,
			Pagination: &query.PageRequest{Limit: 100}, // Default pagination setting.
		},
		&balanceResp,
	)
	sys.OverrideClients("cosmoshub", cmd.ClientOverrides{
		RPCClient: mc,
	})

	// Querying the balances shows nothing on stderr, serialized response on stdout.
	res := sys.MustRun(t, "query", "bank", "balances", ZeroCosmosAddr)
	require.Empty(t, res.Stderr.String())

	var gotBalances banktypes.QueryAllBalancesResponse
	require.NoError(t, json.Unmarshal(res.Stdout.Bytes(), &gotBalances))

	require.Empty(t, cmp.Diff(balanceResp, gotBalances))
}

func TestQueryBankTotalSupply(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	totalSupplyResp := banktypes.QueryTotalSupplyResponse{
		Supply: types.NewCoins(types.Coin{
			Denom:  "atom",
			Amount: types.NewInt(1000),
		}),
	}
	mc := new(mocks.Client)
	MockQueryTotalSupply(
		t, mc,
		&banktypes.QueryTotalSupplyRequest{
			Pagination: &query.PageRequest{Limit: 100}, // Default pagination setting.
		},
		&totalSupplyResp,
	)
	sys.OverrideClients("cosmoshub", cmd.ClientOverrides{
		RPCClient: mc,
	})

	// Querying the total supply shows nothing on stderr, serialized response on stdout.
	res := sys.MustRun(t, "query", "bank", "total-supply")
	require.Empty(t, res.Stderr.String())

	var gotTotalSupply banktypes.QueryTotalSupplyResponse
	require.NoError(t, json.Unmarshal(res.Stdout.Bytes(), &gotTotalSupply))

	require.Empty(t, cmp.Diff(totalSupplyResp, gotTotalSupply))
}

func TestQueryBankDenomsMetadata(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	denomsResp := banktypes.QueryDenomsMetadataResponse{
		Metadatas: []banktypes.Metadata{
			{
				Description: "atoms",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    "uatom",
						Exponent: 0,
						Aliases:  []string{},
					},
					{
						Denom:    "atom",
						Exponent: 6,
						Aliases:  []string{},
					},
				},
				Name:   "Cosmos Atom",
				Symbol: "ATOM",
			},
		},
	}
	mc := new(mocks.Client)
	MockQueryDenomsMetadata(
		t, mc,
		&banktypes.QueryDenomsMetadataRequest{
			Pagination: &query.PageRequest{Limit: 100}, // Default pagination setting.
		},
		&denomsResp,
	)
	sys.OverrideClients("cosmoshub", cmd.ClientOverrides{
		RPCClient: mc,
	})

	// Querying the denom metadata shows nothing on stderr, serialized response on stdout.
	res := sys.MustRun(t, "query", "bank", "denoms-metadata")
	require.Empty(t, res.Stderr.String())

	var gotDenoms banktypes.QueryDenomsMetadataResponse
	require.NoError(t, json.Unmarshal(res.Stdout.Bytes(), &gotDenoms))

	require.Empty(t, cmp.Diff(denomsResp, gotDenoms))
}
