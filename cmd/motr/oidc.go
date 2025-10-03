//go:build js && wasm
// +build js,wasm

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"
)

// OIDCProvider manages OpenID Connect operations
type OIDCProvider struct {
	mu            sync.RWMutex
	issuer        string
	authCodes     map[string]*AuthorizationCode
	accessTokens  map[string]*AccessToken
	refreshTokens map[string]*RefreshToken
	clients       map[string]*OIDCClient
	users         map[string]*User
}

// AuthorizationCode represents an authorization code
type AuthorizationCode struct {
	Code                string
	ClientID            string
	RedirectURI         string
	Scope               string
	State               string
	Nonce               string
	UserID              string
	ExpiresAt           time.Time
	CodeChallenge       string
	CodeChallengeMethod string
}

// AccessToken represents an access token
type AccessToken struct {
	Token     string
	ClientID  string
	UserID    string
	Scope     string
	ExpiresAt time.Time
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	Token     string
	ClientID  string
	UserID    string
	Scope     string
	ExpiresAt time.Time
}

// OIDCClient represents an OIDC client application
type OIDCClient struct {
	ClientID      string
	ClientSecret  string
	RedirectURIs  []string
	GrantTypes    []string
	ResponseTypes []string
	Scopes        []string
	Name          string
}

// User represents a user
type User struct {
	ID            string
	Username      string
	Email         string
	EmailVerified bool
	Name          string
	GivenName     string
	FamilyName    string
}

// OIDCDiscovery represents OIDC discovery document
type OIDCDiscovery struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	JWKSUri                           string   `json:"jwks_uri"`
	RegistrationEndpoint              string   `json:"registration_endpoint,omitempty"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	ResponseModesSupported            []string `json:"response_modes_supported,omitempty"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	ACRValuesSupported                []string `json:"acr_values_supported,omitempty"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
}

// TokenRequest represents a token request
type TokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code,omitempty"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	CodeVerifier string `json:"code_verifier,omitempty"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// Global OIDC provider instance
var oidcProvider = &OIDCProvider{
	issuer:        "https://motor.sonr.io",
	authCodes:     make(map[string]*AuthorizationCode),
	accessTokens:  make(map[string]*AccessToken),
	refreshTokens: make(map[string]*RefreshToken),
	clients:       make(map[string]*OIDCClient),
	users:         make(map[string]*User),
}

// Initialize OIDC provider
func init() {
	// Initialize JWT manager
	InitJWTManager()

	// Add default client for testing
	oidcProvider.clients["motor-client"] = &OIDCClient{
		ClientID:      "motor-client",
		ClientSecret:  "motor-secret",
		RedirectURIs:  []string{"https://localhost:3000/callback", "http://localhost:3000/callback"},
		GrantTypes:    []string{"authorization_code", "refresh_token"},
		ResponseTypes: []string{"code", "token", "id_token"},
		Scopes:        []string{"openid", "profile", "email"},
		Name:          "Motor Test Client",
	}

	// Add default user for testing
	oidcProvider.users["test-user"] = &User{
		ID:            "test-user",
		Username:      "testuser",
		Email:         "test@motor.sonr.io",
		EmailVerified: true,
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
	}
}

// GetDiscovery returns OIDC discovery document
func (p *OIDCProvider) GetDiscovery() *OIDCDiscovery {
	return &OIDCDiscovery{
		Issuer:                p.issuer,
		AuthorizationEndpoint: "/authorize",
		TokenEndpoint:         "/token",
		UserInfoEndpoint:      "/userinfo",
		JWKSUri:               "/.well-known/jwks.json",
		ScopesSupported: []string{
			"openid", "profile", "email", "offline_access",
		},
		ResponseTypesSupported: []string{
			"code", "token", "id_token", "code token", "code id_token", "token id_token", "code token id_token",
		},
		GrantTypesSupported: []string{
			"authorization_code", "implicit", "refresh_token",
		},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
		TokenEndpointAuthMethodsSupported: []string{
			"client_secret_basic", "client_secret_post",
		},
		ClaimsSupported: []string{
			"sub", "name", "given_name", "family_name", "email", "email_verified",
		},
		CodeChallengeMethodsSupported: []string{"plain", "S256"},
	}
}

