package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sonr-io/sonr/crypto/ucan"
)

// OAuth2Provider extends OIDCProvider with full OAuth2 capabilities
type OAuth2Provider struct {
	*OIDCProvider
	clientRegistry    *ClientRegistry
	scopeMapper       *ScopeMapper
	ucanDelegator     *UCANDelegator
	authCodeStore     *AuthCodeStore
	accessTokenStore  *AccessTokenStore
	refreshTokenStore *RefreshTokenStore
	consentStore      *ConsentStore
	config            *OAuth2Config
}

// AuthCodeStore manages authorization codes
type AuthCodeStore struct {
	mu    sync.RWMutex
	codes map[string]*OAuth2AuthorizationCode
}

// AccessTokenStore manages access tokens
type AccessTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*OAuth2AccessToken
}

// RefreshTokenStore manages refresh tokens
type RefreshTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*OAuth2RefreshToken
}

// ConsentStore manages user consent records
type ConsentStore struct {
	mu       sync.RWMutex
	consents map[string]*UserConsent // key: userDID:clientID
}

// UserConsent represents stored user consent
type UserConsent struct {
	UserDID        string    `json:"user_did"`
	ClientID       string    `json:"client_id"`
	ApprovedScopes []string  `json:"approved_scopes"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

var oauth2Provider *OAuth2Provider

// InitializeOAuth2Provider initializes the OAuth2 provider
func InitializeOAuth2Provider() {
	oauth2Provider = &OAuth2Provider{
		OIDCProvider:      oidcProvider,
		clientRegistry:    NewClientRegistry(),
		scopeMapper:       NewScopeMapper(),
		ucanDelegator:     NewUCANDelegator(nil),
		authCodeStore:     &AuthCodeStore{codes: make(map[string]*OAuth2AuthorizationCode)},
		accessTokenStore:  &AccessTokenStore{tokens: make(map[string]*OAuth2AccessToken)},
		refreshTokenStore: &RefreshTokenStore{tokens: make(map[string]*OAuth2RefreshToken)},
		consentStore:      &ConsentStore{consents: make(map[string]*UserConsent)},
		config:            getDefaultOAuth2Config(),
	}

	// Start cleanup goroutine for expired tokens
	go oauth2Provider.cleanupExpiredTokens()
}

// GetOAuth2Discovery returns OAuth2 discovery configuration
func GetOAuth2Discovery(c echo.Context) error {
	if oauth2Provider == nil {
		InitializeOAuth2Provider()
	}
	return c.JSON(http.StatusOK, oauth2Provider.config)
}

// HandleOAuth2Authorize handles OAuth2 authorization requests
func HandleOAuth2Authorize(c echo.Context) error {
	if oauth2Provider == nil {
		InitializeOAuth2Provider()
	}

	req := &OAuth2AuthorizationRequest{
		ResponseType:        c.QueryParam("response_type"),
		ClientID:            c.QueryParam("client_id"),
		RedirectURI:         c.QueryParam("redirect_uri"),
		Scope:               c.QueryParam("scope"),
		State:               c.QueryParam("state"),
		CodeChallenge:       c.QueryParam("code_challenge"),
		CodeChallengeMethod: c.QueryParam("code_challenge_method"),
		Nonce:               c.QueryParam("nonce"),
		Prompt:              c.QueryParam("prompt"),
		LoginHint:           c.QueryParam("login_hint"),
	}

	// Validate client
	client, err := oauth2Provider.clientRegistry.GetClient(req.ClientID)
	if err != nil {
		return oauth2Error(c, "invalid_client", "Unknown client", req.State)
	}

	// Validate redirect URI
	if !client.ValidateRedirectURI(req.RedirectURI) {
		return oauth2Error(c, "invalid_request", "Invalid redirect URI", req.State)
	}

	// Validate response type
	if !isValidResponseType(req.ResponseType) {
		return redirectError(
			c,
			req.RedirectURI,
			"unsupported_response_type",
			"Response type not supported",
			req.State,
		)
	}

	// Validate scopes
	requestedScopes := parseScopes(req.Scope)
	if !client.ValidateScopes(requestedScopes) {
		return redirectError(
			c,
			req.RedirectURI,
			"invalid_scope",
			"Requested scope not allowed",
			req.State,
		)
	}

	// Validate PKCE for public clients
	if client.ClientType == "public" && client.RequirePKCE {
		if req.CodeChallenge == "" {
			return redirectError(
				c,
				req.RedirectURI,
				"invalid_request",
				"PKCE required for public clients",
				req.State,
			)
		}
		if req.CodeChallengeMethod != PKCEMethodS256 {
			return redirectError(
				c,
				req.RedirectURI,
				"invalid_request",
				"Only S256 PKCE method supported",
				req.State,
			)
		}
	}

	// Check authentication
	userDID := c.Get("user_did")
	if userDID == nil {
		// Store authorization request and redirect to authentication
		sessionID := generateSessionID()
		// TODO: Store auth request in session store
		authURL := fmt.Sprintf(
			"/auth/login?session_id=%s&return_to=%s",
			sessionID,
			c.Request().URL.String(),
		)
		return c.Redirect(http.StatusFound, authURL)
	}

	// Check consent
	if client.RequiresConsent &&
		!oauth2Provider.hasValidConsent(userDID.(string), req.ClientID, requestedScopes) {
		// Render consent page
		return renderOAuth2ConsentPage(c, req, client)
	}

	// Generate authorization code
	code := generateSecureToken(32)
	authCode := &OAuth2AuthorizationCode{
		Code:                code,
		ClientID:            req.ClientID,
		UserDID:             userDID.(string),
		RedirectURI:         req.RedirectURI,
		Scopes:              requestedScopes,
		State:               req.State,
		Nonce:               req.Nonce,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ExpiresAt:           time.Now().Add(10 * time.Minute),
		UCANContext:         oauth2Provider.buildUCANContext(userDID.(string)),
	}

	// Store authorization code
	oauth2Provider.authCodeStore.Store(authCode)

	// Build redirect URL
	redirectURL := buildAuthorizationRedirect(req.RedirectURI, code, req.State)
	return c.Redirect(http.StatusFound, redirectURL)
}

// HandleOAuth2Token handles OAuth2 token requests
func HandleOAuth2Token(c echo.Context) error {
	if oauth2Provider == nil {
		InitializeOAuth2Provider()
	}

	var req OAuth2TokenRequest
	if err := c.Bind(&req); err != nil {
		return oauth2TokenError(c, "invalid_request", "Invalid token request")
	}

	// Authenticate client
	client, err := oauth2Provider.authenticateClient(c, &req)
	if err != nil {
		return oauth2TokenError(c, "invalid_client", "Client authentication failed")
	}

	// Handle grant type
	switch req.GrantType {
	case "authorization_code":
		return oauth2Provider.handleAuthorizationCodeGrant(c, client, &req)
	case "refresh_token":
		return oauth2Provider.handleRefreshTokenGrant(c, client, &req)
	case "client_credentials":
		return oauth2Provider.handleClientCredentialsGrant(c, client, &req)
	default:
		return oauth2TokenError(c, "unsupported_grant_type", "Grant type not supported")
	}
}

// HandleOAuth2Introspection handles token introspection requests
func HandleOAuth2Introspection(c echo.Context) error {
	if oauth2Provider == nil {
		InitializeOAuth2Provider()
	}

	var req OAuth2IntrospectionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, &OAuth2IntrospectionResponse{Active: false})
	}

	// Authenticate client
	client, err := oauth2Provider.authenticateClient(c, &OAuth2TokenRequest{
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
	})
	if err != nil {
		return c.JSON(http.StatusUnauthorized, &OAuth2IntrospectionResponse{Active: false})
	}

	// Introspect token
	response := oauth2Provider.introspectToken(req.Token, req.TokenTypeHint, client)
	return c.JSON(http.StatusOK, response)
}

// HandleOAuth2Revocation handles token revocation requests
func HandleOAuth2Revocation(c echo.Context) error {
	if oauth2Provider == nil {
		InitializeOAuth2Provider()
	}

	var req OAuth2RevocationRequest
	if err := c.Bind(&req); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	// Authenticate client
	client, err := oauth2Provider.authenticateClient(c, &OAuth2TokenRequest{
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
	})
	if err != nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	// Revoke token
	oauth2Provider.revokeToken(req.Token, req.TokenTypeHint, client)
	return c.NoContent(http.StatusOK)
}

// Private methods

func (p *OAuth2Provider) handleAuthorizationCodeGrant(
	c echo.Context,
	client *OAuth2Client,
	req *OAuth2TokenRequest,
) error {
	// Retrieve authorization code
	authCode := p.authCodeStore.Exchange(req.Code)
	if authCode == nil {
		return oauth2TokenError(c, "invalid_grant", "Invalid authorization code")
	}

	// Validate code hasn't expired
	if time.Now().After(authCode.ExpiresAt) {
		return oauth2TokenError(c, "invalid_grant", "Authorization code expired")
	}

	// Validate client
	if authCode.ClientID != client.ClientID {
		return oauth2TokenError(c, "invalid_grant", "Code was issued to different client")
	}

	// Validate redirect URI
	if authCode.RedirectURI != req.RedirectURI {
		return oauth2TokenError(c, "invalid_grant", "Redirect URI mismatch")
	}

	// Validate PKCE if present
	if authCode.CodeChallenge != "" {
		if !p.validatePKCE(req.CodeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod) {
			return oauth2TokenError(c, "invalid_grant", "Invalid PKCE verifier")
		}
	}

	// Create UCAN delegation
	ucanToken, err := p.ucanDelegator.CreateDelegation(
		authCode.UserDID,
		client.ClientID,
		authCode.Scopes,
		time.Now().Add(time.Hour),
	)
	if err != nil {
		return oauth2TokenError(c, "server_error", "Failed to create delegation")
	}

	// Generate tokens
	accessToken := p.generateAccessToken(authCode, ucanToken)
	refreshToken := p.generateRefreshToken(authCode)

	// Store tokens
	p.accessTokenStore.Store(accessToken)
	p.refreshTokenStore.Store(refreshToken)

	// Generate ID token if openid scope present
	var idToken string
	if contains(authCode.Scopes, "openid") {
		idToken, _ = generateIDToken(authCode.UserDID, client.ClientID, authCode.Nonce)
	}

	// Return token response
	response := &OAuth2TokenResponse{
		AccessToken:  accessToken.Token,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: refreshToken.Token,
		Scope:        strings.Join(authCode.Scopes, " "),
		IDToken:      idToken,
		UCANToken:    ucanToken.Raw,
	}

	return c.JSON(http.StatusOK, response)
}

func (p *OAuth2Provider) handleRefreshTokenGrant(
	c echo.Context,
	client *OAuth2Client,
	req *OAuth2TokenRequest,
) error {
	// Retrieve refresh token
	oldRefreshToken := p.refreshTokenStore.Get(req.RefreshToken)
	if oldRefreshToken == nil {
		return oauth2TokenError(c, "invalid_grant", "Invalid refresh token")
	}

	// Validate client
	if oldRefreshToken.ClientID != client.ClientID {
		return oauth2TokenError(c, "invalid_grant", "Token was issued to different client")
	}

	// Validate expiration
	if time.Now().After(oldRefreshToken.ExpiresAt) {
		return oauth2TokenError(c, "invalid_grant", "Refresh token expired")
	}

	// Rotate refresh token
	p.refreshTokenStore.Revoke(req.RefreshToken)

	// Create new UCAN delegation
	ucanToken, err := p.ucanDelegator.CreateDelegation(
		oldRefreshToken.UserDID,
		client.ClientID,
		oldRefreshToken.Scopes,
		time.Now().Add(time.Hour),
	)
	if err != nil {
		return oauth2TokenError(c, "server_error", "Failed to create delegation")
	}

	// Generate new tokens
	newAccessToken := &OAuth2AccessToken{
		Token:     generateSecureToken(32),
		UserDID:   oldRefreshToken.UserDID,
		ClientID:  client.ClientID,
		Scopes:    oldRefreshToken.Scopes,
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
		UCANToken: ucanToken,
	}

	newRefreshToken := &OAuth2RefreshToken{
		Token:         generateSecureToken(32),
		AccessToken:   newAccessToken.Token,
		ClientID:      client.ClientID,
		UserDID:       oldRefreshToken.UserDID,
		Scopes:        oldRefreshToken.Scopes,
		ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
		IssuedAt:      time.Now(),
		RotationCount: oldRefreshToken.RotationCount + 1,
	}

	// Store new tokens
	p.accessTokenStore.Store(newAccessToken)
	p.refreshTokenStore.Store(newRefreshToken)

	// Return response
	response := &OAuth2TokenResponse{
		AccessToken:  newAccessToken.Token,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: newRefreshToken.Token,
		Scope:        strings.Join(newAccessToken.Scopes, " "),
		UCANToken:    ucanToken.Raw,
	}

	return c.JSON(http.StatusOK, response)
}

func (p *OAuth2Provider) handleClientCredentialsGrant(
	c echo.Context,
	client *OAuth2Client,
	req *OAuth2TokenRequest,
) error {
	// Client credentials grant is only for confidential clients
	if client.ClientType != "confidential" {
		return oauth2TokenError(
			c,
			"unauthorized_client",
			"Client type not authorized for this grant",
		)
	}

	// Parse requested scopes
	scopes := parseScopes(req.Scope)
	if !client.ValidateScopes(scopes) {
		return oauth2TokenError(c, "invalid_scope", "Requested scope not allowed")
	}

	// Create service-to-service UCAN token
	ucanToken, err := p.ucanDelegator.CreateServiceDelegation(
		client.ClientID,
		scopes,
		time.Now().Add(time.Hour),
	)
	if err != nil {
		return oauth2TokenError(c, "server_error", "Failed to create delegation")
	}

	// Generate access token
	accessToken := &OAuth2AccessToken{
		Token:     generateSecureToken(32),
		UserDID:   "", // No user for client credentials
		ClientID:  client.ClientID,
		Scopes:    scopes,
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
		UCANToken: ucanToken,
		TokenType: "client_credentials",
	}

	// Store token
	p.accessTokenStore.Store(accessToken)

	// Return response
	response := &OAuth2TokenResponse{
		AccessToken: accessToken.Token,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       strings.Join(scopes, " "),
		UCANToken:   ucanToken.Raw,
	}

	return c.JSON(http.StatusOK, response)
}

func (p *OAuth2Provider) authenticateClient(
	c echo.Context,
	req *OAuth2TokenRequest,
) (*OAuth2Client, error) {
	// Try Basic Auth first
	if username, password, ok := c.Request().BasicAuth(); ok {
		client, err := p.clientRegistry.GetClient(username)
		if err != nil {
			return nil, err
		}
		if client.ClientType == "confidential" &&
			subtle.ConstantTimeCompare([]byte(client.ClientSecret), []byte(password)) == 1 {
			return client, nil
		}
		return nil, fmt.Errorf("invalid client credentials")
	}

	// Try client_secret_post
	if req.ClientID != "" && req.ClientSecret != "" {
		client, err := p.clientRegistry.GetClient(req.ClientID)
		if err != nil {
			return nil, err
		}
		if client.ClientType == "confidential" &&
			subtle.ConstantTimeCompare([]byte(client.ClientSecret), []byte(req.ClientSecret)) == 1 {
			return client, nil
		}
		return nil, fmt.Errorf("invalid client credentials")
	}

	// Try client_assertion (JWT)
	if req.ClientAssertion != "" &&
		req.ClientAssertionType == "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
		// TODO: Implement JWT client assertion validation
		return nil, fmt.Errorf("JWT client assertion not yet implemented")
	}

	// Public client (no authentication)
	if req.ClientID != "" {
		client, err := p.clientRegistry.GetClient(req.ClientID)
		if err != nil {
			return nil, err
		}
		if client.ClientType == "public" {
			return client, nil
		}
	}

	return nil, fmt.Errorf("client authentication required")
}

func (p *OAuth2Provider) validatePKCE(verifier, challenge, method string) bool {
	if method == "" {
		method = PKCEMethodPlain
	}
	computed := computePKCEChallenge(verifier, method)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}

func (p *OAuth2Provider) hasValidConsent(userDID, clientID string, scopes []string) bool {
	p.consentStore.mu.RLock()
	defer p.consentStore.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", userDID, clientID)
	consent, exists := p.consentStore.consents[key]
	if !exists {
		return false
	}

	// Check expiration
	if time.Now().After(consent.ExpiresAt) {
		return false
	}

	// Check all requested scopes are approved
	for _, scope := range scopes {
		if !contains(consent.ApprovedScopes, scope) {
			return false
		}
	}

	return true
}

func (p *OAuth2Provider) buildUCANContext(userDID string) *UCANAuthContext {
	// TODO: Fetch actual vault and DID document data
	return &UCANAuthContext{
		VaultAddress:   fmt.Sprintf("vault_%s", userDID),
		EnclaveDataCID: fmt.Sprintf("cid_%s", userDID),
		Capabilities:   []string{"read", "write", "sign"},
	}
}

func (p *OAuth2Provider) generateAccessToken(
	authCode *OAuth2AuthorizationCode,
	ucanToken *ucan.Token,
) *OAuth2AccessToken {
	return &OAuth2AccessToken{
		Token:     generateSecureToken(32),
		UserDID:   authCode.UserDID,
		ClientID:  authCode.ClientID,
		Scopes:    authCode.Scopes,
		ExpiresAt: time.Now().Add(time.Hour),
		IssuedAt:  time.Now(),
		UCANToken: ucanToken,
		SessionID: generateSessionID(),
		TokenType: "authorization_code",
	}
}

func (p *OAuth2Provider) generateRefreshToken(
	authCode *OAuth2AuthorizationCode,
) *OAuth2RefreshToken {
	return &OAuth2RefreshToken{
		Token:         generateSecureToken(32),
		ClientID:      authCode.ClientID,
		UserDID:       authCode.UserDID,
		Scopes:        authCode.Scopes,
		ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
		IssuedAt:      time.Now(),
		RotationCount: 0,
	}
}

func (p *OAuth2Provider) introspectToken(
	token, tokenTypeHint string,
	client *OAuth2Client,
) *OAuth2IntrospectionResponse {
	// Try access token first
	if accessToken := p.accessTokenStore.Get(token); accessToken != nil {
		if accessToken.ClientID != client.ClientID {
			return &OAuth2IntrospectionResponse{Active: false}
		}
		return &OAuth2IntrospectionResponse{
			Active:    time.Now().Before(accessToken.ExpiresAt),
			Scope:     strings.Join(accessToken.Scopes, " "),
			ClientID:  accessToken.ClientID,
			Username:  accessToken.UserDID,
			TokenType: "Bearer",
			ExpiresAt: accessToken.ExpiresAt.Unix(),
			IssuedAt:  accessToken.IssuedAt.Unix(),
			Subject:   accessToken.UserDID,
			UCANToken: accessToken.UCANToken.Raw,
		}
	}

	// Try refresh token
	if refreshToken := p.refreshTokenStore.Get(token); refreshToken != nil {
		if refreshToken.ClientID != client.ClientID {
			return &OAuth2IntrospectionResponse{Active: false}
		}
		return &OAuth2IntrospectionResponse{
			Active:    time.Now().Before(refreshToken.ExpiresAt),
			Scope:     strings.Join(refreshToken.Scopes, " "),
			ClientID:  refreshToken.ClientID,
			Username:  refreshToken.UserDID,
			TokenType: "refresh_token",
			ExpiresAt: refreshToken.ExpiresAt.Unix(),
			IssuedAt:  refreshToken.IssuedAt.Unix(),
			Subject:   refreshToken.UserDID,
		}
	}

	return &OAuth2IntrospectionResponse{Active: false}
}

func (p *OAuth2Provider) revokeToken(token, tokenTypeHint string, client *OAuth2Client) {
	// Try to revoke as access token
	if p.accessTokenStore.Revoke(token) {
		return
	}

	// Try to revoke as refresh token
	p.refreshTokenStore.Revoke(token)
}

func (p *OAuth2Provider) cleanupExpiredTokens() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Cleanup expired authorization codes
		p.authCodeStore.CleanupExpired()

		// Cleanup expired access tokens
		p.accessTokenStore.CleanupExpired()

		// Cleanup expired refresh tokens
		p.refreshTokenStore.CleanupExpired()
	}
}

// Store methods for token stores

func (s *AuthCodeStore) Store(code *OAuth2AuthorizationCode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[code.Code] = code
}

func (s *AuthCodeStore) Exchange(code string) *OAuth2AuthorizationCode {
	s.mu.Lock()
	defer s.mu.Unlock()

	authCode, exists := s.codes[code]
	if !exists || authCode.Used {
		return nil
	}

	authCode.Used = true
	return authCode
}

func (s *AuthCodeStore) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for code, authCode := range s.codes {
		if now.After(authCode.ExpiresAt) {
			delete(s.codes, code)
		}
	}
}

func (s *AccessTokenStore) Store(token *OAuth2AccessToken) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token.Token] = token
}

func (s *AccessTokenStore) Get(token string) *OAuth2AccessToken {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tokens[token]
}

func (s *AccessTokenStore) Revoke(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tokens[token]; exists {
		delete(s.tokens, token)
		return true
	}
	return false
}

func (s *AccessTokenStore) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, accessToken := range s.tokens {
		if now.After(accessToken.ExpiresAt) {
			delete(s.tokens, token)
		}
	}
}

func (s *RefreshTokenStore) Store(token *OAuth2RefreshToken) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token.Token] = token
}

func (s *RefreshTokenStore) Get(token string) *OAuth2RefreshToken {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tokens[token]
}

func (s *RefreshTokenStore) Revoke(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tokens[token]; exists {
		delete(s.tokens, token)
		return true
	}
	return false
}

func (s *RefreshTokenStore) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, refreshToken := range s.tokens {
		if now.After(refreshToken.ExpiresAt) {
			delete(s.tokens, token)
		}
	}
}

// Helper functions

func generateSecureToken(bytes int) string {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func parseScopes(scope string) []string {
	if scope == "" {
		return []string{}
	}
	return strings.Split(scope, " ")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isValidResponseType(responseType string) bool {
	validTypes := []string{
		"code",
		"token",
		"id_token",
		"code id_token",
		"code token",
		"id_token token",
		"code id_token token",
	}
	return contains(validTypes, responseType)
}

func oauth2Error(c echo.Context, error, description, state string) error {
	return c.JSON(http.StatusBadRequest, &OAuth2ErrorResponse{
		Error:            error,
		ErrorDescription: description,
		State:            state,
	})
}

func oauth2TokenError(c echo.Context, error, description string) error {
	return c.JSON(http.StatusBadRequest, &OAuth2ErrorResponse{
		Error:            error,
		ErrorDescription: description,
	})
}

func redirectError(c echo.Context, redirectURI, error, description, state string) error {
	url := fmt.Sprintf("%s?error=%s&error_description=%s&state=%s",
		redirectURI, error, description, state)
	return c.Redirect(http.StatusFound, url)
}

func buildAuthorizationRedirect(redirectURI, code, state string) string {
	if strings.Contains(redirectURI, "?") {
		return fmt.Sprintf("%s&code=%s&state=%s", redirectURI, code, state)
	}
	return fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, state)
}

func renderOAuth2ConsentPage(
	c echo.Context,
	req *OAuth2AuthorizationRequest,
	client *OAuth2Client,
) error {
	// TODO: Render actual consent page
	return c.JSON(http.StatusOK, map[string]any{
		"client":    client,
		"scopes":    parseScopes(req.Scope),
		"state":     req.State,
		"client_id": req.ClientID,
	})
}

func getDefaultOAuth2Config() *OAuth2Config {
	baseURL := "https://localhost:8080"
	return &OAuth2Config{
		Issuer:                baseURL,
		AuthorizationEndpoint: baseURL + "/oauth2/authorize",
		TokenEndpoint:         baseURL + "/oauth2/token",
		UserInfoEndpoint:      baseURL + "/oauth2/userinfo",
		JWKSEndpoint:          baseURL + "/oauth2/jwks",
		RegistrationEndpoint:  baseURL + "/oauth2/register",
		IntrospectionEndpoint: baseURL + "/oauth2/introspect",
		RevocationEndpoint:    baseURL + "/oauth2/revoke",
		ScopesSupported: []string{
			"openid", "profile", "email", "offline_access",
			"vault:read", "vault:write", "vault:sign", "vault:admin",
			"service:manage", "did:read", "did:write",
		},
		ResponseTypesSupported: []string{
			"code", "token", "id_token",
			"code id_token", "code token",
			"id_token token", "code id_token token",
		},
		ResponseModesSupported: []string{
			"query", "fragment", "form_post",
		},
		GrantTypesSupported: []string{
			"authorization_code", "implicit", "refresh_token",
			"client_credentials", "urn:ietf:params:oauth:grant-type:device_code",
		},
		SubjectTypesSupported: []string{
			"public", "pairwise",
		},
		IDTokenSigningAlgValuesSupported: []string{
			"ES256", "RS256", "HS256",
		},
		TokenEndpointAuthMethodsSupported: []string{
			"client_secret_basic", "client_secret_post",
			"client_secret_jwt", "private_key_jwt", "none",
		},
		ClaimsSupported: []string{
			"sub", "iss", "aud", "exp", "iat", "auth_time", "nonce",
			"name", "given_name", "family_name", "middle_name", "nickname",
			"preferred_username", "profile", "picture", "website", "email",
			"email_verified", "did", "vault_id", "ucan_capabilities",
		},
		CodeChallengeMethodsSupported: []string{
			PKCEMethodS256, PKCEMethodPlain,
		},
		ServiceDocumentation: baseURL + "/docs/oauth2",
		UILocalesSupported:   []string{"en-US"},
		UCANSupported:        true,
	}
}
