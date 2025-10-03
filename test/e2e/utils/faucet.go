package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FaucetClient provides HTTP client for Starship faucet API
type FaucetClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewFaucetClient creates a new faucet HTTP client
func NewFaucetClient(baseURL string) *FaucetClient {
	return &FaucetClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FundRequest represents faucet funding request
type FundRequest struct {
	Address string   `json:"address"`
	Coins   []string `json:"coins"`
}

// FundResponse represents faucet funding response
type FundResponse struct {
	Status string `json:"status"`
	TxHash string `json:"tx_hash,omitempty"`
	Error  string `json:"error,omitempty"`
}

// FundAccount requests tokens from the faucet for an account
func (f *FaucetClient) FundAccount(ctx context.Context, address string, coins []sdk.Coin) (*FundResponse, error) {
	url := fmt.Sprintf("%s/credit", f.baseURL)

	// Convert coins to string format
	coinStrs := make([]string, len(coins))
	for i, coin := range coins {
		coinStrs[i] = coin.String()
	}

	reqBody := FundRequest{
		Address: address,
		Coins:   coinStrs,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fund request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create fund request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fund account: %w", err)
	}
	defer resp.Body.Close()

	var fundResp FundResponse
	if err := json.NewDecoder(resp.Body).Decode(&fundResp); err != nil {
		return nil, fmt.Errorf("failed to decode fund response: %w", err)
	}

	if fundResp.Status != "success" {
		return nil, fmt.Errorf("faucet funding failed: %s", fundResp.Error)
	}

	return &fundResp, nil
}

// FundAccountWithRetry funds an account with retry logic
func (f *FaucetClient) FundAccountWithRetry(ctx context.Context, address string, coins []sdk.Coin, maxRetries int) error {
	const retryDelay = 3 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err := f.FundAccount(ctx, address, coins)
		if err == nil {
			return nil
		}

		if attempt == maxRetries-1 {
			return fmt.Errorf("failed to fund account after %d attempts: %w", maxRetries, err)
		}

		time.Sleep(retryDelay)
	}

	return fmt.Errorf("unreachable code")
}

// CreateTestUser represents a test user with funding
type CreateTestUser struct {
	Address string
	Amount  math.Int
	Denom   string
}

// GetDefaultTestUsers returns default test users with addresses from Starship config
func GetDefaultTestUsers(amount math.Int, denom string) []CreateTestUser {
	return []CreateTestUser{
		{
			Address: "idx13a6zjh96w9z9y2defkktdc6vn4r5h3s7jwxuam", // acc0 from Starship config
			Amount:  amount,
			Denom:   denom,
		},
		{
			Address: "idx1xehj0xc24k2c740jslfyd4d6mt8c4dczgntqhg", // acc1 from Starship config
			Amount:  amount,
			Denom:   denom,
		},
		{
			Address: "idx1jyq30438zx0g4urancle25r6tk5td6pgeytpfu", // user0 from Starship config
			Amount:  amount,
			Denom:   denom,
		},
		{
			Address: "idx1wz5qn36kdakkqunkvwuuvpr2l4amd7y0m3qdq6", // user1 from Starship config
			Amount:  amount,
			Denom:   denom,
		},
	}
}

// FundTestUsers funds multiple test users
func (f *FaucetClient) FundTestUsers(ctx context.Context, users []CreateTestUser) error {
	for _, user := range users {
		coins := []sdk.Coin{
			{
				Denom:  user.Denom,
				Amount: user.Amount,
			},
		}

		if err := f.FundAccountWithRetry(ctx, user.Address, coins, 3); err != nil {
			return fmt.Errorf("failed to fund user %s: %w", user.Address, err)
		}
	}

	return nil
}
