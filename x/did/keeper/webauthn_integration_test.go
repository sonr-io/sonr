package keeper_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/suite"

	"github.com/sonr-io/sonr/x/did/keeper"
)

// WebAuthnIntegrationTestSuite tests end-to-end WebAuthn flows
type WebAuthnIntegrationTestSuite struct {
	suite.Suite
	f *testFixture
}

func TestWebAuthnIntegrationSuite(t *testing.T) {
	suite.Run(t, new(WebAuthnIntegrationTestSuite))
}

func (suite *WebAuthnIntegrationTestSuite) SetupTest() {
	suite.f = SetupTest(suite.T())
}

// TestCompleteRegistrationFlow tests the full WebAuthn registration process
func (suite *WebAuthnIntegrationTestSuite) TestCompleteRegistrationFlow() {
	// Test data
	username := "alice"
	credentialID := "test-credential-123"

	// Create valid attestation object (simplified for testing)
	attestationObj := createTestAttestationObject(credentialID)
	clientDataJSON := createTestClientDataJSON("test-challenge", "http://localhost:8080")

	// Extract public key for registration (normally done by VerifyWebAuthnRegistration)
	coseKey := map[int]any{
		1:  2,                // kty: EC2
		3:  -7,               // alg: ES256
		-1: 1,                // crv: P-256
		-2: make([]byte, 32), // x coordinate
		-3: make([]byte, 32), // y coordinate
	}
	publicKeyCOSE, _ := cbor.Marshal(coseKey)

	regData := &keeper.WebAuthnRegistrationData{
		CredentialID:      base64.RawURLEncoding.EncodeToString([]byte(credentialID)),
		RawID:             base64.RawURLEncoding.EncodeToString([]byte(credentialID)),
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString(clientDataJSON),
		AttestationObject: base64.RawURLEncoding.EncodeToString(attestationObj),
		Username:          username,
		PublicKey:         publicKeyCOSE,
		Algorithm:         -7, // ES256
	}

	// Process registration
	didDoc, err := suite.f.k.ProcessWebAuthnRegistration(suite.f.ctx, regData)
	suite.Require().NoError(err, "registration should succeed")
	suite.Require().NotNil(didDoc)

	// Verify DID document was created
	suite.Require().Contains(didDoc.Id, "did:sonr:")
	suite.Require().Len(didDoc.VerificationMethod, 1)

	// Verify WebAuthn credential was stored
	vm := didDoc.VerificationMethod[0]
	suite.Require().NotNil(vm.WebauthnCredential)
	suite.Require().
		Equal(base64.RawURLEncoding.EncodeToString([]byte(credentialID)), vm.WebauthnCredential.CredentialId)
}

// TestCredentialIDUniqueness tests that duplicate credential IDs are rejected
func (suite *WebAuthnIntegrationTestSuite) TestCredentialIDUniqueness() {
	credentialID := "unique-credential-456"

	// First registration
	regData1 := createTestRegistrationData("user1", credentialID)
	didDoc1, err := suite.f.k.ProcessWebAuthnRegistration(suite.f.ctx, regData1)
	suite.Require().NoError(err, "first registration should succeed")
	suite.Require().NotNil(didDoc1)

	// Attempt duplicate registration
	regData2 := createTestRegistrationData("user2", credentialID)
	_, err = suite.f.k.ProcessWebAuthnRegistration(suite.f.ctx, regData2)
	suite.Require().Error(err, "duplicate credential ID should be rejected")
	suite.Require().Contains(err.Error(), "already exists")
}

// TestMultiAlgorithmSupport tests different signature algorithms
func (suite *WebAuthnIntegrationTestSuite) TestMultiAlgorithmSupport() {
	testCases := []struct {
		name      string
		algorithm int32
		keySize   int
	}{
		{"ES256", -7, 64},    // ECDSA P-256
		{"RS256", -257, 256}, // RSA
		// Note: EdDSA (-8) is not currently supported by ValidateAlgorithmSupport
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			credentialID := fmt.Sprintf("algo-test-%s", tc.name)
			username := fmt.Sprintf("user-%s", tc.name)
			regData := createTestRegistrationDataWithAlgorithm(username, credentialID, tc.algorithm)

			didDoc, err := suite.f.k.ProcessWebAuthnRegistration(suite.f.ctx, regData)
			suite.Require().NoError(err, "registration with %s should succeed", tc.name)
			suite.Require().NotNil(didDoc)

			vm := didDoc.VerificationMethod[0]
			suite.Require().Equal(tc.algorithm, vm.WebauthnCredential.Algorithm)
		})
	}
}

