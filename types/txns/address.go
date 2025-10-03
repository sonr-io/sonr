// Package txns implements transaction builders for Cosmos and EVM chains
package txns

import (
	"crypto/ecdsa"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonr-io/sonr/crypto/mpc"
	"github.com/sonr-io/sonr/types/coins"
)

// AddressDeriver interface for deriving addresses from public keys
type AddressDeriver interface {
	// DeriveCosmosAddress derives a Cosmos address from public key
	DeriveCosmosAddress(pubKey []byte, prefix string) (string, error)
	// DeriveEVMAddress derives an EVM address from public key
	DeriveEVMAddress(pubKey []byte) (string, error)
	// ValidateAddress validates an address for the given chain type
	ValidateAddress(address string, chainType TransactionType) error
	// GetAddressFormat returns the address format for the chain type
	GetAddressFormat(chainType TransactionType) string
}

// StandardAddressDeriver implements standard address derivation
type StandardAddressDeriver struct {
	cosmosPrefix string
}

// NewStandardAddressDeriver creates a new standard address deriver
func NewStandardAddressDeriver(cosmosPrefix string) *StandardAddressDeriver {
	return &StandardAddressDeriver{
		cosmosPrefix: cosmosPrefix,
	}
}

// DeriveCosmosAddress implements AddressDeriver interface
func (sad *StandardAddressDeriver) DeriveCosmosAddress(
	pubKey []byte,
	prefix string,
) (string, error) {
	if len(pubKey) != 33 {
		return "", fmt.Errorf("invalid public key length: expected 33, got %d", len(pubKey))
	}

	// Parse the compressed public key
	ecdsaPubKey, err := crypto.DecompressPubkey(pubKey)
	if err != nil {
		return "", fmt.Errorf("failed to decompress public key: %w", err)
	}

	// Convert to Cosmos SDK format and derive address
	pubKeyBytes := crypto.FromECDSAPub(ecdsaPubKey)
	addr := sdk.AccAddress(crypto.Keccak256(pubKeyBytes[1:])[12:]) // Last 20 bytes of hash

	// Use provided prefix or default
	if prefix == "" {
		prefix = sad.cosmosPrefix
	}

	// Convert to bech32 format
	bech32Addr, err := sdk.Bech32ifyAddressBytes(prefix, addr)
	if err != nil {
		return "", fmt.Errorf("failed to convert to bech32: %w", err)
	}

	return bech32Addr, nil
}

// DeriveEVMAddress implements AddressDeriver interface
func (sad *StandardAddressDeriver) DeriveEVMAddress(pubKey []byte) (string, error) {
	if len(pubKey) != 33 && len(pubKey) != 65 {
		return "", fmt.Errorf("invalid public key length: expected 33 or 65, got %d", len(pubKey))
	}

	var ecdsaPubKey *ecdsa.PublicKey
	var err error

	if len(pubKey) == 33 {
		// Compressed public key
		ecdsaPubKey, err = crypto.DecompressPubkey(pubKey)
		if err != nil {
			return "", fmt.Errorf("failed to decompress public key: %w", err)
		}
	} else {
		// Uncompressed public key
		ecdsaPubKey, err = crypto.UnmarshalPubkey(pubKey)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal public key: %w", err)
		}
	}

	// Derive Ethereum address
	address := crypto.PubkeyToAddress(*ecdsaPubKey)
	return address.Hex(), nil
}

// ValidateAddress implements AddressDeriver interface
func (sad *StandardAddressDeriver) ValidateAddress(
	address string,
	chainType TransactionType,
) error {
	switch chainType {
	case TransactionTypeCosmos:
		_, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			return fmt.Errorf("invalid Cosmos address: %w", err)
		}
	case TransactionTypeEVM:
		if !common.IsHexAddress(address) {
			return fmt.Errorf("invalid EVM address format")
		}
	default:
		return ErrUnsupportedChainType
	}
	return nil
}

