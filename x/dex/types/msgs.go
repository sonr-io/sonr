package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

// ValidateBasic performs basic validation of MsgRegisterDEXAccount
func (msg *MsgRegisterDEXAccount) ValidateBasic() error {
	if msg.Did == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "DID cannot be empty")
	}
	if msg.ConnectionId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "connection ID cannot be empty")
	}
	return nil
}

// ValidateBasic performs basic validation of MsgExecuteSwap
func (msg *MsgExecuteSwap) ValidateBasic() error {
	if msg.Did == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "DID cannot be empty")
	}
	if msg.ConnectionId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "connection ID cannot be empty")
	}
	if msg.SourceDenom == "" || msg.TargetDenom == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "denoms cannot be empty")
	}
	if msg.Amount.IsNil() || !msg.Amount.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "amount must be positive")
	}
	if msg.MinAmountOut.IsNil() || !msg.MinAmountOut.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "min amount out must be positive")
	}
	return nil
}

// ValidateBasic performs basic validation of MsgProvideLiquidity
func (msg *MsgProvideLiquidity) ValidateBasic() error {
	if msg.Did == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "DID cannot be empty")
	}
	if msg.ConnectionId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "connection ID cannot be empty")
	}
	if msg.PoolId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "pool ID cannot be empty")
	}
	if len(msg.Assets) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "assets cannot be empty")
	}
	for _, asset := range msg.Assets {
		if !asset.IsValid() || !asset.IsPositive() {
			return errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				fmt.Sprintf("invalid asset amount: %s", asset),
			)
		}
	}
	if msg.MinShares.IsNil() || !msg.MinShares.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "min shares must be positive")
	}
	return nil
}

// ValidateBasic performs basic validation of MsgRemoveLiquidity
func (msg *MsgRemoveLiquidity) ValidateBasic() error {
	if msg.Did == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "DID cannot be empty")
	}
	if msg.ConnectionId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "connection ID cannot be empty")
	}
	if msg.PoolId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "pool ID cannot be empty")
	}
	if msg.Shares.IsNil() || !msg.Shares.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "shares must be positive")
	}
	return nil
}

// ValidateBasic performs basic validation of MsgCreateLimitOrder
func (msg *MsgCreateLimitOrder) ValidateBasic() error {
	if msg.Did == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "DID cannot be empty")
	}
	if msg.ConnectionId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "connection ID cannot be empty")
	}
	if msg.SellDenom == "" || msg.BuyDenom == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "denoms cannot be empty")
	}
	if msg.Amount.IsNil() || !msg.Amount.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "amount must be positive")
	}
	if msg.Price.IsNil() || !msg.Price.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "price must be positive")
	}
	return nil
}

// ValidateBasic performs basic validation of MsgCancelOrder
func (msg *MsgCancelOrder) ValidateBasic() error {
	if msg.Did == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "DID cannot be empty")
	}
	if msg.ConnectionId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "connection ID cannot be empty")
	}
	if msg.OrderId == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "order ID cannot be empty")
	}
	return nil
}
