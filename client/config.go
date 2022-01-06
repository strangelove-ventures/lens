package client

import (
	"github.com/cosmos/relayer/v2/relayer/provider"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/types/module"
)

var (
	_ provider.ProviderConfig = &ChainClientConfig{}
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
	TrustingPeriod string                  `json:"omitempty" yaml:"omitempty"`
	KeyDirectory   string                  `json:"key-directory" yaml:"key-directory"`
	Debug          bool                    `json:"debug" yaml:"debug"`
	Timeout        string                  `json:"timeout" yaml:"timeout"`
	OutputFormat   string                  `json:"output-format" yaml:"output-format"`
	SignModeStr    string                  `json:"sign-mode" yaml:"sign-mode"`
	Modules        []module.AppModuleBasic `json:"-" yaml:"-"`
}

func (ccc *ChainClientConfig) NewProvider(homepath string, debug bool) (provider.ChainProvider, error) {
	if err := ccc.Validate(); err != nil {
		return nil, err
	}
	p, err := NewChainClient(ccc, os.Stdin, os.Stdout)
	if err != nil {
		return nil, err
	}
	return p, err
}

func (ccc *ChainClientConfig) Validate() error {
	if _, err := time.ParseDuration(ccc.Timeout); err != nil {
		return err
	}
	return nil
}

func GetCosmosHubConfig(keyHome string, debug bool) *ChainClientConfig {
	return &ChainClientConfig{
		Key:            "default",
		ChainID:        "cosmoshub-4",
		RPCAddr:        "https://cosmoshub-4.technofractal.com:443",
		GRPCAddr:       "https://gprc.cosmoshub-4.technofractal.com:443",
		AccountPrefix:  "cosmos",
		KeyringBackend: "test",
		GasAdjustment:  1.2,
		GasPrices:      "0.01uatom",
		TrustingPeriod: "336h",
		KeyDirectory:   keyHome,
		Debug:          debug,
		Timeout:        "20s",
		OutputFormat:   "json",
		SignModeStr:    "direct",
	}
}

func GetOsmosisConfig(keyHome string, debug bool) *ChainClientConfig {
	return &ChainClientConfig{
		Key:            "default",
		ChainID:        "osmosis-1",
		RPCAddr:        "https://osmosis-1.technofractal.com:443",
		GRPCAddr:       "https://gprc.osmosis-1.technofractal.com:443",
		AccountPrefix:  "osmo",
		KeyringBackend: "test",
		GasAdjustment:  1.2,
		GasPrices:      "0.01uosmo",
		TrustingPeriod: "300h",
		KeyDirectory:   keyHome,
		Debug:          debug,
		Timeout:        "20s",
		OutputFormat:   "json",
		SignModeStr:    "direct",
	}
}

func GetTestClient() *ChainClient {
	cl, _ := NewChainClient(GetCosmosHubConfig("/tmp", true), nil, nil)
	return cl
}
