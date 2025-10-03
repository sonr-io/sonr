package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/sonr-io/sonr/x/did/types"
)

type QueryServerTestSuite struct {
	suite.Suite
	f *testFixture
}

func TestQueryServerSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

func (suite *QueryServerTestSuite) SetupTest() {
	suite.f = SetupTest(suite.T())
}

// Helper function to create test DID documents
func (suite *QueryServerTestSuite) createTestDIDDocuments(count int) []string {
	dids := make([]string, count)
	for i := 0; i < count; i++ {
		did := fmt.Sprintf("did:example:test%d", i)
		dids[i] = did

		didDoc := types.DIDDocument{
			Id:                did,
			PrimaryController: suite.f.addrs[0].String(),
			AlsoKnownAs:       []string{fmt.Sprintf("alias%d", i)},
			VerificationMethod: []*types.VerificationMethod{
				{
					Id:                     did + "#key-1",
					VerificationMethodKind: "Ed25519VerificationKey2020",
					Controller:             did,
					PublicKeyJwk:           `{"kty":"OKP","crv":"Ed25519","x":"test-key"}`,
				},
			},
			Authentication: []*types.VerificationMethodReference{
				{VerificationMethodId: did + "#key-1"},
			},
			AssertionMethod: []*types.VerificationMethodReference{
				{VerificationMethodId: did + "#key-1"},
			},
			Service: []*types.Service{
				{
					Id:             did + "#service-1",
					ServiceKind:    "LinkedDomains",
					SingleEndpoint: fmt.Sprintf("https://example%d.com", i),
				},
			},
		}

		_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
			Controller:  suite.f.addrs[0].String(),
			DidDocument: didDoc,
		})
		suite.Require().NoError(err)
	}
	return dids
}

// Test ResolveDID
func (suite *QueryServerTestSuite) TestResolveDID() {
	did := "did:example:resolve123"
	didDoc := types.DIDDocument{
		Id:                did,
		PrimaryController: suite.f.addrs[0].String(),
		AlsoKnownAs:       []string{"test-alias"},
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     did + "#key-1",
				VerificationMethodKind: "Ed25519VerificationKey2020",
				Controller:             did,
				PublicKeyJwk:           `{"kty":"OKP","crv":"Ed25519","x":"test-key"}`,
			},
		},
	}

	// Create DID
	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
		Controller:  suite.f.addrs[0].String(),
		DidDocument: didDoc,
	})
	suite.Require().NoError(err)

	testCases := []struct {
		name   string
		req    *types.QueryResolveDIDRequest
		expErr bool
		errMsg string
	}{
		{
			name:   "success",
			req:    &types.QueryResolveDIDRequest{Did: did},
			expErr: false,
		},
		{
			name:   "fail; empty DID",
			req:    &types.QueryResolveDIDRequest{Did: ""},
			expErr: true,
			errMsg: "DID cannot be empty",
		},
		{
			name:   "fail; DID not found",
			req:    &types.QueryResolveDIDRequest{Did: "did:example:notfound"},
			expErr: true,
			errMsg: "DID not found",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.f.queryServer.ResolveDID(suite.f.ctx, tc.req)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Equal(tc.req.Did, resp.DidDocument.Id)
				suite.Require().NotNil(resp.DidDocumentMetadata)
				suite.Require().Equal(int64(0), resp.DidDocumentMetadata.Deactivated)
			}
		})
	}
}

