// Package ucan provides User-Controlled Authorization Networks (UCAN) implementation
// for decentralized authorization and capability delegation in the Sonr network.
// This package handles JWT-based tokens, cryptographic verification, and resource capabilities.
package ucan

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// SupportedSigningMethods returns the list of supported JWT signing methods for UCAN
func SupportedSigningMethods() []jwt.SigningMethod {
	return []jwt.SigningMethod{
		jwt.SigningMethodRS256,
		jwt.SigningMethodRS384,
		jwt.SigningMethodRS512,
		jwt.SigningMethodEdDSA,
	}
}

// ValidateSignature validates the cryptographic signature of a UCAN token
func ValidateSignature(tokenString string, verifyKey any) error {
	// Parse token without verification first to get signing method
	token, err := jwt.ParseWithClaims(
		tokenString,
		jwt.MapClaims{},
		func(token *jwt.Token) (any, error) {
			return verifyKey, nil
		},
	)
	if err != nil {
		return fmt.Errorf("signature validation failed: %w", err)
	}

	if !token.Valid {
		return fmt.Errorf("token signature is invalid")
	}

	return nil
}

// ExtractUnsignedToken extracts the unsigned portion of a JWT token (header + payload)
func ExtractUnsignedToken(tokenString string) (string, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	return strings.Join(parts[:2], "."), nil
}

// ExtractSignature extracts the signature portion of a JWT token
func ExtractSignature(tokenString string) ([]byte, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	signatureBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	return signatureBytes, nil
}

// VerifyRSASignature verifies an RSA signature using the specified hash algorithm
func VerifyRSASignature(
	signingString string,
	signature []byte,
	publicKey *rsa.PublicKey,
	hashAlg crypto.Hash,
) error {
	// Create hash of signing string
	hasher := hashAlg.New()
	hasher.Write([]byte(signingString))
	hashed := hasher.Sum(nil)

	// Verify signature
	err := rsa.VerifyPKCS1v15(publicKey, hashAlg, hashed, signature)
	if err != nil {
		return fmt.Errorf("RSA signature verification failed: %w", err)
	}

	return nil
}

// VerifyEd25519Signature verifies an Ed25519 signature
func VerifyEd25519Signature(
	signingString string,
	signature []byte,
	publicKey ed25519.PublicKey,
) error {
	valid := ed25519.Verify(publicKey, []byte(signingString), signature)
	if !valid {
		return fmt.Errorf("Ed25519 signature verification failed")
	}

	return nil
}

// GetHashAlgorithmForMethod returns the appropriate hash algorithm for a JWT signing method
func GetHashAlgorithmForMethod(method jwt.SigningMethod) (crypto.Hash, error) {
	switch method {
	case jwt.SigningMethodRS256:
		return crypto.SHA256, nil
	case jwt.SigningMethodRS384:
		return crypto.SHA384, nil
	case jwt.SigningMethodRS512:
		return crypto.SHA512, nil
	case jwt.SigningMethodEdDSA:
		// Ed25519 doesn't use a separate hash algorithm
		return crypto.Hash(0), nil
	default:
		return crypto.Hash(0), fmt.Errorf("unsupported signing method: %v", method)
	}
}

// CreateHasher creates a hasher for the given crypto.Hash algorithm
func CreateHasher(hashAlg crypto.Hash) (hash.Hash, error) {
	switch hashAlg {
	case crypto.SHA256:
		return sha256.New(), nil
	case crypto.SHA384:
		return sha512.New384(), nil
	case crypto.SHA512:
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %v", hashAlg)
	}
}

// SigningValidator provides cryptographic validation for UCAN tokens
type SigningValidator struct {
	allowedMethods map[string]jwt.SigningMethod
}

// NewSigningValidator creates a new signing validator with default allowed methods
func NewSigningValidator() *SigningValidator {
	allowed := make(map[string]jwt.SigningMethod)
	for _, method := range SupportedSigningMethods() {
		allowed[method.Alg()] = method
	}

	return &SigningValidator{
		allowedMethods: allowed,
	}
}

// NewSigningValidatorWithMethods creates a validator with specific allowed methods
func NewSigningValidatorWithMethods(methods []jwt.SigningMethod) *SigningValidator {
	allowed := make(map[string]jwt.SigningMethod)
	for _, method := range methods {
		allowed[method.Alg()] = method
	}

	return &SigningValidator{
		allowedMethods: allowed,
	}
}

// ValidateSigningMethod checks if a signing method is allowed
func (sv *SigningValidator) ValidateSigningMethod(method jwt.SigningMethod) error {
	if _, ok := sv.allowedMethods[method.Alg()]; !ok {
		return fmt.Errorf("signing method %s is not allowed", method.Alg())
	}
	return nil
}

