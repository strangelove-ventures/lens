package chain_registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/google/go-github/v43/github"
	"github.com/spf13/viper"
	"github.com/strangelove-ventures/lens/client"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type ChainInfo struct {
	log *zap.Logger

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

// NewChainInfo returns a ChainInfo that is uninitialized other than the provided zap.Logger.
// Typically, the caller will unmarshal JSON content into the ChainInfo after initialization.
func NewChainInfo(log *zap.Logger) ChainInfo {
	return ChainInfo{log: log}
}

func (c ChainInfo) GetAllRPCEndpoints() (out []string, err error) {
	for _, endpoint := range c.Apis.RPC {
		u, err := url.Parse(endpoint.Address)
		if err != nil {
			return nil, err
		}

		var port string
		if u.Port() == "" {
			switch u.Scheme {
			case "https":
				port = "443"
			case "http":
				port = "80"
			default:
				return nil, fmt.Errorf("invalid or unsupported url scheme: %v", u.Scheme)
			}
		} else {
			port = u.Port()
		}

		out = append(out, fmt.Sprintf("%s://%s:%s%s", u.Scheme, u.Hostname(), port, u.Path))
	}

	return
}

func IsHealthyRPC(ctx context.Context, endpoint string) error {
	cl, err := client.NewRPCClient(endpoint, 5*time.Second)
	if err != nil {
		return err
	}
	stat, err := cl.Status(ctx)
	if err != nil {
		return err
	}

	if stat.SyncInfo.CatchingUp {
		return errors.New("still catching up")
	}

	return nil
}

func (c ChainInfo) GetRPCEndpoints(ctx context.Context) (out []string, err error) {
	allRPCEndpoints, err := c.GetAllRPCEndpoints()
	if err != nil {
		return nil, err
	}

	var eg errgroup.Group
	var endpoints []string
	for _, endpoint := range allRPCEndpoints {
		endpoint := endpoint
		eg.Go(func() error {
			err := IsHealthyRPC(ctx, endpoint)
			if err != nil {
				c.log.Info(
					"Ignoring endpoint due to error",
					zap.String("endpoint", endpoint),
					zap.Error(err),
				)
				return nil
			}

			c.log.Info("Verified healthy endpoint", zap.String("endpoint", endpoint))
			endpoints = append(endpoints, endpoint)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return endpoints, nil
}

func (c ChainInfo) GetRandomRPCEndpoint(ctx context.Context) (string, error) {
	rpcs, err := c.GetRPCEndpoints(ctx)
	if err != nil {
		return "", err
	}

	if len(rpcs) == 0 {
		return "", fmt.Errorf("no working RPCs found")
	}

	randomGenerator := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rpcs[randomGenerator.Intn(len(rpcs))], nil
}

func (c ChainInfo) GetAssetList(ctx context.Context) (AssetList, error) {
	cl := github.NewClient(http.DefaultClient)

	chainFileName := path.Join(c.ChainName, "assetlist.json")
	ch, _, res, err := cl.Repositories.GetContents(
		ctx,
		"cosmos",
		"chain-registry",
		chainFileName,
		&github.RepositoryContentGetOptions{})
	if err != nil || res.StatusCode != 200 {
		return AssetList{}, err
	}

	content, err := ch.GetContent()
	if err != nil {
		return AssetList{}, err
	}

	var assetList AssetList
	if err := json.Unmarshal([]byte(content), &assetList); err != nil {
		return AssetList{}, err
	}
	return assetList, nil
}

func (c ChainInfo) GetChainConfig(ctx context.Context) (*client.ChainClientConfig, error) {
	debug := viper.GetBool("debug")
	home := viper.GetString("home")

	assetList, err := c.GetAssetList(ctx)
	if err != nil {
		return nil, err
	}

	var gasPrices string
	if len(assetList.Assets) > 0 {
		gasPrices = fmt.Sprintf("%.2f%s", 0.01, assetList.Assets[0].Base)
	}

	rpc, err := c.GetRandomRPCEndpoint(ctx)
	if err != nil {
		return nil, err
	}

	return &client.ChainClientConfig{
		Key:            "default",
		ChainID:        c.ChainID,
		RPCAddr:        rpc,
		AccountPrefix:  c.Bech32Prefix,
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
