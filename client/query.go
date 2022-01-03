package client

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	transfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
)

// QueryBalanceWithAddress returns the amount of coins in the relayer account with address as input
func (cc *ChainClient) QueryBalanceWithAddress(address sdk.AccAddress, pageReq *query.PageRequest) (sdk.Coins, error) {
	addr, err := cc.EncodeBech32AccAddr(address)
	if err != nil {
		return nil, err
	}
	params := &bankTypes.QueryAllBalancesRequest{Address: addr, Pagination: pageReq}
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
	coins, err := cc.QueryBalanceWithAddress(address, DefaultPageRequest())
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

func (cc *ChainClient) QueryDelegatorValidators(address sdk.AccAddress) ([]string, error) {
	res, err := distTypes.NewQueryClient(cc).DelegatorValidators(context.Background(), &distTypes.QueryDelegatorValidatorsRequest{
		DelegatorAddress: address.String(),
	})
	if err != nil {
		return nil, err
	}
	return res.Validators, nil
}

func (cc *ChainClient) QueryDistributionCommission(address string) (*distTypes.ValidatorAccumulatedCommission, error) {
	request := distTypes.QueryValidatorCommissionRequest{
		ValidatorAddress: address,
	}
	res, err := distTypes.NewQueryClient(cc).ValidatorCommission(context.Background(), &request)

	if err != nil {
		return nil, err
	}

	return &res.Commission, nil
}

func (cc *ChainClient) QueryDistributionCommunityPool() (sdk.DecCoins, error) {
	request := distTypes.QueryCommunityPoolRequest{}

	res, err := distTypes.NewQueryClient(cc).CommunityPool(context.Background(), &request)

	if err != nil {
		return nil, err
	}

	return res.Pool, err
}

func (cc *ChainClient) QueryDistributionParams() (*distTypes.Params, error) {
	res, err := distTypes.NewQueryClient(cc).Params(context.Background(), &distTypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	return &res.Params, nil
}

func (cc *ChainClient) QueryDistributionRewards(delegatorAddress string, validatorAddress string) (sdk.DecCoins, error) {
	request := distTypes.QueryDelegationRewardsRequest{
		DelegatorAddress: delegatorAddress,
		ValidatorAddress: validatorAddress,
	}
	res, err := distTypes.NewQueryClient(cc).DelegationRewards(context.Background(), &request)

	if err != nil {
		return nil, err
	}

	return res.Rewards, nil
}

func (cc *ChainClient) QueryDistributionSlashes(validatorAddress string, startHeight, endHeight uint64, pageReq *querytypes.PageRequest) ([]distTypes.ValidatorSlashEvent, error) {
	request := distTypes.QueryValidatorSlashesRequest{
		ValidatorAddress: validatorAddress,
		StartingHeight:   startHeight,
		EndingHeight:     endHeight,
		Pagination:       pageReq,
	}

	res, err := distTypes.NewQueryClient(cc).ValidatorSlashes(context.Background(), &request)
	if err != nil {
		return nil, err
	}

	return res.Slashes, nil
}

func (cc *ChainClient) QueryDistributionValidatorRewards(validatorAddress string) (*distTypes.ValidatorOutstandingRewards, error) {
	request := distTypes.QueryValidatorOutstandingRewardsRequest{
		ValidatorAddress: validatorAddress,
	}

	res, err := distTypes.NewQueryClient(cc).ValidatorOutstandingRewards(context.Background(), &request)

	if err != nil {
		return nil, err
	}

	return &res.Rewards, nil
}

func (cc *ChainClient) QueryGovernanceProposals(proposalStatus govTypes.ProposalStatus, voter string, depositor string, pageReq *querytypes.PageRequest) ([]govTypes.Proposal, error) {
	request := govTypes.QueryProposalsRequest{
		ProposalStatus: proposalStatus,
		Voter:          voter,
		Depositor:      depositor,
		Pagination:     pageReq,
	}

	res, err := govTypes.NewQueryClient(cc).Proposals(context.Background(), &request)

	if err != nil {
		return nil, err
	}

	fmt.Printf("%v\n", res.Proposals)

	return res.Proposals, nil
}

func DefaultPageRequest() *querytypes.PageRequest {
	return &querytypes.PageRequest{
		Key:        []byte(""),
		Offset:     0,
		Limit:      1000,
		CountTotal: true,
	}
}
