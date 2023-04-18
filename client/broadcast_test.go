package client

import (
	"context"
	"errors"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
)

var (
	errExpected = errors.New("expected unexpected error ;)")
)

type myFakeMsg struct {
	Value string
}

func (m myFakeMsg) Reset()                       {}
func (m myFakeMsg) ProtoMessage()                {}
func (m myFakeMsg) String() string               { return "doesn't matter" }
func (m myFakeMsg) ValidateBasic() error         { return nil }
func (m myFakeMsg) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{sdk.AccAddress(`hello`)} }

type myFakeTx struct {
	msgs []myFakeMsg
}

func (m myFakeTx) GetMsgs() (msgs []sdk.Msg) {
	for _, msg := range m.msgs {
		msgs = append(msgs, msg)
	}
	return
}
func (m myFakeTx) ValidateBasic() error   { return nil }
func (m myFakeTx) AsAny() *codectypes.Any { return &codectypes.Any{} }

type fakeBroadcaster struct {
	tx            func(context.Context, []byte, bool) (*ctypes.ResultTx, error)
	broadcastSync func(context.Context, tmtypes.Tx) (*ctypes.ResultBroadcastTx, error)
}

func (f fakeBroadcaster) Tx(ctx context.Context, hash []byte, prove bool) (*ctypes.ResultTx, error) {
	if f.tx == nil {
		return nil, nil
	}
	return f.tx(ctx, hash, prove)
}

func (f fakeBroadcaster) BroadcastTxSync(ctx context.Context, tx tmtypes.Tx) (*ctypes.ResultBroadcastTx, error) {
	if f.broadcastSync == nil {
		return nil, nil
	}
	return f.broadcastSync(ctx, tx)
}

func TestBroadcast(t *testing.T) {
	ctx := context.Background()

	for _, tt := range []struct {
		name        string
		broadcaster fakeBroadcaster
		txDecoder   sdk.TxDecoder
		txBytes     []byte
		waitTimeout time.Duration
		expectedRes *sdk.TxResponse
		expectedErr error
	}{
		{
			name: "simple success returns result",
			broadcaster: fakeBroadcaster{
				broadcastSync: func(_ context.Context, _ tmtypes.Tx) (*ctypes.ResultBroadcastTx, error) {
					return &ctypes.ResultBroadcastTx{
						Code: 1,
						Hash: []byte(`123bob`),
					}, nil
				},
				tx: func(_ context.Context, hash []byte, _ bool) (*ctypes.ResultTx, error) {
					assert.Equal(t, []byte(`123bob`), hash)
					return &ctypes.ResultTx{}, nil
				},
			},
			txDecoder: func(txBytes []byte) (sdk.Tx, error) {
				return myFakeTx{
					[]myFakeMsg{{"hello"}},
				}, nil
			},
			expectedRes: &sdk.TxResponse{
				Tx: &codectypes.Any{},
			},
		},
		{
			name:        "success but timed out while waiting for tx",
			waitTimeout: time.Microsecond,
			broadcaster: fakeBroadcaster{
				broadcastSync: func(_ context.Context, _ tmtypes.Tx) (*ctypes.ResultBroadcastTx, error) {
					return &ctypes.ResultBroadcastTx{
						Code: 1,
						Hash: []byte(`123bob`),
					}, nil
				},
				tx: func(_ context.Context, hash []byte, _ bool) (*ctypes.ResultTx, error) {
					<-time.After(time.Second)
					// return doesn't matter because it will timeout before return will make sense
					return nil, nil
				},
			},
			expectedErr: ErrTimeoutAfterWaitingForTxBroadcast,
		},
		{
			name: "broadcasting returns an error",
			broadcaster: fakeBroadcaster{
				broadcastSync: func(_ context.Context, _ tmtypes.Tx) (*ctypes.ResultBroadcastTx, error) {
					return nil, errExpected
				},
			},
			expectedErr: errExpected,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			duration := 1 * time.Second
			if tt.waitTimeout > 0 {
				duration = tt.waitTimeout
			}
			gotRes, gotErr := broadcastTx(
				ctx,
				tt.broadcaster,
				tt.txDecoder,
				tt.txBytes,
				duration,
			)
			if gotRes != nil {
				// Ignoring timestamp for tests
				gotRes.Timestamp = ""
			}
			assert.Equal(t, tt.expectedRes, gotRes)
			assert.ErrorIs(t, gotErr, tt.expectedErr)
		})
	}

}
