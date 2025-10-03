package ecies

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"

	eciesgo "github.com/ecies/go/v2"
	"lukechampine.com/blake3"

	"github.com/sonr-io/sonr/crypto/core/curves"
)

type PrivateKey = eciesgo.PrivateKey

type PublicKey = eciesgo.PublicKey

// GenerateKey generates secp256k1 key pair
func GenerateKey() (*PrivateKey, error) {
	curve := curves.SP256()
	p, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("cannot generate key pair: %w", err)
	}
	return &PrivateKey{
		PublicKey: &PublicKey{
			Curve: curve,
			X:     p.X,
			Y:     p.Y,
		},
		D: p.D,
	}, nil
}

// GenerateKeyFromSeed generates secp256k1 key pair from []byte seed
func GenerateKeyFromSeed(seed []byte) (*PrivateKey, error) {
	curve := curves.SP256()
	p, err := ecdsa.GenerateKey(curve, bytes.NewReader(seed[:]))
	if err != nil {
		return nil, fmt.Errorf("cannot generate key pair: %w", err)
	}
	return &PrivateKey{
		PublicKey: &PublicKey{
			Curve: curve,
			X:     p.X,
			Y:     p.Y,
		},
		D: p.D,
	}, nil
}

// HashSeed returns 512 sum hash of byte slice
func HashSeed(seed []byte) []byte {
	bz := blake3.Sum512(seed)
	return bz[:]
}
