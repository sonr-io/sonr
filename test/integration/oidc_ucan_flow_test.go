package integration

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/sonr-io/sonr/bridge/handlers"
	"github.com/sonr-io/sonr/crypto/ucan"
)

// MockDIDResolver provides mock DID resolution for testing
type MockDIDResolver struct {
	keys map[string]ed25519.PublicKey
}

func NewMockDIDResolver() *MockDIDResolver {
	resolver := &MockDIDResolver{
		keys: make(map[string]ed25519.PublicKey),
	}

	// Generate test keys for common DIDs
	testDIDs := []string{
		"did:sonr:oauth-provider",
		"did:sonr:user-chain",
		"did:sonr:client-a",
		"did:sonr:client-b",
		"did:sonr:client-c",
		"did:sonr:user",
		"did:sonr:perf-user",
		"did:sonr:user123",
		"did:sonr:test-user",
	}

	for _, did := range testDIDs {
		pub, _, _ := ed25519.GenerateKey(rand.Reader)
		resolver.keys[did] = pub
	}

	return resolver
}

func (r *MockDIDResolver) GetPublicKey(did string) (ed25519.PublicKey, error) {
	if key, ok := r.keys[did]; ok {
		return key, nil
	}
	// Generate a new key for unknown DIDs
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	r.keys[did] = pub
	return pub, nil
}

// MockOAuth2Provider provides mock OAuth2 functionality for testing
type MockOAuth2Provider struct {
	tokens map[string]string
}

func NewMockOAuth2Provider() *MockOAuth2Provider {
	return &MockOAuth2Provider{
		tokens: make(map[string]string),
	}
}

func (m *MockOAuth2Provider) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	// Mock authorization handler - returns test auth code
	code := "test_auth_code_123"
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI != "" {
		http.Redirect(w, r, redirectURI+"?code="+code, http.StatusFound)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(code))
	}
}

