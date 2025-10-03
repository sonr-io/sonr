package spec

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sonr-io/sonr/crypto/mpc"
)

// MPCSigningMethod implements the SigningMethod interface for MPC-based signing
type MPCSigningMethod struct {
	Name    string
	enclave mpc.Enclave
}

// NewJWTSigningMethod creates a new MPC signing method with the given enclave
func NewJWTSigningMethod(name string, enclave mpc.Enclave) *MPCSigningMethod {
	return &MPCSigningMethod{
		Name:    name,
		enclave: enclave,
	}
}

// WithEnclave sets the enclave for an existing signing method
func (m *MPCSigningMethod) WithEnclave(enclave mpc.Enclave) *MPCSigningMethod {
	return &MPCSigningMethod{
		Name:    m.Name,
		enclave: enclave,
	}
}

// NewMPCSigningMethod is an alias for NewJWTSigningMethod for compatibility
func NewMPCSigningMethod(name string, enclave mpc.Enclave) *MPCSigningMethod {
	return NewJWTSigningMethod(name, enclave)
}

// Alg returns the signing method's name
func (m *MPCSigningMethod) Alg() string {
	return m.Name
}

// Verify verifies the signature using the MPC public key
func (m *MPCSigningMethod) Verify(signingString string, signature []byte, key any) error {
	// Check if enclave is available
	if m.enclave == nil {
		return fmt.Errorf("MPC enclave not available for signature verification")
	}

	// Decode the signature
	sig, err := base64.RawURLEncoding.DecodeString(string(signature))
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Hash the signing string using SHA-256
	hasher := sha256.New()
	hasher.Write([]byte(signingString))
	digest := hasher.Sum(nil)

	// Use MPC enclave to verify signature
	valid, err := m.enclave.Verify(digest, sig)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	if !valid {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// Sign signs the data using MPC
func (m *MPCSigningMethod) Sign(signingString string, key any) ([]byte, error) {
	// Check if enclave is available
	if m.enclave == nil {
		return nil, fmt.Errorf("MPC enclave not available for signing")
	}

	// Hash the signing string using SHA-256
	hasher := sha256.New()
	hasher.Write([]byte(signingString))
	digest := hasher.Sum(nil)

	// Use MPC enclave to sign the digest
	sig, err := m.enclave.Sign(digest)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with MPC: %w", err)
	}

	// Encode the signature as base64url
	encoded := base64.RawURLEncoding.EncodeToString(sig)
	return []byte(encoded), nil
}

func init() {
	// Register the MPC signing method factory
	jwt.RegisterSigningMethod("MPC256", func() jwt.SigningMethod {
		// This factory creates a new instance without enclave
		// The enclave will be provided when creating tokens
		return &MPCSigningMethod{
			Name: "MPC256",
		}
	})
}

// RegisterMPCMethod registers an MPC signing method for the given algorithm name
func RegisterMPCMethod(alg string) {
	jwt.RegisterSigningMethod(alg, func() jwt.SigningMethod {
		return &MPCSigningMethod{
			Name: alg,
		}
	})
}
