package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/avast/retry-go"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/mempool"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// BroadcastTx broadcasts a transactions either synchronously or asynchronously
// based on the context parameters. The result of the broadcast is parsed into
// an intermediate structure which is logged if the context has a logger
// defined.
func (cc *ChainClient) BroadcastTx(ctx context.Context, tx []byte) (res *sdk.TxResponse, err error) {
	// broadcast tx sync waits for check tx to pass
	// this will give us errors from
	res, err = cc.BroadcastTxSync(ctx, tx)
	// TODO: ensure we checking response error codes here
	// and not just the error
	if err != nil {
		return
	}

	for {
		// TODO: wire up context cancellation
		if cc.TxInMempool(ctx, res.TxHash) {
			// break for loop once tx isn't in mempool
			break
		}
		// TODO: make this configurable or just run it with no breaks or fewer breaks
		time.Sleep(time.Millisecond * 100)
	}

	// TODO: add option to return early here?

	txid, err := hex.DecodeString(res.TxHash)
	if err != nil {
		return
	}
	var resTx *ctypes.ResultTx
	if err = retry.Do(func() error {
		resTx, err = cc.RPCClient.Tx(ctx, txid, false)
		// TODO: return error if tx is not found
		return err
		// TODO: retry options here
	}); err != nil {
		return
	}
	return cc.mkTxResult(resTx)
}

func (cc *ChainClient) mkTxResult(resTx *ctypes.ResultTx) (*sdk.TxResponse, error) {
	txb, err := cc.Codec.TxConfig.TxDecoder()(resTx.Tx)
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
	return sdk.NewResponseResultTx(resTx, any, time.Now().Format(time.RFC3339)), nil
}

// Deprecated: this interface is used only internally for scenario we are
// deprecating (StdTxConfig support)
type intoAny interface {
	AsAny() *codectypes.Any
}

func (cc *ChainClient) TxInMempool(ctx context.Context, txHash string) bool {
	limit := 1000
	// TODO: maybe retry this on error?
	// would do this in the case of inconsistent errors from TM
	res, err := cc.RPCClient.UnconfirmedTxs(ctx, &limit)
	if err != nil {
		return false
	}
	for _, txbz := range res.Txs {
		if strings.EqualFold(txHash, fmt.Sprintf("%X", txbz.Hash())) {
			return true
		}
	}
	return false
}

func (cc *ChainClient) BroadcastTxSync(ctx context.Context, tx []byte) (*sdk.TxResponse, error) {
	res, err := cc.RPCClient.BroadcastTxSync(ctx, tx)
	if errRes := CheckTendermintError(err, tx); errRes != nil {
		return errRes, nil
	}

	return sdk.NewResponseFormatBroadcastTx(res), err

}

// func (cc *ChainClient) BroadcastTxAsync(ctx context.Context, tx []byte) (*sdk.TxResponse, error) {
// 	res, err := cc.RPCClient.BroadcastTxAsync(ctx, tx)
// 	if errRes := CheckTendermintError(err, tx); errRes != nil {
// 		return errRes, nil
// 	}

// 	return sdk.NewResponseFormatBroadcastTx(res), err
// }

// func (cc *ChainClient) BroadcastTxCommit(ctx context.Context, txBytes []byte) (*sdk.TxResponse, error) {
// 	res, err := cc.RPCClient.BroadcastTxCommit(ctx, txBytes)
// 	// TODO: why this? need to figure that one out
// 	if err == nil {
// 		return sdk.NewResponseFormatBroadcastTxCommit(res), nil
// 	}
// 	if errRes := CheckTendermintError(err, txBytes); errRes != nil {
// 		return errRes, nil
// 	}
// 	return sdk.NewResponseFormatBroadcastTxCommit(res), err
// }

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
