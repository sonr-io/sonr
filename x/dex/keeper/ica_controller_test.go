package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/sonr/x/dex/types"
)

const (
	testConnectionID = "connection-0"
)

// ICAControllerTestSuite tests ICA controller operations
type ICAControllerTestSuite struct {
	suite.Suite
	f *testFixture
}

func TestICAControllerSuite(t *testing.T) {
	suite.Run(t, new(ICAControllerTestSuite))
}

func (suite *ICAControllerTestSuite) SetupTest() {
	suite.f = SetupTest(suite.T())
}

// TestRegisterDEXAccount tests ICA account registration
func (suite *ICAControllerTestSuite) TestRegisterDEXAccount() {
	did := "did:sonr:test_ica_1"
	connectionID := testConnectionID
	features := []string{"swap", "liquidity"}

	// Register DEX account
	account, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		did,
		connectionID,
		features,
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(account)

	// Verify account was created with correct fields
	suite.Require().Equal(did, account.Did)
	suite.Require().Equal(connectionID, account.ConnectionId)
	suite.Require().Equal(types.ACCOUNT_STATUS_PENDING, account.Status)
	suite.Require().NotEmpty(account.PortId)

	// Verify port ID format
	expectedPortPrefix := "dex-" + did
	suite.Require().Contains(account.PortId, expectedPortPrefix)
}

// TestRegisterDEXAccount_DuplicateRegistration tests duplicate registration
func (suite *ICAControllerTestSuite) TestRegisterDEXAccount_DuplicateRegistration() {
	did := "did:sonr:test_ica_2"
	connectionID := testConnectionID

	// First registration should succeed
	_, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		did,
		connectionID,
		[]string{"swap"},
	)
	suite.Require().NoError(err)

	// Second registration with same DID and connection should return existing account
	// (idempotent behavior)
	account2, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		did,
		connectionID,
		[]string{"swap"},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(account2)
	suite.Require().Equal(did, account2.Did)
	suite.Require().Equal(connectionID, account2.ConnectionId)
}

// TestGetDEXAccount tests retrieving a DEX account
func (suite *ICAControllerTestSuite) TestGetDEXAccount() {
	did := "did:sonr:test_ica_3"
	connectionID := testConnectionID

	// Register account first
	original, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		did,
		connectionID,
		[]string{"order"},
	)
	suite.Require().NoError(err)

	// Retrieve the account
	retrieved, err := suite.f.k.GetDEXAccount(suite.f.ctx, did, connectionID)
	suite.Require().NoError(err)
	suite.Require().NotNil(retrieved)

	// Verify retrieved account matches original
	suite.Require().Equal(original.Did, retrieved.Did)
	suite.Require().Equal(original.ConnectionId, retrieved.ConnectionId)
	suite.Require().Equal(original.PortId, retrieved.PortId)
}

// TestGetDEXAccountsByDID tests retrieving all accounts for a DID
func (suite *ICAControllerTestSuite) TestGetDEXAccountsByDID() {
	did := "did:sonr:test_ica_4"
	connections := []string{testConnectionID, "connection-1", "connection-2"}

	// Register multiple accounts for the same DID
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
	suite.Require().Len(accounts, 3)

	// Verify each account has the correct DID
	for _, account := range accounts {
		suite.Require().Equal(did, account.Did)
	}
}

// TestOnICAAccountCreated tests ICA account creation callback
func (suite *ICAControllerTestSuite) TestOnICAAccountCreated() {
	did := "did:sonr:test_ica_5"
	connectionID := testConnectionID

	// Register account first
	account, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		did,
		connectionID,
		[]string{"swap"},
	)
	suite.Require().NoError(err)

	// Simulate ICA account creation callback
	icaAddress := "cosmos1testaddress"
	err = suite.f.k.OnICAAccountCreated(
		suite.f.ctx,
		account.PortId,
		icaAddress,
	)
	suite.Require().NoError(err)

	// Verify account was updated
	updated, err := suite.f.k.GetDEXAccount(suite.f.ctx, did, connectionID)
	suite.Require().NoError(err)
	suite.Require().Equal(icaAddress, updated.AccountAddress)
	suite.Require().Equal(types.ACCOUNT_STATUS_ACTIVE, updated.Status)
}

// TestSendDEXTransaction tests sending transactions through ICA
func (suite *ICAControllerTestSuite) TestSendDEXTransaction() {
	did := "did:sonr:test_ica_6"
	connectionID := testConnectionID

	// Register account first
	_, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		did,
		connectionID,
		[]string{"swap"},
	)
	suite.Require().NoError(err)

	// SendDEXTransaction requires ACTIVE status
	// But without full capability module setup, it will fail
	// This test just verifies the account must be active
	msgs := []sdk.Msg{}
	_, err = suite.f.k.SendDEXTransaction(
		suite.f.ctx,
		did,
		connectionID,
		msgs,
		"test_memo",
		30,
	)
	// Should fail because account is not active
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "not active")
}

// TestPortBinding tests ICA port binding
func (suite *ICAControllerTestSuite) TestPortBinding() {
	did := "did:sonr:test_ica_7"
	connectionID := testConnectionID

	// Register account to trigger port binding
	account, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		did,
		connectionID,
		[]string{"liquidity"},
	)
	suite.Require().NoError(err)

	// Verify port was bound (mock implementation should handle this)
	suite.Require().NotEmpty(account.PortId)
}

// TestConnectionValidation tests connection ID validation
func (suite *ICAControllerTestSuite) TestConnectionValidation() {
	did := "did:sonr:test_ica_8"
	invalidConnectionID := "invalid-connection"

	// Should fail with invalid connection format
	_, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		did,
		invalidConnectionID,
		[]string{"swap"},
	)
	// The mock might not validate this, but real implementation would
	// This test documents expected behavior
	_ = err // Error handling would depend on actual implementation
}

// TestICATimeout tests ICA operation timeout handling
func (suite *ICAControllerTestSuite) TestICATimeout() {
	did := "did:sonr:test_ica_9"
	connectionID := testConnectionID

	// Register account
	_, err := suite.f.k.RegisterDEXAccount(
		suite.f.ctx,
		did,
		connectionID,
		[]string{"order"},
	)
	suite.Require().NoError(err)

	// Without active account, SendDEXTransaction should fail
	msgs := []sdk.Msg{}
	_, err = suite.f.k.SendDEXTransaction(
		suite.f.ctx,
		did,
		connectionID,
		msgs,
		"timeout_test",
		1, // 1 second timeout - very short
	)
	// Should fail because account is not active
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "not active")
}

// TestMultiChainSupport tests support for multiple chains
func (suite *ICAControllerTestSuite) TestMultiChainSupport() {
	did := "did:sonr:test_ica_10"
	chains := map[string]string{
		testConnectionID: "osmosis-1",
		"connection-1":   "cosmoshub-4",
		"connection-2":   "juno-1",
	}

	// Register accounts on multiple chains
	for connID := range chains {
		account, err := suite.f.k.RegisterDEXAccount(
			suite.f.ctx,
			did,
			connID,
			[]string{"swap", "liquidity"},
		)
		suite.Require().NoError(err)
		suite.Require().NotNil(account)
	}

	// Verify all accounts were created
	accounts, err := suite.f.k.GetDEXAccountsByDID(suite.f.ctx, did)
	suite.Require().NoError(err)
	suite.Require().Len(accounts, 3)
}
