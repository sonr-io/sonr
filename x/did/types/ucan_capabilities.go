package types

import (
	"fmt"
	"strings"

	"github.com/sonr-io/sonr/crypto/ucan"
)

// UCAN Action Constants for DID operations
const (
	// Core DID Actions
	UCANCreate     = "create"     // Create new DID document
	UCANRegister   = "register"   // Register DID with controller
	UCANUpdate     = "update"     // Update DID document
	UCANDeactivate = "deactivate" // Deactivate DID document
	UCANRevoke     = "revoke"     // Revoke DID document (stronger than deactivate)

	// Verification Method Actions
	UCANAddVerificationMethod    = "add-verification-method"    // Add verification method
	UCANRemoveVerificationMethod = "remove-verification-method" // Remove verification method

	// Service Actions
	UCANAddService    = "add-service"    // Add service endpoint
	UCANRemoveService = "remove-service" // Remove service endpoint

	// Credential Actions
	UCANIssueCredential  = "issue-credential"  // Issue verifiable credential
	UCANRevokeCredential = "revoke-credential" // Revoke verifiable credential

	// External Wallet Actions
	UCANLinkWallet = "link-wallet" // Link external wallet

	// WebAuthn Actions
	UCANRegisterWebAuthn = "register-webauthn" // Register WebAuthn credential

	// Standard CRUD Actions (for compatibility)
	UCANRead   = "read"   // Read DID document
	UCANDelete = "delete" // Delete (same as revoke)
	UCANAdmin  = "admin"  // Administrative actions
	UCANAll    = "*"      // Wildcard for all actions
)

// DIDOperation represents the type of DID operation being performed
type DIDOperation string

const (
	DIDOpCreate                   DIDOperation = "create"
	DIDOpRegister                 DIDOperation = "register"
	DIDOpUpdate                   DIDOperation = "update"
	DIDOpDeactivate               DIDOperation = "deactivate"
	DIDOpRevoke                   DIDOperation = "revoke"
	DIDOpAddVerificationMethod    DIDOperation = "add_verification_method"
	DIDOpRemoveVerificationMethod DIDOperation = "remove_verification_method"
	DIDOpAddService               DIDOperation = "add_service"
	DIDOpRemoveService            DIDOperation = "remove_service"
	DIDOpIssueCredential          DIDOperation = "issue_credential"
	DIDOpRevokeCredential         DIDOperation = "revoke_credential"
	DIDOpLinkWallet               DIDOperation = "link_wallet"
	DIDOpRegisterWebAuthn         DIDOperation = "register_webauthn"
)

// String returns the string representation of the DID operation
func (op DIDOperation) String() string {
	return string(op)
}

// UCANCapabilityMapper provides conversion between DID operations and UCAN capabilities
type UCANCapabilityMapper struct{}

// NewUCANCapabilityMapper creates a new capability mapper
func NewUCANCapabilityMapper() *UCANCapabilityMapper {
	return &UCANCapabilityMapper{}
}

// GetUCANCapabilitiesForOperation returns UCAN-specific capabilities for a DID operation
func (m *UCANCapabilityMapper) GetUCANCapabilitiesForOperation(operation DIDOperation) []string {
	switch operation {
	case DIDOpCreate:
		return []string{UCANCreate}
	case DIDOpRegister:
		return []string{UCANRegister, UCANCreate}
	case DIDOpUpdate:
		return []string{UCANUpdate}
	case DIDOpDeactivate:
		return []string{UCANDeactivate, UCANUpdate}
	case DIDOpRevoke:
		return []string{UCANRevoke, UCANDelete, UCANAdmin}

	case DIDOpAddVerificationMethod:
		return []string{UCANAddVerificationMethod, UCANUpdate}
	case DIDOpRemoveVerificationMethod:
		return []string{UCANRemoveVerificationMethod, UCANUpdate}

	case DIDOpAddService:
		return []string{UCANAddService, UCANUpdate}
	case DIDOpRemoveService:
		return []string{UCANRemoveService, UCANUpdate}

	case DIDOpIssueCredential:
		return []string{UCANIssueCredential, UCANCreate}
	case DIDOpRevokeCredential:
		return []string{UCANRevokeCredential, UCANDelete}

	case DIDOpLinkWallet:
		return []string{UCANLinkWallet, UCANUpdate}
	case DIDOpRegisterWebAuthn:
		return []string{UCANRegisterWebAuthn, UCANCreate}

	default:
		return []string{UCANRead} // Default to read permission
	}
}

