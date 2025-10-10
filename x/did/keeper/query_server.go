package keeper

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"lukechampine.com/blake3"

	apiv1 "github.com/sonr-io/sonr/api/did/v1"
	"github.com/sonr-io/common/webauthn"
	"github.com/sonr-io/sonr/x/did/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func NewQuerier(keeper Keeper) Querier {
	return Querier{Keeper: keeper}
}

func (k Querier) Params(
	c context.Context,
	req *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	p, err := k.Keeper.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: &p}, nil
}

// ResolveDID implements types.QueryServer.
func (k Querier) ResolveDID(
	goCtx context.Context,
	req *types.QueryResolveDIDRequest,
) (*types.QueryResolveDIDResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, "request cannot be nil")
	}

	if req.Did == "" {
		return nil, errors.Wrap(types.ErrEmptyDID, "DID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get DID document from ORM
	ormDoc, err := k.OrmDB.DIDDocumentTable().Get(ctx, req.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "%s", req.Did)
	}

	// Convert from ORM type
	didDoc := types.DIDDocumentFromORM(ormDoc)

	// Get metadata
	ormMetadata, err := k.OrmDB.DIDDocumentMetadataTable().Get(ctx, req.Did)
	if err != nil {
		// Metadata might not exist, which is ok
		ormMetadata = nil
	}

	var metadata *types.DIDDocumentMetadata
	if ormMetadata != nil {
		metadata = types.DIDDocumentMetadataFromORM(ormMetadata)
	}

	return &types.QueryResolveDIDResponse{
		DidDocument:         didDoc,
		DidDocumentMetadata: metadata,
	}, nil
}

// GetDIDDocument implements types.QueryServer.
func (k Querier) GetDIDDocument(
	goCtx context.Context,
	req *types.QueryGetDIDDocumentRequest,
) (*types.QueryGetDIDDocumentResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, "request cannot be nil")
	}

	if req.Did == "" {
		return nil, errors.Wrap(types.ErrEmptyDID, "DID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get DID document from ORM
	ormDoc, err := k.OrmDB.DIDDocumentTable().Get(ctx, req.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "%s", req.Did)
	}

	// Convert from ORM type
	didDoc := types.DIDDocumentFromORM(ormDoc)

	return &types.QueryGetDIDDocumentResponse{
		DidDocument: didDoc,
	}, nil
}

// ListDIDDocuments implements types.QueryServer.
func (k Querier) ListDIDDocuments(
	goCtx context.Context,
	req *types.QueryListDIDDocumentsRequest,
) (*types.QueryListDIDDocumentsResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Create pagination query
	var pageReq *query.PageRequest
	if req.Pagination != nil {
		pageReq = req.Pagination
	} else {
		// Default pagination
		pageReq = &query.PageRequest{
			Limit: 100,
		}
	}

	// List DID documents with pagination
	var documents []*types.DIDDocument
	pageRes := &query.PageResponse{}

	// Get all documents from the table
	iter, err := k.OrmDB.DIDDocumentTable().List(ctx, apiv1.DIDDocumentPrimaryKey{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list DID documents")
	}
	defer iter.Close()

	// Apply pagination manually
	offset := pageReq.Offset
	limit := pageReq.Limit
	count := uint64(0)
	totalCount := uint64(0)

	for iter.Next() {
		totalCount++

		// Skip items before offset
		if count < offset {
			count++
			continue
		}

		// Stop if we've reached the limit
		if uint64(len(documents)) >= limit && limit > 0 {
			continue
		}

		ormDoc, err := iter.Value()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get DID document from iterator")
		}

		// Convert from ORM type
		didDoc := types.DIDDocumentFromORM(ormDoc)
		documents = append(documents, didDoc)
		count++
	}

	// Set page response
	pageRes.Total = totalCount
	if uint64(len(documents)) < limit || limit == 0 {
		pageRes.NextKey = nil
	} else {
		pageRes.NextKey = []byte(documents[len(documents)-1].Id)
	}

	return &types.QueryListDIDDocumentsResponse{
		DidDocuments: documents,
		Pagination:   pageRes,
	}, nil
}

