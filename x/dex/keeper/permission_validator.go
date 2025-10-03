package keeper

import (
	"context"
	"fmt"

	"github.com/sonr-io/sonr/crypto/keys"
	"github.com/sonr-io/sonr/crypto/ucan"
	"github.com/sonr-io/sonr/x/dex/types"
)

// PermissionValidator wraps UCAN verifier for DEX-specific permission validation
type PermissionValidator struct {
	verifier    *ucan.Verifier
	keeper      Keeper
	permissions *types.UCANPermissionRegistry
}

// NewPermissionValidator creates a new DEX permission validator
func NewPermissionValidator(keeper Keeper) *PermissionValidator {
	didResolver := &DEXDIDResolver{keeper: keeper}
	verifier := ucan.NewVerifier(didResolver)

	return &PermissionValidator{
		verifier:    verifier,
		keeper:      keeper,
		permissions: types.NewUCANPermissionRegistry(),
	}
}

// ValidatePermission validates UCAN token for DEX operation
func (pv *PermissionValidator) ValidatePermission(
	ctx context.Context,
	tokenString string,
	resourceType string,
	resourceID string,
	operation types.DEXOperation,
) error {
	// Get required UCAN capabilities for the operation
	capabilities, err := pv.permissions.GetRequiredUCANCapabilities(operation)
	if err != nil {
		return fmt.Errorf("failed to get required UCAN capabilities: %w", err)
	}

	// Build resource URI for DEX
	mapper := types.NewUCANCapabilityMapper()
	resourceURI := mapper.CreateDEXResourceURI(resourceType, resourceID)

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

// ValidateSwapPermission validates UCAN token for swap operations
func (pv *PermissionValidator) ValidateSwapPermission(
	ctx context.Context,
	tokenString string,
	poolID string,
	amount string,
	operation types.DEXOperation,
) error {
	// Get required UCAN capabilities for the operation
	capabilities, err := pv.permissions.GetRequiredUCANCapabilities(operation)
	if err != nil {
		return fmt.Errorf("failed to get required UCAN capabilities: %w", err)
	}

	// Build pool resource URI
	mapper := types.NewUCANCapabilityMapper()
	resourceURI := mapper.CreatePoolResourceURI(poolID)

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

	// Additional amount validation
	if err := pv.validateAmountConstraint(token, amount); err != nil {
		return fmt.Errorf("amount constraint validation failed: %w", err)
	}

	return nil
}

// ValidateLiquidityPermission validates UCAN token for liquidity operations
func (pv *PermissionValidator) ValidateLiquidityPermission(
	ctx context.Context,
	tokenString string,
	poolID string,
	operation types.DEXOperation,
) error {
	// Get required UCAN capabilities for the operation
	capabilities, err := pv.permissions.GetRequiredUCANCapabilities(operation)
	if err != nil {
		return fmt.Errorf("failed to get required UCAN capabilities: %w", err)
	}

	// Build pool resource URI
	mapper := types.NewUCANCapabilityMapper()
	resourceURI := mapper.CreatePoolResourceURI(poolID)

	// Verify UCAN token
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

// ValidateOrderPermission validates UCAN token for order operations
func (pv *PermissionValidator) ValidateOrderPermission(
	ctx context.Context,
	tokenString string,
	orderID string,
	operation types.DEXOperation,
) error {
	// Get required UCAN capabilities for the operation
	capabilities, err := pv.permissions.GetRequiredUCANCapabilities(operation)
	if err != nil {
		return fmt.Errorf("failed to get required UCAN capabilities: %w", err)
	}

	// Build order resource URI
	mapper := types.NewUCANCapabilityMapper()
	resourceURI := mapper.CreateOrderResourceURI(orderID)

	// Verify UCAN token
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

// VerifyDelegationChain validates complete UCAN delegation chain
func (pv *PermissionValidator) VerifyDelegationChain(
	ctx context.Context,
	tokenString string,
) error {
	return pv.verifier.VerifyDelegationChain(ctx, tokenString)
}

// Internal validation methods

// validateAmountConstraint validates amount constraints
func (pv *PermissionValidator) validateAmountConstraint(
	token *ucan.Token,
	amount string,
) error {
	// For now, we'll accept all amounts
	// In a real implementation, we'd check against maximum amounts
	// specified in the token's attenuations
	return nil
}

// validatePoolConstraint validates pool constraints
func (pv *PermissionValidator) validatePoolConstraint(
	token *ucan.Token,
	poolID string,
) error {
	// Check if the token's resource matches the pool
	for _, att := range token.Attenuations {
		if simpleResource, ok := att.Resource.(*ucan.SimpleResource); ok {
			// Check if resource matches pool pattern
			if simpleResource.Scheme == "dex" {
				expectedValue := fmt.Sprintf("pool:%s", poolID)
				if simpleResource.Value == expectedValue || simpleResource.Value == "pool:*" {
					return nil
				}
			}
		}
	}

	return fmt.Errorf("no matching pool attenuation found for pool %s", poolID)
}

// Helper methods

// CreateAttenuation creates a UCAN attenuation for DEX operations
func (pv *PermissionValidator) CreateAttenuation(
	actions []string,
	resourceType string,
	resourceID string,
) ucan.Attenuation {
	return pv.permissions.CreateDEXAttenuation(actions, resourceType, resourceID)
}

// CreateAmountLimitedAttenuation creates an amount-limited UCAN attenuation
func (pv *PermissionValidator) CreateAmountLimitedAttenuation(
	actions []string,
	poolID string,
	maxAmount string,
) ucan.Attenuation {
	return pv.permissions.CreateAmountLimitedAttenuation(actions, poolID, maxAmount)
}

// CreatePoolRestrictedAttenuation creates a pool-restricted UCAN attenuation
func (pv *PermissionValidator) CreatePoolRestrictedAttenuation(
	actions []string,
	allowedPools []string,
) ucan.Attenuation {
	return pv.permissions.CreatePoolRestrictedAttenuation(actions, allowedPools)
}

// DEXDIDResolver implements ucan.DIDResolver for DEX module
type DEXDIDResolver struct {
	keeper Keeper
}

// ResolveDIDKey resolves DID to public key for UCAN verification
func (r *DEXDIDResolver) ResolveDIDKey(ctx context.Context, did string) (keys.DID, error) {
	// For DEX module, we need to resolve DIDs from the DID module
	// This would require cross-module keeper access

	// Check if the DEX keeper has access to DID keeper
	if r.keeper.didKeeper != nil {
		didDoc, err := r.keeper.didKeeper.GetDIDDocument(ctx, did)
		if err != nil {
			return keys.DID{}, fmt.Errorf("failed to get DID document: %w", err)
		}

		if didDoc == nil {
			return keys.DID{}, fmt.Errorf("DID document not found")
		}

		// Parse the DID string into a keys.DID
		return keys.Parse(did)
	}

	return keys.DID{}, fmt.Errorf("DID resolver not available in DEX module")
}

// Gasless transaction support

// SupportsGaslessTransaction checks if a UCAN token supports gasless transactions
func (pv *PermissionValidator) SupportsGaslessTransaction(
	ctx context.Context,
	tokenString string,
	poolID string,
	operation types.DEXOperation,
) (bool, uint64, error) {
	// Parse and verify the token
	token, err := pv.verifier.VerifyToken(ctx, tokenString)
	if err != nil {
		return false, 0, fmt.Errorf("token verification failed: %w", err)
	}

	mapper := types.NewUCANCapabilityMapper()
	resourceURI := mapper.CreatePoolResourceURI(poolID)

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
	poolID string,
) (bool, uint64, uint64, error) {
	// Parse and verify the token
	token, err := pv.verifier.VerifyToken(ctx, tokenString)
	if err != nil {
		return false, 0, 0, fmt.Errorf("token verification failed: %w", err)
	}

	mapper := types.NewUCANCapabilityMapper()
	resourceURI := mapper.CreatePoolResourceURI(poolID)

	// Check each attenuation for rate limiting
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
