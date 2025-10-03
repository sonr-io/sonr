package types

// UCANDelegationChain represents the UCAN delegation chain for a DID
type UCANDelegationChain struct {
	// Did is the DID this delegation chain belongs to
	Did string `json:"did"`

	// RootProof is the validator-issued root capability token
	RootProof string `json:"root_proof"`

	// OriginToken is the token for wallet admin operations
	OriginToken string `json:"origin_token"`

	// ValidatorIssuer is the DID of the validator that issued the root proof
	ValidatorIssuer string `json:"validator_issuer"`

	// CreatedAt is the unix timestamp when the chain was created
	CreatedAt int64 `json:"created_at"`

	// ExpiresAt is the unix timestamp when the origin token expires
	ExpiresAt int64 `json:"expires_at"`

	// Metadata contains additional information about the delegation chain
	Metadata map[string]string `json:"metadata"`
}

// EventUCANTokenRefreshed is emitted when a UCAN token is refreshed
type EventUCANTokenRefreshed struct {
	// Did is the DID whose token was refreshed
	Did string `json:"did"`

	// OldToken is the prefix of the old token (for security, only log prefix)
	OldToken string `json:"old_token"`

	// NewToken is the prefix of the new token
	NewToken string `json:"new_token"`

	// RefreshedAt is the unix timestamp when the token was refreshed
	RefreshedAt int64 `json:"refreshed_at"`
}
