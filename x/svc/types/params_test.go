package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/x/svc/types"
)

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()

	// Test default values - Service Limits
	require.Equal(t, uint32(10), params.MaxServicesPerAccount)
	require.Equal(t, uint32(5), params.MaxDomainsPerService)
	require.Equal(t, uint32(20), params.MaxEndpointsPerService)

	// Timeouts and Intervals
	require.Equal(t, int64(86400), params.DomainVerificationTimeout)
	require.Equal(t, int64(300), params.ServiceHealthCheckInterval)
	require.Equal(t, int64(2592000), params.CapabilityDefaultExpiration)

	// Economic Parameters
	require.Equal(t, sdk.NewInt64Coin("usnr", 1000), params.ServiceRegistrationFee)
	require.Equal(t, sdk.NewInt64Coin("usnr", 500), params.DomainVerificationFee)
	require.Equal(t, sdk.NewInt64Coin("usnr", 10000), params.MinServiceStake)

	// UCAN and Capability Settings
	require.Equal(t, uint32(5), params.MaxDelegationChainDepth)
	require.Equal(t, int64(31536000), params.UcanMaxLifetime)
	require.Equal(t, int64(60), params.UcanMinLifetime)
	require.Equal(t, []string{"ES256", "RS256", "EdDSA"}, params.SupportedSignatureAlgorithms)

	// Validation Rules
	require.True(t, params.RequireDomainOwnershipProof)
	require.False(t, params.RequireHttps)
	require.True(t, params.AllowLocalhost)
	require.Equal(t, uint32(1024), params.MaxServiceDescriptionLength)

	// Rate Limiting
	require.Equal(t, uint32(10), params.MaxRegistrationsPerBlock)
	require.Equal(t, uint32(50), params.MaxUpdatesPerBlock)
	require.Equal(t, uint32(100), params.MaxCapabilityGrantsPerBlock)
}

