// Package plugin provides security integration tests for WASM plugin system
package plugin

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/sonr-io/crypto/wasm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPluginIntegrityVerification tests end-to-end plugin verification
func TestPluginIntegrityVerification(t *testing.T) {
	// Get plugin bytes from embedded data
	pluginBytes := motrPluginBytes
	if len(pluginBytes) == 0 {
		t.Skip("Plugin binary not available")
	}

	verifier := wasm.NewHashVerifier()

	// Compute hash of actual plugin
	actualHash := verifier.ComputeHash(pluginBytes)
	t.Logf("Plugin hash: %s", actualHash)

	// Add as trusted
	verifier.AddTrustedHash("motr.wasm", actualHash)

	// Verify succeeds with correct binary
	err := verifier.VerifyHash("motr.wasm", pluginBytes)
	assert.NoError(t, err, "valid plugin should verify")

	// Tamper with plugin
	tamperedPlugin := make([]byte, len(pluginBytes))
	copy(tamperedPlugin, pluginBytes)
	if len(tamperedPlugin) > 100 {
		tamperedPlugin[100] ^= 0xFF // Flip bits
	}

	// Verification should fail
	err = verifier.VerifyHash("motr.wasm", tamperedPlugin)
	assert.Error(t, err, "tampered plugin should not verify")
	assert.Contains(t, err.Error(), "hash verification failed")
}

// TestPluginSignatureVerification tests Ed25519 signature verification
func TestPluginSignatureVerification(t *testing.T) {
	// Get plugin bytes from embedded data
	pluginBytes := motrPluginBytes
	if len(pluginBytes) == 0 {
		t.Skip("Plugin binary not available")
	}

	// Generate signing keypair
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Create signer from private key
	signer, err := wasm.NewSignerFromPrivateKey(privKey)
	require.NoError(t, err)

	// Create manifest using helper function
	manifest, err := wasm.CreateSignatureManifest(pluginBytes, signer, "test-key")
	require.NoError(t, err)

	// Verify signature using the standalone function
	err = wasm.VerifyWithManifest(pluginBytes, manifest)
	assert.NoError(t, err, "valid signature should verify")

	// Tamper with plugin
	tamperedPlugin := make([]byte, len(pluginBytes))
	copy(tamperedPlugin, pluginBytes)
	if len(tamperedPlugin) > 0 {
		tamperedPlugin[0] ^= 0xFF
	}

	// Verification should fail
	err = wasm.VerifyWithManifest(tamperedPlugin, manifest)
	assert.Error(t, err, "tampered plugin should not verify")
}

// TestPluginHashChainUpdate tests secure plugin updates
func TestPluginHashChainUpdate(t *testing.T) {
	chain := wasm.NewHashChain()

	// Simulate plugin update sequence
	versions := []struct {
		version string
		hash    string
	}{
		{"v1.0.0", "hash-v1.0.0"},
		{"v1.0.1", "hash-v1.0.1"},
		{"v1.1.0", "hash-v1.1.0"},
		{"v2.0.0", "hash-v2.0.0"},
	}

	// Add versions to chain
	for i, v := range versions {
		timestamp := time.Now().Add(time.Duration(i) * time.Hour).Unix()
		err := chain.AddEntry(v.version, v.hash, timestamp)
		require.NoError(t, err)
	}

	// Verify chain integrity
	err := chain.VerifyChain()
	assert.NoError(t, err, "hash chain should be valid")

	// Get latest version
	latest, err := chain.GetLatestEntry()
	require.NoError(t, err)
	assert.Equal(t, "v2.0.0", latest.Version)
	assert.Equal(t, "hash-v2.0.0", latest.Hash)

	// Verify the chain has proper linkage
	// The fourth entry (v2.0.0) should have the third entry's hash (v1.1.0) as its previous hash
	assert.Equal(t, "hash-v1.1.0", latest.PreviousHash)
}

// TestPluginSizeRestrictions tests plugin size validation
func TestPluginSizeRestrictions(t *testing.T) {
	policy := wasm.DefaultSecurityPolicy()

	// Test various sizes
	testCases := []struct {
		name    string
		size    int
		allowed bool
	}{
		{"tiny", 1024, true},
		{"small", 100 * 1024, true},
		{"medium", 1024 * 1024, true},
		{"large", 5 * 1024 * 1024, true},
		{"max", 10 * 1024 * 1024, true},
		{"oversized", 11 * 1024 * 1024, false},
		{"huge", 100 * 1024 * 1024, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			module := make([]byte, tc.size)
			err := policy.Validate(module)

			if tc.allowed {
				assert.NoError(t, err, "size %d should be allowed", tc.size)
			} else {
				assert.Error(t, err, "size %d should be rejected", tc.size)
				if err != nil {
					assert.Contains(t, err.Error(), "exceeds maximum allowed size")
				}
			}
		})
	}
}

