package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"slices"
	"time"

	"cosmossdk.io/log"
	"github.com/labstack/echo/v4"
	"github.com/sonr-io/common/webauthn"
	"github.com/sonr-io/common/webauthn/webauthncbor"
	didtypes "github.com/sonr-io/sonr/x/did/types"
)

var logger = log.NewLogger(os.Stderr)

// HandleIndex handles the index route
func HandleIndex(c echo.Context) error {
	return c.String(http.StatusOK, "Sonr Auth Server")
}

// HandleHealth handles the health route
func HandleHealth(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

// HandleLogin handles the basic login route
func HandleLogin(c echo.Context) error {
	return c.String(http.StatusOK, "Login endpoint")
}

// HandleWebAuthnLogin serves the WebAuthn login HTML page
func HandleWebAuthnLogin(c echo.Context) error {
	// Support both username and identifier parameters
	username := c.QueryParam("username")
	if username == "" {
		username = c.QueryParam("identifier")
	}
	if username == "" {
		return c.String(http.StatusBadRequest, "Username or identifier parameter required")
	}

	// Check if user exists
	service := NewWebAuthnCredentialService()
	credentials, err := service.GetByUsername(username)
	if err != nil || len(credentials) == 0 {
		return c.String(
			http.StatusNotFound,
			fmt.Sprintf("No WebAuthn credentials found for user: %s", username),
		)
	}

	// Render the WebAuthn login page
	tmpl := template.Must(template.New("webauthn-login").Parse(webAuthnLoginHTML))
	return tmpl.Execute(c.Response().Writer, map[string]any{
		"Username": username,
		"RPID":     "localhost",
		"RPName":   "Sonr Identity Platform",
	})
}

// HandleBeginLogin starts the WebAuthn authentication ceremony
func HandleBeginLogin(c echo.Context) error {
	var username string

	// Handle both GET and POST requests
	if c.Request().Method == "POST" {
		// For POST requests, try to get username from body
		var body map[string]string
		if err := c.Bind(&body); err == nil {
			username = body["username"]
		}
	}

	// Fall back to query param for both GET and POST
	if username == "" {
		username = c.QueryParam("username")
	}

	// Also check for identifier parameter
	if username == "" {
		username = c.QueryParam("identifier")
	}

	if username == "" {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{"error": "Username or identifier parameter required"},
		)
	}

	logger.Info("Starting WebAuthn authentication", "username", username)

	// Check if user exists and get their credentials
	service := NewWebAuthnCredentialService()
	credentials, err := service.GetByUsername(username)
	if err != nil || len(credentials) == 0 {
		return c.JSON(
			http.StatusNotFound,
			map[string]string{
				"error": fmt.Sprintf("No WebAuthn credentials found for user: %s", username),
			},
		)
	}

	// Generate challenge
	challenge, err := generateChallenge()
	if err != nil {
		logger.Error("Failed to generate challenge", "error", err)
		return c.JSON(
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to generate challenge"},
		)
	}

	// Create authentication options
	allowCredentials := make([]map[string]any, len(credentials))
	for i, cred := range credentials {
		allowCredentials[i] = map[string]any{
			"type": "public-key",
			"id":   cred.CredentialID,
		}
	}

	options := map[string]any{
		"challenge":        challenge,
		"timeout":          60000,
		"rpId":             "localhost",
		"allowCredentials": allowCredentials,
		"userVerification": "preferred", // Changed from required to preferred for broader compatibility
	}

	// Store challenge in session
	if authServer != nil {
		if authServer.sessionStore == nil {
			authServer.sessionStore = make(map[string]string)
		}
		authServer.sessionStore[username] = challenge
	}

	logger.Info(
		"Sending authentication options",
		"username",
		username,
		"challenge",
		challenge,
		"credentialCount",
		len(credentials),
	)
	return c.JSON(http.StatusOK, options)
}

