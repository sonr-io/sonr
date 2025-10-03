package mpc

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnclaveData_GetData(t *testing.T) {
	// Create a new enclave
	enclave, err := NewEnclave()
	require.NoError(t, err)
	require.NotNil(t, enclave)

	// Get the data
	data := enclave.GetData()
	require.NotNil(t, data)

	// Ensure the data is the same instance
	assert.Equal(t, enclave, data.GetEnclave())

	// Ensure the data is valid
	assert.True(t, data.IsValid())
}

func TestEnclaveData_GetEnclave(t *testing.T) {
	// Create a new enclave
	enclave, err := NewEnclave()
	require.NoError(t, err)
	require.NotNil(t, enclave)

	// Get the enclave data
	data := enclave.GetData()
	require.NotNil(t, data)

	// Get the enclave back
	returnedEnclave := data.GetEnclave()
	require.NotNil(t, returnedEnclave)

	// Verify the returned enclave is the same
	assert.Equal(t, enclave, returnedEnclave)
}

func TestEnclaveData_GetPubPoint(t *testing.T) {
	// Create a new enclave
	enclave, err := NewEnclave()
	require.NoError(t, err)
	require.NotNil(t, enclave)

	// Get the enclave data
	data := enclave.GetData()
	require.NotNil(t, data)

	// Get the public point
	pubPoint, err := data.GetPubPoint()
	require.NoError(t, err)
	require.NotNil(t, pubPoint)

	// Verify the public point's serialization matches the stored public bytes
	pointBytes := pubPoint.ToAffineUncompressed()
	assert.Equal(t, data.PubBytes, pointBytes)
}

func TestEnclaveData_PubKeyHex(t *testing.T) {
	// Create a new enclave
	enclave, err := NewEnclave()
	require.NoError(t, err)
	require.NotNil(t, enclave)

	// Get the enclave data
	data := enclave.GetData()
	require.NotNil(t, data)

	// Get the public key hex
	pubKeyHex := data.PubKeyHex()
	require.NotEmpty(t, pubKeyHex)

	// Verify it's a valid hex string
	_, err = hex.DecodeString(pubKeyHex)
	require.NoError(t, err)

	// Verify it matches the stored PubHex
	assert.Equal(t, data.PubHex, pubKeyHex)
}

func TestEnclaveData_PubKeyBytes(t *testing.T) {
	// Create a new enclave
	enclave, err := NewEnclave()
	require.NoError(t, err)
	require.NotNil(t, enclave)

	// Get the enclave data
	data := enclave.GetData()
	require.NotNil(t, data)

	// Get the public key bytes
	pubKeyBytes := data.PubKeyBytes()
	require.NotEmpty(t, pubKeyBytes)

	// Verify it matches the stored PubBytes
	assert.Equal(t, data.PubBytes, pubKeyBytes)
}

func TestEnclaveData_EncryptDecrypt(t *testing.T) {
	// Create a new enclave
	enclave, err := NewEnclave()
	require.NoError(t, err)
	require.NotNil(t, enclave)

	// Get the enclave data
	data := enclave.GetData()
	require.NotNil(t, data)

	// Test key for encryption/decryption
	testKey := []byte("test-key-12345678-test-key-123456")

	// Test encryption
	encrypted, err := data.Encrypt(testKey)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	// Test decryption
	decrypted, err := data.Decrypt(testKey, encrypted)
	require.NoError(t, err)
	require.NotEmpty(t, decrypted)

	// Serialize the original data for comparison
	originalData, err := data.Marshal()
	require.NoError(t, err)

	// Verify the decrypted data matches the original
	assert.Equal(t, originalData, decrypted)

	// Test decryption with wrong key (should fail)
	wrongKey := []byte("wrong-key-12345678-wrong-key-123456")
	_, err = data.Decrypt(wrongKey, encrypted)
	assert.Error(t, err, "Decryption with wrong key should fail")
}

func TestEnclaveData_IsValid(t *testing.T) {
	// Create a new enclave
	enclave, err := NewEnclave()
	require.NoError(t, err)
	require.NotNil(t, enclave)

	// Get the enclave data
	data := enclave.GetData()
	require.NotNil(t, data)

	// Verify it's valid
	assert.True(t, data.IsValid())

	// Create an invalid enclave
	invalidEnclave := &EnclaveData{
		PubHex:   "invalid",
		PubBytes: []byte("invalid"),
		Nonce:    []byte("nonce"),
		Curve:    K256Name,
	}

	// Verify it's invalid
	assert.False(t, invalidEnclave.IsValid())
}

