package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WalletSponsorship represents metadata for a wallet sponsorship
// This stores additional data beyond what BasicAllowance provides
type WalletSponsorship struct {
	// Core sponsorship info
	Granter       string    `json:"granter"`        // Address of the sponsor
	Grantee       string    `json:"grantee"`        // Address of the sponsored wallet
	WalletAddress string    `json:"wallet_address"` // Same as grantee for simplicity
	VaultId       string    `json:"vault_id"`       // Associated vault ID
	CreatedAt     time.Time `json:"created_at"`     // Creation timestamp

	// Additional restrictions (not enforced by BasicAllowance)
	DailyLimit      *sdk.Coins `json:"daily_limit,omitempty"`      // Optional daily limit
	AllowedMessages []string   `json:"allowed_messages,omitempty"` // Optional message restrictions

	// Usage tracking (managed by our keeper)
	DailySpent    sdk.Coins  `json:"daily_spent"`     // Amount spent today
	LastResetDate time.Time  `json:"last_reset_date"` // Last daily reset
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
}

// SponsorshipInfo represents sponsorship information for queries
type SponsorshipInfo struct {
	Granter         string     `json:"granter"`
	Grantee         string     `json:"grantee"`
	WalletAddress   string     `json:"wallet_address"`
	VaultId         string     `json:"vault_id"`
	SpendLimit      *sdk.Coins `json:"spend_limit,omitempty"`
	Expiration      *time.Time `json:"expiration,omitempty"`
	DailyLimit      *sdk.Coins `json:"daily_limit,omitempty"`
	AllowedMessages []string   `json:"allowed_messages,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	DailySpent      sdk.Coins  `json:"daily_spent"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
}

// GaslessTransactionResponse represents the response from a gasless transaction
type GaslessTransactionResponse struct {
	Success       bool      `json:"success"`
	TxHash        string    `json:"tx_hash"`
	WalletAddress string    `json:"wallet_address"`
	GasUsed       uint64    `json:"gas_used"`
	FeesDeducted  sdk.Coins `json:"fees_deducted"`
}

// ValidateBasic performs basic validation on WalletSponsorship
func (ws *WalletSponsorship) ValidateBasic() error {
	if ws.Granter == "" {
		return ErrInvalidWalletAddress.Wrap("granter address is empty")
	}
	if ws.Grantee == "" {
		return ErrInvalidWalletAddress.Wrap("grantee address is empty")
	}
	if ws.VaultId == "" {
		return ErrVaultIDEmpty
	}

	// Validate addresses
	if _, err := sdk.AccAddressFromBech32(ws.Granter); err != nil {
		return ErrInvalidWalletAddress.Wrapf("invalid granter address: %s", err)
	}
	if _, err := sdk.AccAddressFromBech32(ws.Grantee); err != nil {
		return ErrInvalidWalletAddress.Wrapf("invalid grantee address: %s", err)
	}

	// Validate daily limit if present
	if ws.DailyLimit != nil {
		if !ws.DailyLimit.IsValid() {
			return ErrInvalidSpendLimit.Wrap("daily limit is invalid")
		}
		if !ws.DailyLimit.IsAllPositive() {
			return ErrInvalidSpendLimit.Wrap("daily limit must be positive")
		}
	}

	return nil
}

// IsDailyLimitExceeded checks if the daily limit would be exceeded by the given amount
func (ws *WalletSponsorship) IsDailyLimitExceeded(amount sdk.Coins) bool {
	if ws.DailyLimit == nil {
		return false // No daily limit
	}

	// Reset daily spent if it's a new day
	now := time.Now()
	if now.Day() != ws.LastResetDate.Day() || now.Month() != ws.LastResetDate.Month() ||
		now.Year() != ws.LastResetDate.Year() {
		ws.DailySpent = sdk.NewCoins()
		ws.LastResetDate = now
	}

	// Check if adding this amount would exceed the daily limit
	totalSpent := ws.DailySpent.Add(amount...)
	for _, limitCoin := range *ws.DailyLimit {
		spentAmount := totalSpent.AmountOf(limitCoin.Denom)
		if spentAmount.GT(limitCoin.Amount) {
			return true
		}
	}

	return false
}

// AddDailySpent adds to the daily spent amount and updates the last used time
func (ws *WalletSponsorship) AddDailySpent(amount sdk.Coins) {
	now := time.Now()

	// Reset daily spent if it's a new day
	if now.Day() != ws.LastResetDate.Day() || now.Month() != ws.LastResetDate.Month() ||
		now.Year() != ws.LastResetDate.Year() {
		ws.DailySpent = sdk.NewCoins()
		ws.LastResetDate = now
	}

	ws.DailySpent = ws.DailySpent.Add(amount...)
	ws.LastUsedAt = &now
}

// IsMessageAllowed checks if a message type is allowed by this sponsorship
func (ws *WalletSponsorship) IsMessageAllowed(msgTypeURL string) bool {
	if len(ws.AllowedMessages) == 0 {
		return true // No restrictions
	}

	for _, allowed := range ws.AllowedMessages {
		if allowed == msgTypeURL {
			return true
		}
	}
	return false
}
