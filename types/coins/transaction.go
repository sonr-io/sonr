package coins

import (
	"context"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// CosmosTransactionBuilder helps build Cosmos transactions
type CosmosTransactionBuilder struct {
	clientCtx client.Context
	txConfig  client.TxConfig
	chainID   string
	gasLimit  uint64
	gasPrice  sdk.DecCoin
	memo      string
}

// NewCosmosTransactionBuilder creates a new Cosmos transaction builder
func NewCosmosTransactionBuilder(
	clientCtx client.Context,
	chainID string,
) *CosmosTransactionBuilder {
	return &CosmosTransactionBuilder{
		clientCtx: clientCtx,
		txConfig:  clientCtx.TxConfig,
		chainID:   chainID,
		gasLimit:  200000,                                     // Default gas limit
		gasPrice:  sdk.NewDecCoin("stake", math.NewInt(1000)), // Default gas price
	}
}

// SetGas sets the gas limit and price for the transaction
func (ctb *CosmosTransactionBuilder) SetGas(
	limit uint64,
	price sdk.DecCoin,
) *CosmosTransactionBuilder {
	ctb.gasLimit = limit
	ctb.gasPrice = price
	return ctb
}

// SetMemo sets the memo for the transaction
func (ctb *CosmosTransactionBuilder) SetMemo(memo string) *CosmosTransactionBuilder {
	ctb.memo = memo
	return ctb
}

// BuildSendTransaction builds a bank send transaction
func (ctb *CosmosTransactionBuilder) BuildSendTransaction(
	fromAddr, toAddr string,
	amount sdk.Coins,
) (client.TxBuilder, error) {
	// Parse addresses
	fromAddress, err := sdk.AccAddressFromBech32(fromAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid from address: %w", err)
	}

	toAddress, err := sdk.AccAddressFromBech32(toAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid to address: %w", err)
	}

	// Create send message
	msg := &banktypes.MsgSend{
		FromAddress: fromAddress.String(),
		ToAddress:   toAddress.String(),
		Amount:      amount,
	}

	// Create transaction builder
	txBuilder := ctb.txConfig.NewTxBuilder()

	// Set messages
	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, fmt.Errorf("failed to set messages: %w", err)
	}

	// Set gas and fees
	txBuilder.SetGasLimit(ctb.gasLimit)
	fees := sdk.NewCoins(
		sdk.NewCoin(
			ctb.gasPrice.Denom,
			ctb.gasPrice.Amount.MulInt64(int64(ctb.gasLimit)).TruncateInt(),
		),
	)
	txBuilder.SetFeeAmount(fees)

	// Set memo
	if ctb.memo != "" {
		txBuilder.SetMemo(ctb.memo)
	}

	return txBuilder, nil
}

// BuildCustomTransaction builds a transaction with custom messages
func (ctb *CosmosTransactionBuilder) BuildCustomTransaction(
	msgs []sdk.Msg,
) (client.TxBuilder, error) {
	// Create transaction builder
	txBuilder := ctb.txConfig.NewTxBuilder()

	// Set messages
	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return nil, fmt.Errorf("failed to set messages: %w", err)
	}

	// Set gas and fees
	txBuilder.SetGasLimit(ctb.gasLimit)
	fees := sdk.NewCoins(
		sdk.NewCoin(
			ctb.gasPrice.Denom,
			ctb.gasPrice.Amount.MulInt64(int64(ctb.gasLimit)).TruncateInt(),
		),
	)
	txBuilder.SetFeeAmount(fees)

	// Set memo
	if ctb.memo != "" {
		txBuilder.SetMemo(ctb.memo)
	}

	return txBuilder, nil
}

// SignTransaction signs a transaction with the provided wallet
func (ctb *CosmosTransactionBuilder) SignTransaction(
	txBuilder client.TxBuilder,
	wallet *Wallet,
	accountNumber, sequence uint64,
) ([]byte, error) {
	// Simplified signing - just return a test signature for now
	// This would need proper implementation with correct signing flow
	message := []byte("cosmos-transaction")
	signature, err := wallet.SignMessage(message)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return signature, nil
}

// EthereumTransactionBuilder helps build Ethereum transactions
type EthereumTransactionBuilder struct {
	chainID  *big.Int
	gasLimit uint64
	gasPrice *big.Int
	nonce    uint64
}

