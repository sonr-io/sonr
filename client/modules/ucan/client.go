// Package ucan provides a client interface for interacting with UCAN (User-Controlled Authorization Networks) functionality.
package ucan

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"github.com/sonr-io/sonr/client/config"
	"github.com/sonr-io/sonr/client/errors"
	"github.com/sonr-io/sonr/client/keys"
)

// Client provides an interface for UCAN operations.
type Client interface {
	// UCAN Token Operations
	CreateToken(ctx context.Context, req *CreateTokenRequest) (*UCANToken, error)
	AttenuateToken(ctx context.Context, req *AttenuateTokenRequest) (*UCANToken, error)
	ValidateToken(ctx context.Context, token string) (*TokenValidation, error)
	RevokeToken(ctx context.Context, tokenID string) error

	// Capability Operations
	CreateCapability(ctx context.Context, req *CreateCapabilityRequest) (*Capability, error)
	ListCapabilities(ctx context.Context, opts *ListCapabilitiesOptions) (*CapabilityListResponse, error)
	RevokeCapability(ctx context.Context, capabilityID string) error

	// Delegation Operations
	CreateDelegation(ctx context.Context, req *CreateDelegationRequest) (*Delegation, error)
	ListDelegations(ctx context.Context, opts *ListDelegationsOptions) (*DelegationListResponse, error)
	RevokeDelegation(ctx context.Context, delegationID string) error

	// Verification Operations
	VerifyToken(ctx context.Context, token string) (*VerificationResult, error)
	VerifyCapability(ctx context.Context, token string, resource string, action string) (*CapabilityVerification, error)

	// Chain Operations
	ValidateTokenChain(ctx context.Context, tokenChain []string) (*ChainValidation, error)
	ResolveTokenChain(ctx context.Context, token string) (*TokenChain, error)
}

// UCANToken represents a UCAN JWT token.
type UCANToken struct {
	Token        string         `json:"token"`             // JWT string
	ID           string         `json:"id"`                // Token ID
	Issuer       string         `json:"issuer"`            // Issuer DID
	Audience     string         `json:"audience"`          // Audience DID
	Subject      string         `json:"subject,omitempty"` // Subject DID
	IssuedAt     time.Time      `json:"issued_at"`
	ExpiresAt    time.Time      `json:"expires_at"`
	NotBefore    time.Time      `json:"not_before,omitempty"`
	Facts        []string       `json:"facts,omitempty"`
	Capabilities []*Capability  `json:"capabilities"`
	Proof        *Proof         `json:"proof"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// CreateTokenRequest configures UCAN token creation.
type CreateTokenRequest struct {
	Audience     string         `json:"audience"`             // Target audience DID
	Subject      string         `json:"subject,omitempty"`    // Subject DID (if different from issuer)
	Capabilities []*Capability  `json:"capabilities"`         // Granted capabilities
	Facts        []string       `json:"facts,omitempty"`      // Additional facts
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"` // Expiration time
	NotBefore    *time.Time     `json:"not_before,omitempty"` // Validity start time
	Metadata     map[string]any `json:"metadata,omitempty"`   // Additional metadata
}

// AttenuateTokenRequest configures token attenuation.
type AttenuateTokenRequest struct {
	ParentToken  string         `json:"parent_token"`         // Parent token to attenuate
	Audience     string         `json:"audience"`             // New audience DID
	Capabilities []*Capability  `json:"capabilities"`         // Attenuated capabilities
	Facts        []string       `json:"facts,omitempty"`      // Additional facts
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"` // New expiration (must be earlier)
	Metadata     map[string]any `json:"metadata,omitempty"`   // Additional metadata
}

// Capability represents a UCAN capability.
type Capability struct {
	Resource   string         `json:"resource"`             // Resource URI
	Actions    []string       `json:"actions"`              // Allowed actions
	Conditions map[string]any `json:"conditions,omitempty"` // Capability conditions
	Caveats    []*Caveat      `json:"caveats,omitempty"`    // Additional restrictions
}

