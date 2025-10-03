package types

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return Params{
		// Service Limits
		MaxServicesPerAccount:  10,
		MaxDomainsPerService:   5,
		MaxEndpointsPerService: 20,

		// Timeouts and Intervals (in seconds)
		DomainVerificationTimeout:   86400,   // 24 hours
		ServiceHealthCheckInterval:  300,     // 5 minutes
		CapabilityDefaultExpiration: 2592000, // 30 days

		// Economic Parameters
		ServiceRegistrationFee: sdk.NewInt64Coin("usnr", 1000),
		DomainVerificationFee:  sdk.NewInt64Coin("usnr", 500),
		MinServiceStake:        sdk.NewInt64Coin("usnr", 10000),

		// UCAN and Capability Settings
		MaxDelegationChainDepth: 5,
		UcanMaxLifetime:         31536000, // 1 year maximum
		UcanMinLifetime:         60,       // 1 minute minimum
		SupportedSignatureAlgorithms: []string{
			"ES256", // ECDSA with P-256
			"RS256", // RSA with SHA-256
			"EdDSA", // EdDSA signatures
		},

		// Validation Rules
		RequireDomainOwnershipProof: true,
		RequireHttps:                false, // Allow HTTP for development
		AllowLocalhost:              true,  // Development support
		MaxServiceDescriptionLength: 1024,

		// Rate Limiting
		MaxRegistrationsPerBlock:    10,
		MaxUpdatesPerBlock:          50,
		MaxCapabilityGrantsPerBlock: 100,
	}
}

// Stringer method for Params.
func (p Params) String() string {
	bz, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return string(bz)
}

// Validate does the sanity check on the params.
func (p Params) Validate() error {
	// Validate service limits
	if err := validateServiceLimits(p); err != nil {
		return err
	}

	// Validate timeouts
	if err := validateTimeouts(p); err != nil {
		return err
	}

	// Validate economic parameters
	if err := validateEconomicParams(p); err != nil {
		return err
	}

	// Validate UCAN parameters
	if err := validateUCANParams(p); err != nil {
		return err
	}

	// Validate rate limiting
	if err := validateRateLimits(p); err != nil {
		return err
	}

	// Validate other parameters
	if err := validateOtherParams(p); err != nil {
		return err
	}

	return nil
}

// validateServiceLimits validates service-related limits
func validateServiceLimits(p Params) error {
	if p.MaxServicesPerAccount == 0 || p.MaxServicesPerAccount > 100 {
		return fmt.Errorf(
			"max_services_per_account must be between 1 and 100, got %d",
			p.MaxServicesPerAccount,
		)
	}

	if p.MaxDomainsPerService == 0 || p.MaxDomainsPerService > 20 {
		return fmt.Errorf(
			"max_domains_per_service must be between 1 and 20, got %d",
			p.MaxDomainsPerService,
		)
	}

	if p.MaxEndpointsPerService == 0 || p.MaxEndpointsPerService > 100 {
		return fmt.Errorf(
			"max_endpoints_per_service must be between 1 and 100, got %d",
			p.MaxEndpointsPerService,
		)
	}

	return nil
}

// validateTimeouts validates timeout parameters
func validateTimeouts(p Params) error {
	// Domain verification timeout: 1 hour to 7 days
	if p.DomainVerificationTimeout < 3600 || p.DomainVerificationTimeout > 604800 {
		return fmt.Errorf(
			"domain_verification_timeout must be between 3600 and 604800 seconds, got %d",
			p.DomainVerificationTimeout,
		)
	}

	// Service health check interval: 1 minute to 1 hour
	if p.ServiceHealthCheckInterval < 60 || p.ServiceHealthCheckInterval > 3600 {
		return fmt.Errorf(
			"service_health_check_interval must be between 60 and 3600 seconds, got %d",
			p.ServiceHealthCheckInterval,
		)
	}

	// Capability default expiration: 1 hour to 1 year
	if p.CapabilityDefaultExpiration < 3600 || p.CapabilityDefaultExpiration > 31536000 {
		return fmt.Errorf(
			"capability_default_expiration must be between 3600 and 31536000 seconds, got %d",
			p.CapabilityDefaultExpiration,
		)
	}

	return nil
}

