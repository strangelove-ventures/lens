package client

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/avast/retry-go"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v2/modules/core/03-connection/types"
	chantypes "github.com/cosmos/ibc-go/v2/modules/core/04-channel/types"
	commitmenttypes "github.com/cosmos/ibc-go/v2/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v2/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v2/modules/core/exported"
	tmclient "github.com/cosmos/ibc-go/v2/modules/light-clients/07-tendermint/types"
	"github.com/cosmos/relayer/relayer/provider"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	_                  provider.ChainProvider = &ChainClient{}
	_                  provider.KeyProvider   = &ChainClient{}
	_                  provider.QueryProvider = &ChainClient{}
	defaultChainPrefix                        = commitmenttypes.NewMerklePrefix([]byte("ibc"))
	defaultDelayPeriod                        = uint64(0)

	// Strings for parsing events
	spTag       = "send_packet"
	waTag       = "write_acknowledgement"
	srcChanTag  = "packet_src_channel"
	dstChanTag  = "packet_dst_channel"
	srcPortTag  = "packet_src_port"
	dstPortTag  = "packet_dst_port"
	dataTag     = "packet_data"
	ackTag      = "packet_ack"
	toHeightTag = "packet_timeout_height"
	toTSTag     = "packet_timeout_timestamp"
	seqTag      = "packet_sequence"
)

type CosmosMessage struct {
	Msg sdk.Msg
}

func NewCosmosMessage(msg sdk.Msg) provider.RelayerMessage {
	return CosmosMessage{
		Msg: msg,
	}
}

func CosmosMsg(rm provider.RelayerMessage) sdk.Msg {
	if val, ok := rm.(CosmosMessage); !ok {
		fmt.Printf("got data of type %T but wanted provider.CosmosMessage \n", val)
		return nil
	} else {
		return val.Msg
	}
}

func CosmosMsgs(rm ...provider.RelayerMessage) []sdk.Msg {
	sdkMsgs := make([]sdk.Msg, 0)
	for _, rMsg := range rm {
		if val, ok := rMsg.(CosmosMessage); !ok {
			fmt.Printf("got data of type %T but wanted provider.CosmosMessage \n", val)
			return nil
		} else {
			sdkMsgs = append(sdkMsgs, val.Msg)
		}
	}
	return sdkMsgs
}

func (cm CosmosMessage) Type() string {
	return sdk.MsgTypeURL(cm.Msg)
}

func (cm CosmosMessage) MsgBytes() ([]byte, error) {
	return proto.Marshal(cm.Msg)
}

func (cc *ChainClient) ProviderConfig() provider.ProviderConfig {
	return cc.Config
}

func (cc *ChainClient) ChainId() string {
	return cc.Config.ChainID
}

func (cc *ChainClient) Type() string {
	return "cosmos"
}

func (cc *ChainClient) Key() string {
	return cc.Config.Key
}

func (cc *ChainClient) Timeout() string {
	return cc.Config.Timeout
}

// Address returns the chains configured address as a string
func (cc *ChainClient) Address() (string, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	if acc, err = cc.GetKeyAddress(); err != nil {
		return "", err
	}
	return acc.String(), nil
}

func (cc *ChainClient) TrustingPeriod() (time.Duration, error) {
	res, err := cc.QueryStakingParams(context.Background())
	if err != nil {
		return 0, err
	}
	integer, _ := math.Modf(res.UnbondingTime.Hours() * 0.7)
	trustingStr := fmt.Sprintf("%vh", integer)
	tp, err := time.ParseDuration(trustingStr)
	if err != nil {
		return 0, nil
	}
	return tp, nil
}

// CreateClient creates an sdk.Msg to update the client on src with consensus state from dst
func (cc *ChainClient) CreateClient(clientState ibcexported.ClientState, dstHeader ibcexported.Header) (provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	if err := dstHeader.ValidateBasic(); err != nil {
		return nil, err
	}

	tmHeader, ok := dstHeader.(*tmclient.Header)
	if !ok {
		return nil, fmt.Errorf("got data of type %T but wanted tmclient.Header \n", dstHeader)
	}

	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}

	msg, err := clienttypes.NewMsgCreateClient(
		clientState,
		tmHeader.ConsensusState(),
		acc.String(),
	)
	if err != nil {
		return nil, err
	}

	return NewCosmosMessage(msg), msg.ValidateBasic()
}

func (cc *ChainClient) SubmitMisbehavior( /*TBD*/ ) (provider.RelayerMessage, error) {
	return nil, nil
}

func (cc *ChainClient) UpdateClient(srcClientId string, dstHeader ibcexported.Header) (provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	if err := dstHeader.ValidateBasic(); err != nil {
		return nil, err
	}
	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}
	msg, err := clienttypes.NewMsgUpdateClient(
		srcClientId,
		dstHeader,
		acc.String(),
	)
	if err != nil {
		return nil, err
	}
	return NewCosmosMessage(msg), msg.ValidateBasic()
}

