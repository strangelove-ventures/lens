package query

import (
	"fmt"

	txTypes "github.com/cosmos/cosmos-sdk/types/tx"
)

// TxRPC Get Transactions for the given block height.
// Other query options can be specified with the GetTxsEventRequest.
//
// RPC endpoint is defined in cosmos-sdk: proto/cosmos/tx/v1beta1/service.proto,
// See GetTxsEvent(GetTxsEventRequest) returns (GetTxsEventResponse)
func TxsAtHeightRPC(q *Query, height int64) (*txTypes.GetTxsEventResponse, error) {
	req := &txTypes.GetTxsEventRequest{Events: []string{"events=" + fmt.Sprintf("%d", height)}}
	return TxsRPC(q, req)
}

// TxRPC Get Transactions for the given block height.
// Other query options can be specified with the GetTxsEventRequest.
//
// RPC endpoint is defined in cosmos-sdk: proto/cosmos/tx/v1beta1/service.proto,
// See GetTxsEvent(GetTxsEventRequest) returns (GetTxsEventResponse)
func TxsRPC(q *Query, req *txTypes.GetTxsEventRequest) (*txTypes.GetTxsEventResponse, error) {
	queryClient := txTypes.NewServiceClient(q.Client)
	ctx, cancel := q.GetQueryContext()
	defer cancel()

	res, err := queryClient.GetTxsEvent(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

//grpcRes, err := s.queryClient.GetTx(context.Background(), tc.req)
