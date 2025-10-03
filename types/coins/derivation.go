package coins

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"github.com/ethereum/go-ethereum/crypto"
)

// CoinType constants for BIP44 derivation
const (
	CoinTypeCosmos   uint32 = 118 // Cosmos Hub
	CoinTypeEthereum uint32 = 60  // Ethereum
	CoinTypeSonr     uint32 = 60  // Sonr uses Ethereum coin type
)

// DerivationPath represents a BIP44 derivation path
type DerivationPath struct {
	Purpose      uint32 // 44 for BIP44
	CoinType     uint32 // 118 for Cosmos, 60 for Ethereum
	Account      uint32 // Account index
	Change       uint32 // 0 for external, 1 for internal
	AddressIndex uint32 // Address index
}

// String returns the string representation of the derivation path
func (dp DerivationPath) String() string {
	return fmt.Sprintf(
		"m/%d'/%d'/%d'/%d/%d",
		dp.Purpose,
		dp.CoinType,
		dp.Account,
		dp.Change,
		dp.AddressIndex,
	)
}

// DefaultCosmosPath returns the default derivation path for Cosmos
func DefaultCosmosPath() DerivationPath {
	return DerivationPath{
		Purpose:      44,
		CoinType:     CoinTypeCosmos,
		Account:      0,
		Change:       0,
		AddressIndex: 0,
	}
}

// DefaultEthereumPath returns the default derivation path for Ethereum
func DefaultEthereumPath() DerivationPath {
	return DerivationPath{
		Purpose:      44,
		CoinType:     CoinTypeEthereum,
		Account:      0,
		Change:       0,
		AddressIndex: 0,
	}
}

// SeedFromMnemonic generates a seed from a mnemonic phrase
func SeedFromMnemonic(mnemonic, passphrase string) ([]byte, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}
	return bip39.NewSeed(mnemonic, passphrase), nil
}

// SeedFromEntropy generates a seed from entropy (DID + salt)
func SeedFromEntropy(did, salt string) []byte {
	entropy := fmt.Sprintf("%s:%s", did, salt)
	hash := sha256.Sum256([]byte(entropy))
	return hash[:]
}

// MasterKeyFromSeed generates a master key from seed
func MasterKeyFromSeed(seed []byte) (*hdkeychain.ExtendedKey, error) {
	// Use Bitcoin mainnet params for key derivation
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}
	return masterKey, nil
}

// DeriveKey derives a key at the given path from master key
func DeriveKey(
	masterKey *hdkeychain.ExtendedKey,
	path DerivationPath,
) (*hdkeychain.ExtendedKey, error) {
	// Derive purpose
	purpose, err := masterKey.Derive(hdkeychain.HardenedKeyStart + path.Purpose)
	if err != nil {
		return nil, fmt.Errorf("failed to derive purpose: %w", err)
	}

	// Derive coin type
	coinType, err := purpose.Derive(hdkeychain.HardenedKeyStart + path.CoinType)
	if err != nil {
		return nil, fmt.Errorf("failed to derive coin type: %w", err)
	}

	// Derive account
	account, err := coinType.Derive(hdkeychain.HardenedKeyStart + path.Account)
	if err != nil {
		return nil, fmt.Errorf("failed to derive account: %w", err)
	}

	// Derive change
	change, err := account.Derive(path.Change)
	if err != nil {
		return nil, fmt.Errorf("failed to derive change: %w", err)
	}

	// Derive address index
	addressKey, err := change.Derive(path.AddressIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address index: %w", err)
	}

	return addressKey, nil
}

// CosmosAddressFromKey generates a Cosmos address from an extended key
func CosmosAddressFromKey(key *hdkeychain.ExtendedKey, prefix string) (string, error) {
	pubKeyBytes, err := key.ECPubKey()
	if err != nil {
		return "", fmt.Errorf("failed to get public key: %w", err)
	}

	// Convert to Cosmos SDK format
	pubKey := pubKeyBytes.SerializeCompressed()

	// Generate address using SHA256 hash of public key
	hash := sha256.Sum256(pubKey)
	addr := sdk.AccAddress(hash[:20])

	// Convert to bech32 format with custom prefix
	if prefix == "" {
		prefix = "cosmos"
	}

	bech32Addr, err := sdk.Bech32ifyAddressBytes(prefix, addr)
	if err != nil {
		return "", fmt.Errorf("failed to convert to bech32: %w", err)
	}

	return bech32Addr, nil
}

