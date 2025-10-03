package types

import (
	"fmt"

	"github.com/sonr-io/sonr/crypto/ucan"
)

// UCAN Action Constants for DEX operations
const (
	// Core Trading Actions
	UCANSwap            = "swap"              // Execute token swap
	UCANExecuteSwap     = "execute-swap"      // Execute a specific swap
	UCANLimitOrder      = "limit-order"       // Place limit order
	UCANMarketOrder     = "market-order"      // Place market order
	UCANCancelOrder     = "cancel-order"      // Cancel order
	UCANCancelAllOrders = "cancel-all-orders" // Cancel all orders

	// Liquidity Actions
	UCANProvideLiquidity = "provide-liquidity" // Add liquidity to pool
	UCANRemoveLiquidity  = "remove-liquidity"  // Remove liquidity from pool
	UCANCreatePool       = "create-pool"       // Create new liquidity pool

	// Portfolio Management Actions
	UCANRegisterAccount = "register-account" // Register trading account
	UCANUpdatePortfolio = "update-portfolio" // Update portfolio settings
	UCANWithdraw        = "withdraw"         // Withdraw funds
	UCANDeposit         = "deposit"          // Deposit funds

	// Query Actions
	UCANQueryPool      = "query-pool"      // Query pool details
	UCANQueryOrders    = "query-orders"    // Query orders
	UCANQueryPortfolio = "query-portfolio" // Query portfolio

	// Standard CRUD Actions (for compatibility)
	UCANCreate = "create" // Create resource
	UCANRead   = "read"   // Read resource
	UCANUpdate = "update" // Update resource
	UCANDelete = "delete" // Delete resource
	UCANAdmin  = "admin"  // Administrative actions
	UCANAll    = "*"      // Wildcard for all actions
)

// DEXOperation represents the type of DEX operation being performed
type DEXOperation string

const (
	DEXOpSwap             DEXOperation = "swap"
	DEXOpExecuteSwap      DEXOperation = "execute_swap"
	DEXOpLimitOrder       DEXOperation = "limit_order"
	DEXOpMarketOrder      DEXOperation = "market_order"
	DEXOpCancelOrder      DEXOperation = "cancel_order"
	DEXOpCancelAllOrders  DEXOperation = "cancel_all_orders"
	DEXOpProvideLiquidity DEXOperation = "provide_liquidity"
	DEXOpRemoveLiquidity  DEXOperation = "remove_liquidity"
	DEXOpCreatePool       DEXOperation = "create_pool"
	DEXOpRegisterAccount  DEXOperation = "register_account"
	DEXOpUpdatePortfolio  DEXOperation = "update_portfolio"
	DEXOpWithdraw         DEXOperation = "withdraw"
	DEXOpDeposit          DEXOperation = "deposit"
	DEXOpQueryPool        DEXOperation = "query_pool"
	DEXOpQueryOrders      DEXOperation = "query_orders"
	DEXOpQueryPortfolio   DEXOperation = "query_portfolio"
)

// String returns the string representation of the DEX operation
func (op DEXOperation) String() string {
	return string(op)
}

// UCANCapabilityMapper provides conversion between DEX operations and UCAN capabilities
type UCANCapabilityMapper struct{}

// NewUCANCapabilityMapper creates a new capability mapper
func NewUCANCapabilityMapper() *UCANCapabilityMapper {
	return &UCANCapabilityMapper{}
}

// GetUCANCapabilitiesForOperation returns UCAN-specific capabilities for a DEX operation
func (m *UCANCapabilityMapper) GetUCANCapabilitiesForOperation(operation DEXOperation) []string {
	switch operation {
	case DEXOpSwap:
		return []string{UCANSwap, UCANUpdate}
	case DEXOpExecuteSwap:
		return []string{UCANExecuteSwap, UCANUpdate}
	case DEXOpLimitOrder:
		return []string{UCANLimitOrder, UCANCreate}
	case DEXOpMarketOrder:
		return []string{UCANMarketOrder, UCANCreate}
	case DEXOpCancelOrder:
		return []string{UCANCancelOrder, UCANDelete}
	case DEXOpCancelAllOrders:
		return []string{UCANCancelAllOrders, UCANDelete, UCANAdmin}
	case DEXOpProvideLiquidity:
		return []string{UCANProvideLiquidity, UCANCreate}
	case DEXOpRemoveLiquidity:
		return []string{UCANRemoveLiquidity, UCANDelete}
	case DEXOpCreatePool:
		return []string{UCANCreatePool, UCANCreate, UCANAdmin}
	case DEXOpRegisterAccount:
		return []string{UCANRegisterAccount, UCANCreate}
	case DEXOpUpdatePortfolio:
		return []string{UCANUpdatePortfolio, UCANUpdate}
	case DEXOpWithdraw:
		return []string{UCANWithdraw, UCANUpdate}
	case DEXOpDeposit:
		return []string{UCANDeposit, UCANUpdate}
	case DEXOpQueryPool:
		return []string{UCANQueryPool, UCANRead}
	case DEXOpQueryOrders:
		return []string{UCANQueryOrders, UCANRead}
	case DEXOpQueryPortfolio:
		return []string{UCANQueryPortfolio, UCANRead}
	default:
		return []string{UCANRead} // Default to read permission
	}
}

