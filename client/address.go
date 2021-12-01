package client

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (cc *ChainClient) EncodeBech32AccAddr(addr sdk.AccAddress) (string, error) {
	return sdk.Bech32ifyAddressBytes(cc.Config.AccountPrefix, addr)
}
func (cc *ChainClient) EncodeBech32AccPub(addr sdk.AccAddress) (string, error) {
	return sdk.Bech32ifyAddressBytes(fmt.Sprintf("%s%s", cc.Config.AccountPrefix, "pub"), addr)
}
func (cc *ChainClient) EncodeBech32ValAddr(addr sdk.AccAddress) (string, error) {
	return sdk.Bech32ifyAddressBytes(fmt.Sprintf("%s%s", cc.Config.AccountPrefix, "valoper"), addr)
}
func (cc *ChainClient) EncodeBech32ValPub(addr sdk.AccAddress) (string, error) {
	return sdk.Bech32ifyAddressBytes(fmt.Sprintf("%s%s", cc.Config.AccountPrefix, "valoperpub"), addr)
}
func (cc *ChainClient) EncodeBech32ConsAddr(addr sdk.AccAddress) (string, error) {
	return sdk.Bech32ifyAddressBytes(fmt.Sprintf("%s%s", cc.Config.AccountPrefix, "valcons"), addr)
}
func (cc *ChainClient) EncodeBech32ConsPub(addr sdk.AccAddress) (string, error) {
	return sdk.Bech32ifyAddressBytes(fmt.Sprintf("%s%s", cc.Config.AccountPrefix, "valconspub"), addr)
}

func (cc *ChainClient) DecodeBech32AccAddr(addr string) (sdk.AccAddress, error) {
	return sdk.GetFromBech32(addr, cc.Config.AccountPrefix)
}
func (cc *ChainClient) DecodeBech32AccPub(addr string) (sdk.AccAddress, error) {
	return sdk.GetFromBech32(addr, fmt.Sprintf("%s%s", cc.Config.AccountPrefix, "pub"))
}
func (cc *ChainClient) DecodeBech32ValAddr(addr string) (sdk.AccAddress, error) {
	return sdk.GetFromBech32(addr, fmt.Sprintf("%s%s", cc.Config.AccountPrefix, "valoper"))
}
func (cc *ChainClient) DecodeBech32ValPub(addr string) (sdk.AccAddress, error) {
	return sdk.GetFromBech32(addr, fmt.Sprintf("%s%s", cc.Config.AccountPrefix, "valoperpub"))
}
func (cc *ChainClient) DecodeBech32ConsAddr(addr string) (sdk.AccAddress, error) {
	return sdk.GetFromBech32(addr, fmt.Sprintf("%s%s", cc.Config.AccountPrefix, "valcons"))
}
func (cc *ChainClient) DecodeBech32ConsPub(addr string) (sdk.AccAddress, error) {
	return sdk.GetFromBech32(addr, fmt.Sprintf("%s%s", cc.Config.AccountPrefix, "valconspub"))
}
