// Package config provides network configuration and connection settings for the Sonr client SDK.
package config

import (
	"errors"
	"fmt"
	"time"
)

// NetworkConfig defines the configuration for connecting to a Sonr network.
type NetworkConfig struct {
	// Network identification
	ChainID   string `json:"chain_id"`
	Name      string `json:"name"`
	NetworkID string `json:"network_id"`

	// Endpoints for different connection types
	GRPC string `json:"grpc_endpoint"`
	REST string `json:"rest_endpoint"`
	RPC  string `json:"rpc_endpoint"`

	// Token configuration
	Denom        string `json:"denom"`         // Normal denomination (snr)
	StakingDenom string `json:"staking_denom"` // Staking denomination (usnr)

	// Gas configuration
	GasPrice      float64 `json:"gas_price"`      // Default gas price
	GasAdjustment float64 `json:"gas_adjustment"` // Gas adjustment factor

	// Connection settings
	RequestTimeout time.Duration `json:"request_timeout"`
	MaxRetries     int           `json:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay"`

	// TLS configuration
	Insecure bool `json:"insecure"` // Disable TLS verification (for development)

	// Additional endpoints for specialized services
	IPFSGateway string `json:"ipfs_gateway,omitempty"`
	HighwayAPI  string `json:"highway_api,omitempty"`
}

// ClientConfig defines the overall configuration for the Sonr client.
type ClientConfig struct {
	// Network configuration
	Network NetworkConfig `json:"network"`

	// Keyring configuration
	KeyringBackend string `json:"keyring_backend,omitempty"` // os, file, test, memory
	KeyringDir     string `json:"keyring_dir,omitempty"`     // Directory for file backend

	// Logging configuration
	LogLevel  string `json:"log_level,omitempty"`  // debug, info, warn, error
	LogFormat string `json:"log_format,omitempty"` // json, text

	// Transaction configuration
	BroadcastMode string `json:"broadcast_mode,omitempty"` // sync, async, block

	// Feature flags
	EnableMetrics bool `json:"enable_metrics,omitempty"`
	EnableTracing bool `json:"enable_tracing,omitempty"`
}

// DefaultConfig returns a default client configuration using the testnet.
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		Network:        TestnetNetwork(),
		KeyringBackend: "test",
		LogLevel:       "info",
		LogFormat:      "text",
		BroadcastMode:  "sync",
		EnableMetrics:  false,
		EnableTracing:  false,
	}
}

// TestnetConfig returns a client configuration for the Sonr testnet.
func TestnetConfig() *ClientConfig {
	config := DefaultConfig()
	config.Network = TestnetNetwork()
	return config
}

// LocalConfig returns a client configuration for local development.
func LocalConfig() *ClientConfig {
	config := DefaultConfig()
	config.Network = LocalNetwork()
	config.KeyringBackend = "test"
	return config
}

// LocalAPIConfig returns a client configuration for local API development using localhost.
func LocalAPIConfig() *ClientConfig {
	config := DefaultConfig()
	config.Network = LocalAPINetwork()
	config.KeyringBackend = "test"
	return config
}

// TestnetNetwork returns the network configuration for the Sonr testnet.
func TestnetNetwork() NetworkConfig {
	return NetworkConfig{
		ChainID:   "sonrtest_1-1",
		Name:      "Sonr Testnet",
		NetworkID: "testnet",

		// TODO: Update these endpoints with actual testnet endpoints
		GRPC: "grpc.testnet.sonr.io:443",
		REST: "https://api.testnet.sonr.io",
		RPC:  "https://rpc.testnet.sonr.io",

		Denom:        "snr",
		StakingDenom: "usnr",

		GasPrice:      0.001,
		GasAdjustment: 1.5,

		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,

		Insecure: false,

		IPFSGateway: "https://ipfs.testnet.sonr.io",
		HighwayAPI:  "https://highway.testnet.sonr.io",
	}
}

// LocalNetwork returns the network configuration for local development.
func LocalNetwork() NetworkConfig {
	return NetworkConfig{
		ChainID:   "sonrtest_1-1",
		Name:      "Sonr Local",
		NetworkID: "local",

		GRPC: "localhost:9090",
		REST: "http://localhost:1317",
		RPC:  "http://localhost:26657",

		Denom:        "snr",
		StakingDenom: "usnr",

		GasPrice:      0.001,
		GasAdjustment: 1.5,

		RequestTimeout: 10 * time.Second,
		MaxRetries:     3,
		RetryDelay:     500 * time.Millisecond,

		Insecure: true, // Allow insecure connections for local development

		IPFSGateway: "http://localhost:8080",
		HighwayAPI:  "http://localhost:8081",
	}
}

