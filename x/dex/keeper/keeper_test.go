package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
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

	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	portkeeper "github.com/cosmos/ibc-go/v8/modules/core/05-port/keeper"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/sonr-io/sonr/app"
	"github.com/sonr-io/sonr/x/dex/keeper"
	"github.com/sonr-io/sonr/x/dex/types"
	didtypes "github.com/sonr-io/sonr/x/did/types"
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

	accountkeeper authkeeper.AccountKeeper
	bankkeeper    bankkeeper.BaseKeeper
	stakingKeeper *stakingkeeper.Keeper
	mintkeeper    mintkeeper.Keeper

	addrs      []sdk.AccAddress
	govModAddr string
}

// SetupTest creates a new test fixture
func SetupTest(t *testing.T) *testFixture {
	t.Helper()
	f := new(testFixture)

	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	cfg.SetCoinType(app.CoinType)

	validatorAddressCodec := sdkaddress.NewBech32Codec(app.Bech32PrefixValAddr)
	consensusAddressCodec := sdkaddress.NewBech32Codec(app.Bech32PrefixConsAddr)

	// Base setup
	logger := log.NewTestLogger(t)
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// Register auth types interfaces
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	banktypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	minttypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	f.govModAddr = authtypes.NewModuleAddress(govtypes.ModuleName).String()

	// Initialize test addresses
	f.addrs = simtestutil.CreateIncrementalAccounts(3)

	// Setup store keys
	keys := storetypes.NewKVStoreKeys(
		types.StoreKey, authtypes.StoreKey, banktypes.StoreKey,
		stakingtypes.StoreKey, minttypes.StoreKey, capabilitytypes.StoreKey,
	)
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	cdc := encCfg.Codec

	// Initialize keepers
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	maccPerms[types.ModuleName] = nil
	f.accountkeeper = authkeeper.NewAccountKeeper(
		cdc, runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount, maccPerms,
		sdkaddress.NewBech32Codec(app.Bech32PrefixAccAddr),
		app.Bech32PrefixAccAddr, authority.String(),
	)

	f.bankkeeper = bankkeeper.NewBaseKeeper(
		cdc, runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		f.accountkeeper, nil, authority.String(), logger,
	)

	f.stakingKeeper = stakingkeeper.NewKeeper(
		cdc, runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		f.accountkeeper, f.bankkeeper, authority.String(),
		validatorAddressCodec, consensusAddressCodec,
	)

	f.mintkeeper = mintkeeper.NewKeeper(
		cdc, runtime.NewKVStoreService(keys[minttypes.StoreKey]),
		f.stakingKeeper, f.accountkeeper, f.bankkeeper,
		authtypes.FeeCollectorName, authority.String(),
	)

	// Create capability keeper for IBC
	capabilityKeeper := capabilitykeeper.NewKeeper(
		cdc,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)

	// Create scoped keeper for the DEX module
	scopedKeeper := capabilityKeeper.ScopeToModule(types.ModuleName)

	// Create port keeper
	portKeeper := portkeeper.NewKeeper(scopedKeeper)

	// Create mock expected keepers
	mockICS4Wrapper := &mockICS4Wrapper{}
	mockAccountKeeper := &mockAccountKeeper{}
	mockBankKeeper := &mockBankKeeper{}
	mockICAControllerKeeper := &mockICAControllerKeeper{}
	mockConnectionKeeper := &mockConnectionKeeper{}
	mockChannelKeeper := &mockChannelKeeper{}
	mockDIDKeeper := &mockDIDKeeper{}
	mockDWNKeeper := &mockDWNKeeper{}

	// Initialize DEX keeper
	f.k = keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[types.StoreKey]),
		mockICS4Wrapper,
		&portKeeper,
		scopedKeeper,
		mockAccountKeeper,
		mockBankKeeper,
		mockICAControllerKeeper,
		mockConnectionKeeper,
		mockChannelKeeper,
		mockDIDKeeper,
		mockDWNKeeper,
		authority.String(),
	)

	f.msgServer = keeper.NewMsgServerImpl(f.k)
	f.queryServer = keeper.NewQueryServerImpl(f.k)

	// Initialize context with proper multistore
	cms := integration.CreateMultiStore(keys, logger)
	for _, key := range memKeys {
		cms.MountStoreWithDB(key, storetypes.StoreTypeMemory, nil)
	}

	f.ctx = sdk.NewContext(cms, cmtproto.Header{
		Height: 1,
		Time:   time.Now(),
	}, false, logger)

	// Fund test accounts
	initCoins := sdk.NewCoins(sdk.NewCoin("usnr", math.NewInt(1000000000)))
	for _, addr := range f.addrs {
		err := f.bankkeeper.MintCoins(f.ctx, minttypes.ModuleName, initCoins)
		if err != nil {
			panic(err)
		}
		err = f.bankkeeper.SendCoinsFromModuleToAccount(
			f.ctx,
			minttypes.ModuleName,
			addr,
			initCoins,
		)
		if err != nil {
			panic(err)
		}
	}

	return f
}

