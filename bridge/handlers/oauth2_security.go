package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// PKCE challenge methods
	PKCEMethodS256  = "S256"
	PKCEMethodPlain = "plain"
)

// PKCEValidator handles PKCE validation
type PKCEValidator struct{}

// NewPKCEValidator creates a new PKCE validator
func NewPKCEValidator() *PKCEValidator {
	return &PKCEValidator{}
}

// GeneratePKCEPair generates a PKCE verifier and challenge pair
func (p *PKCEValidator) GeneratePKCEPair() (verifier, challenge string, err error) {
	// Generate cryptographically secure verifier (43-128 characters)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Create S256 challenge
	challenge = p.CreateChallenge(verifier, PKCEMethodS256)

	return verifier, challenge, nil
}

// CreateChallenge creates a PKCE challenge from a verifier
func (p *PKCEValidator) CreateChallenge(verifier, method string) string {
	switch method {
	case PKCEMethodS256:
		h := sha256.Sum256([]byte(verifier))
		return base64.RawURLEncoding.EncodeToString(h[:])
	case PKCEMethodPlain:
		return verifier
	default:
		// Default to S256 for security
		h := sha256.Sum256([]byte(verifier))
		return base64.RawURLEncoding.EncodeToString(h[:])
	}
}

// Validate validates a PKCE verifier against a challenge
func (p *PKCEValidator) Validate(verifier, challenge, method string) bool {
	if method == "" {
		method = PKCEMethodPlain
	}

	computed := p.CreateChallenge(verifier, method)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}

// computePKCEChallenge computes a PKCE challenge (helper function)
func computePKCEChallenge(verifier, method string) string {
	validator := NewPKCEValidator()
	return validator.CreateChallenge(verifier, method)
}

// CSRFProtector handles CSRF protection
type CSRFProtector struct {
	secret []byte
}

// NewCSRFProtector creates a new CSRF protector
func NewCSRFProtector(secret string) *CSRFProtector {
	return &CSRFProtector{
		secret: []byte(secret),
	}
}

// GenerateToken generates a CSRF token
func (c *CSRFProtector) GenerateToken(sessionID string) (string, error) {
	// Create a unique token tied to the session
	h := hmac.New(sha256.New, c.secret)
	h.Write([]byte(sessionID))
	h.Write([]byte(time.Now().Format(time.RFC3339)))

	tokenBytes := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(tokenBytes), nil
}

// ValidateToken validates a CSRF token
func (c *CSRFProtector) ValidateToken(token, sessionID string) bool {
	// Decode the token
	tokenBytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return false
	}

	// Recreate the expected token
	h := hmac.New(sha256.New, c.secret)
	h.Write([]byte(sessionID))

	// Note: In production, you'd want to include time validation
	// and possibly store tokens with expiration
	expectedBytes := h.Sum(nil)[:len(tokenBytes)]

	return hmac.Equal(tokenBytes, expectedBytes)
}

// StateValidator validates OAuth state parameters
type StateValidator struct {
	states map[string]*StateEntry
}

// StateEntry represents a stored state parameter
type StateEntry struct {
	Value     string
	ClientID  string
	ExpiresAt time.Time
	Used      bool
}

// NewStateValidator creates a new state validator
func NewStateValidator() *StateValidator {
	validator := &StateValidator{
		states: make(map[string]*StateEntry),
	}

	// Start cleanup routine
	go validator.cleanup()

	return validator
}

// GenerateState generates a secure state parameter
func (s *StateValidator) GenerateState() (string, error) {
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(stateBytes), nil
}

// StoreState stores a state parameter for validation
func (s *StateValidator) StoreState(state, clientID string) {
	s.states[state] = &StateEntry{
		Value:     state,
		ClientID:  clientID,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Used:      false,
	}
}

// ValidateState validates and consumes a state parameter
func (s *StateValidator) ValidateState(state, clientID string) bool {
	entry, exists := s.states[state]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		delete(s.states, state)
		return false
	}

	// Check if already used
	if entry.Used {
		return false
	}

	// Check client ID matches
	if entry.ClientID != clientID {
		return false
	}

	// Mark as used
	entry.Used = true
	return true
}

// cleanup removes expired states
func (s *StateValidator) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		for state, entry := range s.states {
			if now.After(entry.ExpiresAt) {
				delete(s.states, state)
			}
		}
	}
}

// JWTClientAuthenticator handles JWT client authentication
type JWTClientAuthenticator struct {
	clientRegistry *ClientRegistry
}

// NewJWTClientAuthenticator creates a new JWT client authenticator
func NewJWTClientAuthenticator(registry *ClientRegistry) *JWTClientAuthenticator {
	return &JWTClientAuthenticator{
		clientRegistry: registry,
	}
}

