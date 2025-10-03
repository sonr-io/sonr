package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UCANGaslessDecorator allows gasless transactions for UCAN-authorized operations
type UCANGaslessDecorator struct {
	feeDecorator sdk.AnteDecorator
}

// NewUCANGaslessDecorator creates a new UCAN gasless decorator
func NewUCANGaslessDecorator(feeDecorator sdk.AnteDecorator) UCANGaslessDecorator {
	return UCANGaslessDecorator{
		feeDecorator: feeDecorator,
	}
}

// AnteHandle conditionally skips fee deduction for UCAN gasless transactions
func (ugd UCANGaslessDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// Check if transaction is marked as UCAN gasless
	if ctx.Value("gasless_ucan") != nil {
		// Skip fee deduction for gasless UCAN transaction
		return next(ctx, tx, simulate)
	}

	// Apply normal fee deduction
	return ugd.feeDecorator.AnteHandle(ctx, tx, simulate, next)
}
