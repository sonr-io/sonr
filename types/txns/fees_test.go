package txns

import (
	"context"
	"math/big"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestClientCtx() client.Context {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)

	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(marshaler, authtx.DefaultSignModes)

	return client.Context{}.
		WithCodec(marshaler).
		WithTxConfig(txConfig).
		WithInterfaceRegistry(interfaceRegistry)
}

func TestNewCosmosFeeEstimator(t *testing.T) {
	clientCtx := setupTestClientCtx()
	minGasPrice := sdk.NewDecCoin("usnr", math.NewInt(1000))

	estimator := NewCosmosFeeEstimator(clientCtx, minGasPrice)
	require.NotNil(t, estimator)

	assert.Equal(t, clientCtx, estimator.clientCtx)
	assert.Equal(t, minGasPrice, estimator.minGasPrice)
	assert.Equal(t, 1.2, estimator.gasAdjustment)
}

func TestCosmosFeeEstimator_SetGasAdjustment(t *testing.T) {
	clientCtx := setupTestClientCtx()
	minGasPrice := sdk.NewDecCoin("usnr", math.NewInt(1000))
	estimator := NewCosmosFeeEstimator(clientCtx, minGasPrice)

	estimator.SetGasAdjustment(1.5)
	assert.Equal(t, 1.5, estimator.gasAdjustment)
}

func TestCosmosFeeEstimator_EstimateGas(t *testing.T) {
	clientCtx := setupTestClientCtx()
	minGasPrice := sdk.NewDecCoin("usnr", math.NewInt(1000))
	estimator := NewCosmosFeeEstimator(clientCtx, minGasPrice)

	// Test with bank send message
	params := &CosmosTransactionParams{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{
				FromAddress: "snr1test",
				ToAddress:   "snr1test2",
				Amount:      sdk.NewCoins(sdk.NewCoin("usnr", math.NewInt(1000))),
			},
		},
		Memo: "test memo",
	}

	gasLimit, err := estimator.EstimateGas(context.Background(), params)
	require.NoError(t, err)

	// Should be base gas (80000) + overhead (10000) + memo overhead (len("test memo") * 10 = 90)
	expectedGas := uint64(80000 + 10000 + 90)
	assert.Equal(t, expectedGas, gasLimit)
}

func TestCosmosFeeEstimator_EstimateFee(t *testing.T) {
	clientCtx := setupTestClientCtx()
	minGasPrice := sdk.NewDecCoin("usnr", math.NewInt(1000))
	estimator := NewCosmosFeeEstimator(clientCtx, minGasPrice)

	params := &CosmosTransactionParams{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{
				FromAddress: "snr1test",
				ToAddress:   "snr1test2",
				Amount:      sdk.NewCoins(sdk.NewCoin("usnr", math.NewInt(1000))),
			},
		},
		GasLimit: 200000,
	}

	feeEst, err := estimator.EstimateFee(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, feeEst)

	// Check fee estimation structure
	assert.Greater(t, feeEst.GasLimit, uint64(0))
	assert.NotNil(t, feeEst.GasPrice)
	assert.NotNil(t, feeEst.Fee)
	assert.NotEmpty(t, feeEst.Total)

	// Check that gas adjustment was applied (gas limit should be > base gas)
	assert.Greater(t, feeEst.GasLimit, uint64(90000))      // Should be more than base gas
	assert.LessOrEqual(t, feeEst.GasLimit, uint64(250000)) // But reasonable
}

func TestCosmosFeeEstimator_GetGasPrice(t *testing.T) {
	clientCtx := setupTestClientCtx()
	minGasPrice := sdk.NewDecCoin("usnr", math.NewInt(1000))
	estimator := NewCosmosFeeEstimator(clientCtx, minGasPrice)

	gasPrice, err := estimator.GetGasPrice(context.Background())
	require.NoError(t, err)

	assert.Equal(t, minGasPrice, gasPrice)
}