// Test GetDIDDocument
func (suite *QueryServerTestSuite) TestGetDIDDocument() {
	did := "did:example:get123"
	didDoc := types.DIDDocument{
		Id:                did,
		PrimaryController: suite.f.addrs[0].String(),
	}

	// Create DID
	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
		Controller:  suite.f.addrs[0].String(),
		DidDocument: didDoc,
	})
	suite.Require().NoError(err)

	testCases := []struct {
		name   string
		req    *types.QueryGetDIDDocumentRequest
		expErr bool
		errMsg string
	}{
		{
			name:   "success",
			req:    &types.QueryGetDIDDocumentRequest{Did: did},
			expErr: false,
		},
		{
			name:   "fail; empty DID",
			req:    &types.QueryGetDIDDocumentRequest{Did: ""},
			expErr: true,
			errMsg: "DID cannot be empty",
		},
		{
			name:   "fail; DID not found",
			req:    &types.QueryGetDIDDocumentRequest{Did: "did:example:notfound"},
			expErr: true,
			errMsg: "DID not found",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.f.queryServer.GetDIDDocument(suite.f.ctx, tc.req)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Equal(tc.req.Did, resp.DidDocument.Id)
			}
		})
	}
}

// Test ListDIDDocuments
func (suite *QueryServerTestSuite) TestListDIDDocuments() {
	// Create test documents
	dids := suite.createTestDIDDocuments(5)

	testCases := []struct {
		name      string
		req       *types.QueryListDIDDocumentsRequest
		expErr    bool
		expCount  int
		checkDids []string
	}{
		{
			name: "list all documents",
			req: &types.QueryListDIDDocumentsRequest{
				Pagination: &query.PageRequest{Limit: 10},
			},
			expErr:    false,
			expCount:  5,
			checkDids: dids,
		},
		{
			name: "paginate with limit",
			req: &types.QueryListDIDDocumentsRequest{
				Pagination: &query.PageRequest{Limit: 2},
			},
			expErr:   false,
			expCount: 2,
		},
		{
			name: "paginate with offset",
			req: &types.QueryListDIDDocumentsRequest{
				Pagination: &query.PageRequest{Limit: 10, Offset: 3},
			},
			expErr:   false,
			expCount: 2,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.f.queryServer.ListDIDDocuments(suite.f.ctx, tc.req)

			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Len(resp.DidDocuments, tc.expCount)

				if tc.checkDids != nil {
					for i, did := range resp.DidDocuments {
						suite.Require().Equal(tc.checkDids[i], did.Id)
					}
				}
			}
		})
	}
}

// Test GetVerificationMethod
func (suite *QueryServerTestSuite) TestGetVerificationMethod() {
	did := "did:example:vm123"
	methodId := did + "#key-1"

	didDoc := types.DIDDocument{
		Id:                did,
		PrimaryController: suite.f.addrs[0].String(),
		VerificationMethod: []*types.VerificationMethod{
			{
				Id:                     methodId,
				VerificationMethodKind: "Ed25519VerificationKey2020",
				Controller:             did,
				PublicKeyJwk:           `{"kty":"OKP","crv":"Ed25519","x":"test-key"}`,
			},
		},
	}

	// Create DID
	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
		Controller:  suite.f.addrs[0].String(),
		DidDocument: didDoc,
	})
	suite.Require().NoError(err)

	testCases := []struct {
		name   string
		req    *types.QueryGetVerificationMethodRequest
		expErr bool
		errMsg string
	}{
		{
			name: "success",
			req: &types.QueryGetVerificationMethodRequest{
				Did:      did,
				MethodId: methodId,
			},
			expErr: false,
		},
		{
			name: "fail; empty DID",
			req: &types.QueryGetVerificationMethodRequest{
				Did:      "",
				MethodId: methodId,
			},
			expErr: true,
			errMsg: "DID cannot be empty",
		},
		{
			name: "fail; empty method ID",
			req: &types.QueryGetVerificationMethodRequest{
				Did:      did,
				MethodId: "",
			},
			expErr: true,
			errMsg: "method ID cannot be empty",
		},
		{
			name: "fail; DID not found",
			req: &types.QueryGetVerificationMethodRequest{
				Did:      "did:example:notfound",
				MethodId: methodId,
			},
			expErr: true,
			errMsg: "DID not found",
		},
		{
			name: "fail; method not found",
			req: &types.QueryGetVerificationMethodRequest{
				Did:      did,
				MethodId: did + "#notfound",
			},
			expErr: true,
			errMsg: "verification method not found",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.f.queryServer.GetVerificationMethod(suite.f.ctx, tc.req)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Equal(tc.req.MethodId, resp.VerificationMethod.Id)
			}
		})
	}
}

