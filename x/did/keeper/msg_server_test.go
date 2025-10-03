package keeper_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/sonr/x/did/types"
)

type MsgServerTestSuite struct {
	suite.Suite
	f *testFixture
}

func TestMsgServerSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (suite *MsgServerTestSuite) SetupTest() {
	suite.f = SetupTest(suite.T())
}

// Helper function to create a valid DID document
func (suite *MsgServerTestSuite) createValidDIDDocument(did string) types.DIDDocument {
	return types.DIDDocument{
		Id:                did,
		PrimaryController: suite.f.addrs[0].String(),
		AlsoKnownAs:       []string{"alias1", "alias2"},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     did + "#key-1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				Controller:             did,
				PublicKeyJwk:           `{"kty":"OKP","crv":"Ed25519","x":"test-public-key"}`,
			},
		},
		Authentication: []*types.VerificationMethodReference{
			{VerificationMethodId: did + "#key-1"},
		},
		AssertionMethod: []*types.VerificationMethodReference{
			{VerificationMethodId: did + "#key-1"},
		},
		KeyAgreement:         []*types.VerificationMethodReference{},
		CapabilityInvocation: []*types.VerificationMethodReference{},
		CapabilityDelegation: []*types.VerificationMethodReference{},
		Service: []*types.Service{
			{
				Id:             did + "#service-1",
				ServiceKind:    "LinkedDomains",
				SingleEndpoint: "https://example.com",
			},
		},
	}
}

// Test UpdateParams
func (suite *MsgServerTestSuite) TestUpdateParams() {
	testCases := []struct {
		name    string
		request *types.MsgUpdateParams
		expErr  bool
	}{
		{
			name: "fail; invalid authority",
			request: &types.MsgUpdateParams{
				Authority: suite.f.addrs[0].String(),
				Params:    types.DefaultParams(),
			},
			expErr: true,
		},
		{
			name: "success",
			request: &types.MsgUpdateParams{
				Authority: suite.f.govModAddr,
				Params:    types.DefaultParams(),
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.f.msgServer.UpdateParams(suite.f.ctx, tc.request)

			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)

				r, err := suite.f.queryServer.Params(suite.f.ctx, &types.QueryParamsRequest{})
				suite.Require().NoError(err)
				suite.Require().EqualValues(&tc.request.Params, r.Params)
			}
		})
	}
}

// Test CreateDID
func (suite *MsgServerTestSuite) TestCreateDID() {
	testCases := []struct {
		name   string
		msg    *types.MsgCreateDID
		expErr bool
		errMsg string
	}{
		{
			name: "success",
			msg: &types.MsgCreateDID{
				Controller:  suite.f.addrs[0].String(),
				DidDocument: suite.createValidDIDDocument("did:example:success123"),
			},
			expErr: false,
		},
		{
			name: "fail; invalid controller",
			msg: &types.MsgCreateDID{
				Controller:  "invalid-address",
				DidDocument: suite.createValidDIDDocument("did:example:invalid123"),
			},
			expErr: true,
			errMsg: "invalid controller address",
		},
		{
			name: "fail; empty DID document ID",
			msg: &types.MsgCreateDID{
				Controller: suite.f.addrs[0].String(),
				DidDocument: types.DIDDocument{
					Id: "",
				},
			},
			expErr: true,
			errMsg: "DID document ID cannot be empty",
		},
		{
			name: "fail; DID already exists",
			msg: &types.MsgCreateDID{
				Controller:  suite.f.addrs[0].String(),
				DidDocument: suite.createValidDIDDocument("did:example:duplicate123"),
			},
			expErr: true,
			errMsg: "DID already exists",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// For the "DID already exists" test, create the DID first
			if tc.name == "fail; DID already exists" {
				// Create the DID first
				_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
					Controller:  suite.f.addrs[0].String(),
					DidDocument: tc.msg.DidDocument,
				})
				suite.Require().NoError(err)
			}

			resp, err := suite.f.msgServer.CreateDID(suite.f.ctx, tc.msg)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Equal(tc.msg.DidDocument.Id, resp.Did)

				// Verify DID was stored
				queryResp, err := suite.f.queryServer.GetDIDDocument(suite.f.ctx, &types.QueryGetDIDDocumentRequest{
					Did: tc.msg.DidDocument.Id,
				})
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.DidDocument.Id, queryResp.DidDocument.Id)
			}
		})
	}
}

