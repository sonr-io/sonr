package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// SIOPProvider manages Self-Issued OpenID Provider operations
type SIOPProvider struct {
	// No persistent state needed for SIOP as it's self-issued
}

var siopProvider = &SIOPProvider{}

// HandleSIOPAuthorization handles SIOP authorization requests
func HandleSIOPAuthorization(c echo.Context) error {
	var req SIOPRequest

	// Check if request was already parsed by HandleSIOPRequest
	if parsedReq := c.Get("siop_request"); parsedReq != nil {
		req = *(parsedReq.(*SIOPRequest))
	} else {
		// Parse request parameters from query string for GET or form for POST
		req = SIOPRequest{
			ResponseType: c.QueryParam("response_type"),
			ClientID:     c.QueryParam("client_id"),
			RedirectURI:  c.QueryParam("redirect_uri"),
			Scope:        c.QueryParam("scope"),
			Nonce:        c.QueryParam("nonce"),
			State:        c.QueryParam("state"),
		}

		// Parse claims if provided
		if claimsStr := c.QueryParam("claims"); claimsStr != "" {
			_ = json.Unmarshal([]byte(claimsStr), &req.Claims) // Ignore claims parsing errors
		}

		// If no query params, try to bind from body (for POST requests)
		if req.ResponseType == "" {
			_ = c.Bind(&req) // Ignore bind errors and continue with empty request
		}
	}

	// Validate required parameters
	if req.ResponseType == "" || req.ClientID == "" || req.RedirectURI == "" ||
		req.Scope == "" || req.Nonce == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "Missing required parameters",
		})
	}

	// Validate response type (SIOP v2 supports id_token)
	if req.ResponseType != "id_token" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "unsupported_response_type",
			"error_description": "SIOP only supports id_token response type",
		})
	}

	// Validate scope includes openid
	if !strings.Contains(req.Scope, "openid") {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_scope",
			"error_description": "Scope must include 'openid'",
		})
	}

	// Get user DID from context (assumes user is authenticated via WebAuthn)
	userDID := c.Get("user_did")
	if userDID == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error":             "authentication_required",
			"error_description": "User authentication required for SIOP",
		})
	}

	// Generate Self-Issued ID Token
	idToken, err := generateSelfIssuedIDToken(
		userDID.(string),
		req.ClientID,
		req.Nonce,
		req.RedirectURI,
		req.Claims,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to generate self-issued ID token",
		})
	}

	// Build response
	response := &SIOPResponse{
		IDToken: idToken,
		State:   req.State,
	}

	// If VP token is requested, include it
	if req.Claims.VPToken != nil {
		vpToken, err := generateVPToken(userDID.(string), req.Claims.VPToken)
		if err == nil {
			response.VPToken = vpToken
		}
	}

	// Redirect back to client with response
	redirectURL, err := buildSIOPRedirectURL(req.RedirectURI, response)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":             "server_error",
			"error_description": "Failed to build redirect URL",
		})
	}

	return c.Redirect(http.StatusFound, redirectURL)
}

// HandleSIOPRegistration handles dynamic client registration for SIOP
func HandleSIOPRegistration(c echo.Context) error {
	var registration SIOPRegistration
	if err := c.Bind(&registration); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "Invalid registration request",
		})
	}

	// Validate registration parameters
	if registration.ClientName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "Client name is required",
		})
	}

	// Generate client ID for the registration
	clientID := generateClientID()

	// Store registration (in production, persist this)
	// For now, return the registration confirmation
	response := map[string]any{
		"client_id":                      clientID,
		"client_name":                    registration.ClientName,
		"client_purpose":                 registration.ClientPurpose,
		"logo_uri":                       registration.LogoURI,
		"subject_syntax_types_supported": []string{"did"},
		"vp_formats": map[string]any{
			"jwt_vp": map[string]any{
				"alg": []string{"ES256", "EdDSA"},
			},
			"ldp_vp": map[string]any{
				"proof_type": []string{"Ed25519Signature2018"},
			},
		},
	}

	return c.JSON(http.StatusOK, response)
}

// HandleSIOPMetadata returns SIOP provider metadata
func HandleSIOPMetadata(c echo.Context) error {
	metadata := map[string]any{
		"issuer":                   "https://self-issued.me/v2",
		"authorization_endpoint":   "openid://",
		"response_types_supported": []string{"id_token"},
		"scopes_supported": []string{
			"openid",
			"profile",
			"email",
			"did",
		},
		"subject_types_supported":                     []string{"pairwise"},
		"id_token_signing_alg_values_supported":       []string{"ES256", "EdDSA"},
		"request_object_signing_alg_values_supported": []string{"ES256", "EdDSA"},
		"subject_syntax_types_supported": []string{
			"did:key",
			"did:web",
			"did:ion",
		},
		"vp_formats_supported": map[string]any{
			"jwt_vp": map[string]any{
				"alg_values_supported": []string{"ES256", "EdDSA"},
			},
			"ldp_vp": map[string]any{
				"proof_type_values_supported": []string{"Ed25519Signature2018"},
			},
		},
	}

	return c.JSON(http.StatusOK, metadata)
}

