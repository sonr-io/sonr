// Package salt provides cryptographically secure salt generation and management
// for use in key derivation functions and encryption operations.
package salt

import (
	"crypto/rand"
	"fmt"
	"io"
)

const (
	// DefaultSaltSize defines the recommended salt size (256 bits)
	DefaultSaltSize = 32
	// MinSaltSize defines the minimum acceptable salt size (128 bits)
	MinSaltSize = 16
	// MaxSaltSize defines the maximum salt size to prevent resource exhaustion
	MaxSaltSize = 1024
)

// Salt represents a cryptographically secure salt value
type Salt struct {
	value []byte
}

// Generate creates a new cryptographically secure salt of the specified size
func Generate(size int) (*Salt, error) {
	if size < MinSaltSize {
		return nil, fmt.Errorf("salt size too small: minimum %d bytes required", MinSaltSize)
	}
	if size > MaxSaltSize {
		return nil, fmt.Errorf("salt size too large: maximum %d bytes allowed", MaxSaltSize)
	}

	saltBytes := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, saltBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %w", err)
	}

	return &Salt{value: saltBytes}, nil
}

// GenerateDefault creates a new salt with the default recommended size
func GenerateDefault() (*Salt, error) {
	return Generate(DefaultSaltSize)
}

// FromBytes creates a Salt from existing bytes (validates minimum size)
func FromBytes(data []byte) (*Salt, error) {
	if len(data) < MinSaltSize {
		return nil, fmt.Errorf("salt too small: minimum %d bytes required, got %d", MinSaltSize, len(data))
	}
	if len(data) > MaxSaltSize {
		return nil, fmt.Errorf("salt too large: maximum %d bytes allowed, got %d", MaxSaltSize, len(data))
	}

	// Copy to prevent external modification
	saltBytes := make([]byte, len(data))
	copy(saltBytes, data)

	return &Salt{value: saltBytes}, nil
}

// Bytes returns a copy of the salt bytes to prevent external modification
func (s *Salt) Bytes() []byte {
	if s == nil || s.value == nil {
		return nil
	}
	result := make([]byte, len(s.value))
	copy(result, s.value)
	return result
}

// Size returns the size of the salt in bytes
func (s *Salt) Size() int {
	if s == nil || s.value == nil {
		return 0
	}
	return len(s.value)
}

// String returns a redacted string representation for logging (does not expose salt value)
func (s *Salt) String() string {
	if s == nil || s.value == nil {
		return "Salt{<nil>}"
	}
	return fmt.Sprintf("Salt{size=%d}", len(s.value))
}

// Equal compares two salts in constant time to prevent timing attacks
func (s *Salt) Equal(other *Salt) bool {
	if s == nil || other == nil {
		return s == other
	}
	if len(s.value) != len(other.value) {
		return false
	}
	return constantTimeCompare(s.value, other.value)
}

// Clear securely zeros the salt value from memory
func (s *Salt) Clear() {
	if s != nil && s.value != nil {
		for i := range s.value {
			s.value[i] = 0
		}
		s.value = nil
	}
}

// IsEmpty checks if the salt is nil or has zero length
func (s *Salt) IsEmpty() bool {
	return s == nil || s.value == nil || len(s.value) == 0
}

// constantTimeCompare performs constant-time comparison of two byte slices
func constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// SaltStore manages storage and retrieval of salts with metadata
type SaltStore struct {
	salts map[string]*Salt
}

// NewSaltStore creates a new salt store for managing multiple salts
func NewSaltStore() *SaltStore {
	return &SaltStore{
		salts: make(map[string]*Salt),
	}
}

// Store saves a salt with the given identifier
func (ss *SaltStore) Store(id string, salt *Salt) error {
	if id == "" {
		return fmt.Errorf("salt identifier cannot be empty")
	}
	if salt == nil || salt.IsEmpty() {
		return fmt.Errorf("salt cannot be nil or empty")
	}

	// Store a copy to prevent external modification
	ss.salts[id] = &Salt{
		value: salt.Bytes(),
	}

	return nil
}

// Retrieve gets a salt by its identifier
func (ss *SaltStore) Retrieve(id string) (*Salt, error) {
	if id == "" {
		return nil, fmt.Errorf("salt identifier cannot be empty")
	}

	salt, exists := ss.salts[id]
	if !exists {
		return nil, fmt.Errorf("salt not found for identifier: %s", id)
	}

	// Return a copy to prevent external modification
	return &Salt{
		value: salt.Bytes(),
	}, nil
}

// Remove deletes a salt from the store and clears its memory
func (ss *SaltStore) Remove(id string) error {
	if id == "" {
		return fmt.Errorf("salt identifier cannot be empty")
	}

	salt, exists := ss.salts[id]
	if !exists {
		return fmt.Errorf("salt not found for identifier: %s", id)
	}

	// Clear the salt from memory before removal
	salt.Clear()
	delete(ss.salts, id)

	return nil
}

// List returns all stored salt identifiers
func (ss *SaltStore) List() []string {
	ids := make([]string, 0, len(ss.salts))
	for id := range ss.salts {
		ids = append(ids, id)
	}
	return ids
}

// Clear removes all salts and clears their memory
func (ss *SaltStore) Clear() {
	for id, salt := range ss.salts {
		salt.Clear()
		delete(ss.salts, id)
	}
}

// Size returns the number of stored salts
func (ss *SaltStore) Size() int {
	return len(ss.salts)
}

// GenerateAndStore creates a new salt and stores it with the given identifier
func (ss *SaltStore) GenerateAndStore(id string, size int) (*Salt, error) {
	salt, err := Generate(size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	if err := ss.Store(id, salt); err != nil {
		salt.Clear() // Clean up on failure
		return nil, fmt.Errorf("failed to store salt: %w", err)
	}

	return salt, nil
}
