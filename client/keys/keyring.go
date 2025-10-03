// Package keys provides key management functionality for the Sonr client SDK.
// This package integrates with Sonr's DWN plugin architecture for Decentralized Abstracted Smart Wallets.
package keys

import (
	"context"
	"fmt"
	"time"

	"github.com/sonr-io/sonr/client/errors"
	"github.com/sonr-io/sonr/x/dwn/client/plugin"
)

// KeyringManager provides an interface for managing Decentralized Abstracted Smart Wallets
// through the Sonr DWN plugin architecture.
type KeyringManager interface {
	// Wallet Identity Operations
	GetIssuerDID(ctx context.Context) (*WalletIdentity, error)
	GetAddress(ctx context.Context) (string, error)

	// UCAN Token Operations for Authorization
	CreateOriginToken(ctx context.Context, req *UCANRequest) (*UCANToken, error)
	CreateAttenuatedToken(ctx context.Context, req *AttenuatedUCANRequest) (*UCANToken, error)

	// Signing Operations using MPC
	Sign(ctx context.Context, data []byte) (*Signature, error)
	SignTransaction(ctx context.Context, txBytes []byte) (*Signature, error)

	// Verification Operations
	Verify(ctx context.Context, data []byte, signature []byte) (bool, error)

	// Plugin Management
	Plugin() plugin.Plugin
	Close() error
}

// WalletIdentity represents the identity of a Decentralized Abstracted Smart Wallet.
type WalletIdentity struct {
	DID       string `json:"did"`        // W3C DID identifier
	Address   string `json:"address"`    // Sonr blockchain address
	ChainCode string `json:"chain_code"` // Deterministic chain code
}

// UCANRequest represents a request to create a UCAN origin token.
type UCANRequest struct {
	AudienceDID  string           `json:"audience_did"`           // Target audience DID
	Capabilities []map[string]any `json:"capabilities,omitempty"` // Granted capabilities
	Facts        []string         `json:"facts,omitempty"`        // Additional facts
	NotBefore    *time.Time       `json:"not_before,omitempty"`   // Validity start time
	ExpiresAt    *time.Time       `json:"expires_at,omitempty"`   // Expiration time
}

// AttenuatedUCANRequest represents a request to create an attenuated UCAN token.
type AttenuatedUCANRequest struct {
	ParentToken  string           `json:"parent_token"`           // Parent UCAN token
	AudienceDID  string           `json:"audience_did"`           // Target audience DID
	Capabilities []map[string]any `json:"capabilities,omitempty"` // Attenuated capabilities
	Facts        []string         `json:"facts,omitempty"`        // Additional facts
	NotBefore    *time.Time       `json:"not_before,omitempty"`   // Validity start time
	ExpiresAt    *time.Time       `json:"expires_at,omitempty"`   // Expiration time
}

// UCANToken represents a User-Controlled Authorization Network token.
type UCANToken struct {
	Token   string `json:"token"`   // The UCAN JWT token
	Issuer  string `json:"issuer"`  // Issuer DID
	Address string `json:"address"` // Issuer address
}

// Signature represents a cryptographic signature.
type Signature struct {
	Signature []byte `json:"signature"` // Signature bytes
	Algorithm string `json:"algorithm"` // Signature algorithm used
}

// keyringManager implements KeyringManager using the DWN plugin architecture.
type keyringManager struct {
	plugin  plugin.Plugin
	chainID string
}

// KeyringOptions configures keyring creation for Decentralized Abstracted Smart Wallets.
type KeyringOptions struct {
	ChainID     string         `json:"chain_id"`
	EnclaveData []byte         `json:"enclave_data"` // JSON-encoded MPC enclave data
	VaultConfig map[string]any `json:"vault_config,omitempty"`
	UseManager  bool           `json:"use_manager,omitempty"` // Use plugin manager for production
}

