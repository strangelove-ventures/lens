package query

import (
	"context"
	"strconv"
	"time"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/metadata"
)

// Delegation returns the delegations to a particular validator
func Delegation(q *Query, delegator, validator string) (*stakingTypes.DelegationResponse, error) {
	queryClient := stakingTypes.NewQueryClient(q.Client)
	params := &stakingTypes.QueryDelegationRequest{
		DelegatorAddr: delegator,
		ValidatorAddr: validator,
	}
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	strHeight := strconv.Itoa(int(q.Options.Height))
	ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	defer cancel()
	res, err := queryClient.Delegation(ctx, params)
	if err != nil {
		return nil, err
	}

	return res.DelegationResponse, nil
}

// Delegations returns all the delegations
func Delegations(q *Query, delegator string) (*stakingTypes.QueryDelegatorDelegationsResponse, error) {
	queryClient := stakingTypes.NewQueryClient(q.Client)
	params := &stakingTypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegator,
		Pagination:    q.Options.Pagination,
	}
	timeout, _ := time.ParseDuration(q.Client.Config.Timeout) // Timeout is validated in the config so no error check
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	strHeight := strconv.Itoa(int(q.Options.Height))
	ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	defer cancel()
	res, err := queryClient.DelegatorDelegations(ctx, params)
	if err != nil {
		return nil, err
	}

	return res, nil
}