// Test GetService
func (suite *QueryServerTestSuite) TestGetService() {
	did := "did:example:svc123"
	serviceId := did + "#service-1"

	didDoc := types.DIDDocument{
		Id:                did,
		PrimaryController: suite.f.addrs[0].String(),
		Service: []*types.Service{
			{
				Id:             serviceId,
				ServiceKind:    "LinkedDomains",
				SingleEndpoint: "https://example.com",
			},
		},
	}

	// Create DID
	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
		Controller:  suite.f.addrs[0].String(),
		DidDocument: didDoc,
	})
	suite.Require().NoError(err)

	testCases := []struct {
		name   string
		req    *types.QueryGetServiceRequest
		expErr bool
		errMsg string
	}{
		{
			name: "success",
			req: &types.QueryGetServiceRequest{
				Did:       did,
				ServiceId: serviceId,
			},
			expErr: false,
		},
		{
			name: "fail; empty DID",
			req: &types.QueryGetServiceRequest{
				Did:       "",
				ServiceId: serviceId,
			},
			expErr: true,
			errMsg: "DID cannot be empty",
		},
		{
			name: "fail; empty service ID",
			req: &types.QueryGetServiceRequest{
				Did:       did,
				ServiceId: "",
			},
			expErr: true,
			errMsg: "service ID cannot be empty",
		},
		{
			name: "fail; service not found",
			req: &types.QueryGetServiceRequest{
				Did:       did,
				ServiceId: did + "#notfound",
			},
			expErr: true,
			errMsg: "service not found",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.f.queryServer.GetService(suite.f.ctx, tc.req)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Equal(tc.req.ServiceId, resp.Service.Id)
			}
		})
	}
}

// Test GetVerifiableCredential
func (suite *QueryServerTestSuite) TestGetVerifiableCredential() {
	did := "did:example:issuer456"
	credentialId := "https://example.com/credentials/456"

	// Create issuer DID
	didDoc := types.DIDDocument{
		Id:                did,
		PrimaryController: suite.f.addrs[0].String(),
	}
	_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
		Controller:  suite.f.addrs[0].String(),
		DidDocument: didDoc,
	})
	suite.Require().NoError(err)

	// Issue credential
	credential := &types.VerifiableCredential{
		Id:           credentialId,
		Issuer:       did,
		Subject:      "did:example:subject456",
		IssuanceDate: sdk.UnwrapSDKContext(suite.f.ctx).BlockTime().Format(time.RFC3339),
		ExpirationDate: sdk.UnwrapSDKContext(suite.f.ctx).
			BlockTime().
			Add(365 * 24 * time.Hour).
			Format(time.RFC3339),
		CredentialKinds:   []string{"VerifiableCredential"},
		CredentialSubject: []byte(`{"test": "data"}`),
	}

	_, err = suite.f.msgServer.IssueVerifiableCredential(
		suite.f.ctx,
		&types.MsgIssueVerifiableCredential{
			Issuer:     suite.f.addrs[0].String(),
			Credential: *credential,
		},
	)
	suite.Require().NoError(err)

	testCases := []struct {
		name   string
		req    *types.QueryGetVerifiableCredentialRequest
		expErr bool
		errMsg string
	}{
		{
			name:   "success",
			req:    &types.QueryGetVerifiableCredentialRequest{CredentialId: credentialId},
			expErr: false,
		},
		{
			name:   "fail; empty credential ID",
			req:    &types.QueryGetVerifiableCredentialRequest{CredentialId: ""},
			expErr: true,
			errMsg: "credential ID cannot be empty",
		},
		{
			name: "fail; credential not found",
			req: &types.QueryGetVerifiableCredentialRequest{
				CredentialId: "https://example.com/notfound",
			},
			expErr: true,
			errMsg: "credential not found",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.f.queryServer.GetVerifiableCredential(suite.f.ctx, tc.req)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.errMsg)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Equal(tc.req.CredentialId, resp.Credential.Id)
			}
		})
	}
}

