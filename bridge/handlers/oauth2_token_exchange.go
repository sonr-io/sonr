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

// TokenExchangeHandler implements RFC 8693 OAuth 2.0 Token Exchange
type TokenExchangeHandler struct {
	delegator   *UCANDelegator
	signer      *BlockchainUCANSigner
	tokenStore  TokenStore
	clientStore ClientStore
}

// TokenStore interface for token persistence
type TokenStore interface {
	GetToken(ctx context.Context, tokenID string) (*StoredToken, error)
	StoreToken(ctx context.Context, token *StoredToken) error
	RevokeToken(ctx context.Context, tokenID string) error
}

// ClientStore interface for OAuth client information
type ClientStore interface {
	GetClient(ctx context.Context, clientID string) (*OAuth2Client, error)
	ValidateClientCredentials(ctx context.Context, clientID, clientSecret string) error
}

// StoredToken represents a stored OAuth token
type StoredToken struct {
	TokenID      string    `json:"token_id"`
	TokenType    string    `json:"token_type"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scopes       []string  `json:"scopes"`
	ClientID     string    `json:"client_id"`
	UserDID      string    `json:"user_did,omitempty"`
	UCANToken    string    `json:"ucan_token"`
}

// TokenExchangeRequest represents an RFC 8693 token exchange request
type TokenExchangeRequest struct {
	GrantType          string `json:"grant_type"`
	Resource           string `json:"resource,omitempty"`
	Audience           string `json:"audience,omitempty"`
	Scope              string `json:"scope,omitempty"`
	RequestedTokenType string `json:"requested_token_type,omitempty"`
	SubjectToken       string `json:"subject_token"`
	SubjectTokenType   string `json:"subject_token_type"`
	ActorToken         string `json:"actor_token,omitempty"`
	ActorTokenType     string `json:"actor_token_type,omitempty"`
}

// TokenExchangeResponse represents an RFC 8693 token exchange response
type TokenExchangeResponse struct {
	AccessToken     string `json:"access_token"`
	IssuedTokenType string `json:"issued_token_type"`
	TokenType       string `json:"token_type"`
	ExpiresIn       int    `json:"expires_in,omitempty"`
	Scope           string `json:"scope,omitempty"`
	RefreshToken    string `json:"refresh_token,omitempty"`
}

// Token type identifiers from RFC 8693
const (
	TokenTypeAccessToken  = "urn:ietf:params:oauth:token-type:access_token"
	TokenTypeRefreshToken = "urn:ietf:params:oauth:token-type:refresh_token"
	TokenTypeIDToken      = "urn:ietf:params:oauth:token-type:id_token"
	TokenTypeSAML1        = "urn:ietf:params:oauth:token-type:saml1"
	TokenTypeSAML2        = "urn:ietf:params:oauth:token-type:saml2"
	TokenTypeJWT          = "urn:ietf:params:oauth:token-type:jwt"
	TokenTypeUCAN         = "urn:x-oath:params:oauth:token-type:ucan"
)

// NewTokenExchangeHandler creates a new token exchange handler
func NewTokenExchangeHandler(
	delegator *UCANDelegator,
	signer *BlockchainUCANSigner,
	tokenStore TokenStore,
	clientStore ClientStore,
) *TokenExchangeHandler {
	return &TokenExchangeHandler{
		delegator:   delegator,
		signer:      signer,
		tokenStore:  tokenStore,
		clientStore: clientStore,
	}
}

// HandleTokenExchange handles RFC 8693 token exchange requests
func (h *TokenExchangeHandler) HandleTokenExchange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req TokenExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid_request", "Failed to parse request body")
		return
	}

	// Validate grant type
	if req.GrantType != "urn:ietf:params:oauth:grant-type:token-exchange" {
		h.sendError(w, "unsupported_grant_type", "Only token-exchange grant type is supported")
		return
	}

	// Validate required parameters
	if req.SubjectToken == "" || req.SubjectTokenType == "" {
		h.sendError(w, "invalid_request", "Missing required parameters")
		return
	}

	// Authenticate client
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		h.sendError(w, "invalid_client", "Client authentication required")
		return
	}

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

	// Process token exchange based on token types
	response, err := h.processTokenExchange(ctx, &req, client)
	if err != nil {
		h.sendError(w, "invalid_request", err.Error())
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	json.NewEncoder(w).Encode(response)
}

// processTokenExchange processes the token exchange based on token types
func (h *TokenExchangeHandler) processTokenExchange(
	ctx context.Context,
	req *TokenExchangeRequest,
	client *OAuth2Client,
) (*TokenExchangeResponse, error) {
	// Determine requested token type (default to access token)
	requestedType := req.RequestedTokenType
	if requestedType == "" {
		requestedType = TokenTypeAccessToken
	}

	// Handle different subject token types
	switch req.SubjectTokenType {
	case TokenTypeAccessToken:
		return h.exchangeAccessToken(ctx, req, client, requestedType)
	case TokenTypeRefreshToken:
		return h.exchangeRefreshToken(ctx, req, client, requestedType)
	case TokenTypeJWT:
		return h.exchangeJWT(ctx, req, client, requestedType)
	case TokenTypeUCAN:
		return h.exchangeUCAN(ctx, req, client, requestedType)
	default:
		return nil, fmt.Errorf("unsupported subject token type: %s", req.SubjectTokenType)
	}
}

// exchangeAccessToken exchanges an access token for a new token
func (h *TokenExchangeHandler) exchangeAccessToken(
	ctx context.Context,
	req *TokenExchangeRequest,
	client *OAuth2Client,
	requestedType string,
) (*TokenExchangeResponse, error) {
	// Retrieve the subject token
	storedToken, err := h.tokenStore.GetToken(ctx, req.SubjectToken)
	if err != nil {
		return nil, fmt.Errorf("invalid subject token")
	}

	// Validate token hasn't expired
	if time.Now().After(storedToken.ExpiresAt) {
		return nil, fmt.Errorf("subject token has expired")
	}

	// Parse requested scopes (default to original scopes)
	requestedScopes := storedToken.Scopes
	if req.Scope != "" {
		requestedScopes = strings.Split(req.Scope, " ")
		// Validate requested scopes are subset of original
		if !h.isScopeSubset(requestedScopes, storedToken.Scopes) {
			return nil, fmt.Errorf("requested scopes exceed original token scopes")
		}
	}

	// Determine audience (default to requested audience or client ID)
	audience := req.Audience
	if audience == "" {
		// Try to extract DID from client metadata
		if clientDID, ok := client.Metadata["client_did"]; ok {
			audience = clientDID
		} else {
			audience = client.ClientID
		}
	}

	// Create new UCAN delegation based on requested type
	switch requestedType {
	case TokenTypeUCAN:
		return h.createUCANResponse(
			ctx,
			storedToken.UserDID,
			audience,
			requestedScopes,
			storedToken.UCANToken,
		)
	case TokenTypeAccessToken:
		return h.createAccessTokenResponse(ctx, storedToken.UserDID, audience, requestedScopes)
	default:
		return nil, fmt.Errorf("unsupported requested token type: %s", requestedType)
	}
}

// exchangeRefreshToken exchanges a refresh token for new tokens
func (h *TokenExchangeHandler) exchangeRefreshToken(
	ctx context.Context,
	req *TokenExchangeRequest,
	client *OAuth2Client,
	requestedType string,
) (*TokenExchangeResponse, error) {
	// Retrieve the refresh token
	storedToken, err := h.tokenStore.GetToken(ctx, req.SubjectToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Validate it's actually a refresh token
	if storedToken.TokenType != "refresh_token" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	// Create new tokens with same scopes
	clientDID := client.ClientID
	if did, ok := client.Metadata["client_did"]; ok {
		clientDID = did
	}
	return h.createAccessTokenResponse(ctx, storedToken.UserDID, clientDID, storedToken.Scopes)
}

// exchangeJWT exchanges a JWT for a UCAN token
func (h *TokenExchangeHandler) exchangeJWT(
	ctx context.Context,
	req *TokenExchangeRequest,
	client *OAuth2Client,
	requestedType string,
) (*TokenExchangeResponse, error) {
	// Verify the JWT
	ucanToken, err := h.signer.VerifySignature(req.SubjectToken)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT: %w", err)
	}

	// Extract scopes from JWT claims
	scopes := h.extractScopesFromUCAN(ucanToken)

	// Create response based on requested type
	clientDID := client.ClientID
	if did, ok := client.Metadata["client_did"]; ok {
		clientDID = did
	}

	switch requestedType {
	case TokenTypeUCAN:
		return h.createUCANResponse(ctx, ucanToken.Issuer, clientDID, scopes, req.SubjectToken)
	case TokenTypeAccessToken:
		return h.createAccessTokenResponse(ctx, ucanToken.Issuer, clientDID, scopes)
	default:
		return nil, fmt.Errorf("unsupported requested token type: %s", requestedType)
	}
}

// exchangeUCAN exchanges a UCAN token for another token type
func (h *TokenExchangeHandler) exchangeUCAN(
	ctx context.Context,
	req *TokenExchangeRequest,
	client *OAuth2Client,
	requestedType string,
) (*TokenExchangeResponse, error) {
	// Verify the UCAN token
	ucanToken, err := h.signer.VerifySignature(req.SubjectToken)
	if err != nil {
		return nil, fmt.Errorf("invalid UCAN token: %w", err)
	}

	// Validate delegation chain if actor token is provided
	if req.ActorToken != "" {
		actorToken, err := h.signer.VerifySignature(req.ActorToken)
		if err != nil {
			return nil, fmt.Errorf("invalid actor token: %w", err)
		}

		// Validate actor can act on behalf of subject
		if actorToken.Audience != ucanToken.Issuer {
			return nil, fmt.Errorf("actor token audience doesn't match subject issuer")
		}
	}

	// Extract scopes from UCAN
	scopes := h.extractScopesFromUCAN(ucanToken)

	// Handle impersonation/delegation if actor token is present
	issuer := ucanToken.Issuer
	clientDID := client.ClientID
	if did, ok := client.Metadata["client_did"]; ok {
		clientDID = did
	}

	if req.ActorToken != "" {
		// Actor is performing action on behalf of subject
		issuer = clientDID // Actor becomes the new issuer
	}

	// Create response based on requested type
	switch requestedType {
	case TokenTypeAccessToken:
		return h.createAccessTokenResponse(ctx, issuer, clientDID, scopes)
	case TokenTypeUCAN:
		// Create delegated UCAN with proof chain
		proofs := []ucan.Proof{ucan.Proof(req.SubjectToken)}
		if req.ActorToken != "" {
			proofs = append(proofs, ucan.Proof(req.ActorToken))
		}
		return h.createDelegatedUCANResponse(ctx, issuer, clientDID, scopes, proofs)
	default:
		return nil, fmt.Errorf("unsupported requested token type: %s", requestedType)
	}
}

// createUCANResponse creates a UCAN token response
func (h *TokenExchangeHandler) createUCANResponse(
	ctx context.Context,
	issuer, audience string,
	scopes []string,
	proof string,
) (*TokenExchangeResponse, error) {
	// Create UCAN token with delegation
	ucanToken, err := h.delegator.CreateDelegation(
		issuer,
		audience,
		scopes,
		time.Now().Add(time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create UCAN token: %w", err)
	}

	// Add proof if provided
	if proof != "" {
		ucanToken.Proofs = []ucan.Proof{ucan.Proof(proof)}
	}

	// Sign the token
	signedToken, err := h.signer.Sign(ucanToken)
	if err != nil {
		return nil, fmt.Errorf("failed to sign UCAN token: %w", err)
	}

	return &TokenExchangeResponse{
		AccessToken:     signedToken,
		IssuedTokenType: TokenTypeUCAN,
		TokenType:       "Bearer",
		ExpiresIn:       3600,
		Scope:           strings.Join(scopes, " "),
	}, nil
}

// createDelegatedUCANResponse creates a delegated UCAN token with proof chain
func (h *TokenExchangeHandler) createDelegatedUCANResponse(
	ctx context.Context,
	issuer, audience string,
	scopes []string,
	proofs []ucan.Proof,
) (*TokenExchangeResponse, error) {
	// Build resource context
	resourceContext := map[string]string{
		"delegation_type": "token_exchange",
		"issued_at":       fmt.Sprintf("%d", time.Now().Unix()),
	}

	// Map scopes to attenuations
	attenuations := h.delegator.scopeMapper.MapToUCAN(scopes, issuer, audience, resourceContext)

	// Create UCAN token with proof chain
	ucanToken := &ucan.Token{
		Issuer:       issuer,
		Audience:     audience,
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
		NotBefore:    time.Now().Unix(),
		Attenuations: attenuations,
		Proofs:       proofs,
		Facts: []ucan.Fact{
			{
				Data: h.createTokenExchangeFact(scopes),
			},
		},
	}

	// Sign the token
	signedToken, err := h.signer.Sign(ucanToken)
	if err != nil {
		return nil, fmt.Errorf("failed to sign delegated UCAN token: %w", err)
	}

	return &TokenExchangeResponse{
		AccessToken:     signedToken,
		IssuedTokenType: TokenTypeUCAN,
		TokenType:       "Bearer",
		ExpiresIn:       3600,
		Scope:           strings.Join(scopes, " "),
	}, nil
}

// createAccessTokenResponse creates a standard OAuth access token response
func (h *TokenExchangeHandler) createAccessTokenResponse(
	ctx context.Context,
	userDID, clientID string,
	scopes []string,
) (*TokenExchangeResponse, error) {
	// Create UCAN-backed access token
	ucanToken, err := h.delegator.CreateDelegation(
		userDID,
		clientID,
		scopes,
		time.Now().Add(time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// Generate token ID
	tokenID := h.generateTokenID()

	// Store token
	storedToken := &StoredToken{
		TokenID:     tokenID,
		TokenType:   "access_token",
		AccessToken: tokenID,
		ExpiresAt:   time.Now().Add(time.Hour),
		Scopes:      scopes,
		ClientID:    clientID,
		UserDID:     userDID,
		UCANToken:   ucanToken.Raw,
	}

	if err := h.tokenStore.StoreToken(ctx, storedToken); err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	// Generate refresh token
	refreshTokenID := h.generateTokenID()
	refreshToken := &StoredToken{
		TokenID:      refreshTokenID,
		TokenType:    "refresh_token",
		RefreshToken: refreshTokenID,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour), // 30 days
		Scopes:       scopes,
		ClientID:     clientID,
		UserDID:      userDID,
	}

	if err := h.tokenStore.StoreToken(ctx, refreshToken); err != nil {
		// Non-fatal, continue without refresh token
		refreshTokenID = ""
	}

	response := &TokenExchangeResponse{
		AccessToken:     tokenID,
		IssuedTokenType: TokenTypeAccessToken,
		TokenType:       "Bearer",
		ExpiresIn:       3600,
		Scope:           strings.Join(scopes, " "),
	}

	if refreshTokenID != "" {
		response.RefreshToken = refreshTokenID
	}

	return response, nil
}

// extractScopesFromUCAN extracts OAuth scopes from UCAN attenuations
func (h *TokenExchangeHandler) extractScopesFromUCAN(token *ucan.Token) []string {
	scopeMap := make(map[string]bool)

	for _, att := range token.Attenuations {
		scheme := att.Resource.GetScheme()
		actions := att.Capability.GetActions()

		// Map UCAN capabilities back to OAuth scopes
		for _, action := range actions {
			scope := h.mapUCANToScope(scheme, action)
			if scope != "" {
				scopeMap[scope] = true
			}
		}
	}

	scopes := make([]string, 0, len(scopeMap))
	for scope := range scopeMap {
		scopes = append(scopes, scope)
	}

	return scopes
}

// mapUCANToScope maps UCAN capability to OAuth scope
func (h *TokenExchangeHandler) mapUCANToScope(scheme, action string) string {
	// Reverse mapping from UCAN to OAuth scopes
	switch scheme {
	case "vault":
		switch action {
		case "read":
			return "vault:read"
		case "write":
			return "vault:write"
		case "sign":
			return "vault:sign"
		case "*", "admin":
			return "vault:admin"
		}
	case "service", "svc":
		switch action {
		case "read":
			return "service:read"
		case "write":
			return "service:write"
		case "*", "admin":
			return "service:manage"
		}
	case "did":
		switch action {
		case "read":
			return "did:read"
		case "write", "update":
			return "did:write"
		}
	case "dwn":
		switch action {
		case "read":
			return "dwn:read"
		case "write":
			return "dwn:write"
		}
	}

	// Default mapping
	return fmt.Sprintf("%s:%s", scheme, action)
}

// isScopeSubset checks if requested scopes are subset of allowed scopes
func (h *TokenExchangeHandler) isScopeSubset(requested, allowed []string) bool {
	allowedMap := make(map[string]bool)
	for _, scope := range allowed {
		allowedMap[scope] = true
	}

	for _, scope := range requested {
		if !allowedMap[scope] {
			// Check if parent scope is allowed
			if !h.isParentScopeAllowed(scope, allowed) {
				return false
			}
		}
	}

	return true
}

// isParentScopeAllowed checks if a parent scope grants the requested scope
func (h *TokenExchangeHandler) isParentScopeAllowed(requested string, allowed []string) bool {
	for _, scope := range allowed {
		if h.delegator.scopeMapper.IsHierarchicalScope(scope, requested) {
			return true
		}
	}
	return false
}

// createTokenExchangeFact creates a fact for token exchange
func (h *TokenExchangeHandler) createTokenExchangeFact(scopes []string) json.RawMessage {
	fact := map[string]any{
		"type":       "token_exchange",
		"scopes":     scopes,
		"issued_at":  time.Now().Unix(),
		"grant_type": "urn:ietf:params:oauth:grant-type:token-exchange",
	}

	data, _ := json.Marshal(fact)
	return json.RawMessage(data)
}

// generateTokenID generates a unique token identifier
func (h *TokenExchangeHandler) generateTokenID() string {
	// In production, use a proper UUID or random generator
	return fmt.Sprintf("tok_%d_%s", time.Now().UnixNano(), h.randomString(16))
}

// randomString generates a random string of specified length
func (h *TokenExchangeHandler) randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// sendError sends an OAuth error response
func (h *TokenExchangeHandler) sendError(
	w http.ResponseWriter,
	errorCode, errorDescription string,
) {
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
