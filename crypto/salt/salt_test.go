package salt

import (
	"bytes"
	"fmt"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{"minimum size", MinSaltSize, false},
		{"default size", DefaultSaltSize, false},
		{"large size", 512, false},
		{"maximum size", MaxSaltSize, false},
		{"too small", MinSaltSize - 1, true},
		{"too large", MaxSaltSize + 1, true},
		{"zero size", 0, true},
		{"negative size", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			salt, err := Generate(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if salt == nil {
					t.Error("Generate() returned nil salt without error")
					return
				}
				if salt.Size() != tt.size {
					t.Errorf("Generate() size = %d, want %d", salt.Size(), tt.size)
				}
				if salt.IsEmpty() {
					t.Error("Generate() returned empty salt")
				}

				// Test that salt contains random data (not all zeros)
				saltBytes := salt.Bytes()
				allZeros := true
				for _, b := range saltBytes {
					if b != 0 {
						allZeros = false
						break
					}
				}
				if allZeros {
					t.Error("Generate() returned all-zero salt (likely not random)")
				}
			}
		})
	}
}

func TestGenerateDefault(t *testing.T) {
	salt, err := GenerateDefault()
	if err != nil {
		t.Fatalf("GenerateDefault() error = %v", err)
	}

	if salt.Size() != DefaultSaltSize {
		t.Errorf("GenerateDefault() size = %d, want %d", salt.Size(), DefaultSaltSize)
	}
}

func TestFromBytes(t *testing.T) {
	validBytes := make([]byte, DefaultSaltSize)
	for i := range validBytes {
		validBytes[i] = byte(i)
	}

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"valid default size", validBytes, false},
		{"minimum size", make([]byte, MinSaltSize), false},
		{"large size", make([]byte, 512), false},
		{"too small", make([]byte, MinSaltSize-1), true},
		{"too large", make([]byte, MaxSaltSize+1), true},
		{"empty", []byte{}, true},
		{"nil", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			salt, err := FromBytes(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if salt == nil {
					t.Error("FromBytes() returned nil salt without error")
					return
				}
				if salt.Size() != len(tt.data) {
					t.Errorf("FromBytes() size = %d, want %d", salt.Size(), len(tt.data))
				}

				// Verify data is copied correctly
				saltBytes := salt.Bytes()
				if !bytes.Equal(saltBytes, tt.data) {
					t.Error("FromBytes() data doesn't match input")
				}

				// Verify external modification doesn't affect salt
				if len(tt.data) > 0 {
					originalValue := tt.data[0]
					tt.data[0] = ^tt.data[0] // Flip bits
					if salt.Bytes()[0] != originalValue {
						t.Error("FromBytes() salt was affected by external modification")
					}
				}
			}
		})
	}
}

func TestSaltBytes(t *testing.T) {
	salt, err := GenerateDefault()
	if err != nil {
		t.Fatalf("Failed to generate salt: %v", err)
	}

	bytes1 := salt.Bytes()
	bytes2 := salt.Bytes()

	// Should return same data
	if !bytes.Equal(bytes1, bytes2) {
		t.Error("Bytes() returned different data on multiple calls")
	}

	// Should be independent copies
	if &bytes1[0] == &bytes2[0] {
		t.Error("Bytes() returned same underlying array (not a copy)")
	}

	// Modifying returned bytes shouldn't affect salt
	if len(bytes1) > 0 {
		originalValue := bytes1[0]
		bytes1[0] = ^bytes1[0]
		bytes3 := salt.Bytes()
		if bytes3[0] != originalValue {
			t.Error("External modification of Bytes() affected salt")
		}
	}
}