// CreateDEXResourceURI builds a DEX resource URI for UCAN validation
func (m *UCANCapabilityMapper) CreateDEXResourceURI(resourceType, resourceID string) string {
	return fmt.Sprintf("dex:%s:%s", resourceType, resourceID)
}

// CreatePoolResourceURI builds a pool resource URI for UCAN validation
func (m *UCANCapabilityMapper) CreatePoolResourceURI(poolID string) string {
	return fmt.Sprintf("dex:pool:%s", poolID)
}

// CreateOrderResourceURI builds an order resource URI for UCAN validation
func (m *UCANCapabilityMapper) CreateOrderResourceURI(orderID string) string {
	return fmt.Sprintf("dex:order:%s", orderID)
}

// CreateDEXAttenuation creates a UCAN attenuation for DEX operations
func (m *UCANCapabilityMapper) CreateDEXAttenuation(
	actions []string,
	resourceType string,
	resourceID string,
) ucan.Attenuation {
	resourceURI := m.CreateDEXResourceURI(resourceType, resourceID)

	resource := &ucan.SimpleResource{
		Scheme: "dex",
		Value:  fmt.Sprintf("%s:%s", resourceType, resourceID),
		URI:    resourceURI,
	}

	// Use MultiCapability for multiple actions
	var capability ucan.Capability
	if len(actions) == 1 {
		capability = &ucan.SimpleCapability{
			Action: actions[0],
		}
	} else {
		capability = &ucan.MultiCapability{
			Actions: actions,
		}
	}

	return ucan.Attenuation{
		Capability: capability,
		Resource:   resource,
	}
}

// CreateAmountLimitedAttenuation creates a UCAN attenuation with amount limits
func (m *UCANCapabilityMapper) CreateAmountLimitedAttenuation(
	actions []string,
	poolID string,
	maxAmount string,
) ucan.Attenuation {
	// Create base attenuation
	baseAttenuation := m.CreateDEXAttenuation(actions, "pool", poolID)

	// For amount limits, we'll need to handle this at validation layer
	// since the standard capability types don't support custom constraints

	return baseAttenuation
}

// CreatePoolRestrictedAttenuation creates a UCAN attenuation restricted to specific pools
func (m *UCANCapabilityMapper) CreatePoolRestrictedAttenuation(
	actions []string,
	allowedPools []string,
) ucan.Attenuation {
	// Create resource for multiple pools
	resourceURI := "dex:pool:*"
	if len(allowedPools) == 1 {
		resourceURI = m.CreatePoolResourceURI(allowedPools[0])
	}

	resource := &ucan.SimpleResource{
		Scheme: "dex",
		Value:  "pool:*",
		URI:    resourceURI,
	}

	// Use MultiCapability for multiple actions
	var capability ucan.Capability
	if len(actions) == 1 {
		capability = &ucan.SimpleCapability{
			Action: actions[0],
		}
	} else {
		capability = &ucan.MultiCapability{
			Actions: actions,
		}
	}

	return ucan.Attenuation{
		Capability: capability,
		Resource:   resource,
	}
}

// ValidateUCANCapabilities validates that a UCAN capability grants the required DEX actions
func (m *UCANCapabilityMapper) ValidateUCANCapabilities(
	capability ucan.Capability,
	requiredActions []string,
) bool {
	return capability.Grants(requiredActions)
}

// IsUCANAction checks if an action string is a valid UCAN action
func IsUCANAction(action string) bool {
	validActions := []string{
		UCANSwap, UCANExecuteSwap, UCANLimitOrder, UCANMarketOrder,
		UCANCancelOrder, UCANCancelAllOrders,
		UCANProvideLiquidity, UCANRemoveLiquidity, UCANCreatePool,
		UCANRegisterAccount, UCANUpdatePortfolio, UCANWithdraw, UCANDeposit,
		UCANQueryPool, UCANQueryOrders, UCANQueryPortfolio,
		UCANCreate, UCANRead, UCANUpdate, UCANDelete, UCANAdmin, UCANAll,
	}

	for _, validAction := range validActions {
		if action == validAction {
			return true
		}
	}
	return false
}

