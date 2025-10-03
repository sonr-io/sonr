// Package errors defines error types and utilities for the Sonr client SDK.
package errors

import (
	"errors"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
)

// Common error codes for the Sonr client SDK
const (
	// Connection errors
	CodeConnectionFailed uint32 = 1001 + iota
	CodeInvalidEndpoint
	CodeTimeout
	CodeNetworkUnreachable

	// Authentication errors
	CodeInvalidCredentials uint32 = 2001 + iota
	CodeKeyNotFound
	CodeSigningFailed
	CodeWebAuthnFailed

	// Transaction errors
	CodeInvalidTransaction uint32 = 3001 + iota
	CodeInsufficientFunds
	CodeGasEstimationFailed
	CodeBroadcastFailed
	CodeTransactionFailed

	// Query errors
	CodeQueryFailed uint32 = 4001 + iota
	CodeInvalidRequest
	CodeNotFound
	CodeUnauthorized

	// Module-specific errors
	CodeDIDError uint32 = 5001 + iota
	CodeDWNError
	CodeSVCError
	CodeUCANError

	// Configuration errors
	CodeInvalidConfig uint32 = 6001 + iota
	CodeMissingConfig
	CodeInvalidNetwork
)

var (
	// Connection errors
	ErrConnectionFailed   = sdkerrors.Register("sonr_client", CodeConnectionFailed, "failed to connect to endpoint")
	ErrInvalidEndpoint    = sdkerrors.Register("sonr_client", CodeInvalidEndpoint, "invalid endpoint configuration")
	ErrTimeout            = sdkerrors.Register("sonr_client", CodeTimeout, "request timeout")
	ErrNetworkUnreachable = sdkerrors.Register("sonr_client", CodeNetworkUnreachable, "network unreachable")

	// Authentication errors
	ErrInvalidCredentials = sdkerrors.Register("sonr_client", CodeInvalidCredentials, "invalid credentials")
	ErrKeyNotFound        = sdkerrors.Register("sonr_client", CodeKeyNotFound, "key not found in keyring")
	ErrSigningFailed      = sdkerrors.Register("sonr_client", CodeSigningFailed, "transaction signing failed")
	ErrWebAuthnFailed     = sdkerrors.Register("sonr_client", CodeWebAuthnFailed, "WebAuthn operation failed")

	// Transaction errors
	ErrInvalidTransaction  = sdkerrors.Register("sonr_client", CodeInvalidTransaction, "invalid transaction")
	ErrInsufficientFunds   = sdkerrors.Register("sonr_client", CodeInsufficientFunds, "insufficient funds")
	ErrGasEstimationFailed = sdkerrors.Register("sonr_client", CodeGasEstimationFailed, "gas estimation failed")
	ErrBroadcastFailed     = sdkerrors.Register("sonr_client", CodeBroadcastFailed, "transaction broadcast failed")
	ErrTransactionFailed   = sdkerrors.Register("sonr_client", CodeTransactionFailed, "transaction execution failed")

	// Query errors
	ErrQueryFailed    = sdkerrors.Register("sonr_client", CodeQueryFailed, "query execution failed")
	ErrInvalidRequest = sdkerrors.Register("sonr_client", CodeInvalidRequest, "invalid request parameters")
	ErrNotFound       = sdkerrors.Register("sonr_client", CodeNotFound, "resource not found")
	ErrUnauthorized   = sdkerrors.Register("sonr_client", CodeUnauthorized, "unauthorized access")

	// Module-specific errors
	ErrDIDError  = sdkerrors.Register("sonr_client", CodeDIDError, "DID module error")
	ErrDWNError  = sdkerrors.Register("sonr_client", CodeDWNError, "DWN module error")
	ErrSVCError  = sdkerrors.Register("sonr_client", CodeSVCError, "SVC module error")
	ErrUCANError = sdkerrors.Register("sonr_client", CodeUCANError, "UCAN module error")

	// Configuration errors
	ErrInvalidConfig  = sdkerrors.Register("sonr_client", CodeInvalidConfig, "invalid configuration")
	ErrMissingConfig  = sdkerrors.Register("sonr_client", CodeMissingConfig, "missing required configuration")
	ErrInvalidNetwork = sdkerrors.Register("sonr_client", CodeInvalidNetwork, "invalid network configuration")
)

