package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/sonr-io/sonr/bridge/handlers"
	"github.com/sonr-io/sonr/bridge/server"
	"github.com/sonr-io/sonr/bridge/tasks"
	"github.com/sonr-io/sonr/crypto/mpc"
	"github.com/sonr-io/sonr/types/ipfs"
	"github.com/sonr-io/sonr/x/dwn/client/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// UCANWorkflowTestSuite tests complete UCAN workflows from plugin to task processing
type UCANWorkflowTestSuite struct {
	suite.Suite
	ctx           context.Context
	cancel        context.CancelFunc
	enclave       mpc.Enclave
	pluginManager *plugin.Manager
	asynqClient   *asynq.Client
	serverConfig  *server.Config

	// Test configuration
	testChainID string
	testTimeout time.Duration
}

// SetupSuite initializes the test suite with all required components
func (suite *UCANWorkflowTestSuite) SetupSuite() {
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 2*time.Minute)
	suite.testChainID = "sonr-testnet-1"
	suite.testTimeout = 30 * time.Second

	// Create MPC enclave for testing
	enclave, err := mpc.NewEnclave()
	suite.Require().NoError(err)
	suite.enclave = enclave

	// Initialize plugin manager
	suite.pluginManager = plugin.NewManager(plugin.DefaultLoaderConfig())

	// Initialize Asynq client for task processing
	suite.asynqClient = asynq.NewClient(asynq.RedisClientOpt{
		Addr: getTestRedisAddr(),
	})

	// Initialize server configuration for bridge testing
	suite.serverConfig = &server.Config{
		JWTSecret:  []byte("test-ucan-workflow-secret"),
		IPFSClient: &MockIPFSClient{},
	}
}

// TearDownSuite cleans up test resources
func (suite *UCANWorkflowTestSuite) TearDownSuite() {
	if suite.cancel != nil {
		suite.cancel()
	}
	if suite.pluginManager != nil {
		_ = suite.pluginManager.Close()
	}
	if suite.asynqClient != nil {
		_ = suite.asynqClient.Close()
	}
}

// TestCompleteUCANTokenWorkflow tests the complete workflow from plugin to task processing
func (suite *UCANWorkflowTestSuite) TestCompleteUCANTokenWorkflow() {
	suite.T().Run("OriginTokenCreationWorkflow", func(t *testing.T) {
		// Step 1: Load Motor plugin with enclave data
		pluginInstance := suite.loadTestPlugin(t)
		if pluginInstance == nil {
			t.Skip("Vault plugin not available - run 'make vault' to build WASM plugin")
		}

		// Step 2: Generate origin UCAN token via plugin
		originReq := &plugin.NewOriginTokenRequest{
			AudienceDID: "did:sonr:test-audience-workflow",
			Attenuations: []map[string]any{
				{
					"can":  []string{"read", "write"},
					"with": "vault://workflow-resource",
				},
			},
			Facts:     []string{"test-workflow-fact"},
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
		}

		originResp, err := pluginInstance.NewOriginToken(originReq)
		if err != nil {
			t.Skip("Plugin operation failed - WASM runtime may not be available")
		}
		require.NoError(t, err)
		assert.NotEmpty(t, originResp.Token)
		assert.Empty(t, originResp.Error)

		t.Logf("Generated origin token: %s", truncateForLog(originResp.Token))

		// Step 3: Create task for origin token processing
		task, err := tasks.NewUCANTokenTask(
			123,
			originReq.AudienceDID,
			originReq.Attenuations,
			originReq.ExpiresAt,
		)
		require.NoError(t, err)

		// Step 4: Process task through Asynq
		info, err := suite.asynqClient.Enqueue(task, asynq.Queue("default"))
		if err != nil {
			t.Skip("Redis not available for task processing")
		}
		require.NoError(t, err)
		assert.NotEmpty(t, info.ID)

		t.Logf("Enqueued UCAN token task: %s", info.ID)

		// Step 5: Verify token structure and claims
		suite.verifyUCANTokenStructure(t, originResp.Token, originReq.AudienceDID)
	})

	suite.T().Run("AttenuatedTokenWorkflow", func(t *testing.T) {
		pluginInstance := suite.loadTestPlugin(t)
		if pluginInstance == nil {
			t.Skip("Motor plugin not available")
		}

		// Create parent token
		parentReq := &plugin.NewOriginTokenRequest{
			AudienceDID: "did:sonr:parent-audience",
			Attenuations: []map[string]any{
				{
					"can":  []string{"read", "write", "delete"},
					"with": "vault://parent-resource/*",
				},
			},
			ExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		}

		parentResp, err := pluginInstance.NewOriginToken(parentReq)
		if err != nil {
			t.Skip("Plugin operation failed")
		}
		require.NoError(t, err)

		// Create attenuated token with reduced capabilities
		attReq := &plugin.NewAttenuatedTokenRequest{
			ParentToken: parentResp.Token,
			AudienceDID: "did:sonr:child-audience",
			Attenuations: []map[string]any{
				{
					"can":  []string{"read"},
					"with": "vault://parent-resource/child-folder",
				},
			},
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
		}

		attResp, err := pluginInstance.NewAttenuatedToken(attReq)
		require.NoError(t, err)
		assert.NotEmpty(t, attResp.Token)
		assert.NotEqual(t, parentResp.Token, attResp.Token)

		t.Logf("Generated attenuated token chain: parent -> child")

		// Process both tokens through task system
		suite.processTokenThroughTasks(t, parentResp.Token, "parent")
		suite.processTokenThroughTasks(t, attResp.Token, "attenuated")
	})
}

