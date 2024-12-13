package keys

import (
	"crypto/ecdsa"
	"encoding/hex"

	"github.com/onsonr/sonr/crypto/core/curves"
)

type PubKey interface {
	Bytes() []byte
	Type() string
	Hex() string
	Verify(msg []byte, sig []byte) (bool, error)
}

type pubKey struct {
	publicPoint curves.Point
	method      string
}

func NewPubKey(pk curves.Point) PubKey {
	return &pubKey{
		publicPoint: pk,
	}
}

func (p pubKey) Bytes() []byte {
	return p.publicPoint.ToAffineCompressed()
}

func (p pubKey) Hex() string {
	return hex.EncodeToString(p.publicPoint.ToAffineCompressed())
}

func (p pubKey) Type() string {
	return "secp256k1"
}

func (p pubKey) Verify(msgBz []byte, sigBz []byte) (bool, error) {
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
	return ecdsa.Verify(pk, msgBz, sig.R, sig.S), nil
}