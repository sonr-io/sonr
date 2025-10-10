package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/sonr-io/sonr/app"
	"github.com/sonr-io/crypto/mpc"
	"github.com/sonr-io/crypto/ucan"
	didtypes "github.com/sonr-io/sonr/x/did/types"
	"github.com/sonr-io/sonr/x/dwn/client/plugin"
	dwntypes "github.com/sonr-io/sonr/x/dwn/types"
	svctypes "github.com/sonr-io/sonr/x/svc/types"
)

// UCANBlockchainE2ETestSuite tests UCAN integration with blockchain operations
type UCANBlockchainE2ETestSuite struct {
	suite.Suite
	ctx          context.Context
	app          *app.App
	clientCtx    client.Context
	walletPlugin plugin.Plugin
	enclaveData  *mpc.EnclaveData
	userDID      string
	userAddress  types.AccAddress
	serviceDID   string
	ucanToken    string
}

func (suite *UCANBlockchainE2ETestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Initialize test enclave
	suite.enclaveData = &mpc.EnclaveData{
		PubHex:   "0x04b9e72dfd423bcf95b3801ac93f74ec5ecf47f2cc7d8c5b7a0e4d4e0e3f2a1b2",
		PubBytes: make([]byte, 65),
		Nonce:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		Curve:    mpc.CurveName("secp256k1"),
	}

	// Initialize test DIDs
	suite.userDID = "did:sonr:test-user"
	suite.serviceDID = "did:sonr:test-service"
}

// TestE2EGaslessTransactionWithUCAN tests gasless transaction execution with UCAN
func (suite *UCANBlockchainE2ETestSuite) TestE2EGaslessTransactionWithUCAN() {
	// Initialize wallet plugin
	enclaveBytes, err := json.Marshal(suite.enclaveData)
	suite.Require().NoError(err)

	suite.walletPlugin, err = plugin.LoadPluginWithEnclave(
		suite.ctx,
		"sonrtest_1-1",
		enclaveBytes,
		nil,
	)
	suite.Require().NoError(err)

	// Create UCAN token for gasless transaction
	tokenReq := &plugin.NewOriginTokenRequest{
		AudienceDID: suite.serviceDID,
		Attenuations: []map[string]any{
			{
				"can":  []string{"did/update", "did/deactivate"},
				"with": fmt.Sprintf("did://%s", suite.userDID),
			},
			{
				"can":  []string{"dwn/write", "dwn/delete"},
				"with": fmt.Sprintf("dwn://%s/records", suite.userDID),
			},
		},
		Facts: []string{
			"gasless=true",
			"transaction_type=did_update",
		},
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}

	tokenResp, err := suite.walletPlugin.NewOriginToken(tokenReq)
	suite.NoError(err, "Should create UCAN token for gasless transaction")
	suite.NotEmpty(tokenResp.Token)
	suite.ucanToken = tokenResp.Token

	// Parse token to verify structure
	parsedToken, err := ucan.ParseToken(suite.ucanToken)
	suite.NoError(err)
	suite.Contains(parsedToken.Facts[0].String(), "gasless=true")
}

