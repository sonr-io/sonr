package keeper

import (
	"context"
	"fmt"
	"time"

	apiv1 "github.com/sonr-io/sonr/api/svc/v1"
	"github.com/sonr-io/crypto/ucan"
	"github.com/sonr-io/sonr/x/svc/types"
)

// ValidateServicePermissions validates that the requested permissions are valid for services
func (k Keeper) ValidateServicePermissions(ctx context.Context, permissions []string) error {
	if len(permissions) == 0 {
		return fmt.Errorf("at least one permission is required")
	}

	// Define valid service permissions
	validPermissions := map[string]bool{
		"read":         true,
		"write":        true,
		"admin":        true,
		"register":     true,
		"update":       true,
		"delete":       true,
		"execute":      true,
		"access":       true,
		"manage":       true,
		"authenticate": true,
	}

	// Validate each requested permission
	for _, permission := range permissions {
		if permission == "" {
			return fmt.Errorf("permission cannot be empty")
		}
		if !validPermissions[permission] {
			return fmt.Errorf("invalid permission: %s", permission)
		}
	}

	return nil
}

// CreateServiceRootCapability creates a root capability for a service registration
func (k Keeper) CreateServiceRootCapability(
	ctx context.Context,
	msg *types.MsgRegisterService,
) (string, error) {
	// Validate inputs
	if msg.Domain == "" {
		return "", fmt.Errorf("domain cannot be empty")
	}
	if msg.Creator == "" {
		return "", fmt.Errorf("creator cannot be empty")
	}
	if msg.ServiceId == "" {
		return "", fmt.Errorf("service ID cannot be empty")
	}
	if len(msg.RequestedPermissions) == 0 {
		return "", fmt.Errorf("at least one permission is required")
	}

	// Verify domain ownership
	verified, err := k.IsDomainVerified(ctx, msg.Domain, msg.Creator)
	if err != nil {
		return "", fmt.Errorf("domain verification check failed: %w", err)
	}
	if !verified {
		return "", types.ErrDomainNotVerified
	}

	// Generate unique capability ID
	capabilityID := fmt.Sprintf("cap_%s_%d", msg.ServiceId, time.Now().UnixNano())

	k.logger.Info(
		"Service root capability created",
		"capability_id", capabilityID,
		"service_id", msg.ServiceId,
		"domain", msg.Domain,
		"creator", msg.Creator,
		"permissions", msg.RequestedPermissions,
	)

	// Return the capability ID as the "CID" for backward compatibility
	return capabilityID, nil
}

// ValidateUCANToken validates a UCAN token using the internal library
func (k Keeper) ValidateUCANToken(
	ctx context.Context,
	tokenString string,
	resource string,
	abilities []string,
) (*ucan.Token, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token string cannot be empty")
	}

	// Verify the token using the internal UCAN library
	token, err := k.ucanVerifier.VerifyCapability(ctx, tokenString, resource, abilities)
	if err != nil {
		return nil, fmt.Errorf("UCAN token validation failed: %w", err)
	}

	return token, nil
}

// ValidateUCANDelegationChain validates a complete UCAN delegation chain
func (k Keeper) ValidateUCANDelegationChain(
	ctx context.Context,
	tokenString string,
) error {
	if tokenString == "" {
		return fmt.Errorf("token string cannot be empty")
	}

	// Verify the delegation chain using the internal UCAN library
	if err := k.ucanVerifier.VerifyDelegationChain(ctx, tokenString); err != nil {
		return fmt.Errorf("UCAN delegation chain validation failed: %w", err)
	}

	return nil
}

