// Package noop provides a no-operation upgrade handler for testing and development.
// This upgrade performs no actual state migrations but runs module migrations.
package noop

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sonr-io/sonr/app/upgrades"
)

// NewUpgrade creates a new no-operation upgrade with the specified semantic version.
// This upgrade is typically used for testing upgrade mechanisms without performing
// actual state changes beyond standard module migrations.
func NewUpgrade(semver string) upgrades.Upgrade {
	return upgrades.Upgrade{
		UpgradeName:          semver,
		CreateUpgradeHandler: CreateUpgradeHandler,
		StoreUpgrades: storetypes.StoreUpgrades{
			Added:   []string{},
			Deleted: []string{},
		},
	}
}

// CreateUpgradeHandler creates an upgrade handler that performs only module migrations.
// It does not perform any custom upgrade logic, making it suitable for minor version
// upgrades that only require standard module migrations.
func CreateUpgradeHandler(
	mm upgrades.ModuleManager,
	configurator module.Configurator,
	ak *upgrades.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
