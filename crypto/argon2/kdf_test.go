package argon2

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKDF_DeriveKey(t *testing.T) {
	kdf := New(DefaultConfig())

	password := []byte("test-password")
	salt := []byte("salt-must-be-at-least-16-bytes!!")

	// Derive key
	key := kdf.DeriveKey(password, salt)
	assert.Len(t, key, int(kdf.config.KeyLength))

	// Same inputs should produce same key
	key2 := kdf.DeriveKey(password, salt)
	assert.Equal(t, key, key2)

	// Different password should produce different key
	key3 := kdf.DeriveKey([]byte("different"), salt)
	assert.NotEqual(t, key, key3)

	// Different salt should produce different key
	salt2 := []byte("different-salt-at-least-16-bytes")
	key4 := kdf.DeriveKey(password, salt2)
	assert.NotEqual(t, key, key4)
}

func TestKDF_GenerateSalt(t *testing.T) {
	kdf := New(DefaultConfig())

	salt1, err := kdf.GenerateSalt()
	require.NoError(t, err)
	assert.Len(t, salt1, int(kdf.config.SaltLength))

	// Should generate different salt each time
	salt2, err := kdf.GenerateSalt()
	require.NoError(t, err)
	assert.NotEqual(t, salt1, salt2)
}

func TestKDF_HashPassword(t *testing.T) {
	kdf := New(DefaultConfig())
	password := []byte("MySecureP@ssw0rd")

	hash, err := kdf.HashPassword(password)
	require.NoError(t, err)

	// Check format
	assert.True(t, strings.HasPrefix(hash, "$argon2id$"))
	parts := strings.Split(hash, "$")
	assert.Len(t, parts, 6)

	// Verify password
	valid, err := VerifyPassword(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)

	// Wrong password should fail
	valid, err = VerifyPassword([]byte("wrong"), hash)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestVerifyPassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		hash     string
		valid    bool
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			hash:     "$argon2id$v=19$m=65536,t=1,p=4$c2FsdC1tdXN0LWJlLWF0LWxlYXN0LTE2LWJ5dGVzISE$+4smaTt/N7ivKLrqsPIbTplUxDBRMxTKCYOcXWTJOEI",
			valid:    true,
		},
		{
			name:     "invalid password",
			password: "wrongpassword",
			hash:     "$argon2id$v=19$m=65536,t=1,p=4$c2FsdC1tdXN0LWJlLWF0LWxlYXN0LTE2LWJ5dGVzISE$+4smaTt/N7ivKLrqsPIbTplUxDBRMxTKCYOcXWTJOEI",
			valid:    false,
		},
		{
			name:     "invalid format",
			password: "password",
			hash:     "invalid-hash-format",
			wantErr:  true,
		},
		{
			name:     "wrong algorithm",
			password: "password",
			hash:     "$bcrypt$v=19$m=65536,t=1,p=4$salt$hash",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid, err := VerifyPassword([]byte(tc.password), tc.hash)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.valid, valid)
			}
		})
	}
}

func TestConfigurations(t *testing.T) {
	configs := map[string]*Config{
		"default": DefaultConfig(),
		"light":   LightConfig(),
		"high":    HighSecurityConfig(),
	}

	for name, config := range configs {
		t.Run(name, func(t *testing.T) {
			err := ValidateConfig(config)
			assert.NoError(t, err)

			kdf := New(config)
			password := []byte("test-password")

			hash, err := kdf.HashPassword(password)
			require.NoError(t, err)

			valid, err := VerifyPassword(password, hash)
			require.NoError(t, err)
			assert.True(t, valid)
		})
	}
}

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "time too low",
			config: &Config{
				Time:        0,
				Memory:      64 * 1024,
				Parallelism: 4,
				SaltLength:  32,
				KeyLength:   32,
			},
			wantErr: true,
			errMsg:  "time must be at least 1",
		},
		{
			name: "memory too low",
			config: &Config{
				Time:        1,
				Memory:      4 * 1024,
				Parallelism: 4,
				SaltLength:  32,
				KeyLength:   32,
			},
			wantErr: true,
			errMsg:  "memory must be at least 8MB",
		},
		{
			name: "parallelism too low",
			config: &Config{
				Time:        1,
				Memory:      64 * 1024,
				Parallelism: 0,
				SaltLength:  32,
				KeyLength:   32,
			},
			wantErr: true,
			errMsg:  "parallelism must be at least 1",
		},
		{
			name: "salt too short",
			config: &Config{
				Time:        1,
				Memory:      64 * 1024,
				Parallelism: 4,
				SaltLength:  4,
				KeyLength:   32,
			},
			wantErr: true,
			errMsg:  "salt length must be at least 8 bytes",
		},
		{
			name: "key too short",
			config: &Config{
				Time:        1,
				Memory:      64 * 1024,
				Parallelism: 4,
				SaltLength:  32,
				KeyLength:   8,
			},
			wantErr: true,
			errMsg:  "key length must be at least 16 bytes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateConfig(tc.config)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCompareHashes(t *testing.T) {
	hash1 := []byte("hash1")
	hash2 := []byte("hash1")
	hash3 := []byte("hash2")

	assert.True(t, CompareHashes(hash1, hash2))
	assert.False(t, CompareHashes(hash1, hash3))
	assert.False(t, CompareHashes([]byte("short"), []byte("longer")))
}

