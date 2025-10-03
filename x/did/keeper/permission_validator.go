package keeper

import (
	"context"
	"fmt"

	"github.com/sonr-io/sonr/crypto/keys"
	"github.com/sonr-io/sonr/crypto/ucan"
	"github.com/sonr-io/sonr/x/did/types"
)

// PermissionValidator wraps UCAN verifier for DID-specific permission validation
type PermissionValidator struct {
	verifier    *ucan.Verifier
	keeper      Keeper
	permissions *types.UCANPermissionRegistry
}

// NewPermissionValidator creates a new DID permission validator
func NewPermissionValidator(keeper Keeper) *PermissionValidator {
	didResolver := &DIDKeyResolver{keeper: keeper}
	verifier := ucan.NewVerifier(didResolver)

	return &PermissionValidator{
		verifier:    verifier,
		keeper:      keeper,
		permissions: types.NewUCANPermissionRegistry(),
	}
}

// NewPermissionValidatorWithVerifier creates a new DID permission validator with custom verifier (for testing)
func NewPermissionValidatorWithVerifier(
	keeper Keeper,
	verifier *ucan.Verifier,
) *PermissionValidator {
	return &PermissionValidator{
		verifier:    verifier,
		keeper:      keeper,
		permissions: types.NewUCANPermissionRegistry(),
	}
}

