package staking

import (
	"context"

	"github.com/strangelove-ventures/lens/client"

	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func QueryDelegation(chainClient *client.ChainClient, delegator, validator string) (*types.QueryDelegationResponse, error) {
	queryClient := types.NewQueryClient(chainClient)
	params := &types.QueryDelegationRequest{
		DelegatorAddr: delegator,
		ValidatorAddr: validator,
	}

	res, err := queryClient.Delegation(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func QueryDelegations(chainClient *client.ChainClient, delegator string, pageReq *querytypes.PageRequest) (*types.QueryDelegatorDelegationsResponse, error) {
	queryClient := types.NewQueryClient(chainClient)
	params := &types.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegator,
		Pagination:    pageReq,
	}

	res, err := queryClient.DelegatorDelegations(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return res, nil
}
