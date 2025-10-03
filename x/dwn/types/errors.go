package types

import (
	"cosmossdk.io/errors"
)

// DWN module sentinel errors
var (
	// Request validation errors (1-15)
	ErrInvalidRequest         = errors.Register(ModuleName, 1, "invalid request")
	ErrRequestCannotBeNil     = errors.Register(ModuleName, 2, "request cannot be nil")
	ErrTargetDIDEmpty         = errors.Register(ModuleName, 3, "target DID cannot be empty")
	ErrRecordIDEmpty          = errors.Register(ModuleName, 4, "record ID cannot be empty")
	ErrProtocolURIEmpty       = errors.Register(ModuleName, 5, "protocol URI cannot be empty")
	ErrVaultIDEmpty           = errors.Register(ModuleName, 6, "vault ID cannot be empty")
	ErrPermissionIDEmpty      = errors.Register(ModuleName, 7, "permission ID cannot be empty")
	ErrPublicKeyEmpty         = errors.Register(ModuleName, 8, "public key cannot be empty")
	ErrMessageEmpty           = errors.Register(ModuleName, 9, "message cannot be empty")
	ErrSignatureEmpty         = errors.Register(ModuleName, 10, "signature cannot be empty")
	ErrDIDEmpty               = errors.Register(ModuleName, 11, "DID cannot be empty")
	ErrSaltEmpty              = errors.Register(ModuleName, 12, "salt cannot be empty")
	ErrAddressEmpty           = errors.Register(ModuleName, 13, "address cannot be empty")
	ErrInvalidAddressFormat   = errors.Register(ModuleName, 14, "invalid address format")
	ErrInvalidAuthorityFormat = errors.Register(ModuleName, 15, "invalid authority format")

	// Record management errors (16-25)
	ErrRecordNotFound      = errors.Register(ModuleName, 16, "record not found")
	ErrRecordSizeExceeded  = errors.Register(ModuleName, 17, "record size exceeds maximum allowed")
	ErrRecordAlreadyExists = errors.Register(ModuleName, 18, "record already exists")
	ErrRecordDataInvalid   = errors.Register(ModuleName, 19, "record data is invalid")
	ErrRecordSchemaInvalid = errors.Register(ModuleName, 20, "record schema is invalid")
	ErrRecordEncrypted     = errors.Register(ModuleName, 21, "record is encrypted")
	ErrRecordDecryption    = errors.Register(ModuleName, 22, "failed to decrypt record")
	ErrRecordEncryption    = errors.Register(ModuleName, 23, "failed to encrypt record")
	ErrRecordSignature     = errors.Register(ModuleName, 24, "invalid record signature")
	ErrRecordPermission    = errors.Register(ModuleName, 25, "insufficient permissions for record")

	// Protocol management errors (26-35)
	ErrProtocolNotFound      = errors.Register(ModuleName, 26, "protocol not found")
	ErrProtocolAlreadyExists = errors.Register(ModuleName, 27, "protocol already exists")
	ErrProtocolLimitReached  = errors.Register(ModuleName, 28, "protocol limit reached for DWN")
	ErrProtocolInvalid       = errors.Register(ModuleName, 29, "protocol definition is invalid")
	ErrProtocolPermission    = errors.Register(
		ModuleName,
		30,
		"insufficient permissions for protocol",
	)
	ErrProtocolVersionInvalid = errors.Register(ModuleName, 31, "protocol version is invalid")
	ErrProtocolURIInvalid     = errors.Register(ModuleName, 32, "protocol URI is invalid")
	ErrProtocolConfigInvalid  = errors.Register(ModuleName, 33, "protocol configuration is invalid")
	ErrProtocolRuleInvalid    = errors.Register(ModuleName, 34, "protocol rule is invalid")
	ErrProtocolActionInvalid  = errors.Register(ModuleName, 35, "protocol action is invalid")

	// Permission management errors (36-46)
	ErrPermissionNotFound      = errors.Register(ModuleName, 36, "permission not found")
	ErrPermissionAlreadyExists = errors.Register(ModuleName, 37, "permission already exists")
	ErrPermissionLimitReached  = errors.Register(
		ModuleName,
		38,
		"permission limit reached for DWN",
	)
	ErrPermissionAlreadyRevoked = errors.Register(ModuleName, 39, "permission already revoked")
	ErrPermissionInvalid        = errors.Register(ModuleName, 40, "permission is invalid")
	ErrPermissionExpired        = errors.Register(ModuleName, 41, "permission has expired")
	ErrPermissionScope          = errors.Register(ModuleName, 42, "permission scope is invalid")
	ErrPermissionGrantInvalid   = errors.Register(ModuleName, 43, "permission grant is invalid")
	ErrPermissionDenied         = errors.Register(ModuleName, 44, "permission denied")
	ErrPermissionInherited      = errors.Register(
		ModuleName,
		45,
		"cannot modify inherited permission",
	)
	ErrInvalidUCANToken = errors.Register(
		ModuleName,
		46,
		"UCAN token is invalid or insufficient",
	)

	// Vault management errors (47-56)
	ErrVaultNotFound             = errors.Register(ModuleName, 47, "vault not found")
	ErrVaultAlreadyExists        = errors.Register(ModuleName, 48, "vault already exists")
	ErrVaultNotInitialized       = errors.Register(ModuleName, 49, "vault not initialized")
	ErrVaultInitializationFailed = errors.Register(ModuleName, 50, "vault initialization failed")
	ErrVaultOperationFailed      = errors.Register(ModuleName, 51, "vault operation failed")
	ErrVaultPermission           = errors.Register(ModuleName, 52, "insufficient vault permissions")
	ErrVaultKeyNotFound          = errors.Register(ModuleName, 53, "vault key not found")
	ErrVaultKeyInvalid           = errors.Register(ModuleName, 54, "vault key is invalid")
	ErrVaultSecretInvalid        = errors.Register(ModuleName, 55, "vault secret is invalid")
	ErrVaultLocked               = errors.Register(ModuleName, 56, "vault is locked")

	// Wallet derivation errors (57-68)
	ErrWalletDerivationFailed = errors.Register(
		ModuleName,
		57,
		"failed to derive wallet addresses",
	)
	ErrWalletCreateFailed    = errors.Register(ModuleName, 58, "failed to create wallet")
	ErrWalletSignFailed      = errors.Register(ModuleName, 59, "failed to sign message")
	ErrWalletVerifyFailed    = errors.Register(ModuleName, 60, "failed to verify signature")
	ErrWalletAddressMismatch = errors.Register(ModuleName, 61, "wallet address mismatch")
	ErrWalletKeyInvalid      = errors.Register(ModuleName, 62, "wallet key is invalid")
	ErrWalletSeedInvalid     = errors.Register(ModuleName, 63, "wallet seed is invalid")
	ErrWalletPathInvalid     = errors.Register(
		ModuleName,
		64,
		"wallet derivation path is invalid",
	)
	ErrWalletTypeUnsupported      = errors.Register(ModuleName, 65, "wallet type is unsupported")
	ErrWalletNotFound             = errors.Register(ModuleName, 66, "wallet not found")
	ErrUnsupportedTransactionType = errors.Register(ModuleName, 67, "unsupported transaction type")
	ErrWalletOperationFailed      = errors.Register(ModuleName, 68, "wallet operation failed")

	// Cryptographic errors (69-78)
	ErrCryptographicOperation = errors.Register(ModuleName, 69, "cryptographic operation failed")
	ErrSignatureVerification  = errors.Register(ModuleName, 70, "signature verification failed")
	ErrKeyGeneration          = errors.Register(ModuleName, 71, "key generation failed")
	ErrHashGeneration         = errors.Register(ModuleName, 72, "hash generation failed")
	ErrEncryptionFailed       = errors.Register(ModuleName, 73, "encryption failed")
	ErrDecryptionFailed       = errors.Register(ModuleName, 74, "decryption failed")
	ErrKeyExchange            = errors.Register(ModuleName, 75, "key exchange failed")
	ErrKeyDerivation          = errors.Register(ModuleName, 76, "key derivation failed")
	ErrCertificateInvalid     = errors.Register(ModuleName, 77, "certificate is invalid")
	ErrCertificateExpired     = errors.Register(ModuleName, 78, "certificate has expired")

	// Storage and state errors (79-88)
	ErrStorageOperation = errors.Register(ModuleName, 79, "storage operation failed")
	ErrStateCorrupted   = errors.Register(ModuleName, 80, "state is corrupted")
	ErrStateMismatch    = errors.Register(ModuleName, 81, "state mismatch")
	ErrStateNotFound    = errors.Register(ModuleName, 82, "state not found")
	ErrStateInvalid     = errors.Register(ModuleName, 83, "state is invalid")
	ErrStateConflict    = errors.Register(ModuleName, 84, "state conflict")
	ErrStateLocked      = errors.Register(ModuleName, 85, "state is locked")
	ErrStateExpired     = errors.Register(ModuleName, 86, "state has expired")
	ErrStatePermission  = errors.Register(ModuleName, 87, "insufficient state permissions")
	ErrStateVersion     = errors.Register(ModuleName, 88, "state version mismatch")

	// Network and communication errors (89-98)
	ErrNetworkOperation   = errors.Register(ModuleName, 89, "network operation failed")
	ErrConnectionFailed   = errors.Register(ModuleName, 90, "connection failed")
	ErrTimeoutExceeded    = errors.Register(ModuleName, 91, "timeout exceeded")
	ErrRateLimitExceeded  = errors.Register(ModuleName, 92, "rate limit exceeded")
	ErrQuotaExceeded      = errors.Register(ModuleName, 93, "quota exceeded")
	ErrResourceExhausted  = errors.Register(ModuleName, 94, "resource exhausted")
	ErrServiceUnavailable = errors.Register(ModuleName, 95, "service unavailable")
	ErrServiceTimeout     = errors.Register(ModuleName, 96, "service timeout")
	ErrServiceError       = errors.Register(ModuleName, 97, "service error")
	ErrServiceMaintenance = errors.Register(ModuleName, 98, "service under maintenance")

	// Generic operation errors (99-102)
	ErrNotImplemented  = errors.Register(ModuleName, 99, "operation not implemented")
	ErrOperationFailed = errors.Register(ModuleName, 100, "operation failed")
	ErrInternalError   = errors.Register(ModuleName, 101, "internal error")
	ErrUnknownError    = errors.Register(ModuleName, 102, "unknown error")

	// Service verification errors (103-104)
	ErrServiceNotVerified = errors.Register(ModuleName, 103, "service not verified for domain")
	ErrUnauthorized       = errors.Register(ModuleName, 104, "unauthorized access")

	// Fee grant and wallet sponsorship errors (105-115)
	ErrInvalidWalletAddress       = errors.Register(ModuleName, 105, "invalid wallet address")
	ErrInvalidSpendLimit          = errors.Register(ModuleName, 106, "invalid spend limit")
	ErrFeeGrantNotFound           = errors.Register(ModuleName, 107, "fee grant not found")
	ErrFeeGrantAlreadyExists      = errors.Register(ModuleName, 108, "fee grant already exists")
	ErrWalletNotSponsorable       = errors.Register(ModuleName, 109, "wallet is not sponsorable")
	ErrSelfGrantNotAllowed        = errors.Register(ModuleName, 110, "self grants are not allowed")
	ErrDailyLimitExceeded         = errors.Register(ModuleName, 111, "daily limit exceeded")
	ErrMessageTypeNotAllowed      = errors.Register(ModuleName, 112, "message type not allowed")
	ErrSponsorshipExpired         = errors.Register(ModuleName, 113, "sponsorship has expired")
	ErrInsufficientSponsorBalance = errors.Register(ModuleName, 114, "insufficient sponsor balance")
	ErrGasEstimationFailed        = errors.Register(ModuleName, 115, "gas estimation failed")
	ErrFeeGrantExhausted          = errors.Register(ModuleName, 116, "fee grant has been exhausted")

	// IPFS errors (117-126)
	ErrIPFSClientNotAvailable = errors.Register(ModuleName, 117, "IPFS client not available")
)
