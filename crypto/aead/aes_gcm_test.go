package aead

import (
	"crypto/rand"
	"fmt"
	"testing"
)

func TestNewAESGCM(t *testing.T) {
	tests := []struct {
		name    string
		keySize int
		wantErr bool
	}{
		{"valid 256-bit key", 32, false},
		{"invalid 128-bit key", 16, true},
		{"invalid 192-bit key", 24, true},
		{"empty key", 0, true},
		{"oversized key", 64, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keySize)
			_, err := rand.Read(key)
			if err != nil {
				t.Fatalf("Failed to generate test key: %v", err)
			}

			cipher, err := NewAESGCM(key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAESGCM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cipher == nil {
				t.Error("NewAESGCM() returned nil cipher without error")
			}
		})
	}
}

func TestAESGCMEncryptDecrypt(t *testing.T) {
	// Generate random 256-bit key
	key := make([]byte, KeySize)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	cipher, err := NewAESGCM(key)
	if err != nil {
		t.Fatalf("Failed to create AES-GCM cipher: %v", err)
	}

	tests := []struct {
		name      string
		plaintext []byte
		aad       []byte
	}{
		{"empty plaintext", []byte{}, nil},
		{"small plaintext", []byte("hello"), nil},
		{"large plaintext", make([]byte, 1024), nil},
		{"with AAD", []byte("secret data"), []byte("additional auth data")},
		{"unicode plaintext", []byte("Hello, 世界! 🔐"), []byte("metadata")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill large plaintext with test data
			if len(tt.plaintext) == 1024 {
				for i := range tt.plaintext {
					tt.plaintext[i] = byte(i % 256)
				}
			}

			// Encrypt
			ciphertext, err := cipher.Encrypt(tt.plaintext, tt.aad)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// Verify ciphertext structure
			expectedLen := NonceSize + len(tt.plaintext) + TagSize
			if len(ciphertext) != expectedLen {
				t.Errorf("Unexpected ciphertext length: got %d, want %d", len(ciphertext), expectedLen)
			}

			// Decrypt
			decrypted, err := cipher.Decrypt(ciphertext, tt.aad)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			// Verify plaintext matches
			if string(decrypted) != string(tt.plaintext) {
				t.Errorf("Decrypted text doesn't match original: got %q, want %q", string(decrypted), string(tt.plaintext))
			}
		})
	}
}

func TestAESGCMAuthenticationFailure(t *testing.T) {
	key := make([]byte, KeySize)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	cipher, err := NewAESGCM(key)
	if err != nil {
		t.Fatalf("Failed to create AES-GCM cipher: %v", err)
	}

	plaintext := []byte("authenticated data")
	aad := []byte("additional data")

	ciphertext, err := cipher.Encrypt(plaintext, aad)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	tests := []struct {
		name          string
		modifyFunc    func([]byte) []byte
		modifyAAD     func([]byte) []byte
		expectFailure bool
	}{
		{
			"tampered ciphertext",
			func(data []byte) []byte {
				if len(data) > NonceSize+5 {
					data[NonceSize+5] ^= 0x01 // Flip one bit in ciphertext
				}
				return data
			},
			nil,
			true,
		},
		{
			"tampered nonce",
			func(data []byte) []byte {
				if len(data) > 5 {
					data[5] ^= 0x01 // Flip one bit in nonce
				}
				return data
			},
			nil,
			true,
		},
		{
			"tampered AAD",
			nil,
			func(aad []byte) []byte {
				modified := make([]byte, len(aad))
				copy(modified, aad)
				if len(modified) > 0 {
					modified[0] ^= 0x01
				}
				return modified
			},
			true,
		},
		{
			"valid decryption",
			nil,
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCiphertext := make([]byte, len(ciphertext))
			copy(testCiphertext, ciphertext)
			testAAD := make([]byte, len(aad))
			copy(testAAD, aad)

			if tt.modifyFunc != nil {
				testCiphertext = tt.modifyFunc(testCiphertext)
			}
			if tt.modifyAAD != nil {
				testAAD = tt.modifyAAD(testAAD)
			}

			_, err := cipher.Decrypt(testCiphertext, testAAD)
			if tt.expectFailure && err == nil {
				t.Error("Expected decryption to fail due to tampering, but it succeeded")
			}
			if !tt.expectFailure && err != nil {
				t.Errorf("Expected decryption to succeed, but got error: %v", err)
			}
		})
	}
}

