package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/zeebo/blake3"

	"github.com/sonr-io/sonr/app"
	didtypes "github.com/sonr-io/sonr/x/did/types"
	dwntypes "github.com/sonr-io/sonr/x/dwn/types"
)

// WebAuthnRegistrationTestSuite tests the complete WebAuthn registration flow
type WebAuthnRegistrationTestSuite struct {
	suite.Suite
	cfg     network.Config
	network *network.Network
}

func (suite *WebAuthnRegistrationTestSuite) SetupSuite() {
	suite.T().Log("setting up WebAuthn registration test suite")

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	// Custom app constructor
	cfg.AppConstructor = func(val network.Validator) app.TestApp {
		return app.NewTestApp(val.Ctx.Logger, val.Ctx.Config.DBDir(), nil, true, 0)
	}

	suite.cfg = cfg
	suite.network = network.New(suite.T(), cfg)

	suite.Require().NotNil(suite.network)

	// Wait for network to start
	time.Sleep(5 * time.Second)
}

func (suite *WebAuthnRegistrationTestSuite) TearDownSuite() {
	suite.T().Log("tearing down WebAuthn registration test suite")
	suite.network.Cleanup()
}

// TestRegisterStart tests the RegisterStart query handler
func (suite *WebAuthnRegistrationTestSuite) TestRegisterStart() {
	val := suite.network.Validators[0]

	testCases := []struct {
		name           string
		assertionValue string
		assertionType  string
		serviceOrigin  string
		expectErr      bool
		errMsg         string
	}{
		{
			name:           "valid email registration",
			assertionValue: "alice@example.com",
			assertionType:  "email",
			serviceOrigin:  "http://localhost:3000",
			expectErr:      false,
		},
		{
			name:           "valid phone registration",
			assertionValue: "+1234567890",
			assertionType:  "tel",
			serviceOrigin:  "http://localhost:3000",
			expectErr:      false,
		},
		{
			name:           "invalid assertion type",
			assertionValue: "alice",
			assertionType:  "invalid",
			serviceOrigin:  "http://localhost:3000",
			expectErr:      true,
			errMsg:         "unsupported assertion type",
		},
		{
			name:           "invalid origin",
			assertionValue: "alice@example.com",
			assertionType:  "email",
			serviceOrigin:  "http://malicious.com",
			expectErr:      true,
			errMsg:         "invalid origin",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create RegisterStart request
			req := &didtypes.QueryRegisterStartRequest{
				AssertionValue: tc.assertionValue,
				AssertionType:  tc.assertionType,
				ServiceOrigin:  tc.serviceOrigin,
			}

			// Query RegisterStart
			clientCtx := val.ClientCtx
			queryClient := didtypes.NewQueryClient(clientCtx)

			resp, err := queryClient.RegisterStart(context.Background(), req)

			if tc.expectErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().NotNil(resp.DIDDocument)
				suite.Require().NotEmpty(resp.DIDDocument.Id)

				// Verify challenge is generated
				suite.Require().NotNil(resp.DIDDocumentMetadata)
				// Note: In real implementation, challenge would be in metadata
			}
		})
	}
}

