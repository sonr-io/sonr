package types

import (
	"cosmossdk.io/errors"
)

// DID module sentinel errors
var (
	// DID Document errors
	ErrDIDAlreadyExists   = errors.Register(ModuleName, 1, "DID already exists")
	ErrDIDNotFound        = errors.Register(ModuleName, 2, "DID not found")
	ErrDIDDeactivated     = errors.Register(ModuleName, 3, "DID is deactivated")
	ErrInvalidDIDDocument = errors.Register(ModuleName, 4, "invalid DID document")
	ErrUnauthorized       = errors.Register(ModuleName, 5, "unauthorized")

	// Verification Method errors
	ErrInvalidVerificationMethod  = errors.Register(ModuleName, 6, "invalid verification method")
	ErrVerificationMethodNotFound = errors.Register(ModuleName, 7, "verification method not found")

	// Service errors
	ErrInvalidService  = errors.Register(ModuleName, 8, "invalid service")
	ErrServiceNotFound = errors.Register(ModuleName, 9, "service not found")

	// Credential errors
	ErrCredentialNotFound = errors.Register(ModuleName, 10, "credential not found")
	ErrCredentialRevoked  = errors.Register(ModuleName, 11, "credential is revoked")
	ErrInvalidCredential  = errors.Register(ModuleName, 12, "invalid credential")

	// Address errors
	ErrInvalidControllerAddress = errors.Register(ModuleName, 13, "invalid controller address")
	ErrInvalidIssuerAddress     = errors.Register(ModuleName, 14, "invalid issuer address")
	ErrInvalidAuthorityAddress  = errors.Register(ModuleName, 15, "invalid authority address")

	// Validation errors
	ErrEmptyDID           = errors.Register(ModuleName, 16, "DID cannot be empty")
	ErrEmptyDIDDocumentID = errors.Register(
		ModuleName,
		17,
		"DID document ID cannot be empty",
	)
	ErrDIDMismatch = errors.Register(
		ModuleName,
		18,
		"DID and DID document ID must match",
	)
	ErrEmptyVerificationMethodID = errors.Register(
		ModuleName,
		19,
		"verification method ID cannot be empty",
	)
	ErrEmptyVerificationMethodKind = errors.Register(
		ModuleName,
		20,
		"verification method kind cannot be empty",
	)
	ErrEmptyServiceID    = errors.Register(ModuleName, 21, "service ID cannot be empty")
	ErrEmptyServiceKind  = errors.Register(ModuleName, 22, "service kind cannot be empty")
	ErrEmptyCredentialID = errors.Register(
		ModuleName,
		23,
		"credential ID cannot be empty",
	)
	ErrEmptyCredentialIssuer = errors.Register(
		ModuleName,
		24,
		"credential issuer cannot be empty",
	)

	// DID Document validation errors
	ErrInvalidDIDSyntax     = errors.Register(ModuleName, 25, "invalid DID syntax")
	ErrMissingDIDDocumentID = errors.Register(
		ModuleName,
		26,
		"DID document must have an ID",
	)
	ErrMissingVerificationMethodID = errors.Register(
		ModuleName,
		27,
		"verification method must have an ID",
	)
	ErrMissingVerificationMethodKind = errors.Register(
		ModuleName,
		28,
		"verification method must have a kind",
	)
	ErrMissingVerificationMethodController = errors.Register(
		ModuleName,
		29,
		"verification method must have a controller",
	)
	ErrMissingVerificationMethodKey = errors.Register(
		ModuleName,
		30,
		"verification method must have public key material",
	)
	ErrMissingServiceID = errors.Register(
		ModuleName,
		31,
		"service must have an ID",
	)
	ErrMissingServiceKind = errors.Register(
		ModuleName,
		32,
		"service must have a kind",
	)
	ErrMissingServiceEndpoint = errors.Register(
		ModuleName,
		33,
		"service must have an endpoint",
	)

	// Storage errors
	ErrFailedToCheckDIDExists = errors.Register(
		ModuleName,
		34,
		"failed to check if DID exists",
	)
	ErrFailedToStoreDIDDocument = errors.Register(
		ModuleName,
		35,
		"failed to store DID document",
	)
	ErrFailedToStoreDIDMetadata = errors.Register(
		ModuleName,
		36,
		"failed to store DID document metadata",
	)
	ErrFailedToUpdateDIDDocument = errors.Register(
		ModuleName,
		37,
		"failed to update DID document",
	)
	ErrFailedToGetDIDMetadata    = errors.Register(ModuleName, 38, "failed to get DID metadata")
	ErrFailedToUpdateDIDMetadata = errors.Register(
		ModuleName,
		39,
		"failed to update DID metadata",
	)
	ErrFailedToDeactivateDIDDocument = errors.Register(
		ModuleName,
		40,
		"failed to deactivate DID document",
	)
	ErrFailedToCheckCredentialExists = errors.Register(
		ModuleName,
		41,
		"failed to check if credential exists",
	)
	ErrFailedToStoreCredential = errors.Register(
		ModuleName,
		42,
		"failed to store verifiable credential",
	)
	ErrFailedToUpdateCredential = errors.Register(
		ModuleName,
		43,
		"failed to update credential",
	)

	// Existence errors
	ErrVerificationMethodAlreadyExists = errors.Register(
		ModuleName,
		44,
		"verification method with ID already exists",
	)
	ErrServiceAlreadyExists = errors.Register(
		ModuleName,
		45,
		"service with ID already exists",
	)
	ErrCredentialAlreadyExists = errors.Register(
		ModuleName,
		46,
		"credential ID already exists",
	)
	ErrDIDAlreadyDeactivated    = errors.Register(ModuleName, 47, "DID already deactivated")
	ErrCredentialAlreadyRevoked = errors.Register(
		ModuleName,
		48,
		"credential already revoked",
	)

	// Query errors
	ErrInvalidRequest = errors.Register(ModuleName, 49, "invalid request")

	// Parameter errors
	ErrInvalidParams = errors.Register(ModuleName, 62, "invalid module parameters")

	// External Wallet Linking errors
	ErrInvalidBlockchainAccountID = errors.Register(
		ModuleName,
		50,
		"invalid blockchain account ID",
	)
	ErrUnsupportedBlockchainNamespace = errors.Register(
		ModuleName,
		51,
		"unsupported blockchain namespace",
	)
	ErrUnsupportedWalletType = errors.Register(
		ModuleName,
		52,
		"unsupported wallet type",
	)
	ErrInvalidEthereumAddress = errors.Register(
		ModuleName,
		53,
		"invalid Ethereum address",
	)
	ErrInvalidCosmosAddress      = errors.Register(ModuleName, 54, "invalid Cosmos address")
	ErrInvalidWalletVerification = errors.Register(
		ModuleName,
		55,
		"invalid wallet verification",
	)
	ErrWalletSignatureVerificationFailed = errors.Register(
		ModuleName,
		56,
		"wallet signature verification failed",
	)
	ErrWalletAlreadyLinked = errors.Register(
		ModuleName,
		57,
		"wallet already linked to DID",
	)
	ErrDWNVaultControllerRequired = errors.Register(
		ModuleName,
		58,
		"DID must have active DWN vault controller",
	)

	// WebAuthn errors
	ErrInvalidWebAuthnCredential = errors.Register(
		ModuleName,
		59,
		"invalid WebAuthn credential",
	)
	ErrWebAuthnCredentialAlreadyExists = errors.Register(
		ModuleName,
		60,
		"WebAuthn credential already exists",
	)
	ErrMaxWebAuthnCredentialsExceeded = errors.Register(
		ModuleName,
		61,
		"maximum WebAuthn credentials per DID exceeded",
	)
	ErrAssertionNotFound = errors.Register(
		ModuleName,
		64,
		"assertion DID not found",
	)
	ErrInvalidAssertion = errors.Register(
		ModuleName,
		65,
		"invalid assertion",
	)
	ErrNoCredentials = errors.Register(
		ModuleName,
		66,
		"no WebAuthn credentials found",
	)

	// UCAN authorization errors
	ErrUCANValidationFailed = errors.Register(
		ModuleName,
		63,
		"UCAN authorization validation failed",
	)
)
