package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// StarshipClient provides HTTP client for Starship REST API
type StarshipClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewStarshipClient creates a new Starship HTTP client
func NewStarshipClient(baseURL string) *StarshipClient {
	return &StarshipClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ChainQueryResponse represents common chain query response structure
type ChainQueryResponse struct {
	Height string          `json:"height"`
	Result json.RawMessage `json:"result"`
}

// BalanceResponse represents balance query response
type BalanceResponse struct {
	Balance sdk.Coin `json:"balance"`
}

// GetBalance queries the balance of an account
func (c *StarshipClient) GetBalance(ctx context.Context, address, denom string) (math.Int, error) {
	url := fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s/by_denom?denom=%s", c.baseURL, address, denom)

	var balanceResp BalanceResponse
	if err := c.doRequest(ctx, url, &balanceResp); err != nil {
		return math.ZeroInt(), fmt.Errorf("failed to query balance: %w", err)
	}

	return balanceResp.Balance.Amount, nil
}

// AllBalancesResponse represents all balances query response
type AllBalancesResponse struct {
	Balances   []sdk.Coin `json:"balances"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

// GetAllBalances queries all balances of an account
func (c *StarshipClient) GetAllBalances(ctx context.Context, address string) ([]sdk.Coin, error) {
	url := fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", c.baseURL, address)

	var balancesResp AllBalancesResponse
	if err := c.doRequest(ctx, url, &balancesResp); err != nil {
		return nil, fmt.Errorf("failed to query all balances: %w", err)
	}

	return balancesResp.Balances, nil
}

// SupplyResponse represents supply query response
type SupplyResponse struct {
	Amount sdk.Coin `json:"amount"`
}

// GetSupply queries the total supply of a denomination
func (c *StarshipClient) GetSupply(ctx context.Context, denom string) (math.Int, error) {
	url := fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/by_denom?denom=%s", c.baseURL, denom)

	var supplyResp SupplyResponse
	if err := c.doRequest(ctx, url, &supplyResp); err != nil {
		return math.ZeroInt(), fmt.Errorf("failed to query supply: %w", err)
	}

	return supplyResp.Amount.Amount, nil
}

// BankParamsResponse represents bank params query response
type BankParamsResponse struct {
	Params banktypes.Params `json:"params"`
}

// GetBankParams queries bank module parameters
func (c *StarshipClient) GetBankParams(ctx context.Context) (*banktypes.Params, error) {
	url := fmt.Sprintf("%s/cosmos/bank/v1beta1/params", c.baseURL)

	var paramsResp BankParamsResponse
	if err := c.doRequest(ctx, url, &paramsResp); err != nil {
		return nil, fmt.Errorf("failed to query bank params: %w", err)
	}

	return &paramsResp.Params, nil
}

// NodeInfoResponse represents node info query response
type NodeInfoResponse struct {
	DefaultNodeInfo struct {
		Network string `json:"network"`
		Version string `json:"version"`
		Moniker string `json:"moniker"`
	} `json:"default_node_info"`
	ApplicationVersion struct {
		Name      string `json:"name"`
		AppName   string `json:"app_name"`
		Version   string `json:"version"`
		GitCommit string `json:"git_commit"`
	} `json:"application_version"`
}

// GetNodeInfo queries node information
func (c *StarshipClient) GetNodeInfo(ctx context.Context) (*NodeInfoResponse, error) {
	url := fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/node_info", c.baseURL)

	var nodeInfo NodeInfoResponse
	if err := c.doRequest(ctx, url, &nodeInfo); err != nil {
		return nil, fmt.Errorf("failed to query node info: %w", err)
	}

	return &nodeInfo, nil
}

// doRequest performs HTTP GET request with retry logic
func (c *StarshipClient) doRequest(ctx context.Context, url string, target any) error {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt == maxRetries-1 {
				return fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
			}
			time.Sleep(retryDelay)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if attempt == maxRetries-1 {
				return fmt.Errorf("request failed with status %d", resp.StatusCode)
			}
			time.Sleep(retryDelay)
			continue
		}

		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	}

	return fmt.Errorf("unreachable code")
}

// EventSearchResponse represents event search response
type EventSearchResponse struct {
	Events     []EventResult `json:"events"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

// EventResult represents a single event result
type EventResult struct {
	Type       string          `json:"type"`
	Attributes []sdk.Attribute `json:"attributes"`
	Height     string          `json:"height"`
	TxHash     string          `json:"tx_hash"`
}

// BlockEventsResponse represents block events response
type BlockEventsResponse struct {
	Height           string      `json:"height"`
	BeginBlockEvents []sdk.Event `json:"begin_block_events"`
	EndBlockEvents   []sdk.Event `json:"end_block_events"`
	TxEvents         []TxEvents  `json:"tx_events"`
}

// TxEvents represents transaction events
type TxEvents struct {
	TxHash string      `json:"tx_hash"`
	Events []sdk.Event `json:"events"`
}

// QueryEventsByHeight queries events by block height
func (c *StarshipClient) QueryEventsByHeight(ctx context.Context, height int64) (*BlockEventsResponse, error) {
	url := fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/%d/events", c.baseURL, height)

	var eventsResp BlockEventsResponse
	if err := c.doRequest(ctx, url, &eventsResp); err != nil {
		return nil, fmt.Errorf("failed to query events by height: %w", err)
	}

	return &eventsResp, nil
}

// QueryEventsByType queries events by event type
func (c *StarshipClient) QueryEventsByType(ctx context.Context, eventType string, minHeight, maxHeight int64) (*EventSearchResponse, error) {
	query := fmt.Sprintf("message.action='%s'", eventType)
	return c.SearchEvents(ctx, query, minHeight, maxHeight)
}

// QueryEventsByAttribute queries events by attribute key-value pair
func (c *StarshipClient) QueryEventsByAttribute(ctx context.Context, key, value string, minHeight, maxHeight int64) (*EventSearchResponse, error) {
	query := fmt.Sprintf("%s='%s'", key, value)
	return c.SearchEvents(ctx, query, minHeight, maxHeight)
}

// SearchEvents performs a general event search with CometBFT query syntax
func (c *StarshipClient) SearchEvents(ctx context.Context, query string, minHeight, maxHeight int64) (*EventSearchResponse, error) {
	queryParams := url.Values{}
	queryParams.Add("query", query)
	if minHeight > 0 {
		queryParams.Add("min_height", strconv.FormatInt(minHeight, 10))
	}
	if maxHeight > 0 {
		queryParams.Add("max_height", strconv.FormatInt(maxHeight, 10))
	}

	searchURL := fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/events?%s", c.baseURL, queryParams.Encode())

	var eventsResp EventSearchResponse
	if err := c.doRequest(ctx, searchURL, &eventsResp); err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}

	return &eventsResp, nil
}

// GetLatestBlockHeight gets the latest block height
func (c *StarshipClient) GetLatestBlockHeight(ctx context.Context) (int64, error) {
	url := fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/latest", c.baseURL)

	var blockResp struct {
		Block struct {
			Header struct {
				Height string `json:"height"`
			} `json:"header"`
		} `json:"block"`
	}

	if err := c.doRequest(ctx, url, &blockResp); err != nil {
		return 0, fmt.Errorf("failed to get latest block height: %w", err)
	}

	height, err := strconv.ParseInt(blockResp.Block.Header.Height, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse block height: %w", err)
	}

	return height, nil
}

// WaitForNextBlock waits for the next block to be produced
func (c *StarshipClient) WaitForNextBlock(ctx context.Context) (int64, error) {
	currentHeight, err := c.GetLatestBlockHeight(ctx)
	if err != nil {
		return 0, err
	}

	targetHeight := currentHeight + 1
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-ticker.C:
			height, err := c.GetLatestBlockHeight(ctx)
			if err != nil {
				continue
			}
			if height >= targetHeight {
				return height, nil
			}
		}
	}
}

// FilterEventsByType filters events by type from a transaction response
func FilterEventsByType(events []struct {
	Type       string `json:"type"`
	Attributes []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"attributes"`
}, eventType string) []struct {
	Type       string `json:"type"`
	Attributes []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"attributes"`
} {
	var filtered []struct {
		Type       string `json:"type"`
		Attributes []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"attributes"`
	}

	for _, event := range events {
		if strings.Contains(event.Type, eventType) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// GetEventAttribute gets a specific attribute value from an event
func GetEventAttribute(event struct {
	Type       string `json:"type"`
	Attributes []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"attributes"`
}, key string,
) (string, bool) {
	for _, attr := range event.Attributes {
		if attr.Key == key {
			return attr.Value, true
		}
	}
	return "", false
}
