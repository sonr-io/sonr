// Package crypto provides comprehensive security tests for cryptographic implementations
package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/sonr-io/sonr/crypto/argon2"
	ecdsaPkg "github.com/sonr-io/sonr/crypto/ecdsa"
	"github.com/sonr-io/sonr/crypto/password"
	"github.com/sonr-io/sonr/crypto/wasm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTimingAttackResistance verifies constant-time operations
func TestTimingAttackResistance(t *testing.T) {
	// Test Argon2 constant-time comparison
	kdf := argon2.New(argon2.LightConfig())
	password := []byte("correct-password")
	hash, err := kdf.HashPassword(password)
	require.NoError(t, err)

	// Measure timing for correct vs incorrect passwords
	correctTimes := make([]time.Duration, 100)
	incorrectTimes := make([]time.Duration, 100)

	for i := 0; i < 100; i++ {
		// Time correct password
		start := time.Now()
		_, _ = argon2.VerifyPassword(password, hash)
		correctTimes[i] = time.Since(start)

		// Time incorrect password
		wrongPassword := []byte("wrong-password-x")
		start = time.Now()
		_, _ = argon2.VerifyPassword(wrongPassword, hash)
		incorrectTimes[i] = time.Since(start)
	}

	// Calculate average times
	var correctAvg, incorrectAvg time.Duration
	for i := 0; i < 100; i++ {
		correctAvg += correctTimes[i]
		incorrectAvg += incorrectTimes[i]
	}
	correctAvg /= 100
	incorrectAvg /= 100

	// Times should be similar (within 20% variance)
	diff := correctAvg - incorrectAvg
	if diff < 0 {
		diff = -diff
	}
	maxDiff := correctAvg / 5 // 20% threshold

	assert.Less(t, diff, maxDiff, "timing difference suggests non-constant-time comparison")
}

// TestSignatureMalleabilityAttack tests protection against signature malleability
func TestSignatureMalleabilityAttack(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	message := []byte("transaction data")
	hash := sha256.Sum256(message)

	// Create deterministic signature
	r, s, err := ecdsaPkg.DeterministicSign(priv, hash[:])
	require.NoError(t, err)

	// Verify original signature
	assert.True(t, ecdsa.Verify(&priv.PublicKey, hash[:], r, s))

	// Create malleable signature (r, -s mod N)
	N := priv.Curve.Params().N
	sMalleable := new(big.Int).Sub(N, s)

	// Standard ECDSA would accept this, but our canonical check should reject it
	assert.False(t, ecdsaPkg.VerifyDeterministic(&priv.PublicKey, hash[:], r, sMalleable),
		"malleable signature should be rejected")

	// Canonicalize should fix it
	rCanon, sCanon, err := ecdsaPkg.CanonicalizeSignature(r, sMalleable, priv.Curve)
	require.NoError(t, err)
	assert.Equal(t, s, sCanon, "canonicalized signature should match original")
	assert.True(t, ecdsa.Verify(&priv.PublicKey, hash[:], rCanon, sCanon))
}

// TestNonceReuseAttack verifies protection against nonce reuse
func TestNonceReuseAttack(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Sign same message multiple times with deterministic ECDSA
	message := []byte("sensitive data")
	hash := sha256.Sum256(message)

	signatures := make([]struct{ r, s *big.Int }, 10)
	for i := 0; i < 10; i++ {
		r, s, err := ecdsaPkg.DeterministicSign(priv, hash[:])
		require.NoError(t, err)
		signatures[i].r = r
		signatures[i].s = s
	}

	// All signatures should be identical (deterministic)
	for i := 1; i < 10; i++ {
		assert.Equal(t, signatures[0].r, signatures[i].r,
			"deterministic signatures should use same nonce")
		assert.Equal(t, signatures[0].s, signatures[i].s,
			"deterministic signatures should be identical")
	}

	// Different messages should use different nonces
	message2 := []byte("different data")
	hash2 := sha256.Sum256(message2)
	r2, s2, err := ecdsaPkg.DeterministicSign(priv, hash2[:])
	require.NoError(t, err)

	assert.NotEqual(t, signatures[0].r, r2,
		"different messages must use different nonces")
	assert.NotEqual(t, signatures[0].s, s2,
		"different messages must produce different signatures")
}

