package chain_registry

import (
	"github.com/strangelove-ventures/lens/internal/asset_list"
	"github.com/strangelove-ventures/lens/internal/chain_info"
)

type ChainRegistry interface {
	GetChain(name string) (chain_info.ChainInfo, error)
	GetAssets(name string) (asset_list.AssetList, error)
	ListChains() ([]string, error)
	SourceLink() string
}

func DefaultChainRegistry() ChainRegistry {
	return CosmosGithubRegistry{}
}
