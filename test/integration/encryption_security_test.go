package integration_test

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/crypto/aead"
	"github.com/sonr-io/sonr/crypto/salt"
	"github.com/sonr-io/sonr/crypto/secure"
)

// TestEndToEndEncryptionWorkflow tests complete encryption workflow
func TestEndToEndEncryptionWorkflow(t *testing.T) {
	// Test data
	plaintext := []byte("sensitive data that needs protection")
	aad := []byte("additional authenticated data")

	// Generate salt
	saltObj, err := salt.Generate(32)
	require.NoError(t, err)
	require.NotNil(t, saltObj)
	require.Equal(t, 32, saltObj.Size())

	// Create AES-GCM cipher
	key := make([]byte, 32) // 256-bit key
	_, err = rand.Read(key)
	require.NoError(t, err)

	cipher, err := aead.NewAESGCM(key)
	require.NoError(t, err)

	// Encrypt data
	ciphertext, err := cipher.Encrypt(plaintext, aad)
	require.NoError(t, err)
	require.NotEqual(t, plaintext, ciphertext)

	// Decrypt data
	decrypted, err := cipher.Decrypt(ciphertext, aad)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	// Test tampering detection
	tamperedCiphertext := make([]byte, len(ciphertext))
	copy(tamperedCiphertext, ciphertext)
	tamperedCiphertext[len(tamperedCiphertext)-1] ^= 0xFF // Flip last byte

	_, err = cipher.Decrypt(tamperedCiphertext, aad)
	require.Error(t, err)
	require.Contains(t, err.Error(), "authentication failed")
}

// TestSecureMemoryWorkflow tests secure memory handling
func TestSecureMemoryWorkflow(t *testing.T) {
	// Create secure bytes
	sensitiveData := []byte("password123")
	secureData := secure.FromBytes(sensitiveData)

	// Use the data
	data := secureData.Bytes()
	require.Equal(t, sensitiveData, data)

	// Clear the secure bytes
	secureData.Clear()

	// Verify data is nil after clearing (Clear sets data to nil, not zeroed bytes)
	clearedData := secureData.Bytes()
	require.Nil(t, clearedData)
}

// TestSaltUniqueness tests that salts are unique
func TestSaltUniqueness(t *testing.T) {
	const numSalts = 100
	salts := make(map[string]bool)

	for i := 0; i < numSalts; i++ {
		s, err := salt.Generate(16)
		require.NoError(t, err)

		// Use bytes instead of String() for uniqueness check
		key := string(s.Bytes())
		require.False(t, salts[key], "Duplicate salt generated")
		salts[key] = true
	}
}

// TestEncryptionWithSaltDerivation tests key derivation with salt
func TestEncryptionWithSaltDerivation(t *testing.T) {
	password := []byte("user_password")
	plaintext := []byte("sensitive information")

	// Generate salt
	saltObj, err := salt.Generate(32)
	require.NoError(t, err)
	saltValue := saltObj.Bytes()

	// Derive key from password and salt (simplified for test)
	// In production, use proper KDF like PBKDF2 or Argon2
	key := make([]byte, 32)
	copy(key, password)
	for i := range key {
		if i < len(saltValue) {
			key[i] ^= saltValue[i]
		}
	}

	// Encrypt with derived key
	cipher, err := aead.NewAESGCM(key)
	require.NoError(t, err)

	ciphertext, err := cipher.Encrypt(plaintext, nil)
	require.NoError(t, err)

	// Decrypt with same derived key
	decrypted, err := cipher.Decrypt(ciphertext, nil)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	// Test with wrong salt (different key derivation)
	wrongSaltObj, err := salt.Generate(32)
	require.NoError(t, err)
	wrongSalt := wrongSaltObj.Bytes()

	wrongKey := make([]byte, 32)
	copy(wrongKey, password)
	for i := range wrongKey {
		if i < len(wrongSalt) {
			wrongKey[i] ^= wrongSalt[i]
		}
	}

	wrongCipher, err := aead.NewAESGCM(wrongKey)
	require.NoError(t, err)

	_, err = wrongCipher.Decrypt(ciphertext, nil)
	require.Error(t, err)
}

// TestMultipleEncryptionRounds tests multiple encryption/decryption cycles
func TestMultipleEncryptionRounds(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	cipher, err := aead.NewAESGCM(key)
	require.NoError(t, err)

	// Perform multiple rounds
	for i := 0; i < 100; i++ {
		plaintext := make([]byte, 64)
		_, err := rand.Read(plaintext)
		require.NoError(t, err)

		aad := []byte("round-specific AAD")

		ciphertext, err := cipher.Encrypt(plaintext, aad)
		require.NoError(t, err)

		decrypted, err := cipher.Decrypt(ciphertext, aad)
		require.NoError(t, err)

		require.Equal(t, plaintext, decrypted, "Round %d failed", i)
	}
}

// TestLargeDataEncryption tests encryption of large data
func TestLargeDataEncryption(t *testing.T) {
	// Create 1MB of random data
	largeData := make([]byte, 1024*1024)
	_, err := rand.Read(largeData)
	require.NoError(t, err)

	key := make([]byte, 32)
	_, err = rand.Read(key)
	require.NoError(t, err)

	cipher, err := aead.NewAESGCM(key)
	require.NoError(t, err)

	// Encrypt large data
	ciphertext, err := cipher.Encrypt(largeData, nil)
	require.NoError(t, err)

	// Decrypt and verify
	decrypted, err := cipher.Decrypt(ciphertext, nil)
	require.NoError(t, err)
	require.True(t, bytes.Equal(largeData, decrypted))
}

// BenchmarkEndToEndEncryption benchmarks complete encryption workflow
func BenchmarkEndToEndEncryption(b *testing.B) {
	plaintext := make([]byte, 1024) // 1KB data
	_, _ = rand.Read(plaintext)

	key := make([]byte, 32)
	_, _ = rand.Read(key)

	cipher, _ := aead.NewAESGCM(key)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ciphertext, _ := cipher.Encrypt(plaintext, nil)
		_, _ = cipher.Decrypt(ciphertext, nil)
	}
}

// BenchmarkSaltGeneration benchmarks salt generation
func BenchmarkSaltGeneration(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = salt.Generate(32)
	}
}

// BenchmarkSecureMemory benchmarks secure memory operations
func BenchmarkSecureMemory(b *testing.B) {
	data := make([]byte, 1024)
	_, _ = rand.Read(data)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		secureData := secure.FromBytes(data)
		_ = secureData.Bytes()
		secureData.Clear()
	}
}
