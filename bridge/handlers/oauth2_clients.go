package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	// ClientTypePublic represents a public OAuth2 client
	ClientTypePublic = "public"
	// ClientTypeConfidential represents a confidential OAuth2 client
	ClientTypeConfidential = "confidential"
)

// ClientRegistry manages OAuth2 client registrations
type ClientRegistry struct {
	mu      sync.RWMutex
	clients map[string]*OAuth2Client
}

// NewClientRegistry creates a new client registry
func NewClientRegistry() *ClientRegistry {
	registry := &ClientRegistry{
		clients: make(map[string]*OAuth2Client),
	}

	// Initialize with default clients for development
	registry.initializeDefaultClients()

	return registry
}

// initializeDefaultClients adds default clients for development/testing
func (r *ClientRegistry) initializeDefaultClients() {
	// Development public client (e.g., SPA)
	_ = r.RegisterClient(&OAuth2Client{
		ClientID:   "sonr-web-app",
		ClientType: ClientTypePublic,
		RedirectURIs: []string{
			"http://localhost:3000/callback",
			"http://localhost:3001/callback",
		},
		AllowedScopes: []string{
			"openid",
			"profile",
			"vault:read",
			"vault:write",
			"service:manage",
		},
		AllowedGrants:   []string{"authorization_code", "refresh_token"},
		TokenLifetime:   time.Hour,
		RequirePKCE:     true,
		TrustedClient:   true,
		RequiresConsent: false,
		Metadata: map[string]string{
			"name":        "Sonr Web Application",
			"description": "Official Sonr web application",
			"logo_uri":    "https://sonr.io/logo.png",
			"client_uri":  "https://app.sonr.io",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	// Development confidential client (e.g., backend service)
	_ = r.RegisterClient(&OAuth2Client{
		ClientID:        "sonr-backend-service",
		ClientSecret:    "development-secret-change-in-production",
		ClientType:      ClientTypeConfidential,
		RedirectURIs:    []string{"http://localhost:8081/callback"},
		AllowedScopes:   []string{"openid", "profile", "vault:admin", "service:manage"},
		AllowedGrants:   []string{"authorization_code", "refresh_token", "client_credentials"},
		TokenLifetime:   time.Hour * 2,
		RequirePKCE:     false,
		TrustedClient:   true,
		RequiresConsent: false,
		Metadata: map[string]string{
			"name":        "Sonr Backend Service",
			"description": "Backend service for Sonr ecosystem",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	// Example third-party client
	_ = r.RegisterClient(&OAuth2Client{
		ClientID:        "example-third-party",
		ClientSecret:    "third-party-secret",
		ClientType:      ClientTypeConfidential,
		RedirectURIs:    []string{"https://example.com/oauth/callback"},
		AllowedScopes:   []string{"openid", "profile", "vault:read"},
		AllowedGrants:   []string{"authorization_code", "refresh_token"},
		TokenLifetime:   time.Hour,
		RequirePKCE:     true,
		TrustedClient:   false,
		RequiresConsent: true,
		Metadata: map[string]string{
			"name":        "Example Third Party App",
			"description": "Example integration partner",
			"logo_uri":    "https://example.com/logo.png",
			"client_uri":  "https://example.com",
			"policy_uri":  "https://example.com/privacy",
			"tos_uri":     "https://example.com/terms",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
}

// RegisterClient registers a new OAuth2 client
func (r *ClientRegistry) RegisterClient(client *OAuth2Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate client
	if err := r.validateClient(client); err != nil {
		return err
	}

	// Generate client ID if not provided
	if client.ClientID == "" {
		client.ClientID = generateOAuth2ClientID()
	}

	// Generate client secret for confidential clients
	if client.ClientType == ClientTypeConfidential && client.ClientSecret == "" {
		client.ClientSecret = generateClientSecret()
	}

	// Set timestamps
	if client.CreatedAt.IsZero() {
		client.CreatedAt = time.Now()
	}
	client.UpdatedAt = time.Now()

	// Store client
	r.clients[client.ClientID] = client

	return nil
}

// GetClient retrieves a client by ID
func (r *ClientRegistry) GetClient(clientID string) (*OAuth2Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, exists := r.clients[clientID]
	if !exists {
		return nil, fmt.Errorf("client not found: %s", clientID)
	}

	return client, nil
}

// UpdateClient updates an existing client
func (r *ClientRegistry) UpdateClient(clientID string, updates *OAuth2Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	client, exists := r.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	// Update allowed fields
	if len(updates.RedirectURIs) > 0 {
		client.RedirectURIs = updates.RedirectURIs
	}
	if len(updates.AllowedScopes) > 0 {
		client.AllowedScopes = updates.AllowedScopes
	}
	if len(updates.AllowedGrants) > 0 {
		client.AllowedGrants = updates.AllowedGrants
	}
	if updates.TokenLifetime > 0 {
		client.TokenLifetime = updates.TokenLifetime
	}
	if updates.Metadata != nil {
		client.Metadata = updates.Metadata
	}

	client.RequirePKCE = updates.RequirePKCE
	client.RequiresConsent = updates.RequiresConsent
	client.UpdatedAt = time.Now()

	return nil
}

// DeleteClient removes a client from the registry
func (r *ClientRegistry) DeleteClient(clientID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clients[clientID]; !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	delete(r.clients, clientID)
	return nil
}

// ListClients returns all registered clients
func (r *ClientRegistry) ListClients() []*OAuth2Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clients := make([]*OAuth2Client, 0, len(r.clients))
	for _, client := range r.clients {
		clients = append(clients, client)
	}

	return clients
}

// ValidateRedirectURI checks if a redirect URI is valid for the client
func (c *OAuth2Client) ValidateRedirectURI(redirectURI string) bool {
	if redirectURI == "" {
		return false
	}

	// Parse the redirect URI
	parsedURI, err := url.Parse(redirectURI)
	if err != nil {
		return false
	}

	// Check against registered redirect URIs
	for _, registeredURI := range c.RedirectURIs {
		registeredParsed, err := url.Parse(registeredURI)
		if err != nil {
			continue
		}

		// For public clients, allow localhost with any port
		if c.ClientType == ClientTypePublic && registeredParsed.Hostname() == "localhost" &&
			parsedURI.Hostname() == "localhost" {
			if registeredParsed.Path == parsedURI.Path {
				return true
			}
		}

		// Exact match for other cases
		if registeredURI == redirectURI {
			return true
		}

		// Allow subdomain matching for trusted clients
		if c.TrustedClient && matchesWithSubdomain(registeredParsed, parsedURI) {
			return true
		}
	}

	return false
}

// ValidateScopes checks if the requested scopes are allowed for the client
func (c *OAuth2Client) ValidateScopes(requestedScopes []string) bool {
	if len(requestedScopes) == 0 {
		return true // No scopes requested is valid
	}

	for _, scope := range requestedScopes {
		if !c.hasScope(scope) {
			return false
		}
	}

	return true
}

// hasScope checks if a client has a specific scope
func (c *OAuth2Client) hasScope(scope string) bool {
	for _, allowedScope := range c.AllowedScopes {
		if allowedScope == scope {
			return true
		}
		// Check for hierarchical scopes (e.g., vault:admin includes vault:read)
		if isHierarchicalScope(allowedScope, scope) {
			return true
		}
	}
	return false
}

// HasGrantType checks if a client supports a specific grant type
func (c *OAuth2Client) HasGrantType(grantType string) bool {
	for _, allowed := range c.AllowedGrants {
		if allowed == grantType {
			return true
		}
	}
	return false
}

// validateClient validates client configuration
func (r *ClientRegistry) validateClient(client *OAuth2Client) error {
	// Validate client type
	if client.ClientType != ClientTypePublic && client.ClientType != ClientTypeConfidential {
		return fmt.Errorf("invalid client type: %s", client.ClientType)
	}

	// Validate redirect URIs
	if len(client.RedirectURIs) == 0 {
		return fmt.Errorf("at least one redirect URI is required")
	}

	for _, uri := range client.RedirectURIs {
		if _, err := url.Parse(uri); err != nil {
			return fmt.Errorf("invalid redirect URI: %s", uri)
		}
	}

	// Validate grant types
	if len(client.AllowedGrants) == 0 {
		client.AllowedGrants = []string{"authorization_code"}
	}

	validGrants := map[string]bool{
		"authorization_code": true,
		"implicit":           true,
		"refresh_token":      true,
		"client_credentials": true,
		"password":           true,
		"urn:ietf:params:oauth:grant-type:device_code": true,
	}

	for _, grant := range client.AllowedGrants {
		if !validGrants[grant] {
			return fmt.Errorf("invalid grant type: %s", grant)
		}
	}

	// Client credentials grant requires confidential client
	if contains(client.AllowedGrants, "client_credentials") &&
		client.ClientType != ClientTypeConfidential {
		return fmt.Errorf("client_credentials grant requires confidential client")
	}

	// Public clients should use PKCE
	if client.ClientType == ClientTypePublic && !client.RequirePKCE {
		// Log warning but don't fail
		fmt.Printf("Warning: Public client %s should use PKCE\n", client.ClientID)
	}

	return nil
}

// Helper functions

func generateOAuth2ClientID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return fmt.Sprintf("client_%s", base64.RawURLEncoding.EncodeToString(bytes))
}

func generateClientSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func matchesWithSubdomain(registered, requested *url.URL) bool {
	if registered.Scheme != requested.Scheme {
		return false
	}

	if registered.Path != requested.Path {
		return false
	}

	// Check if requested hostname is a subdomain of registered
	registeredHost := registered.Hostname()
	requestedHost := requested.Hostname()

	if registeredHost == requestedHost {
		return true
	}

	// Check subdomain match (e.g., *.example.com matches sub.example.com)
	if strings.HasPrefix(registeredHost, "*.") {
		domain := strings.TrimPrefix(registeredHost, "*.")
		return strings.HasSuffix(requestedHost, domain)
	}

	return false
}

func isHierarchicalScope(allowed, requested string) bool {
	// Define scope hierarchy
	hierarchy := map[string][]string{
		"vault:admin":    {"vault:write", "vault:read", "vault:sign"},
		"vault:write":    {"vault:read"},
		"service:manage": {"service:read", "service:write"},
		"did:write":      {"did:read"},
	}

	childScopes, exists := hierarchy[allowed]
	if !exists {
		return false
	}

	for _, child := range childScopes {
		if child == requested {
			return true
		}
		// Recursive check for nested hierarchies
		if isHierarchicalScope(child, requested) {
			return true
		}
	}

	return false
}