func (cc *ChainClient) ConnectionOpenInit(srcClientId, dstClientId string, dstHeader ibcexported.Header) ([]provider.RelayerMessage, error) {
	var (
		acc     sdk.AccAddress
		err     error
		version *conntypes.Version
	)
	updateMsg, err := cc.UpdateClient(srcClientId, dstHeader)
	if err != nil {
		return nil, err
	}

	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}
	msg := conntypes.NewMsgConnectionOpenInit(
		srcClientId,
		dstClientId,
		defaultChainPrefix,
		version,
		defaultDelayPeriod,
		acc.String(),
	)

	return []provider.RelayerMessage{updateMsg, NewCosmosMessage(msg)}, msg.ValidateBasic()
}

func (cc *ChainClient) ConnectionOpenTry(dstQueryProvider provider.QueryProvider, dstHeader ibcexported.Header, srcClientId, dstClientId, srcConnId, dstConnId string) ([]provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	updateMsg, err := cc.UpdateClient(srcClientId, dstHeader)
	if err != nil {
		return nil, err
	}

	cph, err := dstQueryProvider.QueryLatestHeight()
	if err != nil {
		return nil, err
	}

	clientState, clientStateProof, consensusStateProof, connStateProof, proofHeight, err := dstQueryProvider.GenerateConnHandshakeProof(cph, dstClientId, dstConnId)
	if err != nil {
		return nil, err
	}

	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}

	// TODO: Get DelayPeriod from counterparty connection rather than using default value
	msg := conntypes.NewMsgConnectionOpenTry(
		srcConnId,
		srcClientId,
		dstConnId,
		dstClientId,
		clientState,
		defaultChainPrefix,
		conntypes.ExportedVersionsToProto(conntypes.GetCompatibleVersions()),
		defaultDelayPeriod,
		connStateProof,
		clientStateProof,
		consensusStateProof,
		clienttypes.NewHeight(proofHeight.GetRevisionNumber(), proofHeight.GetRevisionHeight()),
		clientState.GetLatestHeight().(clienttypes.Height),
		acc.String(),
	)

	return []provider.RelayerMessage{updateMsg, NewCosmosMessage(msg)}, msg.ValidateBasic()
}

func (cc *ChainClient) ConnectionOpenAck(dstQueryProvider provider.QueryProvider, dstHeader ibcexported.Header, srcClientId, srcConnId, dstClientId, dstConnId string) ([]provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)

	updateMsg, err := cc.UpdateClient(srcClientId, dstHeader)
	if err != nil {
		return nil, err
	}
	cph, err := dstQueryProvider.QueryLatestHeight()
	if err != nil {
		return nil, err
	}

	clientState, clientStateProof, consensusStateProof, connStateProof,
		proofHeight, err := dstQueryProvider.GenerateConnHandshakeProof(cph, dstClientId, dstConnId)
	if err != nil {
		return nil, err
	}

	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}
	msg := conntypes.NewMsgConnectionOpenAck(
		srcConnId,
		dstConnId,
		clientState,
		connStateProof,
		clientStateProof,
		consensusStateProof,
		clienttypes.NewHeight(proofHeight.GetRevisionNumber(), proofHeight.GetRevisionHeight()),
		clientState.GetLatestHeight().(clienttypes.Height),
		conntypes.DefaultIBCVersion,
		acc.String(),
	)

	return []provider.RelayerMessage{updateMsg, NewCosmosMessage(msg)}, msg.ValidateBasic()
}

func (cc *ChainClient) ConnectionOpenConfirm(dstQueryProvider provider.QueryProvider, dstHeader ibcexported.Header, dstConnId, srcClientId, srcConnId string) ([]provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	updateMsg, err := cc.UpdateClient(srcClientId, dstHeader)
	if err != nil {
		return nil, err
	}

	cph, err := dstQueryProvider.QueryLatestHeight()
	if err != nil {
		return nil, err
	}
	counterpartyConnState, err := dstQueryProvider.QueryConnection(cph, dstConnId)
	if err != nil {
		return nil, err
	}

	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}
	msg := conntypes.NewMsgConnectionOpenConfirm(
		srcConnId,
		counterpartyConnState.Proof,
		counterpartyConnState.ProofHeight,
		acc.String(),
	)

	return []provider.RelayerMessage{updateMsg, NewCosmosMessage(msg)}, msg.ValidateBasic()
}

func (cc *ChainClient) ChannelOpenInit(srcClientId, srcConnId, srcPortId, srcVersion, dstPortId string, order chantypes.Order, dstHeader ibcexported.Header) ([]provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	updateMsg, err := cc.UpdateClient(srcClientId, dstHeader)
	if err != nil {
		return nil, err
	}

	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}
	msg := chantypes.NewMsgChannelOpenInit(
		srcPortId,
		srcVersion,
		order,
		[]string{srcConnId},
		dstPortId,
		acc.String(),
	)

	return []provider.RelayerMessage{updateMsg, NewCosmosMessage(msg)}, msg.ValidateBasic()
}

