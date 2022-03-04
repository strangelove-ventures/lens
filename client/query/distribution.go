package query

import (
	"context"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	distTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"google.golang.org/grpc/metadata"
	"strconv"
	"time"
)

// DelegatorValidatorsRPC returns the validators of a delegator
func DelegatorValidatorsRPC(q *Query, address string) (*distTypes.QueryDelegatorValidatorsResponse, error) {
	var req *distTypes.QueryDelegatorValidatorsRequest
	req = &distTypes.QueryDelegatorValidatorsRequest{DelegatorAddress: address}
	queryClient := distTypes.NewQueryClient(q.Client)
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	strHeight := strconv.Itoa(int(q.Options.Height))
	ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	defer cancel()
	res, err := queryClient.DelegatorValidators(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
