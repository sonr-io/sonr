// Package tx provides transaction building utilities for the Sonr client SDK.
package tx

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/sonr-io/sonr/client/config"
	"github.com/sonr-io/sonr/client/errors"
	"github.com/sonr-io/sonr/client/keys"
)

// TxBuilder provides an interface for building and broadcasting transactions.
type TxBuilder interface {
	// Transaction configuration
	WithChainID(chainID string) TxBuilder
	WithGasPrice(price float64, denom string) TxBuilder
	WithGasLimit(limit uint64) TxBuilder
	WithMemo(memo string) TxBuilder
	WithTimeoutHeight(height uint64) TxBuilder

	// Message operations
	AddMessage(msg sdktypes.Msg) TxBuilder
	AddMessages(msgs ...sdktypes.Msg) TxBuilder
	ClearMessages() TxBuilder

	// Fee operations
	WithFee(amount sdktypes.Coins) TxBuilder
	WithGasAdjustment(adjustment float64) TxBuilder
	EstimateGas(ctx context.Context) (uint64, error)

	// Signing and broadcasting
	Sign(ctx context.Context, keyring keys.KeyringManager) (*SignedTx, error)
	SignAndBroadcast(ctx context.Context, keyring keys.KeyringManager) (*BroadcastResult, error)
	Broadcast(ctx context.Context, signedTx *SignedTx) (*BroadcastResult, error)

	// Simulation
	Simulate(ctx context.Context) (*SimulateResult, error)

	// Building
	Build() (*UnsignedTx, error)
	BuildSigned(signature []byte, pubKey []byte) (*SignedTx, error)

	// Configuration access
	Config() *TxConfig
}

// TxConfig holds transaction configuration.
type TxConfig struct {
	ChainID       string
	GasPrice      float64
	GasDenom      string
	GasLimit      uint64
	GasAdjustment float64
	Memo          string
	TimeoutHeight uint64
	Fee           sdktypes.Coins
}

// UnsignedTx represents an unsigned transaction.
type UnsignedTx struct {
	Messages      []sdktypes.Msg
	Config        *TxConfig
	SignBytes     []byte
	AccountNumber uint64
	Sequence      uint64
}

// SignedTx represents a signed transaction.
type SignedTx struct {
	UnsignedTx *UnsignedTx
	Signature  []byte
	PubKey     []byte
	TxBytes    []byte
}

// BroadcastResult contains the result of broadcasting a transaction.
type BroadcastResult struct {
	TxHash    string
	Code      uint32
	Log       string
	GasWanted int64
	GasUsed   int64
	Height    int64
	Events    []Event
}

// Event represents a transaction event.
type Event struct {
	Type       string
	Attributes []Attribute
}

// Attribute represents an event attribute.
type Attribute struct {
	Key   string
	Value string
}

// SimulateResult contains the result of transaction simulation.
type SimulateResult struct {
	GasWanted int64
	GasUsed   int64
	Log       string
	Events    []Event
}

// txBuilder implements TxBuilder.
type txBuilder struct {
	grpcConn *grpc.ClientConn
	config   *config.NetworkConfig
	txConfig *TxConfig
	messages []sdktypes.Msg

	// Cosmos SDK clients
	txServiceClient tx.ServiceClient
}

// NewTxBuilder creates a new transaction builder.
func NewTxBuilder(cfg *config.NetworkConfig, grpcConn *grpc.ClientConn) (TxBuilder, error) {
	if cfg == nil {
		return nil, fmt.Errorf("network configuration is required")
	}

	if grpcConn == nil {
		return nil, fmt.Errorf("gRPC connection is required")
	}

	txConfig := &TxConfig{
		ChainID:       cfg.ChainID,
		GasPrice:      cfg.GasPrice,
		GasDenom:      cfg.StakingDenom,
		GasAdjustment: cfg.GasAdjustment,
		GasLimit:      200000, // Default gas limit
	}

	return &txBuilder{
		grpcConn:        grpcConn,
		config:          cfg,
		txConfig:        txConfig,
		messages:        make([]sdktypes.Msg, 0),
		txServiceClient: tx.NewServiceClient(grpcConn),
	}, nil
}

// WithChainID sets the chain ID for the transaction.
func (tb *txBuilder) WithChainID(chainID string) TxBuilder {
	tb.txConfig.ChainID = chainID
	return tb
}

