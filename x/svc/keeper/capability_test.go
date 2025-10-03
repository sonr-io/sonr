package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	didtypes "github.com/sonr-io/sonr/x/did/types"
	"github.com/sonr-io/sonr/x/svc/keeper"
	"github.com/sonr-io/sonr/x/svc/types"
)

// CapabilityTestSuite tests capability management
type CapabilityTestSuite struct {
	suite.Suite

	ctx      context.Context
	keeper   keeper.Keeper
	storeKey *storetypes.KVStoreKey
	cdc      codec.BinaryCodec
}

func (suite *CapabilityTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	suite.storeKey = key
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(
		suite.T(),
		key,
		storetypes.NewTransientStoreKey("transient_test"),
	)
	suite.ctx = testCtx.Ctx.WithBlockHeader(sdk.Context{}.BlockHeader())

	encCfg := moduletestutil.MakeTestEncodingConfig()
	suite.cdc = encCfg.Codec

	authority := sdk.AccAddress([]byte("authority"))

	// Mock DID keeper
	mockDIDKeeper := &MockDIDKeeper{}

	suite.keeper = keeper.NewKeeper(
		suite.cdc,
		storeService,
		log.NewNopLogger(),
		authority.String(),
		mockDIDKeeper,
	)
}

func TestCapabilityTestSuite(t *testing.T) {
	suite.Run(t, new(CapabilityTestSuite))
}

// TestCreateCapability tests capability creation
func (suite *CapabilityTestSuite) TestCreateCapability() {
	testCases := []struct {
		name        string
		setup       func() *types.ServiceCapability
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid capability creation",
			setup: func() *types.ServiceCapability {
				return &types.ServiceCapability{
					CapabilityId: "cap_service1_read_1234_0",
					ServiceId:    "service1",
					Domain:       "example.com",
					Abilities:    []string{"read", "write"},
					Owner:        "cosmos1abc123",
					CreatedAt:    time.Now().Unix(),
					ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
					Revoked:      false,
				}
			},
			expectError: false,
		},
		{
			name: "nil capability",
			setup: func() *types.ServiceCapability {
				return nil
			},
			expectError: true,
			errorMsg:    "capability cannot be nil",
		},
		{
			name: "empty capability ID",
			setup: func() *types.ServiceCapability {
				return &types.ServiceCapability{
					ServiceId: "service1",
					Domain:    "example.com",
					Abilities: []string{"read"},
					Owner:     "cosmos1abc123",
				}
			},
			expectError: true,
			errorMsg:    "capability ID cannot be empty",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			capability := tc.setup()
			err := suite.keeper.StoreCapability(suite.ctx, capability)

			if tc.expectError {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errorMsg)
			} else {
				suite.Require().NoError(err)

				// Verify capability was stored
				loaded, err := suite.keeper.LoadCapability(suite.ctx, capability.CapabilityId)
				suite.Require().NoError(err)
				suite.Require().Equal(capability.CapabilityId, loaded.CapabilityId)
				suite.Require().Equal(capability.ServiceId, loaded.ServiceId)
				suite.Require().Equal(capability.Abilities, loaded.Abilities)
			}
		})
	}
}

