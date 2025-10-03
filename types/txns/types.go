package txns

import (
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/sonr-io/sonr/crypto/mpc"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeCosmos  TransactionType = "cosmos"
	TransactionTypeEVM     TransactionType = "evm"
	TransactionTypeUnknown TransactionType = "unknown"
)

// EncodingType represents the encoding format for transactions
type EncodingType string

const (
	EncodingTypeAmino    EncodingType = "amino"
	EncodingTypeProtobuf EncodingType = "protobuf"
	EncodingTypeRLP      EncodingType = "rlp"
)

// Params defines the interface for transaction parameters
type Params interface {
	// Validate validates the transaction parameters
	Validate() error
	// GetType returns the transaction type
	GetType() TransactionType
	// GetMemo returns the transaction memo if any
	GetMemo() string
}

// Builder defines the interface for transaction builders
type Builder interface {
	// SetChainID sets the chain ID for the transaction
	SetChainID(chainID string) Builder
	// SetGas sets gas parameters
	SetGas(limit uint64, price any) Builder
	// SetMemo sets transaction memo
	SetMemo(memo string) Builder
	// BuildUnsigned creates an unsigned transaction
	BuildUnsigned(params Params) (UnsignedTransaction, error)
	// EstimateFee estimates the transaction fee
	EstimateFee(params Params) (*FeeEstimation, error)
	// GetTransactionType returns the transaction type
	GetTransactionType() TransactionType
}

// UnsignedTransaction represents an unsigned transaction that can be signed
type UnsignedTransaction interface {
	// GetSignBytes returns the bytes to be signed
	GetSignBytes() ([]byte, error)
	// GetType returns the transaction type
	GetType() TransactionType
	// GetEncoding returns the encoding type
	GetEncoding() EncodingType
	// Sign signs the transaction with the provided signature
	Sign(signature []byte, pubKey []byte) (SignedTransaction, error)
	// GetRaw returns the raw transaction data
	GetRaw() any
}

// SignedTransaction represents a signed transaction ready for broadcast
type SignedTransaction interface {
	// GetHash returns the transaction hash
	GetHash() string
	// GetBytes returns the serialized transaction bytes
	GetBytes() ([]byte, error)
	// GetType returns the transaction type
	GetType() TransactionType
	// GetEncoding returns the encoding type
	GetEncoding() EncodingType
	// GetRaw returns the raw signed transaction
	GetRaw() any
}

// Signer interface for signing transactions
type Signer interface {
	// Sign signs the transaction bytes and returns signature
	Sign(txBytes []byte) ([]byte, error)
	// GetPublicKey returns the public key
	GetPublicKey() []byte
	// GetAddress returns the address for the given chain
	GetAddress(chainType TransactionType) (string, error)
}

// CosmosTransactionParams holds parameters for Cosmos transactions
type CosmosTransactionParams struct {
	Messages      []sdk.Msg
	GasLimit      uint64
	GasPrice      sdk.DecCoin
	Fee           sdk.Coins
	Memo          string
	TimeoutHeight uint64
	AccountNumber uint64
	Sequence      uint64
	ChainID       string
}

// Validate implements Params interface
func (p *CosmosTransactionParams) Validate() error {
	if len(p.Messages) == 0 {
		return fmt.Errorf("at least one message is required")
	}
	if p.GasLimit == 0 {
		return fmt.Errorf("gas limit must be greater than 0")
	}
	return nil
}

// GetType implements Params interface
func (p *CosmosTransactionParams) GetType() TransactionType {
	return TransactionTypeCosmos
}

// GetMemo implements Params interface
func (p *CosmosTransactionParams) GetMemo() string {
	return p.Memo
}

// EVMTransactionParams holds parameters for EVM transactions
type EVMTransactionParams struct {
	To                   *common.Address
	Value                *big.Int
	Data                 []byte
	GasLimit             uint64
	GasPrice             *big.Int
	MaxFeePerGas         *big.Int
	MaxPriorityFeePerGas *big.Int
	Nonce                uint64
	ChainID              *big.Int
	Memo                 string // Some chains support memos in EVM transactions
}

// Validate implements Params interface
func (p *EVMTransactionParams) Validate() error {
	if p.ChainID == nil {
		return fmt.Errorf("chain ID is required")
	}
	if p.GasLimit == 0 {
		return fmt.Errorf("gas limit must be greater than 0")
	}
	if p.Value == nil {
		p.Value = big.NewInt(0)
	}
	return nil
}

// GetType implements Params interface
func (p *EVMTransactionParams) GetType() TransactionType {
	return TransactionTypeEVM
}

// GetMemo implements Params interface
func (p *EVMTransactionParams) GetMemo() string {
	return p.Memo
}

// FeeEstimation represents fee estimation data
type FeeEstimation struct {
	GasLimit uint64 `json:"gas_limit"`
	GasPrice any    `json:"gas_price"`
	Fee      any    `json:"fee"`
	Total    string `json:"total"`
}

// AddressDerivation represents address derivation information
type AddressDerivation struct {
	CosmosAddress  string `json:"cosmos_address"`
	EVMAddress     string `json:"evm_address"`
	DerivationPath string `json:"derivation_path"`
	PublicKey      []byte `json:"public_key"`
	ChainType      string `json:"chain_type"`
}

// MPCSigner implements Signer interface using MPC enclave
type MPCSigner struct {
	enclave   mpc.Enclave
	publicKey []byte
	chainID   string
}

