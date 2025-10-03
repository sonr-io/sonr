package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	v1 "github.com/sonr-io/sonr/api/svc/v1"
	"github.com/sonr-io/sonr/x/svc/types"
)

func TestParams(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	testCases := []struct {
		name    string
		request *types.MsgUpdateParams
		err     bool
	}{
		{
			name: "fail; invalid authority",
			request: &types.MsgUpdateParams{
				Authority: f.addrs[0].String(),
				Params:    types.DefaultParams(),
			},
			err: true,
		},
		{
			name: "success",
			request: &types.MsgUpdateParams{
				Authority: f.govModAddr,
				Params:    types.DefaultParams(),
			},
			err: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.msgServer.UpdateParams(f.ctx, tc.request)

			if tc.err {
				require.Error(err)
			} else {
				require.NoError(err)

				r, err := f.queryServer.Params(f.ctx, &types.QueryParamsRequest{})
				require.NoError(err)

				require.EqualValues(&tc.request.Params, r.Params)
			}
		})
	}
}

func TestInitiateDomainVerification(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	testCases := []struct {
		name        string
		domain      string
		creator     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "success - valid domain",
			domain:      "example.com",
			creator:     f.addrs[0].String(),
			expectError: false,
		},
		{
			name:        "success - subdomain",
			domain:      "api.example.com",
			creator:     f.addrs[0].String(),
			expectError: false,
		},
		{
			name:        "fail - empty domain",
			domain:      "",
			creator:     f.addrs[0].String(),
			expectError: true,
			errorMsg:    "domain cannot be empty",
		},
		{
			name:        "fail - invalid domain format",
			domain:      "invalid domain with spaces",
			creator:     f.addrs[0].String(),
			expectError: true,
			errorMsg:    "invalid domain format",
		},
		{
			name:        "fail - domain without dot",
			domain:      "localhost",
			creator:     f.addrs[0].String(),
			expectError: true,
			errorMsg:    "must contain at least one dot",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := &types.MsgInitiateDomainVerification{
				Creator: tc.creator,
				Domain:  tc.domain,
			}

			resp, err := f.msgServer.InitiateDomainVerification(f.ctx, msg)

			if tc.expectError {
				require.Error(err)
				require.Contains(err.Error(), tc.errorMsg)
				require.Nil(resp)
			} else {
				require.NoError(err)
				require.NotNil(resp)
				require.NotEmpty(resp.VerificationToken)
				require.Contains(resp.DnsInstruction, tc.domain)
				require.Contains(resp.DnsInstruction, "sonr-verification=")

				// Verify the domain verification was stored
				verification, err := f.k.GetDomainVerification(f.ctx, tc.domain)
				require.NoError(err)
				require.Equal(tc.domain, verification.Domain)
				require.Equal(tc.creator, verification.Owner)
				require.Equal(resp.VerificationToken, verification.VerificationToken)
				require.Equal(v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_PENDING, verification.Status)
				require.Greater(verification.ExpiresAt, time.Now().Unix())
			}
		})
	}
}

func TestInitiateDomainVerification_Duplicate(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	domain := "test.example.com"
	creator := f.addrs[0].String()

	// First verification should succeed
	msg := &types.MsgInitiateDomainVerification{
		Creator: creator,
		Domain:  domain,
	}

	resp1, err := f.msgServer.InitiateDomainVerification(f.ctx, msg)
	require.NoError(err)
	require.NotNil(resp1)

	// Second verification attempt should return error (already exists and valid)
	resp2, err := f.msgServer.InitiateDomainVerification(f.ctx, msg)
	require.Error(err)
	require.Contains(err.Error(), "already exists and is valid")
	require.Nil(resp2)
}

