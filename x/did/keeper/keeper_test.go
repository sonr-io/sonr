package keeper_test

import (
	"testing"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	nftkeeper "cosmossdk.io/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/poa"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	module "github.com/sonr-io/snrd/x/did"
	"github.com/sonr-io/snrd/x/did/keeper"
	"github.com/sonr-io/snrd/x/did/types"
)

var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:     nil,
	stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	minttypes.ModuleName:           {authtypes.Minter},
	govtypes.ModuleName:            {authtypes.Burner},
}

type testFixture struct {
	suite.Suite

	ctx         sdk.Context
	k           keeper.Keeper
	msgServer   types.MsgServer
	queryServer types.QueryServer
	appModule   *module.AppModule

	accountkeeper authkeeper.AccountKeeper
	bankkeeper    bankkeeper.BaseKeeper
	nftKeeper     nftkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper
	mintkeeper    mintkeeper.Keeper

	addrs      []sdk.AccAddress
	govModAddr string
}

func SetupTest(t *testing.T) *testFixture {
	t.Helper()
	f := new(testFixture)
	require := require.New(t)

	// Base setup
	logger := log.NewTestLogger(t)
	encCfg := moduletestutil.MakeTestEncodingConfig()

	f.govModAddr = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	f.addrs = simtestutil.CreateIncrementalAccounts(3)

	key := storetypes.NewKVStoreKey(poa.ModuleName)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	f.ctx = testCtx.Ctx

	// Register SDK modules.
	registerBaseSDKModules(f, encCfg, storeService, logger, require)

	// Setup POA Keeper.
	f.k = keeper.NewKeeper(encCfg.Codec, storeService, logger, f.govModAddr, f.accountkeeper, f.nftKeeper, f.stakingKeeper)
	f.msgServer = keeper.NewMsgServerImpl(f.k)
	f.queryServer = keeper.NewQuerier(f.k)
	f.appModule = module.NewAppModule(encCfg.Codec, f.k, f.nftKeeper)

	return f
}

func registerModuleInterfaces(encCfg moduletestutil.TestEncodingConfig) {
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	types.RegisterInterfaces(encCfg.InterfaceRegistry)
}

func registerBaseSDKModules(
	f *testFixture,
	encCfg moduletestutil.TestEncodingConfig,
	storeService store.KVStoreService,
	logger log.Logger,
	require *require.Assertions,
) {
	registerModuleInterfaces(encCfg)

	// Auth Keeper.
	f.accountkeeper = authkeeper.NewAccountKeeper(
		encCfg.Codec, storeService,
		authtypes.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec(sdk.Bech32MainPrefix), sdk.Bech32MainPrefix,
		f.govModAddr,
	)

	// Bank Keeper.
	f.bankkeeper = bankkeeper.NewBaseKeeper(
		encCfg.Codec, storeService,
		f.accountkeeper,
		nil,
		f.govModAddr, logger,
	)

	// Staking Keeper.
	f.stakingKeeper = stakingkeeper.NewKeeper(
		encCfg.Codec, storeService,
		f.accountkeeper, f.bankkeeper, f.govModAddr,
		authcodec.NewBech32Codec(sdk.Bech32PrefixValAddr),
		authcodec.NewBech32Codec(sdk.Bech32PrefixConsAddr),
	)
	require.NoError(f.stakingKeeper.SetParams(f.ctx, stakingtypes.DefaultParams()))
	f.accountkeeper.SetModuleAccount(f.ctx, f.stakingKeeper.GetNotBondedPool(f.ctx))
	f.accountkeeper.SetModuleAccount(f.ctx, f.stakingKeeper.GetBondedPool(f.ctx))

	// Mint Keeper.
	f.mintkeeper = mintkeeper.NewKeeper(
		encCfg.Codec, storeService,
		f.stakingKeeper, f.accountkeeper, f.bankkeeper,
		authtypes.FeeCollectorName, f.govModAddr,
	)
	f.accountkeeper.SetModuleAccount(f.ctx, f.accountkeeper.GetModuleAccount(f.ctx, minttypes.ModuleName))
	f.mintkeeper.InitGenesis(f.ctx, f.accountkeeper, minttypes.DefaultGenesisState())
}
