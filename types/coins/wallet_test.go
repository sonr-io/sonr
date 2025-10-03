package coins

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalletCreation(t *testing.T) {
	did := "did:example:123456789abcdef"
	salt := "test-salt"
	prefix := "snr"

	wallet, err := WalletFromEntropy(did, salt, prefix)
	require.NoError(t, err)

	assert.Equal(t, did, wallet.DID)
	assert.Equal(t, salt, wallet.Salt)
	assert.True(t, strings.HasPrefix(wallet.CosmosAddress, prefix))
	assert.True(t, strings.HasPrefix(wallet.EthereumAddress, "0x"))
	assert.NotNil(t, wallet.CosmosPrivKey)
	assert.NotNil(t, wallet.EthereumPrivKey)
}

func TestWalletGetCosmosPublicKey(t *testing.T) {
	wallet, err := WalletFromEntropy("did:example:123", "test-salt", "snr")
	require.NoError(t, err)

	pubKey := wallet.GetCosmosPublicKey()
	assert.NotNil(t, pubKey)
	assert.NotNil(t, pubKey.Bytes())
}

func TestWalletGetEthereumPublicKey(t *testing.T) {
	wallet, err := WalletFromEntropy("did:example:123", "test-salt", "snr")
	require.NoError(t, err)

	pubKey := wallet.GetEthereumPublicKey()
	assert.NotNil(t, pubKey)
	assert.NotNil(t, pubKey.X)
	assert.NotNil(t, pubKey.Y)
}

func TestWalletSignEthereumTransaction(t *testing.T) {
	wallet, err := WalletFromEntropy("did:example:123", "test-salt", "snr")
	require.NoError(t, err)

	// Create a simple Ethereum transaction
	to := common.HexToAddress("0x1234567890123456789012345678901234567890")
	amount := big.NewInt(1000000000000000000) // 1 ETH
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(20000000000) // 20 Gwei
	nonce := uint64(0)

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)
	chainID := big.NewInt(1)

	signedTx, err := wallet.SignEthereumTransaction(tx, chainID)
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

func TestWalletSignMessage(t *testing.T) {
	wallet, err := WalletFromEntropy("did:example:123", "test-salt", "snr")
	require.NoError(t, err)

	message := []byte("Hello, World!")
	signature, err := wallet.SignMessage(message)
	require.NoError(t, err)
	assert.NotNil(t, signature)
	assert.True(t, len(signature) > 0)
}

func TestWalletSignEthereumMessage(t *testing.T) {
	wallet, err := WalletFromEntropy("did:example:123", "test-salt", "snr")
	require.NoError(t, err)

	message := []byte("Hello, Ethereum!")
	signature, err := wallet.SignEthereumMessage(message)
	require.NoError(t, err)
	assert.NotNil(t, signature)
	assert.True(t, len(signature) > 0)
}

func TestWalletVerifySignature(t *testing.T) {
	wallet, err := WalletFromEntropy("did:example:123", "test-salt", "snr")
	require.NoError(t, err)

	message := []byte("Hello, World!")
	signature, err := wallet.SignMessage(message)
	require.NoError(t, err)

	// Verify the signature
	isValid := wallet.VerifySignature(message, signature)
	assert.True(t, isValid)

	// Verify with wrong message should fail
	wrongMessage := []byte("Hello, Wrong!")
	isValid = wallet.VerifySignature(wrongMessage, signature)
	assert.False(t, isValid)
}

func TestWalletVerifyEthereumSignature(t *testing.T) {
	wallet, err := WalletFromEntropy("did:example:123", "test-salt", "snr")
	require.NoError(t, err)

	message := []byte("Hello, Ethereum!")
	signature, err := wallet.SignEthereumMessage(message)
	require.NoError(t, err)

	// Verify the signature
	isValid := wallet.VerifyEthereumSignature(message, signature)
	assert.True(t, isValid)

	// Verify with wrong message should fail
	wrongMessage := []byte("Hello, Wrong!")
	isValid = wallet.VerifyEthereumSignature(wrongMessage, signature)
	assert.False(t, isValid)
}

