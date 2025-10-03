// Package secure provides utilities for secure memory handling and sensitive data management.
// It includes functions for explicit memory zeroization and secure data lifecycle management.
package secure

import (
	"crypto/rand"
	"fmt"
	"runtime"
	"sync"
)

// SecureBytes wraps a byte slice with automatic cleanup functionality
type SecureBytes struct {
	data      []byte
	mu        sync.RWMutex
	finalized bool
}

// NewSecureBytes creates a new SecureBytes instance with automatic cleanup
func NewSecureBytes(size int) *SecureBytes {
	if size <= 0 {
		return &SecureBytes{data: nil}
	}

	sb := &SecureBytes{
		data: make([]byte, size),
	}

	// Set finalizer for automatic cleanup if Clear() is not called
	runtime.SetFinalizer(sb, (*SecureBytes).finalize)
	return sb
}

// FromBytes creates a SecureBytes instance from existing data (copies the data)
func FromBytes(data []byte) *SecureBytes {
	if len(data) == 0 {
		return &SecureBytes{data: nil}
	}

	sb := &SecureBytes{
		data: make([]byte, len(data)),
	}
	copy(sb.data, data)

	runtime.SetFinalizer(sb, (*SecureBytes).finalize)
	return sb
}

// Bytes returns a copy of the secure data to prevent external modification
func (sb *SecureBytes) Bytes() []byte {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if sb.data == nil {
		return nil
	}

	result := make([]byte, len(sb.data))
	copy(result, sb.data)
	return result
}

// Size returns the size of the secure data
func (sb *SecureBytes) Size() int {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if sb.data == nil {
		return 0
	}
	return len(sb.data)
}

// IsEmpty checks if the secure data is empty or nil
func (sb *SecureBytes) IsEmpty() bool {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	return sb.data == nil || len(sb.data) == 0
}

// Clear explicitly zeros the memory and removes the finalizer
func (sb *SecureBytes) Clear() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if !sb.finalized && sb.data != nil {
		Zeroize(sb.data)
		sb.data = nil
		sb.finalized = true
		runtime.SetFinalizer(sb, nil) // Remove finalizer since we've cleaned up
	}
}

// finalize is called by the garbage collector if Clear() was not called
func (sb *SecureBytes) finalize() {
	sb.Clear()
}

// CopyTo safely copies data to the secure buffer
func (sb *SecureBytes) CopyTo(data []byte) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.finalized {
		return fmt.Errorf("secure bytes has been finalized")
	}

	if sb.data == nil {
		return fmt.Errorf("secure bytes is nil")
	}

	if len(data) > len(sb.data) {
		return fmt.Errorf("data size %d exceeds secure buffer size %d", len(data), len(sb.data))
	}

	// Zero existing data first
	Zeroize(sb.data)
	copy(sb.data, data)
	return nil
}

// Zeroize explicitly zeros out sensitive data from memory using byte slicing
func Zeroize(data []byte) {
	if len(data) == 0 {
		return
	}

	// Explicitly zero each byte to prevent compiler optimizations
	for i := range data {
		data[i] = 0
	}

	// Force memory barrier to ensure zeroization is not optimized away
	runtime.KeepAlive(data)
}

// ZeroizeString attempts to clear a string reference (limited effectiveness)
// Note: Go strings are immutable, so this only clears the reference, not the underlying data
func ZeroizeString(s *string) {
	if s == nil {
		return
	}
	// Simply clear the reference - Go strings are immutable
	*s = ""
}

// SecureString wraps a string with secure cleanup capabilities
type SecureString struct {
	value     string
	mu        sync.RWMutex
	finalized bool
}

// NewSecureString creates a new SecureString with automatic cleanup
func NewSecureString(s string) *SecureString {
	ss := &SecureString{
		value: s,
	}
	runtime.SetFinalizer(ss, (*SecureString).finalize)
	return ss
}

// String returns the secure string value
func (ss *SecureString) String() string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if ss.finalized {
		return ""
	}
	return ss.value
}

// Clear attempts to zero the string and marks it as finalized
func (ss *SecureString) Clear() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if !ss.finalized {
		ZeroizeString(&ss.value)
		ss.value = ""
		ss.finalized = true
		runtime.SetFinalizer(ss, nil)
	}
}

// finalize is called by the garbage collector
func (ss *SecureString) finalize() {
	ss.Clear()
}

// IsEmpty checks if the secure string is empty
func (ss *SecureString) IsEmpty() bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	return ss.finalized || ss.value == ""
}

// SecureBuffer provides a reusable buffer for sensitive operations
type SecureBuffer struct {
	buffer []byte
	mu     sync.Mutex
}

// NewSecureBuffer creates a new secure buffer with the specified capacity
func NewSecureBuffer(capacity int) *SecureBuffer {
	if capacity <= 0 {
		capacity = 1024 // Default capacity
	}

	sb := &SecureBuffer{
		buffer: make([]byte, 0, capacity),
	}
	runtime.SetFinalizer(sb, (*SecureBuffer).finalize)
	return sb
}

// Write appends data to the secure buffer
func (sb *SecureBuffer) Write(data []byte) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if len(sb.buffer)+len(data) > cap(sb.buffer) {
		return fmt.Errorf("buffer overflow: capacity %d, current size %d, write size %d",
			cap(sb.buffer), len(sb.buffer), len(data))
	}

	sb.buffer = append(sb.buffer, data...)
	return nil
}

// Read returns a copy of the buffer contents
func (sb *SecureBuffer) Read() []byte {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	result := make([]byte, len(sb.buffer))
	copy(result, sb.buffer)
	return result
}

// Reset clears the buffer contents but maintains capacity
func (sb *SecureBuffer) Reset() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if len(sb.buffer) > 0 {
		Zeroize(sb.buffer[:cap(sb.buffer)]) // Zero the entire backing array
		sb.buffer = sb.buffer[:0]           // Reset length to 0
	}
}

// Clear zeros the buffer and releases memory
func (sb *SecureBuffer) Clear() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.buffer != nil {
		Zeroize(sb.buffer[:cap(sb.buffer)]) // Zero entire backing array
		sb.buffer = nil
		runtime.SetFinalizer(sb, nil)
	}
}

// finalize is called by the garbage collector
func (sb *SecureBuffer) finalize() {
	sb.Clear()
}

// Size returns the current size of data in the buffer
func (sb *SecureBuffer) Size() int {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	return len(sb.buffer)
}

// Capacity returns the maximum capacity of the buffer
func (sb *SecureBuffer) Capacity() int {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.buffer == nil {
		return 0
	}
	return cap(sb.buffer)
}

// ZeroizeMultiple zeros multiple byte slices in a single call
func ZeroizeMultiple(slices ...[]byte) {
	for _, slice := range slices {
		Zeroize(slice)
	}
}

// SecureCompare performs constant-time comparison of two byte slices
// Returns true if the slices are equal, false otherwise
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

// SecureRandom fills the provided slice with cryptographically secure random bytes
func SecureRandom(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// Use Go's crypto/rand for secure random generation
	if _, err := rand.Read(data); err != nil {
		return fmt.Errorf("failed to generate secure random bytes: %w", err)
	}

	return nil
}
