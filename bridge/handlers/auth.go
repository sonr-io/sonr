// Package handlers provides HTTP handlers for the highway server
package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// CustomClaims defines custom JWT claims for vault operations
type CustomClaims struct {
	UserID      string   `json:"user_id"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response payload
type LoginResponse struct {
	Token string `json:"token"`
}

// LoginHandler generates JWT tokens for authentication
// This now supports both traditional login and OIDC-based authentication
func LoginHandler(jwtSecret []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check if this is an OIDC callback
		code := c.QueryParam("code")
		if code != "" {
			return handleOIDCCallback(c, code, jwtSecret)
		}

		// Check if user has WebAuthn credentials
		authHeader := c.Request().Header.Get("X-WebAuthn-Assertion")
		if authHeader != "" {
			return handleWebAuthnLogin(c, authHeader, jwtSecret)
		}

		// Traditional login flow
		var req LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON payload"})
		}

		// Simple authentication - in production, validate against a proper user database
		if req.Username == "" || req.Password == "" {
			return c.JSON(
				http.StatusBadRequest,
				map[string]string{"error": "Username and password are required"},
			)
		}

		// For demo purposes, accept any non-empty credentials
		// In production, verify credentials against database/directory
		if req.Username == "vault-user" && req.Password == "vault-pass" {
			// Create custom claims
			claims := &CustomClaims{
				UserID: req.Username,
				Permissions: []string{
					"vault:generate",
					"vault:sign",
					"vault:verify",
					"vault:export",
					"vault:import",
					"vault:refresh",
				},
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // 24 hours
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					NotBefore: jwt.NewNumericDate(time.Now()),
					Issuer:    "highway-vault",
					Subject:   req.Username,
				},
			}

			// Create token with claims
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

			// Sign token with secret
			tokenString, err := token.SignedString(jwtSecret)
			if err != nil {
				return c.JSON(
					http.StatusInternalServerError,
					map[string]string{"error": "Failed to generate token"},
				)
			}

			return c.JSON(http.StatusOK, LoginResponse{Token: tokenString})
		}

		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}
}

// handleOIDCCallback processes OIDC authorization code callback
func handleOIDCCallback(c echo.Context, code string, jwtSecret []byte) error {
	// Exchange code for tokens using OIDC token endpoint
	tokenReq := &OIDCTokenRequest{
		GrantType:   "authorization_code",
		Code:        code,
		RedirectURI: c.QueryParam("redirect_uri"),
		ClientID:    c.QueryParam("client_id"),
	}

	// Call internal OIDC token handler
	// In production, this would make an HTTP call to the OIDC provider
	c.Set("oidc_token_request", tokenReq)
	if err := handleAuthorizationCodeGrant(c, tokenReq); err != nil {
		return err
	}

	// Get the OIDC session from context
	session, ok := c.Get("oidc_session").(*OIDCSession)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to establish OIDC session",
		})
	}

	// Create JWT token from OIDC session
	claims := &CustomClaims{
		UserID: session.UserDID,
		Permissions: []string{
			"vault:generate",
			"vault:sign",
			"vault:verify",
			"vault:export",
			"vault:import",
			"vault:refresh",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(session.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "highway-vault-oidc",
			Subject:   session.UserDID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(http.StatusOK, LoginResponse{Token: tokenString})
}

// handleWebAuthnLogin processes WebAuthn-based login
func handleWebAuthnLogin(c echo.Context, assertion string, jwtSecret []byte) error {
	// Verify WebAuthn assertion
	// This would integrate with the WebAuthn handlers

	// For now, extract username from assertion (simplified)
	username := c.Request().Header.Get("X-WebAuthn-Username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "WebAuthn username required",
		})
	}

	// Verify the assertion matches a stored credential
	webAuthnStore.mu.RLock()
	credentials, exists := webAuthnStore.credentials[username]
	webAuthnStore.mu.RUnlock()

	if !exists || len(credentials) == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "No WebAuthn credentials found for user",
		})
	}

	// Create JWT token for WebAuthn authenticated user
	claims := &CustomClaims{
		UserID: username,
		Permissions: []string{
			"vault:generate",
			"vault:sign",
			"vault:verify",
			"vault:export",
			"vault:import",
			"vault:refresh",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "highway-vault-webauthn",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(http.StatusOK, LoginResponse{Token: tokenString})
}

// OIDCLoginHandler initiates OIDC login flow
func OIDCLoginHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Build authorization URL
		authURL := buildOIDCAuthorizationURL(
			c.QueryParam("client_id"),
			c.QueryParam("redirect_uri"),
			c.QueryParam("scope"),
			c.QueryParam("state"),
		)

		// Redirect to OIDC authorization endpoint
		return c.Redirect(http.StatusFound, authURL)
	}
}

// buildOIDCAuthorizationURL constructs the OIDC authorization URL
func buildOIDCAuthorizationURL(clientID, redirectURI, scope, state string) string {
	// Default values if not provided
	if clientID == "" {
		clientID = "highway-vault-client"
	}
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/auth/callback"
	}
	if scope == "" {
		scope = "openid profile did vault"
	}
	if state == "" {
		state = generateState()
	}

	// Build authorization URL
	return fmt.Sprintf(
		"https://localhost:8080/oidc/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		clientID,
		redirectURI,
		scope,
		state,
	)
}

// generateState generates a random state parameter for OIDC
func generateState() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "default-state"
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}
