package wasm

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSigner_NewSigner(t *testing.T) {
	signer, err := NewSigner()
	require.NoError(t, err)
	require.NotNil(t, signer)

	assert.NotNil(t, signer.privateKey)
	assert.NotNil(t, signer.publicKey)
	assert.Equal(t, ed25519.PrivateKeySize, len(signer.privateKey))
	assert.Equal(t, ed25519.PublicKeySize, len(signer.publicKey))
}

func TestSigner_NewSignerFromPrivateKey(t *testing.T) {
	// Generate a key pair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Create signer from private key
	signer, err := NewSignerFromPrivateKey(priv)
	require.NoError(t, err)

	assert.Equal(t, priv, signer.privateKey)
	assert.Equal(t, pub, signer.publicKey)

	// Test invalid key size
	invalidKey := []byte("too short")
	_, err = NewSignerFromPrivateKey(ed25519.PrivateKey(invalidKey))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid private key size")
}

func TestSigner_Sign(t *testing.T) {
	signer, err := NewSigner()
	require.NoError(t, err)

	// Test data
	wasmBytes := []byte("test wasm module content")

	// Sign the data
	signature, err := signer.Sign(wasmBytes)
	require.NoError(t, err)
	assert.Equal(t, ed25519.SignatureSize, len(signature))

	// Verify the signature
	valid := ed25519.Verify(signer.publicKey, wasmBytes, signature)
	assert.True(t, valid)

	// Test signing different data produces different signature
	differentData := []byte("different content")
	signature2, err := signer.Sign(differentData)
	require.NoError(t, err)
	assert.NotEqual(t, signature, signature2)
}

func TestSignatureVerifier_AddTrustedKey(t *testing.T) {
	verifier := NewSignatureVerifier()

	// Generate a key pair
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Add trusted key
	err = verifier.AddTrustedKey("test-key", pub)
	assert.NoError(t, err)

	// Test invalid key size
	invalidKey := []byte("invalid")
	err = verifier.AddTrustedKey("invalid-key", ed25519.PublicKey(invalidKey))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key size")
}

func TestSignatureVerifier_AddTrustedKeyFromHex(t *testing.T) {
	verifier := NewSignatureVerifier()

	// Generate a key pair
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Add key from hex
	hexKey := hex.EncodeToString(pub)
	err = verifier.AddTrustedKeyFromHex("hex-key", hexKey)
	assert.NoError(t, err)

	// Test invalid hex
	err = verifier.AddTrustedKeyFromHex("bad-hex", "not-hex")
	assert.Error(t, err)
}

func TestSignatureVerifier_Verify(t *testing.T) {
	// Create signer and verifier
	signer, err := NewSigner()
	require.NoError(t, err)

	verifier := NewSignatureVerifier()

	// Test data
	wasmBytes := []byte("test wasm module")

	// Sign the data
	signature, err := signer.Sign(wasmBytes)
	require.NoError(t, err)

	// Test verification without trusted keys
	err = verifier.Verify(wasmBytes, signature)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no trusted keys")

	// Add trusted key
	err = verifier.AddTrustedKey("signer1", signer.publicKey)
	require.NoError(t, err)

	// Test successful verification
	err = verifier.Verify(wasmBytes, signature)
	assert.NoError(t, err)

	// Test verification with wrong data
	wrongData := []byte("wrong data")
	err = verifier.Verify(wrongData, signature)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature verification failed")

	// Test verification with wrong signature
	wrongSignature := make([]byte, ed25519.SignatureSize)
	err = verifier.Verify(wasmBytes, wrongSignature)
	assert.Error(t, err)
}

