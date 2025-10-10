package types

import (
	"fmt"
	"strings"

	"github.com/sonr-io/crypto/ucan"
)

// UCAN Action Constants for Service operations
const (
	// Core Service Actions
	UCANRegisterService   = "register-service"   // Register new service
	UCANUpdateService     = "update-service"     // Update service details
	UCANDeactivateService = "deactivate-service" // Deactivate service
	UCANDeleteService     = "delete-service"     // Delete service

	// Domain Verification Actions
	UCANInitiateDomainVerification = "initiate-domain-verification" // Start domain verification
	UCANVerifyDomain               = "verify-domain"                // Complete domain verification
	UCANRevokeDomainVerification   = "revoke-domain-verification"   // Revoke domain verification

	// Service Discovery Actions
	UCANQueryService   = "query-service"   // Query service details
	UCANListServices   = "list-services"   // List all services
	UCANSearchServices = "search-services" // Search services

	// Standard CRUD Actions (for compatibility)
	UCANCreate = "create" // Create service
	UCANRead   = "read"   // Read service
	UCANUpdate = "update" // Update service
	UCANDelete = "delete" // Delete service
	UCANAdmin  = "admin"  // Administrative actions
	UCANAll    = "*"      // Wildcard for all actions
)

// ServiceOperation represents the type of service operation being performed
type ServiceOperation string

const (
	ServiceOpRegister                   ServiceOperation = "register"
	ServiceOpUpdate                     ServiceOperation = "update"
	ServiceOpDeactivate                 ServiceOperation = "deactivate"
	ServiceOpDelete                     ServiceOperation = "delete"
	ServiceOpInitiateDomainVerification ServiceOperation = "initiate_domain_verification"
	ServiceOpVerifyDomain               ServiceOperation = "verify_domain"
	ServiceOpRevokeDomainVerification   ServiceOperation = "revoke_domain_verification"
	ServiceOpQuery                      ServiceOperation = "query"
	ServiceOpList                       ServiceOperation = "list"
	ServiceOpSearch                     ServiceOperation = "search"
)

// String returns the string representation of the service operation
func (op ServiceOperation) String() string {
	return string(op)
}

// UCANCapabilityMapper provides conversion between Service operations and UCAN capabilities
type UCANCapabilityMapper struct{}

// NewUCANCapabilityMapper creates a new capability mapper
func NewUCANCapabilityMapper() *UCANCapabilityMapper {
	return &UCANCapabilityMapper{}
}

// GetUCANCapabilitiesForOperation returns UCAN-specific capabilities for a Service operation
func (m *UCANCapabilityMapper) GetUCANCapabilitiesForOperation(operation ServiceOperation) []string {
	switch operation {
	case ServiceOpRegister:
		return []string{UCANRegisterService, UCANCreate}
	case ServiceOpUpdate:
		return []string{UCANUpdateService, UCANUpdate}
	case ServiceOpDeactivate:
		return []string{UCANDeactivateService, UCANUpdate}
	case ServiceOpDelete:
		return []string{UCANDeleteService, UCANDelete, UCANAdmin}

	case ServiceOpInitiateDomainVerification:
		return []string{UCANInitiateDomainVerification, UCANUpdate}
	case ServiceOpVerifyDomain:
		return []string{UCANVerifyDomain, UCANUpdate}
	case ServiceOpRevokeDomainVerification:
		return []string{UCANRevokeDomainVerification, UCANUpdate, UCANAdmin}

	case ServiceOpQuery:
		return []string{UCANQueryService, UCANRead}
	case ServiceOpList:
		return []string{UCANListServices, UCANRead}
	case ServiceOpSearch:
		return []string{UCANSearchServices, UCANRead}

	default:
		return []string{UCANRead} // Default to read permission
	}
}

// CreateServiceResourceURI builds a Service resource URI for UCAN validation
func (m *UCANCapabilityMapper) CreateServiceResourceURI(serviceID string) string {
	return fmt.Sprintf("svc:%s", serviceID)
}

