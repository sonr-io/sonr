package keeper_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/x/did/types"
)

func TestVerifyDIDDocumentSignature(t *testing.T) {
	f := SetupTest(t)

	// Generate Ed25519 key pair for testing
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Create test DID document
	did := "did:sonr:test123"
	didDoc := &types.DIDDocument{
		Id:                did,
		PrimaryController: "did:sonr:controller123",
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     did + "#key1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				Controller:             did,
				PublicKeyBase64:        base64.StdEncoding.EncodeToString(publicKey),
			},
		},
		Deactivated: false,
		Version:     1,
		CreatedAt:   1234567890,
		UpdatedAt:   1234567890,
	}

	// Store the DID document
	ormDoc := didDoc.ToORM()
	err = f.k.OrmDB.DIDDocumentTable().Insert(f.ctx, ormDoc)
	require.NoError(t, err)

	// Test signature verification
	testCases := []struct {
		name           string
		did            string
		signature      []byte
		expectedResult bool
		expectedError  bool
	}{
		{
			name:           "Valid signature",
			did:            did,
			signature:      createEd25519Signature(privateKey, []byte("test message")),
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "Invalid signature",
			did:            did,
			signature:      []byte("invalid signature"),
			expectedResult: false,
			expectedError:  true,
		},
		{
			name:           "Non-existent DID",
			did:            "did:sonr:nonexistent",
			signature:      createEd25519Signature(privateKey, []byte("test message")),
			expectedResult: false,
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := f.k.VerifyDIDDocumentSignature(f.ctx, tc.did, tc.signature)

			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestVerifyDIDDocumentSignature_DeactivatedDID(t *testing.T) {
	f := SetupTest(t)

	// Generate Ed25519 key pair for testing
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Create test DID document that is deactivated
	did := "did:sonr:deactivated123"
	didDoc := &types.DIDDocument{
		Id:                did,
		PrimaryController: "did:sonr:controller123",
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     did + "#key1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				Controller:             did,
				PublicKeyBase64:        base64.StdEncoding.EncodeToString(publicKey),
			},
		},
		Deactivated: true, // This is deactivated
		Version:     1,
		CreatedAt:   1234567890,
		UpdatedAt:   1234567890,
	}

	// Store the DID document
	ormDoc := didDoc.ToORM()
	err = f.k.OrmDB.DIDDocumentTable().Insert(f.ctx, ormDoc)
	require.NoError(t, err)

	// Test signature verification should fail for deactivated DID
	result, err := f.k.VerifyDIDDocumentSignature(f.ctx, did, []byte("any signature"))
	require.Error(t, err)
	require.False(t, result)
	require.Contains(t, err.Error(), "deactivated")
}

func TestVerifyDIDDocumentSignature_MultipleVerificationMethods(t *testing.T) {
	f := SetupTest(t)

	// Generate Ed25519 key pairs for testing
	publicKey1, privateKey1, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	publicKey2, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Create test DID document with multiple verification methods
	did := "did:sonr:multi123"
	didDoc := &types.DIDDocument{
		Id:                did,
		PrimaryController: "did:sonr:controller123",
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     did + "#key1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				Controller:             did,
				PublicKeyBase64:        base64.StdEncoding.EncodeToString(publicKey1),
			},
			{
				Id:                     did + "#key2",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				Controller:             did,
				PublicKeyHex:           hex.EncodeToString(publicKey2),
			},
		},
		Deactivated: false,
		Version:     1,
		CreatedAt:   1234567890,
		UpdatedAt:   1234567890,
	}

	// Store the DID document
	ormDoc := didDoc.ToORM()
	err = f.k.OrmDB.DIDDocumentTable().Insert(f.ctx, ormDoc)
	require.NoError(t, err)

	// Test signature verification with first key should succeed
	signature1 := createEd25519Signature(privateKey1, []byte("test message"))
	result, err := f.k.VerifyDIDDocumentSignature(f.ctx, did, signature1)
	require.NoError(t, err)
	require.True(t, result)
}

func TestVerifyDIDDocumentSignature_UnsupportedVerificationMethod(t *testing.T) {
	f := SetupTest(t)

	// Create test DID document with unsupported verification method
	did := "did:sonr:unsupported123"
	didDoc := &types.DIDDocument{
		Id:                did,
		PrimaryController: "did:sonr:controller123",
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     did + "#key1",
				VerificationMethodKind: "UnsupportedMethod2020",
				Controller:             did,
				PublicKeyBase64:        "dummy-key",
			},
		},
		Deactivated: false,
		Version:     1,
		CreatedAt:   1234567890,
		UpdatedAt:   1234567890,
	}

	// Store the DID document
	ormDoc := didDoc.ToORM()
	err := f.k.OrmDB.DIDDocumentTable().Insert(f.ctx, ormDoc)
	require.NoError(t, err)

	// Test signature verification should fail for unsupported method
	result, err := f.k.VerifyDIDDocumentSignature(f.ctx, did, []byte("any signature"))
	require.Error(t, err)
	require.False(t, result)
	require.Contains(t, err.Error(), "signature verification failed")
}

// TestVerifyDIDDocumentSignature_WebAuthnVerificationMethod - REMOVED
// This test was testing deprecated WebAuthn signature verification functionality
// that has been replaced with the gasless transaction approach.

func TestVerifyDIDDocumentSignature_JsonWebSignature2020(t *testing.T) {
	f := SetupTest(t)

	// Create test DID document with JWS verification method
	did := "did:sonr:jws123"
	didDoc := &types.DIDDocument{
		Id:                did,
		PrimaryController: "did:sonr:controller123",
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     did + "#jws1",
				VerificationMethodKind: "JsonWebSignature2020",
				Controller:             did,
				PublicKeyJwk:           `{"kty":"OKP","crv":"Ed25519","x":"dummy-key"}`,
			},
		},
		Deactivated: false,
		Version:     1,
		CreatedAt:   1234567890,
		UpdatedAt:   1234567890,
	}

	// Store the DID document
	ormDoc := didDoc.ToORM()
	err := f.k.OrmDB.DIDDocumentTable().Insert(f.ctx, ormDoc)
	require.NoError(t, err)

	// Test signature verification with JWS method
	// Note: This will fail since we don't have a real JWS signature
	jwsSignature := `{"signature":"dummy-signature","protected":"dummy-protected","header":{}}`
	result, err := f.k.VerifyDIDDocumentSignature(f.ctx, did, []byte(jwsSignature))
	require.Error(t, err)
	require.False(t, result)
}

// Helper function to create Ed25519 signature
func createEd25519Signature(privateKey ed25519.PrivateKey, message []byte) []byte {
	return ed25519.Sign(privateKey, message)
}
