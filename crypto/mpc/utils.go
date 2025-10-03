package mpc

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"math/big"

	"github.com/sonr-io/sonr/crypto/core/curves"
	"github.com/sonr-io/sonr/crypto/core/protocol"
	"github.com/sonr-io/sonr/crypto/tecdsa/dklsv1"
	"golang.org/x/crypto/sha3"
)

func CheckIteratedErrors(aErr, bErr error) error {
	if aErr == protocol.ErrProtocolFinished && bErr == protocol.ErrProtocolFinished {
		return nil
	}
	if aErr != protocol.ErrProtocolFinished {
		return aErr
	}
	if bErr != protocol.ErrProtocolFinished {
		return bErr
	}
	return nil
}

func GetHashKey(key []byte) []byte {
	hash := sha3.New256()
	hash.Write(key)
	return hash.Sum(nil)[:32] // Use first 32 bytes of hash
}

func DecryptKeyshare(msg []byte, key []byte, nonce []byte) ([]byte, error) {
	hashedKey := GetHashKey(key)
	block, err := aes.NewCipher(hashedKey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := aesgcm.Open(nil, nonce, msg, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func EncryptKeyshare(msg Message, key []byte, nonce []byte) ([]byte, error) {
	hashedKey := GetHashKey(key)
	msgBytes, err := protocol.EncodeMessage(msg)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(hashedKey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, []byte(msgBytes), nil)
	return ciphertext, nil
}

func GetAliceOut(msg *protocol.Message) (AliceOut, error) {
	return dklsv1.DecodeAliceDkgResult(msg)
}

func GetAlicePublicPoint(msg *protocol.Message) (Point, error) {
	out, err := dklsv1.DecodeAliceDkgResult(msg)
	if err != nil {
		return nil, err
	}
	return out.PublicKey, nil
}

func GetBobOut(msg *protocol.Message) (BobOut, error) {
	return dklsv1.DecodeBobDkgResult(msg)
}

func GetBobPubPoint(msg *protocol.Message) (Point, error) {
	out, err := dklsv1.DecodeBobDkgResult(msg)
	if err != nil {
		return nil, err
	}
	return out.PublicKey, nil
}

// GetECDSAPoint builds an elliptic curve point from a compressed byte slice
func GetECDSAPoint(pubKey []byte) (*curves.EcPoint, error) {
	crv := curves.K256()
	x := new(big.Int).SetBytes(pubKey[1:33])
	y := new(big.Int).SetBytes(pubKey[33:])
	ecCurve, err := crv.ToEllipticCurve()
	if err != nil {
		return nil, fmt.Errorf("error converting curve: %v", err)
	}
	return &curves.EcPoint{X: x, Y: y, Curve: ecCurve}, nil
}

func SerializeSignature(sig *curves.EcdsaSignature) ([]byte, error) {
	if sig == nil {
		return nil, errors.New("nil signature")
	}

	rBytes := sig.R.Bytes()
	sBytes := sig.S.Bytes()

	// Ensure both components are 32 bytes
	rPadded := make([]byte, 32)
	sPadded := make([]byte, 32)
	copy(rPadded[32-len(rBytes):], rBytes)
	copy(sPadded[32-len(sBytes):], sBytes)

	// Concatenate R and S
	result := make([]byte, 64)
	copy(result[0:32], rPadded)
	copy(result[32:64], sPadded)

	return result, nil
}

func DeserializeSignature(sigBytes []byte) (*curves.EcdsaSignature, error) {
	if len(sigBytes) != 64 {
		return nil, fmt.Errorf("invalid signature length: expected 64 bytes, got %d", len(sigBytes))
	}

	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	return &curves.EcdsaSignature{
		R: r,
		S: s,
	}, nil
}

func GetAliceSignFunc(k *EnclaveData, bz []byte) (SignFunc, error) {
	curve := k.Curve.Curve()
	return dklsv1.NewAliceSign(curve, sha3.New256(), bz, k.ValShare, protocol.Version1)
}

func GetAliceRefreshFunc(k *EnclaveData) (RefreshFunc, error) {
	curve := k.Curve.Curve()
	return dklsv1.NewAliceRefresh(curve, k.ValShare, protocol.Version1)
}

func GetBobSignFunc(k *EnclaveData, bz []byte) (SignFunc, error) {
	curve := curves.K256()
	return dklsv1.NewBobSign(curve, sha3.New256(), bz, k.UserShare, protocol.Version1)
}

func GetBobRefreshFunc(k *EnclaveData) (RefreshFunc, error) {
	curve := curves.K256()
	return dklsv1.NewBobRefresh(curve, k.UserShare, protocol.Version1)
}
