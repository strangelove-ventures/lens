package client

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/ibc-go/v2/modules/core/24-host"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/light"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	transfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v2/modules/core/03-connection/types"
	chantypes "github.com/cosmos/ibc-go/v2/modules/core/04-channel/types"
	commitmenttypes "github.com/cosmos/ibc-go/v2/modules/core/23-commitment/types"
	committypes "github.com/cosmos/ibc-go/v2/modules/core/23-commitment/types"
	ibcexported "github.com/cosmos/ibc-go/v2/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v2/modules/light-clients/07-tendermint/types"
	tmclient "github.com/cosmos/ibc-go/v2/modules/light-clients/07-tendermint/types"
)

// QueryTx takes a transaction hash and returns the transaction
func (cc *ChainClient) QueryTx(hashHex string) (*ctypes.ResultTx, error) {
	hash, err := hex.DecodeString(hashHex)
	if err != nil {
		return &ctypes.ResultTx{}, err
	}

	return cc.RPCClient.Tx(context.Background(), hash, true)
}

// QueryTxs returns an array of transactions given a tag
func (cc *ChainClient) QueryTxs(page, limit int, events []string) ([]*ctypes.ResultTx, error) {
	if len(events) == 0 {
		return nil, errors.New("must declare at least one event to search")
	}

	if page <= 0 {
		return nil, errors.New("page must greater than 0")
	}

	if limit <= 0 {
		return nil, errors.New("limit must greater than 0")
	}

	res, err := cc.RPCClient.TxSearch(context.Background(), strings.Join(events, " AND "), true, &page, &limit, "")
	if err != nil {
		return nil, err
	}
	return res.Txs, nil
}

// QueryBalance returns the amount of coins in the relayer account
func (cc *ChainClient) QueryBalance(keyName string) (sdk.Coins, error) {
	var (
		addr string
		err  error
	)
	if keyName == "" {
		addr, err = cc.Address()
	} else {
		cc.Config.Key = keyName
		addr, err = cc.Address()
	}

	if err != nil {
		return nil, err
	}
	return cc.QueryBalanceWithAddress(addr)
}

// QueryBalanceWithAddress returns the amount of coins in the relayer account with address as input
// TODO add pagination support
func (cc *ChainClient) QueryBalanceWithAddress(address string) (sdk.Coins, error) {
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}

	p := bankTypes.NewQueryAllBalancesRequest(addr, DefaultPageRequest())

	queryClient := bankTypes.NewQueryClient(cc)

	res, err := queryClient.AllBalances(context.Background(), p)
	if err != nil {
		return nil, err
	}

	return res.Balances, nil
}

// QueryBalanceWithAddress returns the amount of coins in the relayer account with address as input
//func (cc *ChainClient) QueryBalanceWithAddress(ctx context.Context, address sdk.AccAddress, pageReq *query.PageRequest) (sdk.Coins, error) {
//	addr, err := cc.EncodeBech32AccAddr(address)
//	if err != nil {
//		return nil, err
//	}
//	params := &bankTypes.QueryAllBalancesRequest{Address: addr, Pagination: pageReq}
//	res, err := bankTypes.NewQueryClient(cc).AllBalances(ctx, params)
//	if err != nil {
//		return nil, err
//	}
//
//	return res.Balances, nil
//}

// QueryUnbondingPeriod returns the unbonding period of the chain
func (cc *ChainClient) QueryUnbondingPeriod() (time.Duration, error) {
	req := stakingtypes.QueryParamsRequest{}

	queryClient := stakingtypes.NewQueryClient(cc)

	res, err := queryClient.Params(context.Background(), &req)
	if err != nil {
		return 0, err
	}

	return res.Params.UnbondingTime, nil
}

