package chain_registry

import "context"

type ChainRegistry interface {
	GetChain(ctx context.Context, name string) (ChainInfo, error)
	ListChains(ctx context.Context) ([]string, error)
	SourceLink() string
}

func DefaultChainRegistry() ChainRegistry {
	return CosmosGithubRegistry{}
}
