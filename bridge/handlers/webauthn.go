package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sonr-io/sonr/types/webauthn"
	"github.com/sonr-io/sonr/types/webauthn/webauthncbor"
)

// WebAuthnStore manages WebAuthn sessions and credentials
type WebAuthnStore struct {
	mu          sync.RWMutex
	sessions    map[string]*WebAuthnSession
	credentials map[string][]*WebAuthnCredential
}

// WebAuthnSession holds session data for WebAuthn ceremonies
type WebAuthnSession struct {
	Challenge   string
	Username    string
	CreatedAt   time.Time
	SessionType string // "registration" or "authentication"
}

var (
	webAuthnStore = &WebAuthnStore{
		sessions:    make(map[string]*WebAuthnSession),
		credentials: make(map[string][]*WebAuthnCredential),
	}
	sessionTimeout = 5 * time.Minute
)

// BeginWebAuthnRegistration starts WebAuthn registration ceremony
func BeginWebAuthnRegistration(c echo.Context) error {
	var req WebAuthnRegistrationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	if req.Username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Username is required",
		})
	}

	// Store the request in context for later use in FinishWebAuthnRegistration
	c.Set("webauthn_registration_request", &req)

	// Generate challenge
	challenge, err := generateChallenge()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate challenge",
		})
	}

	// Store session
	session := &WebAuthnSession{
		Challenge:   challenge,
		Username:    req.Username,
		CreatedAt:   time.Now(),
		SessionType: "registration",
	}

	webAuthnStore.mu.Lock()
	webAuthnStore.sessions[req.Username] = session
	webAuthnStore.mu.Unlock()

	// Create registration response
	response := WebAuthnRegistrationResponse{
		Challenge: challenge,
		RP: WebAuthnRPEntity{
			ID:   "localhost", // TODO: Get from config
			Name: "Sonr Identity Platform",
		},
		User: WebAuthnUserEntity{
			ID:          base64.URLEncoding.EncodeToString([]byte(req.Username)),
			Name:        req.Username,
			DisplayName: req.Username,
		},
		PubKeyCredParams: []WebAuthnCredParam{
			{Type: "public-key", Alg: -7},   // ES256
			{Type: "public-key", Alg: -257}, // RS256
		},
		AuthenticatorSelection: WebAuthnAuthenticatorSelection{
			AuthenticatorAttachment: "platform",
			UserVerification:        "required",
			ResidentKey:             "preferred",
		},
		Timeout:     60000,
		Attestation: "direct",
	}

	return c.JSON(http.StatusOK, response)
}

// FinishWebAuthnRegistration completes WebAuthn registration ceremony
func FinishWebAuthnRegistration(c echo.Context) error {
	username := c.QueryParam("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Username is required",
		})
	}

	// Parse registration response
	var regResponse map[string]any
	if err := c.Bind(&regResponse); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid registration response",
		})
	}

	// Get stored session
	webAuthnStore.mu.RLock()
	session, exists := webAuthnStore.sessions[username]
	webAuthnStore.mu.RUnlock()

	if !exists || session.SessionType != "registration" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No registration session found",
		})
	}

	// Check session timeout
	if time.Since(session.CreatedAt) > sessionTimeout {
		webAuthnStore.mu.Lock()
		delete(webAuthnStore.sessions, username)
		webAuthnStore.mu.Unlock()
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Registration session expired",
		})
	}

	// Extract credential data
	credentialID, _ := regResponse["id"].(string)
	rawID, _ := regResponse["rawId"].(string)
	response, _ := regResponse["response"].(map[string]any)
	clientDataJSON, _ := response["clientDataJSON"].(string)
	attestationObject, _ := response["attestationObject"].(string)

	// Verify client data
	if err := verifyClientData(clientDataJSON, session.Challenge, "webauthn.create"); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Client data verification failed: %v", err),
		})
	}

	// Extract public key from attestation
	publicKey, algorithm, err := extractPublicKeyFromAttestation(attestationObject)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to extract public key: %v", err),
		})
	}

	// Create credential
	credential := &WebAuthnCredential{
		CredentialID:      credentialID,
		RawID:             rawID,
		ClientDataJSON:    clientDataJSON,
		AttestationObject: attestationObject,
		Username:          username,
		Origin:            "localhost", // TODO: Extract from client data
		PublicKey:         publicKey,
		Algorithm:         algorithm,
		CreatedAt:         time.Now(),
	}

	// Store credential
	webAuthnStore.mu.Lock()
	webAuthnStore.credentials[username] = append(webAuthnStore.credentials[username], credential)
	delete(webAuthnStore.sessions, username)
	webAuthnStore.mu.Unlock()

	// Check if we should broadcast to blockchain
	broadcastReq, ok := c.Get("broadcast_to_chain").(bool)
	if !ok {
		// Check from original request stored in context
		if origReq, exists := c.Get("webauthn_registration_request").(*WebAuthnRegistrationRequest); exists {
			broadcastReq = origReq.BroadcastToChain
		}
	}

	var broadcastResult *BroadcastResponse
	if broadcastReq {
		// Broadcast WebAuthn credential to blockchain as gasless transaction
		result, err := BroadcastWebAuthnRegistration(credential, true)
		if err != nil {
			// Log error but don't fail registration
			c.Logger().Error("Failed to broadcast WebAuthn credential:", err)
		} else {
			broadcastResult = result
		}
	}

	// Check if we should create a vault
	autoCreateVault, ok := c.Get("auto_create_vault").(bool)
	if !ok {
		// Check from original request
		if origReq, exists := c.Get("webauthn_registration_request").(*WebAuthnRegistrationRequest); exists {
			autoCreateVault = origReq.AutoCreateVault
		}
	}

	var vaultResult *BroadcastResponse
	if autoCreateVault {
		// Create vault for the user
		vaultConfig := map[string]any{
			"type":       "standard",
			"encryption": "AES256",
			"owner":      username,
		}

		userDID := fmt.Sprintf("did:sonr:%s", username)
		result, err := BroadcastVaultCreation(userDID, vaultConfig)
		if err != nil {
			// Log error but don't fail registration
			c.Logger().Error("Failed to create vault:", err)
		} else {
			vaultResult = result
		}
	}

	finalResponse := map[string]any{
		"success":      true,
		"message":      "Registration completed successfully",
		"credentialId": credentialID,
	}

	if broadcastResult != nil {
		finalResponse["broadcast"] = broadcastResult
	}

	if vaultResult != nil {
		finalResponse["vault"] = vaultResult
	}

	return c.JSON(http.StatusOK, finalResponse)
}