// ValidateTokenSignature validates the cryptographic signature of a token
func (sv *SigningValidator) ValidateTokenSignature(
	tokenString string,
	keyFunc jwt.Keyfunc,
) (*jwt.Token, error) {
	// Parse with validation
	token, err := jwt.Parse(tokenString, keyFunc, jwt.WithValidMethods(sv.getAllowedMethodNames()))
	if err != nil {
		return nil, fmt.Errorf("token signature validation failed: %w", err)
	}

	// Additional signing method validation
	if err := sv.ValidateSigningMethod(token.Method); err != nil {
		return nil, err
	}

	return token, nil
}

// getAllowedMethodNames returns the names of allowed signing methods
func (sv *SigningValidator) getAllowedMethodNames() []string {
	methods := make([]string, 0, len(sv.allowedMethods))
	for name := range sv.allowedMethods {
		methods = append(methods, name)
	}
	return methods
}

// KeyValidator provides validation for cryptographic keys
type KeyValidator struct{}

// NewKeyValidator creates a new key validator
func NewKeyValidator() *KeyValidator {
	return &KeyValidator{}
}

// ValidateRSAPublicKey validates an RSA public key for UCAN usage
func (kv *KeyValidator) ValidateRSAPublicKey(key *rsa.PublicKey) error {
	if key == nil {
		return fmt.Errorf("RSA public key is nil")
	}

	// Check minimum key size (2048 bits recommended for security)
	keySize := key.N.BitLen()
	if keySize < 2048 {
		return fmt.Errorf("RSA key size too small: %d bits (minimum 2048 bits required)", keySize)
	}

	// Check maximum reasonable key size to prevent DoS
	if keySize > 8192 {
		return fmt.Errorf("RSA key size too large: %d bits (maximum 8192 bits allowed)", keySize)
	}

	return nil
}

// ValidateEd25519PublicKey validates an Ed25519 public key for UCAN usage
func (kv *KeyValidator) ValidateEd25519PublicKey(key ed25519.PublicKey) error {
	if key == nil {
		return fmt.Errorf("Ed25519 public key is nil")
	}

	if len(key) != ed25519.PublicKeySize {
		return fmt.Errorf(
			"invalid Ed25519 public key size: %d bytes (expected %d)",
			len(key),
			ed25519.PublicKeySize,
		)
	}

	return nil
}

// SignatureInfo contains information about a token's signature
type SignatureInfo struct {
	Algorithm     string
	KeyType       string
	SigningString string
	Signature     []byte
	Valid         bool
}

// ExtractSignatureInfo extracts signature information from a JWT token
func ExtractSignatureInfo(tokenString string, verifyKey any) (*SignatureInfo, error) {
	// Parse token to get method and claims
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return verifyKey, nil
	})

	var sigInfo SignatureInfo
	sigInfo.Valid = (err == nil && token.Valid)

	if token != nil {
		sigInfo.Algorithm = token.Method.Alg()

		// Get signing string
		parts := strings.Split(tokenString, ".")
		if len(parts) >= 2 {
			sigInfo.SigningString = strings.Join(parts[:2], ".")
		}

		// Get signature
		if len(parts) == 3 {
			sig, decodeErr := base64.RawURLEncoding.DecodeString(parts[2])
			if decodeErr == nil {
				sigInfo.Signature = sig
			}
		}

		// Determine key type
		switch verifyKey.(type) {
		case *rsa.PublicKey:
			sigInfo.KeyType = "RSA"
		case ed25519.PublicKey:
			sigInfo.KeyType = "Ed25519"
		default:
			sigInfo.KeyType = "Unknown"
		}
	}

	return &sigInfo, err
}

// SecurityConfig contains security configuration for UCAN validation
type SecurityConfig struct {
	AllowedSigningMethods []jwt.SigningMethod
	MinRSAKeySize         int
	MaxRSAKeySize         int
	RequireSecureAlgs     bool
}

// DefaultSecurityConfig returns a secure default configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AllowedSigningMethods: SupportedSigningMethods(),
		MinRSAKeySize:         2048,
		MaxRSAKeySize:         8192,
		RequireSecureAlgs:     true,
	}
}

// RestrictiveSecurityConfig returns a more restrictive configuration
func RestrictiveSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AllowedSigningMethods: []jwt.SigningMethod{
			jwt.SigningMethodRS256, // Only RS256 and EdDSA
			jwt.SigningMethodEdDSA,
		},
		MinRSAKeySize:     3072, // Higher minimum
		MaxRSAKeySize:     4096, // Lower maximum
		RequireSecureAlgs: true,
	}
}

// ValidateSecurityConfig validates that a security configuration is reasonable
func ValidateSecurityConfig(config *SecurityConfig) error {
	if len(config.AllowedSigningMethods) == 0 {
		return fmt.Errorf("no signing methods allowed")
	}

	if config.MinRSAKeySize < 1024 {
		return fmt.Errorf("minimum RSA key size too small: %d", config.MinRSAKeySize)
	}

	if config.MaxRSAKeySize < config.MinRSAKeySize {
		return fmt.Errorf("maximum RSA key size smaller than minimum")
	}

	if config.MaxRSAKeySize > 16384 {
		return fmt.Errorf("maximum RSA key size too large: %d", config.MaxRSAKeySize)
	}

	return nil
}
