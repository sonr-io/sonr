package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sonr-io/crypto/mpc"
)

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// DWNKeeper interface defines the methods needed from the DWN keeper for vault operations
type DWNKeeper interface {
	// CreateVaultForDID creates a vault for a given DID with specified parameters
	CreateVaultForDID(
		ctx context.Context,
		data *mpc.EnclaveData,
	) (*CreateVaultResponse, error)

	// GetVaultState retrieves vault state by vault ID
	GetVaultState(ctx context.Context, vaultID string) (*VaultState, error)

	// GetVaultsByDID retrieves all vaults associated with a DID
	GetVaultsByDID(ctx context.Context, did string) ([]*VaultState, error)
}

// CreateVaultResponse represents the response from vault creation
type CreateVaultResponse struct {
	VaultID        string `json:"vault_id"`
	VaultPublicKey string `json:"vault_public_key"`
	EnclaveID      string `json:"enclave_id"`
	IpfsCid        string `json:"ipfs_cid,omitempty"`
}

// VaultState represents the state of a vault
type VaultState struct {
	VaultID    string `json:"vault_id"`
	DID        string `json:"did"`
	Controller string `json:"controller"`
	Status     string `json:"status"` // active, suspended, revoked
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

// ServiceKeeper interface defines the methods needed from the Service keeper for origin validation
type ServiceKeeper interface {
	// VerifyOrigin validates a relying party origin for WebAuthn operations
	VerifyOrigin(ctx context.Context, origin string) error
}
