package secure

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestZeroize(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty slice", []byte{}},
		{"single byte", []byte{0xFF}},
		{"small slice", []byte{1, 2, 3, 4, 5}},
		{"large slice", make([]byte, 1024)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill with non-zero data
			for i := range tt.data {
				tt.data[i] = byte(i%256 + 1)
			}

			// Store original for verification
			original := make([]byte, len(tt.data))
			copy(original, tt.data)

			// Zeroize
			Zeroize(tt.data)

			// Verify all bytes are zero
			for i, b := range tt.data {
				if b != 0 {
					t.Errorf("Byte at index %d not zeroed: got %d, want 0", i, b)
				}
			}

			// Verify original data was actually non-zero (for non-empty slices)
			if len(original) > 0 {
				hasNonZero := false
				for _, b := range original {
					if b != 0 {
						hasNonZero = true
						break
					}
				}
				if !hasNonZero {
					t.Error("Test data was already all zeros - invalid test")
				}
			}
		})
	}
}

func TestSecureBytes(t *testing.T) {
	t.Run("NewSecureBytes", func(t *testing.T) {
		sb := NewSecureBytes(32)
		if sb == nil {
			t.Fatal("NewSecureBytes returned nil")
		}
		if sb.Size() != 32 {
			t.Errorf("Size() = %d, want 32", sb.Size())
		}
		if sb.IsEmpty() {
			t.Error("NewSecureBytes should not be empty")
		}
		sb.Clear()
	})

	t.Run("zero size", func(t *testing.T) {
		sb := NewSecureBytes(0)
		if sb.Size() != 0 {
			t.Errorf("Size() = %d, want 0", sb.Size())
		}
		if !sb.IsEmpty() {
			t.Error("Zero-size SecureBytes should be empty")
		}
	})

	t.Run("FromBytes", func(t *testing.T) {
		original := []byte{1, 2, 3, 4, 5}
		sb := FromBytes(original)

		if sb.Size() != len(original) {
			t.Errorf("Size() = %d, want %d", sb.Size(), len(original))
		}

		retrieved := sb.Bytes()
		if !bytes.Equal(retrieved, original) {
			t.Error("Retrieved bytes don't match original")
		}

		// Verify independence - modifying original shouldn't affect SecureBytes
		original[0] = 99
		retrieved2 := sb.Bytes()
		if retrieved2[0] == 99 {
			t.Error("SecureBytes was affected by external modification")
		}

		sb.Clear()
	})

	t.Run("Bytes returns copy", func(t *testing.T) {
		sb := NewSecureBytes(16)

		bytes1 := sb.Bytes()
		bytes2 := sb.Bytes()

		// Should be equal content
		if !bytes.Equal(bytes1, bytes2) {
			t.Error("Multiple Bytes() calls returned different content")
		}

		// Should be different slices
		if len(bytes1) > 0 && &bytes1[0] == &bytes2[0] {
			t.Error("Bytes() returned same underlying array")
		}

		// Modifying returned slice shouldn't affect SecureBytes
		if len(bytes1) > 0 {
			bytes1[0] = 0xFF
			bytes3 := sb.Bytes()
			if bytes3[0] == 0xFF {
				t.Error("External modification affected SecureBytes")
			}
		}

		sb.Clear()
	})

	t.Run("Clear", func(t *testing.T) {
		data := []byte{1, 2, 3, 4, 5}
		sb := FromBytes(data)

		if sb.IsEmpty() {
			t.Error("SecureBytes should not be empty before Clear")
		}

		sb.Clear()

		if !sb.IsEmpty() {
			t.Error("SecureBytes should be empty after Clear")
		}

		if sb.Size() != 0 {
			t.Error("Size should be 0 after Clear")
		}

		bytes := sb.Bytes()
		if bytes != nil {
			t.Error("Bytes() should return nil after Clear")
		}
	})

	t.Run("CopyTo", func(t *testing.T) {
		sb := NewSecureBytes(10)
		data := []byte{1, 2, 3, 4, 5}

		err := sb.CopyTo(data)
		if err != nil {
			t.Fatalf("CopyTo failed: %v", err)
		}

		retrieved := sb.Bytes()
		if !bytes.Equal(retrieved[:len(data)], data) {
			t.Error("CopyTo didn't copy data correctly")
		}

		// Test overflow
		largeData := make([]byte, 20)
		err = sb.CopyTo(largeData)
		if err == nil {
			t.Error("CopyTo should fail with oversized data")
		}

		sb.Clear()

		// Test copy to finalized
		err = sb.CopyTo(data)
		if err == nil {
			t.Error("CopyTo should fail on finalized SecureBytes")
		}
	})
}

