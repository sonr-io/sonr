package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"

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
	module "github.com/sonr-io/sonr/x/dwn"
	"github.com/sonr-io/sonr/x/dwn/keeper"
	"github.com/sonr-io/sonr/x/dwn/types"
	svctypes "github.com/sonr-io/sonr/x/svc/types"
)

// mockServiceKeeper implements types.ServiceKeeper interface for testing
type mockServiceKeeper struct {
	services        map[string]*svctypes.Service
	verifiedDomains map[string]bool
}

func newMockServiceKeeper() *mockServiceKeeper {
	return &mockServiceKeeper{
		services:        make(map[string]*svctypes.Service),
		verifiedDomains: make(map[string]bool),
	}
}

func (m *mockServiceKeeper) VerifyServiceRegistration(
	ctx context.Context,
	serviceID string,
	domain string,
) (bool, error) {
	if serviceID == "" || domain == "" {
		return false, nil
	}

	service, exists := m.services[serviceID]
	if !exists {
		return false, nil
	}

	// Check if service domain matches and domain is verified
	if service.Domain != domain {
		return false, nil
	}

	verified, exists := m.verifiedDomains[domain]
	if !exists {
		return false, nil
	}

	return verified && service.Status == svctypes.ServiceStatus_SERVICE_STATUS_ACTIVE, nil
}

func (m *mockServiceKeeper) GetService(
	ctx context.Context,
	serviceID string,
) (*svctypes.Service, error) {
	service, exists := m.services[serviceID]
	if !exists {
		return nil, svctypes.ErrInvalidServiceID
	}
	return service, nil
}

func (m *mockServiceKeeper) IsDomainVerified(
	ctx context.Context,
	domain string,
	owner string,
) (bool, error) {
	verified, exists := m.verifiedDomains[domain]
	return exists && verified, nil
}

func (m *mockServiceKeeper) GetServicesByDomain(
	ctx context.Context,
	domain string,
) ([]svctypes.Service, error) {
	var services []svctypes.Service
	for _, service := range m.services {
		if service.Domain == domain {
			services = append(services, *service)
		}
	}
	return services, nil
}

// Helper methods for test setup
func (m *mockServiceKeeper) addService(
	serviceID, domain, owner string,
	status svctypes.ServiceStatus,
) {
	m.services[serviceID] = &svctypes.Service{
		Id:     serviceID,
		Domain: domain,
		Owner:  owner,
		Status: status,
	}
}

func (m *mockServiceKeeper) setDomainVerified(domain string, verified bool) {
	m.verifiedDomains[domain] = verified
}

// testFixtureWithService extends the basic test fixture with service keeper
type testFixtureWithService struct {
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
	serviceKeeper  types.ServiceKeeper

	addrs      []sdk.AccAddress
	govModAddr string

	cleanup func()
}

func SetupTestWithServiceKeeper(t *testing.T) *testFixtureWithService {
	t.Helper()
	f := new(testFixtureWithService)

	cfg := sdk.GetConfig()
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
	f.ctx = sdk.NewContext(
		integration.CreateMultiStore(keys, logger),
		cmtproto.Header{},
		false,
		logger,
	)

	// Register SDK modules
	registerBaseSDKModulesForServiceTest(
		logger,
		f,
		encCfg,
		keys,
		accountAddressCodec,
		validatorAddressCodec,
		consensusAddressCodec,
	)

	// Setup Keeper with mock keepers including service keeper
	mockDIDKeeper := &mockDIDKeeper{}
	mockServiceKeeper := newMockServiceKeeper()
	f.serviceKeeper = mockServiceKeeper

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

	t.Cleanup(f.cleanup)

	return f
}

