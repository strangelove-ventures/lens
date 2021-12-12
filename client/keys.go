package client

import (
	"errors"
	"os"

	ckeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/go-bip39"
)

func (cc *ChainClient) CreateKeystore(path string) error {
	keybase, err := keyring.New(cc.Config.ChainID, cc.Config.KeyringBackend, cc.Config.KeyDirectory, cc.Input, cc.KeyringOptions...)
	if err != nil {
		return err
	}
	cc.Keybase = keybase
	return nil
}

func (cc *ChainClient) KeystoreCreated(path string) bool {
	if _, err := os.Stat(cc.Config.KeyDirectory); errors.Is(err, os.ErrNotExist) {
		return false
	} else if cc.Keybase == nil {
		return false
	}
	return true
}

func (cc *ChainClient) AddKey(name string) (output *KeyOutput, err error) {
	ko, err := cc.KeyAddOrRestore(name, 118)
	if err != nil {
		return nil, err
	}
	return ko, nil
}

func (cc *ChainClient) RestoreKey(name, mnemonic string) (address string, err error) {
	ko, err := cc.KeyAddOrRestore(name, 118, mnemonic)
	if err != nil {
		return "", err
	}
	return ko.Address, nil
}

func (cc *ChainClient) ShowAddress(name string) (address string, err error) {
	info, err := cc.Keybase.Key(name)
	if err != nil {
		return "", err
	}
	out, err := cc.EncodeBech32AccAddr(info.GetAddress())
	if err != nil {
		return "", err
	}
	return out, nil
}

func (cc *ChainClient) ListAddresses() (map[string]string, error) {
	out := map[string]string{}
	info, err := cc.Keybase.List()
	if err != nil {
		return nil, err
	}
	for _, k := range info {
		addr, err := cc.EncodeBech32AccAddr(k.GetAddress())
		if err != nil {
			return nil, err
		}
		out[k.GetName()] = addr
	}
	return out, nil
}

func (cc *ChainClient) DeleteKey(name string) error {
	if err := cc.Keybase.Delete(name); err != nil {
		return err
	}
	return nil
}

func (cc *ChainClient) KeyExists(name string) bool {
	k, err := cc.Keybase.Key(name)
	if err != nil {
		return false
	}

	return k.GetName() == name

}

func (cc *ChainClient) ExportPrivKeyArmor(keyName string) (armor string, err error) {
	return cc.Keybase.ExportPrivKeyArmor(keyName, ckeys.DefaultKeyPass)
}

func (cc *ChainClient) KeyAddOrRestore(keyName string, coinType uint32, mnemonic ...string) (*KeyOutput, error) {
	var mnemonicStr string
	var err error

	if len(mnemonic) > 0 {
		mnemonicStr = mnemonic[0]
	} else {
		mnemonicStr, err = CreateMnemonic()
		if err != nil {
			return nil, err
		}
	}

	info, err := cc.Keybase.NewAccount(keyName, mnemonicStr, "", hd.CreateHDPath(coinType, 0, 0).String(), hd.Secp256k1)
	if err != nil {
		return nil, err
	}

	out, err := cc.EncodeBech32AccAddr(info.GetAddress())
	if err != nil {
		return nil, err
	}
	return &KeyOutput{Mnemonic: mnemonicStr, Address: out}, nil
}

// KeyOutput contains mnemonic and address of key
type KeyOutput struct {
	Mnemonic string `json:"mnemonic" yaml:"mnemonic"`
	Address  string `json:"address" yaml:"address"`
}

// CreateMnemonic creates a new mnemonic
func CreateMnemonic() (string, error) {
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropySeed)
	if err != nil {
		return "", err
	}
	return mnemonic, nil
}