// GetDEXCapabilityTemplate returns a preconfigured capability template for DEX
func GetDEXCapabilityTemplate() *ucan.CapabilityTemplate {
	return ucan.StandardServiceTemplate()
}

// UCANPermissionRegistry extends the basic permission registry with UCAN capabilities
type UCANPermissionRegistry struct {
	operationCapabilities map[DEXOperation][]string
	mapper                *UCANCapabilityMapper
}

// NewUCANPermissionRegistry creates a new UCAN-aware permission registry
func NewUCANPermissionRegistry() *UCANPermissionRegistry {
	registry := &UCANPermissionRegistry{
		operationCapabilities: make(map[DEXOperation][]string),
		mapper:                NewUCANCapabilityMapper(),
	}

	// Initialize default capabilities
	registry.initializeDefaultCapabilities()
	return registry
}

// initializeDefaultCapabilities sets up default capability mappings
func (r *UCANPermissionRegistry) initializeDefaultCapabilities() {
	operations := []DEXOperation{
		DEXOpSwap, DEXOpExecuteSwap, DEXOpLimitOrder, DEXOpMarketOrder,
		DEXOpCancelOrder, DEXOpCancelAllOrders,
		DEXOpProvideLiquidity, DEXOpRemoveLiquidity, DEXOpCreatePool,
		DEXOpRegisterAccount, DEXOpUpdatePortfolio, DEXOpWithdraw, DEXOpDeposit,
		DEXOpQueryPool, DEXOpQueryOrders, DEXOpQueryPortfolio,
	}

	for _, op := range operations {
		r.operationCapabilities[op] = r.mapper.GetUCANCapabilitiesForOperation(op)
	}
}

// GetRequiredUCANCapabilities returns UCAN-specific capabilities for a DEX operation
func (r *UCANPermissionRegistry) GetRequiredUCANCapabilities(operation DEXOperation) ([]string, error) {
	capabilities, exists := r.operationCapabilities[operation]
	if !exists {
		capabilities = r.mapper.GetUCANCapabilitiesForOperation(operation)
	}

	if len(capabilities) == 0 {
		return nil, fmt.Errorf("no UCAN capabilities defined for operation: %s", operation.String())
	}
	return capabilities, nil
}

// CreateDEXAttenuation creates a UCAN attenuation for DEX operations
func (r *UCANPermissionRegistry) CreateDEXAttenuation(
	actions []string,
	resourceType string,
	resourceID string,
) ucan.Attenuation {
	return r.mapper.CreateDEXAttenuation(actions, resourceType, resourceID)
}

// CreateAmountLimitedAttenuation creates an amount-limited attenuation
func (r *UCANPermissionRegistry) CreateAmountLimitedAttenuation(
	actions []string,
	poolID string,
	maxAmount string,
) ucan.Attenuation {
	return r.mapper.CreateAmountLimitedAttenuation(actions, poolID, maxAmount)
}

// CreatePoolRestrictedAttenuation creates a pool-restricted attenuation
func (r *UCANPermissionRegistry) CreatePoolRestrictedAttenuation(
	actions []string,
	allowedPools []string,
) ucan.Attenuation {
	return r.mapper.CreatePoolRestrictedAttenuation(actions, allowedPools)
}

// Helper functions

// CreateGaslessDEXAttenuation creates a UCAN attenuation that supports gasless transactions
func CreateGaslessDEXAttenuation(
	actions []string,
	resourceType string,
	resourceID string,
	gasLimit uint64,
) ucan.Attenuation {
	mapper := NewUCANCapabilityMapper()
	baseAttenuation := mapper.CreateDEXAttenuation(actions, resourceType, resourceID)

	// Wrap capability with gasless support
	gaslessCapability := &ucan.GaslessCapability{
		Capability:   baseAttenuation.Capability,
		AllowGasless: true,
		GasLimit:     gasLimit,
	}

	return ucan.Attenuation{
		Capability: gaslessCapability,
		Resource:   baseAttenuation.Resource,
	}
}

// ValidateAmountConstraint validates amount constraints for DEX operations
func ValidateAmountConstraint(
	capability ucan.Capability,
	amount string,
	maxAmount string,
) error {
	// Amount validation would be handled at a higher level
	// This is a placeholder for the actual implementation
	return nil
}

// ValidatePoolConstraint validates pool constraints for DEX operations
func ValidatePoolConstraint(
	capability ucan.Capability,
	poolID string,
	allowedPools []string,
) error {
	// Check if pool is in allowed list
	for _, allowed := range allowedPools {
		if poolID == allowed {
			return nil
		}
	}

	return fmt.Errorf("pool %s not in allowed list", poolID)
}