func registerBaseSDKModulesForServiceTest(
	logger log.Logger,
	f *testFixtureWithService,
	encCfg moduletestutil.TestEncodingConfig,
	keys map[string]*storetypes.KVStoreKey,
	ac, vc, cc address.Codec,
) {
	// Account keeper
	f.accountkeeper = authkeeper.NewAccountKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		ac,
		app.Bech32PrefixAccAddr,
		f.govModAddr,
	)

	// Bank keeper
	f.bankkeeper = bankkeeper.NewBaseKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		f.accountkeeper,
		map[string]bool{},
		f.govModAddr,
		logger,
	)

	// Staking keeper
	f.stakingKeeper = stakingkeeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		f.accountkeeper,
		f.bankkeeper,
		f.govModAddr,
		vc,
		cc,
	)

	// Mint keeper
	f.mintkeeper = mintkeeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[minttypes.StoreKey]),
		f.stakingKeeper,
		f.accountkeeper,
		f.bankkeeper,
		authtypes.FeeCollectorName,
		f.govModAddr,
	)

	// Feegrant keeper
	f.feegrantkeeper = feegrantkeeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys["feegrant"]),
		f.accountkeeper,
	)
}

func TestValidateServiceForProtocol(t *testing.T) {
	f := SetupTestWithServiceKeeper(t)

	tests := []struct {
		name          string
		target        string
		serviceID     string
		setupMock     func(*mockServiceKeeper)
		expectedError bool
		errorContains string
	}{
		{
			name:      "empty service ID allows operation",
			target:    "did:web:example.com",
			serviceID: "",
			setupMock: func(mock *mockServiceKeeper) {
				// No setup needed
			},
			expectedError: false,
		},
		{
			name:      "non-DID:web target skips verification",
			target:    "did:key:example",
			serviceID: "test-service",
			setupMock: func(mock *mockServiceKeeper) {
				// No setup needed
			},
			expectedError: false,
		},
		{
			name:      "verified service allows operation",
			target:    "did:web:example.com",
			serviceID: "test-service",
			setupMock: func(mock *mockServiceKeeper) {
				mock.addService(
					"test-service",
					"example.com",
					"owner",
					svctypes.ServiceStatus_SERVICE_STATUS_ACTIVE,
				)
				mock.setDomainVerified("example.com", true)
			},
			expectedError: false,
		},
		{
			name:      "unverified service blocks operation",
			target:    "did:web:example.com",
			serviceID: "test-service",
			setupMock: func(mock *mockServiceKeeper) {
				mock.addService(
					"test-service",
					"example.com",
					"owner",
					svctypes.ServiceStatus_SERVICE_STATUS_ACTIVE,
				)
				mock.setDomainVerified("example.com", false)
			},
			expectedError: true,
			errorContains: "service test-service not verified for domain example.com",
		},
		{
			name:      "non-existent service blocks operation",
			target:    "did:web:example.com",
			serviceID: "non-existent-service",
			setupMock: func(mock *mockServiceKeeper) {
				mock.setDomainVerified("example.com", true)
			},
			expectedError: true,
			errorContains: "service non-existent-service not verified for domain example.com",
		},
		{
			name:      "domain mismatch blocks operation",
			target:    "did:web:example.com",
			serviceID: "test-service",
			setupMock: func(mock *mockServiceKeeper) {
				mock.addService(
					"test-service",
					"different.com",
					"owner",
					svctypes.ServiceStatus_SERVICE_STATUS_ACTIVE,
				)
				mock.setDomainVerified("example.com", true)
				mock.setDomainVerified("different.com", true)
			},
			expectedError: true,
			errorContains: "service test-service not verified for domain example.com",
		},
		{
			name:      "suspended service blocks operation",
			target:    "did:web:example.com",
			serviceID: "test-service",
			setupMock: func(mock *mockServiceKeeper) {
				mock.addService(
					"test-service",
					"example.com",
					"owner",
					svctypes.ServiceStatus_SERVICE_STATUS_SUSPENDED,
				)
				mock.setDomainVerified("example.com", true)
			},
			expectedError: true,
			errorContains: "service test-service not verified for domain example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock for this test
			mock := f.serviceKeeper.(*mockServiceKeeper)
			tt.setupMock(mock)

			// Test the validation
			err := f.k.ValidateServiceForProtocol(f.ctx, tt.target, tt.serviceID)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			// Reset mock for next test
			mock.services = make(map[string]*svctypes.Service)
			mock.verifiedDomains = make(map[string]bool)
		})
	}
}

