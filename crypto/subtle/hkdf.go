package subtle

import (
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

const (
	// Minimum tag size in bytes. This provides minimum 80-bit security strength.
	minTagSizeInBytes = uint32(10)
)

// validateHKDFParams validates parameters of HKDF constructor.
func validateHKDFParams(hash string, _ uint32, tagSize uint32) error {
	// validate tag size
	digestSize, err := GetHashDigestSize(hash)
	if err != nil {
		return err
	}
	if tagSize > 255*digestSize {
		return fmt.Errorf("tag size too big")
	}
	if tagSize < minTagSizeInBytes {
		return fmt.Errorf("tag size too small")
	}
	return nil
}

// ComputeHKDF extracts a pseudorandom key.
func ComputeHKDF(
	hashAlg string,
	key []byte,
	salt []byte,
	info []byte,
	tagSize uint32,
) ([]byte, error) {
	keySize := uint32(len(key))
	if err := validateHKDFParams(hashAlg, keySize, tagSize); err != nil {
		return nil, fmt.Errorf("hkdf: %s", err)
	}
	hashFunc := GetHashFunc(hashAlg)
	if hashFunc == nil {
		return nil, fmt.Errorf("hkdf: invalid hash algorithm")
	}
	if len(salt) == 0 {
		salt = make([]byte, hashFunc().Size())
	}

	result := make([]byte, tagSize)
	kdf := hkdf.New(hashFunc, key, salt, info)
	n, err := io.ReadFull(kdf, result)
	if n != len(result) || err != nil {
		return nil, fmt.Errorf("compute of hkdf failed")
	}
	return result, nil
}
