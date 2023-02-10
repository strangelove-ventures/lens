package client

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	abci "github.com/tendermint/tendermint/abci/types"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (cc *ChainClient) TxFactory() tx.Factory {
	return tx.Factory{}.
		WithAccountRetriever(cc).
		WithChainID(cc.Config.ChainID).
		WithTxConfig(cc.Codec.TxConfig).
		WithGasAdjustment(cc.Config.GasAdjustment).
		WithGasPrices(cc.Config.GasPrices).
		WithKeybase(cc.Keybase).
		WithSignMode(cc.Config.SignMode())
}

func (ccc *ChainClientConfig) SignMode() signing.SignMode {
	signMode := signing.SignMode_SIGN_MODE_UNSPECIFIED
	switch ccc.SignModeStr {
	case "direct":
		signMode = signing.SignMode_SIGN_MODE_DIRECT
	case "amino-json":
		signMode = signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	}
	return signMode
}

func (cc *ChainClient) SendMsg(ctx context.Context, msg sdk.Msg, memo string) (*sdk.TxResponse, error) {
	return cc.SendMsgs(ctx, []sdk.Msg{msg}, memo)
}

// SendMsgs wraps the msgs in a StdTx, signs and sends it. An error is returned if there
// was an issue sending the transaction. A successfully sent, but failed transaction will
// not return an error. If a transaction is successfully sent, the result of the execution
// of that transaction will be logged. A boolean indicating if a transaction was successfully
// sent and executed successfully is returned.
//
// feegranterKey - key name of the address set as the feegranter, empty string will not feegrant
func (cc *ChainClient) SendMsgsWith(ctx context.Context, msgs []sdk.Msg, memo string, gas uint64, signingKey string, feegranterKey string) (*sdk.TxResponse, error) {
	sdkConfigMutex.Lock()
	sdkConf := sdk.GetConfig()
	sdkConf.SetBech32PrefixForAccount(cc.Config.AccountPrefix, cc.Config.AccountPrefix+"pub")
	sdkConf.SetBech32PrefixForValidator(cc.Config.AccountPrefix+"valoper", cc.Config.AccountPrefix+"valoperpub")
	sdkConf.SetBech32PrefixForConsensusNode(cc.Config.AccountPrefix+"valcons", cc.Config.AccountPrefix+"valconspub")
	defer sdkConfigMutex.Unlock()

	rand.Seed(time.Now().UnixNano())
	logId := rand.Int()
	signingKeyAcc, _ := cc.GetKeyAddressForKey(signingKey)
	signingAddr, _ := cc.EncodeBech32AccAddr(signingKeyAcc)
	feegrantKeyAcc, _ := cc.GetKeyAddressForKey(feegranterKey)
	feegrantAddr, _ := cc.EncodeBech32AccAddr(feegrantKeyAcc)
	fmt.Printf("[Lens:%d] SIGNER: %s, signer addr: %s, FEEGRANTER: %s, feegrant addr: %s, chain: %s, \n", logId, signingKey, signingAddr, feegranterKey, feegrantAddr, cc.Config.ChainID)

	txf, err := cc.PrepareFactory(cc.TxFactory(), signingKey, logId)
	if err != nil {
		return nil, err
	}

	feegranterAddr := cc.MustEncodeAccAddr(feegrantKeyAcc)

	adjusted := gas

	if gas == 0 {
		// TODO: Make this work with new CalculateGas method
		// TODO: This is related to GRPC client stuff?
		// https://github.com/cosmos/cosmos-sdk/blob/5725659684fc93790a63981c653feee33ecf3225/client/tx/tx.go#L297
		_, adjusted, err = cc.CalculateGas(ctx, txf, signingKey, feegranterAddr, logId, msgs...)

		if err != nil {
			fmt.Printf("[Lens:%d] err CalculateGas: %s\n", logId, err.Error())
			return nil, err
		}
	}

	//Cannot feegrant your own TX
	if signingKey != feegranterKey && feegranterKey != "" {
		//Must be set in Factory to affect gas calculation (sim tx) as well as real tx
		txf = txf.WithFeeGranter(feegrantKeyAcc)
	}

	if memo != "" {
		txf = txf.WithMemo(memo)
	}

	// Set the gas amount on the transaction factory
	txf = txf.WithGas(adjusted)

	// Build the transaction builder
	txb, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	// Attach the signature to the transaction
	// c.LogFailedTx(nil, err, msgs)
	// Force encoding in the chain specific address
	for _, msg := range msgs {
		cc.Codec.Marshaler.MustMarshalJSON(msg)
	}

	err = func() error {
		//done := cc.SetSDKContext()
		// ensure that we allways call done, even in case of an error or panic
		//defer done()

		if err = tx.Sign(txf, signingKey, txb, false); err != nil {
			return err
		}
		return nil
	}()

	if err != nil {
		return nil, err
	}

	// Generate the transaction bytes
	txBytes, err := cc.Codec.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		return nil, err
	}

	// Broadcast those bytes
	res, err := cc.BroadcastTx(ctx, txBytes)
	if res != nil {
		fmt.Printf("TX hash: %s\n", res.TxHash)
	}
	if err != nil {
		return nil, err
	}

	// transaction was executed, log the success or failure using the tx response code
	// NOTE: error is nil, logic should use the returned error to determine if the
	// transaction was successfully executed.
	if res.Code != 0 {
		return res, fmt.Errorf("transaction failed with code: %d", res.Code)
	}

	return res, nil
}

