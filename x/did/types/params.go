package types

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	errors "cosmossdk.io/errors"
)

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return Params{
		Document: &DocumentParams{
			AutoCreateVault:        true,
			MaxVerificationMethods: 20,       // Maximum verification methods per DID
			MaxServiceEndpoints:    10,       // Maximum service endpoints per DID
			MaxControllers:         5,        // Maximum controllers per DID
			DidDocumentMaxSize:     65536,    // 64KB max DID document size
			DidResolutionTimeout:   5,        // 5 seconds resolution timeout
			KeyRotationInterval:    2592000,  // 30 days in seconds
			CredentialLifetime:     31536000, // 1 year in seconds
			SupportedAssertionMethods: []string{
				"Ed25519VerificationKey2018",
				"EcdsaSecp256k1VerificationKey2019",
				"JsonWebKey2020",
			},
			SupportedAuthenticationMethods: []string{
				"Ed25519VerificationKey2018",
				"EcdsaSecp256k1VerificationKey2019",
				"JsonWebKey2020",
				"WebAuthnAuthentication2023",
			},
			SupportedInvocationMethods: []string{
				"Ed25519VerificationKey2018",
				"EcdsaSecp256k1VerificationKey2019",
			},
			SupportedDelegationMethods: []string{
				"Ed25519VerificationKey2018",
				"EcdsaSecp256k1VerificationKey2019",
			},
		},
		Webauthn: &WebauthnParams{
			ChallengeTimeout: 60, // 60 seconds (W3C recommends 60-300s)
			AllowedOrigins: []string{
				"http://localhost:8080",
				"http://localhost:8081",
				"http://localhost:8082",
				"http://localhost:8083",
				"http://localhost:8084",
				"https://localhost:8443",
			},
			SupportedAlgorithms: []string{
				"ES256", // ECDSA with P-256 and SHA-256 (COSE Algorithm -7)
				"RS256", // RSASSA-PKCS1-v1_5 with SHA-256 (COSE Algorithm -257)
				"EdDSA", // EdDSA signature algorithms (COSE Algorithm -8)
			},
			RequireUserVerification: true, // FIDO2 Level 2 certification requirement
			MaxCredentialsPerDid:    10,   // Reasonable limit to prevent resource exhaustion
			DefaultRpId:             "localhost",
			DefaultRpName:           "Sonr Identity Platform",
		},
	}
}

// Stringer method for Params.
func (p Params) String() string {
	bz, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return string(bz)
}

// Validate does the sanity check on the params.
func (p Params) Validate() error {
	// Check that nested params are not nil
	if p.Document == nil {
		return errors.Wrap(ErrInvalidParams, "document params cannot be nil")
	}
	if p.Webauthn == nil {
		return errors.Wrap(ErrInvalidParams, "webauthn params cannot be nil")
	}

	// Validate WebAuthn parameters
	if err := validateWebAuthnParams(p.Webauthn); err != nil {
		return err
	}

	// Validate DID module specific parameters
	if err := validateDIDParams(p.Document); err != nil {
		return err
	}

	return nil
}

// validateWebAuthnParams validates WebAuthn-specific parameters for FIDO2 compliance
func validateWebAuthnParams(p *WebauthnParams) error {
	// Validate challenge timeout (FIDO2: 30-300 seconds recommended)
	if p.ChallengeTimeout < 30 || p.ChallengeTimeout > 300 {
		return errors.Wrap(
			ErrInvalidParams,
			"webauthn_challenge_timeout must be between 30-300 seconds",
		)
	}

	// Validate allowed origins
	if len(p.AllowedOrigins) == 0 {
		return errors.Wrap(ErrInvalidParams, "at least one allowed_origin must be specified")
	}
	for _, origin := range p.AllowedOrigins {
		if err := validateOrigin(origin); err != nil {
			return errors.Wrapf(ErrInvalidParams, "invalid origin %s: %v", origin, err)
		}
	}

	// Validate supported algorithms
	if len(p.SupportedAlgorithms) == 0 {
		return errors.Wrap(ErrInvalidParams, "at least one supported_algorithm must be specified")
	}
	for _, algo := range p.SupportedAlgorithms {
		if !isValidCOSEAlgorithm(algo) {
			return errors.Wrapf(ErrInvalidParams, "unsupported algorithm: %s", algo)
		}
	}

	// Validate max credentials per DID (prevent resource exhaustion)
	if p.MaxCredentialsPerDid < 1 || p.MaxCredentialsPerDid > 100 {
		return errors.Wrap(ErrInvalidParams, "max_credentials_per_did must be between 1-100")
	}

	// Validate RP ID (must be valid domain or "localhost")
	if err := validateRPID(p.DefaultRpId); err != nil {
		return errors.Wrapf(ErrInvalidParams, "invalid default_rp_id: %v", err)
	}

	// Validate RP Name
	if len(p.DefaultRpName) == 0 || len(p.DefaultRpName) > 256 {
		return errors.Wrap(ErrInvalidParams, "default_rp_name must be between 1-256 characters")
	}

	return nil
}