// GetDIDDocumentsByController implements types.QueryServer.
func (k Querier) GetDIDDocumentsByController(
	goCtx context.Context,
	req *types.QueryGetDIDDocumentsByControllerRequest,
) (*types.QueryGetDIDDocumentsByControllerResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, "request cannot be nil")
	}

	if req.Controller == "" {
		return nil, errors.Wrap(
			types.ErrInvalidControllerAddress,
			"controller address cannot be empty",
		)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get all DID documents by controller using the index
	indexKey := apiv1.DIDDocumentPrimaryControllerIndexKey{}.WithPrimaryController(
		req.Controller,
	)
	iter, err := k.OrmDB.DIDDocumentTable().List(ctx, indexKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list DID documents by controller")
	}
	defer iter.Close()

	var documents []*types.DIDDocument
	for iter.Next() {
		ormDoc, err := iter.Value()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get DID document from iterator")
		}

		// Convert from ORM type
		didDoc := types.DIDDocumentFromORM(ormDoc)
		documents = append(documents, didDoc)
	}

	return &types.QueryGetDIDDocumentsByControllerResponse{
		DidDocuments: documents,
	}, nil
}

// GetVerificationMethod implements types.QueryServer.
func (k Querier) GetVerificationMethod(
	goCtx context.Context,
	req *types.QueryGetVerificationMethodRequest,
) (*types.QueryGetVerificationMethodResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, "request cannot be nil")
	}

	if req.Did == "" {
		return nil, errors.Wrap(types.ErrEmptyDID, "DID cannot be empty")
	}

	if req.MethodId == "" {
		return nil, errors.Wrap(
			types.ErrEmptyVerificationMethodID,
			"verification method ID cannot be empty",
		)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get DID document from ORM
	ormDoc, err := k.OrmDB.DIDDocumentTable().Get(ctx, req.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "%s", req.Did)
	}

	// Convert from ORM type
	didDoc := types.DIDDocumentFromORM(ormDoc)

	// Find the verification method
	for _, vm := range didDoc.VerificationMethod {
		if vm.Id == req.MethodId {
			return &types.QueryGetVerificationMethodResponse{
				VerificationMethod: vm,
			}, nil
		}
	}

	return nil, errors.Wrapf(types.ErrVerificationMethodNotFound, "%s", req.MethodId)
}

// GetService implements types.QueryServer.
func (k Querier) GetService(
	goCtx context.Context,
	req *types.QueryGetServiceRequest,
) (*types.QueryGetServiceResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, "request cannot be nil")
	}

	if req.Did == "" {
		return nil, errors.Wrap(types.ErrEmptyDID, "DID cannot be empty")
	}

	if req.ServiceId == "" {
		return nil, errors.Wrap(types.ErrEmptyServiceID, "service ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get DID document from ORM
	ormDoc, err := k.OrmDB.DIDDocumentTable().Get(ctx, req.Did)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "%s", req.Did)
	}

	// Convert from ORM type
	didDoc := types.DIDDocumentFromORM(ormDoc)

	// Find the service
	for _, svc := range didDoc.Service {
		if svc.Id == req.ServiceId {
			return &types.QueryGetServiceResponse{
				Service: svc,
			}, nil
		}
	}

	return nil, errors.Wrapf(types.ErrServiceNotFound, "%s", req.ServiceId)
}

