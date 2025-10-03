package txns

import "errors"

var (
	// ErrUnsupportedChainType is returned when an unsupported chain type is used
	ErrUnsupportedChainType = errors.New("unsupported chain type")

	// ErrInvalidTransactionParams is returned when transaction parameters are invalid
	ErrInvalidTransactionParams = errors.New("invalid transaction parameters")

	// ErrInvalidGasParams is returned when gas parameters are invalid
	ErrInvalidGasParams = errors.New("invalid gas parameters")

	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errors.New("invalid address")

	// ErrInvalidSignature is returned when a signature is invalid
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrSigningFailed is returned when transaction signing fails
	ErrSigningFailed = errors.New("transaction signing failed")

	// ErrEncodingFailed is returned when transaction encoding fails
	ErrEncodingFailed = errors.New("transaction encoding failed")

	// ErrDecodingFailed is returned when transaction decoding fails
	ErrDecodingFailed = errors.New("transaction decoding failed")

	// ErrFeeEstimationFailed is returned when fee estimation fails
	ErrFeeEstimationFailed = errors.New("fee estimation failed")

	// ErrInsufficientFunds is returned when account has insufficient funds
	ErrInsufficientFunds = errors.New("insufficient funds")

	// ErrNonceTooLow is returned when transaction nonce is too low
	ErrNonceTooLow = errors.New("nonce too low")

	// ErrNonceTooHigh is returned when transaction nonce is too high
	ErrNonceTooHigh = errors.New("nonce too high")

	// ErrGasPriceTooLow is returned when gas price is too low
	ErrGasPriceTooLow = errors.New("gas price too low")

	// ErrGasLimitTooLow is returned when gas limit is too low
	ErrGasLimitTooLow = errors.New("gas limit too low")

	// ErrGasLimitTooHigh is returned when gas limit is too high
	ErrGasLimitTooHigh = errors.New("gas limit too high")

	// ErrTransactionTooLarge is returned when transaction is too large
	ErrTransactionTooLarge = errors.New("transaction too large")

	// ErrMemoTooLarge is returned when transaction memo is too large
	ErrMemoTooLarge = errors.New("memo too large")

	// ErrTimeoutHeightInvalid is returned when timeout height is invalid
	ErrTimeoutHeightInvalid = errors.New("timeout height invalid")

	// ErrAccountNotFound is returned when account is not found
	ErrAccountNotFound = errors.New("account not found")

	// ErrSequenceMismatch is returned when account sequence doesn't match
	ErrSequenceMismatch = errors.New("sequence mismatch")

	// ErrChainIDMismatch is returned when chain ID doesn't match
	ErrChainIDMismatch = errors.New("chain ID mismatch")

	// ErrInvalidPublicKey is returned when public key is invalid
	ErrInvalidPublicKey = errors.New("invalid public key")

	// ErrMPCEnclaveNotInitialized is returned when MPC enclave is not initialized
	ErrMPCEnclaveNotInitialized = errors.New("MPC enclave not initialized")

	// ErrAddressDerivationFailed is returned when address derivation fails
	ErrAddressDerivationFailed = errors.New("address derivation failed")

	// ErrUnsupportedEncodingType is returned when encoding type is not supported
	ErrUnsupportedEncodingType = errors.New("unsupported encoding type")

	// ErrTransactionExpired is returned when transaction has expired
	ErrTransactionExpired = errors.New("transaction expired")

	// ErrInvalidContractCall is returned when contract call parameters are invalid
	ErrInvalidContractCall = errors.New("invalid contract call")

	// ErrSimulationFailed is returned when transaction simulation fails
	ErrSimulationFailed = errors.New("transaction simulation failed")
)
