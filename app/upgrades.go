// Package app provides upgrade handling functionality for the Sonr blockchain.
package app

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/sonr-io/sonr/app/upgrades"
	"github.com/sonr-io/sonr/app/upgrades/noop"
)

// Upgrades contains the list of chain upgrades to be applied.
// Each upgrade defines the upgrade name, handler, and store migrations.
var Upgrades = []upgrades.Upgrade{}

// RegisterUpgradeHandlers registers the chain upgrade handlers for all defined upgrades.
// It sets up the upgrade handlers with the module manager and configurator,
// and configures the store loader for the current upgrade if applicable.
// If no upgrades are defined, it registers a no-op upgrade for testing purposes.
func (app *ChainApp) RegisterUpgradeHandlers() {
	// setupLegacyKeyTables(&app.ParamsKeeper)
	if len(Upgrades) == 0 {
		// always have a unique upgrade registered for the current version to test in system tests
		Upgrades = append(Upgrades, noop.NewUpgrade(app.Version()))
	}

	keepers := upgrades.AppKeepers{
		AccountKeeper:         &app.AccountKeeper,
		ParamsKeeper:          &app.ParamsKeeper,
		ConsensusParamsKeeper: &app.ConsensusParamsKeeper,
		CapabilityKeeper:      app.CapabilityKeeper,
		IBCKeeper:             app.IBCKeeper,
		Codec:                 app.appCodec,
		GetStoreKey:           app.GetKey,
	}

	// register all upgrade handlers
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.ModuleManager,
				app.configurator,
				&keepers,
			),
		)
	}

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	// register store loader for current upgrade
	for _, upgrade := range Upgrades {
		if upgradeInfo.Name == upgrade.UpgradeName {
			app.SetStoreLoader(
				upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrade.StoreUpgrades),
			) // nolint:gosec
			break
		}
	}
}