func TestSecureString(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		original := "sensitive data"
		ss := NewSecureString(original)

		if ss.String() != original {
			t.Error("String() doesn't match original")
		}

		if ss.IsEmpty() {
			t.Error("SecureString should not be empty")
		}

		ss.Clear()

		if !ss.IsEmpty() {
			t.Error("SecureString should be empty after Clear")
		}

		if ss.String() != "" {
			t.Error("String() should return empty string after Clear")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		ss := NewSecureString("")
		if !ss.IsEmpty() {
			t.Error("Empty SecureString should report as empty")
		}
		ss.Clear()
	})
}

func TestSecureBuffer(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		sb := NewSecureBuffer(100)

		if sb.Size() != 0 {
			t.Error("New buffer should have size 0")
		}

		if sb.Capacity() != 100 {
			t.Errorf("Capacity() = %d, want 100", sb.Capacity())
		}

		// Write data
		data1 := []byte("hello")
		err := sb.Write(data1)
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		if sb.Size() != len(data1) {
			t.Errorf("Size() = %d, want %d", sb.Size(), len(data1))
		}

		// Write more data
		data2 := []byte(" world")
		err = sb.Write(data2)
		if err != nil {
			t.Fatalf("Second write failed: %v", err)
		}

		expected := append(data1, data2...)
		result := sb.Read()
		if !bytes.Equal(result, expected) {
			t.Errorf("Read() = %q, want %q", string(result), string(expected))
		}

		sb.Clear()
	})

	t.Run("overflow protection", func(t *testing.T) {
		sb := NewSecureBuffer(10)

		// Fill to capacity
		data := make([]byte, 10)
		err := sb.Write(data)
		if err != nil {
			t.Fatalf("Write to capacity failed: %v", err)
		}

		// Try to overflow
		err = sb.Write([]byte{1})
		if err == nil {
			t.Error("Write should fail on buffer overflow")
		}

		sb.Clear()
	})

	t.Run("reset", func(t *testing.T) {
		sb := NewSecureBuffer(50)

		data := []byte("test data")
		err := sb.Write(data)
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		if sb.Size() == 0 {
			t.Error("Buffer should not be empty before reset")
		}

		sb.Reset()

		if sb.Size() != 0 {
			t.Error("Buffer should be empty after reset")
		}

		if sb.Capacity() != 50 {
			t.Error("Capacity should be preserved after reset")
		}

		// Should be able to write again
		err = sb.Write([]byte("new data"))
		if err != nil {
			t.Error("Should be able to write after reset")
		}

		sb.Clear()
	})
}

func TestZeroizeMultiple(t *testing.T) {
	slice1 := []byte{1, 2, 3}
	slice2 := []byte{4, 5, 6}
	slice3 := []byte{7, 8, 9}

	ZeroizeMultiple(slice1, slice2, slice3)

	slices := [][]byte{slice1, slice2, slice3}
	for i, slice := range slices {
		for j, b := range slice {
			if b != 0 {
				t.Errorf("Slice %d, byte %d not zeroed: got %d", i, j, b)
			}
		}
	}
}