// NewMPCSigner creates a new MPC signer
func NewMPCSigner(enclaveData *mpc.EnclaveData, chainID string) (*MPCSigner, error) {
	enclave, err := mpc.ImportEnclave(mpc.WithEnclaveData(enclaveData))
	if err != nil {
		return nil, err
	}

	return &MPCSigner{
		enclave:   enclave,
		publicKey: enclave.PubKeyBytes(),
		chainID:   chainID,
	}, nil
}

// Sign implements Signer interface
func (m *MPCSigner) Sign(txBytes []byte) ([]byte, error) {
	return m.enclave.Sign(txBytes)
}

// GetPublicKey implements Signer interface
func (m *MPCSigner) GetPublicKey() []byte {
	return m.publicKey
}

// GetAddress implements Signer interface
func (m *MPCSigner) GetAddress(chainType TransactionType) (string, error) {
	// This would need to derive address from public key based on chain type
	// Implementation would depend on the specific address derivation logic
	// For now, return a placeholder
	switch chainType {
	case TransactionTypeCosmos:
		return "cosmos1...", nil
	case TransactionTypeEVM:
		return "0x...", nil
	default:
		return "", ErrUnsupportedChainType
	}
}

// CosmosUnsignedTx represents an unsigned Cosmos transaction
type CosmosUnsignedTx struct {
	TxBuilder client.TxBuilder
	ChainID   string
	Encoding  EncodingType
}

// GetSignBytes implements UnsignedTransaction interface
func (c *CosmosUnsignedTx) GetSignBytes() ([]byte, error) {
	// This would return the actual sign bytes for the transaction
	// Implementation depends on the specific signing mode
	return []byte("cosmos-sign-bytes"), nil
}

// GetType implements UnsignedTransaction interface
func (c *CosmosUnsignedTx) GetType() TransactionType {
	return TransactionTypeCosmos
}

// GetEncoding implements UnsignedTransaction interface
func (c *CosmosUnsignedTx) GetEncoding() EncodingType {
	return c.Encoding
}

// Sign implements UnsignedTransaction interface
func (c *CosmosUnsignedTx) Sign(signature []byte, pubKey []byte) (SignedTransaction, error) {
	// Implementation would set the signature on the TxBuilder
	return &CosmosSignedTx{
		TxBuilder: c.TxBuilder,
		ChainID:   c.ChainID,
		Encoding:  c.Encoding,
	}, nil
}

// GetRaw implements UnsignedTransaction interface
func (c *CosmosUnsignedTx) GetRaw() any {
	return c.TxBuilder
}

// CosmosSignedTx represents a signed Cosmos transaction
type CosmosSignedTx struct {
	TxBuilder client.TxBuilder
	ChainID   string
	Encoding  EncodingType
}

// GetHash implements SignedTransaction interface
func (c *CosmosSignedTx) GetHash() string {
	// Implementation would calculate actual transaction hash
	return "0x..."
}

// GetBytes implements SignedTransaction interface
func (c *CosmosSignedTx) GetBytes() ([]byte, error) {
	// Implementation would serialize the transaction
	return []byte("serialized-cosmos-tx"), nil
}

// GetType implements SignedTransaction interface
func (c *CosmosSignedTx) GetType() TransactionType {
	return TransactionTypeCosmos
}

// GetEncoding implements SignedTransaction interface
func (c *CosmosSignedTx) GetEncoding() EncodingType {
	return c.Encoding
}

// GetRaw implements SignedTransaction interface
func (c *CosmosSignedTx) GetRaw() any {
	return c.TxBuilder
}

// EVMUnsignedTx represents an unsigned EVM transaction
type EVMUnsignedTx struct {
	Transaction *ethtypes.Transaction
	ChainID     *big.Int
}

// GetSignBytes implements UnsignedTransaction interface
func (e *EVMUnsignedTx) GetSignBytes() ([]byte, error) {
	signer := ethtypes.NewEIP155Signer(e.ChainID)
	return signer.Hash(e.Transaction).Bytes(), nil
}

// GetType implements UnsignedTransaction interface
func (e *EVMUnsignedTx) GetType() TransactionType {
	return TransactionTypeEVM
}

// GetEncoding implements UnsignedTransaction interface
func (e *EVMUnsignedTx) GetEncoding() EncodingType {
	return EncodingTypeRLP
}

// Sign implements UnsignedTransaction interface
func (e *EVMUnsignedTx) Sign(signature []byte, pubKey []byte) (SignedTransaction, error) {
	// Implementation would create signed transaction from signature
	return &EVMSignedTx{
		Transaction: e.Transaction,
		ChainID:     e.ChainID,
	}, nil
}

// GetRaw implements UnsignedTransaction interface
func (e *EVMUnsignedTx) GetRaw() any {
	return e.Transaction
}

// EVMSignedTx represents a signed EVM transaction
type EVMSignedTx struct {
	Transaction *ethtypes.Transaction
	ChainID     *big.Int
}

// GetHash implements SignedTransaction interface
func (e *EVMSignedTx) GetHash() string {
	return e.Transaction.Hash().Hex()
}

// GetBytes implements SignedTransaction interface
func (e *EVMSignedTx) GetBytes() ([]byte, error) {
	return e.Transaction.MarshalBinary()
}

// GetType implements SignedTransaction interface
func (e *EVMSignedTx) GetType() TransactionType {
	return TransactionTypeEVM
}

// GetEncoding implements SignedTransaction interface
func (e *EVMSignedTx) GetEncoding() EncodingType {
	return EncodingTypeRLP
}

// GetRaw implements SignedTransaction interface
func (e *EVMSignedTx) GetRaw() any {
	return e.Transaction
}