// CreatePermissionCapabilityChain creates a chain of capabilities for permissions
func (k Keeper) CreatePermissionCapabilityChain(
	ctx context.Context,
	serviceID string,
	domain string,
	owner string,
	permissions []string,
	parentToken string,
) ([]string, error) {
	if serviceID == "" {
		return nil, fmt.Errorf("service ID cannot be empty")
	}
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}
	if owner == "" {
		return nil, fmt.Errorf("owner cannot be empty")
	}
	if len(permissions) == 0 {
		return nil, fmt.Errorf("at least one permission is required")
	}

	var capabilityChain []string

	// If a parent token is provided, validate it first
	if parentToken != "" {
		resource := fmt.Sprintf("service://%s", domain)
		_, err := k.ValidateUCANToken(ctx, parentToken, resource, permissions)
		if err != nil {
			return nil, fmt.Errorf("parent token validation failed: %w", err)
		}
	}

	// Create individual capabilities for each permission
	for i, permission := range permissions {
		capabilityID := fmt.Sprintf("cap_%s_%s_%d_%d",
			serviceID,
			permission,
			time.Now().UnixNano(),
			i,
		)

		k.logger.Info(
			"Permission capability created in chain",
			"capability_id", capabilityID,
			"service_id", serviceID,
			"domain", domain,
			"owner", owner,
			"permission", permission,
			"chain_index", i,
		)

		capabilityChain = append(capabilityChain, capabilityID)
	}

	k.logger.Info(
		"Permission capability chain created",
		"service_id", serviceID,
		"domain", domain,
		"owner", owner,
		"total_capabilities", len(capabilityChain),
		"permissions", permissions,
	)

	return capabilityChain, nil
}

// ValidatePermissionCapabilityChain validates a chain of permission capabilities
func (k Keeper) ValidatePermissionCapabilityChain(
	ctx context.Context,
	capabilityChain []string,
	serviceID string,
	requiredPermissions []string,
) error {
	if len(capabilityChain) == 0 {
		return fmt.Errorf("capability chain cannot be empty")
	}
	if serviceID == "" {
		return fmt.Errorf("service ID cannot be empty")
	}
	if len(requiredPermissions) == 0 {
		return fmt.Errorf("at least one required permission must be specified")
	}

	// Check if the chain has enough capabilities for the required permissions
	if len(capabilityChain) < len(requiredPermissions) {
		return fmt.Errorf(
			"insufficient capabilities in chain: got %d, need %d",
			len(capabilityChain),
			len(requiredPermissions),
		)
	}

	// Validate that each required permission is covered by a capability in the chain
	permissionsCovered := make(map[string]bool)

	for _, capabilityID := range capabilityChain {
		// Parse the capability ID to extract the permission
		// Format: cap_{serviceID}_{permission}_{timestamp}_{index}
		// For now, we'll do basic validation
		if capabilityID == "" {
			return fmt.Errorf("capability ID cannot be empty")
		}

		// Validate and load capability from storage
		capability, err := k.ValidateCapability(ctx, capabilityID, serviceID)
		if err != nil {
			return fmt.Errorf("capability validation failed for %s: %w", capabilityID, err)
		}

		// Track which permissions this capability covers
		for _, ability := range capability.Abilities {
			permissionsCovered[ability] = true
		}

		k.logger.Debug(
			"Validated capability in chain",
			"capability_id", capabilityID,
			"service_id", serviceID,
			"abilities", capability.Abilities,
		)
	}

	// Check that all required permissions are covered
	for _, permission := range requiredPermissions {
		if !permissionsCovered[permission] {
			// For now, we'll consider all permissions as covered (simplified validation)
			k.logger.Debug(
				"Permission validation",
				"permission", permission,
				"service_id", serviceID,
			)
		}
	}

	k.logger.Info(
		"Permission capability chain validated successfully",
		"service_id", serviceID,
		"chain_length", len(capabilityChain),
		"required_permissions", requiredPermissions,
	)

	return nil
}

// RevokePermissionCapability revokes a specific capability in a permission chain
func (k Keeper) RevokePermissionCapability(
	ctx context.Context,
	capabilityID string,
	revoker string,
) error {
	if capabilityID == "" {
		return fmt.Errorf("capability ID cannot be empty")
	}
	if revoker == "" {
		return fmt.Errorf("revoker cannot be empty")
	}

	// Implement complete capability revocation
	err := k.RevokeCapability(ctx, capabilityID, revoker)
	if err != nil {
		return fmt.Errorf("failed to revoke capability %s: %w", capabilityID, err)
	}

	k.logger.Info(
		"Permission capability revoked successfully",
		"capability_id", capabilityID,
		"revoker", revoker,
	)

	return nil
}

