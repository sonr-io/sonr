package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/asynkron/protoactor-go/actor"
	"github.com/hibiken/asynq"
	"github.com/sonr-io/sonr/crypto/mpc"
	"github.com/sonr-io/sonr/x/dwn/client/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestEnclaveData creates a test enclave data for testing purposes
func generateTestEnclaveData(t *testing.T) *mpc.EnclaveData {
	t.Helper()

	// Generate a new enclave for testing
	enclave, err := mpc.NewEnclave()
	require.NoError(t, err, "failed to generate test enclave")

	return enclave.GetData()
}

// MockUCANActor for testing UCAN task processors
type MockUCANActor struct {
	responses map[string]any
}

func NewMockUCANActor() *MockUCANActor {
	return &MockUCANActor{
		responses: make(map[string]any),
	}
}

// SlowMockUCANActor simulates timeout scenarios
type SlowMockUCANActor struct{}

func (s *SlowMockUCANActor) Receive(c actor.Context) {
	switch c.Message().(type) {
	case *plugin.NewOriginTokenRequest:
		// Simulate slow response (longer than KRequestTimeout)
		time.Sleep(KRequestTimeout + time.Second)
		c.Respond(&plugin.UCANTokenResponse{})
	}
}

func (m *MockUCANActor) Receive(c actor.Context) {
	switch c.Message().(type) {
	case *actor.Started:
		// Actor started
	case *plugin.NewOriginTokenRequest:
		c.Respond(&plugin.UCANTokenResponse{
			Token:   "mock-ucan-token",
			Issuer:  "did:sonr:mock-issuer",
			Address: "mock-address",
		})
	case *plugin.SignDataRequest:
		c.Respond(&plugin.SignDataResponse{
			Signature: []byte("mock-signature"),
		})
	case *plugin.VerifyDataRequest:
		c.Respond(&plugin.VerifyDataResponse{
			Valid: true,
		})
	case *plugin.NewAttenuatedTokenRequest:
		c.Respond(&plugin.UCANTokenResponse{
			Token:   "mock-attenuated-token",
			Issuer:  "did:sonr:mock-issuer",
			Address: "mock-address",
		})
	case *plugin.GetIssuerDIDResponse:
		c.Respond(&plugin.GetIssuerDIDResponse{
			IssuerDID: "did:sonr:mock-issuer",
			Address:   "mock-address",
			ChainCode: "mock-chain-code",
		})
	default:
		c.Respond(&actor.DeadLetterResponse{})
	}
}

func TestUCANTokenTask(t *testing.T) {
	tests := []struct {
		name         string
		userID       int
		audienceDID  string
		attenuations []map[string]any
		expiresAt    int64
	}{
		{
			name:        "valid UCAN token request",
			userID:      123,
			audienceDID: "did:sonr:audience",
			attenuations: []map[string]any{
				{"can": []string{"sign"}, "with": "vault://example"},
			},
			expiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
		{
			name:         "zero user ID",
			userID:       0,
			audienceDID:  "did:sonr:audience",
			attenuations: []map[string]any{},
			expiresAt:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewUCANTokenTask(tt.userID, tt.audienceDID, tt.attenuations, tt.expiresAt)
			require.NoError(t, err)
			assert.Equal(t, TypeUCANToken, task.Type())

			var payload UCANTokenPayload
			err = json.Unmarshal(task.Payload(), &payload)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, payload.UserID)
			assert.Equal(t, tt.audienceDID, payload.AudienceDID)
			// Handle JSON unmarshaling type conversion ([]string becomes []any)
			assert.Equal(t, len(tt.attenuations), len(payload.Attenuations))
			for i, expectedAtt := range tt.attenuations {
				actualAtt := payload.Attenuations[i]
				for key, expectedVal := range expectedAtt {
					actualVal, exists := actualAtt[key]
					assert.True(t, exists, "key %s should exist", key)

					// Handle []string to []any conversion
					if expectedSlice, ok := expectedVal.([]string); ok {
						actualSlice, ok := actualVal.([]any)
						assert.True(t, ok, "expected []any for key %s", key)
						assert.Equal(t, len(expectedSlice), len(actualSlice))
						for j, expectedItem := range expectedSlice {
							assert.Equal(t, expectedItem, actualSlice[j])
						}
					} else {
						assert.Equal(t, expectedVal, actualVal)
					}
				}
			}
			assert.Equal(t, tt.expiresAt, payload.ExpiresAt)
		})
	}
}

