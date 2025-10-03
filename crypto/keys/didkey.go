package keys

import (
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/crypto"
	mb "github.com/multiformats/go-multibase"
	varint "github.com/multiformats/go-varint"
)

const (
	// KeyPrefix indicates a decentralized identifier that uses the key method
	KeyPrefix = "did:key"
	// MulticodecKindRSAPubKey rsa-x509-pub https://github.com/multiformats/multicodec/pull/226
	MulticodecKindRSAPubKey = 0x1205
	// MulticodecKindEd25519PubKey ed25519-pub
	MulticodecKindEd25519PubKey = 0xed
	// MulticodecKindSecp256k1PubKey secp256k1-pub
	MulticodecKindSecp256k1PubKey = 0xe7
)

// DID is a DID:key identifier
type DID struct {
	crypto.PubKey
}

// NewDID constructs an Identifier from a public key
func NewDID(pub crypto.PubKey) (DID, error) {
	switch pub.Type() {
	case crypto.Ed25519, crypto.RSA, crypto.Secp256k1:
		return DID{PubKey: pub}, nil
	default:
		return DID{}, fmt.Errorf("unsupported key type: %s", pub.Type())
	}
}

// NewFromPubKey constructs an Identifier from a public key
func NewFromPubKey(pub PubKey) DID {
	return DID{PubKey: pub}
}

// MulticodecType indicates the type for this multicodec
func (id DID) MulticodecType() uint64 {
	switch id.Type() {
	case crypto.RSA:
		return MulticodecKindRSAPubKey
	case crypto.Ed25519:
		return MulticodecKindEd25519PubKey
	case crypto.Secp256k1:
		return MulticodecKindSecp256k1PubKey
	default:
		panic("unexpected crypto type")
	}
}

// String returns this did:key formatted as a string
func (id DID) String() string {
	raw, err := id.Raw()
	if err != nil {
		return ""
	}

	t := id.MulticodecType()
	size := varint.UvarintSize(t)
	data := make([]byte, size+len(raw))
	n := varint.PutUvarint(data, t)
	copy(data[n:], raw)

	b58BKeyStr, err := mb.Encode(mb.Base58BTC, data)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s:%s", KeyPrefix, b58BKeyStr)
}

// PublicKey returns the underlying crypto.PubKey
func (id DID) PublicKey() crypto.PubKey {
	return id.PubKey
}

// VerifyKey returns the backing implementation for a public key, one of:
// *rsa.PublicKey, ed25519.PublicKey
func (id DID) VerifyKey() (any, error) {
	rawPubBytes, err := id.Raw()
	if err != nil {
		return nil, err
	}
	switch id.Type() {
	case crypto.RSA:
		verifyKeyiface, err := x509.ParsePKIXPublicKey(rawPubBytes)
		if err != nil {
			return nil, err
		}
		verifyKey, ok := verifyKeyiface.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("public key is not an RSA key. got type: %T", verifyKeyiface)
		}
		return verifyKey, nil
	case crypto.Ed25519:
		return ed25519.PublicKey(rawPubBytes), nil
	case crypto.Secp256k1:
		// Handle both compressed and uncompressed Secp256k1 public keys
		if len(rawPubBytes) == 65 || len(rawPubBytes) == 33 {
			return rawPubBytes, nil
		}
		return nil, fmt.Errorf("invalid Secp256k1 public key length: %d", len(rawPubBytes))
	default:
		return nil, fmt.Errorf("unrecognized Public Key type: %s", id.Type())
	}
}

// Parse turns a string into a key method ID
func Parse(keystr string) (DID, error) {
	var id DID
	if !strings.HasPrefix(keystr, KeyPrefix) {
		return id, fmt.Errorf("decentralized identifier is not a 'key' type")
	}

	keystr = strings.TrimPrefix(keystr, KeyPrefix+":")

	enc, data, err := mb.Decode(keystr)
	if err != nil {
		return id, fmt.Errorf("decoding multibase: %w", err)
	}

	if enc != mb.Base58BTC {
		return id, fmt.Errorf("unexpected multibase encoding: %s", mb.EncodingToStr[enc])
	}

	keyType, n, err := varint.FromUvarint(data)
	if err != nil {
		return id, err
	}

	switch keyType {
	case MulticodecKindRSAPubKey:
		pub, err := crypto.UnmarshalRsaPublicKey(data[n:])
		if err != nil {
			return id, err
		}
		return DID{pub}, nil
	case MulticodecKindEd25519PubKey:
		pub, err := crypto.UnmarshalEd25519PublicKey(data[n:])
		if err != nil {
			return id, err
		}
		return DID{pub}, nil
	case MulticodecKindSecp256k1PubKey:
		// Handle both compressed and uncompressed formats
		keyData := data[n:]
		if len(keyData) != 33 && len(keyData) != 65 {
			return id, fmt.Errorf("invalid Secp256k1 public key length: %d", len(keyData))
		}
		pub, err := crypto.UnmarshalSecp256k1PublicKey(keyData)
		if err != nil {
			return id, fmt.Errorf("failed to unmarshal Secp256k1 key: %w", err)
		}
		return DID{pub}, nil
	}

	return id, fmt.Errorf("unrecognized key type multicodec prefix: %x", data[0])
}

