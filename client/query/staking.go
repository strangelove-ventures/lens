package query

import (
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// delegationRPC returns the delegations to a particular validator
func delegationRPC(q *Query, delegator, validator string) (*stakingTypes.DelegationResponse, error) {
	queryClient := stakingTypes.NewQueryClient(q.Client)
	params := &stakingTypes.QueryDelegationRequest{
		DelegatorAddr: delegator,
		ValidatorAddr: validator,
	}
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.Delegation(ctx, params)
	if err != nil {
		return nil, err
	}

	return res.DelegationResponse, nil
}

// delegationsRPC returns all the delegations
func delegationsRPC(q *Query, delegator string) (*stakingTypes.QueryDelegatorDelegationsResponse, error) {
	queryClient := stakingTypes.NewQueryClient(q.Client)
	params := &stakingTypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegator,
		Pagination:    q.Options.Pagination,
	}
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.DelegatorDelegations(ctx, params)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// validatorDelegationssRPC returns all the delegations for a validator
func validatorDelegationssRPC(q *Query, validator string) (*stakingTypes.QueryValidatorDelegationsResponse, error) {
	// ensure the validator parameter is a valid validator address
	_, err := q.Client.DecodeBech32ValAddr(validator)
	if err != nil {
		return nil, err
	}
	queryClient := stakingTypes.NewQueryClient(q.Client)
	params := &stakingTypes.QueryValidatorDelegationsRequest{
		ValidatorAddr: validator,
		Pagination:    q.Options.Pagination,
	}
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.ValidatorDelegations(ctx, params)
	if err != nil {
		return nil, err
	}

	return res, nil
}