// Test ListVerifiableCredentials with enhanced filtering
func (suite *QueryServerTestSuite) TestListVerifiableCredentials() {
	issuerDid := "did:example:issuer789"
	issuerDid2 := "did:example:issuer790"
	subjectDid := "did:example:subject789"

	// Create issuer DIDs
	for _, did := range []string{issuerDid, issuerDid2} {
		didDoc := types.DIDDocument{
			Id:                did,
			PrimaryController: suite.f.addrs[0].String(),
		}
		_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
			Controller:  suite.f.addrs[0].String(),
			DidDocument: didDoc,
		})
		suite.Require().NoError(err)
	}

	// Issue multiple credentials with different issuers and subjects
	credentialIds := []string{}
	for i := 0; i < 3; i++ {
		// Use different issuer for the third credential
		issuer := issuerDid
		if i == 2 {
			issuer = issuerDid2
		}

		credId := fmt.Sprintf("https://example.com/credentials/list%d", i)
		credentialIds = append(credentialIds, credId)

		credential := &types.VerifiableCredential{
			Id:           credId,
			Issuer:       issuer,
			Subject:      fmt.Sprintf("%s%d", subjectDid, i),
			IssuanceDate: sdk.UnwrapSDKContext(suite.f.ctx).BlockTime().Format(time.RFC3339),
			ExpirationDate: sdk.UnwrapSDKContext(suite.f.ctx).
				BlockTime().
				Add(365 * 24 * time.Hour).
				Format(time.RFC3339),
			CredentialKinds:   []string{"VerifiableCredential"},
			CredentialSubject: []byte(`{"test": "data"}`),
		}

		_, err := suite.f.msgServer.IssueVerifiableCredential(
			suite.f.ctx,
			&types.MsgIssueVerifiableCredential{
				Issuer:     suite.f.addrs[0].String(),
				Credential: *credential,
			},
		)
		suite.Require().NoError(err)
	}

	// Revoke one credential for testing
	_, err := suite.f.msgServer.RevokeVerifiableCredential(
		suite.f.ctx,
		&types.MsgRevokeVerifiableCredential{
			Issuer:       suite.f.addrs[0].String(),
			CredentialId: credentialIds[0],
		},
	)
	suite.Require().NoError(err)

	testCases := []struct {
		name      string
		req       *types.QueryListVerifiableCredentialsRequest
		expCount  int
		checkFunc func(*types.QueryListVerifiableCredentialsResponse)
	}{
		{
			name: "list all credentials without revoked",
			req: &types.QueryListVerifiableCredentialsRequest{
				Pagination:     &query.PageRequest{Limit: 10},
				IncludeRevoked: false,
			},
			expCount: 2, // 3 issued - 1 revoked
		},
		{
			name: "list all credentials including revoked",
			req: &types.QueryListVerifiableCredentialsRequest{
				Pagination:     &query.PageRequest{Limit: 10},
				IncludeRevoked: true,
			},
			expCount: 3,
		},
		{
			name: "filter by issuer",
			req: &types.QueryListVerifiableCredentialsRequest{
				Issuer:         issuerDid,
				Pagination:     &query.PageRequest{Limit: 10},
				IncludeRevoked: true,
			},
			expCount: 2, // First two credentials
		},
		{
			name: "filter by holder/subject",
			req: &types.QueryListVerifiableCredentialsRequest{
				Holder:         fmt.Sprintf("%s1", subjectDid),
				Pagination:     &query.PageRequest{Limit: 10},
				IncludeRevoked: false,
			},
			expCount: 1,
			checkFunc: func(resp *types.QueryListVerifiableCredentialsResponse) {
				suite.Require().Equal(fmt.Sprintf("%s1", subjectDid), resp.Credentials[0].Subject)
			},
		},
		{
			name: "filter by non-existent issuer",
			req: &types.QueryListVerifiableCredentialsRequest{
				Issuer:     "did:example:notfound",
				Pagination: &query.PageRequest{Limit: 10},
			},
			expCount: 0,
		},
		{
			name: "pagination with limit",
			req: &types.QueryListVerifiableCredentialsRequest{
				Pagination:     &query.PageRequest{Limit: 1},
				IncludeRevoked: true,
			},
			expCount: 1,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.f.queryServer.ListVerifiableCredentials(suite.f.ctx, tc.req)
			suite.Require().NoError(err)
			suite.Require().NotNil(resp)
			suite.Require().Len(resp.Credentials, tc.expCount)

			if tc.checkFunc != nil {
				tc.checkFunc(resp)
			}
		})
	}
}

