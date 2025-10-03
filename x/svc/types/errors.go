package types

import (
	"cosmossdk.io/errors"
)

// x/svc module error codes
const (
	DefaultCodespace = ModuleName

	ErrCodeDomainNotVerified        = 1001
	ErrCodeInvalidServiceID         = 1002
	ErrCodeServiceAlreadyExists     = 1003
	ErrCodeDomainAlreadyBound       = 1004
	ErrCodeFailedToSaveService      = 1005
	ErrCodeInvalidPermissions       = 1006
	ErrCodeUCANValidationFailed     = 1007
	ErrCodeInvalidUCANDelegation    = 1008
	ErrCodeFailedToCreateCapability = 1009
	ErrCodeInvalidOwnerDID          = 1010
	ErrCodeServiceNotFound          = 1011
	ErrCodeServiceNotActive         = 1012
	ErrCodeOIDCConfigNotFound       = 1013
	ErrCodeInvalidIssuer            = 1014
)

// x/svc module errors
var (
	ErrDomainNotVerified = errors.Register(
		DefaultCodespace,
		ErrCodeDomainNotVerified,
		"domain is not verified",
	)
	ErrInvalidServiceID = errors.Register(
		DefaultCodespace,
		ErrCodeInvalidServiceID,
		"invalid service ID",
	)
	ErrServiceAlreadyExists = errors.Register(
		DefaultCodespace,
		ErrCodeServiceAlreadyExists,
		"service already exists",
	)
	ErrDomainAlreadyBound = errors.Register(
		DefaultCodespace,
		ErrCodeDomainAlreadyBound,
		"domain is already bound to another service",
	)
	ErrFailedToSaveService = errors.Register(
		DefaultCodespace,
		ErrCodeFailedToSaveService,
		"failed to save service",
	)
	ErrInvalidPermissions = errors.Register(
		DefaultCodespace,
		ErrCodeInvalidPermissions,
		"invalid permissions",
	)
	ErrUCANValidationFailed = errors.Register(
		DefaultCodespace,
		ErrCodeUCANValidationFailed,
		"UCAN validation failed",
	)
	ErrInvalidUCANDelegation = errors.Register(
		DefaultCodespace,
		ErrCodeInvalidUCANDelegation,
		"invalid UCAN delegation chain",
	)
	ErrFailedToCreateCapability = errors.Register(
		DefaultCodespace,
		ErrCodeFailedToCreateCapability,
		"failed to create capability",
	)
	ErrInvalidOwnerDID = errors.Register(
		DefaultCodespace,
		ErrCodeInvalidOwnerDID,
		"invalid owner DID document",
	)
	ErrServiceNotFound = errors.Register(
		DefaultCodespace,
		ErrCodeServiceNotFound,
		"service not found",
	)
	ErrServiceNotActive = errors.Register(
		DefaultCodespace,
		ErrCodeServiceNotActive,
		"service is not active",
	)
	ErrOIDCConfigNotFound = errors.Register(
		DefaultCodespace,
		ErrCodeOIDCConfigNotFound,
		"OIDC configuration not found",
	)
	ErrInvalidIssuer = errors.Register(
		DefaultCodespace,
		ErrCodeInvalidIssuer,
		"invalid OIDC issuer",
	)
)
