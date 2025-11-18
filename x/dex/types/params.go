package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// DefaultParams returns default module parameters.
func DefaultParams() Params {
	return Params{
		Enabled:               true,
		MaxAccountsPerDid:     5,     // Maximum 5 ICA accounts per DID
		DefaultTimeoutSeconds: 600,   // 10 minutes timeout for ICA operations
		AllowedConnections: []string{
			// Noble testnet connection - USDC hub for Cosmos
			"noble-grand-1",
			// Osmosis testnet - Primary DEX
			"osmo-test-5",
			// Other testnets can be added here
		},
		MinSwapAmount:  "1000",          // Minimum 1000 base units
		MaxDailyVolume: "1000000000000", // 1M units daily volume cap
		RateLimits: RateLimitParams{
			MaxOpsPerBlock:       10,  // Maximum 10 operations per block
			MaxOpsPerDidPerDay:   100, // Maximum 100 operations per DID per day
			CooldownBlocks:       5,   // 5 block cooldown between operations
		},
		Fees: FeeParams{
			SwapFeeBps:      30,  // 0.3% swap fee
			LiquidityFeeBps: 20,  // 0.2% liquidity fee
			OrderFeeBps:     10,  // 0.1% order fee
			FeeCollector:    "",  // Empty means use module account
		},
	}
}

// Validate performs basic validation of module parameters.
func (p Params) Validate() error {
	if p.MaxAccountsPerDid == 0 {
		return fmt.Errorf("max_accounts_per_did must be positive")
	}

	if p.MaxAccountsPerDid > 100 {
		return fmt.Errorf("max_accounts_per_did cannot exceed 100")
	}

	if p.DefaultTimeoutSeconds == 0 {
		return fmt.Errorf("default_timeout_seconds must be positive")
	}

	if p.DefaultTimeoutSeconds > 3600 {
		return fmt.Errorf("default_timeout_seconds cannot exceed 3600 (1 hour)")
	}

	// Validate swap amounts
	if p.MinSwapAmount != "" {
		minSwap, ok := math.NewIntFromString(p.MinSwapAmount)
		if !ok {
			return fmt.Errorf("invalid min_swap_amount: %s", p.MinSwapAmount)
		}
		if minSwap.IsNegative() {
			return fmt.Errorf("min_swap_amount cannot be negative")
		}
	}

	if p.MaxDailyVolume != "" {
		maxVolume, ok := math.NewIntFromString(p.MaxDailyVolume)
		if !ok {
			return fmt.Errorf("invalid max_daily_volume: %s", p.MaxDailyVolume)
		}
		if maxVolume.IsNegative() {
			return fmt.Errorf("max_daily_volume cannot be negative")
		}
	}

	// Validate rate limits
	if err := p.RateLimits.Validate(); err != nil {
		return fmt.Errorf("invalid rate_limits: %w", err)
	}

	// Validate fees
	if err := p.Fees.Validate(); err != nil {
		return fmt.Errorf("invalid fees: %w", err)
	}

	return nil
}

// Validate performs basic validation of rate limit parameters.
func (r RateLimitParams) Validate() error {
	if r.MaxOpsPerBlock == 0 {
		return fmt.Errorf("max_ops_per_block must be positive")
	}

	if r.MaxOpsPerBlock > 1000 {
		return fmt.Errorf("max_ops_per_block cannot exceed 1000")
	}

	if r.MaxOpsPerDidPerDay == 0 {
		return fmt.Errorf("max_ops_per_did_per_day must be positive")
	}

	// Cooldown blocks can be 0 (no cooldown)
	if r.CooldownBlocks > 1000 {
		return fmt.Errorf("cooldown_blocks cannot exceed 1000")
	}

	return nil
}

// Validate performs basic validation of fee parameters.
func (f FeeParams) Validate() error {
	// Basis points validation (1 bps = 0.01%)
	maxBps := uint32(10000) // 100%

	if f.SwapFeeBps > maxBps {
		return fmt.Errorf("swap_fee_bps cannot exceed %d (100%%)", maxBps)
	}

	if f.LiquidityFeeBps > maxBps {
		return fmt.Errorf("liquidity_fee_bps cannot exceed %d (100%%)", maxBps)
	}

	if f.OrderFeeBps > maxBps {
		return fmt.Errorf("order_fee_bps cannot exceed %d (100%%)", maxBps)
	}

	// Fee collector address validation is done by the SDK when set
	// Empty string means use module account

	return nil
}
