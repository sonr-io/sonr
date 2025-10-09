package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/sonr-io/crypto/mpc"
	"github.com/sonr-io/crypto/ucan"
	"github.com/sonr-io/sonr/x/dwn/client/plugin"
)

// UCANWalletE2ETestSuite tests end-to-end UCAN operations with wallet functionality
type UCANWalletE2ETestSuite struct {
	suite.Suite
	ctx            context.Context
	chainID        string
	enclaveData    *mpc.EnclaveData
	walletPlugin   plugin.Plugin
	issuerDID      string
	audienceDID    string
	vaultConfig    map[string]any
	originToken    string
	delegatedToken string
}

func (suite *UCANWalletE2ETestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.chainID = "sonrtest_1-1"
	suite.audienceDID = "did:sonr:test-audience"

	// Initialize test enclave data
	suite.enclaveData = &mpc.EnclaveData{
		PubHex:   "0x04b9e72dfd423bcf95b3801ac93f74ec5ecf47f2cc7d8c5b7a0e4d4e0e3f2a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4",
		PubBytes: []byte{4, 185, 231, 45, 253, 66, 59, 207, 149, 179, 128, 26, 201, 63, 116, 236, 94, 207, 71, 242, 204, 125, 140, 91, 122, 14, 77, 78, 14, 63, 42, 27, 44, 61, 78, 95, 106, 123, 140, 157, 14, 31, 42, 59, 76, 93, 110, 127, 138, 155, 12, 29, 46, 63, 74, 91, 108, 125, 142, 159, 10, 27, 44, 61, 78},
		Nonce:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		Curve:    mpc.CurveName("secp256k1"),
	}

	// Vault configuration for wallet
	suite.vaultConfig = map[string]any{
		"enclave_type": "mpc",
		"threshold":    2,
		"parties":      3,
		"network":      "testnet",
	}
}

func (suite *UCANWalletE2ETestSuite) SetupTest() {
	// Initialize plugin with enclave for each test
	enclaveBytes, err := json.Marshal(suite.enclaveData)
	suite.Require().NoError(err, "Failed to marshal enclave data")

	// Load plugin with enclave and vault config
	suite.walletPlugin, err = plugin.LoadPluginWithEnclave(
		suite.ctx,
		suite.chainID,
		enclaveBytes,
		suite.vaultConfig,
	)
	suite.Require().NoError(err, "Failed to load plugin with enclave")

	// Get issuer DID from the wallet
	issuerResp, err := suite.walletPlugin.GetIssuerDID()
	suite.Require().NoError(err, "Failed to get issuer DID")
	suite.issuerDID = issuerResp.DID
}

// TestE2EWalletInitialization tests wallet initialization with EnclaveData
func (suite *UCANWalletE2ETestSuite) TestE2EWalletInitialization() {
	// Verify wallet is initialized
	suite.NotNil(suite.walletPlugin, "Wallet plugin should be initialized")

	// Verify issuer DID is generated
	suite.NotEmpty(suite.issuerDID, "Issuer DID should be generated")
	suite.Contains(suite.issuerDID, "did:sonr:", "Issuer DID should have correct format")

	// Get issuer info again to verify consistency
	issuerResp, err := suite.walletPlugin.GetIssuerDID()
	suite.NoError(err, "Should retrieve issuer DID")
	suite.Equal(suite.issuerDID, issuerResp.DID, "Issuer DID should be consistent")
	suite.NotEmpty(issuerResp.Address, "Address should be generated")
	suite.NotEmpty(issuerResp.ChainCode, "Chain code should be generated")
}

