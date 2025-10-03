package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanonicalizeSignature(t *testing.T) {
	curve := elliptic.P256()
	N := curve.Params().N
	halfN := new(big.Int).Div(N, big.NewInt(2))

	// Test canonical signature (s <= N/2)
	r := big.NewInt(12345)
	s := new(big.Int).Sub(halfN, big.NewInt(1)) // s = N/2 - 1

	rCanon, sCanon, err := CanonicalizeSignature(r, s, curve)
	require.NoError(t, err)
	assert.Equal(t, r, rCanon)
	assert.Equal(t, s, sCanon)

	// Test non-canonical signature (s > N/2)
	sNonCanon := new(big.Int).Add(halfN, big.NewInt(1)) // s = N/2 + 1

	rCanon, sCanon, err = CanonicalizeSignature(r, sNonCanon, curve)
	require.NoError(t, err)
	assert.Equal(t, r, rCanon)

	// sCanon should be N - sNonCanon
	expected := new(big.Int).Sub(N, sNonCanon)
	assert.Equal(t, expected, sCanon)
}

func TestIsSignatureCanonical(t *testing.T) {
	curve := elliptic.P256()
	N := curve.Params().N
	halfN := new(big.Int).Div(N, big.NewInt(2))

	// Test canonical signature
	r := big.NewInt(12345)
	s := halfN // s = N/2 (boundary case, still canonical)

	assert.True(t, IsSignatureCanonical(r, s, curve))

	// Test non-canonical signature
	sNonCanon := new(big.Int).Add(halfN, big.NewInt(1))
	assert.False(t, IsSignatureCanonical(r, sNonCanon, curve))

	// Test invalid r (r = 0)
	assert.False(t, IsSignatureCanonical(big.NewInt(0), s, curve))

	// Test invalid r (r >= N)
	assert.False(t, IsSignatureCanonical(N, s, curve))

	// Test invalid s (s = 0)
	assert.False(t, IsSignatureCanonical(r, big.NewInt(0), curve))

	// Test nil inputs
	assert.False(t, IsSignatureCanonical(nil, s, curve))
	assert.False(t, IsSignatureCanonical(r, nil, curve))
	assert.False(t, IsSignatureCanonical(r, s, nil))
}

func TestValidateAndCanonicalizeSignature(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	message := []byte("test message")
	hash := sha256.Sum256(message)

	// Sign with standard ECDSA
	r, s, err := ecdsa.Sign(rand.Reader, priv, hash[:])
	require.NoError(t, err)

	// Validate and canonicalize
	rCanon, sCanon, err := ValidateAndCanonicalizeSignature(&priv.PublicKey, hash[:], r, s)
	require.NoError(t, err)

	// Verify canonical signature
	valid := ecdsa.Verify(&priv.PublicKey, hash[:], rCanon, sCanon)
	assert.True(t, valid)

	// Ensure signature is canonical
	assert.True(t, IsSignatureCanonical(rCanon, sCanon, priv.Curve))

	// Test with invalid signature
	wrongR := new(big.Int).Add(r, big.NewInt(1))
	_, _, err = ValidateAndCanonicalizeSignature(&priv.PublicKey, hash[:], wrongR, s)
	assert.Error(t, err)
}

func TestRejectNonCanonical(t *testing.T) {
	curve := elliptic.P256()
	N := curve.Params().N
	halfN := new(big.Int).Div(N, big.NewInt(2))

	r := big.NewInt(12345)

	// Test canonical signature
	sCanon := new(big.Int).Sub(halfN, big.NewInt(1))
	err := RejectNonCanonical(r, sCanon, curve)
	assert.NoError(t, err)

	// Test non-canonical signature
	sNonCanon := new(big.Int).Add(halfN, big.NewInt(1))
	err = RejectNonCanonical(r, sNonCanon, curve)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in canonical form")

	// Test nil inputs
	err = RejectNonCanonical(nil, sCanon, curve)
	assert.Error(t, err)

	err = RejectNonCanonical(r, nil, curve)
	assert.Error(t, err)

	err = RejectNonCanonical(r, sCanon, nil)
	assert.Error(t, err)
}