func (cc *ChainClient) ChannelOpenTry(dstQueryProvider provider.QueryProvider, dstHeader ibcexported.Header, srcPortId, dstPortId, srcChanId, dstChanId, srcVersion, srcConnectionId, srcClientId string) ([]provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	updateMsg, err := cc.UpdateClient(srcClientId, dstHeader)
	if err != nil {
		return nil, err
	}
	cph, err := dstQueryProvider.QueryLatestHeight()
	if err != nil {
		return nil, err
	}

	counterpartyChannelRes, err := dstQueryProvider.QueryChannel(cph, dstChanId, dstPortId)
	if err != nil {
		return nil, err
	}

	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}
	msg := chantypes.NewMsgChannelOpenTry(
		srcPortId,
		srcChanId,
		srcVersion,
		counterpartyChannelRes.Channel.Ordering,
		[]string{srcConnectionId},
		dstPortId,
		dstChanId,
		counterpartyChannelRes.Channel.Version,
		counterpartyChannelRes.Proof,
		counterpartyChannelRes.ProofHeight,
		acc.String(),
	)

	return []provider.RelayerMessage{updateMsg, NewCosmosMessage(msg)}, msg.ValidateBasic()
}

func (cc *ChainClient) ChannelOpenAck(dstQueryProvider provider.QueryProvider, dstHeader ibcexported.Header, srcClientId, srcPortId, srcChanId, dstChanId, dstPortId string) ([]provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	updateMsg, err := cc.UpdateClient(srcClientId, dstHeader)
	if err != nil {
		return nil, err
	}

	cph, err := dstQueryProvider.QueryLatestHeight()
	if err != nil {
		return nil, err
	}

	counterpartyChannelRes, err := dstQueryProvider.QueryChannel(cph, dstChanId, dstPortId)
	if err != nil {
		return nil, err
	}

	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}
	msg := chantypes.NewMsgChannelOpenAck(
		srcPortId,
		srcChanId,
		dstChanId,
		counterpartyChannelRes.Channel.Version,
		counterpartyChannelRes.Proof,
		counterpartyChannelRes.ProofHeight,
		acc.String(),
	)

	return []provider.RelayerMessage{updateMsg, NewCosmosMessage(msg)}, msg.ValidateBasic()
}

func (cc *ChainClient) ChannelOpenConfirm(dstQueryProvider provider.QueryProvider, dstHeader ibcexported.Header, srcClientId, srcPortId, srcChanId, dstPortId, dstChanId string) ([]provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	updateMsg, err := cc.UpdateClient(srcClientId, dstHeader)
	if err != nil {
		return nil, err
	}
	cph, err := dstQueryProvider.QueryLatestHeight()
	if err != nil {
		return nil, err
	}

	counterpartyChanState, err := dstQueryProvider.QueryChannel(cph, dstChanId, dstPortId)
	if err != nil {
		return nil, err
	}

	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}
	msg := chantypes.NewMsgChannelOpenConfirm(
		srcPortId,
		srcChanId,
		counterpartyChanState.Proof,
		counterpartyChanState.ProofHeight,
		acc.String(),
	)

	return []provider.RelayerMessage{updateMsg, NewCosmosMessage(msg)}, msg.ValidateBasic()
}

func (cc *ChainClient) ChannelCloseInit(srcPortId, srcChanId string) (provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}
	return NewCosmosMessage(chantypes.NewMsgChannelCloseInit(
		srcPortId,
		srcChanId,
		acc.String(),
	)), nil
}

func (cc *ChainClient) ChannelCloseConfirm(dstQueryProvider provider.QueryProvider, dsth int64, dstChanId, dstPortId, srcPortId, srcChanId string) (provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	dstChanResp, err := dstQueryProvider.QueryChannel(dsth, dstChanId, dstPortId)
	if err != nil {
		return nil, err
	}

	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}
	return NewCosmosMessage(chantypes.NewMsgChannelCloseConfirm(
		srcPortId,
		srcChanId,
		dstChanResp.Proof,
		dstChanResp.ProofHeight,
		acc.String(),
	)), nil
}

// GetIBCUpdateHeader updates the off chain tendermint light client and
// returns an IBC Update Header which can be used to update an on chain
// light client on the destination chain. The source is used to construct
// the header data.
func (cc *ChainClient) GetIBCUpdateHeader(srch int64, dst provider.ChainProvider, dstClientId string) (ibcexported.Header, error) {
	// Construct header data from light client representing source.
	h, err := cc.GetLightSignedHeaderAtHeight(srch)
	if err != nil {
		return nil, err
	}

	// Inject trusted fields based on previous header data from source
	return cc.InjectTrustedFields(h, dst, dstClientId)
}

func (cc *ChainClient) GetLightSignedHeaderAtHeight(h int64) (ibcexported.Header, error) {
	if h == 0 {
		return nil, errors.New("height cannot be 0")
	}
	lightBlock, err := cc.LightProvider.LightBlock(context.Background(), h)
	if err != nil {
		return nil, err
	}
	protoVal, err := tmtypes.NewValidatorSet(lightBlock.ValidatorSet.Validators).ToProto()
	if err != nil {
		return nil, err
	}

	return &tmclient.Header{
		SignedHeader: lightBlock.SignedHeader.ToProto(),
		ValidatorSet: protoVal,
	}, nil
}

