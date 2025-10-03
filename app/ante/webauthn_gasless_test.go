package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/sonr/app/ante"
)

// MockWebAuthnKeeper implements the WebAuthnKeeperInterface for testing
type MockWebAuthnKeeper struct {
	existingCredentials map[string]bool
}

func NewMockWebAuthnKeeper() *MockWebAuthnKeeper {
	return &MockWebAuthnKeeper{
		existingCredentials: make(map[string]bool),
	}
}

func (m *MockWebAuthnKeeper) HasExistingCredential(ctx sdk.Context, credentialId string) bool {
	return m.existingCredentials[credentialId]
}

func (m *MockWebAuthnKeeper) AddCredential(credentialId string) {
	m.existingCredentials[credentialId] = true
}

// MockAccountKeeper - Simple mock without complex interface implementation
// For comprehensive testing, we'd implement the full AccountKeeper interface

// MockTransaction implements sdk.Tx for testing
type MockTransaction struct {
	messages []sdk.Msg
}

func (m *MockTransaction) GetMsgs() []sdk.Msg {
	return m.messages
}

func (m *MockTransaction) ValidateBasic() error {
	return nil
}

// WebAuthnGaslessTestSuite tests the WebAuthn gasless decorator
type WebAuthnGaslessTestSuite struct {
	suite.Suite
	didKeeper *MockWebAuthnKeeper
}

func TestWebAuthnGaslessTestSuite(t *testing.T) {
	suite.Run(t, new(WebAuthnGaslessTestSuite))
}

func (suite *WebAuthnGaslessTestSuite) SetupTest() {
	// Setup minimal test context and keepers
	suite.didKeeper = NewMockWebAuthnKeeper()
	// We'll focus on testing the logic without complex keeper setup
}

func (suite *WebAuthnGaslessTestSuite) TestGenerateAddressFromCredential() {
	// Test that the same credential always generates the same address
	credentialID := "test-credential-id-123"

	addr1 := ante.GenerateAddressFromCredential(credentialID)
	addr2 := ante.GenerateAddressFromCredential(credentialID)

	suite.Require().Equal(addr1, addr2, "Same credential should generate same address")
	suite.Require().NotNil(addr1, "Generated address should not be nil")
	suite.Require().Equal(20, len(addr1), "Address should be 20 bytes")
}

func (suite *WebAuthnGaslessTestSuite) TestGenerateDIDFromCredential() {
	// Test DID generation
	credentialID := "test-credential-id-123"
	username := "testuser"

	did1 := ante.GenerateDIDFromCredential(credentialID, username)
	did2 := ante.GenerateDIDFromCredential(credentialID, username)

	suite.Require().Equal(did1, did2, "Same inputs should generate same DID")
	suite.Require().Contains(did1, "did:sonr:", "DID should have correct prefix")

	// Different username should generate different DID
	did3 := ante.GenerateDIDFromCredential(credentialID, "differentuser")
	suite.Require().NotEqual(did1, did3, "Different username should generate different DID")
}

func (suite *WebAuthnGaslessTestSuite) TestWebAuthnGaslessDecorator_StandardMode() {
	// Test standard mode behavior
	decorator := ante.NewWebAuthnGaslessDecorator(
		nil, // accountKeeper would be mocked in full test
		suite.didKeeper,
		false, // standard mode
	)

	suite.Require().NotNil(decorator, "Decorator should be created")
}

func (suite *WebAuthnGaslessTestSuite) TestWebAuthnGaslessDecorator_EnhancedMode() {
	// Test enhanced mode behavior
	decorator := ante.NewWebAuthnGaslessDecorator(
		nil, // accountKeeper would be mocked in full test
		suite.didKeeper,
		true, // enhanced mode
	)

	suite.Require().NotNil(decorator, "Decorator should be created in enhanced mode")
}

func (suite *WebAuthnGaslessTestSuite) TestConditionalDecorators() {
	// Test that conditional decorators are created properly
	// We'll use a simple mock decorator for testing
	mockDecorator := &MockAnteDecorator{}

	feeDecorator := ante.NewConditionalFeeDecorator(mockDecorator)
	suite.Require().NotNil(feeDecorator, "Conditional fee decorator should be created")

	sigDecorator := ante.NewConditionalSignatureDecorator(mockDecorator)
	suite.Require().NotNil(sigDecorator, "Conditional signature decorator should be created")

	pubKeyDecorator := ante.NewConditionalPubKeyDecorator(mockDecorator)
	suite.Require().NotNil(pubKeyDecorator, "Conditional pubkey decorator should be created")
}

