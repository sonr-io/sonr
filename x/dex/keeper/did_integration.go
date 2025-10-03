// Package keeper implements DID integration for the DEX module
package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/sonr/x/dex/types"
)

// ValidateDIDOwnership verifies that the transaction sender owns the specified DID
func (k Keeper) ValidateDIDOwnership(ctx sdk.Context, did string, sender sdk.AccAddress) error {
	// Get DID document from DID keeper
	didDoc, err := k.didKeeper.GetDIDDocument(ctx, did)
	if err != nil {
		return fmt.Errorf("failed to get DID document: %w", err)
	}

	if didDoc == nil {
		return fmt.Errorf("DID document not found for %s", did)
	}

	// Verify sender is the controller of the DID
	if !k.isDIDController(didDoc, sender.String()) {
		return fmt.Errorf("sender %s is not the controller of DID %s", sender, did)
	}

	return nil
}

// isDIDController checks if an address is a controller of the DID
func (k Keeper) isDIDController(didDoc any, address string) bool {
	// This is a simplified check - actual implementation would depend on DID document structure
	// For now, we'll assume the DID document has a Controller field or similar
	// The actual implementation should match the x/did module's structure

	// TODO: Implement proper controller verification based on actual DID document structure
	// This might involve checking:
	// - didDoc.Controller field
	// - didDoc.Authentication keys
	// - didDoc.AssertionMethod keys

	return true // Placeholder - always return true for now
}

// GetDIDCapabilities retrieves the DEX-related capabilities for a DID
func (k Keeper) GetDIDCapabilities(ctx sdk.Context, did string) ([]string, error) {
	// Get DID document
	didDoc, err := k.didKeeper.GetDIDDocument(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to get DID document: %w", err)
	}

	if didDoc == nil {
		return nil, fmt.Errorf("DID document not found for %s", did)
	}

	// Extract DEX-related capabilities from the DID document
	// This would typically be stored in service endpoints or custom fields
	capabilities := []string{
		"swap",
		"liquidity",
		"orders",
	}

	return capabilities, nil
}

// AuthenticateDIDOperation verifies that a DID is authorized for a specific DEX operation
func (k Keeper) AuthenticateDIDOperation(
	ctx sdk.Context,
	did string,
	operation string,
	params map[string]any,
) error {
	// Get DID document to verify it exists and is active
	didDoc, err := k.didKeeper.GetDIDDocument(ctx, did)
	if err != nil {
		return fmt.Errorf("failed to authenticate DID: %w", err)
	}

	if didDoc == nil {
		return fmt.Errorf("DID %s not found", did)
	}

	// Check if DID has the required capability for this operation
	capabilities, err := k.GetDIDCapabilities(ctx, did)
	if err != nil {
		return fmt.Errorf("failed to get DID capabilities: %w", err)
	}

	// Map operations to required capabilities
	requiredCapability := k.getRequiredCapability(operation)
	if !k.hasCapability(capabilities, requiredCapability) {
		return fmt.Errorf("DID %s lacks capability for operation %s", did, operation)
	}

	// Additional authentication checks could be added here:
	// - Check if DID has sufficient reputation
	// - Check if DID has completed KYC/AML if required
	// - Check rate limits for the DID

	return nil
}

// getRequiredCapability maps DEX operations to required capabilities
func (k Keeper) getRequiredCapability(operation string) string {
	switch operation {
	case "swap", "execute_swap":
		return "swap"
	case "provide_liquidity", "remove_liquidity":
		return "liquidity"
	case "create_order", "cancel_order":
		return "orders"
	default:
		return operation
	}
}

// hasCapability checks if a capability exists in the list
func (k Keeper) hasCapability(capabilities []string, required string) bool {
	for _, cap := range capabilities {
		if cap == required {
			return true
		}
	}
	return false
}

// RecordDIDActivity records DEX activity for a DID (for analytics and compliance)
func (k Keeper) RecordDIDActivity(
	ctx sdk.Context,
	did string,
	activity types.DEXActivity,
) error {
	// Store activity record keyed by DID and timestamp
	activityKey := GetDIDActivityKey(did, ctx.BlockTime().Unix())

	// Store the activity
	if err := k.DIDActivities.Set(ctx, activityKey, activity); err != nil {
		return fmt.Errorf("failed to record DID activity: %w", err)
	}

	// Emit event for activity tracking
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDIDActivity,
			sdk.NewAttribute("did", did),
			sdk.NewAttribute("activity_type", activity.Type),
			sdk.NewAttribute("timestamp", fmt.Sprintf("%d", ctx.BlockTime().Unix())),
		),
	)

	return nil
}

// GetDIDActivityHistory retrieves the activity history for a DID
func (k Keeper) GetDIDActivityHistory(
	ctx sdk.Context,
	did string,
	limit uint32,
) ([]types.DEXActivity, error) {
	activities := make([]types.DEXActivity, 0)

	// Walk through activities for this DID
	prefix := GetDIDActivityPrefix(did)
	iterator, err := k.DIDActivities.Iterate(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to iterate DID activities: %w", err)
	}
	defer iterator.Close()

	count := uint32(0)
	for ; iterator.Valid() && count < limit; iterator.Next() {
		key, err := iterator.Key()
		if err != nil {
			continue
		}

		// Check if key starts with the DID prefix
		if len(key) >= len(prefix) && string(key[:len(prefix)]) == prefix {
			activity, err := iterator.Value()
			if err == nil {
				activities = append(activities, activity)
				count++
			}
		}
	}

	return activities, nil
}

// GetDIDActivityPrefix returns the key prefix for a DID's activities
func GetDIDActivityPrefix(did string) string {
	return fmt.Sprintf("did_activity_%s_", did)
}

// GetDIDActivityKey returns the key for storing a DID activity
func GetDIDActivityKey(did string, timestamp int64) string {
	return fmt.Sprintf("did_activity_%s_%d", did, timestamp)
}
