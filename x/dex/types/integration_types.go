package types

import (
	"time"
)

// UCANCapability represents a UCAN capability for DEX operations
type UCANCapability struct {
	// Resource being accessed (e.g., "dex:swap", "dex:liquidity")
	Resource string `json:"resource"`

	// Ability being granted (e.g., "execute", "read", "write")
	Ability string `json:"ability"`

	// Additional constraints (e.g., max amount, specific pools)
	Constraints map[string]any `json:"constraints,omitempty"`

	// Expiration time
	Expiration time.Time `json:"expiration"`
}

// DWNRecord represents a record stored in DWN
type DWNRecord struct {
	// Record ID
	ID string `json:"id"`

	// DID owner
	DID string `json:"did"`

	// Record type
	Type string `json:"type"`

	// Record data
	Data any `json:"data"`

	// Timestamp
	Timestamp time.Time `json:"timestamp"`

	// Metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}