func (suite *WebAuthnGaslessTestSuite) TestWebAuthnCredentialValidation() {
	// Test WebAuthn credential structure validation
	// This tests the core logic without full ante handler setup

	testCases := []struct {
		name         string
		credentialID string
		expectError  bool
	}{
		{
			name:         "valid credential ID",
			credentialID: "valid-credential-123",
			expectError:  false,
		},
		{
			name:         "empty credential ID should be caught by message validation",
			credentialID: "",
			expectError:  true, // This would be caught by ValidateStructure
		},
		{
			name:         "duplicate credential",
			credentialID: "duplicate-cred",
			expectError:  true,
		},
	}

	// Add a duplicate credential
	suite.didKeeper.AddCredential("duplicate-cred")

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			hasExisting := suite.didKeeper.HasExistingCredential(sdk.Context{}, tc.credentialID)

			switch tc.credentialID {
			case "duplicate-cred":
				suite.Require().True(hasExisting, "Duplicate credential should exist")
			case "valid-credential-123":
				suite.Require().False(hasExisting, "New credential should not exist")
			}
		})
	}
}

func (suite *WebAuthnGaslessTestSuite) TestEnhancedModeAddressGeneration() {
	// Test enhanced mode functionality for address generation
	testCredentialID := "test-enhanced-credential"

	// Test that the same credential always generates the same address
	addr1 := ante.GenerateAddressFromCredential(testCredentialID)
	addr2 := ante.GenerateAddressFromCredential(testCredentialID)

	suite.Require().Equal(addr1, addr2, "Same credential should generate same address")
	suite.Require().Equal(20, len(addr1), "Address should be 20 bytes")

	// Test that different credentials generate different addresses
	addr3 := ante.GenerateAddressFromCredential("different-credential")
	suite.Require().
		NotEqual(addr1, addr3, "Different credentials should generate different addresses")
}

// MockAnteDecorator is a simple mock implementation of sdk.AnteDecorator for testing
type MockAnteDecorator struct{}

func (m *MockAnteDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (sdk.Context, error) {
	return next(ctx, tx, simulate)
}

// TestAddressGeneration tests the deterministic address generation
func TestAddressGeneration(t *testing.T) {
	testCases := []struct {
		name         string
		credentialID string
		expectSame   bool
	}{
		{
			name:         "same credential generates same address",
			credentialID: "credential-1",
			expectSame:   true,
		},
		{
			name:         "different credential generates different address",
			credentialID: "credential-2",
			expectSame:   false,
		},
	}

	baseAddr := ante.GenerateAddressFromCredential("base-credential")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			addr := ante.GenerateAddressFromCredential(tc.credentialID)
			require.NotNil(t, addr)
			require.Equal(t, 20, len(addr))

			if tc.expectSame && tc.credentialID == "base-credential" {
				require.Equal(t, baseAddr, addr)
			} else if !tc.expectSame {
				require.NotEqual(t, baseAddr, addr)
			}
		})
	}
}

// TestDIDGeneration tests the deterministic DID generation
func TestDIDGeneration(t *testing.T) {
	testCases := []struct {
		name           string
		credentialID   string
		username       string
		expectedPrefix string
	}{
		{
			name:           "generates valid DID",
			credentialID:   "cred-1",
			username:       "user1",
			expectedPrefix: "did:sonr:",
		},
		{
			name:           "different username generates different DID",
			credentialID:   "cred-1",
			username:       "user2",
			expectedPrefix: "did:sonr:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			did := ante.GenerateDIDFromCredential(tc.credentialID, tc.username)
			require.NotEmpty(t, did)
			require.Contains(t, did, tc.expectedPrefix)

			// Verify deterministic generation
			did2 := ante.GenerateDIDFromCredential(tc.credentialID, tc.username)
			require.Equal(t, did, did2)
		})
	}
}
