package keeper

import (
	"context"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/snrd/x/did/types"
)

//	func (k Keeper) ResolveController(ctx sdk.Context, did string) (controller.ControllerI, error) {
//		ct, err := k.OrmDB.ControllerTable().GetByDid(ctx, did)
//		if err != nil {
//			return nil, err
//		}
//		c, err := controller.LoadFromTableEntry(ctx, ct)
//		if err != nil {
//			return nil, err
//		}
//		return c, nil
//	}
//
// Logger returns the logger
func (k Keeper) Logger() log.Logger {
	return k.logger
}

// InitGenesis initializes the module's state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
	// this line is used by starport scaffolding # genesis/module/init
	if err := data.Params.Validate(); err != nil {
		return err
	}
	return k.Params.Set(ctx, data.Params)
}

// ExportGenesis exports the module's state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	// this line is used by starport scaffolding # genesis/module/export
	return &types.GenesisState{
		Params: params,
	}
}

// CurrentParams returns the current schema
func (k Keeper) CurrentParams(ctx sdk.Context) (*types.Params, error) {
	p, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
