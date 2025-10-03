package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/sonr-io/sonr/x/dex/types"
)

// CreateLimitOrder creates a limit order through ICA
func (k Keeper) CreateLimitOrder(
	ctx sdk.Context,
	did string,
	connectionID string,
	tokenIn sdk.Coin,
	tokenOutDenom string,
	price math.LegacyDec,
	orderType OrderType,
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

	// Create limit order message for remote chain
	// This is a placeholder - actual implementation would use chain-specific messages
	orderMsg := &banktypes.MsgSend{
		FromAddress: account.AccountAddress,
		ToAddress:   account.AccountAddress, // Placeholder
		Amount:      sdk.NewCoins(tokenIn),
	}

	// Send the order transaction via ICA
	sequence, err := k.SendDEXTransaction(
		ctx,
		did,
		connectionID,
		[]sdk.Msg{orderMsg},
		fmt.Sprintf("limit_order_%s_for_%s", tokenIn.Denom, tokenOutDenom),
		30*time.Second,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to send order transaction: %w", err)
	}

	// Store order ID mapping (sequence -> order details)
	orderID := fmt.Sprintf("%s_%s_%d", did, connectionID, sequence)

	// Emit order created event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOrderCreated,
			sdk.NewAttribute("did", did),
			sdk.NewAttribute("connection", connectionID),
			sdk.NewAttribute("order_id", orderID),
			sdk.NewAttribute("token_in", tokenIn.String()),
			sdk.NewAttribute("token_out", tokenOutDenom),
			sdk.NewAttribute("price", price.String()),
			sdk.NewAttribute("sequence", fmt.Sprintf("%d", sequence)),
		),
	)

	return sequence, nil
}

// CancelOrder cancels an existing order through ICA
func (k Keeper) CancelOrder(
	ctx sdk.Context,
	did string,
	connectionID string,
	orderID string,
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

	// Create cancel order message for remote chain
	// This is a placeholder - actual implementation would use chain-specific messages
	cancelMsg := &banktypes.MsgSend{
		FromAddress: account.AccountAddress,
		ToAddress:   account.AccountAddress, // Placeholder
		Amount:      sdk.NewCoins(),         // Empty amount for cancel
	}

	// Send the cancel transaction via ICA
	sequence, err := k.SendDEXTransaction(
		ctx,
		did,
		connectionID,
		[]sdk.Msg{cancelMsg},
		fmt.Sprintf("cancel_order_%s", orderID),
		30*time.Second,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to send cancel transaction: %w", err)
	}

	// Emit order cancelled event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOrderCancelled,
			sdk.NewAttribute("did", did),
			sdk.NewAttribute("connection", connectionID),
			sdk.NewAttribute("order_id", orderID),
			sdk.NewAttribute("sequence", fmt.Sprintf("%d", sequence)),
		),
	)

	return sequence, nil
}

// OrderType represents the type of order
type OrderType int

const (
	OrderTypeLimit OrderType = iota
	OrderTypeMarket
	OrderTypeStopLoss
	OrderTypeTakeProfit
)

// OrderStatus represents the status of an order
type OrderStatus int

const (
	OrderStatusPending OrderStatus = iota
	OrderStatusOpen
	OrderStatusPartiallyFilled
	OrderStatusFilled
	OrderStatusCancelled
	OrderStatusExpired
)

// OrderInfo represents order information
type OrderInfo struct {
	OrderID         string
	DID             string
	ConnectionID    string
	TokenIn         sdk.Coin
	TokenOut        string
	Price           math.LegacyDec
	Type            OrderType
	Status          OrderStatus
	FilledAmount    math.Int
	RemainingAmount math.Int
	CreatedAt       int64
	UpdatedAt       int64
}

// GetOrderInfo retrieves order information
func (k Keeper) GetOrderInfo(
	ctx sdk.Context,
	did string,
	connectionID string,
	orderID string,
) (*OrderInfo, error) {
	// This would retrieve order info from state or remote chain
	// For now, return placeholder data
	return &OrderInfo{
		OrderID:         orderID,
		DID:             did,
		ConnectionID:    connectionID,
		TokenIn:         sdk.NewCoin("uatom", math.NewInt(1000)),
		TokenOut:        "uosmo",
		Price:           math.LegacyNewDec(10),
		Type:            OrderTypeLimit,
		Status:          OrderStatusOpen,
		FilledAmount:    math.ZeroInt(),
		RemainingAmount: math.NewInt(1000),
		CreatedAt:       ctx.BlockTime().Unix(),
		UpdatedAt:       ctx.BlockTime().Unix(),
	}, nil
}

// GetOrdersByDID retrieves all orders for a DID
func (k Keeper) GetOrdersByDID(
	ctx sdk.Context,
	did string,
	status OrderStatus,
) ([]*OrderInfo, error) {
	// This would query orders from state or remote chain
	// For now, return empty list
	return []*OrderInfo{}, nil
}

// ValidateOrderParameters validates order parameters
func (k Keeper) ValidateOrderParameters(
	tokenIn sdk.Coin,
	tokenOutDenom string,
	price math.LegacyDec,
	orderType OrderType,
) error {
	if tokenIn.IsZero() {
		return fmt.Errorf("token in amount cannot be zero")
	}

	if tokenOutDenom == "" {
		return fmt.Errorf("token out denomination cannot be empty")
	}

	if tokenIn.Denom == tokenOutDenom {
		return fmt.Errorf("cannot create order with same token")
	}

	if price.IsNegative() || price.IsZero() {
		return fmt.Errorf("price must be positive")
	}

	if orderType < OrderTypeLimit || orderType > OrderTypeTakeProfit {
		return fmt.Errorf("invalid order type")
	}

	return nil
}
