package query

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/query"
	txTypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/strangelove-ventures/lens/client"
)

// TxRPC Get Transactions for the given block height.
// Other query options can be specified with the GetTxsEventRequest.
//
// RPC endpoint is defined in cosmos-sdk: proto/cosmos/tx/v1beta1/service.proto,
// See GetTxsEvent(GetTxsEventRequest) returns (GetTxsEventResponse)
func TxsAtHeightRPC(q *Query, height int64, codec client.Codec) (*txTypes.GetTxsEventResponse, error) {
	if q.Options.Pagination == nil {
		pagination := &query.PageRequest{Limit: 100}
		q.Options.Pagination = pagination
	}
	orderBy := txTypes.OrderBy_ORDER_BY_UNSPECIFIED

	req := &txTypes.GetTxsEventRequest{Events: []string{"tx.height=" + fmt.Sprintf("%d", height)}, Pagination: q.Options.Pagination, OrderBy: orderBy}
	return TxsRPC(q, req, codec)
}

// TxRPC Get Transactions for the given block height.
// Other query options can be specified with the GetTxsEventRequest.
//
// RPC endpoint is defined in cosmos-sdk: proto/cosmos/tx/v1beta1/service.proto,
// See GetTxsEvent(GetTxsEventRequest) returns (GetTxsEventResponse)
func TxsRPC(q *Query, req *txTypes.GetTxsEventRequest, codec client.Codec) (*txTypes.GetTxsEventResponse, error) {
	queryClient := txTypes.NewServiceClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()

	res, err := queryClient.GetTxsEvent(ctx, req)
	if err != nil {
		return nil, err
	}

	for _, tx := range res.GetTxs() {
		tx.UnpackInterfaces(codec.InterfaceRegistry)
	}

	return res, nil
}