func TestVerifyDomain(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	domain := "verify.example.com"
	creator := f.addrs[0].String()

	// First initiate domain verification
	initMsg := &types.MsgInitiateDomainVerification{
		Creator: creator,
		Domain:  domain,
	}
	_, err := f.msgServer.InitiateDomainVerification(f.ctx, initMsg)
	require.NoError(err)

	testCases := []struct {
		name           string
		domain         string
		creator        string
		expectVerified bool
		expectError    bool
	}{
		{
			name:           "verify existing domain - will fail DNS lookup",
			domain:         domain,
			creator:        creator,
			expectVerified: false,
			expectError:    false, // Error returned in response, not as gRPC error
		},
		{
			name:        "verify non-existent domain",
			domain:      "nonexistent.example.com",
			creator:     creator,
			expectError: false, // Will return "not found" in response
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := &types.MsgVerifyDomain{
				Creator: tc.creator,
				Domain:  tc.domain,
			}

			resp, err := f.msgServer.VerifyDomain(f.ctx, msg)

			if tc.expectError {
				require.Error(err)
				require.Nil(resp)
			} else {
				require.NoError(err)
				require.NotNil(resp)
				require.Equal(tc.expectVerified, resp.Verified)
				require.NotEmpty(resp.Message)

				if tc.domain == domain {
					// For the initiated domain, check the verification status was updated
					verification, err := f.k.GetDomainVerification(f.ctx, tc.domain)
					require.NoError(err)
					if resp.Verified {
						require.Equal(v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED, verification.Status)
						require.Greater(verification.VerifiedAt, int64(0))
					} else {
						require.Equal(v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_FAILED, verification.Status)
					}
				}
			}
		})
	}
}

func TestRegisterService(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	// Setup: Create and manually verify a domain for testing
	domain := "service.example.com"
	creator := f.addrs[0].String()

	// Insert a verified domain verification record directly
	verification := &v1.DomainVerification{
		Domain:            domain,
		Owner:             creator,
		VerificationToken: "test-token-12345",
		Status:            v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED,
		ExpiresAt:         time.Now().Unix() + 3600,
		VerifiedAt:        time.Now().Unix(),
	}
	err := f.k.OrmDB.DomainVerificationTable().Insert(f.ctx, verification)
	require.NoError(err)

	testCases := []struct {
		name        string
		serviceId   string
		domain      string
		creator     string
		permissions []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "success - register service with verified domain",
			serviceId:   "test-service-1",
			domain:      domain,
			creator:     creator,
			permissions: []string{"register", "update"},
			expectError: false,
		},
		{
			name:        "fail - empty service ID",
			serviceId:   "",
			domain:      domain,
			creator:     creator,
			permissions: []string{"register"},
			expectError: true,
			errorMsg:    "service ID cannot be empty",
		},
		{
			name:        "fail - unverified domain",
			serviceId:   "test-service-2",
			domain:      "unverified.example.com",
			creator:     creator,
			permissions: []string{"register"},
			expectError: true,
			errorMsg:    "domain is not verified",
		},
		{
			name:        "fail - duplicate service ID",
			serviceId:   "test-service-1", // Same as first successful test
			domain:      domain,
			creator:     creator,
			permissions: []string{"register"},
			expectError: true,
			errorMsg:    "service already exists",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := &types.MsgRegisterService{
				Creator:              tc.creator,
				ServiceId:            tc.serviceId,
				Domain:               tc.domain,
				RequestedPermissions: tc.permissions,
				UcanDelegationChain:  "", // Empty for testing - UCAN validation optional
			}

			resp, err := f.msgServer.RegisterService(f.ctx, msg)

			if tc.expectError {
				require.Error(err)
				require.Contains(err.Error(), tc.errorMsg)
				require.Nil(resp)
			} else {
				require.NoError(err)
				require.NotNil(resp)
				require.Equal(tc.serviceId, resp.ServiceId)
				require.NotEmpty(resp.RootCapabilityCid)

				// Verify the service was stored
				service, err := f.k.OrmDB.ServiceTable().Get(f.ctx, tc.serviceId)
				require.NoError(err)
				require.Equal(tc.serviceId, service.Id)
				require.Equal(tc.domain, service.Domain)
				require.Equal(tc.creator, service.Owner)
				require.Equal(tc.permissions, service.Permissions)
				require.Equal(v1.ServiceStatus_SERVICE_STATUS_ACTIVE, service.Status)
				require.Greater(service.CreatedAt, int64(0))
				require.Greater(service.UpdatedAt, int64(0))
			}
		})
	}
}

