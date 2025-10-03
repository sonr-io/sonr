// Package webauthn provides Sonr-specific WebAuthn utility functions
// that support gasless transaction processing and DID generation.
//
// This package contains utility functions moved from app/ante/webauthn_gasless.go
// to eliminate circular dependencies while providing core WebAuthn functionality.
package webauthn

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenerateAddressFromCredential generates a deterministic address from a WebAuthn credential ID.
// This ensures the same credential always generates the same address, allowing for
// predictable account creation without requiring pre-existing blockchain state.
//
// This function is used by the gasless WebAuthn registration system to create
// controller addresses automatically when users don't have existing accounts.
func GenerateAddressFromCredential(credentialID string) sdk.AccAddress {
	// Create a deterministic hash from the credential ID
	// Add a domain separator to prevent collisions with other address generation methods
	domainSeparator := "webauthn_gasless_v1"
	data := domainSeparator + credentialID

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(data))

	// Take the first 20 bytes for the address (Ethereum-compatible)
	return sdk.AccAddress(hash[:20])
}

// GenerateDIDFromCredential generates a deterministic DID from a WebAuthn credential.
// This creates a unique, reproducible DID for each WebAuthn credential.
//
// The generated DID follows the format: did:sonr:<hex-encoded-hash-prefix>
// where the hash is derived from the credential ID and username.
func GenerateDIDFromCredential(credentialID string, username string) string {
	// Create a deterministic hash from credential ID and username
	data := credentialID + ":" + username
	hash := sha256.Sum256([]byte(data))

	// Create a DID with the sonr method
	// Format: did:sonr:<hex-encoded-hash-prefix>
	didSuffix := hex.EncodeToString(hash[:16]) // Use first 16 bytes for shorter DIDs
	return fmt.Sprintf("did:sonr:%s", didSuffix)
}

// GenerateVerificationMethodID creates a verification method ID for WebAuthn credentials.
// This follows DID standards for verification method identifiers.
func GenerateVerificationMethodID(did, credentialID string) string {
	// Create a short hash from the credential ID for uniqueness
	hash := sha256.Sum256([]byte(credentialID))
	hashSuffix := hex.EncodeToString(hash[:8]) // Use first 8 bytes for compactness

	return fmt.Sprintf("%s#webauthn-%s", did, hashSuffix)
}

// IsValidCredentialID validates that a credential ID meets Sonr's requirements.
// WebAuthn credential IDs should be base64url-encoded and of reasonable length.
func IsValidCredentialID(credentialID string) bool {
	if credentialID == "" {
		return false
	}

	// Credential IDs should be at least 16 characters (reasonable minimum)
	// and not exceed 1024 characters (reasonable maximum)
	if len(credentialID) < 16 || len(credentialID) > 1024 {
		return false
	}

	// Basic check for base64url characters
	// WebAuthn credential IDs are typically base64url encoded
	for _, r := range credentialID {
		if !((r >= 'A' && r <= 'Z') ||
			(r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_') {
			return false
		}
	}

	return true
}

// GenerateChallengeHash creates a deterministic challenge hash for WebAuthn operations.
// This can be used for challenge generation in situations where deterministic challenges are needed.
func GenerateChallengeHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// ExtractOriginFromURL extracts the origin from a URL for WebAuthn origin validation.
// Returns the origin in the format expected by WebAuthn (e.g., "https://example.com").
func ExtractOriginFromURL(url string) string {
	// Simple origin extraction - in production this would use proper URL parsing
	// For now, return the input assuming it's already a valid origin
	return url
}

// ValidateOriginFormat validates that an origin meets WebAuthn requirements.
func ValidateOriginFormat(origin string) error {
	if origin == "" {
		return fmt.Errorf("origin cannot be empty")
	}

	// WebAuthn origins must be HTTPS (except localhost for development)
	if origin != "http://localhost" &&
		origin != "http://127.0.0.1" &&
		len(origin) >= 8 &&
		origin[:8] != "https://" {
		return fmt.Errorf("WebAuthn origins must use HTTPS (or localhost for development)")
	}

	return nil
}

// CreateDeterministicSeed creates a deterministic seed from multiple inputs.
// This is useful for generating consistent values across different operations.
func CreateDeterministicSeed(inputs ...string) []byte {
	var combined string
	for _, input := range inputs {
		combined += input + ":"
	}

	hash := sha256.Sum256([]byte(combined))
	return hash[:]
}

// FormatCredentialForDisplay formats a credential ID for user-friendly display.
// Truncates long credential IDs while preserving uniqueness for display purposes.
func FormatCredentialForDisplay(credentialID string) string {
	if len(credentialID) <= 16 {
		return credentialID
	}

	// Show first 8 and last 8 characters with ellipsis in between
	return fmt.Sprintf("%s...%s", credentialID[:8], credentialID[len(credentialID)-8:])
}

// ValidateUsernameFormat validates that a username meets Sonr's requirements.
func ValidateUsernameFormat(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}

	if len(username) > 32 {
		return fmt.Errorf("username cannot exceed 32 characters")
	}

	// Check for valid characters (alphanumeric and some special characters)
	for _, r := range username {
		if !((r >= 'A' && r <= 'Z') ||
			(r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.') {
			return fmt.Errorf("username contains invalid character: %c", r)
		}
	}

	return nil
}

// GenerateAccountSeq generates a deterministic account sequence for gasless transactions.
// This ensures consistent account sequence handling for WebAuthn gasless transactions.
func GenerateAccountSeq(credentialID string, blockHeight int64) uint64 {
	data := fmt.Sprintf("%s:%d", credentialID, blockHeight)
	hash := sha256.Sum256([]byte(data))

	// Convert first 8 bytes to uint64
	var seq uint64
	for i := 0; i < 8; i++ {
		seq = seq<<8 + uint64(hash[i])
	}

	// Ensure it's not zero
	if seq == 0 {
		seq = 1
	}

	return seq
}
