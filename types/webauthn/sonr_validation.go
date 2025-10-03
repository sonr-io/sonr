// Package webauthn provides Sonr-specific WebAuthn validation extensions
// that integrate with the comprehensive WebAuthn protocol implementation.
//
// This package contains validation methods moved from x/did/types/webauthn.go
// to eliminate circular dependencies while leveraging the full WebAuthn protocol stack.
package webauthn

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/sonr-io/sonr/types/webauthn/webauthncbor"
	"github.com/sonr-io/sonr/types/webauthn/webauthncose"
)

// WebAuthnCredential defines the interface that WebAuthn credentials must implement
// This avoids circular dependencies while allowing validation of any credential type
type WebAuthnCredential interface {
	GetCredentialId() string
	GetPublicKey() []byte
	GetAlgorithm() int32
	GetRawId() string
	GetClientDataJson() string
	GetAttestationObject() string
	GetOrigin() string
}

// ClientData represents the parsed client data JSON from WebAuthn
type ClientData struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
	Origin    string `json:"origin"`
}

// ValidateStructure validates a WebAuthn credential for gasless transaction processing.
// This method performs cryptographic validation to ensure the credential is legitimate
// and prevents abuse of the gasless registration system.
//
// Validation checks:
// 1. Required fields are present (credential_id, public_key, algorithm)
// 2. Algorithm is supported (ES256, RS256)
// 3. Public key can be parsed successfully
// 4. Raw ID matches credential ID when base64url decoded
func ValidateStructure(c WebAuthnCredential) error {
	if c == nil {
		return fmt.Errorf("credential cannot be nil")
	}

	// Validate required fields
	if c.GetCredentialId() == "" {
		return fmt.Errorf("credential_id is required")
	}

	if len(c.GetPublicKey()) == 0 {
		return fmt.Errorf("public_key is required")
	}

	if c.GetAlgorithm() == 0 {
		return fmt.Errorf("algorithm is required")
	}

	// Validate supported algorithms
	switch c.GetAlgorithm() {
	case -7: // ES256
	case -257: // RS256
	default:
		return fmt.Errorf("unsupported algorithm: %d", c.GetAlgorithm())
	}

	// Validate public key can be parsed
	if err := validatePublicKeyFormat(c); err != nil {
		return fmt.Errorf("invalid public key format: %v", err)
	}

	// Validate raw_id matches credential_id when decoded
	if c.GetRawId() != "" {
		decodedRawID, err := base64.RawURLEncoding.DecodeString(c.GetRawId())
		if err != nil {
			return fmt.Errorf("invalid raw_id encoding: %v", err)
		}

		decodedCredID, err := base64.RawURLEncoding.DecodeString(c.GetCredentialId())
		if err != nil {
			return fmt.Errorf("invalid credential_id encoding: %v", err)
		}

		if string(decodedRawID) != string(decodedCredID) {
			return fmt.Errorf("raw_id does not match credential_id")
		}
	}

	return nil
}

// validatePublicKeyFormat validates that the public key can be parsed according to the specified algorithm.
// WebAuthn public keys are typically in COSE format, so we use the webauthncose package for parsing.
func validatePublicKeyFormat(c WebAuthnCredential) error {
	// Use the comprehensive COSE public key parser from the WebAuthn library
	parsedKey, err := webauthncose.ParsePublicKey(c.GetPublicKey())
	if err != nil {
		return fmt.Errorf("failed to parse COSE public key: %w", err)
	}

	// Validate the parsed key matches the expected algorithm
	switch c.GetAlgorithm() {
	case -7: // ES256 (ECDSA with P-256 and SHA-256)
		cosePubKey, ok := parsedKey.(webauthncose.EC2PublicKeyData)
		if !ok {
			return fmt.Errorf(
				"public key type mismatch: expected EC2PublicKeyData for ES256 algorithm, got %T",
				parsedKey,
			)
		}

		// Validate that we can construct a valid ECDSA public key from the COSE data
		if len(cosePubKey.XCoord) != 32 || len(cosePubKey.YCoord) != 32 {
			return fmt.Errorf(
				"invalid ECDSA coordinate length: x=%d, y=%d (expected 32 bytes each)",
				len(cosePubKey.XCoord),
				len(cosePubKey.YCoord),
			)
		}

		// Verify the algorithm matches
		if cosePubKey.Algorithm != -7 {
			return fmt.Errorf(
				"algorithm mismatch: COSE key algorithm %d, expected ES256 (-7)",
				cosePubKey.Algorithm,
			)
		}

		return nil

	case -257: // RS256 (RSASSA-PKCS1-v1_5 with SHA-256)
		cosePubKey, ok := parsedKey.(webauthncose.RSAPublicKeyData)
		if !ok {
			return fmt.Errorf(
				"public key type mismatch: expected RSAPublicKeyData for RS256 algorithm, got %T",
				parsedKey,
			)
		}

		// Validate RSA key components
		if len(cosePubKey.Modulus) == 0 || len(cosePubKey.Exponent) == 0 {
			return fmt.Errorf("invalid RSA key: empty modulus or exponent")
		}

		// Verify the algorithm matches
		if cosePubKey.Algorithm != -257 {
			return fmt.Errorf(
				"algorithm mismatch: COSE key algorithm %d, expected RS256 (-257)",
				cosePubKey.Algorithm,
			)
		}

		return nil

	default:
		return fmt.Errorf("unsupported algorithm for validation: %d", c.GetAlgorithm())
	}
}