// validateDIDParams validates DID-specific module parameters
func validateDIDParams(p *DocumentParams) error {
	// Validate max verification methods (1-50)
	if p.MaxVerificationMethods < 1 || p.MaxVerificationMethods > 50 {
		return errors.Wrap(
			ErrInvalidParams,
			"max_verification_methods must be between 1-50",
		)
	}

	// Validate max service endpoints (0-20)
	if p.MaxServiceEndpoints < 0 || p.MaxServiceEndpoints > 20 {
		return errors.Wrap(
			ErrInvalidParams,
			"max_service_endpoints must be between 0-20",
		)
	}

	// Validate max controllers (1-10)
	if p.MaxControllers < 1 || p.MaxControllers > 10 {
		return errors.Wrap(
			ErrInvalidParams,
			"max_controllers must be between 1-10",
		)
	}

	// Validate DID document size limits (1KB-100KB)
	if p.DidDocumentMaxSize < 1024 || p.DidDocumentMaxSize > 102400 {
		return errors.Wrap(
			ErrInvalidParams,
			"did_document_max_size must be between 1024-102400 bytes (1KB-100KB)",
		)
	}

	// Validate DID resolution timeout (1-30 seconds)
	if p.DidResolutionTimeout < 1 || p.DidResolutionTimeout > 30 {
		return errors.Wrap(
			ErrInvalidParams,
			"did_resolution_timeout must be between 1-30 seconds",
		)
	}

	// Validate key rotation interval (1 day - 1 year in seconds)
	if p.KeyRotationInterval < 86400 || p.KeyRotationInterval > 31536000 {
		return errors.Wrap(
			ErrInvalidParams,
			"key_rotation_interval must be between 86400-31536000 seconds (1 day - 1 year)",
		)
	}

	// Validate credential lifetime (1 hour - 10 years in seconds)
	if p.CredentialLifetime < 3600 || p.CredentialLifetime > 315360000 {
		return errors.Wrap(
			ErrInvalidParams,
			"credential_lifetime must be between 3600-315360000 seconds (1 hour - 10 years)",
		)
	}

	// Validate supported assertion methods
	if len(p.SupportedAssertionMethods) == 0 {
		return errors.Wrap(
			ErrInvalidParams,
			"at least one supported_assertion_method must be specified",
		)
	}

	// Validate supported authentication methods
	if len(p.SupportedAuthenticationMethods) == 0 {
		return errors.Wrap(
			ErrInvalidParams,
			"at least one supported_authentication_method must be specified",
		)
	}

	return nil
}

// validateOrigin validates that an origin is a valid URL with http/https scheme
func validateOrigin(origin string) error {
	u, err := url.Parse(origin)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check scheme
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("origin must use http or https scheme")
	}

	// Check host is present
	if u.Host == "" {
		return fmt.Errorf("origin must have a host")
	}

	// Path should be empty for origins
	if u.Path != "" && u.Path != "/" {
		return fmt.Errorf("origin should not include path")
	}

	return nil
}

// isValidCOSEAlgorithm checks if the algorithm is a valid COSE algorithm identifier
func isValidCOSEAlgorithm(algo string) bool {
	// Valid COSE algorithms for WebAuthn
	// Reference: https://www.w3.org/TR/webauthn-3/#sctn-alg-identifier
	validAlgorithms := map[string]bool{
		"ES256": true, // ECDSA with P-256 and SHA-256 (-7)
		"ES384": true, // ECDSA with P-384 and SHA-384 (-35)
		"ES512": true, // ECDSA with P-521 and SHA-512 (-36)
		"RS256": true, // RSASSA-PKCS1-v1_5 with SHA-256 (-257)
		"RS384": true, // RSASSA-PKCS1-v1_5 with SHA-384 (-258)
		"RS512": true, // RSASSA-PKCS1-v1_5 with SHA-512 (-259)
		"PS256": true, // RSASSA-PSS with SHA-256 (-37)
		"PS384": true, // RSASSA-PSS with SHA-384 (-38)
		"PS512": true, // RSASSA-PSS with SHA-512 (-39)
		"EdDSA": true, // EdDSA signature algorithms (-8)
	}
	return validAlgorithms[algo]
}

// validateRPID validates the Relying Party ID according to WebAuthn specs
func validateRPID(rpID string) error {
	if rpID == "" {
		return fmt.Errorf("rp_id cannot be empty")
	}

	// localhost is valid for development
	if rpID == "localhost" {
		return nil
	}

	// Check if it's a valid domain
	// Must not contain scheme, port, or path
	if strings.Contains(rpID, "://") || strings.Contains(rpID, "/") {
		return fmt.Errorf("rp_id must be a domain name without scheme or path")
	}

	// Basic domain validation
	parts := strings.Split(rpID, ".")
	if len(parts) < 2 && rpID != "localhost" {
		return fmt.Errorf("rp_id must be a valid domain")
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return fmt.Errorf("invalid domain label length")
		}
	}

	return nil
}
