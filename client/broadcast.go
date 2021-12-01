package client

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/mempool"
	tmtypes "github.com/tendermint/tendermint/types"
)

// BroadcastTx broadcasts a transactions either synchronously or asynchronously
// based on the context parameters. The result of the broadcast is parsed into
// an intermediate structure which is logged if the context has a logger
// defined.
func (cc *ChainClient) BroadcastTx(ctx context.Context, tx []byte) (res *sdk.TxResponse, err error) {
	switch cc.Config.BroadcastMode {
	case "sync":
		res, err = cc.BroadcastTxSync(ctx, tx)
	case "async":
		res, err = cc.BroadcastTxAsync(ctx, tx)
	case "block":
		res, err = cc.BroadcastTxCommit(ctx, tx)
	default:
		return nil, fmt.Errorf("unsupported return type %s; supported types: sync, async, block", cc.Config.BroadcastMode)
	}
	return res, err
}

func (cc *ChainClient) BroadcastTxSync(ctx context.Context, tx []byte) (*sdk.TxResponse, error) {
	res, err := cc.RPCClient.BroadcastTxSync(ctx, tx)
	if errRes := CheckTendermintError(err, tx); errRes != nil {
		return errRes, nil
	}

	return sdk.NewResponseFormatBroadcastTx(res), err

}

func (cc *ChainClient) BroadcastTxAsync(ctx context.Context, tx []byte) (*sdk.TxResponse, error) {
	res, err := cc.RPCClient.BroadcastTxAsync(ctx, tx)
	if errRes := CheckTendermintError(err, tx); errRes != nil {
		return errRes, nil
	}

	return sdk.NewResponseFormatBroadcastTx(res), err
}

func (cc *ChainClient) BroadcastTxCommit(ctx context.Context, txBytes []byte) (*sdk.TxResponse, error) {
	res, err := cc.RPCClient.BroadcastTxCommit(ctx, txBytes)
	// TODO: why this? need to figure that one out
	if err == nil {
		return sdk.NewResponseFormatBroadcastTxCommit(res), nil
	}
	if errRes := CheckTendermintError(err, txBytes); errRes != nil {
		return errRes, nil
	}
	return sdk.NewResponseFormatBroadcastTxCommit(res), err
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