func TestSaltEqual(t *testing.T) {
	salt1, err := Generate(DefaultSaltSize)
	if err != nil {
		t.Fatalf("Failed to generate salt1: %v", err)
	}

	salt2, err := Generate(DefaultSaltSize)
	if err != nil {
		t.Fatalf("Failed to generate salt2: %v", err)
	}

	// Same salt data
	salt3, err := FromBytes(salt1.Bytes())
	if err != nil {
		t.Fatalf("Failed to create salt3: %v", err)
	}

	tests := []struct {
		name     string
		salt1    *Salt
		salt2    *Salt
		expected bool
	}{
		{"same salt", salt1, salt1, true},
		{"equivalent salts", salt1, salt3, true},
		{"different salts", salt1, salt2, false},
		{"nil salts", nil, nil, true},
		{"one nil salt", salt1, nil, false},
		{"nil vs non-nil", nil, salt1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.salt1.Equal(tt.salt2)
			if result != tt.expected {
				t.Errorf("Equal() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSaltClear(t *testing.T) {
	salt, err := GenerateDefault()
	if err != nil {
		t.Fatalf("Failed to generate salt: %v", err)
	}

	originalSize := salt.Size()
	if originalSize == 0 {
		t.Fatal("Salt size is zero before clear")
	}

	salt.Clear()

	if salt.Size() != 0 {
		t.Error("Salt size is not zero after clear")
	}

	if !salt.IsEmpty() {
		t.Error("Salt is not empty after clear")
	}

	bytes := salt.Bytes()
	if bytes != nil {
		t.Error("Bytes() should return nil after clear")
	}
}

func TestSaltStore(t *testing.T) {
	store := NewSaltStore()

	// Test empty store
	if store.Size() != 0 {
		t.Error("New store should be empty")
	}

	// Generate and store salt
	salt1, err := GenerateDefault()
	if err != nil {
		t.Fatalf("Failed to generate salt: %v", err)
	}

	err = store.Store("test1", salt1)
	if err != nil {
		t.Fatalf("Failed to store salt: %v", err)
	}

	if store.Size() != 1 {
		t.Errorf("Store size = %d, want 1", store.Size())
	}

	// Retrieve salt
	retrieved, err := store.Retrieve("test1")
	if err != nil {
		t.Fatalf("Failed to retrieve salt: %v", err)
	}

	if !salt1.Equal(retrieved) {
		t.Error("Retrieved salt doesn't match stored salt")
	}

	// Test generate and store
	salt2, err := store.GenerateAndStore("test2", DefaultSaltSize)
	if err != nil {
		t.Fatalf("Failed to generate and store salt: %v", err)
	}

	if store.Size() != 2 {
		t.Errorf("Store size = %d, want 2", store.Size())
	}

	if salt2.Size() != DefaultSaltSize {
		t.Errorf("Generated salt size = %d, want %d", salt2.Size(), DefaultSaltSize)
	}

	// Test list
	ids := store.List()
	if len(ids) != 2 {
		t.Errorf("List() returned %d ids, want 2", len(ids))
	}

	foundTest1 := false
	foundTest2 := false
	for _, id := range ids {
		if id == "test1" {
			foundTest1 = true
		}
		if id == "test2" {
			foundTest2 = true
		}
	}
	if !foundTest1 || !foundTest2 {
		t.Error("List() doesn't contain expected IDs")
	}

	// Test remove
	err = store.Remove("test1")
	if err != nil {
		t.Fatalf("Failed to remove salt: %v", err)
	}

	if store.Size() != 1 {
		t.Errorf("Store size = %d, want 1 after removal", store.Size())
	}

	_, err = store.Retrieve("test1")
	if err == nil {
		t.Error("Should not be able to retrieve removed salt")
	}

	// Test clear
	store.Clear()
	if store.Size() != 0 {
		t.Error("Store should be empty after clear")
	}
}

func TestSaltStoreErrors(t *testing.T) {
	store := NewSaltStore()

	// Test empty identifier errors
	err := store.Store("", nil)
	if err == nil {
		t.Error("Should error on empty identifier")
	}

	_, err = store.Retrieve("")
	if err == nil {
		t.Error("Should error on empty identifier")
	}

	err = store.Remove("")
	if err == nil {
		t.Error("Should error on empty identifier")
	}

	// Test nil salt error
	err = store.Store("test", nil)
	if err == nil {
		t.Error("Should error on nil salt")
	}

	// Test empty salt error
	emptySalt := &Salt{}
	err = store.Store("test", emptySalt)
	if err == nil {
		t.Error("Should error on empty salt")
	}

	// Test retrieve non-existent
	_, err = store.Retrieve("nonexistent")
	if err == nil {
		t.Error("Should error when retrieving non-existent salt")
	}

	// Test remove non-existent
	err = store.Remove("nonexistent")
	if err == nil {
		t.Error("Should error when removing non-existent salt")
	}
}

func TestConstantTimeCompare(t *testing.T) {
	a := []byte{1, 2, 3, 4, 5}
	b := []byte{1, 2, 3, 4, 5}
	c := []byte{1, 2, 3, 4, 6}
	d := []byte{1, 2, 3, 4}

	tests := []struct {
		name     string
		a, b     []byte
		expected bool
	}{
		{"equal slices", a, b, true},
		{"different content", a, c, false},
		{"different length", a, d, false},
		{"empty slices", []byte{}, []byte{}, true},
		{"one empty", a, []byte{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constantTimeCompare(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("constantTimeCompare() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func BenchmarkGenerate(b *testing.B) {
	sizes := []int{MinSaltSize, DefaultSaltSize, 512}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				salt, err := Generate(size)
				if err != nil {
					b.Fatalf("Generate error: %v", err)
				}
				salt.Clear() // Clean up
			}
		})
	}
}

func BenchmarkSaltEqual(b *testing.B) {
	salt1, _ := GenerateDefault()
	salt2, _ := GenerateDefault()
	salt3, _ := FromBytes(salt1.Bytes())

	b.Run("equal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			salt1.Equal(salt3)
		}
	})

	b.Run("different", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			salt1.Equal(salt2)
		}
	})
}
