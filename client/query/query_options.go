package query

import (
	"context"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/metadata"
	"strconv"
	"time"
)

type QueryOptions struct {
	Pagination *query.PageRequest
	Height     int64
}

func DefaultOptions() *QueryOptions {
	return &QueryOptions{
		Pagination: &query.PageRequest{
			Key:        []byte(""),
			Offset:     0,
			Limit:      1000,
			CountTotal: true,
		},
		Height: 0,
	}
}

// GetQueryContext returns a context that includes the height and uses the timeout from the config
func (q *Query) GetQueryContext() (context.Context, context.CancelFunc) {
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	strHeight := strconv.Itoa(int(q.Options.Height))
	ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	return ctx, cancel
}
