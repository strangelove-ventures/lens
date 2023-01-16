package client_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
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

func setSdkContext(prefix string) {
	sdkConf := sdk.GetConfig()
	sdkConf.SetBech32PrefixForAccount(prefix, prefix+"pub")
	sdkConf.SetBech32PrefixForValidator(prefix+"valoper", prefix+"valoperpub")
	sdkConf.SetBech32PrefixForConsensusNode(prefix+"valcons", prefix+"valconspub")
}

// To run this test you must first launch a local Juno chain. To do so, git clone the juno repo and run docker-compose up
// ensure feegranting fails if granter has no funds
func TestFeegranterEmptyBalance(t *testing.T) {
	setSdkContext("juno")

	//coinType := uint32(118) //Cosmos coin type
	defaultKey := "default"
	granterKey := "underfundedGranter"
	ctx := context.Background()

	//Mnemonic is from the Juno local testnet scripts (clone juno repo, cat docker/test-user.env)
	junoLocalKeyMnemonic := "clip hire initial neck maid actor venue client foam budget lock catalog sweet steak waste crater broccoli pipe steak sister coyote moment obvious choose"
	cc := newChainClientLocalJuno(t, defaultKey, junoLocalKeyMnemonic)

	//Add another key to the chain client for the grantee
	_, err := cc.AddKey(granterKey, sdk.CoinType)
	if err != nil {
		t.Fail()
	}

	granterAcc, err := cc.GetKeyAddressForKey(granterKey)
	if err != nil {
		t.Fatal(err)
	}
	granterAddr := cc.MustEncodeAccAddr(granterAcc)

	//granter will have 5000ujunox total after this
	fundAccount(t, ctx, cc, granterKey, defaultKey, "5000ujunox", 0)
	q := query.Query{Client: cc, Options: &query.QueryOptions{}}

	//Set the configuration locally for the ChainClient
	cc.ConfigureFeegrants(1, granterKey)

	//gas prices are .01 so this equates to 1000ujunox in gas fees, leaving granter with 4000ujunox
	err = tx.GrantAllGranteesBasicAllowance(cc, ctx, 100000)
	if err != nil {
		t.Fatal(err)
	}

	//Get latest height from the chain, mark feegrant configuration as verified up to that height.
	h, err := cc.QueryLatestHeight(ctx)
	if err != nil {
		t.Fail()
	}
	cc.Config.FeeGrants.BlockHeightVerified = h

	//Grantee will have 2000ujunox, granter will have 1000ujunox
	fundHash := fundAccount(t, ctx, cc, cc.Config.FeeGrants.ManagedGrantees[0], granterKey, "2000ujunox", 100000)
	fmt.Printf("Funded grantee account with 2000ujunox: tx hash %s\n", fundHash)

	//At this point grantee has EXACTLY	2000ujunox.
	//Now send it back to the granter, forcing the granter to pay the TX fee.
	//Since granter is out of funds, feegrant should fail
	granteeKey := cc.Config.FeeGrants.ManagedGrantees[0]
	granteeAcc, err := cc.GetKeyAddressForKey(granteeKey)
	if err != nil {
		t.Fatal(err)
	}
	granteeAddr := cc.MustEncodeAccAddr(granteeAcc)
	balanceBeforeFundReturn, err := q.Bank_Balance(granteeAddr, "ujunox")
	if err != nil {
		t.Fatal(err)
	}
	//Granter sent grantee 2000ujunox earlier.
	assert.Equal(t, balanceBeforeFundReturn.Balance.Amount, sdk.NewInt(2000))

	granterAccBalance, err := q.Bank_Balance(granterAddr, "ujunox")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Granter balance: %+v\n", granterAccBalance.Balance)

	//Send some funds to the default address. Granter now has 1000ujunox to pay this feegranted TX, but we use just over that in fees, so this should fail.
	_, err = feegrantSendFunds(t, ctx, cc, defaultKey, granteeKey, 100001)
	assert.NotNil(t, err)
	fmt.Printf("Error feegranting send funds TX: %s\n", err)
}

