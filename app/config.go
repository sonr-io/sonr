// Package app provides configuration utilities for the Sonr blockchain application.
package app

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

// EVMOptionsFn defines a function type for setting app options specifically for
// the app. The function should receive the chainID and return an error if
// any.
type EVMOptionsFn func(string) error

// NoOpEVMOptions is a no-op function that can be used when the app does not
// need any specific configuration.
func NoOpEVMOptions(_ string) error {
	return nil
}

// sealed tracks whether the EVM configuration has been initialized.
// Once sealed, the configuration cannot be changed.
var sealed = false

// ChainsCoinInfo is a map of the chain id and its corresponding EvmCoinInfo
// that allows initializing the app with different coin info based on the
// chain id
var ChainsCoinInfo = map[string]evmtypes.EvmCoinInfo{
	// Default local development chain
	ChainID: {
		Denom:        BaseDenom,
		DisplayDenom: DisplayDenom,
		Decimals:     evmtypes.EighteenDecimals,
	},
	// Starship testnet configuration
	"sonrtestnet_1-1": {
		Denom:        BaseDenom,
		DisplayDenom: DisplayDenom,
		Decimals:     evmtypes.EighteenDecimals,
	},
	"sonr-testnet-1": {
		Denom:        BaseDenom,
		DisplayDenom: DisplayDenom,
		Decimals:     evmtypes.EighteenDecimals,
	},
	// Additional testnet configurations
	"sonr_1-1": {
		Denom:        BaseDenom,
		DisplayDenom: DisplayDenom,
		Decimals:     evmtypes.EighteenDecimals,
	},
}

// EVMAppOptions sets up the global EVM configuration for the chain.
// It configures the base denomination, chain config, and EVM coin info
// based on the provided chain ID. The configuration is sealed after
// first initialization to prevent modifications.
func EVMAppOptions(chainID string) error {
	if sealed {
		return nil
	}

	if chainID == "" {
		chainID = ChainID
	}

	// Try to find config by base chain ID (without suffix)
	id := strings.Split(chainID, "-")[0]
	coinInfo, found := ChainsCoinInfo[id]
	if !found {
		// Try full chain ID
		coinInfo, found = ChainsCoinInfo[chainID]
		if !found {
			// Use a default configuration for unknown chains with warning
			fmt.Printf("Warning: Unknown chain ID %s, using default usnr configuration\n", chainID)
			coinInfo = evmtypes.EvmCoinInfo{
				Denom:        BaseDenom,
				DisplayDenom: DisplayDenom,
				Decimals:     evmtypes.EighteenDecimals,
			}
		}
	}

	// set the denom info for the chain
	if err := setBaseDenom(coinInfo); err != nil {
		return err
	}

	baseDenom, err := sdk.GetBaseDenom()
	if err != nil {
		return err
	}

	ethCfg := evmtypes.DefaultChainConfig(chainID)

	err = evmtypes.NewEVMConfigurator().
		WithChainConfig(ethCfg).
		WithEVMCoinInfo(baseDenom, uint8(coinInfo.Decimals)).
		Configure()
	if err != nil {
		return err
	}

	sealed = true
	return nil
}

// setBaseDenom registers the display denom and base denom and sets the
// base denom for the chain. It ensures proper decimal conversion between
// the base denomination and display denomination.
func setBaseDenom(ci evmtypes.EvmCoinInfo) error {
	if err := sdk.RegisterDenom(ci.DisplayDenom, math.LegacyOneDec()); err != nil {
		return err
	}

	// sdk.RegisterDenom will automatically overwrite the base denom when the
	// new setBaseDenom() are lower than the current base denom's units.
	return sdk.RegisterDenom(ci.Denom, math.LegacyNewDecWithPrec(1, int64(ci.Decimals)))
}
