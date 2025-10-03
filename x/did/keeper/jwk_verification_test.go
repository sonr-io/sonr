// Package keeper provides integration tests for JWK verification
package keeper

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"
)

// TestECJWKVerification tests EC JWK verification with multiple curves
func TestECJWKVerification(t *testing.T) {
	tests := []struct {
		name  string
		curve elliptic.Curve
		crv   string
	}{
		{"P-256", elliptic.P256(), "P-256"},
		{"P-384", elliptic.P384(), "P-384"},
		{"P-521", elliptic.P521(), "P-521"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate EC key pair
			priv, err := ecdsa.GenerateKey(tt.curve, rand.Reader)
			require.NoError(t, err)

			// Create JWK
			jwk := map[string]any{
				"kty": "EC",
				"crv": tt.crv,
				"x":   base64.RawURLEncoding.EncodeToString(priv.X.Bytes()),
				"y":   base64.RawURLEncoding.EncodeToString(priv.Y.Bytes()),
			}

			// Create test message and signature
			message := []byte("test message")
			var hash []byte
			switch tt.crv {
			case "P-256":
				h := sha256.Sum256(message)
				hash = h[:]
			case "P-384":
				h := sha3.Sum384(message)
				hash = h[:]
			case "P-521":
				h := sha512.Sum512(message)
				hash = h[:]
			}

			sig, err := ecdsa.SignASN1(rand.Reader, priv, hash)
			require.NoError(t, err)

			// Test verification
			k := Keeper{}
			valid, err := k.verifyWithJWKEC(jwk, sig)
			require.NoError(t, err)
			require.True(t, valid, "EC signature verification failed for %s", tt.name)
		})
	}
}

// TestRSAJWKVerification tests RSA JWK verification with different key sizes
func TestRSAJWKVerification(t *testing.T) {
	tests := []struct {
		name    string
		keySize int
		alg     string
	}{
		{"RS256-2048", 2048, "RS256"},
		{"RS384-3072", 3072, "RS384"},
		{"RS512-4096", 4096, "RS512"},
		{"PS256-2048", 2048, "PS256"},
		{"PS384-3072", 3072, "PS384"},
		{"PS512-4096", 4096, "PS512"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate RSA key pair
			priv, err := rsa.GenerateKey(rand.Reader, tt.keySize)
			require.NoError(t, err)

			// Create JWK
			jwk := map[string]any{
				"kty": "RSA",
				"alg": tt.alg,
				"n":   base64.RawURLEncoding.EncodeToString(priv.N.Bytes()),
				"e": base64.RawURLEncoding.EncodeToString(
					big.NewInt(int64(priv.PublicKey.E)).Bytes(),
				),
			}

			// Create test message and signature
			message := []byte("test message")
			var hash []byte
			var hashFunc crypto.Hash

			switch tt.alg {
			case "RS256", "PS256":
				h := sha256.Sum256(message)
				hash = h[:]
				hashFunc = crypto.SHA256
			case "RS384", "PS384":
				h := sha3.Sum384(message)
				hash = h[:]
				hashFunc = crypto.SHA384
			case "RS512", "PS512":
				h := sha512.Sum512(message)
				hash = h[:]
				hashFunc = crypto.SHA512
			}

			var sig []byte
			if tt.alg[:2] == "PS" {
				// PSS signature
				opts := &rsa.PSSOptions{
					SaltLength: rsa.PSSSaltLengthEqualsHash,
					Hash:       hashFunc,
				}
				sig, err = rsa.SignPSS(rand.Reader, priv, hashFunc, hash, opts)
			} else {
				// PKCS#1 v1.5 signature
				sig, err = rsa.SignPKCS1v15(rand.Reader, priv, hashFunc, hash)
			}
			require.NoError(t, err)

			// Test verification
			k := Keeper{}
			valid, err := k.verifyWithJWKRSA(jwk, sig)
			require.NoError(t, err)
			require.True(t, valid, "RSA signature verification failed for %s", tt.name)
		})
	}
}

// TestOKPJWKVerification tests Ed25519 JWK verification
func TestOKPJWKVerification(t *testing.T) {
	// Generate Ed25519 key pair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Create JWK
	jwk := map[string]any{
		"kty": "OKP",
		"crv": "Ed25519",
		"x":   base64.RawURLEncoding.EncodeToString(pub),
	}

	// Create test message and signature
	message := []byte("test message")
	sig := ed25519.Sign(priv, message)

	// Test verification
	k := Keeper{}
	valid, err := k.verifyWithJWKOKP(jwk, sig)
	require.NoError(t, err)
	require.True(t, valid, "Ed25519 signature verification failed")
}

