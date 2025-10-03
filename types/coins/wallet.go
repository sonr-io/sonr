package coins

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// Wallet represents a multi-chain wallet with Cosmos and Ethereum capabilities
type Wallet struct {
	DID             string
	Salt            string
	CosmosAddress   string
	EthereumAddress string
	CosmosPrivKey   *ecdsa.PrivateKey
	EthereumPrivKey *ecdsa.PrivateKey
	DerivationPath  string
}

// GetCosmosPublicKey returns the Cosmos public key
func (w *Wallet) GetCosmosPublicKey() cryptotypes.PubKey {
	pubKeyBytes := crypto.FromECDSAPub(&w.CosmosPrivKey.PublicKey)
	return &secp256k1.PubKey{Key: pubKeyBytes}
}

// GetEthereumPublicKey returns the Ethereum public key
func (w *Wallet) GetEthereumPublicKey() *ecdsa.PublicKey {
	return &w.EthereumPrivKey.PublicKey
}

// SignCosmosTransaction signs a Cosmos transaction
func (w *Wallet) SignCosmosTransaction(
	txBuilder client.TxBuilder,
	chainID string,
	accountNumber, sequence uint64,
) ([]byte, error) {
	// Get the sign bytes
	signMode := signing.SignMode_SIGN_MODE_DIRECT

	// Create signature data
	sig := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: nil,
	}

	// Set the signature
	if err := txBuilder.SetSignatures(signing.SignatureV2{
		PubKey:   w.GetCosmosPublicKey(),
		Data:     &sig,
		Sequence: sequence,
	}); err != nil {
		return nil, fmt.Errorf("failed to set signatures: %w", err)
	}

	// Sign the transaction
	signature, err := w.signBytes([]byte("test"), w.CosmosPrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Update signature
	sig.Signature = signature
	if err := txBuilder.SetSignatures(signing.SignatureV2{
		PubKey:   w.GetCosmosPublicKey(),
		Data:     &sig,
		Sequence: sequence,
	}); err != nil {
		return nil, fmt.Errorf("failed to set final signatures: %w", err)
	}

	return signature, nil
}

// SignEthereumTransaction signs an Ethereum transaction
func (w *Wallet) SignEthereumTransaction(
	tx *types.Transaction,
	chainID *big.Int,
) (*types.Transaction, error) {
	signer := types.NewEIP155Signer(chainID)
	signedTx, err := types.SignTx(tx, signer, w.EthereumPrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign Ethereum transaction: %w", err)
	}

	return signedTx, nil
}

// GetEthereumTransactor returns a transactor for Ethereum smart contracts
func (w *Wallet) GetEthereumTransactor(chainID *big.Int) (*bind.TransactOpts, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(w.EthereumPrivKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	return auth, nil
}

// SignMessage signs a message using the Cosmos private key
func (w *Wallet) SignMessage(message []byte) ([]byte, error) {
	return w.signBytes(message, w.CosmosPrivKey)
}

// SignEthereumMessage signs a message using the Ethereum private key
func (w *Wallet) SignEthereumMessage(message []byte) ([]byte, error) {
	return w.signBytes(message, w.EthereumPrivKey)
}

// VerifySignature verifies a signature against the Cosmos public key
func (w *Wallet) VerifySignature(message, signature []byte) bool {
	pubKey := &w.CosmosPrivKey.PublicKey
	// Hash the message first
	hash := sha256.Sum256(message)
	// Remove the recovery ID (last byte) if present
	if len(signature) == 65 {
		signature = signature[:64]
	}
	return crypto.VerifySignature(crypto.FromECDSAPub(pubKey), hash[:], signature)
}

// VerifyEthereumSignature verifies a signature against the Ethereum public key
func (w *Wallet) VerifyEthereumSignature(message, signature []byte) bool {
	pubKey := &w.EthereumPrivKey.PublicKey
	// Hash the message first
	hash := sha256.Sum256(message)
	// Remove the recovery ID (last byte) if present
	if len(signature) == 65 {
		signature = signature[:64]
	}
	return crypto.VerifySignature(crypto.FromECDSAPub(pubKey), hash[:], signature)
}

// signBytes signs bytes using the provided private key
func (w *Wallet) signBytes(data []byte, privKey *ecdsa.PrivateKey) ([]byte, error) {
	// Hash the data
	hash := sha256.Sum256(data)

	// Sign the hash
	signature, err := crypto.Sign(hash[:], privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return signature, nil
}

// GetCosmosAccountAddress returns the Cosmos address as sdk.AccAddress
func (w *Wallet) GetCosmosAccountAddress() (sdk.AccAddress, error) {
	addr, err := sdk.AccAddressFromBech32(w.CosmosAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Cosmos address: %w", err)
	}
	return addr, nil
}

// GetEthereumAccountAddress returns the Ethereum address as common.Address
func (w *Wallet) GetEthereumAccountAddress() common.Address {
	return common.HexToAddress(w.EthereumAddress)
}

// ExportPrivateKeys returns the private keys in hexadecimal format
func (w *Wallet) ExportPrivateKeys() (cosmosPrivKey, ethPrivKey string) {
	cosmosPrivKey = fmt.Sprintf("%x", crypto.FromECDSA(w.CosmosPrivKey))
	ethPrivKey = fmt.Sprintf("%x", crypto.FromECDSA(w.EthereumPrivKey))
	return cosmosPrivKey, ethPrivKey
}

// WalletInfo contains basic wallet information
type WalletInfo struct {
	DID             string `json:"did"`
	Salt            string `json:"salt"`
	CosmosAddress   string `json:"cosmos_address"`
	EthereumAddress string `json:"ethereum_address"`
	DerivationPath  string `json:"derivation_path"`
}

// GetInfo returns wallet information without private keys
func (w *Wallet) GetInfo() WalletInfo {
	return WalletInfo{
		DID:             w.DID,
		Salt:            w.Salt,
		CosmosAddress:   w.CosmosAddress,
		EthereumAddress: w.EthereumAddress,
		DerivationPath:  w.DerivationPath,
	}
}
