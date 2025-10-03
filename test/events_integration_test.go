package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

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
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"

	"github.com/sonr-io/sonr/app"
	didkeeper "github.com/sonr-io/sonr/x/did/keeper"
	didtypes "github.com/sonr-io/sonr/x/did/types"
	dwnkeeper "github.com/sonr-io/sonr/x/dwn/keeper"
	dwntypes "github.com/sonr-io/sonr/x/dwn/types"
	svckeeper "github.com/sonr-io/sonr/x/svc/keeper"
	svctypes "github.com/sonr-io/sonr/x/svc/types"
)

// EventIntegrationTestSuite tests event emission across modules
type EventIntegrationTestSuite struct {
	suite.Suite

	ctx sdk.Context

	// Keepers
	didKeeper didkeeper.Keeper
	dwnKeeper dwnkeeper.Keeper
	svcKeeper svckeeper.Keeper

	// Message servers
	didMsgServer didtypes.MsgServer
	dwnMsgServer dwntypes.MsgServer
	svcMsgServer svctypes.MsgServer

	// Test addresses
	addrs []sdk.AccAddress
}

func TestEventIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(EventIntegrationTestSuite))
}

func (suite *EventIntegrationTestSuite) SetupTest() {
	// Initialize SDK config
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	cfg.SetCoinType(app.CoinType)

	// Create test addresses
	suite.addrs = simtestutil.CreateIncrementalAccounts(3)

	// Setup logger and encoding config
	logger := log.NewTestLogger(suite.T())
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// Register module interfaces
	didtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	dwntypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	svctypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	banktypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	// Create store keys
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		didtypes.StoreKey,
		dwntypes.StoreKey,
		svctypes.StoreKey,
	)

	// Create context with event manager
	header := cmtproto.Header{
		Time: time.Now(),
	}
	suite.ctx = sdk.NewContext(
		integration.CreateMultiStore(keys, logger),
		header,
		false,
		logger,
	).WithEventManager(sdk.NewEventManager())

	// Setup base SDK keepers
	govModAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	accountAddressCodec := sdkaddress.NewBech32Codec(app.Bech32PrefixAccAddr)
	validatorAddressCodec := sdkaddress.NewBech32Codec(app.Bech32PrefixValAddr)
	consensusAddressCodec := sdkaddress.NewBech32Codec(app.Bech32PrefixConsAddr)

	// Account keeper
	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:     nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}
	accountKeeper := authkeeper.NewAccountKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		accountAddressCodec,
		app.Bech32PrefixAccAddr,
		govModAddr,
	)

	// Bank keeper
	bankKeeper := bankkeeper.NewBaseKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		map[string]bool{},
		govModAddr,
		logger,
	)

	// Staking keeper (minimal setup)
	stakingKeeper := stakingkeeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		govModAddr,
		validatorAddressCodec,
		consensusAddressCodec,
	)

	// Feegrant keeper
	feegrantKeeper := feegrantkeeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		accountKeeper,
	)

	// Client context
	clientCtx := client.Context{}.
		WithCodec(encCfg.Codec).
		WithTxConfig(encCfg.TxConfig)

	// Setup DID keeper
	suite.didKeeper = didkeeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[didtypes.StoreKey]),
		logger,
		govModAddr,
		accountKeeper,
	)
	suite.didMsgServer = didkeeper.NewMsgServerImpl(suite.didKeeper)

	// Setup Service keeper
	suite.svcKeeper = svckeeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[svctypes.StoreKey]),
		logger,
		govModAddr,
		&suite.didKeeper,
	)
	suite.svcMsgServer = svckeeper.NewMsgServerImpl(suite.svcKeeper)

	// Setup DWN keeper with mocks for service
	suite.dwnKeeper = dwnkeeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[dwntypes.StoreKey]),
		logger,
		govModAddr,
		accountKeeper,
		bankKeeper,
		feegrantKeeper,
		stakingKeeper,
		&suite.didKeeper,
		&mockServiceKeeper{},
		clientCtx,
	)
	suite.dwnMsgServer = dwnkeeper.NewMsgServerImpl(suite.dwnKeeper)

	// Initialize genesis for each module
	var err error
	didGenesis := &didtypes.GenesisState{
		Params: didtypes.DefaultParams(),
	}
	err = suite.didKeeper.InitGenesis(suite.ctx, didGenesis)
	suite.Require().NoError(err)

	dwnGenesis := &dwntypes.GenesisState{
		Params: dwntypes.DefaultParams(),
	}
	err = suite.dwnKeeper.InitGenesis(suite.ctx, dwnGenesis)
	suite.Require().NoError(err)

	svcGenesis := &svctypes.GenesisState{
		Params: svctypes.DefaultParams(),
	}
	err = suite.svcKeeper.InitGenesis(suite.ctx, svcGenesis)
	suite.Require().NoError(err)
}