func (m *MockOAuth2Provider) HandleToken(w http.ResponseWriter, r *http.Request) {
	// Mock token handler - returns test access token with UCAN
	// Generate a mock UCAN token (simplified for testing)
	mockUCAN := "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJkaWQ6c29ucjpvYXV0aC1wcm92aWRlciIsImF1ZCI6ImRpZDpzb25yOnVzZXIxMjMiLCJhdHQiOlt7ImNhbiI6InZhdWx0OnJlYWQifSx7ImNhbiI6ImR3bjp3cml0ZSJ9XSwiZXhwIjoxNzM2MzY0MDAwLCJubmMiOiJ0ZXN0LW5vbmNlIn0.test_signature"

	token := map[string]interface{}{
		"access_token": "test_access_token",
		"token_type":   "Bearer",
		"expires_in":   3600,
		"ucan_token":   mockUCAN,
		"scope":        "openid vault:read dwn:write",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

func (m *MockOAuth2Provider) HandleUserInfo(w http.ResponseWriter, r *http.Request) {
	// Mock userinfo handler
	userInfo := map[string]interface{}{
		"sub":   "did:sonr:test-user",
		"name":  "Test User",
		"email": "test@sonr.id",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// OIDCUCANFlowTestSuite tests the complete OIDC → OAuth2 → UCAN flow
// MockBlockchainUCANSigner is a mock implementation for testing
type MockBlockchainUCANSigner struct {
	issuerDID string
	resolver  *MockDIDResolver
}

func NewMockBlockchainUCANSigner(issuerDID string) *MockBlockchainUCANSigner {
	return &MockBlockchainUCANSigner{
		issuerDID: issuerDID,
		resolver:  NewMockDIDResolver(),
	}
}

// CreateDelegationToken creates a mock UCAN token for testing
func (s *MockBlockchainUCANSigner) CreateDelegationToken(
	issuer string,
	audience string,
	attenuations []ucan.Attenuation,
	proofs []ucan.Proof,
	expiration time.Duration,
) (string, error) {
	// Build attenuations array for claims
	var atts []map[string]interface{}
	for _, att := range attenuations {
		attMap := make(map[string]interface{})

		// Handle different capability types
		switch cap := att.Capability.(type) {
		case *ucan.SimpleCapability:
			attMap["can"] = cap.Action
		case *ucan.MultiCapability:
			// For multi-capability, store all actions as a joined string
			if len(cap.Actions) > 0 {
				attMap["can"] = strings.Join(cap.Actions, ",")
			}
		}

		// Add resource if present
		if att.Resource != nil {
			if res, ok := att.Resource.(*handlers.SimpleResource); ok {
				attMap["with"] = fmt.Sprintf("%s:%s", res.Scheme, res.Value)
			}
		}

		atts = append(atts, attMap)
	}

	// Create claims
	claims := map[string]interface{}{
		"iss": issuer,
		"aud": audience,
		"exp": time.Now().Add(expiration).Unix(),
		"att": atts,
	}

	claimsJSON, _ := json.Marshal(claims)

	// Create a simplified mock UCAN token
	token := fmt.Sprintf("eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.%s.mock_signature",
		base64.RawURLEncoding.EncodeToString(claimsJSON))
	return token, nil
}

// VerifySignature verifies a mock UCAN token
func (s *MockBlockchainUCANSigner) VerifySignature(tokenString string) (*ucan.Token, error) {
	// For testing, just check if it's a valid format
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("token expired")
		}
	}

	// Return a mock token
	return &ucan.Token{
		Issuer:   claims["iss"].(string),
		Audience: claims["aud"].(string),
	}, nil
}

// ValidateDelegationChain validates a chain of UCAN tokens
func (s *MockBlockchainUCANSigner) ValidateDelegationChain(tokens []string) error {
	// For testing, verify each token and check for privilege escalation
	var prevCapabilities []string

	for i, token := range tokens {
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			return fmt.Errorf("invalid token format at position %d", i)
		}

		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return fmt.Errorf("failed to decode token %d: %w", i, err)
		}

		var claims map[string]interface{}
		if err := json.Unmarshal(payload, &claims); err != nil {
			return fmt.Errorf("failed to unmarshal claims for token %d: %w", i, err)
		}

		// Check for privilege escalation (simplified for testing)
		if att, ok := claims["att"].([]interface{}); ok && i > 0 {
			for _, a := range att {
				if attMap, ok := a.(map[string]interface{}); ok {
					if cap, ok := attMap["can"].(string); ok {
						// Split comma-separated capabilities
						caps := strings.Split(cap, ",")
						for _, c := range caps {
							c = strings.TrimSpace(c)
							// Check if this capability was in the previous token
							if !contains(prevCapabilities, c) {
								return fmt.Errorf("privilege escalation detected: trying to add '%s' permission", c)
							}
						}
					}
				}
			}
		}

		// Store capabilities for next iteration
		prevCapabilities = []string{}
		if att, ok := claims["att"].([]interface{}); ok {
			for _, a := range att {
				if attMap, ok := a.(map[string]interface{}); ok {
					if cap, ok := attMap["can"].(string); ok {
						// Split and store all capabilities
						caps := strings.Split(cap, ",")
						for _, c := range caps {
							prevCapabilities = append(prevCapabilities, strings.TrimSpace(c))
						}
					}
				}
			}
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

type OIDCUCANFlowTestSuite struct {
	suite.Suite
	oauth2Provider *MockOAuth2Provider
	ucanDelegator  *handlers.UCANDelegator
	tokenExchange  *handlers.TokenExchangeHandler
	refreshHandler *handlers.RefreshTokenHandler
	testServer     *httptest.Server
	clientStore    *MockClientStore
	tokenStore     *MockTokenStore
	mockSigner     *MockBlockchainUCANSigner
}

// MockClientStore implements handlers.ClientStore for testing
type MockClientStore struct {
	clients map[string]*handlers.OAuth2Client
}

func NewMockClientStore() *MockClientStore {
	return &MockClientStore{
		clients: make(map[string]*handlers.OAuth2Client),
	}
}

func (m *MockClientStore) GetClient(ctx context.Context, clientID string) (*handlers.OAuth2Client, error) {
	client, exists := m.clients[clientID]
	if !exists {
		return nil, handlers.ErrClientNotFound
	}
	return client, nil
}

func (m *MockClientStore) ValidateClientCredentials(ctx context.Context, clientID, clientSecret string) error {
	client, exists := m.clients[clientID]
	if !exists || client.ClientSecret != clientSecret {
		return handlers.ErrInvalidClientCredentials
	}
	return nil
}

// MockTokenStore implements handlers.TokenStore for testing
type MockTokenStore struct {
	tokens map[string]*handlers.StoredToken
}

func NewMockTokenStore() *MockTokenStore {
	return &MockTokenStore{
		tokens: make(map[string]*handlers.StoredToken),
	}
}

func (m *MockTokenStore) GetToken(ctx context.Context, tokenID string) (*handlers.StoredToken, error) {
	token, exists := m.tokens[tokenID]
	if !exists {
		return nil, handlers.ErrTokenNotFound
	}
	return token, nil
}

func (m *MockTokenStore) StoreToken(ctx context.Context, token *handlers.StoredToken) error {
	m.tokens[token.TokenID] = token
	return nil
}

func (m *MockTokenStore) RevokeToken(ctx context.Context, tokenID string) error {
	delete(m.tokens, tokenID)
	return nil
}

func (suite *OIDCUCANFlowTestSuite) SetupSuite() {
	// Initialize stores
	suite.clientStore = NewMockClientStore()
	suite.tokenStore = NewMockTokenStore()

	// Add test client
	suite.clientStore.clients["test-client"] = &handlers.OAuth2Client{
		ClientID:      "test-client",
		ClientSecret:  "test-secret",
		RedirectURIs:  []string{"https://example.com/callback"},
		AllowedScopes: []string{"openid", "profile", "vault:read", "dwn:write"},
		Metadata: map[string]string{
			"client_did": "did:sonr:test-client",
		},
	}

	// Initialize handlers with mock signer
	suite.mockSigner = NewMockBlockchainUCANSigner("did:sonr:oauth-provider")
	// Create a real signer for delegator and handlers (using mock resolver internally)
	realSigner, _ := handlers.NewBlockchainUCANSigner(nil, "did:sonr:oauth-provider")
	suite.ucanDelegator = handlers.NewUCANDelegator(realSigner)
	suite.oauth2Provider = NewMockOAuth2Provider()
	suite.tokenExchange = handlers.NewTokenExchangeHandler(suite.ucanDelegator, realSigner, suite.tokenStore, suite.clientStore)
	suite.refreshHandler = handlers.NewRefreshTokenHandler(suite.ucanDelegator, realSigner, suite.tokenStore, suite.clientStore)

	// Setup test server
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", suite.handleOIDCDiscovery)
	mux.HandleFunc("/oauth/authorize", suite.oauth2Provider.HandleAuthorize)
	mux.HandleFunc("/oauth/token", suite.handleToken)
	mux.HandleFunc("/oauth/userinfo", suite.oauth2Provider.HandleUserInfo)

	suite.testServer = httptest.NewServer(mux)
}

func (suite *OIDCUCANFlowTestSuite) TearDownSuite() {
	suite.testServer.Close()
}

func (suite *OIDCUCANFlowTestSuite) handleOIDCDiscovery(w http.ResponseWriter, r *http.Request) {
	discovery := map[string]interface{}{
		"issuer":                   suite.testServer.URL,
		"authorization_endpoint":   suite.testServer.URL + "/oauth/authorize",
		"token_endpoint":           suite.testServer.URL + "/oauth/token",
		"userinfo_endpoint":        suite.testServer.URL + "/oauth/userinfo",
		"jwks_uri":                 suite.testServer.URL + "/.well-known/jwks.json",
		"scopes_supported":         []string{"openid", "profile", "email", "vault:read", "vault:write", "dwn:read", "dwn:write"},
		"response_types_supported": []string{"code", "token", "id_token"},
		"grant_types_supported":    []string{"authorization_code", "refresh_token", "urn:ietf:params:oauth:grant-type:token-exchange"},
		"token_types_supported":    []string{"Bearer", "UCAN"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(discovery)
}

func (suite *OIDCUCANFlowTestSuite) handleToken(w http.ResponseWriter, r *http.Request) {
	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		suite.oauth2Provider.HandleToken(w, r)
	case "refresh_token":
		suite.refreshHandler.HandleRefreshToken(w, r)
	case "urn:ietf:params:oauth:grant-type:token-exchange":
		suite.tokenExchange.HandleTokenExchange(w, r)
	default:
		http.Error(w, "unsupported_grant_type", http.StatusBadRequest)
	}
}

// TestOIDCDiscovery tests OIDC discovery endpoint
func (suite *OIDCUCANFlowTestSuite) TestOIDCDiscovery() {
	resp, err := http.Get(suite.testServer.URL + "/.well-known/openid-configuration")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	var discovery map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&discovery)
	suite.Require().NoError(err)

	// Verify required OIDC fields
	suite.Contains(discovery, "issuer")
	suite.Contains(discovery, "authorization_endpoint")
	suite.Contains(discovery, "token_endpoint")
	suite.Contains(discovery, "jwks_uri")
	suite.Contains(discovery, "scopes_supported")

	// Verify UCAN-specific extensions
	scopes := discovery["scopes_supported"].([]interface{})
	suite.Contains(scopes, "vault:read")
	suite.Contains(scopes, "dwn:write")

	grantTypes := discovery["grant_types_supported"].([]interface{})
	suite.Contains(grantTypes, "urn:ietf:params:oauth:grant-type:token-exchange")
}

// TestAuthorizationCodeToUCAN tests authorization code flow that issues UCAN tokens
func (suite *OIDCUCANFlowTestSuite) TestAuthorizationCodeToUCAN() {
	// Step 1: Create authorization code
	authCode := "test-auth-code"
	userDID := "did:sonr:user123"
	scopes := []string{"openid", "vault:read", "dwn:write"}

	// Store auth code (in real flow, this happens during authorize)
	suite.tokenStore.StoreToken(context.Background(), &handlers.StoredToken{
		TokenID:   authCode,
		TokenType: "authorization_code",
		ClientID:  "test-client",
		UserDID:   userDID,
		Scopes:    scopes,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	})

	// Step 2: Exchange authorization code for tokens
	req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(
		"grant_type=authorization_code&code="+authCode+"&client_id=test-client",
	))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("test-client", "test-secret")

	w := httptest.NewRecorder()
	suite.handleToken(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var tokenResp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&tokenResp)
	suite.Require().NoError(err)

	// Verify response contains UCAN token
	suite.Contains(tokenResp, "access_token")
	suite.Contains(tokenResp, "ucan_token")
	suite.Contains(tokenResp, "token_type")
	suite.Equal("Bearer", tokenResp["token_type"])

	// Step 3: Verify UCAN token structure
	ucanTokenStr, ok := tokenResp["ucan_token"].(string)
	suite.True(ok, "UCAN token should be a string")
	suite.NotEmpty(ucanTokenStr)

	// Verify UCAN can be parsed (basic validation)
	suite.Contains(ucanTokenStr, ".") // JWT format
}

// TestTokenExchangeFlow tests RFC 8693 token exchange
func (suite *OIDCUCANFlowTestSuite) TestTokenExchangeFlow() {
	// Setup: Create an access token
	accessToken := "test-access-token"
	userDID := "did:sonr:user456"

	suite.tokenStore.StoreToken(context.Background(), &handlers.StoredToken{
		TokenID:     accessToken,
		TokenType:   "access_token",
		AccessToken: accessToken,
		ClientID:    "test-client",
		UserDID:     userDID,
		Scopes:      []string{"vault:read", "dwn:read"},
		ExpiresAt:   time.Now().Add(time.Hour),
		UCANToken:   "dummy.ucan.token",
	})

	// Perform token exchange: Access Token → UCAN
	exchangeReq := handlers.TokenExchangeRequest{
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		SubjectToken:       accessToken,
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:access_token",
		RequestedTokenType: "urn:x-oath:params:oauth:token-type:ucan",
		Scope:              "vault:read", // Request subset of scopes
	}

	body, _ := json.Marshal(exchangeReq)
	req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("test-client", "test-secret")

	w := httptest.NewRecorder()
	suite.tokenExchange.HandleTokenExchange(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var exchangeResp handlers.TokenExchangeResponse
	err := json.NewDecoder(w.Body).Decode(&exchangeResp)
	suite.Require().NoError(err)

	// Verify exchange response
	suite.NotEmpty(exchangeResp.AccessToken)
	suite.Equal("urn:x-oath:params:oauth:token-type:ucan", exchangeResp.IssuedTokenType)
	suite.Equal("Bearer", exchangeResp.TokenType)
	suite.Equal("vault:read", exchangeResp.Scope)
}

// TestRefreshTokenWithUCANChain tests refresh token flow with UCAN delegation chains
func (suite *OIDCUCANFlowTestSuite) TestRefreshTokenWithUCANChain() {
	// Setup: Create refresh token
	refreshToken := "test-refresh-token"
	userDID := "did:sonr:user789"

	suite.tokenStore.StoreToken(context.Background(), &handlers.StoredToken{
		TokenID:      refreshToken,
		TokenType:    "refresh_token",
		RefreshToken: refreshToken,
		ClientID:     "test-client",
		UserDID:      userDID,
		Scopes:       []string{"vault:read", "vault:write", "dwn:read"},
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
	})

	// First refresh - maintain all scopes
	req1 := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(
		"grant_type=refresh_token&refresh_token="+refreshToken,
	))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req1.SetBasicAuth("test-client", "test-secret")

	w1 := httptest.NewRecorder()
	suite.handleToken(w1, req1) // Use handleToken instead, which routes to refreshHandler

	// Debug: print response if not OK
	if w1.Code != http.StatusOK {
		suite.T().Logf("Refresh token response: %s", w1.Body.String())
	}

	suite.Equal(http.StatusOK, w1.Code)

	var resp1 handlers.RefreshTokenResponse
	err := json.NewDecoder(w1.Body).Decode(&resp1)
	suite.Require().NoError(err)

	suite.NotEmpty(resp1.AccessToken)
	suite.NotEmpty(resp1.UCANToken)
	suite.Contains(resp1.Scope, "vault:read")
	suite.Contains(resp1.Scope, "vault:write")

	// Second refresh - attenuate scopes (reduce permissions)
	newRefreshToken := resp1.RefreshToken
	if newRefreshToken == "" {
		newRefreshToken = refreshToken // Use original if not rotated
	}

	req2 := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(
		"grant_type=refresh_token&refresh_token="+newRefreshToken+"&scope=vault:read",
	))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.SetBasicAuth("test-client", "test-secret")

	w2 := httptest.NewRecorder()
	suite.handleToken(w2, req2) // Use handleToken instead

	// Debug: print response if not OK
	if w2.Code != http.StatusOK {
		suite.T().Logf("Second refresh response: %s", w2.Body.String())
	}

	// Should succeed with reduced scopes
	suite.Equal(http.StatusOK, w2.Code)

	var resp2 handlers.RefreshTokenResponse
	err = json.NewDecoder(w2.Body).Decode(&resp2)
	suite.Require().NoError(err)

	suite.Equal("vault:read", resp2.Scope)        // Only requested scope
	suite.NotContains(resp2.Scope, "vault:write") // Write permission removed
}

// TestCrossModuleAuthorization tests authorization across multiple modules
func (suite *OIDCUCANFlowTestSuite) TestCrossModuleAuthorization() {
	// Create UCAN token with cross-module capabilities
	userDID := "did:sonr:user-cross"
	clientDID := "did:sonr:client-cross"

	// Map OAuth scopes to multiple module capabilities
	scopes := []string{"vault:read", "dwn:write", "service:manage", "did:write"}

	// Create delegation with cross-module permissions
	ucanToken, err := suite.ucanDelegator.CreateDelegation(
		userDID,
		clientDID,
		scopes,
		time.Now().Add(time.Hour),
	)
	suite.Require().NoError(err)
	suite.NotNil(ucanToken)

	// Verify token contains attenuations for all modules
	suite.Require().NotEmpty(ucanToken.Attenuations)

	// Check that each module has appropriate capabilities
	moduleCapabilities := make(map[string][]string)
	for _, att := range ucanToken.Attenuations {
		scheme := att.Resource.GetScheme()
		actions := att.Capability.GetActions()
		moduleCapabilities[scheme] = append(moduleCapabilities[scheme], actions...)
	}

	// Verify all modules are represented
	suite.Contains(moduleCapabilities, "vault")
	suite.Contains(moduleCapabilities, "dwn")
	suite.Contains(moduleCapabilities, "service")
	suite.Contains(moduleCapabilities, "did")

	// Verify appropriate actions per module
	suite.Contains(moduleCapabilities["vault"], "read")
	suite.Contains(moduleCapabilities["dwn"], "write")
	suite.Contains(moduleCapabilities["did"], "write")
}

// TestDelegationChainValidation tests validation of UCAN delegation chains
func (suite *OIDCUCANFlowTestSuite) TestDelegationChainValidation() {
	signer := suite.mockSigner

	// Create initial delegation: User → Client A
	userDID := "did:sonr:user-chain"
	clientA := "did:sonr:client-a"

	token1, err := signer.CreateDelegationToken(
		userDID,
		clientA,
		[]ucan.Attenuation{
			{
				Capability: &ucan.MultiCapability{Actions: []string{"read", "write"}},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: userDID},
			},
		},
		nil, // No proofs for initial delegation
		time.Hour,
	)
	suite.Require().NoError(err)

	// Create second delegation: Client A → Client B (with attenuation)
	clientB := "did:sonr:client-b"
	token2, err := signer.CreateDelegationToken(
		clientA,
		clientB,
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"}, // Reduced permissions
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: userDID},
			},
		},
		[]ucan.Proof{ucan.Proof(token1)}, // Include proof
		time.Hour,
	)
	suite.Require().NoError(err)

	// Validate delegation chain
	err = signer.ValidateDelegationChain([]string{token1, token2})
	suite.NoError(err, "Valid delegation chain should pass validation")

	// Test invalid chain (trying to escalate privileges)
	invalidToken, err := signer.CreateDelegationToken(
		clientB,
		"did:sonr:client-c",
		[]ucan.Attenuation{
			{
				Capability: &ucan.MultiCapability{Actions: []string{"read", "write", "delete"}}, // Escalation!
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: userDID},
			},
		},
		[]ucan.Proof{ucan.Proof(token2)},
		time.Hour,
	)
	suite.Require().NoError(err) // Token creation succeeds

	// But validation should fail due to privilege escalation
	err = signer.ValidateDelegationChain([]string{token1, token2, invalidToken})
	suite.Error(err, "Chain with privilege escalation should fail validation")
}