// CreateDomainResourceURI builds a domain resource URI for UCAN validation
func (m *UCANCapabilityMapper) CreateDomainResourceURI(domain string) string {
	return fmt.Sprintf("domain:%s", domain)
}

// CreateServiceAttenuation creates a UCAN attenuation for Service operations
func (m *UCANCapabilityMapper) CreateServiceAttenuation(
	actions []string,
	serviceID string,
	caveats []string,
) ucan.Attenuation {
	resourceURI := m.CreateServiceResourceURI(serviceID)

	resource := &ucan.SimpleResource{
		Scheme: "svc",
		Value:  serviceID,
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

// CreateDomainBoundAttenuation creates a UCAN attenuation for domain-bound operations
func (m *UCANCapabilityMapper) CreateDomainBoundAttenuation(
	actions []string,
	domain string,
	serviceID string,
) ucan.Attenuation {
	// Create a domain-specific resource
	resourceURI := m.CreateDomainResourceURI(domain)

	resource := &ucan.SimpleResource{
		Scheme: "domain",
		Value:  domain,
		URI:    resourceURI,
	}

	// Use MultiCapability for domain-bound operations
	// Note: domain binding is enforced through resource matching
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

// CreateRateLimitedAttenuation creates a UCAN attenuation with rate limiting
func (m *UCANCapabilityMapper) CreateRateLimitedAttenuation(
	actions []string,
	serviceID string,
	rateLimit uint64,
	windowSeconds uint64,
) ucan.Attenuation {
	// Rate limiting will be handled at validation layer
	// For now, create a standard service attenuation
	return m.CreateServiceAttenuation(actions, serviceID, nil)
}

// ValidateUCANCapabilities validates that a UCAN capability grants the required Service actions
func (m *UCANCapabilityMapper) ValidateUCANCapabilities(
	capability ucan.Capability,
	requiredActions []string,
) bool {
	return capability.Grants(requiredActions)
}

// ValidateDomainBoundCapability validates domain-bound UCAN capabilities
func (m *UCANCapabilityMapper) ValidateDomainBoundCapability(
	capability ucan.Capability,
	domain string,
) error {
	// Domain validation will be handled through resource matching
	// Check if capability has appropriate actions for domain operations
	actions := capability.GetActions()
	for _, action := range actions {
		if action == UCANInitiateDomainVerification ||
			action == UCANVerifyDomain ||
			action == UCANRevokeDomainVerification {
			return nil
		}
	}

	return fmt.Errorf("capability does not include domain verification actions")
}

// IsUCANAction checks if an action string is a valid UCAN action
func IsUCANAction(action string) bool {
	validActions := []string{
		UCANRegisterService, UCANUpdateService, UCANDeactivateService, UCANDeleteService,
		UCANInitiateDomainVerification, UCANVerifyDomain, UCANRevokeDomainVerification,
		UCANQueryService, UCANListServices, UCANSearchServices,
		UCANCreate, UCANRead, UCANUpdate, UCANDelete, UCANAdmin, UCANAll,
	}

	for _, validAction := range validActions {
		if action == validAction {
			return true
		}
	}
	return false
}

// GetServiceCapabilityTemplate returns a preconfigured capability template for Service
func GetServiceCapabilityTemplate() *ucan.CapabilityTemplate {
	return ucan.StandardServiceTemplate()
}

// UCANPermissionRegistry extends the basic permission registry with UCAN capabilities
type UCANPermissionRegistry struct {
	operationCapabilities map[ServiceOperation][]string
	mapper                *UCANCapabilityMapper
}

// NewUCANPermissionRegistry creates a new UCAN-aware permission registry
func NewUCANPermissionRegistry() *UCANPermissionRegistry {
	registry := &UCANPermissionRegistry{
		operationCapabilities: make(map[ServiceOperation][]string),
		mapper:                NewUCANCapabilityMapper(),
	}

	// Initialize default capabilities
	registry.initializeDefaultCapabilities()
	return registry
}

// initializeDefaultCapabilities sets up default capability mappings
func (r *UCANPermissionRegistry) initializeDefaultCapabilities() {
	operations := []ServiceOperation{
		ServiceOpRegister, ServiceOpUpdate, ServiceOpDeactivate, ServiceOpDelete,
		ServiceOpInitiateDomainVerification, ServiceOpVerifyDomain, ServiceOpRevokeDomainVerification,
		ServiceOpQuery, ServiceOpList, ServiceOpSearch,
	}

	for _, op := range operations {
		r.operationCapabilities[op] = r.mapper.GetUCANCapabilitiesForOperation(op)
	}
}

// GetRequiredUCANCapabilities returns UCAN-specific capabilities for a Service operation
func (r *UCANPermissionRegistry) GetRequiredUCANCapabilities(operation ServiceOperation) ([]string, error) {
	capabilities, exists := r.operationCapabilities[operation]
	if !exists {
		capabilities = r.mapper.GetUCANCapabilitiesForOperation(operation)
	}

	if len(capabilities) == 0 {
		return nil, fmt.Errorf("no UCAN capabilities defined for operation: %s", operation.String())
	}
	return capabilities, nil
}

// CreateServiceAttenuation creates a UCAN attenuation for Service operations
func (r *UCANPermissionRegistry) CreateServiceAttenuation(
	actions []string,
	serviceID string,
	caveats []string,
) ucan.Attenuation {
	return r.mapper.CreateServiceAttenuation(actions, serviceID, caveats)
}

// CreateDomainBoundAttenuation creates a domain-bound attenuation
func (r *UCANPermissionRegistry) CreateDomainBoundAttenuation(
	actions []string,
	domain string,
	serviceID string,
) ucan.Attenuation {
	return r.mapper.CreateDomainBoundAttenuation(actions, domain, serviceID)
}

// CreateRateLimitedAttenuation creates a rate-limited attenuation
func (r *UCANPermissionRegistry) CreateRateLimitedAttenuation(
	actions []string,
	serviceID string,
	rateLimit uint64,
	windowSeconds uint64,
) ucan.Attenuation {
	return r.mapper.CreateRateLimitedAttenuation(actions, serviceID, rateLimit, windowSeconds)
}

// Helper functions

// CreateServiceResourcePattern creates a Service resource pattern for matching
func CreateServiceResourcePattern(serviceType, serviceID string) string {
	if serviceID == "*" {
		return fmt.Sprintf("%s:*", serviceType)
	}
	return fmt.Sprintf("%s:%s", serviceType, serviceID)
}

// MatchesServicePattern checks if a service ID matches a given pattern
func MatchesServicePattern(serviceID, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Handle wildcard patterns like "api:*"
	if strings.HasSuffix(pattern, ":*") {
		prefix := strings.TrimSuffix(pattern, ":*")
		return strings.HasPrefix(serviceID, prefix+":")
	}

	return serviceID == pattern
}

// CreateGaslessServiceAttenuation creates a UCAN attenuation that supports gasless transactions
func CreateGaslessServiceAttenuation(
	actions []string,
	serviceID string,
	gasLimit uint64,
) ucan.Attenuation {
	mapper := NewUCANCapabilityMapper()
	baseAttenuation := mapper.CreateServiceAttenuation(actions, serviceID, nil)

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

// Domain verification helpers

// CreateDomainVerificationURI creates a resource URI for domain verification operations
func CreateDomainVerificationURI(domain string, verificationMethod string) string {
	return fmt.Sprintf("domain:%s/verify/%s", domain, verificationMethod)
}

// ValidateDomainVerificationCapability validates domain verification capability
func ValidateDomainVerificationCapability(
	capability ucan.Capability,
	domain string,
	verificationMethod string,
) error {
	// Check for domain verification actions
	actions := capability.GetActions()
	hasVerificationAction := false
	for _, action := range actions {
		if action == UCANInitiateDomainVerification || action == UCANVerifyDomain {
			hasVerificationAction = true
			break
		}
	}

	if !hasVerificationAction {
		return fmt.Errorf("capability does not include domain verification actions")
	}

	// Domain validation will be handled through resource matching
	return nil
}
