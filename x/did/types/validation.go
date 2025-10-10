package types

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sonr-io/common/webauthn"
	"github.com/sonr-io/common/webauthn/webauthncbor"
)

// ClientData represents the client data for WebAuthn ceremonies
// This is a simplified version for backward compatibility
type ClientData struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Origin    string `json:"origin"`
}

// ValidateStructure validates a WebAuthn credential structure
func ValidateStructure(cred *WebAuthnCredential) error {
	if cred == nil {
		return fmt.Errorf("credential is nil")
	}

	if cred.CredentialId == "" {
		return fmt.Errorf("credential ID is required")
	}

	if len(cred.PublicKey) == 0 {
		return fmt.Errorf("public key is required")
	}

	if cred.AttestationType == "" {
		return fmt.Errorf("attestation type is required")
	}

	return nil
}

// ValidateAttestation performs security validation of WebAuthn credential data
func ValidateAttestation(cred *WebAuthnCredential, challenge, expectedOrigin string) error {
	// Validate structure first
	if err := ValidateStructure(cred); err != nil {
		return fmt.Errorf("structure validation failed: %w", err)
	}

	// Parse and validate client data JSON
	clientData, err := ValidateClientDataJSONFormat(cred.ClientDataJson)
	if err != nil {
		return fmt.Errorf("invalid client data JSON: %w", err)
	}

	// Verify challenge
	if clientData.Challenge != challenge {
		return fmt.Errorf("challenge mismatch: expected %s, got %s", challenge, clientData.Challenge)
	}

	// Verify origin
	if expectedOrigin != "" && !strings.HasPrefix(clientData.Origin, expectedOrigin) {
		return fmt.Errorf("origin mismatch: expected %s, got %s", expectedOrigin, clientData.Origin)
	}

	// Verify ceremony type
	if clientData.Type != "webauthn.create" {
		return fmt.Errorf("invalid ceremony type: expected webauthn.create, got %s", clientData.Type)
	}

	// Validate attestation object format if present
	if cred.AttestationObject != "" {
		if err := ValidateAttestationObjectFormat(cred.AttestationObject); err != nil {
			return fmt.Errorf("invalid attestation object: %w", err)
		}
	}

	return nil
}

// ValidateForGaslessRegistration performs comprehensive validation for gasless WebAuthn registration
func ValidateForGaslessRegistration(cred *WebAuthnCredential, challenge, expectedOrigin string) error {
	// Perform standard attestation validation
	if err := ValidateAttestation(cred, challenge, expectedOrigin); err != nil {
		return fmt.Errorf("attestation validation failed: %w", err)
	}

	// Additional gasless-specific validations
	// User verification is recommended but not strictly required for gasless registration
	// The WebAuthn protocol itself provides sufficient security guarantees

	// Verify algorithm support for gasless transactions
	if err := ValidateAlgorithmSupport(cred.Algorithm); err != nil {
		return fmt.Errorf("unsupported algorithm for gasless registration: %w", err)
	}

	return nil
}

// ValidateCredentialUniqueness validates that a WebAuthn credential is unique
func ValidateCredentialUniqueness(credentialID string, existingCredentials []string) error {
	if credentialID == "" {
		return fmt.Errorf("credential ID cannot be empty")
	}

	for _, existing := range existingCredentials {
		if existing == credentialID {
			return fmt.Errorf("credential ID already exists: %s", credentialID)
		}
	}

	return nil
}

// ValidateAlgorithmSupport validates that the specified algorithm is supported
func ValidateAlgorithmSupport(algorithm int32) error {
	// Supported algorithms for Sonr
	// -7: ES256 (ECDSA with SHA-256)
	// -257: RS256 (RSA with SHA-256)
	// -8: EdDSA
	supportedAlgs := map[int32]string{
		-7:   "ES256",
		-257: "RS256",
		-8:   "EdDSA",
	}

	if _, ok := supportedAlgs[algorithm]; !ok {
		return fmt.Errorf("unsupported algorithm: %d (supported: ES256=-7, RS256=-257, EdDSA=-8)", algorithm)
	}

	return nil
}

// ValidateAttestationObjectFormat validates the attestation object format
func ValidateAttestationObjectFormat(attestationObject string) error {
	if attestationObject == "" {
		return fmt.Errorf("attestation object is empty")
	}

	// Decode base64url
	attestationBytes, err := base64.RawURLEncoding.DecodeString(attestationObject)
	if err != nil {
		// Try standard base64 as fallback
		attestationBytes, err = base64.StdEncoding.DecodeString(attestationObject)
		if err != nil {
			return fmt.Errorf("failed to decode attestation object: %w", err)
		}
	}

	// Parse as CBOR
	var attestationObj webauthn.AttestationObject
	if err := webauthncbor.Unmarshal(attestationBytes, &attestationObj); err != nil {
		return fmt.Errorf("failed to unmarshal attestation object: %w", err)
	}

	// Validate that it has authenticator data
	if len(attestationObj.RawAuthData) == 0 {
		return fmt.Errorf("attestation object missing authenticator data")
	}

	return nil
}

// ValidateClientDataJSONFormat validates the client data JSON format
func ValidateClientDataJSONFormat(clientDataJSON string) (*ClientData, error) {
	if clientDataJSON == "" {
		return nil, fmt.Errorf("client data JSON is empty")
	}

	// Decode base64url
	clientDataBytes, err := base64.RawURLEncoding.DecodeString(clientDataJSON)
	if err != nil {
		// Try standard base64 as fallback
		clientDataBytes, err = base64.StdEncoding.DecodeString(clientDataJSON)
		if err != nil {
			// Try parsing as plain JSON
			clientDataBytes = []byte(clientDataJSON)
		}
	}

	// Parse as JSON
	var clientData ClientData
	if err := json.Unmarshal(clientDataBytes, &clientData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client data JSON: %w", err)
	}

	// Validate required fields
	if clientData.Type == "" {
		return nil, fmt.Errorf("client data missing 'type' field")
	}

	if clientData.Challenge == "" {
		return nil, fmt.Errorf("client data missing 'challenge' field")
	}

	if clientData.Origin == "" {
		return nil, fmt.Errorf("client data missing 'origin' field")
	}

	return &clientData, nil
}

// GenerateAddressFromCredential generates a deterministic address from a WebAuthn credential ID
func GenerateAddressFromCredential(credentialID string) string {
	if credentialID == "" {
		return ""
	}

	// Create a SHA-256 hash of the credential ID
	hash := sha256.Sum256([]byte(credentialID))

	// Take first 20 bytes and encode as hex (Ethereum-style address)
	address := hex.EncodeToString(hash[:20])

	// Return with 0x prefix
	return "0x" + address
}

// GenerateDIDFromCredential generates a deterministic DID from a WebAuthn credential
func GenerateDIDFromCredential(credentialID, username string) string {
	if credentialID == "" {
		return ""
	}

	// Create a hash of credential ID + username for uniqueness
	data := credentialID
	if username != "" {
		data = credentialID + ":" + username
	}

	hash := sha256.Sum256([]byte(data))

	// Encode as base58-like string (using hex for simplicity)
	identifier := hex.EncodeToString(hash[:16])

	// Return DID in did:sonr format
	return fmt.Sprintf("did:sonr:%s", identifier)
}
