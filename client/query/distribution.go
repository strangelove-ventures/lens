package query

import (
	distTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// DistributionParamsRPC returns the distribution params
func DistributionParamsRPC(q *Query, address string) (*distTypes.QueryParamsResponse, error) {
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

// ValidatorSlashes returns slash events for a given validator
func ValidatorSlashesRPC(q *Query, address string, start_height uint64, end_height uint64) (*distTypes.QueryValidatorSlashesResponse, error) {
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

// DelegatorValidatorsRPC returns the validators of a delegator
func DelegatorValidatorsRPC(q *Query, address string) (*distTypes.QueryDelegatorValidatorsResponse, error) {
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

// DelegationRewardsRequestRPC returns rewards for a single delegator/validator tuple
func DelegationRewardsRequestRPC(q *Query, delegator string, validator string) (*distTypes.QueryDelegationRewardsResponse, error) {
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

// DelegationRewardsRequestRPC returns total outstanding rewards for a delegator across one or more validators
func DelegationTotalRewardsRPC(q *Query, address string) (*distTypes.QueryDelegationTotalRewardsResponse, error) {
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

// ValidatorCommissionRPC returns outstanding commission for a validator
func ValidatorCommissionRPC(q *Query, address string) (*distTypes.QueryValidatorCommissionResponse, error) {
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

// ValidatorOutstandingRewardsRPC returns total outstanding reward pool
func ValidatorOutstandingRewardsRPC(q *Query, address string) (*distTypes.QueryValidatorOutstandingRewardsResponse, error) {
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

// DelegatorWithdrawAddressRPC returns withdrawal address for given delegator
func DelegatorWithdrawAddressRPC(q *Query, address string) (*distTypes.QueryDelegatorWithdrawAddressResponse, error) {
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

// CommunityPoolRPC returns balance of community pool
func CommunityPoolRPC(q *Query) (*distTypes.QueryCommunityPoolResponse, error) {
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