// InjectTrustedFields injects the necessary trusted fields for a header to update a light
// client stored on the destination chain, using the information provided by the source
// chain.
// TrustedHeight is the latest height of the IBC client on dst
// TrustedValidators is the validator set of srcChain at the TrustedHeight
// InjectTrustedFields returns a copy of the header with TrustedFields modified
func (cc *ChainClient) InjectTrustedFields(header ibcexported.Header, dst provider.ChainProvider, dstClientId string) (ibcexported.Header, error) {
	// make copy of header stored in mop
	h, ok := header.(*tmclient.Header)
	if !ok {
		return nil, fmt.Errorf("trying to inject fields into non-tendermint headers")
	}

	// retrieve dst client from src chain
	// this is the client that will updated
	cs, err := dst.QueryClientState(0, dstClientId)
	if err != nil {
		return nil, err
	}

	// inject TrustedHeight as latest height stored on dst client
	h.TrustedHeight = cs.GetLatestHeight().(clienttypes.Height)

	// NOTE: We need to get validators from the source chain at height: trustedHeight+1
	// since the last trusted validators for a header at height h is the NextValidators
	// at h+1 committed to in header h by NextValidatorsHash

	// TODO: this is likely a source of off by 1 errors but may be impossible to change? Maybe this is the
	// place where we need to fix the upstream query proof issue?
	var trustedHeader *tmclient.Header
	if err := retry.Do(func() error {
		tmpHeader, err := cc.GetLightSignedHeaderAtHeight(int64(h.TrustedHeight.RevisionHeight) + 1)
		th, ok := tmpHeader.(*tmclient.Header)
		if !ok {
			err = errors.New("non-tm client header")
		}
		trustedHeader = th
		return err
	}, provider.RtyAtt, provider.RtyDel, provider.RtyErr); err != nil {
		return nil, fmt.Errorf(
			"failed to get trusted header, please ensure header at the height %d has not been pruned by the connected node: %w",
			h.TrustedHeight.RevisionHeight, err,
		)
	}

	// inject TrustedValidators into header
	h.TrustedValidators = trustedHeader.ValidatorSet
	return h, nil
}

// MsgRelayAcknowledgement constructs the MsgAcknowledgement which is to be sent to the sending chain.
// The counterparty represents the receiving chain where the acknowledgement would be stored.
func (cc *ChainClient) MsgRelayAcknowledgement(dst provider.ChainProvider, dstChanId, dstPortId, srcChanId, srcPortId string, dsth int64, packet provider.RelayPacket) (provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	msgPacketAck, ok := packet.(*relayMsgPacketAck)
	if !ok {
		return nil, fmt.Errorf("got data of type %T but wanted relayMsgPacketAck \n", packet)
	}

	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}

	ackRes, err := dst.QueryPacketAcknowledgement(dsth, dstChanId, dstPortId, packet.Seq())
	switch {
	case err != nil:
		return nil, err
	case ackRes.Proof == nil || ackRes.Acknowledgement == nil:
		return nil, fmt.Errorf("ack packet acknowledgement query seq(%d) is nil", packet.Seq())
	case ackRes == nil:
		return nil, fmt.Errorf("ack packet [%s]seq{%d} has no associated proofs", dst.ChainId(), packet.Seq())
	default:
		return NewCosmosMessage(chantypes.NewMsgAcknowledgement(
			chantypes.NewPacket(
				packet.Data(),
				packet.Seq(),
				srcPortId,
				srcChanId,
				dstPortId,
				dstChanId,
				packet.Timeout(),
				packet.TimeoutStamp(),
			),
			msgPacketAck.ack,
			ackRes.Proof,
			ackRes.ProofHeight,
			acc.String())), nil
	}
}

// MsgTransfer creates a new transfer message
func (cc *ChainClient) MsgTransfer(amount sdk.Coin, dstChainId, dstAddr, srcPortId, srcChanId string, timeoutHeight, timeoutTimestamp uint64) (provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}

	version := clienttypes.ParseChainID(dstChainId)
	return NewCosmosMessage(transfertypes.NewMsgTransfer(
		srcPortId,
		srcChanId,
		amount,
		acc.String(),
		dstAddr,
		clienttypes.NewHeight(version, timeoutHeight),
		timeoutTimestamp,
	)), nil
}

// MsgRelayTimeout constructs the MsgTimeout which is to be sent to the sending chain.
// The counterparty represents the receiving chain where the receipts would have been
// stored.
func (cc *ChainClient) MsgRelayTimeout(dst provider.ChainProvider, dsth int64, packet provider.RelayPacket, dstChanId, dstPortId, srcChanId, srcPortId string) (provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}

	recvRes, err := dst.QueryPacketReceipt(dsth, dstChanId, dstPortId, packet.Seq())
	switch {
	case err != nil:
		return nil, err
	case recvRes.Proof == nil:
		return nil, fmt.Errorf("timeout packet receipt proof seq(%d) is nil", packet.Seq())
	case recvRes == nil:
		return nil, fmt.Errorf("timeout packet [%s]seq{%d} has no associated proofs", cc.Config.ChainID, packet.Seq())
	default:
		return NewCosmosMessage(chantypes.NewMsgTimeout(
			chantypes.NewPacket(
				packet.Data(),
				packet.Seq(),
				srcPortId,
				srcChanId,
				dstPortId,
				dstChanId,
				packet.Timeout(),
				packet.TimeoutStamp(),
			),
			packet.Seq(),
			recvRes.Proof,
			recvRes.ProofHeight,
			acc.String(),
		)), nil
	}
}