func TestWalletGetCosmosAccountAddress(t *testing.T) {
	wallet, err := WalletFromEntropy("did:example:123", "test-salt", "cosmos")
	require.NoError(t, err)

	addr, err := wallet.GetCosmosAccountAddress()
	require.NoError(t, err)
	assert.NotNil(t, addr)
	assert.True(t, len(addr) > 0)
}

func TestWalletGetEthereumAccountAddress(t *testing.T) {
	wallet, err := WalletFromEntropy("did:example:123", "test-salt", "snr")
	require.NoError(t, err)

	addr := wallet.GetEthereumAccountAddress()
	assert.NotEqual(t, common.Address{}, addr)
	assert.Equal(t, wallet.EthereumAddress, addr.Hex())
}

func TestWalletExportPrivateKeys(t *testing.T) {
	wallet, err := WalletFromEntropy("did:example:123", "test-salt", "snr")
	require.NoError(t, err)

	cosmosPrivKey, ethPrivKey := wallet.ExportPrivateKeys()
	assert.True(t, len(cosmosPrivKey) > 0)
	assert.True(t, len(ethPrivKey) > 0)
	assert.NotEqual(t, cosmosPrivKey, ethPrivKey)
}

func TestWalletGetInfo(t *testing.T) {
	did := "did:example:123456789abcdef"
	salt := "test-salt"
	prefix := "snr"

	wallet, err := WalletFromEntropy(did, salt, prefix)
	require.NoError(t, err)

	info := wallet.GetInfo()
	assert.Equal(t, did, info.DID)
	assert.Equal(t, salt, info.Salt)
	assert.Equal(t, wallet.CosmosAddress, info.CosmosAddress)
	assert.Equal(t, wallet.EthereumAddress, info.EthereumAddress)
	assert.Equal(t, wallet.DerivationPath, info.DerivationPath)
}

func TestWalletGetEthereumTransactor(t *testing.T) {
	wallet, err := WalletFromEntropy("did:example:123", "test-salt", "snr")
	require.NoError(t, err)

	chainID := big.NewInt(1)
	transactor, err := wallet.GetEthereumTransactor(chainID)
	require.NoError(t, err)
	assert.NotNil(t, transactor)
	assert.NotNil(t, transactor)
}

func TestWalletDeterministic(t *testing.T) {
	did := "did:example:123456789abcdef"
	salt := "test-salt"
	prefix := "snr"

	// Create two wallets with same parameters
	wallet1, err := WalletFromEntropy(did, salt, prefix)
	require.NoError(t, err)

	wallet2, err := WalletFromEntropy(did, salt, prefix)
	require.NoError(t, err)

	// They should be identical
	assert.Equal(t, wallet1.DID, wallet2.DID)
	assert.Equal(t, wallet1.Salt, wallet2.Salt)
	assert.Equal(t, wallet1.CosmosAddress, wallet2.CosmosAddress)
	assert.Equal(t, wallet1.EthereumAddress, wallet2.EthereumAddress)
	assert.Equal(t, wallet1.DerivationPath, wallet2.DerivationPath)

	// Private keys should be the same
	cosmosPrivKey1, ethPrivKey1 := wallet1.ExportPrivateKeys()
	cosmosPrivKey2, ethPrivKey2 := wallet2.ExportPrivateKeys()
	assert.Equal(t, cosmosPrivKey1, cosmosPrivKey2)
	assert.Equal(t, ethPrivKey1, ethPrivKey2)
}

func TestWalletDifferentInputs(t *testing.T) {
	prefix := "snr"

	// Create wallets with different DIDs
	wallet1, err := WalletFromEntropy("did:example:123", "test-salt", prefix)
	require.NoError(t, err)

	wallet2, err := WalletFromEntropy("did:example:456", "test-salt", prefix)
	require.NoError(t, err)

	// They should be different
	assert.NotEqual(t, wallet1.CosmosAddress, wallet2.CosmosAddress)
	assert.NotEqual(t, wallet1.EthereumAddress, wallet2.EthereumAddress)

	// Create wallets with different salts
	wallet3, err := WalletFromEntropy("did:example:123", "different-salt", prefix)
	require.NoError(t, err)

	// They should be different
	assert.NotEqual(t, wallet1.CosmosAddress, wallet3.CosmosAddress)
	assert.NotEqual(t, wallet1.EthereumAddress, wallet3.EthereumAddress)
}
