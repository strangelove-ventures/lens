package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	// injectivecodec "github.com/InjectiveLabs/sdk-go/chain/crypto/codec"
	// injectiveeth "github.com/InjectiveLabs/sdk-go/chain/crypto/ethsecp256k1"
	// injectivetypes "github.com/InjectiveLabs/sdk-go/chain/types"
	// ethcodec "github.com/evmos/ethermint/crypto/codec"
	// "github.com/evmos/ethermint/crypto/ethsecp256k1"
	// ethermint "github.com/evmos/ethermint/types"
)

type Codec struct {
	InterfaceRegistry types.InterfaceRegistry
	Marshaler         codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

func MakeCodec(moduleBasics []module.AppModuleBasic, extraCodecs []string) Codec {
	modBasic := module.NewBasicManager(moduleBasics...)
	encodingConfig := MakeCodecConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	modBasic.RegisterLegacyAminoCodec(encodingConfig.Amino)
	modBasic.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	// for _, c := range extraCodecs {
	// 	switch c {
	// 	case "ethermint":
	// 		ethcodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	// 		ethermint.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	// 		encodingConfig.Amino.RegisterConcrete(&ethsecp256k1.PubKey{}, ethsecp256k1.PubKeyName, nil)
	// 		encodingConfig.Amino.RegisterConcrete(&ethsecp256k1.PrivKey{}, ethsecp256k1.PrivKeyName, nil)
	// 	case "injective":
	// 		injectivetypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	// 		injectivecodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	// 		encodingConfig.Amino.RegisterConcrete(&injectiveeth.PubKey{}, injectiveeth.PubKeyName, nil)
	// 		encodingConfig.Amino.RegisterConcrete(&injectiveeth.PrivKey{}, injectiveeth.PrivKeyName, nil)
	// 	}
	// }

	return encodingConfig
}

func MakeCodecConfig() Codec {
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	return Codec{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		Amino:             codec.NewLegacyAmino(),
	}
}
