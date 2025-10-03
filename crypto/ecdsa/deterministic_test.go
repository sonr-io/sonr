package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeterministicSign(t *testing.T) {
	// Generate test key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Test message
	message := []byte("test message for deterministic signing")
	hash := sha256.Sum256(message)

	// Sign with deterministic algorithm
	r1, s1, err := DeterministicSign(priv, hash[:])
	require.NoError(t, err)
	assert.NotNil(t, r1)
	assert.NotNil(t, s1)

	// Sign again - should produce identical signature
	r2, s2, err := DeterministicSign(priv, hash[:])
	require.NoError(t, err)
	assert.Equal(t, r1, r2, "deterministic signatures should be identical")
	assert.Equal(t, s1, s2, "deterministic signatures should be identical")

	// Verify signature
	valid := ecdsa.Verify(&priv.PublicKey, hash[:], r1, s1)
	assert.True(t, valid, "signature should be valid")

	// Different message should produce different signature
	message2 := []byte("different message")
	hash2 := sha256.Sum256(message2)
	r3, s3, err := DeterministicSign(priv, hash2[:])
	require.NoError(t, err)
	assert.NotEqual(t, r1, r3, "different messages should produce different signatures")
	// Also verify the s component is different
	assert.NotEqual(t, s1, s3, "different messages should produce different s values")
}

func TestCanonicalSignature(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	message := []byte("test canonical signature")
	hash := sha256.Sum256(message)

	// Sign multiple times and check all signatures are canonical
	for i := 0; i < 10; i++ {
		r, s, err := DeterministicSign(priv, hash[:])
		require.NoError(t, err)

		// Check signature is canonical (s <= N/2)
		N := priv.Curve.Params().N
		assert.True(t, IsCanonical(s, N), "signature should be canonical")

		// Verify signature
		valid := ecdsa.Verify(&priv.PublicKey, hash[:], r, s)
		assert.True(t, valid, "canonical signature should be valid")
	}
}

func TestMakeCanonical(t *testing.T) {
	curve := elliptic.P256()
	N := curve.Params().N
	halfN := new(big.Int).Div(N, big.NewInt(2))

	// Test with non-canonical s (s > N/2)
	r := big.NewInt(12345)
	s := new(big.Int).Add(halfN, big.NewInt(1)) // s = N/2 + 1

	assert.False(t, IsCanonical(s, N), "s > N/2 should not be canonical")

	// Make canonical
	rCanon, sCanon := MakeCanonical(r, s, N)

	assert.Equal(t, r, rCanon, "r should not change")
	assert.True(t, IsCanonical(sCanon, N), "canonicalized s should be <= N/2")

	// sCanon should equal N - s
	expected := new(big.Int).Sub(N, s)
	assert.Equal(t, expected, sCanon, "canonical s should be N - s")
}

func TestVerifyDeterministic(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	message := []byte("test verification")
	hash := sha256.Sum256(message)

	// Create deterministic signature
	r, s, err := DeterministicSign(priv, hash[:])
	require.NoError(t, err)

	// Verify with our function
	valid := VerifyDeterministic(&priv.PublicKey, hash[:], r, s)
	assert.True(t, valid, "signature should verify")

	// Test with non-canonical signature (should fail)
	N := priv.Curve.Params().N
	sNonCanon := new(big.Int).Sub(N, s) // Create non-canonical s

	valid = VerifyDeterministic(&priv.PublicKey, hash[:], r, sNonCanon)
	assert.False(t, valid, "non-canonical signature should not verify")

	// Test with wrong hash
	wrongHash := sha256.Sum256([]byte("wrong message"))
	valid = VerifyDeterministic(&priv.PublicKey, wrongHash[:], r, s)
	assert.False(t, valid, "signature with wrong hash should not verify")
}

func TestInvalidInputs(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	hash := sha256.Sum256([]byte("test"))

	// Test with nil private key
	r, s, err := DeterministicSign(nil, hash[:])
	assert.Error(t, err)
	assert.Nil(t, r)
	assert.Nil(t, s)

	// Test with empty hash
	r, s, err = DeterministicSign(priv, []byte{})
	assert.Error(t, err)
	assert.Nil(t, r)
	assert.Nil(t, s)

	// Test verify with nil inputs
	assert.False(t, VerifyDeterministic(nil, hash[:], big.NewInt(1), big.NewInt(1)))
	assert.False(t, VerifyDeterministic(&priv.PublicKey, hash[:], nil, big.NewInt(1)))
	assert.False(t, VerifyDeterministic(&priv.PublicKey, hash[:], big.NewInt(1), nil))
}

