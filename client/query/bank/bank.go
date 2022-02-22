package bank

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/strangelove-ventures/lens/client"
	"time"
)

// QueryBalanceWithAddress returns the balance of all coins for a single account.
func QueryBalanceWithAddress(cc *client.ChainClient, address string, pr *querytypes.PageRequest) (sdk.Coins, error) {
	req := &bankTypes.QueryAllBalancesRequest{Address: address, Pagination: pr}
	queryClient := bankTypes.NewQueryClient(cc)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*10000))
	defer cancel()
	res, err := queryClient.AllBalances(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.Balances, nil
}

// QueryTotalSupply returns the supply of all coins
func QueryTotalSupply(cc *client.ChainClient, pr *querytypes.PageRequest) (sdk.Coins, error) {
	req := &bankTypes.QueryTotalSupplyRequest{Pagination: pr}
	queryClient := bankTypes.NewQueryClient(cc)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*10000))
	defer cancel()
	res, err := queryClient.TotalSupply(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.Supply, nil
}

// QueryDenomsMetadata returns the metadata for all denoms
func QueryDenomsMetadata(cc *client.ChainClient, pr *querytypes.PageRequest) ([]bankTypes.Metadata, error) {
	req := &bankTypes.QueryDenomsMetadataRequest{Pagination: pr}
	queryClient := bankTypes.NewQueryClient(cc)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*10000))
	defer cancel()
	res, err := queryClient.DenomsMetadata(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.Metadatas, nil
}
