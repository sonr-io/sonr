// Package tx provides gas estimation and fee calculation utilities for the Sonr client SDK.
package tx

import (
	"context"
	"fmt"
	"math"

	"google.golang.org/grpc"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/sonr-io/sonr/client/config"
	"github.com/sonr-io/sonr/client/errors"
)

// GasEstimator provides an interface for estimating gas costs and calculating fees.
type GasEstimator interface {
	// Gas estimation
	EstimateGas(ctx context.Context, msgs []sdktypes.Msg) (*GasEstimate, error)
	EstimateGasForTx(ctx context.Context, unsignedTx *UnsignedTx) (*GasEstimate, error)

	// Fee calculation
	CalculateFee(gasUsed uint64, gasPrice float64, denom string) sdktypes.Coins
	CalculateFeeWithAdjustment(gasUsed uint64, gasPrice float64, adjustment float64, denom string) sdktypes.Coins

	// Gas configuration
	WithGasAdjustment(adjustment float64) GasEstimator
	WithMinGasPrice(price float64) GasEstimator
	WithMaxGasLimit(limit uint64) GasEstimator

	// Utility methods
	GetRecommendedGasPrice(ctx context.Context) (float64, error)
	GetNetworkGasInfo(ctx context.Context) (*NetworkGasInfo, error)
}

// GasEstimate contains the result of gas estimation.
type GasEstimate struct {
	GasWanted     uint64         // Estimated gas needed
	GasUsed       uint64         // Gas used in simulation
	GasLimit      uint64         // Recommended gas limit (with adjustment)
	Fee           sdktypes.Coins // Calculated fee
	GasPrice      float64        // Gas price used
	GasAdjustment float64        // Adjustment factor applied
}

// NetworkGasInfo contains network-wide gas information.
type NetworkGasInfo struct {
	MinGasPrice         float64 // Minimum gas price accepted by validators
	MedianGasPrice      float64 // Median gas price from recent transactions
	RecommendedGasPrice float64 // Recommended gas price for fast inclusion
	MaxGasLimit         uint64  // Maximum gas limit per transaction
}

// GasConfig holds gas estimation configuration.
type GasConfig struct {
	Adjustment  float64 // Gas adjustment factor (default: 1.5)
	MinGasPrice float64 // Minimum gas price
	MaxGasLimit uint64  // Maximum gas limit
	Denom       string  // Gas fee denomination
}

// gasEstimator implements GasEstimator.
type gasEstimator struct {
	grpcConn        *grpc.ClientConn
	config          *config.NetworkConfig
	txServiceClient tx.ServiceClient
	gasConfig       GasConfig
}

// NewGasEstimator creates a new gas estimator.
func NewGasEstimator(grpcConn *grpc.ClientConn, cfg *config.NetworkConfig) GasEstimator {
	gasConfig := GasConfig{
		Adjustment:  cfg.GasAdjustment,
		MinGasPrice: cfg.GasPrice,
		MaxGasLimit: 10000000, // 10M gas limit
		Denom:       cfg.StakingDenom,
	}

	return &gasEstimator{
		grpcConn:        grpcConn,
		config:          cfg,
		txServiceClient: tx.NewServiceClient(grpcConn),
		gasConfig:       gasConfig,
	}
}

// EstimateGas estimates gas for a list of messages.
func (ge *gasEstimator) EstimateGas(ctx context.Context, msgs []sdktypes.Msg) (*GasEstimate, error) {
	if len(msgs) == 0 {
		return nil, fmt.Errorf("no messages provided for gas estimation")
	}

	// Create a temporary transaction builder to build the transaction for simulation
	builder, err := NewTxBuilder(ge.config, ge.grpcConn)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrGasEstimationFailed, "failed to create transaction builder")
	}

	// Add messages and build unsigned transaction
	for _, msg := range msgs {
		builder.AddMessage(msg)
	}

	unsignedTx, err := builder.Build()
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrGasEstimationFailed, "failed to build transaction for estimation")
	}

	return ge.EstimateGasForTx(ctx, unsignedTx)
}

