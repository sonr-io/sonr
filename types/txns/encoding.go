package txns

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// Encoder interface for encoding transactions
type Encoder interface {
	// EncodeTx encodes a transaction
	EncodeTx(tx any) ([]byte, error)
	// DecodeTx decodes a transaction
	DecodeTx(data []byte) (any, error)
	// GetEncodingType returns the encoding type
	GetEncodingType() EncodingType
}

// CosmosProtobufEncoder encodes/decodes Cosmos transactions using Protobuf
type CosmosProtobufEncoder struct {
	txConfig client.TxConfig
	cdc      codec.Codec
}

// NewCosmosProtobufEncoder creates a new Protobuf encoder for Cosmos
func NewCosmosProtobufEncoder(clientCtx client.Context) *CosmosProtobufEncoder {
	return &CosmosProtobufEncoder{
		txConfig: clientCtx.TxConfig,
		cdc:      clientCtx.Codec,
	}
}

// EncodeTx implements Encoder interface
func (e *CosmosProtobufEncoder) EncodeTx(tx any) ([]byte, error) {
	switch t := tx.(type) {
	case client.TxBuilder:
		return e.txConfig.TxEncoder()(t.GetTx())
	case sdk.Tx:
		return e.txConfig.TxEncoder()(t)
	case *CosmosSignedTx:
		return e.txConfig.TxEncoder()(t.TxBuilder.GetTx())
	case *CosmosUnsignedTx:
		return e.txConfig.TxEncoder()(t.TxBuilder.GetTx())
	default:
		return nil, fmt.Errorf("unsupported transaction type for Protobuf encoding: %T", tx)
	}
}

// DecodeTx implements Encoder interface
func (e *CosmosProtobufEncoder) DecodeTx(data []byte) (any, error) {
	tx, err := e.txConfig.TxDecoder()(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Protobuf transaction: %w", err)
	}
	return tx, nil
}

// GetEncodingType implements Encoder interface
func (e *CosmosProtobufEncoder) GetEncodingType() EncodingType {
	return EncodingTypeProtobuf
}

// CosmosAminoEncoder encodes/decodes Cosmos transactions using Amino
type CosmosAminoEncoder struct {
	cdc *codec.LegacyAmino
}

// NewCosmosAminoEncoder creates a new Amino encoder for Cosmos
func NewCosmosAminoEncoder() *CosmosAminoEncoder {
	// Create a legacy amino codec
	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	return &CosmosAminoEncoder{
		cdc: cdc,
	}
}

// EncodeTx implements Encoder interface
func (e *CosmosAminoEncoder) EncodeTx(tx any) ([]byte, error) {
	switch t := tx.(type) {
	case sdk.Tx:
		return e.cdc.Marshal(t)
	case *CosmosSignedTx:
		return e.cdc.Marshal(t.TxBuilder.GetTx())
	case *CosmosUnsignedTx:
		return e.cdc.Marshal(t.TxBuilder.GetTx())
	default:
		return nil, fmt.Errorf("unsupported transaction type for Amino encoding: %T", tx)
	}
}

// DecodeTx implements Encoder interface
func (e *CosmosAminoEncoder) DecodeTx(data []byte) (any, error) {
	var tx sdk.Tx
	err := e.cdc.Unmarshal(data, &tx)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Amino transaction: %w", err)
	}
	return tx, nil
}

// GetEncodingType implements Encoder interface
func (e *CosmosAminoEncoder) GetEncodingType() EncodingType {
	return EncodingTypeAmino
}

// EVMRLPEncoder encodes/decodes EVM transactions using RLP
type EVMRLPEncoder struct{}

// NewEVMRLPEncoder creates a new RLP encoder for EVM
func NewEVMRLPEncoder() *EVMRLPEncoder {
	return &EVMRLPEncoder{}
}

// EncodeTx implements Encoder interface
func (e *EVMRLPEncoder) EncodeTx(tx any) ([]byte, error) {
	switch t := tx.(type) {
	case *ethtypes.Transaction:
		return t.MarshalBinary()
	case *EVMSignedTx:
		return t.Transaction.MarshalBinary()
	case *EVMUnsignedTx:
		return t.Transaction.MarshalBinary()
	default:
		return nil, fmt.Errorf("unsupported transaction type for RLP encoding: %T", tx)
	}
}

// DecodeTx implements Encoder interface
func (e *EVMRLPEncoder) DecodeTx(data []byte) (any, error) {
	var tx ethtypes.Transaction
	err := tx.UnmarshalBinary(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode RLP transaction: %w", err)
	}
	return &tx, nil
}

// GetEncodingType implements Encoder interface
func (e *EVMRLPEncoder) GetEncodingType() EncodingType {
	return EncodingTypeRLP
}

// EncoderRegistry manages different encoders
type EncoderRegistry struct {
	encoders map[string]Encoder
}

// NewEncoderRegistry creates a new encoder registry
func NewEncoderRegistry() *EncoderRegistry {
	return &EncoderRegistry{
		encoders: make(map[string]Encoder),
	}
}

