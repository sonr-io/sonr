// Package wasm provides cryptographic verification for WebAssembly modules
package wasm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

// HashVerifier provides SHA256 hash verification for WASM modules
type HashVerifier struct {
	// trustedHashes stores SHA256 hashes of trusted WASM modules
	trustedHashes map[string]string
	mu            sync.RWMutex
}

// NewHashVerifier creates a new WASM hash verifier
func NewHashVerifier() *HashVerifier {
	return &HashVerifier{
		trustedHashes: make(map[string]string),
	}
}

// ComputeHash calculates SHA256 hash of WASM bytecode
func (v *HashVerifier) ComputeHash(wasmBytes []byte) string {
	hash := sha256.Sum256(wasmBytes)
	return hex.EncodeToString(hash[:])
}

// AddTrustedHash adds a trusted hash for a named WASM module
func (v *HashVerifier) AddTrustedHash(name, hash string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.trustedHashes[name] = hash
}

// VerifyHash verifies WASM bytecode against trusted hash
func (v *HashVerifier) VerifyHash(name string, wasmBytes []byte) error {
	v.mu.RLock()
	trustedHash, exists := v.trustedHashes[name]
	v.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no trusted hash found for WASM module: %s", name)
	}

	computedHash := v.ComputeHash(wasmBytes)
	if computedHash != trustedHash {
		return fmt.Errorf(
			"WASM hash verification failed for %s: expected %s, got %s",
			name, trustedHash, computedHash,
		)
	}

	return nil
}

// VerifyHashWithFallback verifies against primary hash or fallback list
func (v *HashVerifier) VerifyHashWithFallback(name string, wasmBytes []byte, fallbackHashes []string) error {
	// Try primary verification first
	if err := v.VerifyHash(name, wasmBytes); err == nil {
		return nil
	}

	// Check against fallback hashes
	computedHash := v.ComputeHash(wasmBytes)
	for _, fallbackHash := range fallbackHashes {
		if computedHash == fallbackHash {
			// Update trusted hash for future use
			v.AddTrustedHash(name, computedHash)
			return nil
		}
	}

	return fmt.Errorf(
		"WASM hash verification failed: computed hash %s not in trusted set",
		computedHash,
	)
}

// GetTrustedHash retrieves the trusted hash for a module
func (v *HashVerifier) GetTrustedHash(name string) (string, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	hash, exists := v.trustedHashes[name]
	return hash, exists
}

// ClearTrustedHashes removes all trusted hashes
func (v *HashVerifier) ClearTrustedHashes() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.trustedHashes = make(map[string]string)
}

// HashChain provides hash chain verification for plugin updates
type HashChain struct {
	chain []HashEntry
	mu    sync.RWMutex
}

// HashEntry represents a single entry in the hash chain
type HashEntry struct {
	Version      string `json:"version"`
	Hash         string `json:"hash"`
	PreviousHash string `json:"previous_hash"`
	Timestamp    int64  `json:"timestamp"`
}

// NewHashChain creates a new hash chain
func NewHashChain() *HashChain {
	return &HashChain{
		chain: make([]HashEntry, 0),
	}
}

// AddEntry adds a new entry to the hash chain
func (hc *HashChain) AddEntry(version, hash string, timestamp int64) error {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	previousHash := ""
	if len(hc.chain) > 0 {
		previousHash = hc.chain[len(hc.chain)-1].Hash
	}

	entry := HashEntry{
		Version:      version,
		Hash:         hash,
		PreviousHash: previousHash,
		Timestamp:    timestamp,
	}

	hc.chain = append(hc.chain, entry)
	return nil
}

// VerifyChain verifies the integrity of the hash chain
func (hc *HashChain) VerifyChain() error {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	if len(hc.chain) == 0 {
		return nil
	}

	// First entry should have empty previous hash
	if hc.chain[0].PreviousHash != "" {
		return fmt.Errorf("invalid hash chain: first entry has non-empty previous hash")
	}

	// Verify chain continuity
	for i := 1; i < len(hc.chain); i++ {
		if hc.chain[i].PreviousHash != hc.chain[i-1].Hash {
			return fmt.Errorf(
				"hash chain broken at version %s: expected previous hash %s, got %s",
				hc.chain[i].Version,
				hc.chain[i-1].Hash,
				hc.chain[i].PreviousHash,
			)
		}
	}

	return nil
}

// GetLatestEntry returns the most recent hash chain entry
func (hc *HashChain) GetLatestEntry() (*HashEntry, error) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	if len(hc.chain) == 0 {
		return nil, fmt.Errorf("hash chain is empty")
	}

	latest := hc.chain[len(hc.chain)-1]
	return &latest, nil
}

// VerificationError represents a WASM verification failure
type VerificationError struct {
	Module       string
	ExpectedHash string
	ActualHash   string
	Reason       string
}

// Error implements the error interface
func (e *VerificationError) Error() string {
	return fmt.Sprintf(
		"WASM verification failed for %s: %s (expected: %s, actual: %s)",
		e.Module, e.Reason, e.ExpectedHash, e.ActualHash,
	)
}

// SecurityPolicy defines verification requirements
type SecurityPolicy struct {
	RequireHashVerification bool
	RequireSignature        bool
	AllowedHashes           []string
	MaxModuleSize           int64
}

// DefaultSecurityPolicy returns a secure default policy
func DefaultSecurityPolicy() *SecurityPolicy {
	return &SecurityPolicy{
		RequireHashVerification: true,
		RequireSignature:        false, // Will be enabled in next phase
		AllowedHashes:           []string{},
		MaxModuleSize:           10 * 1024 * 1024, // 10MB max
	}
}

// Validate checks if WASM module meets security policy
func (p *SecurityPolicy) Validate(wasmBytes []byte) error {
	if p.MaxModuleSize > 0 && int64(len(wasmBytes)) > p.MaxModuleSize {
		return fmt.Errorf(
			"WASM module size %d exceeds maximum allowed size %d",
			len(wasmBytes), p.MaxModuleSize,
		)
	}
	return nil
}
