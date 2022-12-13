package tx

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/strangelove-ventures/lens/client"
	"github.com/strangelove-ventures/lens/client/query"
)

// Ensure all Basic Allowance grants are in place for the given ChainClient.
// This will query (RPC) for existing grants and create new grants if they don't exist.
func EnsureBasicGrants(cc *client.ChainClient) error {
	if cc.Config.FeeGrants == nil {
		return errors.New("ChainClient must be a FeeGranter to establish grants")
	} else if len(cc.Config.FeeGrants.ManagedGrantees) == 0 {
		return errors.New("ChainClient is a FeeGranter, but is not managing any Grantees")
	}

	granterKey := cc.Config.FeeGrants.GranterKey
	if granterKey == "" {
		granterKey = cc.Config.Key
	}

	granterAcc, err := cc.GetKeyAddressForKey(granterKey)
	if err != nil {
		fmt.Printf("ChainClient FeeGranter misconfiguration: %s", err.Error())
		return err
	}

	validGrants, err := query.GetValidBasicGrants(cc)
	if err != nil {
		return err
	}

	msgs := []sdk.Msg{}

	for _, grantee := range cc.Config.FeeGrants.ManagedGrantees {
		granteeAcc, err := cc.GetKeyAddressForKey(grantee)
		if err != nil {
			fmt.Printf("Misconfiguration for grantee key %s. Error: %s\n", grantee, err.Error())
			return err
		}

		granteeAddr, granteeAddrErr := cc.EncodeBech32AccAddr(granteeAcc)
		if granteeAddrErr != nil {
			return granteeAddrErr
		}

		hasGrant := false
		for _, basicGrant := range validGrants {
			if basicGrant.Grantee == granteeAddr {
				hasGrant = true
			}
		}

		if !hasGrant {
			grantMsg, err := getMsgGrantBasicAllowance(cc, granterAcc, granteeAcc)
			if err != nil {
				return err
			}
			msgs = append(msgs, grantMsg)
		}
	}

	if len(msgs) > 0 {
		txResp, err := cc.SubmitTxAwaitResponse(context.Background(), msgs, "", 0, granterKey)
		if err != nil {
			fmt.Printf("Error: SubmitTxAwaitResponse: %s", err.Error())
			return err
		} else if txResp != nil && txResp.TxResponse != nil && txResp.TxResponse.Code != 0 {
			fmt.Printf("Submitting grants for granter %s failed. Code: %d, TX hash: %s\n", granterKey, txResp.TxResponse.Code, txResp.TxResponse.TxHash)
			return fmt.Errorf("could not configure feegrant for granter %s", granterKey)
		}
	}

	return nil
}

// GrantBasicAllowance Send a feegrant with the basic allowance type.
// This function does not check for existing feegrant authorizations.
// TODO: check for existing authorizations prior to attempting new one.
func GrantAllGranteesBasicAllowance(cc *client.ChainClient, ctx context.Context) error {
	if cc.Config.FeeGrants == nil {
		return errors.New("ChainClient must be a FeeGranter to establish grants")
	} else if len(cc.Config.FeeGrants.ManagedGrantees) == 0 {
		return errors.New("ChainClient is a FeeGranter, but is not managing any Grantees")
	}

	granterKey := cc.Config.FeeGrants.GranterKey
	if granterKey == "" {
		granterKey = cc.Config.Key
	}
	granterAddr, err := cc.GetKeyAddressForKey(granterKey)
	if err != nil {
		fmt.Printf("ChainClient FeeGranter misconfiguration: %s", err.Error())
		return err
	}

	for _, grantee := range cc.Config.FeeGrants.ManagedGrantees {
		granteeAddr, err := cc.GetKeyAddressForKey(grantee)

		if err != nil {
			fmt.Printf("Misconfiguration for grantee %s. Error: %s\n", grantee, err.Error())
			return err
		}

		grantResp, err := GrantBasicAllowance(cc, ctx, granterAddr, granterKey, granteeAddr)
		if err != nil {
			return err
		} else if grantResp != nil && grantResp.TxResponse != nil && grantResp.TxResponse.Code != 0 {
			fmt.Printf("grantee %s and granter %s. Code: %d\n", granterAddr.String(), granteeAddr.String(), grantResp.TxResponse.Code)
			return fmt.Errorf("could not configure feegrant for granter %s and grantee %s", granterAddr.String(), granteeAddr.String())
		}
	}
	return nil
}

func getMsgGrantBasicAllowance(cc *client.ChainClient, granter sdk.AccAddress, grantee sdk.AccAddress) (sdk.Msg, error) {
	//thirtyMin := time.Now().Add(30 * time.Minute)
	feeGrantBasic := &feegrant.BasicAllowance{
		//Expiration: &thirtyMin,
	}
	msgGrantAllowance, err := feegrant.NewMsgGrantAllowance(feeGrantBasic, granter, grantee)
	if err != nil {
		fmt.Printf("Error: GrantBasicAllowance.NewMsgGrantAllowance: %s", err.Error())
		return nil, err
	}

	//Due to the way Lens configures the SDK, addresses will have the 'cosmos' prefix which
	//doesn't necessarily match the chain prefix of the ChainClient config. So calling the internal
	//'NewMsgGrantAllowance' function will return the *incorrect* 'cosmos' prefixed bech32 address.

	//Update the Grant to ensure the correct chain-specific granter is set
	granterAddr, granterAddrErr := cc.EncodeBech32AccAddr(granter)
	if granterAddrErr != nil {
		fmt.Printf("EncodeBech32AccAddr: %s", granterAddrErr.Error())
		return nil, granterAddrErr
	}

	//Update the Grant to ensure the correct chain-specific grantee is set
	granteeAddr, granteeAddrErr := cc.EncodeBech32AccAddr(grantee)
	if granteeAddrErr != nil {
		fmt.Printf("EncodeBech32AccAddr: %s", granteeAddrErr.Error())
		return nil, granteeAddrErr
	}

	//override the 'cosmos' prefixed bech32 addresses with the correct chain prefix
	msgGrantAllowance.Grantee = granteeAddr
	msgGrantAllowance.Granter = granterAddr

	return msgGrantAllowance, nil
}

func GrantBasicAllowance(cc *client.ChainClient, ctx context.Context, granter sdk.AccAddress, granterKeyName string, grantee sdk.AccAddress) (*txtypes.GetTxResponse, error) {
	msgGrantAllowance, err := getMsgGrantBasicAllowance(cc, granter, grantee)
	if err != nil {
		return nil, err
	}

	msgs := []sdk.Msg{msgGrantAllowance}
	txResp, err := cc.SubmitTxAwaitResponse(ctx, msgs, "", 80000, granterKeyName)
	if err != nil {
		fmt.Printf("Error: GrantBasicAllowance.SubmitTxAwaitResponse: %s", err.Error())
		return nil, err
	}

	return txResp, nil
}
