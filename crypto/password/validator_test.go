package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_Validate(t *testing.T) {
	validator := NewValidator(DefaultPasswordConfig())

	testCases := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "too short",
			password: "Short1!",
			wantErr:  true,
			errMsg:   "at least 12 characters",
		},
		{
			name:     "too long",
			password: string(make([]byte, 129)),
			wantErr:  true,
			errMsg:   "not exceed 128 characters",
		},
		{
			name:     "missing uppercase",
			password: "longenoughpassword123!",
			wantErr:  true,
			errMsg:   "uppercase letter",
		},
		{
			name:     "missing lowercase",
			password: "LONGENOUGHPASSWORD123!",
			wantErr:  true,
			errMsg:   "lowercase letter",
		},
		{
			name:     "missing digit",
			password: "LongEnoughPassword!",
			wantErr:  true,
			errMsg:   "one digit",
		},
		{
			name:     "missing special",
			password: "LongEnoughPassword123",
			wantErr:  true,
			errMsg:   "special character",
		},
		{
			name:     "valid password",
			password: "ValidPassword123!",
			wantErr:  false,
		},
		{
			name:     "complex valid password",
			password: "MyS3cur3P@ssw0rd!2024",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.Validate([]byte(tc.password))
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_CustomConfig(t *testing.T) {
	config := &PasswordConfig{
		MinLength:        8,
		MaxLength:        64,
		RequireUppercase: false,
		RequireLowercase: true,
		RequireDigits:    true,
		RequireSpecial:   false,
		MinEntropy:       30.0,
	}

	validator := NewValidator(config)

	// Should pass with custom config
	err := validator.Validate([]byte("simple123"))
	assert.NoError(t, err)

	// Should fail - too short
	err = validator.Validate([]byte("short1"))
	assert.Error(t, err)

	// Should fail - no digits
	err = validator.Validate([]byte("simplepass"))
	assert.Error(t, err)
}

func TestGenerateSalt(t *testing.T) {
	// Test valid salt generation
	salt, err := GenerateSalt(32)
	require.NoError(t, err)
	assert.Len(t, salt, 32)

	// Test different salt each time
	salt2, err := GenerateSalt(32)
	require.NoError(t, err)
	assert.NotEqual(t, salt, salt2)

	// Test minimum size enforcement
	_, err = GenerateSalt(8)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 16 bytes")
}

func TestSecureCompare(t *testing.T) {
	// Test equal slices
	a := []byte("password")
	b := []byte("password")
	assert.True(t, SecureCompare(a, b))

	// Test different slices
	c := []byte("different")
	assert.False(t, SecureCompare(a, c))

	// Test different lengths
	d := []byte("pass")
	assert.False(t, SecureCompare(a, d))

	// Test empty slices
	assert.True(t, SecureCompare([]byte{}, []byte{}))
}

func TestZeroBytes(t *testing.T) {
	password := []byte("sensitive")
	ZeroBytes(password)

	for _, b := range password {
		assert.Equal(t, byte(0), b)
	}
}

func TestCalculateEntropy(t *testing.T) {
	validator := NewValidator(nil)

	testCases := []struct {
		name       string
		password   string
		minEntropy float64
	}{
		{
			name:       "lowercase only",
			password:   "abcdefghij",
			minEntropy: 40,
		},
		{
			name:       "alphanumeric",
			password:   "Abc123",
			minEntropy: 30,
		},
		{
			name:       "complex",
			password:   "MyP@ssw0rd!",
			minEntropy: 50,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entropy := validator.calculateEntropy([]byte(tc.password))
			assert.GreaterOrEqual(t, entropy, tc.minEntropy)
		})
	}
}

func BenchmarkValidate(b *testing.B) {
	validator := NewValidator(DefaultPasswordConfig())
	password := []byte("ValidPassword123!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(password)
	}
}

func BenchmarkGenerateSalt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateSalt(32)
	}
}

func BenchmarkSecureCompare(b *testing.B) {
	a := []byte("password123")
	c := []byte("password123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SecureCompare(a, c)
	}
}
