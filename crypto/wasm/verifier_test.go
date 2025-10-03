package wasm

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashVerifier_ComputeHash(t *testing.T) {
	verifier := NewHashVerifier()

	testCases := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty input",
			input:    []byte{},
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "simple wasm header",
			input:    []byte{0x00, 0x61, 0x73, 0x6d}, // \0asm
			expected: "cd5d4935a48c0672cb06407bb443bc0087aff947c6b864bac886982c73b3027f",
		},
		{
			name:     "test module",
			input:    []byte("test wasm module content"),
			expected: "945acabcfc93e347e8c08ea44afd3122670f04a89f9a0a0a5ce16ab849bbac06",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := verifier.ComputeHash(tc.input)
			assert.Equal(t, tc.expected, hash)
		})
	}
}

func TestHashVerifier_TrustedHashes(t *testing.T) {
	verifier := NewHashVerifier()

	// Add trusted hash
	moduleName := "test-module"
	trustedHash := "abc123def456"
	verifier.AddTrustedHash(moduleName, trustedHash)

	// Retrieve trusted hash
	hash, exists := verifier.GetTrustedHash(moduleName)
	assert.True(t, exists)
	assert.Equal(t, trustedHash, hash)

	// Non-existent module
	_, exists = verifier.GetTrustedHash("non-existent")
	assert.False(t, exists)

	// Clear hashes
	verifier.ClearTrustedHashes()
	_, exists = verifier.GetTrustedHash(moduleName)
	assert.False(t, exists)
}

func TestHashVerifier_VerifyHash(t *testing.T) {
	verifier := NewHashVerifier()

	// Test data
	wasmBytes := []byte("test wasm module")
	expectedHash := verifier.ComputeHash(wasmBytes)
	moduleName := "test-module"

	// Test missing trusted hash
	err := verifier.VerifyHash(moduleName, wasmBytes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no trusted hash found")

	// Add trusted hash
	verifier.AddTrustedHash(moduleName, expectedHash)

	// Test successful verification
	err = verifier.VerifyHash(moduleName, wasmBytes)
	assert.NoError(t, err)

	// Test failed verification
	verifier.AddTrustedHash(moduleName, "wrong-hash")
	err = verifier.VerifyHash(moduleName, wasmBytes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hash verification failed")
}

func TestHashVerifier_VerifyHashWithFallback(t *testing.T) {
	verifier := NewHashVerifier()

	wasmBytes := []byte("test wasm module")
	actualHash := verifier.ComputeHash(wasmBytes)
	moduleName := "test-module"

	// Test with fallback hashes
	fallbackHashes := []string{
		"wrong-hash-1",
		actualHash,
		"wrong-hash-2",
	}

	err := verifier.VerifyHashWithFallback(moduleName, wasmBytes, fallbackHashes)
	assert.NoError(t, err)

	// Verify that the hash was added as trusted
	trustedHash, exists := verifier.GetTrustedHash(moduleName)
	assert.True(t, exists)
	assert.Equal(t, actualHash, trustedHash)

	// Test with no matching fallback
	fallbackHashes = []string{"wrong-1", "wrong-2"}
	verifier.ClearTrustedHashes()

	err = verifier.VerifyHashWithFallback(moduleName, wasmBytes, fallbackHashes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in trusted set")
}

func TestHashChain(t *testing.T) {
	chain := NewHashChain()

	// Add entries
	timestamp1 := time.Now().Unix()
	err := chain.AddEntry("v1.0.0", "hash1", timestamp1)
	require.NoError(t, err)

	timestamp2 := time.Now().Unix()
	err = chain.AddEntry("v1.0.1", "hash2", timestamp2)
	require.NoError(t, err)

	timestamp3 := time.Now().Unix()
	err = chain.AddEntry("v1.0.2", "hash3", timestamp3)
	require.NoError(t, err)

	// Verify chain integrity
	err = chain.VerifyChain()
	assert.NoError(t, err)

	// Get latest entry
	latest, err := chain.GetLatestEntry()
	require.NoError(t, err)
	assert.Equal(t, "v1.0.2", latest.Version)
	assert.Equal(t, "hash3", latest.Hash)
	assert.Equal(t, "hash2", latest.PreviousHash)
}

func TestHashChain_BrokenChain(t *testing.T) {
	chain := NewHashChain()

	// Manually create a broken chain
	chain.chain = []HashEntry{
		{Version: "v1", Hash: "hash1", PreviousHash: ""},
		{Version: "v2", Hash: "hash2", PreviousHash: "hash1"},
		{Version: "v3", Hash: "hash3", PreviousHash: "wrong-hash"}, // Broken link
	}

	err := chain.VerifyChain()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hash chain broken")
}

func TestSecurityPolicy(t *testing.T) {
	policy := DefaultSecurityPolicy()

	// Test size validation
	smallModule := make([]byte, 1024)
	err := policy.Validate(smallModule)
	assert.NoError(t, err)

	// Test oversized module
	largeModule := make([]byte, 11*1024*1024)
	err = policy.Validate(largeModule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

func TestVerificationError(t *testing.T) {
	err := &VerificationError{
		Module:       "test.wasm",
		ExpectedHash: "expected123",
		ActualHash:   "actual456",
		Reason:       "hash mismatch",
	}

	errStr := err.Error()
	assert.Contains(t, errStr, "test.wasm")
	assert.Contains(t, errStr, "hash mismatch")
	assert.Contains(t, errStr, "expected123")
	assert.Contains(t, errStr, "actual456")
}

func BenchmarkComputeHash(b *testing.B) {
	verifier := NewHashVerifier()
	data := make([]byte, 1024*1024) // 1MB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = verifier.ComputeHash(data)
	}
}

func BenchmarkVerifyHash(b *testing.B) {
	verifier := NewHashVerifier()
	data := make([]byte, 1024*1024) // 1MB
	hash := verifier.ComputeHash(data)
	verifier.AddTrustedHash("bench-module", hash)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = verifier.VerifyHash("bench-module", data)
	}
}

// Helper function to compute SHA256 hash for testing
func computeSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
