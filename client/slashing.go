package client

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func (cc *ChainClient) QuerySlashingSigningInfo(publicKey cryptotypes.PubKey) (*types.QuerySigningInfoResponse, error) {
	params := &types.QuerySigningInfoRequest{ConsAddress: sdk.ConsAddress(publicKey.Address()).String()}
	res, err := types.NewQueryClient(cc).SigningInfo(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (cc *ChainClient) QuerySlashingSigningInfos(pageReq *querytypes.PageRequest) (*types.QuerySigningInfosResponse, error) {
	res, err := types.NewQueryClient(cc).SigningInfos(context.Background(),
		&types.QuerySigningInfosRequest{Pagination: pageReq})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (cc *ChainClient) QuerySlashingParams() (*types.QueryParamsResponse, error) {
	res, err := types.NewQueryClient(cc).Params(context.Background(), &types.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}
	return res, nil
}
