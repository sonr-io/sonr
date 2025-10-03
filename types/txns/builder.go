package txns

import (
	"context"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/sonr-io/sonr/types/coins"
)

// TransactionBuilder is the main builder interface that wraps both Cosmos and EVM builders
type TransactionBuilder struct {
	cosmosBuilder   *CosmosBuilder
	evmBuilder      *EVMBuilder
	coinsManager    *coins.Manager
	defaultChainID  string
	defaultEncoding EncodingType
}

// NewTransactionBuilder creates a new transaction builder
func NewTransactionBuilder(coinsManager *coins.Manager, defaultChainID string) *TransactionBuilder {
	return &TransactionBuilder{
		coinsManager:    coinsManager,
		defaultChainID:  defaultChainID,
		defaultEncoding: EncodingTypeProtobuf,
	}
}

// Cosmos returns a Cosmos transaction builder
func (tb *TransactionBuilder) Cosmos(clientCtx client.Context) *CosmosBuilder {
	if tb.cosmosBuilder == nil {
		tb.cosmosBuilder = NewCosmosBuilder(clientCtx, tb.defaultChainID, tb.defaultEncoding)
	}
	return tb.cosmosBuilder
}

// EVM returns an EVM transaction builder
func (tb *TransactionBuilder) EVM(chainID *big.Int) *EVMBuilder {
	if tb.evmBuilder == nil {
		tb.evmBuilder = NewEVMBuilder(chainID)
	}
	return tb.evmBuilder
}

// CreateSigner creates a signer from wallet
func (tb *TransactionBuilder) CreateSigner(
	wallet *coins.Wallet,
	chainType TransactionType,
) (Signer, error) {
	return NewWalletSigner(wallet, chainType)
}

// DeriveAddresses derives addresses for both chains
func (tb *TransactionBuilder) DeriveAddresses(did, salt string) (*AddressDerivation, error) {
	cosmosAddr, evmAddr, derivationPath, err := tb.coinsManager.DeriveAddresses(did, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive addresses: %w", err)
	}

	// Get public key from wallet creation
	wallet, err := tb.coinsManager.CreateWalletFromEntropy(did, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet for public key: %w", err)
	}

	return &AddressDerivation{
		CosmosAddress:  cosmosAddr,
		EVMAddress:     evmAddr,
		DerivationPath: derivationPath,
		PublicKey:      wallet.GetCosmosPublicKey().Bytes(),
		ChainType:      "multi-chain",
	}, nil
}

// CosmosBuilder builds Cosmos transactions
type CosmosBuilder struct {
	clientCtx      client.Context
	chainID        string
	encoding       EncodingType
	gasLimit       uint64
	gasPrice       sdk.DecCoin
	memo           string
	timeoutHeight  uint64
	coinsTxBuilder *coins.CosmosTransactionBuilder
}

// NewCosmosBuilder creates a new Cosmos transaction builder
func NewCosmosBuilder(
	clientCtx client.Context,
	chainID string,
	encoding EncodingType,
) *CosmosBuilder {
	coinsTxBuilder := coins.NewCosmosTransactionBuilder(clientCtx, chainID)

	return &CosmosBuilder{
		clientCtx:      clientCtx,
		chainID:        chainID,
		encoding:       encoding,
		gasLimit:       200000,
		gasPrice:       sdk.NewDecCoin("usnr", math.NewInt(1000)),
		coinsTxBuilder: coinsTxBuilder,
	}
}

// SetChainID implements Builder interface
func (cb *CosmosBuilder) SetChainID(chainID string) Builder {
	cb.chainID = chainID
	return cb
}

// SetGas implements Builder interface
func (cb *CosmosBuilder) SetGas(limit uint64, price any) Builder {
	cb.gasLimit = limit
	if gasPrice, ok := price.(sdk.DecCoin); ok {
		cb.gasPrice = gasPrice
		cb.coinsTxBuilder.SetGas(limit, gasPrice)
	}
	return cb
}

// SetMemo implements Builder interface
func (cb *CosmosBuilder) SetMemo(memo string) Builder {
	cb.memo = memo
	cb.coinsTxBuilder.SetMemo(memo)
	return cb
}

// SetTimeoutHeight sets the timeout height for the transaction
func (cb *CosmosBuilder) SetTimeoutHeight(height uint64) *CosmosBuilder {
	cb.timeoutHeight = height
	return cb
}