// SendMsgs wraps the msgs in a StdTx, signs and sends it. An error is returned if there
// was an issue sending the transaction. A successfully sent, but failed transaction will
// not return an error. If a transaction is successfully sent, the result of the execution
// of that transaction will be logged. A boolean indicating if a transaction was successfully
// sent and executed successfully is returned.
func (cc *ChainClient) SendMsgs(ctx context.Context, msgs []sdk.Msg, memo string) (*sdk.TxResponse, error) {
	signer, feegranter, err := cc.GetTxFeeGrant()
	if err != nil {
		return nil, err
	}
	return cc.SendMsgsWith(ctx, msgs, memo, 0, signer, feegranter)
}

func (cc *ChainClient) SubmitTxAwaitResponse(ctx context.Context, msgs []sdk.Msg, memo string, gas uint64, signingKeyName string) (*txtypes.GetTxResponse, error) {
	resp, err := cc.SendMsgsWith(ctx, msgs, memo, gas, signingKeyName, "")
	if err != nil {
		return nil, err
	}
	fmt.Printf("TX result code: %d. Waiting for TX with hash %s\n", resp.Code, resp.TxHash)
	tx1resp, err := cc.AwaitTx(resp.TxHash, 15*time.Second)
	if err != nil {
		return nil, err
	}

	return tx1resp, err
}

// Get the TX by hash, waiting for it to be included in a block
func (cc *ChainClient) AwaitTx(txHash string, timeout time.Duration) (*txtypes.GetTxResponse, error) {
	var txByHash *txtypes.GetTxResponse
	var txLookupErr error
	startTime := time.Now()
	timeBetweenQueries := 100

	txClient := txtypes.NewServiceClient(cc)

	for txByHash == nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		if time.Since(startTime) > timeout {
			cancel()
			return nil, txLookupErr
		}

		txByHash, txLookupErr = txClient.GetTx(ctx, &txtypes.GetTxRequest{Hash: txHash})
		if txLookupErr != nil {
			time.Sleep(time.Duration(timeBetweenQueries) * time.Millisecond)
		}
		cancel()
	}

	return txByHash, nil
}

