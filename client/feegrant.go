package client

import (
	"context"
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"go.uber.org/zap"
)

type FeeGrantConfiguration struct {
	GranteesWanted int
	//Normally this is the default ChainClient key
	GranterKey string
	//List of keys (by name) that this FeeGranter manages
	ManagedGrantees []string
}

func (cc *ChainClient) GetFeeGranterAddress(txKey string) (sdk.AccAddress, error) {
	if cc.FeeGrants == nil {
		return sdk.AccAddress{}, errors.New("no feegranter configured")
	}

	granterKey := cc.FeeGrants.GranterKey
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

// GrantBasicAllowance Send a feegrant with the basic allowance type.
// This function does not check for existing feegrant authorizations.
// TODO: check for existing authorizations prior to attempting new one.
func (cc *ChainClient) GrantAllGranteesBasicAllowance(ctx context.Context) error {
	if cc.FeeGrants == nil {
		return errors.New("ChainClient must be a FeeGranter to establish grants")
	} else if len(cc.FeeGrants.ManagedGrantees) == 0 {
		return errors.New("ChainClient is a FeeGranter, but is not managing any Grantees")
	}

	granterKey := cc.FeeGrants.GranterKey
	if granterKey == "" {
		granterKey = cc.Config.Key
	}
	granterAddr, err := cc.GetKeyAddressForKey(granterKey)
	if err != nil {
		cc.log.Error("ChainClient FeeGranter misconfiguration", zap.Error(err))
		return err
	}

	for _, grantee := range cc.FeeGrants.ManagedGrantees {
		granteeAddr, err := cc.GetKeyAddressForKey(grantee)

		if err != nil {
			cc.log.Error("ChainClient FeeGrantee.GranterAddress misconfiguration",
				zap.String("GranteeAddress", grantee),
				zap.Error(err),
			)
			return err
		}

		grantResp, err := cc.GrantBasicAllowance(ctx, granterAddr, granteeAddr)
		if err != nil {
			return err
		} else {
			cc.log.Debug("FeeGrant",
				zap.String("type", "BasicAllowance"),
				zap.String("granter address", granterAddr.String()),
				zap.String("grantee address", granteeAddr.String()),
				zap.String("hash", grantResp.TxResponse.TxHash),
				zap.Uint32("result", grantResp.TxResponse.Code),
			)
		}
	}
	return nil
}

// TODO: the SubmitTxAwaitResponse should use the granter key, not the default key used for submitting TXs. These could potentially be different.
func (cc *ChainClient) GrantBasicAllowance(ctx context.Context, granter sdk.AccAddress, grantee sdk.AccAddress) (*txtypes.GetTxResponse, error) {
	thirtyMin := time.Now().Add(30 * time.Minute)
	feeGrantBasic := &feegrant.BasicAllowance{
		Expiration: &thirtyMin,
	}
	msgGrantAllowance, err := feegrant.NewMsgGrantAllowance(feeGrantBasic, granter, grantee)

	if err != nil {
		cc.log.Error("GrantBasicAllowance.NewMsgGrantAllowance", zap.Error(err))
		return nil, err
	}

	msgs := []sdk.Msg{msgGrantAllowance}
	txResp, err := cc.SubmitTxAwaitResponse(ctx, msgs, "", 80000)
	if err != nil {
		cc.log.Error("GrantBasicAllowance.SubmitTxAwaitResponse", zap.Error(err))
		return nil, err
	}

	return txResp, nil
}
