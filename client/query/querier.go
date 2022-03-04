package query

import (
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/lens/client"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Query struct {
	Client  *client.ChainClient
	Options *QueryOptions
}

// Bank queries

func (q *Query) Balances(address string) (*bankTypes.QueryAllBalancesResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return BalanceWithAddressRPC(q, address)
}

func (q *Query) TotalSupply() (*bankTypes.QueryTotalSupplyResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return TotalSupplyRPC(q)
}

func (q *Query) DenomsMetadata() (*bankTypes.QueryDenomsMetadataResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return DenomsMetadataRPC(q)
}

// Staking queries

func (q *Query) Delegation(delegator, validator string) (*stakingTypes.DelegationResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return DelegationRPC(q, delegator, validator)
}

func (q *Query) Delegations(delegator string) (*stakingTypes.QueryDelegatorDelegationsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return DelegationsRPC(q, delegator)
}

// Distribution queries

func (q *Query) DelegatorValidators(delegator string) (*distributionTypes.QueryDelegatorValidatorsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return DelegatorValidatorsRPC(q, delegator)
}

// Tendermint queries

func (q *Query) Block() (*coretypes.ResultBlock, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return BlockRPC(q)
}
