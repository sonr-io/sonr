package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdkaddress "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"

	"github.com/sonr-io/sonr/app"
	didtypes "github.com/sonr-io/sonr/x/did/types"
	module "github.com/sonr-io/sonr/x/dwn"
	"github.com/sonr-io/sonr/x/dwn/keeper"
	"github.com/sonr-io/sonr/x/dwn/types"
	svctypes "github.com/sonr-io/sonr/x/svc/types"
)

var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:     nil,
	stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	minttypes.ModuleName:           {authtypes.Minter},
	govtypes.ModuleName:            {authtypes.Burner},
}

// mockDIDKeeper implements types.DIDKeeper interface for testing
type mockDIDKeeper struct{}

func (m *mockDIDKeeper) ResolveDID(
	ctx context.Context,
	did string,
) (*didtypes.DIDDocument, *didtypes.DIDDocumentMetadata, error) {
	// Return mock DID document for testing
	return &didtypes.DIDDocument{
			Id: did,
		}, &didtypes.DIDDocumentMetadata{
			Did:     did,
			Created: 1672531200, // 2023-01-01T00:00:00Z as Unix timestamp
			Updated: 1672531200, // 2023-01-01T00:00:00Z as Unix timestamp
		}, nil
}

func (m *mockDIDKeeper) GetDIDDocument(
	ctx context.Context,
	did string,
) (*didtypes.DIDDocument, error) {
	// Return mock DID document for testing
	return &didtypes.DIDDocument{
		Id: did,
	}, nil
}

// mockServiceKeeperForStandardTest implements types.ServiceKeeper interface for standard tests
type mockServiceKeeperForStandardTest struct{}

func (m *mockServiceKeeperForStandardTest) VerifyServiceRegistration(
	ctx context.Context,
	serviceID string,
	domain string,
) (bool, error) {
	// Always return true for standard tests to avoid breaking existing functionality
	return true, nil
}

func (m *mockServiceKeeperForStandardTest) GetService(
	ctx context.Context,
	serviceID string,
) (*svctypes.Service, error) {
	// Return a basic service for testing
	return &svctypes.Service{
		Id:     serviceID,
		Domain: "test.com",
		Owner:  "test-owner",
		Status: svctypes.ServiceStatus_SERVICE_STATUS_ACTIVE,
	}, nil
}

func (m *mockServiceKeeperForStandardTest) IsDomainVerified(
	ctx context.Context,
	domain string,
	owner string,
) (bool, error) {
	// Always return true for standard tests
	return true, nil
}

func (m *mockServiceKeeperForStandardTest) GetServicesByDomain(
	ctx context.Context,
	domain string,
) ([]svctypes.Service, error) {
	// Return empty list for standard tests
	return []svctypes.Service{}, nil
}

type testFixture struct {
	suite.Suite

	ctx         sdk.Context
	k           keeper.Keeper
	msgServer   types.MsgServer
	queryServer types.QueryServer
	appModule   *module.AppModule

	accountkeeper  authkeeper.AccountKeeper
	bankkeeper     bankkeeper.BaseKeeper
	stakingKeeper  *stakingkeeper.Keeper
	mintkeeper     mintkeeper.Keeper
	feegrantkeeper feegrantkeeper.Keeper

	addrs      []sdk.AccAddress
	govModAddr string

	// Add cleanup function
	cleanup func()
}

