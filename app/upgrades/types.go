// Package upgrades provides types and interfaces for chain upgrade handling.
package upgrades

import (
	"context"

	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
)

// AppKeepers holds references to all the keepers needed during chain upgrades.
// This struct is passed to upgrade handlers to provide access to various
// module keepers and core functionality.
type AppKeepers struct {
	AccountKeeper         *authkeeper.AccountKeeper
	ParamsKeeper          *paramskeeper.Keeper
	ConsensusParamsKeeper *consensusparamkeeper.Keeper
	Codec                 codec.Codec
	GetStoreKey           func(storeKey string) *storetypes.KVStoreKey
	CapabilityKeeper      *capabilitykeeper.Keeper
	IBCKeeper             *ibckeeper.Keeper
}

// ModuleManager defines the interface for running module migrations during upgrades.
// It provides methods to execute migrations and retrieve the current version map.
type ModuleManager interface {
	RunMigrations(
		ctx context.Context,
		cfg module.Configurator,
		fromVM module.VersionMap,
	) (module.VersionMap, error)
	GetVersionMap() module.VersionMap
}

// Upgrade defines a struct containing necessary fields that a SoftwareUpgradeProposal
// must have written, in order for the state migration to go smoothly.
// An upgrade must implement this struct, and then set it in the app.go.
// The app.go will then define the handler.
type Upgrade struct {
	// UpgradeName is the version name for the upgrade handler, e.g. `v7`.
	// This must match the upgrade name in the governance proposal.
	UpgradeName string

	// CreateUpgradeHandler defines the function that creates an upgrade handler.
	// The handler performs the actual upgrade logic when the chain reaches the upgrade height.
	CreateUpgradeHandler func(ModuleManager, module.Configurator, *AppKeepers) upgradetypes.UpgradeHandler
	// StoreUpgrades defines any store migrations needed for this upgrade,
	// including adding, renaming, or deleting stores.
	StoreUpgrades storetypes.StoreUpgrades
}
