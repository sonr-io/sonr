package usdcswapdid

import (
	"context"
	"fmt"
	"time"

	"github.com/sonr-io/sonr/test/e2e/utils"
	dextypes "github.com/sonr-io/sonr/x/dex/types"
	didtypes "github.com/sonr-io/sonr/x/did/types"
)

// CreateTestDID creates a DID document for testing
// Returns the created DID document or error
// Note: For E2E tests, DIDs should be created using CLI commands before running tests
// This function checks if the DID exists and returns it, or returns instructions to create it
func CreateTestDID(ctx context.Context, cfg *utils.TestConfig, account, didID string) (*didtypes.DIDDocument, error) {
	// Check if DID already exists
	existing, err := QueryDID(ctx, cfg, didID)
	if err == nil && existing != nil {
		return existing, nil
	}

	// DID doesn't exist - return error with instructions
	return nil, fmt.Errorf("DID %s not found. Create it using CLI:\n"+
		"  snrd tx did create-did %s --from %s --chain-id %s --yes",
		didID, didID, account, cfg.ChainID)
}

// QueryDID queries a DID document by ID
func QueryDID(ctx context.Context, cfg *utils.TestConfig, didID string) (*didtypes.DIDDocument, error) {
	// Build query URL
	url := fmt.Sprintf("%s/sonr/did/v1/did/%s", cfg.BaseURL, didID)

	// Response structure
	var response struct {
		DidDocument *didtypes.DIDDocument `json:"did_document"`
	}

	// Make HTTP request
	if err := cfg.Client.DoRequest(ctx, url, &response); err != nil {
		return nil, fmt.Errorf("failed to query DID: %w", err)
	}

	return response.DidDocument, nil
}

// QueryDEXAccount queries a DEX account by DID and connection ID
func QueryDEXAccount(ctx context.Context, cfg *utils.TestConfig, didID, connectionID string) (*dextypes.InterchainDEXAccount, error) {
	// Build query URL
	url := fmt.Sprintf("%s/sonr/dex/v1/account/%s/%s", cfg.BaseURL, didID, connectionID)

	// Response structure
	var response struct {
		Account *dextypes.InterchainDEXAccount `json:"account"`
	}

	// Make HTTP request
	if err := cfg.Client.DoRequest(ctx, url, &response); err != nil {
		return nil, fmt.Errorf("failed to query DEX account: %w", err)
	}

	return response.Account, nil
}

// QueryDEXHistory queries transaction history for a DID
func QueryDEXHistory(ctx context.Context, cfg *utils.TestConfig, didID string) ([]*dextypes.DEXActivity, error) {
	// Build query URL
	url := fmt.Sprintf("%s/sonr/dex/v1/history/%s", cfg.BaseURL, didID)

	// Response structure
	var response struct {
		History []*dextypes.DEXActivity `json:"history"`
	}

	// Make HTTP request
	if err := cfg.Client.DoRequest(ctx, url, &response); err != nil {
		return nil, fmt.Errorf("failed to query DEX history: %w", err)
	}

	return response.History, nil
}

// QueryAllDEXAccounts queries all DEX accounts
func QueryAllDEXAccounts(ctx context.Context, cfg *utils.TestConfig) ([]*dextypes.InterchainDEXAccount, error) {
	// Build query URL
	url := fmt.Sprintf("%s/sonr/dex/v1/accounts", cfg.BaseURL)

	// Response structure
	var response struct {
		Accounts []*dextypes.InterchainDEXAccount `json:"accounts"`
	}

	// Make HTTP request
	if err := cfg.Client.DoRequest(ctx, url, &response); err != nil {
		return nil, fmt.Errorf("failed to query all DEX accounts: %w", err)
	}

	return response.Accounts, nil
}

// WaitForDEXAccountActivation waits for a DEX account to become active
// Returns true if account is active, false if timeout
func WaitForDEXAccountActivation(ctx context.Context, cfg *utils.TestConfig, didID, connectionID string, timeout time.Duration) (bool, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		account, err := QueryDEXAccount(ctx, cfg, didID, connectionID)
		if err != nil {
			// Account not found yet, continue waiting
			time.Sleep(cfg.BlockTime)
			continue
		}

		if account.Status == dextypes.ACCOUNT_STATUS_ACTIVE {
			return true, nil
		}

		// Check for failed status
		if account.Status == dextypes.ACCOUNT_STATUS_FAILED {
			return false, fmt.Errorf("DEX account activation failed")
		}

		time.Sleep(cfg.BlockTime)
	}

	return false, fmt.Errorf("timeout waiting for account activation")
}

