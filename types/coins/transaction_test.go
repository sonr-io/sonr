package coins

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEthereumTransactionBuilder(t *testing.T) {
	chainID := big.NewInt(1)
	builder := NewEthereumTransactionBuilder(chainID)

	assert.NotNil(t, builder)
	assert.Equal(t, chainID, builder.chainID)
	assert.Equal(t, uint64(21000), builder.gasLimit)
	assert.Equal(t, big.NewInt(20000000000), builder.gasPrice)
	assert.Equal(t, uint64(0), builder.nonce)
}

func TestEthereumTransactionBuilderSetGas(t *testing.T) {
	builder := NewEthereumTransactionBuilder(big.NewInt(1))

	gasLimit := uint64(50000)
	gasPrice := big.NewInt(30000000000)

	result := builder.SetGas(gasLimit, gasPrice)

	assert.Equal(t, builder, result) // Should return self for chaining
	assert.Equal(t, gasLimit, builder.gasLimit)
	assert.Equal(t, gasPrice, builder.gasPrice)
}

func TestEthereumTransactionBuilderSetNonce(t *testing.T) {
	builder := NewEthereumTransactionBuilder(big.NewInt(1))

	nonce := uint64(42)
	result := builder.SetNonce(nonce)

	assert.Equal(t, builder, result) // Should return self for chaining
	assert.Equal(t, nonce, builder.nonce)
}

func TestEthereumTransactionBuilderBuildTransferTransaction(t *testing.T) {
	builder := NewEthereumTransactionBuilder(big.NewInt(1))

	to := common.HexToAddress("0x1234567890123456789012345678901234567890")
	amount := big.NewInt(1000000000000000000) // 1 ETH

	tx := builder.BuildTransferTransaction(to, amount)

	assert.NotNil(t, tx)
	assert.Equal(t, to, *tx.To())
	assert.Equal(t, amount, tx.Value())
	assert.Equal(t, builder.gasLimit, tx.Gas())
	assert.Equal(t, builder.gasPrice, tx.GasPrice())
	assert.Equal(t, builder.nonce, tx.Nonce())
}

func TestEthereumTransactionBuilderBuildContractTransaction(t *testing.T) {
	builder := NewEthereumTransactionBuilder(big.NewInt(1))

	to := common.HexToAddress("0x1234567890123456789012345678901234567890")
	value := big.NewInt(0)
	data := []byte("contract call data")

	tx := builder.BuildContractTransaction(to, value, data)

	assert.NotNil(t, tx)
	assert.Equal(t, to, *tx.To())
	assert.Equal(t, value, tx.Value())
	assert.Equal(t, data, tx.Data())
	assert.Equal(t, builder.gasLimit, tx.Gas())
	assert.Equal(t, builder.gasPrice, tx.GasPrice())
	assert.Equal(t, builder.nonce, tx.Nonce())
}

func TestEthereumTransactionBuilderBuildEIP1559Transaction(t *testing.T) {
	builder := NewEthereumTransactionBuilder(big.NewInt(1))

	to := common.HexToAddress("0x1234567890123456789012345678901234567890")
	amount := big.NewInt(1000000000000000000) // 1 ETH
	maxFeePerGas := big.NewInt(30000000000)
	maxPriorityFeePerGas := big.NewInt(2000000000)
	data := []byte("test data")

	tx := builder.BuildEIP1559Transaction(to, amount, maxFeePerGas, maxPriorityFeePerGas, data)

	assert.NotNil(t, tx)
	assert.Equal(t, uint8(types.DynamicFeeTxType), tx.Type())
	assert.Equal(t, to, *tx.To())
	assert.Equal(t, amount, tx.Value())
	assert.Equal(t, data, tx.Data())
	assert.Equal(t, builder.gasLimit, tx.Gas())
	assert.Equal(t, maxFeePerGas, tx.GasFeeCap())
	assert.Equal(t, maxPriorityFeePerGas, tx.GasTipCap())
	assert.Equal(t, builder.nonce, tx.Nonce())
}

func TestEthereumTransactionBuilderSignTransaction(t *testing.T) {
	builder := NewEthereumTransactionBuilder(big.NewInt(1))
	wallet, err := WalletFromEntropy("did:example:123", "test-salt", "snr")
	require.NoError(t, err)

	to := common.HexToAddress("0x1234567890123456789012345678901234567890")
	amount := big.NewInt(1000000000000000000) // 1 ETH

	tx := builder.BuildTransferTransaction(to, amount)
	signedTx, err := builder.SignTransaction(tx, wallet)

	require.NoError(t, err)
	assert.NotNil(t, signedTx)

	// Verify the transaction is signed
	v, r, s := signedTx.RawSignatureValues()
	assert.NotNil(t, v)
	assert.NotNil(t, r)
	assert.NotNil(t, s)
	assert.True(t, v.Cmp(big.NewInt(0)) > 0)
	assert.True(t, r.Cmp(big.NewInt(0)) > 0)
	assert.True(t, s.Cmp(big.NewInt(0)) > 0)
}