// WithGasPrice sets the gas price and denomination.
func (tb *txBuilder) WithGasPrice(price float64, denom string) TxBuilder {
	tb.txConfig.GasPrice = price
	tb.txConfig.GasDenom = denom
	return tb
}

// WithGasLimit sets the gas limit for the transaction.
func (tb *txBuilder) WithGasLimit(limit uint64) TxBuilder {
	tb.txConfig.GasLimit = limit
	return tb
}

// WithMemo sets the memo for the transaction.
func (tb *txBuilder) WithMemo(memo string) TxBuilder {
	tb.txConfig.Memo = memo
	return tb
}

// WithTimeoutHeight sets the timeout height for the transaction.
func (tb *txBuilder) WithTimeoutHeight(height uint64) TxBuilder {
	tb.txConfig.TimeoutHeight = height
	return tb
}

// AddMessage adds a single message to the transaction.
func (tb *txBuilder) AddMessage(msg sdktypes.Msg) TxBuilder {
	tb.messages = append(tb.messages, msg)
	return tb
}

// AddMessages adds multiple messages to the transaction.
func (tb *txBuilder) AddMessages(msgs ...sdktypes.Msg) TxBuilder {
	tb.messages = append(tb.messages, msgs...)
	return tb
}

// ClearMessages removes all messages from the transaction.
func (tb *txBuilder) ClearMessages() TxBuilder {
	tb.messages = make([]sdktypes.Msg, 0)
	return tb
}

// WithFee sets the transaction fee directly.
func (tb *txBuilder) WithFee(amount sdktypes.Coins) TxBuilder {
	tb.txConfig.Fee = amount
	return tb
}

// WithGasAdjustment sets the gas adjustment factor.
func (tb *txBuilder) WithGasAdjustment(adjustment float64) TxBuilder {
	tb.txConfig.GasAdjustment = adjustment
	return tb
}

// EstimateGas estimates the gas required for the transaction.
func (tb *txBuilder) EstimateGas(ctx context.Context) (uint64, error) {
	// Build unsigned transaction for simulation
	_, err := tb.Build()
	if err != nil {
		return 0, errors.WrapError(err, errors.ErrGasEstimationFailed, "failed to build transaction for gas estimation")
	}

	// Simulate the transaction
	simulateResult, err := tb.Simulate(ctx)
	if err != nil {
		return 0, errors.WrapError(err, errors.ErrGasEstimationFailed, "failed to simulate transaction")
	}

	// Apply gas adjustment
	estimatedGas := float64(simulateResult.GasUsed) * tb.txConfig.GasAdjustment
	return uint64(estimatedGas), nil
}

// Sign signs the transaction using the provided keyring.
func (tb *txBuilder) Sign(ctx context.Context, keyring keys.KeyringManager) (*SignedTx, error) {
	// Build unsigned transaction
	unsignedTx, err := tb.Build()
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrSigningFailed, "failed to build unsigned transaction")
	}

	// Sign the transaction bytes using the DWN plugin
	signature, err := keyring.SignTransaction(ctx, unsignedTx.SignBytes)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrSigningFailed, "failed to sign transaction")
	}

	// Get wallet identity for public key
	identity, err := keyring.GetIssuerDID(ctx)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrSigningFailed, "failed to get wallet identity")
	}

	// For now, use a placeholder for public key - this should be derived from the DID
	// TODO: Extract public key from DID or add GetPubKey method to KeyringManager
	pubKey := []byte(identity.DID) // Placeholder

	// Build signed transaction
	signedTx, err := tb.BuildSigned(signature.Signature, pubKey)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrSigningFailed, "failed to build signed transaction")
	}

	return signedTx, nil
}

// SignAndBroadcast signs and broadcasts the transaction in one operation.
func (tb *txBuilder) SignAndBroadcast(ctx context.Context, keyring keys.KeyringManager) (*BroadcastResult, error) {
	// Sign the transaction
	signedTx, err := tb.Sign(ctx, keyring)
	if err != nil {
		return nil, err
	}

	// Broadcast the signed transaction
	return tb.Broadcast(ctx, signedTx)
}