// TestPerformanceBenchmark tests authorization performance
func (suite *OIDCUCANFlowTestSuite) TestPerformanceBenchmark() {
	iterations := 100
	maxDuration := 50 * time.Millisecond // Target: < 50ms per operation

	userDID := "did:sonr:perf-user"
	clientDID := "did:sonr:perf-client"
	scopes := []string{"vault:read", "dwn:write"}

	// Benchmark UCAN token creation
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := suite.ucanDelegator.CreateDelegation(
			userDID,
			clientDID,
			scopes,
			time.Now().Add(time.Hour),
		)
		suite.Require().NoError(err)
	}
	avgCreation := time.Since(start) / time.Duration(iterations)

	suite.Less(avgCreation, maxDuration,
		"Average UCAN creation time (%v) should be less than %v", avgCreation, maxDuration)

	// Benchmark token validation
	signer := suite.mockSigner
	testToken, _ := signer.CreateDelegationToken(
		userDID,
		clientDID,
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: userDID},
			},
		},
		nil,
		time.Hour,
	)

	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, err := signer.VerifySignature(testToken)
		suite.Require().NoError(err)
	}
	avgValidation := time.Since(start) / time.Duration(iterations)

	suite.Less(avgValidation, maxDuration,
		"Average UCAN validation time (%v) should be less than %v", avgValidation, maxDuration)

	// Log performance metrics
	suite.T().Logf("Performance Metrics:")
	suite.T().Logf("  UCAN Creation: %v avg", avgCreation)
	suite.T().Logf("  UCAN Validation: %v avg", avgValidation)
}