func TestEnclaveData_RefreshAndSign(t *testing.T) {
	// Create a new enclave
	enclave, err := NewEnclave()
	require.NoError(t, err)
	require.NotNil(t, enclave)

	// Get the original public key
	originalPubKeyHex := enclave.PubKeyHex()
	originalPubKeyBytes := enclave.PubKeyBytes()
	require.NotEmpty(t, originalPubKeyHex)
	require.NotEmpty(t, originalPubKeyBytes)

	// Sign a message with the original enclave to verify it works
	testMessage := []byte("test message before refresh")
	originalSignature, err := enclave.Sign(testMessage)
	require.NoError(t, err)
	require.NotEmpty(t, originalSignature)

	// Verify the original signature
	valid, err := enclave.Verify(testMessage, originalSignature)
	require.NoError(t, err)
	assert.True(t, valid, "Original signature should be valid")

	// Refresh the enclave
	refreshedEnclave, err := enclave.Refresh()
	require.NoError(t, err)
	require.NotNil(t, refreshedEnclave)

	// CRITICAL TEST: The public key should remain the same after refresh
	refreshedPubKeyHex := refreshedEnclave.PubKeyHex()
	refreshedPubKeyBytes := refreshedEnclave.PubKeyBytes()

	assert.Equal(t, originalPubKeyHex, refreshedPubKeyHex,
		"Public key hex should not change after refresh")
	assert.Equal(t, originalPubKeyBytes, refreshedPubKeyBytes,
		"Public key bytes should not change after refresh")

	// Verify the refreshed enclave is valid
	assert.True(t, refreshedEnclave.IsValid(), "Refreshed enclave should be valid")

	// Test that the refreshed enclave can still sign messages
	testMessage2 := []byte("test message after refresh")
	refreshedSignature, err := refreshedEnclave.Sign(testMessage2)
	require.NoError(t, err)
	require.NotEmpty(t, refreshedSignature)

	// Verify the signature from the refreshed enclave with its own key
	valid, err = refreshedEnclave.Verify(testMessage2, refreshedSignature)
	require.NoError(t, err)
	assert.True(t, valid, "Signature from refreshed enclave should be valid")

	// CRITICAL TEST: The original enclave should be able to verify the signature
	// from the refreshed enclave since they have the same public key
	valid, err = enclave.Verify(testMessage2, refreshedSignature)
	require.NoError(t, err)
	assert.True(t, valid, "Original enclave should be able to verify refreshed enclave's signature")

	// CRITICAL TEST: The refreshed enclave should be able to verify the signature
	// from the original enclave since they have the same public key
	valid, err = refreshedEnclave.Verify(testMessage, originalSignature)
	require.NoError(t, err)
	assert.True(t, valid, "Refreshed enclave should be able to verify original enclave's signature")

	// Test with wrong message (should fail)
	wrongMessage := []byte("wrong message")
	valid, err = refreshedEnclave.Verify(wrongMessage, refreshedSignature)
	require.NoError(t, err)
	assert.False(t, valid, "Wrong message verification should fail")
}

func TestEnclaveData_MarshalUnmarshal(t *testing.T) {
	// Create a new enclave
	enclave, err := NewEnclave()
	require.NoError(t, err)
	require.NotNil(t, enclave)

	// Get the enclave data
	data := enclave.GetData()
	require.NotNil(t, data)

	// Marshal the enclave
	encoded, err := data.Marshal()
	require.NoError(t, err)
	require.NotEmpty(t, encoded)

	// Create a new empty enclave
	newEnclave := &EnclaveData{}

	// Unmarshal the encoded data
	err = newEnclave.Unmarshal(encoded)
	require.NoError(t, err)

	// Verify the unmarshaled enclave matches the original
	assert.Equal(t, data.PubHex, newEnclave.PubHex)
	assert.Equal(t, data.Curve, newEnclave.Curve)
	assert.True(t, bytes.Equal(data.PubBytes, newEnclave.PubBytes))
	assert.True(t, bytes.Equal(data.Nonce, newEnclave.Nonce))
	assert.True(t, newEnclave.IsValid())

	// Verify the public key matches
	assert.Equal(t, data.PubKeyHex(), newEnclave.PubKeyHex())
}

func TestEnclaveData_Verify(t *testing.T) {
	// Create a new enclave
	enclave, err := NewEnclave()
	require.NoError(t, err)
	require.NotNil(t, enclave)

	// Sign a message
	testMessage := []byte("test message")
	signature, err := enclave.Sign(testMessage)
	require.NoError(t, err)
	require.NotEmpty(t, signature)

	// Verify the signature
	valid, err := enclave.Verify(testMessage, signature)
	require.NoError(t, err)
	assert.True(t, valid)

	// Verify with wrong message
	wrongMessage := []byte("wrong message")
	valid, err = enclave.Verify(wrongMessage, signature)
	require.NoError(t, err)
	assert.False(t, valid)

	// Corrupt the signature
	corruptedSig := make([]byte, len(signature))
	copy(corruptedSig, signature)
	corruptedSig[0] ^= 0x01 // flip a bit

	// Verify with corrupted signature (should fail)
	valid, err = enclave.Verify(testMessage, corruptedSig)
	require.NoError(t, err)
	assert.False(t, valid)

	// We don't need to manually create ECDSA signatures here
	// as we already verified the Sign and Verify functions work together.
	// This completes the verification of the enclave's signature functionality.
}