const (
	eventTypeDIDCreated = "did.v1.EventDIDCreated"
)

// TestDIDModuleEventEmission tests DID module event emissions
func (suite *EventIntegrationTestSuite) TestDIDModuleEventEmission() {
	did := "did:sonr:test123"
	controller := suite.addrs[0].String()

	// Clear any previous events
	suite.ctx = suite.ctx.WithEventManager(sdk.NewEventManager())

	// Create DID
	createMsg := &didtypes.MsgCreateDID{
		Controller: controller,
		DidDocument: didtypes.DIDDocument{
			Id:                did,
			PrimaryController: controller,
		},
	}

	_, err := suite.didMsgServer.CreateDID(suite.ctx, createMsg)
	suite.Require().NoError(err)

	// Check for EventDIDCreated
	events := suite.ctx.EventManager().Events()
	suite.Require().NotEmpty(events, "Expected events to be emitted")

	foundEvent := false
	for _, event := range events {
		if event.Type == eventTypeDIDCreated {
			foundEvent = true
			break
		}
	}
	suite.Require().True(foundEvent, "EventDIDCreated not found")
}

// TestDWNModuleEventEmission tests DWN module event emissions
func (suite *EventIntegrationTestSuite) TestDWNModuleEventEmission() {
	target := "did:sonr:dwn123"

	// Clear any previous events
	suite.ctx = suite.ctx.WithEventManager(sdk.NewEventManager())

	// Write a record
	writeMsg := &dwntypes.MsgRecordsWrite{
		Target: target,
		Author: suite.addrs[0].String(),
		Descriptor_: &dwntypes.DWNMessageDescriptor{
			InterfaceName:    "Records",
			Method:           "Write",
			MessageTimestamp: time.Now().Format(time.RFC3339),
			DataFormat:       "application/json",
		},
		Data:     []byte(`{"test": "data"}`),
		Protocol: "test-protocol",
		Schema:   "test-schema",
	}

	resp, err := suite.dwnMsgServer.RecordsWrite(suite.ctx, writeMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)

	// Check for EventRecordWritten
	events := suite.ctx.EventManager().Events()
	suite.Require().NotEmpty(events, "Expected events to be emitted")

	foundEvent := false
	for _, event := range events {
		if event.Type == "dwn.v1.EventRecordWritten" {
			foundEvent = true
			break
		}
	}
	suite.Require().True(foundEvent, "EventRecordWritten not found")
}

// TestServiceModuleEventEmission tests Service module event emissions
func (suite *EventIntegrationTestSuite) TestServiceModuleEventEmission() {
	domain := "test.example.com"
	creator := suite.addrs[0].String()

	// Clear any previous events
	suite.ctx = suite.ctx.WithEventManager(sdk.NewEventManager())

	// Initiate domain verification
	initMsg := &svctypes.MsgInitiateDomainVerification{
		Domain:  domain,
		Creator: creator,
	}

	resp, err := suite.svcMsgServer.InitiateDomainVerification(suite.ctx, initMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)

	// Check for EventDomainVerificationInitiated
	events := suite.ctx.EventManager().Events()
	suite.Require().NotEmpty(events, "Expected events to be emitted")

	foundEvent := false
	for _, event := range events {
		if event.Type == "svc.v1.EventDomainVerificationInitiated" {
			foundEvent = true
			break
		}
	}
	suite.Require().True(foundEvent, "EventDomainVerificationInitiated not found")
}