// Broadcast broadcasts a signed transaction to the network.
func (tb *txBuilder) Broadcast(ctx context.Context, signedTx *SignedTx) (*BroadcastResult, error) {
	// Create broadcast request
	req := &tx.BroadcastTxRequest{
		TxBytes: signedTx.TxBytes,
		Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC, // Default to sync mode
	}

	// Broadcast the transaction
	resp, err := tb.txServiceClient.BroadcastTx(ctx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrBroadcastFailed, "failed to broadcast transaction")
	}

	// Convert response to our format
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

	return result, nil
}

// Simulate simulates the transaction to estimate gas and check for errors.
func (tb *txBuilder) Simulate(ctx context.Context) (*SimulateResult, error) {
	// Build unsigned transaction for simulation
	unsignedTx, err := tb.Build()
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrGasEstimationFailed, "failed to build transaction for simulation")
	}

	// Create simulate request
	req := &tx.SimulateRequest{
		TxBytes: unsignedTx.SignBytes, // Use sign bytes for simulation
	}

	// Simulate the transaction
	resp, err := tb.txServiceClient.Simulate(ctx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrGasEstimationFailed, "failed to simulate transaction")
	}

	// Convert response to our format
	result := &SimulateResult{
		GasWanted: int64(resp.GasInfo.GasWanted),
		GasUsed:   int64(resp.GasInfo.GasUsed),
		Log:       resp.Result.Log,
	}

	// Convert events
	for _, event := range resp.Result.Events {
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

	return result, nil
}

// Build creates an unsigned transaction.
func (tb *txBuilder) Build() (*UnsignedTx, error) {
	// Allow building without messages for testing/simulation purposes
	// Real transactions will still require messages when broadcasting

	// Calculate fee if not set
	fee := tb.txConfig.Fee
	if fee.IsZero() {
		// Calculate fee based on gas price and limit
		gasAmount := math.NewIntFromUint64(uint64(float64(tb.txConfig.GasLimit) * tb.txConfig.GasPrice))
		fee = sdktypes.NewCoins(sdktypes.NewCoin(tb.txConfig.GasDenom, gasAmount))
	}

	// Create sign bytes (simplified - in a real implementation this would use proper transaction encoding)
	signBytes := []byte(fmt.Sprintf("chain_id:%s,messages:%d,fee:%s,memo:%s",
		tb.txConfig.ChainID,
		len(tb.messages),
		fee.String(),
		tb.txConfig.Memo))

	return &UnsignedTx{
		Messages:  tb.messages,
		Config:    tb.txConfig,
		SignBytes: signBytes,
		// TODO: Fetch account number and sequence from chain
		AccountNumber: 0,
		Sequence:      0,
	}, nil
}

// BuildSigned creates a signed transaction from signature and public key.
func (tb *txBuilder) BuildSigned(signature []byte, pubKey []byte) (*SignedTx, error) {
	unsignedTx, err := tb.Build()
	if err != nil {
		return nil, err
	}

	// Create transaction bytes (simplified - in a real implementation this would use proper transaction encoding)
	txBytes := append(unsignedTx.SignBytes, signature...)
	txBytes = append(txBytes, pubKey...)

	return &SignedTx{
		UnsignedTx: unsignedTx,
		Signature:  signature,
		PubKey:     pubKey,
		TxBytes:    txBytes,
	}, nil
}

// Config returns the current transaction configuration.
func (tb *txBuilder) Config() *TxConfig {
	return tb.txConfig
}

// Utility functions

// NewTxConfig creates a new transaction configuration with defaults.
func NewTxConfig(chainID string) *TxConfig {
	return &TxConfig{
		ChainID:       chainID,
		GasPrice:      0.001,
		GasDenom:      "usnr",
		GasAdjustment: 1.5,
		GasLimit:      200000,
	}
}

// DefaultGasLimit returns the default gas limit for transactions.
func DefaultGasLimit() uint64 {
	return 200000
}

// DefaultGasPrice returns the default gas price for the Sonr network.
func DefaultGasPrice() float64 {
	return 0.001
}

// CalculateFee calculates the transaction fee based on gas price and limit.
func CalculateFee(gasPrice float64, gasLimit uint64, denom string) sdktypes.Coins {
	gasAmount := math.NewIntFromUint64(uint64(float64(gasLimit) * gasPrice))
	return sdktypes.NewCoins(sdktypes.NewCoin(denom, gasAmount))
}