// GetVerifiableCredential implements types.QueryServer.
func (k Querier) GetVerifiableCredential(
	goCtx context.Context,
	req *types.QueryGetVerifiableCredentialRequest,
) (*types.QueryGetVerifiableCredentialResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, "request cannot be nil")
	}

	if req.CredentialId == "" {
		return nil, errors.Wrap(types.ErrEmptyCredentialID, "credential ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get credential from ORM
	ormCred, err := k.OrmDB.VerifiableCredentialTable().Get(ctx, req.CredentialId)
	if err != nil {
		return nil, errors.Wrapf(types.ErrCredentialNotFound, "%s", req.CredentialId)
	}

	// Convert from ORM type
	credential := types.VerifiableCredentialFromORM(ormCred)

	return &types.QueryGetVerifiableCredentialResponse{
		Credential: credential,
	}, nil
}

// ListVerifiableCredentials implements types.QueryServer with filtering support.
func (k Querier) ListVerifiableCredentials(
	goCtx context.Context,
	req *types.QueryListVerifiableCredentialsRequest,
) (*types.QueryListVerifiableCredentialsResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, "request cannot be nil")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Create pagination query
	var pageReq *query.PageRequest
	if req.Pagination != nil {
		pageReq = req.Pagination
	} else {
		// Default pagination
		pageReq = &query.PageRequest{
			Limit: 100,
		}
	}

	var credentials []*types.VerifiableCredential
	pageRes := &query.PageResponse{}

	// Determine which index to use based on filters
	var iter apiv1.VerifiableCredentialIterator
	var err error

	if req.Issuer != "" {
		// Use issuer index
		indexKey := apiv1.VerifiableCredentialIssuerIndexKey{}.WithIssuer(req.Issuer)
		iter, err = k.OrmDB.VerifiableCredentialTable().List(ctx, indexKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list credentials by issuer")
		}
	} else if req.Holder != "" {
		// Use subject/holder index
		indexKey := apiv1.VerifiableCredentialSubjectIndexKey{}.WithSubject(req.Holder)
		iter, err = k.OrmDB.VerifiableCredentialTable().List(ctx, indexKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list credentials by holder")
		}
	} else {
		// List all credentials
		iter, err = k.OrmDB.VerifiableCredentialTable().
			List(ctx, apiv1.VerifiableCredentialPrimaryKey{})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list credentials")
		}
	}
	defer iter.Close()

	// Apply pagination and filters
	offset := pageReq.Offset
	limit := pageReq.Limit
	count := uint64(0)
	totalCount := uint64(0)

	for iter.Next() {
		ormCred, err := iter.Value()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get credential from iterator")
		}

		// Convert from ORM type
		credential := types.VerifiableCredentialFromORM(ormCred)

		// Apply additional filters
		if req.Issuer != "" && credential.Issuer != req.Issuer {
			continue
		}
		if req.Holder != "" && credential.Subject != req.Holder {
			continue
		}
		if !req.IncludeRevoked && credential.Revoked {
			continue
		}

		totalCount++

		// Skip items before offset
		if count < offset {
			count++
			continue
		}

		// Stop if we've reached the limit
		if uint64(len(credentials)) >= limit && limit > 0 {
			continue
		}

		credentials = append(credentials, credential)
		count++
	}

	// Set page response
	pageRes.Total = totalCount
	if uint64(len(credentials)) < limit || limit == 0 {
		pageRes.NextKey = nil
	} else if len(credentials) > 0 {
		pageRes.NextKey = []byte(credentials[len(credentials)-1].Id)
	}

	return &types.QueryListVerifiableCredentialsResponse{
		Credentials: credentials,
		Pagination:  pageRes,
	}, nil
}

// getVaultInfoForDID retrieves vault information for a DID
func (k Querier) getVaultInfoForDID(
	ctx sdk.Context,
	did string,
) (vaultId string, isEncrypted bool) {
	// Default values
	vaultId = ""
	isEncrypted = false

	// Check if DWN keeper is available
	if k.dwnKeeper == nil {
		return vaultId, isEncrypted
	}

	// Query vaults associated with the DID
	vaults, err := k.dwnKeeper.GetVaultsByDID(ctx, did)
	if err != nil || len(vaults) == 0 {
		return vaultId, isEncrypted
	}

	// Use the first active vault
	for _, vault := range vaults {
		if vault.Status == "active" {
			vaultId = vault.VaultID
			isEncrypted = true // Assume vault storage means encryption
			return vaultId, isEncrypted
		}
	}

	return vaultId, isEncrypted
}

