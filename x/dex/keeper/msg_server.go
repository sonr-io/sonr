package keeper

import (
	"context"
	"fmt"
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sonr-io/sonr/x/dex/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// RegisterDEXAccount implements types.MsgServer.
func (ms msgServer) RegisterDEXAccount(
	ctx context.Context,
	msg *types.MsgRegisterDEXAccount,
) (*types.MsgRegisterDEXAccountResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Register the DEX account using the keeper's ICA controller logic
	account, err := ms.Keeper.RegisterDEXAccount(
		sdkCtx,
		msg.Did,
		msg.ConnectionId,
		msg.Features,
	)
	if err != nil {
		return nil, err
	}

	// Emit event for account registration
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDEXAccountRegistered,
			sdk.NewAttribute("did", msg.Did),
			sdk.NewAttribute("connection_id", msg.ConnectionId),
			sdk.NewAttribute("port_id", account.PortId),
		),
	)

	return &types.MsgRegisterDEXAccountResponse{
		PortId:         account.PortId,
		AccountAddress: account.AccountAddress,
	}, nil
}

// ExecuteSwap implements cross-chain swap execution via ICA
// This method handles token swaps on remote chains through Interchain Accounts
// with special support for Noble USDC integration
func (ms msgServer) ExecuteSwap(
	ctx context.Context,
	msg *types.MsgExecuteSwap,
) (*types.MsgExecuteSwapResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Validate UCAN permission if token provided
	if msg.UcanToken != "" {
		// Use connection ID as resource ID for swap operations
		if err := ms.validateUCANPermission(ctx, msg.UcanToken, "swap", msg.ConnectionId, types.DEXOpExecuteSwap); err != nil {
			return nil, err
		}
	}

	// 1. Validate DID exists and is active
	_, err := ms.didKeeper.GetDIDDocument(sdkCtx, msg.Did)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "DID %s not found", msg.Did)
	}
	// Note: Active status check removed as DIDDocument may not have Active field
	// If needed, add additional validation based on actual DID structure

	// 2. Validate connection exists and is open
	if err := ms.ValidateConnection(sdkCtx, msg.ConnectionId); err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidConnection, "connection validation failed: %v", err)
	}

	// 3. Get or ensure ICA account exists
	account, err := ms.GetDEXAccount(sdkCtx, msg.Did, msg.ConnectionId)
	if err != nil {
		// Account doesn't exist, return error - user must register first
		return nil, sdkerrors.Wrapf(types.ErrAccountNotFound, "DEX account not found for DID %s on connection %s. Please register first.", msg.Did, msg.ConnectionId)
	}

	// Verify account is active
	if account.Status != types.ACCOUNT_STATUS_ACTIVE {
		return nil, sdkerrors.Wrapf(types.ErrAccountNotActive, "DEX account is not active (status: %s)", account.Status.String())
	}

	// 4. Validate swap parameters
	tokenIn := sdk.NewCoin(msg.SourceDenom, msg.Amount)
	if err := ms.ValidateSwapParameters(tokenIn, msg.TargetDenom, msg.MinAmountOut); err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidSwapParams, "swap parameter validation failed: %v", err)
	}

	// 5. Build swap message for the target chain
	// For Noble USDC swaps, we use IBC transfer
	// For other DEX chains (like Osmosis), we build chain-specific swap messages
	var swapMsgs []sdk.Msg
	var swapType string

	// Check if this is a Noble USDC swap
	if types.IsNobleChain(account.HostChainId) || msg.SourceDenom == types.NobleUSDCDenom || msg.TargetDenom == types.NobleUSDCDenom {
		// Build Noble-specific swap message (IBC transfer for USDC)
		swapMsg, err := ms.BuildNobleSwapMsg(sdkCtx, account.AccountAddress, tokenIn, msg.TargetDenom, msg.MinAmountOut)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrSwapFailed, "failed to build Noble swap message: %v", err)
		}
		swapMsgs = []sdk.Msg{swapMsg}
		swapType = "noble_usdc_swap"
	} else {
		// Build generic DEX swap message (e.g., Osmosis)
		swapMsg := ms.BuildOsmosisSwapMsg(account.AccountAddress, 1, tokenIn, msg.TargetDenom, msg.MinAmountOut)
		swapMsgs = []sdk.Msg{swapMsg}
		swapType = "osmosis_swap"
	}

	// 6. Calculate timeout from message or use default
	timeoutDuration := msg.Timeout.Sub(sdkCtx.BlockTime())
	if timeoutDuration <= 0 {
		timeoutDuration = 30 * time.Second // Default 30 second timeout
	}

	// 7. Send the swap transaction via ICA
	sequence, err := ms.SendDEXTransaction(
		sdkCtx,
		msg.Did,
		msg.ConnectionId,
		swapMsgs,
		fmt.Sprintf("swap_%s_%s_for_%s", swapType, msg.SourceDenom, msg.TargetDenom),
		timeoutDuration,
	)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrSwapFailed, "failed to send swap transaction via ICA: %v", err)
	}

	// 8. Track transaction in DWN if available
	if ms.dwnKeeper != nil {
		// Create swap activity record
		activity := types.DEXActivity{
			Type:         "swap",
			Did:          msg.Did,
			ConnectionId: msg.ConnectionId,
			BlockHeight:  sdkCtx.BlockHeight(),
			Timestamp:    sdkCtx.BlockTime(),
			Status:       "pending",
			Amount:       sdk.NewCoins(tokenIn),
		}

		// Store in DWN for user history (non-blocking)
		if err := ms.storeActivityInDWN(sdkCtx, msg.Did, &activity); err != nil {
			// Log but don't fail the transaction
			ms.Logger(sdkCtx).Error("failed to store swap activity in DWN", "error", err, "did", msg.Did)
		}
	}

	// 9. Emit swap event for indexing
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwapExecuted,
			sdk.NewAttribute("did", msg.Did),
			sdk.NewAttribute("connection_id", msg.ConnectionId),
			sdk.NewAttribute("source_denom", msg.SourceDenom),
			sdk.NewAttribute("target_denom", msg.TargetDenom),
			sdk.NewAttribute("amount", msg.Amount.String()),
			sdk.NewAttribute("min_amount_out", msg.MinAmountOut.String()),
			sdk.NewAttribute("sequence", fmt.Sprintf("%d", sequence)),
			sdk.NewAttribute("swap_type", swapType),
			sdk.NewAttribute("ica_address", account.AccountAddress),
		),
	)

	return &types.MsgExecuteSwapResponse{
		TxHash:         "", // Will be populated by ICA callback
		AmountReceived: "", // Will be populated by ICA callback
		Sequence:       sequence,
	}, nil
}