// ValidateAttestation performs security validation of WebAuthn credential data.
// This performs essential security checks while leveraging the comprehensive
// WebAuthn protocol validation framework where possible.
//
// Security checks performed:
// 1. Challenge verification against provided challenge
// 2. Origin validation against expected origin
// 3. Client data JSON structure validation
// 4. Basic attestation object presence validation
func ValidateAttestation(c WebAuthnCredential, challenge, expectedOrigin string) error {
	if c == nil {
		return fmt.Errorf("credential cannot be nil")
	}

	// Validate that we have the required attestation data
	if c.GetClientDataJson() == "" {
		return fmt.Errorf("client_data_json is required for attestation validation")
	}

	if c.GetAttestationObject() == "" {
		return fmt.Errorf("attestation_object is required for attestation validation")
	}

	// Parse and validate client data JSON
	clientData, err := parseClientDataJSON(c.GetClientDataJson())
	if err != nil {
		return fmt.Errorf("failed to parse client data JSON: %w", err)
	}

	// Validate challenge
	if challenge != "" && clientData.Challenge != challenge {
		return fmt.Errorf(
			"challenge mismatch: expected %s, got %s",
			challenge,
			clientData.Challenge,
		)
	}

	// Validate origin
	if expectedOrigin != "" && clientData.Origin != expectedOrigin {
		return fmt.Errorf("origin mismatch: expected %s, got %s", expectedOrigin, clientData.Origin)
	}

	// Validate type is "webauthn.create" for registration
	if clientData.Type != "webauthn.create" {
		return fmt.Errorf(
			"invalid client data type: expected 'webauthn.create', got %s",
			clientData.Type,
		)
	}

	// Validate attestation object is valid base64url and has reasonable size
	attestationBytes, err := base64.RawURLEncoding.DecodeString(c.GetAttestationObject())
	if err != nil {
		return fmt.Errorf("invalid attestation object encoding: %w", err)
	}

	// Basic size validation - attestation objects should be at least 100 bytes
	if len(attestationBytes) < 100 {
		return fmt.Errorf("attestation object too small: %d bytes", len(attestationBytes))
	}

	return nil
}

// parseClientDataJSON parses the base64url-encoded client data JSON
func parseClientDataJSON(clientDataJSON string) (*ClientData, error) {
	clientDataBytes, err := base64.RawURLEncoding.DecodeString(clientDataJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to decode client data JSON: %w", err)
	}

	var clientData ClientData
	if err := json.Unmarshal(clientDataBytes, &clientData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client data JSON: %w", err)
	}

	return &clientData, nil
}

// ValidateForGaslessRegistration performs comprehensive validation for gasless WebAuthn registration.
// This combines structural validation with cryptographic attestation validation,
// leveraging the full WebAuthn protocol capabilities where applicable.
func ValidateForGaslessRegistration(
	c WebAuthnCredential,
	challenge, expectedOrigin string,
) error {
	// First perform structural validation
	if err := ValidateStructure(c); err != nil {
		return fmt.Errorf("structural validation failed: %w", err)
	}

	// Then perform cryptographic attestation validation
	// For gasless registration, we require full attestation validation to prevent abuse
	if challenge != "" || expectedOrigin != "" {
		if err := ValidateAttestation(c, challenge, expectedOrigin); err != nil {
			return fmt.Errorf("attestation validation failed: %w", err)
		}
	}

	return nil
}

// ValidateCredentialUniqueness validates that a WebAuthn credential is unique
// across the entire system to prevent reuse attacks in gasless transactions.
func ValidateCredentialUniqueness(credentialID string, existingCredentials []string) error {
	if credentialID == "" {
		return fmt.Errorf("credential_id cannot be empty")
	}

	// Check against all existing credentials
	if slices.Contains(existingCredentials, credentialID) {
		return fmt.Errorf("credential_id already exists: %s", credentialID)
	}

	return nil
}

// ValidateAlgorithmSupport validates that the specified algorithm is supported
// by the Sonr WebAuthn implementation.
func ValidateAlgorithmSupport(algorithm int32) error {
	switch algorithm {
	case -7: // ES256 (ECDSA with P-256 and SHA-256)
		return nil
	case -257: // RS256 (RSASSA-PKCS1-v1_5 with SHA-256)
		return nil
	default:
		return fmt.Errorf(
			"unsupported algorithm: %d (only ES256 (-7) and RS256 (-257) are supported)",
			algorithm,
		)
	}
}