// MsgRelayRecvPacket constructs the MsgRecvPacket which is to be sent to the receiving chain.
// The counterparty represents the sending chain where the packet commitment would be stored.
func (cc *ChainClient) MsgRelayRecvPacket(dst provider.ChainProvider, dsth int64, packet provider.RelayPacket, dstChanId, dstPortId, srcChanId, srcPortId string) (provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}

	comRes, err := dst.QueryPacketCommitment(dsth, dstChanId, dstPortId, packet.Seq())
	switch {
	case err != nil:
		return nil, err
	case comRes.Proof == nil || comRes.Commitment == nil:
		return nil, fmt.Errorf("recv packet commitment query seq(%d) is nil", packet.Seq())
	case comRes == nil:
		return nil, fmt.Errorf("receive packet [%s]seq{%d} has no associated proofs", cc.Config.ChainID, packet.Seq())
	default:
		return NewCosmosMessage(chantypes.NewMsgRecvPacket(
			chantypes.NewPacket(
				packet.Data(),
				packet.Seq(),
				dstPortId,
				dstChanId,
				srcPortId,
				srcChanId,
				packet.Timeout(),
				packet.TimeoutStamp(),
			),
			comRes.Proof,
			comRes.ProofHeight,
			acc.String(),
		)), nil
	}
}

// RelayPacketFromSequence relays a packet with a given seq on src and returns recvPacket msgs, timeoutPacketmsgs and error
func (cc *ChainClient) RelayPacketFromSequence(src, dst provider.ChainProvider, srch, dsth, seq uint64, dstChanId, dstPortId, srcChanId, srcPortId, srcClientId string) (provider.RelayerMessage, provider.RelayerMessage, error) {
	txs, err := cc.QueryTxs(1, 1000, rcvPacketQuery(srcChanId, int(seq)))
	switch {
	case err != nil:
		return nil, nil, err
	case len(txs) == 0:
		return nil, nil, fmt.Errorf("no transactions returned with query")
	case len(txs) > 1:
		return nil, nil, fmt.Errorf("more than one transaction returned with query")
	}

	rcvPackets, timeoutPackets, err := relayPacketsFromResultTx(src, dst, int64(dsth), txs[0], dstChanId, dstPortId, srcChanId, srcPortId, srcClientId)
	switch {
	case err != nil:
		return nil, nil, err
	case len(rcvPackets) == 0 && len(timeoutPackets) == 0:
		return nil, nil, fmt.Errorf("no relay msgs created from query response")
	case len(rcvPackets)+len(timeoutPackets) > 1:
		return nil, nil, fmt.Errorf("more than one relay msg found in tx query")
	}

	if len(rcvPackets) == 1 {
		pkt := rcvPackets[0]
		if seq != pkt.Seq() {
			return nil, nil, fmt.Errorf("wrong sequence: expected(%d) got(%d)", seq, pkt.Seq())
		}

		packet, err := dst.MsgRelayRecvPacket(src, int64(srch), pkt, srcChanId, srcPortId, dstChanId, dstPortId)
		if err != nil {
			return nil, nil, err
		}

		return packet, nil, nil
	}

	if len(timeoutPackets) == 1 {
		pkt := timeoutPackets[0]
		if seq != pkt.Seq() {
			return nil, nil, fmt.Errorf("wrong sequence: expected(%d) got(%d)", seq, pkt.Seq())
		}

		timeout, err := src.MsgRelayTimeout(dst, int64(dsth), pkt, dstChanId, dstPortId, srcChanId, srcPortId)
		if err != nil {
			return nil, nil, err
		}
		return nil, timeout, nil
	}

	return nil, nil, fmt.Errorf("should have errored before here")
}

// AcknowledgementFromSequence relays an acknowledgement with a given seq on src, source is the sending chain, destination is the receiving chain
func (cc *ChainClient) AcknowledgementFromSequence(dst provider.ChainProvider, dsth, seq uint64, dstChanId, dstPortId, srcChanId, srcPortId string) (provider.RelayerMessage, error) {
	txs, err := dst.QueryTxs(1, 1000, ackPacketQuery(dstChanId, int(seq)))
	switch {
	case err != nil:
		return nil, err
	case len(txs) == 0:
		return nil, fmt.Errorf("no transactions returned with query")
	case len(txs) > 1:
		return nil, fmt.Errorf("more than one transaction returned with query")
	}

	acks, err := acknowledgementsFromResultTx(dstChanId, dstPortId, srcChanId, srcPortId, txs[0])
	switch {
	case err != nil:
		return nil, err
	case len(acks) == 0:
		return nil, fmt.Errorf("no ack msgs created from query response")
	}

	var out provider.RelayerMessage
	for _, ack := range acks {
		if seq != ack.Seq() {
			continue
		}
		msg, err := cc.MsgRelayAcknowledgement(dst, dstChanId, dstPortId, srcChanId, srcPortId, int64(dsth), ack)
		if err != nil {
			return nil, err
		}
		out = msg
	}
	return out, nil
}

