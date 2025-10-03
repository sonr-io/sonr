package mpc

import (
	"encoding/hex"
	"errors"
	"fmt"
)

// ImportEnclave creates an Enclave instance from various import options.
// It prioritizes enclave bytes over keyshares if both are provided.
func ImportEnclave(options ...ImportOption) (Enclave, error) {
	if len(options) == 0 {
		return nil, errors.New("no import options provided")
	}

	opts := Options{}
	for _, opt := range options {
		opts = opt(opts)
	}
	return opts.Apply()
}

// Options is a struct that holds the import options
type Options struct {
	valKeyshare   Message
	userKeyshare  Message
	enclaveBytes  []byte
	enclaveData   *EnclaveData
	initialShares bool
	isEncrypted   bool
	secretKey     []byte
	curve         CurveName
}

// ImportOption is a function that modifies the import options
type ImportOption func(Options) Options

// WithInitialShares creates an option to import an enclave from validator and user keyshares.
func WithInitialShares(valKeyshare Message, userKeyshare Message, curve CurveName) ImportOption {
	return func(opts Options) Options {
		opts.valKeyshare = valKeyshare
		opts.userKeyshare = userKeyshare
		opts.initialShares = true
		opts.curve = curve
		return opts
	}
}

// WithEncryptedData creates an option to import an enclave from encrypted data.
func WithEncryptedData(data []byte, key []byte) ImportOption {
	return func(opts Options) Options {
		opts.enclaveBytes = data
		opts.initialShares = false
		opts.isEncrypted = true
		opts.secretKey = key
		return opts
	}
}

// WithEnclaveData creates an option to import an enclave from a data struct.
func WithEnclaveData(data *EnclaveData) ImportOption {
	return func(opts Options) Options {
		opts.enclaveData = data
		opts.initialShares = false
		return opts
	}
}

// Apply applies the import options to create an Enclave instance.
func (opts Options) Apply() (Enclave, error) {
	// Load from encrypted data if provided
	if opts.isEncrypted {
		if len(opts.enclaveBytes) == 0 {
			return nil, errors.New("enclave bytes cannot be empty")
		}
		return RestoreEncryptedEnclave(opts.enclaveBytes, opts.secretKey)
	}
	// Generate from keyshares if provided
	if opts.initialShares {
		// Then try to build from keyshares
		if opts.valKeyshare == nil {
			return nil, errors.New("validator share cannot be nil")
		}
		if opts.userKeyshare == nil {
			return nil, errors.New("user share cannot be nil")
		}
		return BuildEnclave(opts.valKeyshare, opts.userKeyshare, opts)
	}
	// Load from enclave data if provided
	return RestoreEnclaveFromData(opts.enclaveData)
}

// BuildEnclave creates a new enclave from validator and user keyshares.
func BuildEnclave(valShare, userShare Message, options Options) (Enclave, error) {
	if valShare == nil {
		return nil, errors.New("validator share cannot be nil")
	}
	if userShare == nil {
		return nil, errors.New("user share cannot be nil")
	}

	pubPoint, err := GetAlicePublicPoint(valShare)
	if err != nil {
		return nil, fmt.Errorf("failed to get public point: %w", err)
	}
	return &EnclaveData{
		PubBytes:  pubPoint.ToAffineUncompressed(),
		PubHex:    hex.EncodeToString(pubPoint.ToAffineCompressed()),
		ValShare:  valShare,
		UserShare: userShare,
		Nonce:     randNonce(),
		Curve:     options.curve,
	}, nil
}

// RestoreEnclaveFromData deserializes an enclave from its data struct.
func RestoreEnclaveFromData(data *EnclaveData) (Enclave, error) {
	if data == nil {
		return nil, errors.New("enclave data cannot be nil")
	}
	return data, nil
}

// RestoreEncryptedEnclave decrypts an enclave from its binary representation. and key
func RestoreEncryptedEnclave(data []byte, key []byte) (Enclave, error) {
	keyclave := &EnclaveData{}
	err := keyclave.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal enclave: %w", err)
	}
	decryptedData, err := keyclave.Decrypt(key, data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt enclave: %w", err)
	}
	err = keyclave.Unmarshal(decryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted enclave: %w", err)
	}
	return keyclave, nil
}
