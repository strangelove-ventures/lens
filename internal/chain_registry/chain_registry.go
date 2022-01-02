package chain_registry

import (
	"github.com/strangelove-ventures/lens/internal/chain_info"
)

type ChainRegistry interface {
	GetChain(name string) (chain_info.ChainInfo, error)
	ListChains() ([]string, error)
	SourceLink() string
}

func DefaultChainRegistry() ChainRegistry {
	return CosmosGithubRegistry{}
}
