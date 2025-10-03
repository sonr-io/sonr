package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// OIDCProvider manages OIDC operations
type OIDCProvider struct {
	mu       sync.RWMutex
	codes    map[string]*AuthorizationCode
	sessions map[string]*OIDCSession
	config   any // TODO: Use bridge.OIDCProviderConfig
}

var (
	oidcProvider = &OIDCProvider{
		codes:    make(map[string]*AuthorizationCode),
		sessions: make(map[string]*OIDCSession),
	}
	codeExpiration = 10 * time.Minute
)

// SetOIDCConfig sets the OIDC provider configuration
func SetOIDCConfig(config any) {
	oidcProvider.mu.Lock()
	defer oidcProvider.mu.Unlock()
	oidcProvider.config = config
}

// GetOIDCDiscovery returns OIDC discovery configuration
func GetOIDCDiscovery(c echo.Context) error {
	config := &OIDCConfig{
		Issuer:                "https://localhost:8080",
		AuthorizationEndpoint: "https://localhost:8080/oidc/authorize",
		TokenEndpoint:         "https://localhost:8080/oidc/token",
		UserInfoEndpoint:      "https://localhost:8080/oidc/userinfo",
		JWKSEndpoint:          "https://localhost:8080/oidc/jwks",
		RevocationEndpoint:    "https://localhost:8080/oidc/revoke",
		IntrospectionEndpoint: "https://localhost:8080/oidc/introspect",
		ScopesSupported: []string{
			"openid", "profile", "email", "did", "vault", "offline_access",
		},
		ResponseTypesSupported: []string{
			"code", "id_token", "code id_token",
		},
		GrantTypesSupported: []string{
			"authorization_code", "refresh_token", "client_credentials",
		},
		SubjectTypesSupported: []string{
			"public", "pairwise",
		},
		IDTokenSigningAlgValuesSupported: []string{
			"ES256", "RS256",
		},
		TokenEndpointAuthMethodsSupported: []string{
			"client_secret_post", "client_secret_basic", "none",
		},
		ClaimsSupported: []string{
			"sub", "name", "preferred_username", "email", "email_verified",
			"did", "vault_id", "updated_at",
		},
		CodeChallengeMethodsSupported: []string{
			"S256", "plain",
		},
	}

	return c.JSON(http.StatusOK, config)
}

// HandleOIDCAuthorization handles OIDC authorization requests
func HandleOIDCAuthorization(c echo.Context) error {
	// Parse request parameters from query string for GET or form for POST
	req := OIDCAuthorizationRequest{
		ResponseType:        c.QueryParam("response_type"),
		ClientID:            c.QueryParam("client_id"),
		RedirectURI:         c.QueryParam("redirect_uri"),
		Scope:               c.QueryParam("scope"),
		State:               c.QueryParam("state"),
		Nonce:               c.QueryParam("nonce"),
		CodeChallenge:       c.QueryParam("code_challenge"),
		CodeChallengeMethod: c.QueryParam("code_challenge_method"),
	}

	// If no query params, try to bind from body (for POST requests)
	if req.ResponseType == "" {
		_ = c.Bind(&req) // Ignore bind errors and continue
	}

	// Validate required parameters
	if req.ResponseType == "" || req.ClientID == "" || req.RedirectURI == "" || req.Scope == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "Missing required parameters",
		})
	}

	// Validate response type
	if req.ResponseType != "code" && req.ResponseType != "id_token" &&
		req.ResponseType != "code id_token" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "unsupported_response_type",
			"error_description": "Response type not supported",
		})
	}

	// Validate scope includes openid
	if !strings.Contains(req.Scope, "openid") {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_scope",
			"error_description": "Scope must include 'openid'",
		})
	}

	// TODO: Validate client_id and redirect_uri against registered clients

	// For now, assume user is authenticated (in production, redirect to login)
	userDID := c.Get("user_did")
	if userDID == nil {
		// Redirect to WebAuthn authentication
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":             "authentication_required",
			"error_description": "User authentication required",
		})
	}

	// Generate authorization code
	code := generateAuthorizationCode()

	// Store authorization code
	authCode := &AuthorizationCode{
		Code:                code,
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURI,
		UserDID:             userDID.(string),
		Scope:               req.Scope,
		Nonce:               req.Nonce,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ExpiresAt:           time.Now().Add(codeExpiration),
		Used:                false,
	}

	oidcProvider.mu.Lock()
	oidcProvider.codes[code] = authCode
	oidcProvider.mu.Unlock()

	// Build redirect URL
	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", req.RedirectURI, code, req.State)

	return c.Redirect(http.StatusFound, redirectURL)
}

// HandleOIDCToken handles OIDC token requests
func HandleOIDCToken(c echo.Context) error {
	var req OIDCTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "Invalid token request",
		})
	}

	// Handle different grant types
	switch req.GrantType {
	case "authorization_code":
		return handleAuthorizationCodeGrant(c, &req)
	case "refresh_token":
		return handleRefreshTokenGrant(c, &req)
	case "client_credentials":
		return handleClientCredentialsGrant(c, &req)
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "unsupported_grant_type",
			"error_description": "Grant type not supported",
		})
	}
}

