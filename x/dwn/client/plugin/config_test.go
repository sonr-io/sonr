package plugin

import (
	"fmt"
	"testing"
	"time"

	"github.com/sonr-io/crypto/mpc"
)

// createTestEnclaveData creates mock enclave data for testing
func createTestEnclaveData() *mpc.EnclaveData {
	// Generate a real enclave for testing to ensure IsValid() returns true
	enclave, err := mpc.NewEnclave()
	if err != nil {
		// Fallback to mock data if real enclave generation fails
		// This provides compatibility for environments without proper MPC support
		testPubBytes := make([]byte, 65)
		for i := range testPubBytes {
			testPubBytes[i] = byte(i % 256)
		}
		return &mpc.EnclaveData{
			PubHex:    "03a1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12",
			PubBytes:  testPubBytes,
			ValShare:  nil,
			UserShare: nil,
			Nonce:     make([]byte, 32),
			Curve:     mpc.K256Name,
		}
	}

	return enclave.GetData()
}

func TestEnclaveConfigValidation(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := DefaultEnclaveConfig()
		config.EnclaveData = createTestEnclaveData()

		if err := config.Validate(); err != nil {
			t.Errorf("Valid config failed validation: %v", err)
		}
	})

	t.Run("missing chain_id", func(t *testing.T) {
		config := DefaultEnclaveConfig()
		config.ChainID = ""
		config.EnclaveData = createTestEnclaveData()

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for missing chain_id")
		}
	})

	t.Run("missing enclave_data", func(t *testing.T) {
		config := DefaultEnclaveConfig()
		config.EnclaveData = nil

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for missing enclave_data")
		}
	})

	t.Run("invalid vault config", func(t *testing.T) {
		config := DefaultEnclaveConfig()
		config.EnclaveData = createTestEnclaveData()
		config.VaultConfig.MaxVaultSize = -1 // Invalid size

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for invalid vault config")
		}
	})
}

func TestVaultConfigValidation(t *testing.T) {
	t.Run("valid vault config", func(t *testing.T) {
		config := DefaultEnclaveConfig().VaultConfig

		if err := config.Validate(); err != nil {
			t.Errorf("Valid vault config failed validation: %v", err)
		}
	})

	t.Run("negative max vault size", func(t *testing.T) {
		config := VaultConfig{
			MaxVaultSize: -1,
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for negative max vault size")
		}
	})

	t.Run("excessive max vault size", func(t *testing.T) {
		config := VaultConfig{
			MaxVaultSize: 200 * 1024 * 1024, // 200MB (exceeds 100MB limit)
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for excessive max vault size")
		}
	})
}

func TestSecurityConfigValidation(t *testing.T) {
	t.Run("valid security config", func(t *testing.T) {
		config := DefaultEnclaveConfig().Security

		if err := config.Validate(); err != nil {
			t.Errorf("Valid security config failed validation: %v", err)
		}
	})

	t.Run("negative token lifetime", func(t *testing.T) {
		config := SecurityConfig{
			MaxTokenLifetime: -time.Hour,
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for negative token lifetime")
		}
	})

	t.Run("excessive token lifetime", func(t *testing.T) {
		config := SecurityConfig{
			MaxTokenLifetime: 40 * 24 * time.Hour, // 40 days (exceeds 30 day limit)
		}

		if err := config.Validate(); err == nil {
			t.Error("Expected validation error for excessive token lifetime")
		}
	})
}

func TestToManifestConfig(t *testing.T) {
	config := DefaultEnclaveConfig()
	config.EnclaveData = createTestEnclaveData()

	manifestConfig, err := config.ToManifestConfig()
	if err != nil {
		t.Fatalf("ToManifestConfig failed: %v", err)
	}

	// Check required keys are present
	requiredKeys := []string{
		"chain_id",
		"enclave",
		"vault_config",
		"security_config",
		"timeout_config",
	}
	for _, key := range requiredKeys {
		if _, exists := manifestConfig[key]; !exists {
			t.Errorf("Missing required key in manifest config: %s", key)
		}
	}

	// Check chain_id value
	if manifestConfig["chain_id"] != config.ChainID {
		t.Errorf(
			"Chain ID mismatch: expected %s, got %s",
			config.ChainID,
			manifestConfig["chain_id"],
		)
	}
}

