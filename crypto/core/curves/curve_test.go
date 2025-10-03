// Package curves provides integration tests for elliptic curve support
package curves

import (
	"crypto/elliptic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/crypto/core"
)

// TestAdditionalCurveSupport tests the new curve support in hash.go
func TestAdditionalCurveSupport(t *testing.T) {
	tests := []struct {
		name             string
		curve            elliptic.Curve
		expectedSecurity int
		expectedL        int
	}{
		{
			name:             "P-256",
			curve:            elliptic.P256(),
			expectedSecurity: 128,
			expectedL:        48,
		},
		{
			name:             "P-384",
			curve:            elliptic.P384(),
			expectedSecurity: 192,
			expectedL:        72,
		},
		{
			name:             "P-521",
			curve:            elliptic.P521(),
			expectedSecurity: 256,
			expectedL:        98,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Hash function with the curve
			msg := []byte("test message for curve support")
			result, err := core.Hash(msg, tt.curve)

			require.NoError(t, err, "Hash should succeed for %s", tt.name)
			require.NotNil(t, result, "Hash result should not be nil for %s", tt.name)

			// Verify the result is a valid big integer
			require.Greater(t, result.BitLen(), 0, "Hash result should have non-zero bit length")
		})
	}
}

// BenchmarkHashWithCurves benchmarks the hash function with different curves
func BenchmarkHashWithCurves(b *testing.B) {
	curves := []struct {
		name  string
		curve elliptic.Curve
	}{
		{"P256", elliptic.P256()},
		{"P384", elliptic.P384()},
		{"P521", elliptic.P521()},
	}

	msg := []byte("benchmark message for hash function")

	for _, c := range curves {
		b.Run(c.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = core.Hash(msg, c.curve)
			}
		})
	}
}

// TestBackwardCompatibility ensures old curves still work
func TestBackwardCompatibility(t *testing.T) {
	// Test that P-256 still works as before
	msg := []byte("backward compatibility test")

	result256, err := core.Hash(msg, elliptic.P256())
	require.NoError(t, err)
	require.NotNil(t, result256)

	// Test with same message multiple times for consistency
	for i := 0; i < 10; i++ {
		result, err := core.Hash(msg, elliptic.P256())
		require.NoError(t, err)
		require.Equal(t, result256.Cmp(result), 0, "Hash should be consistent")
	}
}