// BeginWebAuthnAuthentication starts WebAuthn authentication ceremony
func BeginWebAuthnAuthentication(c echo.Context) error {
	var req WebAuthnAuthenticationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	if req.Username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Username is required",
		})
	}

	// Check if user has credentials
	webAuthnStore.mu.RLock()
	credentials, exists := webAuthnStore.credentials[req.Username]
	webAuthnStore.mu.RUnlock()

	if !exists || len(credentials) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No credentials found for user",
		})
	}

	// Generate challenge
	challenge, err := generateChallenge()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate challenge",
		})
	}

	// Store session
	session := &WebAuthnSession{
		Challenge:   challenge,
		Username:    req.Username,
		CreatedAt:   time.Now(),
		SessionType: "authentication",
	}

	webAuthnStore.mu.Lock()
	webAuthnStore.sessions[req.Username] = session
	webAuthnStore.mu.Unlock()

	// Build allowed credentials
	allowCredentials := make([]WebAuthnAllowedCred, len(credentials))
	for i, cred := range credentials {
		allowCredentials[i] = WebAuthnAllowedCred{
			Type: "public-key",
			ID:   cred.CredentialID,
		}
	}

	// Create authentication response
	response := WebAuthnAuthenticationResponse{
		Challenge:        challenge,
		Timeout:          60000,
		RPID:             "localhost", // TODO: Get from config
		AllowCredentials: allowCredentials,
		UserVerification: "required",
	}

	return c.JSON(http.StatusOK, response)
}

