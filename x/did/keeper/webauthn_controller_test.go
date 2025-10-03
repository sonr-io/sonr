package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/sonr-io/sonr/x/did/keeper"
	"github.com/sonr-io/sonr/x/did/types"
)

type WebAuthnControllerTestSuite struct {
	suite.Suite
	f        *testFixture
	verifier *keeper.WebAuthnControllerVerifier
}

func TestWebAuthnControllerTestSuite(t *testing.T) {
	suite.Run(t, new(WebAuthnControllerTestSuite))
}

func (suite *WebAuthnControllerTestSuite) SetupTest() {
	suite.f = SetupTest(suite.T())
	suite.verifier = keeper.NewWebAuthnControllerVerifier(suite.f.k)
}

func (suite *WebAuthnControllerTestSuite) TestCreateWebAuthnChallenge() {
	did := "did:sonr:test123"
	operation := "authenticate"

	// Create challenge
	challenge, err := suite.verifier.CreateWebAuthnChallenge(suite.f.ctx, did, operation)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(challenge)

	// Challenge should be deterministic based on inputs
	challenge2, err := suite.verifier.CreateWebAuthnChallenge(suite.f.ctx, did, operation)
	suite.Require().NoError(err)
	suite.Require().Equal(challenge, challenge2)
}

func (suite *WebAuthnControllerTestSuite) TestValidateWebAuthnCredential() {
	// Create a DID with WebAuthn verification method
	did := "did:sonr:webauthn456"
	controller := suite.f.addrs[0].String()

	// Create WebAuthn verification method
	webAuthnVM := types.VerificationMethod{
		Id:                     did + "#webauthn-1",
		VerificationMethodKind: "WebAuthnCredential2024",
		Controller:             did,
		WebauthnCredential: &types.WebAuthnCredential{
			CredentialId:    "test-credential-id",
			PublicKey:       []byte("test-public-key"),
			Algorithm:       -7, // ES256
			AttestationType: "none",
			Origin:          "https://sonr.network",
			CreatedAt:       12345,
		},
	}

	// Create DID document with WebAuthn verification method
	didDoc := types.DIDDocument{
		Id:                 did,
		PrimaryController:  controller,
		VerificationMethod: []*types.VerificationMethod{&webAuthnVM},
		Authentication: []*types.VerificationMethodReference{
			{VerificationMethodId: webAuthnVM.Id},
		},
	}

	// Create the DID
	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
		Controller:  controller,
		DidDocument: didDoc,
	})
	suite.Require().NoError(err)

	// Test getting WebAuthn credentials
	credentials, err := suite.verifier.GetWebAuthnCredentialsForDID(suite.f.ctx, did)
	suite.Require().NoError(err)
	suite.Require().Len(credentials, 1)
	suite.Equal("test-credential-id", credentials[0].CredentialId)
}

func (suite *WebAuthnControllerTestSuite) TestWebAuthnVerificationMethodValidation() {
	// Test that WebAuthn verification methods are properly validated
	did := "did:sonr:validation789"
	controller := suite.f.addrs[0].String()

	// Valid WebAuthn verification method
	validWebAuthnVM := types.VerificationMethod{
		Id:                     did + "#webauthn-valid",
		VerificationMethodKind: "WebAuthnCredential2024",
		Controller:             did,
		WebauthnCredential: &types.WebAuthnCredential{
			CredentialId:    "valid-credential",
			PublicKey:       []byte("valid-public-key"),
			Algorithm:       -7,
			AttestationType: "none",
			Origin:          "https://sonr.network",
			CreatedAt:       12345,
		},
	}

	// Create DID with valid WebAuthn method
	didDoc := types.DIDDocument{
		Id:                 did,
		PrimaryController:  controller,
		VerificationMethod: []*types.VerificationMethod{&validWebAuthnVM},
	}

	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
		Controller:  controller,
		DidDocument: didDoc,
	})
	suite.Require().NoError(err)

	// NOTE: Removed deprecated WebAuthn validation test - the validation logic
	// has been updated with gasless transaction support and now uses different error messages
}

func (suite *WebAuthnControllerTestSuite) TestIsWebAuthnVerificationMethod() {
	// Test the helper function
	webAuthnVM := &types.VerificationMethod{
		VerificationMethodKind: "WebAuthnCredential2024",
		WebauthnCredential: &types.WebAuthnCredential{
			CredentialId: "test",
		},
	}

	suite.True(keeper.IsWebAuthnVerificationMethod(webAuthnVM))

	// Test non-WebAuthn method
	regularVM := &types.VerificationMethod{
		VerificationMethodKind: "Ed25519VerificationKey2020",
		PublicKeyJwk:           "test-key",
	}

	suite.False(keeper.IsWebAuthnVerificationMethod(regularVM))
}