// generateSelfIssuedIDToken generates a self-issued ID token
func generateSelfIssuedIDToken(
	userDID, clientID, nonce, redirectURI string,
	requestedClaims SIOPClaims,
) (string, error) {
	// Create base claims
	claims := jwt.MapClaims{
		"iss":     "https://self-issued.me/v2",
		"sub":     userDID,
		"aud":     clientID,
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"nonce":   nonce,
		"_sd_alg": "sha-256", // Selective disclosure algorithm
		"sub_jwk": map[string]any{
			// TODO: Include actual public key from DID document
			"kty": "EC",
			"crv": "P-256",
			"x":   "placeholder_x",
			"y":   "placeholder_y",
		},
	}

	// Add requested claims from ID token
	if requestedClaims.IDToken != nil {
		for claimName := range requestedClaims.IDToken {
			// Add claim based on request
			switch claimName {
			case "email":
				claims["email"] = "user@example.com" // TODO: Get from DID document
				claims["email_verified"] = true
			case "name":
				claims["name"] = "User Name" // TODO: Get from DID document
			case "did":
				claims["did"] = userDID
			}
		}
	}

	// Sign with user's DID key (simplified for now)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("temporary-siop-key"))
}

// generateVPToken generates a Verifiable Presentation token
func generateVPToken(userDID string, vpClaims map[string]ClaimRequest) (string, error) {
	// Create VP structure
	vp := map[string]any{
		"@context": []string{
			"https://www.w3.org/2018/credentials/v1",
		},
		"type":                 []string{"VerifiablePresentation"},
		"holder":               userDID,
		"verifiableCredential": []any{
			// TODO: Include actual verifiable credentials
		},
	}

	// Convert to JWT
	vpJSON, err := json.Marshal(vp)
	if err != nil {
		return "", err
	}

	// Sign VP (simplified for now)
	claims := jwt.MapClaims{
		"vp":    string(vpJSON),
		"iss":   userDID,
		"aud":   "verifier",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"nonce": generateNonce(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("temporary-vp-key"))
}

// buildSIOPRedirectURL builds the redirect URL with SIOP response
func buildSIOPRedirectURL(redirectURI string, response *SIOPResponse) (string, error) {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return "", err
	}

	// Add response parameters
	q := u.Query()
	q.Set("id_token", response.IDToken)
	if response.VPToken != "" {
		q.Set("vp_token", response.VPToken)
	}
	if response.State != "" {
		q.Set("state", response.State)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// generateClientID generates a unique client ID
func generateClientID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return fmt.Sprintf("siop_client_%s", base64.RawURLEncoding.EncodeToString(bytes))
}

// generateNonce generates a nonce for tokens
func generateNonce() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

// ValidateSIOPRequest validates an incoming SIOP request
func ValidateSIOPRequest(requestURI string) (*SIOPRequest, error) {
	// Parse the request URI
	u, err := url.Parse(requestURI)
	if err != nil {
		return nil, fmt.Errorf("invalid request URI: %w", err)
	}

	// Validate that it's a proper URI with scheme and host
	if u.Scheme == "" && u.Host == "" {
		return nil, fmt.Errorf("invalid request URI: missing scheme or host")
	}

	// Extract parameters
	q := u.Query()
	req := &SIOPRequest{
		ResponseType: q.Get("response_type"),
		ClientID:     q.Get("client_id"),
		RedirectURI:  q.Get("redirect_uri"),
		Scope:        q.Get("scope"),
		Nonce:        q.Get("nonce"),
		State:        q.Get("state"),
	}

	// Parse claims if present
	if claimsStr := q.Get("claims"); claimsStr != "" {
		if err := json.Unmarshal([]byte(claimsStr), &req.Claims); err != nil {
			return nil, fmt.Errorf("invalid claims parameter: %w", err)
		}
	}

	// Parse registration if present
	if regStr := q.Get("registration"); regStr != "" {
		if err := json.Unmarshal([]byte(regStr), &req.Registration); err != nil {
			return nil, fmt.Errorf("invalid registration parameter: %w", err)
		}
	}

	return req, nil
}

// HandleSIOPRequest processes a complete SIOP request flow
func HandleSIOPRequest(c echo.Context) error {
	// Get request URI from query parameter
	requestURI := c.QueryParam("request_uri")
	if requestURI == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": "Missing request_uri parameter",
		})
	}

	// Validate and parse the request
	req, err := ValidateSIOPRequest(requestURI)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             "invalid_request",
			"error_description": err.Error(),
		})
	}

	// Process as regular SIOP authorization
	c.Set("siop_request", req)
	return HandleSIOPAuthorization(c)
}