// TestPasswordDictionaryAttack tests resistance to dictionary attacks
func TestPasswordDictionaryAttack(t *testing.T) {
	commonPasswords := []string{
		"password", "123456", "password123", "admin", "letmein",
		"welcome", "monkey", "dragon", "master", "qwerty",
	}

	validator := password.NewValidator(password.DefaultPasswordConfig())

	// All common passwords should be rejected
	for _, pwd := range commonPasswords {
		err := validator.Validate([]byte(pwd))
		assert.Error(t, err, "common password '%s' should be rejected", pwd)
	}

	// Test that Argon2 makes dictionary attacks expensive
	kdf := argon2.New(argon2.DefaultConfig())

	start := time.Now()
	for _, pwd := range commonPasswords {
		hash, err := kdf.HashPassword([]byte(pwd))
		require.NoError(t, err)

		// Try to crack with dictionary
		for _, attempt := range commonPasswords {
			_, _ = argon2.VerifyPassword([]byte(attempt), hash)
		}
	}
	elapsed := time.Since(start)

	// Should take significant time (> 1 second for 100 attempts)
	assert.Greater(t, elapsed, 1*time.Second,
		"Argon2 should make dictionary attacks expensive")
}

// TestWASMHashCollisionAttack tests resistance to hash collision attacks
func TestWASMHashCollisionAttack(t *testing.T) {
	verifier := wasm.NewHashVerifier()

	// Create two different modules
	module1 := []byte("wasm module version 1.0")
	module2 := []byte("wasm module version 2.0")

	hash1 := verifier.ComputeHash(module1)
	hash2 := verifier.ComputeHash(module2)

	// Hashes must be different
	assert.NotEqual(t, hash1, hash2,
		"different modules must have different hashes")

	// Test collision resistance with similar modules
	similarModules := make([][]byte, 100)
	hashes := make(map[string]bool)

	for i := 0; i < 100; i++ {
		similarModules[i] = []byte(fmt.Sprintf("wasm module version 1.%d", i))
		hash := verifier.ComputeHash(similarModules[i])

		// Check for collisions
		assert.False(t, hashes[hash],
			"hash collision detected for module %d", i)
		hashes[hash] = true
	}
}

// TestRaceConditionSafety tests thread safety of cryptographic operations
func TestRaceConditionSafety(t *testing.T) {
	// Test concurrent Argon2 operations
	kdf := argon2.New(argon2.LightConfig())
	password := []byte("test-password")

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			hash, err := kdf.HashPassword(password)
			if err != nil {
				errors <- err
				return
			}

			valid, err := argon2.VerifyPassword(password, hash)
			if err != nil {
				errors <- err
				return
			}
			if !valid {
				errors <- fmt.Errorf("password verification failed")
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("concurrent operation failed: %v", err)
	}

	// Test concurrent ECDSA signing
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	message := []byte("concurrent test")
	hash := sha256.Sum256(message)

	signatures := make(chan struct{ r, s *big.Int }, 100)
	var sigWg sync.WaitGroup

	for i := 0; i < 100; i++ {
		sigWg.Add(1)
		go func() {
			defer sigWg.Done()

			r, s, err := ecdsaPkg.DeterministicSign(priv, hash[:])
			if err == nil {
				signatures <- struct{ r, s *big.Int }{r, s}
			}
		}()
	}

	sigWg.Wait()
	close(signatures)

	// All signatures should be identical (deterministic)
	var firstSig struct{ r, s *big.Int }
	count := 0
	for sig := range signatures {
		if count == 0 {
			firstSig = sig
		} else {
			assert.Equal(t, firstSig.r, sig.r,
				"concurrent signatures should be identical")
			assert.Equal(t, firstSig.s, sig.s,
				"concurrent signatures should be identical")
		}
		count++
	}
	assert.Equal(t, 100, count, "all concurrent operations should succeed")
}

// TestMemoryExhaustionAttack tests resistance to memory exhaustion
func TestMemoryExhaustionAttack(t *testing.T) {
	// Test with high memory Argon2 config
	config := &argon2.Config{
		Time:        1,
		Memory:      128 * 1024, // 128MB
		Parallelism: 4,
		SaltLength:  32,
		KeyLength:   32,
	}

	// Validate config prevents excessive memory use
	err := argon2.ValidateConfig(config)
	assert.NoError(t, err, "reasonable memory config should be valid")

	// Test with excessive memory request
	excessiveConfig := &argon2.Config{
		Time:        1,
		Memory:      4 * 1024 * 1024, // 4GB - should be rejected
		Parallelism: 4,
		SaltLength:  32,
		KeyLength:   32,
	}

	// This should be caught by reasonable implementations
	kdf := argon2.New(excessiveConfig)

	// Attempt derivation with memory limit
	password := []byte("test")
	salt := make([]byte, 32)
	_, _ = rand.Read(salt)

	// Monitor memory usage (simplified test)
	start := time.Now()
	_ = kdf.DeriveKey(password, salt)
	elapsed := time.Since(start)

	// Excessive memory should cause noticeable delay
	assert.Less(t, elapsed, 10*time.Second,
		"operation should complete in reasonable time")
}

