package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/snrd/x/svc/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func NewQuerier(keeper Keeper) Querier {
	return Querier{Keeper: keeper}
}

func (k Querier) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	p, err := k.Keeper.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: &p}, nil
}

// OriginExists implements types.QueryServer.
func (k Querier) OriginExists(goCtx context.Context, req *types.QueryOriginExistsRequest) (*types.QueryOriginExistsResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.QueryOriginExistsResponse{}, nil
}

// ResolveOrigin implements types.QueryServer.
func (k Querier) ResolveOrigin(goCtx context.Context, req *types.QueryResolveOriginRequest) (*types.QueryResolveOriginResponse, error) {
	// ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.QueryResolveOriginResponse{}, nil
}