// KeeperTestSuite runs all keeper tests
type KeeperTestSuite struct {
	suite.Suite
	f *testFixture
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.f = SetupTest(suite.T())
}

// Test basic keeper operations
func (suite *KeeperTestSuite) TestRegisterDEXAccount() {
	did := "did:sonr:test123"
	connectionID := "connection-0"

	// Register a new DEX account through keeper method
	account, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		did,
		connectionID,
		[]string{"swap", "liquidity"},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(account)

	// Retrieve the account
	retrieved, err := suite.f.k.GetDEXAccount(suite.f.ctx, did, connectionID)
	suite.Require().NoError(err)
	suite.Require().NotNil(retrieved)
	suite.Require().Equal(did, retrieved.Did)
	suite.Require().Equal(connectionID, retrieved.ConnectionId)
	suite.Require().Equal(types.ACCOUNT_STATUS_PENDING, retrieved.Status)
}

func (suite *KeeperTestSuite) TestGetDEXAccountsByDID() {
	did := "did:sonr:test456"

	// Register multiple accounts for the same DID
	connections := []string{"connection-0", "connection-1"}
	for _, connID := range connections {
		_, err := suite.f.k.RegisterDEXAccount(
			suite.f.ctx,
			did,
			connID,
			[]string{"swap"},
		)
		suite.Require().NoError(err)
	}

	// Retrieve all accounts for the DID
	accounts, err := suite.f.k.GetDEXAccountsByDID(suite.f.ctx, did)
	suite.Require().NoError(err)
	suite.Require().Len(accounts, 2)
}

