package mpc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"

	"github.com/sonr-io/sonr/crypto/core/curves"
	"golang.org/x/crypto/sha3"
)

// EnclaveData implements the Enclave interface
type EnclaveData struct {
	PubHex    string    `json:"pub_hex"`   // PubHex is the hex-encoded compressed public key
	PubBytes  []byte    `json:"pub_bytes"` // PubBytes is the uncompressed public key
	ValShare  Message   `json:"val_share"`
	UserShare Message   `json:"user_share"`
	Nonce     []byte    `json:"nonce"`
	Curve     CurveName `json:"curve"`
}

// GetData returns the data of the keyEnclave
func (k *EnclaveData) GetData() *EnclaveData {
	return k
}

// GetEnclave returns the enclave of the keyEnclave
func (k *EnclaveData) GetEnclave() Enclave {
	return k
}

// GetPubPoint returns the public point of the keyEnclave
func (k *EnclaveData) GetPubPoint() (curves.Point, error) {
	curve := k.Curve.Curve()
	return curve.NewIdentityPoint().FromAffineUncompressed(k.PubBytes)
}

// PubKeyHex returns the public key of the keyEnclave
func (k *EnclaveData) PubKeyHex() string {
	return k.PubHex
}

// PubKeyBytes returns the public key of the keyEnclave
func (k *EnclaveData) PubKeyBytes() []byte {
	return k.PubBytes
}

// Decrypt returns decrypted enclave data
func (k *EnclaveData) Decrypt(key []byte, encryptedData []byte) ([]byte, error) {
	hashedKey := GetHashKey(key)
	block, err := aes.NewCipher(hashedKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt the data using AES-GCM
	plaintext, err := aesgcm.Open(nil, k.Nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}
	return plaintext, nil
}

// Encrypt returns encrypted enclave data
func (k *EnclaveData) Encrypt(key []byte) ([]byte, error) {
	data, err := k.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize enclave: %w", err)
	}

	hashedKey := GetHashKey(key)
	block, err := aes.NewCipher(hashedKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm.Seal(nil, k.Nonce, data, nil), nil
}

// IsValid returns true if the keyEnclave is valid
func (k *EnclaveData) IsValid() bool {
	return k.ValShare != nil && k.UserShare != nil
}

// Refresh returns a new keyEnclave
func (k *EnclaveData) Refresh() (Enclave, error) {
	refreshFuncVal, err := GetAliceRefreshFunc(k)
	if err != nil {
		return nil, err
	}
	refreshFuncUser, err := GetBobRefreshFunc(k)
	if err != nil {
		return nil, err
	}
	return ExecuteRefresh(refreshFuncVal, refreshFuncUser, k.Curve)
}

// Sign returns the signature of the data
func (k *EnclaveData) Sign(data []byte) ([]byte, error) {
	userSign, err := GetBobSignFunc(k, data)
	if err != nil {
		return nil, err
	}
	valSign, err := GetAliceSignFunc(k, data)
	if err != nil {
		return nil, err
	}
	return ExecuteSigning(valSign, userSign)
}

// Verify returns true if the signature is valid
func (k *EnclaveData) Verify(data []byte, sig []byte) (bool, error) {
	edSig, err := DeserializeSignature(sig)
	if err != nil {
		return false, err
	}
	ePub, err := GetECDSAPoint(k.PubBytes)
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

// Marshal returns the JSON encoding of keyEnclave
func (k *EnclaveData) Marshal() ([]byte, error) {
	return json.Marshal(k)
}

// Unmarshal unmarshals the JSON encoding of keyEnclave
func (k *EnclaveData) Unmarshal(data []byte) error {
	if err := json.Unmarshal(data, k); err != nil {
		return err
	}
	return nil
}