// TestOriginValidation tests that only allowed origins are accepted
func (suite *WebAuthnIntegrationTestSuite) TestOriginValidation() {
	testCases := []struct {
		name        string
		origin      string
		shouldError bool
	}{
		{"valid localhost", "http://localhost:8080", false},
		{"valid localhost alt port", "http://localhost:8081", false},
		{"invalid origin", "http://evil.com", true},
		{"invalid protocol", "ftp://localhost:8080", true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			challenge := "test-challenge"
			credentialID := fmt.Sprintf("origin-test-%s", tc.name)

			clientData := createTestClientDataJSON(challenge, tc.origin)
			attestationObj := createTestAttestationObject(credentialID)

			// Create a valid COSE public key for ES256
			coseKey := map[int]any{
				1:  2,                // kty: EC2
				3:  -7,               // alg: ES256
				-1: 1,                // crv: P-256
				-2: make([]byte, 32), // x coordinate
				-3: make([]byte, 32), // y coordinate
			}
			publicKey, _ := cbor.Marshal(coseKey)

			regData := &keeper.WebAuthnRegistrationData{
				CredentialID:      base64.RawURLEncoding.EncodeToString([]byte(credentialID)),
				RawID:             base64.RawURLEncoding.EncodeToString([]byte(credentialID)),
				ClientDataJSON:    base64.RawURLEncoding.EncodeToString(clientData),
				AttestationObject: base64.RawURLEncoding.EncodeToString(attestationObj),
				Username:          "testuser",
				PublicKey:         publicKey,
				Algorithm:         -7, // ES256
				Origin:            tc.origin,
			}

			err := suite.f.k.VerifyWebAuthnRegistration(suite.f.ctx, regData, challenge)

			if tc.shouldError {
				suite.Require().Error(err, "origin %s should be rejected", tc.origin)
			} else {
				suite.Require().NoError(err, "origin %s should be accepted", tc.origin)
			}
		})
	}
}

// TestChallengeVerification tests challenge validation
func (suite *WebAuthnIntegrationTestSuite) TestChallengeVerification() {
	credentialID := "challenge-test-789"
	correctChallenge := "correct-challenge"
	wrongChallenge := "wrong-challenge"

	// Create registration data with correct challenge
	clientData := createTestClientDataJSON(correctChallenge, "http://localhost:8080")
	attestationObj := createTestAttestationObject(credentialID)

	// Create a valid COSE public key for ES256
	coseKey := map[int]any{
		1:  2,                // kty: EC2
		3:  -7,               // alg: ES256
		-1: 1,                // crv: P-256
		-2: make([]byte, 32), // x coordinate
		-3: make([]byte, 32), // y coordinate
	}
	publicKey, _ := cbor.Marshal(coseKey)

	regData := &keeper.WebAuthnRegistrationData{
		CredentialID:      base64.RawURLEncoding.EncodeToString([]byte(credentialID)),
		RawID:             base64.RawURLEncoding.EncodeToString([]byte(credentialID)),
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString(clientData),
		AttestationObject: base64.RawURLEncoding.EncodeToString(attestationObj),
		Username:          "testuser",
		PublicKey:         publicKey,
		Algorithm:         -7, // ES256
		Origin:            "http://localhost:8080",
	}

	// Verify with correct challenge
	err := suite.f.k.VerifyWebAuthnRegistration(suite.f.ctx, regData, correctChallenge)
	suite.Require().NoError(err, "correct challenge should pass")

	// Verify with wrong challenge
	err = suite.f.k.VerifyWebAuthnRegistration(suite.f.ctx, regData, wrongChallenge)
	suite.Require().Error(err, "wrong challenge should fail")
	suite.Require().Contains(err.Error(), "challenge mismatch")
}

