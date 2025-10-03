package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sonr-io/sonr/x/dwn/types"
)

func TestErrorDefinitions(t *testing.T) {
	// Test that all errors are properly defined
	require.NotNil(t, types.ErrInvalidRequest)
	require.NotNil(t, types.ErrRequestCannotBeNil)
	require.NotNil(t, types.ErrTargetDIDEmpty)
	require.NotNil(t, types.ErrRecordIDEmpty)
	require.NotNil(t, types.ErrProtocolURIEmpty)
	require.NotNil(t, types.ErrVaultIDEmpty)
	require.NotNil(t, types.ErrPermissionIDEmpty)
	require.NotNil(t, types.ErrPublicKeyEmpty)
	require.NotNil(t, types.ErrMessageEmpty)
	require.NotNil(t, types.ErrSignatureEmpty)
	require.NotNil(t, types.ErrDIDEmpty)
	require.NotNil(t, types.ErrSaltEmpty)
	require.NotNil(t, types.ErrAddressEmpty)
	require.NotNil(t, types.ErrInvalidAddressFormat)
	require.NotNil(t, types.ErrInvalidAuthorityFormat)

	// Record management errors
	require.NotNil(t, types.ErrRecordNotFound)
	require.NotNil(t, types.ErrRecordSizeExceeded)
	require.NotNil(t, types.ErrRecordAlreadyExists)
	require.NotNil(t, types.ErrRecordDataInvalid)
	require.NotNil(t, types.ErrRecordSchemaInvalid)
	require.NotNil(t, types.ErrRecordEncrypted)
	require.NotNil(t, types.ErrRecordDecryption)
	require.NotNil(t, types.ErrRecordEncryption)
	require.NotNil(t, types.ErrRecordSignature)
	require.NotNil(t, types.ErrRecordPermission)

	// Protocol management errors
	require.NotNil(t, types.ErrProtocolNotFound)
	require.NotNil(t, types.ErrProtocolAlreadyExists)
	require.NotNil(t, types.ErrProtocolLimitReached)
	require.NotNil(t, types.ErrProtocolInvalid)
	require.NotNil(t, types.ErrProtocolPermission)
	require.NotNil(t, types.ErrProtocolVersionInvalid)
	require.NotNil(t, types.ErrProtocolURIInvalid)
	require.NotNil(t, types.ErrProtocolConfigInvalid)
	require.NotNil(t, types.ErrProtocolRuleInvalid)
	require.NotNil(t, types.ErrProtocolActionInvalid)

	// Permission management errors
	require.NotNil(t, types.ErrPermissionNotFound)
	require.NotNil(t, types.ErrPermissionAlreadyExists)
	require.NotNil(t, types.ErrPermissionLimitReached)
	require.NotNil(t, types.ErrPermissionAlreadyRevoked)
	require.NotNil(t, types.ErrPermissionInvalid)
	require.NotNil(t, types.ErrPermissionExpired)
	require.NotNil(t, types.ErrPermissionScope)
	require.NotNil(t, types.ErrPermissionGrantInvalid)
	require.NotNil(t, types.ErrPermissionDenied)
	require.NotNil(t, types.ErrPermissionInherited)
	require.NotNil(t, types.ErrInvalidUCANToken)

	// Vault management errors
	require.NotNil(t, types.ErrVaultNotFound)
	require.NotNil(t, types.ErrVaultAlreadyExists)
	require.NotNil(t, types.ErrVaultNotInitialized)
	require.NotNil(t, types.ErrVaultInitializationFailed)
	require.NotNil(t, types.ErrVaultOperationFailed)
	require.NotNil(t, types.ErrVaultPermission)
	require.NotNil(t, types.ErrVaultKeyNotFound)
	require.NotNil(t, types.ErrVaultKeyInvalid)
	require.NotNil(t, types.ErrVaultSecretInvalid)
	require.NotNil(t, types.ErrVaultLocked)

	// Wallet derivation errors
	require.NotNil(t, types.ErrWalletDerivationFailed)
	require.NotNil(t, types.ErrWalletCreateFailed)
	require.NotNil(t, types.ErrWalletSignFailed)
	require.NotNil(t, types.ErrWalletVerifyFailed)
	require.NotNil(t, types.ErrWalletAddressMismatch)
	require.NotNil(t, types.ErrWalletKeyInvalid)
	require.NotNil(t, types.ErrWalletSeedInvalid)
	require.NotNil(t, types.ErrWalletPathInvalid)
	require.NotNil(t, types.ErrWalletTypeUnsupported)
	require.NotNil(t, types.ErrWalletNotFound)

	// Cryptographic errors
	require.NotNil(t, types.ErrCryptographicOperation)
	require.NotNil(t, types.ErrSignatureVerification)
	require.NotNil(t, types.ErrKeyGeneration)
	require.NotNil(t, types.ErrHashGeneration)
	require.NotNil(t, types.ErrEncryptionFailed)
	require.NotNil(t, types.ErrDecryptionFailed)
	require.NotNil(t, types.ErrKeyExchange)
	require.NotNil(t, types.ErrKeyDerivation)
	require.NotNil(t, types.ErrCertificateInvalid)
	require.NotNil(t, types.ErrCertificateExpired)

	// Storage and state errors
	require.NotNil(t, types.ErrStorageOperation)
	require.NotNil(t, types.ErrStateCorrupted)
	require.NotNil(t, types.ErrStateMismatch)
	require.NotNil(t, types.ErrStateNotFound)
	require.NotNil(t, types.ErrStateInvalid)
	require.NotNil(t, types.ErrStateConflict)
	require.NotNil(t, types.ErrStateLocked)
	require.NotNil(t, types.ErrStateExpired)
	require.NotNil(t, types.ErrStatePermission)
	require.NotNil(t, types.ErrStateVersion)

	// Network and communication errors
	require.NotNil(t, types.ErrNetworkOperation)
	require.NotNil(t, types.ErrConnectionFailed)
	require.NotNil(t, types.ErrTimeoutExceeded)
	require.NotNil(t, types.ErrRateLimitExceeded)
	require.NotNil(t, types.ErrQuotaExceeded)
	require.NotNil(t, types.ErrResourceExhausted)
	require.NotNil(t, types.ErrServiceUnavailable)
	require.NotNil(t, types.ErrServiceTimeout)
	require.NotNil(t, types.ErrServiceError)
	require.NotNil(t, types.ErrServiceMaintenance)

	// Generic operation errors
	require.NotNil(t, types.ErrNotImplemented)
	require.NotNil(t, types.ErrOperationFailed)
	require.NotNil(t, types.ErrInternalError)
	require.NotNil(t, types.ErrUnknownError)

	// Service verification errors
	require.NotNil(t, types.ErrServiceNotVerified)
}

