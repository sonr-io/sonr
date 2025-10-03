package plugin

// UCAN Token Request/Response types matching Motor plugin

// NewOriginTokenRequest represents a request to create a new UCAN origin token.
type NewOriginTokenRequest struct {
	AudienceDID  string           `json:"audience_did"`           // Target audience DID for the token
	Attenuations []map[string]any `json:"attenuations,omitempty"` // Capability attenuations
	Facts        []string         `json:"facts,omitempty"`        // Additional facts to include
	NotBefore    int64            `json:"not_before,omitempty"`   // Token validity start time
	ExpiresAt    int64            `json:"expires_at,omitempty"`   // Token expiration time
}

// NewAttenuatedTokenRequest represents a request to create a delegated UCAN token.
type NewAttenuatedTokenRequest struct {
	ParentToken  string           `json:"parent_token"`           // Parent token to derive from
	AudienceDID  string           `json:"audience_did"`           // Target audience DID for the token
	Attenuations []map[string]any `json:"attenuations,omitempty"` // Capability attenuations
	Facts        []string         `json:"facts,omitempty"`        // Additional facts to include
	NotBefore    int64            `json:"not_before,omitempty"`   // Token validity start time
	ExpiresAt    int64            `json:"expires_at,omitempty"`   // Token expiration time
}

// UCANTokenResponse contains the result of UCAN token creation.
type UCANTokenResponse struct {
	Token   string `json:"token"`           // Generated UCAN token
	Issuer  string `json:"issuer"`          // Issuer DID of the token
	Address string `json:"address"`         // Address derived from issuer
	Error   string `json:"error,omitempty"` // Error message if creation failed
}

// SignDataRequest represents a request to sign arbitrary data.
type SignDataRequest struct {
	Data []byte `json:"data"` // Data bytes to sign
}

// SignDataResponse contains the result of data signing.
type SignDataResponse struct {
	Signature []byte `json:"signature"`       // Generated signature bytes
	Error     string `json:"error,omitempty"` // Error message if signing failed
}

// VerifyDataRequest represents a request to verify a signature.
type VerifyDataRequest struct {
	Data      []byte `json:"data"`      // Original data that was signed
	Signature []byte `json:"signature"` // Signature bytes to verify
}

// VerifyDataResponse contains the result of signature verification.
type VerifyDataResponse struct {
	Valid bool   `json:"valid"`           // Whether the signature is valid
	Error string `json:"error,omitempty"` // Error message if verification failed
}

// GetIssuerDIDResponse contains issuer DID and address information.
type GetIssuerDIDResponse struct {
	IssuerDID string `json:"issuer_did"`      // Issuer DID derived from enclave
	Address   string `json:"address"`         // Address derived from enclave
	ChainCode string `json:"chain_code"`      // Deterministic chain code
	Error     string `json:"error,omitempty"` // Error message if retrieval failed
}