// GetAddressFormat implements AddressDeriver interface
func (sad *StandardAddressDeriver) GetAddressFormat(chainType TransactionType) string {
	switch chainType {
	case TransactionTypeCosmos:
		return fmt.Sprintf("bech32 with prefix '%s'", sad.cosmosPrefix)
	case TransactionTypeEVM:
		return "hex format (0x...)"
	default:
		return "unknown"
	}
}

// MPCAddressDeriver derives addresses from MPC public keys
type MPCAddressDeriver struct {
	*StandardAddressDeriver
}

// NewMPCAddressDeriver creates a new MPC address deriver
func NewMPCAddressDeriver(cosmosPrefix string) *MPCAddressDeriver {
	return &MPCAddressDeriver{
		StandardAddressDeriver: NewStandardAddressDeriver(cosmosPrefix),
	}
}

// DeriveFromEnclaveData derives addresses from MPC enclave data
func (mad *MPCAddressDeriver) DeriveFromEnclaveData(
	enclaveData *mpc.EnclaveData,
	prefix string,
) (*AddressDerivation, error) {
	// Import the enclave to get the public key
	enclave, err := mpc.ImportEnclave(mpc.WithEnclaveData(enclaveData))
	if err != nil {
		return nil, fmt.Errorf("failed to import enclave: %w", err)
	}

	pubKey := enclave.PubKeyBytes()

	// Derive Cosmos address
	cosmosAddr, err := mad.DeriveCosmosAddress(pubKey, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to derive Cosmos address: %w", err)
	}

	// Derive EVM address
	evmAddr, err := mad.DeriveEVMAddress(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive EVM address: %w", err)
	}

	return &AddressDerivation{
		CosmosAddress:  cosmosAddr,
		EVMAddress:     evmAddr,
		DerivationPath: "mpc-enclave",
		PublicKey:      pubKey,
		ChainType:      "multi-chain",
	}, nil
}

// EntropyAddressDeriver derives addresses from entropy (DID + salt)
type EntropyAddressDeriver struct {
	*StandardAddressDeriver
	coinsManager *coins.Manager
}

// NewEntropyAddressDeriver creates a new entropy-based address deriver
func NewEntropyAddressDeriver(
	cosmosPrefix string,
	coinsManager *coins.Manager,
) *EntropyAddressDeriver {
	return &EntropyAddressDeriver{
		StandardAddressDeriver: NewStandardAddressDeriver(cosmosPrefix),
		coinsManager:           coinsManager,
	}
}

