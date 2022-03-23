package query

import (
	"encoding/hex"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

// blockRPC returns information about a block
func blockRPC(q *Query) (*coretypes.ResultBlock, error) {
	var height int64
	ctx, cancel := q.GetQueryContext()
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

// blockByHashRPC returns information about a block by hash
func blockByHashRPC(q *Query, hash string) (*coretypes.ResultBlock, error) {
	ctx, cancel := q.GetQueryContext()
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

// blockResultsRPC returns information about a block by hash
func blockResultsRPC(q *Query) (*coretypes.ResultBlockResults, error) {
	var height int64
	ctx, cancel := q.GetQueryContext()
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

// statusRPC returns information about a node status
func statusRPC(q *Query) (*coretypes.ResultStatus, error) {
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := q.Client.RPCClient.Status(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// abciInfoRPC returns information about the ABCI application
func abciInfoRPC(q *Query) (*coretypes.ResultABCIInfo, error) {
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := q.Client.RPCClient.ABCIInfo(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// abciQueryRPC returns data from a particular path in the ABCI application
func abciQueryRPC(q *Query, path string, data string, prove bool) (*coretypes.ResultABCIQuery, error) {
	ctx, cancel := q.GetQueryContext()
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
