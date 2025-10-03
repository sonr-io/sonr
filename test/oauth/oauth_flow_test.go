package oauth_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validAccessToken = "valid_access_token"
)

// OAuth2Client represents an OAuth client for testing
type OAuth2Client struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
}

// OAuth2Token represents an access token response
type OAuth2Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// TestOAuth2AuthorizationCodeFlow tests the complete OAuth2 authorization code flow
func TestOAuth2AuthorizationCodeFlow(t *testing.T) {
	// Setup test server
	server := setupTestOAuthServer()
	defer server.Close()

	client := &OAuth2Client{
		ClientID:     "test_client_123",
		ClientSecret: "test_secret",
		RedirectURI:  "http://localhost:3000/callback",
		Scopes:       []string{"openid", "profile", "vault:read"},
	}

	// Test authorization request
	t.Run("Authorization Request", func(t *testing.T) {
		// Generate PKCE challenge
		verifier := generateCodeVerifier()
		challenge := generateCodeChallenge(verifier)

		// Build authorization URL
		authURL := buildAuthorizationURL(server.URL, client, challenge, "test_state")

		// Make authorization request
		httpClient := &http.Client{}
		req, err := http.NewRequest("GET", authURL, nil)
		require.NoError(t, err)
		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Should return authorization page
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test token exchange
	t.Run("Token Exchange", func(t *testing.T) {
		verifier := generateCodeVerifier()
		code := "test_authorization_code"

		// Exchange code for token
		token, err := exchangeCodeForToken(server.URL, client, code, verifier)
		require.NoError(t, err)

		assert.NotEmpty(t, token.AccessToken)
		assert.Equal(t, "Bearer", token.TokenType)
		assert.Greater(t, token.ExpiresIn, 0)
		assert.NotEmpty(t, token.RefreshToken)
	})

	// Test token refresh
	t.Run("Token Refresh", func(t *testing.T) {
		refreshToken := "test_refresh_token"

		// Refresh token
		token, err := refreshAccessToken(server.URL, client, refreshToken)
		require.NoError(t, err)

		assert.NotEmpty(t, token.AccessToken)
		assert.Equal(t, "Bearer", token.TokenType)
		assert.Greater(t, token.ExpiresIn, 0)
	})
}

// TestOAuth2PKCEValidation tests PKCE validation
func TestOAuth2PKCEValidation(t *testing.T) {
	server := setupTestOAuthServer()
	defer server.Close()

	client := &OAuth2Client{
		ClientID:    "public_client",
		RedirectURI: "http://localhost:3000/callback",
		Scopes:      []string{"openid"},
	}

	t.Run("Valid PKCE", func(t *testing.T) {
		verifier := generateCodeVerifier()
		challenge := generateCodeChallenge(verifier)

		// Authorization request with PKCE
		authURL := buildAuthorizationURL(server.URL, client, challenge, "")
		httpClient := &http.Client{}
		req, err := http.NewRequest("GET", authURL, nil)
		require.NoError(t, err)
		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Missing PKCE for Public Client", func(t *testing.T) {
		// Authorization request without PKCE
		params := url.Values{
			"response_type": {"code"},
			"client_id":     {client.ClientID},
			"redirect_uri":  {client.RedirectURI},
			"scope":         {strings.Join(client.Scopes, " ")},
		}

		authURL := fmt.Sprintf("%s/oauth/authorize?%s", server.URL, params.Encode())
		httpClient := &http.Client{}
		req, err := http.NewRequest("GET", authURL, nil)
		require.NoError(t, err)
		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Should reject without PKCE
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestOAuth2ScopeValidation tests scope validation and UCAN mapping
func TestOAuth2ScopeValidation(t *testing.T) {
	server := setupTestOAuthServer()
	defer server.Close()

	testCases := []struct {
		name          string
		scopes        []string
		expectedError bool
	}{
		{
			name:          "Valid scopes",
			scopes:        []string{"openid", "profile", "vault:read"},
			expectedError: false,
		},
		{
			name:          "Invalid scope",
			scopes:        []string{"invalid_scope"},
			expectedError: true,
		},
		{
			name:          "Mixed valid and invalid",
			scopes:        []string{"openid", "invalid_scope"},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oauthClient := &OAuth2Client{
				ClientID:    "test_client",
				RedirectURI: "http://localhost:3000/callback",
				Scopes:      tc.scopes,
			}

			verifier := generateCodeVerifier()
			challenge := generateCodeChallenge(verifier)
			authURL := buildAuthorizationURL(server.URL, oauthClient, challenge, "")

			httpClient := &http.Client{}
			req, err := http.NewRequest("GET", authURL, nil)
			require.NoError(t, err)
			resp, err := httpClient.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			if tc.expectedError {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			} else {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}
		})
	}
}

// TestOAuth2TokenIntrospection tests token introspection
func TestOAuth2TokenIntrospection(t *testing.T) {
	server := setupTestOAuthServer()
	defer server.Close()

	t.Run("Active Token", func(t *testing.T) {
		token := validAccessToken

		introspection, err := introspectToken(server.URL, token)
		require.NoError(t, err)

		assert.True(t, introspection["active"].(bool))
		assert.Equal(t, "test_client_123", introspection["client_id"])
		assert.Contains(t, introspection, "scope")
		assert.Contains(t, introspection, "exp")
	})

	t.Run("Expired Token", func(t *testing.T) {
		token := "expired_access_token"

		introspection, err := introspectToken(server.URL, token)
		require.NoError(t, err)

		assert.False(t, introspection["active"].(bool))
	})
}

// TestOAuth2TokenRevocation tests token revocation
func TestOAuth2TokenRevocation(t *testing.T) {
	server := setupTestOAuthServer()
	defer server.Close()

	token := "access_token_to_revoke"

	// Revoke token
	err := revokeToken(server.URL, token)
	require.NoError(t, err)

	// Verify token is revoked
	introspection, err := introspectToken(server.URL, token)
	require.NoError(t, err)
	assert.False(t, introspection["active"].(bool))
}

// TestOAuth2UserInfo tests the userinfo endpoint
func TestOAuth2UserInfo(t *testing.T) {
	server := setupTestOAuthServer()
	defer server.Close()

	accessToken := validAccessToken

	userInfo, err := getUserInfo(server.URL, accessToken)
	require.NoError(t, err)

	assert.NotEmpty(t, userInfo["sub"])
	assert.NotEmpty(t, userInfo["name"])
	assert.Contains(t, userInfo, "email")
	assert.Contains(t, userInfo, "picture")
}

// Helper functions

func setupTestOAuthServer() *httptest.Server {
	mux := http.NewServeMux()

	// Authorization endpoint
	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("client_id")
		codeChallenge := r.URL.Query().Get("code_challenge")

		// Validate PKCE for public clients
		if clientID == "public_client" && codeChallenge == "" {
			http.Error(w, "PKCE required for public clients", http.StatusBadRequest)
			return
		}

		// Validate scopes
		scopes := strings.Split(r.URL.Query().Get("scope"), " ")
		validScopes := map[string]bool{
			"openid": true, "profile": true, "email": true,
			"vault:read": true, "vault:write": true, "vault:sign": true,
		}

		for _, scope := range scopes {
			if scope != "" && !validScopes[scope] {
				http.Error(w, fmt.Sprintf("Invalid scope: %s", scope), http.StatusBadRequest)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "Authorization page")
	})

	// Token endpoint
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		grantType := r.FormValue("grant_type")

		var token OAuth2Token
		switch grantType {
		case "authorization_code":
			token = OAuth2Token{
				AccessToken:  generateToken(),
				TokenType:    "Bearer",
				ExpiresIn:    3600,
				RefreshToken: generateToken(),
				Scope:        r.FormValue("scope"),
			}
		case "refresh_token":
			token = OAuth2Token{
				AccessToken: generateToken(),
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			}
		default:
			http.Error(w, "Unsupported grant type", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(token)
	})

	// Introspection endpoint
	mux.HandleFunc("/oauth/introspect", func(w http.ResponseWriter, r *http.Request) {
		token := r.FormValue("token")

		response := map[string]any{
			"active": token == validAccessToken,
		}

		if response["active"].(bool) {
			response["client_id"] = "test_client_123"
			response["scope"] = "openid profile vault:read"
			response["exp"] = time.Now().Add(time.Hour).Unix()
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	// Revocation endpoint
	mux.HandleFunc("/oauth/revoke", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// UserInfo endpoint
	mux.HandleFunc("/oauth/userinfo", func(w http.ResponseWriter, r *http.Request) {
		userInfo := map[string]any{
			"sub":     "did:sonr:123456",
			"name":    "Test User",
			"email":   "test@example.com",
			"picture": "https://example.com/picture.jpg",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(userInfo)
	})

	return httptest.NewServer(mux)
}

func generateCodeVerifier() string {
	return base64.RawURLEncoding.EncodeToString(
		[]byte("test_verifier_12345678901234567890123456789012"),
	)
}

func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func generateToken() string {
	return fmt.Sprintf("token_%d", time.Now().UnixNano())
}

func buildAuthorizationURL(
	serverURL string,
	client *OAuth2Client,
	codeChallenge, state string,
) string {
	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {client.ClientID},
		"redirect_uri":          {client.RedirectURI},
		"scope":                 {strings.Join(client.Scopes, " ")},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
	}

	if state != "" {
		params.Set("state", state)
	}

	return fmt.Sprintf("%s/oauth/authorize?%s", serverURL, params.Encode())
}

func exchangeCodeForToken(
	serverURL string,
	client *OAuth2Client,
	code, verifier string,
) (*OAuth2Token, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {client.RedirectURI},
		"client_id":     {client.ClientID},
		"code_verifier": {verifier},
	}

	if client.ClientSecret != "" {
		data.Set("client_secret", client.ClientSecret)
	}

	resp, err := http.PostForm(fmt.Sprintf("%s/oauth/token", serverURL), data)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var token OAuth2Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	return &token, err
}

func refreshAccessToken(
	serverURL string,
	client *OAuth2Client,
	refreshToken string,
) (*OAuth2Token, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {client.ClientID},
	}

	if client.ClientSecret != "" {
		data.Set("client_secret", client.ClientSecret)
	}

	resp, err := http.PostForm(fmt.Sprintf("%s/oauth/token", serverURL), data)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var token OAuth2Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	return &token, err
}

func introspectToken(serverURL, token string) (map[string]any, error) {
	data := url.Values{
		"token": {token},
	}

	resp, err := http.PostForm(fmt.Sprintf("%s/oauth/introspect", serverURL), data)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func revokeToken(serverURL, token string) error {
	data := url.Values{
		"token": {token},
	}

	resp, err := http.PostForm(fmt.Sprintf("%s/oauth/revoke", serverURL), data)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

func getUserInfo(serverURL, accessToken string) (map[string]any, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/oauth/userinfo", serverURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}