// QueryTendermintProof performs an ABCI query with the given key and returns
// the value of the query, the proto encoded merkle proof, and the height of
// the Tendermint block containing the state root. The desired tendermint height
// to perform the query should be set in the client context. The query will be
// performed at one below this height (at the IAVL version) in order to obtain
// the correct merkle proof. Proof queries at height less than or equal to 2 are
// not supported. Queries with a client context height of 0 will perform a query
// at the lastest state available.
// Issue: https://github.com/cosmos/cosmos-sdk/issues/6567
func (cc *ChainClient) QueryTendermintProof(height int64, key []byte) ([]byte, []byte, clienttypes.Height, error) {
	// ABCI queries at heights 1, 2 or less than or equal to 0 are not supported.
	// Base app does not support queries for height less than or equal to 1.
	// Therefore, a query at height 2 would be equivalent to a query at height 3.
	// A height of 0 will query with the lastest state.
	if height != 0 && height <= 2 {
		return nil, nil, clienttypes.Height{}, fmt.Errorf("proof queries at height <= 2 are not supported")
	}

	// Use the IAVL height if a valid tendermint height is passed in.
	// A height of 0 will query with the latest state.
	if height != 0 {
		height--
	}

	req := abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
		Height: height,
		Data:   key,
		Prove:  true,
	}

	res, err := cc.QueryABCI(req)
	if err != nil {
		return nil, nil, clienttypes.Height{}, err
	}

	merkleProof, err := commitmenttypes.ConvertProofs(res.ProofOps)
	if err != nil {
		return nil, nil, clienttypes.Height{}, err
	}

	cdc := codec.NewProtoCodec(cc.Codec.InterfaceRegistry)

	proofBz, err := cdc.Marshal(&merkleProof)
	if err != nil {
		return nil, nil, clienttypes.Height{}, err
	}

	revision := clienttypes.ParseChainID(cc.Config.ChainID)
	return res.Value, proofBz, clienttypes.NewHeight(revision, uint64(res.Height)+1), nil
}

// QueryClientStateResponse retrieves the latest consensus state for a client in state at a given height
func (cc *ChainClient) QueryClientStateResponse(height int64, srcClientId string) (*clienttypes.QueryClientStateResponse, error) {
	key := host.FullClientStateKey(srcClientId)

	value, proofBz, proofHeight, err := cc.QueryTendermintProof(height, key)
	if err != nil {
		return nil, err
	}

	// check if client exists
	if len(value) == 0 {
		return nil, sdkerrors.Wrap(clienttypes.ErrClientNotFound, srcClientId)
	}

	cdc := codec.NewProtoCodec(cc.Codec.InterfaceRegistry)

	clientState, err := clienttypes.UnmarshalClientState(cdc, value)
	if err != nil {
		return nil, err
	}

	anyClientState, err := clienttypes.PackClientState(clientState)
	if err != nil {
		return nil, err
	}

	clientStateRes := clienttypes.NewQueryClientStateResponse(anyClientState, proofBz, proofHeight)
	return clientStateRes, nil
}

// QueryClientState retrevies the latest consensus state for a client in state at a given height
// and unpacks it to exported client state interface
func (cc *ChainClient) QueryClientState(height int64, clientid string) (ibcexported.ClientState, error) {
	clientStateRes, err := cc.QueryClientStateResponse(height, clientid)
	if err != nil {
		return nil, err
	}

	clientStateExported, err := clienttypes.UnpackClientState(clientStateRes.ClientState)
	if err != nil {
		return nil, err
	}

	return clientStateExported, nil
}

// QueryClientConsensusState retrieves the latest consensus state for a client in state at a given height
func (cc *ChainClient) QueryClientConsensusState(chainHeight int64, clientid string, clientHeight ibcexported.Height) (*clienttypes.QueryConsensusStateResponse, error) {
	key := host.FullConsensusStateKey(clientid, clientHeight)

	value, proofBz, proofHeight, err := cc.QueryTendermintProof(chainHeight, key)
	if err != nil {
		return nil, err
	}

	// check if consensus state exists
	if len(value) == 0 {
		return nil, sdkerrors.Wrap(clienttypes.ErrConsensusStateNotFound, clientid)
	}

	cdc := codec.NewProtoCodec(cc.Codec.InterfaceRegistry)

	cs, err := clienttypes.UnmarshalConsensusState(cdc, value)
	if err != nil {
		return nil, err
	}

	anyConsensusState, err := clienttypes.PackConsensusState(cs)
	if err != nil {
		return nil, err
	}

	return clienttypes.NewQueryConsensusStateResponse(anyConsensusState, proofBz, proofHeight), nil
}

//DefaultUpgradePath is the default IBC upgrade path set for an on-chain light client
var defaultUpgradePath = []string{"upgrade", "upgradedIBCState"}

