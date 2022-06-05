package query

import (
	distTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// distribution_ParamsRPC returns the distribution params
func distribution_ParamsRPC(q *Query) (*distTypes.QueryParamsResponse, error) {
	req := &distTypes.QueryParamsRequest{}
	queryClient := distTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.Params(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// distribution_ValidatorSlashesRPC returns slash events for a given validator
func distribution_ValidatorSlashesRPC(q *Query, address string, start_height uint64, end_height uint64) (*distTypes.QueryValidatorSlashesResponse, error) {
	req := &distTypes.QueryValidatorSlashesRequest{ValidatorAddress: address, StartingHeight: start_height, EndingHeight: end_height}
	queryClient := distTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.ValidatorSlashes(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// distribution_DelegatorValidatorsRPC returns the validators of a delegator
func distribution_DelegatorValidatorsRPC(q *Query, address string) (*distTypes.QueryDelegatorValidatorsResponse, error) {
	req := &distTypes.QueryDelegatorValidatorsRequest{DelegatorAddress: address}
	queryClient := distTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.DelegatorValidators(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// distribution_DelegationRewardsRPC returns rewards for a single delegator/validator tuple
func distribution_DelegationRewardsRPC(q *Query, delegator string, validator string) (*distTypes.QueryDelegationRewardsResponse, error) {
	req := &distTypes.QueryDelegationRewardsRequest{DelegatorAddress: delegator, ValidatorAddress: validator}
	queryClient := distTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.DelegationRewards(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// distribution_DelegationTotalRewardsRPC returns total outstanding rewards for a delegator across one or more validators
func distribution_DelegationTotalRewardsRPC(q *Query, address string) (*distTypes.QueryDelegationTotalRewardsResponse, error) {
	req := &distTypes.QueryDelegationTotalRewardsRequest{DelegatorAddress: address}
	queryClient := distTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.DelegationTotalRewards(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// distribution_ValidatorCommissionRPC returns outstanding commission for a validator
func distribution_ValidatorCommissionRPC(q *Query, address string) (*distTypes.QueryValidatorCommissionResponse, error) {
	req := &distTypes.QueryValidatorCommissionRequest{ValidatorAddress: address}
	queryClient := distTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.ValidatorCommission(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// distribution_ValidatorOutstandingRewardsRPC returns total outstanding reward pool
func distribution_ValidatorOutstandingRewardsRPC(q *Query, address string) (*distTypes.QueryValidatorOutstandingRewardsResponse, error) {
	req := &distTypes.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: address}
	queryClient := distTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.ValidatorOutstandingRewards(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// distribution_DelegatorWithdrawAddressRPC returns withdrawal address for given delegator
func distribution_DelegatorWithdrawAddressRPC(q *Query, address string) (*distTypes.QueryDelegatorWithdrawAddressResponse, error) {
	req := &distTypes.QueryDelegatorWithdrawAddressRequest{DelegatorAddress: address}
	queryClient := distTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.DelegatorWithdrawAddress(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// distribution_CommunityPoolRPC returns balance of community pool
func distribution_CommunityPoolRPC(q *Query) (*distTypes.QueryCommunityPoolResponse, error) {
	req := &distTypes.QueryCommunityPoolRequest{}
	queryClient := distTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.CommunityPool(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
