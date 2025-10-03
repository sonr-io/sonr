// Package password provides secure password handling and validation
package password

import (
	"crypto/rand"
	"fmt"
	"unicode"
)

// PasswordConfig defines password policy requirements
type PasswordConfig struct {
	MinLength        int
	MaxLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireDigits    bool
	RequireSpecial   bool
	MinEntropy       float64
}

// DefaultPasswordConfig returns secure default password requirements
func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		MinLength:        12,
		MaxLength:        128,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigits:    true,
		RequireSpecial:   true,
		MinEntropy:       50.0, // bits
	}
}

// Validator validates passwords against security policies
type Validator struct {
	config *PasswordConfig
}

// NewValidator creates a password validator with the given configuration
func NewValidator(config *PasswordConfig) *Validator {
	if config == nil {
		config = DefaultPasswordConfig()
	}
	return &Validator{config: config}
}

// Validate checks if a password meets security requirements
func (v *Validator) Validate(password []byte) error {
	// Check length
	if len(password) < v.config.MinLength {
		return fmt.Errorf("password must be at least %d characters", v.config.MinLength)
	}
	if len(password) > v.config.MaxLength {
		return fmt.Errorf("password must not exceed %d characters", v.config.MaxLength)
	}

	// Check character requirements
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range string(password) {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsSpace(ch):
			// Spaces are allowed but not counted as special
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if v.config.RequireUppercase && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if v.config.RequireLowercase && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if v.config.RequireDigits && !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if v.config.RequireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check entropy
	entropy := v.calculateEntropy(password)
	if entropy < v.config.MinEntropy {
		return fmt.Errorf("password entropy too low: %.1f bits (minimum: %.1f)",
			entropy, v.config.MinEntropy)
	}

	return nil
}

// calculateEntropy estimates password entropy in bits
func (v *Validator) calculateEntropy(password []byte) float64 {
	// Count unique characters
	charSet := make(map[byte]bool)
	for _, b := range password {
		charSet[b] = true
	}

	// Estimate character pool size
	poolSize := 0
	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for ch := range charSet {
		r := rune(ch)
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if hasLower {
		poolSize += 26
	}
	if hasUpper {
		poolSize += 26
	}
	if hasDigit {
		poolSize += 10
	}
	if hasSpecial {
		poolSize += 32 // Common special characters
	}

	if poolSize == 0 {
		return 0
	}

	// Calculate entropy: log2(poolSize^length)
	// Simplified: length * log2(poolSize)
	bitsPerChar := 0.0
	temp := poolSize
	for temp > 0 {
		bitsPerChar++
		temp >>= 1
	}

	return float64(len(password)) * bitsPerChar
}

// GenerateSalt generates a cryptographically secure random salt
func GenerateSalt(size int) ([]byte, error) {
	if size < 16 {
		return nil, fmt.Errorf("salt size must be at least 16 bytes")
	}

	salt := make([]byte, size)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	return salt, nil
}

// SecureCompare performs constant-time comparison of two byte slices
func SecureCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// ZeroBytes overwrites a byte slice with zeros
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