// TestMultiAlgorithmDetection tests the main JWK verification router
func TestMultiAlgorithmDetection(t *testing.T) {
	tests := []struct {
		name string
		jwk  map[string]any
		err  bool
	}{
		{
			name: "EC key",
			jwk: map[string]any{
				"kty": "EC",
				"crv": "P-256",
				"x":   base64.RawURLEncoding.EncodeToString(make([]byte, 32)),
				"y":   base64.RawURLEncoding.EncodeToString(make([]byte, 32)),
			},
			err: false,
		},
		{
			name: "RSA key",
			jwk: map[string]any{
				"kty": "RSA",
				"n":   base64.RawURLEncoding.EncodeToString(make([]byte, 256)),
				"e":   base64.RawURLEncoding.EncodeToString([]byte{1, 0, 1}),
			},
			err: false,
		},
		{
			name: "OKP key",
			jwk: map[string]any{
				"kty": "OKP",
				"crv": "Ed25519",
				"x":   base64.RawURLEncoding.EncodeToString(make([]byte, 32)),
			},
			err: false,
		},
		{
			name: "Unsupported key type",
			jwk: map[string]any{
				"kty": "INVALID",
			},
			err: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwkStr, err := json.Marshal(tt.jwk)
			require.NoError(t, err)

			k := Keeper{}
			_, err = k.verifyWithJWK(string(jwkStr), []byte("dummy signature"))

			if tt.err {
				require.Error(t, err, "Expected error for %s", tt.name)
			} else {
				// Note: Will fail signature verification but should parse correctly
				if err != nil {
					require.NotContains(t, err.Error(), "unsupported JWK key type")
				}
			}
		})
	}
}

// TestInvalidJWKHandling tests error handling for invalid JWKs
func TestInvalidJWKHandling(t *testing.T) {
	tests := []struct {
		name string
		jwk  map[string]any
		err  string
	}{
		{
			name: "Missing curve in EC JWK",
			jwk: map[string]any{
				"kty": "EC",
				"x":   "test",
				"y":   "test",
			},
			err: "missing or invalid 'crv' parameter",
		},
		{
			name: "Missing x coordinate in EC JWK",
			jwk: map[string]any{
				"kty": "EC",
				"crv": "P-256",
				"y":   "test",
			},
			err: "missing or invalid 'x' coordinate",
		},
		{
			name: "Missing modulus in RSA JWK",
			jwk: map[string]any{
				"kty": "RSA",
				"e":   "AQAB",
			},
			err: "missing or invalid 'n' (modulus)",
		},
		{
			name: "Small RSA key",
			jwk: map[string]any{
				"kty": "RSA",
				"n":   base64.RawURLEncoding.EncodeToString(make([]byte, 128)), // 1024 bits
				"e":   "AQAB",
			},
			err: "RSA key size too small",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := Keeper{}

			switch tt.jwk["kty"] {
			case "EC":
				_, err := k.verifyWithJWKEC(tt.jwk, []byte{})
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
			case "RSA":
				_, err := k.verifyWithJWKRSA(tt.jwk, []byte{})
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
			}
		})
	}
}

// BenchmarkECJWKVerification benchmarks EC JWK verification
func BenchmarkECJWKVerification(b *testing.B) {
	curves := []struct {
		name  string
		curve elliptic.Curve
		crv   string
	}{
		{"P256", elliptic.P256(), "P-256"},
		{"P384", elliptic.P384(), "P-384"},
		{"P521", elliptic.P521(), "P-521"},
	}

	for _, c := range curves {
		b.Run(c.name, func(b *testing.B) {
			// Setup
			priv, _ := ecdsa.GenerateKey(c.curve, rand.Reader)
			jwk := map[string]any{
				"kty": "EC",
				"crv": c.crv,
				"x":   base64.RawURLEncoding.EncodeToString(priv.X.Bytes()),
				"y":   base64.RawURLEncoding.EncodeToString(priv.Y.Bytes()),
			}

			message := []byte("test message")
			h := sha256.Sum256(message)
			sig, _ := ecdsa.SignASN1(rand.Reader, priv, h[:])

			k := Keeper{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = k.verifyWithJWKEC(jwk, sig)
			}
		})
	}
}

// BenchmarkRSAJWKVerification benchmarks RSA JWK verification
func BenchmarkRSAJWKVerification(b *testing.B) {
	keySizes := []int{2048, 3072, 4096}

	for _, size := range keySizes {
		b.Run(fmt.Sprintf("RSA%d", size), func(b *testing.B) {
			// Setup
			priv, _ := rsa.GenerateKey(rand.Reader, size)
			jwk := map[string]any{
				"kty": "RSA",
				"n":   base64.RawURLEncoding.EncodeToString(priv.N.Bytes()),
				"e": base64.RawURLEncoding.EncodeToString(
					big.NewInt(int64(priv.PublicKey.E)).Bytes(),
				),
			}

			message := []byte("test message")
			h := sha256.Sum256(message)
			sig, _ := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, h[:])

			k := Keeper{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = k.verifyWithJWKRSA(jwk, sig)
			}
		})
	}
}
