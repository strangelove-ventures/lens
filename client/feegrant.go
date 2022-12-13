package client

import (
	"errors"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"
)

// By default, TXs will be signed by the feegrantees 'ManagedGrantees' keys in a round robin fashion.
// ChainClients can use other signing keys by invoking 'tx.SendMsgsWith' and specifying the signing key.
type FeeGrantConfiguration struct {
	GranteesWanted int
	//Normally this is the default ChainClient key
	GranterKey string
	//List of keys (by name) that this FeeGranter manages
	ManagedGrantees []string
	//Last checked on chain (0 means grants never checked and may not exist)
	BlockHeightVerified int64
}

func (cc *ChainClient) ConfigureFeegrants() error {
	cc.Config.FeeGrants = &FeeGrantConfiguration{
		GranteesWanted:  10,
		GranterKey:      cc.Config.Key,
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

func (cc *ChainClient) GetFeeGranterAddress(txKey string) (sdk.AccAddress, error) {
	if cc.Config.FeeGrants == nil {
		return sdk.AccAddress{}, errors.New("no feegranter configured")
	}

	granterKey := cc.Config.FeeGrants.GranterKey
	if granterKey == "" {
		granterKey = cc.Config.Key
	}

	if granterKey == txKey {
		return sdk.AccAddress{}, errors.New("cannot feegrant your own TX")
	}

	granterAddr, err := cc.GetKeyAddressForKey(granterKey)
	if err != nil {
		cc.log.Error("ChainClient FeeGrantee.GranterAddress misconfiguration",
			zap.String("Granter key", granterKey),
			zap.Error(err),
		)
		return granterAddr, err
	}

	return granterAddr, err
}

func (cc *ChainClient) EnsureBasicFeeGrants() {

}
