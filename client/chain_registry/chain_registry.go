package chain_registry

import (
	"context"

	"github.com/strangelove-ventures/lens/client"
	"go.uber.org/zap"
)

type ChainRegistry interface {
	AddChainInfo(ctx context.Context, name string) (*client.ChainClientConfig, error)
	ListChains(ctx context.Context) ([]string, error)
	SourceLink() string
}

func DefaultChainRegistry(log *zap.Logger) ChainRegistry {
	return NewCosmosGithubRegistry(log.With(zap.String("registry", "cosmos_github")))
}

func EcoStakeRegistryAPI(log *zap.Logger) ChainRegistry {
	return NewChainRegistryAPI(log.With(zap.String("registry", "eco-stake/cosmos-directory api")))
}
