package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/x/did/types"
)

func TestDIDDocumentConversions(t *testing.T) {
	// Create a test DID document
	doc := &types.DIDDocument{
		Id:                "did:example:123",
		PrimaryController: "controller123",
		AlsoKnownAs:       []string{"alias1", "alias2"},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     "did:example:123#key-1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				Controller:             "did:example:123",
				PublicKeyJwk:           `{"kty":"OKP"}`,
			},
		},
		Authentication: []*types.VerificationMethodReference{
			{VerificationMethodId: "did:example:123#key-1"},
		},
		Service: []*types.Service{
			{
				Id:             "did:example:123#service-1",
				ServiceKind:    "LinkedDomains",
				SingleEndpoint: "https://example.com",
			},
		},
		CreatedAt:   12345,
		UpdatedAt:   12346,
		Deactivated: false,
		Version:     1,
	}

	// Convert to ORM
	ormDoc := doc.ToORM()
	require.NotNil(t, ormDoc)
	require.Equal(t, doc.Id, ormDoc.Id)
	require.Equal(t, doc.PrimaryController, ormDoc.PrimaryController)
	require.Equal(t, doc.AlsoKnownAs, ormDoc.AlsoKnownAs)
	require.Len(t, ormDoc.VerificationMethod, 1)
	require.Len(t, ormDoc.Authentication, 1)
	require.Len(t, ormDoc.Service, 1)

	// Convert back from ORM
	convertedDoc := types.DIDDocumentFromORM(ormDoc)
	require.NotNil(t, convertedDoc)
	require.Equal(t, doc.Id, convertedDoc.Id)
	require.Equal(t, doc.PrimaryController, convertedDoc.PrimaryController)
	require.Equal(t, doc.AlsoKnownAs, convertedDoc.AlsoKnownAs)
	require.Len(t, convertedDoc.VerificationMethod, 1)
	require.Equal(t, doc.VerificationMethod[0].Id, convertedDoc.VerificationMethod[0].Id)
}

func TestVerifiableCredentialConversions(t *testing.T) {
	// Create a test credential
	vc := &types.VerifiableCredential{
		Id:                "https://example.com/credentials/123",
		Context:           []string{"https://www.w3.org/2018/credentials/v1"},
		CredentialKinds:   []string{"VerifiableCredential"},
		Issuer:            "did:example:issuer",
		Subject:           "did:example:subject",
		IssuanceDate:      "2024-01-01T00:00:00Z",
		ExpirationDate:    "2025-01-01T00:00:00Z",
		CredentialSubject: []byte(`{"name":"John Doe"}`),
		Proof: []*types.CredentialProof{
			{
				ProofKind:          "Ed25519Signature2020",
				Created:            "2024-01-01T00:00:00Z",
				VerificationMethod: "did:example:issuer#key-1",
				ProofPurpose:       "assertionMethod",
				Signature:          "signature123",
			},
		},
	}

	// Convert to ORM
	ormVC := vc.ToORM()
	require.NotNil(t, ormVC)
	require.Equal(t, vc.Id, ormVC.Id)
	require.Equal(t, vc.Issuer, ormVC.Issuer)
	require.Equal(t, vc.Subject, ormVC.Subject)
	require.Len(t, ormVC.Proof, 1)

	// Convert back from ORM
	convertedVC := types.VerifiableCredentialFromORM(ormVC)
	require.NotNil(t, convertedVC)
	require.Equal(t, vc.Id, convertedVC.Id)
	require.Equal(t, vc.Issuer, convertedVC.Issuer)
	require.Equal(t, vc.Subject, convertedVC.Subject)
	require.Equal(t, vc.CredentialSubject, convertedVC.CredentialSubject)
}
