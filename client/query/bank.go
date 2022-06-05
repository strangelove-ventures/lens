package query

import (
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// bank_ParamsRPC returns the distribution params
func bank_ParamsRPC(q *Query) (*bankTypes.QueryParamsResponse, error) {
	req := &bankTypes.QueryParamsRequest{}
	queryClient := bankTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.Params(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// bank_BalanceRPC returns the balance of specified denom coins for a single account.
func bank_BalanceRPC(q *Query, address string, denom string) (*bankTypes.QueryBalanceResponse, error) {
	req := &bankTypes.QueryBalanceRequest{Address: address, Denom: denom}
	queryClient := bankTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.Balance(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// bank_AllBalancesRPC returns the balance of all coins for a single account.
func bank_AllBalancesRPC(q *Query, address string) (*bankTypes.QueryAllBalancesResponse, error) {
	req := &bankTypes.QueryAllBalancesRequest{Address: address, Pagination: q.Options.Pagination}
	queryClient := bankTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.AllBalances(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// bank_SupplyOfRPC returns the supply of all coins
func bank_SupplyOfRPC(q *Query, denom string) (*bankTypes.QuerySupplyOfResponse, error) {
	req := &bankTypes.QuerySupplyOfRequest{Denom: denom}
	queryClient := bankTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.SupplyOf(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// bank_TotalSupplyRPC returns the supply of all coins
func bank_TotalSupplyRPC(q *Query) (*bankTypes.QueryTotalSupplyResponse, error) {
	req := &bankTypes.QueryTotalSupplyRequest{Pagination: q.Options.Pagination}
	queryClient := bankTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.TotalSupply(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// bank_DenomMetadataRPC returns the metadata for given denom
func bank_DenomMetadataRPC(q *Query, denom string) (*bankTypes.QueryDenomMetadataResponse, error) {
	req := &bankTypes.QueryDenomMetadataRequest{Denom: denom}
	queryClient := bankTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.DenomMetadata(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// bank_DenomsMetadataRPC returns the metadata for all denoms
func bank_DenomsMetadataRPC(q *Query) (*bankTypes.QueryDenomsMetadataResponse, error) {
	req := &bankTypes.QueryDenomsMetadataRequest{Pagination: q.Options.Pagination}
	queryClient := bankTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.DenomsMetadata(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