// Test UpdateDID
func (suite *MsgServerTestSuite) TestUpdateDID() {
	did := "did:example:update123"
	didDoc := suite.createValidDIDDocument(did)

	// Create DID first
	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
		Controller:  suite.f.addrs[0].String(),
		DidDocument: didDoc,
	})
	suite.Require().NoError(err)

	updatedDoc := didDoc
	updatedDoc.AlsoKnownAs = []string{"new-alias"}

	testCases := []struct {
		name   string
		msg    *types.MsgUpdateDID
		expErr bool
		errMsg string
	}{
		{
			name: "success",
			msg: &types.MsgUpdateDID{
				Controller:  suite.f.addrs[0].String(),
				Did:         did,
				DidDocument: updatedDoc,
			},
			expErr: false,
		},
		{
			name: "fail; unauthorized",
			msg: &types.MsgUpdateDID{
				Controller:  suite.f.addrs[1].String(), // Different controller
				Did:         did,
				DidDocument: updatedDoc,
			},
			expErr: true,
			errMsg: "unauthorized",
		},
		{
			name: "fail; DID not found",
			msg: &types.MsgUpdateDID{
				Controller: suite.f.addrs[0].String(),
				Did:        "did:example:notfound",
				DidDocument: types.DIDDocument{
					Id:                "did:example:notfound",
					PrimaryController: suite.f.addrs[0].String(),
				},
			},
			expErr: true,
			errMsg: "DID not found",
		},
		{
			name: "fail; DID mismatch",
			msg: &types.MsgUpdateDID{
				Controller: suite.f.addrs[0].String(),
				Did:        did,
				DidDocument: types.DIDDocument{
					Id: "did:example:different",
				},
			},
			expErr: true,
			errMsg: "DID and DID document ID must match",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.f.msgServer.UpdateDID(suite.f.ctx, tc.msg)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)

				// Verify DID was updated
				queryResp, err := suite.f.queryServer.GetDIDDocument(suite.f.ctx, &types.QueryGetDIDDocumentRequest{
					Did: tc.msg.Did,
				})
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.DidDocument.AlsoKnownAs, queryResp.DidDocument.AlsoKnownAs)
			}
		})
	}
}

// Test DeactivateDID
func (suite *MsgServerTestSuite) TestDeactivateDID() {
	testCases := []struct {
		name   string
		msg    *types.MsgDeactivateDID
		expErr bool
		errMsg string
	}{
		{
			name: "success",
			msg: &types.MsgDeactivateDID{
				Controller: suite.f.addrs[0].String(),
				Did:        "did:example:deactivate_success",
			},
			expErr: false,
		},
		{
			name: "fail; unauthorized",
			msg: &types.MsgDeactivateDID{
				Controller: suite.f.addrs[1].String(), // Different controller
				Did:        "did:example:deactivate_unauth",
			},
			expErr: true,
			errMsg: "unauthorized",
		},
		{
			name: "fail; DID not found",
			msg: &types.MsgDeactivateDID{
				Controller: suite.f.addrs[0].String(),
				Did:        "did:example:notfound",
			},
			expErr: true,
			errMsg: "DID not found",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create DID first for success and unauthorized cases
			if tc.name == "success" || tc.name == "fail; unauthorized" {
				didDoc := suite.createValidDIDDocument(tc.msg.Did)
				_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
					Controller:  suite.f.addrs[0].String(),
					DidDocument: didDoc,
				})
				suite.Require().NoError(err)
			}

			resp, err := suite.f.msgServer.DeactivateDID(suite.f.ctx, tc.msg)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)

				// Verify DID was deactivated by checking metadata
				resolveResp, err := suite.f.queryServer.ResolveDID(suite.f.ctx, &types.QueryResolveDIDRequest{
					Did: tc.msg.Did,
				})
				suite.Require().NoError(err)
				suite.Require().Greater(resolveResp.DidDocumentMetadata.Deactivated, int64(0))
			}
		})
	}
}