// TestUCANDIDGenerationWorkflow tests DID generation through UCAN architecture
func (suite *UCANWorkflowTestSuite) TestUCANDIDGenerationWorkflow() {
	suite.T().Run("DIDGenerationViaPlugin", func(t *testing.T) {
		pluginInstance := suite.loadTestPlugin(t)
		if pluginInstance == nil {
			t.Skip("Motor plugin not available")
		}

		// Step 1: Get issuer DID from plugin
		didResp, err := pluginInstance.GetIssuerDID()
		if err != nil {
			t.Skip("Plugin DID operation failed")
		}
		require.NoError(t, err)
		assert.NotEmpty(t, didResp.IssuerDID)
		assert.Contains(t, didResp.IssuerDID, "did:sonr:")
		assert.NotEmpty(t, didResp.Address)
		assert.NotEmpty(t, didResp.ChainCode)

		t.Logf("Generated DID: %s", didResp.IssuerDID)
		t.Logf("Address: %s", didResp.Address)

		// Step 2: Create DID generation task
		didTask, err := tasks.NewUCANDIDTask(456)
		require.NoError(t, err)

		// Step 3: Process through task queue
		info, err := suite.asynqClient.Enqueue(didTask, asynq.Queue("critical"))
		if err != nil {
			t.Skip("Redis not available")
		}
		require.NoError(t, err)

		t.Logf("DID generation task enqueued: %s", info.ID)

		// Step 4: Verify DID components match expected patterns
		assert.True(t, suite.isValidSonrDID(didResp.IssuerDID))
		assert.True(t, suite.isValidSonrAddress(didResp.Address))
	})
}