func rcvPacketQuery(channelID string, seq int) []string {
	return []string{fmt.Sprintf("%s.packet_src_channel='%s'", spTag, channelID),
		fmt.Sprintf("%s.packet_sequence='%d'", spTag, seq)}
}

func ackPacketQuery(channelID string, seq int) []string {
	return []string{fmt.Sprintf("%s.packet_dst_channel='%s'", waTag, channelID),
		fmt.Sprintf("%s.packet_sequence='%d'", waTag, seq)}
}

// relayPacketsFromResultTx looks through the events in a *ctypes.ResultTx
// and returns relayPackets with the appropriate data
func relayPacketsFromResultTx(src, dst provider.ChainProvider, dsth int64, res *ctypes.ResultTx, dstChanId, dstPortId, srcChanId, srcPortId, srcClientId string) ([]provider.RelayPacket, []provider.RelayPacket, error) {
	var (
		rcvPackets     []provider.RelayPacket
		timeoutPackets []provider.RelayPacket
	)

	for _, e := range res.TxResult.Events {
		if e.Type == spTag {
			// NOTE: Src and Dst are switched here
			rp := &relayMsgRecvPacket{pass: false}
			for _, p := range e.Attributes {
				if string(p.Key) == srcChanTag {
					if string(p.Value) != srcChanId {
						rp.pass = true
						continue
					}
				}
				if string(p.Key) == dstChanTag {
					if string(p.Value) != dstChanId {
						rp.pass = true
						continue
					}
				}
				if string(p.Key) == srcPortTag {
					if string(p.Value) != srcPortId {
						rp.pass = true
						continue
					}
				}
				if string(p.Key) == dstPortTag {
					if string(p.Value) != dstPortId {
						rp.pass = true
						continue
					}
				}
				if string(p.Key) == dataTag {
					rp.packetData = p.Value
				}
				if string(p.Key) == toHeightTag {
					timeout, err := clienttypes.ParseHeight(string(p.Value))
					if err != nil {
						return nil, nil, err
					}

					rp.timeout = timeout
				}
				if string(p.Key) == toTSTag {
					timeout, _ := strconv.ParseUint(string(p.Value), 10, 64)
					rp.timeoutStamp = timeout
				}
				if string(p.Key) == seqTag {
					seq, _ := strconv.ParseUint(string(p.Value), 10, 64)
					rp.seq = seq
				}
			}

			// fetch the header which represents a block produced on destination
			block, err := dst.GetIBCUpdateHeader(dsth, src, srcClientId)
			if err != nil {
				return nil, nil, err
			}

			switch {
			// If the packet has a timeout height, and it has been reached, return a timeout packet
			case !rp.timeout.IsZero() && block.GetHeight().GTE(rp.timeout):
				timeoutPackets = append(timeoutPackets, rp.timeoutPacket())
			// If the packet matches the relay constraints relay it as a MsgReceivePacket
			case !rp.pass:
				rcvPackets = append(rcvPackets, rp)
			}
		}
	}

	// If there is a relayPacket, return it
	if len(rcvPackets)+len(timeoutPackets) > 0 {
		return rcvPackets, timeoutPackets, nil
	}

	return nil, nil, fmt.Errorf("no packet data found")
}

// acknowledgementsFromResultTx looks through the events in a *ctypes.ResultTx and returns
// relayPackets with the appropriate data
func acknowledgementsFromResultTx(dstChanId, dstPortId, srcChanId, srcPortId string, res *ctypes.ResultTx) ([]provider.RelayPacket, error) {
	var ackPackets []provider.RelayPacket
	for _, e := range res.TxResult.Events {
		if e.Type == waTag {
			// NOTE: Src and Dst are switched here
			rp := &relayMsgPacketAck{pass: false}
			for _, p := range e.Attributes {
				if string(p.Key) == srcChanTag {
					if string(p.Value) != srcChanId {
						rp.pass = true
						continue
					}
				}
				if string(p.Key) == dstChanTag {
					if string(p.Value) != dstChanId {
						rp.pass = true
						continue
					}
				}
				if string(p.Key) == srcPortTag {
					if string(p.Value) != srcPortId {
						rp.pass = true
						continue
					}
				}
				if string(p.Key) == dstPortTag {
					if string(p.Value) != dstPortId {
						rp.pass = true
						continue
					}
				}
				if string(p.Key) == ackTag {
					rp.ack = p.Value
				}
				if string(p.Key) == dataTag {
					rp.packetData = p.Value
				}
				if string(p.Key) == toHeightTag {
					timeout, err := clienttypes.ParseHeight(string(p.Value))
					if err != nil {
						return nil, err
					}
					rp.timeout = timeout
				}
				if string(p.Key) == toTSTag {
					timeout, _ := strconv.ParseUint(string(p.Value), 10, 64)
					rp.timeoutStamp = timeout
				}
				if string(p.Key) == seqTag {
					seq, _ := strconv.ParseUint(string(p.Value), 10, 64)
					rp.seq = seq
				}
			}
			if !rp.pass {
				ackPackets = append(ackPackets, rp)
			}
		}
	}

	// If there is a relayPacket, return it
	if len(ackPackets) > 0 {
		return ackPackets, nil
	}

	return nil, fmt.Errorf("no packet data found")
}

