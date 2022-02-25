package query

import (
	"context"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Delegation returns the delegations to a particular validator
func Delegation(q *Query, delegator, validator string) (*types.DelegationResponse, error) {
	queryClient := types.NewQueryClient(q.Client)
	params := &types.QueryDelegationRequest{
		DelegatorAddr: delegator,
		ValidatorAddr: validator,
	}

	res, err := queryClient.Delegation(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return res.DelegationResponse, nil
}

// Delegations returns all the delegations
func Delegations(q *Query, delegator string) (types.DelegationResponses, error) {
	queryClient := types.NewQueryClient(q.Client)
	params := &types.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegator,
		Pagination:    q.Options.Pagination,
	}

	res, err := queryClient.DelegatorDelegations(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return res.DelegationResponses, nil
}
