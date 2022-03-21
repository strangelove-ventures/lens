package cmd_test

import (
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/client/mocks"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
)

var protoCodec = encoding.GetCodec(proto.Name)

func MockQueryAllBalances(t *testing.T, mc *mocks.Client, req *banktypes.QueryAllBalancesRequest, resp *banktypes.QueryAllBalancesResponse) {
	t.Helper()

	// Encode the query account request to match the mock's expected input.
	reqBz, err := protoCodec.Marshal(req)
	require.NoError(t, err)

	// Serialize the query account response.
	respBz, err := protoCodec.Marshal(resp)
	require.NoError(t, err)

	mc.On(
		"ABCIQueryWithOptions",
		mock.Anything, // Some context.Context value.
		"/cosmos.bank.v1beta1.Query/AllBalances",
		bytes.HexBytes(reqBz),
		client.ABCIQueryOptions{}, // TODO: support specific query options?
	).Return(&coretypes.ResultABCIQuery{
		Response: abci.ResponseQuery{
			Value: respBz,
		},
	}, nil)
}

func MockQueryTotalSupply(t *testing.T, mc *mocks.Client, req *banktypes.QueryTotalSupplyRequest, resp *banktypes.QueryTotalSupplyResponse) {
	t.Helper()

	// Encode the query total supply request to match the mock's expected input.
	reqBz, err := protoCodec.Marshal(req)
	require.NoError(t, err)

	// Serialize the query total supply response.
	respBz, err := protoCodec.Marshal(resp)
	require.NoError(t, err)

	mc.On(
		"ABCIQueryWithOptions",
		mock.Anything, // Some context.Context value.
		"/cosmos.bank.v1beta1.Query/TotalSupply",
		bytes.HexBytes(reqBz),
		client.ABCIQueryOptions{}, // TODO: support specific query options?
	).Return(&coretypes.ResultABCIQuery{
		Response: abci.ResponseQuery{
			Value: respBz,
		},
	}, nil)
}

func MockQueryDenomsMetadata(t *testing.T, mc *mocks.Client, req *banktypes.QueryDenomsMetadataRequest, resp *banktypes.QueryDenomsMetadataResponse) {
	t.Helper()

	// Encode the denoms metadata request to match the mock's expected input.
	reqBz, err := protoCodec.Marshal(req)
	require.NoError(t, err)

	// Serialize the denoms metadata response.
	respBz, err := protoCodec.Marshal(resp)
	require.NoError(t, err)

	mc.On(
		"ABCIQueryWithOptions",
		mock.Anything, // Some context.Context value.
		"/cosmos.bank.v1beta1.Query/DenomsMetadata",
		bytes.HexBytes(reqBz),
		client.ABCIQueryOptions{}, // TODO: support specific query options?
	).Return(&coretypes.ResultABCIQuery{
		Response: abci.ResponseQuery{
			Value: respBz,
		},
	}, nil)
}
