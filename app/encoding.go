// Package app provides encoding utilities for the Sonr blockchain application.
package app

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/log"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/sonr/app/params"
)

// MakeEncodingConfig creates a new EncodingConfig with all modules registered.
// This function is intended for testing purposes only, as it creates a temporary
// application instance to extract the encoding configuration.
func MakeEncodingConfig(t testing.TB) params.EncodingConfig {
	t.Helper()
	// we "pre"-instantiate the application for getting the injected/configured encoding configuration
	// note, this is not necessary when using app wiring, as depinject can be directly used (see root_v2.go)
	tempApp := NewChainApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
		EVMAppOptions,
	)
	return makeEncodingConfig(tempApp)
}

// makeEncodingConfig extracts the encoding configuration from a ChainApp instance.
// It returns an EncodingConfig struct containing the interface registry, codec,
// transaction config, and amino codec.
func makeEncodingConfig(tempApp *ChainApp) params.EncodingConfig {
	encodingConfig := params.EncodingConfig{
		InterfaceRegistry: tempApp.InterfaceRegistry(),
		Codec:             tempApp.AppCodec(),
		TxConfig:          tempApp.TxConfig(),
		Amino:             tempApp.LegacyAmino(),
	}
	return encodingConfig
}

// GetInterfaceRegistry creates and returns a new interface registry with proper
// address codecs configured. This registry is used for protobuf Any type
// registration and message routing.
func GetInterfaceRegistry() types.InterfaceRegistry {
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})
	if err != nil {
		panic(err)
	}
	return interfaceRegistry
}