// HandleFinishLogin completes the WebAuthn authentication ceremony
func HandleFinishLogin(c echo.Context) error {
	username := c.QueryParam("username")
	if username == "" {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{"error": "Username parameter required"},
		)
	}

	// Parse authentication response from client
	var authResponse map[string]any
	if err := c.Bind(&authResponse); err != nil {
		logger.Error("Failed to parse authentication response", "error", err)
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{"error": "Invalid authentication response"},
		)
	}

	logger.Info("Received authentication response", "username", username)

	// Get stored challenge
	var storedChallenge string
	if authServer != nil && authServer.sessionStore != nil {
		storedChallenge = authServer.sessionStore[username]
	}

	if storedChallenge == "" {
		logger.Error("No stored challenge found", "username", username)
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{"error": "No challenge found for user"},
		)
	}

	// Extract credential data from the response
	credentialID, ok := authResponse["id"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid credential ID"})
	}

	response, ok := authResponse["response"].(map[string]any)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid response object"})
	}

	clientDataJSON, ok := response["clientDataJSON"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid client data JSON"})
	}

	// Verify client data and challenge for authentication
	if err := verifyClientDataForAuthentication(clientDataJSON, storedChallenge); err != nil {
		logger.Error("Client data verification failed for authentication", "error", err)
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{"error": "Authentication verification failed"},
		)
	}

	// Verify the credential exists for this user
	service := NewWebAuthnCredentialService()
	credential, err := service.GetByCredentialID(credentialID)
	if err != nil {
		logger.Error("Credential not found", "error", err, "credentialID", credentialID)
		return c.JSON(
			http.StatusNotFound,
			map[string]string{"error": "Credential not found"},
		)
	}

	if credential.Username != username {
		logger.Error(
			"Credential belongs to different user",
			"credentialUser",
			credential.Username,
			"requestedUser",
			username,
		)
		return c.JSON(
			http.StatusUnauthorized,
			map[string]string{"error": "Credential does not belong to this user"},
		)
	}

	// Clean up session
	if authServer != nil && authServer.sessionStore != nil {
		delete(authServer.sessionStore, username)
	}

	// Signal completion to CLI
	if authServer != nil && authServer.registrationDone != nil {
		select {
		case authServer.registrationDone <- nil:
			logger.Info("Authentication completion signaled to CLI", "username", username)
		default:
			logger.Warn(
				"Failed to signal authentication completion - channel full",
				"username",
				username,
			)
		}
	}

	logger.Info(
		"WebAuthn authentication completed successfully",
		"username",
		username,
		"credentialID",
		credentialID,
	)
	return c.JSON(http.StatusOK, map[string]any{
		"success":      true,
		"message":      "Authentication completed successfully",
		"credentialId": credentialID,
	})
}

// HandleWebAuthnRegister serves the WebAuthn registration HTML page
func HandleWebAuthnRegister(c echo.Context) error {
	// Support both username and identifier parameters
	username := c.QueryParam("username")
	if username == "" {
		username = c.QueryParam("identifier")
	}
	if username == "" {
		return c.String(http.StatusBadRequest, "Username or identifier parameter required")
	}

	// Render the WebAuthn registration page
	tmpl := template.Must(template.New("webauthn-register").Parse(webAuthnRegistrationHTML))
	return tmpl.Execute(c.Response().Writer, map[string]any{
		"Username": username,
		"RPID":     "localhost",
		"RPName":   "Sonr Identity Platform",
	})
}

