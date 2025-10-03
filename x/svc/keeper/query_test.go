package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/x/svc/types"
)

func TestQueryDomainVerification(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	// First create a domain verification
	_, err := f.k.InitiateDomainVerification(f.ctx, "example.com", "idx1test")
	require.NoError(err)

	// Query the domain verification
	resp, err := f.queryServer.DomainVerification(f.ctx, &types.QueryDomainVerificationRequest{
		Domain: "example.com",
	})
	require.NoError(err)
	require.NotNil(resp.DomainVerification)
	require.Equal("example.com", resp.DomainVerification.Domain)
	require.Equal("idx1test", resp.DomainVerification.Owner)
	require.NotEmpty(resp.DomainVerification.VerificationToken)
}

func TestQueryService(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	// First register a service (need domain verified first)
	_, err := f.k.InitiateDomainVerification(f.ctx, "example.com", "idx1test")
	require.NoError(err)

	err = f.k.SetDomainVerified(f.ctx, "example.com")
	require.NoError(err)

	registerResp, err := f.msgServer.RegisterService(f.ctx, &types.MsgRegisterService{
		Creator:              "idx1test",
		ServiceId:            "test-service",
		Domain:               "example.com",
		RequestedPermissions: []string{"register", "update"},
		UcanDelegationChain:  "",
	})
	require.NoError(err)
	require.NotNil(registerResp)

	// Query the service
	resp, err := f.queryServer.Service(f.ctx, &types.QueryServiceRequest{
		ServiceId: "test-service",
	})
	require.NoError(err)
	require.NotNil(resp.Service)
	require.Equal("test-service", resp.Service.Id)
	require.Equal("example.com", resp.Service.Domain)
	require.Equal("idx1test", resp.Service.Owner)
	require.Contains(resp.Service.Permissions, "register")
	require.Contains(resp.Service.Permissions, "update")
}

func TestQueryServicesByOwner(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	// Setup verified domain and register multiple services
	_, err := f.k.InitiateDomainVerification(f.ctx, "example.com", "idx1test")
	require.NoError(err)
	err = f.k.SetDomainVerified(f.ctx, "example.com")
	require.NoError(err)

	_, err = f.k.InitiateDomainVerification(f.ctx, "test.org", "idx1test")
	require.NoError(err)
	err = f.k.SetDomainVerified(f.ctx, "test.org")
	require.NoError(err)

	// Register first service
	_, err = f.msgServer.RegisterService(f.ctx, &types.MsgRegisterService{
		Creator:              "idx1test",
		ServiceId:            "service1",
		Domain:               "example.com",
		RequestedPermissions: []string{"register"},
		UcanDelegationChain:  "",
	})
	require.NoError(err)

	// Register second service
	_, err = f.msgServer.RegisterService(f.ctx, &types.MsgRegisterService{
		Creator:              "idx1test",
		ServiceId:            "service2",
		Domain:               "test.org",
		RequestedPermissions: []string{"register", "update"},
		UcanDelegationChain:  "",
	})
	require.NoError(err)

	// Query services by owner
	resp, err := f.queryServer.ServicesByOwner(f.ctx, &types.QueryServicesByOwnerRequest{
		Owner: "idx1test",
	})
	require.NoError(err)
	require.Len(resp.Services, 2)

	// Check that both services are returned
	serviceIds := make([]string, len(resp.Services))
	for i, service := range resp.Services {
		serviceIds[i] = service.Id
		require.Equal("idx1test", service.Owner)
	}
	require.Contains(serviceIds, "service1")
	require.Contains(serviceIds, "service2")
}

func TestQueryServicesByDomain(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	// Setup verified domain
	_, err := f.k.InitiateDomainVerification(f.ctx, "example.com", "idx1test")
	require.NoError(err)
	err = f.k.SetDomainVerified(f.ctx, "example.com")
	require.NoError(err)

	// Register service for the domain
	_, err = f.msgServer.RegisterService(f.ctx, &types.MsgRegisterService{
		Creator:              "idx1test",
		ServiceId:            "domain-service",
		Domain:               "example.com",
		RequestedPermissions: []string{"register"},
		UcanDelegationChain:  "",
	})
	require.NoError(err)

	// Query services by domain
	resp, err := f.queryServer.ServicesByDomain(f.ctx, &types.QueryServicesByDomainRequest{
		Domain: "example.com",
	})
	require.NoError(err)
	require.Len(resp.Services, 1)
	require.Equal("domain-service", resp.Services[0].Id)
	require.Equal("example.com", resp.Services[0].Domain)
}

func TestQueryErrors(t *testing.T) {
	f := SetupTest(t)
	require := require.New(t)

	// Test empty domain
	_, err := f.queryServer.DomainVerification(f.ctx, &types.QueryDomainVerificationRequest{
		Domain: "",
	})
	require.Error(err)
	require.Contains(err.Error(), "domain cannot be empty")

	// Test empty service ID
	_, err = f.queryServer.Service(f.ctx, &types.QueryServiceRequest{
		ServiceId: "",
	})
	require.Error(err)
	require.Contains(err.Error(), "service_id cannot be empty")

	// Test empty owner
	_, err = f.queryServer.ServicesByOwner(f.ctx, &types.QueryServicesByOwnerRequest{
		Owner: "",
	})
	require.Error(err)
	require.Contains(err.Error(), "owner cannot be empty")

	// Test empty domain for services query
	_, err = f.queryServer.ServicesByDomain(f.ctx, &types.QueryServicesByDomainRequest{
		Domain: "",
	})
	require.Error(err)
	require.Contains(err.Error(), "domain cannot be empty")

	// Test non-existent domain
	_, err = f.queryServer.DomainVerification(f.ctx, &types.QueryDomainVerificationRequest{
		Domain: "nonexistent.com",
	})
	require.Error(err)
	require.Contains(err.Error(), "domain verification not found")

	// Test non-existent service
	_, err = f.queryServer.Service(f.ctx, &types.QueryServiceRequest{
		ServiceId: "nonexistent-service",
	})
	require.Error(err)
	require.Contains(err.Error(), "service not found")
}