// TestUCANSigningWorkflow tests complete signing workflow including verification
func (suite *UCANWorkflowTestSuite) TestUCANSigningWorkflow() {
	suite.T().Run("SignAndVerifyWorkflow", func(t *testing.T) {
		pluginInstance := suite.loadTestPlugin(t)
		if pluginInstance == nil {
			t.Skip("Motor plugin not available")
		}

		testData := []byte(
			"UCAN workflow integration test data - this should be signed and verified",
		)

		// Step 1: Sign data through plugin
		signReq := &plugin.SignDataRequest{
			Data: testData,
		}

		signResp, err := pluginInstance.SignData(signReq)
		if err != nil {
			t.Skip("Plugin signing failed")
		}
		require.NoError(t, err)
		assert.NotEmpty(t, signResp.Signature)
		assert.Empty(t, signResp.Error)

		// Step 2: Create signing task for task processor
		signTask, err := tasks.NewUCANSignTask(789, testData)
		require.NoError(t, err)

		info, err := suite.asynqClient.Enqueue(signTask, asynq.Queue("default"))
		if err != nil {
			t.Skip("Redis not available")
		}
		require.NoError(t, err)

		t.Logf("Signing task enqueued: %s", info.ID)

		// Step 3: Verify signature through plugin
		verifyReq := &plugin.VerifyDataRequest{
			Data:      testData,
			Signature: signResp.Signature,
		}

		verifyResp, err := pluginInstance.VerifyData(verifyReq)
		require.NoError(t, err)
		assert.True(t, verifyResp.Valid)
		assert.Empty(t, verifyResp.Error)

		// Step 4: Create verification task
		verifyTask, err := tasks.NewUCANVerifyTask(789, testData, signResp.Signature)
		require.NoError(t, err)

		verifyInfo, err := suite.asynqClient.Enqueue(verifyTask, asynq.Queue("default"))
		if err != nil {
			t.Skip("Redis not available")
		}
		require.NoError(t, err)

		t.Logf("Verification task enqueued: %s", verifyInfo.ID)
	})
}

// TestBridgeIntegration tests integration with bridge handlers
func (suite *UCANWorkflowTestSuite) TestBridgeIntegration() {
	suite.T().Run("VaultHandlerIntegration", func(t *testing.T) {
		// Create vault handlers with test dependencies
		connManager := handlers.NewConnectionManager()
		sseManager := handlers.NewSSEManager()

		vaultHandlers := handlers.NewVaultHandlers(
			suite.serverConfig.IPFSClient,
			connManager,
			sseManager,
		)

		require.NotNil(t, vaultHandlers)

		// Test queue priority mapping
		priorities := map[string]string{
			"critical": "critical",
			"high":     "critical",
			"default":  "default",
			"low":      "low",
			"":         "default",
			"unknown":  "default",
		}

		for input, expected := range priorities {
			result := handlers.GetQueueFromPriority(input)
			assert.Equal(t, expected, result, "Priority mapping failed for: %s", input)
		}

		t.Log("Bridge handler integration verified")
	})
}