// GetCredentialsByDID implements types.QueryServer - unified method for all credentials.
func (k Querier) GetCredentialsByDID(
	goCtx context.Context,
	req *types.QueryGetCredentialsByDIDRequest,
) (*types.QueryGetCredentialsByDIDResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, "request cannot be nil")
	}

	if req.Did == "" {
		return nil, errors.Wrap(types.ErrEmptyDID, "DID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var credentialInfos []*types.CredentialInfo

	// 1. Get WebAuthn credentials if requested (default: true)
	if req.IncludeWebauthn || (!req.IncludeVerifiable && !req.IncludeWebauthn) {
		webauthnCreds, err := k.GetWebAuthnCredentialsByDID(ctx, req.Did)
		if err == nil {
			for _, wc := range webauthnCreds {
				info := &types.CredentialInfo{
					Credential: &types.CredentialInfo_WebauthnCredential{
						WebauthnCredential: wc,
					},
				}

				// Check if stored in vault
				vaultId, isEncrypted := k.getVaultInfoForDID(ctx, req.Did)
				info.VaultId = vaultId
				info.IsEncrypted = isEncrypted

				credentialInfos = append(credentialInfos, info)
			}
		}
	}

	// 2. Get Verifiable Credentials if requested (default: true)
	if req.IncludeVerifiable || (!req.IncludeVerifiable && !req.IncludeWebauthn) {
		// Get credentials where DID is issuer
		issuerIndex := apiv1.VerifiableCredentialIssuerIndexKey{}.WithIssuer(req.Did)
		issuerIter, err := k.OrmDB.VerifiableCredentialTable().List(ctx, issuerIndex)
		if err == nil {
			defer issuerIter.Close()
			for issuerIter.Next() {
				ormCred, err := issuerIter.Value()
				if err != nil {
					continue
				}

				credential := types.VerifiableCredentialFromORM(ormCred)

				// Skip revoked if not requested
				if credential.Revoked && !req.IncludeRevoked {
					continue
				}

				info := &types.CredentialInfo{
					Credential: &types.CredentialInfo_VerifiableCredential{
						VerifiableCredential: credential,
					},
				}

				// Check if stored in vault
				vaultId, isEncrypted := k.getVaultInfoForDID(ctx, req.Did)
				info.VaultId = vaultId
				info.IsEncrypted = isEncrypted

				credentialInfos = append(credentialInfos, info)
			}
		}

		// Get credentials where DID is holder
		holderIndex := apiv1.VerifiableCredentialSubjectIndexKey{}.WithSubject(req.Did)
		holderIter, err := k.OrmDB.VerifiableCredentialTable().List(ctx, holderIndex)
		if err == nil {
			defer holderIter.Close()
			for holderIter.Next() {
				ormCred, err := holderIter.Value()
				if err != nil {
					continue
				}

				credential := types.VerifiableCredentialFromORM(ormCred)

				// Skip if already added as issuer
				isDuplicate := false
				for _, existing := range credentialInfos {
					if vc := existing.GetVerifiableCredential(); vc != nil &&
						vc.Id == credential.Id {
						isDuplicate = true
						break
					}
				}
				if isDuplicate {
					continue
				}

				// Skip revoked if not requested
				if credential.Revoked && !req.IncludeRevoked {
					continue
				}

				info := &types.CredentialInfo{
					Credential: &types.CredentialInfo_VerifiableCredential{
						VerifiableCredential: credential,
					},
				}

				// Check if stored in vault
				vaultId, isEncrypted := k.getVaultInfoForDID(ctx, req.Did)
				info.VaultId = vaultId
				info.IsEncrypted = isEncrypted

				credentialInfos = append(credentialInfos, info)
			}
		}
	}

	// Apply pagination
	pageRes := &query.PageResponse{}
	if req.Pagination != nil {
		start := int(req.Pagination.Offset)
		end := start + int(req.Pagination.Limit)

		if start > len(credentialInfos) {
			credentialInfos = []*types.CredentialInfo{}
		} else if end > len(credentialInfos) {
			credentialInfos = credentialInfos[start:]
		} else {
			credentialInfos = credentialInfos[start:end]
			pageRes.NextKey = []byte("next")
		}
	}

	pageRes.Total = uint64(len(credentialInfos))

	return &types.QueryGetCredentialsByDIDResponse{
		Credentials: credentialInfos,
		Pagination:  pageRes,
	}, nil
}

// RegisterStart implements types.QueryServer.
// TODO: Complete implementation per issue #269
func (k Querier) RegisterStart(
	goCtx context.Context,
	req *types.QueryRegisterStartRequest,
) (*types.QueryRegisterStartResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, "request cannot be nil")
	}

	// Validate assertion DID format
	if req.AssertionDid == "" {
		return nil, errors.Wrap(types.ErrInvalidRequest, "assertion_did cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check that assertion DID does NOT exist (opposite of LoginStart)
	existingAssertion, err := k.OrmDB.AssertionTable().Get(ctx, req.AssertionDid)
	if err == nil && existingAssertion != nil {
		return nil, errors.Wrapf(types.ErrDIDAlreadyExists, "assertion already exists: %s", req.AssertionDid)
	}

	// Generate deterministic challenge
	challenge, nonce, err := k.generateWebAuthnChallenge(ctx, req.AssertionDid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate challenge")
	}

	// Get relying party ID from params
	params, err := k.Keeper.Params.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get module params")
	}

	// Extract user information from assertion DID
	userName, displayName := k.ExtractUserInfoFromAssertionDID(req.AssertionDid)
	user := map[string]string{
		"id":          req.AssertionDid, // Full DID as unique identifier
		"name":        userName,         // Human-readable name (email/phone type)
		"displayName": displayName,      // User-friendly display name
	}

	// Store registration session for later validation
	if err := k.storeRegistrationSession(ctx, req.AssertionDid, challenge, nonce); err != nil {
		return nil, errors.Wrap(err, "failed to store registration session")
	}

	// Convert challenge string to bytes
	challengeBytes := []byte(challenge)

	// Get relying party ID from webauthn params
	rpId := ""
	if params.Webauthn != nil {
		rpId = params.Webauthn.DefaultRpId
	}

	return &types.QueryRegisterStartResponse{
		Challenge:      challengeBytes,
		RelyingPartyId: rpId,
		User:           user,
	}, nil
}

