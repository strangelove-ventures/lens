package chain_registry

type ChainRegistry interface {
	GetChain(name string) (ChainInfo, error)
	ListChains() ([]string, error)
	SourceLink() string
}

func DefaultChainRegistry() ChainRegistry {
	return CosmosGithubRegistry{}
}
