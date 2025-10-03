package types

import (
	"fmt"
	"strings"

	"cosmossdk.io/errors"
)

// BlockchainAccountID represents a blockchain account identifier following CAIP-10 standard
// Format: <namespace>:<chain_id>:<address>
type BlockchainAccountID struct {
	Namespace string // "eip155" for Ethereum, "cosmos" for Cosmos chains
	ChainID   string // "1" for Ethereum mainnet, "cosmoshub-4" for Cosmos Hub
	Address   string // The account address
}

// String returns the CAIP-10 formatted blockchain account ID
func (b BlockchainAccountID) String() string {
	return fmt.Sprintf("%s:%s:%s", b.Namespace, b.ChainID, b.Address)
}

// ParseBlockchainAccountID parses a CAIP-10 formatted blockchain account ID
func ParseBlockchainAccountID(accountID string) (*BlockchainAccountID, error) {
	parts := strings.Split(accountID, ":")
	if len(parts) != 3 {
		return nil, errors.Wrapf(ErrInvalidBlockchainAccountID, "invalid format: %s", accountID)
	}

	return &BlockchainAccountID{
		Namespace: parts[0],
		ChainID:   parts[1],
		Address:   parts[2],
	}, nil
}

// Validate checks if the blockchain account ID is valid
func (b BlockchainAccountID) Validate() error {
	if b.Namespace == "" {
		return errors.Wrap(ErrInvalidBlockchainAccountID, "namespace cannot be empty")
	}
	if b.ChainID == "" {
		return errors.Wrap(ErrInvalidBlockchainAccountID, "chain_id cannot be empty")
	}
	if b.Address == "" {
		return errors.Wrap(ErrInvalidBlockchainAccountID, "address cannot be empty")
	}

	// Validate specific namespaces
	switch b.Namespace {
	case "eip155":
		return b.validateEIP155Address()
	case "cosmos":
		return b.validateCosmosAddress()
	default:
		return errors.Wrapf(ErrUnsupportedBlockchainNamespace, "namespace: %s", b.Namespace)
	}
}

// validateEIP155Address validates Ethereum addresses
func (b BlockchainAccountID) validateEIP155Address() error {
	if !strings.HasPrefix(b.Address, "0x") {
		return errors.Wrap(ErrInvalidEthereumAddress, "address must start with 0x")
	}
	if len(b.Address) != 42 { // 0x + 40 hex characters
		return errors.Wrap(ErrInvalidEthereumAddress, "address must be 42 characters long")
	}

	// Check if all characters after 0x are valid hex
	for _, r := range b.Address[2:] {
		if !isHexChar(r) {
			return errors.Wrap(ErrInvalidEthereumAddress, "address contains invalid hex characters")
		}
	}

	return nil
}

// validateCosmosAddress validates Cosmos addresses
func (b BlockchainAccountID) validateCosmosAddress() error {
	// Basic validation - Cosmos addresses typically start with a prefix
	if len(b.Address) < 10 {
		return errors.Wrap(ErrInvalidCosmosAddress, "address too short")
	}

	// More detailed validation could be added here based on bech32 format
	// For now, we'll do basic length and character checks
	if len(b.Address) > 100 {
		return errors.Wrap(ErrInvalidCosmosAddress, "address too long")
	}

	return nil
}

// isHexChar checks if a rune is a valid hexadecimal character
func isHexChar(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
}

// WalletType represents the type of external wallet
type WalletType string

const (
	WalletTypeEthereum WalletType = "ethereum"
	WalletTypeCosmos   WalletType = "cosmos"
)

// String returns the string representation of WalletType
func (w WalletType) String() string {
	return string(w)
}

// Validate checks if the wallet type is supported
func (w WalletType) Validate() error {
	switch w {
	case WalletTypeEthereum, WalletTypeCosmos:
		return nil
	default:
		return errors.Wrapf(ErrUnsupportedWalletType, "wallet type: %s", w)
	}
}

// ToVerificationMethodType returns the W3C verification method type for the wallet
func (w WalletType) ToVerificationMethodType() string {
	switch w {
	case WalletTypeEthereum:
		return "EcdsaSecp256k1RecoveryMethod2020"
	case WalletTypeCosmos:
		return "Secp256k1VerificationKey2018"
	default:
		return "UnknownVerificationMethod"
	}
}

// GetNamespace returns the CAIP-10 namespace for the wallet type
func (w WalletType) GetNamespace() string {
	switch w {
	case WalletTypeEthereum:
		return "eip155"
	case WalletTypeCosmos:
		return "cosmos"
	default:
		return ""
	}
}

// WalletVerification contains verification data for wallet ownership proof
type WalletVerification struct {
	Challenge  []byte     // The challenge message that was signed
	Signature  []byte     // The signature proving ownership
	WalletType WalletType // Type of wallet
	Verified   bool       // Whether the verification was successful
}

// Validate checks if the wallet verification data is complete
func (wv WalletVerification) Validate() error {
	if len(wv.Challenge) == 0 {
		return errors.Wrap(ErrInvalidWalletVerification, "challenge cannot be empty")
	}
	if len(wv.Signature) == 0 {
		return errors.Wrap(ErrInvalidWalletVerification, "signature cannot be empty")
	}
	if err := wv.WalletType.Validate(); err != nil {
		return errors.Wrap(ErrInvalidWalletVerification, err.Error())
	}
	return nil
}