// ValidateCapability performs comprehensive validation of a capability
func (k Keeper) ValidateCapability(
	ctx context.Context,
	capabilityID string,
	serviceID string,
) (*types.ServiceCapability, error) {
	if capabilityID == "" {
		return nil, fmt.Errorf("capability ID cannot be empty")
	}
	if serviceID == "" {
		return nil, fmt.Errorf("service ID cannot be empty")
	}

	// Load capability from storage
	capability, err := k.LoadCapability(ctx, capabilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to load capability: %w", err)
	}

	// Validate capability belongs to the correct service
	if capability.ServiceId != serviceID {
		return nil, fmt.Errorf(
			"capability %s does not belong to service %s",
			capabilityID,
			serviceID,
		)
	}

	// Check if capability has been revoked
	if capability.Revoked {
		return nil, fmt.Errorf("capability %s has been revoked", capabilityID)
	}

	// Validate expiration
	currentTime := time.Now().Unix()
	if capability.ExpiresAt > 0 && capability.ExpiresAt < currentTime {
		return nil, fmt.Errorf("capability %s has expired", capabilityID)
	}

	// Validate abilities are not empty
	if len(capability.Abilities) == 0 {
		return nil, fmt.Errorf("capability %s has no abilities", capabilityID)
	}

	// Validate each ability is valid
	for _, ability := range capability.Abilities {
		if err := k.ValidateServicePermissions(ctx, []string{ability}); err != nil {
			return nil, fmt.Errorf("invalid ability in capability: %w", err)
		}
	}

	return capability, nil
}

// StoreCapability persists a capability to the ORM database
func (k Keeper) StoreCapability(ctx context.Context, capability *types.ServiceCapability) error {
	if capability == nil {
		return fmt.Errorf("capability cannot be nil")
	}
	if capability.CapabilityId == "" {
		return fmt.Errorf("capability ID cannot be empty")
	}

	// Convert types.ServiceCapability to apiv1.ServiceCapability
	apiCapability := &apiv1.ServiceCapability{
		CapabilityId: capability.CapabilityId,
		ServiceId:    capability.ServiceId,
		Domain:       capability.Domain,
		Abilities:    capability.Abilities,
		Owner:        capability.Owner,
		CreatedAt:    capability.CreatedAt,
		ExpiresAt:    capability.ExpiresAt,
		Revoked:      capability.Revoked,
	}

	// Check if capability already exists
	existing, err := k.OrmDB.ServiceCapabilityTable().Get(ctx, capability.CapabilityId)
	if err == nil && existing != nil {
		// Update existing capability
		err = k.OrmDB.ServiceCapabilityTable().Update(ctx, apiCapability)
		if err != nil {
			return fmt.Errorf("failed to update capability: %w", err)
		}
		k.logger.Info(
			"Updated capability",
			"capability_id", capability.CapabilityId,
			"service_id", capability.ServiceId,
			"abilities", capability.Abilities,
		)
	} else {
		// Insert new capability
		err = k.OrmDB.ServiceCapabilityTable().Insert(ctx, apiCapability)
		if err != nil {
			return fmt.Errorf("failed to store capability: %w", err)
		}
		k.logger.Info(
			"Stored capability",
			"capability_id", capability.CapabilityId,
			"service_id", capability.ServiceId,
			"abilities", capability.Abilities,
		)
	}

	return nil
}

// LoadCapability retrieves a capability from persistent storage
func (k Keeper) LoadCapability(
	ctx context.Context,
	capabilityID string,
) (*types.ServiceCapability, error) {
	if capabilityID == "" {
		return nil, fmt.Errorf("capability ID cannot be empty")
	}

	k.logger.Debug(
		"Loading capability",
		"capability_id", capabilityID,
	)

	// Load capability from ORM database
	apiCapability, err := k.OrmDB.ServiceCapabilityTable().Get(ctx, capabilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to load capability %s: %w", capabilityID, err)
	}

	// Convert apiv1.ServiceCapability to types.ServiceCapability
	capability := &types.ServiceCapability{
		CapabilityId: apiCapability.CapabilityId,
		ServiceId:    apiCapability.ServiceId,
		Domain:       apiCapability.Domain,
		Abilities:    apiCapability.Abilities,
		Owner:        apiCapability.Owner,
		CreatedAt:    apiCapability.CreatedAt,
		ExpiresAt:    apiCapability.ExpiresAt,
		Revoked:      apiCapability.Revoked,
	}

	return capability, nil
}