// TestDIDDocumentStorage tests that DID documents are properly stored
func (suite *WebAuthnIntegrationTestSuite) TestDIDDocumentStorage() {
	username := "bob"
	credentialID := "storage-test-abc"

	regData := createTestRegistrationData(username, credentialID)
	didDoc, err := suite.f.k.ProcessWebAuthnRegistration(suite.f.ctx, regData)
	suite.Require().NoError(err)
	suite.Require().NotNil(didDoc)

	// Verify we can retrieve the stored DID document
	credentials, err := suite.f.k.GetWebAuthnCredentialsByDID(suite.f.ctx, didDoc.Id)
	suite.Require().NoError(err)
	suite.Require().Len(credentials, 1)
	suite.Require().
		Equal(base64.RawURLEncoding.EncodeToString([]byte(credentialID)), credentials[0].CredentialId)
}

// TestInvalidAttestationHandling tests rejection of invalid attestation data
func (suite *WebAuthnIntegrationTestSuite) TestInvalidAttestationHandling() {
	testCases := []struct {
		name              string
		attestationObject string
		clientDataJSON    string
		expectedError     string
	}{
		{
			"empty attestation",
			"",
			base64.RawURLEncoding.EncodeToString(
				[]byte(
					`{"type":"webauthn.create","challenge":"test","origin":"http://localhost:8080"}`,
				),
			),
			"attestation_object is required",
		},
		{
			"invalid base64",
			"not-base64!@#$",
			base64.RawURLEncoding.EncodeToString(
				[]byte(
					`{"type":"webauthn.create","challenge":"test","origin":"http://localhost:8080"}`,
				),
			),
			"illegal base64 data",
		},
		{
			"empty client data",
			base64.RawURLEncoding.EncodeToString(createTestAttestationObject("test")),
			"",
			"failed to parse client data: unexpected end of JSON input",
		},
		{
			"invalid client data JSON",
			base64.RawURLEncoding.EncodeToString(createTestAttestationObject("test")),
			base64.RawURLEncoding.EncodeToString([]byte("not json")),
			"failed to decode client data JSON: illegal base64 data",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create a valid public key for the test
			coseKey := map[int]any{
				1:  2,                // kty: EC2
				3:  -7,               // alg: ES256
				-1: 1,                // crv: P-256
				-2: make([]byte, 32), // x coordinate
				-3: make([]byte, 32), // y coordinate
			}
			publicKey, _ := cbor.Marshal(coseKey)

			regData := &keeper.WebAuthnRegistrationData{
				CredentialID:      "test",
				RawID:             base64.RawURLEncoding.EncodeToString([]byte("test")),
				ClientDataJSON:    tc.clientDataJSON,
				AttestationObject: tc.attestationObject,
				Username:          "testuser",
				PublicKey:         publicKey,
				Algorithm:         -7,
				Origin:            "http://localhost:8080",
			}

			err := suite.f.k.VerifyWebAuthnRegistration(suite.f.ctx, regData, "test")
			suite.Require().Error(err)
			suite.Require().Contains(err.Error(), tc.expectedError)
		})
	}
}

// Helper functions

func createTestRegistrationData(username, credentialID string) *keeper.WebAuthnRegistrationData {
	return createTestRegistrationDataWithAlgorithm(username, credentialID, -7) // ES256
}

