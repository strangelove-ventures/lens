package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/mempool"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	defaultBroadcastWaitTimeout = 10 * time.Minute
)

func (cc *ChainClient) BroadcastTx(ctx context.Context, tx []byte) (*sdk.TxResponse, error) {
	var (
		blockTimeout time.Duration = defaultBroadcastWaitTimeout
		err          error
	)

	if cc.Config.BlockTimeout != "" {
		blockTimeout, err = time.ParseDuration(cc.Config.BlockTimeout)
		if err != nil {
			// Did you call Validate() method on ChainClientConfig struct
			// before coming here?
			return nil, err
		}
	}

	return broadcastTx(
		ctx,
		cc.RPCClient,
		cc.Codec.TxConfig.TxDecoder(),
		tx,
		blockTimeout,
	)
}

type rpcTxBroadcaster interface {
	Tx(ctx context.Context, hash []byte, prove bool) (*ctypes.ResultTx, error)
	BroadcastTxSync(context.Context, tmtypes.Tx) (*ctypes.ResultBroadcastTx, error)

	// TODO: implement commit and async as well
	// BroadcastTxCommit(context.Context, tmtypes.Tx) (*ctypes.ResultBroadcastTxCommit, error)
	// BroadcastTxAsync(context.Context, tmtypes.Tx) (*ctypes.ResultBroadcastTx, error)
}

// broadcastTx broadcasts a TX and then waits for the TX to be included in the block.
// The waiting will either be canceled after the waitTimeout has run out or the context
// exited.
func broadcastTx(
	ctx context.Context,
	broadcaster rpcTxBroadcaster,
	txDecoder sdk.TxDecoder,
	tx []byte,
	waitTimeout time.Duration,
) (*sdk.TxResponse, error) {
	// broadcast tx sync waits for check tx to pass
	// NOTE: this can return w/ a timeout
	// need to investigate if this will leave the tx
	// in the mempool or we can retry the broadcast at that
	// point
	syncRes, err := broadcaster.BroadcastTxSync(ctx, tx)
	if err != nil {
		errRes := CheckTendermintError(err, tx)
		if errRes != nil {
			return errRes, nil
		}
		return nil, err
	}

	// TODO: maybe we need to check if the node has tx indexing enabled?
	// if not, we need to find a new way to block until inclusion in a block

	// wait for tx to be included in a block
	exitAfter := time.After(waitTimeout)
	for {
		select {
		case <-exitAfter:
			return nil, fmt.Errorf("timed out after: %d; %w", waitTimeout, ErrTimeoutAfterWaitingForTxBroadcast)
		// TODO: this is potentially less than optimal and may
		// be better as something configurable
		case <-time.After(time.Millisecond * 100):
			resTx, err := broadcaster.Tx(ctx, syncRes.Hash, false)
			if err == nil {
				return mkTxResult(txDecoder, resTx)
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func mkTxResult(txDecoder sdk.TxDecoder, resTx *ctypes.ResultTx) (*sdk.TxResponse, error) {
	txb, err := txDecoder(resTx.Tx)
	if err != nil {
		return nil, err
	}
	p, ok := txb.(intoAny)
	if !ok {
		return nil, fmt.Errorf("expecting a type implementing intoAny, got: %T", txb)
	}
	any := p.AsAny()
	// TODO: maybe don't make up the time here?
	// we can fetch the block for the block time buts thats
	// more round trips
	// TODO: logs get rendered as base64 encoded, need to fix this somehow
	return sdk.NewResponseResultTx(resTx, any, time.Now().Format(time.RFC3339)), nil
}

// Deprecated: this interface is used only internally for scenario we are
// deprecating (StdTxConfig support)
type intoAny interface {
	AsAny() *codectypes.Any
}

// CheckTendermintError checks if the error returned from BroadcastTx is a
// Tendermint error that is returned before the tx is submitted due to
// precondition checks that failed. If an Tendermint error is detected, this
// function returns the correct code back in TxResponse.
//
// TODO: Avoid brittle string matching in favor of error matching. This requires
// a change to Tendermint's RPCError type to allow retrieval or matching against
// a concrete error type.
func CheckTendermintError(err error, tx tmtypes.Tx) *sdk.TxResponse {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())
	txHash := fmt.Sprintf("%X", tx.Hash())

	switch {
	case strings.Contains(errStr, strings.ToLower(mempool.ErrTxInCache.Error())):
		return &sdk.TxResponse{
			Code:      sdkerrors.ErrTxInMempoolCache.ABCICode(),
			Codespace: sdkerrors.ErrTxInMempoolCache.Codespace(),
			TxHash:    txHash,
		}

	case strings.Contains(errStr, "mempool is full"):
		return &sdk.TxResponse{
			Code:      sdkerrors.ErrMempoolIsFull.ABCICode(),
			Codespace: sdkerrors.ErrMempoolIsFull.Codespace(),
			TxHash:    txHash,
		}

	case strings.Contains(errStr, "tx too large"):
		return &sdk.TxResponse{
			Code:      sdkerrors.ErrTxTooLarge.ABCICode(),
			Codespace: sdkerrors.ErrTxTooLarge.Codespace(),
			TxHash:    txHash,
		}
	default:
		// More error debugging here!!
		return nil
	}
}