// Caveat represents a capability caveat (restriction).
type Caveat struct {
	Type      string         `json:"type"`      // Caveat type
	Condition map[string]any `json:"condition"` // Caveat condition
}

// Proof represents cryptographic proof of authority.
type Proof struct {
	Type      string `json:"type"`                // Proof type (e.g., "Ed25519", "ECDSA")
	Created   string `json:"created"`             // Proof creation time
	Signature string `json:"signature"`           // Cryptographic signature
	Challenge string `json:"challenge,omitempty"` // Challenge if required
}

// TokenValidation contains token validation results.
type TokenValidation struct {
	Valid     bool         `json:"valid"`
	Token     *UCANToken   `json:"token,omitempty"`
	Errors    []string     `json:"errors,omitempty"`
	Warnings  []string     `json:"warnings,omitempty"`
	ExpiresAt time.Time    `json:"expires_at"`
	Chain     []*UCANToken `json:"chain,omitempty"`
}

// CreateCapabilityRequest configures capability creation.
type CreateCapabilityRequest struct {
	Resource   string         `json:"resource"`
	Actions    []string       `json:"actions"`
	Conditions map[string]any `json:"conditions,omitempty"`
	Caveats    []*Caveat      `json:"caveats,omitempty"`
	ExpiresAt  *time.Time     `json:"expires_at,omitempty"`
}

// ListCapabilitiesOptions configures capability listing.
type ListCapabilitiesOptions struct {
	Resource string `json:"resource,omitempty"`
	Action   string `json:"action,omitempty"`
	Owner    string `json:"owner,omitempty"`
	Limit    uint64 `json:"limit,omitempty"`
	Offset   uint64 `json:"offset,omitempty"`
}

// CapabilityListResponse contains a list of capabilities.
type CapabilityListResponse struct {
	Capabilities []*Capability `json:"capabilities"`
	TotalCount   uint64        `json:"total_count"`
	Limit        uint64        `json:"limit"`
	Offset       uint64        `json:"offset"`
}

// Delegation represents a UCAN delegation.
type Delegation struct {
	ID        string     `json:"id"`
	From      string     `json:"from"`  // Delegator DID
	To        string     `json:"to"`    // Delegatee DID
	Token     *UCANToken `json:"token"` // Delegation token
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	Revoked   bool       `json:"revoked"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// CreateDelegationRequest configures delegation creation.
type CreateDelegationRequest struct {
	To           string         `json:"to"`                   // Delegatee DID
	Capabilities []*Capability  `json:"capabilities"`         // Delegated capabilities
	ExpiresAt    *time.Time     `json:"expires_at,omitempty"` // Delegation expiration
	Facts        []string       `json:"facts,omitempty"`      // Additional facts
	Metadata     map[string]any `json:"metadata,omitempty"`   // Additional metadata
}

// ListDelegationsOptions configures delegation listing.
type ListDelegationsOptions struct {
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Active *bool  `json:"active,omitempty"` // Filter by active status
	Limit  uint64 `json:"limit,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
}

// DelegationListResponse contains a list of delegations.
type DelegationListResponse struct {
	Delegations []*Delegation `json:"delegations"`
	TotalCount  uint64        `json:"total_count"`
	Limit       uint64        `json:"limit"`
	Offset      uint64        `json:"offset"`
}

// VerificationResult contains token verification results.
type VerificationResult struct {
	Valid        bool          `json:"valid"`
	Token        *UCANToken    `json:"token,omitempty"`
	Chain        []*UCANToken  `json:"chain,omitempty"`
	Errors       []string      `json:"errors,omitempty"`
	Capabilities []*Capability `json:"capabilities,omitempty"`
}

// CapabilityVerification contains capability verification results.
type CapabilityVerification struct {
	Authorized bool        `json:"authorized"`
	Capability *Capability `json:"capability,omitempty"`
	Token      *UCANToken  `json:"token,omitempty"`
	Reason     string      `json:"reason,omitempty"`
	Conditions []string    `json:"conditions,omitempty"`
}