func (suite *KeeperTestSuite) TestParamsOperations() {
	// Set params
	params := types.Params{
		Enabled:               true,
		MaxAccountsPerDid:     5,
		DefaultTimeoutSeconds: 600,
		AllowedConnections:    []string{"connection-0", "connection-1"},
		MinSwapAmount:         "100",
		MaxDailyVolume:        "1000000",
		RateLimits: types.RateLimitParams{
			MaxOpsPerBlock:     10,
			MaxOpsPerDidPerDay: 100,
			CooldownBlocks:     5,
		},
		Fees: types.FeeParams{
			SwapFeeBps:      30, // 0.3%
			LiquidityFeeBps: 10, // 0.1%
			OrderFeeBps:     20, // 0.2%
			FeeCollector:    "sonr1feecolllector",
		},
	}

	err := suite.f.k.Params.Set(suite.f.ctx, params)
	suite.Require().NoError(err)

	// Get params
	retrieved, err := suite.f.k.Params.Get(suite.f.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(params.Enabled, retrieved.Enabled)
	suite.Require().Equal(params.MaxAccountsPerDid, retrieved.MaxAccountsPerDid)
	suite.Require().Equal(params.AllowedConnections, retrieved.AllowedConnections)
}

// Mock implementations for expected keepers
type mockICS4Wrapper struct{}

func (m *mockICS4Wrapper) SendPacket(
	ctx sdk.Context,
	channelCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (uint64, error) {
	return 1, nil
}

func (m *mockICS4Wrapper) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet ibcexported.PacketI,
	acknowledgement ibcexported.Acknowledgement,
) error {
	return nil
}

func (m *mockICS4Wrapper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return "ics27-1", true
}

type mockAccountKeeper struct{}

func (m *mockAccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}

func (m *mockAccountKeeper) SetAccount(ctx context.Context, acc sdk.AccountI) {}

func (m *mockAccountKeeper) NewAccountWithAddress(
	ctx sdk.Context,
	addr sdk.AccAddress,
) sdk.AccountI {
	return nil
}

func (m *mockAccountKeeper) GetModuleAccount(
	ctx context.Context,
	moduleName string,
) sdk.ModuleAccountI {
	return nil
}

func (m *mockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return sdk.AccAddress{}
}

type mockBankKeeper struct{}

func (m *mockBankKeeper) SendCoins(
	ctx context.Context,
	fromAddr, toAddr sdk.AccAddress,
	amt sdk.Coins,
) error {
	return nil
}

func (m *mockBankKeeper) SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins()
}

type mockICAControllerKeeper struct{}

func (m *mockICAControllerKeeper) RegisterInterchainAccount(
	ctx sdk.Context,
	connectionID, owner, version string,
) error {
	return nil
}

func (m *mockICAControllerKeeper) GetInterchainAccountAddress(
	ctx sdk.Context,
	connectionID, portID string,
) (string, bool) {
	return "cosmos1test", true
}

func (m *mockICAControllerKeeper) SendTx(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	connectionID, portID string,
	icaPacketData icatypes.InterchainAccountPacketData,
	timeoutTimestamp uint64,
) (uint64, error) {
	return 1, nil
}

func (m *mockICAControllerKeeper) GetActiveChannelID(
	ctx sdk.Context,
	connectionID, portID string,
) (string, bool) {
	return "channel-0", true
}

type mockConnectionKeeper struct{}

func (m *mockConnectionKeeper) GetConnection(
	ctx sdk.Context,
	connectionID string,
) (connectiontypes.ConnectionEnd, bool) {
	return connectiontypes.ConnectionEnd{
		ClientId: "07-tendermint-0",
		Versions: []*connectiontypes.Version{{
			Identifier: "1",
			Features:   []string{"ORDER_ORDERED", "ORDER_UNORDERED"},
		}},
		State: connectiontypes.OPEN,
		Counterparty: connectiontypes.Counterparty{
			ClientId:     "07-tendermint-0",
			ConnectionId: "connection-0",
		},
	}, true
}

type mockChannelKeeper struct{}

func (m *mockChannelKeeper) GetChannel(
	ctx sdk.Context,
	portID, channelID string,
) (channeltypes.Channel, bool) {
	return channeltypes.Channel{
		State:    channeltypes.OPEN,
		Ordering: channeltypes.ORDERED,
		Counterparty: channeltypes.Counterparty{
			PortId:    "icahost",
			ChannelId: "channel-0",
		},
		ConnectionHops: []string{"connection-0"},
		Version:        "ics27-1",
	}, true
}

func (m *mockChannelKeeper) GetNextSequenceSend(
	ctx sdk.Context,
	portID, channelID string,
) (uint64, bool) {
	return 1, true
}

func (m *mockChannelKeeper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (uint64, error) {
	return 1, nil
}

type mockDIDKeeper struct{}

func (m *mockDIDKeeper) GetDIDDocument(
	ctx context.Context,
	did string,
) (*didtypes.DIDDocument, error) {
	return &didtypes.DIDDocument{
		Id: did,
	}, nil
}

type mockDWNKeeper struct{}