func (cc *ChainClient) NewClientState(dstUpdateHeader ibcexported.Header, dstTrustingPeriod, dstUbdPeriod time.Duration, allowUpdateAfterExpiry, allowUpdateAfterMisbehaviour bool) (ibcexported.ClientState, error) {
	dstTmHeader, ok := dstUpdateHeader.(*tmclient.Header)
	if !ok {
		return nil, fmt.Errorf("got data of type %T but wanted  tmclient.Header \n", dstUpdateHeader)
	}
	// Create the ClientState we want on 'c' tracking 'dst'
	return tmclient.NewClientState(
		dstTmHeader.GetHeader().GetChainID(),
		tmclient.NewFractionFromTm(light.DefaultTrustLevel),
		dstTrustingPeriod,
		dstUbdPeriod,
		time.Minute*10,
		dstUpdateHeader.GetHeight().(clienttypes.Height),
		committypes.GetSDKSpecs(),
		defaultUpgradePath,
		allowUpdateAfterExpiry,
		allowUpdateAfterMisbehaviour,
	), nil
}

// QueryUpgradeProof performs an abci query with the given key and returns the proto encoded merkle proof
// for the query and the height at which the proof will succeed on a tendermint verifier.
func (cc *ChainClient) QueryUpgradeProof(key []byte, height uint64) ([]byte, clienttypes.Height, error) {
	res, err := cc.QueryABCI(abci.RequestQuery{
		Path:   "store/upgrade/key",
		Height: int64(height - 1),
		Data:   key,
		Prove:  true,
	})
	if err != nil {
		return nil, clienttypes.Height{}, err
	}

	merkleProof, err := committypes.ConvertProofs(res.ProofOps)
	if err != nil {
		return nil, clienttypes.Height{}, err
	}

	proof, err := cc.Codec.Marshaler.Marshal(&merkleProof)
	if err != nil {
		return nil, clienttypes.Height{}, err
	}

	revision := clienttypes.ParseChainID(cc.Config.ChainID)

	// proof height + 1 is returned as the proof created corresponds to the height the proof
	// was created in the IAVL tree. Tendermint and subsequently the clients that rely on it
	// have heights 1 above the IAVL tree. Thus we return proof height + 1
	return proof, clienttypes.NewHeight(revision, uint64(res.Height+1)), nil
}

