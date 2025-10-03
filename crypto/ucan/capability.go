// Package ucan provides User-Controlled Authorization Networks (UCAN) implementation
// for decentralized authorization and capability delegation in the Sonr network.
// This package handles JWT-based tokens, cryptographic verification, and resource capabilities.
package ucan

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Token represents a UCAN JWT token with parsed claims
type Token struct {
	Raw          string        `json:"raw"`
	Issuer       string        `json:"iss"`
	Audience     string        `json:"aud"`
	ExpiresAt    int64         `json:"exp,omitempty"`
	NotBefore    int64         `json:"nbf,omitempty"`
	Attenuations []Attenuation `json:"att"`
	Proofs       []Proof       `json:"prf,omitempty"`
	Facts        []Fact        `json:"fct,omitempty"`
}

// Attenuation represents a UCAN capability attenuation
type Attenuation struct {
	Capability Capability `json:"can"`
	Resource   Resource   `json:"with"`
}

// Proof represents a UCAN delegation proof (either JWT or CID)
type Proof string

// Fact represents arbitrary facts in UCAN tokens
type Fact struct {
	Data json.RawMessage `json:"data"`
}

// Capability defines what actions can be performed
type Capability interface {
	// GetActions returns the list of actions this capability grants
	GetActions() []string
	// Grants checks if this capability grants the required abilities
	Grants(abilities []string) bool
	// Contains checks if this capability contains another capability
	Contains(other Capability) bool
	// String returns a string representation
	String() string
}

// Resource defines what resource the capability applies to
type Resource interface {
	// GetScheme returns the resource scheme (e.g., "https", "ipfs")
	GetScheme() string
	// GetValue returns the resource value/path
	GetValue() string
	// GetURI returns the full URI string
	GetURI() string
	// Matches checks if this resource matches another resource
	Matches(other Resource) bool
}

// SimpleCapability implements Capability for single actions
type SimpleCapability struct {
	Action string `json:"action"`
}

// GetActions returns the single action
func (c *SimpleCapability) GetActions() []string {
	return []string{c.Action}
}

// Grants checks if the capability grants all required abilities
func (c *SimpleCapability) Grants(abilities []string) bool {
	if len(abilities) != 1 {
		return false
	}
	return c.Action == abilities[0] || c.Action == "*"
}

// Contains checks if this capability contains another capability
func (c *SimpleCapability) Contains(other Capability) bool {
	if c.Action == "*" {
		return true
	}

	otherActions := other.GetActions()
	if len(otherActions) != 1 {
		return false
	}

	return c.Action == otherActions[0]
}

// String returns string representation
func (c *SimpleCapability) String() string {
	return c.Action
}

// MultiCapability implements Capability for multiple actions
type MultiCapability struct {
	Actions []string `json:"actions"`
}

// GetActions returns all actions
func (c *MultiCapability) GetActions() []string {
	return c.Actions
}

// Grants checks if the capability grants all required abilities
func (c *MultiCapability) Grants(abilities []string) bool {
	actionSet := make(map[string]bool)
	for _, action := range c.Actions {
		actionSet[action] = true
	}

	// Check if we have wildcard permission
	if actionSet["*"] {
		return true
	}

	// Check each required ability
	for _, ability := range abilities {
		if !actionSet[ability] {
			return false
		}
	}

	return true
}

// Contains checks if this capability contains another capability
func (c *MultiCapability) Contains(other Capability) bool {
	actionSet := make(map[string]bool)
	for _, action := range c.Actions {
		actionSet[action] = true
	}

	// Wildcard contains everything
	if actionSet["*"] {
		return true
	}

	// Check if all other actions are contained
	for _, otherAction := range other.GetActions() {
		if !actionSet[otherAction] {
			return false
		}
	}

	return true
}

// String returns string representation
func (c *MultiCapability) String() string {
	return strings.Join(c.Actions, ",")
}