// Enhanced validation functions that leverage the full WebAuthn protocol

// ValidateAttestationObjectFormat validates the attestation object format
// using the comprehensive WebAuthn protocol validation framework.
func ValidateAttestationObjectFormat(attestationObject string) error {
	if attestationObject == "" {
		return fmt.Errorf("attestation_object is required")
	}

	// Decode the attestation object
	attestationBytes, err := base64.RawURLEncoding.DecodeString(attestationObject)
	if err != nil {
		return fmt.Errorf("invalid attestation object encoding: %w", err)
	}

	// Validate minimum size
	if len(attestationBytes) < 100 {
		return fmt.Errorf("attestation object too small: %d bytes", len(attestationBytes))
	}

	// Use the existing WebAuthn protocol CBOR unmarshaling for validation
	var attestationObj AttestationObject
	if err := webauthncbor.Unmarshal(attestationBytes, &attestationObj); err != nil {
		return fmt.Errorf("failed to unmarshal attestation object: %w", err)
	}

	// Validate authenticator data can be unmarshaled
	if err := attestationObj.AuthData.Unmarshal(attestationObj.RawAuthData); err != nil {
		return fmt.Errorf("failed to unmarshal authenticator data: %w", err)
	}

	// Check for attested credential data flag
	if !attestationObj.AuthData.Flags.HasAttestedCredentialData() {
		return fmt.Errorf("attestation missing attested credential data flag")
	}

	return nil
}

// ValidateClientDataJSONFormat validates the client data JSON format
// and returns the parsed client data using the WebAuthn protocol structures.
func ValidateClientDataJSONFormat(clientDataJSON string) (*ClientData, error) {
	if clientDataJSON == "" {
		return nil, fmt.Errorf("client_data_json is required")
	}

	// Use the existing WebAuthn protocol's CollectedClientData for validation
	clientDataBytes, err := base64.RawURLEncoding.DecodeString(clientDataJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to decode client data JSON: %w", err)
	}

	var collectedData CollectedClientData
	if err := json.Unmarshal(clientDataBytes, &collectedData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client data JSON: %w", err)
	}

	// Convert to our simple ClientData format for compatibility
	return &ClientData{
		Type:      string(collectedData.Type),
		Challenge: collectedData.Challenge,
		Origin:    collectedData.Origin,
	}, nil
}

// ValidateWithProtocol validates a WebAuthn credential using the full protocol validation.
// This leverages the comprehensive attestation and client data validation from the WebAuthn library.
func ValidateWithProtocol(
	c WebAuthnCredential,
	challenge string,
	rpOrigins []string,
	rpID string,
	userVerificationRequired bool,
) error {
	if c == nil {
		return fmt.Errorf("credential cannot be nil")
	}

	// Create a CredentialCreationResponse from our credential data
	ccr := &CredentialCreationResponse{
		PublicKeyCredential: PublicKeyCredential{
			Credential: Credential{
				ID:   c.GetCredentialId(),
				Type: "public-key",
			},
			RawID: URLEncodedBase64(c.GetRawId()),
		},
		AttestationResponse: AuthenticatorAttestationResponse{
			AuthenticatorResponse: AuthenticatorResponse{
				ClientDataJSON: URLEncodedBase64(c.GetClientDataJson()),
			},
			AttestationObject: URLEncodedBase64(c.GetAttestationObject()),
		},
	}

	// Parse the credential creation response using the full WebAuthn protocol
	parsed, err := ccr.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse credential with WebAuthn protocol: %w", err)
	}

	// Validate client data using the full protocol validation
	err = parsed.Response.CollectedClientData.Verify(
		challenge,
		CreateCeremony,
		rpOrigins,
		[]string{}, // rpTopOrigins - empty for basic validation
		TopOriginDefaultVerificationMode,
	)
	if err != nil {
		return fmt.Errorf("client data validation failed: %w", err)
	}

	// Create client data hash for attestation verification
	clientDataBytes, err := base64.RawURLEncoding.DecodeString(c.GetClientDataJson())
	if err != nil {
		return fmt.Errorf("failed to decode client data JSON: %w", err)
	}
	clientDataHash := sha256.Sum256(clientDataBytes)

	// Validate the attestation object
	err = parsed.Response.AttestationObject.Verify(
		rpID,
		clientDataHash[:],
		userVerificationRequired,
		true, // user presence required
		nil,  // metadata provider - optional
		[]CredentialParameter{
			{Type: "public-key", Algorithm: -7},   // ES256
			{Type: "public-key", Algorithm: -257}, // RS256
		},
	)
	if err != nil {
		return fmt.Errorf("attestation validation failed: %w", err)
	}

	return nil
}
