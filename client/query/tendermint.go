package query

import (
	"context"
	"encoding/hex"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"time"
)

// BlockRPC returns information about a block
func BlockRPC(q *Query) (*coretypes.ResultBlock, error) {
	var height int64
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// If height is not specified, default value is 0, query the latest available block then
	if q.Options.Height == 0 {
		resStatus, err := q.Client.RPCClient.Status(ctx)
		if err != nil {
			return nil, err
		}
		height = resStatus.SyncInfo.LatestBlockHeight
	} else {
		height = q.Options.Height
	}
	res, err := q.Client.RPCClient.Block(ctx, &height)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// BlockByHashRPC returns information about a block by hash
func BlockByHashRPC(q *Query, hash string) (*coretypes.ResultBlock, error) {
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	h, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}
	res, err := q.Client.RPCClient.BlockByHash(ctx, h)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// BlockResultsRPC returns information about a block by hash
func BlockResultsRPC(q *Query) (*coretypes.ResultBlockResults, error) {
	var height int64
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// If height is not specified, default value is 0, query the latest available block then
	if q.Options.Height == 0 {
		resStatus, err := q.Client.RPCClient.Status(ctx)
		if err != nil {
			return nil, err
		}
		height = resStatus.SyncInfo.LatestBlockHeight
	} else {
		height = q.Options.Height
	}
	res, err := q.Client.RPCClient.BlockResults(ctx, &height)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// StatusRPC returns information about a node status
func StatusRPC(q *Query) (*coretypes.ResultStatus, error) {
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	res, err := q.Client.RPCClient.Status(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ABCIInfoRPC returns information about the ABCI application
func ABCIInfoRPC(q *Query) (*coretypes.ResultABCIInfo, error) {
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	res, err := q.Client.RPCClient.ABCIInfo(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ABCIQueryRPC returns data from a particular path in the ABCI application
func ABCIQueryRPC(q *Query, path string, data string, prove bool) (*coretypes.ResultABCIQuery, error) {
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// If height is not specified, default value is 0, query the latest available block then
	opts := rpcclient.ABCIQueryOptions{
		Height: q.Options.Height,
		Prove:  prove,
	}
	res, err := q.Client.RPCClient.ABCIQueryWithOptions(ctx, path, []byte(data), opts)
	if err != nil {
		return nil, err
	}

	return res, nil
}