func TestRegisterService_DomainAlreadyBound(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	// Setup: Create and verify a domain
	domain := "bound.example.com"
	creator := f.addrs[0].String()

	verification := &v1.DomainVerification{
		Domain:            domain,
		Owner:             creator,
		VerificationToken: "test-token-bound",
		Status:            v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED,
		ExpiresAt:         time.Now().Unix() + 3600,
		VerifiedAt:        time.Now().Unix(),
	}
	err := f.k.OrmDB.DomainVerificationTable().Insert(f.ctx, verification)
	require.NoError(err)

	// Register first service successfully
	msg1 := &types.MsgRegisterService{
		Creator:              creator,
		ServiceId:            "service-1",
		Domain:               domain,
		RequestedPermissions: []string{"register"},
		UcanDelegationChain:  "", // Empty for testing - UCAN validation optional
	}

	resp1, err := f.msgServer.RegisterService(f.ctx, msg1)
	require.NoError(err)
	require.NotNil(resp1)

	// Try to register second service with same domain - should fail
	msg2 := &types.MsgRegisterService{
		Creator:              creator,
		ServiceId:            "service-2",
		Domain:               domain, // Same domain
		RequestedPermissions: []string{"update"},
		UcanDelegationChain:  "", // Empty for testing - UCAN validation optional
	}

	resp2, err := f.msgServer.RegisterService(f.ctx, msg2)
	require.Error(err)
	require.Contains(err.Error(), "domain is already bound to another service")
	require.Nil(resp2)
}

func TestDomainVerificationWorkflow(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	domain := "workflow.example.com"
	creator := f.addrs[0].String()

	// Step 1: Initiate domain verification
	initiateMsg := &types.MsgInitiateDomainVerification{
		Creator: creator,
		Domain:  domain,
	}

	initiateResp, err := f.msgServer.InitiateDomainVerification(f.ctx, initiateMsg)
	require.NoError(err)
	require.NotNil(initiateResp)
	require.NotEmpty(initiateResp.VerificationToken)

	// Step 2: Try to register service before verification - should fail
	registerMsg := &types.MsgRegisterService{
		Creator:              creator,
		ServiceId:            "workflow-service",
		Domain:               domain,
		RequestedPermissions: []string{"register"},
		UcanDelegationChain:  "", // Empty for testing - UCAN validation optional
	}

	_, err = f.msgServer.RegisterService(f.ctx, registerMsg)
	require.Error(err)
	require.Contains(err.Error(), "domain is not verified")

	// Step 3: Manually mark domain as verified (simulating successful DNS verification)
	verification, err := f.k.GetDomainVerification(f.ctx, domain)
	require.NoError(err)
	verification.Status = v1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED
	verification.VerifiedAt = time.Now().Unix()
	err = f.k.OrmDB.DomainVerificationTable().Update(f.ctx, verification)
	require.NoError(err)

	// Step 4: Now service registration should succeed
	registerResp, err := f.msgServer.RegisterService(f.ctx, registerMsg)
	require.NoError(err)
	require.NotNil(registerResp)
	require.Equal("workflow-service", registerResp.ServiceId)

	// Step 5: Verify the complete workflow
	service, err := f.k.OrmDB.ServiceTable().Get(f.ctx, "workflow-service")
	require.NoError(err)
	require.Equal(domain, service.Domain)
	require.Equal(creator, service.Owner)
}