// VerifyICAAccountAddress verifies that an ICA account has been created
func VerifyICAAccountAddress(ctx context.Context, cfg *utils.TestConfig, didID, connectionID string) (string, error) {
	account, err := QueryDEXAccount(ctx, cfg, didID, connectionID)
	if err != nil {
		return "", fmt.Errorf("failed to query DEX account: %w", err)
	}

	if account.AccountAddress == "" {
		return "", fmt.Errorf("ICA account address not yet assigned")
	}

	return account.AccountAddress, nil
}

// ValidateSwapParameters validates swap parameters before execution
func ValidateSwapParameters(sourceDenom, targetDenom string, amount, minOut int64) error {
	if sourceDenom == "" {
		return fmt.Errorf("source denom cannot be empty")
	}

	if targetDenom == "" {
		return fmt.Errorf("target denom cannot be empty")
	}

	if sourceDenom == targetDenom {
		return fmt.Errorf("source and target denoms must be different")
	}

	if amount <= 0 {
		return fmt.Errorf("swap amount must be positive")
	}

	if minOut < 0 {
		return fmt.Errorf("min output cannot be negative")
	}

	if minOut > amount {
		return fmt.Errorf("min output cannot exceed input amount (unless exchange rate > 1)")
	}

	return nil
}

// CalculateMinimumOutput calculates minimum output based on slippage tolerance
// slippagePct should be between 0 and 100 (e.g., 5 for 5% slippage)
func CalculateMinimumOutput(inputAmount int64, slippagePct float64) int64 {
	if slippagePct < 0 {
		slippagePct = 0
	}
	if slippagePct > 100 {
		slippagePct = 100
	}

	multiplier := (100 - slippagePct) / 100
	return int64(float64(inputAmount) * multiplier)
}

// FormatUSDCAmount formats a USDC amount for display
// USDC has 6 decimals, so 1000000 = 1.000000 USDC
func FormatUSDCAmount(amount int64) string {
	whole := amount / 1_000_000
	fractional := amount % 1_000_000
	return fmt.Sprintf("%d.%06d USDC", whole, fractional)
}

// FormatSNRAmount formats a SNR amount for display
// SNR has 6 decimals (usnr), so 1000000 = 1.000000 SNR
func FormatSNRAmount(amount int64) string {
	whole := amount / 1_000_000
	fractional := amount % 1_000_000
	return fmt.Sprintf("%d.%06d SNR", whole, fractional)
}

// ParseIBCDenom parses an IBC denom to extract the base denom
// IBC denoms have format: ibc/HASH
func ParseIBCDenom(ibcDenom string) (hash string, isIBC bool) {
	if len(ibcDenom) > 4 && ibcDenom[:4] == "ibc/" {
		return ibcDenom[4:], true
	}
	return ibcDenom, false
}

// BuildNobleIBCDenom builds the IBC denom for Noble USDC on Sonr chain
// This requires knowing the IBC channel from Noble to Sonr
func BuildNobleIBCDenom(channelID string) string {
	// In actual implementation, this would use IBC denom trace hashing
	// For now, return a placeholder
	return fmt.Sprintf("ibc/noble_usdc_via_%s", channelID)
}

// ExtractSwapEventData extracts relevant data from a swap event
type SwapEventData struct {
	DID          string
	ConnectionID string
	SourceDenom  string
	TargetDenom  string
	Amount       string
	MinAmountOut string
	Sequence     string
	SwapType     string
	ICAAddress   string
}

// ParseSwapEvent parses a swap event from transaction logs
func ParseSwapEvent(events []interface{}) (*SwapEventData, error) {
	// This would parse the actual event structure from the transaction
	// For now, return a placeholder
	return &SwapEventData{}, fmt.Errorf("event parsing not yet implemented")
}

// VerifyNobleConnection verifies that a Noble IBC connection exists
func VerifyNobleConnection(ctx context.Context, cfg *utils.TestConfig, connectionID string) error {
	// Build query URL for connection
	url := fmt.Sprintf("%s/ibc/core/connection/v1/connections/%s", cfg.BaseURL, connectionID)

	var response struct {
		Connection struct {
			State string `json:"state"`
		} `json:"connection"`
	}

	if err := cfg.Client.DoRequest(ctx, url, &response); err != nil {
		return fmt.Errorf("failed to query connection: %w", err)
	}

	if response.Connection.State != "STATE_OPEN" {
		return fmt.Errorf("connection is not open: %s", response.Connection.State)
	}

	return nil
}
