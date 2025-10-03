package txns

import (
	"context"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

// FeeEstimator interface for estimating transaction fees
type FeeEstimator interface {
	// EstimateFee estimates the fee for a transaction
	EstimateFee(ctx context.Context, params Params) (*FeeEstimation, error)
	// EstimateGas estimates the gas required for a transaction
	EstimateGas(ctx context.Context, params Params) (uint64, error)
	// GetGasPrice retrieves current gas price
	GetGasPrice(ctx context.Context) (any, error)
	// ValidateFee validates if a fee is sufficient
	ValidateFee(fee any, gasUsed uint64) error
}

// CosmosFeeEstimator estimates fees for Cosmos transactions
type CosmosFeeEstimator struct {
	clientCtx     client.Context
	minGasPrice   sdk.DecCoin
	gasAdjustment float64
}

// NewCosmosFeeEstimator creates a new Cosmos fee estimator
func NewCosmosFeeEstimator(clientCtx client.Context, minGasPrice sdk.DecCoin) *CosmosFeeEstimator {
	return &CosmosFeeEstimator{
		clientCtx:     clientCtx,
		minGasPrice:   minGasPrice,
		gasAdjustment: 1.2, // Default gas adjustment factor
	}
}

// SetGasAdjustment sets the gas adjustment factor
func (cfe *CosmosFeeEstimator) SetGasAdjustment(adjustment float64) {
	cfe.gasAdjustment = adjustment
}

// EstimateFee implements FeeEstimator interface
func (cfe *CosmosFeeEstimator) EstimateFee(
	ctx context.Context,
	params Params,
) (*FeeEstimation, error) {
	cosmosParams, ok := params.(*CosmosTransactionParams)
	if !ok {
		return nil, ErrInvalidTransactionParams
	}

	if err := cosmosParams.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Estimate gas usage
	gasLimit, err := cfe.EstimateGas(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Apply gas adjustment
	adjustedGasLimit := uint64(float64(gasLimit) * cfe.gasAdjustment)

	// Calculate fee
	feeAmount := cfe.minGasPrice.Amount.MulInt64(int64(adjustedGasLimit))
	fee := sdk.NewCoins(sdk.NewCoin(cfe.minGasPrice.Denom, feeAmount.TruncateInt()))

	return &FeeEstimation{
		GasLimit: adjustedGasLimit,
		GasPrice: cfe.minGasPrice,
		Fee:      fee,
		Total:    fee.String(),
	}, nil
}

// EstimateGas implements FeeEstimator interface
func (cfe *CosmosFeeEstimator) EstimateGas(ctx context.Context, params Params) (uint64, error) {
	cosmosParams, ok := params.(*CosmosTransactionParams)
	if !ok {
		return 0, ErrInvalidTransactionParams
	}

	// Base gas estimation by message type
	baseGas := uint64(0)
	for _, msg := range cosmosParams.Messages {
		baseGas += cfe.estimateGasForMessage(msg)
	}

	// Add overhead for transaction processing
	overhead := uint64(10000) // Base transaction overhead
	if cosmosParams.Memo != "" {
		overhead += uint64(len(cosmosParams.Memo)) * 10 // Memo overhead
	}

	return baseGas + overhead, nil
}

// GetGasPrice implements FeeEstimator interface
func (cfe *CosmosFeeEstimator) GetGasPrice(ctx context.Context) (any, error) {
	// In Cosmos, gas price is typically fixed or queried from chain parameters
	// For now, return the configured minimum gas price
	return cfe.minGasPrice, nil
}

// ValidateFee implements FeeEstimator interface
func (cfe *CosmosFeeEstimator) ValidateFee(fee any, gasUsed uint64) error {
	feeCoins, ok := fee.(sdk.Coins)
	if !ok {
		return ErrInvalidTransactionParams
	}

	// Calculate minimum required fee
	minFeeAmount := cfe.minGasPrice.Amount.MulInt64(int64(gasUsed))
	minFee := sdk.NewCoins(sdk.NewCoin(cfe.minGasPrice.Denom, minFeeAmount.TruncateInt()))

	// Check if provided fee is sufficient
	if !feeCoins.IsAllGTE(minFee) {
		return fmt.Errorf("insufficient fee: got %s, need at least %s", feeCoins, minFee)
	}

	return nil
}

// estimateGasForMessage estimates gas usage for a specific message type
func (cfe *CosmosFeeEstimator) estimateGasForMessage(msg sdk.Msg) uint64 {
	switch msg.(type) {
	case *banktypes.MsgSend:
		return 80000 // Base gas for bank send
	case *banktypes.MsgMultiSend:
		return 120000 // Higher gas for multi-send
	default:
		return 100000 // Default gas estimate
	}
}

// EVMFeeEstimator estimates fees for EVM transactions
type EVMFeeEstimator struct {
	client        *ethclient.Client
	chainID       *big.Int
	gasAdjustment float64
}

// NewEVMFeeEstimator creates a new EVM fee estimator
func NewEVMFeeEstimator(client *ethclient.Client, chainID *big.Int) *EVMFeeEstimator {
	return &EVMFeeEstimator{
		client:        client,
		chainID:       chainID,
		gasAdjustment: 1.1, // Default gas adjustment factor
	}
}

// SetGasAdjustment sets the gas adjustment factor
func (efe *EVMFeeEstimator) SetGasAdjustment(adjustment float64) {
	efe.gasAdjustment = adjustment
}

// EstimateFee implements FeeEstimator interface
func (efe *EVMFeeEstimator) EstimateFee(
	ctx context.Context,
	params Params,
) (*FeeEstimation, error) {
	evmParams, ok := params.(*EVMTransactionParams)
	if !ok {
		return nil, ErrInvalidTransactionParams
	}

	if err := evmParams.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Estimate gas usage
	gasLimit, err := efe.EstimateGas(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Get current gas price
	gasPrice, err := efe.GetGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Apply gas adjustment
	adjustedGasLimit := uint64(float64(gasLimit) * efe.gasAdjustment)

	// Calculate fee
	var totalFee *big.Int
	var feeData any

	if evmParams.MaxFeePerGas != nil && evmParams.MaxPriorityFeePerGas != nil {
		// EIP-1559 transaction
		totalFee = new(big.Int).Mul(evmParams.MaxFeePerGas, big.NewInt(int64(adjustedGasLimit)))
		feeData = map[string]*big.Int{
			"maxFeePerGas":         evmParams.MaxFeePerGas,
			"maxPriorityFeePerGas": evmParams.MaxPriorityFeePerGas,
		}
	} else {
		// Legacy transaction
		gasPriceBig := gasPrice.(*big.Int)
		totalFee = new(big.Int).Mul(gasPriceBig, big.NewInt(int64(adjustedGasLimit)))
		feeData = gasPriceBig
	}

	return &FeeEstimation{
		GasLimit: adjustedGasLimit,
		GasPrice: feeData,
		Fee:      totalFee,
		Total:    totalFee.String(),
	}, nil
}

// EstimateGas implements FeeEstimator interface
func (efe *EVMFeeEstimator) EstimateGas(ctx context.Context, params Params) (uint64, error) {
	evmParams, ok := params.(*EVMTransactionParams)
	if !ok {
		return 0, ErrInvalidTransactionParams
	}

	if efe.client == nil {
		// Fallback estimation without client
		return efe.estimateGasOffline(evmParams), nil
	}

	// Create a call message for gas estimation
	callMsg := ethereum.CallMsg{
		To:    evmParams.To,
		Value: evmParams.Value,
		Data:  evmParams.Data,
	}

	// Estimate gas using the client
	gasLimit, err := efe.client.EstimateGas(ctx, callMsg)
	if err != nil {
		// Fallback to offline estimation
		return efe.estimateGasOffline(evmParams), nil
	}

	return gasLimit, nil
}

// GetGasPrice implements FeeEstimator interface
func (efe *EVMFeeEstimator) GetGasPrice(ctx context.Context) (any, error) {
	if efe.client == nil {
		// Return default gas price
		return big.NewInt(params.GWei * 20), nil // 20 Gwei
	}

	gasPrice, err := efe.client.SuggestGasPrice(ctx)
	if err != nil {
		// Fallback to default
		return big.NewInt(params.GWei * 20), nil
	}

	return gasPrice, nil
}

// ValidateFee implements FeeEstimator interface
func (efe *EVMFeeEstimator) ValidateFee(fee any, gasUsed uint64) error {
	switch f := fee.(type) {
	case *big.Int:
		// Legacy transaction
		minFee := new(big.Int).Mul(big.NewInt(params.GWei), big.NewInt(int64(gasUsed)))
		if f.Cmp(minFee) < 0 {
			return fmt.Errorf("insufficient fee: got %s, need at least %s", f, minFee)
		}
	case map[string]*big.Int:
		// EIP-1559 transaction
		maxFeePerGas, ok := f["maxFeePerGas"]
		if !ok {
			return fmt.Errorf("missing maxFeePerGas in fee data")
		}
		minFee := new(big.Int).Mul(big.NewInt(params.GWei), big.NewInt(int64(gasUsed)))
		totalMaxFee := new(big.Int).Mul(maxFeePerGas, big.NewInt(int64(gasUsed)))
		if totalMaxFee.Cmp(minFee) < 0 {
			return fmt.Errorf("insufficient max fee: got %s, need at least %s", totalMaxFee, minFee)
		}
	default:
		return fmt.Errorf("unsupported fee type: %T", fee)
	}

	return nil
}

// estimateGasOffline provides offline gas estimation
func (efe *EVMFeeEstimator) estimateGasOffline(params *EVMTransactionParams) uint64 {
	baseGas := uint64(21000) // Base transaction gas

	if params.Data != nil && len(params.Data) > 0 {
		// Contract interaction
		baseGas += uint64(len(params.Data)) * 16 // Rough estimate for data
		if params.To == nil {
			// Contract deployment
			baseGas += 200000
		} else {
			// Contract call
			baseGas += 100000
		}
	}

	return baseGas
}

// FeeManager manages fee estimation for multiple transaction types
type FeeManager struct {
	cosmosEstimator *CosmosFeeEstimator
	evmEstimator    *EVMFeeEstimator
}

// NewFeeManager creates a new fee manager
func NewFeeManager(cosmosEstimator *CosmosFeeEstimator, evmEstimator *EVMFeeEstimator) *FeeManager {
	return &FeeManager{
		cosmosEstimator: cosmosEstimator,
		evmEstimator:    evmEstimator,
	}
}

// EstimateFee estimates fee for any transaction type
func (fm *FeeManager) EstimateFee(
	ctx context.Context,
	txType TransactionType,
	params Params,
) (*FeeEstimation, error) {
	switch txType {
	case TransactionTypeCosmos:
		if fm.cosmosEstimator == nil {
			return nil, fmt.Errorf("cosmos fee estimator not configured")
		}
		return fm.cosmosEstimator.EstimateFee(ctx, params)
	case TransactionTypeEVM:
		if fm.evmEstimator == nil {
			return nil, fmt.Errorf("EVM fee estimator not configured")
		}
		return fm.evmEstimator.EstimateFee(ctx, params)
	default:
		return nil, ErrUnsupportedChainType
	}
}

// GetEstimator returns the appropriate fee estimator for a transaction type
func (fm *FeeManager) GetEstimator(txType TransactionType) (FeeEstimator, error) {
	switch txType {
	case TransactionTypeCosmos:
		if fm.cosmosEstimator == nil {
			return nil, fmt.Errorf("cosmos fee estimator not configured")
		}
		return fm.cosmosEstimator, nil
	case TransactionTypeEVM:
		if fm.evmEstimator == nil {
			return nil, fmt.Errorf("EVM fee estimator not configured")
		}
		return fm.evmEstimator, nil
	default:
		return nil, ErrUnsupportedChainType
	}
}

// DefaultFeeConfig holds default fee configuration
type DefaultFeeConfig struct {
	CosmosMinGasPrice sdk.DecCoin
	EVMGasPrice       *big.Int
	GasAdjustment     float64
}

// GetDefaultFeeConfig returns default fee configuration
func GetDefaultFeeConfig() *DefaultFeeConfig {
	return &DefaultFeeConfig{
		CosmosMinGasPrice: sdk.NewDecCoin("usnr", math.NewInt(1000)),
		EVMGasPrice:       big.NewInt(params.GWei * 20), // 20 Gwei
		GasAdjustment:     1.2,
	}
}

// CreateDefaultFeeManager creates a fee manager with default configuration
func CreateDefaultFeeManager(
	clientCtx client.Context,
	evmClient *ethclient.Client,
	chainID *big.Int,
) *FeeManager {
	config := GetDefaultFeeConfig()

	cosmosEstimator := NewCosmosFeeEstimator(clientCtx, config.CosmosMinGasPrice)
	cosmosEstimator.SetGasAdjustment(config.GasAdjustment)

	evmEstimator := NewEVMFeeEstimator(evmClient, chainID)
	evmEstimator.SetGasAdjustment(config.GasAdjustment)

	return NewFeeManager(cosmosEstimator, evmEstimator)
}
