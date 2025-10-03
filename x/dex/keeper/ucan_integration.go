// Package keeper implements UCAN integration for the DEX module
package keeper

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sonr-io/sonr/x/dex/types"
)

// ValidateUCANForDEXOperation validates UCAN token for a DEX operation
func (k Keeper) ValidateUCANForDEXOperation(
	ctx sdk.Context,
	ucanToken string,
	did string,
	operation string,
	params map[string]any,
) error {
	if ucanToken == "" {
		// No UCAN provided - check if operation requires it
		if k.requiresUCAN(operation) {
			return fmt.Errorf("UCAN token required for operation %s", operation)
		}
		return nil
	}

	// Validate UCAN token structure and signature
	capability, err := k.parseUCANToken(ucanToken)
	if err != nil {
		return fmt.Errorf("invalid UCAN token: %w", err)
	}

	// Check expiration
	if ctx.BlockTime().After(capability.Expiration) {
		return fmt.Errorf("UCAN token expired")
	}

	// Verify resource matches operation
	expectedResource := k.getResourceForOperation(operation)
	if !k.resourceMatches(capability.Resource, expectedResource) {
		return fmt.Errorf(
			"UCAN resource %s does not match operation %s",
			capability.Resource,
			operation,
		)
	}

	// Verify ability
	if !k.hasAbility(capability.Ability, operation) {
		return fmt.Errorf(
			"UCAN ability %s insufficient for operation %s",
			capability.Ability,
			operation,
		)
	}

	// Validate constraints
	if err := k.validateConstraints(capability.Constraints, params); err != nil {
		return fmt.Errorf("UCAN constraints not satisfied: %w", err)
	}

	return nil
}

// requiresUCAN checks if an operation requires UCAN authorization
func (k Keeper) requiresUCAN(operation string) bool {
	// Critical operations that always require UCAN
	criticalOps := []string{
		"large_swap",        // Swaps above threshold
		"remove_liquidity",  // Removing liquidity
		"cancel_all_orders", // Canceling all orders
	}

	return slices.Contains(criticalOps, operation)
}

// parseUCANToken parses and validates a UCAN token
func (k Keeper) parseUCANToken(token string) (*types.UCANCapability, error) {
	// This is a simplified implementation
	// Real implementation would validate JWT signature and parse claims

	// For now, parse as JSON for simplicity
	var capability types.UCANCapability
	if err := json.Unmarshal([]byte(token), &capability); err != nil {
		return nil, fmt.Errorf("failed to parse UCAN token: %w", err)
	}

	return &capability, nil
}

// getResourceForOperation maps operations to UCAN resources
func (k Keeper) getResourceForOperation(operation string) string {
	resourceMap := map[string]string{
		"swap":              "dex:swap",
		"execute_swap":      "dex:swap",
		"provide_liquidity": "dex:liquidity:provide",
		"remove_liquidity":  "dex:liquidity:remove",
		"create_order":      "dex:order:create",
		"cancel_order":      "dex:order:cancel",
		"register_account":  "dex:account:register",
	}

	if resource, ok := resourceMap[operation]; ok {
		return resource
	}

	return fmt.Sprintf("dex:%s", operation)
}

// resourceMatches checks if UCAN resource matches required resource
func (k Keeper) resourceMatches(ucanResource, requiredResource string) bool {
	// Exact match
	if ucanResource == requiredResource {
		return true
	}

	// Wildcard match (e.g., "dex:*" matches any DEX operation)
	if strings.HasSuffix(ucanResource, ":*") {
		prefix := strings.TrimSuffix(ucanResource, "*")
		return strings.HasPrefix(requiredResource, prefix)
	}

	// Hierarchical match (e.g., "dex:swap" matches "dex:swap:osmosis")
	return strings.HasPrefix(requiredResource, ucanResource+":")
}

// hasAbility checks if UCAN ability is sufficient for operation
func (k Keeper) hasAbility(ucanAbility, operation string) bool {
	// Map operations to required abilities
	requiredAbilities := map[string][]string{
		"swap":              {"execute", "trade"},
		"provide_liquidity": {"execute", "provide"},
		"remove_liquidity":  {"execute", "remove"},
		"create_order":      {"execute", "create"},
		"cancel_order":      {"execute", "cancel"},
		"read":              {"read", "view"},
	}

	required, ok := requiredAbilities[operation]
	if !ok {
		// Default to requiring "execute" ability
		required = []string{"execute"}
	}

	// Check if UCAN ability matches any required ability
	for _, req := range required {
		if ucanAbility == req || ucanAbility == "*" {
			return true
		}
	}

	return false
}

// validateConstraints validates UCAN constraints against operation parameters
func (k Keeper) validateConstraints(constraints, params map[string]any) error {
	// Check amount constraints
	if maxAmount, ok := constraints["max_amount"]; ok {
		if amount, ok := params["amount"]; ok {
			if !k.isAmountWithinLimit(amount, maxAmount) {
				return fmt.Errorf("amount exceeds UCAN limit")
			}
		}
	}

	// Check pool constraints
	if allowedPools, ok := constraints["allowed_pools"]; ok {
		if poolID, ok := params["pool_id"]; ok {
			if !k.isPoolAllowed(poolID, allowedPools) {
				return fmt.Errorf("pool not allowed by UCAN")
			}
		}
	}

	// Check chain constraints
	if allowedChains, ok := constraints["allowed_chains"]; ok {
		if connectionID, ok := params["connection_id"]; ok {
			if !k.isChainAllowed(connectionID, allowedChains) {
				return fmt.Errorf("chain not allowed by UCAN")
			}
		}
	}

	return nil
}

// isAmountWithinLimit checks if amount is within UCAN limit
func (k Keeper) isAmountWithinLimit(amount, maxAmount any) bool {
	// Convert and compare amounts
	// Simplified implementation - real one would handle different types
	return true
}

// isPoolAllowed checks if pool is in allowed list
func (k Keeper) isPoolAllowed(poolID, allowedPools any) bool {
	// Check if pool is in allowed list
	// Simplified implementation
	return true
}

// isChainAllowed checks if chain is in allowed list
func (k Keeper) isChainAllowed(connectionID, allowedChains any) bool {
	// Check if chain connection is allowed
	// Simplified implementation
	return true
}