func TestDefaultConfigurations(t *testing.T) {
	t.Run("default enclave config", func(t *testing.T) {
		config := DefaultEnclaveConfig()

		if config.ChainID == "" {
			t.Error("Default enclave config should have non-empty chain_id")
		}

		if config.VaultConfig.MaxVaultSize <= 0 {
			t.Error("Default vault config should have positive max_vault_size")
		}

		if config.Security.MaxTokenLifetime <= 0 {
			t.Error("Default security config should have positive max_token_lifetime")
		}
	})

	t.Run("default loader config", func(t *testing.T) {
		config := DefaultLoaderConfig()

		if !config.EnableWASI {
			t.Error("Default loader config should enable WASI")
		}

		if config.MemoryLimit <= 0 {
			t.Error("Default loader config should have positive memory limit")
		}

		if config.MaxConcurrentPlugins <= 0 {
			t.Error("Default loader config should have positive max concurrent plugins")
		}
	})
}

func TestPluginStateManagement(t *testing.T) {
	config := DefaultEnclaveConfig()
	config.EnclaveData = createTestEnclaveData()

	state := NewPluginState("test-plugin", config, nil)

	t.Run("initial state", func(t *testing.T) {
		if state.ID != "test-plugin" {
			t.Errorf("Expected plugin ID 'test-plugin', got %s", state.ID)
		}

		if !state.IsHealthy {
			t.Error("New plugin state should be healthy")
		}

		if state.ErrorCount != 0 {
			t.Errorf("New plugin state should have zero error count, got %d", state.ErrorCount)
		}
	})

	t.Run("health updates", func(t *testing.T) {
		// Test successful operation
		state.UpdateHealth(nil)
		if !state.IsHealthy {
			t.Error("Plugin should remain healthy after successful operation")
		}
		if state.ErrorCount != 0 {
			t.Error("Error count should reset after successful operation")
		}

		// Test error handling
		testError := fmt.Errorf("test error")
		for i := 0; i < state.MaxErrors; i++ {
			state.UpdateHealth(testError)
		}

		if state.IsHealthy {
			t.Error("Plugin should be unhealthy after max errors")
		}
		if state.ErrorCount != state.MaxErrors {
			t.Errorf("Expected error count %d, got %d", state.MaxErrors, state.ErrorCount)
		}
	})

	t.Run("expiration check", func(t *testing.T) {
		maxIdleTime := 1 * time.Second

		// Plugin should not be expired immediately
		if state.IsExpired(maxIdleTime) {
			t.Error("Plugin should not be expired immediately")
		}

		// Simulate old last used time
		state.LastUsed = time.Now().Add(-2 * time.Second)

		if !state.IsExpired(maxIdleTime) {
			t.Error("Plugin should be expired after max idle time")
		}
	})
}

func TestManagerConfiguration(t *testing.T) {
	loaderConfig := DefaultLoaderConfig()
	manager := NewManager(loaderConfig)
	defer manager.Close()

	if manager.loaderConfig != loaderConfig {
		t.Error("Manager should use provided loader config")
	}

	if len(manager.plugins) != 0 {
		t.Error("New manager should have no plugins initially")
	}
}

func TestCreateEnclaveConfig(t *testing.T) {
	chainID := "test-chain"
	enclaveData := createTestEnclaveData()

	config := CreateEnclaveConfig(chainID, enclaveData)

	if config.ChainID != chainID {
		t.Errorf("Expected chain ID %s, got %s", chainID, config.ChainID)
	}

	if config.EnclaveData != enclaveData {
		t.Error("Enclave data should match provided data")
	}

	// Should have default values for other fields
	if config.VaultConfig.MaxVaultSize <= 0 {
		t.Error("Should have default vault config values")
	}
}