func SetupTest(t *testing.T) *testFixture {
	t.Helper()
	f := new(testFixture)

	cfg := sdk.GetConfig() // do not seal, more set later
	cfg.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	cfg.SetCoinType(app.CoinType)

	validatorAddressCodec := sdkaddress.NewBech32Codec(app.Bech32PrefixValAddr)
	accountAddressCodec := sdkaddress.NewBech32Codec(app.Bech32PrefixAccAddr)
	consensusAddressCodec := sdkaddress.NewBech32Codec(app.Bech32PrefixConsAddr)

	// Base setup
	logger := log.NewTestLogger(t)
	encCfg := moduletestutil.MakeTestEncodingConfig()

	f.govModAddr = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	f.addrs = simtestutil.CreateIncrementalAccounts(3)

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		minttypes.ModuleName,
		"feegrant",
		types.ModuleName,
	)
	// Set a proper block time for fee grant expiration validation
	header := cmtproto.Header{
		Time: time.Now(),
	}
	f.ctx = sdk.NewContext(integration.CreateMultiStore(keys, logger), header, false, logger)

	// Register SDK modules.
	registerBaseSDKModules(
		logger,
		f,
		encCfg,
		keys,
		accountAddressCodec,
		validatorAddressCodec,
		consensusAddressCodec,
	)

	// Setup Keeper with mock DID, UCAN, and Service keepers.
	mockDIDKeeper := &mockDIDKeeper{}
	mockServiceKeeper := &mockServiceKeeperForStandardTest{}

	// Create client context for transaction building
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithCodec(encCfg.Codec).WithTxConfig(encCfg.TxConfig)

	f.k = keeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[types.ModuleName]),
		logger,
		f.govModAddr,
		f.accountkeeper,
		f.bankkeeper,
		f.feegrantkeeper,
		f.stakingKeeper,
		mockDIDKeeper,
		mockServiceKeeper,
		clientCtx,
	)
	f.msgServer = keeper.NewMsgServerImpl(f.k)
	f.queryServer = keeper.NewQuerier(f.k)
	f.appModule = module.NewAppModule(encCfg.Codec, f.k)

	// Initialize with default genesis
	genesisState := &types.GenesisState{
		Params: types.DefaultParams(),
	}
	f.k.InitGenesis(f.ctx, genesisState)

	// Set up cleanup function (no-op for now, can be extended if needed)
	f.cleanup = func() {
		// Currently no cleanup needed, but placeholder for future use
	}

	// Register cleanup to run when test finishes
	t.Cleanup(f.cleanup)

	return f
}

func registerModuleInterfaces(encCfg moduletestutil.TestEncodingConfig) {
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	banktypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	minttypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	types.RegisterInterfaces(encCfg.InterfaceRegistry)
}

func registerBaseSDKModules(
	logger log.Logger,
	f *testFixture,
	encCfg moduletestutil.TestEncodingConfig,
	keys map[string]*storetypes.KVStoreKey,
	ac address.Codec,
	validator address.Codec,
	consensus address.Codec,
) {
	registerModuleInterfaces(encCfg)

	// Auth Keeper.
	f.accountkeeper = authkeeper.NewAccountKeeper(
		encCfg.Codec, runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		ac, app.Bech32PrefixAccAddr,
		f.govModAddr,
	)

	// Bank Keeper.
	f.bankkeeper = bankkeeper.NewBaseKeeper(
		encCfg.Codec, runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		f.accountkeeper,
		nil,
		f.govModAddr, logger,
	)

	// Staking Keeper.
	f.stakingKeeper = stakingkeeper.NewKeeper(
		encCfg.Codec, runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		f.accountkeeper, f.bankkeeper, f.govModAddr,
		validator,
		consensus,
	)

	// Mint Keeper.
	f.mintkeeper = mintkeeper.NewKeeper(
		encCfg.Codec, runtime.NewKVStoreService(keys[minttypes.StoreKey]),
		f.stakingKeeper, f.accountkeeper, f.bankkeeper,
		authtypes.FeeCollectorName, f.govModAddr,
	)

	// Feegrant Keeper.
	f.feegrantkeeper = feegrantkeeper.NewKeeper(
		encCfg.Codec, runtime.NewKVStoreService(keys["feegrant"]),
		f.accountkeeper,
	)
	// Set the bank keeper using the SetBankKeeper method
	f.feegrantkeeper = f.feegrantkeeper.SetBankKeeper(f.bankkeeper)
}
