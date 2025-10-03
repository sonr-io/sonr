package types

import sdkerrors "cosmossdk.io/errors"

var (
	ErrInvalidGenesisState    = sdkerrors.Register(ModuleName, 1, "invalid genesis state")
	ErrInvalidActivityType    = sdkerrors.Register(ModuleName, 2, "invalid activity type")
	ErrInvalidDID             = sdkerrors.Register(ModuleName, 3, "invalid DID")
	ErrInvalidConnectionID    = sdkerrors.Register(ModuleName, 4, "invalid connection ID")
	ErrAccountNotFound        = sdkerrors.Register(ModuleName, 5, "DEX account not found")
	ErrAccountNotActive       = sdkerrors.Register(ModuleName, 6, "DEX account not active")
	ErrUnauthorized           = sdkerrors.Register(ModuleName, 7, "unauthorized")
	ErrInvalidSwapParams      = sdkerrors.Register(ModuleName, 8, "invalid swap parameters")
	ErrInvalidLiquidityParams = sdkerrors.Register(ModuleName, 9, "invalid liquidity parameters")
	ErrInvalidOrderParams     = sdkerrors.Register(ModuleName, 10, "invalid order parameters")
	ErrICAOperationFailed     = sdkerrors.Register(ModuleName, 11, "ICA operation failed")
)
