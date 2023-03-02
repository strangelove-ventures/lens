package tx

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/strangelove-ventures/lens/client"
	"github.com/strangelove-ventures/lens/client/query"
)

// Ensure all Basic Allowance grants are in place for the given ChainClient.
// This will query (RPC) for existing grants and create new grants if they don't exist.
func EnsureBasicGrants(ctx context.Context, memo string, cc *client.ChainClient) (*sdk.TxResponse, error) {
	if cc.Config.FeeGrants == nil {
		return nil, errors.New("ChainClient must be a FeeGranter to establish grants")
	} else if len(cc.Config.FeeGrants.ManagedGrantees) == 0 {
		return nil, errors.New("ChainClient is a FeeGranter, but is not managing any Grantees")
	}

	granterKey := cc.Config.FeeGrants.GranterKey
	if granterKey == "" {
		granterKey = cc.Config.Key
	}

	granterAcc, err := cc.GetKeyAddressForKey(granterKey)
	if err != nil {
		fmt.Printf("Retrieving key '%s': ChainClient FeeGranter misconfiguration: %s", granterKey, err.Error())
		return nil, err
	}

	granterAddr, granterAddrErr := cc.EncodeBech32AccAddr(granterAcc)
	if granterAddrErr != nil {
		return nil, granterAddrErr
	}

	validGrants, err := query.GetValidBasicGrants(cc)
	failedLookupGrantsByGranter := err != nil

	msgs := []sdk.Msg{}
	numGrantees := len(cc.Config.FeeGrants.ManagedGrantees)
	grantsNeeded := 0

	for _, grantee := range cc.Config.FeeGrants.ManagedGrantees {

		//Searching for all grants with the given granter failed, so we will search by the grantee.
		//Reason this lookup sometimes fails is because the 'Search by granter' request is in SDK v0.46+
		if failedLookupGrantsByGranter {
			validGrants, err = query.GetGranteeValidBasicGrants(cc, grantee)
			if err != nil {
				return nil, err
			}
		}

		granteeAcc, err := cc.GetKeyAddressForKey(grantee)
		if err != nil {
			fmt.Printf("Misconfiguration for grantee key %s. Error: %s\n", grantee, err.Error())
			return nil, err
		}

		granteeAddr, granteeAddrErr := cc.EncodeBech32AccAddr(granteeAcc)
		if granteeAddrErr != nil {
			return nil, granteeAddrErr
		}

		hasGrant := false
		for _, basicGrant := range validGrants {
			if basicGrant.Grantee == granteeAddr {
				fmt.Printf("Valid grant found for granter %s, grantee %s\n", basicGrant.Granter, basicGrant.Grantee)
				hasGrant = true
			}
		}

		if !hasGrant {
			grantsNeeded += 1
			fmt.Printf("Grant will be created on chain for granter %s and grantee %s\n", granterAddr, granteeAddr)
			grantMsg, err := getMsgGrantBasicAllowance(cc, granterAcc, granteeAcc)
			if err != nil {
				return nil, err
			}
			msgs = append(msgs, grantMsg)
		}
	}

	if len(msgs) > 0 {
		//Make sure the granter has funds on chain, if not, we can't even pay TX fees.
		//Also, depending how the config was initialized, the key might only exist locally, not on chain.
		options := query.QueryOptions{}
		query := query.Query{Client: cc, Options: &options}
		balance, err := query.Bank_Balances(granterAddr)
		if err != nil {
			return nil, err
		}

		//Check to ensure the feegranter has funds on chain that can pay TX fees
		weBroke := true
		gasDenom, err := getGasTokenDenom(cc.Config.GasPrices)
		if err != nil {
			return nil, err
		}

		for _, coin := range balance.Balances {
			if coin.Denom == gasDenom {
				if coin.Amount.GT(sdk.ZeroInt()) {
					weBroke = false
				}
			}
		}

		//Feegranter can pay TX fees
		if !weBroke {
			txResp, err := cc.SubmitTxAwaitResponse(ctx, msgs, memo, 0, granterKey)
			if err != nil {
				fmt.Printf("Error: SubmitTxAwaitResponse: %s", err.Error())
				return nil, err
			} else if txResp != nil && txResp.TxResponse != nil && txResp.TxResponse.Code != 0 {
				fmt.Printf("Submitting grants for granter %s failed. Code: %d, TX hash: %s\n", granterKey, txResp.TxResponse.Code, txResp.TxResponse.TxHash)
				return nil, fmt.Errorf("could not configure feegrant for granter %s", granterKey)
			}

			fmt.Printf("TX succeeded, %d new grants configured, %d grants already in place. TX hash: %s\n", grantsNeeded, numGrantees-grantsNeeded, txResp.TxResponse.TxHash)
			return txResp.TxResponse, err
		} else {
			return nil, fmt.Errorf("granter %s does not have funds on chain in fee denom '%s' (no TXs submitted)", granterKey, gasDenom)
		}
	} else {
		fmt.Printf("All grantees (%d total) already had valid feegrants. Feegrant configuration verified.\n", numGrantees)
	}

	return nil, nil
}