// NewKeyringManager creates a new keyring manager using the DWN plugin architecture.
func NewKeyringManager(backend, dir, chainID string) (KeyringManager, error) {
	if chainID == "" {
		return nil, fmt.Errorf("chain ID is required")
	}

	// For backwards compatibility, create with minimal config
	// In practice, clients should use NewKeyringManagerWithOptions
	opts := KeyringOptions{
		ChainID:     chainID,
		EnclaveData: []byte(`{}`), // Minimal enclave data
		UseManager:  false,        // Use simple plugin loading
	}

	return NewKeyringManagerWithOptions(opts)
}

// NewKeyringManagerWithOptions creates a new keyring manager with detailed options
// for Decentralized Abstracted Smart Wallets.
func NewKeyringManagerWithOptions(opts KeyringOptions) (KeyringManager, error) {
	if opts.ChainID == "" {
		return nil, fmt.Errorf("chain ID is required")
	}

	ctx := context.Background()
	var p plugin.Plugin
	var err error

	if opts.UseManager {
		// For production use with plugin manager, we would need to properly construct
		// the EnclaveConfig with the correct types. For now, fall back to simple loading.
		// TODO: Implement proper EnclaveConfig construction when plugin manager is ready
		p, err = plugin.LoadPluginWithEnclave(ctx, opts.ChainID, opts.EnclaveData, opts.VaultConfig)
	} else {
		// Use simple plugin loading for development
		p, err = plugin.LoadPluginWithEnclave(ctx, opts.ChainID, opts.EnclaveData, opts.VaultConfig)
	}

	if err != nil {
		return nil, errors.WrapError(err, errors.ErrSigningFailed, "failed to load DWN plugin for chain %s", opts.ChainID)
	}

	return &keyringManager{
		plugin:  p,
		chainID: opts.ChainID,
	}, nil
}

// GetIssuerDID retrieves the wallet's DID identity information.
func (km *keyringManager) GetIssuerDID(ctx context.Context) (*WalletIdentity, error) {
	resp, err := km.plugin.GetIssuerDID()
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrKeyNotFound, "failed to get issuer DID from wallet")
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("plugin error: %s", resp.Error)
	}

	return &WalletIdentity{
		DID:       resp.IssuerDID,
		Address:   resp.Address,
		ChainCode: resp.ChainCode,
	}, nil
}

// GetAddress retrieves the wallet's blockchain address.
func (km *keyringManager) GetAddress(ctx context.Context) (string, error) {
	identity, err := km.GetIssuerDID(ctx)
	if err != nil {
		return "", err
	}
	return identity.Address, nil
}

// CreateOriginToken creates a new UCAN origin token for authorization.
func (km *keyringManager) CreateOriginToken(ctx context.Context, req *UCANRequest) (*UCANToken, error) {
	// Convert to plugin request format
	pluginReq := &plugin.NewOriginTokenRequest{
		AudienceDID:  req.AudienceDID,
		Attenuations: req.Capabilities,
		Facts:        req.Facts,
	}

	if req.NotBefore != nil {
		pluginReq.NotBefore = req.NotBefore.Unix()
	}

	if req.ExpiresAt != nil {
		pluginReq.ExpiresAt = req.ExpiresAt.Unix()
	}

	resp, err := km.plugin.NewOriginToken(pluginReq)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrSigningFailed, "failed to create origin UCAN token")
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("plugin error: %s", resp.Error)
	}

	return &UCANToken{
		Token:   resp.Token,
		Issuer:  resp.Issuer,
		Address: resp.Address,
	}, nil
}

// CreateAttenuatedToken creates an attenuated UCAN token by delegating from a parent token.
func (km *keyringManager) CreateAttenuatedToken(ctx context.Context, req *AttenuatedUCANRequest) (*UCANToken, error) {
	// Convert to plugin request format
	pluginReq := &plugin.NewAttenuatedTokenRequest{
		ParentToken:  req.ParentToken,
		AudienceDID:  req.AudienceDID,
		Attenuations: req.Capabilities,
		Facts:        req.Facts,
	}

	if req.NotBefore != nil {
		pluginReq.NotBefore = req.NotBefore.Unix()
	}

	if req.ExpiresAt != nil {
		pluginReq.ExpiresAt = req.ExpiresAt.Unix()
	}

	resp, err := km.plugin.NewAttenuatedToken(pluginReq)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrSigningFailed, "failed to create attenuated UCAN token")
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("plugin error: %s", resp.Error)
	}

	return &UCANToken{
		Token:   resp.Token,
		Issuer:  resp.Issuer,
		Address: resp.Address,
	}, nil
}