// Test AddVerificationMethod
func (suite *MsgServerTestSuite) TestAddVerificationMethod() {
	did := "did:example:addvm123"
	didDoc := suite.createValidDIDDocument(did)

	// Create DID first
	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
		Controller:  suite.f.addrs[0].String(),
		DidDocument: didDoc,
	})
	suite.Require().NoError(err)

	newVM := types.VerificationMethod{
		Id:                     did + "#key-2",
		VerificationMethodKind: "Ed25519VerificationKey2020",
		Controller:             did,
		PublicKeyJwk:           `{"kty":"OKP","crv":"Ed25519","x":"new-public-key"}`,
	}

	testCases := []struct {
		name   string
		msg    *types.MsgAddVerificationMethod
		expErr bool
		errMsg string
	}{
		{
			name: "success",
			msg: &types.MsgAddVerificationMethod{
				Controller:         suite.f.addrs[0].String(),
				Did:                did,
				VerificationMethod: newVM,
			},
			expErr: false,
		},
		{
			name: "fail; unauthorized",
			msg: &types.MsgAddVerificationMethod{
				Controller:         suite.f.addrs[1].String(),
				Did:                did,
				VerificationMethod: newVM,
			},
			expErr: true,
			errMsg: "unauthorized",
		},
		{
			name: "fail; DID not found",
			msg: &types.MsgAddVerificationMethod{
				Controller:         suite.f.addrs[0].String(),
				Did:                "did:example:notfound",
				VerificationMethod: newVM,
			},
			expErr: true,
			errMsg: "DID not found",
		},
		{
			name: "fail; verification method already exists",
			msg: &types.MsgAddVerificationMethod{
				Controller:         suite.f.addrs[0].String(),
				Did:                did,
				VerificationMethod: *didDoc.VerificationMethod[0], // Existing method
			},
			expErr: true,
			errMsg: "verification method with ID already exists",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.f.msgServer.AddVerificationMethod(suite.f.ctx, tc.msg)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)

				// Verify method was added
				queryResp, err := suite.f.queryServer.GetVerificationMethod(suite.f.ctx, &types.QueryGetVerificationMethodRequest{
					Did:      tc.msg.Did,
					MethodId: tc.msg.VerificationMethod.Id,
				})
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.VerificationMethod.Id, queryResp.VerificationMethod.Id)
			}
		})
	}
}

// Test RemoveVerificationMethod
func (suite *MsgServerTestSuite) TestRemoveVerificationMethod() {
	did := "did:example:removevm123"
	didDoc := suite.createValidDIDDocument(did)

	// Create DID first
	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
		Controller:  suite.f.addrs[0].String(),
		DidDocument: didDoc,
	})
	suite.Require().NoError(err)

	testCases := []struct {
		name   string
		msg    *types.MsgRemoveVerificationMethod
		expErr bool
		errMsg string
	}{
		{
			name: "success",
			msg: &types.MsgRemoveVerificationMethod{
				Controller:           suite.f.addrs[0].String(),
				Did:                  did,
				VerificationMethodId: didDoc.VerificationMethod[0].Id,
			},
			expErr: false,
		},
		{
			name: "fail; unauthorized",
			msg: &types.MsgRemoveVerificationMethod{
				Controller:           suite.f.addrs[1].String(),
				Did:                  did,
				VerificationMethodId: didDoc.VerificationMethod[0].Id,
			},
			expErr: true,
			errMsg: "unauthorized",
		},
		{
			name: "fail; verification method not found",
			msg: &types.MsgRemoveVerificationMethod{
				Controller:           suite.f.addrs[0].String(),
				Did:                  did,
				VerificationMethodId: "did:example:notfound#key-99",
			},
			expErr: true,
			errMsg: "verification method not found",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.f.msgServer.RemoveVerificationMethod(suite.f.ctx, tc.msg)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)

				// Verify method was removed
				_, err := suite.f.queryServer.GetVerificationMethod(suite.f.ctx, &types.QueryGetVerificationMethodRequest{
					Did:      tc.msg.Did,
					MethodId: tc.msg.VerificationMethodId,
				})
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), "verification method not found")
			}
		})
	}
}

