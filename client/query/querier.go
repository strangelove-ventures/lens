package query

import (
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v5/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/strangelove-ventures/lens/client"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type Query struct {
	Client  *client.ChainClient
	Options *QueryOptions
}

// Bank queries

// Return params for bank module.
func (q *Query) Bank_Params() (*bankTypes.QueryParamsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return bank_ParamsRPC(q)
}

// Balances returns the balance of specific denom for a single account.
func (q *Query) Bank_Balance(address string, denom string) (*bankTypes.QueryBalanceResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return bank_BalanceRPC(q, address, denom)
}

// Balances returns the balance of all coins for a single account.
func (q *Query) Bank_Balances(address string) (*bankTypes.QueryAllBalancesResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return bank_AllBalancesRPC(q, address)
}

// SupplyOf returns the supply of given coin
func (q *Query) Bank_SupplyOf(denom string) (*bankTypes.QuerySupplyOfResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return bank_SupplyOfRPC(q, denom)
}

// TotalSupply returns the supply of all coins
func (q *Query) Bank_TotalSupply() (*bankTypes.QueryTotalSupplyResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return bank_TotalSupplyRPC(q)
}

// DenomMetadata returns the metadata for given denoms
func (q *Query) Bank_DenomMetadata(denom string) (*bankTypes.QueryDenomMetadataResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return bank_DenomMetadataRPC(q, denom)
}

// DenomsMetadata returns the metadata for all denoms
func (q *Query) Bank_DenomsMetadata() (*bankTypes.QueryDenomsMetadataResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return bank_DenomsMetadataRPC(q)
}

// Staking queries

// Return params for staking module.
func (q *Query) Staking_Params() (*stakingTypes.QueryParamsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_ParamsRPC(q)
}

// Return balance of staking pool.
func (q *Query) Staking_Pool() (*stakingTypes.QueryPoolResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_PoolRPC(q)
}

// Return specified validator.
func (q *Query) Staking_Validator(address string) (*stakingTypes.QueryValidatorResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_ValidatorRPC(q, address)
}

// Return validators for given status.
func (q *Query) Staking_Validators(status string) (*stakingTypes.QueryValidatorsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_ValidatorsRPC(q, status)
}

// ValidatorDelegations returns all the delegations for a validator
func (q *Query) Staking_ValidatorDelegations(validator string) (*stakingTypes.QueryValidatorDelegationsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_ValidatorDelegationsRPC(q, validator)
}

// ValidatorDelegations returns all the unbonding delegations for a validator
func (q *Query) Staking_ValidatorUnbondingDelegations(validator string) (*stakingTypes.QueryValidatorUnbondingDelegationsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_ValidatorUnbondingDelegationsRPC(q, validator)
}

// Delegation returns the delegations for a particular validator / delegator tuple
func (q *Query) Staking_Delegation(delegator string, validator string) (*stakingTypes.QueryDelegationResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_DelegationRPC(q, delegator, validator)
}

// UnbondingDelegation returns the unbonding delegations for a particular validator / delegator tuple
func (q *Query) Staking_UnbondingDelegation(delegator string, validator string) (*stakingTypes.QueryUnbondingDelegationResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_UnbondingDelegationRPC(q, delegator, validator)
}

// DelegatorDelegations returns all the delegations for a given delegator
func (q *Query) Staking_DelegatorDelegations(delegator string) (*stakingTypes.QueryDelegatorDelegationsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_DelegatorDelegationsRPC(q, delegator)
}

// Delegations returns all the unbonding delegations for a given delegator
func (q *Query) Staking_DelegatorUnbondingDelegations(delegator string) (*stakingTypes.QueryDelegatorUnbondingDelegationsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_DelegatorUnbondingDelegationsRPC(q, delegator)
}

// Delegation returns the delegations for a particular validator / delegator tuple
func (q *Query) Staking_Redelegations(delegator string, src_validator string, dst_validator string) (*stakingTypes.QueryRedelegationsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_RedelegationsRPC(q, delegator, src_validator, dst_validator)
}

// DelegatorValidators returns all the validators for a given delegator
func (q *Query) Staking_DelegatorValidators(delegator string) (*stakingTypes.QueryDelegatorValidatorsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_DelegatorValidatorsRPC(q, delegator)
}

// DelegatorValidators returns the validator for a given delegator / validator tuple
func (q *Query) Staking_DelegatorValidator(delegator string, validator string) (*stakingTypes.QueryDelegatorValidatorResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_DelegatorValidatorRPC(q, delegator, validator)
}

// HistoricalInfoRPC return histrical info for a given height
func (q *Query) Staking_HistoricalInfo(height int64) (*stakingTypes.QueryHistoricalInfoResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return staking_HistoricalInfoRPC(q, height)
}

// Distribution queries

// Return params for staking module.
func (q *Query) Distribution_Params() (*distributionTypes.QueryParamsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return distribution_ParamsRPC(q)
}

// Return balance of community pool.
func (q *Query) Distribution_CommunityPool() (*distributionTypes.QueryCommunityPoolResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return distribution_CommunityPoolRPC(q)
}

