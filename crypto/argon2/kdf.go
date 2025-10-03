// Package argon2 provides secure key derivation using Argon2id
package argon2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Config defines Argon2id parameters
type Config struct {
	Time        uint32 // Number of iterations
	Memory      uint32 // Memory in KB
	Parallelism uint8  // Number of threads
	SaltLength  uint32 // Salt length in bytes
	KeyLength   uint32 // Output key length in bytes
}

// DefaultConfig returns secure default parameters
func DefaultConfig() *Config {
	return &Config{
		Time:        1,
		Memory:      64 * 1024, // 64MB
		Parallelism: 4,
		SaltLength:  32,
		KeyLength:   32,
	}
}

// LightConfig returns lighter parameters for testing
func LightConfig() *Config {
	return &Config{
		Time:        1,
		Memory:      16 * 1024, // 16MB
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// HighSecurityConfig returns high-security parameters
func HighSecurityConfig() *Config {
	return &Config{
		Time:        3,
		Memory:      128 * 1024, // 128MB
		Parallelism: 4,
		SaltLength:  32,
		KeyLength:   32,
	}
}

// KDF implements Argon2id key derivation
type KDF struct {
	config *Config
}

// New creates a new Argon2id KDF with the given configuration
func New(config *Config) *KDF {
	if config == nil {
		config = DefaultConfig()
	}
	return &KDF{config: config}
}

// DeriveKey derives a key from password and salt
func (k *KDF) DeriveKey(password []byte, salt []byte) []byte {
	return argon2.IDKey(
		password,
		salt,
		k.config.Time,
		k.config.Memory,
		k.config.Parallelism,
		k.config.KeyLength,
	)
}

// GenerateSalt generates a cryptographically secure salt
func (k *KDF) GenerateSalt() ([]byte, error) {
	salt := make([]byte, k.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// HashPassword generates a hash with embedded salt and parameters
func (k *KDF) HashPassword(password []byte) (string, error) {
	salt, err := k.GenerateSalt()
	if err != nil {
		return "", err
	}

	hash := k.DeriveKey(password, salt)

	// Encode in PHC format: $argon2id$v=19$m=65536,t=1,p=4$salt$hash
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		k.config.Memory,
		k.config.Time,
		k.config.Parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

// VerifyPassword verifies a password against a PHC-formatted hash
func VerifyPassword(password []byte, encodedHash string) (bool, error) {
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	kdf := &KDF{config: params}
	derivedHash := kdf.DeriveKey(password, salt)

	// Constant-time comparison
	return subtle.ConstantTimeCompare(hash, derivedHash) == 1, nil
}

// decodeHash parses PHC-formatted Argon2id hash
func decodeHash(encodedHash string) (*Config, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, fmt.Errorf("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse version: %w", err)
	}

	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("unsupported Argon2 version: %d", version)
	}

	var memory, time uint32
	var parallelism uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &parallelism)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	config := &Config{
		Time:        time,
		Memory:      memory,
		Parallelism: parallelism,
		SaltLength:  uint32(len(salt)),
		KeyLength:   uint32(len(hash)),
	}

	return config, salt, hash, nil
}

// CompareHashes performs constant-time comparison of two hashes
func CompareHashes(hash1, hash2 []byte) bool {
	return subtle.ConstantTimeCompare(hash1, hash2) == 1
}

// ValidateConfig validates Argon2id parameters
func ValidateConfig(config *Config) error {
	if config.Time < 1 {
		return fmt.Errorf("time must be at least 1")
	}
	if config.Memory < 8*1024 {
		return fmt.Errorf("memory must be at least 8MB")
	}
	if config.Parallelism < 1 {
		return fmt.Errorf("parallelism must be at least 1")
	}
	if config.SaltLength < 8 {
		return fmt.Errorf("salt length must be at least 8 bytes")
	}
	if config.KeyLength < 16 {
		return fmt.Errorf("key length must be at least 16 bytes")
	}
	return nil
}

// EstimateTime estimates the time required for key derivation
func EstimateTime(config *Config, iterations int) string {
	// This is a rough estimate - actual time depends on hardware
	baseTime := float64(config.Time) * float64(config.Memory) / (64 * 1024)
	totalTime := baseTime * float64(iterations)

	if totalTime < 1 {
		return fmt.Sprintf("%.2f ms", totalTime*1000)
	} else if totalTime < 60 {
		return fmt.Sprintf("%.2f s", totalTime)
	}
	return fmt.Sprintf("%.2f min", totalTime/60)
}
