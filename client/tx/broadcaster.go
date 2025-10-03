// Package tx provides transaction broadcasting utilities for the Sonr client SDK.
package tx

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/sonr-io/sonr/client/config"
	"github.com/sonr-io/sonr/client/errors"
)

// BroadcastMode defines different transaction broadcasting modes.
type BroadcastMode string

const (
	// BroadcastModeSync waits for the transaction to be included in a block and returns the result.
	BroadcastModeSync BroadcastMode = "sync"

	// BroadcastModeAsync submits the transaction and returns immediately without waiting.
	BroadcastModeAsync BroadcastMode = "async"

	// BroadcastModeBlock waits for the transaction to be committed and returns the full result.
	BroadcastModeBlock BroadcastMode = "block"
)

// Broadcaster provides an interface for broadcasting transactions with different modes and retry logic.
type Broadcaster interface {
	// Broadcasting operations
	Broadcast(ctx context.Context, txBytes []byte, mode BroadcastMode) (*BroadcastResult, error)
	BroadcastSync(ctx context.Context, txBytes []byte) (*BroadcastResult, error)
	BroadcastAsync(ctx context.Context, txBytes []byte) (*BroadcastResult, error)
	BroadcastBlock(ctx context.Context, txBytes []byte) (*BroadcastResult, error)

	// Retry and monitoring
	BroadcastWithRetry(ctx context.Context, txBytes []byte, mode BroadcastMode, maxRetries int) (*BroadcastResult, error)
	WaitForConfirmation(ctx context.Context, txHash string, timeout time.Duration) (*TxConfirmation, error)

	// Configuration
	WithRetryConfig(config RetryConfig) Broadcaster
	WithTimeout(timeout time.Duration) Broadcaster
}

// TxConfirmation contains information about a confirmed transaction.
type TxConfirmation struct {
	TxHash      string
	BlockHeight int64
	BlockTime   time.Time
	Code        uint32
	Log         string
	GasWanted   int64
	GasUsed     int64
	Events      []Event
}

// RetryConfig defines retry behavior for failed broadcasts.
type RetryConfig struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// broadcaster implements Broadcaster.
type broadcaster struct {
	grpcConn        *grpc.ClientConn
	config          *config.NetworkConfig
	txServiceClient tx.ServiceClient
	retryConfig     RetryConfig
	timeout         time.Duration
}

// NewBroadcaster creates a new transaction broadcaster.
func NewBroadcaster(grpcConn *grpc.ClientConn, cfg *config.NetworkConfig) Broadcaster {
	return &broadcaster{
		grpcConn:        grpcConn,
		config:          cfg,
		txServiceClient: tx.NewServiceClient(grpcConn),
		retryConfig:     DefaultRetryConfig(),
		timeout:         30 * time.Second,
	}
}

// DefaultRetryConfig returns sensible defaults for retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
	}
}

// Broadcast broadcasts a transaction with the specified mode.
func (b *broadcaster) Broadcast(ctx context.Context, txBytes []byte, mode BroadcastMode) (*BroadcastResult, error) {
	// Convert our mode to SDK broadcast mode
	var sdkMode tx.BroadcastMode
	switch mode {
	case BroadcastModeSync:
		sdkMode = tx.BroadcastMode_BROADCAST_MODE_SYNC
	case BroadcastModeAsync:
		sdkMode = tx.BroadcastMode_BROADCAST_MODE_ASYNC
	case BroadcastModeBlock:
		sdkMode = tx.BroadcastMode_BROADCAST_MODE_BLOCK
	default:
		return nil, fmt.Errorf("invalid broadcast mode: %s", mode)
	}

	// Create broadcast request
	req := &tx.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    sdkMode,
	}

	// Apply timeout to context
	broadcastCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	// Broadcast the transaction
	resp, err := b.txServiceClient.BroadcastTx(broadcastCtx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrBroadcastFailed, "failed to broadcast transaction")
	}

	// Convert response
	return convertBroadcastResponse(resp), nil
}

// BroadcastSync broadcasts a transaction synchronously.
func (b *broadcaster) BroadcastSync(ctx context.Context, txBytes []byte) (*BroadcastResult, error) {
	return b.Broadcast(ctx, txBytes, BroadcastModeSync)
}

// BroadcastAsync broadcasts a transaction asynchronously.
func (b *broadcaster) BroadcastAsync(ctx context.Context, txBytes []byte) (*BroadcastResult, error) {
	return b.Broadcast(ctx, txBytes, BroadcastModeAsync)
}

// BroadcastBlock broadcasts a transaction and waits for block confirmation.
func (b *broadcaster) BroadcastBlock(ctx context.Context, txBytes []byte) (*BroadcastResult, error) {
	return b.Broadcast(ctx, txBytes, BroadcastModeBlock)
}

// BroadcastWithRetry broadcasts a transaction with retry logic.
func (b *broadcaster) BroadcastWithRetry(ctx context.Context, txBytes []byte, mode BroadcastMode, maxRetries int) (*BroadcastResult, error) {
	var lastErr error
	delay := b.retryConfig.InitialDelay

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err := b.Broadcast(ctx, txBytes, mode)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry on the last attempt
		if attempt == maxRetries {
			break
		}

		// Check if error is retryable
		if !isRetryableError(err) {
			break
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// Exponential backoff
			delay = time.Duration(float64(delay) * b.retryConfig.BackoffFactor)
			if delay > b.retryConfig.MaxDelay {
				delay = b.retryConfig.MaxDelay
			}
		}
	}

	return nil, errors.WrapError(lastErr, errors.ErrBroadcastFailed, "failed to broadcast transaction after %d retries", maxRetries)
}

