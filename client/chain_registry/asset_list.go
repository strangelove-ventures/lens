package chain_registry

type AssetList struct {
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
