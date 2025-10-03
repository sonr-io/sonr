package types

import (
	"context"

	didtypes "github.com/sonr-io/sonr/x/did/types"
)

// DIDKeeper interface defines the methods needed from the DID keeper
type DIDKeeper interface {
	// ResolveDID resolves a DID to its DID document and metadata
	ResolveDID(
		ctx context.Context,
		did string,
	) (*didtypes.DIDDocument, *didtypes.DIDDocumentMetadata, error)

	// GetDIDDocument gets a DID document by its ID
	GetDIDDocument(ctx context.Context, did string) (*didtypes.DIDDocument, error)

	// VerifyDIDDocumentSignature verifies a DID document signature
	VerifyDIDDocumentSignature(ctx context.Context, did string, signature []byte) (bool, error)
}