// SimpleResource implements Resource for basic URI resources
type SimpleResource struct {
	Scheme string `json:"scheme"`
	Value  string `json:"value"`
	URI    string `json:"uri"`
}

// GetScheme returns the resource scheme
func (r *SimpleResource) GetScheme() string {
	return r.Scheme
}

// GetValue returns the resource value
func (r *SimpleResource) GetValue() string {
	return r.Value
}

// GetURI returns the full URI
func (r *SimpleResource) GetURI() string {
	return r.URI
}

// Matches checks if resources are equivalent
func (r *SimpleResource) Matches(other Resource) bool {
	return r.URI == other.GetURI()
}

// VaultResource represents vault-specific resources with metadata
type VaultResource struct {
	SimpleResource
	VaultAddress   string            `json:"vault_address,omitempty"`
	EnclaveDataCID string            `json:"enclave_data_cid,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// ServiceResource represents service-specific resources
type ServiceResource struct {
	SimpleResource
	ServiceID string            `json:"service_id"`
	Domain    string            `json:"domain"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// CreateSimpleAttenuation creates a basic attenuation
func CreateSimpleAttenuation(action, resourceURI string) Attenuation {
	return Attenuation{
		Capability: &SimpleCapability{Action: action},
		Resource:   parseResourceURI(resourceURI),
	}
}

// CreateMultiAttenuation creates an attenuation with multiple actions
func CreateMultiAttenuation(actions []string, resourceURI string) Attenuation {
	return Attenuation{
		Capability: &MultiCapability{Actions: actions},
		Resource:   parseResourceURI(resourceURI),
	}
}

// CreateVaultAttenuation creates a vault-specific attenuation
func CreateVaultAttenuation(actions []string, enclaveDataCID, vaultAddress string) Attenuation {
	resource := &VaultResource{
		SimpleResource: SimpleResource{
			Scheme: "ipfs",
			Value:  enclaveDataCID,
			URI:    fmt.Sprintf("ipfs://%s", enclaveDataCID),
		},
		VaultAddress:   vaultAddress,
		EnclaveDataCID: enclaveDataCID,
	}

	return Attenuation{
		Capability: &MultiCapability{Actions: actions},
		Resource:   resource,
	}
}

// CreateServiceAttenuation creates a service-specific attenuation
func CreateServiceAttenuation(actions []string, serviceID, domain string) Attenuation {
	resourceURI := fmt.Sprintf("service://%s", serviceID)
	resource := &ServiceResource{
		SimpleResource: SimpleResource{
			Scheme: "service",
			Value:  serviceID,
			URI:    resourceURI,
		},
		ServiceID: serviceID,
		Domain:    domain,
	}

	return Attenuation{
		Capability: &MultiCapability{Actions: actions},
		Resource:   resource,
	}
}

// parseResourceURI creates a Resource from URI string
func parseResourceURI(uri string) Resource {
	parts := strings.SplitN(uri, "://", 2)
	if len(parts) != 2 {
		return &SimpleResource{
			Scheme: "unknown",
			Value:  uri,
			URI:    uri,
		}
	}

	return &SimpleResource{
		Scheme: parts[0],
		Value:  parts[1],
		URI:    uri,
	}
}

// CapabilityTemplate provides validation and construction utilities
type CapabilityTemplate struct {
	AllowedActions    map[string][]string `json:"allowed_actions"`    // resource_type -> []actions
	DefaultExpiration time.Duration       `json:"default_expiration"` // default token lifetime
	MaxExpiration     time.Duration       `json:"max_expiration"`     // maximum allowed lifetime
}

// NewCapabilityTemplate creates a new capability template
func NewCapabilityTemplate() *CapabilityTemplate {
	return &CapabilityTemplate{
		AllowedActions:    make(map[string][]string),
		DefaultExpiration: 24 * time.Hour,
		MaxExpiration:     30 * 24 * time.Hour, // 30 days
	}
}

// AddAllowedActions adds allowed actions for a resource type
func (ct *CapabilityTemplate) AddAllowedActions(resourceType string, actions []string) {
	ct.AllowedActions[resourceType] = actions
}

// ValidateAttenuation validates an attenuation against the template
func (ct *CapabilityTemplate) ValidateAttenuation(att Attenuation) error {
	resourceType := att.Resource.GetScheme()
	allowedActions, exists := ct.AllowedActions[resourceType]

	if !exists {
		// Allow unknown resource types for backward compatibility
		return nil
	}

	// Create action set for efficient lookup
	actionSet := make(map[string]bool)
	for _, action := range allowedActions {
		actionSet[action] = true
	}

	// Check if all capability actions are allowed
	for _, action := range att.Capability.GetActions() {
		if action == "*" {
			// Wildcard requires explicit permission
			if !actionSet["*"] {
				return fmt.Errorf("wildcard action not allowed for resource type %s", resourceType)
			}
			continue
		}

		if !actionSet[action] {
			return fmt.Errorf("action %s not allowed for resource type %s", action, resourceType)
		}
	}

	return nil
}

// ValidateExpiration validates token expiration time
func (ct *CapabilityTemplate) ValidateExpiration(expiresAt int64) error {
	if expiresAt == 0 {
		return nil // No expiration is allowed
	}

	now := time.Now()
	expiry := time.Unix(expiresAt, 0)

	if expiry.Before(now) {
		return fmt.Errorf("token expiration is in the past")
	}

	if expiry.Sub(now) > ct.MaxExpiration {
		return fmt.Errorf("token expiration exceeds maximum allowed duration")
	}

	return nil
}

// GetDefaultExpirationTime returns the default expiration timestamp
func (ct *CapabilityTemplate) GetDefaultExpirationTime() int64 {
	return time.Now().Add(ct.DefaultExpiration).Unix()
}

// StandardVaultTemplate returns a standard template for vault operations
func StandardVaultTemplate() *CapabilityTemplate {
	template := NewCapabilityTemplate()
	template.AddAllowedActions(
		"ipfs",
		[]string{"read", "write", "sign", "export", "import", "delete", VaultAdminAction},
	)
	template.AddAllowedActions(
		"vault",
		[]string{"read", "write", "sign", "export", "import", "delete", "admin", "*"},
	)
	return template
}

// StandardServiceTemplate returns a standard template for service operations
func StandardServiceTemplate() *CapabilityTemplate {
	template := NewCapabilityTemplate()
	template.AddAllowedActions(
		"service",
		[]string{"read", "write", "admin", "register", "update", "delete"},
	)
	template.AddAllowedActions("https", []string{"read", "write"})
	template.AddAllowedActions("http", []string{"read", "write"})
	return template
}

// AttenuationList provides utilities for working with multiple attenuations
type AttenuationList []Attenuation

// Contains checks if the list contains attenuations for a specific resource
func (al AttenuationList) Contains(resourceURI string) bool {
	for _, att := range al {
		if att.Resource.GetURI() == resourceURI {
			return true
		}
	}
	return false
}

// GetCapabilitiesForResource returns all capabilities for a specific resource
func (al AttenuationList) GetCapabilitiesForResource(resourceURI string) []Capability {
	var capabilities []Capability
	for _, att := range al {
		if att.Resource.GetURI() == resourceURI {
			capabilities = append(capabilities, att.Capability)
		}
	}
	return capabilities
}

// CanPerform checks if the attenuations allow specific actions on a resource
func (al AttenuationList) CanPerform(resourceURI string, actions []string) bool {
	capabilities := al.GetCapabilitiesForResource(resourceURI)
	for _, cap := range capabilities {
		if cap.Grants(actions) {
			return true
		}
	}
	return false
}

// IsSubsetOf checks if this list is a subset of another list
func (al AttenuationList) IsSubsetOf(parent AttenuationList) bool {
	for _, childAtt := range al {
		if !parent.containsAttenuation(childAtt) {
			return false
		}
	}
	return true
}

// containsAttenuation checks if the list contains an equivalent attenuation
func (al AttenuationList) containsAttenuation(att Attenuation) bool {
	for _, parentAtt := range al {
		if parentAtt.Resource.Matches(att.Resource) {
			if parentAtt.Capability.Contains(att.Capability) {
				return true
			}
		}
	}
	return false
}

// Module-Specific Capability Types

// DIDCapability implements Capability for DID module operations
type DIDCapability struct {
	Action   string            `json:"action"`
	Actions  []string          `json:"actions,omitempty"`
	Caveats  []string          `json:"caveats,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// GetActions returns the actions this DID capability grants
func (c *DIDCapability) GetActions() []string {
	if len(c.Actions) > 0 {
		return c.Actions
	}
	return []string{c.Action}
}

// Grants checks if this capability grants the required abilities
func (c *DIDCapability) Grants(abilities []string) bool {
	if c.Action == "*" {
		return true
	}

	grantedActions := make(map[string]bool)
	for _, action := range c.GetActions() {
		grantedActions[action] = true
	}

	for _, ability := range abilities {
		if !grantedActions[ability] {
			return false
		}
	}
	return true
}

// Contains checks if this capability contains another capability
func (c *DIDCapability) Contains(other Capability) bool {
	if c.Action == "*" {
		return true
	}

	ourActions := make(map[string]bool)
	for _, action := range c.GetActions() {
		ourActions[action] = true
	}

	for _, otherAction := range other.GetActions() {
		if !ourActions[otherAction] {
			return false
		}
	}
	return true
}

// String returns string representation
func (c *DIDCapability) String() string {
	if len(c.Actions) > 1 {
		return strings.Join(c.Actions, ",")
	}
	return c.Action
}

// DWNCapability implements Capability for DWN module operations
type DWNCapability struct {
	Action   string            `json:"action"`
	Actions  []string          `json:"actions,omitempty"`
	Caveats  []string          `json:"caveats,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// GetActions returns the actions this DWN capability grants
func (c *DWNCapability) GetActions() []string {
	if len(c.Actions) > 0 {
		return c.Actions
	}
	return []string{c.Action}
}

// Grants checks if this capability grants the required abilities
func (c *DWNCapability) Grants(abilities []string) bool {
	if c.Action == "*" {
		return true
	}

	grantedActions := make(map[string]bool)
	for _, action := range c.GetActions() {
		grantedActions[action] = true
	}

	for _, ability := range abilities {
		if !grantedActions[ability] {
			return false
		}
	}
	return true
}

// Contains checks if this capability contains another capability
func (c *DWNCapability) Contains(other Capability) bool {
	if c.Action == "*" {
		return true
	}

	ourActions := make(map[string]bool)
	for _, action := range c.GetActions() {
		ourActions[action] = true
	}

	for _, otherAction := range other.GetActions() {
		if !ourActions[otherAction] {
			return false
		}
	}
	return true
}

// String returns string representation
func (c *DWNCapability) String() string {
	if len(c.Actions) > 1 {
		return strings.Join(c.Actions, ",")
	}
	return c.Action
}

// DEXCapability implements Capability for DEX module operations
type DEXCapability struct {
	Action    string            `json:"action"`
	Actions   []string          `json:"actions,omitempty"`
	Caveats   []string          `json:"caveats,omitempty"`
	MaxAmount string            `json:"max_amount,omitempty"` // For swap limits
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// GetActions returns the actions this DEX capability grants
func (c *DEXCapability) GetActions() []string {
	if len(c.Actions) > 0 {
		return c.Actions
	}
	return []string{c.Action}
}

// Grants checks if this capability grants the required abilities
func (c *DEXCapability) Grants(abilities []string) bool {
	if c.Action == "*" {
		return true
	}

	grantedActions := make(map[string]bool)
	for _, action := range c.GetActions() {
		grantedActions[action] = true
	}

	for _, ability := range abilities {
		if !grantedActions[ability] {
			return false
		}
	}
	return true
}

// Contains checks if this capability contains another capability
func (c *DEXCapability) Contains(other Capability) bool {
	if c.Action == "*" {
		return true
	}

	ourActions := make(map[string]bool)
	for _, action := range c.GetActions() {
		ourActions[action] = true
	}

	for _, otherAction := range other.GetActions() {
		if !ourActions[otherAction] {
			return false
		}
	}
	return true
}

// String returns string representation
func (c *DEXCapability) String() string {
	if len(c.Actions) > 1 {
		return strings.Join(c.Actions, ",")
	}
	return c.Action
}

// Module-Specific Resource Types

// DIDResource represents DID-specific resources
type DIDResource struct {
	SimpleResource
	DIDMethod  string            `json:"did_method,omitempty"`
	DIDSubject string            `json:"did_subject,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// DWNResource represents DWN-specific resources
type DWNResource struct {
	SimpleResource
	RecordType string            `json:"record_type,omitempty"`
	Protocol   string            `json:"protocol,omitempty"`
	Owner      string            `json:"owner,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// DEXResource represents DEX-specific resources
type DEXResource struct {
	SimpleResource
	PoolID    string            `json:"pool_id,omitempty"`
	AssetPair string            `json:"asset_pair,omitempty"`
	OrderID   string            `json:"order_id,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Enhanced ServiceResource adds delegation capabilities
func (r *ServiceResource) SupportsDelegate() bool {
	return r.Metadata != nil && r.Metadata["supports_delegation"] == "true"
}

// Module-Specific Capability Templates

// StandardDIDTemplate returns a standard template for DID operations
func StandardDIDTemplate() *CapabilityTemplate {
	template := NewCapabilityTemplate()
	template.AddAllowedActions("did", []string{
		"create", "register", "update", "deactivate", "revoke",
		"add-verification-method", "remove-verification-method",
		"add-service", "remove-service", "issue-credential",
		"revoke-credential", "link-wallet", "register-webauthn", "*",
	})
	return template
}

// StandardDWNTemplate returns a standard template for DWN operations
func StandardDWNTemplate() *CapabilityTemplate {
	template := NewCapabilityTemplate()
	template.AddAllowedActions("dwn", []string{
		"records-write", "records-delete", "protocols-configure",
		"permissions-grant", "permissions-revoke", "create", "read",
		"update", "delete", "*",
	})
	return template
}

// EnhancedServiceTemplate returns enhanced service template with delegation support
func EnhancedServiceTemplate() *CapabilityTemplate {
	template := NewCapabilityTemplate()
	template.AddAllowedActions("service", []string{
		"register", "update", "delete", "verify-domain",
		"initiate-domain-verification", "delegate", "*",
	})
	template.AddAllowedActions("svc", []string{
		"register", "verify-domain", "delegate", "*",
	})
	template.AddAllowedActions("https", []string{"read", "write"})
	template.AddAllowedActions("http", []string{"read", "write"})
	return template
}

// StandardDEXTemplate returns a standard template for DEX operations
func StandardDEXTemplate() *CapabilityTemplate {
	template := NewCapabilityTemplate()
	template.AddAllowedActions("dex", []string{
		"register-account", "swap", "provide-liquidity", "remove-liquidity",
		"create-limit-order", "cancel-order", "*",
	})
	template.AddAllowedActions("pool", []string{
		"swap", "provide-liquidity", "remove-liquidity", "*",
	})
	return template
}

// Module-Specific Attenuation Constructors

// CreateDIDAttenuation creates a DID-specific attenuation
func CreateDIDAttenuation(actions []string, didPattern string, caveats []string) Attenuation {
	resourceURI := fmt.Sprintf("did:%s", didPattern)
	resource := &DIDResource{
		SimpleResource: SimpleResource{
			Scheme: "did",
			Value:  didPattern,
			URI:    resourceURI,
		},
	}

	return Attenuation{
		Capability: &DIDCapability{
			Actions: actions,
			Caveats: caveats,
		},
		Resource: resource,
	}
}

// CreateDWNAttenuation creates a DWN-specific attenuation
func CreateDWNAttenuation(actions []string, recordPattern string, caveats []string) Attenuation {
	resourceURI := fmt.Sprintf("dwn:records/%s", recordPattern)
	resource := &DWNResource{
		SimpleResource: SimpleResource{
			Scheme: "dwn",
			Value:  fmt.Sprintf("records/%s", recordPattern),
			URI:    resourceURI,
		},
		RecordType: recordPattern,
	}

	return Attenuation{
		Capability: &DWNCapability{
			Actions: actions,
			Caveats: caveats,
		},
		Resource: resource,
	}
}

// CreateDEXAttenuation creates a DEX-specific attenuation
func CreateDEXAttenuation(actions []string, poolPattern string, caveats []string, maxAmount string) Attenuation {
	resourceURI := fmt.Sprintf("dex:pool/%s", poolPattern)
	resource := &DEXResource{
		SimpleResource: SimpleResource{
			Scheme: "dex",
			Value:  fmt.Sprintf("pool/%s", poolPattern),
			URI:    resourceURI,
		},
		PoolID: poolPattern,
	}

	return Attenuation{
		Capability: &DEXCapability{
			Actions:   actions,
			Caveats:   caveats,
			MaxAmount: maxAmount,
		},
		Resource: resource,
	}
}

// Cross-Module Capability Composition

// CrossModuleCapability allows composing capabilities across modules
type CrossModuleCapability struct {
	Modules map[string]Capability `json:"modules"`
}

// GetActions returns all actions across all modules
func (c *CrossModuleCapability) GetActions() []string {
	var actions []string
	for _, cap := range c.Modules {
		actions = append(actions, cap.GetActions()...)
	}
	return actions
}

// Grants checks if required abilities are granted across modules
func (c *CrossModuleCapability) Grants(abilities []string) bool {
	allActions := make(map[string]bool)
	for _, cap := range c.Modules {
		for _, action := range cap.GetActions() {
			allActions[action] = true
		}
	}

	for _, ability := range abilities {
		if !allActions[ability] {
			return false
		}
	}
	return true
}

// Contains checks if this cross-module capability contains another
func (c *CrossModuleCapability) Contains(other Capability) bool {
	// For cross-module capabilities, check each module
	if otherCross, ok := other.(*CrossModuleCapability); ok {
		for module, otherCap := range otherCross.Modules {
			if ourCap, exists := c.Modules[module]; exists {
				if !ourCap.Contains(otherCap) {
					return false
				}
			} else {
				return false
			}
		}
		return true
	}

	// For single capabilities, check if any module contains it
	for _, cap := range c.Modules {
		if cap.Contains(other) {
			return true
		}
	}
	return false
}

// String returns string representation
func (c *CrossModuleCapability) String() string {
	var moduleStrs []string
	for module, cap := range c.Modules {
		moduleStrs = append(moduleStrs, fmt.Sprintf("%s:%s", module, cap.String()))
	}
	return strings.Join(moduleStrs, ";")
}

// Gasless Transaction Support

// GaslessCapability wraps other capabilities with gasless transaction support
type GaslessCapability struct {
	Capability
	AllowGasless bool   `json:"allow_gasless"`
	GasLimit     uint64 `json:"gas_limit,omitempty"`
}

// SupportsGasless returns whether this capability supports gasless transactions
func (c *GaslessCapability) SupportsGasless() bool {
	return c.AllowGasless
}

// GetGasLimit returns the gas limit for gasless transactions
func (c *GaslessCapability) GetGasLimit() uint64 {
	return c.GasLimit
}
