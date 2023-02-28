package query

import (
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

// ibc_ParamsRPC returns the distribution params
func ibc_ClientParamsRPC(q *Query) (*clienttypes.QueryClientParamsResponse, error) {
	req := &clienttypes.QueryClientParamsRequest{}
	queryClient := clienttypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.ClientParams(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ibc_ClientStateRPC returns the state of the specified IBC client.
func ibc_ClientStateRPC(q *Query, clientId string) (*clienttypes.QueryClientStateResponse, error) {
	req := &clienttypes.QueryClientStateRequest{ClientId: clientId}
	queryClient := clienttypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.ClientState(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ibc_ClientStatesRPC returns the state of the all IBC clients.
func ibc_ClientStatesRPC(q *Query) (*clienttypes.QueryClientStatesResponse, error) {
	req := &clienttypes.QueryClientStatesRequest{Pagination: q.Options.Pagination}
	queryClient := clienttypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.ClientStates(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ibc_ConsensusStateRPC returns the consensus state of the specified IBC client.
func ibc_ConsensusStateRPC(q *Query, clientId string, height clienttypes.Height) (*clienttypes.QueryConsensusStateResponse, error) {
	req := &clienttypes.QueryConsensusStateRequest{ClientId: clientId}
	if (height.RevisionHeight == 0) && (height.RevisionNumber == 0) {
		req.LatestHeight = true
	} else {
		req.RevisionNumber = height.RevisionNumber
		req.RevisionHeight = height.RevisionHeight
	}

	queryClient := clienttypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.ConsensusState(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ibc_ConsensusStatesRPC returns the consensus states of given IBC client.
func ibc_ConsensusStatesRPC(q *Query, clientId string) (*clienttypes.QueryConsensusStatesResponse, error) {
	req := &clienttypes.QueryConsensusStatesRequest{ClientId: clientId, Pagination: q.Options.Pagination}

	queryClient := clienttypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.ConsensusStates(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ibc_ConnectionRPC returns the state of the specified IBC connection.
func ibc_ConnectionRPC(q *Query, connectionId string) (*connectiontypes.QueryConnectionResponse, error) {
	req := &connectiontypes.QueryConnectionRequest{ConnectionId: connectionId}

	queryClient := connectiontypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.Connection(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ibc_ConnectionsRPC returns the state of all IBC connections.
func ibc_ConnectionsRPC(q *Query) (*connectiontypes.QueryConnectionsResponse, error) {
	req := &connectiontypes.QueryConnectionsRequest{Pagination: q.Options.Pagination}

	queryClient := connectiontypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.Connections(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ibc_ChannelRPC returns the state of the specified IBC channel.
func ibc_ChannelRPC(q *Query, channelId string, portId string) (*channeltypes.QueryChannelResponse, error) {
	req := &channeltypes.QueryChannelRequest{PortId: portId, ChannelId: channelId}

	queryClient := channeltypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.Channel(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ibc_ChannelsRPC returns the state of all IBC channels.
func ibc_ChannelsRPC(q *Query) (*channeltypes.QueryChannelsResponse, error) {
	req := &channeltypes.QueryChannelsRequest{Pagination: q.Options.Pagination}

	queryClient := channeltypes.NewQueryClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()
	res, err := queryClient.Channels(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