func getGasTokenDenom(gasPrices string) (string, error) {
	r := regexp.MustCompile(`(?P<digits>[0-9.]*)(?P<denom>.*)`)
	submatches := r.FindStringSubmatch(gasPrices)
	if len(submatches) != 3 {
		return "", errors.New("could not find fee denom")
	}

	return submatches[2], nil
}

// GrantBasicAllowance Send a feegrant with the basic allowance type.
// This function does not check for existing feegrant authorizations.
// TODO: check for existing authorizations prior to attempting new one.
func GrantAllGranteesBasicAllowance(cc *client.ChainClient, ctx context.Context, gas uint64) error {
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

		grantResp, err := GrantBasicAllowance(cc, ctx, granterAddr, granterKey, granteeAddr, gas)
		if err != nil {
			return err
		} else if grantResp != nil && grantResp.TxResponse != nil && grantResp.TxResponse.Code != 0 {
			fmt.Printf("grantee %s and granter %s. Code: %d\n", granterAddr.String(), granteeAddr.String(), grantResp.TxResponse.Code)
			return fmt.Errorf("could not configure feegrant for granter %s and grantee %s", granterAddr.String(), granteeAddr.String())
		}
	}
	return nil
}

// GrantBasicAllowance Send a feegrant with the basic allowance type.
// This function does not check for existing feegrant authorizations.
// TODO: check for existing authorizations prior to attempting new one.
func GrantAllGranteesBasicAllowanceWithExpiration(cc *client.ChainClient, ctx context.Context, gas uint64, expiration time.Time) error {
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

		grantResp, err := GrantBasicAllowanceWithExpiration(cc, ctx, granterAddr, granterKey, granteeAddr, gas, expiration)
		if err != nil {
			return err
		} else if grantResp != nil && grantResp.TxResponse != nil && grantResp.TxResponse.Code != 0 {
			fmt.Printf("grantee %s and granter %s. Code: %d\n", granterAddr.String(), granteeAddr.String(), grantResp.TxResponse.Code)
			return fmt.Errorf("could not configure feegrant for granter %s and grantee %s", granterAddr.String(), granteeAddr.String())
		}
	}
	return nil
}

func getMsgGrantBasicAllowanceWithExpiration(cc *client.ChainClient, granter sdk.AccAddress, grantee sdk.AccAddress, expiration time.Time) (sdk.Msg, error) {
	//thirtyMin := time.Now().Add(30 * time.Minute)
	feeGrantBasic := &feegrant.BasicAllowance{
		Expiration: &expiration,
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

func GrantBasicAllowance(cc *client.ChainClient, ctx context.Context, granter sdk.AccAddress, granterKeyName string, grantee sdk.AccAddress, gas uint64) (*txtypes.GetTxResponse, error) {
	msgGrantAllowance, err := getMsgGrantBasicAllowance(cc, granter, grantee)
	if err != nil {
		return nil, err
	}

	msgs := []sdk.Msg{msgGrantAllowance}
	txResp, err := cc.SubmitTxAwaitResponse(ctx, msgs, "", gas, granterKeyName)
	if err != nil {
		fmt.Printf("Error: GrantBasicAllowance.SubmitTxAwaitResponse: %s", err.Error())
		return nil, err
	}

	return txResp, nil
}

func GrantBasicAllowanceWithExpiration(cc *client.ChainClient, ctx context.Context, granter sdk.AccAddress, granterKeyName string, grantee sdk.AccAddress, gas uint64, expiration time.Time) (*txtypes.GetTxResponse, error) {
	msgGrantAllowance, err := getMsgGrantBasicAllowanceWithExpiration(cc, granter, grantee, expiration)
	if err != nil {
		return nil, err
	}

	msgs := []sdk.Msg{msgGrantAllowance}
	txResp, err := cc.SubmitTxAwaitResponse(ctx, msgs, "", gas, granterKeyName)
	if err != nil {
		fmt.Printf("Error: GrantBasicAllowance.SubmitTxAwaitResponse: %s", err.Error())
		return nil, err
	}

	return txResp, nil
}