// LocalAPINetwork returns the network configuration for local API development using localhost.
func LocalAPINetwork() NetworkConfig {
	return NetworkConfig{
		ChainID:   "sonrtest_1-1",
		Name:      "Sonr Local API",
		NetworkID: "local-api",

		GRPC: "localhost:9090",
		REST: "http://localhost:1317",
		RPC:  "http://localhost:26657",

		Denom:        "snr",
		StakingDenom: "usnr",

		GasPrice:      0.001,
		GasAdjustment: 1.5,

		RequestTimeout: 10 * time.Second,
		MaxRetries:     3,
		RetryDelay:     500 * time.Millisecond,

		Insecure: true, // Allow insecure connections for local development

		IPFSGateway: "http://localhost:8080",
		HighwayAPI:  "http://localhost:8081",
	}
}

// DevnetNetwork returns the network configuration for the development network.
func DevnetNetwork() NetworkConfig {
	return NetworkConfig{
		ChainID:   "sonrtest_1-1",
		Name:      "Sonr Devnet",
		NetworkID: "devnet",

		// TODO: Update these endpoints with actual devnet endpoints
		GRPC: "grpc.devnet.sonr.io:443",
		REST: "https://api.devnet.sonr.io",
		RPC:  "https://rpc.devnet.sonr.io",

		Denom:        "snr",
		StakingDenom: "usnr",

		GasPrice:      0.001,
		GasAdjustment: 1.5,

		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,

		Insecure: false,

		IPFSGateway: "https://ipfs.devnet.sonr.io",
		HighwayAPI:  "https://highway.devnet.sonr.io",
	}
}

// MainnetNetwork returns the network configuration for the Sonr mainnet.
// Note: Mainnet is not yet available.
func MainnetNetwork() NetworkConfig {
	return NetworkConfig{
		ChainID:   "sonr_1-1",
		Name:      "Sonr Mainnet",
		NetworkID: "mainnet",

		// TODO: Update these endpoints when mainnet is available
		GRPC: "grpc.sonr.io:443",
		REST: "https://api.sonr.io",
		RPC:  "https://rpc.sonr.io",

		Denom:        "snr",
		StakingDenom: "usnr",

		GasPrice:      0.001,
		GasAdjustment: 1.2,

		RequestTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     2 * time.Second,

		Insecure: false,

		IPFSGateway: "https://ipfs.sonr.io",
		HighwayAPI:  "https://highway.sonr.io",
	}
}

// Validate checks if the network configuration is valid.
func (nc *NetworkConfig) Validate() error {
	if nc.ChainID == "" {
		return errors.New("chain ID is required")
	}

	if nc.GRPC == "" && nc.REST == "" && nc.RPC == "" {
		return errors.New("at least one endpoint (GRPC, REST, or RPC) is required")
	}

	if nc.Denom == "" {
		return errors.New("denom is required")
	}

	if nc.StakingDenom == "" {
		return errors.New("staking denom is required")
	}

	if nc.GasPrice <= 0 {
		return errors.New("gas price must be positive")
	}

	if nc.GasAdjustment <= 0 {
		nc.GasAdjustment = 1.5 // Default value
	}

	if nc.RequestTimeout <= 0 {
		nc.RequestTimeout = 30 * time.Second // Default value
	}

	if nc.MaxRetries < 0 {
		nc.MaxRetries = 3 // Default value
	}

	return nil
}

// Validate checks if the client configuration is valid.
func (cc *ClientConfig) Validate() error {
	if err := cc.Network.Validate(); err != nil {
		return fmt.Errorf("network config validation failed: %w", err)
	}

	// Validate keyring backend
	validBackends := map[string]bool{
		"os":     true,
		"file":   true,
		"test":   true,
		"memory": true,
	}

	if cc.KeyringBackend != "" && !validBackends[cc.KeyringBackend] {
		return fmt.Errorf("invalid keyring backend: %s", cc.KeyringBackend)
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if cc.LogLevel != "" && !validLogLevels[cc.LogLevel] {
		return fmt.Errorf("invalid log level: %s", cc.LogLevel)
	}

	// Validate broadcast mode
	validBroadcastModes := map[string]bool{
		"sync":  true,
		"async": true,
		"block": true,
	}

	if cc.BroadcastMode != "" && !validBroadcastModes[cc.BroadcastMode] {
		return fmt.Errorf("invalid broadcast mode: %s", cc.BroadcastMode)
	}

	return nil
}

// IsTestnet returns true if the network is a test network.
func (nc *NetworkConfig) IsTestnet() bool {
	return nc.NetworkID == "testnet" || nc.NetworkID == "devnet" || nc.NetworkID == "local" || nc.NetworkID == "local-api"
}

// IsMainnet returns true if the network is the main network.
func (nc *NetworkConfig) IsMainnet() bool {
	return nc.NetworkID == "mainnet"
}

// GetNetworkByChainID returns a pre-configured network by chain ID.
func GetNetworkByChainID(chainID string) (NetworkConfig, bool) {
	networks := map[string]NetworkConfig{
		"sonrtest_1-1": TestnetNetwork(),
		"sonr_1-1":     MainnetNetwork(),
	}

	network, exists := networks[chainID]
	return network, exists
}
