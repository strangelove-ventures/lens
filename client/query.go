package client

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	transfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
)

// QueryBalanceWithAddress returns the amount of coins in the relayer account with address as input
func (cc *ChainClient) QueryBalanceWithAddress(address sdk.AccAddress) (sdk.Coins, error) {
	addr, err := cc.EncodeBech32AccAddr(address)
	if err != nil {
		return nil, err
	}
	params := &bankTypes.QueryAllBalancesRequest{Address: addr, Pagination: DefaultPageRequest()}
	res, err := bankTypes.NewQueryClient(cc).AllBalances(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return res.Balances, nil
}

func (cc *ChainClient) QueryLatestHeight() (int64, error) {
	stat, err := cc.RPCClient.Status(context.Background())
	if err != nil {
		return 0, err
	}
	return stat.SyncInfo.LatestBlockHeight, nil
}

// QueryDenomTraces returns all the denom traces from a given chain
func (cc *ChainClient) QueryDenomTraces(pageReq *querytypes.PageRequest, height int64) (*transfertypes.QueryDenomTracesResponse, error) {
	ctx := SetHeightOnContext(context.Background(), height)
	return transfertypes.NewQueryClient(cc).DenomTraces(ctx, &transfertypes.QueryDenomTracesRequest{
		Pagination: pageReq,
	})
}

func (cc *ChainClient) QueryAccount(address sdk.AccAddress) (authtypes.AccountI, error) {
	addr, err := cc.EncodeBech32AccAddr(address)
	if err != nil {
		return nil, err
	}
	res, err := authtypes.NewQueryClient(cc).Account(context.Background(), &authtypes.QueryAccountRequest{Address: addr})
	if err != nil {
		return nil, err
	}
	var acc authtypes.AccountI
	if err := cc.Codec.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return nil, err
	}
	return acc, nil
}

// QueryBalance is a helper function for query balance
func (cc *ChainClient) QueryBalance(address sdk.AccAddress, showDenoms bool) (sdk.Coins, error) {
	coins, err := cc.QueryBalanceWithAddress(address)
	if err != nil {
		return nil, err
	}

	if showDenoms {
		return coins, nil
	}

	h, err := cc.QueryLatestHeight()
	if err != nil {
		return nil, err
	}

	dts, err := cc.QueryDenomTraces(DefaultPageRequest(), h)
	if err != nil {
		return nil, err
	}

	if len(dts.DenomTraces) == 0 {
		return coins, nil
	}

	var out sdk.Coins
	for _, c := range coins {
		if c.Amount.Equal(sdk.NewInt(0)) {
			continue
		}

		for i, d := range dts.DenomTraces {
			if c.Denom == d.IBCDenom() {
				out = append(out, sdk.Coin{Denom: d.GetFullDenomPath(), Amount: c.Amount})
				break
			}

			if i == len(dts.DenomTraces)-1 {
				out = append(out, c)
			}
		}
	}
	return out, nil
}

func (cc *ChainClient) QueryDistributionCommission(address string) (*disttypes.ValidatorAccumulatedCommission, error) {
	request := disttypes.QueryValidatorCommissionRequest{
		ValidatorAddress: address,
	}
	res, err := disttypes.NewQueryClient(cc).ValidatorCommission(context.Background(), &request)

	if err != nil {
		return nil, err
	}

	return &res.Commission, nil
}

func (cc *ChainClient) QueryDistributionCommunityPool() (sdk.DecCoins, error) {
	request := disttypes.QueryCommunityPoolRequest{}

	res, err := disttypes.NewQueryClient(cc).CommunityPool(context.Background(), &request)

	if err != nil {
		return nil, err
	}

	return res.Pool, err
}

func (cc *ChainClient) QueryDistributionParams() (*disttypes.Params, error) {
	res, err := disttypes.NewQueryClient(cc).Params(context.Background(), &disttypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	return &res.Params, nil
}

func (cc *ChainClient) QueryDistributionRewards(delegatorAddress string, validatorAddress string) (sdk.DecCoins, error) {
	request := disttypes.QueryDelegationRewardsRequest{
		DelegatorAddress: delegatorAddress,
		ValidatorAddress: validatorAddress,
	}
	res, err := disttypes.NewQueryClient(cc).DelegationRewards(context.Background(), &request)

	if err != nil {
		return nil, err
	}

	return res.Rewards, nil
}

func (cc *ChainClient) QueryDistributionSlashes(validatorAddress string) ([]disttypes.ValidatorSlashEvent, error) {
	request := disttypes.QueryValidatorSlashesRequest{
		ValidatorAddress: validatorAddress,
	}

	res, err := disttypes.NewQueryClient(cc).ValidatorSlashes(context.Background(), &request)

	if err != nil {
		return nil, err
	}

	return res.Slashes, nil
}

func (cc *ChainClient) QueryDistributionValidatorRewards(validatorAddress string) (*disttypes.ValidatorOutstandingRewards, error) {
	request := disttypes.QueryValidatorOutstandingRewardsRequest{
		ValidatorAddress: validatorAddress,
	}

	res, err := disttypes.NewQueryClient(cc).ValidatorOutstandingRewards(context.Background(), &request)

	if err != nil {
		return nil, err
	}

	return &res.Rewards, nil
}

func DefaultPageRequest() *querytypes.PageRequest {
	return &querytypes.PageRequest{
		Key:        []byte(""),
		Offset:     0,
		Limit:      1000,
		CountTotal: true,
	}
}
