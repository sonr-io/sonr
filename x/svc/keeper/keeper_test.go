package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
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

	"github.com/sonr-io/sonr/app"
	didtypes "github.com/sonr-io/sonr/x/did/types"
	module "github.com/sonr-io/sonr/x/svc"
	"github.com/sonr-io/sonr/x/svc/keeper"
	"github.com/sonr-io/sonr/x/svc/types"
)

var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:     nil,
	stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	minttypes.ModuleName:           {authtypes.Minter},
	govtypes.ModuleName:            {authtypes.Burner},
}

// SVCMockDIDKeeper provides a minimal mock implementation for SVC testing
type SVCMockDIDKeeper struct{}

func (m *SVCMockDIDKeeper) ResolveDID(
	ctx context.Context,
	did string,
) (*didtypes.DIDDocument, *didtypes.DIDDocumentMetadata, error) {
	if did == "did:example:deactivated" {
		return &didtypes.DIDDocument{
				Id:          did,
				Deactivated: true,
			}, &didtypes.DIDDocumentMetadata{
				Did: did,
			}, nil
	}
	if did == "did:example:no-verification" {
		return &didtypes.DIDDocument{
				Id:                 did,
				VerificationMethod: []*didtypes.VerificationMethod{},
			}, &didtypes.DIDDocumentMetadata{
				Did: did,
			}, nil
	}
	return &didtypes.DIDDocument{
		Id:                 did,
		VerificationMethod: []*didtypes.VerificationMethod{{Id: did + "#key-1"}},
	}, &didtypes.DIDDocumentMetadata{Did: did}, nil
}

func (m *SVCMockDIDKeeper) GetDIDDocument(
	ctx context.Context,
	did string,
) (*didtypes.DIDDocument, error) {
	if did == "did:example:deactivated" {
		return &didtypes.DIDDocument{Id: did, Deactivated: true}, nil
	}
	if did == "did:example:no-verification" {
		return &didtypes.DIDDocument{
			Id:                 did,
			VerificationMethod: []*didtypes.VerificationMethod{},
		}, nil
	}
	return &didtypes.DIDDocument{
		Id:                 did,
		VerificationMethod: []*didtypes.VerificationMethod{{Id: did + "#key-1"}},
	}, nil
}

func (m *SVCMockDIDKeeper) VerifyDIDDocumentSignature(
	ctx context.Context,
	did string,
	signature []byte,
) (bool, error) {
	return true, nil
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
	stakingKeeper *stakingkeeper.Keeper
	mintkeeper    mintkeeper.Keeper

	addrs      []sdk.AccAddress
	govModAddr string
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
		authtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		minttypes.ModuleName,
		types.ModuleName,
	)
	f.ctx = sdk.NewContext(
		integration.CreateMultiStore(keys, logger),
		cmtproto.Header{},
		false,
		logger,
	)

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

	// Setup SVC Keeper with DID dependency only (UCAN is now internal).
	mockDIDKeeper := &SVCMockDIDKeeper{}
	f.k = keeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[types.ModuleName]),
		logger,
		f.govModAddr,
		mockDIDKeeper,
	)
	f.msgServer = keeper.NewMsgServerImpl(f.k)
	f.queryServer = keeper.NewQuerier(f.k)
	f.appModule = module.NewAppModule(encCfg.Codec, f.k)

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
}

// Test for VerifyServiceRegistration method
func TestVerifyServiceRegistration(t *testing.T) {
	f := SetupTest(t)

	// Test case 1: Invalid service ID
	valid, err := f.k.VerifyServiceRegistration(f.ctx, "", "example.com")
	if err != types.ErrInvalidServiceID {
		t.Errorf("Expected ErrInvalidServiceID, got %v", err)
	}
	if valid {
		t.Error("Expected invalid service registration")
	}

	// Test case 2: Invalid domain
	valid, err = f.k.VerifyServiceRegistration(f.ctx, "test-service", "")
	if err != types.ErrDomainNotVerified {
		t.Errorf("Expected ErrDomainNotVerified, got %v", err)
	}
	if valid {
		t.Error("Expected invalid service registration")
	}

	// Test case 3: Non-existent service
	valid, err = f.k.VerifyServiceRegistration(f.ctx, "non-existent-service", "example.com")
	if err != types.ErrInvalidServiceID {
		t.Errorf("Expected ErrInvalidServiceID, got %v", err)
	}
	if valid {
		t.Error("Expected invalid service registration")
	}
}

