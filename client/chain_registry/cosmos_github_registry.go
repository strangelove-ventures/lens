package chain_registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/google/go-github/v43/github"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/strangelove-ventures/lens/client"
	"go.uber.org/zap"
)

type assetListJson struct {
	Schema  string `json:"$schema"`
	ChainID string `json:"chain_id"`
	Assets  []struct {
		Description string `json:"description"`
		DenomUnits  []struct {
			Denom    string `json:"denom"`
			Exponent int    `json:"exponent"`
		} `json:"denom_units"`
		Base     string `json:"base"`
		Name     string `json:"name"`
		Display  string `json:"display"`
		Symbol   string `json:"symbol"`
		LogoURIs struct {
			Png string `json:"png"`
			Svg string `json:"svg"`
		} `json:"logo_URIs"`
		CoingeckoID string `json:"coingecko_id"`
	} `json:"assets"`
}

type ChainJson struct {
	// log *zap.Logger

	Schema       string `json:"$schema"`
	ChainName    string `json:"chain_name"`
	Status       string `json:"status"`
	NetworkType  string `json:"network_type"`
	PrettyName   string `json:"pretty_name"`
	ChainID      string `json:"chain_id"`
	Bech32Prefix string `json:"bech32_prefix"`
	DaemonName   string `json:"daemon_name"`
	NodeHome     string `json:"node_home"`
	Genesis      struct {
		GenesisURL string `json:"genesis_url"`
	} `json:"genesis"`
	Slip44   int `json:"slip44"`
	Codebase struct {
		GitRepo            string   `json:"git_repo"`
		RecommendedVersion string   `json:"recommended_version"`
		CompatibleVersions []string `json:"compatible_versions"`
	} `json:"codebase"`
	Peers struct {
		Seeds []struct {
			ID       string `json:"id"`
			Address  string `json:"address"`
			Provider string `json:"provider,omitempty"`
		} `json:"seeds"`
		PersistentPeers []struct {
			ID      string `json:"id"`
			Address string `json:"address"`
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
	} `json:"apis"`
}

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
func (c CosmosGithubRegistry) AddChainInfo(ctx context.Context, name string) (*client.ChainClientConfig, error) {
	var chainClientConfig *client.ChainClientConfig
	chainInfo, err := getChainJson(ctx, name)
	if err != nil {
		c.log.Info(
			"chain not found",
			zap.String("chain_name", name),
		)
		return chainClientConfig, err
	}
	var baseDenom string
	baseDenom, err = getBaseDenom(ctx, name)
	if err != nil {
		c.log.Info(
			"base denom not found",
			zap.String("chain_name", name),
		)
		return chainClientConfig, err
	}

	rpc, err := GetRandomRPCEndpoint(ctx, chainInfo.Apis.RPC)
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
	if baseDenom != "" {
		gasPrices = fmt.Sprintf("%.2f%s", 0.01, baseDenom)
	}

	return &client.ChainClientConfig{
		Key:            "default",
		ChainID:        chainInfo.ChainID,
		RPCAddr:        rpc,
		AccountPrefix:  chainInfo.Bech32Prefix,
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

func getChainJson(ctx context.Context, name string) (chainJson ChainJson, err error) {
	client := github.NewClient(http.DefaultClient)

	chainFileName := path.Join(name, "chain.json")
	fileContent, _, res, err := client.Repositories.GetContents(
		ctx,
		"cosmos",
		"chain-registry",
		chainFileName,
		&github.RepositoryContentGetOptions{})
	if res.StatusCode == 404 {
		return chainJson, errors.Wrapf(err, "chain not found")
	}
	if err != nil || res.StatusCode != 200 {
		return chainJson, errors.Wrap(err, fmt.Sprintf("error fetching %s", chainFileName))
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return chainJson, err
	}

	result := chainJson
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return chainJson, err
	}

	return result, nil
}

func getBaseDenom(ctx context.Context, name string) (string, error) {
	baseDenom := ""
	cl := github.NewClient(http.DefaultClient)

	chainFileName := path.Join(name, "assetlist.json")
	ch, _, res, err := cl.Repositories.GetContents(
		ctx,
		"cosmos",
		"chain-registry",
		chainFileName,
		&github.RepositoryContentGetOptions{})
	if err != nil || res.StatusCode != 200 {
		return baseDenom, err
	}

	content, err := ch.GetContent()
	if err != nil {
		return baseDenom, err
	}

	var assetListJson assetListJson
	if err := json.Unmarshal([]byte(content), &assetListJson); err != nil {
		return baseDenom, err
	}

	if len(assetListJson.Assets) > 0 {
		baseDenom = assetListJson.Assets[0].Base
	}

	return baseDenom, nil
}

func (c CosmosGithubRegistry) SourceLink() string {
	return "https://github.com/cosmos/chain-registry"
}
