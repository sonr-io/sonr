package keeper

import (
	"context"
	"fmt"

	"github.com/sonr-io/crypto/keys"
	"github.com/sonr-io/crypto/ucan"
	"github.com/sonr-io/sonr/x/dwn/types"
)

// PermissionValidator wraps UCAN verifier for DWN-specific permission validation
type PermissionValidator struct {
	verifier    *ucan.Verifier
	didKeeper   types.DIDKeeper
	permissions *types.UCANPermissionRegistry
}

// NewPermissionValidator creates a new DWN permission validator
func NewPermissionValidator(didKeeper types.DIDKeeper) *PermissionValidator {
	didResolver := &DIDKeyResolver{didKeeper: didKeeper}
	verifier := ucan.NewVerifier(didResolver)

	return &PermissionValidator{
		verifier:    verifier,
		didKeeper:   didKeeper,
		permissions: types.NewUCANPermissionRegistry(),
	}
}

// NewPermissionValidatorWithVerifier creates a new DWN permission validator with custom verifier (for testing)
func NewPermissionValidatorWithVerifier(
	didKeeper types.DIDKeeper,
	verifier *ucan.Verifier,
) *PermissionValidator {
	return &PermissionValidator{
		verifier:    verifier,
		didKeeper:   didKeeper,
		permissions: types.NewUCANPermissionRegistry(),
	}
}

// ValidatePermission validates UCAN token for DWN operation
func (pv *PermissionValidator) ValidatePermission(
	ctx context.Context,
	tokenString string,
	target string,
	operation types.DWNOperation,
) error {
	// Get required UCAN capabilities for the operation
	capabilities, err := pv.permissions.GetRequiredUCANCapabilities(operation)
	if err != nil {
		return fmt.Errorf("failed to get required UCAN capabilities: %w", err)
	}

	// Build resource URI for DWN target
	resourceURI := pv.buildResourceURI(target, operation)

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

// ValidateRecordOperation validates UCAN token for record-specific operations
func (pv *PermissionValidator) ValidateRecordOperation(
	ctx context.Context,
	tokenString string,
	target string,
	recordID string,
	operation types.RecordOperation,
) error {
	// Get required UCAN capabilities for record operation
	capabilities := pv.permissions.GetRecordUCANCapabilities(operation)

	// Build resource URI for specific record
	resourceURI := pv.buildRecordResourceURI(target, recordID)

	// Verify UCAN token
	_, err := pv.verifier.VerifyCapability(
		ctx,
		tokenString,
		resourceURI,
		capabilities,
	)
	if err != nil {
		return fmt.Errorf("record operation validation failed: %w", err)
	}

	return nil
}

// ValidateProtocolOperation validates UCAN token for protocol operations
func (pv *PermissionValidator) ValidateProtocolOperation(
	ctx context.Context,
	tokenString string,
	target string,
	protocolURI string,
	operation types.ProtocolOperation,
) error {
	// Get required UCAN capabilities for protocol operation
	capabilities := pv.permissions.GetProtocolUCANCapabilities(operation)

	// Build resource URI for protocol
	resourceURI := pv.buildProtocolResourceURI(target, protocolURI)

	// Verify UCAN token
	_, err := pv.verifier.VerifyCapability(
		ctx,
		tokenString,
		resourceURI,
		capabilities,
	)
	if err != nil {
		return fmt.Errorf("protocol operation validation failed: %w", err)
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

// buildResourceURI constructs DWN resource URI
func (pv *PermissionValidator) buildResourceURI(
	target string,
	operation types.DWNOperation,
) string {
	return fmt.Sprintf("dwn://%s/%s", target, operation.String())
}

// buildRecordResourceURI constructs resource URI for specific record
func (pv *PermissionValidator) buildRecordResourceURI(target, recordID string) string {
	return fmt.Sprintf("dwn://%s/records/%s", target, recordID)
}

// buildProtocolResourceURI constructs resource URI for protocol
func (pv *PermissionValidator) buildProtocolResourceURI(target, protocolURI string) string {
	return fmt.Sprintf("dwn://%s/protocols/%s", target, protocolURI)
}

// DIDKeyResolver implements ucan.DIDResolver for DWN module
type DIDKeyResolver struct {
	didKeeper types.DIDKeeper
}

// ResolveDIDKey resolves DID to public key for UCAN verification
func (r *DIDKeyResolver) ResolveDIDKey(ctx context.Context, did string) (keys.DID, error) {
	doc, err := r.didKeeper.GetDIDDocument(ctx, did)
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

	// For other DID methods, we'd need to extract public key from verification method
	// For now, return an error for unsupported DID types
	return keys.DID{}, fmt.Errorf("unsupported DID method: %s", doc.Id)
}