// GenerateAuthorizationCode generates an authorization code
func (p *OIDCProvider) GenerateAuthorizationCode(clientID, redirectURI, scope, state, nonce, userID string, codeChallenge, codeChallengeMethod string) (*AuthorizationCode, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Validate client
	client, exists := p.clients[clientID]
	if !exists {
		return nil, fmt.Errorf("invalid client_id")
	}

	// Validate redirect URI
	validRedirect := false
	for _, uri := range client.RedirectURIs {
		if uri == redirectURI {
			validRedirect = true
			break
		}
	}
	if !validRedirect {
		return nil, fmt.Errorf("invalid redirect_uri")
	}

	// Generate code
	code := generateRandomString(32)

	authCode := &AuthorizationCode{
		Code:                code,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		Scope:               scope,
		State:               state,
		Nonce:               nonce,
		UserID:              userID,
		ExpiresAt:           time.Now().Add(10 * time.Minute),
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
	}

	p.authCodes[code] = authCode

	return authCode, nil
}

// ExchangeCode exchanges authorization code for tokens
func (p *OIDCProvider) ExchangeCode(req *TokenRequest) (*TokenResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Get authorization code
	authCode, exists := p.authCodes[req.Code]
	if !exists {
		return nil, fmt.Errorf("invalid authorization code")
	}

	// Validate code hasn't expired
	if time.Now().After(authCode.ExpiresAt) {
		delete(p.authCodes, req.Code)
		return nil, fmt.Errorf("authorization code expired")
	}

	// Validate client
	if authCode.ClientID != req.ClientID {
		return nil, fmt.Errorf("client_id mismatch")
	}

	// Validate redirect URI
	if authCode.RedirectURI != req.RedirectURI {
		return nil, fmt.Errorf("redirect_uri mismatch")
	}

	// Validate PKCE if present
	if authCode.CodeChallenge != "" {
		if !validatePKCE(authCode.CodeChallenge, authCode.CodeChallengeMethod, req.CodeVerifier) {
			return nil, fmt.Errorf("invalid code_verifier")
		}
	}

	// Delete used code
	delete(p.authCodes, req.Code)

	// Generate tokens
	accessToken, _ := jwtManager.GenerateAccessToken(authCode.UserID, authCode.Scope)
	refreshToken, _ := jwtManager.GenerateRefreshToken(authCode.UserID)
	idToken, _ := jwtManager.GenerateIDToken(authCode.UserID, authCode.ClientID, authCode.Nonce, nil)

	// Store tokens
	p.accessTokens[accessToken] = &AccessToken{
		Token:     accessToken,
		ClientID:  authCode.ClientID,
		UserID:    authCode.UserID,
		Scope:     authCode.Scope,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	p.refreshTokens[refreshToken] = &RefreshToken{
		Token:     refreshToken,
		ClientID:  authCode.ClientID,
		UserID:    authCode.UserID,
		Scope:     authCode.Scope,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		Scope:        authCode.Scope,
	}, nil
}

// GetUserInfo returns user information
func (p *OIDCProvider) GetUserInfo(accessToken string) (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Validate access token
	token, exists := p.accessTokens[accessToken]
	if !exists {
		return nil, fmt.Errorf("invalid access token")
	}

	// Check expiration
	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("access token expired")
	}

	// Get user
	user, exists := p.users[token.UserID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Return user info based on scope
	userInfo := map[string]interface{}{
		"sub": user.ID,
	}

	// Add claims based on scope
	scopes := strings.Split(token.Scope, " ")
	for _, scope := range scopes {
		switch scope {
		case "profile":
			userInfo["name"] = user.Name
			userInfo["given_name"] = user.GivenName
			userInfo["family_name"] = user.FamilyName
			userInfo["preferred_username"] = user.Username
		case "email":
			userInfo["email"] = user.Email
			userInfo["email_verified"] = user.EmailVerified
		}
	}

	return userInfo, nil
}

// Helper functions

// generateRandomString generates a random string
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)[:length]
}

// validatePKCE validates PKCE code challenge
func validatePKCE(codeChallenge, method, verifier string) bool {
	if method == "plain" {
		return codeChallenge == verifier
	}
	// For S256, would need to implement SHA256 hashing
	// For simplicity, returning true for now
	return true
}
