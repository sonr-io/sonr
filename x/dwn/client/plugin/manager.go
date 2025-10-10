package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	extism "github.com/extism/go-sdk"
	"github.com/sonr-io/crypto/mpc"
	"github.com/sonr-io/crypto/wasm"
)

// Manager handles the lifecycle of Motor plugin instances with health monitoring,
// automatic recovery, and efficient resource management.
type Manager struct {
	mu           sync.RWMutex
	plugins      map[string]*PluginState
	loaderConfig *LoaderConfig

	// Cleanup configuration
	cleanupInterval time.Duration
	maxIdleTime     time.Duration

	// Background cleanup goroutine
	stopCleanup chan struct{}
	cleanupWG   sync.WaitGroup
}

// NewManager creates a new plugin manager with the specified configuration.
func NewManager(loaderConfig *LoaderConfig) *Manager {
	if loaderConfig == nil {
		loaderConfig = DefaultLoaderConfig()
	}

	m := &Manager{
		plugins:         make(map[string]*PluginState),
		loaderConfig:    loaderConfig,
		cleanupInterval: 5 * time.Minute,
		maxIdleTime:     30 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}

	// Start background cleanup goroutine
	m.cleanupWG.Add(1)
	go m.cleanupLoop()

	return m
}

// LoadPlugin loads a Motor plugin with the specified enclave configuration.
// Returns a cached instance if available, or creates a new one.
func (m *Manager) LoadPlugin(ctx context.Context, config *EnclaveConfig) (Plugin, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid enclave configuration: %w", err)
	}

	// Generate plugin ID based on configuration
	pluginID := m.generatePluginID(config)

	m.mu.RLock()
	state, exists := m.plugins[pluginID]
	m.mu.RUnlock()

	// Check if we have a healthy cached instance
	if exists && state.IsHealthy && !state.IsExpired(m.maxIdleTime) {
		state.UpdateHealth(nil) // Update last used timestamp
		return &managedPluginImpl{
			state:   state,
			manager: m,
		}, nil
	}

	// Create new plugin instance
	return m.createPlugin(ctx, pluginID, config)
}

// LoadPluginWithID loads a plugin with a specific ID for testing or debugging.
func (m *Manager) LoadPluginWithID(
	ctx context.Context,
	id string,
	config *EnclaveConfig,
) (Plugin, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid enclave configuration: %w", err)
	}

	return m.createPlugin(ctx, id, config)
}

// createPlugin creates a new plugin instance with the given configuration.
func (m *Manager) createPlugin(
	ctx context.Context,
	id string,
	config *EnclaveConfig,
) (Plugin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check concurrent plugin limit
	if len(m.plugins) >= m.loaderConfig.MaxConcurrentPlugins {
		return nil, fmt.Errorf(
			"maximum concurrent plugins limit reached (%d)",
			m.loaderConfig.MaxConcurrentPlugins,
		)
	}

	// Verify WASM integrity before loading
	if err := VerifyPluginIntegrity(motrPluginBytes); err != nil {
		return nil, fmt.Errorf("WASM integrity verification failed: %w", err)
	}

	// Verify signature if manifest is available (optional for now)
	if manifest := GetPluginSignatureManifest(); manifest != nil {
		if err := wasm.VerifyWithManifest(motrPluginBytes, manifest); err != nil {
			// Log warning but don't fail for backward compatibility
			// In production, this should return an error
			// Using fmt.Printf as we don't have access to pdk here
			fmt.Printf("WARNING: WASM signature verification failed: %v\n", err)
		}
	}

	// Convert configuration to manifest format
	manifestConfig, err := config.ToManifestConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to convert config to manifest: %w", err)
	}

	// Create Extism manifest
	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmData{
				Data: motrPluginBytes,
			},
		},
		Config: manifestConfig,
	}

	// Create plugin with timeout
	pluginConfig := m.loaderConfig.ToPluginConfig()

	// Create context with timeout for plugin initialization
	initCtx, cancel := context.WithTimeout(ctx, config.Timeouts.PluginInit)
	defer cancel()

	plugin, err := extism.NewPlugin(initCtx, manifest, pluginConfig, []extism.HostFunction{})
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin: %w", err)
	}

	// Create plugin state
	state := NewPluginState(id, config, plugin)
	m.plugins[id] = state

	return &managedPluginImpl{
		state:   state,
		manager: m,
	}, nil
}