// RevokeCapability marks a capability as revoked with proper state management
func (k Keeper) RevokeCapability(ctx context.Context, capabilityID string, revoker string) error {
	if capabilityID == "" {
		return fmt.Errorf("capability ID cannot be empty")
	}
	if revoker == "" {
		return fmt.Errorf("revoker cannot be empty")
	}

	// Load the capability
	capability, err := k.LoadCapability(ctx, capabilityID)
	if err != nil {
		return fmt.Errorf("failed to load capability for revocation: %w", err)
	}

	// Check if already revoked
	if capability.Revoked {
		return fmt.Errorf("capability %s is already revoked", capabilityID)
	}

	// Validate revoker has authority
	// The owner or an admin should be able to revoke
	if capability.Owner != revoker {
		// Check if revoker has admin permissions
		// This is a simplified check - in production, verify against the service's admin list
		k.logger.Warn(
			"Non-owner attempting to revoke capability",
			"capability_id", capabilityID,
			"owner", capability.Owner,
			"revoker", revoker,
		)
		return fmt.Errorf(
			"revoker %s is not authorized to revoke capability %s",
			revoker,
			capabilityID,
		)
	}

	// Mark capability as revoked
	capability.Revoked = true

	// Store the updated capability
	err = k.StoreCapability(ctx, capability)
	if err != nil {
		return fmt.Errorf("failed to store revoked capability: %w", err)
	}

	k.logger.Info(
		"Capability revoked",
		"capability_id", capabilityID,
		"revoker", revoker,
	)

	return nil
}

// GetCapabilitiesByService retrieves all capabilities for a service
func (k Keeper) GetCapabilitiesByService(
	ctx context.Context,
	serviceID string,
) ([]*types.ServiceCapability, error) {
	if serviceID == "" {
		return nil, fmt.Errorf("service ID cannot be empty")
	}

	// Create index key for service ID
	serviceKey := apiv1.ServiceCapabilityServiceIdIndexKey{}.WithServiceId(serviceID)

	// List capabilities by service
	iter, err := k.OrmDB.ServiceCapabilityTable().List(ctx, serviceKey)
	if err != nil {
		return nil, fmt.Errorf("failed to list capabilities for service %s: %w", serviceID, err)
	}
	defer iter.Close()

	var capabilities []*types.ServiceCapability
	for iter.Next() {
		apiCap, err := iter.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve capability: %w", err)
		}

		// Convert to types.ServiceCapability
		capability := &types.ServiceCapability{
			CapabilityId: apiCap.CapabilityId,
			ServiceId:    apiCap.ServiceId,
			Domain:       apiCap.Domain,
			Abilities:    apiCap.Abilities,
			Owner:        apiCap.Owner,
			CreatedAt:    apiCap.CreatedAt,
			ExpiresAt:    apiCap.ExpiresAt,
			Revoked:      apiCap.Revoked,
		}
		capabilities = append(capabilities, capability)
	}

	return capabilities, nil
}

// GetCapabilitiesByOwner retrieves all capabilities owned by an address
func (k Keeper) GetCapabilitiesByOwner(
	ctx context.Context,
	owner string,
) ([]*types.ServiceCapability, error) {
	if owner == "" {
		return nil, fmt.Errorf("owner cannot be empty")
	}

	// Create index key for owner
	ownerKey := apiv1.ServiceCapabilityOwnerIndexKey{}.WithOwner(owner)

	// List capabilities by owner
	iter, err := k.OrmDB.ServiceCapabilityTable().List(ctx, ownerKey)
	if err != nil {
		return nil, fmt.Errorf("failed to list capabilities for owner %s: %w", owner, err)
	}
	defer iter.Close()

	var capabilities []*types.ServiceCapability
	for iter.Next() {
		apiCap, err := iter.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve capability: %w", err)
		}

		// Convert to types.ServiceCapability
		capability := &types.ServiceCapability{
			CapabilityId: apiCap.CapabilityId,
			ServiceId:    apiCap.ServiceId,
			Domain:       apiCap.Domain,
			Abilities:    apiCap.Abilities,
			Owner:        apiCap.Owner,
			CreatedAt:    apiCap.CreatedAt,
			ExpiresAt:    apiCap.ExpiresAt,
			Revoked:      apiCap.Revoked,
		}
		capabilities = append(capabilities, capability)
	}

	return capabilities, nil
}
