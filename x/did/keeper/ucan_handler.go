package keeper

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/sonr/crypto/ucan"
	"github.com/sonr-io/sonr/x/did/types"
)

// InitializeUCANDelegationChain creates a UCAN delegation chain for a new DID
// with the validator as root proof issuer
func (k Keeper) InitializeUCANDelegationChain(
	ctx context.Context,
	didID string,
	controllerAddress string,
	webauthnCredentialID string,
) (*types.UCANDelegationChain, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get validator address/key (use block proposer as validator)
	proposer := sdkCtx.BlockHeader().ProposerAddress
	validatorDID := fmt.Sprintf("did:sonr:validator:%s", base64.URLEncoding.EncodeToString(proposer))

	// Create root capability - validator grants full admin rights to the DID controller
	rootAttenuation, err := createRootAttenuation(didID, controllerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create root attenuation: %w", err)
	}

	// Generate validator-issued root token (24 hour expiry for initial registration)
	rootToken, err := ucan.GenerateModuleJWTToken(
		[]ucan.Attenuation{rootAttenuation},
		validatorDID,      // issuer: validator
		controllerAddress, // audience: controller
		24*time.Hour,      // duration
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate root token: %w", err)
	}

	// Create origin token for wallet admin operations
	// This token is scoped to WebAuthn credential and allows wallet operations
	originAttenuation, err := createOriginAttenuation(didID, webauthnCredentialID)
	if err != nil {
		return nil, fmt.Errorf("failed to create origin attenuation: %w", err)
	}

	// Generate origin token (30 day expiry for wallet operations)
	originToken, err := ucan.GenerateModuleJWTToken(
		[]ucan.Attenuation{originAttenuation},
		controllerAddress, // issuer: controller (delegating from root)
		didID,             // audience: the DID itself
		30*24*time.Hour,   // duration: 30 days
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate origin token: %w", err)
	}

	// Create delegation chain structure
	delegationChain := &types.UCANDelegationChain{
		Did:             didID,
		RootProof:       rootToken,
		OriginToken:     originToken,
		ValidatorIssuer: validatorDID,
		CreatedAt:       sdkCtx.BlockTime().Unix(),
		ExpiresAt:       sdkCtx.BlockTime().Add(30 * 24 * time.Hour).Unix(),
		Metadata: map[string]string{
			"webauthn_credential": webauthnCredentialID,
			"controller":          controllerAddress,
			"registration_type":   "webauthn",
			"block_height":        fmt.Sprintf("%d", sdkCtx.BlockHeight()),
		},
	}

	// Store delegation chain in keeper state (if we have a storage mechanism)
	if err := k.storeUCANDelegationChain(ctx, delegationChain); err != nil {
		return nil, fmt.Errorf("failed to store delegation chain: %w", err)
	}

	return delegationChain, nil
}

// createRootAttenuation creates the root capability granting full admin rights
func createRootAttenuation(didID string, controllerAddress string) (ucan.Attenuation, error) {
	// Create DID capability with full admin rights
	capability := &ucan.DIDCapability{
		Action: "*", // Full access
		Caveats: []string{
			fmt.Sprintf("controller:%s", controllerAddress),
			"registration:webauthn",
		},
		Metadata: map[string]string{
			"purpose": "root_delegation",
			"scope":   "full_admin",
		},
	}

	// Create DID resource using embedded SimpleResource
	resource := &ucan.DIDResource{
		SimpleResource: ucan.SimpleResource{
			Scheme: "did",
			Value:  didID,
			URI:    didID,
		},
		DIDMethod:  "sonr",
		DIDSubject: controllerAddress,
	}

	return ucan.Attenuation{
		Capability: capability,
		Resource:   resource,
	}, nil
}