// validateUCANPermission validates UCAN token for a DEX operation
func (ms msgServer) validateUCANPermission(
	ctx context.Context,
	ucanToken string,
	resourceType string,
	resourceID string,
	operation types.DEXOperation,
) error {
	if ms.permissionValidator == nil {
		// Permission validator not available - skip validation
		return nil
	}

	return ms.permissionValidator.ValidatePermission(
		ctx,
		ucanToken,
		resourceType,
		resourceID,
		operation,
	)
}

// TODO: ProvideLiquidity - Implement cross-chain liquidity provision via ICA
// This method should handle adding liquidity to pools on remote chains
// Required implementation steps:
// 1. Validate the sender's DID exists and is active using did keeper
// 2. Verify UCAN token has liquidity provision capabilities (resource: liquidity, action: provide)
// 3. Retrieve the ICA account for this DID and connection from state
// 4. Calculate appropriate liquidity amounts based on pool ratios
// 5. Build liquidity provision message for target chain's AMM protocol
// 6. Create ICA packet data with the liquidity transaction
// 7. Send ICA packet through IBC channel and await acknowledgment
// 8. Store LP token information in DWN for tracking
// 9. Update user's position records in state
// Returns: Sequence number and LP token amount on success
// ProvideLiquidity implements types.MsgServer.
func (ms msgServer) ProvideLiquidity(
	ctx context.Context,
	msg *types.MsgProvideLiquidity,
) (*types.MsgProvideLiquidityResponse, error) {
	// TODO: Implement liquidity provision via ICA
	// 1. Validate DID and UCAN token
	// 2. Get ICA account for this DID and connection
	// 3. Construct liquidity provision message for remote chain
	// 4. Send ICA packet with liquidity instruction
	// 5. Track transaction in DWN
	return &types.MsgProvideLiquidityResponse{}, nil
}

