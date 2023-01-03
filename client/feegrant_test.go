package client_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/strangelove-ventures/lens/client"
	"github.com/strangelove-ventures/lens/client/query"
	tx "github.com/strangelove-ventures/lens/client/tx"
	"github.com/stretchr/testify/assert"
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

func TestGasPricesRegex(t *testing.T) {
	gasPrices := "0.01ujunox"
	denom, err := getGasTokenDenom(gasPrices)
	if err != nil || denom != "ujunox" {
		t.Fail()
	}
}

func getGasTokenDenom(gasPrices string) (string, error) {
	r := regexp.MustCompile(`(?P<digits>[0-9.]*)(?P<denom>.*)`)
	submatches := r.FindStringSubmatch(gasPrices)
	if len(submatches) != 3 {
		return "", errors.New("could not find fee denom")
	}
	//fmt.Printf("%#v\n", r.SubexpNames())

	return submatches[2], nil
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

	//Ensure grantee's account exists on chain before attempting a FeeGrant.
	//Therefore we just send some funds to that address.
	fundHash := fundAccount(t, ctx, chainClient, granteeKey, granterKey)
	fmt.Printf("Funded grantee account: tx hash %s\n", fundHash)

	chainClient.Config.FeeGrants = &client.FeeGrantConfiguration{
		GranteesWanted:  1,
		GranterKey:      granterKey,
		ManagedGrantees: []string{granteeKey},
	}

	err = tx.GrantAllGranteesBasicAllowance(chainClient, ctx)
	if err != nil {
		t.Fatal(err)
	}
}

// To run this test you must first launch a local Juno chain. To do so, git clone the juno repo and run docker-compose up
func TestFeeGrantRoundRobin(t *testing.T) {
	// This is, unfortunately, very necessary due to the SDK's internal caching of bech32 address translation.
	// In other words, lets say you DON'T set the bech32 prefix. The default is 'cosmos'. Then you call any function
	// in the cosmos SDK that returns the account address string. From that point on, it will ALWAYS cache 'cosmos'
	// as the prefix and you'll never be able to produce 'juno' as the prefix for that bech32 pub key, no matter what.
	// This is a huge problem for feegrants. Actually, it MIGHT be necessary to use the reflect package to set some fields,
	// depending on how the relayer works.
	prefix := "juno"
	sdkConf := sdk.GetConfig()
	sdkConf.SetBech32PrefixForAccount(prefix, prefix+"pub")
	sdkConf.SetBech32PrefixForValidator(prefix+"valoper", prefix+"valoperpub")
	sdkConf.SetBech32PrefixForConsensusNode(prefix+"valcons", prefix+"valconspub")

	//coinType := uint32(118) //Cosmos coin type
	granterKey := "ccGranterKey"
	ctx := context.Background()

	//Mnemonic is from the Juno local testnet scripts (clone juno repo, cat docker/test-user.env)
	junoLocalKeyMnemonic := "clip hire initial neck maid actor venue client foam budget lock catalog sweet steak waste crater broccoli pipe steak sister coyote moment obvious choose"
	cc := newChainClientLocalJuno(t, granterKey, junoLocalKeyMnemonic)

	//Set the configuration locally for the ChainClient
	cc.ConfigureFeegrants(10, cc.Config.Key)

	_, err := tx.EnsureBasicGrants(context.Background(), "feegrant configuration", cc)
	if err != nil {
		t.Fail()
	}

	//Get latest height from the chain, mark feegrant configuration as verified up to that height.
	h, err := cc.QueryLatestHeight(ctx)
	if err != nil {
		t.Fail()
	}
	cc.Config.FeeGrants.BlockHeightVerified = h

	//Send every grantee 1000ujunox
	for _, granteeKey := range cc.Config.FeeGrants.ManagedGrantees {
		fundHash := fundAccount(t, ctx, cc, granteeKey, granterKey)
		fmt.Printf("Funded grantee account with 1000ujunox: tx hash %s\n", fundHash)
	}

	//At this point every grantee's account has EXACTLY	1000ujunox.
	//Now send it back to the granter, forcing the granter to pay the TX fee.
	//Since the grantee only has exactly 1000ujunox, and we are sending it back to the granter,
	//every grantee will end up with 0 balance again after this.
	for _, granteeKey := range cc.Config.FeeGrants.ManagedGrantees {
		q := query.Query{Client: cc, Options: &query.QueryOptions{}}
		granteeAcc, err := cc.GetKeyAddressForKey(granteeKey)
		if err != nil {
			t.Fatal(err)
		}
		granteeAddr := cc.MustEncodeAccAddr(granteeAcc)
		balanceBeforeFundReturn, err := q.Bank_Balance(granteeAddr, "ujunox")
		if err != nil {
			t.Fatal(err)
		}
		//Granter sent grantee 1000ujunox earlier.
		assert.Equal(t, balanceBeforeFundReturn.Balance.Amount, sdk.NewInt(1000))

		//Send it back
		fundHash := feegrantSendFunds(t, ctx, cc, granterKey, granteeKey)
		fmt.Printf("Returned granter's funds: tx hash %s\n", fundHash)

		balanceAfterFundReturn, err := q.Bank_Balance(granteeAddr, "ujunox")
		if err != nil {
			t.Fatal(err)
		}

		//Grantee returned the 1000ujunox to the grantee
		expectedBalance := false
		if balanceAfterFundReturn.Balance == nil {
			expectedBalance = true
		} else if balanceAfterFundReturn.Balance.Amount.Equal(sdk.ZeroInt()) {
			expectedBalance = true
		}

		assert.True(t, expectedBalance)
	}
}

func feegrantSendFunds(t *testing.T, ctx context.Context, cc *client.ChainClient, keyNameReceiveFunds string, keyNameSendFunds string) string {
	fromAddr, err := cc.GetKeyAddressForKey(keyNameSendFunds)
	if err != nil {
		t.Fatal(err)
	}

	toAddr, err := cc.GetKeyAddressForKey(keyNameReceiveFunds)
	if err != nil {
		t.Fatal(err)
	}

	coins, err := sdk.ParseCoinsNormalized("1000ujunox")
	if err != nil {
		t.Fatal(err)
	}

	req := &banktypes.MsgSend{
		FromAddress: cc.MustEncodeAccAddr(fromAddr),
		ToAddress:   cc.MustEncodeAccAddr(toAddr),
		Amount:      coins,
	}

	res, err := cc.AwaitFeegrantedTx(ctx, []sdk.Msg{req}, "")
	if err != nil {
		t.Fatal(err)
	}
	return res.TxResponse.TxHash
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

	coins, err := sdk.ParseCoinsNormalized("1000ujunox")
	if err != nil {
		t.Fatal(err)
	}

	req := &banktypes.MsgSend{
		FromAddress: cc.MustEncodeAccAddr(fromAddr),
		ToAddress:   cc.MustEncodeAccAddr(toAddr),
		Amount:      coins,
	}

	res, err := cc.SubmitTxAwaitResponse(ctx, []sdk.Msg{req}, "", 0, keyNameSendFunds)
	if err != nil {
		t.Fatal(err)
	}
	return res.TxResponse.TxHash
}