// ChainValidation contains token chain validation results.
type ChainValidation struct {
	Valid  bool         `json:"valid"`
	Chain  []*UCANToken `json:"chain"`
	Errors []string     `json:"errors,omitempty"`
	Root   *UCANToken   `json:"root,omitempty"`
}

// TokenChain represents a resolved token chain.
type TokenChain struct {
	Token   *UCANToken   `json:"token"`
	Parents []*UCANToken `json:"parents"`
	Root    *UCANToken   `json:"root"`
	Depth   int          `json:"depth"`
	Valid   bool         `json:"valid"`
	Errors  []string     `json:"errors,omitempty"`
}

// client implements the UCAN Client interface.
type client struct {
	grpcConn *grpc.ClientConn
	config   *config.NetworkConfig
	keyring  keys.KeyringManager

	// UCAN operations are primarily handled through the DWN plugin
	// and don't require separate gRPC clients
}

// NewClient creates a new UCAN client.
func NewClient(grpcConn *grpc.ClientConn, cfg *config.NetworkConfig) Client {
	return &client{
		grpcConn: grpcConn,
		config:   cfg,
		// keyring will be injected when needed
	}
}

// WithKeyring sets the keyring for UCAN operations.
func (c *client) WithKeyring(keyring keys.KeyringManager) Client {
	c.keyring = keyring
	return c
}

// CreateToken creates a new UCAN token.
func (c *client) CreateToken(ctx context.Context, req *CreateTokenRequest) (*UCANToken, error) {
	if c.keyring == nil {
		return nil, fmt.Errorf("keyring required for token creation")
	}

	// Convert request to keyring format
	ucanReq := &keys.UCANRequest{
		AudienceDID:  req.Audience,
		Capabilities: capabilitiesToMap(req.Capabilities),
		Facts:        req.Facts,
		NotBefore:    req.NotBefore,
		ExpiresAt:    req.ExpiresAt,
	}

	// Create token using keyring (DWN plugin)
	token, err := c.keyring.CreateOriginToken(ctx, ucanReq)
	if err != nil {
		return nil, errors.NewModuleError("ucan", "CreateToken", err)
	}

	// Convert to our format
	return convertToUCANToken(token, req), nil
}

// AttenuateToken creates an attenuated UCAN token.
func (c *client) AttenuateToken(ctx context.Context, req *AttenuateTokenRequest) (*UCANToken, error) {
	if c.keyring == nil {
		return nil, fmt.Errorf("keyring required for token attenuation")
	}

	// Convert request to keyring format
	attenuateReq := &keys.AttenuatedUCANRequest{
		ParentToken:  req.ParentToken,
		AudienceDID:  req.Audience,
		Capabilities: capabilitiesToMap(req.Capabilities),
		Facts:        req.Facts,
		ExpiresAt:    req.ExpiresAt,
	}

	// Create attenuated token using keyring
	token, err := c.keyring.CreateAttenuatedToken(ctx, attenuateReq)
	if err != nil {
		return nil, errors.NewModuleError("ucan", "AttenuateToken", err)
	}

	// Convert to our format
	return convertToUCANToken(token, nil), nil
}

// ValidateToken validates a UCAN token.
func (c *client) ValidateToken(ctx context.Context, token string) (*TokenValidation, error) {
	// TODO: Implement UCAN token validation using internal/ucan package
	// Should parse JWT, validate signature, check expiration, verify capability chain
	// Use ucan.ValidateToken() to perform cryptographic verification
	// Return structured validation results with errors and warnings

	return nil, errors.NewModuleError("ucan", "ValidateToken",
		fmt.Errorf("token validation not yet implemented"))
}

// RevokeToken revokes a UCAN token.
func (c *client) RevokeToken(ctx context.Context, tokenID string) error {
	// TODO: Implement UCAN token revocation mechanism
	// Should add token to on-chain revocation list or registry
	// Integrate with DWN module to store revocation records
	// Notify dependent systems of token revocation

	return errors.NewModuleError("ucan", "RevokeToken",
		fmt.Errorf("token revocation not yet implemented"))
}

