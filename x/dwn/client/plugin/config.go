package plugin

import (
	"encoding/json"
	"fmt"
	"time"

	extism "github.com/extism/go-sdk"
	"github.com/sonr-io/crypto/mpc"
)

// EnclaveConfig represents the MPC enclave configuration for the Motor plugin.
// This configuration is passed to the plugin via PDK environment variables.
type EnclaveConfig struct {
	// ChainID specifies the blockchain network identifier (e.g., "sonr-testnet-1")
	ChainID string `json:"chain_id" yaml:"chain_id"`

	// EnclaveData contains the MPC enclave data with private key material
	EnclaveData *mpc.EnclaveData `json:"enclave_data" yaml:"enclave_data"`

	// VaultConfig provides additional vault configuration parameters
	VaultConfig VaultConfig `json:"vault_config" yaml:"vault_config"`

	// Security settings for plugin operations
	Security SecurityConfig `json:"security" yaml:"security"`

	// Timeout configurations for various operations
	Timeouts TimeoutConfig `json:"timeouts" yaml:"timeouts"`
}

// VaultConfig specifies vault-specific configuration parameters.
type VaultConfig struct {
	// IPFSEndpoint specifies the IPFS endpoint for vault operations
	IPFSEndpoint string `json:"ipfs_endpoint" yaml:"ipfs_endpoint"`

	// MaxVaultSize limits the maximum size of vault data in bytes
	MaxVaultSize int64 `json:"max_vault_size" yaml:"max_vault_size"`

	// EnableCompression enables compression for vault data
	EnableCompression bool `json:"enable_compression" yaml:"enable_compression"`

	// BackupEnabled enables automatic backup of vault data
	BackupEnabled bool `json:"backup_enabled" yaml:"backup_enabled"`

	// Custom metadata for vault operations
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// SecurityConfig defines security parameters for plugin operations.
type SecurityConfig struct {
	// RequiredAttestations specifies required security attestations
	RequiredAttestations []string `json:"required_attestations" yaml:"required_attestations"`

	// MaxTokenLifetime limits the maximum lifetime of generated tokens
	MaxTokenLifetime time.Duration `json:"max_token_lifetime" yaml:"max_token_lifetime"`

	// RequireAudience enforces audience validation for all tokens
	RequireAudience bool `json:"require_audience" yaml:"require_audience"`

	// AllowedOrigins specifies allowed origins for token delegation
	AllowedOrigins []string `json:"allowed_origins" yaml:"allowed_origins"`
}

// TimeoutConfig specifies timeout values for various plugin operations.
type TimeoutConfig struct {
	// TokenCreation timeout for UCAN token creation operations
	TokenCreation time.Duration `json:"token_creation" yaml:"token_creation"`

	// Signature timeout for cryptographic signing operations
	Signature time.Duration `json:"signature" yaml:"signature"`

	// Verification timeout for signature verification operations
	Verification time.Duration `json:"verification" yaml:"verification"`

	// PluginInit timeout for plugin initialization
	PluginInit time.Duration `json:"plugin_init" yaml:"plugin_init"`
}

// DefaultEnclaveConfig returns a default enclave configuration with sensible defaults.
func DefaultEnclaveConfig() *EnclaveConfig {
	return &EnclaveConfig{
		ChainID: "sonr-testnet-1",
		VaultConfig: VaultConfig{
			IPFSEndpoint:      "127.0.0.1:5001",
			MaxVaultSize:      10 * 1024 * 1024, // 10MB
			EnableCompression: true,
			BackupEnabled:     false,
			Metadata:          make(map[string]string),
		},
		Security: SecurityConfig{
			RequiredAttestations: []string{},
			MaxTokenLifetime:     24 * time.Hour,
			RequireAudience:      true,
			AllowedOrigins:       []string{"*"},
		},
		Timeouts: TimeoutConfig{
			TokenCreation: 30 * time.Second,
			Signature:     10 * time.Second,
			Verification:  5 * time.Second,
			PluginInit:    15 * time.Second,
		},
	}
}

// Validate checks that the enclave configuration is valid and complete.
func (c *EnclaveConfig) Validate() error {
	if c.ChainID == "" {
		return fmt.Errorf("chain_id is required")
	}

	if c.EnclaveData == nil {
		return fmt.Errorf("enclave_data is required")
	}

	if !c.EnclaveData.IsValid() {
		return fmt.Errorf("enclave_data is invalid")
	}

	// Validate vault configuration
	if err := c.VaultConfig.Validate(); err != nil {
		return fmt.Errorf("vault_config validation failed: %w", err)
	}

	// Validate security configuration
	if err := c.Security.Validate(); err != nil {
		return fmt.Errorf("security configuration validation failed: %w", err)
	}

	return nil
}

// ToManifestConfig converts the enclave configuration to Extism manifest config.
// This is used to pass configuration to the WASM plugin via environment variables.
func (c *EnclaveConfig) ToManifestConfig() (map[string]string, error) {
	config := make(map[string]string)

	// Add chain ID
	config["chain_id"] = c.ChainID

	// Serialize and add enclave data
	if c.EnclaveData != nil {
		enclaveBytes, err := json.Marshal(c.EnclaveData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal enclave data: %w", err)
		}
		config["enclave"] = string(enclaveBytes)
	}

	// Serialize and add vault configuration
	vaultBytes, err := json.Marshal(c.VaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal vault config: %w", err)
	}
	config["vault_config"] = string(vaultBytes)

	// Serialize and add security configuration
	securityBytes, err := json.Marshal(c.Security)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal security config: %w", err)
	}
	config["security_config"] = string(securityBytes)

	// Serialize and add timeout configuration
	timeoutBytes, err := json.Marshal(c.Timeouts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal timeout config: %w", err)
	}
	config["timeout_config"] = string(timeoutBytes)

	return config, nil
}