func createTestRegistrationDataWithAlgorithm(
	username, credentialID string,
	algorithm int32,
) *keeper.WebAuthnRegistrationData {
	attestationObj := createTestAttestationObject(credentialID)
	clientDataJSON := createTestClientDataJSON("test-challenge", "http://localhost:8080")

	// Create COSE public key based on algorithm
	var publicKey []byte
	switch algorithm {
	case -7: // ES256
		coseKey := map[int]any{
			1:  2,                // kty: EC2
			3:  -7,               // alg: ES256
			-1: 1,                // crv: P-256
			-2: make([]byte, 32), // x coordinate
			-3: make([]byte, 32), // y coordinate
		}
		publicKey, _ = cbor.Marshal(coseKey)
	case -257: // RS256
		coseKey := map[int]any{
			1:  3,                 // kty: RSA
			3:  -257,              // alg: RS256
			-1: make([]byte, 256), // n (modulus)
			-2: []byte{1, 0, 1},   // e (exponent = 65537)
		}
		publicKey, _ = cbor.Marshal(coseKey)
	case -8: // EdDSA
		coseKey := map[int]any{
			1:  1,                // kty: OKP
			3:  -8,               // alg: EdDSA
			-1: 6,                // crv: Ed25519
			-2: make([]byte, 32), // x coordinate
		}
		publicKey, _ = cbor.Marshal(coseKey)
	default: // Default to ES256
		coseKey := map[int]any{
			1:  2,                // kty: EC2
			3:  -7,               // alg: ES256
			-1: 1,                // crv: P-256
			-2: make([]byte, 32), // x coordinate
			-3: make([]byte, 32), // y coordinate
		}
		publicKey, _ = cbor.Marshal(coseKey)
		algorithm = -7
	}

	return &keeper.WebAuthnRegistrationData{
		CredentialID:      base64.RawURLEncoding.EncodeToString([]byte(credentialID)),
		RawID:             base64.RawURLEncoding.EncodeToString([]byte(credentialID)),
		ClientDataJSON:    base64.RawURLEncoding.EncodeToString(clientDataJSON),
		AttestationObject: base64.RawURLEncoding.EncodeToString(attestationObj),
		Username:          username,
		PublicKey:         publicKey,
		Algorithm:         algorithm,
		Origin:            "http://localhost:8080",
	}
}

func createTestClientDataJSON(challenge, origin string) []byte {
	// Create client data that matches WebAuthn format
	clientData := map[string]any{
		"type":        "webauthn.create",
		"challenge":   challenge, // Keep challenge as-is, will be base64 encoded by caller
		"origin":      origin,
		"crossOrigin": false,
	}
	data, _ := json.Marshal(clientData)
	return data
}

func createTestAttestationObject(credentialID string) []byte {
	// Create a proper CBOR attestation object with valid structure

	// Create COSE public key for ES256
	coseKey := map[int]any{
		1:  2,                // kty: EC2
		3:  -7,               // alg: ES256
		-1: 1,                // crv: P-256
		-2: make([]byte, 32), // x coordinate (dummy)
		-3: make([]byte, 32), // y coordinate (dummy)
	}
	publicKeyCOSE, _ := cbor.Marshal(coseKey)

	// Create authenticator data
	authData := createValidAuthenticatorData([]byte(credentialID), publicKeyCOSE)

	// Create attestation object
	attestationObj := map[string]any{
		"fmt":      "none",
		"attStmt":  map[string]any{},
		"authData": authData,
	}

	attestationObjCBOR, _ := cbor.Marshal(attestationObj)
	return attestationObjCBOR
}

func createValidAuthenticatorData(credentialID, publicKey []byte) []byte {
	// RP ID hash (32 bytes) - SHA256 of "localhost"
	rpIDHash := sha256.Sum256([]byte("localhost"))

	// Flags byte: UP=1, UV=1, AT=1 (0x45)
	flags := byte(0x45)

	// Sign count (4 bytes)
	signCount := make([]byte, 4)
	binary.BigEndian.PutUint32(signCount, 0)

	// Build authenticator data
	authData := make([]byte, 0)
	authData = append(authData, rpIDHash[:]...)
	authData = append(authData, flags)
	authData = append(authData, signCount...)

	// Add attested credential data (since AT flag is set)
	// AAGUID (16 bytes) - all zeros for testing
	aaguid := make([]byte, 16)
	authData = append(authData, aaguid...)

	// Credential ID length (2 bytes)
	credIDLen := make([]byte, 2)
	binary.BigEndian.PutUint16(credIDLen, uint16(len(credentialID)))
	authData = append(authData, credIDLen...)

	// Credential ID
	authData = append(authData, credentialID...)

	// Public key
	authData = append(authData, publicKey...)

	return authData
}
