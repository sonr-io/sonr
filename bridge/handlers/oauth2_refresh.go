package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sonr-io/sonr/crypto/ucan"
)

// RefreshTokenHandler handles OAuth2 refresh token flows with UCAN chains
type RefreshTokenHandler struct {
	delegator   *UCANDelegator
	signer      *BlockchainUCANSigner
	tokenStore  TokenStore
	clientStore ClientStore
}

// RefreshTokenRequest represents an OAuth2 refresh token request
type RefreshTokenRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

// RefreshTokenResponse represents an OAuth2 refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	UCANToken    string `json:"ucan_token,omitempty"`
}

// UCANRefreshMetadata stores metadata for UCAN refresh chains
type UCANRefreshMetadata struct {
	OriginalIssuer  string        `json:"original_issuer"`
	DelegationChain []string      `json:"delegation_chain"`
	RefreshCount    int           `json:"refresh_count"`
	MaxRefreshCount int           `json:"max_refresh_count"`
	CreatedAt       time.Time     `json:"created_at"`
	LastRefreshedAt time.Time     `json:"last_refreshed_at"`
	AttenuationPath []Attenuation `json:"attenuation_path"`
}

// Attenuation represents scope reduction in the delegation chain
type Attenuation struct {
	FromScopes []string  `json:"from_scopes"`
	ToScopes   []string  `json:"to_scopes"`
	Timestamp  time.Time `json:"timestamp"`
	Reason     string    `json:"reason,omitempty"`
}

// NewRefreshTokenHandler creates a new refresh token handler
func NewRefreshTokenHandler(
	delegator *UCANDelegator,
	signer *BlockchainUCANSigner,
	tokenStore TokenStore,
	clientStore ClientStore,
) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		delegator:   delegator,
		signer:      signer,
		tokenStore:  tokenStore,
		clientStore: clientStore,
	}
}

// HandleRefreshToken handles OAuth2 refresh token requests
func (h *RefreshTokenHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request (handle both JSON and form-encoded)
	var req RefreshTokenRequest

	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.sendError(w, "invalid_request", "Failed to parse JSON request body")
			return
		}
	} else {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			h.sendError(w, "invalid_request", "Failed to parse form data")
			return
		}

		req.GrantType = r.FormValue("grant_type")
		req.RefreshToken = r.FormValue("refresh_token")
		req.Scope = r.FormValue("scope")
		req.ClientID = r.FormValue("client_id")
		req.ClientSecret = r.FormValue("client_secret")
	}

	// Validate grant type
	if req.GrantType != "refresh_token" {
		h.sendError(w, "unsupported_grant_type", "Only refresh_token grant type is supported")
		return
	}

	// Validate refresh token
	if req.RefreshToken == "" {
		h.sendError(w, "invalid_request", "Missing refresh_token parameter")
		return
	}

	// Authenticate client
	clientID, clientSecret := h.extractClientCredentials(r, &req)
	ctx := r.Context()

	if err := h.clientStore.ValidateClientCredentials(ctx, clientID, clientSecret); err != nil {
		h.sendError(w, "invalid_client", "Client authentication failed")
		return
	}

	// Get client information
	client, err := h.clientStore.GetClient(ctx, clientID)
	if err != nil {
		h.sendError(w, "invalid_client", "Client not found")
		return
	}

	// Process refresh token
	response, err := h.processRefreshToken(ctx, &req, client)
	if err != nil {
		h.sendError(w, "invalid_grant", err.Error())
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	json.NewEncoder(w).Encode(response)
}