// TestWebAuthnCredentialRegistration tests the complete registration flow
func (suite *WebAuthnRegistrationTestSuite) TestWebAuthnCredentialRegistration() {
	val := suite.network.Validators[0]

	// Test data
	username := "testuser"
	email := "testuser@example.com"

	// Generate mock WebAuthn credential data
	credentialID := base64.StdEncoding.EncodeToString([]byte("test-credential-id"))
	publicKey := generateMockPublicKey()
	attestationObject := generateMockAttestationObject()
	clientDataJSON := generateMockClientDataJSON()

	// Create registration message
	msg := &didtypes.MsgRegisterWebAuthnCredential{
		Creator:        val.Address.String(),
		Username:       username,
		AssertionValue: email,
		AssertionType:  "email",
		WebauthnCredential: &didtypes.WebAuthnCredential{
			CredentialId:      credentialID,
			PublicKey:         publicKey,
			AttestationObject: attestationObject,
			ClientDataJson:    clientDataJSON,
		},
		CreateVault: true,
	}

	// Broadcast transaction
	clientCtx := val.ClientCtx
	txResp, err := broadcastTx(clientCtx, msg)

	suite.Require().NoError(err)
	suite.Require().Equal(uint32(0), txResp.Code)

	// Wait for transaction to be processed
	time.Sleep(2 * time.Second)

	// Verify DID document was created
	queryClient := didtypes.NewQueryClient(clientCtx)

	// Calculate expected DID
	hasher := blake3.New()
	hasher.Write([]byte(email))
	hash := hasher.Sum(nil)
	expectedDID := fmt.Sprintf("did:email:%x", hash)

	// Query for the DID document
	didResp, err := queryClient.GetDIDDocument(context.Background(), &didtypes.QueryGetDIDDocumentRequest{
		Did: expectedDID,
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(didResp)
	suite.Require().NotNil(didResp.DidDocument)
	suite.Require().Equal(expectedDID, didResp.DidDocument.Id)

	// Verify assertion methods
	suite.Require().Len(didResp.DidDocument.AssertionMethod, 2)

	// Verify authentication method (WebAuthn)
	suite.Require().Len(didResp.DidDocument.Authentication, 1)

	// Verify controller is set
	suite.Require().Len(didResp.DidDocument.Controller, 1)
}

// TestVaultCreation tests that vault is created during registration
func (suite *WebAuthnRegistrationTestSuite) TestVaultCreation() {
	val := suite.network.Validators[0]

	username := "vaultuser"
	email := "vaultuser@example.com"

	// Create registration with vault
	msg := &didtypes.MsgRegisterWebAuthnCredential{
		Creator:        val.Address.String(),
		Username:       username,
		AssertionValue: email,
		AssertionType:  "email",
		WebauthnCredential: &didtypes.WebAuthnCredential{
			CredentialId:      "vault-cred-id",
			PublicKey:         generateMockPublicKey(),
			AttestationObject: generateMockAttestationObject(),
			ClientDataJson:    generateMockClientDataJSON(),
		},
		CreateVault: true,
	}

	clientCtx := val.ClientCtx
	txResp, err := broadcastTx(clientCtx, msg)

	suite.Require().NoError(err)
	suite.Require().Equal(uint32(0), txResp.Code)

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Query vault
	dwnQueryClient := dwntypes.NewQueryClient(clientCtx)

	// Calculate expected vault ID (based on DID)
	hasher := blake3.New()
	hasher.Write([]byte(email))
	hash := hasher.Sum(nil)
	expectedDID := fmt.Sprintf("did:email:%x", hash)

	vaultResp, err := dwnQueryClient.GetVaultByDID(context.Background(), &dwntypes.QueryGetVaultByDIDRequest{
		Did: expectedDID,
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(vaultResp)
	suite.Require().NotNil(vaultResp.Vault)
	suite.Require().Equal(expectedDID, vaultResp.Vault.Did)
	suite.Require().NotEmpty(vaultResp.Vault.PublicKey)
}

// TestAssertionUniqueness tests that duplicate assertions are rejected
func (suite *WebAuthnRegistrationTestSuite) TestAssertionUniqueness() {
	val := suite.network.Validators[0]

	email := "unique@example.com"

	// First registration
	msg1 := &didtypes.MsgRegisterWebAuthnCredential{
		Creator:        val.Address.String(),
		Username:       "user1",
		AssertionValue: email,
		AssertionType:  "email",
		WebauthnCredential: &didtypes.WebAuthnCredential{
			CredentialId:      "cred-1",
			PublicKey:         generateMockPublicKey(),
			AttestationObject: generateMockAttestationObject(),
			ClientDataJson:    generateMockClientDataJSON(),
		},
		CreateVault: false,
	}

	clientCtx := val.ClientCtx
	txResp1, err := broadcastTx(clientCtx, msg1)

	suite.Require().NoError(err)
	suite.Require().Equal(uint32(0), txResp1.Code)

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Attempt duplicate registration with same email
	msg2 := &didtypes.MsgRegisterWebAuthnCredential{
		Creator:        val.Address.String(),
		Username:       "user2",
		AssertionValue: email,
		AssertionType:  "email",
		WebauthnCredential: &didtypes.WebAuthnCredential{
			CredentialId:      "cred-2",
			PublicKey:         generateMockPublicKey(),
			AttestationObject: generateMockAttestationObject(),
			ClientDataJson:    generateMockClientDataJSON(),
		},
		CreateVault: false,
	}

	txResp2, err := broadcastTx(clientCtx, msg2)

	// Should fail due to duplicate assertion
	suite.Require().Error(err)
	if txResp2 != nil {
		suite.Require().NotEqual(uint32(0), txResp2.Code)
	}
}

// TestUCANDelegationChain tests UCAN token creation during registration
func (suite *WebAuthnRegistrationTestSuite) TestUCANDelegationChain() {
	val := suite.network.Validators[0]

	username := "ucanuser"
	email := "ucanuser@example.com"

	msg := &didtypes.MsgRegisterWebAuthnCredential{
		Creator:        val.Address.String(),
		Username:       username,
		AssertionValue: email,
		AssertionType:  "email",
		WebauthnCredential: &didtypes.WebAuthnCredential{
			CredentialId:      "ucan-cred-id",
			PublicKey:         generateMockPublicKey(),
			AttestationObject: generateMockAttestationObject(),
			ClientDataJson:    generateMockClientDataJSON(),
		},
		CreateVault: true,
	}

	clientCtx := val.ClientCtx
	txResp, err := broadcastTx(clientCtx, msg)

	suite.Require().NoError(err)
	suite.Require().Equal(uint32(0), txResp.Code)

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Query DID document
	queryClient := didtypes.NewQueryClient(clientCtx)

	hasher := blake3.New()
	hasher.Write([]byte(email))
	hash := hasher.Sum(nil)
	expectedDID := fmt.Sprintf("did:email:%x", hash)

	didResp, err := queryClient.GetDIDDocument(context.Background(), &didtypes.QueryGetDIDDocumentRequest{
		Did: expectedDID,
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(didResp.DidDocumentMetadata)

	// Verify UCAN delegation chain reference exists in metadata
	// In real implementation, this would check for the actual UCAN token
	suite.Require().NotEmpty(didResp.DidDocumentMetadata.Created)
}

// TestMultipleAssertionTypes tests registration with both email and phone
func (suite *WebAuthnRegistrationTestSuite) TestMultipleAssertionTypes() {
	val := suite.network.Validators[0]

	testCases := []struct {
		name           string
		username       string
		assertionType  string
		assertionValue string
	}{
		{
			name:           "register with email",
			username:       "emailuser",
			assertionType:  "email",
			assertionValue: "multi@example.com",
		},
		{
			name:           "register with phone",
			username:       "phoneuser",
			assertionType:  "tel",
			assertionValue: "+9876543210",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			msg := &didtypes.MsgRegisterWebAuthnCredential{
				Creator:        val.Address.String(),
				Username:       tc.username,
				AssertionValue: tc.assertionValue,
				AssertionType:  tc.assertionType,
				WebauthnCredential: &didtypes.WebAuthnCredential{
					CredentialId:      fmt.Sprintf("cred-%s", tc.username),
					PublicKey:         generateMockPublicKey(),
					AttestationObject: generateMockAttestationObject(),
					ClientDataJson:    generateMockClientDataJSON(),
				},
				CreateVault: false,
			}

			clientCtx := val.ClientCtx
			txResp, err := broadcastTx(clientCtx, msg)

			suite.Require().NoError(err)
			suite.Require().Equal(uint32(0), txResp.Code)

			// Wait for processing
			time.Sleep(2 * time.Second)

			// Verify DID was created with correct prefix
			hasher := blake3.New()
			hasher.Write([]byte(tc.assertionValue))
			hash := hasher.Sum(nil)
			expectedDID := fmt.Sprintf("did:%s:%x", tc.assertionType, hash)

			queryClient := didtypes.NewQueryClient(clientCtx)
			didResp, err := queryClient.GetDIDDocument(context.Background(), &didtypes.QueryGetDIDDocumentRequest{
				Did: expectedDID,
			})

			suite.Require().NoError(err)
			suite.Require().NotNil(didResp.DidDocument)
			suite.Require().Equal(expectedDID, didResp.DidDocument.Id)
		})
	}
}

// Helper functions

func broadcastTx(clientCtx client.Context, msg sdk.Msg) (*sdk.TxResponse, error) {
	// This is a simplified version - in real tests, use proper tx broadcasting
	// with proper fee calculation and signing
	return nil, nil
}

func generateMockPublicKey() string {
	// Generate a mock public key for testing
	// In real tests, this would be a proper WebAuthn public key
	return base64.StdEncoding.EncodeToString([]byte("mock-public-key"))
}

func generateMockAttestationObject() string {
	// Generate a mock attestation object
	attestation := map[string]interface{}{
		"fmt":      "packed",
		"attStmt":  map[string]interface{}{},
		"authData": base64.StdEncoding.EncodeToString([]byte("mock-auth-data")),
	}
	data, _ := json.Marshal(attestation)
	return base64.StdEncoding.EncodeToString(data)
}

func generateMockClientDataJSON() string {
	// Generate mock client data
	clientData := map[string]interface{}{
		"type":      "webauthn.create",
		"challenge": "test-challenge",
		"origin":    "http://localhost:3000",
	}
	data, _ := json.Marshal(clientData)
	return base64.StdEncoding.EncodeToString(data)
}

func TestWebAuthnRegistrationTestSuite(t *testing.T) {
	suite.Run(t, new(WebAuthnRegistrationTestSuite))
}
