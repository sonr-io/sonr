// Package ecdsa provides ECDSA signature canonicalization
package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"math/big"
)

// CanonicalizeSignature ensures ECDSA signature is in canonical form
// This prevents signature malleability attacks where (r, s) and (r, -s mod N) are both valid
func CanonicalizeSignature(r, s *big.Int, curve elliptic.Curve) (*big.Int, *big.Int, error) {
	if r == nil || s == nil {
		return nil, nil, fmt.Errorf("r and s cannot be nil")
	}

	if curve == nil {
		return nil, nil, fmt.Errorf("curve cannot be nil")
	}

	N := curve.Params().N
	if N == nil {
		return nil, nil, fmt.Errorf("invalid curve parameters")
	}

	// Create copies to avoid modifying originals
	rCopy := new(big.Int).Set(r)
	sCopy := new(big.Int).Set(s)

	// Check if r is in valid range [1, N-1]
	if rCopy.Sign() <= 0 || rCopy.Cmp(N) >= 0 {
		return nil, nil, fmt.Errorf("r is not in valid range [1, N-1]")
	}

	// Check if s is in valid range [1, N-1]
	if sCopy.Sign() <= 0 || sCopy.Cmp(N) >= 0 {
		return nil, nil, fmt.Errorf("s is not in valid range [1, N-1]")
	}

	// Ensure s is canonical (s <= N/2)
	halfN := new(big.Int).Div(N, big.NewInt(2))
	if sCopy.Cmp(halfN) > 0 {
		// Use N - s to get canonical form
		sCopy.Sub(N, sCopy)
	}

	return rCopy, sCopy, nil
}

// IsSignatureCanonical checks if an ECDSA signature is in canonical form
func IsSignatureCanonical(r, s *big.Int, curve elliptic.Curve) bool {
	if r == nil || s == nil || curve == nil {
		return false
	}

	N := curve.Params().N
	if N == nil {
		return false
	}

	// Check r is in valid range [1, N-1]
	if r.Sign() <= 0 || r.Cmp(N) >= 0 {
		return false
	}

	// Check s is in valid range [1, N/2]
	halfN := new(big.Int).Div(N, big.NewInt(2))
	if s.Sign() <= 0 || s.Cmp(halfN) > 0 {
		return false
	}

	return true
}

// ValidateAndCanonicalizeSignature validates and canonicalizes an ECDSA signature
func ValidateAndCanonicalizeSignature(pub *ecdsa.PublicKey, hash []byte, r, s *big.Int) (*big.Int, *big.Int, error) {
	if pub == nil {
		return nil, nil, fmt.Errorf("public key cannot be nil")
	}

	if len(hash) == 0 {
		return nil, nil, fmt.Errorf("hash cannot be empty")
	}

	// Canonicalize the signature
	rCanon, sCanon, err := CanonicalizeSignature(r, s, pub.Curve)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to canonicalize signature: %w", err)
	}

	// Verify the canonical signature
	if !ecdsa.Verify(pub, hash, rCanon, sCanon) {
		// If canonical signature doesn't verify, try the original
		// This handles the case where the signature was already canonical but negated
		if !ecdsa.Verify(pub, hash, r, s) {
			return nil, nil, fmt.Errorf("signature verification failed")
		}
		// Original verified, return it canonicalized
		return CanonicalizeSignature(r, s, pub.Curve)
	}

	return rCanon, sCanon, nil
}

// RejectNonCanonical rejects non-canonical signatures outright
// This is stricter than canonicalization and prevents accepting malleable signatures
func RejectNonCanonical(r, s *big.Int, curve elliptic.Curve) error {
	if r == nil || s == nil {
		return fmt.Errorf("r and s cannot be nil")
	}

	if curve == nil {
		return fmt.Errorf("curve cannot be nil")
	}

	if !IsSignatureCanonical(r, s, curve) {
		return fmt.Errorf("signature is not in canonical form")
	}

	return nil
}

// NormalizeSignature normalizes an ECDSA signature to ensure consistent representation
// This is useful for signature aggregation and comparison
func NormalizeSignature(r, s *big.Int, curve elliptic.Curve) (*big.Int, *big.Int, error) {
	// First canonicalize
	rNorm, sNorm, err := CanonicalizeSignature(r, s, curve)
	if err != nil {
		return nil, nil, err
	}

	// Additional normalization can be added here if needed
	// For example, ensuring consistent byte representation

	return rNorm, sNorm, nil
}

// CompareSignatures compares two ECDSA signatures for equality after canonicalization
func CompareSignatures(r1, s1, r2, s2 *big.Int, curve elliptic.Curve) (bool, error) {
	// Canonicalize both signatures
	r1Canon, s1Canon, err := CanonicalizeSignature(r1, s1, curve)
	if err != nil {
		return false, fmt.Errorf("failed to canonicalize first signature: %w", err)
	}

	r2Canon, s2Canon, err := CanonicalizeSignature(r2, s2, curve)
	if err != nil {
		return false, fmt.Errorf("failed to canonicalize second signature: %w", err)
	}

	// Compare canonical forms
	return r1Canon.Cmp(r2Canon) == 0 && s1Canon.Cmp(s2Canon) == 0, nil
}

// SignatureBytes converts signature to bytes in canonical form
// Returns 64 bytes for P-256 (32 bytes for r, 32 bytes for s)
func SignatureBytes(r, s *big.Int, curve elliptic.Curve) ([]byte, error) {
	// Canonicalize first
	rCanon, sCanon, err := CanonicalizeSignature(r, s, curve)
	if err != nil {
		return nil, err
	}

	// Get the byte size for the curve
	byteSize := (curve.Params().BitSize + 7) / 8

	// Convert to bytes with proper padding
	rBytes := rCanon.Bytes()
	sBytes := sCanon.Bytes()

	// Pad if necessary
	signature := make([]byte, 2*byteSize)
	copy(signature[byteSize-len(rBytes):byteSize], rBytes)
	copy(signature[2*byteSize-len(sBytes):], sBytes)

	return signature, nil
}

// SignatureFromBytes reconstructs signature from bytes and ensures it's canonical
func SignatureFromBytes(sig []byte, curve elliptic.Curve) (*big.Int, *big.Int, error) {
	byteSize := (curve.Params().BitSize + 7) / 8

	if len(sig) != 2*byteSize {
		return nil, nil, fmt.Errorf("invalid signature length: expected %d, got %d", 2*byteSize, len(sig))
	}

	r := new(big.Int).SetBytes(sig[:byteSize])
	s := new(big.Int).SetBytes(sig[byteSize:])

	// Ensure canonical form
	return CanonicalizeSignature(r, s, curve)
}