func (cc *ChainClient) PrepareFactory(txf tx.Factory, signingKey string, id int) (tx.Factory, error) {
	var (
		err      error
		from     sdk.AccAddress
		num, seq uint64
	)

	// Get key address and retry if fail
	if err = retry.Do(func() error {
		from, err = cc.GetKeyAddressForKey(signingKey)
		if err != nil {
			return err
		}
		return err
	}, RtyAtt, RtyDel, RtyErr); err != nil {
		return tx.Factory{}, err
	}

	fmt.Printf("[Lens:%d] PrepareFactory: SigningKey: %s, SDK address: %s\n", id, signingKey, from.String())

	cliCtx := client.Context{}.WithClient(cc.RPCClient).
		WithInterfaceRegistry(cc.Codec.InterfaceRegistry).
		WithChainID(cc.Config.ChainID).
		WithCodec(cc.Codec.Marshaler).
		WithFromAddress(from)

		//.WithFeePayerAddress(from)

	// Account defines a read-only version of the auth module's AccountI.
	// type Account interface {
	// 	GetAddress() sdk.AccAddress
	// 	GetPubKey() cryptotypes.PubKey // can return nil.
	// 	GetAccountNumber() uint64
	// 	GetSequence() uint64
	// }

	// Set the account number and sequence on the transaction factory and retry if fail
	if err = retry.Do(func() error {
		if err = txf.AccountRetriever().EnsureExists(cliCtx, from); err != nil {
			fmt.Printf("[Lens:%d] PrepareFactory: Error EnsureExists: %s\n", id, err.Error())
			return err
		}
		return err
	}, RtyAtt, RtyDel, RtyErr); err != nil {
		return txf, err
	}

	acct, err := txf.AccountRetriever().GetAccount(cliCtx, from)
	if err == nil {
		acctAddr := acct.GetAddress()
		//acctPub := acct.GetPubKey()
		acctSeq := acct.GetSequence()
		acctNum := acct.GetAccountNumber()

		keyInfo, _ := cc.Keybase.Key(signingKey)
		pk, _ := keyInfo.GetPubKey()

		sEnc := "Unknown"
		if pk != nil {
			sEnc = b64.StdEncoding.EncodeToString(pk.Bytes())
		}
		fmt.Printf("Lens[%d]:PrepareFactory: pubkey b64: %s, acctAddr: %s, seq: %d, num: %d\n", id, sEnc, acctAddr.String(), acctSeq, acctNum)
	} else {
		fmt.Printf("Lens[%d]:PrepareFactory: GetAccount err %s\n", id, err.Error())
	}

	// TODO: why this code? this may potentially require another query when we don't want one
	initNum, initSeq := txf.AccountNumber(), txf.Sequence()
	if initNum == 0 || initSeq == 0 {
		if err = retry.Do(func() error {
			num, seq, err = txf.AccountRetriever().GetAccountNumberSequence(cliCtx, from)
			if err != nil {
				fmt.Printf("[Lens:%d] PrepareFactory: Error getting sequence: %s\n", id, err.Error())
				return err
			}
			return err
		}, RtyAtt, RtyDel, RtyErr); err != nil {
			return txf, err
		}

		if initNum == 0 {
			txf = txf.WithAccountNumber(num)
			fmt.Printf("[Lens:%d] PrepareFactory: AccountNumber: %d\n", id, num)
		}

		if initSeq == 0 {
			txf = txf.WithSequence(seq)
			fmt.Printf("[Lens:%d] PrepareFactory: Sequence: %d\n", id, seq)
		}
	}

	if cc.Config.MinGasAmount != 0 {
		txf = txf.WithGas(cc.Config.MinGasAmount)
	}

	return txf, nil
}

