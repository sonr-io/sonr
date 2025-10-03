package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// DynamicClientRegistrationRequest represents a client registration request per RFC 7591
type DynamicClientRegistrationRequest struct {
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	Scope                   string   `json:"scope,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	ApplicationType         string   `json:"application_type,omitempty"`
	Contacts                []string `json:"contacts,omitempty"`
	LogoURI                 string   `json:"logo_uri,omitempty"`
	ClientURI               string   `json:"client_uri,omitempty"`
	PolicyURI               string   `json:"policy_uri,omitempty"`
	TosURI                  string   `json:"tos_uri,omitempty"`
	JwksURI                 string   `json:"jwks_uri,omitempty"`
	SoftwareID              string   `json:"software_id,omitempty"`
	SoftwareVersion         string   `json:"software_version,omitempty"`
}

// DynamicClientRegistrationResponse represents the response for client registration
type DynamicClientRegistrationResponse struct {
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret,omitempty"`
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	Scope                   string   `json:"scope"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	ApplicationType         string   `json:"application_type"`
	ClientIDIssuedAt        int64    `json:"client_id_issued_at"`
	ClientSecretExpiresAt   int64    `json:"client_secret_expires_at,omitempty"`
	LogoURI                 string   `json:"logo_uri,omitempty"`
	ClientURI               string   `json:"client_uri,omitempty"`
	PolicyURI               string   `json:"policy_uri,omitempty"`
	TosURI                  string   `json:"tos_uri,omitempty"`
	JwksURI                 string   `json:"jwks_uri,omitempty"`
}

// HandleDynamicClientRegistration handles dynamic client registration per RFC 7591
func (s *OAuth2Provider) HandleDynamicClientRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse registration request
	var req DynamicClientRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid registration request", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ClientName == "" {
		http.Error(w, "client_name is required", http.StatusBadRequest)
		return
	}

	if len(req.RedirectURIs) == 0 {
		http.Error(w, "redirect_uris is required", http.StatusBadRequest)
		return
	}

	// Validate redirect URIs
	for _, uri := range req.RedirectURIs {
		if !isValidRedirectURI(uri) {
			http.Error(w, "Invalid redirect URI: "+uri, http.StatusBadRequest)
			return
		}
	}

	// Set defaults if not provided
	if len(req.GrantTypes) == 0 {
		req.GrantTypes = []string{"authorization_code"}
	}

	if len(req.ResponseTypes) == 0 {
		req.ResponseTypes = []string{"code"}
	}

	if req.TokenEndpointAuthMethod == "" {
		// Default based on application type
		if req.ApplicationType == "native" || req.ApplicationType == "browser" {
			req.TokenEndpointAuthMethod = "none" // Public client
		} else {
			req.TokenEndpointAuthMethod = "client_secret_basic"
		}
	}

	if req.ApplicationType == "" {
		req.ApplicationType = "web"
	}

	// Validate grant types and response types
	if !validateGrantTypes(req.GrantTypes) {
		http.Error(w, "Invalid grant types", http.StatusBadRequest)
		return
	}

	if !validateResponseTypes(req.ResponseTypes) {
		http.Error(w, "Invalid response types", http.StatusBadRequest)
		return
	}

	// Validate scopes if scope mapper is available
	if req.Scope != "" && s.scopeMapper != nil {
		scopes := strings.Split(req.Scope, " ")
		for _, scope := range scopes {
			// Check if scope is valid using the scope mapper
			if _, exists := s.scopeMapper.GetScope(scope); !exists {
				http.Error(w, "Invalid scope: "+scope, http.StatusBadRequest)
				return
			}
		}
	}

	// Generate client credentials
	clientID := generateDynamicClientID()
	var clientSecret string
	var clientSecretExpiresAt int64

	// Only generate secret for confidential clients
	if req.TokenEndpointAuthMethod != "none" {
		clientSecret = generateDynamicClientSecret()
		// Client secrets expire in 1 year by default
		clientSecretExpiresAt = time.Now().Add(365 * 24 * time.Hour).Unix()
	}

	// Create OAuth2 client
	client := &OAuth2Client{
		ClientID:        clientID,
		ClientSecret:    clientSecret,
		RedirectURIs:    req.RedirectURIs,
		AllowedScopes:   strings.Split(req.Scope, " "),
		AllowedGrants:   req.GrantTypes,
		TokenLifetime:   time.Hour, // Default 1 hour
		RequirePKCE:     false,
		TrustedClient:   false,
		RequiresConsent: true,
		Metadata: map[string]string{
			"client_name":      req.ClientName,
			"application_type": req.ApplicationType,
			"logo_uri":         req.LogoURI,
			"client_uri":       req.ClientURI,
			"policy_uri":       req.PolicyURI,
			"tos_uri":          req.TosURI,
			"jwks_uri":         req.JwksURI,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set client type based on auth method
	if req.TokenEndpointAuthMethod == "none" {
		client.ClientType = "public"
	} else {
		client.ClientType = "confidential"
	}

	// Determine if client requires PKCE
	if req.ApplicationType == "native" || req.ApplicationType == "browser" {
		client.RequirePKCE = true
	}

	// Store client in the registry
	if s.clientRegistry != nil {
		if err := s.clientRegistry.RegisterClient(client); err != nil {
			http.Error(w, "Failed to register client", http.StatusInternalServerError)
			return
		}
	}

	// Build response
	resp := DynamicClientRegistrationResponse{
		ClientID:                clientID,
		ClientSecret:            clientSecret,
		ClientName:              req.ClientName,
		RedirectURIs:            req.RedirectURIs,
		GrantTypes:              req.GrantTypes,
		ResponseTypes:           req.ResponseTypes,
		Scope:                   req.Scope,
		TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
		ApplicationType:         req.ApplicationType,
		ClientIDIssuedAt:        time.Now().Unix(),
		ClientSecretExpiresAt:   clientSecretExpiresAt,
		LogoURI:                 req.LogoURI,
		ClientURI:               req.ClientURI,
		PolicyURI:               req.PolicyURI,
		TosURI:                  req.TosURI,
		JwksURI:                 req.JwksURI,
	}

	// Return registration response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// HandleClientConfiguration handles client configuration retrieval
func (s *OAuth2Provider) HandleClientConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract client ID from path or query
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		// Try to extract from path (e.g., /register/{client_id})
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 2 {
			clientID = parts[len(parts)-1]
		}
	}

	if clientID == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}

	// Validate access token for client management
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.Header().Set("WWW-Authenticate", `Bearer realm="client_configuration"`)
		http.Error(w, "Access token required", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate token in the access token store
	if s.accessTokenStore == nil {
		http.Error(w, "Token store not available", http.StatusInternalServerError)
		return
	}

	// Check token validity (simplified for now)
	// In production, this should validate the token properly
	if token == "" {
		http.Error(w, "Invalid or expired access token", http.StatusUnauthorized)
		return
	}

	// Get client from registry
	if s.clientRegistry == nil {
		http.Error(w, "Client registry not available", http.StatusInternalServerError)
		return
	}

	client, err := s.clientRegistry.GetClient(clientID)
	if err != nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	// Build response
	resp := DynamicClientRegistrationResponse{
		ClientID:                client.ClientID,
		ClientName:              client.Metadata["client_name"],
		RedirectURIs:            client.RedirectURIs,
		GrantTypes:              client.AllowedGrants,
		ResponseTypes:           []string{"code", "token"}, // Default response types
		Scope:                   strings.Join(client.AllowedScopes, " "),
		TokenEndpointAuthMethod: getTokenEndpointAuthMethod(client),
		ApplicationType:         client.Metadata["application_type"],
		ClientIDIssuedAt:        client.CreatedAt.Unix(),
		LogoURI:                 client.Metadata["logo_uri"],
		ClientURI:               client.Metadata["client_uri"],
		PolicyURI:               client.Metadata["policy_uri"],
		TosURI:                  client.Metadata["tos_uri"],
		JwksURI:                 client.Metadata["jwks_uri"],
	}

	// Return client configuration
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(resp)
}

// getTokenEndpointAuthMethod determines the auth method from client type
func getTokenEndpointAuthMethod(client *OAuth2Client) string {
	if client.ClientType == "public" {
		return "none"
	}
	return "client_secret_basic"
}

// Helper functions for dynamic registration

func generateDynamicClientID() string {
	// Generate a random client ID for dynamic registration
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "dyn_client_" + base64.RawURLEncoding.EncodeToString(b)
}

func generateDynamicClientSecret() string {
	// Generate a secure random secret for dynamic registration
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func isValidRedirectURI(uri string) bool {
	// Basic validation - in production, this should be more comprehensive
	if uri == "" {
		return false
	}

	// Allow localhost for development
	if strings.HasPrefix(uri, "http://localhost") || strings.HasPrefix(uri, "http://127.0.0.1") {
		return true
	}

	// Require HTTPS for production URIs
	if !strings.HasPrefix(uri, "https://") {
		// Allow custom schemes for native apps
		if strings.Contains(uri, "://") {
			return true
		}
		return false
	}

	return true
}

func validateGrantTypes(grantTypes []string) bool {
	validGrants := map[string]bool{
		"authorization_code": true,
		"implicit":           true,
		"refresh_token":      true,
		"client_credentials": true,
		"password":           true,
	}

	for _, grant := range grantTypes {
		if !validGrants[grant] {
			return false
		}
	}
	return true
}

func validateResponseTypes(responseTypes []string) bool {
	validTypes := map[string]bool{
		"code":     true,
		"token":    true,
		"id_token": true,
	}

	for _, respType := range responseTypes {
		// Handle composite types like "code id_token"
		parts := strings.Split(respType, " ")
		for _, part := range parts {
			if !validTypes[part] {
				return false
			}
		}
	}
	return true
}

func hasScope(scopes []string, requiredScope string) bool {
	for _, scope := range scopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}
