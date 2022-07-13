package chain_registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/spf13/viper"
	"github.com/strangelove-ventures/lens/client"
	"go.uber.org/zap"
)

type res_chains struct {
	Repository struct {
		URL       string `json:"url"`
		Branch    string `json:"branch"`
		Commit    string `json:"commit"`
		Timestamp int    `json:"timestamp"`
	} `json:"repository"`
	Chains []struct {
		Name        string `json:"name"`
		Path        string `json:"path"`
		ChainName   string `json:"chain_name"`
		NetworkType string `json:"network_type"`
		PrettyName  string `json:"pretty_name"`
		ChainID     string `json:"chain_id"`
		Status      string `json:"status"`
		Symbol      string `json:"symbol"`
		Denom       string `json:"denom"`
		Decimals    int    `json:"decimals"`
		Image       string `json:"image"`
		Height      int    `json:"height"`
		BestApis    struct {
			Rest []struct {
				Address string `json:"address"`
			} `json:"rest"`
			RPC []struct {
				Address  string `json:"address"`
				Provider string `json:"provider,omitempty"`
			} `json:"rpc"`
		} `json:"best_apis"`
		Params struct {
			Authz           bool    `json:"authz"`
			BondedTokens    string  `json:"bonded_tokens"`
			TotalSupply     string  `json:"total_supply"`
			ActualBlockTime float64 `json:"actual_block_time"`
			CalculatedApr   float64 `json:"calculated_apr"`
		} `json:"params"`
	} `json:"chains"`
}

type res_chain struct {
	Repository struct {
		URL       string `json:"url"`
		Branch    string `json:"branch"`
		Commit    string `json:"commit"`
		Timestamp int    `json:"timestamp"`
	} `json:"repository"`
	Chain struct {
		Schema       string `json:"$schema"`
		ChainName    string `json:"chain_name"`
		Status       string `json:"status"`
		NetworkType  string `json:"network_type"`
		Updatelink   string `json:"updatelink"`
		PrettyName   string `json:"pretty_name"`
		ChainID      string `json:"chain_id"`
		Bech32Prefix string `json:"bech32_prefix"`
		DaemonName   string `json:"daemon_name"`
		NodeHome     string `json:"node_home"`
		Genesis      struct {
			GenesisURL string `json:"genesis_url"`
		} `json:"genesis"`
		KeyAlgos []string `json:"key_algos"`
		Slip44   int      `json:"slip44"`
		Fees     struct {
			FeeTokens []struct {
				Denom            string  `json:"denom"`
				FixedMinGasPrice int     `json:"fixed_min_gas_price"`
				LowGasPrice      int     `json:"low_gas_price"`
				AverageGasPrice  float64 `json:"average_gas_price"`
				HighGasPrice     float64 `json:"high_gas_price"`
			} `json:"fee_tokens"`
		} `json:"fees"`
		Codebase struct {
			GitRepo            string   `json:"git_repo"`
			RecommendedVersion string   `json:"recommended_version"`
			CompatibleVersions []string `json:"compatible_versions"`
			Binaries           struct {
				LinuxAmd64 string `json:"linux/amd64"`
				LinuxArm64 string `json:"linux/arm64"`
			} `json:"binaries"`
		} `json:"codebase"`
		Peers struct {
			Seeds []struct {
				ID       string `json:"id"`
				Address  string `json:"address"`
				Provider string `json:"provider"`
			} `json:"seeds"`
			PersistentPeers []struct {
				ID       string `json:"id"`
				Address  string `json:"address"`
				Provider string `json:"provider"`
			} `json:"persistent_peers"`
		} `json:"peers"`
		Apis struct {
			RPC []struct {
				Address  string `json:"address"`
				Provider string `json:"provider"`
			} `json:"rpc"`
			Rest []struct {
				Address  string `json:"address"`
				Provider string `json:"provider"`
			} `json:"rest"`
			Grpc []struct {
				Address  string `json:"address"`
				Provider string `json:"provider"`
			} `json:"grpc"`
		} `json:"apis"`
		Explorers []struct {
			Kind        string `json:"kind"`
			URL         string `json:"url"`
			TxPage      string `json:"tx_page"`
			AccountPage string `json:"account_page,omitempty"`
		} `json:"explorers"`
		LogoURIs struct {
			Png string `json:"png"`
		} `json:"logo_URIs"`
		Name        string `json:"name"`
		Path        string `json:"path"`
		Symbol      string `json:"symbol"`
		Denom       string `json:"denom"`
		Decimals    int    `json:"decimals"`
		CoingeckoID string `json:"coingecko_id"`
		Image       string `json:"image"`
		Height      int    `json:"height"`
		BestApis    struct {
			Rest []struct {
				Address  string `json:"address"`
				Provider string `json:"provider"`
			} `json:"rest"`
			RPC []struct {
				Address  string `json:"address"`
				Provider string `json:"provider"`
			} `json:"rpc"`
		} `json:"best_apis"`
	} `json:"chain"`
}

