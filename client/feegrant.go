package client

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// By default, TXs will be signed by the feegrantees 'ManagedGrantees' keys in a round robin fashion.
// ChainClients can use other signing keys by invoking 'tx.SendMsgsWith' and specifying the signing key.
type FeeGrantConfiguration struct {
	GranteesWanted int `json:"num_grantees" yaml:"num_grantees"`
	//Normally this is the default ChainClient key
	GranterKey string `json:"granter" yaml:"granter"`
	//List of keys (by name) that this FeeGranter manages
	ManagedGrantees []string `json:"grantees" yaml:"grantees"`
	//Last checked on chain (0 means grants never checked and may not exist)
	BlockHeightVerified int64 `json:"block_last_verified" yaml:"block_last_verified"`
	//Index of the last ManagedGrantee used as a TX signer
	GranteeLastSignerIndex int
}

func (cc *ChainClient) ConfigureFeegrants(numGrantees int, granterKey string) error {
	cc.Config.FeeGrants = &FeeGrantConfiguration{
		GranteesWanted:  numGrantees,
		GranterKey:      granterKey,
		ManagedGrantees: []string{},
	}

	return cc.Config.FeeGrants.AddGranteeKeys(cc)
}

func (fg *FeeGrantConfiguration) AddGranteeKeys(cc *ChainClient) error {
	for i := len(fg.ManagedGrantees); i < fg.GranteesWanted; i++ {
		newGranteeIdx := strconv.Itoa(len(fg.ManagedGrantees) + 1)
		newGrantee := "grantee" + newGranteeIdx

		//Add another key to the chain client for the grantee
		_, err := cc.AddKey(newGrantee, sdk.CoinType)
		if err != nil {
			return err
		}

		fg.ManagedGrantees = append(fg.ManagedGrantees, newGrantee)
	}

	return nil
}

// Get the feegrant params to use for the next TX. If feegrants are not configured for the chain client, the default key will be used for TX signing.
// Otherwise, a configured feegrantee will be chosen for TX signing in round-robin fashion.
func (cc *ChainClient) GetTxFeeGrant() (txSignerKey string, feeGranterKey string, err error) {
	//By default, we should sign TXs with the ChainClient's default key
	txSignerKey = cc.Config.Key

	if cc.Config.FeeGrants == nil {
		return
	}

	// Use the ChainClient's configured Feegranter key for the next TX.
	feeGranterKey = cc.Config.FeeGrants.GranterKey

	// The ChainClient Feegrant configuration has never been verified on chain.
	// Don't use Feegrants as it could cause the TX to fail on chain.
	if feeGranterKey == "" || cc.Config.FeeGrants.BlockHeightVerified <= 0 {
		feeGranterKey = ""
		return
	}

	//Pick the next managed grantee in the list as the TX signer
	lastGranteeIdx := cc.Config.FeeGrants.GranteeLastSignerIndex
	if lastGranteeIdx >= 0 && lastGranteeIdx <= len(cc.Config.FeeGrants.ManagedGrantees)-1 {
		txSignerKey = cc.Config.FeeGrants.ManagedGrantees[lastGranteeIdx]
		cc.Config.FeeGrants.GranteeLastSignerIndex = cc.Config.FeeGrants.GranteeLastSignerIndex + 1

		//Restart the round robin at 0 if we reached the end of the list of grantees
		if cc.Config.FeeGrants.GranteeLastSignerIndex == len(cc.Config.FeeGrants.ManagedGrantees) {
			cc.Config.FeeGrants.GranteeLastSignerIndex = 0
		}
	}

	return
}