// CreateDIDResourceURI builds a DID resource URI for UCAN validation
func (m *UCANCapabilityMapper) CreateDIDResourceURI(didPattern string) string {
	return fmt.Sprintf("did:%s", didPattern)
}

// CreateDIDAttenuation creates a UCAN attenuation for DID operations
func (m *UCANCapabilityMapper) CreateDIDAttenuation(
	actions []string,
	didPattern string,
	caveats []string,
) ucan.Attenuation {
	resourceURI := m.CreateDIDResourceURI(didPattern)

	// Extract method and subject from DID pattern
	didMethod, didSubject := parseDIDPattern(didPattern)

	resource := &ucan.DIDResource{
		SimpleResource: ucan.SimpleResource{
			Scheme: "did",
			Value:  didPattern,
			URI:    resourceURI,
		},
		DIDMethod:  didMethod,
		DIDSubject: didSubject,
	}

	capability := &ucan.DIDCapability{
		Actions: actions,
		Caveats: caveats,
	}

	return ucan.Attenuation{
		Capability: capability,
		Resource:   resource,
	}
}

// CreateControllerAttenuation creates a UCAN attenuation for controller-specific operations
func (m *UCANCapabilityMapper) CreateControllerAttenuation(
	actions []string,
	didPattern string,
	controllerAddress string,
) ucan.Attenuation {
	caveats := []string{"controller"}
	attenuation := m.CreateDIDAttenuation(actions, didPattern, caveats)

	// Add controller metadata to the DID resource
	if resource, ok := attenuation.Resource.(*ucan.DIDResource); ok {
		if resource.Metadata == nil {
			resource.Metadata = make(map[string]string)
		}
		resource.Metadata["controller"] = controllerAddress
	}

	return attenuation
}

// CreateOwnerAttenuation creates a UCAN attenuation for owner-specific operations
func (m *UCANCapabilityMapper) CreateOwnerAttenuation(
	actions []string,
	didPattern string,
	ownerAddress string,
) ucan.Attenuation {
	caveats := []string{"owner"}
	attenuation := m.CreateDIDAttenuation(actions, didPattern, caveats)

	// Add owner metadata to the DID resource
	if resource, ok := attenuation.Resource.(*ucan.DIDResource); ok {
		if resource.Metadata == nil {
			resource.Metadata = make(map[string]string)
		}
		resource.Metadata["owner"] = ownerAddress
	}

	return attenuation
}

// CreateWebAuthnDelegationAttenuation creates UCAN attenuation for WebAuthn credential delegation
func (m *UCANCapabilityMapper) CreateWebAuthnDelegationAttenuation(
	actions []string,
	didPattern string,
	credentialID string,
) ucan.Attenuation {
	caveats := []string{"webauthn-delegation"}
	attenuation := m.CreateDIDAttenuation(actions, didPattern, caveats)

	// Add WebAuthn metadata to the capability
	if capability, ok := attenuation.Capability.(*ucan.DIDCapability); ok {
		if capability.Metadata == nil {
			capability.Metadata = make(map[string]string)
		}
		capability.Metadata["webauthn_credential_id"] = credentialID
		capability.Metadata["delegation_type"] = "webauthn"
	}

	return attenuation
}

// ValidateUCANCapabilities validates that a UCAN capability grants the required DID actions
func (m *UCANCapabilityMapper) ValidateUCANCapabilities(
	capability ucan.Capability,
	requiredActions []string,
) bool {
	return capability.Grants(requiredActions)
}

