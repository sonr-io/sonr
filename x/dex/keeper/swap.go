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

// BuildNobleSwapMsg builds a Noble-specific swap message using IBC transfer
// Noble swaps typically involve transferring USDC between chains via IBC
func (k Keeper) BuildNobleSwapMsg(
	ctx sdk.Context,
	senderAddress string,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	minAmountOut math.Int,
) (sdk.Msg, error) {
	// Validate Noble swap parameters
	params := types.NobleSwapParams{
		InputDenom:  tokenIn.Denom,
		OutputDenom: tokenOutDenom,
		Amount:      tokenIn.Amount,
		MinOutput:   minAmountOut,
		Receiver:    senderAddress,
	}

	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Noble swap params: %w", err)
	}

	// For Noble USDC, we primarily use IBC transfers
	// In a full implementation, this would:
	// 1. Determine the IBC channel to Noble
	// 2. Build an IBC transfer message to send USDC
	// 3. Include proper timeout and memo for swap routing

	// For now, return a placeholder bank send message
	// In production, this should be an IBC transfer message:
	// transferMsg := &transfertypes.MsgTransfer{
	//     SourcePort:       "transfer",
	//     SourceChannel:    channelID,
	//     Token:            tokenIn,
	//     Sender:           senderAddress,
	//     Receiver:         senderAddress,
	//     TimeoutHeight:    clienttypes.NewHeight(0, 0),
	//     TimeoutTimestamp: uint64(ctx.BlockTime().Add(30 * time.Second).UnixNano()),
	//     Memo:             fmt.Sprintf("swap:%s:%s", tokenOutDenom, minAmountOut.String()),
	// }

	return &banktypes.MsgSend{
		FromAddress: senderAddress,
		ToAddress:   senderAddress,
		Amount:      sdk.NewCoins(tokenIn),
	}, nil
}

// BuildSwapRoute determines the optimal swap route, potentially using USDC as intermediary
func (k Keeper) BuildSwapRoute(
	ctx sdk.Context,
	tokenInDenom string,
	tokenOutDenom string,
	connectionID string,
) ([]types.TradingPair, error) {
	// Simple routing logic:
	// 1. If either token is USDC, direct swap
	// 2. Otherwise, route through USDC as intermediary

	if tokenInDenom == types.NobleUSDCDenom || tokenOutDenom == types.NobleUSDCDenom {
		// Direct swap
		return []types.TradingPair{
			{
				Base:        tokenInDenom,
				Quote:       tokenOutDenom,
				Description: fmt.Sprintf("%s/%s direct", tokenInDenom, tokenOutDenom),
			},
		}, nil
	}

	// Route through USDC
	return []types.TradingPair{
		{
			Base:        tokenInDenom,
			Quote:       types.NobleUSDCDenom,
			Description: fmt.Sprintf("%s/USDC", tokenInDenom),
		},
		{
			Base:        types.NobleUSDCDenom,
			Quote:       tokenOutDenom,
			Description: fmt.Sprintf("USDC/%s", tokenOutDenom),
		},
	}, nil
}

// EstimateNobleSwapOutput estimates output for a Noble USDC swap
func (k Keeper) EstimateNobleSwapOutput(
	ctx sdk.Context,
	tokenIn sdk.Coin,
	tokenOutDenom string,
) (math.Int, error) {
	// In a full implementation, this would:
	// 1. Query Noble chain for current exchange rates
	// 2. Query any DEX pools for pricing
	// 3. Calculate expected output accounting for fees

	// For now, use a simple 1% fee model
	estimatedOutput := tokenIn.Amount.MulRaw(99).QuoRaw(100)

	return estimatedOutput, nil
}

// CalculateSwapSlippage calculates the slippage percentage for a swap
func (k Keeper) CalculateSwapSlippage(
	expectedOutput math.Int,
	minOutput math.Int,
) math.LegacyDec {
	if expectedOutput.IsZero() {
		return math.LegacyZeroDec()
	}

	slippage := math.LegacyNewDecFromInt(expectedOutput.Sub(minOutput)).Quo(math.LegacyNewDecFromInt(expectedOutput))
	return slippage.Mul(math.LegacyNewDec(100)) // Convert to percentage
}