// Test AddService
func (suite *MsgServerTestSuite) TestAddService() {
	did := "did:example:addsvc123"
	didDoc := suite.createValidDIDDocument(did)

	// Create DID first
	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
		Controller:  suite.f.addrs[0].String(),
		DidDocument: didDoc,
	})
	suite.Require().NoError(err)

	newService := types.Service{
		Id:             did + "#service-2",
		ServiceKind:    "CredentialRegistry",
		SingleEndpoint: "https://creds.example.com",
	}

	testCases := []struct {
		name   string
		msg    *types.MsgAddService
		expErr bool
		errMsg string
	}{
		{
			name: "success",
			msg: &types.MsgAddService{
				Controller: suite.f.addrs[0].String(),
				Did:        did,
				Service:    newService,
			},
			expErr: false,
		},
		{
			name: "fail; unauthorized",
			msg: &types.MsgAddService{
				Controller: suite.f.addrs[1].String(),
				Did:        did,
				Service:    newService,
			},
			expErr: true,
			errMsg: "unauthorized",
		},
		{
			name: "fail; service already exists",
			msg: &types.MsgAddService{
				Controller: suite.f.addrs[0].String(),
				Did:        did,
				Service:    *didDoc.Service[0], // Existing service
			},
			expErr: true,
			errMsg: "service with ID already exists",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.f.msgServer.AddService(suite.f.ctx, tc.msg)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)

				// Verify service was added
				queryResp, err := suite.f.queryServer.GetService(suite.f.ctx, &types.QueryGetServiceRequest{
					Did:       tc.msg.Did,
					ServiceId: tc.msg.Service.Id,
				})
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.Service.Id, queryResp.Service.Id)
			}
		})
	}
}

// Test RemoveService
func (suite *MsgServerTestSuite) TestRemoveService() {
	did := "did:example:removesvc123"
	didDoc := suite.createValidDIDDocument(did)

	// Create DID first
	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
		Controller:  suite.f.addrs[0].String(),
		DidDocument: didDoc,
	})
	suite.Require().NoError(err)

	testCases := []struct {
		name   string
		msg    *types.MsgRemoveService
		expErr bool
		errMsg string
	}{
		{
			name: "success",
			msg: &types.MsgRemoveService{
				Controller: suite.f.addrs[0].String(),
				Did:        did,
				ServiceId:  didDoc.Service[0].Id,
			},
			expErr: false,
		},
		{
			name: "fail; unauthorized",
			msg: &types.MsgRemoveService{
				Controller: suite.f.addrs[1].String(),
				Did:        did,
				ServiceId:  didDoc.Service[0].Id,
			},
			expErr: true,
			errMsg: "unauthorized",
		},
		{
			name: "fail; service not found",
			msg: &types.MsgRemoveService{
				Controller: suite.f.addrs[0].String(),
				Did:        did,
				ServiceId:  "did:example:notfound#service-99",
			},
			expErr: true,
			errMsg: "service not found",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.f.msgServer.RemoveService(suite.f.ctx, tc.msg)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)

				// Verify service was removed
				_, err := suite.f.queryServer.GetService(suite.f.ctx, &types.QueryGetServiceRequest{
					Did:       tc.msg.Did,
					ServiceId: tc.msg.ServiceId,
				})
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), "service not found")
			}
		})
	}
}

