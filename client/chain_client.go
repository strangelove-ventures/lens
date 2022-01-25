package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/relayer/relayer/provider"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"
	provtypes "github.com/tendermint/tendermint/light/provider"
	prov "github.com/tendermint/tendermint/light/provider/http"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	libclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"gopkg.in/yaml.v3"
)

type ChainClient struct {
	Config         *ChainClientConfig
	Keybase        keyring.Keyring
	KeyringOptions []keyring.Option
	RPCClient      rpcclient.Client
	LightProvider  provtypes.Provider
	Input          io.Reader
	Output         io.Writer
	// TODO: GRPC Client type?

	Codec  Codec
	Logger log.Logger
}

func NewChainClient(ccc *ChainClientConfig, homepath string, input io.Reader, output io.Writer, kro ...keyring.Option) (*ChainClient, error) {
	ccc.KeyDirectory = keysDir(homepath, ccc.ChainID)
	cc := &ChainClient{
		KeyringOptions: kro,
		Config:         ccc,
		Input:          input,
		Output:         output,
		Codec:          MakeCodec(ccc.Modules),
		Logger:         log.NewTMLogger(log.NewSyncWriter(output)),
	}
	if err := cc.Init(); err != nil {
		return nil, err
	}
	return cc, nil
}

func (cc *ChainClient) Init() error {
	// TODO: test key directory and return error if not created
	keybase, err := keyring.New(cc.Config.ChainID, cc.Config.KeyringBackend, cc.Config.KeyDirectory, cc.Input, cc.KeyringOptions...)
	if err != nil {
		return err
	}
	// TODO: figure out how to deal with input or maybe just make all keyring backends test?

	timeout, _ := time.ParseDuration(cc.Config.Timeout)
	rpcClient, err := NewRPCClient(cc.Config.RPCAddr, timeout)
	if err != nil {
		return err
	}

	lightprovider, err := prov.New(cc.Config.ChainID, cc.Config.RPCAddr)
	if err != nil {
		return err
	}

	cc.RPCClient = rpcClient
	cc.LightProvider = lightprovider
	cc.Keybase = keybase

	return nil
}

// Log takes a string and logs the data
func (cc *ChainClient) Log(s string) {
	cc.Logger.Info(s)
}

// TODO: actually do something different here have a couple of levels of verbosity
func (cc *ChainClient) PrintTxResponse(res *sdk.TxResponse) error {
	return cc.PrintObject(res)
}

func (cc *ChainClient) HandleAndPrintMsgSend(res *sdk.TxResponse, err error) error {
	if err != nil {
		if res != nil {
			return fmt.Errorf("failed to withdraw rewards: code(%d) msg(%s)", res.Code, res.Logs)
		}
		return fmt.Errorf("failed to withdraw rewards: err(%w)", err)
	}
	return cc.PrintTxResponse(res)
}

// LogFailedTx takes the transaction and the messages to create it and logs the appropriate data
func (cc *ChainClient) LogFailedTx(res *provider.RelayerTxResponse, err error, msgs []provider.RelayerMessage) {
	if cc.Config.Debug {
		cc.Log(fmt.Sprintf("- [%s] -> failed sending transaction:", cc.ChainId()))
		for _, msg := range msgs {
			_ = cc.PrintObject(msg)
		}
	}

	if err != nil {
		cc.Logger.Error(fmt.Errorf("- [%s] -> err(%v)", cc.ChainId(), err).Error())
		if res == nil {
			return
		}
	}

	if res.Code != 0 && res.Data != "" {
		cc.Log(fmt.Sprintf("✘ [%s]@{%d} - msg(%s) err(%d:%s)", cc.ChainId(), res.Height, getMsgTypes(msgs), res.Code, res.Data))
	}

	if cc.Config.Debug && res != nil {
		cc.Log("- transaction response:")
		_ = cc.PrintObject(res)
	}
}

func getMsgTypes(msgs []provider.RelayerMessage) string {
	var out string
	for i, msg := range msgs {
		out += fmt.Sprintf("%d:%s,", i, msg.Type())
	}
	return strings.TrimSuffix(out, ",")
}

// LogSuccessTx take the transaction and the messages to create it and logs the appropriate data
func (cc *ChainClient) LogSuccessTx(res *sdk.TxResponse, msgs []provider.RelayerMessage) {
	cc.Logger.Info(fmt.Sprintf("✔ [%s]@{%d} - msg(%s) hash(%s)", cc.ChainId(), res.Height, getMsgTypes(msgs), res.TxHash))
}

func (cc *ChainClient) PrintObject(res interface{}) error {
	var (
		bz  []byte
		err error
	)
	switch cc.Config.OutputFormat {
	case "json":
		if m, ok := res.(proto.Message); ok {
			bz, err = cc.MarshalProto(m)
		} else {
			bz, err = json.Marshal(res)
		}
		if err != nil {
			return err
		}
	case "indent":
		if m, ok := res.(proto.Message); ok {
			bz, err = cc.MarshalProto(m)
			if err != nil {
				return err
			}
			buf := bytes.NewBuffer([]byte{})
			if err = json.Indent(buf, bz, "", "  "); err != nil {
				return err
			}
			bz = buf.Bytes()
		} else {
			bz, err = json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}
		}
	case "yaml":
		bz, err = yaml.Marshal(res)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown output type: %s", cc.Config.OutputFormat)
	}
	fmt.Fprint(cc.Output, string(bz), "\n")
	return nil
}

func (cc *ChainClient) MarshalProto(res proto.Message) ([]byte, error) {
	return cc.Codec.Marshaler.MarshalJSON(res)
}

func (cc *ChainClient) GetKeyAddress() (sdk.AccAddress, error) {
	info, err := cc.Keybase.Key(cc.Config.Key)
	if err != nil {
		return nil, err
	}
	return info.GetAddress(), nil
}

func NewRPCClient(addr string, timeout time.Duration) (*rpchttp.HTTP, error) {
	httpClient, err := libclient.DefaultHTTPClient(addr)
	if err != nil {
		return nil, err
	}
	httpClient.Timeout = timeout
	rpcClient, err := rpchttp.NewWithClient(addr, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}
	return rpcClient, nil
}

// AccountFromKeyOrAddress returns an account from either a key or an address
// if empty string is passed in this returns the default key's address
func (cc *ChainClient) AccountFromKeyOrAddress(keyOrAddress string) (out sdk.AccAddress, err error) {
	switch {
	case keyOrAddress == "":
		out, err = cc.GetKeyAddress()
	case cc.KeyExists(keyOrAddress):
		cc.Config.Key = keyOrAddress
		out, err = cc.GetKeyAddress()
	default:
		out, err = cc.DecodeBech32AccAddr(keyOrAddress)
	}
	return
}

func keysDir(home, chainID string) string {
	return path.Join(home, "keys", chainID)
}
