package keeper

import (
	"context"
	"fmt"

	"github.com/sonr-io/sonr/crypto/keys"
	"github.com/sonr-io/sonr/crypto/ucan"
	"github.com/sonr-io/sonr/x/svc/types"
)

// PermissionValidator wraps UCAN verifier for Service-specific permission validation
type PermissionValidator struct {
	verifier    *ucan.Verifier
	keeper      Keeper
	permissions *types.UCANPermissionRegistry
}

// NewPermissionValidator creates a new Service permission validator
func NewPermissionValidator(keeper Keeper) *PermissionValidator {
	didResolver := &ServiceDIDResolver{keeper: keeper}
	verifier := ucan.NewVerifier(didResolver)

	return &PermissionValidator{
		verifier:    verifier,
		keeper:      keeper,
		permissions: types.NewUCANPermissionRegistry(),
	}
}

// ValidatePermission validates UCAN token for Service operation
func (pv *PermissionValidator) ValidatePermission(
	ctx context.Context,
	tokenString string,
	serviceID string,
	operation types.ServiceOperation,
) error {
	// Get required UCAN capabilities for the operation
	capabilities, err := pv.permissions.GetRequiredUCANCapabilities(operation)
	if err != nil {
		return fmt.Errorf("failed to get required UCAN capabilities: %w", err)
	}

	// Build resource URI for Service
	resourceURI := pv.buildResourceURI(serviceID)

	// Verify UCAN token grants required capabilities
	_, err = pv.verifier.VerifyCapability(
		ctx,
		tokenString,
		resourceURI,
		capabilities,
	)
	if err != nil {
		return fmt.Errorf("UCAN validation failed: %w", err)
	}

	return nil
}

// ValidateDomainBoundPermission validates UCAN token for domain-bound operations
func (pv *PermissionValidator) ValidateDomainBoundPermission(
	ctx context.Context,
	tokenString string,
	domain string,
	serviceID string,
	operation types.ServiceOperation,
) error {
	// Get required UCAN capabilities for the operation
	capabilities, err := pv.permissions.GetRequiredUCANCapabilities(operation)
	if err != nil {
		return fmt.Errorf("failed to get required UCAN capabilities: %w", err)
	}

	// Build domain resource URI
	resourceURI := pv.buildDomainResourceURI(domain)

	// Verify UCAN token with domain-bound validation
	token, err := pv.verifier.VerifyCapability(
		ctx,
		tokenString,
		resourceURI,
		capabilities,
	)
	if err != nil {
		return fmt.Errorf("UCAN validation failed: %w", err)
	}

	// Additional domain-bound validation
	if err := pv.validateDomainBoundCaveat(token, domain, serviceID); err != nil {
		return fmt.Errorf("domain-bound validation failed: %w", err)
	}

	return nil
}

// ValidateDomainVerificationPermission validates UCAN token for domain verification
func (pv *PermissionValidator) ValidateDomainVerificationPermission(
	ctx context.Context,
	tokenString string,
	domain string,
	verificationMethod string,
	operation types.ServiceOperation,
) error {
	// Get required UCAN capabilities for the operation
	capabilities, err := pv.permissions.GetRequiredUCANCapabilities(operation)
	if err != nil {
		return fmt.Errorf("failed to get required UCAN capabilities: %w", err)
	}

	// Build domain verification resource URI
	resourceURI := types.CreateDomainVerificationURI(domain, verificationMethod)

	// Verify UCAN token
	token, err := pv.verifier.VerifyCapability(
		ctx,
		tokenString,
		resourceURI,
		capabilities,
	)
	if err != nil {
		return fmt.Errorf("UCAN validation failed: %w", err)
	}

	// Additional domain verification validation
	if err := pv.validateDomainVerification(token, domain, verificationMethod); err != nil {
		return fmt.Errorf("domain verification validation failed: %w", err)
	}

	return nil
}

// VerifyDelegationChain validates complete UCAN delegation chain
func (pv *PermissionValidator) VerifyDelegationChain(
	ctx context.Context,
	tokenString string,
) error {
	return pv.verifier.VerifyDelegationChain(ctx, tokenString)
}

// Internal validation methods

// validateDomainBoundCaveat validates that the token has proper domain binding
func (pv *PermissionValidator) validateDomainBoundCaveat(
	token *ucan.Token,
	domain string,
	serviceID string,
) error {
	// Check each attenuation for domain-bound resources
	for _, att := range token.Attenuations {
		if simpleResource, ok := att.Resource.(*ucan.SimpleResource); ok {
			if simpleResource.Scheme == "domain" && simpleResource.Value == domain {
				// Found matching domain resource
				return nil
			}
		}
	}

	return fmt.Errorf("no matching domain-bound attenuation found for domain %s", domain)
}