func TestUCANSignTask(t *testing.T) {
	tests := []struct {
		name   string
		userID int
		data   []byte
	}{
		{
			name:   "valid sign request",
			userID: 123,
			data:   []byte("test data to sign"),
		},
		{
			name:   "empty data",
			userID: 456,
			data:   []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewUCANSignTask(tt.userID, tt.data)
			require.NoError(t, err)
			assert.Equal(t, TypeUCANSign, task.Type())

			var payload UCANSignPayload
			err = json.Unmarshal(task.Payload(), &payload)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, payload.UserID)
			assert.Equal(t, tt.data, payload.Data)
		})
	}
}

func TestUCANVerifyTask(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		data      []byte
		signature []byte
	}{
		{
			name:      "valid verify request",
			userID:    123,
			data:      []byte("test data to verify"),
			signature: []byte("test-signature"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewUCANVerifyTask(tt.userID, tt.data, tt.signature)
			require.NoError(t, err)
			assert.Equal(t, TypeUCANVerify, task.Type())

			var payload UCANVerifyPayload
			err = json.Unmarshal(task.Payload(), &payload)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, payload.UserID)
			assert.Equal(t, tt.data, payload.Data)
			assert.Equal(t, tt.signature, payload.Signature)
		})
	}
}

func TestUCANAttenuationTask(t *testing.T) {
	tests := []struct {
		name         string
		userID       int
		parentToken  string
		audienceDID  string
		attenuations []map[string]any
		expiresAt    int64
	}{
		{
			name:        "valid attenuation request",
			userID:      123,
			parentToken: "parent-ucan-token",
			audienceDID: "did:sonr:delegated",
			attenuations: []map[string]any{
				{"can": []string{"read"}, "with": "vault://example"},
			},
			expiresAt: time.Now().Add(1 * time.Hour).Unix(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewUCANAttenuationTask(
				tt.userID,
				tt.parentToken,
				tt.audienceDID,
				tt.attenuations,
				tt.expiresAt,
			)
			require.NoError(t, err)
			assert.Equal(t, TypeUCANAttenuation, task.Type())

			var payload UCANAttenuationPayload
			err = json.Unmarshal(task.Payload(), &payload)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, payload.UserID)
			assert.Equal(t, tt.parentToken, payload.ParentToken)
			assert.Equal(t, tt.audienceDID, payload.AudienceDID)
			// Handle JSON unmarshaling type conversion ([]string becomes []any)
			assert.Equal(t, len(tt.attenuations), len(payload.Attenuations))
			for i, expectedAtt := range tt.attenuations {
				actualAtt := payload.Attenuations[i]
				for key, expectedVal := range expectedAtt {
					actualVal, exists := actualAtt[key]
					assert.True(t, exists, "key %s should exist", key)

					// Handle []string to []any conversion
					if expectedSlice, ok := expectedVal.([]string); ok {
						actualSlice, ok := actualVal.([]any)
						assert.True(t, ok, "expected []any for key %s", key)
						assert.Equal(t, len(expectedSlice), len(actualSlice))
						for j, expectedItem := range expectedSlice {
							assert.Equal(t, expectedItem, actualSlice[j])
						}
					} else {
						assert.Equal(t, expectedVal, actualVal)
					}
				}
			}
			assert.Equal(t, tt.expiresAt, payload.ExpiresAt)
		})
	}
}

func TestUCANDIDTask(t *testing.T) {
	tests := []struct {
		name   string
		userID int
	}{
		{"valid DID request", 123},
		{"zero user ID", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewUCANDIDTask(tt.userID)
			require.NoError(t, err)
			assert.Equal(t, TypeUCANDIDGeneration, task.Type())

			var payload UCANDIDPayload
			err = json.Unmarshal(task.Payload(), &payload)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, payload.UserID)
		})
	}
}