// Calculates how much gas is needed by simulating the TX without submitting it on chain (submits to a node, but does not cost gas, etc).
func (cc *ChainClient) CalculateGas(ctx context.Context, txf tx.Factory, signingKey string, feegranterAddr string, id int, msgs ...sdk.Msg) (txtypes.SimulateResponse, uint64, error) {
	keyInfo, err := cc.Keybase.Key(signingKey)
	if err != nil {
		return txtypes.SimulateResponse{}, 0, err
	}

	addr, _ := keyInfo.GetAddress()
	fmt.Printf("Lens[%d]: Signing address in CalcGas: %s", id, addr.String())

	var txBytes []byte
	var txb client.TxBuilder
	if err := retry.Do(func() error {
		var err error
		txBytes, err, txb = BuildSimTx(keyInfo, txf, feegranterAddr, id, msgs...)
		if err != nil {
			return err
		}
		return nil
	}, retry.Context(ctx), RtyAtt, RtyDel, RtyErr); err != nil {
		return txtypes.SimulateResponse{}, 0, err
	}

	simQuery := abci.RequestQuery{
		Path: "/cosmos.tx.v1beta1.Service/Simulate",
		Data: txBytes,
	}

	signingTx := txb.GetTx()
	pubKeys, err := signingTx.GetPubKeys()
	if err != nil {
		fmt.Printf("[Lens:%d] decoder pubkey err: %s \n", id, err.Error())
		return txtypes.SimulateResponse{}, 0, err
	}
	for _, curr := range pubKeys {
		if curr != nil {
			sEnc := b64.StdEncoding.EncodeToString(curr.Bytes())
			fmt.Printf("[Lens:%d] signingTx public key: %s \n", id, sEnc)
		} else {
			fmt.Printf("[Lens:%d] signingTx public key nil \n", id)
		}
	}

	sigs, err := signingTx.GetSignaturesV2()
	if err != nil {
		fmt.Printf("[Lens:%d] decoder sigs err: %s \n", id, err.Error())
		return txtypes.SimulateResponse{}, 0, err
	}
	for _, curr := range sigs {
		if curr.PubKey != nil {
			sEnc := b64.StdEncoding.EncodeToString(curr.PubKey.Bytes())
			fmt.Printf("[Lens:%d] signingTx SIG public key: %s \n", id, sEnc)
		} else {
			fmt.Printf("[Lens:%d] signingTx SIG public key nil \n", id)
		}
	}

	signers := signingTx.GetSigners()
	for _, curr := range signers {
		if curr != nil {
			fmt.Printf("[Lens:%d] signer: %s \n", id, curr.String())
		} else {
			fmt.Printf("[Lens:%d] signer nil \n", id)
		}
	}

	var res abci.ResponseQuery
	if err := retry.Do(func() error {
		var err error
		res, err = cc.QueryABCI(ctx, simQuery)
		if err != nil {
			return err
		}
		return nil
	}, retry.Context(ctx), RtyAtt, RtyDel, RtyErr); err != nil {
		return txtypes.SimulateResponse{}, 0, err
	}

	var simRes txtypes.SimulateResponse
	if err := simRes.Unmarshal(res.Value); err != nil {
		return txtypes.SimulateResponse{}, 0, err
	}

	return simRes, uint64(txf.GasAdjustment() * float64(simRes.GasInfo.GasUsed)), nil
}

func (cc *ChainClient) QueryABCI(ctx context.Context, req abci.RequestQuery) (abci.ResponseQuery, error) {
	opts := rpcclient.ABCIQueryOptions{
		Height: req.Height,
		Prove:  req.Prove,
	}
	result, err := cc.RPCClient.ABCIQueryWithOptions(ctx, req.Path, req.Data, opts)
	if err != nil {
		return abci.ResponseQuery{}, err
	}

	if !result.Response.IsOK() {
		return abci.ResponseQuery{}, sdkErrorToGRPCError(result.Response)
	}

	// data from trusted node or subspace query doesn't need verification
	if !opts.Prove || !isQueryStoreWithProof(req.Path) {
		return result.Response, nil
	}

	return result.Response, nil
}

func sdkErrorToGRPCError(resp abci.ResponseQuery) error {
	switch resp.Code {
	case sdkerrors.ErrInvalidRequest.ABCICode():
		return status.Error(codes.InvalidArgument, resp.Log)
	case sdkerrors.ErrUnauthorized.ABCICode():
		return status.Error(codes.Unauthenticated, resp.Log)
	case sdkerrors.ErrKeyNotFound.ABCICode():
		return status.Error(codes.NotFound, resp.Log)
	default:
		return status.Error(codes.Unknown, resp.Log)
	}
}