func TestCosmosFeeEstimator_ValidateFee(t *testing.T) {
	clientCtx := setupTestClientCtx()
	minGasPrice := sdk.NewDecCoin("usnr", math.NewInt(1000))
	estimator := NewCosmosFeeEstimator(clientCtx, minGasPrice)

	gasUsed := uint64(100000)

	// Test sufficient fee
	sufficientFee := sdk.NewCoins(sdk.NewCoin("usnr", math.NewInt(100000000))) // 100000 * 1000
	err := estimator.ValidateFee(sufficientFee, gasUsed)
	assert.NoError(t, err)

	// Test insufficient fee
	insufficientFee := sdk.NewCoins(sdk.NewCoin("usnr", math.NewInt(50000)))
	err = estimator.ValidateFee(insufficientFee, gasUsed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient fee")

	// Test invalid fee type
	err = estimator.ValidateFee("invalid", gasUsed)
	assert.Error(t, err)
}

func TestNewEVMFeeEstimator(t *testing.T) {
	chainID := big.NewInt(1)
	estimator := NewEVMFeeEstimator(nil, chainID)
	require.NotNil(t, estimator)

	assert.Equal(t, chainID, estimator.chainID)
	assert.Equal(t, 1.1, estimator.gasAdjustment)
	assert.Nil(t, estimator.client)
}

func TestEVMFeeEstimator_EstimateGas(t *testing.T) {
	chainID := big.NewInt(1)
	estimator := NewEVMFeeEstimator(nil, chainID) // No client for offline estimation

	// Test simple transfer
	toAddr := common.HexToAddress("0x742d35Cc6634C0532925a3b8D80C")
	params := &EVMTransactionParams{
		To:    &toAddr,
		Value: big.NewInt(1000000000000000000),
	}

	gasLimit, err := estimator.EstimateGas(context.Background(), params)
	require.NoError(t, err)
	assert.Equal(t, uint64(21000), gasLimit) // Base gas for transfer

	// Test contract call
	params.Data = []byte("contract call data")
	gasLimit, err = estimator.EstimateGas(context.Background(), params)
	require.NoError(t, err)
	assert.Greater(t, gasLimit, uint64(21000)) // Should be more than base gas

	// Test contract deployment
	params.To = nil
	gasLimit, err = estimator.EstimateGas(context.Background(), params)
	require.NoError(t, err)
	assert.Greater(t, gasLimit, uint64(200000)) // Should include deployment gas
}

func TestEVMFeeEstimator_EstimateFee(t *testing.T) {
	chainID := big.NewInt(1)
	estimator := NewEVMFeeEstimator(nil, chainID)

	toAddr := common.HexToAddress("0x742d35Cc6634C0532925a3b8D80C")

	// Test legacy transaction
	params := &EVMTransactionParams{
		To:       &toAddr,
		Value:    big.NewInt(1000000000000000000),
		GasLimit: 21000,
		GasPrice: big.NewInt(20000000000), // 20 Gwei
		ChainID:  big.NewInt(1),
	}

	feeEst, err := estimator.EstimateFee(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, feeEst)

	// Check fee estimation
	expectedGas := uint64(23100) // 21000 * 1.1 with adjustment
	assert.Equal(t, expectedGas, feeEst.GasLimit)
	assert.NotNil(t, feeEst.GasPrice)
	assert.NotNil(t, feeEst.Fee)

	// Test EIP-1559 transaction
	params.MaxFeePerGas = big.NewInt(30000000000)
	params.MaxPriorityFeePerGas = big.NewInt(2000000000)

	feeEst1559, err := estimator.EstimateFee(context.Background(), params)
	require.NoError(t, err)
	require.NotNil(t, feeEst1559)

	assert.Equal(t, expectedGas, feeEst1559.GasLimit)

	// Check that fee data contains EIP-1559 fields
	feeData, ok := feeEst1559.GasPrice.(map[string]*big.Int)
	require.True(t, ok)
	assert.Contains(t, feeData, "maxFeePerGas")
	assert.Contains(t, feeData, "maxPriorityFeePerGas")
}

func TestEVMFeeEstimator_GetGasPrice(t *testing.T) {
	chainID := big.NewInt(1)
	estimator := NewEVMFeeEstimator(nil, chainID)

	gasPrice, err := estimator.GetGasPrice(context.Background())
	require.NoError(t, err)

	expectedPrice := big.NewInt(params.GWei * 20) // 20 Gwei default
	assert.Equal(t, expectedPrice, gasPrice)
}

func TestEVMFeeEstimator_ValidateFee(t *testing.T) {
	chainID := big.NewInt(1)
	estimator := NewEVMFeeEstimator(nil, chainID)

	gasUsed := uint64(21000)

	// Test legacy transaction with sufficient fee
	sufficientFee := big.NewInt(21000000000000000) // 21000 * 1 Gwei
	err := estimator.ValidateFee(sufficientFee, gasUsed)
	assert.NoError(t, err)

	// Test legacy transaction with insufficient fee
	insufficientFee := big.NewInt(1000)
	err = estimator.ValidateFee(insufficientFee, gasUsed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient fee")

	// Test EIP-1559 transaction
	eip1559Fee := map[string]*big.Int{
		"maxFeePerGas":         big.NewInt(2000000000), // 2 Gwei
		"maxPriorityFeePerGas": big.NewInt(1000000000), // 1 Gwei
	}
	err = estimator.ValidateFee(eip1559Fee, gasUsed)
	assert.NoError(t, err)

	// Test invalid fee type
	err = estimator.ValidateFee("invalid", gasUsed)
	assert.Error(t, err)
}

func TestFeeManager(t *testing.T) {
	clientCtx := setupTestClientCtx()
	minGasPrice := sdk.NewDecCoin("usnr", math.NewInt(1000))
	cosmosEstimator := NewCosmosFeeEstimator(clientCtx, minGasPrice)

	chainID := big.NewInt(1)
	evmEstimator := NewEVMFeeEstimator(nil, chainID)

	manager := NewFeeManager(cosmosEstimator, evmEstimator)
	require.NotNil(t, manager)

	// Test Cosmos fee estimation
	cosmosParams := &CosmosTransactionParams{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{
				FromAddress: "snr1test",
				ToAddress:   "snr1test2",
				Amount:      sdk.NewCoins(sdk.NewCoin("usnr", math.NewInt(1000))),
			},
		},
		GasLimit: 200000,
	}

	cosmosFee, err := manager.EstimateFee(context.Background(), TransactionTypeCosmos, cosmosParams)
	require.NoError(t, err)
	assert.NotNil(t, cosmosFee)

	// Test EVM fee estimation
	toAddr := common.HexToAddress("0x742d35Cc6634C0532925a3b8D80C")
	evmParams := &EVMTransactionParams{
		To:       &toAddr,
		Value:    big.NewInt(1000000000000000000),
		GasLimit: 21000,
		GasPrice: big.NewInt(20000000000),
		ChainID:  big.NewInt(1),
	}

	evmFee, err := manager.EstimateFee(context.Background(), TransactionTypeEVM, evmParams)
	require.NoError(t, err)
	assert.NotNil(t, evmFee)

	// Test unsupported transaction type
	_, err = manager.EstimateFee(context.Background(), TransactionTypeUnknown, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrUnsupportedChainType, err)
}

func TestFeeManager_GetEstimator(t *testing.T) {
	clientCtx := setupTestClientCtx()
	minGasPrice := sdk.NewDecCoin("usnr", math.NewInt(1000))
	cosmosEstimator := NewCosmosFeeEstimator(clientCtx, minGasPrice)

	chainID := big.NewInt(1)
	evmEstimator := NewEVMFeeEstimator(nil, chainID)

	manager := NewFeeManager(cosmosEstimator, evmEstimator)

	// Test getting Cosmos estimator
	cosmosEst, err := manager.GetEstimator(TransactionTypeCosmos)
	require.NoError(t, err)
	assert.Equal(t, cosmosEstimator, cosmosEst)

	// Test getting EVM estimator
	evmEst, err := manager.GetEstimator(TransactionTypeEVM)
	require.NoError(t, err)
	assert.Equal(t, evmEstimator, evmEst)

	// Test unsupported type
	_, err = manager.GetEstimator(TransactionTypeUnknown)
	assert.Error(t, err)
}

func TestGetDefaultFeeConfig(t *testing.T) {
	config := GetDefaultFeeConfig()
	require.NotNil(t, config)

	assert.Equal(t, "usnr", config.CosmosMinGasPrice.Denom)
	expectedDec := math.LegacyNewDecFromInt(math.NewInt(1000))
	assert.True(t, config.CosmosMinGasPrice.Amount.Equal(expectedDec))
	assert.Equal(t, big.NewInt(params.GWei*20), config.EVMGasPrice)
	assert.Equal(t, 1.2, config.GasAdjustment)
}

func TestCreateDefaultFeeManager(t *testing.T) {
	clientCtx := setupTestClientCtx()
	chainID := big.NewInt(1)

	manager := CreateDefaultFeeManager(clientCtx, nil, chainID)
	require.NotNil(t, manager)
	require.NotNil(t, manager.cosmosEstimator)
	require.NotNil(t, manager.evmEstimator)

	// Test that estimators work
	cosmosEst, err := manager.GetEstimator(TransactionTypeCosmos)
	require.NoError(t, err)
	assert.NotNil(t, cosmosEst)

	evmEst, err := manager.GetEstimator(TransactionTypeEVM)
	require.NoError(t, err)
	assert.NotNil(t, evmEst)
}

// Benchmark tests
func BenchmarkCosmosFeeEstimator_EstimateGas(b *testing.B) {
	clientCtx := setupTestClientCtx()
	minGasPrice := sdk.NewDecCoin("usnr", math.NewInt(1000))
	estimator := NewCosmosFeeEstimator(clientCtx, minGasPrice)

	params := &CosmosTransactionParams{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{
				FromAddress: "snr1test",
				ToAddress:   "snr1test2",
				Amount:      sdk.NewCoins(sdk.NewCoin("usnr", math.NewInt(1000))),
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = estimator.EstimateGas(context.Background(), params)
	}
}

func BenchmarkEVMFeeEstimator_EstimateGas(b *testing.B) {
	chainID := big.NewInt(1)
	estimator := NewEVMFeeEstimator(nil, chainID)

	toAddr := common.HexToAddress("0x742d35Cc6634C0532925a3b8D80C")
	params := &EVMTransactionParams{
		To:    &toAddr,
		Value: big.NewInt(1000000000000000000),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = estimator.EstimateGas(context.Background(), params)
	}
}
