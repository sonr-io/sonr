package handlers

import (
	"fmt"
	"strings"

	"github.com/sonr-io/sonr/crypto/ucan"
)

// ScopeMapper manages OAuth scope to UCAN capability mapping
type ScopeMapper struct {
	scopeDefinitions map[string]*OAuth2ScopeDefinition
	ucanTemplates    map[string]*ucan.Attenuation
}

// NewScopeMapper creates a new scope mapper with standard scopes
func NewScopeMapper() *ScopeMapper {
	mapper := &ScopeMapper{
		scopeDefinitions: make(map[string]*OAuth2ScopeDefinition),
		ucanTemplates:    make(map[string]*ucan.Attenuation),
	}

	mapper.initializeStandardScopes()
	return mapper
}

// initializeStandardScopes defines the standard OAuth scopes and their UCAN mappings
func (m *ScopeMapper) initializeStandardScopes() {
	// OpenID Connect scopes
	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "openid",
		Description:  "OpenID Connect authentication",
		UCANActions:  []string{"authenticate"},
		ResourceType: "identity",
		RequiresAuth: true,
		Sensitive:    false,
	})

	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "profile",
		Description:  "Access to user profile information",
		UCANActions:  []string{"read"},
		ResourceType: "did",
		RequiresAuth: true,
		Sensitive:    false,
	})

	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "email",
		Description:  "Access to user email",
		UCANActions:  []string{"read"},
		ResourceType: "contact",
		RequiresAuth: true,
		Sensitive:    true,
	})

	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "offline_access",
		Description:  "Maintain access when user is not present",
		UCANActions:  []string{"refresh"},
		ResourceType: "session",
		RequiresAuth: true,
		Sensitive:    true,
	})

	// Vault scopes
	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "vault:read",
		Description:  "Read access to vault data",
		UCANActions:  []string{"read"},
		ResourceType: "vault",
		RequiresAuth: true,
		Sensitive:    false,
		ParentScope:  "",
	})

	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "vault:write",
		Description:  "Write access to vault data",
		UCANActions:  []string{"read", "write"},
		ResourceType: "vault",
		RequiresAuth: true,
		Sensitive:    true,
		ParentScope:  "vault:read",
	})

	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "vault:sign",
		Description:  "Signing operations with vault keys",
		UCANActions:  []string{"read", "sign"},
		ResourceType: "vault",
		RequiresAuth: true,
		Sensitive:    true,
		ParentScope:  "vault:read",
	})

	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "vault:admin",
		Description:  "Full administrative access to vault",
		UCANActions:  []string{"read", "write", "sign", "export", "import", "delete", "admin"},
		ResourceType: "vault",
		RequiresAuth: true,
		Sensitive:    true,
		ParentScope:  "vault:write",
		ChildScopes:  []string{"vault:read", "vault:write", "vault:sign"},
	})

	// Service scopes
	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "service:read",
		Description:  "Read service information",
		UCANActions:  []string{"read"},
		ResourceType: "service",
		RequiresAuth: true,
		Sensitive:    false,
	})

	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "service:write",
		Description:  "Create and update services",
		UCANActions:  []string{"read", "write"},
		ResourceType: "service",
		RequiresAuth: true,
		Sensitive:    true,
		ParentScope:  "service:read",
	})

	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "service:manage",
		Description:  "Full service management capabilities",
		UCANActions:  []string{"read", "write", "delete", "admin"},
		ResourceType: "service",
		RequiresAuth: true,
		Sensitive:    true,
		ParentScope:  "service:write",
		ChildScopes:  []string{"service:read", "service:write"},
	})

	// DID scopes
	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "did:read",
		Description:  "Read DID documents",
		UCANActions:  []string{"read"},
		ResourceType: "did",
		RequiresAuth: false,
		Sensitive:    false,
	})

	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "did:write",
		Description:  "Update DID documents",
		UCANActions:  []string{"read", "write"},
		ResourceType: "did",
		RequiresAuth: true,
		Sensitive:    true,
		ParentScope:  "did:read",
	})

	// DWN (Decentralized Web Node) scopes
	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "dwn:read",
		Description:  "Read data from DWN",
		UCANActions:  []string{"read"},
		ResourceType: "dwn",
		RequiresAuth: true,
		Sensitive:    false,
	})

	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "dwn:write",
		Description:  "Write data to DWN",
		UCANActions:  []string{"read", "write"},
		ResourceType: "dwn",
		RequiresAuth: true,
		Sensitive:    true,
		ParentScope:  "dwn:read",
	})

	_ = m.RegisterScope(&OAuth2ScopeDefinition{
		Name:         "dwn:admin",
		Description:  "Full administrative access to DWN",
		UCANActions:  []string{"read", "write", "admin", "protocols-configure"},
		ResourceType: "dwn",
		RequiresAuth: true,
		Sensitive:    true,
		ParentScope:  "dwn:write",
		ChildScopes:  []string{"dwn:read", "dwn:write"},
	})
}

// RegisterScope registers a new scope definition
func (m *ScopeMapper) RegisterScope(scope *OAuth2ScopeDefinition) error {
	if scope.Name == "" {
		return fmt.Errorf("scope name is required")
	}

	m.scopeDefinitions[scope.Name] = scope

	// Create UCAN template for this scope
	m.createUCANTemplate(scope)

	return nil
}

