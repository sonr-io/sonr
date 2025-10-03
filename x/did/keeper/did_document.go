package keeper

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/sonr/x/did/types"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CreateEnhancedDIDDocument creates a DID document with proper controller and verification methods
// This is used during WebAuthn registration to create a complete DID document
func (k Keeper) CreateEnhancedDIDDocument(
	ctx context.Context,
	did string,
	controllerAddress string,
	webauthnCredential *types.WebAuthnCredential,
	assertionType string,
	assertionValue string,
	enclavePublicKey []byte,
) (*types.DIDDocument, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Derive controller DID from enclave public key
	controllerDID := k.deriveControllerDID(enclavePublicKey)

	// Create WebAuthn authentication method
	webauthnMethod := &types.VerificationMethod{
		Id:                     fmt.Sprintf("%s#webauthn-1", did),
		Controller:             did,
		VerificationMethodKind: "WebAuthnCredential2024",
		WebauthnCredential:     webauthnCredential,
	}

	// Create assertion method based on type (email/tel)
	var assertionMethod *types.VerificationMethod
	if assertionType == "email" || assertionType == "tel" {
		assertionMethod = &types.VerificationMethod{
			Id:                     fmt.Sprintf("%s#%s-assertion", did, assertionType),
			Controller:             did,
			VerificationMethodKind: "AssertionMethod2024",
			BlockchainAccountId:    fmt.Sprintf("did:%s:%s", assertionType, types.HashAssertionValue(assertionValue)),
		}
	}

	// Create Sonr account assertion method
	sonrAccountMethod := &types.VerificationMethod{
		Id:                     fmt.Sprintf("%s#sonr-account", did),
		Controller:             did,
		VerificationMethodKind: "BlockchainAccountId2024",
		BlockchainAccountId:    fmt.Sprintf("sonr:%s", controllerAddress),
	}

	// Create enclave key agreement method if public key is provided
	var enclaveMethod *types.VerificationMethod
	if len(enclavePublicKey) > 0 {
		// Create JWK string representation
		jwkString := fmt.Sprintf(`{"kty":"EC","crv":"secp256k1","x":"%s","y":"%s"}`,
			base64.URLEncoding.EncodeToString(enclavePublicKey[:min(32, len(enclavePublicKey))]),
			base64.URLEncoding.EncodeToString(enclavePublicKey[min(32, len(enclavePublicKey)):]),
		)

		enclaveMethod = &types.VerificationMethod{
			Id:                     fmt.Sprintf("%s#enclave-key", did),
			Controller:             did,
			VerificationMethodKind: "JsonWebKey2020",
			PublicKeyJwk:           jwkString,
		}
	}

	// Build verification methods array
	verificationMethods := []*types.VerificationMethod{
		webauthnMethod,
		sonrAccountMethod,
	}
	if assertionMethod != nil {
		verificationMethods = append(verificationMethods, assertionMethod)
	}
	if enclaveMethod != nil {
		verificationMethods = append(verificationMethods, enclaveMethod)
	}

	// Create verification method references
	authRefs := []*types.VerificationMethodReference{
		{VerificationMethodId: webauthnMethod.Id},
	}

	assertRefs := []*types.VerificationMethodReference{
		{VerificationMethodId: sonrAccountMethod.Id},
	}
	if assertionMethod != nil {
		assertRefs = append(assertRefs, &types.VerificationMethodReference{
			VerificationMethodId: assertionMethod.Id,
		})
	}

	keyAgreementRefs := []*types.VerificationMethodReference{}
	if enclaveMethod != nil {
		keyAgreementRefs = append(keyAgreementRefs, &types.VerificationMethodReference{
			VerificationMethodId: enclaveMethod.Id,
		})
	}

	capabilityInvocationRefs := []*types.VerificationMethodReference{
		{VerificationMethodId: webauthnMethod.Id},
	}

	// Add service endpoints
	services := k.createDefaultServices(did)

	// Create the DID document
	didDoc := &types.DIDDocument{
		Id:                   did,
		PrimaryController:    controllerDID,
		VerificationMethod:   verificationMethods,
		Authentication:       authRefs,
		AssertionMethod:      assertRefs,
		KeyAgreement:         keyAgreementRefs,
		CapabilityInvocation: capabilityInvocationRefs,
		CapabilityDelegation: []*types.VerificationMethodReference{},
		Service:              services,
		AlsoKnownAs:          k.generateAlsoKnownAs(assertionType, assertionValue),
		CreatedAt:            sdkCtx.BlockHeight(),
		UpdatedAt:            sdkCtx.BlockHeight(),
		Version:              1,
		Deactivated:          false,
	}

	return didDoc, nil
}

// deriveControllerDID derives a controller DID from enclave public key
func (k Keeper) deriveControllerDID(enclavePublicKey []byte) string {
	if len(enclavePublicKey) == 0 {
		// If no enclave key, use a default controller pattern
		return "did:sonr:controller"
	}

	// Create deterministic controller DID from public key
	// Use first 16 bytes of public key for identifier
	identifier := base64.URLEncoding.EncodeToString(enclavePublicKey[:16])
	identifier = strings.TrimRight(identifier, "=") // Remove padding

	return fmt.Sprintf("did:sonr:idx%s", identifier)
}

