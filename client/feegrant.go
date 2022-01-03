package client

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	feetypes "github.com/cosmos/cosmos-sdk/x/feegrant"
)

func (cc *ChainClient) QueryFeeGrants(address sdk.AccAddress, pageReq *querytypes.PageRequest) (*feetypes.QueryAllowancesResponse, error) {
	addr, err := cc.EncodeBech32AccAddr(address)
	if err != nil {
		return nil, err
	}

	request := feetypes.QueryAllowancesRequest{
		Grantee:    addr,
		Pagination: pageReq,
	}

	res, err := feetypes.NewQueryClient(cc).Allowances(context.Background(), &request)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (cc *ChainClient) QueryFeeGrant(granteeAddress sdk.AccAddress, granterAddress sdk.AccAddress) (*feetypes.QueryAllowanceResponse, error) {
	granteeAddr, err := cc.EncodeBech32AccAddr(granteeAddress)
	if err != nil {
		return nil, err
	}

	granterAddr, err := cc.EncodeBech32AccAddr(granterAddress)
	if err != nil {
		return nil, err
	}

	request := feetypes.QueryAllowanceRequest{
		Grantee: granteeAddr,
		Granter: granterAddr,
	}

	res, err := feetypes.NewQueryClient(cc).Allowance(context.Background(), &request)

	if err != nil {
		return nil, err
	}

	return res, nil
}
