package mpc

import (
	"crypto/ecdsa"

	"golang.org/x/crypto/sha3"
)

func VerifyWithPubKey(pubKeyCompressed []byte, data []byte, sig []byte) (bool, error) {
	edSig, err := DeserializeSignature(sig)
	if err != nil {
		return false, err
	}
	ePub, err := GetECDSAPoint(pubKeyCompressed)
	if err != nil {
		return false, err
	}
	pk := &ecdsa.PublicKey{
		Curve: ePub.Curve,
		X:     ePub.X,
		Y:     ePub.Y,
	}

	// Hash the message using SHA3-256
	hash := sha3.New256()
	hash.Write(data)
	digest := hash.Sum(nil)
	return ecdsa.Verify(pk, digest, edSig.R, edSig.S), nil
}