// ValidatorOutstandingRewards returns the outstanding reward pool for given validator
func (q *Query) Distribution_ValidatorOutstandingRewards(validator string) (*distributionTypes.QueryValidatorOutstandingRewardsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return distribution_ValidatorOutstandingRewardsRPC(q, validator)
}

// ValidatorCommission returns the outstanding commission for given validator
func (q *Query) Distribution_ValidatorCommission(validator string) (*distributionTypes.QueryValidatorCommissionResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return distribution_ValidatorCommissionRPC(q, validator)
}

// ValidatorSlashes returns slashing events for given validator between the optional start and end height
func (q *Query) Distribution_ValidatorSlashes(validator string, start uint64, end uint64) (*distributionTypes.QueryValidatorSlashesResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return distribution_ValidatorSlashesRPC(q, validator, start, end)
}

// DelegationRewards returns the validators of a delegator
func (q *Query) Distribution_DelegationRewards(delegator string, validator string) (*distributionTypes.QueryDelegationRewardsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return distribution_DelegationRewardsRPC(q, delegator, validator)
}

// DelegationTotalRewards returns the validators of a delegator
func (q *Query) Distribution_DelegationTotalRewards(delegator string) (*distributionTypes.QueryDelegationTotalRewardsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return distribution_DelegationTotalRewardsRPC(q, delegator)
}

// DelegatorValidators returns the validators of a delegator
func (q *Query) Distribution_DelegatorValidators(delegator string) (*distributionTypes.QueryDelegatorValidatorsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return distribution_DelegatorValidatorsRPC(q, delegator)
}

// DelegatorWithdrawAddress returns the validators of a delegator
func (q *Query) Distribution_DelegatorWithdrawAddress(delegator string) (*distributionTypes.QueryDelegatorWithdrawAddressResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return distribution_DelegatorWithdrawAddressRPC(q, delegator)
}

// Tendermint queries

// Block returns information about a block
func (q *Query) Block() (*coretypes.ResultBlock, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return BlockRPC(q)
}

// BlockByHash returns information about a block by hash
func (q *Query) BlockByHash(hash string) (*coretypes.ResultBlock, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return BlockByHashRPC(q, hash)
}

// BlockResults returns information about a block by hash
func (q *Query) BlockResults() (*coretypes.ResultBlockResults, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return BlockResultsRPC(q)
}

// Status returns information about a node status
func (q *Query) Status() (*coretypes.ResultStatus, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return StatusRPC(q)
}

// ABCIInfo returns general information about the ABCI application
func (q *Query) ABCIInfo() (*coretypes.ResultABCIInfo, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return ABCIInfoRPC(q)
}

// ABCIQuery returns data from a particular path in the ABCI application
func (q *Query) ABCIQuery(path string, data string, prove bool) (*coretypes.ResultABCIQuery, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return ABCIQueryRPC(q, path, data, prove)
}

// IBC Queries

// IBCQuery returns parameters for the IBC client submodule.
func (q *Query) Ibc_ClientParams() (*clienttypes.QueryClientParamsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return ibc_ClientParamsRPC(q)
}

// Ibc_ClientState returns the client state for the specified IBC client.
func (q *Query) Ibc_ClientState(clientId string) (*clienttypes.QueryClientStateResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return ibc_ClientStateRPC(q, clientId)
}

// Ibc_ClientStates returns the client state for all IBC clients.
func (q *Query) Ibc_ClientStates() (*clienttypes.QueryClientStatesResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return ibc_ClientStatesRPC(q)
}

// Ibc_ConsensusState returns the consensus state for the specified IBC client and the given height.
func (q *Query) Ibc_ConsensusState(clientId string, height clienttypes.Height) (*clienttypes.QueryConsensusStateResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return ibc_ConsensusStateRPC(q, clientId, height)
}

// Ibc_ConsensusState returns all consensus states for the specified IBC client.
func (q *Query) Ibc_ConsensusStates(clientId string) (*clienttypes.QueryConsensusStatesResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return ibc_ConsensusStatesRPC(q, clientId)
}

// Ibc_Connection returns the connection state for the specified IBC connection.
func (q *Query) Ibc_Connection(connectionId string) (*connectiontypes.QueryConnectionResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return ibc_ConnectionRPC(q, connectionId)
}

// Ibc_Connections returns the connection state for all IBC connections.
func (q *Query) Ibc_Connections() (*connectiontypes.QueryConnectionsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return ibc_ConnectionsRPC(q)
}

// Ibc_Channel returns the channel state for the specified IBC channel and port.
func (q *Query) Ibc_Channel(channelId string, portId string) (*channeltypes.QueryChannelResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return ibc_ChannelRPC(q, channelId, portId)
}

// Ibc_Channels returns the channel state for all IBC channels.
func (q *Query) Ibc_Channels() (*channeltypes.QueryChannelsResponse, error) {
	/// TODO: In the future have some logic to route the query to the appropriate client (gRPC or RPC)
	return ibc_ChannelsRPC(q)
}