// Test GetCredentialsByDID - new unified method
func (suite *QueryServerTestSuite) TestGetCredentialsByDID() {
	issuerDid := "did:example:issuer_unified"
	holderDid := "did:example:holder_unified"
	otherIssuerDid := "did:example:other_issuer"

	// Create DIDs
	for _, did := range []string{issuerDid, holderDid, otherIssuerDid} {
		// Add WebAuthn credential for the holder DID
		var verificationMethod []*types.VerificationMethod
		if did == holderDid {
			verificationMethod = []*types.VerificationMethod{
				{
					Id:                     did + "#webauthn-1",
					VerificationMethodKind: "WebAuthnCredential2024",
					Controller:             did,
					WebauthnCredential: &types.WebAuthnCredential{
						CredentialId:       "webauthn-cred-1",
						PublicKey:          []byte("test-public-key"),
						Algorithm:          -7, // ES256
						AttestationType:    "none",
						Origin:             "https://example.com",
						RpId:               "example.com",
						RpName:             "Example",
						SignatureAlgorithm: "ES256",
					},
				},
			}
		}

		didDoc := types.DIDDocument{
			Id:                 did,
			PrimaryController:  suite.f.addrs[0].String(),
			VerificationMethod: verificationMethod,
		}
		_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
			Controller:  suite.f.addrs[0].String(),
			DidDocument: didDoc,
		})
		suite.Require().NoError(err)
	}

	// Issue verifiable credentials
	// 1. Credential issued by issuerDid
	_, err := suite.f.msgServer.IssueVerifiableCredential(
		suite.f.ctx,
		&types.MsgIssueVerifiableCredential{
			Issuer: suite.f.addrs[0].String(),
			Credential: types.VerifiableCredential{
				Id:      "https://example.com/cred/1",
				Issuer:  issuerDid,
				Subject: holderDid,
				IssuanceDate: sdk.UnwrapSDKContext(suite.f.ctx).
					BlockTime().
					Format(time.RFC3339),
				CredentialKinds:   []string{"VerifiableCredential"},
				CredentialSubject: []byte(`{"test": "data1"}`),
			},
		},
	)
	suite.Require().NoError(err)

	// 2. Credential held by holderDid (different issuer)
	_, err = suite.f.msgServer.IssueVerifiableCredential(
		suite.f.ctx,
		&types.MsgIssueVerifiableCredential{
			Issuer: suite.f.addrs[0].String(),
			Credential: types.VerifiableCredential{
				Id:      "https://example.com/cred/2",
				Issuer:  otherIssuerDid,
				Subject: holderDid,
				IssuanceDate: sdk.UnwrapSDKContext(suite.f.ctx).
					BlockTime().
					Format(time.RFC3339),
				CredentialKinds:   []string{"VerifiableCredential"},
				CredentialSubject: []byte(`{"test": "data2"}`),
			},
		},
	)
	suite.Require().NoError(err)

	testCases := []struct {
		name               string
		req                *types.QueryGetCredentialsByDIDRequest
		expVerifiableCount int
		expWebAuthnCount   int
		expTotalCount      int
	}{
		{
			name: "get all credentials for issuer DID",
			req: &types.QueryGetCredentialsByDIDRequest{
				Did:               issuerDid,
				IncludeVerifiable: true,
				IncludeWebauthn:   true,
			},
			expVerifiableCount: 1, // 1 credential issued by this DID
			expWebAuthnCount:   0, // No WebAuthn credentials
			expTotalCount:      1,
		},
		{
			name: "get all credentials for holder DID",
			req: &types.QueryGetCredentialsByDIDRequest{
				Did:               holderDid,
				IncludeVerifiable: true,
				IncludeWebauthn:   true,
			},
			expVerifiableCount: 2, // 2 credentials where this DID is subject
			expWebAuthnCount:   1, // 1 WebAuthn credential
			expTotalCount:      3,
		},
		{
			name: "get only verifiable credentials",
			req: &types.QueryGetCredentialsByDIDRequest{
				Did:               holderDid,
				IncludeVerifiable: true,
				IncludeWebauthn:   false,
			},
			expVerifiableCount: 2,
			expWebAuthnCount:   0,
			expTotalCount:      2,
		},
		{
			name: "get only WebAuthn credentials",
			req: &types.QueryGetCredentialsByDIDRequest{
				Did:               holderDid,
				IncludeVerifiable: false,
				IncludeWebauthn:   true,
			},
			expVerifiableCount: 0,
			expWebAuthnCount:   1,
			expTotalCount:      1,
		},
		{
			name: "non-existent DID",
			req: &types.QueryGetCredentialsByDIDRequest{
				Did:               "did:example:notfound",
				IncludeVerifiable: true,
				IncludeWebauthn:   true,
			},
			expVerifiableCount: 0,
			expWebAuthnCount:   0,
			expTotalCount:      0,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.f.queryServer.GetCredentialsByDID(suite.f.ctx, tc.req)
			suite.Require().NoError(err)
			suite.Require().NotNil(resp)
			suite.Require().Len(resp.Credentials, tc.expTotalCount)

			// Count credential types
			verifiableCount := 0
			webauthnCount := 0
			for _, cred := range resp.Credentials {
				if cred.GetVerifiableCredential() != nil {
					verifiableCount++
				}
				if cred.GetWebauthnCredential() != nil {
					webauthnCount++
				}
			}

			suite.Require().
				Equal(tc.expVerifiableCount, verifiableCount, "verifiable credential count mismatch")
			suite.Require().
				Equal(tc.expWebAuthnCount, webauthnCount, "WebAuthn credential count mismatch")
		})
	}
}

