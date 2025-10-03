// Package wasm provides cryptographic signing and verification for WebAssembly modules
package wasm

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Signer provides Ed25519 digital signature operations for WASM modules
type Signer struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewSigner creates a new signer with a generated Ed25519 key pair
func NewSigner() (*Signer, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}

	return &Signer{
		privateKey: priv,
		publicKey:  pub,
	}, nil
}

// NewSignerFromPrivateKey creates a signer from an existing private key
func NewSignerFromPrivateKey(privateKey ed25519.PrivateKey) (*Signer, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d",
			ed25519.PrivateKeySize, len(privateKey))
	}

	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &Signer{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

// Sign creates an Ed25519 signature for the given WASM bytecode
func (s *Signer) Sign(wasmBytes []byte) ([]byte, error) {
	if s.privateKey == nil {
		return nil, fmt.Errorf("private key not initialized")
	}

	signature := ed25519.Sign(s.privateKey, wasmBytes)
	return signature, nil
}

// GetPublicKey returns the public key bytes
func (s *Signer) GetPublicKey() []byte {
	return s.publicKey
}

// GetPublicKeyHex returns the public key as hex string
func (s *Signer) GetPublicKeyHex() string {
	return hex.EncodeToString(s.publicKey)
}

// ExportPrivateKey exports the private key (handle with care)
func (s *Signer) ExportPrivateKey() []byte {
	return s.privateKey
}

// SignatureVerifier verifies Ed25519 signatures on WASM modules
type SignatureVerifier struct {
	trustedKeys map[string]ed25519.PublicKey
	mu          sync.RWMutex
}

// NewSignatureVerifier creates a new signature verifier
func NewSignatureVerifier() *SignatureVerifier {
	return &SignatureVerifier{
		trustedKeys: make(map[string]ed25519.PublicKey),
	}
}

// AddTrustedKey adds a trusted public key for signature verification
func (v *SignatureVerifier) AddTrustedKey(keyID string, publicKey ed25519.PublicKey) error {
	if len(publicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size: expected %d, got %d",
			ed25519.PublicKeySize, len(publicKey))
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	v.trustedKeys[keyID] = publicKey
	return nil
}

// AddTrustedKeyFromHex adds a trusted public key from hex string
func (v *SignatureVerifier) AddTrustedKeyFromHex(keyID string, publicKeyHex string) error {
	publicKey, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return fmt.Errorf("failed to decode public key hex: %w", err)
	}

	return v.AddTrustedKey(keyID, ed25519.PublicKey(publicKey))
}

// Verify verifies a signature against trusted public keys
func (v *SignatureVerifier) Verify(wasmBytes []byte, signature []byte) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if len(v.trustedKeys) == 0 {
		return fmt.Errorf("no trusted keys configured")
	}

	// Try to verify with any trusted key
	for keyID, publicKey := range v.trustedKeys {
		if ed25519.Verify(publicKey, wasmBytes, signature) {
			// Signature valid with this key
			return nil
		}
		_ = keyID // Key ID available for logging if needed
	}

	return fmt.Errorf("signature verification failed: no matching trusted key")
}

// VerifyWithKey verifies a signature with a specific key
func (v *SignatureVerifier) VerifyWithKey(keyID string, wasmBytes []byte, signature []byte) error {
	v.mu.RLock()
	publicKey, exists := v.trustedKeys[keyID]
	v.mu.RUnlock()

	if !exists {
		return fmt.Errorf("trusted key not found: %s", keyID)
	}

	if !ed25519.Verify(publicKey, wasmBytes, signature) {
		return fmt.Errorf("signature verification failed for key: %s", keyID)
	}

	return nil
}

// RemoveTrustedKey removes a trusted key
func (v *SignatureVerifier) RemoveTrustedKey(keyID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.trustedKeys, keyID)
}

// GetTrustedKeyIDs returns all trusted key IDs
func (v *SignatureVerifier) GetTrustedKeyIDs() []string {
	v.mu.RLock()
	defer v.mu.RUnlock()

	ids := make([]string, 0, len(v.trustedKeys))
	for id := range v.trustedKeys {
		ids = append(ids, id)
	}
	return ids
}

// SignedModule represents a WASM module with its signature
type SignedModule struct {
	Module    []byte    `json:"-"`         // WASM bytecode (excluded from JSON)
	Hash      string    `json:"hash"`      // SHA256 hash of module
	Signature []byte    `json:"signature"` // Ed25519 signature
	SignerID  string    `json:"signer_id"` // ID of signing key
	Timestamp time.Time `json:"timestamp"` // Signing timestamp
	Version   string    `json:"version"`   // Module version
}