// QueryUpgradedClient returns upgraded client info
func (cc *ChainClient) QueryUpgradedClient(height int64) (*clienttypes.QueryClientStateResponse, error) {
	req := clienttypes.QueryUpgradedClientStateRequest{}

	queryClient := clienttypes.NewQueryClient(cc)

	res, err := queryClient.UpgradedClientState(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	if res == nil || res.UpgradedClientState == nil {
		return nil, fmt.Errorf("upgraded client state plan does not exist at height %d", height)
	}

	proof, proofHeight, err := cc.QueryUpgradeProof(upgradetypes.UpgradedClientKey(height), uint64(height))
	if err != nil {
		return nil, err
	}

	return &clienttypes.QueryClientStateResponse{
		ClientState: res.UpgradedClientState,
		Proof:       proof,
		ProofHeight: proofHeight,
	}, nil
}

// QueryUpgradedConsState returns upgraded consensus state and height of client
func (cc *ChainClient) QueryUpgradedConsState(height int64) (*clienttypes.QueryConsensusStateResponse, error) {
	req := clienttypes.QueryUpgradedConsensusStateRequest{}

	queryClient := clienttypes.NewQueryClient(cc)

	res, err := queryClient.UpgradedConsensusState(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	if res == nil || res.UpgradedConsensusState == nil {
		return nil, fmt.Errorf("upgraded consensus state plan does not exist at height %d", height)
	}

	proof, proofHeight, err := cc.QueryUpgradeProof(upgradetypes.UpgradedConsStateKey(height), uint64(height))
	if err != nil {
		return nil, err
	}

	return &clienttypes.QueryConsensusStateResponse{
		ConsensusState: res.UpgradedConsensusState,
		Proof:          proof,
		ProofHeight:    proofHeight,
	}, nil
}

// QueryConsensusState returns a consensus state for a given chain to be used as a
// client in another chain, fetches latest height when passed 0 as arg
func (cc *ChainClient) QueryConsensusState(height int64) (ibcexported.ConsensusState, int64, error) {
	commit, err := cc.RPCClient.Commit(context.Background(), &height)
	if err != nil {
		return &ibctmtypes.ConsensusState{}, 0, err
	}

	page := 1
	count := 10_000

	nextHeight := height + 1
	nextVals, err := cc.RPCClient.Validators(context.Background(), &nextHeight, &page, &count)
	if err != nil {
		return &ibctmtypes.ConsensusState{}, 0, err
	}

	state := &ibctmtypes.ConsensusState{
		Timestamp:          commit.Time,
		Root:               commitmenttypes.NewMerkleRoot(commit.AppHash),
		NextValidatorsHash: tmtypes.NewValidatorSet(nextVals.Validators).Hash(),
	}

	return state, height, nil
}

// QueryClients queries all the clients!
// TODO add pagination support
func (cc *ChainClient) QueryClients() (clienttypes.IdentifiedClientStates, error) {
	qc := clienttypes.NewQueryClient(cc)
	state, err := qc.ClientStates(context.Background(), &clienttypes.QueryClientStatesRequest{
		Pagination: DefaultPageRequest(),
	})
	if err != nil {
		return nil, err
	}
	return state.ClientStates, nil
}

// QueryConnection returns the remote end of a given connection
func (cc *ChainClient) QueryConnection(height int64, connectionid string) (*conntypes.QueryConnectionResponse, error) {
	res, err := cc.queryConnectionABCI(height, connectionid)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return conntypes.NewQueryConnectionResponse(
			conntypes.NewConnectionEnd(
				conntypes.UNINITIALIZED,
				"client",
				conntypes.NewCounterparty(
					"client",
					"connection",
					committypes.NewMerklePrefix([]byte{}),
				),
				[]*conntypes.Version{},
				0,
			), []byte{}, clienttypes.NewHeight(0, 0)), nil
	} else if err != nil {
		return nil, err
	}
	return res, nil
}

func (cc *ChainClient) queryConnectionABCI(height int64, connectionID string) (*conntypes.QueryConnectionResponse, error) {
	key := host.ConnectionKey(connectionID)

	value, proofBz, proofHeight, err := cc.QueryTendermintProof(height, key)
	if err != nil {
		return nil, err
	}

	// check if connection exists
	if len(value) == 0 {
		return nil, sdkerrors.Wrap(conntypes.ErrConnectionNotFound, connectionID)
	}

	cdc := codec.NewProtoCodec(cc.Codec.InterfaceRegistry)

	var connection conntypes.ConnectionEnd
	if err := cdc.Unmarshal(value, &connection); err != nil {
		return nil, err
	}

	return conntypes.NewQueryConnectionResponse(connection, proofBz, proofHeight), nil
}

// QueryConnections gets any connections on a chain
// TODO add pagination support
func (cc *ChainClient) QueryConnections() (conns []*conntypes.IdentifiedConnection, err error) {
	qc := conntypes.NewQueryClient(cc)
	res, err := qc.Connections(context.Background(), &conntypes.QueryConnectionsRequest{
		Pagination: DefaultPageRequest(),
	})
	return res.Connections, err
}

// QueryConnectionsUsingClient gets any connections that exist between chain and counterparty
// TODO add pagination support
func (cc *ChainClient) QueryConnectionsUsingClient(height int64, clientid string) (*conntypes.QueryConnectionsResponse, error) {
	qc := conntypes.NewQueryClient(cc)
	res, err := qc.Connections(context.Background(), &conntypes.QueryConnectionsRequest{
		Pagination: DefaultPageRequest(),
	})
	return res, err
}

// GenerateConnHandshakeProof generates all the proofs needed to prove the existence of the
// connection state on this chain. A counterparty should use these generated proofs.
func (cc *ChainClient) GenerateConnHandshakeProof(height int64, clientId, connId string) (clientState ibcexported.ClientState, clientStateProof []byte, consensusProof []byte, connectionProof []byte, connectionProofHeight ibcexported.Height, err error) {
	var (
		clientStateRes     *clienttypes.QueryClientStateResponse
		consensusStateRes  *clienttypes.QueryConsensusStateResponse
		connectionStateRes *conntypes.QueryConnectionResponse
		eg                 = new(errgroup.Group)
	)

	// query for the client state for the proof and get the height to query the consensus state at.
	clientStateRes, err = cc.QueryClientStateResponse(height, clientId)
	if err != nil {
		return nil, nil, nil, nil, clienttypes.Height{}, err
	}

	clientState, err = clienttypes.UnpackClientState(clientStateRes.ClientState)
	if err != nil {
		return nil, nil, nil, nil, clienttypes.Height{}, err
	}

	eg.Go(func() error {
		var err error
		consensusStateRes, err = cc.QueryClientConsensusState(height, clientId, clientState.GetLatestHeight())
		return err
	})
	eg.Go(func() error {
		var err error
		connectionStateRes, err = cc.QueryConnection(height, connId)
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, nil, nil, nil, clienttypes.Height{}, err
	}

	return clientState, clientStateRes.Proof, consensusStateRes.Proof, connectionStateRes.Proof, connectionStateRes.ProofHeight, nil
}

// QueryChannel returns the channel associated with a channelID
func (cc *ChainClient) QueryChannel(height int64, channelid, portid string) (chanRes *chantypes.QueryChannelResponse, err error) {
	res, err := cc.queryChannelABCI(height, portid, channelid)
	if err != nil && strings.Contains(err.Error(), "not found") {
		return chantypes.NewQueryChannelResponse(
			chantypes.NewChannel(
				chantypes.UNINITIALIZED,
				chantypes.UNORDERED,
				chantypes.NewCounterparty(
					"port",
					"channel",
				),
				[]string{},
				"version",
			),
			[]byte{},
			clienttypes.NewHeight(0, 0)), nil
	} else if err != nil {
		return nil, err
	}
	return res, nil
}

func (cc *ChainClient) queryChannelABCI(height int64, portID, channelID string) (*chantypes.QueryChannelResponse, error) {
	key := host.ChannelKey(portID, channelID)

	value, proofBz, proofHeight, err := cc.QueryTendermintProof(height, key)
	if err != nil {
		return nil, err
	}

	// check if channel exists
	if len(value) == 0 {
		return nil, sdkerrors.Wrapf(chantypes.ErrChannelNotFound, "portID (%s), channelID (%s)", portID, channelID)
	}

	cdc := codec.NewProtoCodec(cc.Codec.InterfaceRegistry)

	var channel chantypes.Channel
	if err := cdc.Unmarshal(value, &channel); err != nil {
		return nil, err
	}

	return chantypes.NewQueryChannelResponse(channel, proofBz, proofHeight), nil
}

// QueryChannelClient returns the client state of the client supporting a given channel
func (cc *ChainClient) QueryChannelClient(height int64, channelid, portid string) (*clienttypes.IdentifiedClientState, error) {
	qc := chantypes.NewQueryClient(cc)
	cState, err := qc.ChannelClientState(context.Background(), &chantypes.QueryChannelClientStateRequest{
		PortId:    portid,
		ChannelId: channelid,
	})
	if err != nil {
		return nil, err
	}
	return cState.IdentifiedClientState, nil
}

// QueryConnectionChannels queries the channels associated with a connection
func (cc *ChainClient) QueryConnectionChannels(height int64, connectionid string) ([]*chantypes.IdentifiedChannel, error) {
	qc := chantypes.NewQueryClient(cc)
	chans, err := qc.ConnectionChannels(context.Background(), &chantypes.QueryConnectionChannelsRequest{
		Connection: connectionid,
		Pagination: DefaultPageRequest(),
	})
	if err != nil {
		return nil, err
	}
	return chans.Channels, nil
}

// QueryChannels returns all the channels that are registered on a chain
// TODO add pagination support
func (cc *ChainClient) QueryChannels() ([]*chantypes.IdentifiedChannel, error) {
	qc := chantypes.NewQueryClient(cc)
	res, err := qc.Channels(context.Background(), &chantypes.QueryChannelsRequest{
		Pagination: DefaultPageRequest(),
	})
	if err != nil {
		return nil, err
	}
	return res.Channels, err
}

// QueryPacketCommitments returns an array of packet commitments
// TODO add pagination support
func (cc *ChainClient) QueryPacketCommitments(height uint64, channelid, portid string) (commitments *chantypes.QueryPacketCommitmentsResponse, err error) {
	qc := chantypes.NewQueryClient(cc)
	c, err := qc.PacketCommitments(context.Background(), &chantypes.QueryPacketCommitmentsRequest{
		PortId:     portid,
		ChannelId:  channelid,
		Pagination: DefaultPageRequest(),
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}

// QueryPacketAcknowledgements returns an array of packet acks
// TODO add pagination support
func (cc *ChainClient) QueryPacketAcknowledgements(height uint64, channelid, portid string) (acknowledgements []*chantypes.PacketState, err error) {
	qc := chantypes.NewQueryClient(cc)
	acks, err := qc.PacketAcknowledgements(context.Background(), &chantypes.QueryPacketAcknowledgementsRequest{
		PortId:     portid,
		ChannelId:  channelid,
		Pagination: DefaultPageRequest(),
	})
	if err != nil {
		return nil, err
	}
	return acks.Acknowledgements, nil
}

// QueryUnreceivedPackets returns a list of unrelayed packet commitments
func (cc *ChainClient) QueryUnreceivedPackets(height uint64, channelid, portid string, seqs []uint64) ([]uint64, error) {
	qc := chantypes.NewQueryClient(cc)
	res, err := qc.UnreceivedPackets(context.Background(), &chantypes.QueryUnreceivedPacketsRequest{
		PortId:                    portid,
		ChannelId:                 channelid,
		PacketCommitmentSequences: seqs,
	})
	if err != nil {
		return nil, err
	}
	return res.Sequences, nil
}

// QueryUnreceivedAcknowledgements returns a list of unrelayed packet acks
func (cc *ChainClient) QueryUnreceivedAcknowledgements(height uint64, channelid, portid string, seqs []uint64) ([]uint64, error) {
	qc := chantypes.NewQueryClient(cc)
	res, err := qc.UnreceivedAcks(context.Background(), &chantypes.QueryUnreceivedAcksRequest{
		PortId:             portid,
		ChannelId:          channelid,
		PacketAckSequences: seqs,
	})
	if err != nil {
		return nil, err
	}
	return res.Sequences, nil
}

// QueryNextSeqRecv returns the next seqRecv for a configured channel
func (cc *ChainClient) QueryNextSeqRecv(height int64, channelid, portid string) (recvRes *chantypes.QueryNextSequenceReceiveResponse, err error) {
	key := host.NextSequenceRecvKey(portid, channelid)

	value, proofBz, proofHeight, err := cc.QueryTendermintProof(height, key)
	if err != nil {
		return nil, err
	}

	// check if next sequence receive exists
	if len(value) == 0 {
		return nil, sdkerrors.Wrapf(chantypes.ErrChannelNotFound, "portID (%s), channelID (%s)", portid, channelid)
	}

	sequence := binary.BigEndian.Uint64(value)

	return chantypes.NewQueryNextSequenceReceiveResponse(sequence, proofBz, proofHeight), nil
}

// QueryPacketCommitment returns the packet commitment proof at a given height
func (cc *ChainClient) QueryPacketCommitment(height int64, channelid, portid string, seq uint64) (comRes *chantypes.QueryPacketCommitmentResponse, err error) {
	key := host.PacketCommitmentKey(portid, channelid, seq)

	value, proofBz, proofHeight, err := cc.QueryTendermintProof(height, key)
	if err != nil {
		return nil, err
	}

	// check if packet commitment exists
	if len(value) == 0 {
		return nil, sdkerrors.Wrapf(chantypes.ErrPacketCommitmentNotFound, "portID (%s), channelID (%s), sequence (%d)", portid, channelid, seq)
	}

	return chantypes.NewQueryPacketCommitmentResponse(value, proofBz, proofHeight), nil
}

// QueryPacketAcknowledgement returns the packet ack proof at a given height
func (cc *ChainClient) QueryPacketAcknowledgement(height int64, channelid, portid string, seq uint64) (ackRes *chantypes.QueryPacketAcknowledgementResponse, err error) {
	key := host.PacketAcknowledgementKey(portid, channelid, seq)

	value, proofBz, proofHeight, err := cc.QueryTendermintProof(height, key)
	if err != nil {
		return nil, err
	}

	if len(value) == 0 {
		return nil, sdkerrors.Wrapf(chantypes.ErrInvalidAcknowledgement, "portID (%s), channelID (%s), sequence (%d)", portid, channelid, seq)
	}

	return chantypes.NewQueryPacketAcknowledgementResponse(value, proofBz, proofHeight), nil
}

// QueryPacketReceipt returns the packet receipt proof at a given height
func (cc *ChainClient) QueryPacketReceipt(height int64, channelid, portid string, seq uint64) (recRes *chantypes.QueryPacketReceiptResponse, err error) {
	key := host.PacketReceiptKey(portid, channelid, seq)

	value, proofBz, proofHeight, err := cc.QueryTendermintProof(height, key)
	if err != nil {
		return nil, err
	}

	return chantypes.NewQueryPacketReceiptResponse(value != nil, proofBz, proofHeight), nil
}

func (cc *ChainClient) QueryLatestHeight() (int64, error) {
	stat, err := cc.RPCClient.Status(context.Background())
	if err != nil {
		return -1, err
	} else if stat.SyncInfo.CatchingUp {
		return -1, fmt.Errorf("node at %s running chain %s not caught up", cc.Config.RPCAddr, cc.Config.ChainID)
	}
	return stat.SyncInfo.LatestBlockHeight, nil
}

// QueryHeaderAtHeight returns the header at a given height
func (cc *ChainClient) QueryHeaderAtHeight(height int64) (ibcexported.Header, error) {
	var (
		page    = 1
		perPage = 100000
	)
	if height <= 0 {
		return nil, fmt.Errorf("must pass in valid height, %d not valid", height)
	}

	res, err := cc.RPCClient.Commit(context.Background(), &height)
	if err != nil {
		return nil, err
	}

	val, err := cc.RPCClient.Validators(context.Background(), &height, &page, &perPage)
	if err != nil {
		return nil, err
	}

	protoVal, err := tmtypes.NewValidatorSet(val.Validators).ToProto()
	if err != nil {
		return nil, err
	}

	return &tmclient.Header{
		// NOTE: This is not a SignedHeader
		// We are missing a light.Commit type here
		SignedHeader: res.SignedHeader.ToProto(),
		ValidatorSet: protoVal,
	}, nil
}

// QueryDenomTrace takes a denom from IBC and queries the information about it
func (cc *ChainClient) QueryDenomTrace(denom string) (*transfertypes.DenomTrace, error) {
	transfers, err := transfertypes.NewQueryClient(cc).DenomTrace(context.Background(),
		&transfertypes.QueryDenomTraceRequest{
			Hash: denom,
		})
	if err != nil {
		return nil, err
	}
	return transfers.DenomTrace, nil
}

// QueryDenomTraces returns all the denom traces from a given chain
// TODO add pagination support
func (cc *ChainClient) QueryDenomTraces(offset, limit uint64, height int64) ([]transfertypes.DenomTrace, error) {
	transfers, err := transfertypes.NewQueryClient(cc).DenomTraces(context.Background(),
		&transfertypes.QueryDenomTracesRequest{
			Pagination: DefaultPageRequest(),
		})
	if err != nil {
		return nil, err
	}
	return transfers.DenomTraces, nil
}

// QueryDenomTraces returns all the denom traces from a given chain
//func (cc *ChainClient) QueryDenomTraces(pageReq *querytypes.PageRequest, height int64) (*transfertypes.QueryDenomTracesResponse, error) {
//	ctx := SetHeightOnContext(context.Background(), height)
//	return transfertypes.NewQueryClient(cc).DenomTraces(ctx, &transfertypes.QueryDenomTracesRequest{
//		Pagination: pageReq,
//	})
//}

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

// QueryBalanceWithDenomTraces is a helper function for query balance
func (cc *ChainClient) QueryBalanceWithDenomTraces(ctx context.Context, address sdk.AccAddress, pageReq *query.PageRequest) (sdk.Coins, error) {
	coins, err := cc.QueryBalanceWithAddress(address.String())
	if err != nil {
		return nil, err
	}

	h, err := cc.QueryLatestHeight()
	if err != nil {
		return nil, err
	}

	// TODO: figure out how to handle this
	// we don't want to expose user to this
	// so maybe we need a QueryAllDenomTraces function
	// that will paginate the responses automatically
	// TODO fix pagination here later
	dts, err := cc.QueryDenomTraces(0, 1000, h)
	if err != nil {
		return nil, err
	}

	if len(dts) == 0 {
		return coins, nil
	}

	var out sdk.Coins
	for _, c := range coins {
		if c.Amount.Equal(sdk.NewInt(0)) {
			continue
		}

		for i, d := range dts {
			if c.Denom == d.IBCDenom() {
				out = append(out, sdk.Coin{Denom: d.GetFullDenomPath(), Amount: c.Amount})
				break
			}

			if i == len(dts)-1 {
				out = append(out, c)
			}
		}
	}
	return out, nil
}

func (cc *ChainClient) QueryDelegatorValidators(ctx context.Context, address sdk.AccAddress) ([]string, error) {
	res, err := distTypes.NewQueryClient(cc).DelegatorValidators(ctx, &distTypes.QueryDelegatorValidatorsRequest{
		DelegatorAddress: cc.MustEncodeAccAddr(address),
	})
	if err != nil {
		return nil, err
	}
	return res.Validators, nil
}

func (cc *ChainClient) QueryDistributionCommission(ctx context.Context, address sdk.ValAddress) (sdk.DecCoins, error) {
	valAddr, err := cc.EncodeBech32ValAddr(address)
	if err != nil {
		return nil, err
	}
	request := distTypes.QueryValidatorCommissionRequest{
		ValidatorAddress: valAddr,
	}
	res, err := distTypes.NewQueryClient(cc).ValidatorCommission(ctx, &request)
	if err != nil {
		return nil, err
	}
	return res.Commission.Commission, nil
}

func (cc *ChainClient) QueryDistributionCommunityPool(ctx context.Context) (sdk.DecCoins, error) {
	res, err := distTypes.NewQueryClient(cc).CommunityPool(ctx, &distTypes.QueryCommunityPoolRequest{})
	if err != nil {
		return nil, err
	}
	return res.Pool, err
}

func (cc *ChainClient) QueryDistributionParams(ctx context.Context) (*distTypes.Params, error) {
	res, err := distTypes.NewQueryClient(cc).Params(ctx, &distTypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}
	return &res.Params, nil
}

func (cc *ChainClient) QueryDistributionRewards(ctx context.Context, delegatorAddress sdk.AccAddress, validatorAddress sdk.ValAddress) (sdk.DecCoins, error) {
	delAddr, err := cc.EncodeBech32AccAddr(delegatorAddress)
	if err != nil {
		return nil, err
	}
	valAddr, err := cc.EncodeBech32ValAddr(validatorAddress)
	if err != nil {
		return nil, err
	}
	request := distTypes.QueryDelegationRewardsRequest{
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
	}
	res, err := distTypes.NewQueryClient(cc).DelegationRewards(ctx, &request)
	if err != nil {
		return nil, err
	}
	return res.Rewards, nil
}

// QueryDistributionSlashes returns all slashes of a validator, optionally pass the start and end height
func (cc *ChainClient) QueryDistributionSlashes(ctx context.Context, validatorAddress sdk.ValAddress, startHeight, endHeight uint64, pageReq *querytypes.PageRequest) (*distTypes.QueryValidatorSlashesResponse, error) {
	valAddr, err := cc.EncodeBech32ValAddr(validatorAddress)
	if err != nil {
		return nil, err
	}
	request := distTypes.QueryValidatorSlashesRequest{
		ValidatorAddress: valAddr,
		StartingHeight:   startHeight,
		EndingHeight:     endHeight,
		Pagination:       pageReq,
	}
	return distTypes.NewQueryClient(cc).ValidatorSlashes(ctx, &request)
}

// QueryDistributionValidatorRewards returns all the validator distribution rewards from a given height
func (cc *ChainClient) QueryDistributionValidatorRewards(ctx context.Context, validatorAddress sdk.ValAddress) (sdk.DecCoins, error) {
	valAddr, err := cc.EncodeBech32ValAddr(validatorAddress)
	if err != nil {
		return nil, err
	}
	request := distTypes.QueryValidatorOutstandingRewardsRequest{
		ValidatorAddress: valAddr,
	}
	res, err := distTypes.NewQueryClient(cc).ValidatorOutstandingRewards(ctx, &request)
	if err != nil {
		return nil, err
	}
	return res.Rewards.Rewards, nil
}

// QueryTotalSupply returns the total supply of coins on a chain
func (cc *ChainClient) QueryTotalSupply(ctx context.Context, pageReq *query.PageRequest) (*bankTypes.QueryTotalSupplyResponse, error) {
	return bankTypes.NewQueryClient(cc).TotalSupply(ctx, &bankTypes.QueryTotalSupplyRequest{Pagination: pageReq})
}

func (cc *ChainClient) QueryDenomsMetadata(ctx context.Context, pageReq *query.PageRequest) (*bankTypes.QueryDenomsMetadataResponse, error) {
	return bankTypes.NewQueryClient(cc).DenomsMetadata(ctx, &bankTypes.QueryDenomsMetadataRequest{Pagination: pageReq})
}

func (cc *ChainClient) QueryStakingParams(ctx context.Context) (*stakingtypes.Params, error) {
	res, err := stakingtypes.NewQueryClient(cc).Params(ctx, &stakingtypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}
	return &res.Params, nil
}

func DefaultPageRequest() *querytypes.PageRequest {
	return &querytypes.PageRequest{
		Key:        []byte(""),
		Offset:     0,
		Limit:      1000,
		CountTotal: true,
	}
}