func TestDifferentCurves(t *testing.T) {
	curves := []elliptic.Curve{
		elliptic.P224(),
		elliptic.P256(),
		elliptic.P384(),
		elliptic.P521(),
	}

	message := []byte("test message for different curves")
	hash := sha256.Sum256(message)

	for _, curve := range curves {
		t.Run(curve.Params().Name, func(t *testing.T) {
			priv, err := ecdsa.GenerateKey(curve, rand.Reader)
			require.NoError(t, err)

			// Sign deterministically
			r, s, err := DeterministicSign(priv, hash[:])
			require.NoError(t, err)

			// Verify signature is canonical
			N := curve.Params().N
			assert.True(t, IsCanonical(s, N))

			// Verify signature
			valid := ecdsa.Verify(&priv.PublicKey, hash[:], r, s)
			assert.True(t, valid)

			// Verify deterministic property
			r2, s2, err := DeterministicSign(priv, hash[:])
			require.NoError(t, err)
			assert.Equal(t, r, r2)
			assert.Equal(t, s, s2)
		})
	}
}

// TestRFC6979Vectors tests against known test vectors
// These are simplified vectors - in production, use the full RFC 6979 test vectors
func TestRFC6979Vectors(t *testing.T) {
	// Test vector for P-256 with SHA-256
	// This is a simplified example - real implementation should use official test vectors
	privKeyHex := "c9afa9d845ba75166b5c215767b1d6934e50c3db36e89b127b8a622b120f6721"
	messageHex := "73616d706c65" // "sample"

	privKeyBytes, err := hex.DecodeString(privKeyHex)
	require.NoError(t, err)

	message, err := hex.DecodeString(messageHex)
	require.NoError(t, err)

	// Create private key
	priv := new(ecdsa.PrivateKey)
	priv.Curve = elliptic.P256()
	priv.D = new(big.Int).SetBytes(privKeyBytes)
	priv.PublicKey.Curve = priv.Curve
	priv.PublicKey.X, priv.PublicKey.Y = priv.Curve.ScalarBaseMult(privKeyBytes)

	// Hash message
	hash := sha256.Sum256(message)

	// Sign deterministically
	r, s, err := DeterministicSign(priv, hash[:])
	require.NoError(t, err)
	assert.NotNil(t, r)
	assert.NotNil(t, s)

	// Verify signature
	valid := ecdsa.Verify(&priv.PublicKey, hash[:], r, s)
	assert.True(t, valid, "RFC 6979 test vector signature should verify")
}

func BenchmarkDeterministicSign(b *testing.B) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(b, err)

	message := []byte("benchmark message")
	hash := sha256.Sum256(message)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = DeterministicSign(priv, hash[:])
	}
}

func BenchmarkVerifyDeterministic(b *testing.B) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(b, err)

	message := []byte("benchmark message")
	hash := sha256.Sum256(message)

	r, s, err := DeterministicSign(priv, hash[:])
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifyDeterministic(&priv.PublicKey, hash[:], r, s)
	}
}

func TestConcurrentSigning(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	message := []byte("concurrent test")
	hash := sha256.Sum256(message)

	// Sign concurrently
	const goroutines = 10
	results := make(chan struct {
		r, s *big.Int
		err  error
	}, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			r, s, err := DeterministicSign(priv, hash[:])
			results <- struct {
				r, s *big.Int
				err  error
			}{r, s, err}
		}()
	}

	// Collect results
	var firstR, firstS *big.Int
	for i := 0; i < goroutines; i++ {
		result := <-results
		require.NoError(t, result.err)

		if i == 0 {
			firstR, firstS = result.r, result.s
		} else {
			// All signatures should be identical (deterministic)
			assert.Equal(t, firstR, result.r)
			assert.Equal(t, firstS, result.s)
		}
	}
}
