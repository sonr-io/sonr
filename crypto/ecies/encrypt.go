package ecies

import eciesgo "github.com/ecies/go/v2"

// Encrypt encrypts a plaintext using a public key
func Encrypt(pub *PublicKey, plaintext []byte) ([]byte, error) {
	return eciesgo.Encrypt(pub, plaintext)
}

// Decrypt decrypts a ciphertext using a private key
func Decrypt(priv *PrivateKey, ciphertext []byte) ([]byte, error) {
	return eciesgo.Decrypt(priv, ciphertext)
}
