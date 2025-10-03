package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestSIOPAuthorization tests Self-Issued OpenID Provider authorization
func TestSIOPAuthorization(t *testing.T) {
	e := echo.New()

	t.Run("ValidSIOPRequest", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/siop/authorize", nil)
		q := req.URL.Query()
		q.Set("response_type", "id_token")
		q.Set("client_id", "test-client")
		q.Set("redirect_uri", "http://localhost:3000/callback")
		q.Set("scope", "openid did")
		q.Set("nonce", "test-nonce")
		q.Set("state", "test-state")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Set authenticated user DID
		c.Set("user_did", "did:sonr:testuser")

		err := HandleSIOPAuthorization(c)
		assert.NoError(t, err)

		// Should redirect with ID token
		assert.Equal(t, http.StatusFound, rec.Code)
		location := rec.Header().Get("Location")
		assert.Contains(t, location, "id_token=")
		assert.Contains(t, location, "state=test-state")
	})

	t.Run("InvalidResponseType", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/siop/authorize", nil)
		q := req.URL.Query()
		q.Set("response_type", "code") // SIOP only supports id_token
		q.Set("client_id", "test-client")
		q.Set("redirect_uri", "http://localhost:3000/callback")
		q.Set("scope", "openid")
		q.Set("nonce", "test-nonce")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleSIOPAuthorization(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "unsupported_response_type", errorResp["error"])
	})

	t.Run("MissingScopeOpenID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/siop/authorize", nil)
		q := req.URL.Query()
		q.Set("response_type", "id_token")
		q.Set("client_id", "test-client")
		q.Set("redirect_uri", "http://localhost:3000/callback")
		q.Set("scope", "profile email") // Missing 'openid'
		q.Set("nonce", "test-nonce")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleSIOPAuthorization(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_scope", errorResp["error"])
	})

	t.Run("UnauthenticatedUser", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/siop/authorize", nil)
		q := req.URL.Query()
		q.Set("response_type", "id_token")
		q.Set("client_id", "test-client")
		q.Set("redirect_uri", "http://localhost:3000/callback")
		q.Set("scope", "openid")
		q.Set("nonce", "test-nonce")
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// No user_did set - user not authenticated

		err := HandleSIOPAuthorization(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var errorResp map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "authentication_required", errorResp["error"])
	})
}

// TestSIOPRegistration tests dynamic client registration for SIOP
func TestSIOPRegistration(t *testing.T) {
	e := echo.New()

	t.Run("ValidRegistration", func(t *testing.T) {
		registration := map[string]any{
			"client_name":    "Test SIOP Client",
			"client_purpose": "Testing SIOP functionality",
			"logo_uri":       "https://example.com/logo.png",
		}
		body, _ := json.Marshal(registration)

		req := httptest.NewRequest(http.MethodPost, "/siop/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleSIOPRegistration(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]any
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response["client_id"])
		assert.Equal(t, "Test SIOP Client", response["client_name"])
		assert.Contains(t, response["subject_syntax_types_supported"], "did")
	})

	t.Run("MissingClientName", func(t *testing.T) {
		registration := map[string]any{
			"client_purpose": "Testing",
		}
		body, _ := json.Marshal(registration)

		req := httptest.NewRequest(http.MethodPost, "/siop/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleSIOPRegistration(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_request", errorResp["error"])
		assert.Contains(t, errorResp["error_description"], "Client name is required")
	})
}

// TestSIOPMetadata tests SIOP metadata endpoint
func TestSIOPMetadata(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/siop/metadata", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := HandleSIOPMetadata(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var metadata map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &metadata)
	assert.NoError(t, err)

	// Verify required SIOP metadata fields
	assert.Equal(t, "https://self-issued.me/v2", metadata["issuer"])
	assert.Equal(t, "openid://", metadata["authorization_endpoint"])
	assert.Contains(t, metadata["response_types_supported"], "id_token")
	assert.Contains(t, metadata["scopes_supported"], "openid")
	assert.Contains(t, metadata["subject_types_supported"], "pairwise")
	assert.NotNil(t, metadata["vp_formats_supported"])
}