func TestParams_Validate(t *testing.T) {
	tests := []struct {
		name        string
		modifyFunc  func(*types.Params)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "default params are valid",
			modifyFunc:  func(p *types.Params) {},
			expectError: false,
		},
		// Service Limits Tests
		{
			name: "zero max services per account",
			modifyFunc: func(p *types.Params) {
				p.MaxServicesPerAccount = 0
			},
			expectError: true,
			errorMsg:    "max_services_per_account must be between 1 and 100",
		},
		{
			name: "excessive max services per account",
			modifyFunc: func(p *types.Params) {
				p.MaxServicesPerAccount = 101
			},
			expectError: true,
			errorMsg:    "max_services_per_account must be between 1 and 100",
		},
		{
			name: "zero max domains per service",
			modifyFunc: func(p *types.Params) {
				p.MaxDomainsPerService = 0
			},
			expectError: true,
			errorMsg:    "max_domains_per_service must be between 1 and 20",
		},
		{
			name: "excessive max domains per service",
			modifyFunc: func(p *types.Params) {
				p.MaxDomainsPerService = 21
			},
			expectError: true,
			errorMsg:    "max_domains_per_service must be between 1 and 20",
		},
		{
			name: "zero max endpoints per service",
			modifyFunc: func(p *types.Params) {
				p.MaxEndpointsPerService = 0
			},
			expectError: true,
			errorMsg:    "max_endpoints_per_service must be between 1 and 100",
		},
		{
			name: "excessive max endpoints per service",
			modifyFunc: func(p *types.Params) {
				p.MaxEndpointsPerService = 101
			},
			expectError: true,
			errorMsg:    "max_endpoints_per_service must be between 1 and 100",
		},
		// Timeout Tests
		{
			name: "domain verification timeout too low",
			modifyFunc: func(p *types.Params) {
				p.DomainVerificationTimeout = 3599
			},
			expectError: true,
			errorMsg:    "domain_verification_timeout must be between 3600 and 604800 seconds",
		},
		{
			name: "domain verification timeout too high",
			modifyFunc: func(p *types.Params) {
				p.DomainVerificationTimeout = 604801
			},
			expectError: true,
			errorMsg:    "domain_verification_timeout must be between 3600 and 604800 seconds",
		},
		{
			name: "service health check interval too low",
			modifyFunc: func(p *types.Params) {
				p.ServiceHealthCheckInterval = 59
			},
			expectError: true,
			errorMsg:    "service_health_check_interval must be between 60 and 3600 seconds",
		},
		{
			name: "service health check interval too high",
			modifyFunc: func(p *types.Params) {
				p.ServiceHealthCheckInterval = 3601
			},
			expectError: true,
			errorMsg:    "service_health_check_interval must be between 60 and 3600 seconds",
		},
		{
			name: "capability expiration too low",
			modifyFunc: func(p *types.Params) {
				p.CapabilityDefaultExpiration = 3599
			},
			expectError: true,
			errorMsg:    "capability_default_expiration must be between 3600 and 31536000 seconds",
		},
		{
			name: "capability expiration too high",
			modifyFunc: func(p *types.Params) {
				p.CapabilityDefaultExpiration = 31536001
			},
			expectError: true,
			errorMsg:    "capability_default_expiration must be between 3600 and 31536000 seconds",
		},
		// Economic Parameters Tests
		{
			name: "negative service registration fee",
			modifyFunc: func(p *types.Params) {
				p.ServiceRegistrationFee = sdk.Coin{Denom: "usnr", Amount: math.NewInt(-1)}
			},
			expectError: true,
			errorMsg:    "service_registration_fee cannot be negative",
		},
		{
			name: "wrong denom for service registration fee",
			modifyFunc: func(p *types.Params) {
				p.ServiceRegistrationFee = sdk.NewInt64Coin("snr", 1000)
			},
			expectError: true,
			errorMsg:    "service_registration_fee must use usnr denomination",
		},
		{
			name: "excessive service registration fee",
			modifyFunc: func(p *types.Params) {
				p.ServiceRegistrationFee = sdk.NewInt64Coin("usnr", 1000001)
			},
			expectError: true,
			errorMsg:    "service_registration_fee exceeds maximum of 1000000 usnr",
		},
		{
			name: "negative domain verification fee",
			modifyFunc: func(p *types.Params) {
				p.DomainVerificationFee = sdk.Coin{Denom: "usnr", Amount: math.NewInt(-1)}
			},
			expectError: true,
			errorMsg:    "domain_verification_fee cannot be negative",
		},
		{
			name: "domain fee exceeds service fee",
			modifyFunc: func(p *types.Params) {
				p.ServiceRegistrationFee = sdk.NewInt64Coin("usnr", 100)
				p.DomainVerificationFee = sdk.NewInt64Coin("usnr", 200)
			},
			expectError: true,
			errorMsg:    "domain_verification_fee should not exceed service_registration_fee",
		},
		{
			name: "zero min service stake",
			modifyFunc: func(p *types.Params) {
				p.MinServiceStake = sdk.NewInt64Coin("usnr", 0)
			},
			expectError: true,
			errorMsg:    "min_service_stake must be positive",
		},
		{
			name: "min stake less than registration fee",
			modifyFunc: func(p *types.Params) {
				p.ServiceRegistrationFee = sdk.NewInt64Coin("usnr", 1000)
				p.MinServiceStake = sdk.NewInt64Coin("usnr", 999)
			},
			expectError: true,
			errorMsg:    "min_service_stake should be greater than service_registration_fee",
		},
		// UCAN Parameters Tests
		{
			name: "zero delegation chain depth",
			modifyFunc: func(p *types.Params) {
				p.MaxDelegationChainDepth = 0
			},
			expectError: true,
			errorMsg:    "max_delegation_chain_depth must be between 1 and 10",
		},
		{
			name: "excessive delegation chain depth",
			modifyFunc: func(p *types.Params) {
				p.MaxDelegationChainDepth = 11
			},
			expectError: true,
			errorMsg:    "max_delegation_chain_depth must be between 1 and 10",
		},
		{
			name: "UCAN max lifetime too low",
			modifyFunc: func(p *types.Params) {
				p.UcanMaxLifetime = 59
			},
			expectError: true,
			errorMsg:    "ucan_max_lifetime must be between 60 and 315360000 seconds",
		},
		{
			name: "UCAN max lifetime too high",
			modifyFunc: func(p *types.Params) {
				p.UcanMaxLifetime = 315360001
			},
			expectError: true,
			errorMsg:    "ucan_max_lifetime must be between 60 and 315360000 seconds",
		},
		{
			name: "UCAN min lifetime zero",
			modifyFunc: func(p *types.Params) {
				p.UcanMinLifetime = 0
			},
			expectError: true,
			errorMsg:    "ucan_min_lifetime must be positive and less than ucan_max_lifetime",
		},
		{
			name: "UCAN min lifetime exceeds max",
			modifyFunc: func(p *types.Params) {
				p.UcanMaxLifetime = 100
				p.UcanMinLifetime = 101
			},
			expectError: true,
			errorMsg:    "ucan_min_lifetime must be positive and less than ucan_max_lifetime",
		},
		{
			name: "no signature algorithms",
			modifyFunc: func(p *types.Params) {
				p.SupportedSignatureAlgorithms = []string{}
			},
			expectError: true,
			errorMsg:    "at least one signature algorithm must be supported",
		},
		{
			name: "invalid signature algorithm",
			modifyFunc: func(p *types.Params) {
				p.SupportedSignatureAlgorithms = []string{"INVALID"}
			},
			expectError: true,
			errorMsg:    "unsupported signature algorithm: INVALID",
		},
		// Rate Limiting Tests
		{
			name: "zero max registrations per block",
			modifyFunc: func(p *types.Params) {
				p.MaxRegistrationsPerBlock = 0
			},
			expectError: true,
			errorMsg:    "max_registrations_per_block must be between 1 and 100",
		},
		{
			name: "excessive max registrations per block",
			modifyFunc: func(p *types.Params) {
				p.MaxRegistrationsPerBlock = 101
			},
			expectError: true,
			errorMsg:    "max_registrations_per_block must be between 1 and 100",
		},
		{
			name: "zero max updates per block",
			modifyFunc: func(p *types.Params) {
				p.MaxUpdatesPerBlock = 0
			},
			expectError: true,
			errorMsg:    "max_updates_per_block must be between 1 and 1000",
		},
		{
			name: "excessive max updates per block",
			modifyFunc: func(p *types.Params) {
				p.MaxUpdatesPerBlock = 1001
			},
			expectError: true,
			errorMsg:    "max_updates_per_block must be between 1 and 1000",
		},
		{
			name: "zero max capability grants per block",
			modifyFunc: func(p *types.Params) {
				p.MaxCapabilityGrantsPerBlock = 0
			},
			expectError: true,
			errorMsg:    "max_capability_grants_per_block must be between 1 and 500",
		},
		{
			name: "excessive max capability grants per block",
			modifyFunc: func(p *types.Params) {
				p.MaxCapabilityGrantsPerBlock = 501
			},
			expectError: true,
			errorMsg:    "max_capability_grants_per_block must be between 1 and 500",
		},
		// Other Parameters Tests
		{
			name: "service description length too low",
			modifyFunc: func(p *types.Params) {
				p.MaxServiceDescriptionLength = 9
			},
			expectError: true,
			errorMsg:    "max_service_description_length must be between 10 and 10000",
		},
		{
			name: "service description length too high",
			modifyFunc: func(p *types.Params) {
				p.MaxServiceDescriptionLength = 10001
			},
			expectError: true,
			errorMsg:    "max_service_description_length must be between 10 and 10000",
		},
		// Valid edge cases
		{
			name: "all minimum valid values",
			modifyFunc: func(p *types.Params) {
				p.MaxServicesPerAccount = 1
				p.MaxDomainsPerService = 1
				p.MaxEndpointsPerService = 1
				p.DomainVerificationTimeout = 3600
				p.ServiceHealthCheckInterval = 60
				p.CapabilityDefaultExpiration = 3600
				p.ServiceRegistrationFee = sdk.NewInt64Coin("usnr", 0)
				p.DomainVerificationFee = sdk.NewInt64Coin("usnr", 0)
				p.MinServiceStake = sdk.NewInt64Coin("usnr", 1)
				p.MaxDelegationChainDepth = 1
				p.UcanMaxLifetime = 60
				p.UcanMinLifetime = 1
				p.MaxRegistrationsPerBlock = 1
				p.MaxUpdatesPerBlock = 1
				p.MaxCapabilityGrantsPerBlock = 1
				p.MaxServiceDescriptionLength = 10
			},
			expectError: false,
		},
		{
			name: "all maximum valid values",
			modifyFunc: func(p *types.Params) {
				p.MaxServicesPerAccount = 100
				p.MaxDomainsPerService = 20
				p.MaxEndpointsPerService = 100
				p.DomainVerificationTimeout = 604800
				p.ServiceHealthCheckInterval = 3600
				p.CapabilityDefaultExpiration = 31536000
				p.ServiceRegistrationFee = sdk.NewInt64Coin("usnr", 1000000)
				p.DomainVerificationFee = sdk.NewInt64Coin("usnr", 1000000)
				p.MinServiceStake = sdk.NewInt64Coin("usnr", 1000001)
				p.MaxDelegationChainDepth = 10
				p.UcanMaxLifetime = 315360000
				p.UcanMinLifetime = 315359999
				p.MaxRegistrationsPerBlock = 100
				p.MaxUpdatesPerBlock = 1000
				p.MaxCapabilityGrantsPerBlock = 500
				p.MaxServiceDescriptionLength = 10000
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := types.DefaultParams()
			tc.modifyFunc(&params)

			err := params.Validate()

			if tc.expectError {
				require.Error(t, err)
				if tc.errorMsg != "" {
					require.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestParamsString tests the String method of Params
func TestParamsString(t *testing.T) {
	params := types.DefaultParams()
	str := params.String()

	// Test that string representation contains key fields
	require.Contains(t, str, "max_services_per_account")
	require.Contains(t, str, "max_domains_per_service")
	require.Contains(t, str, "service_registration_fee")
	require.Contains(t, str, "max_delegation_chain_depth")
	require.Contains(t, str, "max_registrations_per_block")
}