// LoginStart implements types.QueryServer.
// TODO: Complete implementation per issue #269
func (k Querier) LoginStart(
	goCtx context.Context,
	req *types.QueryLoginStartRequest,
) (*types.QueryLoginStartResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidRequest, "request cannot be nil")
	}

	// Validate assertion DID format
	if req.AssertionDid == "" {
		return nil, errors.Wrap(types.ErrInvalidRequest, "assertion_did cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify assertion DID exists (opposite of RegisterStart)
	assertion, err := k.OrmDB.AssertionTable().Get(ctx, req.AssertionDid)
	if err != nil {
		return nil, errors.Wrapf(types.ErrAssertionNotFound, "assertion DID %s not found", req.AssertionDid)
	}

	// Log login attempt for security audit trail
	ctx.Logger().Info("WebAuthn login start",
		"assertion_did", req.AssertionDid,
		"controller", assertion.Controller,
		"block_height", ctx.BlockHeight(),
	)

	// Get controller DID from assertion
	if assertion.Controller == "" {
		return nil, errors.Wrapf(types.ErrInvalidAssertion, "assertion %s has no controller", req.AssertionDid)
	}

	// Get controller's DID document
	controllerDoc, err := k.OrmDB.DIDDocumentTable().Get(ctx, assertion.Controller)
	if err != nil {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "controller DID %s not found", assertion.Controller)
	}

	// Check if DID is deactivated
	if controllerDoc.Deactivated {
		return nil, errors.Wrapf(types.ErrDIDDeactivated, "controller DID %s is deactivated", assertion.Controller)
	}

	// Extract credential IDs from authentication methods (WebAuthn only)
	credentialIds, err := k.ExtractWebAuthnCredentialIDs(ctx, controllerDoc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract WebAuthn credential IDs")
	}

	if len(credentialIds) == 0 {
		return nil, errors.Wrapf(types.ErrNoCredentials, "no WebAuthn credentials found for controller %s", assertion.Controller)
	}

	ctx.Logger().Debug("extracted WebAuthn credentials",
		"controller", assertion.Controller,
		"credential_count", len(credentialIds),
	)

	// Generate deterministic challenge
	challenge, nonce, err := k.generateWebAuthnChallenge(ctx, req.AssertionDid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate challenge")
	}

	// Get relying party ID from params
	params, err := k.Keeper.Params.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get module params")
	}

	// Store login session for later validation
	if err := k.storeLoginSession(ctx, req.AssertionDid, challenge, nonce); err != nil {
		return nil, errors.Wrap(err, "failed to store login session")
	}

	// Convert challenge string to bytes
	challengeBytes := []byte(challenge)

	// Get relying party ID from webauthn params
	rpId := ""
	if params.Webauthn != nil {
		rpId = params.Webauthn.DefaultRpId
	}

	return &types.QueryLoginStartResponse{
		CredentialIds:  credentialIds,
		Challenge:      challengeBytes,
		RelyingPartyId: rpId,
	}, nil
}

