package context

import (
	"os"
	"path/filepath"
	"testing"

	"cosmossdk.io/log"
	"github.com/stretchr/testify/require"

	"github.com/sonr-io/crypto/vrf"
)

// TestSonrContextInitialization tests SonrContext initialization with VRF keys
func TestSonrContextInitialization(t *testing.T) {
	require := require.New(t)

	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Test initialization without VRF keys
	t.Setenv("HOME", tmpDir)
	logger := log.NewNopLogger()
	ctx := NewSonrContext(logger)

	err := ctx.Initialize()
	require.Error(err, "Should fail to initialize without VRF keys")
	require.False(ctx.IsInitialized(), "Context should not be initialized without VRF keys")
}

// TestSonrContextWithValidKeys tests SonrContext with valid VRF keys
func TestSonrContextWithValidKeys(t *testing.T) {
	require := require.New(t)

	// Create temporary directory for test
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create .sonr directory
	sonrDir := filepath.Join(tmpDir, ".sonr")
	err := os.MkdirAll(sonrDir, 0o750)
	require.NoError(err)

	// Generate VRF keys
	privateKey, err := vrf.GenerateKey(nil)
	require.NoError(err)

	// Write VRF keys
	vrfKeyPath := filepath.Join(sonrDir, "vrf_secret.key")
	err = os.WriteFile(vrfKeyPath, privateKey, 0o600)
	require.NoError(err)

	// Test initialization with valid VRF keys
	logger := log.NewNopLogger()
	ctx := NewSonrContext(logger)

	err = ctx.Initialize()
	require.NoError(err, "Should initialize successfully with valid VRF keys")
	require.True(ctx.IsInitialized(), "Context should be initialized")

	// Test VRF key retrieval
	privKey, err := ctx.GetVRFPrivateKey()
	require.NoError(err)
	require.Len(privKey, vrf.PrivateKeySize)

	pubKey, err := ctx.GetVRFPublicKey()
	require.NoError(err)
	require.Len(pubKey, vrf.PublicKeySize)
}

// TestSonrContextInvalidKeySize tests handling of invalid key size
func TestSonrContextInvalidKeySize(t *testing.T) {
	require := require.New(t)

	// Create temporary directory for test
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create .sonr directory
	sonrDir := filepath.Join(tmpDir, ".sonr")
	err := os.MkdirAll(sonrDir, 0o750)
	require.NoError(err)

	// Write invalid VRF key (wrong size)
	vrfKeyPath := filepath.Join(sonrDir, "vrf_secret.key")
	invalidKey := make([]byte, 32) // Should be 64 bytes
	err = os.WriteFile(vrfKeyPath, invalidKey, 0o600)
	require.NoError(err)

	// Test initialization with invalid key size
	logger := log.NewNopLogger()
	ctx := NewSonrContext(logger)

	err = ctx.Initialize()
	require.Error(err, "Should fail to initialize with invalid key size")
	require.Contains(err.Error(), "invalid VRF private key size")
	require.False(ctx.IsInitialized(), "Context should not be initialized with invalid keys")
}

// TestSonrContextThreadSafety tests thread-safe access to VRF keys
func TestSonrContextThreadSafety(t *testing.T) {
	require := require.New(t)

	// Create temporary directory for test
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create .sonr directory with valid keys
	sonrDir := filepath.Join(tmpDir, ".sonr")
	err := os.MkdirAll(sonrDir, 0o750)
	require.NoError(err)

	privateKey, err := vrf.GenerateKey(nil)
	require.NoError(err)

	vrfKeyPath := filepath.Join(sonrDir, "vrf_secret.key")
	err = os.WriteFile(vrfKeyPath, privateKey, 0o600)
	require.NoError(err)

	// Initialize context
	logger := log.NewNopLogger()
	ctx := NewSonrContext(logger)
	err = ctx.Initialize()
	require.NoError(err)

	// Test concurrent access (simple check - not exhaustive)
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := ctx.GetVRFPrivateKey()
			require.NoError(err)
			_, err = ctx.GetVRFPublicKey()
			require.NoError(err)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestSonrContextErrorMessages tests that error messages are helpful
func TestSonrContextErrorMessages(t *testing.T) {
	require := require.New(t)

	// Create temporary directory for test
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	logger := log.NewNopLogger()
	ctx := NewSonrContext(logger)

	err := ctx.Initialize()
	require.Error(err)
	require.Contains(err.Error(), "failed to read VRF secret key")
	require.Contains(err.Error(), "VRF keys are required")
	require.Contains(err.Error(), "snrd init")
}
