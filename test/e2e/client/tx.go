package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// TxResponse represents transaction broadcast response
type TxResponse struct {
	TxHash    string `json:"txhash"`
	Code      uint32 `json:"code"`
	RawLog    string `json:"raw_log"`
	GasUsed   string `json:"gas_used"`
	GasWanted string `json:"gas_wanted"`
	Height    string `json:"height"`
	Events    []struct {
		Type       string `json:"type"`
		Attributes []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"attributes"`
	} `json:"events"`
}

// BroadcastTxRequest represents transaction broadcast request
type BroadcastTxRequest struct {
	TxBytes []byte        `json:"tx_bytes"`
	Mode    BroadcastMode `json:"mode"`
}

// BroadcastMode represents different broadcast modes
type BroadcastMode string

const (
	BroadcastModeSync  BroadcastMode = "BROADCAST_MODE_SYNC"
	BroadcastModeAsync BroadcastMode = "BROADCAST_MODE_ASYNC"
	BroadcastModeBlock BroadcastMode = "BROADCAST_MODE_BLOCK"
)

// BroadcastTx broadcasts a transaction to the network
func (c *StarshipClient) BroadcastTx(ctx context.Context, txBytes []byte, mode BroadcastMode) (*TxResponse, error) {
	url := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs", c.baseURL)

	reqBody := BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    mode,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("broadcast failed with status %d", resp.StatusCode)
	}

	var broadcastResp struct {
		TxResponse TxResponse `json:"tx_response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&broadcastResp); err != nil {
		return nil, fmt.Errorf("failed to decode broadcast response: %w", err)
	}

	return &broadcastResp.TxResponse, nil
}

// GetTxResponse represents get transaction response
type GetTxResponse struct {
	Tx         tx.Tx      `json:"tx"`
	TxResponse TxResponse `json:"tx_response"`
}

// GetTx queries a transaction by hash
func (c *StarshipClient) GetTx(ctx context.Context, txHash string) (*GetTxResponse, error) {
	url := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", c.baseURL, txHash)

	var txResp GetTxResponse
	if err := c.doRequest(ctx, url, &txResp); err != nil {
		return nil, fmt.Errorf("failed to query transaction: %w", err)
	}

	return &txResp, nil
}

// SimulateRequest represents transaction simulation request
type SimulateRequest struct {
	TxBytes []byte `json:"tx_bytes"`
}

// SimulateResponse represents transaction simulation response
type SimulateResponse struct {
	GasInfo struct {
		GasWanted string `json:"gas_wanted"`
		GasUsed   string `json:"gas_used"`
	} `json:"gas_info"`
	Result struct {
		Data   string            `json:"data"`
		Log    string            `json:"log"`
		Events []sdk.StringEvent `json:"events"`
	} `json:"result"`
}

// SimulateTx simulates a transaction to estimate gas
func (c *StarshipClient) SimulateTx(ctx context.Context, txBytes []byte) (*SimulateResponse, error) {
	url := fmt.Sprintf("%s/cosmos/tx/v1beta1/simulate", c.baseURL)

	reqBody := SimulateRequest{
		TxBytes: txBytes,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal simulate request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create simulate request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to simulate transaction: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("simulation failed with status %d", resp.StatusCode)
	}

	var simResp SimulateResponse
	if err := json.NewDecoder(resp.Body).Decode(&simResp); err != nil {
		return nil, fmt.Errorf("failed to decode simulation response: %w", err)
	}

	return &simResp, nil
}

// WaitForTx waits for a transaction to be included in a block
func (c *StarshipClient) WaitForTx(ctx context.Context, txHash string, timeout time.Duration) (*GetTxResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for transaction %s", txHash)
		case <-ticker.C:
			tx, err := c.GetTx(ctx, txHash)
			if err == nil && tx.TxResponse.Code == 0 {
				return tx, nil
			}
			// Continue waiting if transaction not found or failed
		}
	}
}
