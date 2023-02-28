package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	ethermintcodecs "github.com/strangelove-ventures/lens/client/codecs/ethermint"
	injectivecodecs "github.com/strangelove-ventures/lens/client/codecs/injective"
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
	for _, c := range extraCodecs {
		switch c {
		case "ethermint":
			ethermintcodecs.RegisterInterfaces(encodingConfig.InterfaceRegistry)
			encodingConfig.Amino.RegisterConcrete(&ethermintcodecs.PubKey{}, ethermintcodecs.PubKeyName, nil)
			encodingConfig.Amino.RegisterConcrete(&ethermintcodecs.PrivKey{}, ethermintcodecs.PrivKeyName, nil)
		case "injective":
			injectivecodecs.RegisterInterfaces(encodingConfig.InterfaceRegistry)
			encodingConfig.Amino.RegisterConcrete(&injectivecodecs.PubKey{}, injectivecodecs.PubKeyName, nil)
			encodingConfig.Amino.RegisterConcrete(&injectivecodecs.PrivKey{}, injectivecodecs.PrivKeyName, nil)
		}
	}

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