// Sign signs arbitrary data using the MPC-based wallet.
func (km *keyringManager) Sign(ctx context.Context, data []byte) (*Signature, error) {
	req := &plugin.SignDataRequest{
		Data: data,
	}

	resp, err := km.plugin.SignData(req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrSigningFailed, "failed to sign data")
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("plugin error: %s", resp.Error)
	}

	return &Signature{
		Signature: resp.Signature,
		Algorithm: "MPC", // The plugin uses MPC-based signing
	}, nil
}

// SignTransaction signs a transaction using the MPC-based wallet.
func (km *keyringManager) SignTransaction(ctx context.Context, txBytes []byte) (*Signature, error) {
	// Transaction signing uses the same underlying data signing mechanism
	return km.Sign(ctx, txBytes)
}

// Verify verifies a signature against data using the MPC-based wallet.
func (km *keyringManager) Verify(ctx context.Context, data []byte, signature []byte) (bool, error) {
	req := &plugin.VerifyDataRequest{
		Data:      data,
		Signature: signature,
	}

	resp, err := km.plugin.VerifyData(req)
	if err != nil {
		return false, errors.WrapError(err, errors.ErrSigningFailed, "failed to verify signature")
	}

	if resp.Error != "" {
		return false, fmt.Errorf("plugin error: %s", resp.Error)
	}

	return resp.Valid, nil
}

// Plugin returns the underlying DWN plugin for advanced operations.
func (km *keyringManager) Plugin() plugin.Plugin {
	return km.plugin
}

// Close closes the keyring manager and releases resources.
func (km *keyringManager) Close() error {
	// Note: The plugin interface doesn't currently expose a Close method
	// This is here for future compatibility and to satisfy the interface
	return nil
}

// Utility functions for working with Sonr addresses and DIDs

// SonrBech32Prefix returns the bech32 prefix used by Sonr addresses.
func SonrBech32Prefix() string {
	return "sonr"
}

// WalletInfo provides formatted information about a Decentralized Abstracted Smart Wallet.
type WalletInfo struct {
	DID       string `json:"did"`
	Address   string `json:"address"`
	ChainCode string `json:"chain_code"`
	ChainID   string `json:"chain_id"`
}

// GetWalletInfo returns formatted information about the wallet.
func (km *keyringManager) GetWalletInfo(ctx context.Context) (*WalletInfo, error) {
	identity, err := km.GetIssuerDID(ctx)
	if err != nil {
		return nil, err
	}

	return &WalletInfo{
		DID:       identity.DID,
		Address:   identity.Address,
		ChainCode: identity.ChainCode,
		ChainID:   km.chainID,
	}, nil
}

// CreateDefaultUCANRequest creates a UCAN request with sensible defaults.
func CreateDefaultUCANRequest(audienceDID string) *UCANRequest {
	// Create a token that expires in 1 hour with basic capabilities
	expiresAt := time.Now().Add(time.Hour)

	return &UCANRequest{
		AudienceDID: audienceDID,
		Capabilities: []map[string]any{
			{
				"can":  []string{"sign", "verify"},
				"with": "vault://default",
			},
		},
		ExpiresAt: &expiresAt,
	}
}

// CreateTransactionUCANRequest creates a UCAN request specifically for transaction signing.
func CreateTransactionUCANRequest(audienceDID string, txHash string) *UCANRequest {
	// Create a short-lived token specifically for transaction signing
	expiresAt := time.Now().Add(10 * time.Minute)

	return &UCANRequest{
		AudienceDID: audienceDID,
		Capabilities: []map[string]any{
			{
				"can":  []string{"sign"},
				"with": fmt.Sprintf("tx://%s", txHash),
			},
		},
		Facts: []string{
			fmt.Sprintf("transaction:%s", txHash),
		},
		ExpiresAt: &expiresAt,
	}
}
