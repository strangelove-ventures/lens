package cmd_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/strangelove-ventures/lens/cmd"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/rpc/client/mocks"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/types"
)

func TestTendermintBlock_SpecificHeight(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	mockBlock := coretypes.ResultBlock{
		BlockID: types.BlockID{
			Hash: bytes.HexBytes{0, 0, 0, 0},
			PartSetHeader: types.PartSetHeader{
				Total: 1234,
				Hash:  bytes.HexBytes{1, 1, 1, 1},
			},
		},
		// Partially populated block just to confirm data flows through.
		Block: &types.Block{
			Header: types.Header{
				ChainID: "cosmoshub",
				Height:  100,
				Time:    time.Now().Add(-time.Second),
			},
		},
	}
	mc := new(mocks.Client)
	expHeight := new(int64)
	*expHeight = 100
	mc.On("Block", mock.Anything, expHeight).Return(&mockBlock, nil)

	sys.OverrideClients("cosmoshub", cmd.ClientOverrides{
		RPCClient: mc,
	})

	// tm status prints the received block as JSON.
	// Nothing should output on stderr.
	res := sys.MustRun(t, "tendermint", "block", "--height=100")
	require.Empty(t, res.Stderr.String())

	var gotBlock coretypes.ResultBlock
	require.NoError(t, json.Unmarshal(res.Stdout.Bytes(), &gotBlock))

	require.Empty(
		t,
		cmp.Diff(
			mockBlock,
			gotBlock,
			cmpopts.IgnoreUnexported(types.Block{}, types.Data{}, types.EvidenceData{}),
			cmpopts.EquateEmpty(),
		),
	)
}

func TestTendermintTx(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	mockTx := coretypes.ResultTx{
		Hash:   bytes.HexBytes{0x12, 0x34},
		Height: 100,
		Index:  5,
		Tx:     []byte("some tx"),
	}
	mc := new(mocks.Client)
	mc.On("Tx", mock.Anything, []byte{0x12, 0x34}, false).Return(&mockTx, nil)

	sys.OverrideClients("cosmoshub", cmd.ClientOverrides{
		RPCClient: mc,
	})

	// tm status prints the received tx as JSON.
	// Nothing should output on stderr.
	res := sys.MustRun(t, "tendermint", "tx", "1234")
	require.Empty(t, res.Stderr.String())

	var gotTx coretypes.ResultTx
	require.NoError(t, json.Unmarshal(res.Stdout.Bytes(), &gotTx))

	require.Empty(t, cmp.Diff(mockTx, gotTx, cmpopts.EquateEmpty()))
}

func TestTendermintStatus(t *testing.T) {
	t.Parallel()

	sys := NewSystem(t)

	// Arbitrary status response with a few fields filled in.
	mockStatus := coretypes.ResultStatus{
		NodeInfo: p2p.DefaultNodeInfo{
			Moniker: "foo bar",
		},
		SyncInfo: coretypes.SyncInfo{
			LatestBlockHeight: 123,
		},
		ValidatorInfo: coretypes.ValidatorInfo{
			VotingPower: 5,
		},
	}
	mc := new(mocks.Client)
	mc.On("Status", mock.Anything).Return(&mockStatus, nil)

	sys.OverrideClients("cosmoshub", cmd.ClientOverrides{
		RPCClient: mc,
	})

	// tm status prints the received status as JSON.
	// Nothing should output on stderr.
	res := sys.MustRun(t, "tendermint", "status")
	require.Empty(t, res.Stderr.String())

	var gotStatus coretypes.ResultStatus
	require.NoError(t, json.Unmarshal(res.Stdout.Bytes(), &gotStatus))

	require.Empty(t, cmp.Diff(mockStatus, gotStatus, cmpopts.EquateEmpty()))
}