func TestSecureCompare(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []byte
		expected bool
	}{
		{"equal slices", []byte{1, 2, 3}, []byte{1, 2, 3}, true},
		{"different content", []byte{1, 2, 3}, []byte{1, 2, 4}, false},
		{"different length", []byte{1, 2, 3}, []byte{1, 2}, false},
		{"both empty", []byte{}, []byte{}, true},
		{"one empty", []byte{1}, []byte{}, false},
		{"both nil", nil, nil, true},
		{"one nil", []byte{1}, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SecureCompare(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("SecureCompare() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSecureRandom(t *testing.T) {
	sizes := []int{0, 1, 16, 32, 1024}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("%d bytes", size), func(t *testing.T) {
			data := make([]byte, size)
			err := SecureRandom(data)
			if err != nil {
				t.Fatalf("SecureRandom failed: %v", err)
			}

			if size == 0 {
				return // Nothing to verify for empty slice
			}

			// For non-zero sizes, verify we got some randomness
			// (Note: there's a tiny chance this could fail with truly random data)
			allZeros := true
			allSame := true
			first := data[0]

			for _, b := range data {
				if b != 0 {
					allZeros = false
				}
				if b != first {
					allSame = false
				}
			}

			if size > 1 {
				if allZeros {
					t.Error("SecureRandom returned all zeros (suspicious)")
				}
				if allSame {
					t.Error("SecureRandom returned all same values (suspicious)")
				}
			}
		})
	}
}

// Test that finalizers work correctly (this is tricky to test reliably)
func TestFinalizers(t *testing.T) {
	t.Run("SecureBytes finalizer", func(t *testing.T) {
		// Create a SecureBytes and let it go out of scope
		func() {
			sb := NewSecureBytes(32)
			// Fill with test data
			data := make([]byte, 32)
			for i := range data {
				data[i] = byte(i + 1)
			}
			sb.CopyTo(data)
			// sb goes out of scope here
		}()

		// Force garbage collection
		runtime.GC()
		runtime.GC()
		time.Sleep(10 * time.Millisecond)

		// We can't easily verify the finalizer ran, but this tests that
		// the finalizer doesn't cause a panic
	})

	t.Run("SecureString finalizer", func(t *testing.T) {
		func() {
			ss := NewSecureString("test data")
			_ = ss.String()
			// ss goes out of scope here
		}()

		runtime.GC()
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("SecureBuffer finalizer", func(t *testing.T) {
		func() {
			sb := NewSecureBuffer(64)
			sb.Write([]byte("test data"))
			// sb goes out of scope here
		}()

		runtime.GC()
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	})
}

func BenchmarkZeroize(b *testing.B) {
	sizes := []int{32, 256, 1024, 4096}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			data := make([]byte, size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Fill with data
				for j := range data {
					data[j] = byte(j)
				}
				// Zeroize
				Zeroize(data)
			}
		})
	}
}

func BenchmarkSecureCompare(b *testing.B) {
	sizes := []int{16, 32, 256, 1024}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			a := make([]byte, size)
			b_slice := make([]byte, size)

			// Fill with identical data
			for i := range a {
				a[i] = byte(i)
				b_slice[i] = byte(i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				SecureCompare(a, b_slice)
			}
		})
	}
}

func BenchmarkSecureBytes(b *testing.B) {
	b.Run("NewSecureBytes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sb := NewSecureBytes(32)
			sb.Clear()
		}
	})

	b.Run("Bytes", func(b *testing.B) {
		sb := NewSecureBytes(32)
		defer sb.Clear()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data := sb.Bytes()
			_ = data
		}
	})

	b.Run("CopyTo", func(b *testing.B) {
		sb := NewSecureBytes(32)
		defer sb.Clear()

		testData := make([]byte, 16)
		for i := range testData {
			testData[i] = byte(i)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sb.CopyTo(testData)
		}
	})
}
