package client

import (
	"time"

	"github.com/cosmos/cosmos-sdk/types/module"
)

type ChainClientConfig struct {
	Key            string                  `json:"key" yaml:"key"`
	ChainID        string                  `json:"chain-id" yaml:"chain-id"`
	RPCAddr        string                  `json:"rpc-addr" yaml:"rpc-addr"`
	GRPCAddr       string                  `json:"grpc-addr" yaml:"grpc-addr"`
	AccountPrefix  string                  `json:"account-prefix" yaml:"account-prefix"`
	KeyringBackend string                  `json:"keyring-backend" yaml:"keyring-backend"`
	GasAdjustment  float64                 `json:"gas-adjustment" yaml:"gas-adjustment"`
	GasPrices      string                  `json:"gas-prices" yaml:"gas-prices"`
	KeyDirectory   string                  `json:"key-directory" yaml:"key-directory"`
	Debug          bool                    `json:"debug" yaml:"debug"`
	Timeout        string                  `json:"timeout" yaml:"timeout"`
	OutputFormat   string                  `json:"output-format" yaml:"output-format"`
	BroadcastMode  string                  `json:"broadcast-mode" yaml:"broadcast-mode"`
	SignModeStr    string                  `json:"sign-mode" yaml:"sign-mode"`
	Modules        []module.AppModuleBasic `json:"-" yaml:"-"`
}

func (ccc *ChainClientConfig) Validate() error {
	if _, err := time.ParseDuration(ccc.Timeout); err != nil {
		return err
	}
	return nil
}