// Helper methods for WebAuthn validation

// ValidateServiceOrigin validates the service origin for WebAuthn operations.
// Validation rules:
// 1. Origin format must be valid (scheme://host[:port])
// 2. HTTPS is required except for localhost/127.0.0.1
// 3. Check against x/svc module domain registry (if available)
// 4. Check against module params allowed origins list
// 5. Support wildcard subdomain matching (*.example.com)
func (k Querier) ValidateServiceOrigin(ctx sdk.Context, origin string) error {
	// Validate origin format
	if origin == "" {
		return errors.Wrap(types.ErrInvalidRequest, "origin cannot be empty")
	}

	// Check for valid scheme
	if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
		return errors.Wrapf(types.ErrInvalidRequest, "origin must start with http:// or https://, got: %s", origin)
	}

	// Extract domain from origin
	domain := ExtractDomainFromOrigin(origin)
	if domain == "" {
		return errors.Wrapf(types.ErrInvalidRequest, "invalid origin format: %s", origin)
	}

	// Allow localhost and 127.0.0.1 with both HTTP and HTTPS
	if k.IsLocalhostOrigin(domain) {
		ctx.Logger().Debug("allowing localhost origin", "origin", origin)
		return nil
	}

	// Non-localhost origins must use HTTPS
	if strings.HasPrefix(origin, "http://") {
		return errors.Wrapf(types.ErrInvalidRequest, "non-localhost origins must use HTTPS, got: %s", origin)
	}

	// Try x/svc module verification first (if keeper available)
	if k.serviceKeeper != nil {
		if err := k.serviceKeeper.VerifyOrigin(ctx, origin); err == nil {
			ctx.Logger().Debug("origin verified via x/svc module", "origin", origin)
			return nil
		}
		// Continue to check module params if x/svc verification fails
		ctx.Logger().Debug("origin not found in x/svc registry, checking module params", "origin", origin)
	}

	// Check module params allowed origins list
	params, err := k.Keeper.Params.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get module params")
	}

	if params.Webauthn == nil || len(params.Webauthn.AllowedOrigins) == 0 {
		return errors.Wrapf(types.ErrInvalidRequest,
			"origin %s not registered in x/svc and no allowed origins configured in module params", origin)
	}

	// Check against allowed origins with wildcard support
	for _, allowedOrigin := range params.Webauthn.AllowedOrigins {
		if k.MatchesOrigin(origin, domain, allowedOrigin) {
			ctx.Logger().Debug("origin matched in allowed origins", "origin", origin, "pattern", allowedOrigin)
			return nil
		}
	}

	return errors.Wrapf(types.ErrInvalidRequest,
		"origin %s not registered in x/svc module and not in allowed origins list", origin)
}

// IsLocalhostOrigin checks if the domain is localhost or 127.0.0.1
func (k Querier) IsLocalhostOrigin(domain string) bool {
	return domain == "localhost" || domain == "127.0.0.1" || domain == "[::1]"
}

// MatchesOrigin checks if an origin matches an allowed origin pattern.
// Supports wildcard subdomain matching (*.example.com matches app.example.com)
func (k Querier) MatchesOrigin(fullOrigin, domain, allowedOrigin string) bool {
	// Exact match
	if fullOrigin == allowedOrigin {
		return true
	}

	// Extract domain from allowed origin for wildcard matching
	allowedDomain := ExtractDomainFromOrigin(allowedOrigin)

	// Check for wildcard subdomain pattern (*.example.com)
	if strings.HasPrefix(allowedDomain, "*.") {
		baseDomain := strings.TrimPrefix(allowedDomain, "*.")
		// Match if domain ends with .baseDomain
		if strings.HasSuffix(domain, "."+baseDomain) {
			return true
		}
		// Also match the base domain itself (example.com matches *.example.com)
		if domain == baseDomain {
			return true
		}
	}

	return false
}

