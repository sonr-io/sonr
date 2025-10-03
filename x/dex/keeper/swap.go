package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/sonr-io/sonr/x/dex/types"
)

// ExecuteSwap handles swap execution through ICA
func (k Keeper) ExecuteSwap(
	ctx sdk.Context,
	did string,
	connectionID string,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	minAmountOut math.Int,
	poolID uint64,
) (uint64, error) {
	// Get the DEX account
	account, err := k.GetDEXAccount(ctx, did, connectionID)
	if err != nil {
		return 0, fmt.Errorf("DEX account not found: %w", err)
	}

	// Verify account is active
	if account.Status != types.ACCOUNT_STATUS_ACTIVE {
		return 0, fmt.Errorf("DEX account is not active")
	}

	// Create swap message for remote chain
	// This example uses a generic bank send as placeholder
	// Actual implementation would use chain-specific swap messages
	swapMsg := &banktypes.MsgSend{
		FromAddress: account.AccountAddress,
		ToAddress:   account.AccountAddress, // Swap to self as example
		Amount:      sdk.NewCoins(tokenIn),
	}

	// Send the swap transaction via ICA
	sequence, err := k.SendDEXTransaction(
		ctx,
		did,
		connectionID,
		[]sdk.Msg{swapMsg},
		fmt.Sprintf("swap_%s_for_%s", tokenIn.Denom, tokenOutDenom),
		30*time.Second,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to send swap transaction: %w", err)
	}

	// Emit swap event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwapExecuted,
			sdk.NewAttribute("did", did),
			sdk.NewAttribute("connection", connectionID),
			sdk.NewAttribute("token_in", tokenIn.String()),
			sdk.NewAttribute("token_out_denom", tokenOutDenom),
			sdk.NewAttribute("sequence", fmt.Sprintf("%d", sequence)),
		),
	)

	return sequence, nil
}

// BuildOsmosisSwapMsg builds an Osmosis-specific swap message
func (k Keeper) BuildOsmosisSwapMsg(
	senderAddress string,
	poolID uint64,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	minAmountOut math.Int,
) sdk.Msg {
	// This would build an actual Osmosis swap message
	// For now, return a placeholder bank send
	return &banktypes.MsgSend{
		FromAddress: senderAddress,
		ToAddress:   senderAddress,
		Amount:      sdk.NewCoins(tokenIn),
	}
}

// EstimateSwapOutput estimates the output of a swap
func (k Keeper) EstimateSwapOutput(
	ctx sdk.Context,
	connectionID string,
	poolID uint64,
	tokenIn sdk.Coin,
	tokenOutDenom string,
) (math.Int, error) {
	// This would query the remote chain for swap estimation
	// For now, return a placeholder value
	return tokenIn.Amount.MulRaw(95).QuoRaw(100), nil // 95% of input as example
}

// ValidateSwapParameters validates swap parameters
func (k Keeper) ValidateSwapParameters(
	tokenIn sdk.Coin,
	tokenOutDenom string,
	minAmountOut math.Int,
) error {
	if tokenIn.IsZero() {
		return fmt.Errorf("token in amount cannot be zero")
	}

	if tokenOutDenom == "" {
		return fmt.Errorf("token out denomination cannot be empty")
	}

	if tokenIn.Denom == tokenOutDenom {
		return fmt.Errorf("cannot swap same token")
	}

	if minAmountOut.IsNegative() {
		return fmt.Errorf("minimum amount out cannot be negative")
	}

	return nil
}
