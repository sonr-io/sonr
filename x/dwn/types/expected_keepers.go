package types

import (
	"context"

	didtypes "github.com/sonr-io/sonr/x/did/types"
	svctypes "github.com/sonr-io/sonr/x/svc/types"
)

// DIDKeeper interface defines the methods needed from the DID keeper
type DIDKeeper interface {
	// ResolveDID resolves a DID to its DID document
	ResolveDID(
		ctx context.Context,
		did string,
	) (*didtypes.DIDDocument, *didtypes.DIDDocumentMetadata, error)

	// GetDIDDocument gets a DID document by its ID
	GetDIDDocument(ctx context.Context, did string) (*didtypes.DIDDocument, error)
}

// ServiceKeeper interface defines the methods needed from the Service keeper
type ServiceKeeper interface {
	// VerifyServiceRegistration verifies service registration and domain ownership
	VerifyServiceRegistration(ctx context.Context, serviceID string, domain string) (bool, error)

	// GetService gets service by ID
	GetService(ctx context.Context, serviceID string) (*svctypes.Service, error)

	// IsDomainVerified checks if domain is verified
	IsDomainVerified(ctx context.Context, domain string, owner string) (bool, error)

	// GetServicesByDomain gets services by domain
	GetServicesByDomain(ctx context.Context, domain string) ([]svctypes.Service, error)
}