// isQueryStoreWithProof expects a format like /<queryType>/<storeName>/<subpath>
// queryType must be "store" and subpath must be "key" to require a proof.
func isQueryStoreWithProof(path string) bool {
	if !strings.HasPrefix(path, "/") {
		return false
	}

	paths := strings.SplitN(path[1:], "/", 3)

	switch {
	case len(paths) != 3:
		return false
	case paths[0] != "store":
		return false
	case rootmulti.RequireProof("/" + paths[2]):
		return true
	}

	return false
}

// protoTxProvider is a type which can provide a proto transaction. It is a
// workaround to get access to the wrapper TxBuilder's method GetProtoTx().
type protoTxProvider interface {
	GetProtoTx() *txtypes.Tx
}

// BuildSimTx creates an unsigned tx with an empty single signature and returns
// the encoded transaction or an error if the unsigned transaction cannot be built.
func BuildSimTx(info *keyring.Record, txf tx.Factory, feegranterAddr string, logId int, msgs ...sdk.Msg) ([]byte, error, client.TxBuilder) {
	txb, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err, nil
	}

	var pk cryptotypes.PubKey = &secp256k1.PubKey{} // use default public key type

	pk, err = info.GetPubKey()
	if err != nil {
		return nil, err, nil
	}

	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	sig := signing.SignatureV2{
		PubKey: pk,
		Data: &signing.SingleSignatureData{
			SignMode: txf.SignMode(),
		},
		Sequence: txf.Sequence(),
	}
	if err := txb.SetSignatures(sig); err != nil {
		return nil, err, nil
	}

	sEnc := b64.StdEncoding.EncodeToString(pk.Bytes())
	fmt.Printf("Lens[%d]: pubkey b64 in simtx (new modified.....): %s\n", logId, sEnc)

	protoProvider, ok := txb.(protoTxProvider)
	if !ok {
		return nil, fmt.Errorf("cannot simulate amino tx"), nil
	}

	tBytes, err := protoProvider.GetProtoTx().Marshal()
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
		panic("marshalling getprototx is a bad idea, bucko")
	}
	//simReq := txtypes.SimulateRequest{Tx: protoProvider.GetProtoTx()}
	simReq := txtypes.SimulateRequest{TxBytes: tBytes}

	//if feegranterAddr != "" {
	//fmt.Printf("Lens[%d]: feegranter in simtx: %s\n", logId, simReq.Tx.AuthInfo.Fee.Granter)
	//simReq.Tx.AuthInfo.Fee.Granter = ""
	//simReq.Tx.AuthInfo.Fee.Granter = feegranterAddr
	//	}
	data, err := simReq.Marshal()
	return data, err, txb
}

func (cc *ChainClient) BuildTx(ctx context.Context, msgs []sdk.Msg, memo string, gas uint64) (
	txBytes []byte,
	sequence uint64,
	fees sdk.Coins,
	err error,
) {
	signingKey, feegranterKey, err := cc.GetTxFeeGrant()
	if err != nil {
		return nil, 0, sdk.Coins{}, err
	}

	rand.Seed(time.Now().UnixNano())
	logId := rand.Int()
	return cc.buildTxWith(ctx, msgs, memo, gas, signingKey, feegranterKey, logId)
}