func TestErrorMessages(t *testing.T) {
	// Test that error messages are correct
	require.Contains(t, types.ErrInvalidRequest.Error(), "invalid request")
	require.Contains(t, types.ErrRequestCannotBeNil.Error(), "request cannot be nil")
	require.Contains(t, types.ErrTargetDIDEmpty.Error(), "target DID cannot be empty")
	require.Contains(t, types.ErrRecordIDEmpty.Error(), "record ID cannot be empty")
	require.Contains(t, types.ErrProtocolURIEmpty.Error(), "protocol URI cannot be empty")
	require.Contains(t, types.ErrVaultIDEmpty.Error(), "vault ID cannot be empty")
	require.Contains(t, types.ErrPermissionIDEmpty.Error(), "permission ID cannot be empty")
	require.Contains(t, types.ErrPublicKeyEmpty.Error(), "public key cannot be empty")
	require.Contains(t, types.ErrMessageEmpty.Error(), "message cannot be empty")
	require.Contains(t, types.ErrSignatureEmpty.Error(), "signature cannot be empty")
	require.Contains(t, types.ErrDIDEmpty.Error(), "DID cannot be empty")
	require.Contains(t, types.ErrSaltEmpty.Error(), "salt cannot be empty")
	require.Contains(t, types.ErrAddressEmpty.Error(), "address cannot be empty")
	require.Contains(t, types.ErrInvalidAddressFormat.Error(), "invalid address format")
	require.Contains(t, types.ErrInvalidAuthorityFormat.Error(), "invalid authority format")

	// Record management error messages
	require.Contains(t, types.ErrRecordNotFound.Error(), "record not found")
	require.Contains(t, types.ErrRecordSizeExceeded.Error(), "record size exceeds maximum allowed")
	require.Contains(t, types.ErrRecordAlreadyExists.Error(), "record already exists")
	require.Contains(t, types.ErrRecordDataInvalid.Error(), "record data is invalid")
	require.Contains(t, types.ErrRecordSchemaInvalid.Error(), "record schema is invalid")
	require.Contains(t, types.ErrRecordEncrypted.Error(), "record is encrypted")
	require.Contains(t, types.ErrRecordDecryption.Error(), "failed to decrypt record")
	require.Contains(t, types.ErrRecordEncryption.Error(), "failed to encrypt record")
	require.Contains(t, types.ErrRecordSignature.Error(), "invalid record signature")
	require.Contains(t, types.ErrRecordPermission.Error(), "insufficient permissions for record")

	// Protocol management error messages
	require.Contains(t, types.ErrProtocolNotFound.Error(), "protocol not found")
	require.Contains(t, types.ErrProtocolAlreadyExists.Error(), "protocol already exists")
	require.Contains(t, types.ErrProtocolLimitReached.Error(), "protocol limit reached for DWN")
	require.Contains(t, types.ErrProtocolInvalid.Error(), "protocol definition is invalid")
	require.Contains(
		t,
		types.ErrProtocolPermission.Error(),
		"insufficient permissions for protocol",
	)

	// Permission management error messages
	require.Contains(t, types.ErrPermissionNotFound.Error(), "permission not found")
	require.Contains(t, types.ErrPermissionAlreadyExists.Error(), "permission already exists")
	require.Contains(t, types.ErrPermissionLimitReached.Error(), "permission limit reached for DWN")
	require.Contains(t, types.ErrPermissionAlreadyRevoked.Error(), "permission already revoked")
	require.Contains(t, types.ErrPermissionInvalid.Error(), "permission is invalid")
	require.Contains(t, types.ErrPermissionExpired.Error(), "permission has expired")
	require.Contains(t, types.ErrPermissionScope.Error(), "permission scope is invalid")
	require.Contains(t, types.ErrPermissionGrantInvalid.Error(), "permission grant is invalid")
	require.Contains(t, types.ErrPermissionDenied.Error(), "permission denied")
	require.Contains(t, types.ErrPermissionInherited.Error(), "cannot modify inherited permission")
	require.Contains(t, types.ErrInvalidUCANToken.Error(), "UCAN token is invalid or insufficient")

	// Vault management error messages
	require.Contains(t, types.ErrVaultNotFound.Error(), "vault not found")
	require.Contains(t, types.ErrVaultAlreadyExists.Error(), "vault already exists")
	require.Contains(t, types.ErrVaultNotInitialized.Error(), "vault not initialized")
	require.Contains(t, types.ErrVaultInitializationFailed.Error(), "vault initialization failed")
	require.Contains(t, types.ErrVaultOperationFailed.Error(), "vault operation failed")
	require.Contains(t, types.ErrVaultPermission.Error(), "insufficient vault permissions")
	require.Contains(t, types.ErrVaultKeyNotFound.Error(), "vault key not found")
	require.Contains(t, types.ErrVaultKeyInvalid.Error(), "vault key is invalid")
	require.Contains(t, types.ErrVaultSecretInvalid.Error(), "vault secret is invalid")
	require.Contains(t, types.ErrVaultLocked.Error(), "vault is locked")

	// Wallet derivation error messages
	require.Contains(
		t,
		types.ErrWalletDerivationFailed.Error(),
		"failed to derive wallet addresses",
	)
	require.Contains(t, types.ErrWalletCreateFailed.Error(), "failed to create wallet")
	require.Contains(t, types.ErrWalletSignFailed.Error(), "failed to sign message")
	require.Contains(t, types.ErrWalletVerifyFailed.Error(), "failed to verify signature")
	require.Contains(t, types.ErrWalletAddressMismatch.Error(), "wallet address mismatch")
	require.Contains(t, types.ErrWalletKeyInvalid.Error(), "wallet key is invalid")
	require.Contains(t, types.ErrWalletSeedInvalid.Error(), "wallet seed is invalid")
	require.Contains(t, types.ErrWalletPathInvalid.Error(), "wallet derivation path is invalid")
	require.Contains(t, types.ErrWalletTypeUnsupported.Error(), "wallet type is unsupported")
	require.Contains(t, types.ErrWalletNotFound.Error(), "wallet not found")

	// Cryptographic error messages
	require.Contains(t, types.ErrCryptographicOperation.Error(), "cryptographic operation failed")
	require.Contains(t, types.ErrSignatureVerification.Error(), "signature verification failed")
	require.Contains(t, types.ErrKeyGeneration.Error(), "key generation failed")
	require.Contains(t, types.ErrHashGeneration.Error(), "hash generation failed")
	require.Contains(t, types.ErrEncryptionFailed.Error(), "encryption failed")
	require.Contains(t, types.ErrDecryptionFailed.Error(), "decryption failed")
	require.Contains(t, types.ErrKeyExchange.Error(), "key exchange failed")
	require.Contains(t, types.ErrKeyDerivation.Error(), "key derivation failed")
	require.Contains(t, types.ErrCertificateInvalid.Error(), "certificate is invalid")
	require.Contains(t, types.ErrCertificateExpired.Error(), "certificate has expired")

	// Storage and state error messages
	require.Contains(t, types.ErrStorageOperation.Error(), "storage operation failed")
	require.Contains(t, types.ErrStateCorrupted.Error(), "state is corrupted")
	require.Contains(t, types.ErrStateMismatch.Error(), "state mismatch")
	require.Contains(t, types.ErrStateNotFound.Error(), "state not found")
	require.Contains(t, types.ErrStateInvalid.Error(), "state is invalid")
	require.Contains(t, types.ErrStateConflict.Error(), "state conflict")
	require.Contains(t, types.ErrStateLocked.Error(), "state is locked")
	require.Contains(t, types.ErrStateExpired.Error(), "state has expired")
	require.Contains(t, types.ErrStatePermission.Error(), "insufficient state permissions")
	require.Contains(t, types.ErrStateVersion.Error(), "state version mismatch")

	// Network and communication error messages
	require.Contains(t, types.ErrNetworkOperation.Error(), "network operation failed")
	require.Contains(t, types.ErrConnectionFailed.Error(), "connection failed")
	require.Contains(t, types.ErrTimeoutExceeded.Error(), "timeout exceeded")
	require.Contains(t, types.ErrRateLimitExceeded.Error(), "rate limit exceeded")
	require.Contains(t, types.ErrQuotaExceeded.Error(), "quota exceeded")
	require.Contains(t, types.ErrResourceExhausted.Error(), "resource exhausted")
	require.Contains(t, types.ErrServiceUnavailable.Error(), "service unavailable")
	require.Contains(t, types.ErrServiceTimeout.Error(), "service timeout")
	require.Contains(t, types.ErrServiceError.Error(), "service error")
	require.Contains(t, types.ErrServiceMaintenance.Error(), "service under maintenance")

	// Generic operation error messages
	require.Contains(t, types.ErrNotImplemented.Error(), "operation not implemented")
	require.Contains(t, types.ErrOperationFailed.Error(), "operation failed")
	require.Contains(t, types.ErrInternalError.Error(), "internal error")
	require.Contains(t, types.ErrUnknownError.Error(), "unknown error")

	// Service verification error messages
	require.Contains(t, types.ErrServiceNotVerified.Error(), "service not verified for domain")
}