// RecoverPlugin attempts to recover a failed plugin instance.
func (m *Manager) RecoverPlugin(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.plugins[id]
	if !exists {
		return fmt.Errorf("plugin %s not found", id)
	}

	// Close existing plugin
	if state.Plugin != nil {
		state.Plugin.Close(ctx)
	}

	// Recreate plugin with same configuration
	manifestConfig, err := state.Config.ToManifestConfig()
	if err != nil {
		return fmt.Errorf("failed to convert config to manifest: %w", err)
	}

	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmData{
				Data: motrPluginBytes,
			},
		},
		Config: manifestConfig,
	}

	pluginConfig := m.loaderConfig.ToPluginConfig()

	initCtx, cancel := context.WithTimeout(ctx, state.Config.Timeouts.PluginInit)
	defer cancel()

	plugin, err := extism.NewPlugin(initCtx, manifest, pluginConfig, []extism.HostFunction{})
	if err != nil {
		return fmt.Errorf("failed to recover plugin: %w", err)
	}

	// Update state
	state.Plugin = plugin
	state.IsHealthy = true
	state.ErrorCount = 0
	state.LastUsed = time.Now()

	return nil
}

// GetPluginStats returns statistics for a specific plugin.
func (m *Manager) GetPluginStats(id string) (*PluginStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.plugins[id]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", id)
	}

	return &PluginStats{
		ID:             state.ID,
		CreatedAt:      state.CreatedAt,
		LastUsed:       state.LastUsed,
		IsHealthy:      state.IsHealthy,
		ErrorCount:     state.ErrorCount,
		ChainID:        state.Config.ChainID,
		UptimeDuration: time.Since(state.CreatedAt),
		IdleDuration:   time.Since(state.LastUsed),
	}, nil
}

// ListPlugins returns a list of all currently managed plugins.
func (m *Manager) ListPlugins() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.plugins))
	for id := range m.plugins {
		ids = append(ids, id)
	}
	return ids
}

// ClosePlugin closes and removes a specific plugin instance.
func (m *Manager) ClosePlugin(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.plugins[id]
	if !exists {
		return fmt.Errorf("plugin %s not found", id)
	}

	if state.Plugin != nil {
		state.Plugin.Close(context.Background())
	}

	delete(m.plugins, id)
	return nil
}

// Close shuts down the manager and all managed plugins.
func (m *Manager) Close() error {
	// Stop cleanup goroutine
	close(m.stopCleanup)
	m.cleanupWG.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all plugins
	for id, state := range m.plugins {
		if state.Plugin != nil {
			state.Plugin.Close(context.Background())
		}
		delete(m.plugins, id)
	}

	return nil
}

// generatePluginID generates a unique plugin ID based on configuration.
func (m *Manager) generatePluginID(config *EnclaveConfig) string {
	// Use chain ID and enclave data hash for unique identification
	if config.EnclaveData != nil && len(config.EnclaveData.PubBytes) > 8 {
		pubKeyHash := fmt.Sprintf("%x", config.EnclaveData.PubBytes[:8])
		return fmt.Sprintf("%s_%s", config.ChainID, pubKeyHash)
	}
	return fmt.Sprintf("%s_%d", config.ChainID, time.Now().UnixNano())
}

// cleanupLoop runs periodic cleanup of expired and unhealthy plugins.
func (m *Manager) cleanupLoop() {
	defer m.cleanupWG.Done()

	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupExpiredPlugins()
		case <-m.stopCleanup:
			return
		}
	}
}

// cleanupExpiredPlugins removes expired and unhealthy plugin instances.
func (m *Manager) cleanupExpiredPlugins() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var toRemove []string

	for id, state := range m.plugins {
		if !state.IsHealthy || state.IsExpired(m.maxIdleTime) {
			if state.Plugin != nil {
				state.Plugin.Close(context.Background())
			}
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		delete(m.plugins, id)
	}
}

// PluginStats contains statistics and status information for a plugin instance.
type PluginStats struct {
	ID             string        `json:"id"`
	CreatedAt      time.Time     `json:"created_at"`
	LastUsed       time.Time     `json:"last_used"`
	IsHealthy      bool          `json:"is_healthy"`
	ErrorCount     int           `json:"error_count"`
	ChainID        string        `json:"chain_id"`
	UptimeDuration time.Duration `json:"uptime_duration"`
	IdleDuration   time.Duration `json:"idle_duration"`
}

// managedPluginImpl implements the Plugin interface with health monitoring.
type managedPluginImpl struct {
	state   *PluginState
	manager *Manager
}

