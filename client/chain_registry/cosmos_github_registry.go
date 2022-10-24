package chain_registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v43/github"
	"go.uber.org/zap"
)

type CosmosGithubRegistry struct {
	log *zap.Logger
}

func NewCosmosGithubRegistry(log *zap.Logger) CosmosGithubRegistry {
	return CosmosGithubRegistry{log: log}
}

func (c CosmosGithubRegistry) ListChains(ctx context.Context) ([]string, error) {
	client := github.NewClient(http.DefaultClient)
	var chains []string

	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	tree, res, err := client.Git.GetTree(
		ctx,
		"cosmos",
		"chain-registry",
		"master",
		false)
	if err != nil || res.StatusCode != 200 {
		return chains, err
	}

	for _, entry := range tree.Entries {
		if *entry.Type == "tree" && !strings.HasPrefix(*entry.Path, ".") {
			chains = append(chains, *entry.Path)
		}
	}
	return chains, nil
}

func (c CosmosGithubRegistry) GetChain(ctx context.Context, name string) (ChainInfo, error) {
	chainRegURL := fmt.Sprintf("https://raw.githubusercontent.com/cosmos/chain-registry/master/%s/chain.json", name)

	res, err := http.Get(chainRegURL)
	if err != nil {
		return ChainInfo{}, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return ChainInfo{}, fmt.Errorf("chain not found on registry: response code: %d: GET failed: %s", res.StatusCode, chainRegURL)
	}
	if res.StatusCode != http.StatusOK {
		return ChainInfo{}, fmt.Errorf("response code: %d: GET failed: %s", res.StatusCode, chainRegURL)
	}

	result := NewChainInfo(c.log.With(zap.String("chain_name", name)))
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ChainInfo{}, err
	}
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return ChainInfo{}, err
	}
	return result, nil
}

func (c CosmosGithubRegistry) SourceLink() string {
	return "https://github.com/cosmos/chain-registry"
}
