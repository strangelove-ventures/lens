package chain_registry

import (
	"context"

	"go.uber.org/zap"
)

type ChainRegistry interface {
	GetChain(ctx context.Context, name string) (ChainInfo, error)
	ListChains(ctx context.Context) ([]string, error)
	SourceLink() string
}

func DefaultChainRegistry(log *zap.Logger) ChainRegistry {
	return NewCosmosGithubRegistry(log.With(zap.String("registry", "cosmos_github")))
}
