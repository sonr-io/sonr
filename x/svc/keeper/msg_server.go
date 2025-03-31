package keeper

import (
	"context"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"cosmossdk.io/errors"
	"github.com/sonr-io/snrd/x/svc/types"
)

type msgServer struct {
	k Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{k: keeper}
}

func (ms msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.k.authority != msg.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.k.authority, msg.Authority)
	}

	return nil, ms.k.Params.Set(ctx, msg.Params)
}

// RegisterService implements types.MsgServer.
func (ms msgServer) RegisterService(ctx context.Context, msg *types.MsgRegisterService) (*types.MsgRegisterServiceResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	panic("RegisterService is unimplemented")
	return &types.MsgRegisterServiceResponse{}, nil
}