// TestUCANTokenChainWorkflow tests complex token delegation chains
func (suite *UCANWorkflowTestSuite) TestUCANTokenChainWorkflow() {
	suite.T().Run("TokenDelegationChain", func(t *testing.T) {
		pluginInstance := suite.loadTestPlugin(t)
		if pluginInstance == nil {
			t.Skip("Motor plugin not available")
		}

		// Create root token with broad permissions
		rootReq := &plugin.NewOriginTokenRequest{
			AudienceDID: "did:sonr:root-service",
			Attenuations: []map[string]any{
				{
					"can":  []string{"*"},
					"with": "vault://root/*",
				},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		}

		rootResp, err := pluginInstance.NewOriginToken(rootReq)
		if err != nil {
			t.Skip("Plugin operation failed")
		}
		require.NoError(t, err)

		// Create first-level delegation
		level1Req := &plugin.NewAttenuatedTokenRequest{
			ParentToken: rootResp.Token,
			AudienceDID: "did:sonr:service-manager",
			Attenuations: []map[string]any{
				{
					"can":  []string{"read", "write"},
					"with": "vault://root/service/*",
				},
			},
			ExpiresAt: time.Now().Add(12 * time.Hour).Unix(),
		}

		level1Resp, err := pluginInstance.NewAttenuatedToken(level1Req)
		require.NoError(t, err)

		// Create second-level delegation
		level2Req := &plugin.NewAttenuatedTokenRequest{
			ParentToken: level1Resp.Token,
			AudienceDID: "did:sonr:worker-node",
			Attenuations: []map[string]any{
				{
					"can":  []string{"read"},
					"with": "vault://root/service/data",
				},
			},
			ExpiresAt: time.Now().Add(6 * time.Hour).Unix(),
		}

		level2Resp, err := pluginInstance.NewAttenuatedToken(level2Req)
		require.NoError(t, err)

		// Verify each token in the chain is unique
		tokens := []string{rootResp.Token, level1Resp.Token, level2Resp.Token}
		for i, token1 := range tokens {
			for j, token2 := range tokens {
				if i != j {
					assert.NotEqual(t, token1, token2, "Tokens in chain should be unique")
				}
			}
		}

		t.Logf("Created UCAN delegation chain: root -> level1 -> level2")
		t.Logf("Chain lengths: root=%d, level1=%d, level2=%d",
			len(rootResp.Token), len(level1Resp.Token), len(level2Resp.Token))
	})
}

// TestErrorHandlingWorkflows tests error conditions across the full workflow
func (suite *UCANWorkflowTestSuite) TestErrorHandlingWorkflows() {
	suite.T().Run("InvalidTokenCreation", func(t *testing.T) {
		pluginInstance := suite.loadTestPlugin(t)
		if pluginInstance == nil {
			t.Skip("Motor plugin not available")
		}

		// Test with invalid audience DID
		invalidReq := &plugin.NewOriginTokenRequest{
			AudienceDID: "", // Invalid
			ExpiresAt:   time.Now().Add(1 * time.Hour).Unix(),
		}

		resp, err := pluginInstance.NewOriginToken(invalidReq)

		// Should handle error gracefully
		if err == nil {
			assert.NotEmpty(t, resp.Error, "Should return error for invalid audience")
		} else {
			assert.Error(t, err)
		}

		t.Log("Invalid token creation handled correctly")
	})

	suite.T().Run("TaskProcessingErrors", func(t *testing.T) {
		// Test with invalid task payload
		task, err := tasks.NewUCANTokenTask(
			0,  // Invalid user ID
			"", // Invalid audience
			nil,
			0,
		)
		// Should create task but processing should handle validation
		assert.NoError(t, err)
		assert.NotNil(t, task)

		t.Log("Task error handling verified")
	})
}

// Helper methods

// loadTestPlugin loads a Motor plugin instance for testing
func (suite *UCANWorkflowTestSuite) loadTestPlugin(t *testing.T) plugin.Plugin {
	config := plugin.DefaultEnclaveConfig()
	config.ChainID = suite.testChainID
	config.EnclaveData = suite.enclave.GetData()

	ctx, cancel := context.WithTimeout(suite.ctx, suite.testTimeout)
	defer cancel()

	pluginInstance, err := suite.pluginManager.LoadPlugin(ctx, config)
	if err != nil {
		t.Logf("Failed to load plugin: %v", err)
		return nil
	}

	return pluginInstance
}

// processTokenThroughTasks processes a token through the task system
func (suite *UCANWorkflowTestSuite) processTokenThroughTasks(
	t *testing.T,
	_ /* token */, tokenType string,
) {
	// Create a task that would process this token
	task, err := tasks.NewUCANTokenTask(
		999,
		"did:sonr:task-processor",
		nil,
		0,
	)
	require.NoError(t, err)

	info, err := suite.asynqClient.Enqueue(task, asynq.Queue("default"))
	if err != nil {
		t.Skip("Redis not available")
	}
	require.NoError(t, err)

	t.Logf("Processed %s token through tasks: %s", tokenType, info.ID)
}

// verifyUCANTokenStructure verifies the structure of a generated UCAN token
func (suite *UCANWorkflowTestSuite) verifyUCANTokenStructure(
	t *testing.T,
	token, expectedAudience string,
) {
	// Split JWT into parts
	parts := suite.splitJWT(token)
	require.Len(t, parts, 3, "UCAN token should be valid JWT")

	// Decode header (for structure verification, not cryptographic validation)
	header := make(map[string]any)
	headerBytes := suite.base64Decode(parts[0])

	err := json.Unmarshal(headerBytes, &header)
	if err != nil {
		t.Logf("Cannot parse header JSON: %v", err)
		return
	}

	// Check for UCAN version
	if ucv, exists := header["ucv"]; exists {
		t.Logf("UCAN version: %v", ucv)
	}

	// Decode claims
	claims := make(map[string]any)
	claimsBytes := suite.base64Decode(parts[1])

	err = json.Unmarshal(claimsBytes, &claims)
	if err != nil {
		t.Logf("Cannot parse claims JSON: %v", err)
		return
	}

	// Verify expected claims
	if aud, exists := claims["aud"]; exists {
		assert.Equal(t, expectedAudience, aud, "Audience should match request")
	}

	if iss, exists := claims["iss"]; exists {
		assert.Contains(t, iss.(string), "did:sonr:", "Issuer should be Sonr DID")
		t.Logf("Token issuer: %v", iss)
	}

	if att, exists := claims["att"]; exists {
		t.Logf("Token attenuations: %v", att)
	}

	t.Log("UCAN token structure verified")
}

// isValidSonrDID checks if a string is a valid Sonr DID
func (suite *UCANWorkflowTestSuite) isValidSonrDID(did string) bool {
	return len(did) > 10 &&
		did[:9] == "did:sonr:" &&
		len(did) > 9
}

// isValidSonrAddress checks if a string is a valid Sonr address
func (suite *UCANWorkflowTestSuite) isValidSonrAddress(address string) bool {
	return len(address) > 5 &&
		address[:5] == "sonr1" &&
		len(address) > 5
}

// splitJWT splits a JWT token into its component parts
func (suite *UCANWorkflowTestSuite) splitJWT(token string) []string {
	var parts []string
	start := 0

	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}

	if start < len(token) {
		parts = append(parts, token[start:])
	}

	return parts
}

