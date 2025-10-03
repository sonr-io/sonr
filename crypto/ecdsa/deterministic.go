// Package ecdsa provides RFC 6979 deterministic ECDSA implementation
package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"hash"
	"math/big"
)

// DeterministicSign implements RFC 6979 deterministic ECDSA signing
// This prevents nonce reuse and bias attacks by generating k deterministically
func DeterministicSign(priv *ecdsa.PrivateKey, hash []byte) (*big.Int, *big.Int, error) {
	if priv == nil || priv.D == nil {
		return nil, nil, fmt.Errorf("invalid private key")
	}

	if len(hash) == 0 {
		return nil, nil, fmt.Errorf("hash cannot be empty")
	}

	// Generate deterministic k using RFC 6979
	k := generateK(priv, hash, sha256.New)

	// Sign with deterministic k
	return signWithK(priv, hash, k)
}

// generateK implements RFC 6979 deterministic nonce generation
func generateK(priv *ecdsa.PrivateKey, hash []byte, hashFunc func() hash.Hash) *big.Int {
	curve := priv.Curve
	N := curve.Params().N
	bitSize := N.BitLen()
	byteSize := (bitSize + 7) / 8

	// Step a: Process hash
	h1 := hashToInt(hash, curve)

	// Step b: Convert private key to bytes
	x := priv.D.Bytes()
	if len(x) < byteSize {
		// Pad with zeros on the left
		padding := make([]byte, byteSize-len(x))
		x = append(padding, x...)
	}

	// Step c: Create HMAC-DRBG instance
	hm := hmac.New(hashFunc, nil)
	hlen := hm.Size()

	// Step d: Set V = 0x01 0x01 0x01 ... 0x01
	v := bytes(hlen, 0x01)

	// Step e: Set K = 0x00 0x00 0x00 ... 0x00
	k := bytes(hlen, 0x00)

	// Step f: K = HMAC_K(V || 0x00 || x || h1)
	k = hmacCompute(hashFunc, k, v, []byte{0x00}, x, h1.Bytes())

	// Step g: V = HMAC_K(V)
	v = hmacCompute(hashFunc, k, v)

	// Step h: K = HMAC_K(V || 0x01 || x || h1)
	k = hmacCompute(hashFunc, k, v, []byte{0x01}, x, h1.Bytes())

	// Step i: V = HMAC_K(V)
	v = hmacCompute(hashFunc, k, v)

	// Step j: Generate k
	for {
		// Step j.1: Set T = empty sequence
		var t []byte

		// Step j.2: While tlen < qlen
		for len(t)*8 < bitSize {
			// V = HMAC_K(V)
			v = hmacCompute(hashFunc, k, v)
			// T = T || V
			t = append(t, v...)
		}

		// Step j.3: k = bits2int(T)
		kInt := hashToInt(t, curve)

		// Check if k is valid (0 < k < N)
		if kInt.Sign() > 0 && kInt.Cmp(N) < 0 {
			return kInt
		}

		// Step j.4: K = HMAC_K(V || 0x00)
		k = hmacCompute(hashFunc, k, v, []byte{0x00})
		// V = HMAC_K(V)
		v = hmacCompute(hashFunc, k, v)
	}
}

// signWithK performs ECDSA signing with a given k value
func signWithK(priv *ecdsa.PrivateKey, hash []byte, k *big.Int) (*big.Int, *big.Int, error) {
	curve := priv.Curve
	N := curve.Params().N

	// Calculate r = x-coordinate of k*G mod N
	x, _ := curve.ScalarBaseMult(k.Bytes())
	r := new(big.Int).Set(x)
	r.Mod(r, N)

	if r.Sign() == 0 {
		return nil, nil, fmt.Errorf("invalid r value")
	}

	// Calculate s = k^(-1) * (h + r*d) mod N
	e := hashToInt(hash, curve)

	kInv := new(big.Int).ModInverse(k, N)
	if kInv == nil {
		return nil, nil, fmt.Errorf("k has no inverse")
	}

	s := new(big.Int).Mul(r, priv.D)
	s.Add(s, e)
	s.Mul(s, kInv)
	s.Mod(s, N)

	if s.Sign() == 0 {
		return nil, nil, fmt.Errorf("invalid s value")
	}

	// Canonicalize signature (ensure s <= N/2)
	rFinal, sFinal := canonicalize(r, s, N)
	return rFinal, sFinal, nil
}

// canonicalize ensures the signature is in canonical form (s <= N/2)
// This prevents signature malleability
func canonicalize(r, s, N *big.Int) (*big.Int, *big.Int) {
	halfN := new(big.Int).Div(N, big.NewInt(2))

	// If s > N/2, use N - s instead
	if s.Cmp(halfN) > 0 {
		s = new(big.Int).Sub(N, s)
	}

	return r, s
}

// hashToInt converts a hash value to an integer for ECDSA operations
func hashToInt(hash []byte, curve elliptic.Curve) *big.Int {
	N := curve.Params().N
	orderBits := N.BitLen()
	orderBytes := (orderBits + 7) / 8

	if len(hash) > orderBytes {
		hash = hash[:orderBytes]
	}

	ret := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - orderBits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}

	return ret
}

// hmacCompute computes HMAC with concatenated data
func hmacCompute(hashFunc func() hash.Hash, key []byte, data ...[]byte) []byte {
	mac := hmac.New(hashFunc, key)
	for _, d := range data {
		mac.Write(d)
	}
	return mac.Sum(nil)
}

// bytes creates a byte slice filled with value
func bytes(size int, value byte) []byte {
	b := make([]byte, size)
	for i := range b {
		b[i] = value
	}
	return b
}

// VerifyDeterministic verifies a deterministic ECDSA signature
func VerifyDeterministic(pub *ecdsa.PublicKey, hash []byte, r, s *big.Int) bool {
	if pub == nil || r == nil || s == nil {
		return false
	}

	// Ensure signature is canonical
	N := pub.Curve.Params().N
	halfN := new(big.Int).Div(N, big.NewInt(2))

	// Check r and s are in valid range
	if r.Sign() <= 0 || r.Cmp(N) >= 0 {
		return false
	}
	if s.Sign() <= 0 || s.Cmp(halfN) > 0 {
		return false // s must be <= N/2 for canonical form
	}

	return ecdsa.Verify(pub, hash, r, s)
}

// IsCanonical checks if a signature is in canonical form
func IsCanonical(s, N *big.Int) bool {
	if s == nil || N == nil {
		return false
	}

	halfN := new(big.Int).Div(N, big.NewInt(2))
	return s.Cmp(halfN) <= 0
}

// MakeCanonical converts a signature to canonical form
func MakeCanonical(r, s, N *big.Int) (*big.Int, *big.Int) {
	if r == nil || s == nil || N == nil {
		return r, s
	}

	return canonicalize(r, s, N)
}