// To run this test you must first launch a local Juno chain. To do so, git clone the juno repo and run docker-compose up
// test what happens if a feegrantee's grant expired
func TestFeegranterExpiredGrants(t *testing.T) {
	setSdkContext("juno")

	//coinType := uint32(118) //Cosmos coin type
	defaultKey := "default"
	granterKey := "granter1"
	ctx := context.Background()

	//Mnemonic is from the Juno local testnet scripts (clone juno repo, cat docker/test-user.env)
	junoLocalKeyMnemonic := "clip hire initial neck maid actor venue client foam budget lock catalog sweet steak waste crater broccoli pipe steak sister coyote moment obvious choose"
	cc := newChainClientLocalJuno(t, defaultKey, junoLocalKeyMnemonic)

	//Add another key to the chain client for the grantee
	_, err := cc.AddKey(granterKey, sdk.CoinType)
	if err != nil {
		t.Fail()
	}

	//fund granter
	fundAccount(t, ctx, cc, granterKey, defaultKey, "10000ujunox", 0)
	q := query.Query{Client: cc, Options: &query.QueryOptions{}}

	//Set the configuration locally for the ChainClient
	cc.ConfigureFeegrants(1, granterKey)

	expiresAt := time.Now().Add(30 * time.Second)
	//gas prices are .01 so this equates to 1000ujunox in gas fees, leaving granter with 4000ujunox
	err = tx.GrantAllGranteesBasicAllowanceWithExpiration(cc, ctx, 100000, expiresAt)
	if err != nil {
		t.Fatal(err)
	}

	//Get latest height from the chain, mark feegrant configuration as verified up to that height.
	h, err := cc.QueryLatestHeight(ctx)
	if err != nil {
		t.Fail()
	}
	cc.Config.FeeGrants.BlockHeightVerified = h

	//Grantee will have 2000ujunox, granter will have 1000ujunox
	fundHash := fundAccount(t, ctx, cc, cc.Config.FeeGrants.ManagedGrantees[0], granterKey, "2000ujunox", 100000)
	fmt.Printf("Funded grantee account with 2000ujunox: tx hash %s\n", fundHash)

	//At this point grantee has EXACTLY	2000ujunox.
	//Now send it back to the granter, forcing the granter to pay the TX fee.
	granteeKey := cc.Config.FeeGrants.ManagedGrantees[0]
	granteeAcc, err := cc.GetKeyAddressForKey(granteeKey)
	if err != nil {
		t.Fatal(err)
	}
	granteeAddr := cc.MustEncodeAccAddr(granteeAcc)
	balanceBeforeFundReturn, err := q.Bank_Balance(granteeAddr, "ujunox")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, balanceBeforeFundReturn.Balance.Amount, sdk.NewInt(2000))

	//Send some funds to the default address. Should succeed
	_, err = feegrantSendFunds(t, ctx, cc, defaultKey, granteeKey, 100000)
	assert.Nil(t, err)

	//Wait for the feegrant to expire
	time.Sleep(30 * time.Second)

	//Send some funds to the default address again. Should FAIL
	_, err = feegrantSendFunds(t, ctx, cc, defaultKey, granteeKey, 100000)
	assert.NotNil(t, err)
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
	fundHash := fundAccount(t, ctx, chainClient, granteeKey, granterKey, "1000ujunox", 0)
	fmt.Printf("Funded grantee account: tx hash %s\n", fundHash)

	chainClient.Config.FeeGrants = &client.FeeGrantConfiguration{
		GranteesWanted:  1,
		GranterKey:      granterKey,
		ManagedGrantees: []string{granteeKey},
	}

	err = tx.GrantAllGranteesBasicAllowance(chainClient, ctx, 80000)
	if err != nil {
		t.Fatal(err)
	}
}

// To run this test you must first launch a local Juno chain. To do so, git clone the juno repo and run docker-compose up
func TestFeeGrantRoundRobin(t *testing.T) {
	setSdkContext("juno")

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
		fundHash := fundAccount(t, ctx, cc, granteeKey, granterKey, "1000ujunox", 0)
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
		fundHash, err := feegrantSendFunds(t, ctx, cc, granterKey, granteeKey, 0) //autocalculate gas
		if err != nil {
			t.Fatal()
		}
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

func feegrantSendFunds(t *testing.T, ctx context.Context, cc *client.ChainClient, keyNameReceiveFunds string, keyNameSendFunds string, gas uint64) (*txtypes.GetTxResponse, error) {
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

	signer, feegranter, err := cc.GetTxFeeGrant()
	if err != nil {
		return nil, err
	}
	res, err := cc.SendMsgsWith(ctx, []sdk.Msg{req}, "", gas, signer, feegranter)
	if err != nil {
		return nil, err
	}

	tx1resp, err := cc.AwaitTx(res.TxHash, 15*time.Second)
	if err != nil {
		return nil, err
	}

	return tx1resp, err
}

func fundAccount(t *testing.T, ctx context.Context, cc *client.ChainClient, keyNameReceiveFunds string, keyNameSendFunds string, amountCoin string, gas uint64) string {
	fromAddr, err := cc.GetKeyAddressForKey(keyNameSendFunds)
	if err != nil {
		t.Fatal(err)
	}

	toAddr, err := cc.GetKeyAddressForKey(keyNameReceiveFunds)
	if err != nil {
		t.Fatal(err)
	}

	coins, err := sdk.ParseCoinsNormalized(amountCoin)
	if err != nil {
		t.Fatal(err)
	}

	req := &banktypes.MsgSend{
		FromAddress: cc.MustEncodeAccAddr(fromAddr),
		ToAddress:   cc.MustEncodeAccAddr(toAddr),
		Amount:      coins,
	}

	res, err := cc.SubmitTxAwaitResponse(ctx, []sdk.Msg{req}, "", gas, keyNameSendFunds)
	if err != nil {
		t.Fatal(err)
	}
	return res.TxResponse.TxHash
}