// EstimateGasForTx estimates gas for an unsigned transaction.
func (ge *gasEstimator) EstimateGasForTx(ctx context.Context, unsignedTx *UnsignedTx) (*GasEstimate, error) {
	// Create simulate request
	req := &tx.SimulateRequest{
		TxBytes: unsignedTx.SignBytes, // Use sign bytes for simulation
	}

	// Simulate the transaction
	resp, err := ge.txServiceClient.Simulate(ctx, req)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrGasEstimationFailed, "failed to simulate transaction")
	}

	gasUsed := resp.GasInfo.GasUsed
	gasWanted := resp.GasInfo.GasWanted

	// Apply gas adjustment
	gasLimit := uint64(float64(gasUsed) * ge.gasConfig.Adjustment)

	// Ensure gas limit doesn't exceed maximum
	if gasLimit > ge.gasConfig.MaxGasLimit {
		gasLimit = ge.gasConfig.MaxGasLimit
	}

	// Calculate fee
	fee := ge.CalculateFee(gasLimit, ge.gasConfig.MinGasPrice, ge.gasConfig.Denom)

	return &GasEstimate{
		GasWanted:     gasWanted,
		GasUsed:       gasUsed,
		GasLimit:      gasLimit,
		Fee:           fee,
		GasPrice:      ge.gasConfig.MinGasPrice,
		GasAdjustment: ge.gasConfig.Adjustment,
	}, nil
}

// CalculateFee calculates the transaction fee based on gas usage and price.
func (ge *gasEstimator) CalculateFee(gasUsed uint64, gasPrice float64, denom string) sdktypes.Coins {
	// Calculate fee amount
	feeAmount := math.Ceil(float64(gasUsed) * gasPrice)

	// Create coin
	feeCoin := sdktypes.NewInt64Coin(denom, int64(feeAmount))

	return sdktypes.NewCoins(feeCoin)
}

// CalculateFeeWithAdjustment calculates fee with a custom gas adjustment.
func (ge *gasEstimator) CalculateFeeWithAdjustment(gasUsed uint64, gasPrice float64, adjustment float64, denom string) sdktypes.Coins {
	adjustedGas := uint64(float64(gasUsed) * adjustment)
	return ge.CalculateFee(adjustedGas, gasPrice, denom)
}

// WithGasAdjustment sets the gas adjustment factor.
func (ge *gasEstimator) WithGasAdjustment(adjustment float64) GasEstimator {
	ge.gasConfig.Adjustment = adjustment
	return ge
}

// WithMinGasPrice sets the minimum gas price.
func (ge *gasEstimator) WithMinGasPrice(price float64) GasEstimator {
	ge.gasConfig.MinGasPrice = price
	return ge
}

// WithMaxGasLimit sets the maximum gas limit.
func (ge *gasEstimator) WithMaxGasLimit(limit uint64) GasEstimator {
	ge.gasConfig.MaxGasLimit = limit
	return ge
}

// GetRecommendedGasPrice returns the recommended gas price for the network.
func (ge *gasEstimator) GetRecommendedGasPrice(ctx context.Context) (float64, error) {
	// TODO: Implement dynamic gas price discovery based on network conditions
	// Should query recent transactions to analyze gas price trends
	// Calculate percentile-based recommendations (e.g., 25th, 50th, 75th)
	// Consider network congestion and validator preferences
	// Return optimal gas price for desired transaction inclusion speed
	return ge.gasConfig.MinGasPrice, nil
}