// GetScope retrieves a scope definition
func (m *ScopeMapper) GetScope(name string) (*OAuth2ScopeDefinition, bool) {
	scope, exists := m.scopeDefinitions[name]
	return scope, exists
}

// MapToUCAN maps OAuth scopes to UCAN attenuations
func (m *ScopeMapper) MapToUCAN(
	scopes []string,
	userDID string,
	clientID string,
	resourceContext map[string]string,
) []ucan.Attenuation {
	attenuations := []ucan.Attenuation{}

	for _, scopeName := range scopes {
		scope, exists := m.scopeDefinitions[scopeName]
		if !exists {
			continue
		}

		// Create attenuation for this scope
		attenuation := m.createAttenuation(scope, userDID, clientID, resourceContext)
		attenuations = append(attenuations, attenuation)

		// Add child scope attenuations if this is a parent scope
		for _, childScope := range scope.ChildScopes {
			if childDef, exists := m.scopeDefinitions[childScope]; exists {
				childAttenuation := m.createAttenuation(
					childDef,
					userDID,
					clientID,
					resourceContext,
				)
				attenuations = append(attenuations, childAttenuation)
			}
		}
	}

	return attenuations
}

// ValidateScopes validates that the requested scopes are valid
func (m *ScopeMapper) ValidateScopes(scopes []string) error {
	for _, scope := range scopes {
		if _, exists := m.scopeDefinitions[scope]; !exists {
			return fmt.Errorf("invalid scope: %s", scope)
		}
	}
	return nil
}

// GetScopeDescriptions returns human-readable descriptions for scopes
func (m *ScopeMapper) GetScopeDescriptions(scopes []string) map[string]string {
	descriptions := make(map[string]string)

	for _, scope := range scopes {
		if def, exists := m.scopeDefinitions[scope]; exists {
			descriptions[scope] = def.Description
		}
	}

	return descriptions
}

// GetSensitiveScopes returns only the sensitive scopes from a list
func (m *ScopeMapper) GetSensitiveScopes(scopes []string) []string {
	sensitive := []string{}

	for _, scope := range scopes {
		if def, exists := m.scopeDefinitions[scope]; exists && def.Sensitive {
			sensitive = append(sensitive, scope)
		}
	}

	return sensitive
}

// IsHierarchicalScope checks if one scope includes another
func (m *ScopeMapper) IsHierarchicalScope(parentScope, childScope string) bool {
	parent, exists := m.scopeDefinitions[parentScope]
	if !exists {
		return false
	}

	// Check direct children
	for _, child := range parent.ChildScopes {
		if child == childScope {
			return true
		}
		// Recursive check
		if m.IsHierarchicalScope(child, childScope) {
			return true
		}
	}

	return false
}

// Private helper methods

func (m *ScopeMapper) createUCANTemplate(scope *OAuth2ScopeDefinition) {
	// Create a template attenuation for this scope
	capability := &ucan.SimpleCapability{
		Action: strings.Join(scope.UCANActions, ","),
	}

	resource := &SimpleResource{
		Scheme: scope.ResourceType,
		Value:  "{resource_id}",
	}

	m.ucanTemplates[scope.Name] = &ucan.Attenuation{
		Capability: capability,
		Resource:   resource,
	}
}

func (m *ScopeMapper) createAttenuation(
	scope *OAuth2ScopeDefinition,
	userDID string,
	clientID string,
	resourceContext map[string]string,
) ucan.Attenuation {
	// Create capability based on scope actions
	var capability ucan.Capability
	if len(scope.UCANActions) == 1 {
		capability = &ucan.SimpleCapability{
			Action: scope.UCANActions[0],
		}
	} else {
		capability = &ucan.MultiCapability{
			Actions: scope.UCANActions,
		}
	}

	// Create resource based on scope type and context
	resourceValue := m.resolveResourceValue(scope, userDID, resourceContext)
	resource := &SimpleResource{
		Scheme: scope.ResourceType,
		Value:  resourceValue,
	}

	return ucan.Attenuation{
		Capability: capability,
		Resource:   resource,
	}
}

func (m *ScopeMapper) resolveResourceValue(
	scope *OAuth2ScopeDefinition,
	userDID string,
	context map[string]string,
) string {
	switch scope.ResourceType {
	case "did":
		return fmt.Sprintf("did:sonr:%s", userDID)
	case "vault":
		if vaultAddr, exists := context["vault_address"]; exists {
			return vaultAddr
		}
		return fmt.Sprintf("vault:%s", userDID)
	case "service":
		if serviceID, exists := context["service_id"]; exists {
			return serviceID
		}
		return "service:*"
	case "dwn":
		if dwnID, exists := context["dwn_id"]; exists {
			return dwnID
		}
		return fmt.Sprintf("dwn:%s", userDID)
	default:
		return "*"
	}
}

// SimpleResource implements the Resource interface for OAuth scopes
type SimpleResource struct {
	Scheme string
	Value  string
}

func (r *SimpleResource) GetScheme() string {
	return r.Scheme
}

func (r *SimpleResource) GetValue() string {
	return r.Value
}

func (r *SimpleResource) GetURI() string {
	return fmt.Sprintf("%s://%s", r.Scheme, r.Value)
}

func (r *SimpleResource) Matches(other ucan.Resource) bool {
	if r.Scheme != other.GetScheme() {
		return false
	}

	// Wildcard matching
	if r.Value == "*" || other.GetValue() == "*" {
		return true
	}

	// Exact match
	return r.Value == other.GetValue()
}