// HandleBeginRegister starts the WebAuthn registration ceremony
func HandleBeginRegister(c echo.Context) error {
	var username string

	// Handle both GET and POST requests
	if c.Request().Method == "POST" {
		// For POST requests, try to get username from body
		var body map[string]string
		if err := c.Bind(&body); err == nil {
			username = body["username"]
		}
	}

	// Fall back to query param for both GET and POST
	if username == "" {
		username = c.QueryParam("username")
	}

	// Also check for identifier parameter
	if username == "" {
		username = c.QueryParam("identifier")
	}

	if username == "" {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{"error": "Username or identifier parameter required"},
		)
	}

	logger.Info("Starting WebAuthn registration", "username", username)

	// Generate challenge
	challenge, err := generateChallenge()
	if err != nil {
		logger.Error("Failed to generate challenge", "error", err)
		return c.JSON(
			http.StatusInternalServerError,
			map[string]string{"error": "Failed to generate challenge"},
		)
	}

	// Create registration options
	options := map[string]any{
		"challenge": challenge,
		"rp": map[string]string{
			"id":   "localhost",
			"name": "Sonr Identity Platform",
		},
		"user": map[string]any{
			"id":          base64.URLEncoding.EncodeToString([]byte(username)),
			"name":        username,
			"displayName": username,
		},
		"pubKeyCredParams": []map[string]any{
			{
				"type": "public-key",
				"alg":  -7, // ES256 algorithm (most common)
			},
			{
				"type": "public-key",
				"alg":  -257, // RS256 algorithm
			},
			{
				"type": "public-key",
				"alg":  -8, // EdDSA algorithm
			},
		},
		"authenticatorSelection": map[string]any{
			// Remove authenticatorAttachment to allow both platform and cross-platform authenticators
			// "authenticatorAttachment": "platform", // Commented out to allow QR codes
			"userVerification":   "preferred", // Changed from required to preferred for broader compatibility
			"residentKey":        "preferred",
			"requireResidentKey": false, // Allow non-resident keys for broader compatibility
		},
		"timeout":     60000,
		"attestation": "none", // Changed from direct to none for broader compatibility
	}

	// Store challenge in session (in production, use proper session store)
	if authServer != nil {
		if authServer.sessionStore == nil {
			authServer.sessionStore = make(map[string]string)
		}
		authServer.sessionStore[username] = challenge
	}

	logger.Info("Sending registration options", "username", username, "challenge", challenge)
	return c.JSON(http.StatusOK, options)
}

// HandleFinishRegister completes the WebAuthn registration ceremony
func HandleFinishRegister(c echo.Context) error {
	username := c.QueryParam("username")
	if username == "" {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{"error": "Username parameter required"},
		)
	}

	// Parse registration response from client
	var regResponse map[string]any
	if err := c.Bind(&regResponse); err != nil {
		logger.Error("Failed to parse registration response", "error", err)
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{"error": "Invalid registration response"},
		)
	}

	logger.Info("Received registration response", "username", username)

	// Get stored challenge
	var storedChallenge string
	if authServer != nil && authServer.sessionStore != nil {
		storedChallenge = authServer.sessionStore[username]
	}

	if storedChallenge == "" {
		logger.Error("No stored challenge found", "username", username)
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{"error": "No challenge found for user"},
		)
	}

	// Extract credential data from the response
	credentialID, ok := regResponse["id"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid credential ID"})
	}

	rawID, ok := regResponse["rawId"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid raw ID"})
	}

	response, ok := regResponse["response"].(map[string]any)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid response object"})
	}

	clientDataJSON, ok := response["clientDataJSON"].(string)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid client data JSON"})
	}

	attestationObject, ok := response["attestationObject"].(string)
	if !ok {
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{"error": "Invalid attestation object"},
		)
	}

	// Verify client data and challenge
	if err := verifyClientData(clientDataJSON, storedChallenge); err != nil {
		logger.Error("Client data verification failed", "error", err)
		return c.JSON(
			http.StatusBadRequest,
			map[string]string{"error": "Client data verification failed"},
		)
	}

	// Create WebAuthn credential record
	webAuthnCredential := &WebAuthnCredential{
		CredentialID:      credentialID,
		RawID:             rawID,
		ClientDataJSON:    clientDataJSON,
		AttestationObject: attestationObject,
		Username:          username,
		CreatedAt:         time.Now(),
	}

	// Process the registration and store in database
	if err := processWebAuthnRegistration(webAuthnCredential); err != nil {
		logger.Error("Failed to process WebAuthn registration", "error", err)
		return c.JSON(
			http.StatusInternalServerError,
			map[string]string{"error": "Registration processing failed"},
		)
	}

	// Store WebAuthn credential in database
	if err := storeWebAuthnCredential(webAuthnCredential); err != nil {
		logger.Error("Failed to store WebAuthn credential in database", "error", err)
		// Don't fail the registration if database storage fails
		logger.Warn("Continuing registration despite database storage failure")
	}

	// Clean up session
	if authServer != nil && authServer.sessionStore != nil {
		delete(authServer.sessionStore, username)
	}

	// Send credential data to CLI if channel is available
	if authServer != nil && authServer.credentialData != nil {
		select {
		case authServer.credentialData <- webAuthnCredential:
			logger.Info(
				"WebAuthn credential data sent to CLI",
				"username",
				username,
				"credentialID",
				credentialID,
			)
		default:
			logger.Warn("Failed to send credential data - channel full", "username", username)
		}
	}

	// Signal completion to CLI
	if authServer != nil && authServer.registrationDone != nil {
		select {
		case authServer.registrationDone <- nil:
			logger.Info("Registration completion signaled to CLI", "username", username)
		default:
			logger.Warn(
				"Failed to signal registration completion - channel full",
				"username",
				username,
			)
		}
	}

	logger.Info(
		"WebAuthn registration completed successfully",
		"username",
		username,
		"credentialID",
		credentialID,
	)
	return c.JSON(http.StatusOK, map[string]any{
		"success":      true,
		"message":      "Registration completed successfully",
		"credentialId": credentialID,
	})
}