// CreateCapability creates a new capability.
func (c *client) CreateCapability(ctx context.Context, req *CreateCapabilityRequest) (*Capability, error) {
	// TODO: Implement capability creation with proper validation
	// Should validate resource URIs and action permissions
	// Create capability following UCAN spec format
	// Store capability in persistent storage for later use

	return nil, errors.NewModuleError("ucan", "CreateCapability",
		fmt.Errorf("capability creation not yet implemented"))
}

// ListCapabilities lists capabilities with filtering.
func (c *client) ListCapabilities(ctx context.Context, opts *ListCapabilitiesOptions) (*CapabilityListResponse, error) {
	// TODO: Implement capability listing with filtering and pagination
	// Should query stored capabilities by resource, action, owner
	// Support pagination with limit/offset
	// Return capabilities with metadata and expiration info

	return nil, errors.NewModuleError("ucan", "ListCapabilities",
		fmt.Errorf("capability listing not yet implemented"))
}

// RevokeCapability revokes a capability.
func (c *client) RevokeCapability(ctx context.Context, capabilityID string) error {
	// TODO: Implement capability revocation mechanism
	// Should invalidate capability and update revocation registry
	// Cascade revocation to dependent capabilities
	// Notify systems using the revoked capability

	return errors.NewModuleError("ucan", "RevokeCapability",
		fmt.Errorf("capability revocation not yet implemented"))
}

// CreateDelegation creates a new delegation.
func (c *client) CreateDelegation(ctx context.Context, req *CreateDelegationRequest) (*Delegation, error) {
	// Delegation is essentially creating an attenuated token for someone else
	attenuateReq := &AttenuateTokenRequest{
		Audience:     req.To,
		Capabilities: req.Capabilities,
		Facts:        req.Facts,
		ExpiresAt:    req.ExpiresAt,
	}

	token, err := c.AttenuateToken(ctx, attenuateReq)
	if err != nil {
		return nil, errors.NewModuleError("ucan", "CreateDelegation", err)
	}

	// Convert to delegation format
	delegation := &Delegation{
		ID:        fmt.Sprintf("delegation_%d", time.Now().UnixNano()),
		From:      token.Issuer,
		To:        req.To,
		Token:     token,
		CreatedAt: token.IssuedAt,
		ExpiresAt: token.ExpiresAt,
		Revoked:   false,
	}

	return delegation, nil
}

// ListDelegations lists delegations with filtering.
func (c *client) ListDelegations(ctx context.Context, opts *ListDelegationsOptions) (*DelegationListResponse, error) {
	// TODO: Implement delegation listing with filtering
	// Should query delegations by grantor, grantee, active status
	// Support pagination and date range filtering
	// Include delegation status and expiration information

	return nil, errors.NewModuleError("ucan", "ListDelegations",
		fmt.Errorf("delegation listing not yet implemented"))
}

// RevokeDelegation revokes a delegation.
func (c *client) RevokeDelegation(ctx context.Context, delegationID string) error {
	// TODO: Implement delegation revocation mechanism
	// Should revoke underlying UCAN token for delegation
	// Update delegation status in storage
	// Notify grantee of delegation revocation

	return errors.NewModuleError("ucan", "RevokeDelegation",
		fmt.Errorf("delegation revocation not yet implemented"))
}

// VerifyToken verifies a UCAN token and its chain.
func (c *client) VerifyToken(ctx context.Context, token string) (*VerificationResult, error) {
	// TODO: Implement comprehensive UCAN token verification
	// Should verify entire delegation chain from root to current token
	// Check cryptographic signatures and capability bounds
	// Validate against revocation lists and expiration times
	// Use internal/ucan verification functions

	return nil, errors.NewModuleError("ucan", "VerifyToken",
		fmt.Errorf("token verification not yet implemented"))
}

