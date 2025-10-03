// Package bip39 provides BIP39 mnemonic functionality.
// This is a stub package to replace the missing tyler-smith/go-bip39 repository.
// The original repository at github.com/tyler-smith/go-bip39 no longer exists.
package bip39

import (
	"errors"
)

// NewMnemonic generates a new mnemonic sequence
func NewMnemonic(bitSize int) (string, error) {
	return "", errors.New("tyler-smith/go-bip39 is deprecated, use github.com/cosmos/go-bip39 instead")
}

// NewSeed creates a seed from a mnemonic and passphrase
func NewSeed(mnemonic string, password string) []byte {
	panic("tyler-smith/go-bip39 is deprecated, use github.com/cosmos/go-bip39 instead")
}

// IsMnemonicValid validates a mnemonic sequence
func IsMnemonicValid(mnemonic string) bool {
	return false
}

// NewEntropy generates new entropy
func NewEntropy(bitSize int) ([]byte, error) {
	return nil, errors.New("tyler-smith/go-bip39 is deprecated, use github.com/cosmos/go-bip39 instead")
}

// NewMnemonicFromEntropy creates a mnemonic from entropy
func NewMnemonicFromEntropy(entropy []byte) (string, error) {
	return "", errors.New("tyler-smith/go-bip39 is deprecated, use github.com/cosmos/go-bip39 instead")
}