// ConvertLegacyCapabilities converts old string-based capabilities to UCAN format
func (m *UCANCapabilityMapper) ConvertLegacyCapabilities(legacyCapabilities []string) []string {
	var ucanCapabilities []string

	for _, legacy := range legacyCapabilities {
		switch strings.ToLower(legacy) {
		case "create":
			ucanCapabilities = append(ucanCapabilities, UCANCreate)
		case "register":
			ucanCapabilities = append(ucanCapabilities, UCANRegister)
		case "update":
			ucanCapabilities = append(ucanCapabilities, UCANUpdate)
		case "deactivate":
			ucanCapabilities = append(ucanCapabilities, UCANDeactivate)
		case "revoke":
			ucanCapabilities = append(ucanCapabilities, UCANRevoke)
		case "read", "get":
			ucanCapabilities = append(ucanCapabilities, UCANRead)
		case "delete":
			ucanCapabilities = append(ucanCapabilities, UCANDelete)
		case "admin":
			ucanCapabilities = append(ucanCapabilities, UCANAdmin)
		case "*":
			ucanCapabilities = append(ucanCapabilities, UCANAll)
		default:
			// Pass through unknown capabilities
			ucanCapabilities = append(ucanCapabilities, legacy)
		}
	}

	return ucanCapabilities
}

// IsUCANAction checks if an action string is a valid UCAN action
func IsUCANAction(action string) bool {
	validActions := []string{
		UCANCreate, UCANRegister, UCANUpdate, UCANDeactivate, UCANRevoke,
		UCANAddVerificationMethod, UCANRemoveVerificationMethod,
		UCANAddService, UCANRemoveService,
		UCANIssueCredential, UCANRevokeCredential,
		UCANLinkWallet, UCANRegisterWebAuthn,
		UCANRead, UCANDelete, UCANAdmin, UCANAll,
	}

	for _, validAction := range validActions {
		if action == validAction {
			return true
		}
	}
	return false
}

// GetDIDCapabilityTemplate returns a preconfigured capability template for DID
func GetDIDCapabilityTemplate() *ucan.CapabilityTemplate {
	return ucan.StandardDIDTemplate()
}

// UCANPermissionRegistry extends the basic permission registry with UCAN capabilities
type UCANPermissionRegistry struct {
	operationCapabilities map[DIDOperation][]string
	mapper                *UCANCapabilityMapper
}

// NewUCANPermissionRegistry creates a new UCAN-aware permission registry
func NewUCANPermissionRegistry() *UCANPermissionRegistry {
	registry := &UCANPermissionRegistry{
		operationCapabilities: make(map[DIDOperation][]string),
		mapper:                NewUCANCapabilityMapper(),
	}

	// Initialize default capabilities
	registry.initializeDefaultCapabilities()
	return registry
}

// initializeDefaultCapabilities sets up default capability mappings
func (r *UCANPermissionRegistry) initializeDefaultCapabilities() {
	operations := []DIDOperation{
		DIDOpCreate, DIDOpRegister, DIDOpUpdate, DIDOpDeactivate, DIDOpRevoke,
		DIDOpAddVerificationMethod, DIDOpRemoveVerificationMethod,
		DIDOpAddService, DIDOpRemoveService,
		DIDOpIssueCredential, DIDOpRevokeCredential,
		DIDOpLinkWallet, DIDOpRegisterWebAuthn,
	}

	for _, op := range operations {
		r.operationCapabilities[op] = r.mapper.GetUCANCapabilitiesForOperation(op)
	}
}

// GetRequiredUCANCapabilities returns UCAN-specific capabilities for a DID operation
func (r *UCANPermissionRegistry) GetRequiredUCANCapabilities(operation DIDOperation) ([]string, error) {
	capabilities, exists := r.operationCapabilities[operation]
	if !exists {
		capabilities = r.mapper.GetUCANCapabilitiesForOperation(operation)
	}

	if len(capabilities) == 0 {
		return nil, fmt.Errorf("no UCAN capabilities defined for operation: %s", operation.String())
	}
	return capabilities, nil
}

// CreateDIDAttenuation creates a UCAN attenuation for DID operations
func (r *UCANPermissionRegistry) CreateDIDAttenuation(
	actions []string,
	didPattern string,
	caveats []string,
) ucan.Attenuation {
	return r.mapper.CreateDIDAttenuation(actions, didPattern, caveats)
}