func TestCompareSignatures(t *testing.T) {
	curve := elliptic.P256()
	N := curve.Params().N
	halfN := new(big.Int).Div(N, big.NewInt(2))

	r := big.NewInt(12345)
	s1 := new(big.Int).Sub(halfN, big.NewInt(1))

	// Same signature should be equal
	equal, err := CompareSignatures(r, s1, r, s1, curve)
	require.NoError(t, err)
	assert.True(t, equal)

	// Canonical and non-canonical versions of same signature should be equal
	s1NonCanon := new(big.Int).Sub(N, s1)
	equal, err = CompareSignatures(r, s1, r, s1NonCanon, curve)
	require.NoError(t, err)
	assert.True(t, equal)

	// Different signatures should not be equal
	s2 := new(big.Int).Sub(halfN, big.NewInt(10))
	equal, err = CompareSignatures(r, s1, r, s2, curve)
	require.NoError(t, err)
	assert.False(t, equal)
}

func TestSignatureBytes(t *testing.T) {
	curve := elliptic.P256()
	N := curve.Params().N
	halfN := new(big.Int).Div(N, big.NewInt(2))

	r := big.NewInt(12345)
	s := new(big.Int).Sub(halfN, big.NewInt(1))

	// Convert to bytes
	sigBytes, err := SignatureBytes(r, s, curve)
	require.NoError(t, err)
	assert.Len(t, sigBytes, 64) // 32 bytes for r, 32 bytes for s on P-256

	// Test with non-canonical s - should be canonicalized
	sNonCanon := new(big.Int).Sub(N, s)
	sigBytesNonCanon, err := SignatureBytes(r, sNonCanon, curve)
	require.NoError(t, err)

	// Both should produce the same bytes (after canonicalization)
	assert.Equal(t, sigBytes, sigBytesNonCanon)
}

func TestSignatureFromBytes(t *testing.T) {
	curve := elliptic.P256()

	// Create a signature
	r := big.NewInt(12345)
	s := big.NewInt(67890)

	// Convert to bytes
	sigBytes, err := SignatureBytes(r, s, curve)
	require.NoError(t, err)

	// Convert back from bytes
	rRecovered, sRecovered, err := SignatureFromBytes(sigBytes, curve)
	require.NoError(t, err)

	// Should be canonical
	assert.True(t, IsSignatureCanonical(rRecovered, sRecovered, curve))

	// Values should match (after canonicalization)
	rCanon, sCanon, err := CanonicalizeSignature(r, s, curve)
	require.NoError(t, err)
	assert.Equal(t, rCanon, rRecovered)
	assert.Equal(t, sCanon, sRecovered)

	// Test with invalid length
	_, _, err = SignatureFromBytes([]byte("too short"), curve)
	assert.Error(t, err)
}

func TestNormalizeSignature(t *testing.T) {
	curve := elliptic.P256()
	N := curve.Params().N
	halfN := new(big.Int).Div(N, big.NewInt(2))

	r := big.NewInt(12345)
	s := new(big.Int).Add(halfN, big.NewInt(1)) // Non-canonical

	// Normalize should canonicalize
	rNorm, sNorm, err := NormalizeSignature(r, s, curve)
	require.NoError(t, err)

	assert.True(t, IsSignatureCanonical(rNorm, sNorm, curve))
	assert.Equal(t, r, rNorm)

	// sNorm should be N - s
	expected := new(big.Int).Sub(N, s)
	assert.Equal(t, expected, sNorm)
}

func TestCanonicalWithRealSignatures(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	message := []byte("test canonical signatures")
	hash := sha256.Sum256(message)

	// Generate multiple signatures and ensure all can be canonicalized
	for i := 0; i < 10; i++ {
		r, s, err := ecdsa.Sign(rand.Reader, priv, hash[:])
		require.NoError(t, err)

		// Canonicalize
		rCanon, sCanon, err := CanonicalizeSignature(r, s, priv.Curve)
		require.NoError(t, err)

		// Should be canonical
		assert.True(t, IsSignatureCanonical(rCanon, sCanon, priv.Curve))

		// Should still verify
		valid := ecdsa.Verify(&priv.PublicKey, hash[:], rCanon, sCanon)
		assert.True(t, valid)
	}
}

func BenchmarkCanonicalizeSignature(b *testing.B) {
	curve := elliptic.P256()
	N := curve.Params().N
	halfN := new(big.Int).Div(N, big.NewInt(2))

	r := big.NewInt(12345)
	s := new(big.Int).Add(halfN, big.NewInt(1))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = CanonicalizeSignature(r, s, curve)
	}
}

func BenchmarkIsSignatureCanonical(b *testing.B) {
	curve := elliptic.P256()
	N := curve.Params().N
	halfN := new(big.Int).Div(N, big.NewInt(2))

	r := big.NewInt(12345)
	s := new(big.Int).Sub(halfN, big.NewInt(1))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsSignatureCanonical(r, s, curve)
	}
}