// createOriginAttenuation creates the origin token for wallet admin operations
func createOriginAttenuation(didID string, webauthnCredentialID string) (ucan.Attenuation, error) {
	// Create wallet-specific capabilities
	capability := &ucan.MultiCapability{
		Actions: []string{
			"vault:read",
			"vault:write",
			"vault:sign",
			"vault:export",
			"did:update",
			"did:add-verification-method",
			"did:link-wallet",
			"dwn:records-write",
			"dwn:records-delete",
			"dwn:permissions-grant",
		},
	}

	// Create DID resource scoped to WebAuthn credential
	resource := &ucan.SimpleResource{
		Scheme: "did",
		Value:  fmt.Sprintf("%s#%s", didID, webauthnCredentialID),
		URI:    fmt.Sprintf("%s#%s", didID, webauthnCredentialID),
	}

	return ucan.Attenuation{
		Capability: capability,
		Resource:   resource,
	}, nil
}

// storeUCANDelegationChain stores the delegation chain in keeper state
func (k Keeper) storeUCANDelegationChain(ctx context.Context, chain *types.UCANDelegationChain) error {
	// Store in a dedicated UCAN delegation chain table or as part of DID document metadata
	// For now, we'll store it as part of the DID document metadata

	// TODO: Implement actual storage mechanism
	// This could be:
	// 1. A separate ORM table for UCAN delegation chains
	// 2. Part of the DID document's metadata field
	// 3. A separate key-value store entry

	// For now, we'll just validate the chain
	if chain.Did == "" || chain.RootProof == "" || chain.OriginToken == "" {
		return fmt.Errorf("invalid delegation chain: missing required fields")
	}

	return nil
}

// RefreshUCANToken refreshes an expiring UCAN token
func (k Keeper) RefreshUCANToken(
	ctx context.Context,
	didID string,
	oldToken string,
) (string, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Parse the old token to extract capabilities
	parsedToken, err := ucan.VerifyModuleJWTToken(oldToken, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to parse old token: %w", err)
	}

	// Check if token is close to expiry (within 7 days)
	expiryTime := time.Unix(parsedToken.ExpiresAt, 0)
	if time.Until(expiryTime) > 7*24*time.Hour {
		// Token still has plenty of time, no need to refresh
		return oldToken, nil
	}

	// Generate new token with same capabilities but extended expiry
	newToken, err := ucan.GenerateModuleJWTToken(
		parsedToken.Attenuations,
		parsedToken.Issuer,
		parsedToken.Audience,
		30*24*time.Hour, // Refresh for another 30 days
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate refreshed token: %w", err)
	}

	// Update stored delegation chain with new token
	// TODO: Update storage with new token

	// Emit event for token refresh
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"ucan_token_refreshed",
			sdk.NewAttribute("did", didID),
			sdk.NewAttribute("old_token_prefix", oldToken[:20]+"..."), // Only log prefix for security
			sdk.NewAttribute("new_token_prefix", newToken[:20]+"..."),
			sdk.NewAttribute("refreshed_at", fmt.Sprintf("%d", sdkCtx.BlockTime().Unix())),
		),
	)

	return newToken, nil
}

// ValidateUCANToken validates a UCAN token for a specific DID and action
func (k Keeper) ValidateUCANToken(
	ctx context.Context,
	token string,
	didID string,
	requiredAction string,
) error {
	// Parse and verify the token
	parsedToken, err := ucan.VerifyModuleJWTToken(token, "", didID)
	if err != nil {
		return fmt.Errorf("token verification failed: %w", err)
	}

	// Check if token has required capability
	hasCapability := false
	for _, att := range parsedToken.Attenuations {
		actions := att.Capability.GetActions()
		for _, action := range actions {
			if action == "*" || action == requiredAction {
				// Also check if resource matches the DID
				resourceURI := att.Resource.GetURI()
				if resourceURI == didID || resourceURI == "*" {
					hasCapability = true
					break
				}
			}
		}
		if hasCapability {
			break
		}
	}

	if !hasCapability {
		return fmt.Errorf("token does not have required capability: %s for DID: %s", requiredAction, didID)
	}

	return nil
}

// GetUCANDelegationChain retrieves the delegation chain for a DID
func (k Keeper) GetUCANDelegationChain(ctx context.Context, didID string) (*types.UCANDelegationChain, error) {
	// TODO: Implement retrieval from storage
	// This would fetch from wherever we store the delegation chains

	// For now, return a placeholder error
	return nil, fmt.Errorf("delegation chain retrieval not yet implemented for DID: %s", didID)
}