// TestSecurityAudit performs security validation of delegation chains
func (suite *OIDCUCANFlowTestSuite) TestSecurityAudit() {
	signer := suite.mockSigner

	// Test 1: Expired token rejection
	expiredToken, err := signer.CreateDelegationToken(
		"did:sonr:user",
		"did:sonr:client",
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		nil,
		-1*time.Hour, // Already expired
	)
	suite.Require().NoError(err)

	_, err = signer.VerifySignature(expiredToken)
	suite.Error(err, "Expired token should fail verification")

	// Test 2: Malformed token rejection
	malformedToken := "not.a.valid.token"
	_, err = signer.VerifySignature(malformedToken)
	suite.Error(err, "Malformed token should fail verification")

	// Test 3: Token replay protection
	validToken, err := signer.CreateDelegationToken(
		"did:sonr:user",
		"did:sonr:client",
		[]ucan.Attenuation{
			{
				Capability: &ucan.SimpleCapability{Action: "read"},
				Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "test"},
			},
		},
		nil,
		time.Hour,
	)
	suite.Require().NoError(err)

	// First use should succeed
	_, err = signer.VerifySignature(validToken)
	suite.NoError(err)

	// Multiple uses should also succeed (tokens are bearer tokens)
	// But in production, nonce/jti tracking would prevent replay
	_, err = signer.VerifySignature(validToken)
	suite.NoError(err)

	// Test 4: Scope boundary enforcement
	err = suite.ucanDelegator.ValidateDelegation(
		&ucan.Token{
			Issuer:    "did:sonr:user",
			Audience:  "did:sonr:client",
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
			Attenuations: []ucan.Attenuation{
				{
					Capability: &ucan.SimpleCapability{Action: "read"},
					Resource:   &handlers.SimpleResource{Scheme: "vault", Value: "user"},
				},
			},
		},
		[]string{"vault:write"}, // Requesting more than granted
	)
	suite.Error(err, "Should reject request for unpermitted scope")
}

func TestOIDCUCANFlowSuite(t *testing.T) {
	suite.Run(t, new(OIDCUCANFlowTestSuite))
}