// ExtractDomainFromOrigin extracts the domain from an origin URL
func ExtractDomainFromOrigin(origin string) string {
	// Remove protocol
	domain := strings.TrimPrefix(origin, "https://")
	domain = strings.TrimPrefix(domain, "http://")

	// Remove path if present (do this before port to handle IPv6)
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// Handle IPv6 addresses (enclosed in brackets)
	if strings.HasPrefix(domain, "[") {
		// Find closing bracket
		if idx := strings.Index(domain, "]"); idx != -1 {
			// Return everything up to and including the closing bracket
			domain = domain[:idx+1]
		}
	} else {
		// For non-IPv6, remove port if present
		if idx := strings.Index(domain, ":"); idx != -1 {
			domain = domain[:idx]
		}
	}

	return domain
}

// ExtractUserInfoFromAssertionDID extracts human-readable user information from assertion DID.
// Assertion DID format: did:sonr:email:<blake3_hash> or did:sonr:phone:<blake3_hash>
// Returns: (name, displayName) suitable for WebAuthn user object
func (k Querier) ExtractUserInfoFromAssertionDID(assertionDid string) (name string, displayName string) {
	// Remove "did:" prefix
	trimmed := types.TrimDIDMethodPrefix(assertionDid)

	// Parse format: sonr:email:<hash> or sonr:phone:<hash>
	parts := strings.Split(trimmed, ":")
	if len(parts) < 3 {
		// Fallback if format doesn't match expected
		return assertionDid, assertionDid
	}

	// parts[0] = "sonr"
	// parts[1] = assertion type (email, phone, etc.)
	// parts[2] = blake3 hash
	assertionType := parts[1]
	hashValue := parts[2]

	// Create human-readable name based on assertion type
	switch assertionType {
	case "email":
		name = fmt.Sprintf("Email User")
		displayName = fmt.Sprintf("Email (%s...)", hashValue[:8])
	case "phone", "tel":
		name = fmt.Sprintf("Phone User")
		displayName = fmt.Sprintf("Phone (%s...)", hashValue[:8])
	case "github":
		name = fmt.Sprintf("GitHub User")
		displayName = fmt.Sprintf("GitHub (%s...)", hashValue[:8])
	case "google":
		name = fmt.Sprintf("Google User")
		displayName = fmt.Sprintf("Google (%s...)", hashValue[:8])
	default:
		name = fmt.Sprintf("%s User", strings.Title(assertionType))
		displayName = fmt.Sprintf("%s (%s...)", strings.Title(assertionType), hashValue[:8])
	}

	return name, displayName
}

// ExtractWebAuthnCredentialIDs extracts WebAuthn credential IDs from a DID document's authentication methods.
// Only returns credentials with WebAuthn type verification methods.
func (k Querier) ExtractWebAuthnCredentialIDs(ctx sdk.Context, didDoc *apiv1.DIDDocument) ([]string, error) {
	var credentialIds []string

	// Iterate through authentication verification method references
	for _, authRef := range didDoc.Authentication {
		// Check if it's an embedded verification method
		if authRef.EmbeddedVerificationMethod != nil {
			vm := authRef.EmbeddedVerificationMethod

			// Check if this is a WebAuthn verification method
			if vm.WebauthnCredential != nil && vm.WebauthnCredential.CredentialId != "" {
				credentialIds = append(credentialIds, vm.WebauthnCredential.CredentialId)
				ctx.Logger().Debug("found WebAuthn credential",
					"credential_id", vm.WebauthnCredential.CredentialId,
					"verification_method_id", vm.Id,
				)
			}
		} else if authRef.VerificationMethodId != "" {
			// It's a reference - need to look up in verificationMethod array
			for _, vm := range didDoc.VerificationMethod {
				if vm.Id == authRef.VerificationMethodId {
					// Check if this is a WebAuthn verification method
					if vm.WebauthnCredential != nil && vm.WebauthnCredential.CredentialId != "" {
						credentialIds = append(credentialIds, vm.WebauthnCredential.CredentialId)
						ctx.Logger().Debug("found WebAuthn credential",
							"credential_id", vm.WebauthnCredential.CredentialId,
							"verification_method_id", vm.Id,
						)
					}
					break
				}
			}
		}
	}

	return credentialIds, nil
}