// Validate checks that the vault configuration is valid.
func (v *VaultConfig) Validate() error {
	if v.MaxVaultSize <= 0 {
		return fmt.Errorf("max_vault_size must be positive")
	}

	if v.MaxVaultSize > 100*1024*1024 { // 100MB limit
		return fmt.Errorf("max_vault_size exceeds maximum allowed (100MB)")
	}

	return nil
}

// Validate checks that the security configuration is valid.
func (s *SecurityConfig) Validate() error {
	if s.MaxTokenLifetime <= 0 {
		return fmt.Errorf("max_token_lifetime must be positive")
	}

	if s.MaxTokenLifetime > 30*24*time.Hour { // 30 days limit
		return fmt.Errorf("max_token_lifetime exceeds maximum allowed (30 days)")
	}

	return nil
}

// LoaderConfig represents configuration for the plugin loader itself.
type LoaderConfig struct {
	// EnableWASI enables WebAssembly System Interface for the plugin
	EnableWASI bool

	// MemoryLimit sets the maximum memory limit for the plugin in bytes
	MemoryLimit uint32

	// AllowHttpRequests enables HTTP requests from the plugin
	AllowHttpRequests bool

	// LogLevel sets the logging level for plugin operations
	LogLevel string

	// MaxConcurrentPlugins limits the number of concurrent plugin instances
	MaxConcurrentPlugins int
}

// DefaultLoaderConfig returns a default loader configuration.
func DefaultLoaderConfig() *LoaderConfig {
	return &LoaderConfig{
		EnableWASI:           true,
		MemoryLimit:          64 * 1024 * 1024, // 64MB
		AllowHttpRequests:    false,
		LogLevel:             "info",
		MaxConcurrentPlugins: 10,
	}
}

// ToPluginConfig converts the loader configuration to Extism plugin config.
func (l *LoaderConfig) ToPluginConfig() extism.PluginConfig {
	return extism.PluginConfig{
		EnableWasi: l.EnableWASI,
	}
}

// PluginState represents the runtime state of a plugin instance.
type PluginState struct {
	// ID is the unique identifier for this plugin instance
	ID string

	// Config is the configuration used to create this plugin
	Config *EnclaveConfig

	// Plugin is the underlying Extism plugin instance
	Plugin *extism.Plugin

	// CreatedAt is the timestamp when the plugin was created
	CreatedAt time.Time

	// LastUsed is the timestamp of the last plugin operation
	LastUsed time.Time

	// IsHealthy indicates whether the plugin is in a healthy state
	IsHealthy bool

	// ErrorCount tracks the number of errors encountered
	ErrorCount int

	// MaxErrors is the maximum number of errors before marking unhealthy
	MaxErrors int
}

// UpdateHealth updates the plugin health status based on operation result.
func (s *PluginState) UpdateHealth(err error) {
	s.LastUsed = time.Now()

	if err != nil {
		s.ErrorCount++
		if s.ErrorCount >= s.MaxErrors {
			s.IsHealthy = false
		}
	} else {
		// Reset error count on successful operation
		s.ErrorCount = 0
		s.IsHealthy = true
	}
}

// IsExpired checks if the plugin instance should be considered expired.
func (s *PluginState) IsExpired(maxIdleTime time.Duration) bool {
	return time.Since(s.LastUsed) > maxIdleTime
}

// NewPluginState creates a new plugin state with default values.
func NewPluginState(id string, config *EnclaveConfig, plugin *extism.Plugin) *PluginState {
	return &PluginState{
		ID:         id,
		Config:     config,
		Plugin:     plugin,
		CreatedAt:  time.Now(),
		LastUsed:   time.Now(),
		IsHealthy:  true,
		ErrorCount: 0,
		MaxErrors:  5, // Allow up to 5 errors before marking as unhealthy
	}
}