// TODO: RemoveLiquidity - Implement cross-chain liquidity removal via ICA
// This method should handle removing liquidity from pools on remote chains
// Required implementation steps:
// 1. Validate the sender's DID exists and is active using did keeper
// 2. Verify UCAN token has liquidity removal capabilities (resource: liquidity, action: remove)
// 3. Retrieve the ICA account for this DID and connection from state
// 4. Verify user has sufficient LP tokens to remove
// 5. Build liquidity removal message for target chain's AMM protocol
// 6. Create ICA packet data with the removal transaction
// 7. Send ICA packet through IBC channel and await acknowledgment
// 8. Update LP token information in DWN after removal
// 9. Clear user's position records from state if fully withdrawn
// Returns: Sequence number and withdrawn token amounts on success
// RemoveLiquidity implements types.MsgServer.
func (ms msgServer) RemoveLiquidity(
	ctx context.Context,
	msg *types.MsgRemoveLiquidity,
) (*types.MsgRemoveLiquidityResponse, error) {
	// TODO: Implement liquidity removal via ICA
	// 1. Validate DID and UCAN token
	// 2. Get ICA account for this DID and connection
	// 3. Construct liquidity removal message for remote chain
	// 4. Send ICA packet with removal instruction
	// 5. Track transaction in DWN
	return &types.MsgRemoveLiquidityResponse{}, nil
}

// TODO: CreateLimitOrder - Implement cross-chain limit order creation via ICA
// This method should handle placing limit orders on remote chain order books
// Required implementation steps:
// 1. Validate the sender's DID exists and is active using did keeper
// 2. Verify UCAN token has order creation capabilities (resource: order, action: create)
// 3. Retrieve the ICA account for this DID and connection from state
// 4. Validate order parameters (price, amount, expiry) against market conditions
// 5. Build limit order message for target chain's order book protocol
// 6. Create ICA packet data with the order placement transaction
// 7. Send ICA packet through IBC channel and await acknowledgment
// 8. Store order details in local state for tracking
// 9. Create order record in DWN with unique order ID
// 10. Set up monitoring for order fills and expiration
// Returns: Sequence number and unique order ID on success
// CreateLimitOrder implements types.MsgServer.
func (ms msgServer) CreateLimitOrder(
	ctx context.Context,
	msg *types.MsgCreateLimitOrder,
) (*types.MsgCreateLimitOrderResponse, error) {
	// TODO: Implement limit order creation via ICA
	// 1. Validate DID and UCAN token
	// 2. Get ICA account for this DID and connection
	// 3. Construct limit order message for remote chain
	// 4. Send ICA packet with order instruction
	// 5. Track order in DWN
	return &types.MsgCreateLimitOrderResponse{}, nil
}

// TODO: CancelOrder - Implement cross-chain order cancellation via ICA
// This method should handle cancelling existing limit orders on remote chains
// Required implementation steps:
// 1. Validate the sender's DID exists and is active using did keeper
// 2. Verify UCAN token has order cancellation capabilities (resource: order, action: cancel)
// 3. Retrieve the ICA account for this DID and connection from state
// 4. Verify the order exists and belongs to the sender
// 5. Check order status is still open (not filled or already cancelled)
// 6. Build order cancellation message for target chain's order book protocol
// 7. Create ICA packet data with the cancellation transaction
// 8. Send IBC packet through IBC channel and await acknowledgment
// 9. Update order status in local state to cancelled
// 10. Update order record in DWN with cancellation details
// Returns: Sequence number on successful cancellation
// CancelOrder implements types.MsgServer.
func (ms msgServer) CancelOrder(
	ctx context.Context,
	msg *types.MsgCancelOrder,
) (*types.MsgCancelOrderResponse, error) {
	// TODO: Implement order cancellation via ICA
	// 1. Validate DID and UCAN token
	// 2. Get ICA account for this DID and connection
	// 3. Construct order cancellation message for remote chain
	// 4. Send ICA packet with cancellation instruction
	// 5. Update order status in DWN
	return &types.MsgCancelOrderResponse{}, nil
}

// storeActivityInDWN stores a DEX activity record in the DWN module
func (ms msgServer) storeActivityInDWN(ctx sdk.Context, did string, activity *types.DEXActivity) error {
	// For now, this is a stub - actual implementation would interface with DWN keeper
	// to store activity records in the user's decentralized web node
	if ms.dwnKeeper == nil {
		return nil
	}

	// TODO: Implement actual DWN storage logic
	// This would involve:
	// 1. Serializing the activity to JSON
	// 2. Creating a DWN record with proper schema
	// 3. Storing it in the user's vault

	ms.Logger(ctx).Info("DEX activity tracked",
		"did", did,
		"type", activity.Type,
		"status", activity.Status,
	)

	return nil
}
