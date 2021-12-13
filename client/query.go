package client

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
)

// QueryBalanceWithAddress returns the amount of coins in the relayer account with address as input
func (cc *ChainClient) QueryBalanceWithAddress(address sdk.AccAddress) (sdk.Coins, error) {
	params := bankTypes.NewQueryAllBalancesRequest(address, DefaultPageRequest())
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

func DefaultPageRequest() *querytypes.PageRequest {
	return &querytypes.PageRequest{
		Key:        []byte(""),
		Offset:     0,
		Limit:      1000,
		CountTotal: true,
	}
}