// RegisterEncoder registers an encoder
func (r *EncoderRegistry) RegisterEncoder(name string, encoder Encoder) {
	r.encoders[name] = encoder
}

// GetEncoder retrieves an encoder by name
func (r *EncoderRegistry) GetEncoder(name string) (Encoder, error) {
	encoder, exists := r.encoders[name]
	if !exists {
		return nil, fmt.Errorf("encoder not found: %s", name)
	}
	return encoder, nil
}

// GetEncoderByType retrieves an encoder by encoding type and transaction type
func (r *EncoderRegistry) GetEncoderByType(
	encodingType EncodingType,
	txType TransactionType,
) (Encoder, error) {
	key := fmt.Sprintf("%s-%s", txType, encodingType)
	return r.GetEncoder(key)
}

// DefaultEncoderRegistry creates a registry with default encoders
func DefaultEncoderRegistry(clientCtx client.Context) *EncoderRegistry {
	registry := NewEncoderRegistry()

	// Register Cosmos encoders
	registry.RegisterEncoder(
		fmt.Sprintf("%s-%s", TransactionTypeCosmos, EncodingTypeProtobuf),
		NewCosmosProtobufEncoder(clientCtx),
	)
	registry.RegisterEncoder(
		fmt.Sprintf("%s-%s", TransactionTypeCosmos, EncodingTypeAmino),
		NewCosmosAminoEncoder(),
	)

	// Register EVM encoder
	registry.RegisterEncoder(
		fmt.Sprintf("%s-%s", TransactionTypeEVM, EncodingTypeRLP),
		NewEVMRLPEncoder(),
	)

	return registry
}

// TransactionData represents decoded transaction data
type TransactionData struct {
	Type     TransactionType `json:"type"`
	Encoding EncodingType    `json:"encoding"`
	Hash     string          `json:"hash"`
	Size     int             `json:"size"`
	Raw      any             `json:"raw"`
	Metadata any             `json:"metadata,omitempty"`
}

// DecodeTransaction decodes a transaction and returns structured data
func DecodeTransaction(
	data []byte,
	encodingType EncodingType,
	txType TransactionType,
	clientCtx client.Context,
) (*TransactionData, error) {
	registry := DefaultEncoderRegistry(clientCtx)
	encoder, err := registry.GetEncoderByType(encodingType, txType)
	if err != nil {
		return nil, fmt.Errorf("failed to get encoder: %w", err)
	}

	decodedTx, err := encoder.DecodeTx(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %w", err)
	}

	// Calculate hash based on transaction type
	var hash string
	switch txType {
	case TransactionTypeCosmos:
		if cosmosTx, ok := decodedTx.(sdk.Tx); ok {
			// Calculate Cosmos transaction hash
			txBytes, err := encoder.EncodeTx(cosmosTx)
			if err == nil {
				hash = fmt.Sprintf("%x", txBytes[:32]) // Simple hash for demo
			}
		}
	case TransactionTypeEVM:
		if evmTx, ok := decodedTx.(*ethtypes.Transaction); ok {
			hash = evmTx.Hash().Hex()
		}
	}

	return &TransactionData{
		Type:     txType,
		Encoding: encodingType,
		Hash:     hash,
		Size:     len(data),
		Raw:      decodedTx,
	}, nil
}

// EncodeTransaction encodes a transaction using the specified encoding
func EncodeTransaction(
	tx any,
	encodingType EncodingType,
	txType TransactionType,
	clientCtx client.Context,
) ([]byte, error) {
	registry := DefaultEncoderRegistry(clientCtx)
	encoder, err := registry.GetEncoderByType(encodingType, txType)
	if err != nil {
		return nil, fmt.Errorf("failed to get encoder: %w", err)
	}

	return encoder.EncodeTx(tx)
}

// ConvertEncoding converts a transaction from one encoding to another
func ConvertEncoding(
	data []byte,
	fromEncoding, toEncoding EncodingType,
	txType TransactionType,
	clientCtx client.Context,
) ([]byte, error) {
	// Decode with source encoding
	decoded, err := DecodeTransaction(data, fromEncoding, txType, clientCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %w", err)
	}

	// Encode with target encoding
	return EncodeTransaction(decoded.Raw, toEncoding, txType, clientCtx)
}

// ValidateTransactionEncoding validates that transaction data is properly encoded
func ValidateTransactionEncoding(
	data []byte,
	encodingType EncodingType,
	txType TransactionType,
	clientCtx client.Context,
) error {
	_, err := DecodeTransaction(data, encodingType, txType, clientCtx)
	if err != nil {
		return fmt.Errorf("invalid transaction encoding: %w", err)
	}
	return nil
}

// GetTransactionSize returns the size of an encoded transaction
func GetTransactionSize(
	tx any,
	encodingType EncodingType,
	txType TransactionType,
	clientCtx client.Context,
) (int, error) {
	data, err := EncodeTransaction(tx, encodingType, txType, clientCtx)
	if err != nil {
		return 0, fmt.Errorf("failed to encode transaction: %w", err)
	}
	return len(data), nil
}