// generateChallenge generates a cryptographically secure challenge
func generateChallenge() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// verifyClientData verifies the client data JSON and challenge using centralized WebAuthn validation
func verifyClientData(clientDataJSON, expectedChallenge string) error {
	// Parse client data using the local types validation
	clientData, err := didtypes.ValidateClientDataJSONFormat(clientDataJSON)
	if err != nil {
		return fmt.Errorf("failed to parse client data: %w", err)
	}

	// Verify challenge
	if clientData.Challenge != expectedChallenge {
		return fmt.Errorf("challenge mismatch")
	}

	// Verify type
	if clientData.Type != "webauthn.create" {
		return fmt.Errorf("invalid client data type: %s", clientData.Type)
	}

	// Verify origin (adjust for your domain)
	expectedOrigin := "http://localhost"
	if clientData.Origin != expectedOrigin &&
		!containsString(
			clientData.Origin,
			[]string{
				"http://localhost:8080",
				"http://localhost:8081",
				"http://localhost:8082",
				"http://localhost:8083",
				"http://localhost:8084",
			},
		) {
		return fmt.Errorf("invalid origin: %s", clientData.Origin)
	}

	return nil
}

// verifyClientDataForAuthentication verifies the client data JSON and challenge for authentication
func verifyClientDataForAuthentication(clientDataJSON, expectedChallenge string) error {
	// Parse client data using the local types validation
	clientData, err := didtypes.ValidateClientDataJSONFormat(clientDataJSON)
	if err != nil {
		return fmt.Errorf("failed to parse client data: %w", err)
	}

	// Verify challenge
	if clientData.Challenge != expectedChallenge {
		return fmt.Errorf("challenge mismatch")
	}

	// Verify type for authentication (webauthn.get instead of webauthn.create)
	if clientData.Type != "webauthn.get" {
		return fmt.Errorf("invalid client data type for authentication: %s", clientData.Type)
	}

	// Verify origin (adjust for your domain)
	expectedOrigin := "http://localhost"
	if clientData.Origin != expectedOrigin &&
		!containsString(
			clientData.Origin,
			[]string{
				"http://localhost:8080",
				"http://localhost:8081",
				"http://localhost:8082",
				"http://localhost:8083",
				"http://localhost:8084",
				"http://localhost:8085",
				"http://localhost:8086",
				"http://localhost:8087",
				"http://localhost:8088",
				"http://localhost:8089",
			},
		) {
		return fmt.Errorf("invalid origin: %s", clientData.Origin)
	}

	return nil
}

// containsString checks if a string is in a slice
func containsString(str string, slice []string) bool {
	return slices.Contains(slice, str)
}