func TestEthereumTransactionBuilderEstimateGas(t *testing.T) {
	builder := NewEthereumTransactionBuilder(big.NewInt(1))

	to := common.HexToAddress("0x1234567890123456789012345678901234567890")
	amount := big.NewInt(1000000000000000000) // 1 ETH

	tx := builder.BuildTransferTransaction(to, amount)

	// This is a placeholder implementation that returns the current gas limit
	gasEstimate, err := builder.EstimateGas(nil, tx)
	require.NoError(t, err)
	assert.Equal(t, builder.gasLimit, gasEstimate)
}

func TestEthereumTransactionBuilderChaining(t *testing.T) {
	builder := NewEthereumTransactionBuilder(big.NewInt(1))

	// Test method chaining
	result := builder.SetGas(50000, big.NewInt(30000000000)).SetNonce(42)

	assert.Equal(t, builder, result)
	assert.Equal(t, uint64(50000), builder.gasLimit)
	assert.Equal(t, big.NewInt(30000000000), builder.gasPrice)
	assert.Equal(t, uint64(42), builder.nonce)
}

func TestGetDefaultEthereumParamsValues(t *testing.T) {
	defaultParams := GetDefaultEthereumParams()

	assert.Equal(t, big.NewInt(1), defaultParams.ChainID)
	assert.Equal(t, uint64(0), defaultParams.Nonce)
	assert.Equal(t, uint64(21000), defaultParams.GasLimit)
	assert.NotNil(t, defaultParams.GasPrice)
	assert.NotNil(t, defaultParams.MaxFeePerGas)
	assert.NotNil(t, defaultParams.MaxPriorityFeePerGas)
}

func TestTransactionParamsStructure(t *testing.T) {
	params := &TransactionParams{
		ChainID:       "test-chain",
		AccountNumber: 1,
		Sequence:      2,
		GasLimit:      200000,
		GasPrice:      DefaultGasPrice(),
		Memo:          "test memo",
	}

	assert.Equal(t, "test-chain", params.ChainID)
	assert.Equal(t, uint64(1), params.AccountNumber)
	assert.Equal(t, uint64(2), params.Sequence)
	assert.Equal(t, uint64(200000), params.GasLimit)
	assert.NotNil(t, params.GasPrice)
	assert.Equal(t, "test memo", params.Memo)
}

func TestEthereumTransactionParamsStructure(t *testing.T) {
	params := &EthereumTransactionParams{
		ChainID:              big.NewInt(1),
		Nonce:                5,
		GasLimit:             21000,
		GasPrice:             big.NewInt(20000000000),
		MaxFeePerGas:         big.NewInt(30000000000),
		MaxPriorityFeePerGas: big.NewInt(2000000000),
	}

	assert.Equal(t, big.NewInt(1), params.ChainID)
	assert.Equal(t, uint64(5), params.Nonce)
	assert.Equal(t, uint64(21000), params.GasLimit)
	assert.NotNil(t, params.GasPrice)
	assert.NotNil(t, params.MaxFeePerGas)
	assert.NotNil(t, params.MaxPriorityFeePerGas)
}

func TestTransactionTypesIntegration(t *testing.T) {
	// Test that different transaction types can be created and are compatible
	builder := NewEthereumTransactionBuilder(big.NewInt(1))
	to := common.HexToAddress("0x1234567890123456789012345678901234567890")
	amount := big.NewInt(1000000000000000000) // 1 ETH

	// Legacy transaction
	legacyTx := builder.BuildTransferTransaction(to, amount)
	assert.Equal(t, uint8(types.LegacyTxType), legacyTx.Type())

	// EIP-1559 transaction
	eip1559Tx := builder.BuildEIP1559Transaction(
		to,
		amount,
		big.NewInt(30000000000),
		big.NewInt(2000000000),
		nil,
	)
	assert.Equal(t, uint8(types.DynamicFeeTxType), eip1559Tx.Type())

	// Contract transaction
	contractTx := builder.BuildContractTransaction(to, big.NewInt(0), []byte("test"))
	assert.Equal(t, uint8(types.LegacyTxType), contractTx.Type())
	assert.Equal(t, []byte("test"), contractTx.Data())
}