// WrapError wraps an existing error with additional context and a Sonr-specific error code.
// This follows the Cosmos SDK pattern for error handling.
func WrapError(err error, sdkErr *sdkerrors.Error, format string, args ...any) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(format, args...)
	return sdkerrors.Wrapf(sdkErr, "%s: %v", msg, err)
}

// IsConnectionError returns true if the error is related to network connectivity.
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionFailed) ||
		errors.Is(err, ErrInvalidEndpoint) ||
		errors.Is(err, ErrTimeout) ||
		errors.Is(err, ErrNetworkUnreachable)
}

// IsAuthenticationError returns true if the error is related to authentication.
func IsAuthenticationError(err error) bool {
	return errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrKeyNotFound) ||
		errors.Is(err, ErrSigningFailed) ||
		errors.Is(err, ErrWebAuthnFailed)
}

// IsTransactionError returns true if the error is related to transaction processing.
func IsTransactionError(err error) bool {
	return errors.Is(err, ErrInvalidTransaction) ||
		errors.Is(err, ErrInsufficientFunds) ||
		errors.Is(err, ErrGasEstimationFailed) ||
		errors.Is(err, ErrBroadcastFailed) ||
		errors.Is(err, ErrTransactionFailed)
}

// IsQueryError returns true if the error is related to query operations.
func IsQueryError(err error) bool {
	return errors.Is(err, ErrQueryFailed) ||
		errors.Is(err, ErrInvalidRequest) ||
		errors.Is(err, ErrNotFound) ||
		errors.Is(err, ErrUnauthorized)
}

// IsConfigurationError returns true if the error is related to configuration.
func IsConfigurationError(err error) bool {
	return errors.Is(err, ErrInvalidConfig) ||
		errors.Is(err, ErrMissingConfig) ||
		errors.Is(err, ErrInvalidNetwork)
}

// GetErrorCode extracts the error code from a Cosmos SDK error.
// Returns 0 if the error is not a Cosmos SDK error or doesn't have a code.
func GetErrorCode(err error) uint32 {
	var sdkErr *sdkerrors.Error
	if errors.As(err, &sdkErr) {
		return sdkErr.ABCICode()
	}
	return 0
}

// NewConnectionError creates a new connection-related error with context.
func NewConnectionError(endpoint string, underlying error) error {
	return WrapError(underlying, ErrConnectionFailed, "failed to connect to %s", endpoint)
}

// NewAuthenticationError creates a new authentication-related error with context.
func NewAuthenticationError(operation string, underlying error) error {
	return WrapError(underlying, ErrInvalidCredentials, "authentication failed for %s", operation)
}

// NewTransactionError creates a new transaction-related error with context.
func NewTransactionError(txHash string, underlying error) error {
	return WrapError(underlying, ErrTransactionFailed, "transaction %s failed", txHash)
}

// NewQueryError creates a new query-related error with context.
func NewQueryError(query string, underlying error) error {
	return WrapError(underlying, ErrQueryFailed, "query %s failed", query)
}

// NewModuleError creates a new module-specific error with context.
func NewModuleError(module string, operation string, underlying error) error {
	var baseErr *sdkerrors.Error

	switch module {
	case "did":
		baseErr = ErrDIDError
	case "dwn":
		baseErr = ErrDWNError
	case "svc":
		baseErr = ErrSVCError
	case "ucan":
		baseErr = ErrUCANError
	default:
		baseErr = ErrQueryFailed
	}

	return WrapError(underlying, baseErr, "%s module %s operation failed", module, operation)
}