func TestSignatureVerifier_VerifyWithKey(t *testing.T) {
	signer1, err := NewSigner()
	require.NoError(t, err)

	signer2, err := NewSigner()
	require.NoError(t, err)

	verifier := NewSignatureVerifier()
	verifier.AddTrustedKey("key1", signer1.publicKey)
	verifier.AddTrustedKey("key2", signer2.publicKey)

	wasmBytes := []byte("test module")
	signature1, _ := signer1.Sign(wasmBytes)
	signature2, _ := signer2.Sign(wasmBytes)

	// Verify with correct key
	err = verifier.VerifyWithKey("key1", wasmBytes, signature1)
	assert.NoError(t, err)

	err = verifier.VerifyWithKey("key2", wasmBytes, signature2)
	assert.NoError(t, err)

	// Verify with wrong key
	err = verifier.VerifyWithKey("key1", wasmBytes, signature2)
	assert.Error(t, err)

	// Verify with non-existent key
	err = verifier.VerifyWithKey("key3", wasmBytes, signature1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trusted key not found")
}

func TestSignatureVerifier_Management(t *testing.T) {
	verifier := NewSignatureVerifier()

	// Generate keys
	pub1, _, _ := ed25519.GenerateKey(rand.Reader)
	pub2, _, _ := ed25519.GenerateKey(rand.Reader)

	// Add keys
	verifier.AddTrustedKey("key1", pub1)
	verifier.AddTrustedKey("key2", pub2)

	// Get key IDs
	ids := verifier.GetTrustedKeyIDs()
	assert.Len(t, ids, 2)
	assert.Contains(t, ids, "key1")
	assert.Contains(t, ids, "key2")

	// Remove key
	verifier.RemoveTrustedKey("key1")
	ids = verifier.GetTrustedKeyIDs()
	assert.Len(t, ids, 1)
	assert.NotContains(t, ids, "key1")
	assert.Contains(t, ids, "key2")
}

func TestSignedModule(t *testing.T) {
	signer, err := NewSigner()
	require.NoError(t, err)

	module := []byte("test wasm module")
	signerID := "test-signer"
	version := "v1.0.0"

	// Create signed module
	signed, err := SignModule(signer, module, signerID, version)
	require.NoError(t, err)

	assert.Equal(t, module, signed.Module)
	assert.NotEmpty(t, signed.Hash)
	assert.NotEmpty(t, signed.Signature)
	assert.Equal(t, signerID, signed.SignerID)
	assert.Equal(t, version, signed.Version)
	assert.False(t, signed.Timestamp.IsZero())

	// Verify signed module
	verifier := NewSignatureVerifier()
	verifier.AddTrustedKey(signerID, signer.publicKey)

	err = VerifySignedModule(verifier, signed)
	assert.NoError(t, err)

	// Test with tampered module
	signed.Module = []byte("tampered")
	err = VerifySignedModule(verifier, signed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hash mismatch")
}

func TestSignatureManifest(t *testing.T) {
	signer, err := NewSigner()
	require.NoError(t, err)

	module := []byte("test wasm module")
	signerID := "manifest-signer"

	// Create manifest
	manifest, err := CreateSignatureManifest(module, signer, signerID)
	require.NoError(t, err)

	assert.NotEmpty(t, manifest.ModuleHash)
	assert.Len(t, manifest.Signatures, 1)
	assert.Len(t, manifest.TrustedKeys, 1)
	assert.Equal(t, signerID, manifest.Signatures[0].SignerID)
	assert.Equal(t, "Ed25519", manifest.Signatures[0].Algorithm)
	assert.Equal(t, signerID, manifest.TrustedKeys[0].KeyID)
	assert.Equal(t, "code-signing", manifest.TrustedKeys[0].Purpose)

	// Verify with manifest
	err = VerifyWithManifest(module, manifest)
	assert.NoError(t, err)

	// Test with wrong module
	wrongModule := []byte("wrong module")
	err = VerifyWithManifest(wrongModule, manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hash mismatch")

	// Test with expired manifest
	expired := time.Now().Add(-1 * time.Hour)
	manifest.ExpiresAt = &expired
	err = VerifyWithManifest(module, manifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestManifestSerialization(t *testing.T) {
	signer, err := NewSigner()
	require.NoError(t, err)

	module := []byte("test module")
	manifest, err := CreateSignatureManifest(module, signer, "test-key")
	require.NoError(t, err)

	// Export manifest
	data, err := ExportManifest(manifest)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Import manifest
	imported, err := ImportManifest(data)
	require.NoError(t, err)

	assert.Equal(t, manifest.ModuleHash, imported.ModuleHash)
	assert.Len(t, imported.Signatures, 1)
	assert.Len(t, imported.TrustedKeys, 1)

	// Verify imported manifest
	err = VerifyWithManifest(module, imported)
	assert.NoError(t, err)

	// Test invalid JSON
	_, err = ImportManifest([]byte("invalid json"))
	assert.Error(t, err)
}

func TestMultipleSignatures(t *testing.T) {
	// Create multiple signers
	signer1, _ := NewSigner()
	signer2, _ := NewSigner()

	module := []byte("multi-signed module")

	// Create manifest with first signature
	manifest, err := CreateSignatureManifest(module, signer1, "signer1")
	require.NoError(t, err)

	// Add second signature
	signature2, _ := signer2.Sign(module)
	manifest.Signatures = append(manifest.Signatures, SignatureEntry{
		Signature: base64.StdEncoding.EncodeToString(signature2),
		SignerID:  "signer2",
		Timestamp: time.Now(),
		Algorithm: "Ed25519",
	})

	manifest.TrustedKeys = append(manifest.TrustedKeys, TrustedKeyEntry{
		KeyID:     "signer2",
		PublicKey: base64.StdEncoding.EncodeToString(signer2.GetPublicKey()),
		AddedAt:   time.Now(),
		Purpose:   "code-signing",
	})

	// Verify with either signature
	err = VerifyWithManifest(module, manifest)
	assert.NoError(t, err)

	// Remove first signature and key
	manifest.Signatures = manifest.Signatures[1:]
	manifest.TrustedKeys = manifest.TrustedKeys[1:]

	// Should still verify with second signature
	err = VerifyWithManifest(module, manifest)
	assert.NoError(t, err)
}

func BenchmarkSign(b *testing.B) {
	signer, _ := NewSigner()
	data := make([]byte, 1024*1024) // 1MB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = signer.Sign(data)
	}
}

func BenchmarkVerify(b *testing.B) {
	signer, _ := NewSigner()
	verifier := NewSignatureVerifier()
	verifier.AddTrustedKey("bench", signer.publicKey)

	data := make([]byte, 1024*1024) // 1MB
	signature, _ := signer.Sign(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = verifier.Verify(data, signature)
	}
}