// processRefreshToken processes the refresh token and returns new tokens
func (h *RefreshTokenHandler) processRefreshToken(
	ctx context.Context,
	req *RefreshTokenRequest,
	client *OAuth2Client,
) (*RefreshTokenResponse, error) {
	// Retrieve stored refresh token
	storedToken, err := h.tokenStore.GetToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Validate token type
	if storedToken.TokenType != "refresh_token" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	// Validate client binding
	if storedToken.ClientID != client.ClientID {
		return nil, fmt.Errorf("refresh token was issued to a different client")
	}

	// Check if refresh token has expired
	if time.Now().After(storedToken.ExpiresAt) {
		return nil, fmt.Errorf("refresh token has expired")
	}

	// Get refresh metadata
	metadata, err := h.getRefreshMetadata(ctx, req.RefreshToken)
	if err != nil {
		// Initialize metadata for first refresh
		metadata = &UCANRefreshMetadata{
			OriginalIssuer:  storedToken.UserDID,
			DelegationChain: []string{},
			RefreshCount:    0,
			MaxRefreshCount: 10, // Default max refresh count
			CreatedAt:       time.Now(),
			AttenuationPath: []Attenuation{},
		}
	}

	// Check refresh count limit
	if metadata.RefreshCount >= metadata.MaxRefreshCount {
		return nil, fmt.Errorf("refresh token has reached maximum refresh count")
	}

	// Parse requested scopes
	requestedScopes := storedToken.Scopes
	if req.Scope != "" {
		requestedScopes = strings.Split(req.Scope, " ")

		// Validate scope reduction (attenuate permissions)
		if !h.validateScopeAttenuation(requestedScopes, storedToken.Scopes) {
			return nil, fmt.Errorf("requested scopes exceed refresh token scopes")
		}

		// Record attenuation
		if !h.scopesEqual(requestedScopes, storedToken.Scopes) {
			metadata.AttenuationPath = append(metadata.AttenuationPath, Attenuation{
				FromScopes: storedToken.Scopes,
				ToScopes:   requestedScopes,
				Timestamp:  time.Now(),
				Reason:     "Client requested scope reduction",
			})
		}
	}

	// Create new UCAN token with delegation chain
	newUCANToken, err := h.createRefreshedUCANToken(
		ctx,
		metadata,
		storedToken,
		client,
		requestedScopes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create refreshed UCAN token: %w", err)
	}

	// Generate new access token
	accessTokenID := h.generateTokenID()
	accessToken := &StoredToken{
		TokenID:     accessTokenID,
		TokenType:   "access_token",
		AccessToken: accessTokenID,
		ExpiresAt:   time.Now().Add(time.Hour),
		Scopes:      requestedScopes,
		ClientID:    client.ClientID,
		UserDID:     storedToken.UserDID,
		UCANToken:   newUCANToken,
	}

	if err := h.tokenStore.StoreToken(ctx, accessToken); err != nil {
		return nil, fmt.Errorf("failed to store new access token: %w", err)
	}

	// Update refresh metadata
	metadata.RefreshCount++
	metadata.LastRefreshedAt = time.Now()
	metadata.DelegationChain = append(metadata.DelegationChain, newUCANToken)

	// Optionally rotate refresh token
	newRefreshTokenID := ""
	if h.shouldRotateRefreshToken(metadata) {
		newRefreshTokenID = h.generateTokenID()
		newRefreshToken := &StoredToken{
			TokenID:      newRefreshTokenID,
			TokenType:    "refresh_token",
			RefreshToken: newRefreshTokenID,
			ExpiresAt:    time.Now().Add(30 * 24 * time.Hour), // 30 days
			Scopes:       requestedScopes,
			ClientID:     client.ClientID,
			UserDID:      storedToken.UserDID,
		}

		if err := h.tokenStore.StoreToken(ctx, newRefreshToken); err != nil {
			// Non-fatal, continue with existing refresh token
			newRefreshTokenID = ""
		} else {
			// Revoke old refresh token
			h.tokenStore.RevokeToken(ctx, req.RefreshToken)

			// Store metadata for new refresh token
			h.storeRefreshMetadata(ctx, newRefreshTokenID, metadata)
		}
	} else {
		// Update metadata for existing refresh token
		h.storeRefreshMetadata(ctx, req.RefreshToken, metadata)
	}

	// Build response
	response := &RefreshTokenResponse{
		AccessToken: accessTokenID,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       strings.Join(requestedScopes, " "),
		UCANToken:   newUCANToken,
	}

	if newRefreshTokenID != "" {
		response.RefreshToken = newRefreshTokenID
	}

	return response, nil
}

