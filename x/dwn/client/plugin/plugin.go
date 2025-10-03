// Package plugin provides a high-level interface for interacting with the Motor WebAssembly enclave plugin.
//
// The Motor plugin operates as an MPC-based UCAN (User-Controlled Authorization Networks)
// KeyshareSource, providing sophisticated decentralized authorization capabilities. This package abstracts the
// underlying WASM implementation and provides type-safe method calls for:
//
// - UCAN token creation and delegation
// - MPC-based cryptographic signing and verification
// - DID generation and identity management
// - Secure enclave configuration and management
//
// # Usage Example
//
// Basic usage with enclave configuration:
//
//	ctx := context.Background()
//	enclaveData, _ := json.Marshal(&mpc.EnclaveData{...})
//	plugin, err := LoadPluginWithEnclave(ctx, "sonr-testnet-1", enclaveData, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create UCAN token
//	req := &NewOriginTokenRequest{
//		AudienceDID: "did:sonr:audience",
//		Attenuations: []map[string]any{
//			{"can": []string{"sign"}, "with": "vault://example"},
//		},
//	}
//	resp, err := plugin.NewOriginToken(req)
package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	extism "github.com/extism/go-sdk"
)

// Plugin defines the interface for cryptographic operations provided by the WebAssembly enclave.
// It abstracts the underlying WASM implementation and provides type-safe method calls.
// Updated to match the refactored MPC-based UCAN KeyshareSource Motor plugin.
type Plugin interface {
	// UCAN Token Operations

	// NewOriginToken creates a new UCAN origin token using MPC signing.
	NewOriginToken(req *NewOriginTokenRequest) (*UCANTokenResponse, error)

	// NewAttenuatedToken creates a delegated UCAN token with attenuated permissions.
	NewAttenuatedToken(req *NewAttenuatedTokenRequest) (*UCANTokenResponse, error)

	// Cryptographic Operations

	// SignData signs arbitrary data using the MPC enclave.
	SignData(req *SignDataRequest) (*SignDataResponse, error)

	// VerifyData verifies a signature against data using the MPC enclave.
	VerifyData(req *VerifyDataRequest) (*VerifyDataResponse, error)

	// Identity Operations

	// GetIssuerDID retrieves the issuer DID, address, and chain code from the enclave.
	GetIssuerDID() (*GetIssuerDIDResponse, error)
}

// LoadPluginWithEnclave initializes and loads the WebAssembly MPC-based UCAN enclave plugin
// with the specified enclave configuration. This method provides basic enclave configuration
// but does not include advanced features like health monitoring and automatic recovery.
//
// For production use, consider LoadPluginWithManager which provides enhanced lifecycle management.
//
// Parameters:
//   - ctx: Context for plugin initialization
//   - chainID: Chain ID for the enclave (e.g., "sonr-testnet-1")
//   - enclaveData: JSON-encoded MPC enclave data
//   - vaultConfig: Optional vault configuration parameters
//
// Returns a Plugin interface for UCAN token operations and MPC cryptographic functions.
func LoadPluginWithEnclave(
	ctx context.Context,
	chainID string,
	enclaveData []byte,
	vaultConfig map[string]any,
) (Plugin, error) {
	// Verify WASM integrity before loading
	if err := VerifyPluginIntegrity(motrPluginBytes); err != nil {
		return nil, fmt.Errorf("WASM integrity verification failed: %w", err)
	}

	manifest := GetManifestWithEnclave(chainID, enclaveData, vaultConfig)
	cfg := GetPluginConfig()
	plugin, err := extism.NewPlugin(ctx, manifest, cfg, []extism.HostFunction{})
	if err != nil {
		return nil, err
	}
	return &pluginImpl{plugin: plugin}, nil
}

// LoadPluginWithManager loads a Motor plugin using the enhanced plugin manager.
// This is the recommended method for production use as it provides:
//   - Health monitoring and automatic recovery
//   - Plugin instance caching and reuse
//   - Comprehensive configuration validation
//   - Background cleanup of expired instances
//
// Parameters:
//   - ctx: Context for plugin initialization
//   - config: Complete enclave configuration including timeouts, security settings, etc.
//
// Returns a managed Plugin interface with enhanced error handling and recovery.
func LoadPluginWithManager(ctx context.Context, config *EnclaveConfig) (Plugin, error) {
	return DefaultManager.LoadPlugin(ctx, config)
}

type pluginImpl struct {
	plugin *extism.Plugin
}

// UCAN Token Operations - Primary interface for the refactored Motor plugin

func (p *pluginImpl) NewOriginToken(req *NewOriginTokenRequest) (*UCANTokenResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	_, r, err := p.plugin.Call("new_origin_token", reqBytes)
	if err != nil {
		return nil, err
	}
	var resp UCANTokenResponse
	if err := json.Unmarshal(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// NewAttenuatedToken creates an attenuated UCAN token by delegating from a parent token.
func (p *pluginImpl) NewAttenuatedToken(
	req *NewAttenuatedTokenRequest,
) (*UCANTokenResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	_, r, err := p.plugin.Call("new_attenuated_token", reqBytes)
	if err != nil {
		return nil, err
	}
	var resp UCANTokenResponse
	if err := json.Unmarshal(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Cryptographic Operations

func (p *pluginImpl) SignData(req *SignDataRequest) (*SignDataResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	_, r, err := p.plugin.Call("sign_data", reqBytes)
	if err != nil {
		return nil, err
	}
	var resp SignDataResponse
	if err := json.Unmarshal(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *pluginImpl) VerifyData(req *VerifyDataRequest) (*VerifyDataResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	_, r, err := p.plugin.Call("verify_data", reqBytes)
	if err != nil {
		return nil, err
	}
	var resp VerifyDataResponse
	if err := json.Unmarshal(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Identity Operations

func (p *pluginImpl) GetIssuerDID() (*GetIssuerDIDResponse, error) {
	_, r, err := p.plugin.Call("get_issuer_did", []byte{})
	if err != nil {
		return nil, err
	}
	var resp GetIssuerDIDResponse
	if err := json.Unmarshal(r, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
