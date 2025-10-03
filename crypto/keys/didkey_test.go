package keys

import (
	"crypto/rand"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// generateSecp256k1Key generates a test Secp256k1 key pair
func generateSecp256k1Key(t *testing.T) crypto.PrivKey {
	privKey, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate Secp256k1 key: %v", err)
	}
	return privKey
}

// TestNewFromMPCPubKey tests creating DIDs from MPC public key bytes
func TestNewFromMPCPubKey(t *testing.T) {
	privKey := generateSecp256k1Key(t)
	pubKey := privKey.GetPublic()

	// Get raw public key bytes
	pubKeyBytes, err := pubKey.Raw()
	if err != nil {
		t.Fatalf("Failed to get raw public key: %v", err)
	}

	// Test with compressed key (33 bytes)
	if len(pubKeyBytes) == 33 {
		did, err := NewFromMPCPubKey(pubKeyBytes)
		if err != nil {
			t.Errorf("NewFromMPCPubKey failed with compressed key: %v", err)
		}

		if did.Type() != crypto.Secp256k1 {
			t.Errorf("Expected Secp256k1 key type, got %v", did.Type())
		}
	}

	// Test with invalid key lengths
	invalidKeys := [][]byte{
		make([]byte, 32), // Too short
		make([]byte, 34), // Wrong length
		make([]byte, 66), // Too long
	}

	for _, invalidKey := range invalidKeys {
		_, err := NewFromMPCPubKey(invalidKey)
		if err == nil {
			t.Errorf("Expected error with invalid key length %d, got nil", len(invalidKey))
		}
	}
}

// TestSecp256k1MulticodecFix tests that the multicodec value is correct
func TestSecp256k1MulticodecFix(t *testing.T) {
	if MulticodecKindSecp256k1PubKey != 0xe7 {
		t.Errorf(
			"Expected Secp256k1 multicodec to be 0xe7, got 0x%x",
			MulticodecKindSecp256k1PubKey,
		)
	}

	privKey := generateSecp256k1Key(t)
	pubKey := privKey.GetPublic()

	did := DID{PubKey: pubKey}
	multicodecType := did.MulticodecType()

	if multicodecType != 0xe7 {
		t.Errorf("Expected multicodec type 0xe7, got 0x%x", multicodecType)
	}
}

// TestAddress tests blockchain-compatible address derivation
func TestAddress(t *testing.T) {
	privKey := generateSecp256k1Key(t)
	pubKey := privKey.GetPublic()

	did := DID{PubKey: pubKey}

	address, err := did.Address()
	if err != nil {
		t.Fatalf("Failed to derive address: %v", err)
	}

	// Check that address starts with "sonr1"
	if len(address) < 5 || address[:5] != "sonr1" {
		t.Errorf("Expected address to start with 'sonr1', got %s", address)
	}

	// Check that address is deterministic
	address2, err := did.Address()
	if err != nil {
		t.Fatalf("Failed to derive address second time: %v", err)
	}

	if address != address2 {
		t.Errorf("Address derivation not deterministic: %s != %s", address, address2)
	}
}

// TestCompressedPubKey tests public key compression
func TestCompressedPubKey(t *testing.T) {
	privKey := generateSecp256k1Key(t)
	pubKey := privKey.GetPublic()

	did := DID{PubKey: pubKey}

	compressed, err := did.CompressedPubKey()
	if err != nil {
		t.Fatalf("Failed to get compressed public key: %v", err)
	}

	// Check that compressed key is 33 bytes for Secp256k1
	if len(compressed) != 33 {
		t.Errorf("Expected compressed key length 33, got %d", len(compressed))
	}

	// Check that compression prefix is valid (0x02 or 0x03)
	if compressed[0] != 0x02 && compressed[0] != 0x03 {
		t.Errorf("Invalid compression prefix: 0x%02x", compressed[0])
	}
}