// SetEncoding sets the encoding type
func (cb *CosmosBuilder) SetEncoding(encoding EncodingType) *CosmosBuilder {
	cb.encoding = encoding
	return cb
}

// BuildUnsigned implements Builder interface
func (cb *CosmosBuilder) BuildUnsigned(params Params) (UnsignedTransaction, error) {
	cosmosParams, ok := params.(*CosmosTransactionParams)
	if !ok {
		return nil, ErrInvalidTransactionParams
	}

	if err := cosmosParams.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Use the underlying coins transaction builder
	txBuilder, err := cb.coinsTxBuilder.BuildCustomTransaction(cosmosParams.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	// Set additional parameters
	if cosmosParams.TimeoutHeight > 0 {
		txBuilder.SetTimeoutHeight(cosmosParams.TimeoutHeight)
	}

	return &CosmosUnsignedTx{
		TxBuilder: txBuilder,
		ChainID:   cb.chainID,
		Encoding:  cb.encoding,
	}, nil
}

// EstimateFee implements Builder interface
func (cb *CosmosBuilder) EstimateFee(params Params) (*FeeEstimation, error) {
	cosmosParams, ok := params.(*CosmosTransactionParams)
	if !ok {
		return nil, ErrInvalidTransactionParams
	}

	if err := cosmosParams.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Simple fee estimation based on gas limit and price
	fee := cb.gasPrice.Amount.MulInt64(int64(cb.gasLimit))
	totalFee := sdk.NewCoins(sdk.NewCoin(cb.gasPrice.Denom, fee.TruncateInt()))

	return &FeeEstimation{
		GasLimit: cb.gasLimit,
		GasPrice: cb.gasPrice,
		Fee:      totalFee,
		Total:    totalFee.String(),
	}, nil
}

// GetTransactionType implements Builder interface
func (cb *CosmosBuilder) GetTransactionType() TransactionType {
	return TransactionTypeCosmos
}

// BuildSendTransaction builds a simple send transaction
func (cb *CosmosBuilder) BuildSendTransaction(
	fromAddr, toAddr string,
	amount sdk.Coins,
) (UnsignedTransaction, error) {
	txBuilder, err := cb.coinsTxBuilder.BuildSendTransaction(fromAddr, toAddr, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to build send transaction: %w", err)
	}

	return &CosmosUnsignedTx{
		TxBuilder: txBuilder,
		ChainID:   cb.chainID,
		Encoding:  cb.encoding,
	}, nil
}

// EVMBuilder builds EVM transactions
type EVMBuilder struct {
	chainID              *big.Int
	gasLimit             uint64
	gasPrice             *big.Int
	maxFeePerGas         *big.Int
	maxPriorityFeePerGas *big.Int
	nonce                uint64
	coinsEthBuilder      *coins.EthereumTransactionBuilder
}

// NewEVMBuilder creates a new EVM transaction builder
func NewEVMBuilder(chainID *big.Int) *EVMBuilder {
	coinsEthBuilder := coins.NewEthereumTransactionBuilder(chainID)

	return &EVMBuilder{
		chainID:         chainID,
		gasLimit:        21000,
		gasPrice:        big.NewInt(20000000000), // 20 Gwei
		coinsEthBuilder: coinsEthBuilder,
	}
}

// SetChainID implements Builder interface
func (eb *EVMBuilder) SetChainID(chainID string) Builder {
	if chainIDBig, ok := new(big.Int).SetString(chainID, 10); ok {
		eb.chainID = chainIDBig
	}
	return eb
}

// SetGas implements Builder interface
func (eb *EVMBuilder) SetGas(limit uint64, price any) Builder {
	eb.gasLimit = limit
	if gasPrice, ok := price.(*big.Int); ok {
		eb.gasPrice = gasPrice
		eb.coinsEthBuilder.SetGas(limit, gasPrice)
	}
	return eb
}

// SetMemo implements Builder interface (no-op for EVM)
func (eb *EVMBuilder) SetMemo(memo string) Builder {
	// EVM transactions don't have memos
	return eb
}

// SetNonce sets the nonce for the transaction
func (eb *EVMBuilder) SetNonce(nonce uint64) *EVMBuilder {
	eb.nonce = nonce
	eb.coinsEthBuilder.SetNonce(nonce)
	return eb
}

// SetMaxFee sets the max fee per gas for EIP-1559 transactions
func (eb *EVMBuilder) SetMaxFee(maxFeePerGas, maxPriorityFeePerGas *big.Int) *EVMBuilder {
	eb.maxFeePerGas = maxFeePerGas
	eb.maxPriorityFeePerGas = maxPriorityFeePerGas
	return eb
}

// BuildUnsigned implements Builder interface
func (eb *EVMBuilder) BuildUnsigned(params Params) (UnsignedTransaction, error) {
	evmParams, ok := params.(*EVMTransactionParams)
	if !ok {
		return nil, ErrInvalidTransactionParams
	}

	if err := evmParams.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	var tx *ethtypes.Transaction

	// Choose transaction type based on fee parameters
	if evmParams.MaxFeePerGas != nil && evmParams.MaxPriorityFeePerGas != nil {
		// EIP-1559 transaction
		tx = eb.coinsEthBuilder.BuildEIP1559Transaction(
			*evmParams.To,
			evmParams.Value,
			evmParams.MaxFeePerGas,
			evmParams.MaxPriorityFeePerGas,
			evmParams.Data,
		)
	} else if evmParams.Data != nil && len(evmParams.Data) > 0 {
		// Contract transaction
		tx = eb.coinsEthBuilder.BuildContractTransaction(
			*evmParams.To,
			evmParams.Value,
			evmParams.Data,
		)
	} else {
		// Simple transfer
		tx = eb.coinsEthBuilder.BuildTransferTransaction(
			*evmParams.To,
			evmParams.Value,
		)
	}

	return &EVMUnsignedTx{
		Transaction: tx,
		ChainID:     eb.chainID,
	}, nil
}

// EstimateFee implements Builder interface
func (eb *EVMBuilder) EstimateFee(params Params) (*FeeEstimation, error) {
	evmParams, ok := params.(*EVMTransactionParams)
	if !ok {
		return nil, ErrInvalidTransactionParams
	}

	if err := evmParams.Validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Simple fee estimation
	totalFee := new(big.Int).Mul(eb.gasPrice, big.NewInt(int64(eb.gasLimit)))

	return &FeeEstimation{
		GasLimit: eb.gasLimit,
		GasPrice: eb.gasPrice,
		Fee:      totalFee,
		Total:    totalFee.String(),
	}, nil
}

// GetTransactionType implements Builder interface
func (eb *EVMBuilder) GetTransactionType() TransactionType {
	return TransactionTypeEVM
}

// EstimateGas estimates gas for a transaction
func (eb *EVMBuilder) EstimateGas(
	ctx context.Context,
	params *EVMTransactionParams,
) (uint64, error) {
	// Use the underlying coins builder for gas estimation
	tx := eb.coinsEthBuilder.BuildContractTransaction(*params.To, params.Value, params.Data)
	return eb.coinsEthBuilder.EstimateGas(ctx, tx)
}

// WalletSigner implements Signer interface using coins.Wallet
type WalletSigner struct {
	wallet    *coins.Wallet
	chainType TransactionType
}

// NewWalletSigner creates a new wallet-based signer
func NewWalletSigner(wallet *coins.Wallet, chainType TransactionType) (*WalletSigner, error) {
	if wallet == nil {
		return nil, fmt.Errorf("wallet cannot be nil")
	}

	return &WalletSigner{
		wallet:    wallet,
		chainType: chainType,
	}, nil
}

// Sign implements Signer interface
func (ws *WalletSigner) Sign(txBytes []byte) ([]byte, error) {
	switch ws.chainType {
	case TransactionTypeCosmos:
		return ws.wallet.SignMessage(txBytes)
	case TransactionTypeEVM:
		return ws.wallet.SignEthereumMessage(txBytes)
	default:
		return nil, ErrUnsupportedChainType
	}
}

// GetPublicKey implements Signer interface
func (ws *WalletSigner) GetPublicKey() []byte {
	switch ws.chainType {
	case TransactionTypeCosmos:
		return ws.wallet.GetCosmosPublicKey().Bytes()
	case TransactionTypeEVM:
		pubKey := ws.wallet.GetEthereumPublicKey()
		// Convert ECDSA public key to bytes
		return append(pubKey.X.Bytes(), pubKey.Y.Bytes()...)
	default:
		return nil
	}
}

// GetAddress implements Signer interface
func (ws *WalletSigner) GetAddress(chainType TransactionType) (string, error) {
	switch chainType {
	case TransactionTypeCosmos:
		return ws.wallet.CosmosAddress, nil
	case TransactionTypeEVM:
		return ws.wallet.EthereumAddress, nil
	default:
		return "", ErrUnsupportedChainType
	}
}