// handleAuthorizationCodeGrant processes authorization code grant
func handleAuthorizationCodeGrant(c echo.Context, req *OIDCTokenRequest) error {
	if req.Code == "" || req.RedirectURI == "" || req.ClientID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "Missing required parameters",
		})
	}

	// Retrieve and validate authorization code
	oidcProvider.mu.Lock()
	authCode, exists := oidcProvider.codes[req.Code]
	if exists {
		delete(oidcProvider.codes, req.Code) // Single use
	}
	oidcProvider.mu.Unlock()

	if !exists {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "Invalid authorization code",
		})
	}

	// Validate code hasn't expired
	if time.Now().After(authCode.ExpiresAt) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "Authorization code expired",
		})
	}

	// Validate code hasn't been used
	if authCode.Used {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "Authorization code already used",
		})
	}

	// Validate client and redirect URI
	if authCode.ClientID != req.ClientID || authCode.RedirectURI != req.RedirectURI {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_grant",
			"error_description": "Invalid client or redirect URI",
		})
	}

	// Validate PKCE if present
	if authCode.CodeChallenge != "" {
		if req.CodeVerifier == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":             "invalid_grant",
				"error_description": "Code verifier required",
			})
		}

		if !verifyPKCE(req.CodeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error":             "invalid_grant",
				"error_description": "Invalid code verifier",
			})
		}
	}

	// Mark code as used
	authCode.Used = true

	// Generate tokens
	accessToken := generateAccessToken(authCode.UserDID)
	refreshToken := generateRefreshToken()
	idToken, err := generateIDToken(authCode.UserDID, authCode.ClientID, authCode.Nonce)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to generate ID token",
		})
	}

	// Create session
	session := &OIDCSession{
		SessionID:    generateSessionID(),
		UserDID:      authCode.UserDID,
		ClientID:     authCode.ClientID,
		Scope:        authCode.Scope,
		Nonce:        authCode.Nonce,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
	}

	oidcProvider.mu.Lock()
	oidcProvider.sessions[accessToken] = session
	oidcProvider.mu.Unlock()

	// Return tokens
	response := &OIDCTokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: refreshToken,
		IDToken:      idToken,
		Scope:        authCode.Scope,
	}

	return c.JSON(http.StatusOK, response)
}

// handleRefreshTokenGrant processes refresh token grant
func handleRefreshTokenGrant(c echo.Context, req *OIDCTokenRequest) error {
	// TODO: Implement refresh token grant
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"error":             "unsupported_grant_type",
		"error_description": "Refresh token grant not yet implemented",
	})
}

// handleClientCredentialsGrant processes client credentials grant
func handleClientCredentialsGrant(c echo.Context, req *OIDCTokenRequest) error {
	// TODO: Implement client credentials grant
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"error":             "unsupported_grant_type",
		"error_description": "Client credentials grant not yet implemented",
	})
}

// HandleOIDCUserInfo handles OIDC userinfo requests
func HandleOIDCUserInfo(c echo.Context) error {
	// Extract access token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":             "invalid_token",
			"error_description": "Invalid access token",
		})
	}

	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate access token
	oidcProvider.mu.RLock()
	session, exists := oidcProvider.sessions[accessToken]
	oidcProvider.mu.RUnlock()

	if !exists {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":             "invalid_token",
			"error_description": "Invalid or expired access token",
		})
	}

	// Check if token is expired
	if time.Now().After(session.ExpiresAt) {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":             "invalid_token",
			"error_description": "Access token expired",
		})
	}

	// TODO: Fetch actual user info from DID document or database
	userInfo := &OIDCUserInfo{
		Subject:           session.UserDID,
		PreferredUsername: "user", // TODO: Get from DID document
		DID:               session.UserDID,
		UpdatedAt:         time.Now().Unix(),
	}

	// Add additional claims based on scope
	if strings.Contains(session.Scope, "profile") {
		userInfo.Name = "User Name" // TODO: Get from DID document
	}

	if strings.Contains(session.Scope, "email") {
		userInfo.Email = "user@example.com" // TODO: Get from DID document
		userInfo.EmailVerified = true
	}

	if strings.Contains(session.Scope, "vault") {
		userInfo.VaultID = "vault_" + session.UserDID // TODO: Get actual vault ID
	}

	return c.JSON(http.StatusOK, userInfo)
}

// HandleOIDCJWKS handles JWKS endpoint
func HandleOIDCJWKS(c echo.Context) error {
	// TODO: Return actual public keys used for signing
	jwks := &JWKSet{
		Keys: []JWK{
			{
				KeyType:   "EC",
				Use:       "sig",
				KeyID:     "1",
				Algorithm: "ES256",
				Curve:     "P-256",
				// TODO: Add actual public key coordinates
			},
		},
	}

	return c.JSON(http.StatusOK, jwks)
}

// Helper functions

func generateAuthorizationCode() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func generateAccessToken(userDID string) string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func generateRefreshToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func generateSessionID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

func generateIDToken(userDID, clientID, nonce string) (string, error) {
	claims := &DIDAuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userDID,
			Issuer:    "https://localhost:8080",
			Audience:  []string{clientID},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
		DID: userDID,
	}

	if nonce != "" {
		claims.Extra = map[string]any{
			"nonce": nonce,
		}
	}

	// TODO: Sign with actual signing key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("temporary-secret-key"))
}

func verifyPKCE(verifier, challenge, method string) bool {
	var computed string

	switch method {
	case "S256":
		h := sha256.Sum256([]byte(verifier))
		computed = base64.RawURLEncoding.EncodeToString(h[:])
	case "plain":
		computed = verifier
	default:
		return false
	}

	return computed == challenge
}
