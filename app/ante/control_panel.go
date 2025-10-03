package ante

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	anteinterfaces "github.com/cosmos/evm/ante/interfaces"
)

// Ensure ControlPanelKeeper implements the interface
var _ anteinterfaces.ControlPanelKeeper = (*ControlPanelKeeper)(nil)

// ControlPanelKeeper provides control panel functionality for sponsored transactions.
// This is a simple implementation that can be extended to support sponsored addresses
// and custom transaction priorities in the future.
type ControlPanelKeeper struct {
	// sponsoredAddresses could be loaded from state or configuration
	sponsoredAddresses map[string]bool
	// priority for sponsored transactions
	sponsoredTxPriority int64
}

// NewControlPanelKeeper creates a new ControlPanelKeeper instance
func NewControlPanelKeeper() *ControlPanelKeeper {
	return &ControlPanelKeeper{
		sponsoredAddresses:  make(map[string]bool),
		sponsoredTxPriority: 0, // Default priority
	}
}

// IsSponsoredAddress checks if an address is sponsored for gasless transactions
func (k *ControlPanelKeeper) IsSponsoredAddress(ctx context.Context, addr []byte) bool {
	// For now, return false for all addresses
	// This can be extended to check against a whitelist or state
	return false
}

// GetSponsoredTransactionPriority returns the priority for sponsored transactions
func (k *ControlPanelKeeper) GetSponsoredTransactionPriority(ctx context.Context) int64 {
	// Return default priority
	// This can be made configurable or dynamic based on chain state
	return k.sponsoredTxPriority
}

// SetSponsoredAddress adds or removes an address from the sponsored list
// This is a helper method for future use
func (k *ControlPanelKeeper) SetSponsoredAddress(addr sdk.AccAddress, sponsored bool) {
	if sponsored {
		k.sponsoredAddresses[addr.String()] = true
	} else {
		delete(k.sponsoredAddresses, addr.String())
	}
}

// SetSponsoredTransactionPriority updates the priority for sponsored transactions
// This is a helper method for future use
func (k *ControlPanelKeeper) SetSponsoredTransactionPriority(priority int64) {
	k.sponsoredTxPriority = priority
}
