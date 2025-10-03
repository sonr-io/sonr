package coins

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeedFromEntropy(t *testing.T) {
	did := "did:example:123456789abcdef"
	salt := "test-salt"

	seed1 := SeedFromEntropy(did, salt)
	seed2 := SeedFromEntropy(did, salt)

	// Seeds should be deterministic
	assert.Equal(t, seed1, seed2, "Seeds should be deterministic")

	// Different inputs should produce different seeds
	seed3 := SeedFromEntropy(did, "different-salt")
	assert.NotEqual(t, seed1, seed3, "Different salts should produce different seeds")

	seed4 := SeedFromEntropy("did:example:different", salt)
	assert.NotEqual(t, seed1, seed4, "Different DIDs should produce different seeds")
}

func TestDerivationPath(t *testing.T) {
	cosmosPath := DefaultCosmosPath()
	ethPath := DefaultEthereumPath()

	assert.Equal(t, uint32(44), cosmosPath.Purpose)
	assert.Equal(t, uint32(44), ethPath.Purpose)

	assert.Equal(t, CoinTypeCosmos, cosmosPath.CoinType)
	assert.Equal(t, CoinTypeEthereum, ethPath.CoinType)

	// Test string representation
	cosmosStr := cosmosPath.String()
	ethStr := ethPath.String()

	assert.Contains(t, cosmosStr, "m/44'/118'/0'/0/0")
	assert.Contains(t, ethStr, "m/44'/60'/0'/0/0")
}

func TestMasterKeyFromSeed(t *testing.T) {
	seed := SeedFromEntropy("did:example:123", "test-salt")

	masterKey, err := MasterKeyFromSeed(seed)
	require.NoError(t, err)
	assert.NotNil(t, masterKey)

	// Test with same seed produces same key
	masterKey2, err := MasterKeyFromSeed(seed)
	require.NoError(t, err)
	assert.Equal(t, masterKey.String(), masterKey2.String())
}

func TestDeriveKey(t *testing.T) {
	seed := SeedFromEntropy("did:example:123", "test-salt")
	masterKey, err := MasterKeyFromSeed(seed)
	require.NoError(t, err)

	cosmosPath := DefaultCosmosPath()
	cosmosKey, err := DeriveKey(masterKey, cosmosPath)
	require.NoError(t, err)
	assert.NotNil(t, cosmosKey)

	ethPath := DefaultEthereumPath()
	ethKey, err := DeriveKey(masterKey, ethPath)
	require.NoError(t, err)
	assert.NotNil(t, ethKey)

	// Keys should be different
	assert.NotEqual(t, cosmosKey.String(), ethKey.String())
}

func TestCosmosAddressFromKey(t *testing.T) {
	seed := SeedFromEntropy("did:example:123", "test-salt")
	masterKey, err := MasterKeyFromSeed(seed)
	require.NoError(t, err)

	cosmosPath := DefaultCosmosPath()
	cosmosKey, err := DeriveKey(masterKey, cosmosPath)
	require.NoError(t, err)

	// Test with default prefix
	addr1, err := CosmosAddressFromKey(cosmosKey, "")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(addr1, "cosmos"))

	// Test with custom prefix
	addr2, err := CosmosAddressFromKey(cosmosKey, "snr")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(addr2, "snr"))

	// Addresses should be different with different prefixes
	assert.NotEqual(t, addr1, addr2)
}

func TestEthereumAddressFromKey(t *testing.T) {
	seed := SeedFromEntropy("did:example:123", "test-salt")
	masterKey, err := MasterKeyFromSeed(seed)
	require.NoError(t, err)

	ethPath := DefaultEthereumPath()
	ethKey, err := DeriveKey(masterKey, ethPath)
	require.NoError(t, err)

	addr, err := EthereumAddressFromKey(ethKey)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(addr, "0x"))
	assert.Len(t, addr, 42) // 0x + 40 hex chars
}

func TestPrivateKeyFromExtendedKey(t *testing.T) {
	seed := SeedFromEntropy("did:example:123", "test-salt")
	masterKey, err := MasterKeyFromSeed(seed)
	require.NoError(t, err)

	cosmosPath := DefaultCosmosPath()
	cosmosKey, err := DeriveKey(masterKey, cosmosPath)
	require.NoError(t, err)

	privKey, err := PrivateKeyFromExtendedKey(cosmosKey)
	require.NoError(t, err)
	assert.NotNil(t, privKey)
	assert.NotNil(t, privKey.PublicKey)
}

func TestDeriveAddressesFromEntropy(t *testing.T) {
	did := "did:example:123456789abcdef"
	salt := "test-salt"
	prefix := "snr"

	cosmosAddr, ethAddr, derivationPath, err := DeriveAddressesFromEntropy(did, salt, prefix)
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(cosmosAddr, prefix))
	assert.True(t, strings.HasPrefix(ethAddr, "0x"))
	assert.Contains(t, derivationPath, "m/44'/118'/0'/0/0")

	// Test deterministic generation
	cosmosAddr2, ethAddr2, derivationPath2, err := DeriveAddressesFromEntropy(did, salt, prefix)
	require.NoError(t, err)

	assert.Equal(t, cosmosAddr, cosmosAddr2)
	assert.Equal(t, ethAddr, ethAddr2)
	assert.Equal(t, derivationPath, derivationPath2)
}

func TestWalletFromEntropy(t *testing.T) {
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
	assert.Contains(t, wallet.DerivationPath, "m/44'/118'/0'/0/0")

	// Test deterministic generation
	wallet2, err := WalletFromEntropy(did, salt, prefix)
	require.NoError(t, err)

	assert.Equal(t, wallet.CosmosAddress, wallet2.CosmosAddress)
	assert.Equal(t, wallet.EthereumAddress, wallet2.EthereumAddress)
}

func TestCoinTypeConstants(t *testing.T) {
	assert.Equal(t, uint32(118), CoinTypeCosmos)
	assert.Equal(t, uint32(60), CoinTypeEthereum)
	assert.Equal(t, uint32(60), CoinTypeSonr)
}
