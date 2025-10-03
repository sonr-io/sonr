// Package aead provides authenticated encryption with associated data (AEAD) implementations
// following NIST SP 800-38D standards for secure data encryption and integrity verification.
package aead

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

const (
	// NonceSize defines the standard 96-bit nonce size for optimal GCM performance
	NonceSize = 12
	// TagSize defines the 128-bit authentication tag size for GCM
	TagSize = 16
	// KeySize defines the AES-256 key size
	KeySize = 32
)

// AESGCMCipher wraps AES-GCM operations with secure defaults
type AESGCMCipher struct {
	gcm cipher.AEAD
}

// NewAESGCM creates a new AES-GCM cipher with the provided 256-bit key
func NewAESGCM(key []byte) (*AESGCMCipher, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", KeySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	return &AESGCMCipher{gcm: gcm}, nil
}

// Encrypt encrypts plaintext with additional authenticated data (AAD) using AES-GCM
// Returns nonce + ciphertext + tag concatenated for easy storage
func (a *AESGCMCipher) Encrypt(plaintext, aad []byte) ([]byte, error) {
	// Generate cryptographically secure random nonce
	nonce, err := generateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt with GCM (includes authentication tag)
	ciphertext := a.gcm.Seal(nil, nonce, plaintext, aad)

	// Prepend nonce to ciphertext for transport
	result := make([]byte, NonceSize+len(ciphertext))
	copy(result[:NonceSize], nonce)
	copy(result[NonceSize:], ciphertext)

	return result, nil
}

// Decrypt decrypts and authenticates ciphertext with AAD using AES-GCM
// Expects input format: nonce + ciphertext + tag
func (a *AESGCMCipher) Decrypt(data, aad []byte) ([]byte, error) {
	if len(data) < NonceSize+TagSize {
		return nil, fmt.Errorf("invalid ciphertext length: minimum %d bytes required", NonceSize+TagSize)
	}

	// Extract nonce and ciphertext
	nonce := data[:NonceSize]
	ciphertext := data[NonceSize:]

	// Decrypt and verify authentication tag
	plaintext, err := a.gcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, fmt.Errorf("decryption and authentication failed: %w", err)
	}

	return plaintext, nil
}

// EncryptWithNonce encrypts plaintext using a provided nonce (for testing purposes)
// WARNING: Nonce reuse can compromise security. Use only for testing.
func (a *AESGCMCipher) EncryptWithNonce(plaintext, aad, nonce []byte) ([]byte, error) {
	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d bytes, got %d", NonceSize, len(nonce))
	}

	// Encrypt with provided nonce
	ciphertext := a.gcm.Seal(nil, nonce, plaintext, aad)

	// Prepend nonce to ciphertext
	result := make([]byte, NonceSize+len(ciphertext))
	copy(result[:NonceSize], nonce)
	copy(result[NonceSize:], ciphertext)

	return result, nil
}

// generateNonce creates a cryptographically secure 96-bit nonce
func generateNonce() ([]byte, error) {
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to read random bytes: %w", err)
	}
	return nonce, nil
}

// GetNonceSize returns the nonce size used by this cipher
func (a *AESGCMCipher) GetNonceSize() int {
	return NonceSize
}

// GetTagSize returns the authentication tag size
func (a *AESGCMCipher) GetTagSize() int {
	return TagSize
}