// TestValidateSIOPRequest tests SIOP request validation
func TestValidateSIOPRequest(t *testing.T) {
	t.Run("ValidRequest", func(t *testing.T) {
		requestURI := "openid://?" +
			"response_type=id_token" +
			"&client_id=test-client" +
			"&redirect_uri=" + url.QueryEscape("http://localhost:3000/callback") +
			"&scope=openid+did" +
			"&nonce=test-nonce" +
			"&state=test-state"

		req, err := ValidateSIOPRequest(requestURI)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.Equal(t, "id_token", req.ResponseType)
		assert.Equal(t, "test-client", req.ClientID)
		assert.Equal(t, "openid did", req.Scope)
		assert.Equal(t, "test-nonce", req.Nonce)
	})

	t.Run("InvalidURI", func(t *testing.T) {
		req, err := ValidateSIOPRequest("not-a-valid-uri")
		assert.Error(t, err)
		assert.Nil(t, req)
	})

	t.Run("WithClaims", func(t *testing.T) {
		claims := map[string]any{
			"id_token": map[string]any{
				"email": map[string]any{
					"essential": true,
				},
				"name": nil,
			},
		}
		claimsJSON, _ := json.Marshal(claims)

		requestURI := "openid://?" +
			"response_type=id_token" +
			"&client_id=test-client" +
			"&redirect_uri=" + url.QueryEscape("http://localhost:3000/callback") +
			"&scope=openid" +
			"&nonce=test-nonce" +
			"&claims=" + url.QueryEscape(string(claimsJSON))

		req, err := ValidateSIOPRequest(requestURI)
		assert.NoError(t, err)
		assert.NotNil(t, req)
		assert.NotNil(t, req.Claims.IDToken)
	})
}

// TestSIOPWithVPToken tests SIOP with Verifiable Presentation
func TestSIOPWithVPToken(t *testing.T) {
	e := echo.New()

	vpClaims := map[string]any{
		"vp_token": map[string]any{
			"presentation_definition": map[string]any{
				"id": "test-presentation",
				"input_descriptors": []map[string]any{
					{
						"id": "id_credential",
						"constraints": map[string]any{
							"fields": []map[string]any{
								{
									"path": []string{"$.type"},
									"filter": map[string]any{
										"type":  "string",
										"const": "IdentityCredential",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	claimsJSON, _ := json.Marshal(vpClaims)

	req := httptest.NewRequest(http.MethodPost, "/siop/authorize", nil)
	q := req.URL.Query()
	q.Set("response_type", "id_token")
	q.Set("client_id", "test-client")
	q.Set("redirect_uri", "http://localhost:3000/callback")
	q.Set("scope", "openid")
	q.Set("nonce", "test-nonce")
	q.Set("claims", string(claimsJSON))
	req.URL.RawQuery = q.Encode()

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_did", "did:sonr:testuser")

	// Parse request to set claims
	siopReq := &SIOPRequest{
		ResponseType: "id_token",
		ClientID:     "test-client",
		RedirectURI:  "http://localhost:3000/callback",
		Scope:        "openid",
		Nonce:        "test-nonce",
	}
	err := json.Unmarshal(claimsJSON, &siopReq.Claims)
	assert.NoError(t, err)
	c.Set("siop_request", siopReq)

	err = HandleSIOPAuthorization(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)

	location := rec.Header().Get("Location")
	assert.Contains(t, location, "id_token=")
	// VP token generation is optional, so we don't assert its presence
}

// TestHandleSIOPRequest tests complete SIOP request flow
func TestHandleSIOPRequest(t *testing.T) {
	e := echo.New()

	t.Run("ValidRequestURI", func(t *testing.T) {
		requestURI := "openid://?" +
			"response_type=id_token" +
			"&client_id=test-client" +
			"&redirect_uri=" + url.QueryEscape("http://localhost:3000/callback") +
			"&scope=openid" +
			"&nonce=test-nonce"

		req := httptest.NewRequest(http.MethodGet, "/siop/request", nil)
		q := req.URL.Query()
		q.Set("request_uri", requestURI)
		req.URL.RawQuery = q.Encode()

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user_did", "did:sonr:testuser")

		err := HandleSIOPRequest(c)
		assert.NoError(t, err)
		// Should process as authorization request
		assert.Equal(t, http.StatusFound, rec.Code)
	})

	t.Run("MissingRequestURI", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/siop/request", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := HandleSIOPRequest(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResp map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_request", errorResp["error"])
		assert.Contains(t, errorResp["error_description"], "Missing request_uri")
	})
}
