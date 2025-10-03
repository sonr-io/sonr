package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/x/dwn/types"
)

// TestEncryptionGracefulDegradation tests that encryption features are disabled when params say so
func TestEncryptionGracefulDegradation(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	// Get current params
	params, err := f.k.Params.Get(f.ctx)
	require.NoError(err)

	// Disable encryption
	params.EncryptionEnabled = false
	err = f.k.Params.Set(f.ctx, params)
	require.NoError(err)

	// Test that CheckAndPerformRotation returns nil when encryption is disabled
	encryptionSubkeeper := f.k.GetEncryptionSubkeeper()
	err = encryptionSubkeeper.CheckAndPerformRotation(f.ctx)
	require.NoError(err, "CheckAndPerformRotation should succeed when encryption is disabled")

	// Verify encryption is still disabled
	params, err = f.k.Params.Get(f.ctx)
	require.NoError(err)
	require.False(params.EncryptionEnabled, "Encryption should remain disabled")
}

// TestShouldEncryptRecord tests encryption decision logic
func TestShouldEncryptRecord(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	// Test with encryption disabled
	params, err := f.k.Params.Get(f.ctx)
	require.NoError(err)
	params.EncryptionEnabled = false
	err = f.k.Params.Set(f.ctx, params)
	require.NoError(err)

	shouldEncrypt, err := f.k.ShouldEncryptRecord(f.ctx, "test-protocol", "test-schema")
	require.NoError(err)
	require.False(shouldEncrypt, "Should not encrypt when encryption is globally disabled")

	// Test with encryption enabled but protocol not in list
	params.EncryptionEnabled = true
	params.EncryptedProtocols = []string{"encrypted-protocol"}
	params.EncryptedSchemas = []string{}
	err = f.k.Params.Set(f.ctx, params)
	require.NoError(err)

	shouldEncrypt, err = f.k.ShouldEncryptRecord(f.ctx, "test-protocol", "test-schema")
	require.NoError(err)
	require.False(shouldEncrypt, "Should not encrypt protocol not in encrypted list")

	// Test with protocol in encrypted list
	shouldEncrypt, err = f.k.ShouldEncryptRecord(f.ctx, "encrypted-protocol", "test-schema")
	require.NoError(err)
	require.True(shouldEncrypt, "Should encrypt when protocol is in encrypted list")

	// Test with schema in encrypted list
	params.EncryptedSchemas = []string{"encrypted-schema"}
	err = f.k.Params.Set(f.ctx, params)
	require.NoError(err)

	shouldEncrypt, err = f.k.ShouldEncryptRecord(f.ctx, "test-protocol", "encrypted-schema")
	require.NoError(err)
	require.True(shouldEncrypt, "Should encrypt when schema is in encrypted list")
}

// TestVRFKeysNotRequired tests that operations work when VRF keys are not available but encryption is disabled
func TestVRFKeysNotRequired(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	// Disable encryption
	params, err := f.k.Params.Get(f.ctx)
	require.NoError(err)
	params.EncryptionEnabled = false
	err = f.k.Params.Set(f.ctx, params)
	require.NoError(err)

	// Should not attempt to rotate keys when encryption is disabled
	encryptionSubkeeper := f.k.GetEncryptionSubkeeper()
	err = encryptionSubkeeper.CheckAndPerformRotation(f.ctx)
	require.NoError(err, "Should succeed without VRF keys when encryption is disabled")
}

// TestDefaultEncryptionParams tests that default params have encryption enabled
func TestDefaultEncryptionParams(t *testing.T) {
	require := require.New(t)

	params := types.DefaultParams()
	require.True(params.EncryptionEnabled, "Default params should have encryption enabled")
	require.NotEmpty(params.EncryptedProtocols, "Default params should have encrypted protocols")
	require.NotEmpty(params.EncryptedSchemas, "Default params should have encrypted schemas")
}