type ChainRegistryAPI struct {
	log *zap.Logger
}

func NewChainRegistryAPI(log *zap.Logger) ChainRegistryAPI {
	return ChainRegistryAPI{log: log}
}

func (c ChainRegistryAPI) ListChains(ctx context.Context) ([]string, error) {
	var chains []string
	req, err := http.NewRequest(http.MethodGet, "https://chains.cosmos.directory/", nil)
	if err != nil {
		return chains, err
	}
	req = req.WithContext(ctx)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return chains, err
	}
	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var apiResponse res_chains
	json.Unmarshal([]byte(bytes), &apiResponse)

	for _, value := range apiResponse.Chains {
		chains = append(chains, value.ChainName)
	}
	return chains, nil
}

func (c ChainRegistryAPI) AddChainInfo(ctx context.Context, name string) (*client.ChainClientConfig, error) {
	var chainInfo res_chain
	var chainClientConfig *client.ChainClientConfig

	chainInfo, err := getChainInfo(ctx, name)
	if err != nil {
		c.log.Info(
			"chain not found",
			zap.String("chain_name", name),
		)
		return chainClientConfig, err
	}

	rpc, err := GetRandomRPCEndpoint(ctx, chainInfo.Chain.BestApis.RPC)
	if err != nil {
		c.log.Info(
			"error getting/checking rpc",
			zap.String("chain_name", name),
		)
		return nil, err
	}

	debug := viper.GetBool("debug")
	home := viper.GetString("home")

	var gasPrices string
	baseDenom := chainInfo.Chain.Denom
	if baseDenom != "" {
		gasPrices = fmt.Sprintf("%.2f%s", 0.01, baseDenom)
	}

	return &client.ChainClientConfig{
		Key:            "default",
		ChainID:        chainInfo.Chain.ChainID,
		RPCAddr:        rpc,
		AccountPrefix:  chainInfo.Chain.Bech32Prefix,
		KeyringBackend: "test",
		GasAdjustment:  1.2,
		GasPrices:      gasPrices,
		KeyDirectory:   home,
		Debug:          debug,
		Timeout:        "20s",
		OutputFormat:   "json",
		SignModeStr:    "direct",
	}, nil
}

func getChainInfo(ctx context.Context, name string) (chainInfo res_chain, err error) {
	getUrl := ("https://chains.cosmos.directory/" + name)
	req, err := http.NewRequest(http.MethodGet, (getUrl), nil)
	if err != nil {
		return chainInfo, err
	}
	req = req.WithContext(ctx)
	client := &http.Client{}
	res, err := client.Do(req)
	if res.StatusCode == 404 {
		// c.log.Info(
		// 	"chain not found",
		// 	zap.String("chain_name", name),
		// 	zap.String("GET", getUrl),
		// )
		fmt.Printf("chain %s not found", name)
		return chainInfo, nil
	}
	if err != nil {
		return chainInfo, err
	}
	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	result := chainInfo
	if err := json.Unmarshal([]byte(bytes), &result); err != nil {
		return chainInfo, err
	}
	return result, nil

}

func (c ChainRegistryAPI) SourceLink() string {
	return ("https://chains.cosmos.directory/")
}