// TestValidateCapability tests capability validation
func (suite *CapabilityTestSuite) TestValidateCapability() {
	// Store a test capability
	capability := &types.ServiceCapability{
		CapabilityId: "cap_test_read_1234_0",
		ServiceId:    "test-service",
		Domain:       "test.com",
		Abilities:    []string{"read", "write"},
		Owner:        "cosmos1test",
		CreatedAt:    time.Now().Unix(),
		ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
		Revoked:      false,
	}
	err := suite.keeper.StoreCapability(suite.ctx, capability)
	suite.Require().NoError(err)

	testCases := []struct {
		name         string
		capabilityID string
		serviceID    string
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "valid capability",
			capabilityID: "cap_test_read_1234_0",
			serviceID:    "test-service",
			expectError:  false,
		},
		{
			name:         "empty capability ID",
			capabilityID: "",
			serviceID:    "test-service",
			expectError:  true,
			errorMsg:     "capability ID cannot be empty",
		},
		{
			name:         "empty service ID",
			capabilityID: "cap_test_read_1234_0",
			serviceID:    "",
			expectError:  true,
			errorMsg:     "service ID cannot be empty",
		},
		{
			name:         "wrong service ID",
			capabilityID: "cap_test_read_1234_0",
			serviceID:    "wrong-service",
			expectError:  true,
			errorMsg:     "does not belong to service",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			validated, err := suite.keeper.ValidateCapability(
				suite.ctx,
				tc.capabilityID,
				tc.serviceID,
			)

			if tc.expectError {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errorMsg)
				suite.Require().Nil(validated)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(validated)
				suite.Require().Equal(capability.CapabilityId, validated.CapabilityId)
			}
		})
	}
}

// TestRevokeCapability tests capability revocation
func (suite *CapabilityTestSuite) TestRevokeCapability() {
	// Store a test capability
	capability := &types.ServiceCapability{
		CapabilityId: "cap_revoke_test_1234_0",
		ServiceId:    "revoke-service",
		Domain:       "revoke.com",
		Abilities:    []string{"admin"},
		Owner:        "cosmos1owner",
		CreatedAt:    time.Now().Unix(),
		ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
		Revoked:      false,
	}
	err := suite.keeper.StoreCapability(suite.ctx, capability)
	suite.Require().NoError(err)

	testCases := []struct {
		name         string
		capabilityID string
		revoker      string
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "valid revocation by owner",
			capabilityID: "cap_revoke_test_1234_0",
			revoker:      "cosmos1owner",
			expectError:  false,
		},
		{
			name:         "empty capability ID",
			capabilityID: "",
			revoker:      "cosmos1owner",
			expectError:  true,
			errorMsg:     "capability ID cannot be empty",
		},
		{
			name:         "empty revoker",
			capabilityID: "cap_revoke_test_1234_0",
			revoker:      "",
			expectError:  true,
			errorMsg:     "revoker cannot be empty",
		},
		{
			name:         "non-owner revocation",
			capabilityID: "cap_revoke_test_1234_0",
			revoker:      "cosmos1other",
			expectError:  true,
			errorMsg:     "not authorized to revoke",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Reset capability state before tests that need it unrevoked
			if tc.name == "valid revocation by owner" || tc.name == "non-owner revocation" {
				capability.Revoked = false
				err := suite.keeper.StoreCapability(suite.ctx, capability)
				suite.Require().NoError(err)
			}

			err := suite.keeper.RevokeCapability(suite.ctx, tc.capabilityID, tc.revoker)

			if tc.expectError {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errorMsg)
			} else {
				suite.Require().NoError(err)

				// Verify capability is revoked
				loaded, err := suite.keeper.LoadCapability(suite.ctx, tc.capabilityID)
				suite.Require().NoError(err)
				suite.Require().True(loaded.Revoked)
			}
		})
	}
}

// TestExpiredCapability tests expired capability validation
func (suite *CapabilityTestSuite) TestExpiredCapability() {
	// Store an expired capability
	expiredCapability := &types.ServiceCapability{
		CapabilityId: "cap_expired_test_1234_0",
		ServiceId:    "expired-service",
		Domain:       "expired.com",
		Abilities:    []string{"read"},
		Owner:        "cosmos1expired",
		CreatedAt:    time.Now().Add(-48 * time.Hour).Unix(),
		ExpiresAt:    time.Now().Add(-24 * time.Hour).Unix(), // Expired
		Revoked:      false,
	}
	err := suite.keeper.StoreCapability(suite.ctx, expiredCapability)
	suite.Require().NoError(err)

	// Validation should fail for expired capability
	validated, err := suite.keeper.ValidateCapability(
		suite.ctx,
		"cap_expired_test_1234_0",
		"expired-service",
	)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "has expired")
	suite.Require().Nil(validated)
}

