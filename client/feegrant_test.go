package client_test

import (
	"context"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/strangelove-ventures/lens/client"
	"go.uber.org/zap/zaptest"
)

func newChainClientLocalJuno(t *testing.T, keyName string, mnemonic string) *client.ChainClient {
	coinType := uint32(118) //Cosmos coin type

	homepath := t.TempDir()
	cl, err := client.NewChainClient(
		zaptest.NewLogger(t),
		client.GetJunoLocalConfig(homepath, true),
		homepath, nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = cl.RestoreKey(keyName, mnemonic, coinType)
	if err != nil {
		t.Fatal(err)
	}

	//Set new key to the default
	cl.Config.Key = keyName
	return cl
}

// To run this test you must first launch a local Juno chain. To do so, git clone the juno repo and run docker-compose up
func TestFeeGrantBasic(t *testing.T) {
	coinType := uint32(118) //Cosmos coin type
	granterKey := "ccGranterKey"
	granteeKey := "ccGranteeKey"
	ctx := context.Background()

	//Mnemonic is from the Juno local testnet scripts (clone juno repo, cat docker/test-user.env)
	junoLocalKeyMnemonic := "clip hire initial neck maid actor venue client foam budget lock catalog sweet steak waste crater broccoli pipe steak sister coyote moment obvious choose"
	chainClient := newChainClientLocalJuno(t, granterKey, junoLocalKeyMnemonic)

	//Add another key to the chain client for the grantee
	_, err := chainClient.AddKey(granteeKey, coinType)
	if err != nil {
		t.Fatal(err)
	}

	//Need to make sure the grantee's account exists on chain before attempting a FeeGrant.
	//Therefore we just send some funds to that address.
	fundHash := fundAccount(t, ctx, chainClient, granteeKey, granterKey)
	fmt.Printf("Funded grantee account: tx hash %s\n", fundHash)

	chainClient.FeeGrants = &client.FeeGrantConfiguration{
		GranteesWanted:  1,
		GranterKey:      granterKey,
		ManagedGrantees: []string{granteeKey},
	}

	err = chainClient.GrantAllGranteesBasicAllowance(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func fundAccount(t *testing.T, ctx context.Context, cc *client.ChainClient, keyNameReceiveFunds string, keyNameSendFunds string) string {
	fromAddr, err := cc.GetKeyAddressForKey(keyNameSendFunds)
	if err != nil {
		t.Fatal(err)
	}

	toAddr, err := cc.GetKeyAddressForKey(keyNameReceiveFunds)
	if err != nil {
		t.Fatal(err)
	}

	coins, err := sdk.ParseCoinsNormalized("1000ustake")
	if err != nil {
		t.Fatal(err)
	}

	req := &banktypes.MsgSend{
		FromAddress: cc.MustEncodeAccAddr(fromAddr),
		ToAddress:   cc.MustEncodeAccAddr(toAddr),
		Amount:      coins,
	}

	//TODO: keyNameSendFunds needs to be used for actually submitting the TX, in case its different than the default key
	res, err := cc.SubmitTxAwaitResponse(ctx, []sdk.Msg{req}, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	return res.TxResponse.TxHash
}