// GetNetworkGasInfo retrieves network-wide gas information.
func (ge *gasEstimator) GetNetworkGasInfo(ctx context.Context) (*NetworkGasInfo, error) {
	// TODO: Implement dynamic network gas info retrieval
	// Should query validator minimum gas prices via gRPC
	// Analyze recent block gas usage patterns and limits
	// Calculate median and recommended gas prices from mempool
	// Monitor network congestion metrics for pricing recommendations
	// Query chain parameters for maximum gas limits and constraints
	return &NetworkGasInfo{
		MinGasPrice:         ge.gasConfig.MinGasPrice,
		MedianGasPrice:      ge.gasConfig.MinGasPrice,
		RecommendedGasPrice: ge.gasConfig.MinGasPrice,
		MaxGasLimit:         ge.gasConfig.MaxGasLimit,
	}, nil
}

// Utility functions and constants

// Default gas values for different transaction types
const (
	// DefaultGasLimitValue is the default gas limit for transactions
	DefaultGasLimitValue = 200000

	// SendGasLimit is the typical gas limit for send transactions
	SendGasLimit = 100000

	// DelegateGasLimit is the typical gas limit for delegation transactions
	DelegateGasLimit = 150000

	// ContractCallGasLimit is the typical gas limit for smart contract calls
	ContractCallGasLimit = 500000

	// MinGasAdjustment is the minimum recommended gas adjustment
	MinGasAdjustment = 1.1

	// MaxGasAdjustment is the maximum reasonable gas adjustment
	MaxGasAdjustment = 3.0
)

// GasLimitForMessageType returns a recommended gas limit for different message types.
func GasLimitForMessageType(msgType string) uint64 {
	switch msgType {
	case "/cosmos.bank.v1beta1.MsgSend":
		return SendGasLimit
	case "/cosmos.staking.v1beta1.MsgDelegate":
		return DelegateGasLimit
	case "/cosmos.staking.v1beta1.MsgUndelegate":
		return DelegateGasLimit
	case "/cosmos.staking.v1beta1.MsgRedelegate":
		return DelegateGasLimit * 2
	default:
		return DefaultGasLimitValue
	}
}

// EstimateGasForMessages provides a quick gas estimate based on message types.
func EstimateGasForMessages(msgs []sdktypes.Msg) uint64 {
	var totalGas uint64

	for _, msg := range msgs {
		msgType := sdktypes.MsgTypeURL(msg)
		gas := GasLimitForMessageType(msgType)
		totalGas += gas
	}

	// Add base transaction overhead
	totalGas += 50000

	return totalGas
}

// ValidateGasPrice checks if a gas price is reasonable.
func ValidateGasPrice(gasPrice float64) error {
	if gasPrice <= 0 {
		return fmt.Errorf("gas price must be positive")
	}

	if gasPrice > 1.0 { // 1 SNR per gas unit seems excessive
		return fmt.Errorf("gas price %f seems too high", gasPrice)
	}

	return nil
}

// ValidateGasLimit checks if a gas limit is reasonable.
func ValidateGasLimit(gasLimit uint64) error {
	if gasLimit == 0 {
		return fmt.Errorf("gas limit must be positive")
	}

	if gasLimit > 50000000 { // 50M gas limit seems excessive
		return fmt.Errorf("gas limit %d seems too high", gasLimit)
	}

	return nil
}

// OptimizeGasConfig optimizes gas configuration based on network conditions.
func OptimizeGasConfig(config *GasConfig, networkInfo *NetworkGasInfo) *GasConfig {
	optimized := *config

	// Use recommended gas price if it's higher than our minimum
	if networkInfo.RecommendedGasPrice > config.MinGasPrice {
		optimized.MinGasPrice = networkInfo.RecommendedGasPrice
	}

	// Ensure gas adjustment is within reasonable bounds
	if optimized.Adjustment < MinGasAdjustment {
		optimized.Adjustment = MinGasAdjustment
	}
	if optimized.Adjustment > MaxGasAdjustment {
		optimized.Adjustment = MaxGasAdjustment
	}

	// Use network max gas limit if it's lower than our configured max
	if networkInfo.MaxGasLimit > 0 && networkInfo.MaxGasLimit < config.MaxGasLimit {
		optimized.MaxGasLimit = networkInfo.MaxGasLimit
	}

	return &optimized
}
