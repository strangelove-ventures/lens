package chain_registry

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"go.uber.org/zap"
)

type ChainRegistryAPI struct {
	log *zap.Logger
}

func NewChainRegistryAPI(log *zap.Logger) ChainRegistryAPI {
	return ChainRegistryAPI{log: log}
}

type Chains struct {
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
	// var theChain Chains
	var theChain Chains
	json.Unmarshal([]byte(bytes), &theChain)

	for _, value := range theChain.Chains {
		chains = append(chains, value.ChainName)
	}
	return chains, nil
}

func (c ChainRegistryAPI) GetChain(ctx context.Context, name string) (ChainInfo, error) {
	req, err := http.NewRequest(http.MethodGet, ("https://chains.cosmos.directory/" + name), nil)
	if err != nil {
		return ChainInfo{}, err
	}
	req = req.WithContext(ctx)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return ChainInfo{}, err
	}
	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	result := NewChainInfo(c.log.With(zap.String("chain_name", name)))
	if err := json.Unmarshal([]byte(bytes), &result); err != nil {
		return ChainInfo{}, err
	}
	return result, nil

}

func (c ChainRegistryAPI) SourceLink() string {
	return "https://github.com/cosmos/chain-registry"
}