// validateDomainVerification validates domain verification capability
func (pv *PermissionValidator) validateDomainVerification(
	token *ucan.Token,
	domain string,
	verificationMethod string,
) error {
	// Find the relevant attenuation for this domain
	for _, att := range token.Attenuations {
		if err := types.ValidateDomainVerificationCapability(att.Capability, domain, verificationMethod); err == nil {
			return nil
		}
	}

	return fmt.Errorf("no valid domain verification capability found for domain %s", domain)
}

// Helper methods

// buildResourceURI constructs Service resource URI
func (pv *PermissionValidator) buildResourceURI(serviceID string) string {
	return fmt.Sprintf("svc:%s", serviceID)
}

// buildDomainResourceURI constructs domain resource URI
func (pv *PermissionValidator) buildDomainResourceURI(domain string) string {
	return fmt.Sprintf("domain:%s", domain)
}

// CreateAttenuation creates a UCAN attenuation for Service operations
func (pv *PermissionValidator) CreateAttenuation(
	actions []string,
	serviceID string,
	caveats []string,
) ucan.Attenuation {
	return pv.permissions.CreateServiceAttenuation(actions, serviceID, caveats)
}

// CreateDomainBoundAttenuation creates a domain-bound UCAN attenuation
func (pv *PermissionValidator) CreateDomainBoundAttenuation(
	actions []string,
	domain string,
	serviceID string,
) ucan.Attenuation {
	return pv.permissions.CreateDomainBoundAttenuation(actions, domain, serviceID)
}

// CreateRateLimitedAttenuation creates a rate-limited UCAN attenuation
func (pv *PermissionValidator) CreateRateLimitedAttenuation(
	actions []string,
	serviceID string,
	rateLimit uint64,
	windowSeconds uint64,
) ucan.Attenuation {
	return pv.permissions.CreateRateLimitedAttenuation(actions, serviceID, rateLimit, windowSeconds)
}

// ServiceDIDResolver implements ucan.DIDResolver for Service module
type ServiceDIDResolver struct {
	keeper Keeper
}

// ResolveDIDKey resolves DID to public key for UCAN verification
func (r *ServiceDIDResolver) ResolveDIDKey(ctx context.Context, did string) (keys.DID, error) {
	// Get the DID document from the keeper
	didDoc, err := r.keeper.didKeeper.GetDIDDocument(ctx, did)
	if err != nil {
		return keys.DID{}, fmt.Errorf("failed to get DID document: %w", err)
	}

	if didDoc == nil {
		return keys.DID{}, types.ErrInvalidOwnerDID
	}

	// Parse the DID string into a keys.DID
	// This assumes the DID keeper can provide the public key information
	return keys.Parse(did)
}

// Gasless transaction support

// SupportsGaslessTransaction checks if a UCAN token supports gasless transactions
func (pv *PermissionValidator) SupportsGaslessTransaction(
	ctx context.Context,
	tokenString string,
	serviceID string,
	operation types.ServiceOperation,
) (bool, uint64, error) {
	// Parse and verify the token
	token, err := pv.verifier.VerifyToken(ctx, tokenString)
	if err != nil {
		return false, 0, fmt.Errorf("token verification failed: %w", err)
	}

	resourceURI := pv.buildResourceURI(serviceID)

	// Check each attenuation for gasless support
	for _, att := range token.Attenuations {
		if att.Resource.GetURI() == resourceURI {
			// Check if capability supports gasless transactions
			if gaslessCapability, ok := att.Capability.(*ucan.GaslessCapability); ok {
				if gaslessCapability.SupportsGasless() {
					// Verify the capability grants the required operation
					capabilities, err := pv.permissions.GetRequiredUCANCapabilities(operation)
					if err != nil {
						continue
					}

					if gaslessCapability.Grants(capabilities) {
						return true, gaslessCapability.GetGasLimit(), nil
					}
				}
			}
		}
	}

	return false, 0, nil
}

// ValidateRateLimit checks if a UCAN token has rate limiting and if it's within limits
func (pv *PermissionValidator) ValidateRateLimit(
	ctx context.Context,
	tokenString string,
	serviceID string,
) (bool, uint64, uint64, error) {
	// Parse and verify the token
	token, err := pv.verifier.VerifyToken(ctx, tokenString)
	if err != nil {
		return false, 0, 0, fmt.Errorf("token verification failed: %w", err)
	}

	resourceURI := pv.buildResourceURI(serviceID)

	// Check each attenuation for rate limiting
	// Rate limiting would typically be implemented with custom capability types
	// or through external state management
	for _, att := range token.Attenuations {
		if att.Resource.GetURI() == resourceURI {
			// Check if this is a gasless capability with limits
			if gaslessCapability, ok := att.Capability.(*ucan.GaslessCapability); ok {
				if gaslessCapability.AllowGasless && gaslessCapability.GasLimit > 0 {
					// Use gas limit as a proxy for rate limiting
					return true, gaslessCapability.GasLimit, 60, nil // 60 second window
				}
			}
		}
	}

	return false, 0, 0, nil
}