// CreateControllerAttenuation creates a controller-specific attenuation
func (r *UCANPermissionRegistry) CreateControllerAttenuation(
	actions []string,
	didPattern string,
	controllerAddress string,
) ucan.Attenuation {
	return r.mapper.CreateControllerAttenuation(actions, didPattern, controllerAddress)
}

// CreateWebAuthnDelegationAttenuation creates WebAuthn delegation attenuation
func (r *UCANPermissionRegistry) CreateWebAuthnDelegationAttenuation(
	actions []string,
	didPattern string,
	credentialID string,
) ucan.Attenuation {
	return r.mapper.CreateWebAuthnDelegationAttenuation(actions, didPattern, credentialID)
}

// Helper functions

// parseDIDPattern extracts method and subject from a DID pattern
func parseDIDPattern(didPattern string) (method, subject string) {
	// Handle patterns like "sonr:alice" or "key:z6MkV..."
	parts := strings.SplitN(didPattern, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	// If no colon, treat entire pattern as subject with default method
	return "sonr", didPattern
}

// CreateDIDResourcePattern creates a DID resource pattern for matching
func CreateDIDResourcePattern(method, subject string) string {
	if subject == "*" {
		return fmt.Sprintf("%s:*", method)
	}
	return fmt.Sprintf("%s:%s", method, subject)
}

// MatchesDIDPattern checks if a DID matches a given pattern
func MatchesDIDPattern(did, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Extract DID components
	didParts := strings.SplitN(did, ":", 3) // ["did", "method", "subject"]
	if len(didParts) != 3 {
		return false
	}

	// Extract pattern components
	patternParts := strings.SplitN(pattern, ":", 2) // ["method", "subject"]
	if len(patternParts) != 2 {
		return false
	}

	didMethod := didParts[1]
	didSubject := didParts[2]
	patternMethod := patternParts[0]
	patternSubject := patternParts[1]

	// Check method match
	if patternMethod != "*" && patternMethod != didMethod {
		return false
	}

	// Check subject match
	if patternSubject != "*" && patternSubject != didSubject {
		return false
	}

	return true
}

// CreateGaslessAttenuation creates a UCAN attenuation that supports gasless transactions
func CreateGaslessAttenuation(
	actions []string,
	didPattern string,
	gasLimit uint64,
) ucan.Attenuation {
	mapper := NewUCANCapabilityMapper()
	baseAttenuation := mapper.CreateDIDAttenuation(actions, didPattern, nil)

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

// WebAuthn-specific helpers

// CreateWebAuthnResourceURI creates a resource URI for WebAuthn operations
func CreateWebAuthnResourceURI(did, credentialID string) string {
	return fmt.Sprintf("did:%s/webauthn/%s", strings.TrimPrefix(did, "did:"), credentialID)
}

// ValidateWebAuthnDelegation validates WebAuthn capability delegation
func ValidateWebAuthnDelegation(
	capability ucan.Capability,
	credentialID string,
) error {
	didCapability, ok := capability.(*ucan.DIDCapability)
	if !ok {
		return fmt.Errorf("capability is not a DID capability")
	}

	// Check for WebAuthn delegation caveat
	hasWebAuthnCaveat := false
	for _, caveat := range didCapability.Caveats {
		if caveat == "webauthn-delegation" {
			hasWebAuthnCaveat = true
			break
		}
	}

	if !hasWebAuthnCaveat {
		return fmt.Errorf("capability does not include WebAuthn delegation caveat")
	}

	// Validate credential ID in metadata
	if didCapability.Metadata == nil {
		return fmt.Errorf("missing WebAuthn metadata")
	}

	storedCredentialID, exists := didCapability.Metadata["webauthn_credential_id"]
	if !exists {
		return fmt.Errorf("missing WebAuthn credential ID in metadata")
	}

	if storedCredentialID != credentialID {
		return fmt.Errorf("WebAuthn credential ID mismatch")
	}

	return nil
}
