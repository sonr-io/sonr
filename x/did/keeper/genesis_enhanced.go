package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	apiv1 "github.com/sonr-io/sonr/api/did/v1"
	"github.com/sonr-io/sonr/x/did/types"
)

// GenesisOrmData holds all ORM table data for genesis import/export
type GenesisOrmData struct {
	DidDocuments    []*apiv1.DIDDocument
	Assertions      []*apiv1.Assertion
	Controllers     []*apiv1.Controller
	Authentications []*apiv1.Authentication
	DidMetadata     []*apiv1.DIDDocumentMetadata
	Credentials     []*apiv1.VerifiableCredential
	Delegations     []*apiv1.Delegation
	Invocations     []*apiv1.Invocation
	DidControllers  []*apiv1.DIDController
}

// InitGenesisWithORM initializes the module's state from genesis including all ORM tables
// This function handles the ORM data separately from the base GenesisState
func (k *Keeper) InitGenesisWithORM(ctx context.Context, data *types.GenesisState, ormData *GenesisOrmData) error {
	// Initialize params first
	if err := data.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	if err := k.Params.Set(ctx, data.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// If no ORM data provided, return early
	if ormData == nil {
		return nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Import DID Documents
	if ormData.DidDocuments != nil {
		for _, doc := range ormData.DidDocuments {
			if err := k.OrmDB.DIDDocumentTable().Insert(ctx, doc); err != nil {
				sdkCtx.Logger().Error(
					"Failed to import DID document",
					"did", doc.Id,
					"error", err,
				)
			}
		}
	}

	// Import Assertions
	if ormData.Assertions != nil {
		for _, assertion := range ormData.Assertions {
			if err := k.OrmDB.AssertionTable().Insert(ctx, assertion); err != nil {
				sdkCtx.Logger().Error(
					"Failed to import assertion",
					"did", assertion.Did,
					"error", err,
				)
			}
		}
	}

	// Import Controllers
	if ormData.Controllers != nil {
		for _, controller := range ormData.Controllers {
			if err := k.OrmDB.ControllerTable().Insert(ctx, controller); err != nil {
				sdkCtx.Logger().Error(
					"Failed to import controller",
					"did", controller.Did,
					"error", err,
				)
			}
		}
	}

	// Import Authentications
	if ormData.Authentications != nil {
		for _, auth := range ormData.Authentications {
			if err := k.OrmDB.AuthenticationTable().Insert(ctx, auth); err != nil {
				sdkCtx.Logger().Error(
					"Failed to import authentication",
					"did", auth.Did,
					"error", err,
				)
			}
		}
	}

	// Import DID Document Metadata
	if ormData.DidMetadata != nil {
		for _, metadata := range ormData.DidMetadata {
			if err := k.OrmDB.DIDDocumentMetadataTable().Insert(ctx, metadata); err != nil {
				sdkCtx.Logger().Error(
					"Failed to import DID metadata",
					"did", metadata.Did,
					"error", err,
				)
			}
		}
	}

	// Import Verifiable Credentials
	if ormData.Credentials != nil {
		for _, cred := range ormData.Credentials {
			if err := k.OrmDB.VerifiableCredentialTable().Insert(ctx, cred); err != nil {
				sdkCtx.Logger().Error(
					"Failed to import credential",
					"id", cred.Id,
					"error", err,
				)
			}
		}
	}

	sdkCtx.Logger().Info(
		"Genesis import completed",
		"did_documents", len(ormData.DidDocuments),
		"assertions", len(ormData.Assertions),
		"controllers", len(ormData.Controllers),
		"authentications", len(ormData.Authentications),
		"credentials", len(ormData.Credentials),
	)

	return nil
}

// ExportGenesisWithORM exports the module's complete state to genesis
func (k *Keeper) ExportGenesisWithORM(ctx context.Context) (*types.GenesisState, *GenesisOrmData, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get params: %w", err)
	}

	genesis := &types.GenesisState{
		Params: params,
	}

	ormData := &GenesisOrmData{}

	// Export DID Documents
	didDocs, err := k.exportDIDDocuments(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to export DID documents: %w", err)
	}
	ormData.DidDocuments = didDocs

	// Export Assertions
	assertions, err := k.exportAssertions(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to export assertions: %w", err)
	}
	ormData.Assertions = assertions

	// Export Controllers
	controllers, err := k.exportControllers(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to export controllers: %w", err)
	}
	ormData.Controllers = controllers

	// Export Authentications
	auths, err := k.exportAuthentications(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to export authentications: %w", err)
	}
	ormData.Authentications = auths

	// Export DID Metadata
	metadata, err := k.exportDIDMetadata(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to export DID metadata: %w", err)
	}
	ormData.DidMetadata = metadata

	// Export Verifiable Credentials
	creds, err := k.exportCredentials(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to export credentials: %w", err)
	}
	ormData.Credentials = creds

	return genesis, ormData, nil
}

// Helper functions for exporting each table

func (k *Keeper) exportDIDDocuments(ctx context.Context) ([]*apiv1.DIDDocument, error) {
	var documents []*apiv1.DIDDocument

	iter, err := k.OrmDB.DIDDocumentTable().List(ctx, apiv1.DIDDocumentPrimaryKey{})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.Next() {
		doc, err := iter.Value()
		if err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}

	return documents, nil
}

func (k *Keeper) exportAssertions(ctx context.Context) ([]*apiv1.Assertion, error) {
	var assertions []*apiv1.Assertion

	iter, err := k.OrmDB.AssertionTable().List(ctx, apiv1.AssertionPrimaryKey{})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.Next() {
		assertion, err := iter.Value()
		if err != nil {
			return nil, err
		}
		assertions = append(assertions, assertion)
	}

	return assertions, nil
}

func (k *Keeper) exportControllers(ctx context.Context) ([]*apiv1.Controller, error) {
	var controllers []*apiv1.Controller

	iter, err := k.OrmDB.ControllerTable().List(ctx, apiv1.ControllerPrimaryKey{})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.Next() {
		controller, err := iter.Value()
		if err != nil {
			return nil, err
		}
		controllers = append(controllers, controller)
	}

	return controllers, nil
}

func (k *Keeper) exportAuthentications(ctx context.Context) ([]*apiv1.Authentication, error) {
	var auths []*apiv1.Authentication

	iter, err := k.OrmDB.AuthenticationTable().List(ctx, apiv1.AuthenticationPrimaryKey{})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.Next() {
		auth, err := iter.Value()
		if err != nil {
			return nil, err
		}
		auths = append(auths, auth)
	}

	return auths, nil
}

func (k *Keeper) exportDIDMetadata(ctx context.Context) ([]*apiv1.DIDDocumentMetadata, error) {
	var metadata []*apiv1.DIDDocumentMetadata

	iter, err := k.OrmDB.DIDDocumentMetadataTable().List(ctx, apiv1.DIDDocumentMetadataPrimaryKey{})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.Next() {
		meta, err := iter.Value()
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, meta)
	}

	return metadata, nil
}

