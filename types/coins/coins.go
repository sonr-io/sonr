// Package coins provides address derivation and transaction building utilities
// for multi-chain wallets supporting both Cosmos SDK and Ethereum-based chains.
package coins

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Manager provides high-level wallet and transaction management
type Manager struct {
	cosmosPrefix string
	chainID      string
	ethChainID   *big.Int
}

// NewManager creates a new coins manager
func NewManager(cosmosPrefix, chainID string, ethChainID *big.Int) *Manager {
	return &Manager{
		cosmosPrefix: cosmosPrefix,
		chainID:      chainID,
		ethChainID:   ethChainID,
	}
}

// CreateWalletFromEntropy creates a new wallet from DID and salt
func (m *Manager) CreateWalletFromEntropy(did, salt string) (*Wallet, error) {
	return WalletFromEntropy(did, salt, m.cosmosPrefix)
}

// DeriveAddresses derives addresses from DID and salt without creating a full wallet
func (m *Manager) DeriveAddresses(
	did, salt string,
) (cosmosAddr, ethAddr, derivationPath string, err error) {
	return DeriveAddressesFromEntropy(did, salt, m.cosmosPrefix)
}

// CreateCosmosTransactionBuilder creates a Cosmos transaction builder
func (m *Manager) CreateCosmosTransactionBuilder(
	clientCtx client.Context,
) *CosmosTransactionBuilder {
	return NewCosmosTransactionBuilder(clientCtx, m.chainID)
}

// CreateEthereumTransactionBuilder creates an Ethereum transaction builder
func (m *Manager) CreateEthereumTransactionBuilder() *EthereumTransactionBuilder {
	return NewEthereumTransactionBuilder(m.ethChainID)
}

// SignAndBuildCosmosTransaction signs and builds a Cosmos transaction
func (m *Manager) SignAndBuildCosmosTransaction(
	clientCtx client.Context,
	wallet *Wallet,
	msgs []sdk.Msg,
	params *TransactionParams,
) ([]byte, error) {
	// Create transaction builder
	txBuilder := m.CreateCosmosTransactionBuilder(clientCtx)

	// Set gas and memo if provided
	if params != nil {
		if params.GasLimit > 0 {
			txBuilder.SetGas(params.GasLimit, params.GasPrice)
		}
		if params.Memo != "" {
			txBuilder.SetMemo(params.Memo)
		}
	}

	// Build transaction
	tx, err := txBuilder.BuildCustomTransaction(msgs)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	// Sign transaction
	if params == nil {
		return nil, fmt.Errorf("transaction parameters required for signing")
	}

	signedTx, err := txBuilder.SignTransaction(tx, wallet, params.AccountNumber, params.Sequence)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return signedTx, nil
}

// SignAndBuildEthereumTransaction signs and builds an Ethereum transaction
func (m *Manager) SignAndBuildEthereumTransaction(
	wallet *Wallet,
	to common.Address,
	amount *big.Int,
	params *EthereumTransactionParams,
) (*types.Transaction, error) {
	// Create transaction builder
	txBuilder := m.CreateEthereumTransactionBuilder()

	// Set parameters if provided
	if params != nil {
		if params.GasLimit > 0 {
			txBuilder.SetGas(params.GasLimit, params.GasPrice)
		}
		txBuilder.SetNonce(params.Nonce)
	}

	// Build transaction
	tx := txBuilder.BuildTransferTransaction(to, amount)

	// Sign transaction
	signedTx, err := txBuilder.SignTransaction(tx, wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return signedTx, nil
}

// ValidateAddress validates an address for the given chain
func (m *Manager) ValidateAddress(address, chain string) error {
	switch chain {
	case "cosmos":
		_, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			return fmt.Errorf("invalid Cosmos address: %w", err)
		}
	case "ethereum":
		if !common.IsHexAddress(address) {
			return fmt.Errorf("invalid Ethereum address")
		}
	default:
		return fmt.Errorf("unsupported chain: %s", chain)
	}

	return nil
}

// GetAddressFormat returns the address format for the given chain
func (m *Manager) GetAddressFormat(chain string) string {
	switch chain {
	case "cosmos":
		return fmt.Sprintf("bech32 with prefix '%s'", m.cosmosPrefix)
	case "ethereum":
		return "hex format (0x...)"
	default:
		return "unknown"
	}
}

// SupportedChains returns a list of supported chains
func (m *Manager) SupportedChains() []string {
	return []string{"cosmos", "ethereum"}
}

// ChainConfig holds chain-specific configuration
type ChainConfig struct {
	ChainID     string
	Prefix      string
	CoinType    uint32
	GasLimit    uint64
	GasPrice    sdk.DecCoin
	EthChainID  *big.Int
	EthGasPrice *big.Int
	EthGasLimit uint64
}

// GetDefaultChainConfig returns default chain configuration
func GetDefaultChainConfig() *ChainConfig {
	return &ChainConfig{
		ChainID:     "sonr-1",
		Prefix:      "snr",
		CoinType:    CoinTypeSonr,
		GasLimit:    200000,
		GasPrice:    sdk.NewDecCoin("usnr", math.NewInt(1000)),
		EthChainID:  big.NewInt(1),
		EthGasPrice: big.NewInt(20000000000), // 20 Gwei
		EthGasLimit: 21000,
	}
}

// NewManagerFromConfig creates a manager from chain configuration
func NewManagerFromConfig(config *ChainConfig) *Manager {
	return &Manager{
		cosmosPrefix: config.Prefix,
		chainID:      config.ChainID,
		ethChainID:   config.EthChainID,
	}
}