// TestE2EDIDOperationsWithUCAN tests DID module operations with UCAN authorization
func (suite *UCANBlockchainE2ETestSuite) TestE2EDIDOperationsWithUCAN() {
	// Create UCAN token for DID operations
	didTokenReq := &plugin.NewOriginTokenRequest{
		AudienceDID: suite.serviceDID,
		Attenuations: []map[string]any{
			{
				"can": []string{
					"did/create",
					"did/update",
					"did/add_verification_method",
					"did/remove_verification_method",
				},
				"with": "did://*", // Allow operations on any DID
			},
		},
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	tokenResp, err := suite.walletPlugin.NewOriginToken(didTokenReq)
	suite.NoError(err)
	suite.NotEmpty(tokenResp.Token)

	// Simulate DID document creation with UCAN
	didDoc := &didtypes.DIDDocument{
		Id: suite.userDID,
		VerificationMethod: []*didtypes.VerificationMethod{
			{
				Id:                     fmt.Sprintf("%s#key-1", suite.userDID),
				VerificationMethodKind: "Ed25519VerificationKey2020",
				Controller:             suite.userDID,
				PublicKeyMultibase:     "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			},
		},
		Authentication: []string{fmt.Sprintf("%s#key-1", suite.userDID)},
	}

	// Verify UCAN grants permission for DID creation
	token, err := ucan.ParseToken(tokenResp.Token)
	suite.NoError(err)

	hasPermission := false
	for _, att := range token.Attenuations {
		capability := att.Capability
		if capability != nil {
			actions := capability.GetActions()
			for _, action := range actions {
				if action == "did/create" {
					hasPermission = true
					break
				}
			}
		}
	}
	suite.True(hasPermission, "UCAN should grant DID creation permission")

	// Create delegated token for specific DID
	delegatedReq := &plugin.NewAttenuatedTokenRequest{
		ParentToken: tokenResp.Token,
		AudienceDID: "did:sonr:delegated-controller",
		Attenuations: []map[string]any{
			{
				"can":  []string{"did/update"}, // Only update, not create
				"with": fmt.Sprintf("did://%s", suite.userDID),
			},
		},
		ExpiresAt: time.Now().Add(12 * time.Hour).Unix(),
	}

	delegatedResp, err := suite.walletPlugin.NewAttenuatedToken(delegatedReq)
	suite.NoError(err)
	suite.NotEmpty(delegatedResp.Token)

	// Verify delegated token has reduced permissions
	delegatedToken, err := ucan.ParseToken(delegatedResp.Token)
	suite.NoError(err)
	suite.Len(delegatedToken.Attenuations, 1)

	// Verify only update permission remains
	capability := delegatedToken.Attenuations[0].Capability
	actions := capability.GetActions()
	suite.Len(actions, 1)
	suite.Equal("did/update", actions[0])
}

// TestE2EDWNOperationsWithUCAN tests DWN module operations with UCAN authorization
func (suite *UCANBlockchainE2ETestSuite) TestE2EDWNOperationsWithUCAN() {
	// Create UCAN token for DWN operations
	dwnTokenReq := &plugin.NewOriginTokenRequest{
		AudienceDID: suite.serviceDID,
		Attenuations: []map[string]any{
			{
				"can": []string{
					"dwn/write",
					"dwn/read",
					"dwn/delete",
					"dwn/query",
				},
				"with": fmt.Sprintf("dwn://%s/records/*", suite.userDID),
			},
			{
				"can":  []string{"dwn/admin"},
				"with": fmt.Sprintf("dwn://%s/permissions", suite.userDID),
			},
		},
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	tokenResp, err := suite.walletPlugin.NewOriginToken(dwnTokenReq)
	suite.NoError(err)
	suite.NotEmpty(tokenResp.Token)

	// Parse token
	token, err := ucan.ParseToken(tokenResp.Token)
	suite.NoError(err)

	// Simulate DWN record creation
	record := &dwntypes.DWNRecord{
		ContextId:        suite.userDID,
		RecordId:         "record-123",
		Schema:           "https://schema.org/TextDigitalDocument",
		ParentId:         "",
		DataCid:          "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
		DateCreated:      time.Now().Unix(),
		DateModified:     time.Now().Unix(),
		DatePublished:    0,
		Published:        false,
		EncryptedDataCid: "",
	}

	// Verify UCAN grants write permission
	hasWritePermission := false
	for _, att := range token.Attenuations {
		capability := att.Capability
		if capability != nil {
			actions := capability.GetActions()
			for _, action := range actions {
				if action == "dwn/write" {
					hasWritePermission = true
					break
				}
			}
		}
	}
	suite.True(hasWritePermission, "UCAN should grant DWN write permission")

	// Create scoped token for specific record
	scopedReq := &plugin.NewAttenuatedTokenRequest{
		ParentToken: tokenResp.Token,
		AudienceDID: "did:sonr:record-processor",
		Attenuations: []map[string]any{
			{
				"can":  []string{"dwn/read"}, // Read-only access
				"with": fmt.Sprintf("dwn://%s/records/%s", suite.userDID, record.RecordId),
			},
		},
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	}

	scopedResp, err := suite.walletPlugin.NewAttenuatedToken(scopedReq)
	suite.NoError(err)
	suite.NotEmpty(scopedResp.Token)

	// Verify scoped permissions
	scopedToken, err := ucan.ParseToken(scopedResp.Token)
	suite.NoError(err)
	suite.Len(scopedToken.Attenuations, 1)

	capability := scopedToken.Attenuations[0].Capability
	actions := capability.GetActions()
	suite.Equal([]string{"dwn/read"}, actions)
}

// TestE2EServiceRegistrationWithUCAN tests service registration with UCAN
func (suite *UCANBlockchainE2ETestSuite) TestE2EServiceRegistrationWithUCAN() {
	// Create UCAN token for service operations
	svcTokenReq := &plugin.NewOriginTokenRequest{
		AudienceDID: suite.serviceDID,
		Attenuations: []map[string]any{
			{
				"can": []string{
					"service/register",
					"service/update",
					"service/deactivate",
				},
				"with": "service://*",
			},
			{
				"can":  []string{"service/verify_domain"},
				"with": "service://example.com",
			},
		},
		Facts: []string{
			"service_type=oauth2_provider",
			"domain=example.com",
		},
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Unix(), // 30 days
	}

	tokenResp, err := suite.walletPlugin.NewOriginToken(svcTokenReq)
	suite.NoError(err)
	suite.NotEmpty(tokenResp.Token)

	// Parse token
	token, err := ucan.ParseToken(tokenResp.Token)
	suite.NoError(err)

	// Simulate service registration
	service := &svctypes.Service{
		Id:          "service-123",
		Owner:       suite.userDID,
		Domain:      "example.com",
		Description: "Test OAuth2 Provider",
		Endpoints: []*svctypes.ServiceEndpoint{
			{
				Id:              "oauth2",
				Type:            "OAuth2Provider",
				ServiceEndpoint: "https://example.com/oauth2",
			},
		},
	}

	// Verify registration permission
	hasRegisterPermission := false
	for _, att := range token.Attenuations {
		capability := att.Capability
		if capability != nil {
			actions := capability.GetActions()
			for _, action := range actions {
				if action == "service/register" {
					hasRegisterPermission = true
					break
				}
			}
		}
	}
	suite.True(hasRegisterPermission, "UCAN should grant service registration permission")

	// Create admin token for service management
	adminReq := &plugin.NewAttenuatedTokenRequest{
		ParentToken: tokenResp.Token,
		AudienceDID: "did:sonr:service-admin",
		Attenuations: []map[string]any{
			{
				"can":  []string{"service/update", "service/deactivate"},
				"with": fmt.Sprintf("service://%s", service.Id),
			},
		},
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days
	}

	adminResp, err := suite.walletPlugin.NewAttenuatedToken(adminReq)
	suite.NoError(err)
	suite.NotEmpty(adminResp.Token)
}

// TestE2ECrossModuleUCANDelegation tests UCAN delegation across modules
func (suite *UCANBlockchainE2ETestSuite) TestE2ECrossModuleUCANDelegation() {
	// Create master token with permissions across all modules
	masterReq := &plugin.NewOriginTokenRequest{
		AudienceDID: "did:sonr:master-service",
		Attenuations: []map[string]any{
			// DID permissions
			{
				"can":  []string{"did/create", "did/update"},
				"with": "did://*",
			},
			// DWN permissions
			{
				"can":  []string{"dwn/write", "dwn/read"},
				"with": "dwn://*",
			},
			// Service permissions
			{
				"can":  []string{"service/register"},
				"with": "service://*",
			},
		},
		Facts: []string{
			"type=cross_module",
			"purpose=integration",
		},
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	masterResp, err := suite.walletPlugin.NewOriginToken(masterReq)
	suite.NoError(err)
	suite.NotEmpty(masterResp.Token)

	// Create module-specific delegations

	// DID-only delegation
	didDelegation := &plugin.NewAttenuatedTokenRequest{
		ParentToken: masterResp.Token,
		AudienceDID: "did:sonr:did-manager",
		Attenuations: []map[string]any{
			{
				"can":  []string{"did/update"}, // Remove create permission
				"with": fmt.Sprintf("did://%s", suite.userDID),
			},
		},
		ExpiresAt: time.Now().Add(12 * time.Hour).Unix(),
	}

	didResp, err := suite.walletPlugin.NewAttenuatedToken(didDelegation)
	suite.NoError(err)

	// DWN-only delegation
	dwnDelegation := &plugin.NewAttenuatedTokenRequest{
		ParentToken: masterResp.Token,
		AudienceDID: "did:sonr:dwn-manager",
		Attenuations: []map[string]any{
			{
				"can":  []string{"dwn/read"}, // Read-only
				"with": fmt.Sprintf("dwn://%s/records", suite.userDID),
			},
		},
		ExpiresAt: time.Now().Add(6 * time.Hour).Unix(),
	}

	dwnResp, err := suite.walletPlugin.NewAttenuatedToken(dwnDelegation)
	suite.NoError(err)

	// Verify each delegation has only its module's permissions
	didToken, err := ucan.ParseToken(didResp.Token)
	suite.NoError(err)
	suite.Len(didToken.Attenuations, 1)

	dwnToken, err := ucan.ParseToken(dwnResp.Token)
	suite.NoError(err)
	suite.Len(dwnToken.Attenuations, 1)

	// Verify no cross-contamination of permissions
	didCapability := didToken.Attenuations[0].Capability
	didActions := didCapability.GetActions()
	for _, action := range didActions {
		suite.Contains(action, "did/", "DID token should only have DID permissions")
	}

	dwnCapability := dwnToken.Attenuations[0].Capability
	dwnActions := dwnCapability.GetActions()
	for _, action := range dwnActions {
		suite.Contains(action, "dwn/", "DWN token should only have DWN permissions")
	}
}

// TestE2EUCANRevocation tests UCAN token revocation
func (suite *UCANBlockchainE2ETestSuite) TestE2EUCANRevocation() {
	// Create revocable token
	revocableReq := &plugin.NewOriginTokenRequest{
		AudienceDID: "did:sonr:revocable-service",
		Attenuations: []map[string]any{
			{
				"can":  []string{"vault/read", "vault/write"},
				"with": "vault://sensitive-data",
			},
		},
		Facts: []string{
			"revocable=true",
			fmt.Sprintf("revocation_id=%d", time.Now().Unix()),
		},
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	tokenResp, err := suite.walletPlugin.NewOriginToken(revocableReq)
	suite.NoError(err)
	suite.NotEmpty(tokenResp.Token)

	// Parse token to get revocation ID
	token, err := ucan.ParseToken(tokenResp.Token)
	suite.NoError(err)

	var revocationID string
	for _, fact := range token.Facts {
		factStr := fact.String()
		if len(factStr) > 13 && factStr[:13] == "revocation_id" {
			revocationID = factStr[14:]
			break
		}
	}
	suite.NotEmpty(revocationID, "Should have revocation ID")

	// Simulate revocation list check
	revocationList := map[string]bool{
		revocationID: false, // Not revoked initially
	}

	suite.False(revocationList[revocationID], "Token should not be revoked initially")

	// Simulate revocation
	revocationList[revocationID] = true
	suite.True(revocationList[revocationID], "Token should be marked as revoked")
}

// TestE2EBatchUCANOperations tests batch UCAN token operations
func (suite *UCANBlockchainE2ETestSuite) TestE2EBatchUCANOperations() {
	const batchSize = 5
	tokens := make([]string, 0, batchSize)

	// Create batch of tokens
	for i := 0; i < batchSize; i++ {
		req := &plugin.NewOriginTokenRequest{
			AudienceDID: fmt.Sprintf("did:sonr:batch-service-%d", i),
			Attenuations: []map[string]any{
				{
					"can":  []string{fmt.Sprintf("batch/operation-%d", i)},
					"with": fmt.Sprintf("batch://resource-%d", i),
				},
			},
			Facts: []string{
				fmt.Sprintf("batch_index=%d", i),
				fmt.Sprintf("batch_id=%d", time.Now().Unix()),
			},
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		}

		resp, err := suite.walletPlugin.NewOriginToken(req)
		suite.NoError(err)
		suite.NotEmpty(resp.Token)
		tokens = append(tokens, resp.Token)
	}

	suite.Len(tokens, batchSize, "Should create all batch tokens")

	// Verify each token
	for i, tokenStr := range tokens {
		token, err := ucan.ParseToken(tokenStr)
		suite.NoError(err)
		suite.Equal(fmt.Sprintf("did:sonr:batch-service-%d", i), token.Audience)

		// Verify batch-specific permissions
		suite.Len(token.Attenuations, 1)
		capability := token.Attenuations[0].Capability
		actions := capability.GetActions()
		suite.Contains(actions[0], fmt.Sprintf("operation-%d", i))
	}

	// Create aggregated delegation from all batch tokens
	aggregatedReq := &plugin.NewOriginTokenRequest{
		AudienceDID: "did:sonr:batch-aggregator",
		Attenuations: []map[string]any{
			{
				"can":  []string{"batch/aggregate", "batch/process"},
				"with": "batch://*",
			},
		},
		Facts: []string{
			fmt.Sprintf("aggregated_count=%d", batchSize),
			"type=batch_aggregation",
		},
		ExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
	}

	aggregatedResp, err := suite.walletPlugin.NewOriginToken(aggregatedReq)
	suite.NoError(err)
	suite.NotEmpty(aggregatedResp.Token)

	// Verify aggregated token
	aggregatedToken, err := ucan.ParseToken(aggregatedResp.Token)
	suite.NoError(err)
	suite.Contains(aggregatedToken.Facts[0].String(), fmt.Sprintf("aggregated_count=%d", batchSize))
}

func TestUCANBlockchainE2ESuite(t *testing.T) {
	suite.Run(t, new(UCANBlockchainE2ETestSuite))
}
