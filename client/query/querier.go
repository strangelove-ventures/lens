package query

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/lens/client"
)

type Query struct {
	Client  *client.ChainClient
	Options *QueryOptions
}

func (q *Query) Balances(address string) (sdk.Coins, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return BalanceWithAddressRPC(q, address)
}

func (q *Query) TotalSupply() (sdk.Coins, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return TotalSupplyRPC(q)
}

func (q *Query) DenomsMetadata() ([]bankTypes.Metadata, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return DenomsMetadataRPC(q)
}

func (q *Query) Delegation(delegator, validator string) (*types.DelegationResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return Delegation(q, delegator, validator)
}

func (q *Query) Delegations(delegator string) (types.DelegationResponses, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return Delegations(q, delegator)
}