func (k *Keeper) exportCredentials(ctx context.Context) ([]*apiv1.VerifiableCredential, error) {
	var credentials []*apiv1.VerifiableCredential

	iter, err := k.OrmDB.VerifiableCredentialTable().List(ctx, apiv1.VerifiableCredentialPrimaryKey{})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.Next() {
		cred, err := iter.Value()
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, cred)
	}

	return credentials, nil
}

// ValidateGenesisOrmData validates the ORM data for consistency
func ValidateGenesisOrmData(ormData *GenesisOrmData) error {
	if ormData == nil {
		return nil
	}

	// Check for duplicate DIDs
	didSet := make(map[string]bool)
	for _, doc := range ormData.DidDocuments {
		if didSet[doc.Id] {
			return fmt.Errorf("duplicate DID document: %s", doc.Id)
		}
		didSet[doc.Id] = true
	}

	// Check for duplicate assertions (controller+subject must be unique)
	assertionSet := make(map[string]bool)
	for _, assertion := range ormData.Assertions {
		key := fmt.Sprintf("%s:%s", assertion.Controller, assertion.Subject)
		if assertionSet[key] {
			return fmt.Errorf("duplicate assertion for controller=%s, subject=%s",
				assertion.Controller, assertion.Subject)
		}
		assertionSet[key] = true
	}

	// Check for duplicate controllers (address must be unique)
	addressSet := make(map[string]bool)
	for _, controller := range ormData.Controllers {
		if addressSet[controller.Address] {
			return fmt.Errorf("duplicate controller address: %s", controller.Address)
		}
		addressSet[controller.Address] = true
	}

	return nil
}

// isValidDerivedDID checks if a DID is a valid derived DID (email/tel)
func isValidDerivedDID(did string) bool {
	// Check for email or tel DIDs
	if len(did) > 10 {
		prefix := did[:10]
		if prefix == "did:email:" || len(did) > 8 && did[:8] == "did:tel:" {
			return true
		}
	}
	return false
}