// Test for GetService method
func TestGetService(t *testing.T) {
	f := SetupTest(t)

	// Test case 1: Invalid service ID
	service, err := f.k.GetService(f.ctx, "")
	if err != types.ErrInvalidServiceID {
		t.Errorf("Expected ErrInvalidServiceID, got %v", err)
	}
	if service != nil {
		t.Error("Expected nil service")
	}

	// Test case 2: Non-existent service
	service, err = f.k.GetService(f.ctx, "non-existent-service")
	if err != types.ErrInvalidServiceID {
		t.Errorf("Expected ErrInvalidServiceID, got %v", err)
	}
	if service != nil {
		t.Error("Expected nil service")
	}
}

// Test for IsDomainVerified method
func TestIsDomainVerified(t *testing.T) {
	f := SetupTest(t)

	// Test case 1: Empty domain
	verified, err := f.k.IsDomainVerified(f.ctx, "", "owner")
	if err != types.ErrDomainNotVerified {
		t.Errorf("Expected ErrDomainNotVerified, got %v", err)
	}
	if verified {
		t.Error("Expected domain not verified")
	}

	// Test case 2: Non-existent domain
	verified, err = f.k.IsDomainVerified(f.ctx, "non-existent.com", "owner")
	if err != types.ErrDomainNotVerified {
		t.Errorf("Expected ErrDomainNotVerified, got %v", err)
	}
	if verified {
		t.Error("Expected domain not verified")
	}
}

// Test for GetServicesByDomain method
func TestGetServicesByDomain(t *testing.T) {
	f := SetupTest(t)

	// Test case 1: Empty domain
	services, err := f.k.GetServicesByDomain(f.ctx, "")
	if err != types.ErrDomainNotVerified {
		t.Errorf("Expected ErrDomainNotVerified, got %v", err)
	}
	if services != nil {
		t.Error("Expected nil services")
	}

	// Test case 2: Non-existent domain
	services, err = f.k.GetServicesByDomain(f.ctx, "non-existent.com")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(services) != 0 {
		t.Error("Expected empty services slice")
	}
}

// Test for ValidateServiceOwnerDID method
func TestValidateServiceOwnerDID(t *testing.T) {
	f := SetupTest(t)

	// Test case 1: Empty DID
	err := f.k.ValidateServiceOwnerDID(f.ctx, "")
	if err != types.ErrInvalidOwnerDID {
		t.Errorf("Expected ErrInvalidOwnerDID, got %v", err)
	}

	// Test case 2: Valid DID (mocked)
	err = f.k.ValidateServiceOwnerDID(f.ctx, "did:example:123")
	if err != nil {
		t.Errorf("Expected no error for valid DID, got %v", err)
	}

	// Test case 3: Deactivated DID
	err = f.k.ValidateServiceOwnerDID(f.ctx, "did:example:deactivated")
	if err != types.ErrInvalidOwnerDID {
		t.Errorf("Expected ErrInvalidOwnerDID for deactivated DID, got %v", err)
	}

	// Test case 4: DID without verification methods
	err = f.k.ValidateServiceOwnerDID(f.ctx, "did:example:no-verification")
	if err != types.ErrInvalidOwnerDID {
		t.Errorf("Expected ErrInvalidOwnerDID for DID without verification methods, got %v", err)
	}
}

// Test for internal UCAN integration basic functionality
func TestInternalUCANIntegrationBasic(t *testing.T) {
	f := SetupTest(t)

	// Test that keeper was created successfully with internal UCAN integration
	if f.k.Logger() == nil {
		t.Error("Expected keeper to be properly initialized")
	}

	// Test ValidateServicePermissions method (which uses internal validation)
	permissions := []string{"register", "update"}
	err := f.k.ValidateServicePermissions(f.ctx, permissions)
	if err != nil {
		t.Errorf("ValidateServicePermissions failed: %v", err)
	}

	// Test UCAN delegation chain validation (which uses internal UCAN library)
	// Note: This will fail with properly formatted error since we don't have valid UCAN tokens
	invalidChain := "invalid_token"
	err = f.k.ValidateUCANDelegationChain(f.ctx, invalidChain)
	if err == nil {
		t.Error("Expected ValidateUCANDelegationChain to fail with invalid token")
	}
}