// Test GetDIDDocumentsByController
func (suite *QueryServerTestSuite) TestGetDIDDocumentsByController() {
	controllerAddr := suite.f.addrs[0].String()

	// Create multiple DIDs controlled by the same controller
	for i := 0; i < 3; i++ {
		did := fmt.Sprintf("did:example:bycontroller%d", i)
		didDoc := types.DIDDocument{
			Id:                did,
			PrimaryController: controllerAddr,
		}
		_, err := suite.f.msgServer.CreateDID(suite.f.ctx, &types.MsgCreateDID{
			Controller:  controllerAddr,
			DidDocument: didDoc,
		})
		suite.Require().NoError(err)
	}

	// Test retrieving DIDs by controller
	resp, err := suite.f.queryServer.GetDIDDocumentsByController(
		suite.f.ctx,
		&types.QueryGetDIDDocumentsByControllerRequest{
			Controller: controllerAddr,
			Pagination: &query.PageRequest{Limit: 10},
		},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().GreaterOrEqual(len(resp.DidDocuments), 3)

	// Test with non-existent controller
	emptyResp, err := suite.f.queryServer.GetDIDDocumentsByController(
		suite.f.ctx,
		&types.QueryGetDIDDocumentsByControllerRequest{
			Controller: "idx1notfound123456789",
			Pagination: &query.PageRequest{Limit: 10},
		},
	)
	suite.Require().NoError(err)
	suite.Require().NotNil(emptyResp)
	suite.Require().Len(emptyResp.DidDocuments, 0)
}