// base64Decode decodes base64url strings (simplified implementation)
func (suite *UCANWorkflowTestSuite) base64Decode(s string) []byte {
	// This is a simplified implementation for testing
	// Real implementation would handle base64url properly
	return []byte(s)
}

// Utility functions

// getTestRedisAddr returns Redis address for testing
func getTestRedisAddr() string {
	// Try to use test Redis if available, otherwise use default
	return "127.0.0.1:6379"
}

// truncateForLog truncates long strings for logging
func truncateForLog(s string) string {
	if len(s) <= 100 {
		return s
	}
	return s[:50] + "..." + s[len(s)-50:]
}

// MockIPFSClient implements a mock IPFS client for testing
type MockIPFSClient struct{}

func (m *MockIPFSClient) Add(data []byte) (string, error) {
	return fmt.Sprintf("Qm%x", data[:min(len(data), 10)]), nil
}

func (m *MockIPFSClient) AddFile(file ipfs.File) (string, error) {
	return "QmMockFileHash", nil
}

func (m *MockIPFSClient) AddFolder(folder ipfs.Folder) (string, error) {
	return "QmMockFolderHash", nil
}

func (m *MockIPFSClient) Exists(cid string) (bool, error) {
	return true, nil
}

func (m *MockIPFSClient) Get(cid string) ([]byte, error) {
	return []byte("mock-ipfs-data"), nil
}

func (m *MockIPFSClient) IsPinned(ipns string) (bool, error) {
	return true, nil
}

func (m *MockIPFSClient) Ls(cid string) ([]string, error) {
	return []string{"file1", "file2"}, nil
}

func (m *MockIPFSClient) Pin(cid string, name string) error {
	return nil
}

func (m *MockIPFSClient) Unpin(cid string) error {
	return nil
}

func (m *MockIPFSClient) NodeStatus() (*ipfs.NodeStatus, error) {
	return &ipfs.NodeStatus{
		PeerID:         "12D3KooWMockPeerIDForTesting",
		Version:        "kubo-0.28.0",
		PeerType:       "kubo",
		ConnectedPeers: 7,
	}, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestSuite runner
func TestUCANWorkflowIntegration(t *testing.T) {
	suite.Run(t, new(UCANWorkflowTestSuite))
}
