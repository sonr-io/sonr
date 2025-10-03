package keeper

import (
	"fmt"

	"github.com/sonr-io/sonr/x/dex/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a specified GenesisState
func (k Keeper) InitGenesis(ctx sdk.Context, state types.GenesisState) {
	// Set params
	if err := k.Params.Set(ctx, state.Params); err != nil {
		panic(fmt.Sprintf("failed to set params: %v", err))
	}

	// Set port ID - use default if empty
	portID := state.PortId
	if portID == "" {
		portID = types.PortID
	}

	// Only try to bind to port if it is not already bound
	if !k.IsBound(ctx, portID) {
		// Module binds to the port on InitChain
		// and claims the returned capability
		if err := k.BindPort(ctx, portID); err != nil {
			panic(fmt.Sprintf("could not claim port capability: %v", err))
		}
	}

	// Restore accounts
	for _, account := range state.Accounts {
		accountKey := GetAccountKey(account.Did, account.ConnectionId)
		if err := k.Accounts.Set(ctx, accountKey, *account); err != nil {
			panic(fmt.Sprintf("failed to set account: %v", err))
		}
	}

	// Set account sequence
	if err := k.AccountSequence.Set(ctx, state.AccountSequence); err != nil {
		panic(fmt.Sprintf("failed to set account sequence: %v", err))
	}
}

// ExportGenesis exports the module's state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		params = types.Params{} // Use default params if not set
	}

	var accounts []*types.InterchainDEXAccount
	err = k.Accounts.Walk(
		ctx,
		nil,
		func(key string, value types.InterchainDEXAccount) (bool, error) {
			accounts = append(accounts, &value)
			return false, nil
		},
	)
	if err != nil {
		panic(fmt.Sprintf("failed to export accounts: %v", err))
	}

	sequence, err := k.AccountSequence.Peek(ctx)
	if err != nil {
		sequence = 0
	}

	return &types.GenesisState{
		Params:          params,
		PortId:          types.PortID,
		Accounts:        accounts,
		AccountSequence: sequence,
	}
}

// IsBound checks if the port is already bound
func (k Keeper) IsBound(ctx sdk.Context, portID string) bool {
	_, ok := k.ScopedKeeper.GetCapability(ctx, fmt.Sprintf("ports/%s", portID))
	return ok
}