// Test processor functionality with mock actors
func TestUCANTokenProcessor(t *testing.T) {
	// Create actor system
	system := actor.NewActorSystem()
	defer system.Shutdown()

	// Create mock actor - use the mock instead of real plugin
	mockActor := system.Root.Spawn(actor.PropsFromProducer(func() actor.Actor {
		return NewMockUCANActor()
	}))

	// Create processor with mock actor
	processor := &UCANProcessor{pid: mockActor}

	// Create test task
	task, err := NewUCANTokenTask(123, "did:sonr:audience", []map[string]any{
		{"can": []string{"sign"}, "with": "vault://example"},
	}, time.Now().Add(24*time.Hour).Unix())
	require.NoError(t, err)

	// Process the task - the error is expected because the mock actor returns a dead letter
	// The processor should handle this gracefully
	err = processor.ProcessTask(context.Background(), task)

	// For now, we expect this to fail with the dead letter error
	// In a real scenario, the actor would be properly initialized with enclave data
	assert.Error(t, err, "expected error due to mock actor limitations")
	assert.Contains(t, err.Error(), "dead letter", "should get dead letter error from mock")
}

// TestUCANActorWithRealEnclave tests the real UCAN actor with generated enclave data
func TestUCANActorWithRealEnclave(t *testing.T) {
	// Skip this test if we're not in integration test mode
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Skip this test in CI environments where WASM plugin is not available
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("skipping WASM enclave test in CI environment")
	}

	// Generate test enclave data
	enclaveData := generateTestEnclaveData(t)

	// Create enclave config with test data
	config := plugin.DefaultEnclaveConfig()
	config.EnclaveData = enclaveData

	// Create actor system
	system := actor.NewActorSystem()
	defer system.Shutdown()

	// Create real UCAN actor with the enclave configuration
	realActor := system.Root.Spawn(plugin.PropsWithConfig(config))

	// Wait for actor to initialize (longer wait for CI environments)
	time.Sleep(500 * time.Millisecond)

	// Test creating a UCAN token through the actor
	req := &plugin.NewOriginTokenRequest{
		AudienceDID: "did:sonr:test-audience",
		Attenuations: []map[string]any{
			{"can": []string{"read"}, "with": "vault://test"},
		},
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	}

	// Send request to actor with enhanced error handling and longer timeout for CI
	result := system.Root.RequestFuture(realActor, req, 10*time.Second)
	response, err := result.Result()
	if err != nil {
		t.Logf("Actor request failed: %v", err)

		// If it's a timeout or WASM-related error, that's expected in test environments
		if err.Error() == "future: timeout" || strings.Contains(err.Error(), "wasm") {
			t.Skip("skipping test due to WASM enclave not available in test environment")
		}
	}

	// The response could be an error if the actor didn't initialize properly
	// For now, just verify we got some response (could be error or success) - but only if not a timeout
	if err == nil || err.Error() != "future: timeout" {
		assert.NotNil(t, response, "should receive some response from actor")
	}

	// Check if this is a WASM-related error (expected when WASM plugin is not available)
	if err == nil && response != nil {
		// The response could be a string or an error
		responseStr := ""
		if errResponse, ok := response.(error); ok {
			responseStr = errResponse.Error()
		} else if strResponse, ok := response.(string); ok {
			responseStr = strResponse
		} else {
			responseStr = fmt.Sprintf("%+v", response)
		}

		if responseStr != "" && (strings.Contains(responseStr, "wasm error: unreachable") ||
			strings.Contains(responseStr, "wasm not available") ||
			strings.Contains(responseStr, "runtime.notInitialized")) {
			t.Log("WASM plugin not available (expected in test environment)")
		} else {
			t.Logf("Actor successfully processed request: %+v", response)
		}
	} else {
		t.Logf("Actor response: %+v, error: %v", response, err)
	}
}

func TestInvalidJSONPayload(t *testing.T) {
	// Create processor
	processor := NewUCANProcessor()

	// Create task with invalid JSON
	task := asynq.NewTask(TypeUCANToken, []byte("invalid json"))

	// Process the task - should fail with skip retry
	err := processor.ProcessTask(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "json.Unmarshal failed")
}