// VerifyCapability verifies if a token grants access to a specific resource/action.
func (c *client) VerifyCapability(ctx context.Context, token string, resource string, action string) (*CapabilityVerification, error) {
	// TODO: Implement capability-specific verification
	// Should check if token contains capability for resource and action
	// Verify capability conditions and caveats are satisfied
	// Check resource URI patterns and action permissions
	// Return detailed authorization result with reasoning

	return nil, errors.NewModuleError("ucan", "VerifyCapability",
		fmt.Errorf("capability verification not yet implemented"))
}

// ValidateTokenChain validates a chain of UCAN tokens.
func (c *client) ValidateTokenChain(ctx context.Context, tokenChain []string) (*ChainValidation, error) {
	// TODO: Implement UCAN delegation chain validation
	// Should verify each token in chain is properly attenuated
	// Check parent-child relationships and capability inheritance
	// Validate chronological order and expiration bounds
	// Ensure no capability escalation in delegation chain

	return nil, errors.NewModuleError("ucan", "ValidateTokenChain",
		fmt.Errorf("token chain validation not yet implemented"))
}

// ResolveTokenChain resolves the full chain for a token.
func (c *client) ResolveTokenChain(ctx context.Context, token string) (*TokenChain, error) {
	// TODO: Implement UCAN delegation chain resolution
	// Should trace token back to root authority
	// Build complete chain with parent tokens and proofs
	// Resolve delegator DIDs and verify signatures
	// Return structured chain with validation status

	return nil, errors.NewModuleError("ucan", "ResolveTokenChain",
		fmt.Errorf("token chain resolution not yet implemented"))
}

// Utility functions

// capabilitiesToMap converts capabilities to map format for keyring.
func capabilitiesToMap(capabilities []*Capability) []map[string]any {
	var result []map[string]any

	for _, cap := range capabilities {
		capMap := map[string]any{
			"can":  cap.Actions,
			"with": cap.Resource,
		}

		if len(cap.Conditions) > 0 {
			capMap["conditions"] = cap.Conditions
		}

		if len(cap.Caveats) > 0 {
			capMap["caveats"] = cap.Caveats
		}

		result = append(result, capMap)
	}

	return result
}

// convertToUCANToken converts keyring token to UCAN token format.
func convertToUCANToken(token *keys.UCANToken, req *CreateTokenRequest) *UCANToken {
	ucanToken := &UCANToken{
		Token:    token.Token,
		ID:       fmt.Sprintf("ucan_%d", time.Now().UnixNano()),
		Issuer:   token.Issuer,
		IssuedAt: time.Now(),
	}

	if req != nil {
		ucanToken.Audience = req.Audience
		ucanToken.Subject = req.Subject
		ucanToken.Facts = req.Facts
		ucanToken.Capabilities = req.Capabilities
		ucanToken.Metadata = req.Metadata

		if req.ExpiresAt != nil {
			ucanToken.ExpiresAt = *req.ExpiresAt
		} else {
			ucanToken.ExpiresAt = time.Now().Add(time.Hour) // Default 1 hour
		}

		if req.NotBefore != nil {
			ucanToken.NotBefore = *req.NotBefore
		}
	}

	return ucanToken
}

// CreateDefaultCapability creates a basic capability.
func CreateDefaultCapability(resource string, actions []string) *Capability {
	return &Capability{
		Resource: resource,
		Actions:  actions,
	}
}

// CreateVaultCapability creates a capability for vault operations.
func CreateVaultCapability(vaultID string) *Capability {
	return &Capability{
		Resource: fmt.Sprintf("vault://%s", vaultID),
		Actions:  []string{"read", "write", "sign", "export"},
	}
}

// CreateServiceCapability creates a capability for service operations.
func CreateServiceCapability(serviceID string, actions []string) *Capability {
	return &Capability{
		Resource: fmt.Sprintf("service://%s", serviceID),
		Actions:  actions,
	}
}

// ValidateCapability validates a capability structure.
func ValidateCapability(cap *Capability) error {
	if cap.Resource == "" {
		return fmt.Errorf("capability resource cannot be empty")
	}

	if len(cap.Actions) == 0 {
		return fmt.Errorf("capability must have at least one action")
	}

	return nil
}
