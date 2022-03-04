package query

import (
	"context"
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