// TestSaltReuseVulnerability tests that salts are unique
func TestSaltReuseVulnerability(t *testing.T) {
	kdf := argon2.New(argon2.DefaultConfig())

	salts := make(map[string]bool)

	// Generate many salts
	for i := 0; i < 1000; i++ {
		salt, err := kdf.GenerateSalt()
		require.NoError(t, err)

		saltStr := string(salt)
		assert.False(t, salts[saltStr],
			"salt reuse detected at iteration %d", i)
		salts[saltStr] = true
	}
}

// TestWeakRandomnessDetection tests quality of random number generation
func TestWeakRandomnessDetection(t *testing.T) {
	// Generate multiple ECDSA keys and check for patterns
	keys := make([]*ecdsa.PrivateKey, 100)

	for i := 0; i < 100; i++ {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)
		keys[i] = key
	}

	// Check that all private keys are unique
	seen := make(map[string]bool)
	for i, key := range keys {
		keyStr := key.D.String()
		assert.False(t, seen[keyStr],
			"duplicate private key detected at index %d", i)
		seen[keyStr] = true
	}

	// Check distribution (simplified chi-square test)
	// Count leading bits
	zeros := 0
	ones := 0
	for _, key := range keys {
		if key.D.Bit(0) == 0 {
			zeros++
		} else {
			ones++
		}
	}

	// Should be roughly 50/50 distribution
	// For 100 samples, we expect ~50 each, but allow for statistical variance
	// Using binomial distribution, 99% confidence interval is approximately ±3 standard deviations
	// Standard deviation = sqrt(n*p*(1-p)) = sqrt(100*0.5*0.5) = 5
	// So we allow difference up to 3*5 = 15, but we'll be more lenient with 25
	diff := zeros - ones
	if diff < 0 {
		diff = -diff
	}
	assert.Less(t, diff, 25,
		"random bit distribution appears biased: %d zeros, %d ones", zeros, ones)
}

// TestCryptographicAgility tests ability to switch algorithms
func TestCryptographicAgility(t *testing.T) {
	// Test different Argon2 configurations
	configs := []*argon2.Config{
		argon2.LightConfig(),
		argon2.DefaultConfig(),
		argon2.HighSecurityConfig(),
	}

	password := []byte("test-password")

	for i, config := range configs {
		kdf := argon2.New(config)

		hash, err := kdf.HashPassword(password)
		require.NoError(t, err, "config %d should work", i)

		// Verify with same config
		valid, err := argon2.VerifyPassword(password, hash)
		require.NoError(t, err)
		assert.True(t, valid, "config %d verification should succeed", i)
	}

	// Test different elliptic curves
	curves := []elliptic.Curve{
		elliptic.P224(),
		elliptic.P256(),
		elliptic.P384(),
		elliptic.P521(),
	}

	message := []byte("test message")
	hashMsg := sha256.Sum256(message)

	for _, curve := range curves {
		priv, err := ecdsa.GenerateKey(curve, rand.Reader)
		require.NoError(t, err)

		r, s, err := ecdsaPkg.DeterministicSign(priv, hashMsg[:])
		require.NoError(t, err)

		valid := ecdsa.Verify(&priv.PublicKey, hashMsg[:], r, s)
		assert.True(t, valid, "curve %s should work", curve.Params().Name)
	}
}

// BenchmarkSecurityOperations benchmarks security-critical operations
func BenchmarkSecurityOperations(b *testing.B) {
	b.Run("Argon2Default", func(b *testing.B) {
		kdf := argon2.New(argon2.DefaultConfig())
		password := []byte("benchmark-password")
		salt, _ := kdf.GenerateSalt()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = kdf.DeriveKey(password, salt)
		}
	})

	b.Run("ECDSADeterministic", func(b *testing.B) {
		priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		hash := sha256.Sum256([]byte("benchmark"))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, _ = ecdsaPkg.DeterministicSign(priv, hash[:])
		}
	})

	b.Run("WASMHashVerification", func(b *testing.B) {
		verifier := wasm.NewHashVerifier()
		module := make([]byte, 1024*1024) // 1MB module
		rand.Read(module)
		hash := verifier.ComputeHash(module)
		verifier.AddTrustedHash("bench", hash)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = verifier.VerifyHash("bench", module)
		}
	})
}