// ValidateClientAssertion validates a JWT client assertion
func (j *JWTClientAuthenticator) ValidateClientAssertion(
	assertion, expectedAudience string,
) (*OAuth2Client, error) {
	// Parse the JWT without verification first to get the claims
	token, _, err := jwt.NewParser().ParseUnverified(assertion, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse client assertion: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	// Extract client ID from issuer and subject
	clientID, ok := claims["iss"].(string)
	if !ok {
		return nil, fmt.Errorf("missing issuer claim")
	}

	subject, ok := claims["sub"].(string)
	if !ok || subject != clientID {
		return nil, fmt.Errorf("issuer and subject must match")
	}

	// Validate audience
	audience, ok := claims["aud"].(string)
	if !ok || audience != expectedAudience {
		return nil, fmt.Errorf("invalid audience")
	}

	// Validate expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("assertion expired")
		}
	} else {
		return nil, fmt.Errorf("missing expiration")
	}

	// Validate not before
	if nbf, ok := claims["nbf"].(float64); ok {
		if time.Now().Unix() < int64(nbf) {
			return nil, fmt.Errorf("assertion not yet valid")
		}
	}

	// Validate issued at
	if iat, ok := claims["iat"].(float64); ok {
		// Check that the assertion is not too old (5 minutes max)
		if time.Now().Unix()-int64(iat) > 300 {
			return nil, fmt.Errorf("assertion too old")
		}
	}

	// Validate JTI for replay protection
	if jti, ok := claims["jti"].(string); !ok || jti == "" {
		return nil, fmt.Errorf("missing jti claim")
	}
	// TODO: Store and check JTI to prevent replay attacks

	// Get the client
	client, err := j.clientRegistry.GetClient(clientID)
	if err != nil {
		return nil, fmt.Errorf("unknown client: %w", err)
	}

	// TODO: Verify the JWT signature using the client's registered public key
	// This requires storing client public keys in the registry

	return client, nil
}

// SecureTokenGenerator generates cryptographically secure tokens
type SecureTokenGenerator struct {
	entropy int // bits of entropy
}

// NewSecureTokenGenerator creates a new secure token generator
func NewSecureTokenGenerator(entropyBits int) *SecureTokenGenerator {
	if entropyBits < 128 {
		entropyBits = 256 // Default to 256 bits for security
	}
	return &SecureTokenGenerator{
		entropy: entropyBits,
	}
}

// GenerateToken generates a secure random token
func (g *SecureTokenGenerator) GenerateToken() (string, error) {
	bytes := make([]byte, g.entropy/8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// GenerateHexToken generates a secure random token in hex format
func (g *SecureTokenGenerator) GenerateHexToken() (string, error) {
	bytes := make([]byte, g.entropy/8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// RateLimiter implements rate limiting for OAuth endpoints
type RateLimiter struct {
	attempts    map[string][]time.Time
	maxAttempts int
	window      time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxAttempts int, window time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		attempts:    make(map[string][]time.Time),
		maxAttempts: maxAttempts,
		window:      window,
	}

	// Start cleanup routine
	go limiter.cleanup()

	return limiter
}

// Allow checks if a request should be allowed
func (r *RateLimiter) Allow(key string) bool {
	now := time.Now()
	windowStart := now.Add(-r.window)

	// Get attempts for this key
	attempts := r.attempts[key]

	// Filter out attempts outside the window
	validAttempts := []time.Time{}
	for _, attempt := range attempts {
		if attempt.After(windowStart) {
			validAttempts = append(validAttempts, attempt)
		}
	}

	// Check if under limit
	if len(validAttempts) >= r.maxAttempts {
		return false
	}

	// Add this attempt
	validAttempts = append(validAttempts, now)
	r.attempts[key] = validAttempts

	return true
}

// cleanup removes old entries
func (r *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		windowStart := now.Add(-r.window)

		for key, attempts := range r.attempts {
			validAttempts := []time.Time{}
			for _, attempt := range attempts {
				if attempt.After(windowStart) {
					validAttempts = append(validAttempts, attempt)
				}
			}

			if len(validAttempts) == 0 {
				delete(r.attempts, key)
			} else {
				r.attempts[key] = validAttempts
			}
		}
	}
}

// OriginValidator validates request origins for CORS
type OriginValidator struct {
	allowedOrigins  map[string]bool
	allowSubdomains bool
}

// NewOriginValidator creates a new origin validator
func NewOriginValidator(origins []string, allowSubdomains bool) *OriginValidator {
	validator := &OriginValidator{
		allowedOrigins:  make(map[string]bool),
		allowSubdomains: allowSubdomains,
	}

	for _, origin := range origins {
		validator.allowedOrigins[origin] = true
	}

	return validator
}

// IsAllowed checks if an origin is allowed
func (o *OriginValidator) IsAllowed(origin string) bool {
	// Direct match
	if o.allowedOrigins[origin] {
		return true
	}

	// Check subdomain matching if enabled
	if o.allowSubdomains {
		for allowed := range o.allowedOrigins {
			if o.isSubdomainOf(origin, allowed) {
				return true
			}
		}
	}

	return false
}

// isSubdomainOf checks if origin is a subdomain of allowed
func (o *OriginValidator) isSubdomainOf(origin, allowed string) bool {
	// Simple subdomain check
	// In production, use proper URL parsing
	if strings.HasPrefix(allowed, "*.") {
		domain := strings.TrimPrefix(allowed, "*")
		return strings.HasSuffix(origin, domain)
	}
	return false
}