// processWebAuthnRegistration processes the WebAuthn registration and extracts required fields
func processWebAuthnRegistration(credential *WebAuthnCredential) error {
	logger.Info(
		"Processing WebAuthn registration",
		"username",
		credential.Username,
		"credentialID",
		credential.CredentialID,
	)

	// Extract origin from client data JSON
	origin, err := extractOriginFromClientData(credential.ClientDataJSON)
	if err != nil {
		logger.Error("Failed to extract origin from client data", "error", err)
		return fmt.Errorf("failed to extract origin: %w", err)
	}
	credential.Origin = origin
	logger.Info("Extracted origin from client data", "origin", origin)

	// Extract public key and algorithm from attestation object
	publicKey, algorithm, err := extractPublicKeyFromAttestation(credential.AttestationObject)
	if err != nil {
		logger.Error("Failed to extract public key from attestation", "error", err)
		return fmt.Errorf("failed to extract public key: %w", err)
	}
	credential.PublicKey = publicKey
	credential.Algorithm = algorithm
	logger.Info("Extracted public key from attestation",
		"algorithm", algorithm,
		"publicKeyLength", len(publicKey))

	logger.Info(
		"WebAuthn credential data collected - ready for blockchain transaction",
		"credentialID",
		credential.CredentialID,
		"username",
		credential.Username,
		"origin",
		credential.Origin,
		"algorithm",
		credential.Algorithm,
	)
	return nil
}

// extractOriginFromClientData extracts the origin from client data JSON using centralized WebAuthn parsing
func extractOriginFromClientData(clientDataJSON string) (string, error) {
	// Use the local types validation
	clientData, err := didtypes.ValidateClientDataJSONFormat(clientDataJSON)
	if err != nil {
		return "", fmt.Errorf("failed to parse client data: %w", err)
	}

	if clientData.Origin == "" {
		return "", fmt.Errorf("origin not found in client data JSON")
	}

	return clientData.Origin, nil
}

// extractPublicKeyFromAttestation extracts public key and algorithm from attestation object using centralized WebAuthn parsing
func extractPublicKeyFromAttestation(attestationObject string) ([]byte, int32, error) {
	// Use the local types validation first
	if err := didtypes.ValidateAttestationObjectFormat(attestationObject); err != nil {
		return nil, 0, fmt.Errorf("invalid attestation object format: %w", err)
	}

	// Decode the attestation object
	attestationBytes, err := base64.RawURLEncoding.DecodeString(attestationObject)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode attestation object: %w", err)
	}

	// Parse the attestation object using the centralized WebAuthn CBOR parsing
	var attestationObj webauthn.AttestationObject
	if err := webauthncbor.Unmarshal(attestationBytes, &attestationObj); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal attestation object: %w", err)
	}

	// Unmarshal the authenticator data
	if err := attestationObj.AuthData.Unmarshal(attestationObj.RawAuthData); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal authenticator data: %w", err)
	}

	// Extract the attested credential data
	if !attestationObj.AuthData.Flags.HasAttestedCredentialData() {
		return nil, 0, fmt.Errorf("attestation object missing attested credential data")
	}

	publicKey := attestationObj.AuthData.AttData.CredentialPublicKey
	if len(publicKey) == 0 {
		return nil, 0, fmt.Errorf("no public key found in attested credential data")
	}

	// Assume ES256 algorithm for now. In the future, this could be extracted
	// from the COSE key format in the public key bytes
	algorithm := int32(-7) // ES256

	return publicKey, algorithm, nil
}

// WebAuthnCredential represents a WebAuthn credential for processing
type WebAuthnCredential struct {
	CredentialID      string
	RawID             string
	ClientDataJSON    string
	AttestationObject string
	Username          string
	CreatedAt         time.Time
	// Extracted fields
	Origin    string
	PublicKey []byte
	Algorithm int32
}

// storeWebAuthnCredential stores the WebAuthn credential in the database
func storeWebAuthnCredential(credential *WebAuthnCredential) error {
	// Initialize database if not already done
	if db == nil {
		if err := InitDB(); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
	}

	// Convert WebAuthn credential to database model
	storedCredential := &StoredWebAuthnCredential{
		CredentialID:      credential.CredentialID,
		RawID:             credential.RawID,
		ClientDataJSON:    credential.ClientDataJSON,
		AttestationObject: credential.AttestationObject,
		Username:          credential.Username,
		Origin:            "localhost", // Default for CLI registration
		RPID:              "localhost",
		Algorithm:         -7, // ES256 algorithm by default
	}

	// Store using service
	service := NewWebAuthnCredentialService()
	return service.Store(storedCredential)
}
