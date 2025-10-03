package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestNetworkConfigurations tests pre-configured networks.
func TestNetworkConfigurations(t *testing.T) {
	tests := []struct {
		name        string
		network     NetworkConfig
		expectValid bool
	}{
		{
			name:        "testnet config",
			network:     TestnetNetwork(),
			expectValid: true,
		},
		{
			name:        "local config",
			network:     LocalNetwork(),
			expectValid: true,
		},
		{
			name:        "local API config",
			network:     LocalAPINetwork(),
			expectValid: true,
		},
		{
			name:        "devnet config",
			network:     DevnetNetwork(),
			expectValid: true,
		},
		{
			name:        "mainnet config",
			network:     MainnetNetwork(),
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.network.Validate()
			if tt.expectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// TestNetworkValidation tests network configuration validation.
func TestNetworkValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    NetworkConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid config",
			config: NetworkConfig{
				ChainID:        "test-chain",
				GRPC:           "localhost:9090",
				Denom:          "snr",
				StakingDenom:   "usnr",
				GasPrice:       0.001,
				GasAdjustment:  1.5,
				RequestTimeout: 30 * time.Second,
			},
			wantError: false,
		},
		{
			name: "missing chain ID",
			config: NetworkConfig{
				GRPC:         "localhost:9090",
				Denom:        "snr",
				StakingDenom: "usnr",
				GasPrice:     0.001,
			},
			wantError: true,
			errorMsg:  "chain ID is required",
		},
		{
			name: "no endpoints",
			config: NetworkConfig{
				ChainID:      "test-chain",
				Denom:        "snr",
				StakingDenom: "usnr",
				GasPrice:     0.001,
			},
			wantError: true,
			errorMsg:  "at least one endpoint",
		},
		{
			name: "missing denom",
			config: NetworkConfig{
				ChainID:      "test-chain",
				GRPC:         "localhost:9090",
				StakingDenom: "usnr",
				GasPrice:     0.001,
			},
			wantError: true,
			errorMsg:  "denom is required",
		},
		{
			name: "missing staking denom",
			config: NetworkConfig{
				ChainID:  "test-chain",
				GRPC:     "localhost:9090",
				Denom:    "snr",
				GasPrice: 0.001,
			},
			wantError: true,
			errorMsg:  "staking denom is required",
		},
		{
			name: "invalid gas price",
			config: NetworkConfig{
				ChainID:      "test-chain",
				GRPC:         "localhost:9090",
				Denom:        "snr",
				StakingDenom: "usnr",
				GasPrice:     -0.001,
			},
			wantError: true,
			errorMsg:  "gas price must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestClientConfigValidation tests client configuration validation.
func TestClientConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    ClientConfig
		wantError bool
		errorMsg  string
	}{
		{
			name:      "default config",
			config:    *DefaultConfig(),
			wantError: false,
		},
		{
			name:      "testnet config",
			config:    *TestnetConfig(),
			wantError: false,
		},
		{
			name:      "local config",
			config:    *LocalConfig(),
			wantError: false,
		},
		{
			name: "invalid keyring backend",
			config: ClientConfig{
				Network:        TestnetNetwork(),
				KeyringBackend: "invalid",
			},
			wantError: true,
			errorMsg:  "invalid keyring backend",
		},
		{
			name: "invalid log level",
			config: ClientConfig{
				Network:  TestnetNetwork(),
				LogLevel: "invalid",
			},
			wantError: true,
			errorMsg:  "invalid log level",
		},
		{
			name: "invalid broadcast mode",
			config: ClientConfig{
				Network:       TestnetNetwork(),
				BroadcastMode: "invalid",
			},
			wantError: true,
			errorMsg:  "invalid broadcast mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestIsTestnet tests network type detection.
func TestIsTestnet(t *testing.T) {
	tests := []struct {
		networkID string
		isTestnet bool
	}{
		{"testnet", true},
		{"devnet", true},
		{"local", true},
		{"local-api", true},
		{"mainnet", false},
		{"custom", false},
	}

	for _, tt := range tests {
		t.Run(tt.networkID, func(t *testing.T) {
			network := NetworkConfig{NetworkID: tt.networkID}
			require.Equal(t, tt.isTestnet, network.IsTestnet())
		})
	}
}

// TestIsMainnet tests mainnet detection.
func TestIsMainnet(t *testing.T) {
	tests := []struct {
		networkID string
		isMainnet bool
	}{
		{"mainnet", true},
		{"testnet", false},
		{"devnet", false},
		{"local", false},
		{"custom", false},
	}

	for _, tt := range tests {
		t.Run(tt.networkID, func(t *testing.T) {
			network := NetworkConfig{NetworkID: tt.networkID}
			require.Equal(t, tt.isMainnet, network.IsMainnet())
		})
	}
}

// TestGetNetworkByChainID tests network lookup by chain ID.
func TestGetNetworkByChainID(t *testing.T) {
	tests := []struct {
		chainID     string
		expectFound bool
	}{
		{"sonrtest_1-1", true},
		{"sonr_1-1", true},
		{"unknown-chain", false},
	}

	for _, tt := range tests {
		t.Run(tt.chainID, func(t *testing.T) {
			network, found := GetNetworkByChainID(tt.chainID)
			require.Equal(t, tt.expectFound, found)
			if found {
				require.Equal(t, tt.chainID, network.ChainID)
			}
		})
	}
}

// TestNetworkConfigDefaults tests default value assignment.
func TestNetworkConfigDefaults(t *testing.T) {
	cfg := &NetworkConfig{
		ChainID:      "test-chain",
		GRPC:         "localhost:9090",
		Denom:        "snr",
		StakingDenom: "usnr",
		GasPrice:     0.001,
		// Leave defaults unset
		GasAdjustment:  0,
		RequestTimeout: 0,
		MaxRetries:     -1,
	}

	err := cfg.Validate()
	require.NoError(t, err)

	// Should set defaults
	require.Equal(t, 1.5, cfg.GasAdjustment)
	require.Equal(t, 30*time.Second, cfg.RequestTimeout)
	require.Equal(t, 3, cfg.MaxRetries)
}

// TestConfigurationConsistency tests that all configs are internally consistent.
func TestConfigurationConsistency(t *testing.T) {
	configs := []NetworkConfig{
		TestnetNetwork(),
		LocalNetwork(),
		LocalAPINetwork(),
		DevnetNetwork(),
		MainnetNetwork(),
	}

	for _, cfg := range configs {
		t.Run(cfg.Name, func(t *testing.T) {
			// Check required fields
			require.NotEmpty(t, cfg.ChainID)
			require.NotEmpty(t, cfg.Name)
			require.NotEmpty(t, cfg.NetworkID)
			require.NotEmpty(t, cfg.Denom)
			require.NotEmpty(t, cfg.StakingDenom)

			// Check at least one endpoint exists
			hasEndpoint := cfg.GRPC != "" || cfg.REST != "" || cfg.RPC != ""
			require.True(t, hasEndpoint)

			// Check gas configuration
			require.Greater(t, cfg.GasPrice, 0.0)
			require.Greater(t, cfg.GasAdjustment, 0.0)

			// Check timeouts
			require.Greater(t, cfg.RequestTimeout, time.Duration(0))
			require.GreaterOrEqual(t, cfg.MaxRetries, 0)
			require.GreaterOrEqual(t, cfg.RetryDelay, time.Duration(0))
		})
	}
}
