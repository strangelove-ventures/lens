package query

import (
	distTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// delegatorValidatorsRPC returns the validators of a delegator
func delegatorValidatorsRPC(q *Query, address string) (*distTypes.QueryDelegatorValidatorsResponse, error) {
	var req *distTypes.QueryDelegatorValidatorsRequest
	req = &distTypes.QueryDelegatorValidatorsRequest{DelegatorAddress: address}
	queryClient := distTypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.DelegatorValidators(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