func TestEstimateTime(t *testing.T) {
	config := DefaultConfig()

	estimate := EstimateTime(config, 1)
	assert.NotEmpty(t, estimate)
	assert.True(t, strings.HasSuffix(estimate, "ms") ||
		strings.HasSuffix(estimate, "s") ||
		strings.HasSuffix(estimate, "min"))

	// Test different scales
	estimate = EstimateTime(config, 100)
	assert.NotEmpty(t, estimate)

	// High security config should take longer
	highConfig := HighSecurityConfig()
	highEstimate := EstimateTime(highConfig, 1)
	assert.NotEmpty(t, highEstimate)
}

func TestDecodeHash(t *testing.T) {
	validHash := "$argon2id$v=19$m=65536,t=1,p=4$c2FsdA$aGFzaA"

	config, salt, hash, err := decodeHash(validHash)
	require.NoError(t, err)

	assert.Equal(t, uint32(65536), config.Memory)
	assert.Equal(t, uint32(1), config.Time)
	assert.Equal(t, uint8(4), config.Parallelism)
	assert.Equal(t, []byte("salt"), salt)
	assert.Equal(t, []byte("hash"), hash)

	// Test invalid formats
	invalidHashes := []string{
		"invalid",
		"$bcrypt$v=19$m=65536,t=1,p=4$salt$hash",
		"$argon2id$v=18$m=65536,t=1,p=4$salt$hash", // wrong version
		"$argon2id$v=19$invalid$salt$hash",
		"$argon2id$v=19$m=65536,t=1,p=4$!invalid!$hash",
	}

	for _, h := range invalidHashes {
		_, _, _, err := decodeHash(h)
		assert.Error(t, err)
	}
}

func BenchmarkDeriveKey(b *testing.B) {
	configs := map[string]*Config{
		"light":   LightConfig(),
		"default": DefaultConfig(),
		"high":    HighSecurityConfig(),
	}

	password := []byte("benchmark-password")
	salt := []byte("benchmark-salt-at-least-16-bytes")

	for name, config := range configs {
		b.Run(name, func(b *testing.B) {
			kdf := New(config)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = kdf.DeriveKey(password, salt)
			}
		})
	}
}

func BenchmarkHashPassword(b *testing.B) {
	kdf := New(LightConfig()) // Use light config for benchmarks
	password := []byte("benchmark-password")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = kdf.HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	kdf := New(LightConfig())
	password := []byte("benchmark-password")
	hash, _ := kdf.HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = VerifyPassword(password, hash)
	}
}

func TestConcurrentDerivation(t *testing.T) {
	kdf := New(DefaultConfig())
	password := []byte("concurrent-test")

	// Generate multiple salts
	salts := make([][]byte, 10)
	for i := range salts {
		salt, err := kdf.GenerateSalt()
		require.NoError(t, err)
		salts[i] = salt
	}

	// Derive keys concurrently
	results := make([][]byte, len(salts))
	done := make(chan int, len(salts))

	for i, salt := range salts {
		go func(idx int, s []byte) {
			results[idx] = kdf.DeriveKey(password, s)
			done <- idx
		}(i, salt)
	}

	// Wait for all goroutines
	for i := 0; i < len(salts); i++ {
		<-done
	}

	// Verify all keys are different (different salts)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			assert.False(t, bytes.Equal(results[i], results[j]))
		}
	}
}

func TestPerformanceBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance benchmark in short mode")
	}

	configs := []struct {
		name   string
		config *Config
		maxMs  int64
	}{
		{"light", LightConfig(), 100},
		{"default", DefaultConfig(), 500},
	}

	password := []byte("perf-test")

	for _, tc := range configs {
		t.Run(tc.name, func(t *testing.T) {
			kdf := New(tc.config)
			salt, _ := kdf.GenerateSalt()

			start := time.Now()
			_ = kdf.DeriveKey(password, salt)
			elapsed := time.Since(start).Milliseconds()

			t.Logf("%s config took %dms", tc.name, elapsed)
			assert.Less(t, elapsed, tc.maxMs,
				"derivation took too long: %dms > %dms", elapsed, tc.maxMs)
		})
	}
}