func TestAESGCMWithNonce(t *testing.T) {
	key := make([]byte, KeySize)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	cipher, err := NewAESGCM(key)
	if err != nil {
		t.Fatalf("Failed to create AES-GCM cipher: %v", err)
	}

	plaintext := []byte("test message")
	nonce := make([]byte, NonceSize)
	_, err = rand.Read(nonce)
	if err != nil {
		t.Fatalf("Failed to generate nonce: %v", err)
	}

	// Test encryption with provided nonce
	ciphertext, err := cipher.EncryptWithNonce(plaintext, nil, nonce)
	if err != nil {
		t.Fatalf("EncryptWithNonce() error = %v", err)
	}

	// Verify nonce is properly prepended
	if string(ciphertext[:NonceSize]) != string(nonce) {
		t.Error("Nonce not properly prepended to ciphertext")
	}

	// Test decryption works
	decrypted, err := cipher.Decrypt(ciphertext, nil)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text doesn't match: got %q, want %q", string(decrypted), string(plaintext))
	}

	// Test invalid nonce size
	invalidNonce := make([]byte, 8)
	_, err = cipher.EncryptWithNonce(plaintext, nil, invalidNonce)
	if err == nil {
		t.Error("Expected error for invalid nonce size")
	}
}

func TestAESGCMInvalidCiphertext(t *testing.T) {
	key := make([]byte, KeySize)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	cipher, err := NewAESGCM(key)
	if err != nil {
		t.Fatalf("Failed to create AES-GCM cipher: %v", err)
	}

	tests := []struct {
		name       string
		ciphertext []byte
	}{
		{"empty ciphertext", []byte{}},
		{"too short ciphertext", make([]byte, NonceSize)},
		{"minimal invalid", make([]byte, NonceSize+TagSize-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cipher.Decrypt(tt.ciphertext, nil)
			if err == nil {
				t.Error("Expected error for invalid ciphertext length")
			}
		})
	}
}

func BenchmarkAESGCMEncrypt(b *testing.B) {
	key := make([]byte, KeySize)
	_, err := rand.Read(key)
	if err != nil {
		b.Fatalf("Failed to generate test key: %v", err)
	}

	cipher, err := NewAESGCM(key)
	if err != nil {
		b.Fatalf("Failed to create AES-GCM cipher: %v", err)
	}

	sizes := []int{64, 512, 1024, 4096}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			plaintext := make([]byte, size)
			_, err := rand.Read(plaintext)
			if err != nil {
				b.Fatalf("Failed to generate test data: %v", err)
			}

			b.ResetTimer()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				_, err := cipher.Encrypt(plaintext, nil)
				if err != nil {
					b.Fatalf("Encrypt error: %v", err)
				}
			}
		})
	}
}

func BenchmarkAESGCMDecrypt(b *testing.B) {
	key := make([]byte, KeySize)
	_, err := rand.Read(key)
	if err != nil {
		b.Fatalf("Failed to generate test key: %v", err)
	}

	cipher, err := NewAESGCM(key)
	if err != nil {
		b.Fatalf("Failed to create AES-GCM cipher: %v", err)
	}

	sizes := []int{64, 512, 1024, 4096}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			plaintext := make([]byte, size)
			_, err := rand.Read(plaintext)
			if err != nil {
				b.Fatalf("Failed to generate test data: %v", err)
			}

			ciphertext, err := cipher.Encrypt(plaintext, nil)
			if err != nil {
				b.Fatalf("Failed to encrypt test data: %v", err)
			}

			b.ResetTimer()
			b.SetBytes(int64(size))

			for i := 0; i < b.N; i++ {
				_, err := cipher.Decrypt(ciphertext, nil)
				if err != nil {
					b.Fatalf("Decrypt error: %v", err)
				}
			}
		})
	}
}