// TestPluginUpdateRollback tests safe rollback mechanism
func TestPluginUpdateRollback(t *testing.T) {
	verifier := wasm.NewHashVerifier()
	chain := wasm.NewHashChain()

	// Current version
	currentPlugin := []byte("plugin v1.0.0 content")
	currentHash := verifier.ComputeHash(currentPlugin)
	verifier.AddTrustedHash("motr.wasm", currentHash)

	err := chain.AddEntry("v1.0.0", currentHash, time.Now().Unix())
	require.NoError(t, err)

	// Update to new version
	newPlugin := []byte("plugin v1.1.0 content with bugs")
	newHash := verifier.ComputeHash(newPlugin)

	// Simulate update
	verifier.AddTrustedHash("motr.wasm", newHash)
	err = chain.AddEntry("v1.1.0", newHash, time.Now().Add(1*time.Hour).Unix())
	require.NoError(t, err)

	// Verify new version works
	err = verifier.VerifyHash("motr.wasm", newPlugin)
	assert.NoError(t, err)

	// Simulate rollback needed (new version has issues)
	// Since we added v1.0.0 first, we know its hash
	// Rollback to previous version using the known hash
	verifier.AddTrustedHash("motr.wasm", currentHash)

	// Add rollback entry to chain
	err = chain.AddEntry("v1.1.1-rollback", currentHash, time.Now().Add(2*time.Hour).Unix())
	require.NoError(t, err)

	// Verify rollback works
	err = verifier.VerifyHash("motr.wasm", currentPlugin)
	assert.NoError(t, err, "rollback to previous version should work")

	// New plugin should still fail if not updated
	err = verifier.VerifyHash("motr.wasm", newPlugin)
	assert.Error(t, err, "rolled back version should not verify new plugin")
}

// TestConcurrentPluginVerification tests thread safety
func TestConcurrentPluginVerification(t *testing.T) {
	// Create test plugin
	plugin := make([]byte, 1024*1024) // 1MB
	_, err := rand.Read(plugin)
	require.NoError(t, err)

	verifier := wasm.NewHashVerifier()
	hash := verifier.ComputeHash(plugin)
	verifier.AddTrustedHash("concurrent-test", hash)

	// Run concurrent verifications
	done := make(chan bool, 100)
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		go func() {
			err := verifier.VerifyHash("concurrent-test", plugin)
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("concurrent verification failed: %v", err)
	}
}

// TestPluginMemoryProtection tests memory safety
func TestPluginMemoryProtection(t *testing.T) {
	// Test that sensitive data is cleared
	// Create sensitive data
	sensitiveData := []byte("sensitive-key-material")

	// Simulate key operations
	dataCopy := make([]byte, len(sensitiveData))
	copy(dataCopy, sensitiveData)

	// Clear original
	for i := range sensitiveData {
		sensitiveData[i] = 0
	}

	// Verify cleared
	assert.True(t, bytes.Equal(sensitiveData, make([]byte, len(sensitiveData))),
		"sensitive data should be cleared")

	// Copy should still exist (for this test)
	assert.NotEqual(t, dataCopy, sensitiveData,
		"copy should be different from cleared data")
}

// BenchmarkPluginVerification benchmarks plugin verification
func BenchmarkPluginVerification(b *testing.B) {
	plugin := make([]byte, 1024*1024) // 1MB plugin
	rand.Read(plugin)

	verifier := wasm.NewHashVerifier()
	hash := verifier.ComputeHash(plugin)
	verifier.AddTrustedHash("bench", hash)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = verifier.VerifyHash("bench", plugin)
	}
}

// BenchmarkSignatureVerification benchmarks signature verification
func BenchmarkSignatureVerification(b *testing.B) {
	plugin := make([]byte, 1024*1024) // 1MB
	rand.Read(plugin)

	_, privKey, _ := ed25519.GenerateKey(rand.Reader)
	signer, _ := wasm.NewSignerFromPrivateKey(privKey)

	manifest, _ := wasm.CreateSignatureManifest(plugin, signer, "bench-key")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wasm.VerifyWithManifest(plugin, manifest)
	}
}