func (cc *ChainClient) MsgUpgradeClient(srcClientId string, consRes *clienttypes.QueryConsensusStateResponse, clientRes *clienttypes.QueryClientStateResponse) (provider.RelayerMessage, error) {
	var (
		acc sdk.AccAddress
		err error
	)
	if acc, err = cc.GetKeyAddress(); err != nil {
		return nil, err
	}
	return NewCosmosMessage(&clienttypes.MsgUpgradeClient{ClientId: srcClientId, ClientState: clientRes.ClientState,
		ConsensusState: consRes.ConsensusState, ProofUpgradeClient: consRes.GetProof(),
		ProofUpgradeConsensusState: consRes.ConsensusState.Value, Signer: acc.String()}), nil
}

// AutoUpdateClient update client automatically to prevent expiry
func (cc *ChainClient) AutoUpdateClient(dst provider.ChainProvider, thresholdTime time.Duration, srcClientId, dstClientId string) (time.Duration, error) {
	srch, err := cc.QueryLatestHeight()
	if err != nil {
		return 0, err
	}
	dsth, err := dst.QueryLatestHeight()
	if err != nil {
		return 0, err
	}

	clientState, err := cc.queryTMClientState(srch, srcClientId)
	if err != nil {
		return 0, err
	}

	if clientState.TrustingPeriod <= thresholdTime {
		return 0, fmt.Errorf("client (%s) trusting period time is less than or equal to threshold time", srcClientId)
	}

	// query the latest consensus state of the potential matching client
	consensusStateResp, err := cc.QueryConsensusStateABCI(srcClientId, clientState.GetLatestHeight())
	if err != nil {
		return 0, err
	}

	exportedConsState, err := clienttypes.UnpackConsensusState(consensusStateResp.ConsensusState)
	if err != nil {
		return 0, err
	}

	consensusState, ok := exportedConsState.(*tmclient.ConsensusState)
	if !ok {
		return 0, fmt.Errorf("consensus state with clientID %s from chain %s is not IBC tendermint type",
			srcClientId, cc.Config.ChainID)
	}

	expirationTime := consensusState.Timestamp.Add(clientState.TrustingPeriod)

	timeToExpiry := time.Until(expirationTime)

	if timeToExpiry > thresholdTime {
		return timeToExpiry, nil
	}

	if clientState.IsExpired(consensusState.Timestamp, time.Now()) {
		return 0, fmt.Errorf("client (%s) is already expired on chain: %s", srcClientId, cc.Config.ChainID)
	}

	srcUpdateHeader, err := cc.GetIBCUpdateHeader(srch, dst, dstClientId)
	if err != nil {
		return 0, err
	}

	dstUpdateHeader, err := dst.GetIBCUpdateHeader(dsth, cc, srcClientId)
	if err != nil {
		return 0, err
	}

	updateMsg, err := cc.UpdateClient(srcClientId, dstUpdateHeader)
	if err != nil {
		return 0, err
	}

	msgs := []provider.RelayerMessage{updateMsg}

	res, success, err := cc.SendMessages(msgs)
	if err != nil {
		// cp.LogFailedTx(res, err, CosmosMsgs(msgs...))
		return 0, err
	}
	if !success {
		return 0, fmt.Errorf("tx failed: %s", res.Data)
	}
	cc.Log(fmt.Sprintf("★ Client updated: [%s]client(%s) {%d}->{%d}",
		cc.Config.ChainID,
		srcClientId,
		provider.MustGetHeight(srcUpdateHeader.GetHeight()),
		srcUpdateHeader.GetHeight().GetRevisionHeight(),
	))

	return clientState.TrustingPeriod, nil
}

