package query

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
)

// BlockRPC returns information about a block
func BlockRPC(q *Query) (*coretypes.ResultBlock, error) {
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

// BlockByHashRPC returns information about a block by hash
func BlockByHashRPC(q *Query, hash string) (*coretypes.ResultBlock, error) {
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

// BlockResultsRPC returns information about a block by hash
func BlockResultsRPC(q *Query) (*coretypes.ResultBlockResults, error) {
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

// StatusRPC returns information about a node status
func StatusRPC(q *Query) (*coretypes.ResultStatus, error) {
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := q.Client.RPCClient.Status(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ABCIInfoRPC returns information about the ABCI application
func ABCIInfoRPC(q *Query) (*coretypes.ResultABCIInfo, error) {
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := q.Client.RPCClient.ABCIInfo(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ABCIQueryRPC returns data from a particular path in the ABCI application
func ABCIQueryRPC(q *Query, path string, data string, prove bool) (*coretypes.ResultABCIQuery, error) {
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

// QueryTxs returns an results of a TxSearch for a given tag
func TxsRPC(q *Query, events []string) (*coretypes.ResultTxSearch, error) {
	if len(events) == 0 {
		return nil, errors.New("must declare at least one event to search")
	}

	page := int(q.Options.Pagination.Offset/q.Options.Pagination.Limit) + 1 // page is 1-indexed, not 0-indexed
	limit := int(q.Options.Pagination.Limit)

	res, err := q.Client.RPCClient.TxSearch(context.Background(), strings.Join(events, " AND "), true, &page, &limit, "")
	if err != nil {
		return nil, err
	}

	return res, nil
}