// DeriveFromEntropy derives addresses from DID and salt using the coins package
func (ead *EntropyAddressDeriver) DeriveFromEntropy(
	did, salt, prefix string,
) (*AddressDerivation, error) {
	// Use the coins manager to derive addresses
	cosmosAddr, evmAddr, derivationPath, err := ead.coinsManager.DeriveAddresses(did, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to derive addresses from entropy: %w", err)
	}

	// Create a temporary wallet to get the public key
	wallet, err := ead.coinsManager.CreateWalletFromEntropy(did, salt)
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

// AddressManager manages address derivation for multiple methods
type AddressManager struct {
	standardDeriver *StandardAddressDeriver
	mpcDeriver      *MPCAddressDeriver
	entropyDeriver  *EntropyAddressDeriver
	cosmosPrefix    string
}

// NewAddressManager creates a new address manager
func NewAddressManager(cosmosPrefix string, coinsManager *coins.Manager) *AddressManager {
	return &AddressManager{
		standardDeriver: NewStandardAddressDeriver(cosmosPrefix),
		mpcDeriver:      NewMPCAddressDeriver(cosmosPrefix),
		entropyDeriver:  NewEntropyAddressDeriver(cosmosPrefix, coinsManager),
		cosmosPrefix:    cosmosPrefix,
	}
}

// DeriveFromPublicKey derives addresses from a raw public key
func (am *AddressManager) DeriveFromPublicKey(
	pubKey []byte,
	prefix string,
) (*AddressDerivation, error) {
	cosmosAddr, err := am.standardDeriver.DeriveCosmosAddress(pubKey, prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to derive Cosmos address: %w", err)
	}

	evmAddr, err := am.standardDeriver.DeriveEVMAddress(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive EVM address: %w", err)
	}

	return &AddressDerivation{
		CosmosAddress:  cosmosAddr,
		EVMAddress:     evmAddr,
		DerivationPath: "direct-public-key",
		PublicKey:      pubKey,
		ChainType:      "multi-chain",
	}, nil
}

// DeriveFromMPCEnclave derives addresses from MPC enclave data
func (am *AddressManager) DeriveFromMPCEnclave(
	enclaveData *mpc.EnclaveData,
	prefix string,
) (*AddressDerivation, error) {
	return am.mpcDeriver.DeriveFromEnclaveData(enclaveData, prefix)
}

// DeriveFromEntropy derives addresses from DID and salt
func (am *AddressManager) DeriveFromEntropy(did, salt, prefix string) (*AddressDerivation, error) {
	return am.entropyDeriver.DeriveFromEntropy(did, salt, prefix)
}

// ValidateAddress validates an address for any supported chain type
func (am *AddressManager) ValidateAddress(address string, chainType TransactionType) error {
	return am.standardDeriver.ValidateAddress(address, chainType)
}

// GetSupportedChainTypes returns the supported chain types
func (am *AddressManager) GetSupportedChainTypes() []TransactionType {
	return []TransactionType{TransactionTypeCosmos, TransactionTypeEVM}
}

// ConvertAddress converts an address between different formats (if applicable)
func (am *AddressManager) ConvertAddress(
	address string,
	fromChain, toChain TransactionType,
) (string, error) {
	// For now, only validation since we can't directly convert between Cosmos and EVM addresses
	// without the underlying public key
	if fromChain == toChain {
		return address, nil
	}

	// Validate the source address
	err := am.ValidateAddress(address, fromChain)
	if err != nil {
		return "", fmt.Errorf("invalid source address: %w", err)
	}

	return "", fmt.Errorf(
		"direct address conversion between %s and %s is not supported without public key",
		fromChain,
		toChain,
	)
}

// AddressBatch represents a batch of addresses derived together
type AddressBatch struct {
	Addresses []AddressDerivation `json:"addresses"`
	Metadata  map[string]string   `json:"metadata"`
}

// DeriveAddressBatch derives multiple addresses from different sources
func (am *AddressManager) DeriveAddressBatch(requests []AddressRequest) (*AddressBatch, error) {
	batch := &AddressBatch{
		Addresses: make([]AddressDerivation, 0, len(requests)),
		Metadata:  make(map[string]string),
	}

	for i, req := range requests {
		var derivation *AddressDerivation
		var err error

		switch req.Type {
		case "public_key":
			derivation, err = am.DeriveFromPublicKey(req.PublicKey, req.Prefix)
		case "entropy":
			derivation, err = am.DeriveFromEntropy(req.DID, req.Salt, req.Prefix)
		case "mpc_enclave":
			if req.EnclaveData == nil {
				err = fmt.Errorf("enclave data required for MPC derivation")
			} else {
				derivation, err = am.DeriveFromMPCEnclave(req.EnclaveData, req.Prefix)
			}
		default:
			err = fmt.Errorf("unsupported derivation type: %s", req.Type)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to derive address for request %d: %w", i, err)
		}

		batch.Addresses = append(batch.Addresses, *derivation)
	}

	batch.Metadata["total_addresses"] = fmt.Sprintf("%d", len(batch.Addresses))
	batch.Metadata["cosmos_prefix"] = am.cosmosPrefix

	return batch, nil
}

// AddressRequest represents a request to derive addresses
type AddressRequest struct {
	Type        string           `json:"type"`
	PublicKey   []byte           `json:"public_key,omitempty"`
	DID         string           `json:"did,omitempty"`
	Salt        string           `json:"salt,omitempty"`
	EnclaveData *mpc.EnclaveData `json:"enclave_data,omitempty"`
	Prefix      string           `json:"prefix,omitempty"`
}