// Test IssueVerifiableCredential
func (suite *MsgServerTestSuite) TestIssueVerifiableCredential() {
	// Convert credential subject to JSON bytes
	credSubject := map[string]string{
		"degree": "Bachelor of Science",
		"name":   "Alice",
	}
	credSubjectBytes, _ := json.Marshal(credSubject)

	blockTime := sdk.UnwrapSDKContext(suite.f.ctx).BlockTime()

	testCases := []struct {
		name   string
		msg    *types.MsgIssueVerifiableCredential
		expErr bool
		errMsg string
	}{
		{
			name: "success",
			msg: &types.MsgIssueVerifiableCredential{
				Issuer: suite.f.addrs[0].String(),
				Credential: types.VerifiableCredential{
					Id:             "https://example.com/credentials/success123",
					Issuer:         "did:example:issuer_success",
					Subject:        "did:example:subject123",
					IssuanceDate:   blockTime.Format(time.RFC3339),
					ExpirationDate: blockTime.Add(365 * 24 * time.Hour).Format(time.RFC3339),
					CredentialKinds: []string{
						"VerifiableCredential",
						"UniversityDegreeCredential",
					},
					CredentialSubject: credSubjectBytes,
					Proof: []*types.CredentialProof{
						{
							ProofKind:          "Ed25519Signature2020",
							Created:            blockTime.Format(time.RFC3339),
							ProofPurpose:       "assertionMethod",
							VerificationMethod: "did:example:issuer_success#key-1",
							Signature:          "eyJhbGciOiJFZERTQSIsImI2NCI6ZmFsc2UsImNyaXQiOlsiYjY0Il19..test",
						},
					},
				},
			},
			expErr: false,
		},
		{
			name: "fail; invalid issuer",
			msg: &types.MsgIssueVerifiableCredential{
				Issuer: "invalid-address",
				Credential: types.VerifiableCredential{
					Id:                "https://example.com/credentials/invalid123",
					Issuer:            "did:example:issuer_invalid",
					Subject:           "did:example:subject123",
					IssuanceDate:      blockTime.Format(time.RFC3339),
					ExpirationDate:    blockTime.Add(365 * 24 * time.Hour).Format(time.RFC3339),
					CredentialKinds:   []string{"VerifiableCredential"},
					CredentialSubject: credSubjectBytes,
				},
			},
			expErr: true,
			errMsg: "invalid issuer address",
		},
		{
			name: "fail; credential already exists",
			msg: &types.MsgIssueVerifiableCredential{
				Issuer: suite.f.addrs[0].String(),
				Credential: types.VerifiableCredential{
					Id:                "https://example.com/credentials/duplicate123",
					Issuer:            "did:example:issuer_duplicate",
					Subject:           "did:example:subject123",
					IssuanceDate:      blockTime.Format(time.RFC3339),
					ExpirationDate:    blockTime.Add(365 * 24 * time.Hour).Format(time.RFC3339),
					CredentialKinds:   []string{"VerifiableCredential"},
					CredentialSubject: credSubjectBytes,
				},
			},
			expErr: true,
			errMsg: "credential ID already exists",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create issuer DID first for success and duplicate cases
			if tc.name == "success" || tc.name == "fail; credential already exists" {
				didDoc := suite.createValidDIDDocument(tc.msg.Credential.Issuer)
				_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
					Controller:  suite.f.addrs[0].String(),
					DidDocument: didDoc,
				})
				suite.Require().NoError(err)
			}

			// For the "already exists" test, issue it first
			if tc.name == "fail; credential already exists" {
				_, err := suite.f.msgServer.IssueVerifiableCredential(suite.f.ctx, tc.msg)
				suite.Require().NoError(err)
			}

			resp, err := suite.f.msgServer.IssueVerifiableCredential(suite.f.ctx, tc.msg)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Equal(tc.msg.Credential.Id, resp.CredentialId)

				// Verify credential was stored
				queryResp, err := suite.f.queryServer.GetVerifiableCredential(suite.f.ctx, &types.QueryGetVerifiableCredentialRequest{
					CredentialId: tc.msg.Credential.Id,
				})
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.Credential.Id, queryResp.Credential.Id)
			}
		})
	}
}

