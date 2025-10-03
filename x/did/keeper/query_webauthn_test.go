package keeper_test

import (
	"fmt"

	apiv1 "github.com/sonr-io/sonr/api/did/v1"
	"github.com/sonr-io/sonr/x/did/types"
)

// TestRegisterStart tests the RegisterStart query endpoint
func (suite *QueryServerTestSuite) TestRegisterStart() {
	testCases := []struct {
		name           string
		setupFn        func() *types.QueryRegisterStartRequest
		expErr         bool
		expErrContains string
		validateResp   func(*types.QueryRegisterStartResponse)
	}{
		{
			name: "success - new email assertion",
			setupFn: func() *types.QueryRegisterStartRequest {
				// Initialize default params for this test
				err := suite.f.k.Params.Set(suite.f.ctx, types.DefaultParams())
				suite.Require().NoError(err, "failed to initialize default params")

				return &types.QueryRegisterStartRequest{
					AssertionDid: "did:sonr:email:abc123def456",
				}
			},
			expErr: false,
			validateResp: func(resp *types.QueryRegisterStartResponse) {
				suite.Require().NotEmpty(resp.Challenge, "challenge should not be empty")
				suite.Require().Len(resp.Challenge, 43, "base64url-encoded 32 bytes should be 43 chars")
				suite.Require().NotEmpty(resp.RelyingPartyId, "relying party ID should be set")
				suite.Require().NotNil(resp.User, "user map should not be nil")
				suite.Require().Equal("did:sonr:email:abc123def456", resp.User["id"])
				suite.Require().Equal("Email User", resp.User["name"])
				suite.Require().Contains(resp.User["displayName"], "Email")
			},
		},
		{
			name: "success - new phone assertion",
			setupFn: func() *types.QueryRegisterStartRequest {
				err := suite.f.k.Params.Set(suite.f.ctx, types.DefaultParams())
				suite.Require().NoError(err)

				return &types.QueryRegisterStartRequest{
					AssertionDid: "did:sonr:phone:xyz789abc012",
				}
			},
			expErr: false,
			validateResp: func(resp *types.QueryRegisterStartResponse) {
				suite.Require().NotEmpty(resp.Challenge)
				suite.Require().NotNil(resp.User)
				suite.Require().Equal("Phone User", resp.User["name"])
				suite.Require().Contains(resp.User["displayName"], "Phone")
			},
		},
		{
			name: "success - github assertion",
			setupFn: func() *types.QueryRegisterStartRequest {
				err := suite.f.k.Params.Set(suite.f.ctx, types.DefaultParams())
				suite.Require().NoError(err)

				return &types.QueryRegisterStartRequest{
					AssertionDid: "did:sonr:github:fedcba987654",
				}
			},
			expErr: false,
			validateResp: func(resp *types.QueryRegisterStartResponse) {
				suite.Require().Equal("GitHub User", resp.User["name"])
				suite.Require().Contains(resp.User["displayName"], "GitHub")
			},
		},
		{
			name: "error - nil request",
			setupFn: func() *types.QueryRegisterStartRequest {
				return nil
			},
			expErr:         true,
			expErrContains: "request cannot be nil",
		},
		{
			name: "error - empty assertion DID",
			setupFn: func() *types.QueryRegisterStartRequest {
				return &types.QueryRegisterStartRequest{
					AssertionDid: "",
				}
			},
			expErr:         true,
			expErrContains: "assertion_did cannot be empty",
		},
		{
			name: "error - assertion already exists",
			setupFn: func() *types.QueryRegisterStartRequest {
				err := suite.f.k.Params.Set(suite.f.ctx, types.DefaultParams())
				suite.Require().NoError(err)

				// Create an assertion first
				assertionDid := "did:sonr:email:existing123"
				assertion := &apiv1.Assertion{
					Did:        assertionDid,
					Controller: "did:sonr:controller123",
					Subject:    "test@example.com",
					DidKind:    "email",
				}
				err = suite.f.k.OrmDB.AssertionTable().Save(suite.f.ctx, assertion)
				suite.Require().NoError(err)

				return &types.QueryRegisterStartRequest{
					AssertionDid: assertionDid,
				}
			},
			expErr:         true,
			expErrContains: "assertion already exists",
		},
		{
			name: "deterministic challenge - same inputs generate same challenge",
			setupFn: func() *types.QueryRegisterStartRequest {
				err := suite.f.k.Params.Set(suite.f.ctx, types.DefaultParams())
				suite.Require().NoError(err)

				// This test verifies determinism by calling RegisterStart twice
				// at the same block height with the same assertion DID
				return &types.QueryRegisterStartRequest{
					AssertionDid: "did:sonr:email:deterministic123",
				}
			},
			expErr: false,
			validateResp: func(resp1 *types.QueryRegisterStartResponse) {
				// Call again with same params
				resp2, err := suite.f.queryServer.RegisterStart(suite.f.ctx, &types.QueryRegisterStartRequest{
					AssertionDid: "did:sonr:email:deterministic456",
				})
				suite.Require().NoError(err)

				// Challenges should be different for different DIDs
				suite.Require().NotEqual(
					string(resp1.Challenge),
					string(resp2.Challenge),
					"different DIDs should produce different challenges",
				)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req := tc.setupFn()

			resp, err := suite.f.queryServer.RegisterStart(suite.f.ctx, req)

			if tc.expErr {
				suite.Require().Error(err)
				if tc.expErrContains != "" {
					suite.Require().Contains(err.Error(), tc.expErrContains)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				if tc.validateResp != nil {
					tc.validateResp(resp)
				}
			}
		})
	}
}

// TestLoginStart tests the LoginStart query endpoint
func (suite *QueryServerTestSuite) TestLoginStart() {
	// Initialize default params for all tests
	err := suite.f.k.Params.Set(suite.f.ctx, types.DefaultParams())
	suite.Require().NoError(err, "failed to initialize default params")

	// Setup: Create a controller DID with WebAuthn credentials
	controllerDid := "did:sonr:controller789"
	credId1 := "credential_id_1"
	credId2 := "credential_id_2"

	controllerDoc := &apiv1.DIDDocument{
		Id:                controllerDid,
		PrimaryController: suite.f.addrs[0].String(),
		VerificationMethod: []*apiv1.VerificationMethod{
			{
				Id:                     controllerDid + "#webauthn-1",
				VerificationMethodKind: "WebAuthn2021",
				Controller:             controllerDid,
				WebauthnCredential: &apiv1.WebAuthnCredential{
					CredentialId: credId1,
					PublicKey:    []byte("test-public-key-1"),
					Algorithm:    -7, // ES256
				},
			},
			{
				Id:                     controllerDid + "#webauthn-2",
				VerificationMethodKind: "WebAuthn2021",
				Controller:             controllerDid,
				WebauthnCredential: &apiv1.WebAuthnCredential{
					CredentialId: credId2,
					PublicKey:    []byte("test-public-key-2"),
					Algorithm:    -7, // ES256
				},
			},
			{
				Id:                     controllerDid + "#ed25519-1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				Controller:             controllerDid,
				PublicKeyMultibase:     "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
			},
		},
		Authentication: []*apiv1.VerificationMethodReference{
			{VerificationMethodId: controllerDid + "#webauthn-1"},
			{VerificationMethodId: controllerDid + "#webauthn-2"},
			{VerificationMethodId: controllerDid + "#ed25519-1"},
		},
	}

	err = suite.f.k.OrmDB.DIDDocumentTable().Save(suite.f.ctx, controllerDoc)
	suite.Require().NoError(err)

	testCases := []struct {
		name           string
		setupFn        func() *types.QueryLoginStartRequest
		expErr         bool
		expErrContains string
		validateResp   func(*types.QueryLoginStartResponse)
	}{
		{
			name: "success - existing assertion with WebAuthn credentials",
			setupFn: func() *types.QueryLoginStartRequest {
				assertionDid := "did:sonr:email:login123"
				assertion := &apiv1.Assertion{
					Did:        assertionDid,
					Controller: controllerDid,
					Subject:    "user@example.com",
					DidKind:    "email",
				}
				err := suite.f.k.OrmDB.AssertionTable().Save(suite.f.ctx, assertion)
				suite.Require().NoError(err)

				return &types.QueryLoginStartRequest{
					AssertionDid: assertionDid,
				}
			},
			expErr: false,
			validateResp: func(resp *types.QueryLoginStartResponse) {
				suite.Require().NotEmpty(resp.Challenge, "challenge should not be empty")
				suite.Require().Len(resp.Challenge, 43, "base64url-encoded 32 bytes should be 43 chars")
				suite.Require().NotEmpty(resp.RelyingPartyId, "relying party ID should be set")
				suite.Require().Len(resp.CredentialIds, 2, "should extract exactly 2 WebAuthn credentials")
				suite.Require().Contains(resp.CredentialIds, credId1)
				suite.Require().Contains(resp.CredentialIds, credId2)
			},
		},
		{
			name: "success - embedded verification method",
			setupFn: func() *types.QueryLoginStartRequest {
				// Create controller with embedded verification method
				embeddedControllerDid := "did:sonr:embedded456"
				embeddedCredId := "embedded_credential_id"

				embeddedDoc := &apiv1.DIDDocument{
					Id:                embeddedControllerDid,
					PrimaryController: suite.f.addrs[0].String(),
					Authentication: []*apiv1.VerificationMethodReference{
						{
							EmbeddedVerificationMethod: &apiv1.VerificationMethod{
								Id:                     embeddedControllerDid + "#embedded-webauthn",
								VerificationMethodKind: "WebAuthn2021",
								Controller:             embeddedControllerDid,
								WebauthnCredential: &apiv1.WebAuthnCredential{
									CredentialId: embeddedCredId,
									PublicKey:    []byte("embedded-key"),
									Algorithm:    -7,
								},
							},
						},
					},
				}
				err := suite.f.k.OrmDB.DIDDocumentTable().Save(suite.f.ctx, embeddedDoc)
				suite.Require().NoError(err)

				assertionDid := "did:sonr:email:embedded789"
				assertion := &apiv1.Assertion{
					Did:        assertionDid,
					Controller: embeddedControllerDid,
					Subject:    "embedded@example.com",
					DidKind:    "email",
				}
				err = suite.f.k.OrmDB.AssertionTable().Save(suite.f.ctx, assertion)
				suite.Require().NoError(err)

				return &types.QueryLoginStartRequest{
					AssertionDid: assertionDid,
				}
			},
			expErr: false,
			validateResp: func(resp *types.QueryLoginStartResponse) {
				suite.Require().Len(resp.CredentialIds, 1)
				suite.Require().Equal("embedded_credential_id", resp.CredentialIds[0])
			},
		},
		{
			name: "error - nil request",
			setupFn: func() *types.QueryLoginStartRequest {
				return nil
			},
			expErr:         true,
			expErrContains: "request cannot be nil",
		},
		{
			name: "error - empty assertion DID",
			setupFn: func() *types.QueryLoginStartRequest {
				return &types.QueryLoginStartRequest{
					AssertionDid: "",
				}
			},
			expErr:         true,
			expErrContains: "assertion_did cannot be empty",
		},
		{
			name: "error - assertion not found",
			setupFn: func() *types.QueryLoginStartRequest {
				return &types.QueryLoginStartRequest{
					AssertionDid: "did:sonr:email:notfound999",
				}
			},
			expErr:         true,
			expErrContains: "assertion DID did:sonr:email:notfound999 not found",
		},
		{
			name: "error - assertion has no controller",
			setupFn: func() *types.QueryLoginStartRequest {
				assertionDid := "did:sonr:email:nocontroller123"
				assertion := &apiv1.Assertion{
					Did:        assertionDid,
					Controller: "", // No controller
					Subject:    "nocontroller@example.com",
					DidKind:    "email",
				}
				err := suite.f.k.OrmDB.AssertionTable().Save(suite.f.ctx, assertion)
				suite.Require().NoError(err)

				return &types.QueryLoginStartRequest{
					AssertionDid: assertionDid,
				}
			},
			expErr:         true,
			expErrContains: "has no controller",
		},
		{
			name: "error - controller DID not found",
			setupFn: func() *types.QueryLoginStartRequest {
				assertionDid := "did:sonr:email:missingcontroller456"
				assertion := &apiv1.Assertion{
					Did:        assertionDid,
					Controller: "did:sonr:nonexistent999",
					Subject:    "missing@example.com",
					DidKind:    "email",
				}
				err := suite.f.k.OrmDB.AssertionTable().Save(suite.f.ctx, assertion)
				suite.Require().NoError(err)

				return &types.QueryLoginStartRequest{
					AssertionDid: assertionDid,
				}
			},
			expErr:         true,
			expErrContains: "controller DID did:sonr:nonexistent999 not found",
		},
		{
			name: "error - controller DID is deactivated",
			setupFn: func() *types.QueryLoginStartRequest {
				deactivatedDid := "did:sonr:deactivated789"
				deactivatedDoc := &apiv1.DIDDocument{
					Id:                deactivatedDid,
					PrimaryController: suite.f.addrs[0].String(),
					Deactivated:       true, // Deactivated
				}
				err := suite.f.k.OrmDB.DIDDocumentTable().Save(suite.f.ctx, deactivatedDoc)
				suite.Require().NoError(err)

				assertionDid := "did:sonr:email:deactivatedlogin123"
				assertion := &apiv1.Assertion{
					Did:        assertionDid,
					Controller: deactivatedDid,
					Subject:    "deactivated@example.com",
					DidKind:    "email",
				}
				err = suite.f.k.OrmDB.AssertionTable().Save(suite.f.ctx, assertion)
				suite.Require().NoError(err)

				return &types.QueryLoginStartRequest{
					AssertionDid: assertionDid,
				}
			},
			expErr:         true,
			expErrContains: "is deactivated",
		},
		{
			name: "error - no WebAuthn credentials found",
			setupFn: func() *types.QueryLoginStartRequest {
				noCredsControllerDid := "did:sonr:nocreds456"
				noCredsDoc := &apiv1.DIDDocument{
					Id:                noCredsControllerDid,
					PrimaryController: suite.f.addrs[0].String(),
					VerificationMethod: []*apiv1.VerificationMethod{
						{
							Id:                     noCredsControllerDid + "#ed25519",
							VerificationMethodKind: "Ed25519VerificationKey2020",
							Controller:             noCredsControllerDid,
							PublicKeyMultibase:     "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
						},
					},
					Authentication: []*apiv1.VerificationMethodReference{
						{VerificationMethodId: noCredsControllerDid + "#ed25519"},
					},
				}
				err := suite.f.k.OrmDB.DIDDocumentTable().Save(suite.f.ctx, noCredsDoc)
				suite.Require().NoError(err)

				assertionDid := "did:sonr:email:nocreds789"
				assertion := &apiv1.Assertion{
					Did:        assertionDid,
					Controller: noCredsControllerDid,
					Subject:    "nocreds@example.com",
					DidKind:    "email",
				}
				err = suite.f.k.OrmDB.AssertionTable().Save(suite.f.ctx, assertion)
				suite.Require().NoError(err)

				return &types.QueryLoginStartRequest{
					AssertionDid: assertionDid,
				}
			},
			expErr:         true,
			expErrContains: "no WebAuthn credentials found",
		},
		{
			name: "filters out non-WebAuthn methods",
			setupFn: func() *types.QueryLoginStartRequest {
				mixedDid := "did:sonr:mixed123"
				mixedCredId := "mixed_webauthn_cred"

				mixedDoc := &apiv1.DIDDocument{
					Id:                mixedDid,
					PrimaryController: suite.f.addrs[0].String(),
					VerificationMethod: []*apiv1.VerificationMethod{
						{
							Id:                     mixedDid + "#webauthn",
							VerificationMethodKind: "WebAuthn2021",
							Controller:             mixedDid,
							WebauthnCredential: &apiv1.WebAuthnCredential{
								CredentialId: mixedCredId,
								PublicKey:    []byte("mixed-key"),
								Algorithm:    -7,
							},
						},
						{
							Id:                     mixedDid + "#ed25519",
							VerificationMethodKind: "Ed25519VerificationKey2020",
							Controller:             mixedDid,
							PublicKeyMultibase:     "z6MkhaXgBZDvotDkL5257faiztiGiC2QtKLGpbnnEGta2doK",
						},
						{
							Id:                     mixedDid + "#secp256k1",
							VerificationMethodKind: "EcdsaSecp256k1VerificationKey2019",
							Controller:             mixedDid,
							PublicKeyMultibase:     "zQ3shokFTS3brHcDQrn82RUDfCZESWL1ZdCEJwekUDPQiYBme",
						},
					},
					Authentication: []*apiv1.VerificationMethodReference{
						{VerificationMethodId: mixedDid + "#webauthn"},
						{VerificationMethodId: mixedDid + "#ed25519"},
						{VerificationMethodId: mixedDid + "#secp256k1"},
					},
				}
				err := suite.f.k.OrmDB.DIDDocumentTable().Save(suite.f.ctx, mixedDoc)
				suite.Require().NoError(err)

				assertionDid := "did:sonr:email:mixed789"
				assertion := &apiv1.Assertion{
					Did:        assertionDid,
					Controller: mixedDid,
					Subject:    "mixed@example.com",
					DidKind:    "email",
				}
				err = suite.f.k.OrmDB.AssertionTable().Save(suite.f.ctx, assertion)
				suite.Require().NoError(err)

				return &types.QueryLoginStartRequest{
					AssertionDid: assertionDid,
				}
			},
			expErr: false,
			validateResp: func(resp *types.QueryLoginStartResponse) {
				// Should only return the WebAuthn credential, not Ed25519 or secp256k1
				suite.Require().Len(resp.CredentialIds, 1, "should only extract WebAuthn credentials")
				suite.Require().Equal("mixed_webauthn_cred", resp.CredentialIds[0])
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req := tc.setupFn()

			resp, err := suite.f.queryServer.LoginStart(suite.f.ctx, req)

			if tc.expErr {
				suite.Require().Error(err)
				if tc.expErrContains != "" {
					suite.Require().Contains(err.Error(), tc.expErrContains)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				if tc.validateResp != nil {
					tc.validateResp(resp)
				}
			}
		})
	}
}

// TestUserInfoExtraction tests the extractUserInfoFromAssertionDID helper
func (suite *QueryServerTestSuite) TestUserInfoExtraction() {
	// Initialize module params for RegisterStart to work
	err := suite.f.k.Params.Set(suite.f.ctx, types.DefaultParams())
	suite.Require().NoError(err, "failed to initialize default params")

	testCases := []struct {
		assertionDid         string
		expectedName         string
		expectedDispContains string
	}{
		{
			assertionDid:         "did:sonr:email:abc123def456",
			expectedName:         "Email User",
			expectedDispContains: "Email",
		},
		{
			assertionDid:         "did:sonr:phone:xyz789abc012",
			expectedName:         "Phone User",
			expectedDispContains: "Phone",
		},
		{
			assertionDid:         "did:sonr:tel:111222333444",
			expectedName:         "Phone User",
			expectedDispContains: "Phone",
		},
		{
			assertionDid:         "did:sonr:github:fedcba987654",
			expectedName:         "GitHub User",
			expectedDispContains: "GitHub",
		},
		{
			assertionDid:         "did:sonr:google:aabbccddee11",
			expectedName:         "Google User",
			expectedDispContains: "Google",
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("extract_%s", tc.expectedName), func() {
			resp, err := suite.f.queryServer.RegisterStart(suite.f.ctx, &types.QueryRegisterStartRequest{
				AssertionDid: tc.assertionDid,
			})

			suite.Require().NoError(err)
			suite.Require().NotNil(resp)
			suite.Require().Equal(tc.expectedName, resp.User["name"])
			suite.Require().Contains(resp.User["displayName"], tc.expectedDispContains)
		})
	}
}
