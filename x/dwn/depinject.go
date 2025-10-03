package module

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	modulev1 "github.com/sonr-io/sonr/api/dwn/module/v1"
	didkeeper "github.com/sonr-io/sonr/x/did/keeper"
	"github.com/sonr-io/sonr/x/dwn/keeper"
	svckeeper "github.com/sonr-io/sonr/x/svc/keeper"
)

var _ appmodule.AppModule = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (a AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (a AppModule) IsAppModule() {}

func init() {
	appmodule.Register(
		&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Cdc          codec.Codec
	StoreService store.KVStoreService
	AddressCodec address.Codec

	AccountKeeper  authkeeper.AccountKeeper
	BankKeeper     bankkeeper.Keeper
	StakingKeeper  *stakingkeeper.Keeper
	SlashingKeeper slashingkeeper.Keeper
	FeegrantKeeper feegrantkeeper.Keeper
	DIDKeeper      didkeeper.Keeper
	ServiceKeeper  svckeeper.Keeper
}

type ModuleOutputs struct {
	depinject.Out

	Module appmodule.AppModule
	Keeper keeper.Keeper
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	// Create a default client context for transaction building
	// Note: TxConfig will need to be set when available in the app initialization
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithCodec(in.Cdc)

	k := keeper.NewKeeper(
		in.Cdc,
		in.StoreService,
		log.NewLogger(os.Stderr),
		govAddr,
		in.AccountKeeper,
		in.BankKeeper,
		in.FeegrantKeeper,
		in.StakingKeeper,
		in.DIDKeeper,
		in.ServiceKeeper,
		clientCtx,
	)
	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{Module: m, Keeper: k, Out: depinject.Out{}}
}