// validateEconomicParams validates economic-related parameters
func validateEconomicParams(p Params) error {
	// Validate service registration fee
	if p.ServiceRegistrationFee.IsNegative() {
		return fmt.Errorf("service_registration_fee cannot be negative")
	}
	if p.ServiceRegistrationFee.Denom != "usnr" {
		return fmt.Errorf(
			"service_registration_fee must use usnr denomination, got %s",
			p.ServiceRegistrationFee.Denom,
		)
	}
	// Upper bound: 1M usnr
	if p.ServiceRegistrationFee.Amount.GT(math.NewInt(1000000)) {
		return fmt.Errorf("service_registration_fee exceeds maximum of 1000000 usnr")
	}

	// Validate domain verification fee
	if p.DomainVerificationFee.IsNegative() {
		return fmt.Errorf("domain_verification_fee cannot be negative")
	}
	if p.DomainVerificationFee.Denom != "usnr" {
		return fmt.Errorf(
			"domain_verification_fee must use usnr denomination, got %s",
			p.DomainVerificationFee.Denom,
		)
	}
	// Should be less than or equal to service registration fee
	if p.DomainVerificationFee.Amount.GT(p.ServiceRegistrationFee.Amount) {
		return fmt.Errorf("domain_verification_fee should not exceed service_registration_fee")
	}

	// Validate minimum service stake
	if !p.MinServiceStake.IsPositive() {
		return fmt.Errorf("min_service_stake must be positive")
	}
	if p.MinServiceStake.Denom != "usnr" {
		return fmt.Errorf(
			"min_service_stake must use usnr denomination, got %s",
			p.MinServiceStake.Denom,
		)
	}
	// Should be greater than registration fee
	if p.MinServiceStake.Amount.LTE(p.ServiceRegistrationFee.Amount) {
		return fmt.Errorf("min_service_stake should be greater than service_registration_fee")
	}

	return nil
}

// validateUCANParams validates UCAN-related parameters
func validateUCANParams(p Params) error {
	// Max delegation chain depth: 1-10
	if p.MaxDelegationChainDepth == 0 || p.MaxDelegationChainDepth > 10 {
		return fmt.Errorf(
			"max_delegation_chain_depth must be between 1 and 10, got %d",
			p.MaxDelegationChainDepth,
		)
	}

	// UCAN lifetime: min 1 minute, max 10 years
	if p.UcanMaxLifetime < 60 || p.UcanMaxLifetime > 315360000 {
		return fmt.Errorf(
			"ucan_max_lifetime must be between 60 and 315360000 seconds, got %d",
			p.UcanMaxLifetime,
		)
	}

	if p.UcanMinLifetime < 1 || p.UcanMinLifetime >= p.UcanMaxLifetime {
		return fmt.Errorf(
			"ucan_min_lifetime must be positive and less than ucan_max_lifetime, got %d",
			p.UcanMinLifetime,
		)
	}

	// Validate signature algorithms
	if len(p.SupportedSignatureAlgorithms) == 0 {
		return fmt.Errorf("at least one signature algorithm must be supported")
	}

	validAlgorithms := map[string]bool{
		"ES256": true, // ECDSA with P-256
		"ES384": true, // ECDSA with P-384
		"ES512": true, // ECDSA with P-521
		"RS256": true, // RSA with SHA-256
		"RS384": true, // RSA with SHA-384
		"RS512": true, // RSA with SHA-512
		"EdDSA": true, // EdDSA signatures
	}

	for _, algo := range p.SupportedSignatureAlgorithms {
		if !validAlgorithms[algo] {
			return fmt.Errorf("unsupported signature algorithm: %s", algo)
		}
	}

	return nil
}

// validateRateLimits validates rate limiting parameters
func validateRateLimits(p Params) error {
	// Max registrations per block: 1-100
	if p.MaxRegistrationsPerBlock == 0 || p.MaxRegistrationsPerBlock > 100 {
		return fmt.Errorf(
			"max_registrations_per_block must be between 1 and 100, got %d",
			p.MaxRegistrationsPerBlock,
		)
	}

	// Max updates per block: 1-1000
	if p.MaxUpdatesPerBlock == 0 || p.MaxUpdatesPerBlock > 1000 {
		return fmt.Errorf(
			"max_updates_per_block must be between 1 and 1000, got %d",
			p.MaxUpdatesPerBlock,
		)
	}

	// Max capability grants per block: 1-500
	if p.MaxCapabilityGrantsPerBlock == 0 || p.MaxCapabilityGrantsPerBlock > 500 {
		return fmt.Errorf(
			"max_capability_grants_per_block must be between 1 and 500, got %d",
			p.MaxCapabilityGrantsPerBlock,
		)
	}

	return nil
}

// validateOtherParams validates miscellaneous parameters
func validateOtherParams(p Params) error {
	// Max service description length: 10-10000 characters
	if p.MaxServiceDescriptionLength < 10 || p.MaxServiceDescriptionLength > 10000 {
		return fmt.Errorf(
			"max_service_description_length must be between 10 and 10000, got %d",
			p.MaxServiceDescriptionLength,
		)
	}

	// Validation rules are boolean, no special validation needed
	// but we can add logical checks
	if !p.AllowLocalhost && !p.RequireHttps {
		// This is a valid configuration for production
	}

	return nil
}