// WaitForConfirmation waits for a transaction to be confirmed on-chain.
func (b *broadcaster) WaitForConfirmation(ctx context.Context, txHash string, timeout time.Duration) (*TxConfirmation, error) {
	// Create timeout context
	confirmCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Poll for transaction confirmation
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-confirmCtx.Done():
			return nil, errors.WrapError(confirmCtx.Err(), errors.ErrTimeout, "timeout waiting for transaction confirmation")
		case <-ticker.C:
			// Try to fetch transaction
			req := &tx.GetTxRequest{Hash: txHash}
			resp, err := b.txServiceClient.GetTx(confirmCtx, req)
			if err != nil {
				// Transaction not found yet, continue polling
				continue
			}

			// Transaction found, convert to confirmation
			confirmation := &TxConfirmation{
				TxHash:      resp.TxResponse.TxHash,
				BlockHeight: resp.TxResponse.Height,
				Code:        resp.TxResponse.Code,
				Log:         resp.TxResponse.RawLog,
				GasWanted:   resp.TxResponse.GasWanted,
				GasUsed:     resp.TxResponse.GasUsed,
				// BlockTime would need to be fetched from block info
			}

			// Convert events
			for _, event := range resp.TxResponse.Events {
				e := Event{
					Type: event.Type,
				}
				for _, attr := range event.Attributes {
					e.Attributes = append(e.Attributes, Attribute{
						Key:   attr.Key,
						Value: attr.Value,
					})
				}
				confirmation.Events = append(confirmation.Events, e)
			}

			return confirmation, nil
		}
	}
}

// WithRetryConfig sets the retry configuration.
func (b *broadcaster) WithRetryConfig(config RetryConfig) Broadcaster {
	b.retryConfig = config
	return b
}

// WithTimeout sets the broadcast timeout.
func (b *broadcaster) WithTimeout(timeout time.Duration) Broadcaster {
	b.timeout = timeout
	return b
}

// Helper functions

// convertBroadcastResponse converts SDK broadcast response to our format.
func convertBroadcastResponse(resp *tx.BroadcastTxResponse) *BroadcastResult {
	result := &BroadcastResult{
		TxHash:    resp.TxResponse.TxHash,
		Code:      resp.TxResponse.Code,
		Log:       resp.TxResponse.RawLog,
		GasWanted: resp.TxResponse.GasWanted,
		GasUsed:   resp.TxResponse.GasUsed,
		Height:    resp.TxResponse.Height,
	}

	// Convert events
	for _, event := range resp.TxResponse.Events {
		e := Event{
			Type: event.Type,
		}
		for _, attr := range event.Attributes {
			e.Attributes = append(e.Attributes, Attribute{
				Key:   attr.Key,
				Value: attr.Value,
			})
		}
		result.Events = append(result.Events, e)
	}

	return result
}

// isRetryableError determines if an error is worth retrying.
func isRetryableError(err error) bool {
	// Check for specific error types that are retryable
	if errors.IsConnectionError(err) {
		return true
	}

	// Timeouts are generally retryable
	if errors.GetErrorCode(err) == errors.CodeTimeout {
		return true
	}

	// Network unreachable errors are retryable
	if errors.GetErrorCode(err) == errors.CodeNetworkUnreachable {
		return true
	}

	// Other errors like invalid transaction, insufficient funds, etc. are not retryable
	return false
}

// BroadcastConfig provides configuration options for broadcasting.
type BroadcastConfig struct {
	Mode         BroadcastMode
	Timeout      time.Duration
	RetryConfig  RetryConfig
	WaitForBlock bool
}

// DefaultBroadcastConfig returns sensible defaults for broadcasting.
func DefaultBroadcastConfig() BroadcastConfig {
	return BroadcastConfig{
		Mode:         BroadcastModeSync,
		Timeout:      30 * time.Second,
		RetryConfig:  DefaultRetryConfig(),
		WaitForBlock: false,
	}
}

// BroadcastWithConfig broadcasts a transaction using the provided configuration.
func (b *broadcaster) BroadcastWithConfig(ctx context.Context, txBytes []byte, config BroadcastConfig) (*BroadcastResult, error) {
	// Set timeout
	originalTimeout := b.timeout
	b.timeout = config.Timeout
	defer func() { b.timeout = originalTimeout }()

	// Set retry config
	originalRetryConfig := b.retryConfig
	b.retryConfig = config.RetryConfig
	defer func() { b.retryConfig = originalRetryConfig }()

	// Broadcast with retry
	result, err := b.BroadcastWithRetry(ctx, txBytes, config.Mode, config.RetryConfig.MaxRetries)
	if err != nil {
		return nil, err
	}

	// Wait for block confirmation if requested
	if config.WaitForBlock && result.TxHash != "" {
		confirmation, err := b.WaitForConfirmation(ctx, result.TxHash, config.Timeout)
		if err != nil {
			// Return the broadcast result even if we couldn't wait for confirmation
			return result, fmt.Errorf("transaction broadcast succeeded but confirmation failed: %w", err)
		}

		// Update result with confirmation data
		result.Height = confirmation.BlockHeight
		result.Code = confirmation.Code
		result.Log = confirmation.Log
		result.GasWanted = confirmation.GasWanted
		result.GasUsed = confirmation.GasUsed
		result.Events = confirmation.Events
	}

	return result, nil
}