// TestValidateFormat tests DID string format validation
func TestValidateFormat(t *testing.T) {
	privKey := generateSecp256k1Key(t)
	pubKey := privKey.GetPublic()

	did := DID{PubKey: pubKey}
	didString := did.String()

	// Valid DID should pass validation
	err := ValidateFormat(didString)
	if err != nil {
		t.Errorf("Valid DID failed validation: %v", err)
	}

	// Invalid DIDs should fail validation
	invalidDIDs := []string{
		"invalid:key:z123", // Wrong method
		"did:invalid:z123", // Wrong key method
		"did:key:invalid",  // Invalid encoding
		"not-a-did-at-all", // Not a DID
		"",                 // Empty string
	}

	for _, invalidDID := range invalidDIDs {
		err := ValidateFormat(invalidDID)
		if err == nil {
			t.Errorf("Invalid DID '%s' passed validation", invalidDID)
		}
	}
}

// TestGetMulticodecType tests multicodec type lookup
func TestGetMulticodecType(t *testing.T) {
	testCases := []struct {
		keyType  int
		expected uint64
	}{
		{int(crypto.RSA), MulticodecKindRSAPubKey},
		{int(crypto.Ed25519), MulticodecKindEd25519PubKey},
		{int(crypto.Secp256k1), MulticodecKindSecp256k1PubKey},
	}

	for _, tc := range testCases {
		result, err := GetMulticodecType(tc.keyType)
		if err != nil {
			t.Errorf("GetMulticodecType failed for type %d: %v", tc.keyType, err)
		}
		if result != tc.expected {
			t.Errorf(
				"GetMulticodecType for type %d: expected 0x%x, got 0x%x",
				tc.keyType,
				tc.expected,
				result,
			)
		}
	}

	// Test invalid key type
	_, err := GetMulticodecType(999)
	if err == nil {
		t.Errorf("GetMulticodecType should fail for invalid key type")
	}
}

// TestDIDStringFormat tests complete DID string generation and parsing
func TestDIDStringFormat(t *testing.T) {
	privKey := generateSecp256k1Key(t)
	pubKey := privKey.GetPublic()

	did := DID{PubKey: pubKey}
	didString := did.String()

	// Check that DID starts with "did:key:z"
	if len(didString) < 9 || didString[:9] != "did:key:z" {
		t.Errorf("DID string should start with 'did:key:z', got %s", didString)
	}

	// Parse the DID back
	parsedDID, err := Parse(didString)
	if err != nil {
		t.Fatalf("Failed to parse generated DID: %v", err)
	}

	// Verify parsed DID generates same string
	parsedString := parsedDID.String()
	if parsedString != didString {
		t.Errorf("Parsed DID string mismatch: %s != %s", parsedString, didString)
	}

	// Verify key types match
	if parsedDID.Type() != did.Type() {
		t.Errorf("Parsed DID key type mismatch: %v != %v", parsedDID.Type(), did.Type())
	}
}

// TestMPCIntegration tests end-to-end MPC public key integration
func TestMPCIntegration(t *testing.T) {
	// Generate a valid MPC enclave public key for testing
	privKey := generateSecp256k1Key(t)
	pubKey := privKey.GetPublic()
	mpcPubKeyBytes, err := pubKey.Raw()
	if err != nil {
		t.Fatalf("Failed to get raw public key: %v", err)
	}

	// Create DID from MPC public key
	did, err := NewFromMPCPubKey(mpcPubKeyBytes)
	if err != nil {
		t.Fatalf("Failed to create DID from MPC public key: %v", err)
	}

	// Test DID string generation
	didString := did.String()
	if len(didString) == 0 {
		t.Error("Generated DID string is empty")
	}

	// Test address derivation
	address, err := did.Address()
	if err != nil {
		t.Fatalf("Failed to derive address: %v", err)
	}

	if len(address) == 0 || address[:5] != "sonr1" {
		t.Errorf("Invalid address format: %s", address)
	}

	// Test compressed public key
	compressed, err := did.CompressedPubKey()
	if err != nil {
		t.Fatalf("Failed to get compressed key: %v", err)
	}

	if len(compressed) != 33 {
		t.Errorf("Expected 33-byte compressed key, got %d bytes", len(compressed))
	}

	// Verify round-trip: DID -> string -> parse -> DID
	parsedDID, err := Parse(didString)
	if err != nil {
		t.Fatalf("Failed to parse generated DID: %v", err)
	}

	parsedAddress, err := parsedDID.Address()
	if err != nil {
		t.Fatalf("Failed to derive address from parsed DID: %v", err)
	}

	if address != parsedAddress {
		t.Errorf("Address mismatch after round-trip: %s != %s", address, parsedAddress)
	}
}