// createDefaultServices creates default service endpoints for a DID
func (k Keeper) createDefaultServices(did string) []*types.Service {
	return []*types.Service{
		{
			Id:             fmt.Sprintf("%s#dwn", did),
			ServiceKind:    "DecentralizedWebNode",
			SingleEndpoint: "https://dwn.sonr.io",
		},
		{
			Id:             fmt.Sprintf("%s#messaging", did),
			ServiceKind:    "MessagingService",
			SingleEndpoint: "https://msg.sonr.io",
		},
	}
}

// generateAlsoKnownAs generates alternative identifiers for the DID
func (k Keeper) generateAlsoKnownAs(assertionType string, assertionValue string) []string {
	alsoKnownAs := []string{}

	if assertionType == "email" {
		// Add email-based identifier
		alsoKnownAs = append(alsoKnownAs, fmt.Sprintf("mailto:%s", assertionValue))
	} else if assertionType == "tel" {
		// Add phone-based identifier
		alsoKnownAs = append(alsoKnownAs, fmt.Sprintf("tel:%s", assertionValue))
	}

	return alsoKnownAs
}

// UpdateDIDDocumentWithUCAN updates a DID document with UCAN delegation chain reference
func (k Keeper) UpdateDIDDocumentWithUCAN(
	ctx context.Context,
	did string,
	ucanRootProof string,
	ucanOriginToken string,
) error {
	// Get existing DID document
	ormDoc, err := k.OrmDB.DIDDocumentTable().Get(ctx, did)
	if err != nil {
		return fmt.Errorf("failed to get DID document: %w", err)
	}

	didDoc := types.DIDDocumentFromORM(ormDoc)

	// Add UCAN service endpoint to indicate UCAN support
	ucanService := &types.Service{
		Id:             fmt.Sprintf("%s#ucan", did),
		ServiceKind:    "UCANDelegation",
		SingleEndpoint: "ucan:enabled:true",
	}

	// Check if service already exists
	serviceExists := false
	for _, svc := range didDoc.Service {
		if svc.ServiceKind == "UCANDelegation" {
			serviceExists = true
			break
		}
	}

	if !serviceExists {
		didDoc.Service = append(didDoc.Service, ucanService)
	}

	// Update version and timestamp
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	didDoc.UpdatedAt = sdkCtx.BlockHeight()
	didDoc.Version = didDoc.Version + 1

	// Store updated document
	ormUpdated := didDoc.ToORM()
	if err := k.OrmDB.DIDDocumentTable().Update(ctx, ormUpdated); err != nil {
		return fmt.Errorf("failed to update DID document with UCAN: %w", err)
	}

	return nil
}

// GetDIDDocumentWithEnhancements retrieves a DID document with all enhancements
func (k Keeper) GetDIDDocumentWithEnhancements(
	ctx context.Context,
	did string,
) (*types.DIDDocument, error) {
	// Get DID document from ORM
	ormDoc, err := k.OrmDB.DIDDocumentTable().Get(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("DID document not found: %s", did)
	}

	didDoc := types.DIDDocumentFromORM(ormDoc)

	// Ensure all required fields are populated
	if didDoc.PrimaryController == "" {
		// Try to derive from verification methods
		for _, vm := range didDoc.VerificationMethod {
			if vm.Controller != "" {
				didDoc.PrimaryController = vm.Controller
				break
			}
		}
	}

	return didDoc, nil
}

// ValidateDIDDocumentStructure validates the structure of an enhanced DID document
func (k Keeper) ValidateDIDDocumentStructure(didDoc *types.DIDDocument) error {
	// Check required fields
	if didDoc.Id == "" {
		return fmt.Errorf("DID document must have an ID")
	}

	// Verify controller
	if didDoc.PrimaryController == "" {
		return fmt.Errorf("DID document must have a primary controller")
	}

	// Check verification methods
	if len(didDoc.VerificationMethod) == 0 {
		return fmt.Errorf("DID document must have at least one verification method")
	}

	// Verify authentication methods
	if len(didDoc.Authentication) == 0 {
		return fmt.Errorf("DID document must have at least one authentication method")
	}

	// Verify assertion methods (should have at least 2: Sonr account + email/tel)
	if len(didDoc.AssertionMethod) < 1 {
		return fmt.Errorf("DID document must have at least one assertion method")
	}

	// Check for WebAuthn credential
	hasWebAuthn := false
	for _, vm := range didDoc.VerificationMethod {
		if vm.WebauthnCredential != nil {
			hasWebAuthn = true
			break
		}
	}

	if !hasWebAuthn {
		return fmt.Errorf("DID document must have a WebAuthn credential for authentication")
	}

	return nil
}