func TestErrorCodes(t *testing.T) {
	// Test that error codes are in the expected ranges

	// Request validation errors (1-15)
	require.Equal(t, uint32(1), types.ErrInvalidRequest.ABCICode())
	require.Equal(t, uint32(2), types.ErrRequestCannotBeNil.ABCICode())
	require.Equal(t, uint32(3), types.ErrTargetDIDEmpty.ABCICode())
	require.Equal(t, uint32(4), types.ErrRecordIDEmpty.ABCICode())
	require.Equal(t, uint32(5), types.ErrProtocolURIEmpty.ABCICode())
	require.Equal(t, uint32(6), types.ErrVaultIDEmpty.ABCICode())
	require.Equal(t, uint32(7), types.ErrPermissionIDEmpty.ABCICode())
	require.Equal(t, uint32(8), types.ErrPublicKeyEmpty.ABCICode())
	require.Equal(t, uint32(9), types.ErrMessageEmpty.ABCICode())
	require.Equal(t, uint32(10), types.ErrSignatureEmpty.ABCICode())
	require.Equal(t, uint32(11), types.ErrDIDEmpty.ABCICode())
	require.Equal(t, uint32(12), types.ErrSaltEmpty.ABCICode())
	require.Equal(t, uint32(13), types.ErrAddressEmpty.ABCICode())
	require.Equal(t, uint32(14), types.ErrInvalidAddressFormat.ABCICode())
	require.Equal(t, uint32(15), types.ErrInvalidAuthorityFormat.ABCICode())

	// Record management errors (16-25)
	require.Equal(t, uint32(16), types.ErrRecordNotFound.ABCICode())
	require.Equal(t, uint32(17), types.ErrRecordSizeExceeded.ABCICode())
	require.Equal(t, uint32(18), types.ErrRecordAlreadyExists.ABCICode())
	require.Equal(t, uint32(19), types.ErrRecordDataInvalid.ABCICode())
	require.Equal(t, uint32(20), types.ErrRecordSchemaInvalid.ABCICode())
	require.Equal(t, uint32(21), types.ErrRecordEncrypted.ABCICode())
	require.Equal(t, uint32(22), types.ErrRecordDecryption.ABCICode())
	require.Equal(t, uint32(23), types.ErrRecordEncryption.ABCICode())
	require.Equal(t, uint32(24), types.ErrRecordSignature.ABCICode())
	require.Equal(t, uint32(25), types.ErrRecordPermission.ABCICode())

	// Protocol management errors (26-35)
	require.Equal(t, uint32(26), types.ErrProtocolNotFound.ABCICode())
	require.Equal(t, uint32(27), types.ErrProtocolAlreadyExists.ABCICode())
	require.Equal(t, uint32(28), types.ErrProtocolLimitReached.ABCICode())
	require.Equal(t, uint32(29), types.ErrProtocolInvalid.ABCICode())
	require.Equal(t, uint32(30), types.ErrProtocolPermission.ABCICode())
	require.Equal(t, uint32(31), types.ErrProtocolVersionInvalid.ABCICode())
	require.Equal(t, uint32(32), types.ErrProtocolURIInvalid.ABCICode())
	require.Equal(t, uint32(33), types.ErrProtocolConfigInvalid.ABCICode())
	require.Equal(t, uint32(34), types.ErrProtocolRuleInvalid.ABCICode())
	require.Equal(t, uint32(35), types.ErrProtocolActionInvalid.ABCICode())

	// Permission management errors (36-45)
	require.Equal(t, uint32(36), types.ErrPermissionNotFound.ABCICode())
	require.Equal(t, uint32(37), types.ErrPermissionAlreadyExists.ABCICode())
	require.Equal(t, uint32(38), types.ErrPermissionLimitReached.ABCICode())
	require.Equal(t, uint32(39), types.ErrPermissionAlreadyRevoked.ABCICode())
	require.Equal(t, uint32(40), types.ErrPermissionInvalid.ABCICode())
	require.Equal(t, uint32(41), types.ErrPermissionExpired.ABCICode())
	require.Equal(t, uint32(42), types.ErrPermissionScope.ABCICode())
	require.Equal(t, uint32(43), types.ErrPermissionGrantInvalid.ABCICode())
	require.Equal(t, uint32(44), types.ErrPermissionDenied.ABCICode())
	require.Equal(t, uint32(45), types.ErrPermissionInherited.ABCICode())
	require.Equal(t, uint32(46), types.ErrInvalidUCANToken.ABCICode())

	// Vault management errors (47-56)
	require.Equal(t, uint32(47), types.ErrVaultNotFound.ABCICode())
	require.Equal(t, uint32(48), types.ErrVaultAlreadyExists.ABCICode())
	require.Equal(t, uint32(49), types.ErrVaultNotInitialized.ABCICode())
	require.Equal(t, uint32(50), types.ErrVaultInitializationFailed.ABCICode())
	require.Equal(t, uint32(51), types.ErrVaultOperationFailed.ABCICode())
	require.Equal(t, uint32(52), types.ErrVaultPermission.ABCICode())
	require.Equal(t, uint32(53), types.ErrVaultKeyNotFound.ABCICode())
	require.Equal(t, uint32(54), types.ErrVaultKeyInvalid.ABCICode())
	require.Equal(t, uint32(55), types.ErrVaultSecretInvalid.ABCICode())
	require.Equal(t, uint32(56), types.ErrVaultLocked.ABCICode())

	// Wallet derivation errors (57-66)
	require.Equal(t, uint32(57), types.ErrWalletDerivationFailed.ABCICode())
	require.Equal(t, uint32(58), types.ErrWalletCreateFailed.ABCICode())
	require.Equal(t, uint32(59), types.ErrWalletSignFailed.ABCICode())
	require.Equal(t, uint32(60), types.ErrWalletVerifyFailed.ABCICode())
	require.Equal(t, uint32(61), types.ErrWalletAddressMismatch.ABCICode())
	require.Equal(t, uint32(62), types.ErrWalletKeyInvalid.ABCICode())
	require.Equal(t, uint32(63), types.ErrWalletSeedInvalid.ABCICode())
	require.Equal(t, uint32(64), types.ErrWalletPathInvalid.ABCICode())
	require.Equal(t, uint32(65), types.ErrWalletTypeUnsupported.ABCICode())
	require.Equal(t, uint32(66), types.ErrWalletNotFound.ABCICode())
	require.Equal(t, uint32(67), types.ErrUnsupportedTransactionType.ABCICode())
	require.Equal(t, uint32(68), types.ErrWalletOperationFailed.ABCICode())

	// Cryptographic errors (69-78)
	require.Equal(t, uint32(69), types.ErrCryptographicOperation.ABCICode())
	require.Equal(t, uint32(70), types.ErrSignatureVerification.ABCICode())
	require.Equal(t, uint32(71), types.ErrKeyGeneration.ABCICode())
	require.Equal(t, uint32(72), types.ErrHashGeneration.ABCICode())
	require.Equal(t, uint32(73), types.ErrEncryptionFailed.ABCICode())
	require.Equal(t, uint32(74), types.ErrDecryptionFailed.ABCICode())
	require.Equal(t, uint32(75), types.ErrKeyExchange.ABCICode())
	require.Equal(t, uint32(76), types.ErrKeyDerivation.ABCICode())
	require.Equal(t, uint32(77), types.ErrCertificateInvalid.ABCICode())
	require.Equal(t, uint32(78), types.ErrCertificateExpired.ABCICode())

	// Storage and state errors (79-88)
	require.Equal(t, uint32(79), types.ErrStorageOperation.ABCICode())
	require.Equal(t, uint32(80), types.ErrStateCorrupted.ABCICode())
	require.Equal(t, uint32(81), types.ErrStateMismatch.ABCICode())
	require.Equal(t, uint32(82), types.ErrStateNotFound.ABCICode())
	require.Equal(t, uint32(83), types.ErrStateInvalid.ABCICode())
	require.Equal(t, uint32(84), types.ErrStateConflict.ABCICode())
	require.Equal(t, uint32(85), types.ErrStateLocked.ABCICode())
	require.Equal(t, uint32(86), types.ErrStateExpired.ABCICode())
	require.Equal(t, uint32(87), types.ErrStatePermission.ABCICode())
	require.Equal(t, uint32(88), types.ErrStateVersion.ABCICode())

	// Network and communication errors (89-98)
	require.Equal(t, uint32(89), types.ErrNetworkOperation.ABCICode())
	require.Equal(t, uint32(90), types.ErrConnectionFailed.ABCICode())
	require.Equal(t, uint32(91), types.ErrTimeoutExceeded.ABCICode())
	require.Equal(t, uint32(92), types.ErrRateLimitExceeded.ABCICode())
	require.Equal(t, uint32(93), types.ErrQuotaExceeded.ABCICode())
	require.Equal(t, uint32(94), types.ErrResourceExhausted.ABCICode())
	require.Equal(t, uint32(95), types.ErrServiceUnavailable.ABCICode())
	require.Equal(t, uint32(96), types.ErrServiceTimeout.ABCICode())
	require.Equal(t, uint32(97), types.ErrServiceError.ABCICode())
	require.Equal(t, uint32(98), types.ErrServiceMaintenance.ABCICode())

	// Generic operation errors (99-102)
	require.Equal(t, uint32(99), types.ErrNotImplemented.ABCICode())
	require.Equal(t, uint32(100), types.ErrOperationFailed.ABCICode())
	require.Equal(t, uint32(101), types.ErrInternalError.ABCICode())
	require.Equal(t, uint32(102), types.ErrUnknownError.ABCICode())

	// Service verification errors (103)
	require.Equal(t, uint32(103), types.ErrServiceNotVerified.ABCICode())
}