// TestCapabilityChainValidation tests permission chain validation
func (suite *CapabilityTestSuite) TestCapabilityChainValidation() {
	// Store multiple capabilities for chain validation
	capabilities := []*types.ServiceCapability{
		{
			CapabilityId: "cap_chain_read_1234_0",
			ServiceId:    "chain-service",
			Domain:       "chain.com",
			Abilities:    []string{"read"},
			Owner:        "cosmos1chain",
			CreatedAt:    time.Now().Unix(),
			ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
			Revoked:      false,
		},
		{
			CapabilityId: "cap_chain_write_1234_1",
			ServiceId:    "chain-service",
			Domain:       "chain.com",
			Abilities:    []string{"write"},
			Owner:        "cosmos1chain",
			CreatedAt:    time.Now().Unix(),
			ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
			Revoked:      false,
		},
	}

	for _, cap := range capabilities {
		err := suite.keeper.StoreCapability(suite.ctx, cap)
		suite.Require().NoError(err)
	}

	testCases := []struct {
		name                string
		capabilityChain     []string
		serviceID           string
		requiredPermissions []string
		expectError         bool
		errorMsg            string
	}{
		{
			name:                "valid chain with all permissions",
			capabilityChain:     []string{"cap_chain_read_1234_0", "cap_chain_write_1234_1"},
			serviceID:           "chain-service",
			requiredPermissions: []string{"read", "write"},
			expectError:         false,
		},
		{
			name:                "empty capability chain",
			capabilityChain:     []string{},
			serviceID:           "chain-service",
			requiredPermissions: []string{"read"},
			expectError:         true,
			errorMsg:            "capability chain cannot be empty",
		},
		{
			name:                "empty service ID",
			capabilityChain:     []string{"cap_chain_read_1234_0"},
			serviceID:           "",
			requiredPermissions: []string{"read"},
			expectError:         true,
			errorMsg:            "service ID cannot be empty",
		},
		{
			name:                "insufficient capabilities",
			capabilityChain:     []string{"cap_chain_read_1234_0"},
			serviceID:           "chain-service",
			requiredPermissions: []string{"read", "write", "admin"},
			expectError:         true,
			errorMsg:            "insufficient capabilities",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := suite.keeper.ValidatePermissionCapabilityChain(
				suite.ctx,
				tc.capabilityChain,
				tc.serviceID,
				tc.requiredPermissions,
			)

			if tc.expectError {
				suite.Require().Error(err)
				if tc.errorMsg != "" {
					suite.Require().Contains(err.Error(), tc.errorMsg)
				}
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// MockDIDKeeper is a mock implementation of DIDKeeper for testing
type MockDIDKeeper struct{}

func (m *MockDIDKeeper) ResolveDID(
	ctx context.Context,
	did string,
) (*didtypes.DIDDocument, *didtypes.DIDDocumentMetadata, error) {
	doc := &didtypes.DIDDocument{
		Id:                did,
		PrimaryController: did,
		VerificationMethod: []*didtypes.VerificationMethod{
			{
				Id:                     did + "#key1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				Controller:             did,
			},
		},
		Deactivated: false,
	}

	metadata := &didtypes.DIDDocumentMetadata{
		VersionId:   "1",
		Created:     time.Now().Unix(),
		Updated:     time.Now().Unix(),
		Deactivated: 0,
	}

	return doc, metadata, nil
}

func (m *MockDIDKeeper) GetDIDDocument(
	ctx context.Context,
	did string,
) (*didtypes.DIDDocument, error) {
	return &didtypes.DIDDocument{
		Id:                did,
		PrimaryController: did,
		VerificationMethod: []*didtypes.VerificationMethod{
			{
				Id:                     did + "#key1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				Controller:             did,
			},
		},
		Deactivated: false,
	}, nil
}

func (m *MockDIDKeeper) VerifyDIDDocumentSignature(
	ctx context.Context,
	did string,
	signature []byte,
) (bool, error) {
	// For testing, always return true
	return true, nil
}
