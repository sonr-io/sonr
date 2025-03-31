package keeper

import (
	"context"

	"cosmossdk.io/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/sonr-io/snrd/x/did/types"
)

type msgServer struct {
	k Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{k: keeper}
}

// UpdateParams updates the x/did module parameters.
func (ms msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.k.authority != msg.Authority {
		return nil, errors.Wrapf(
			govtypes.ErrInvalidSigner,
			"invalid authority; expected %s, got %s",
			ms.k.authority,
			msg.Authority,
		)
	}
	return nil, ms.k.Params.Set(ctx, msg.Params)
}

// ExecuteTx implements types.MsgServer.
func (ms msgServer) ExecuteTx(ctx context.Context, msg *types.MsgExecuteTx) (*types.MsgExecuteTxResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.MsgExecuteTxResponse{}, nil
}

// LinkAssertion implements types.MsgServer.
func (ms msgServer) LinkAssertion(ctx context.Context, msg *types.MsgLinkAssertion) (*types.MsgLinkAssertionResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.MsgLinkAssertionResponse{}, nil
}

// LinkAuthentication implements types.MsgServer.
func (ms msgServer) LinkAuthentication(ctx context.Context, msg *types.MsgLinkAuthentication) (*types.MsgLinkAuthenticationResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.MsgLinkAuthenticationResponse{}, nil
}

// UnlinkAssertion implements types.MsgServer.
func (ms msgServer) UnlinkAssertion(ctx context.Context, msg *types.MsgUnlinkAssertion) (*types.MsgUnlinkAssertionResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.MsgUnlinkAssertionResponse{}, nil
}

// UnlinkAuthentication implements types.MsgServer.
func (ms msgServer) UnlinkAuthentication(ctx context.Context, msg *types.MsgUnlinkAuthentication) (*types.MsgUnlinkAuthenticationResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.MsgUnlinkAuthenticationResponse{}, nil
}