func (cc *ChainClient) buildTxWith(ctx context.Context, msgs []sdk.Msg, memo string, gas uint64, signingKey string, feegranterKey string, id int) (
	txBytes []byte,
	sequence uint64,
	fees sdk.Coins,
	err error,
) {
	sdkConfigMutex.Lock()
	sdkConf := sdk.GetConfig()
	sdkConf.SetBech32PrefixForAccount(cc.Config.AccountPrefix, cc.Config.AccountPrefix+"pub")
	sdkConf.SetBech32PrefixForValidator(cc.Config.AccountPrefix+"valoper", cc.Config.AccountPrefix+"valoperpub")
	sdkConf.SetBech32PrefixForConsensusNode(cc.Config.AccountPrefix+"valcons", cc.Config.AccountPrefix+"valconspub")
	defer sdkConfigMutex.Unlock()

	signingKeyAcc, _ := cc.GetKeyAddressForKey(signingKey)
	signingAddr, _ := cc.EncodeBech32AccAddr(signingKeyAcc)
	feegrantKeyAcc, _ := cc.GetKeyAddressForKey(feegranterKey)
	feegrantAddr, _ := cc.EncodeBech32AccAddr(feegrantKeyAcc)
	fmt.Printf("[Lens:%d][app=relayer] SIGNER: %s, signer addr: %s, FEEGRANTER: %s, feegrant addr: %s, chain: %s, account prefix: %s \n", id, signingKey, signingAddr, feegranterKey, feegrantAddr, cc.Config.ChainID, cc.Config.AccountPrefix)

	txf, err := cc.PrepareFactory(cc.TxFactory(), signingKey, id)
	if err != nil {
		return nil, 0, sdk.Coins{}, err
	}

	feegranterAddr := ""
	adjusted := gas

	fmt.Printf("[Lens:%d] before CalculateGas\n", id)

	if gas == 0 {
		txf = txf.WithMemo(strconv.Itoa(id)) //ID will help trace on the SDK side
		_, adjusted, err = cc.CalculateGas(ctx, txf, signingKey, feegranterAddr, id, msgs...)

		if err != nil {
			fmt.Printf("[Lens:%d] err CalculateGas: %s\n", id, err.Error())
			return nil, 0, sdk.Coins{}, err
		}
	}
	fmt.Printf("[Lens:%d] after CalculateGas\n", id)

	if memo != "" {
		txf = txf.WithMemo(memo)
	}

	//Cannot feegrant your own TX
	if signingKey != feegranterKey && feegranterKey != "" {
		granterAddr, err := cc.GetKeyAddressForKey(feegranterKey)
		if err != nil {
			return nil, 0, sdk.Coins{}, err
		}

		//Must be set in Factory to affect gas calculation (sim tx) as well as real tx
		txf = txf.WithFeeGranter(granterAddr)
		//txf = txf.WithFeePayer(granterAddr)
		feegranterAddr = cc.MustEncodeAccAddr(granterAddr)
		fmt.Printf("[Lens:%d] feegranter addr: %s \n", id, feegranterAddr)
	}

	// Set the gas amount on the transaction factory
	txf = txf.WithGas(adjusted)

	// Build the transaction builder
	txb, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, 0, sdk.Coins{}, err
	}

	//err = func() error {
	k, err := txf.Keybase().Key(signingKey)
	if err != nil {
		fmt.Printf("[Lens:%d] Key err %s\n", id, err.Error())
		//return err
		return nil, 0, sdk.Coins{}, err
	}

	pubKey, err := k.GetPubKey()
	if err != nil {
		fmt.Printf("[Lens:%d] GetPubKey err %s\n", id, err.Error())
		//return err
		return nil, 0, sdk.Coins{}, err
	}

	fmt.Printf("[Lens:%d] before sdk.AccAddress\n", id)
	addr := sdk.AccAddress(pubKey.Address()).String()
	fmt.Printf("[Lens:%d] Signing TX with key %s, which has address %s\n", id, signingKey, addr)

	if err = tx.Sign(txf, signingKey, txb, false); err != nil {
		//return err
		return nil, 0, sdk.Coins{}, err
	}
	//return nil
	//}()

	if err != nil {
		return nil, 0, sdk.Coins{}, err
	}

	// Generate the transaction bytes
	txBytes, err = cc.Codec.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		return nil, 0, sdk.Coins{}, err
	}

	// decoder := cc.Codec.TxConfig.TxDecoder()
	// sdktx, err := decoder(txBytes)
	// if err != nil {
	// 	fmt.Printf("[Lens:%d] err: %s \n", id, err.Error())
	// 	return nil, 0, sdk.Coins{}, err
	// }
	// sdktx.GetMsgs()
	//fmt.Printf("[Lens:%d] feegrant addr: %s \n", id, feegranterAddr)

	return txBytes, txf.Sequence(), fees, nil
}
