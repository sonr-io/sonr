// Package mpc implements the Sonr MPC protocol
package mpc

import (
	"crypto/rand"

	"github.com/sonr-io/sonr/crypto/core/curves"
	"github.com/sonr-io/sonr/crypto/core/protocol"
	"github.com/sonr-io/sonr/crypto/tecdsa/dklsv1/dkg"
)

type CurveName string

const (
	K256Name       CurveName = "secp256k1"
	BLS12381G1Name CurveName = "BLS12381G1"
	BLS12381G2Name CurveName = "BLS12381G2"
	BLS12831Name   CurveName = "BLS12831"
	P256Name       CurveName = "P-256"
	ED25519Name    CurveName = "ed25519"
	PallasName     CurveName = "pallas"
	BLS12377G1Name CurveName = "BLS12377G1"
	BLS12377G2Name CurveName = "BLS12377G2"
	BLS12377Name   CurveName = "BLS12377"
)

func (c CurveName) String() string {
	return string(c)
}

func (c CurveName) Curve() *curves.Curve {
	switch c {
	case K256Name:
		return curves.K256()
	case BLS12381G1Name:
		return curves.BLS12381G1()
	case BLS12381G2Name:
		return curves.BLS12381G2()
	case BLS12831Name:
		return curves.BLS12381G1()
	case P256Name:
		return curves.P256()
	case ED25519Name:
		return curves.ED25519()
	case PallasName:
		return curves.PALLAS()
	case BLS12377G1Name:
		return curves.BLS12377G1()
	case BLS12377G2Name:
		return curves.BLS12377G2()
	case BLS12377Name:
		return curves.BLS12377G1()
	default:
		return curves.K256()
	}
}

// ╭───────────────────────────────────────────────────────────╮
// │                    Exported Generics                      │
// ╰───────────────────────────────────────────────────────────╯

type (
	AliceOut    *dkg.AliceOutput
	BobOut      *dkg.BobOutput
	Point       curves.Point
	Role        string                         // Role is the type for the role
	Message     *protocol.Message              // Message is the protocol.Message that is used for MPC
	Signature   *curves.EcdsaSignature         // Signature is the type for the signature
	RefreshFunc interface{ protocol.Iterator } // RefreshFunc is the type for the refresh function
	SignFunc    interface{ protocol.Iterator } // SignFunc is the type for the sign function
)

const (
	RoleVal  = "validator"
	RoleUser = "user"
)

func randNonce() []byte {
	nonce := make([]byte, 12)
	rand.Read(nonce)
	return nonce
}

// Enclave defines the interface for key management operations
type Enclave interface {
	GetData() *EnclaveData // GetData returns the data of the keyEnclave
	GetEnclave() Enclave   // GetEnclave returns the enclave of the keyEnclave
	Decrypt(
		key []byte,
		encryptedData []byte,
	) ([]byte, error) // Decrypt returns decrypted enclave data
	Encrypt(
		key []byte,
	) ([]byte, error) // Encrypt returns encrypted enclave data
	IsValid() bool             // IsValid returns true if the keyEnclave is valid
	PubKeyBytes() []byte       // PubKeyBytes returns the public key of the keyEnclave
	PubKeyHex() string         // PubKeyHex returns the public key of the keyEnclave
	Refresh() (Enclave, error) // Refresh returns a new keyEnclave
	Marshal() ([]byte, error)  // Serialize returns the serialized keyEnclave
	Sign(
		data []byte,
	) ([]byte, error) // Sign returns the signature of the data
	Unmarshal(
		data []byte,
	) error // Verify returns true if the signature is valid
	Verify(
		data []byte,
		sig []byte,
	) (bool, error) // Verify returns true if the signature is valid
}