// UCAN Token Operations with health monitoring

// callTokenMethod is a helper method to reduce duplication between token creation methods
func (p *managedPluginImpl) callTokenMethod(
	methodName string,
	request any,
) (*UCANTokenResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.state.Config.Timeouts.TokenCreation)
	defer cancel()

	reqBytes, err := json.Marshal(request)
	if err != nil {
		p.state.UpdateHealth(err)
		return nil, err
	}

	_, r, err := p.state.Plugin.CallWithContext(ctx, methodName, reqBytes)
	if err != nil {
		p.state.UpdateHealth(err)

		// Attempt recovery on failure
		if !p.state.IsHealthy {
			if recoverErr := p.manager.RecoverPlugin(ctx, p.state.ID); recoverErr == nil {
				// Retry after recovery
				_, r, err = p.state.Plugin.CallWithContext(ctx, methodName, reqBytes)
			}
		}

		if err != nil {
			return nil, err
		}
	}

	var resp UCANTokenResponse
	if err := json.Unmarshal(r, &resp); err != nil {
		p.state.UpdateHealth(err)
		return nil, err
	}

	p.state.UpdateHealth(nil)
	return &resp, nil
}

// NewOriginToken creates a new origin UCAN token with health monitoring and recovery.
func (p *managedPluginImpl) NewOriginToken(req *NewOriginTokenRequest) (*UCANTokenResponse, error) {
	return p.callTokenMethod("new_origin_token", req)
}

func (p *managedPluginImpl) NewAttenuatedToken(
	req *NewAttenuatedTokenRequest,
) (*UCANTokenResponse, error) {
	return p.callTokenMethod("new_attenuated_token", req)
}

// Cryptographic Operations with health monitoring

func (p *managedPluginImpl) SignData(req *SignDataRequest) (*SignDataResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.state.Config.Timeouts.Signature)
	defer cancel()

	reqBytes, err := json.Marshal(req)
	if err != nil {
		p.state.UpdateHealth(err)
		return nil, err
	}

	_, r, err := p.state.Plugin.CallWithContext(ctx, "sign_data", reqBytes)
	if err != nil {
		p.state.UpdateHealth(err)
		return nil, err
	}

	var resp SignDataResponse
	if err := json.Unmarshal(r, &resp); err != nil {
		p.state.UpdateHealth(err)
		return nil, err
	}

	p.state.UpdateHealth(nil)
	return &resp, nil
}

func (p *managedPluginImpl) VerifyData(req *VerifyDataRequest) (*VerifyDataResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.state.Config.Timeouts.Verification)
	defer cancel()

	reqBytes, err := json.Marshal(req)
	if err != nil {
		p.state.UpdateHealth(err)
		return nil, err
	}

	_, r, err := p.state.Plugin.CallWithContext(ctx, "verify_data", reqBytes)
	if err != nil {
		p.state.UpdateHealth(err)
		return nil, err
	}

	var resp VerifyDataResponse
	if err := json.Unmarshal(r, &resp); err != nil {
		p.state.UpdateHealth(err)
		return nil, err
	}

	p.state.UpdateHealth(nil)
	return &resp, nil
}

// Identity Operations with health monitoring

func (p *managedPluginImpl) GetIssuerDID() (*GetIssuerDIDResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, r, err := p.state.Plugin.CallWithContext(ctx, "get_issuer_did", []byte{})
	if err != nil {
		p.state.UpdateHealth(err)
		return nil, err
	}

	var resp GetIssuerDIDResponse
	if err := json.Unmarshal(r, &resp); err != nil {
		p.state.UpdateHealth(err)
		return nil, err
	}

	p.state.UpdateHealth(nil)
	return &resp, nil
}

// DefaultManager is a package-level default manager instance.
var DefaultManager *Manager

// init initializes the default manager.
func init() {
	DefaultManager = NewManager(DefaultLoaderConfig())
}

// LoadPluginWithDefaultManager is a convenience function that uses the default manager.
func LoadPluginWithDefaultManager(ctx context.Context, config *EnclaveConfig) (Plugin, error) {
	return DefaultManager.LoadPlugin(ctx, config)
}

// CreateEnclaveConfig is a helper function to create enclave configuration from MPC data.
func CreateEnclaveConfig(chainID string, enclaveData *mpc.EnclaveData) *EnclaveConfig {
	config := DefaultEnclaveConfig()
	config.ChainID = chainID
	config.EnclaveData = enclaveData
	return config
}