// createRefreshedUCANToken creates a new UCAN token with proper delegation chain
func (h *RefreshTokenHandler) createRefreshedUCANToken(
	ctx context.Context,
	metadata *UCANRefreshMetadata,
	storedToken *StoredToken,
	client *OAuth2Client,
	scopes []string,
) (string, error) {
	// Build proof chain from previous delegations
	proofs := make([]ucan.Proof, 0, len(metadata.DelegationChain))
	for _, tokenStr := range metadata.DelegationChain {
		proofs = append(proofs, ucan.Proof(tokenStr))
	}

	// Add original token as proof if exists
	if storedToken.UCANToken != "" {
		proofs = append([]ucan.Proof{ucan.Proof(storedToken.UCANToken)}, proofs...)
	}

	// Determine issuer and audience
	issuer := metadata.OriginalIssuer
	if issuer == "" {
		issuer = storedToken.UserDID
	}

	audience := client.ClientID
	if did, ok := client.Metadata["client_did"]; ok {
		audience = did
	}

	// Create resource context with refresh metadata
	resourceContext := map[string]string{
		"refresh_count":   fmt.Sprintf("%d", metadata.RefreshCount),
		"original_issuer": metadata.OriginalIssuer,
		"delegation_type": "refresh_token",
		"client_id":       client.ClientID,
	}

	// Map OAuth scopes to UCAN attenuations
	attenuations := h.delegator.scopeMapper.MapToUCAN(scopes, issuer, audience, resourceContext)

	// Create UCAN token with delegation chain
	ucanToken := &ucan.Token{
		Issuer:       issuer,
		Audience:     audience,
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
		NotBefore:    time.Now().Unix(),
		Attenuations: attenuations,
		Proofs:       proofs,
		Facts: []ucan.Fact{
			{
				Data: h.createRefreshFact(metadata, scopes),
			},
		},
	}

	// Sign the token
	signedToken, err := h.signer.Sign(ucanToken)
	if err != nil {
		return "", fmt.Errorf("failed to sign UCAN token: %w", err)
	}

	// Validate the delegation chain
	if len(metadata.DelegationChain) > 0 {
		allTokens := append([]string{storedToken.UCANToken}, metadata.DelegationChain...)
		allTokens = append(allTokens, signedToken)

		if err := h.signer.ValidateDelegationChain(allTokens); err != nil {
			return "", fmt.Errorf("invalid delegation chain: %w", err)
		}
	}

	return signedToken, nil
}