// SignModule creates a signed module package
func SignModule(signer *Signer, module []byte, signerID string, version string) (*SignedModule, error) {
	// Compute hash
	hashVerifier := NewHashVerifier()
	hash := hashVerifier.ComputeHash(module)

	// Sign the module
	signature, err := signer.Sign(module)
	if err != nil {
		return nil, fmt.Errorf("failed to sign module: %w", err)
	}

	return &SignedModule{
		Module:    module,
		Hash:      hash,
		Signature: signature,
		SignerID:  signerID,
		Timestamp: time.Now(),
		Version:   version,
	}, nil
}

// VerifySignedModule verifies a signed module
func VerifySignedModule(verifier *SignatureVerifier, module *SignedModule) error {
	// Verify hash matches
	hashVerifier := NewHashVerifier()
	computedHash := hashVerifier.ComputeHash(module.Module)
	if computedHash != module.Hash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", module.Hash, computedHash)
	}

	// Verify signature
	if module.SignerID != "" {
		return verifier.VerifyWithKey(module.SignerID, module.Module, module.Signature)
	}

	return verifier.Verify(module.Module, module.Signature)
}

// SignatureManifest contains signature metadata for a WASM module
type SignatureManifest struct {
	ModuleHash  string            `json:"module_hash"`
	Signatures  []SignatureEntry  `json:"signatures"`
	TrustedKeys []TrustedKeyEntry `json:"trusted_keys"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// SignatureEntry represents a single signature in the manifest
type SignatureEntry struct {
	Signature string    `json:"signature"` // Base64 encoded
	SignerID  string    `json:"signer_id"`
	Timestamp time.Time `json:"timestamp"`
	Algorithm string    `json:"algorithm"` // Always "Ed25519"
}

// TrustedKeyEntry represents a trusted public key
type TrustedKeyEntry struct {
	KeyID     string     `json:"key_id"`
	PublicKey string     `json:"public_key"` // Base64 encoded
	AddedAt   time.Time  `json:"added_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Purpose   string     `json:"purpose"` // e.g., "code-signing"
}

// CreateSignatureManifest creates a manifest for module signatures
func CreateSignatureManifest(module []byte, signer *Signer, signerID string) (*SignatureManifest, error) {
	hashVerifier := NewHashVerifier()
	moduleHash := hashVerifier.ComputeHash(module)

	signature, err := signer.Sign(module)
	if err != nil {
		return nil, fmt.Errorf("failed to sign module: %w", err)
	}

	manifest := &SignatureManifest{
		ModuleHash: moduleHash,
		Signatures: []SignatureEntry{
			{
				Signature: base64.StdEncoding.EncodeToString(signature),
				SignerID:  signerID,
				Timestamp: time.Now(),
				Algorithm: "Ed25519",
			},
		},
		TrustedKeys: []TrustedKeyEntry{
			{
				KeyID:     signerID,
				PublicKey: base64.StdEncoding.EncodeToString(signer.GetPublicKey()),
				AddedAt:   time.Now(),
				Purpose:   "code-signing",
			},
		},
		CreatedAt: time.Now(),
	}

	return manifest, nil
}

// VerifyWithManifest verifies a module using a signature manifest
func VerifyWithManifest(module []byte, manifest *SignatureManifest) error {
	// Verify hash
	hashVerifier := NewHashVerifier()
	computedHash := hashVerifier.ComputeHash(module)
	if computedHash != manifest.ModuleHash {
		return fmt.Errorf("module hash mismatch")
	}

	// Check expiration
	if manifest.ExpiresAt != nil && time.Now().After(*manifest.ExpiresAt) {
		return fmt.Errorf("signature manifest has expired")
	}

	// Create verifier with trusted keys from manifest
	verifier := NewSignatureVerifier()
	for _, key := range manifest.TrustedKeys {
		// Check key expiration
		if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
			continue // Skip expired keys
		}

		publicKey, err := base64.StdEncoding.DecodeString(key.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to decode public key: %w", err)
		}

		if err := verifier.AddTrustedKey(key.KeyID, ed25519.PublicKey(publicKey)); err != nil {
			return fmt.Errorf("failed to add trusted key: %w", err)
		}
	}

	// Verify at least one signature
	for _, sig := range manifest.Signatures {
		signature, err := base64.StdEncoding.DecodeString(sig.Signature)
		if err != nil {
			continue // Skip invalid signatures
		}

		if err := verifier.VerifyWithKey(sig.SignerID, module, signature); err == nil {
			// At least one valid signature found
			return nil
		}
	}

	return fmt.Errorf("no valid signatures found in manifest")
}

// ExportManifest exports a signature manifest as JSON
func ExportManifest(manifest *SignatureManifest) ([]byte, error) {
	return json.MarshalIndent(manifest, "", "  ")
}

// ImportManifest imports a signature manifest from JSON
func ImportManifest(data []byte) (*SignatureManifest, error) {
	var manifest SignatureManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	return &manifest, nil
}