// ValidatePermission validates UCAN token for DID operation
func (pv *PermissionValidator) ValidatePermission(
	ctx context.Context,
	tokenString string,
	did string,
	operation types.DIDOperation,
) error {
	// Get required UCAN capabilities for the operation
	capabilities, err := pv.permissions.GetRequiredUCANCapabilities(operation)
	if err != nil {
		return fmt.Errorf("failed to get required UCAN capabilities: %w", err)
	}

	// Build resource URI for DID
	resourceURI := pv.buildResourceURI(did)

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

// ValidateControllerPermission validates UCAN token for controller-specific DID operations
func (pv *PermissionValidator) ValidateControllerPermission(
	ctx context.Context,
	tokenString string,
	did string,
	controllerAddress string,
	operation types.DIDOperation,
) error {
	// Get required UCAN capabilities for the operation
	capabilities, err := pv.permissions.GetRequiredUCANCapabilities(operation)
	if err != nil {
		return fmt.Errorf("failed to get required UCAN capabilities: %w", err)
	}

	// Build resource URI for DID
	resourceURI := pv.buildResourceURI(did)

	// Verify UCAN token with controller caveat validation
	token, err := pv.verifier.VerifyCapability(
		ctx,
		tokenString,
		resourceURI,
		capabilities,
	)
	if err != nil {
		return fmt.Errorf("UCAN validation failed: %w", err)
	}

	// Additional controller validation
	if err := pv.validateControllerCaveat(token, did, controllerAddress); err != nil {
		return fmt.Errorf("controller validation failed: %w", err)
	}

	return nil
}

// ValidateWebAuthnDelegation validates UCAN token for WebAuthn-delegated operations
func (pv *PermissionValidator) ValidateWebAuthnDelegation(
	ctx context.Context,
	tokenString string,
	did string,
	credentialID string,
	operation types.DIDOperation,
) error {
	// Get required UCAN capabilities for the operation
	capabilities, err := pv.permissions.GetRequiredUCANCapabilities(operation)
	if err != nil {
		return fmt.Errorf("failed to get required UCAN capabilities: %w", err)
	}

	// Build resource URI for DID
	resourceURI := pv.buildResourceURI(did)

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

	// Additional WebAuthn validation
	if err := pv.validateWebAuthnDelegation(token, did, credentialID); err != nil {
		return fmt.Errorf("WebAuthn delegation validation failed: %w", err)
	}

	return nil
}

// ValidateCredentialOperation validates UCAN token for credential operations
func (pv *PermissionValidator) ValidateCredentialOperation(
	ctx context.Context,
	tokenString string,
	issuerDID string,
	subjectDID string,
	operation types.DIDOperation,
) error {
	// For credential operations, validate against issuer DID
	return pv.ValidatePermission(ctx, tokenString, issuerDID, operation)
}

// VerifyDelegationChain validates complete UCAN delegation chain
func (pv *PermissionValidator) VerifyDelegationChain(
	ctx context.Context,
	tokenString string,
) error {
	return pv.verifier.VerifyDelegationChain(ctx, tokenString)
}

// Internal validation methods

// validateControllerCaveat validates that the token has proper controller authorization
func (pv *PermissionValidator) validateControllerCaveat(
	token *ucan.Token,
	did string,
	controllerAddress string,
) error {
	// Check each attenuation for controller caveats
	for _, att := range token.Attenuations {
		if att.Resource.GetURI() == pv.buildResourceURI(did) {
			// Check if this is a DID capability with controller caveat
			if didCapability, ok := att.Capability.(*ucan.DIDCapability); ok {
				return pv.validateDIDControllerCaveat(didCapability, controllerAddress)
			}
		}
	}

	// If no specific controller caveat found, check if token issuer is the controller
	return pv.validateTokenIssuerAsController(token, controllerAddress)
}

// validateDIDControllerCaveat validates controller-specific DID capability caveats
func (pv *PermissionValidator) validateDIDControllerCaveat(
	capability *ucan.DIDCapability,
	controllerAddress string,
) error {
	// Check for controller caveat
	hasControllerCaveat := false
	for _, caveat := range capability.Caveats {
		if caveat == "controller" {
			hasControllerCaveat = true
			break
		}
	}

	if !hasControllerCaveat {
		return nil // No controller caveat, proceed with normal validation
	}

	// Validate controller metadata
	if capability.Metadata == nil {
		return fmt.Errorf("missing controller metadata for controller caveat")
	}

	allowedController, exists := capability.Metadata["controller"]
	if !exists {
		return fmt.Errorf("missing controller address in capability metadata")
	}

	if allowedController != controllerAddress {
		return fmt.Errorf(
			"controller address mismatch: expected %s, got %s",
			allowedController,
			controllerAddress,
		)
	}

	return nil
}

// validateTokenIssuerAsController validates that the token issuer is the controller
func (pv *PermissionValidator) validateTokenIssuerAsController(
	token *ucan.Token,
	controllerAddress string,
) error {
	// For now, we accept any valid token issuer as a potential controller
	// In a more sophisticated implementation, we could:
	// 1. Resolve the issuer DID to get its controller address
	// 2. Validate that the controller address matches
	// 3. Check delegation chains for proper authorization

	if token.Issuer == "" {
		return fmt.Errorf("token issuer is required for controller validation")
	}

	return nil
}

// validateWebAuthnDelegation validates WebAuthn-specific delegation
func (pv *PermissionValidator) validateWebAuthnDelegation(
	token *ucan.Token,
	did string,
	credentialID string,
) error {
	// Find the relevant attenuation for this DID
	for _, att := range token.Attenuations {
		if att.Resource.GetURI() == pv.buildResourceURI(did) {
			// Validate WebAuthn delegation capability
			if err := types.ValidateWebAuthnDelegation(att.Capability, credentialID); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("no matching attenuation found for DID %s", did)
}

// Helper methods

// buildResourceURI constructs DID resource URI
func (pv *PermissionValidator) buildResourceURI(did string) string {
	return fmt.Sprintf("did:%s", pv.extractDIDPattern(did))
}

// extractDIDPattern extracts the method and subject from a full DID
func (pv *PermissionValidator) extractDIDPattern(did string) string {
	// Remove "did:" prefix if present
	if len(did) > 4 && did[:4] == "did:" {
		return did[4:]
	}
	return did
}

// CreateAttenuation creates a UCAN attenuation for DID operations
func (pv *PermissionValidator) CreateAttenuation(
	actions []string,
	did string,
	caveats []string,
) ucan.Attenuation {
	didPattern := pv.extractDIDPattern(did)
	return pv.permissions.CreateDIDAttenuation(actions, didPattern, caveats)
}

// CreateControllerAttenuation creates a controller-specific UCAN attenuation
func (pv *PermissionValidator) CreateControllerAttenuation(
	actions []string,
	did string,
	controllerAddress string,
) ucan.Attenuation {
	didPattern := pv.extractDIDPattern(did)
	return pv.permissions.CreateControllerAttenuation(actions, didPattern, controllerAddress)
}

// CreateWebAuthnDelegationAttenuation creates a WebAuthn delegation attenuation
func (pv *PermissionValidator) CreateWebAuthnDelegationAttenuation(
	actions []string,
	did string,
	credentialID string,
) ucan.Attenuation {
	didPattern := pv.extractDIDPattern(did)
	return pv.permissions.CreateWebAuthnDelegationAttenuation(actions, didPattern, credentialID)
}

// DIDKeyResolver implements ucan.DIDResolver for DID module
type DIDKeyResolver struct {
	keeper Keeper
}

// ResolveDIDKey resolves DID to public key for UCAN verification
func (r *DIDKeyResolver) ResolveDIDKey(ctx context.Context, did string) (keys.DID, error) {
	doc, err := r.keeper.GetDIDDocument(ctx, did)
	if err != nil {
		return keys.DID{}, fmt.Errorf("failed to resolve DID: %w", err)
	}

	// Extract verification method for signature verification
	if len(doc.VerificationMethod) == 0 {
		return keys.DID{}, fmt.Errorf("no verification methods found in DID document")
	}

	// Use the first verification method to parse the DID key
	verificationMethod := doc.VerificationMethod[0]
	if verificationMethod == nil {
		return keys.DID{}, fmt.Errorf("verification method is nil")
	}

	// If the DID document ID is a did:key, parse it directly
	if len(doc.Id) > 8 && doc.Id[:8] == "did:key:" {
		didKey, err := keys.Parse(doc.Id)
		if err != nil {
			return keys.DID{}, fmt.Errorf("failed to parse did:key: %w", err)
		}
		return didKey, nil
	}

	// For other DID methods (like did:sonr), extract public key from verification method
	return r.extractKeyFromVerificationMethod(verificationMethod)
}

// extractKeyFromVerificationMethod extracts a DID key from a verification method
func (r *DIDKeyResolver) extractKeyFromVerificationMethod(
	vm *types.VerificationMethod,
) (keys.DID, error) {
	// Try different public key formats
	if vm.PublicKeyMultibase != "" {
		// Convert multibase to did:key format
		didKeyString := fmt.Sprintf("did:key:%s", vm.PublicKeyMultibase)
		return keys.Parse(didKeyString)
	}

	if vm.PublicKeyBase58 != "" {
		// Try to parse base58 key directly
		didKeyString := fmt.Sprintf("did:key:z%s", vm.PublicKeyBase58)
		return keys.Parse(didKeyString)
	}

	if vm.PublicKeyJwk != "" {
		// For JWK format, we'd need to parse the JSON and extract the key
		// This is more complex and would require JWK parsing
		return keys.DID{}, fmt.Errorf(
			"JWK public key format not yet supported for UCAN verification",
		)
	}

	// Check for WebAuthn credential
	if vm.WebauthnCredential != nil && vm.WebauthnCredential.CredentialId != "" {
		// For WebAuthn credentials, we need to create a pseudo-DID key
		// This is a simplified approach - in practice, you might want to use
		// the actual WebAuthn public key for verification
		return keys.DID{}, fmt.Errorf(
			"WebAuthn credential keys require special handling for UCAN verification",
		)
	}

	return keys.DID{}, fmt.Errorf("no supported public key format found in verification method")
}

// Gasless transaction support

// SupportsGaslessTransaction checks if a UCAN token supports gasless transactions
func (pv *PermissionValidator) SupportsGaslessTransaction(
	ctx context.Context,
	tokenString string,
	did string,
	operation types.DIDOperation,
) (bool, uint64, error) {
	// Parse and verify the token
	token, err := pv.verifier.VerifyToken(ctx, tokenString)
	if err != nil {
		return false, 0, fmt.Errorf("token verification failed: %w", err)
	}

	resourceURI := pv.buildResourceURI(did)

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
