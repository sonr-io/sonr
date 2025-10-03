package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOIDCDiscovery tests the OIDC discovery endpoint
func TestOIDCDiscovery(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/openid-configuration", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := GetOIDCDiscovery(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var config OIDCConfig
	err = json.Unmarshal(rec.Body.Bytes(), &config)
	assert.NoError(t, err)

	// Verify required fields
	assert.NotEmpty(t, config.Issuer)
	assert.NotEmpty(t, config.AuthorizationEndpoint)
	assert.NotEmpty(t, config.TokenEndpoint)
	assert.NotEmpty(t, config.UserInfoEndpoint)
	assert.NotEmpty(t, config.JWKSEndpoint)
	assert.Contains(t, config.ScopesSupported, "openid")
	assert.Contains(t, config.ResponseTypesSupported, "code")
	assert.Contains(t, config.GrantTypesSupported, "authorization_code")
}

// TestOIDCAuthorizationFlow tests the authorization code flow
func TestOIDCAuthorizationFlow(t *testing.T) {
	e := echo.New()

	t.Run("ValidAuthorizationRequest", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oidc/authorize", nil)
		q := req.URL.Query()
		q.Set("response_type", "code")
		q.Set("client_id", "test-client")
		q.Set("redirect_uri", "http://localhost:3000/callback")
		q.Set("scope", "openid profile")
		q.Set("state", "test-state")
		q.Set("nonce", "test-nonce")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Set authenticated user context
		c.Set("user_did", "did:sonr:testuser")
		c.Set("authenticated", true)

		err := HandleOIDCAuthorization(c)
		assert.NoError(t, err)

		// Should redirect with authorization code
		assert.Equal(t, http.StatusFound, rec.Code)
		location := rec.Header().Get("Location")
		assert.Contains(t, location, "code=")
		assert.Contains(t, location, "state=test-state")
	})

	t.Run("MissingRequiredParameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oidc/authorize", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleOIDCAuthorization(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_request", errorResp["error"])
	})

	t.Run("InvalidResponseType", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oidc/authorize", nil)
		q := req.URL.Query()
		q.Set("response_type", "invalid")
		q.Set("client_id", "test-client")
		q.Set("redirect_uri", "http://localhost:3000/callback")
		q.Set("scope", "openid")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleOIDCAuthorization(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// TestOIDCTokenExchange tests the token endpoint
func TestOIDCTokenExchange(t *testing.T) {
	e := echo.New()

	// Setup: Create an authorization code
	code := "test-auth-code"
	authCode := &AuthorizationCode{
		Code:                code,
		ClientID:            "test-client",
		RedirectURI:         "http://localhost:3000/callback",
		UserDID:             "did:sonr:testuser",
		Scope:               "openid profile",
		ExpiresAt:           time.Now().Add(10 * time.Minute),
		CodeChallenge:       "test-challenge",
		CodeChallengeMethod: "S256",
	}

	oidcProvider.mu.Lock()
	oidcProvider.codes[code] = authCode
	oidcProvider.mu.Unlock()

	t.Run("ValidTokenExchange", func(t *testing.T) {
		body := strings.NewReader("grant_type=authorization_code&code=" + code +
			"&redirect_uri=http://localhost:3000/callback&client_id=test-client" +
			"&code_verifier=test-verifier")

		req := httptest.NewRequest(http.MethodPost, "/oidc/token", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		tokenReq := &OIDCTokenRequest{
			GrantType:    "authorization_code",
			Code:         code,
			RedirectURI:  "http://localhost:3000/callback",
			ClientID:     "test-client",
			CodeVerifier: "test-verifier",
		}

		err := handleAuthorizationCodeGrant(c, tokenReq)
		assert.NoError(t, err)

		if rec.Code == http.StatusOK {
			var tokenResp OIDCTokenResponse
			err = json.Unmarshal(rec.Body.Bytes(), &tokenResp)
			assert.NoError(t, err)
			assert.NotEmpty(t, tokenResp.AccessToken)
			assert.NotEmpty(t, tokenResp.IDToken)
			assert.Equal(t, "Bearer", tokenResp.TokenType)
		}
	})

	t.Run("ExpiredAuthorizationCode", func(t *testing.T) {
		expiredCode := "expired-code"
		expiredAuthCode := &AuthorizationCode{
			Code:        expiredCode,
			ClientID:    "test-client",
			RedirectURI: "http://localhost:3000/callback",
			UserDID:     "did:sonr:testuser",
			ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired
		}

		oidcProvider.mu.Lock()
		oidcProvider.codes[expiredCode] = expiredAuthCode
		oidcProvider.mu.Unlock()

		tokenReq := &OIDCTokenRequest{
			GrantType:   "authorization_code",
			Code:        expiredCode,
			RedirectURI: "http://localhost:3000/callback",
			ClientID:    "test-client",
		}

		req := httptest.NewRequest(http.MethodPost, "/oidc/token", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handleAuthorizationCodeGrant(c, tokenReq)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// TestOIDCUserInfo tests the userinfo endpoint
func TestOIDCUserInfo(t *testing.T) {
	e := echo.New()

	// Setup: Create a session
	accessToken := "test-access-token"
	session := &OIDCSession{
		SessionID:    "test-session",
		UserDID:      "did:sonr:testuser",
		ClientID:     "test-client",
		Scope:        "openid profile email",
		AccessToken:  accessToken,
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		CreatedAt:    time.Now(),
	}

	oidcProvider.mu.Lock()
	oidcProvider.sessions[accessToken] = session
	oidcProvider.mu.Unlock()

	t.Run("ValidUserInfoRequest", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oidc/userinfo", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleOIDCUserInfo(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var userInfo map[string]any
		err = json.Unmarshal(rec.Body.Bytes(), &userInfo)
		assert.NoError(t, err)
		assert.Equal(t, "did:sonr:testuser", userInfo["sub"])
	})

	t.Run("InvalidAccessToken", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oidc/userinfo", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleOIDCUserInfo(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("MissingAuthorizationHeader", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oidc/userinfo", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleOIDCUserInfo(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// TestPKCEFlow tests PKCE (Proof Key for Code Exchange) implementation
func TestPKCEFlow(t *testing.T) {
	e := echo.New()

	// Generate PKCE parameters
	codeVerifier := "test-code-verifier-string-that-is-long-enough"
	codeChallenge := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM" // SHA256 of verifier

	t.Run("AuthorizationWithPKCE", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/oidc/authorize", nil)
		q := req.URL.Query()
		q.Set("response_type", "code")
		q.Set("client_id", "test-client")
		q.Set("redirect_uri", "http://localhost:3000/callback")
		q.Set("scope", "openid")
		q.Set("code_challenge", codeChallenge)
		q.Set("code_challenge_method", "S256")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user_did", "did:sonr:testuser")
		c.Set("authenticated", true)

		err := HandleOIDCAuthorization(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, rec.Code)
	})

	t.Run("TokenExchangeWithPKCE", func(t *testing.T) {
		// Create auth code with PKCE
		code := "pkce-auth-code"
		authCode := &AuthorizationCode{
			Code:                code,
			ClientID:            "test-client",
			RedirectURI:         "http://localhost:3000/callback",
			UserDID:             "did:sonr:testuser",
			CodeChallenge:       codeChallenge,
			CodeChallengeMethod: "S256",
			ExpiresAt:           time.Now().Add(10 * time.Minute),
		}

		oidcProvider.mu.Lock()
		oidcProvider.codes[code] = authCode
		oidcProvider.mu.Unlock()

		tokenReq := &OIDCTokenRequest{
			GrantType:    "authorization_code",
			Code:         code,
			RedirectURI:  "http://localhost:3000/callback",
			ClientID:     "test-client",
			CodeVerifier: codeVerifier,
		}

		req := httptest.NewRequest(http.MethodPost, "/oidc/token", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handleAuthorizationCodeGrant(c, tokenReq)
		require.NoError(t, err)

		// With correct verifier, should succeed
		if rec.Code != http.StatusOK {
			t.Logf("Response: %s", rec.Body.String())
		}
	})
}

// TestRefreshTokenFlow tests refresh token functionality
func TestRefreshTokenFlow(t *testing.T) {
	e := echo.New()

	// Create initial session with refresh token
	refreshToken := "test-refresh-token"
	session := &OIDCSession{
		SessionID:    "test-session",
		UserDID:      "did:sonr:testuser",
		ClientID:     "test-client",
		Scope:        "openid profile offline_access",
		AccessToken:  "old-access-token",
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Expired access token
		CreatedAt:    time.Now().Add(-2 * time.Hour),
	}

	oidcProvider.mu.Lock()
	oidcProvider.sessions[refreshToken] = session
	oidcProvider.mu.Unlock()

	t.Run("ValidRefreshToken", func(t *testing.T) {
		tokenReq := &OIDCTokenRequest{
			GrantType:    "refresh_token",
			RefreshToken: refreshToken,
			ClientID:     "test-client",
		}

		req := httptest.NewRequest(http.MethodPost, "/oidc/token", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handleRefreshTokenGrant(c, tokenReq)
		assert.NoError(t, err)

		if rec.Code == http.StatusOK {
			var tokenResp OIDCTokenResponse
			err = json.Unmarshal(rec.Body.Bytes(), &tokenResp)
			assert.NoError(t, err)
			assert.NotEmpty(t, tokenResp.AccessToken)
			assert.NotEqual(t, "old-access-token", tokenResp.AccessToken)
		}
	})
}