// EthereumAddressFromKey generates an Ethereum address from an extended key
func EthereumAddressFromKey(key *hdkeychain.ExtendedKey) (string, error) {
	pubKeyBytes, err := key.ECPubKey()
	if err != nil {
		return "", fmt.Errorf("failed to get public key: %w", err)
	}

	// Convert to Ethereum format
	pubKey := pubKeyBytes.ToECDSA()
	address := crypto.PubkeyToAddress(*pubKey)

	return address.Hex(), nil
}

// PrivateKeyFromExtendedKey extracts the private key from an extended key
func PrivateKeyFromExtendedKey(key *hdkeychain.ExtendedKey) (*ecdsa.PrivateKey, error) {
	privKeyBytes, err := key.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	return privKeyBytes.ToECDSA(), nil
}

// DeriveAddressesFromEntropy derives both Cosmos and Ethereum addresses from DID and salt
func DeriveAddressesFromEntropy(
	did, salt, cosmosPrefix string,
) (cosmosAddr, ethAddr, derivationPath string, err error) {
	// Generate seed from DID and salt
	seed := SeedFromEntropy(did, salt)

	// Generate master key
	masterKey, err := MasterKeyFromSeed(seed)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate master key: %w", err)
	}

	// Derive Cosmos address
	cosmosPath := DefaultCosmosPath()
	cosmosKey, err := DeriveKey(masterKey, cosmosPath)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to derive Cosmos key: %w", err)
	}

	cosmosAddr, err = CosmosAddressFromKey(cosmosKey, cosmosPrefix)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate Cosmos address: %w", err)
	}

	// Derive Ethereum address
	ethPath := DefaultEthereumPath()
	ethKey, err := DeriveKey(masterKey, ethPath)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to derive Ethereum key: %w", err)
	}

	ethAddr, err = EthereumAddressFromKey(ethKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate Ethereum address: %w", err)
	}

	derivationPath = cosmosPath.String()
	return cosmosAddr, ethAddr, derivationPath, nil
}

// WalletFromEntropy creates a wallet with both Cosmos and Ethereum keys from entropy
func WalletFromEntropy(did, salt, cosmosPrefix string) (*Wallet, error) {
	seed := SeedFromEntropy(did, salt)

	masterKey, err := MasterKeyFromSeed(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}

	// Derive Cosmos key
	cosmosPath := DefaultCosmosPath()
	cosmosKey, err := DeriveKey(masterKey, cosmosPath)
	if err != nil {
		return nil, fmt.Errorf("failed to derive Cosmos key: %w", err)
	}

	cosmosPrivKey, err := PrivateKeyFromExtendedKey(cosmosKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get Cosmos private key: %w", err)
	}

	cosmosAddr, err := CosmosAddressFromKey(cosmosKey, cosmosPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Cosmos address: %w", err)
	}

	// Derive Ethereum key
	ethPath := DefaultEthereumPath()
	ethKey, err := DeriveKey(masterKey, ethPath)
	if err != nil {
		return nil, fmt.Errorf("failed to derive Ethereum key: %w", err)
	}

	ethPrivKey, err := PrivateKeyFromExtendedKey(ethKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get Ethereum private key: %w", err)
	}

	ethAddr, err := EthereumAddressFromKey(ethKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ethereum address: %w", err)
	}

	return &Wallet{
		DID:             did,
		Salt:            salt,
		CosmosAddress:   cosmosAddr,
		EthereumAddress: ethAddr,
		CosmosPrivKey:   cosmosPrivKey,
		EthereumPrivKey: ethPrivKey,
		DerivationPath:  cosmosPath.String(),
	}, nil
}