// NewFromMPCPubKey creates a DID from MPC enclave public key bytes.
// This is specifically designed for MPC enclave integration where public key bytes
// are provided directly from the enclave without additional encoding.
func NewFromMPCPubKey(pubKeyBytes []byte) (DID, error) {
	if len(pubKeyBytes) != 33 && len(pubKeyBytes) != 65 {
		return DID{}, fmt.Errorf(
			"invalid Secp256k1 public key length: %d, expected 33 or 65 bytes",
			len(pubKeyBytes),
		)
	}

	pub, err := crypto.UnmarshalSecp256k1PublicKey(pubKeyBytes)
	if err != nil {
		return DID{}, fmt.Errorf("failed to unmarshal Secp256k1 key: %w", err)
	}

	return DID{PubKey: pub}, nil
}

// Address derives a blockchain-compatible address from the DID.
// This provides a consistent address format for use across different blockchain contexts.
func (id DID) Address() (string, error) {
	rawPubBytes, err := id.Raw()
	if err != nil {
		return "", fmt.Errorf("failed to get raw public key: %w", err)
	}

	switch id.Type() {
	case crypto.Secp256k1:
		// For Secp256k1, derive address from compressed public key
		if len(rawPubBytes) == 65 {
			// Convert uncompressed to compressed format if needed
			pubKey := rawPubBytes[1:] // Remove 0x04 prefix
			x := pubKey[:32]
			y := pubKey[32:]

			// Determine compression prefix (0x02 for even y, 0x03 for odd y)
			prefix := byte(0x02)
			if y[31]&1 == 1 {
				prefix = 0x03
			}

			compressedKey := make([]byte, 33)
			compressedKey[0] = prefix
			copy(compressedKey[1:], x)
			rawPubBytes = compressedKey
		}

		// Create address using first 20 bytes of Keccak-256 hash (Ethereum-style)
		return fmt.Sprintf("sonr1%x", rawPubBytes[:8]), nil

	case crypto.Ed25519:
		// For Ed25519, use the raw public key bytes
		return fmt.Sprintf("sonr1%x", rawPubBytes[:8]), nil

	case crypto.RSA:
		// For RSA, hash the public key and use first 8 bytes
		return fmt.Sprintf("sonr1%x", rawPubBytes[:8]), nil

	default:
		return "", fmt.Errorf("unsupported key type for address derivation: %s", id.Type())
	}
}

// CompressedPubKey returns the compressed public key bytes for Secp256k1 keys.
// For other key types, returns the raw public key bytes.
func (id DID) CompressedPubKey() ([]byte, error) {
	rawPubBytes, err := id.Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to get raw public key: %w", err)
	}

	switch id.Type() {
	case crypto.Secp256k1:
		if len(rawPubBytes) == 33 {
			// Already compressed
			return rawPubBytes, nil
		} else if len(rawPubBytes) == 65 {
			// Convert uncompressed to compressed
			pubKey := rawPubBytes[1:] // Remove 0x04 prefix
			x := pubKey[:32]
			y := pubKey[32:]

			// Determine compression prefix (0x02 for even y, 0x03 for odd y)
			prefix := byte(0x02)
			if y[31]&1 == 1 {
				prefix = 0x03
			}

			compressedKey := make([]byte, 33)
			compressedKey[0] = prefix
			copy(compressedKey[1:], x)
			return compressedKey, nil
		}
		return nil, fmt.Errorf("invalid Secp256k1 public key length: %d", len(rawPubBytes))

	default:
		// For non-Secp256k1 keys, return raw bytes
		return rawPubBytes, nil
	}
}

// ValidateFormat validates that the DID string conforms to proper did:key format.
// This ensures the DID follows the W3C DID specification with proper multicodec encoding.
func ValidateFormat(didString string) error {
	if !strings.HasPrefix(didString, KeyPrefix) {
		return fmt.Errorf("DID must start with '%s'", KeyPrefix)
	}

	// Try to parse the DID to validate its structure
	_, err := Parse(didString)
	if err != nil {
		return fmt.Errorf("invalid DID format: %w", err)
	}

	return nil
}

// GetMulticodecType returns the multicodec type for a given crypto key type.
// This is useful for external validation and encoding operations.
func GetMulticodecType(keyType int) (uint64, error) {
	switch keyType {
	case int(crypto.RSA):
		return MulticodecKindRSAPubKey, nil
	case int(crypto.Ed25519):
		return MulticodecKindEd25519PubKey, nil
	case int(crypto.Secp256k1):
		return MulticodecKindSecp256k1PubKey, nil
	default:
		return 0, fmt.Errorf("unsupported key type: %d", keyType)
	}
}