// TestCrossModuleEventSequence tests events from multiple modules in sequence
func (suite *EventIntegrationTestSuite) TestCrossModuleEventSequence() {
	controller := suite.addrs[0].String()
	did := "did:sonr:crosstest"

	// Clear events
	suite.ctx = suite.ctx.WithEventManager(sdk.NewEventManager())

	// 1. Create DID
	createDIDMsg := &didtypes.MsgCreateDID{
		Controller: controller,
		DidDocument: didtypes.DIDDocument{
			Id:                did,
			PrimaryController: controller,
		},
	}
	_, err := suite.didMsgServer.CreateDID(suite.ctx, createDIDMsg)
	suite.Require().NoError(err)

	// 2. Write DWN record
	writeRecordMsg := &dwntypes.MsgRecordsWrite{
		Target: did,
		Author: controller,
		Descriptor_: &dwntypes.DWNMessageDescriptor{
			InterfaceName:    "Records",
			Method:           "Write",
			MessageTimestamp: time.Now().Format(time.RFC3339),
			DataFormat:       "application/json",
		},
		Data:     []byte(`{"crossModule": true}`),
		Protocol: "test-protocol",
		Schema:   "test-schema",
	}
	recordResp, err := suite.dwnMsgServer.RecordsWrite(suite.ctx, writeRecordMsg)
	suite.Require().NoError(err)
	suite.Require().NotNil(recordResp)

	// 3. Initiate domain verification
	initDomainMsg := &svctypes.MsgInitiateDomainVerification{
		Domain:  "cross.test.com",
		Creator: controller,
	}
	_, err = suite.svcMsgServer.InitiateDomainVerification(suite.ctx, initDomainMsg)
	suite.Require().NoError(err)

	// Check that we have events from all modules
	events := suite.ctx.EventManager().Events()
	suite.Require().NotEmpty(events, "Expected events to be emitted")

	// Verify we have events from each module
	hasDIDEvent := false
	hasDWNEvent := false
	hasSVCEvent := false

	for _, event := range events {
		switch event.Type {
		case "did.v1.EventDIDCreated":
			hasDIDEvent = true
		case "dwn.v1.EventRecordWritten":
			hasDWNEvent = true
		case "svc.v1.EventDomainVerificationInitiated":
			hasSVCEvent = true
		}
	}

	suite.Require().True(hasDIDEvent, "DID event not found in cross-module sequence")
	suite.Require().True(hasDWNEvent, "DWN event not found in cross-module sequence")
	suite.Require().True(hasSVCEvent, "Service event not found in cross-module sequence")
}

// TestEventAttributeFiltering tests filtering events by attributes
func (suite *EventIntegrationTestSuite) TestEventAttributeFiltering() {
	controller1 := suite.addrs[0].String()
	controller2 := suite.addrs[1].String()

	// Clear events
	suite.ctx = suite.ctx.WithEventManager(sdk.NewEventManager())

	// Create multiple DIDs with different controllers
	for i, controller := range []string{controller1, controller2, controller1} {
		did := fmt.Sprintf("did:sonr:filter%d", i)
		msg := &didtypes.MsgCreateDID{
			Controller: controller,
			DidDocument: didtypes.DIDDocument{
				Id:                did,
				PrimaryController: controller,
			},
		}
		_, err := suite.didMsgServer.CreateDID(suite.ctx, msg)
		suite.Require().NoError(err)
	}

	// Get all events
	events := suite.ctx.EventManager().Events()

	// Filter events by controller1
	controller1Events := 0
	controller2Events := 0
	for _, event := range events {
		if event.Type == eventTypeDIDCreated {
			for _, attr := range event.Attributes {
				if attr.Key == "creator" {
					if attr.Value == fmt.Sprintf("\"%s\"", controller1) {
						controller1Events++
					} else if attr.Value == fmt.Sprintf("\"%s\"", controller2) {
						controller2Events++
					}
				}
			}
		}
	}

	suite.Require().Equal(2, controller1Events, "Should have 2 events from controller1")
	suite.Require().Equal(1, controller2Events, "Should have 1 event from controller2")
}

// Mock service keeper for DWN tests
type mockServiceKeeper struct{}

func (m *mockServiceKeeper) VerifyServiceRegistration(
	ctx context.Context,
	serviceID string,
	domain string,
) (bool, error) {
	return true, nil
}

func (m *mockServiceKeeper) GetService(
	ctx context.Context,
	serviceID string,
) (*svctypes.Service, error) {
	return &svctypes.Service{
		Id:     serviceID,
		Domain: "test.com",
		Owner:  "test-owner",
		Status: svctypes.ServiceStatus_SERVICE_STATUS_ACTIVE,
	}, nil
}

func (m *mockServiceKeeper) IsDomainVerified(
	ctx context.Context,
	domain string,
	owner string,
) (bool, error) {
	return true, nil
}

func (m *mockServiceKeeper) GetServicesByDomain(
	ctx context.Context,
	domain string,
) ([]svctypes.Service, error) {
	return []svctypes.Service{}, nil
}