// validateScopeAttenuation validates that requested scopes are properly attenuated
func (h *RefreshTokenHandler) validateScopeAttenuation(requested, allowed []string) bool {
	// Build allowed scope map
	allowedMap := make(map[string]bool)
	for _, scope := range allowed {
		allowedMap[scope] = true
	}

	// Check each requested scope
	for _, scope := range requested {
		if !allowedMap[scope] {
			// Check if a parent scope allows this
			found := false
			for _, allowedScope := range allowed {
				if h.delegator.scopeMapper.IsHierarchicalScope(allowedScope, scope) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// shouldRotateRefreshToken determines if refresh token should be rotated
func (h *RefreshTokenHandler) shouldRotateRefreshToken(metadata *UCANRefreshMetadata) bool {
	// Rotate on every use for maximum security
	// Could be configured based on policy
	return true
}

// extractClientCredentials extracts client credentials from request
func (h *RefreshTokenHandler) extractClientCredentials(
	r *http.Request,
	req *RefreshTokenRequest,
) (string, string) {
	// Try Basic Auth first
	if clientID, clientSecret, ok := r.BasicAuth(); ok {
		return clientID, clientSecret
	}

	// Fall back to request body
	return req.ClientID, req.ClientSecret
}

// getRefreshMetadata retrieves refresh metadata from storage
func (h *RefreshTokenHandler) getRefreshMetadata(
	ctx context.Context,
	refreshTokenID string,
) (*UCANRefreshMetadata, error) {
	// In production, this would retrieve from persistent storage
	// For now, return error to initialize new metadata
	return nil, fmt.Errorf("metadata not found")
}

// storeRefreshMetadata stores refresh metadata
func (h *RefreshTokenHandler) storeRefreshMetadata(
	ctx context.Context,
	refreshTokenID string,
	metadata *UCANRefreshMetadata,
) error {
	// In production, this would persist to storage
	// For now, just return success
	return nil
}

// createRefreshFact creates a fact for refresh token
func (h *RefreshTokenHandler) createRefreshFact(
	metadata *UCANRefreshMetadata,
	scopes []string,
) json.RawMessage {
	fact := map[string]any{
		"type":              "refresh_token",
		"refresh_count":     metadata.RefreshCount,
		"original_issuer":   metadata.OriginalIssuer,
		"scopes":            scopes,
		"refreshed_at":      time.Now().Unix(),
		"delegation_length": len(metadata.DelegationChain),
	}

	// Add attenuation info if present
	if len(metadata.AttenuationPath) > 0 {
		fact["attenuations"] = len(metadata.AttenuationPath)
		lastAttenuation := metadata.AttenuationPath[len(metadata.AttenuationPath)-1]
		fact["last_attenuation"] = map[string]any{
			"from": strings.Join(lastAttenuation.FromScopes, " "),
			"to":   strings.Join(lastAttenuation.ToScopes, " "),
			"at":   lastAttenuation.Timestamp.Unix(),
		}
	}

	data, _ := json.Marshal(fact)
	return json.RawMessage(data)
}

// scopesEqual checks if two scope slices are equal
func (h *RefreshTokenHandler) scopesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]bool)
	for _, scope := range a {
		aMap[scope] = true
	}

	for _, scope := range b {
		if !aMap[scope] {
			return false
		}
	}

	return true
}

// generateTokenID generates a unique token identifier
func (h *RefreshTokenHandler) generateTokenID() string {
	// In production, use a proper UUID or secure random generator
	return fmt.Sprintf("tok_%d_%s", time.Now().UnixNano(), h.randomString(16))
}

// randomString generates a random string
func (h *RefreshTokenHandler) randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// sendError sends an OAuth error response
func (h *RefreshTokenHandler) sendError(w http.ResponseWriter, errorCode, errorDescription string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusBadRequest)

	response := map[string]string{
		"error":             errorCode,
		"error_description": errorDescription,
	}

	json.NewEncoder(w).Encode(response)
}

// HandleUCANRefresh handles UCAN-specific refresh requests
func (h *RefreshTokenHandler) HandleUCANRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse UCAN token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		h.sendError(w, "invalid_request", "Missing or invalid Authorization header")
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Verify the UCAN token
	ucanToken, err := h.signer.VerifySignature(tokenString)
	if err != nil {
		h.sendError(w, "invalid_grant", "Invalid UCAN token")
		return
	}

	// Check if token can be refreshed (not expired beyond grace period)
	gracePeriod := int64(300) // 5 minutes grace period
	if time.Now().Unix() > ucanToken.ExpiresAt+gracePeriod {
		h.sendError(w, "invalid_grant", "Token expired beyond grace period")
		return
	}

	// Create refreshed token with extended expiration
	newToken, err := h.signer.RefreshToken(tokenString, time.Hour)
	if err != nil {
		h.sendError(w, "server_error", "Failed to refresh token")
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]any{
		"ucan_token": newToken,
		"token_type": "UCAN",
		"expires_in": 3600,
	})
}
