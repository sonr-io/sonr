package keys

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"

	p2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	p2ppb "github.com/libp2p/go-libp2p/core/crypto/pb"
	"github.com/sonr-io/sonr/crypto/core/curves"
	"golang.org/x/crypto/sha3"
)

type PubKey interface {
	Bytes() []byte
	Raw() ([]byte, error)
	Equals(b p2pcrypto.Key) bool
	Type() p2ppb.KeyType
	Hex() string
	Verify(msg []byte, sig []byte) (bool, error)
}

type pubKey struct {
	publicPoint curves.Point
}

func NewPubKey(pk curves.Point) PubKey {
	return &pubKey{
		publicPoint: pk,
	}
}

func (p pubKey) Bytes() []byte {
	return p.publicPoint.ToAffineCompressed()
}

func (p pubKey) Raw() ([]byte, error) {
	return p.publicPoint.ToAffineCompressed(), nil
}

func (p pubKey) Equals(b p2pcrypto.Key) bool {
	if b == nil {
		return false
	}
	apbz, err := b.Raw()
	if err != nil {
		return false
	}
	bbz, err := p.Raw()
	if err != nil {
		return false
	}
	return bytes.Equal(apbz, bbz)
}

func (p pubKey) Hex() string {
	return hex.EncodeToString(p.publicPoint.ToAffineCompressed())
}

func (p pubKey) Type() p2ppb.KeyType {
	return p2ppb.KeyType_Secp256k1
}

func (p pubKey) Verify(data []byte, sigBz []byte) (bool, error) {
	sig, err := deserializeSignature(sigBz)
	if err != nil {
		return false, err
	}
	pp, err := getEcdsaPoint(p.Bytes())
	if err != nil {
		return false, err
	}
	pk := &ecdsa.PublicKey{
		Curve: pp.Curve,
		X:     pp.X,
		Y:     pp.Y,
	}

	// Hash the message using SHA3-256
	hash := sha3.New256()
	hash.Write(data)
	digest := hash.Sum(nil)

	return ecdsa.Verify(pk, digest, sig.R, sig.S), nil
}
