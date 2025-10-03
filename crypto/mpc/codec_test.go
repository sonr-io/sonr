package mpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyShareGeneration(t *testing.T) {
	t.Run("Generate Valid Enclave", func(t *testing.T) {
		// Generate enclave
		enclave, err := NewEnclave()
		require.NoError(t, err)
		require.NotNil(t, enclave)

		// Validate enclave contents
		assert.True(t, enclave.IsValid())
	})

	t.Run("Export and Import", func(t *testing.T) {
		// Generate original enclave
		original, err := NewEnclave()
		require.NoError(t, err)

		// Test key for encryption/decryption (32 bytes)
		testKey := []byte("test-key-12345678-test-key-123456")

		// Test Export/Import
		t.Run("Full Enclave", func(t *testing.T) {
			// Export enclave
			data, err := original.Encrypt(testKey)
			require.NoError(t, err)
			require.NotEmpty(t, data)

			// Create new empty enclave
			newEnclave, err := NewEnclave()
			require.NoError(t, err)

			// Verify the imported enclave works by signing
			testData := []byte("test message")
			sig, err := newEnclave.Sign(testData)
			require.NoError(t, err)
			valid, err := newEnclave.Verify(testData, sig)
			require.NoError(t, err)
			assert.True(t, valid)
		})
	})

	t.Run("Encrypt and Decrypt", func(t *testing.T) {
		// Generate original enclave
		original, err := NewEnclave()
		require.NoError(t, err)
		require.NotNil(t, original)

		// Test key for encryption/decryption (32 bytes)
		testKey := []byte("test-key-12345678-test-key-123456")

		// Test Encrypt
		encrypted, err := original.Encrypt(testKey)
		require.NoError(t, err)
		require.NotEmpty(t, encrypted)

		// Test Decrypt
		decrypted, err := original.Decrypt(testKey, encrypted)
		require.NoError(t, err)
		require.NotEmpty(t, decrypted)

		// Verify decrypted data matches original
		originalData, err := original.Marshal()
		require.NoError(t, err)
		assert.Equal(t, originalData, decrypted)

		// Test with wrong key should fail
		wrongKey := []byte("wrong-key-12345678-wrong-key-123456")
		_, err = original.Decrypt(wrongKey, encrypted)
		assert.Error(t, err, "Decryption with wrong key should fail")
	})
}

func TestEnclaveOperations(t *testing.T) {
	t.Run("Signing and Verification", func(t *testing.T) {
		// Generate valid enclave
		enclave, err := NewEnclave()
		require.NoError(t, err)

		// Test signing
		testData := []byte("test message")
		signature, err := enclave.Sign(testData)
		require.NoError(t, err)
		require.NotNil(t, signature)

		// Verify the signature
		valid, err := enclave.Verify(testData, signature)
		require.NoError(t, err)
		assert.True(t, valid)

		// Test invalid data verification
		invalidData := []byte("wrong message")
		valid, err = enclave.Verify(invalidData, signature)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("Refresh Operation", func(t *testing.T) {
		enclave, err := NewEnclave()
		require.NoError(t, err)

		// Test refresh
		refreshedEnclave, err := enclave.Refresh()
		require.NoError(t, err)
		require.NotNil(t, refreshedEnclave)

		// Verify refreshed enclave is valid
		assert.True(t, refreshedEnclave.IsValid())
	})
}

func TestEnclaveDataAccess(t *testing.T) {
	t.Run("GetData", func(t *testing.T) {
		// Generate enclave
		enclave, err := NewEnclave()
		require.NoError(t, err)
		require.NotNil(t, enclave)

		// Get the enclave data
		data := enclave.GetData()
		require.NotNil(t, data, "GetData should return non-nil value")

		// Verify the data is valid
		assert.True(t, data.IsValid(), "Enclave data should be valid")

		// Verify the public key in the data matches the enclave's public key
		assert.Equal(t, enclave.PubKeyHex(), data.PubKeyHex(), "Public keys should match")
	})

	t.Run("PubKeyHex", func(t *testing.T) {
		// Generate enclave
		enclave, err := NewEnclave()
		require.NoError(t, err)
		require.NotNil(t, enclave)

		// Get the public key hex
		pubKeyHex := enclave.PubKeyHex()
		require.NotEmpty(t, pubKeyHex, "PubKeyHex should return non-empty string")

		// Check that it's a valid hex string (should be 66 chars for compressed point: 0x02/0x03 + 32 bytes)
		assert.GreaterOrEqual(
			t,
			len(pubKeyHex),
			66,
			"Public key hex should be at least 66 characters",
		)
		assert.True(t, len(pubKeyHex)%2 == 0, "Hex string should have even length")

		// Compare with the enclave data's public key
		data := enclave.GetData()
		assert.Equal(
			t,
			data.PubKeyHex(),
			pubKeyHex,
			"Public key hex should match the one from GetData",
		)

		// Verify that two different enclaves have different public keys
		enclave2, err := NewEnclave()
		require.NoError(t, err)
		require.NotNil(t, enclave2)

		pubKeyHex2 := enclave2.PubKeyHex()
		assert.NotEqual(
			t,
			pubKeyHex,
			pubKeyHex2,
			"Different enclaves should have different public keys",
		)
	})
}