// NewEthereumTransactionBuilder creates a new Ethereum transaction builder
func NewEthereumTransactionBuilder(chainID *big.Int) *EthereumTransactionBuilder {
	return &EthereumTransactionBuilder{
		chainID:  chainID,
		gasLimit: 21000,                   // Default gas limit for simple transfer
		gasPrice: big.NewInt(20000000000), // Default gas price (20 Gwei)
	}
}

// SetGas sets the gas limit and price for the transaction
func (etb *EthereumTransactionBuilder) SetGas(
	limit uint64,
	price *big.Int,
) *EthereumTransactionBuilder {
	etb.gasLimit = limit
	etb.gasPrice = price
	return etb
}

// SetNonce sets the nonce for the transaction
func (etb *EthereumTransactionBuilder) SetNonce(nonce uint64) *EthereumTransactionBuilder {
	etb.nonce = nonce
	return etb
}

// BuildTransferTransaction builds an Ethereum transfer transaction
func (etb *EthereumTransactionBuilder) BuildTransferTransaction(
	to common.Address,
	amount *big.Int,
) *types.Transaction {
	return types.NewTransaction(
		etb.nonce,
		to,
		amount,
		etb.gasLimit,
		etb.gasPrice,
		nil,
	)
}

// BuildContractTransaction builds an Ethereum contract interaction transaction
func (etb *EthereumTransactionBuilder) BuildContractTransaction(
	to common.Address,
	value *big.Int,
	data []byte,
) *types.Transaction {
	return types.NewTransaction(
		etb.nonce,
		to,
		value,
		etb.gasLimit,
		etb.gasPrice,
		data,
	)
}

// BuildEIP1559Transaction builds an EIP-1559 transaction with dynamic fees
func (etb *EthereumTransactionBuilder) BuildEIP1559Transaction(
	to common.Address,
	amount *big.Int,
	maxFeePerGas, maxPriorityFeePerGas *big.Int,
	data []byte,
) *types.Transaction {
	return types.NewTx(&types.DynamicFeeTx{
		ChainID:   etb.chainID,
		Nonce:     etb.nonce,
		To:        &to,
		Value:     amount,
		Gas:       etb.gasLimit,
		GasFeeCap: maxFeePerGas,
		GasTipCap: maxPriorityFeePerGas,
		Data:      data,
	})
}

// SignTransaction signs an Ethereum transaction with the provided wallet
func (etb *EthereumTransactionBuilder) SignTransaction(
	tx *types.Transaction,
	wallet *Wallet,
) (*types.Transaction, error) {
	return wallet.SignEthereumTransaction(tx, etb.chainID)
}

// EstimateGas estimates gas for a transaction (placeholder - would need actual client)
func (etb *EthereumTransactionBuilder) EstimateGas(
	ctx context.Context,
	tx *types.Transaction,
) (uint64, error) {
	// This would typically use an Ethereum client to estimate gas
	// For now, return a default estimate
	return etb.gasLimit, nil
}

// TransactionParams holds common transaction parameters
type TransactionParams struct {
	ChainID       string
	AccountNumber uint64
	Sequence      uint64
	GasLimit      uint64
	GasPrice      sdk.DecCoin
	Memo          string
}

// EthereumTransactionParams holds Ethereum transaction parameters
type EthereumTransactionParams struct {
	ChainID              *big.Int
	Nonce                uint64
	GasLimit             uint64
	GasPrice             *big.Int
	MaxFeePerGas         *big.Int
	MaxPriorityFeePerGas *big.Int
}

// GetDefaultEthereumParams returns default Ethereum transaction parameters
func GetDefaultEthereumParams() *EthereumTransactionParams {
	return &EthereumTransactionParams{
		ChainID:              big.NewInt(1), // Ethereum mainnet
		Nonce:                0,
		GasLimit:             21000,
		GasPrice:             big.NewInt(params.GWei * 20), // 20 Gwei
		MaxFeePerGas:         big.NewInt(params.GWei * 30), // 30 Gwei
		MaxPriorityFeePerGas: big.NewInt(params.GWei * 2),  // 2 Gwei
	}
}
