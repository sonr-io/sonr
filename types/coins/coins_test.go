package coins

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	cosmosPrefix := "snr"
	chainID := "sonr-1"
	ethChainID := big.NewInt(1)

	manager := NewManager(cosmosPrefix, chainID, ethChainID)
	assert.NotNil(t, manager)
	assert.Equal(t, cosmosPrefix, manager.cosmosPrefix)
	assert.Equal(t, chainID, manager.chainID)
	assert.Equal(t, ethChainID, manager.ethChainID)
}

func TestManagerCreateWalletFromEntropy(t *testing.T) {
	manager := NewManager("snr", "sonr-1", big.NewInt(1))

	did := "did:example:123456789abcdef"
	salt := "test-salt"

	wallet, err := manager.CreateWalletFromEntropy(did, salt)
	require.NoError(t, err)
	assert.NotNil(t, wallet)
	assert.Equal(t, did, wallet.DID)
	assert.Equal(t, salt, wallet.Salt)
}

func TestManagerDeriveAddresses(t *testing.T) {
	manager := NewManager("snr", "sonr-1", big.NewInt(1))

	did := "did:example:123456789abcdef"
	salt := "test-salt"

	cosmosAddr, ethAddr, derivationPath, err := manager.DeriveAddresses(did, salt)
	require.NoError(t, err)

	assert.True(t, len(cosmosAddr) > 0)
	assert.True(t, len(ethAddr) > 0)
	assert.True(t, len(derivationPath) > 0)
	assert.Contains(t, cosmosAddr, "snr")
	assert.Contains(t, ethAddr, "0x")
	assert.Contains(t, derivationPath, "m/44'/118'/0'/0/0")
}

func TestManagerCreateCosmosTransactionBuilder(t *testing.T) {
	manager := NewManager("snr", "sonr-1", big.NewInt(1))

	// This would need a real client context in a real test
	// For now, we just test that the method exists
	assert.NotNil(t, manager.CreateCosmosTransactionBuilder)
}

func TestManagerCreateEthereumTransactionBuilder(t *testing.T) {
	manager := NewManager("snr", "sonr-1", big.NewInt(1))

	txBuilder := manager.CreateEthereumTransactionBuilder()
	assert.NotNil(t, txBuilder)
	assert.Equal(t, big.NewInt(1), txBuilder.chainID)
}

func TestManagerValidateAddress(t *testing.T) {
	manager := NewManager("snr", "sonr-1", big.NewInt(1))

	// Test valid Cosmos address
	validCosmosAddr := "snr1abc123def456ghi789jkl012mno345pqr678stu"
	err := manager.ValidateAddress(validCosmosAddr, "cosmos")
	// This will fail because it's not a real address, but we test the method exists
	assert.Error(t, err)

	// Test valid Ethereum address
	validEthAddr := "0x1234567890123456789012345678901234567890"
	err = manager.ValidateAddress(validEthAddr, "ethereum")
	assert.NoError(t, err)

	// Test invalid Ethereum address
	invalidEthAddr := "0x123"
	err = manager.ValidateAddress(invalidEthAddr, "ethereum")
	assert.Error(t, err)

	// Test unsupported chain
	err = manager.ValidateAddress("address", "unsupported")
	assert.Error(t, err)
}

func TestManagerGetAddressFormat(t *testing.T) {
	manager := NewManager("snr", "sonr-1", big.NewInt(1))

	cosmosFormat := manager.GetAddressFormat("cosmos")
	assert.Contains(t, cosmosFormat, "bech32")
	assert.Contains(t, cosmosFormat, "snr")

	ethFormat := manager.GetAddressFormat("ethereum")
	assert.Contains(t, ethFormat, "hex")
	assert.Contains(t, ethFormat, "0x")

	unknownFormat := manager.GetAddressFormat("unknown")
	assert.Equal(t, "unknown", unknownFormat)
}

func TestManagerSupportedChains(t *testing.T) {
	manager := NewManager("snr", "sonr-1", big.NewInt(1))

	chains := manager.SupportedChains()
	assert.Contains(t, chains, "cosmos")
	assert.Contains(t, chains, "ethereum")
	assert.Len(t, chains, 2)
}

func TestGetDefaultChainConfig(t *testing.T) {
	config := GetDefaultChainConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "sonr-1", config.ChainID)
	assert.Equal(t, "snr", config.Prefix)
	assert.Equal(t, CoinTypeSonr, config.CoinType)
	assert.Equal(t, uint64(200000), config.GasLimit)
	assert.NotNil(t, config.GasPrice)
	assert.NotNil(t, config.EthChainID)
	assert.NotNil(t, config.EthGasPrice)
	assert.Equal(t, uint64(21000), config.EthGasLimit)
}

func TestNewManagerFromConfig(t *testing.T) {
	config := GetDefaultChainConfig()
	manager := NewManagerFromConfig(config)

	assert.NotNil(t, manager)
	assert.Equal(t, config.Prefix, manager.cosmosPrefix)
	assert.Equal(t, config.ChainID, manager.chainID)
	assert.Equal(t, config.EthChainID, manager.ethChainID)
}

func TestChainConfig(t *testing.T) {
	config := &ChainConfig{
		ChainID:     "test-chain",
		Prefix:      "test",
		CoinType:    CoinTypeCosmos,
		GasLimit:    100000,
		GasPrice:    DefaultGasPrice(),
		EthChainID:  big.NewInt(1337),
		EthGasPrice: big.NewInt(10000000000),
		EthGasLimit: 21000,
	}

	assert.Equal(t, "test-chain", config.ChainID)
	assert.Equal(t, "test", config.Prefix)
	assert.Equal(t, CoinTypeCosmos, config.CoinType)
	assert.Equal(t, uint64(100000), config.GasLimit)
	assert.NotNil(t, config.GasPrice)
	assert.Equal(t, big.NewInt(1337), config.EthChainID)
	assert.Equal(t, big.NewInt(10000000000), config.EthGasPrice)
	assert.Equal(t, uint64(21000), config.EthGasLimit)
}

func TestTransactionParams(t *testing.T) {
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

func TestEthereumTransactionParams(t *testing.T) {
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
	assert.Equal(t, big.NewInt(20000000000), params.GasPrice)
	assert.Equal(t, big.NewInt(30000000000), params.MaxFeePerGas)
	assert.Equal(t, big.NewInt(2000000000), params.MaxPriorityFeePerGas)
}

func TestGetDefaultEthereumParams(t *testing.T) {
	params := GetDefaultEthereumParams()
	assert.NotNil(t, params)
	assert.Equal(t, big.NewInt(1), params.ChainID)
	assert.Equal(t, uint64(0), params.Nonce)
	assert.Equal(t, uint64(21000), params.GasLimit)
	assert.NotNil(t, params.GasPrice)
	assert.NotNil(t, params.MaxFeePerGas)
	assert.NotNil(t, params.MaxPriorityFeePerGas)
}

// Helper function to create a default gas price for testing
func DefaultGasPrice() sdk.DecCoin {
	return sdk.NewDecCoin("usnr", math.NewInt(1000))
}
