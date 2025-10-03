package subtle

import (
	"crypto/rand"

	"golang.org/x/crypto/curve25519"
)

// GeneratePrivateKeyX25519 generates a new 32-byte private key.
func GeneratePrivateKeyX25519() ([]byte, error) {
	privKey := make([]byte, curve25519.ScalarSize)
	_, err := rand.Read(privKey)
	return privKey, err
}

// ComputeSharedSecretX25519 returns the 32-byte shared key, i.e.
// privKey * pubValue on the curve.
func ComputeSharedSecretX25519(privKey, pubValue []byte) ([]byte, error) {
	return curve25519.X25519(privKey, pubValue)
}

// PublicFromPrivateX25519 computes privKey's corresponding public key.
func PublicFromPrivateX25519(privKey []byte) ([]byte, error) {
	return ComputeSharedSecretX25519(privKey, curve25519.Basepoint)
}