// generateWebAuthnChallenge generates a deterministic cryptographic challenge for WebAuthn ceremonies
// The challenge is a BLAKE3 hash of: assertion_did || block_height || chain_id || nonce
// Returns: base64url-encoded challenge (32 bytes), nonce, error
func (k Querier) generateWebAuthnChallenge(ctx sdk.Context, assertionDid string) (string, string, error) {
	// Generate a unique nonce based on block time and assertion DID
	// This ensures uniqueness even within the same block
	nonce := fmt.Sprintf("%d:%s:%d",
		ctx.BlockHeight(),
		assertionDid,
		ctx.BlockTime().UnixNano())

	// Construct deterministic input for BLAKE3 hash
	// Format: assertion_did || block_height || chain_id || nonce
	input := fmt.Sprintf("%s||%d||%s||%s",
		assertionDid,
		ctx.BlockHeight(),
		ctx.ChainID(),
		nonce)

	// Generate BLAKE3 hash (32 bytes)
	hash := blake3.Sum256([]byte(input))

	// Encode as base64url for WebAuthn (32 bytes meets spec requirement of 16-64 bytes)
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return challenge, nonce, nil
}

// storeRegistrationSession stores the registration session for later validation
func (k Querier) storeRegistrationSession(
	ctx sdk.Context,
	assertionDid string,
	challenge string,
	nonce string,
) error {
	// TODO: Implement session storage in state
	// This will store:
	// - assertion_did
	// - challenge
	// - nonce (for replay prevention)
	// - expiration timestamp (e.g., 5 minutes from now)
	// - session_type: "registration"

	// For now, just log that we would store this
	ctx.Logger().Debug("storing registration session",
		"assertion_did", assertionDid,
		"challenge_length", len(challenge),
		"nonce", nonce,
	)

	return nil
}

// storeLoginSession stores the login session for later validation
func (k Querier) storeLoginSession(
	ctx sdk.Context,
	assertionDid string,
	challenge string,
	nonce string,
) error {
	// TODO: Implement session storage in state
	// This will store:
	// - assertion_did
	// - challenge
	// - nonce (for replay prevention)
	// - expiration timestamp (e.g., 5 minutes from now)
	// - session_type: "login"

	// For now, just log that we would store this
	ctx.Logger().Debug("storing login session",
		"assertion_did", assertionDid,
		"challenge_length", len(challenge),
		"nonce", nonce,
	)

	return nil
}

// createWebAuthnOptions creates WebAuthn credential creation options
func (k Querier) createWebAuthnOptions(challenge, origin, username string) []byte {
	// Extract RP ID from origin
	rpID := ExtractDomainFromOrigin(origin)
	if rpID == "" {
		rpID = "localhost"
	}

	// Create WebAuthn options
	options := webauthn.PublicKeyCredentialCreationOptions{
		Challenge: webauthn.URLEncodedBase64(challenge),
		RelyingParty: webauthn.RelyingPartyEntity{
			ID: rpID,
		},
		User: webauthn.UserEntity{
			ID:          []byte(username),
			DisplayName: username,
		},
		Parameters: []webauthn.CredentialParameter{
			{
				Type:      webauthn.PublicKeyCredentialType,
				Algorithm: -7, // ES256
			},
			{
				Type:      webauthn.PublicKeyCredentialType,
				Algorithm: -257, // RS256
			},
		},
		Timeout: 60000, // 60 seconds
		AuthenticatorSelection: webauthn.AuthenticatorSelection{
			AuthenticatorAttachment: webauthn.Platform,
			RequireResidentKey:      boolPtr(false),
			ResidentKey:             webauthn.ResidentKeyRequirementDiscouraged,
			UserVerification:        webauthn.VerificationPreferred,
		},
		Attestation: webauthn.PreferNoAttestation,
	}

	// Marshal to JSON
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		k.logger.Error("Failed to marshal WebAuthn options", "error", err)
		return []byte("{}")
	}

	return optionsJSON
}

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}