// FindMatchingClient will determine if there exists a client with identical client and consensus states
// to the client which would have been created. Source is the chain that would be adding a client
// which would track the counterparty. Therefore we query source for the existing clients
// and check if any match the counterparty. The counterparty must have a matching consensus state
// to the latest consensus state of a potential match. The provided client state is the client
// state that will be created if there exist no matches.
func (cc *ChainClient) FindMatchingClient(counterparty provider.ChainProvider, clientState ibcexported.ClientState) (string, bool) {
	// TODO: add appropriate offset and limits, along with retries
	clientsResp, err := cc.QueryClients()
	if err != nil {
		if cc.Config.Debug {
			cc.Log(fmt.Sprintf("Error: querying clients on %s failed: %v", cc.Config.ChainID, err))
		}
		return "", false
	}

	for _, identifiedClientState := range clientsResp {
		// unpack any into ibc tendermint client state
		existingClientState, err := castClientStateToTMType(identifiedClientState.ClientState)
		if err != nil {
			return "", false
		}

		tmClientState, ok := clientState.(*tmclient.ClientState)
		if !ok {
			if cc.Config.Debug {
				fmt.Printf("got data of type %T but wanted tmclient.ClientState \n", clientState)
			}
			return "", false
		}

		// check if the client states match
		// NOTE: FrozenHeight.IsZero() is a sanity check, the client to be created should always
		// have a zero frozen height and therefore should never match with a frozen client
		if isMatchingClient(tmClientState, existingClientState) && existingClientState.FrozenHeight.IsZero() {

			// query the latest consensus state of the potential matching client
			consensusStateResp, err := cc.QueryConsensusStateABCI(identifiedClientState.ClientId, existingClientState.GetLatestHeight())
			if err != nil {
				if cc.Config.Debug {
					cc.Log(fmt.Sprintf("Error: failed to query latest consensus state for existing client on chain %s: %v",
						cc.Config.ChainID, err))
				}
				continue
			}

			//nolint:lll
			header, err := counterparty.GetLightSignedHeaderAtHeight(int64(existingClientState.GetLatestHeight().GetRevisionHeight()))
			if err != nil {
				if cc.Config.Debug {
					cc.Log(fmt.Sprintf("Error: failed to query header for chain %s at height %d: %v",
						counterparty.ChainId(), existingClientState.GetLatestHeight().GetRevisionHeight(), err))
				}
				continue
			}

			exportedConsState, err := clienttypes.UnpackConsensusState(consensusStateResp.ConsensusState)
			if err != nil {
				if cc.Config.Debug {
					cc.Log(fmt.Sprintf("Error: failed to consensus state on chain %s: %v", counterparty.ChainId(), err))
				}
				continue
			}
			existingConsensusState, ok := exportedConsState.(*tmclient.ConsensusState)
			if !ok {
				if cc.Config.Debug {
					cc.Log(fmt.Sprintf("Error: consensus state is not tendermint type on chain %s", counterparty.ChainId()))
				}
				continue
			}

			if existingClientState.IsExpired(existingConsensusState.Timestamp, time.Now()) {
				continue
			}

			tmHeader, ok := header.(*tmclient.Header)
			if !ok {
				if cc.Config.Debug {
					fmt.Printf("got data of type %T but wanted tmclient.Header \n", header)
				}
				return "", false
			}

			if isMatchingConsensusState(existingConsensusState, tmHeader.ConsensusState()) {
				// found matching client
				return identifiedClientState.ClientId, true
			}
		}
	}
	return "", false
}

func (cc *ChainClient) QueryConsensusStateABCI(clientID string, height ibcexported.Height) (*clienttypes.QueryConsensusStateResponse, error) {
	key := host.FullConsensusStateKey(clientID, height)

	value, proofBz, proofHeight, err := cc.QueryTendermintProof(int64(height.GetRevisionHeight()), key)
	if err != nil {
		return nil, err
	}

	// check if consensus state exists
	if len(value) == 0 {
		return nil, sdkerrors.Wrap(clienttypes.ErrConsensusStateNotFound, clientID)
	}

	// TODO do we really want to create a new codec? ChainClient exposes proto.Marshaler
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

// isMatchingClient determines if the two provided clients match in all fields
// except latest height. They are assumed to be IBC tendermint light clients.
// NOTE: we don't pass in a pointer so upstream references don't have a modified
// latest height set to zero.
func isMatchingClient(clientStateA, clientStateB *tmclient.ClientState) bool {
	// zero out latest client height since this is determined and incremented
	// by on-chain updates. Changing the latest height does not fundamentally
	// change the client. The associated consensus state at the latest height
	// determines this last check
	clientStateA.LatestHeight = clienttypes.ZeroHeight()
	clientStateB.LatestHeight = clienttypes.ZeroHeight()

	return reflect.DeepEqual(clientStateA, clientStateB)
}

// isMatchingConsensusState determines if the two provided consensus states are
// identical. They are assumed to be IBC tendermint light clients.
func isMatchingConsensusState(consensusStateA, consensusStateB *tmclient.ConsensusState) bool {
	return reflect.DeepEqual(*consensusStateA, *consensusStateB)
}

// queryTMClientState retrieves the latest consensus state for a client in state at a given height
// and unpacks/cast it to tendermint clientstate
func (cc *ChainClient) queryTMClientState(srch int64, srcClientId string) (*tmclient.ClientState, error) {
	clientStateRes, err := cc.QueryClientStateResponse(srch, srcClientId)
	if err != nil {
		return &tmclient.ClientState{}, err
	}

	return castClientStateToTMType(clientStateRes.ClientState)
}

// castClientStateToTMType casts client state to tendermint type
func castClientStateToTMType(cs *codectypes.Any) (*tmclient.ClientState, error) {
	clientStateExported, err := clienttypes.UnpackClientState(cs)
	if err != nil {
		return &tmclient.ClientState{}, err
	}

	// cast from interface to concrete type
	clientState, ok := clientStateExported.(*tmclient.ClientState)
	if !ok {
		return &tmclient.ClientState{},
			fmt.Errorf("error when casting exported clientstate to tendermint type")
	}

	return clientState, nil
}

// WaitForNBlocks blocks until the next block on a given chain
func (cc *ChainClient) WaitForNBlocks(n int64) error {
	var initial int64
	h, err := cc.RPCClient.Status(context.Background())
	if err != nil {
		return err
	}
	if h.SyncInfo.CatchingUp {
		return fmt.Errorf("chain catching up")
	}
	initial = h.SyncInfo.LatestBlockHeight
	for {
		h, err = cc.RPCClient.Status(context.Background())
		if err != nil {
			return err
		}
		if h.SyncInfo.LatestBlockHeight > initial+n {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
}