// FinishWebAuthnAuthentication completes WebAuthn authentication ceremony
func FinishWebAuthnAuthentication(c echo.Context) error {
	username := c.QueryParam("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Username is required",
		})
	}

	// Parse authentication response
	var authResponse map[string]any
	if err := c.Bind(&authResponse); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid authentication response",
		})
	}

	// Get stored session
	webAuthnStore.mu.RLock()
	session, exists := webAuthnStore.sessions[username]
	webAuthnStore.mu.RUnlock()

	if !exists || session.SessionType != "authentication" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No authentication session found",
		})
	}

	// Check session timeout
	if time.Since(session.CreatedAt) > sessionTimeout {
		webAuthnStore.mu.Lock()
		delete(webAuthnStore.sessions, username)
		webAuthnStore.mu.Unlock()
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Authentication session expired",
		})
	}

	// Extract response data
	credentialID, _ := authResponse["id"].(string)
	response, _ := authResponse["response"].(map[string]any)
	clientDataJSON, _ := response["clientDataJSON"].(string)
	authenticatorData, _ := response["authenticatorData"].(string)
	signature, _ := response["signature"].(string)
	userHandle, _ := response["userHandle"].(string)

	// Verify client data
	if err := verifyClientData(clientDataJSON, session.Challenge, "webauthn.get"); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Client data verification failed: %v", err),
		})
	}

	// Find matching credential
	webAuthnStore.mu.RLock()
	credentials := webAuthnStore.credentials[username]
	webAuthnStore.mu.RUnlock()

	var matchedCredential *WebAuthnCredential
	for _, cred := range credentials {
		if cred.CredentialID == credentialID {
			matchedCredential = cred
			break
		}
	}

	if matchedCredential == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid credential",
		})
	}

	// TODO: Verify signature using the stored public key
	// This would require implementing proper WebAuthn signature verification

	// Clean up session
	webAuthnStore.mu.Lock()
	delete(webAuthnStore.sessions, username)
	webAuthnStore.mu.Unlock()

	// Create authenticated session
	userDID := fmt.Sprintf("did:sonr:%s", username)
	authSession := &OIDCSession{
		SessionID:    generateSessionID(),
		UserDID:      userDID,
		ClientID:     "webauthn-client",
		Scope:        "openid profile did vault",
		AccessToken:  generateAccessToken(userDID),
		RefreshToken: generateRefreshToken(),
		ExpiresAt:    time.Now().Add(time.Hour),
		CreatedAt:    time.Now(),
	}

	// Store session for OIDC compatibility
	oidcProvider.mu.Lock()
	oidcProvider.sessions[authSession.AccessToken] = authSession
	oidcProvider.mu.Unlock()

	// Set user context for downstream handlers
	c.Set("user_did", userDID)
	c.Set("authenticated", true)
	c.Set("auth_method", "webauthn")
	c.Set("credential_id", credentialID)

	return c.JSON(http.StatusOK, map[string]any{
		"success":           true,
		"message":           "Authentication successful",
		"credentialId":      credentialID,
		"accessToken":       authSession.AccessToken,
		"expiresIn":         3600,
		"userDID":           userDID,
		"sessionId":         authSession.SessionID,
		"authenticatorData": authenticatorData,
		"signature":         signature,
		"userHandle":        userHandle,
	})
}

// generateChallenge creates a cryptographically secure challenge
func generateChallenge() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// verifyClientData verifies client data JSON
func verifyClientData(clientDataJSON, expectedChallenge, expectedType string) error {
	clientData, err := webauthn.ValidateClientDataJSONFormat(clientDataJSON)
	if err != nil {
		return fmt.Errorf("failed to parse client data: %w", err)
	}

	if clientData.Challenge != expectedChallenge {
		return fmt.Errorf("challenge mismatch")
	}

	if clientData.Type != expectedType {
		return fmt.Errorf(
			"invalid client data type: expected %s, got %s",
			expectedType,
			clientData.Type,
		)
	}

	// TODO: Verify origin from config
	expectedOrigins := []string{
		"http://localhost",
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
		"http://localhost:8084",
	}

	validOrigin := false
	for _, origin := range expectedOrigins {
		if clientData.Origin == origin {
			validOrigin = true
			break
		}
	}

	if !validOrigin {
		return fmt.Errorf("invalid origin: %s", clientData.Origin)
	}

	return nil
}

// extractPublicKeyFromAttestation extracts public key from attestation object
func extractPublicKeyFromAttestation(attestationObjectB64 string) ([]byte, int32, error) {
	// Validate format
	if err := webauthn.ValidateAttestationObjectFormat(attestationObjectB64); err != nil {
		return nil, 0, fmt.Errorf("invalid attestation format: %w", err)
	}

	// Decode attestation object
	attestationBytes, err := base64.RawURLEncoding.DecodeString(attestationObjectB64)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode attestation: %w", err)
	}

	// Parse CBOR
	var attestationObj webauthn.AttestationObject
	if err := webauthncbor.Unmarshal(attestationBytes, &attestationObj); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal attestation: %w", err)
	}

	// Unmarshal authenticator data
	if err := attestationObj.AuthData.Unmarshal(attestationObj.RawAuthData); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal auth data: %w", err)
	}

	// Check for attested credential data
	if !attestationObj.AuthData.Flags.HasAttestedCredentialData() {
		return nil, 0, fmt.Errorf("no attested credential data")
	}

	publicKey := attestationObj.AuthData.AttData.CredentialPublicKey
	if len(publicKey) == 0 {
		return nil, 0, fmt.Errorf("no public key found")
	}

	// Default to ES256 algorithm
	algorithm := int32(-7)

	return publicKey, algorithm, nil
}

// GetWebAuthnCredentials retrieves credentials for a user
func GetWebAuthnCredentials(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Username is required",
		})
	}

	webAuthnStore.mu.RLock()
	credentials, exists := webAuthnStore.credentials[username]
	webAuthnStore.mu.RUnlock()

	if !exists {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No credentials found for user",
		})
	}

	// Return sanitized credentials (without sensitive data)
	sanitized := make([]map[string]any, len(credentials))
	for i, cred := range credentials {
		sanitized[i] = map[string]any{
			"credentialId": cred.CredentialID,
			"createdAt":    cred.CreatedAt,
			"algorithm":    cred.Algorithm,
		}
	}

	return c.JSON(http.StatusOK, sanitized)
}