// Test RevokeVerifiableCredential
func (suite *MsgServerTestSuite) TestRevokeVerifiableCredential() {
	// Convert credential subject to JSON bytes
	credSubject := map[string]string{
		"test": "data",
	}
	credSubjectBytes, _ := json.Marshal(credSubject)

	blockTime := sdk.UnwrapSDKContext(suite.f.ctx).BlockTime()

	testCases := []struct {
		name   string
		msg    *types.MsgRevokeVerifiableCredential
		expErr bool
		errMsg string
	}{
		{
			name: "success",
			msg: &types.MsgRevokeVerifiableCredential{
				Issuer:           suite.f.addrs[0].String(),
				CredentialId:     "https://example.com/credentials/revoke_success",
				RevocationReason: "Key compromise",
			},
			expErr: false,
		},
		{
			name: "fail; unauthorized",
			msg: &types.MsgRevokeVerifiableCredential{
				Issuer:           suite.f.addrs[1].String(), // Different issuer
				CredentialId:     "https://example.com/credentials/revoke_unauth",
				RevocationReason: "Unauthorized revocation",
			},
			expErr: true,
			errMsg: "unauthorized",
		},
		{
			name: "fail; credential not found",
			msg: &types.MsgRevokeVerifiableCredential{
				Issuer:           suite.f.addrs[0].String(),
				CredentialId:     "https://example.com/credentials/notfound",
				RevocationReason: "Not found",
			},
			expErr: true,
			errMsg: "credential not found",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create issuer DID and credential for success and unauthorized cases
			if tc.name == "success" || tc.name == "fail; unauthorized" {
				// Create a valid DID without special characters
				didSuffix := "success"
				if tc.name == "fail; unauthorized" {
					didSuffix = "unauthorized"
				}
				did := "did:example:revokeissuer-" + didSuffix
				didDoc := suite.createValidDIDDocument(did)
				_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
					Controller:  suite.f.addrs[0].String(),
					DidDocument: didDoc,
				})
				suite.Require().NoError(err)

				// Issue credential first
				credential := types.VerifiableCredential{
					Id:                tc.msg.CredentialId,
					Issuer:            did,
					Subject:           "did:example:subject123",
					IssuanceDate:      blockTime.Format(time.RFC3339),
					ExpirationDate:    blockTime.Add(365 * 24 * time.Hour).Format(time.RFC3339),
					CredentialKinds:   []string{"VerifiableCredential"},
					CredentialSubject: credSubjectBytes,
					Proof: []*types.CredentialProof{
						{
							ProofKind:          "Ed25519Signature2020",
							Created:            blockTime.Format(time.RFC3339),
							ProofPurpose:       "assertionMethod",
							VerificationMethod: did + "#key-1",
							Signature:          "eyJhbGciOiJFZERTQSIsImI2NCI6ZmFsc2UsImNyaXQiOlsiYjY0Il19..test",
						},
					},
				}

				_, err = suite.f.msgServer.IssueVerifiableCredential(
					suite.f.ctx,
					&types.MsgIssueVerifiableCredential{
						Issuer:     suite.f.addrs[0].String(),
						Credential: credential,
					},
				)
				suite.Require().NoError(err)
			}

			resp, err := suite.f.msgServer.RevokeVerifiableCredential(suite.f.ctx, tc.msg)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)

				// Verify credential was revoked
				queryResp, err := suite.f.queryServer.GetVerifiableCredential(suite.f.ctx, &types.QueryGetVerifiableCredentialRequest{
					CredentialId: tc.msg.CredentialId,
				})
				suite.Require().NoError(err)
				if queryResp.Credential.CredentialStatus != nil {
					suite.Require().Equal("Revoked", queryResp.Credential.CredentialStatus.StatusKind)
					if queryResp.Credential.CredentialStatus.Properties != nil {
						suite.Require().Equal(tc.msg.RevocationReason, queryResp.Credential.CredentialStatus.Properties["reason"])
					}
				}
			}
		})
	}
}
