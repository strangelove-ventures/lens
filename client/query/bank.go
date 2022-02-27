package query

import (
	"context"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"google.golang.org/grpc/metadata"
	"strconv"
	"time"
)

// BalanceWithAddressRPC returns the balance of all coins for a single account.
func BalanceWithAddressRPC(q *Query, address string) (*bankTypes.QueryAllBalancesResponse, error) {
	var req *bankTypes.QueryAllBalancesRequest
	req = &bankTypes.QueryAllBalancesRequest{Address: address, Pagination: q.Options.Pagination}
	queryClient := bankTypes.NewQueryClient(q.Client)
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	strHeight := strconv.Itoa(int(q.Options.Height))
	ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	defer cancel()
	res, err := queryClient.AllBalances(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// TotalSupplyRPC returns the supply of all coins
func TotalSupplyRPC(q *Query) (*bankTypes.QueryTotalSupplyResponse, error) {
	req := &bankTypes.QueryTotalSupplyRequest{Pagination: q.Options.Pagination}
	queryClient := bankTypes.NewQueryClient(q.Client)
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	strHeight := strconv.Itoa(int(q.Options.Height))
	ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	defer cancel()
	res, err := queryClient.TotalSupply(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// DenomsMetadataRPC returns the metadata for all denoms
func DenomsMetadataRPC(q *Query) (*bankTypes.QueryDenomsMetadataResponse, error) {
	req := &bankTypes.QueryDenomsMetadataRequest{Pagination: q.Options.Pagination}
	queryClient := bankTypes.NewQueryClient(q.Client)
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	strHeight := strconv.Itoa(int(q.Options.Height))
	ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	defer cancel()
	res, err := queryClient.DenomsMetadata(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