func TestProtocolsConfigureWithServiceVerification(t *testing.T) {
	f := SetupTestWithServiceKeeper(t)
	mock := f.serviceKeeper.(*mockServiceKeeper)

	// Setup a verified service
	mock.addService(
		"test-service",
		"example.com",
		"owner",
		svctypes.ServiceStatus_SERVICE_STATUS_ACTIVE,
	)
	mock.setDomainVerified("example.com", true)

	tests := []struct {
		name          string
		msg           *types.MsgProtocolsConfigure
		expectedError bool
		errorContains string
	}{
		{
			name: "protocol configure with verified service succeeds",
			msg: &types.MsgProtocolsConfigure{
				Author:        "test-author",
				Target:        "did:web:example.com",
				Authorization: "service:test-service",
				ProtocolUri:   "https://example.com/protocol",
				Definition:    []byte(`{"protocol": "test"}`),
				Published:     true,
			},
			expectedError: false,
		},
		{
			name: "protocol configure with unverified service fails",
			msg: &types.MsgProtocolsConfigure{
				Author:        "test-author",
				Target:        "did:web:example.com",
				Authorization: "service:unverified-service",
				ProtocolUri:   "https://example.com/protocol",
				Definition:    []byte(`{"protocol": "test"}`),
				Published:     true,
			},
			expectedError: true,
			errorContains: "service unverified-service not verified for domain example.com",
		},
		{
			name: "protocol configure without service authorization succeeds",
			msg: &types.MsgProtocolsConfigure{
				Author:        "test-author",
				Target:        "did:web:example.com",
				Authorization: "",
				ProtocolUri:   "https://example.com/protocol",
				Definition:    []byte(`{"protocol": "test"}`),
				Published:     true,
			},
			expectedError: false,
		},
		{
			name: "protocol configure with non-service authorization succeeds",
			msg: &types.MsgProtocolsConfigure{
				Author:        "test-author",
				Target:        "did:web:example.com",
				Authorization: "some-jwt-token",
				ProtocolUri:   "https://example.com/protocol",
				Definition:    []byte(`{"protocol": "test"}`),
				Published:     true,
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.k.ProtocolsConfigure(f.ctx, tt.msg)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRecordsWriteWithServiceVerification(t *testing.T) {
	f := SetupTestWithServiceKeeper(t)
	mock := f.serviceKeeper.(*mockServiceKeeper)

	// Setup a verified service
	mock.addService(
		"test-service",
		"example.com",
		"owner",
		svctypes.ServiceStatus_SERVICE_STATUS_ACTIVE,
	)
	mock.setDomainVerified("example.com", true)

	tests := []struct {
		name          string
		msg           *types.MsgRecordsWrite
		expectedError bool
		errorContains string
	}{
		{
			name: "record write with verified service succeeds",
			msg: &types.MsgRecordsWrite{
				Author:        "test-author",
				Target:        "did:web:example.com",
				Authorization: "service:test-service",
				Data:          []byte("test data"),
				Descriptor_: &types.DWNMessageDescriptor{
					InterfaceName:    "Records",
					Method:           "Write",
					MessageTimestamp: "2023-01-01T00:00:00Z",
					DataCid:          "test-cid",
					DataSize:         9,
					DataFormat:       "text/plain",
				},
			},
			expectedError: false,
		},
		{
			name: "record write with unverified service fails",
			msg: &types.MsgRecordsWrite{
				Author:        "test-author",
				Target:        "did:web:example.com",
				Authorization: "service:unverified-service",
				Data:          []byte("test data"),
				Descriptor_: &types.DWNMessageDescriptor{
					InterfaceName:    "Records",
					Method:           "Write",
					MessageTimestamp: "2023-01-01T00:00:00Z",
					DataCid:          "test-cid",
					DataSize:         9,
					DataFormat:       "text/plain",
				},
			},
			expectedError: true,
			errorContains: "service unverified-service not verified for domain example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.k.RecordsWrite(f.ctx, tt.msg)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