// TestE2ECreateOriginUCANToken tests creating an origin UCAN token
func (suite *UCANWalletE2ETestSuite) TestE2ECreateOriginUCANToken() {
	// Create origin token request
	req := &plugin.NewOriginTokenRequest{
		AudienceDID: suite.audienceDID,
		Attenuations: []map[string]any{
			{
				"can":  []string{"vault/read", "vault/write"},
				"with": "vault://user-vault",
			},
			{
				"can":  []string{"dwn/read"},
				"with": "dwn://user-records",
			},
		},
		Facts: []string{
			"origin=test-wallet",
			"purpose=e2e-testing",
		},
		NotBefore: time.Now().Unix(),
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	// Create origin token
	resp, err := suite.walletPlugin.NewOriginToken(req)
	suite.NoError(err, "Should create origin token")
	suite.NotEmpty(resp.Token, "Token should be generated")
	suite.Equal(suite.issuerDID, resp.Issuer, "Issuer should match wallet DID")
	suite.NotEmpty(resp.Address, "Address should be included")
	suite.Empty(resp.Error, "Should not have error")

	// Store for later tests
	suite.originToken = resp.Token

	// Verify token can be parsed
	parsedToken, err := ucan.ParseToken(resp.Token)
	suite.NoError(err, "Should parse generated token")
	suite.Equal(suite.issuerDID, parsedToken.Issuer, "Parsed issuer should match")
	suite.Equal(suite.audienceDID, parsedToken.Audience, "Parsed audience should match")
}

// TestE2ECreateDelegatedUCANToken tests creating a delegated UCAN token
func (suite *UCANWalletE2ETestSuite) TestE2ECreateDelegatedUCANToken() {
	// First create an origin token
	suite.TestE2ECreateOriginUCANToken()
	suite.Require().NotEmpty(suite.originToken, "Origin token required for delegation")

	// Create delegated token request with attenuated permissions
	req := &plugin.NewAttenuatedTokenRequest{
		ParentToken: suite.originToken,
		AudienceDID: "did:sonr:delegated-service",
		Attenuations: []map[string]any{
			{
				"can":  []string{"vault/read"}, // Reduced from read/write
				"with": "vault://user-vault",
			},
			// Removed dwn permissions - further attenuation
		},
		Facts: []string{
			"delegation=level-1",
			"delegator=test-wallet",
		},
		NotBefore: time.Now().Unix(),
		ExpiresAt: time.Now().Add(12 * time.Hour).Unix(), // Shorter expiration
	}

	// Create delegated token
	resp, err := suite.walletPlugin.NewAttenuatedToken(req)
	suite.NoError(err, "Should create delegated token")
	suite.NotEmpty(resp.Token, "Delegated token should be generated")
	suite.Equal(suite.issuerDID, resp.Issuer, "Issuer should be original wallet")
	suite.Empty(resp.Error, "Should not have error")

	// Store for later tests
	suite.delegatedToken = resp.Token

	// Verify delegation chain
	delegatedParsed, err := ucan.ParseToken(resp.Token)
	suite.NoError(err, "Should parse delegated token")
	suite.Equal("did:sonr:delegated-service", delegatedParsed.Audience)
	suite.NotEmpty(delegatedParsed.Proofs, "Should have proof chain")
}

// TestE2ESignAndVerifyData tests signing and verifying data with the wallet
func (suite *UCANWalletE2ETestSuite) TestE2ESignAndVerifyData() {
	testData := []byte("test message for signing")

	// Sign data
	signReq := &plugin.SignDataRequest{
		Data: testData,
	}
	signResp, err := suite.walletPlugin.SignData(signReq)
	suite.NoError(err, "Should sign data")
	suite.NotEmpty(signResp.Signature, "Signature should be generated")
	suite.Empty(signResp.Error, "Should not have error")

	// Verify signature
	verifyReq := &plugin.VerifyDataRequest{
		Data:      testData,
		Signature: signResp.Signature,
	}
	verifyResp, err := suite.walletPlugin.VerifyData(verifyReq)
	suite.NoError(err, "Should verify signature")
	suite.True(verifyResp.Valid, "Signature should be valid")
	suite.Empty(verifyResp.Error, "Should not have error")

	// Verify with wrong data should fail
	wrongData := []byte("different message")
	wrongVerifyReq := &plugin.VerifyDataRequest{
		Data:      wrongData,
		Signature: signResp.Signature,
	}
	wrongVerifyResp, err := suite.walletPlugin.VerifyData(wrongVerifyReq)
	suite.NoError(err, "Should handle verification")
	suite.False(wrongVerifyResp.Valid, "Signature should be invalid for wrong data")
}

// TestE2EUCANTokenChainValidation tests validation of UCAN delegation chains
func (suite *UCANWalletE2ETestSuite) TestE2EUCANTokenChainValidation() {
	// Create origin token
	suite.TestE2ECreateOriginUCANToken()
	suite.Require().NotEmpty(suite.originToken)

	// Create first delegation
	firstDelegation := &plugin.NewAttenuatedTokenRequest{
		ParentToken: suite.originToken,
		AudienceDID: "did:sonr:service-a",
		Attenuations: []map[string]any{
			{
				"can":  []string{"vault/read", "vault/write"},
				"with": "vault://user-vault",
			},
		},
		ExpiresAt: time.Now().Add(20 * time.Hour).Unix(),
	}

	firstResp, err := suite.walletPlugin.NewAttenuatedToken(firstDelegation)
	suite.NoError(err, "Should create first delegation")
	suite.NotEmpty(firstResp.Token)

	// Create second delegation from first
	secondDelegation := &plugin.NewAttenuatedTokenRequest{
		ParentToken: firstResp.Token,
		AudienceDID: "did:sonr:service-b",
		Attenuations: []map[string]any{
			{
				"can":  []string{"vault/read"}, // Further attenuated
				"with": "vault://user-vault",
			},
		},
		ExpiresAt: time.Now().Add(10 * time.Hour).Unix(),
	}

	secondResp, err := suite.walletPlugin.NewAttenuatedToken(secondDelegation)
	suite.NoError(err, "Should create second delegation")
	suite.NotEmpty(secondResp.Token)

	// Parse and validate chain
	finalToken, err := ucan.ParseToken(secondResp.Token)
	suite.NoError(err, "Should parse final token")
	suite.Len(finalToken.Proofs, 2, "Should have full proof chain")

	// Verify attenuation is properly reduced
	suite.Len(finalToken.Attenuations, 1, "Should have one attenuation")
	capability := finalToken.Attenuations[0].Capability
	suite.NotNil(capability, "Should have capability")
}

// TestE2EWalletRecovery tests wallet recovery from EnclaveData
func (suite *UCANWalletE2ETestSuite) TestE2EWalletRecovery() {
	// Get original issuer info
	originalIssuer, err := suite.walletPlugin.GetIssuerDID()
	suite.NoError(err)

	// Create a token with original wallet
	tokenReq := &plugin.NewOriginTokenRequest{
		AudienceDID: "did:sonr:recovery-test",
		Attenuations: []map[string]any{
			{"can": []string{"test"}, "with": "test://resource"},
		},
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	originalResp, err := suite.walletPlugin.NewOriginToken(tokenReq)
	suite.NoError(err)
	suite.NotEmpty(originalResp.Token)

	// Simulate wallet recovery - create new plugin with same enclave data
	enclaveBytes, err := json.Marshal(suite.enclaveData)
	suite.NoError(err)

	recoveredPlugin, err := plugin.LoadPluginWithEnclave(
		suite.ctx,
		suite.chainID,
		enclaveBytes,
		suite.vaultConfig,
	)
	suite.NoError(err, "Should recover wallet from enclave data")

	// Verify recovered wallet has same identity
	recoveredIssuer, err := recoveredPlugin.GetIssuerDID()
	suite.NoError(err)
	suite.Equal(originalIssuer.DID, recoveredIssuer.DID, "Recovered DID should match")
	suite.Equal(originalIssuer.Address, recoveredIssuer.Address, "Recovered address should match")

	// Verify recovered wallet can create compatible tokens
	recoveredResp, err := recoveredPlugin.NewOriginToken(tokenReq)
	suite.NoError(err)
	suite.NotEmpty(recoveredResp.Token)
	suite.Equal(originalIssuer.DID, recoveredResp.Issuer, "Issuer should be consistent")
}

// TestE2EMultiPartyWallet tests multi-party wallet operations
func (suite *UCANWalletE2ETestSuite) TestE2EMultiPartyWallet() {
	// Configure multi-party vault
	multiPartyConfig := map[string]any{
		"enclave_type": "mpc",
		"threshold":    2,
		"parties":      3,
		"party_ids":    []string{"party-1", "party-2", "party-3"},
		"network":      "testnet",
	}

	// Create wallet with multi-party config
	enclaveBytes, err := json.Marshal(suite.enclaveData)
	suite.NoError(err)

	multiPartyPlugin, err := plugin.LoadPluginWithEnclave(
		suite.ctx,
		suite.chainID,
		enclaveBytes,
		multiPartyConfig,
	)
	suite.NoError(err, "Should create multi-party wallet")

	// Create token requiring multi-party consensus
	consensusReq := &plugin.NewOriginTokenRequest{
		AudienceDID: "did:sonr:multi-party-service",
		Attenuations: []map[string]any{
			{
				"can":  []string{"vault/admin", "vault/transfer"},
				"with": "vault://multi-party-vault",
			},
		},
		Facts: []string{
			"type=multi-party",
			"threshold=2/3",
		},
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	resp, err := multiPartyPlugin.NewOriginToken(consensusReq)
	suite.NoError(err, "Should create multi-party token")
	suite.NotEmpty(resp.Token, "Multi-party token should be generated")

	// Verify token contains multi-party facts
	parsedToken, err := ucan.ParseToken(resp.Token)
	suite.NoError(err)
	suite.NotEmpty(parsedToken.Facts, "Should have facts")
}

// TestE2EPermissionBoundaries tests UCAN permission boundaries
func (suite *UCANWalletE2ETestSuite) TestE2EPermissionBoundaries() {
	// Create token with specific permissions
	boundedReq := &plugin.NewOriginTokenRequest{
		AudienceDID: "did:sonr:bounded-service",
		Attenuations: []map[string]any{
			{
				"can":  []string{"vault/read"},
				"with": "vault://user-vault/documents/*",
			},
			{
				"can":  []string{"dwn/write"},
				"with": "dwn://user-records/profile",
			},
		},
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}

	resp, err := suite.walletPlugin.NewOriginToken(boundedReq)
	suite.NoError(err)
	suite.NotEmpty(resp.Token)

	// Parse and verify permissions
	token, err := ucan.ParseToken(resp.Token)
	suite.NoError(err)
	suite.Len(token.Attenuations, 2, "Should have two permission sets")

	// Verify permission boundaries are maintained
	for _, att := range token.Attenuations {
		capability := att.Capability
		suite.NotNil(capability)

		actions := capability.GetActions()
		suite.NotEmpty(actions, "Should have actions")

		// Verify no escalation beyond original permissions
		for _, action := range actions {
			suite.Contains([]string{"vault/read", "dwn/write"}, action,
				"Action should be within bounded permissions")
		}
	}
}

// TestE2ETokenExpiration tests UCAN token expiration
func (suite *UCANWalletE2ETestSuite) TestE2ETokenExpiration() {
	// Create token with very short expiration
	shortLivedReq := &plugin.NewOriginTokenRequest{
		AudienceDID: "did:sonr:short-lived",
		Attenuations: []map[string]any{
			{"can": []string{"test"}, "with": "test://resource"},
		},
		ExpiresAt: time.Now().Add(100 * time.Millisecond).Unix(),
	}

	resp, err := suite.walletPlugin.NewOriginToken(shortLivedReq)
	suite.NoError(err)
	suite.NotEmpty(resp.Token)

	// Parse immediately - should be valid
	token, err := ucan.ParseToken(resp.Token)
	suite.NoError(err)
	suite.NotNil(token)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Verify token is expired
	now := time.Now().Unix()
	suite.Less(token.ExpiresAt, now, "Token should be expired")
}

// TestE2EConcurrentWalletOperations tests concurrent wallet operations
func (suite *UCANWalletE2ETestSuite) TestE2EConcurrentWalletOperations() {
	const numOperations = 10
	results := make(chan *plugin.UCANTokenResponse, numOperations)
	errors := make(chan error, numOperations)

	// Launch concurrent token creation
	for i := 0; i < numOperations; i++ {
		go func(index int) {
			req := &plugin.NewOriginTokenRequest{
				AudienceDID: fmt.Sprintf("did:sonr:concurrent-%d", index),
				Attenuations: []map[string]any{
					{"can": []string{"test"}, "with": fmt.Sprintf("test://resource-%d", index)},
				},
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
			}

			resp, err := suite.walletPlugin.NewOriginToken(req)
			if err != nil {
				errors <- err
			} else {
				results <- resp
			}
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numOperations; i++ {
		select {
		case resp := <-results:
			suite.NotEmpty(resp.Token, "Concurrent token should be generated")
			suite.Equal(suite.issuerDID, resp.Issuer, "Issuer should be consistent")
			successCount++
		case err := <-errors:
			suite.NoError(err, "Concurrent operation should not fail")
		case <-time.After(5 * time.Second):
			suite.Fail("Concurrent operation timed out")
		}
	}

	suite.Equal(numOperations, successCount, "All concurrent operations should succeed")
}

func TestUCANWalletE2ESuite(t *testing.T) {
	suite.Run(t, new(UCANWalletE2ETestSuite))
}
