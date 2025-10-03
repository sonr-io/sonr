package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmante "github.com/cosmos/evm/ante/evm"
)

// newMonoEVMAnteHandler creates the ante handler for Ethereum Virtual Machine transactions.
// It uses a single decorator that handles all EVM-specific validation and processing,
// including gas calculation, fee market dynamics, and account management.
//
// The mono decorator performs all EVM ante operations in a single pass for efficiency.
func newMonoEVMAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		evmante.NewEVMMonoDecorator(
			options.AccountKeeper,
			options.FeeMarketKeeper,
			options.EvmKeeper,
			options.ControlPanelKeeper,
			options.MaxTxGasWanted,
		),
	)
}
