package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/orm/types/ormerrors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/ipfs/go-cid"

	apiv1 "github.com/sonr-io/sonr/api/dwn/v1"
	"github.com/sonr-io/sonr/x/dwn/types"
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

// Records queries DWN records with filters
func (k Querier) Records(
	c context.Context,
	req *types.QueryRecordsRequest,
) (*types.QueryRecordsResponse, error) {
	if req == nil {
		return nil, types.ErrRequestCannotBeNil
	}

	if req.Target == "" {
		return nil, types.ErrTargetDIDEmpty
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Build index key based on filters
	var indexKey apiv1.DWNRecordIndexKey

	if req.Protocol != "" {
		indexKey = apiv1.DWNRecordTargetProtocolIndexKey{}.WithTargetProtocol(
			req.Target,
			req.Protocol,
		)
	} else if req.Schema != "" {
		indexKey = apiv1.DWNRecordTargetSchemaIndexKey{}.WithTargetSchema(req.Target, req.Schema)
	} else if req.ParentId != "" {
		indexKey = apiv1.DWNRecordParentIdIndexKey{}.WithParentId(req.ParentId)
	} else {
		indexKey = apiv1.DWNRecordTargetProtocolIndexKey{}.WithTarget(req.Target)
	}

	// Query with pagination
	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &query.PageRequest{Limit: 100}
	}

	records := []types.DWNRecord{}
	pageRes := &query.PageResponse{}

	iter, err := k.OrmDB.DWNRecordTable().List(ctx, indexKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list records")
	}
	defer iter.Close()

	count := uint64(0)
	offset := pageReq.Offset
	limit := pageReq.Limit

	for iter.Next() {
		record, err := iter.Value()
		if err != nil {
			continue
		}

		// Apply published filter
		if req.PublishedOnly && !record.Published {
			continue
		}

		count++

		// Handle pagination
		if count <= offset {
			continue
		}

		if uint64(len(records)) >= limit {
			pageRes.NextKey = []byte(record.RecordId)
			break
		}

		records = append(records, types.ConvertAPIRecordToType(record))
	}

	pageRes.Total = count

	return &types.QueryRecordsResponse{
		Records:    records,
		Pagination: pageRes,
	}, nil
}

// Record queries a specific DWN record by ID
func (k Querier) Record(
	c context.Context,
	req *types.QueryRecordRequest,
) (*types.QueryRecordResponse, error) {
	if req == nil {
		return nil, types.ErrRequestCannotBeNil
	}

	if req.Target == "" {
		return nil, types.ErrTargetDIDEmpty
	}

	if req.RecordId == "" {
		return nil, types.ErrRecordIDEmpty
	}

	ctx := sdk.UnwrapSDKContext(c)

	record, err := k.OrmDB.DWNRecordTable().Get(ctx, req.RecordId)
	if err != nil {
		if ormerrors.IsNotFound(err) {
			return nil, errors.Wrapf(types.ErrRecordNotFound, "record %s not found", req.RecordId)
		}
		return nil, errors.Wrap(err, "failed to get record")
	}

	// Verify the record belongs to the target DWN
	if record.Target != req.Target {
		return nil, errors.Wrap(sdkerrors.ErrUnauthorized, "record does not belong to target DWN")
	}
	rec := types.ConvertAPIRecordToType(record)
	return &types.QueryRecordResponse{
		Record: &rec,
	}, nil
}

// Protocols queries DWN protocols
func (k Querier) Protocols(
	c context.Context,
	req *types.QueryProtocolsRequest,
) (*types.QueryProtocolsResponse, error) {
	if req == nil {
		return nil, types.ErrRequestCannotBeNil
	}

	if req.Target == "" {
		return nil, types.ErrTargetDIDEmpty
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Query with pagination
	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &query.PageRequest{Limit: 100}
	}

	protocols := []types.DWNProtocol{}
	pageRes := &query.PageResponse{}

	indexKey := apiv1.DWNProtocolTargetProtocolUriIndexKey{}.WithTarget(req.Target)
	iter, err := k.OrmDB.DWNProtocolTable().List(ctx, indexKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list protocols")
	}
	defer iter.Close()

	count := uint64(0)
	offset := pageReq.Offset
	limit := pageReq.Limit

	for iter.Next() {
		protocol, err := iter.Value()
		if err != nil {
			continue
		}

		// Apply published filter
		if req.PublishedOnly && !protocol.Published {
			continue
		}

		count++

		// Handle pagination
		if count <= offset {
			continue
		}

		if uint64(len(protocols)) >= limit {
			pageRes.NextKey = []byte(protocol.ProtocolUri)
			break
		}

		protocols = append(protocols, types.ConvertAPIProtocolToType(protocol))
	}

	pageRes.Total = count

	return &types.QueryProtocolsResponse{
		Protocols:  protocols,
		Pagination: pageRes,
	}, nil
}

// Protocol queries a specific DWN protocol
func (k Querier) Protocol(
	c context.Context,
	req *types.QueryProtocolRequest,
) (*types.QueryProtocolResponse, error) {
	if req == nil {
		return nil, types.ErrRequestCannotBeNil
	}

	if req.Target == "" {
		return nil, types.ErrTargetDIDEmpty
	}

	if req.ProtocolUri == "" {
		return nil, types.ErrProtocolURIEmpty
	}

	ctx := sdk.UnwrapSDKContext(c)

	protocol, err := k.OrmDB.DWNProtocolTable().Get(ctx, req.Target, req.ProtocolUri)
	if err != nil {
		if ormerrors.IsNotFound(err) {
			return nil, errors.Wrapf(
				types.ErrProtocolNotFound,
				"protocol %s not found",
				req.ProtocolUri,
			)
		}
		return nil, errors.Wrap(err, "failed to get protocol")
	}
	prot := types.ConvertAPIProtocolToType(protocol)
	return &types.QueryProtocolResponse{
		Protocol: &prot,
	}, nil
}

// Permissions queries DWN permissions
func (k Querier) Permissions(
	c context.Context,
	req *types.QueryPermissionsRequest,
) (*types.QueryPermissionsResponse, error) {
	if req == nil {
		return nil, types.ErrRequestCannotBeNil
	}

	if req.Target == "" {
		return nil, types.ErrTargetDIDEmpty
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Build index key based on filters
	var indexKey apiv1.DWNPermissionIndexKey

	if req.Grantor != "" && req.Grantee != "" {
		indexKey = apiv1.DWNPermissionGrantorGranteeIndexKey{}.WithGrantorGrantee(
			req.Grantor,
			req.Grantee,
		)
	} else if req.Grantor != "" {
		indexKey = apiv1.DWNPermissionGrantorGranteeIndexKey{}.WithGrantor(req.Grantor)
	} else if req.InterfaceName != "" && req.Method != "" {
		indexKey = apiv1.DWNPermissionTargetInterfaceNameMethodIndexKey{}.WithTargetInterfaceNameMethod(req.Target, req.InterfaceName, req.Method)
	} else if req.InterfaceName != "" {
		indexKey = apiv1.DWNPermissionTargetInterfaceNameMethodIndexKey{}.WithTargetInterfaceName(req.Target, req.InterfaceName)
	} else {
		indexKey = apiv1.DWNPermissionTargetInterfaceNameMethodIndexKey{}.WithTarget(req.Target)
	}

	// Query with pagination
	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &query.PageRequest{Limit: 100}
	}

	permissions := []types.DWNPermission{}
	pageRes := &query.PageResponse{}

	iter, err := k.OrmDB.DWNPermissionTable().List(ctx, indexKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list permissions")
	}
	defer iter.Close()

	count := uint64(0)
	offset := pageReq.Offset
	limit := pageReq.Limit

	for iter.Next() {
		permission, err := iter.Value()
		if err != nil {
			continue
		}

		// Apply filters
		if permission.Target != req.Target {
			continue
		}

		if !req.IncludeRevoked && permission.Revoked {
			continue
		}

		count++

		// Handle pagination
		if count <= offset {
			continue
		}

		if uint64(len(permissions)) >= limit {
			pageRes.NextKey = []byte(permission.PermissionId)
			break
		}

		permissions = append(permissions, types.ConvertAPIPermissionToType(permission))
	}

	pageRes.Total = count

	return &types.QueryPermissionsResponse{
		Permissions: permissions,
		Pagination:  pageRes,
	}, nil
}

// Vault queries a specific vault
func (k Querier) Vault(
	c context.Context,
	req *types.QueryVaultRequest,
) (*types.QueryVaultResponse, error) {
	if req == nil {
		return nil, types.ErrRequestCannotBeNil
	}

	if req.VaultId == "" {
		return nil, types.ErrVaultIDEmpty
	}

	ctx := sdk.UnwrapSDKContext(c)

	vault, err := k.OrmDB.VaultStateTable().Get(ctx, req.VaultId)
	if err != nil {
		if ormerrors.IsNotFound(err) {
			return nil, errors.Wrapf(types.ErrVaultNotFound, "vault %s not found", req.VaultId)
		}
		return nil, errors.Wrap(err, "failed to get vault")
	}
	vlt := types.ConvertAPIVaultToType(vault)
	return &types.QueryVaultResponse{
		Vault: &vlt,
	}, nil
}

// Vaults queries vaults by owner
func (k Querier) Vaults(
	c context.Context,
	req *types.QueryVaultsRequest,
) (*types.QueryVaultsResponse, error) {
	if req == nil {
		return nil, types.ErrRequestCannotBeNil
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Query with pagination
	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &query.PageRequest{Limit: 100}
	}

	vaults := []types.VaultState{}
	pageRes := &query.PageResponse{}

	var iter apiv1.VaultStateIterator
	var err error

	if req.Owner != "" {
		indexKey := apiv1.VaultStateOwnerIndexKey{}.WithOwner(req.Owner)
		iter, err = k.OrmDB.VaultStateTable().List(ctx, indexKey)
	} else {
		// List all vaults
		indexKey := apiv1.VaultStatePrimaryKey{}
		iter, err = k.OrmDB.VaultStateTable().List(ctx, indexKey)
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to list vaults")
	}
	defer iter.Close()

	count := uint64(0)
	offset := pageReq.Offset
	limit := pageReq.Limit

	for iter.Next() {
		vault, err := iter.Value()
		if err != nil {
			continue
		}

		count++

		// Handle pagination
		if count <= offset {
			continue
		}

		if uint64(len(vaults)) >= limit {
			pageRes.NextKey = []byte(vault.VaultId)
			break
		}

		vaults = append(vaults, types.ConvertAPIVaultToType(vault))
	}

	pageRes.Total = count

	return &types.QueryVaultsResponse{
		Vaults:     vaults,
		Pagination: pageRes,
	}, nil
}

// TODO: Implement IPFS query functionality - connects to IPFS nodes and retrieves status information
// Should integrate with internal/ipfs package for IPFS client operations
// Query IPFS node health, connectivity, and peer information
// Return storage statistics and pinned content summary
// Support multiple IPFS endpoints with failover capability

// IPFS implements types.QueryServer.
func (k Querier) IPFS(
	goCtx context.Context,
	req *types.QueryIPFSRequest,
) (*types.QueryIPFSResponse, error) {
	// Check if IPFS client is available
	if k.ipfsClient == nil {
		k.Logger().Debug("IPFS client not available")
		return &types.QueryIPFSResponse{
			Status: &types.IPFSStatus{
				PeerId:   "",
				PeerName: "unavailable",
				PeerType: "ipfs",
				Version:  "unknown",
			},
		}, nil
	}

	// Get node status using the new NodeStatus method
	nodeStatus, err := k.ipfsClient.NodeStatus()
	if err != nil {
		k.Logger().Error("Failed to get IPFS node status", "error", err)
		return &types.QueryIPFSResponse{
			Status: &types.IPFSStatus{
				PeerId:   "",
				PeerName: "connection_failed",
				PeerType: "ipfs",
				Version:  "unknown",
			},
		}, nil
	}

	// Convert from internal NodeStatus to types.IPFSStatus
	status := &types.IPFSStatus{
		PeerId:   nodeStatus.PeerID,
		PeerName: "kubo-node",
		PeerType: nodeStatus.PeerType,
		Version:  nodeStatus.Version,
	}

	// Log successful status retrieval for debugging
	k.Logger().Debug("IPFS node status retrieved successfully",
		"peer_id", status.PeerId,
		"version", status.Version,
		"peer_type", status.PeerType,
		"connected_peers", nodeStatus.ConnectedPeers,
	)

	return &types.QueryIPFSResponse{
		Status: status,
	}, nil
}

// TODO: Implement CID query functionality - retrieves data from IPFS using content identifiers
// Should validate CID format and check if content exists in IPFS
// Retrieve content metadata without downloading full data
// Support different CID versions and hash algorithms
// Include content size, availability, and pin status

// CID implements types.QueryServer.
func (k Querier) CID(
	goCtx context.Context,
	req *types.QueryCIDRequest,
) (*types.QueryCIDResponse, error) {
	// Validate input
	if req == nil || req.Cid == "" {
		return &types.QueryCIDResponse{
			StatusCode: 400, // Bad Request
			Data:       nil,
		}, nil
	}

	// Validate CID format using go-cid library
	_, err := cid.Parse(req.Cid)
	if err != nil {
		k.Logger().Debug("Invalid CID format", "cid", req.Cid, "error", err)
		return &types.QueryCIDResponse{
			StatusCode: 400, // Bad Request
			Data:       nil,
		}, nil
	}

	// Get IPFS client with connectivity check
	ipfsClient, err := k.GetIPFSClient()
	if err != nil {
		k.Logger().Error("IPFS client not available", "error", err)
		return &types.QueryCIDResponse{
			StatusCode: 500, // Internal Server Error
			Data:       nil,
		}, nil
	}

	// Check if content exists in IPFS
	exists, err := ipfsClient.Exists(req.Cid)
	if err != nil {
		k.Logger().Error("Error checking CID existence", "cid", req.Cid, "error", err)
		return &types.QueryCIDResponse{
			StatusCode: 500, // Internal Server Error
			Data:       nil,
		}, nil
	}

	if !exists {
		k.Logger().Debug("CID not found in IPFS", "cid", req.Cid)
		return &types.QueryCIDResponse{
			StatusCode: 404, // Not Found
			Data:       nil,
		}, nil
	}

	// Retrieve content from IPFS
	data, err := ipfsClient.Get(req.Cid)
	if err != nil {
		k.Logger().Error("Error retrieving content from IPFS", "cid", req.Cid, "error", err)
		return &types.QueryCIDResponse{
			StatusCode: 500, // Internal Server Error
			Data:       nil,
		}, nil
	}

	k.Logger().Debug("Successfully retrieved content from IPFS",
		"cid", req.Cid,
		"size", len(data))

	return &types.QueryCIDResponse{
		StatusCode: 200, // Success
		Data:       data,
	}, nil
}

// EncryptedRecord queries a specific encrypted record with automatic decryption
func (k Querier) EncryptedRecord(
	c context.Context,
	req *types.QueryEncryptedRecordRequest,
) (*types.QueryEncryptedRecordResponse, error) {
	if req == nil {
		return nil, types.ErrRequestCannotBeNil
	}

	if req.Target == "" {
		return nil, types.ErrTargetDIDEmpty
	}

	if req.RecordId == "" {
		return nil, types.ErrRecordIDEmpty
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Get the record first
	record, err := k.OrmDB.DWNRecordTable().Get(ctx, req.RecordId)
	if err != nil {
		if ormerrors.IsNotFound(err) {
			return nil, errors.Wrapf(types.ErrRecordNotFound, "record %s not found", req.RecordId)
		}
		return nil, errors.Wrap(err, "failed to get record")
	}

	// Verify the record belongs to the target DWN
	if record.Target != req.Target {
		return nil, errors.Wrap(sdkerrors.ErrUnauthorized, "record does not belong to target DWN")
	}

	rec := types.ConvertAPIRecordToType(record)

	// Check if the record has encryption metadata
	if record.EncryptionMetadata == nil {
		return &types.QueryEncryptedRecordResponse{
			Record:       &rec,
			WasDecrypted: false,
		}, nil
	}

	// If return_encrypted is true, return the encrypted record without decryption
	if req.ReturnEncrypted {
		return &types.QueryEncryptedRecordResponse{
			Record:             &rec,
			EncryptionMetadata: types.ConvertAPIEncryptionMetadataToType(record.EncryptionMetadata),
			WasDecrypted:       false,
		}, nil
	}

	// Attempt to decrypt the record data if we have encryption subkeeper
	if k.encryptionSubkeeper != nil && len(rec.Data) > 0 {
		decryptedData, err := k.encryptionSubkeeper.DecryptWithConsensusKey(
			c,
			rec.Data,
			types.ConvertAPIEncryptionMetadataToType(record.EncryptionMetadata),
		)
		if err != nil {
			k.Logger().Error("Failed to decrypt record data",
				"record_id", req.RecordId,
				"error", err,
			)
			// Return the encrypted record if decryption fails
			return &types.QueryEncryptedRecordResponse{
				Record: &rec,
				EncryptionMetadata: types.ConvertAPIEncryptionMetadataToType(
					record.EncryptionMetadata,
				),
				WasDecrypted: false,
			}, nil
		}

		// Update record with decrypted data
		rec.Data = decryptedData
		return &types.QueryEncryptedRecordResponse{
			Record:             &rec,
			EncryptionMetadata: types.ConvertAPIEncryptionMetadataToType(record.EncryptionMetadata),
			WasDecrypted:       true,
		}, nil
	}

	// No encryption subkeeper or no encrypted data
	return &types.QueryEncryptedRecordResponse{
		Record:       &rec,
		WasDecrypted: false,
	}, nil
}

// EncryptionStatus queries current encryption key state and version
func (k Querier) EncryptionStatus(
	c context.Context,
	req *types.QueryEncryptionStatusRequest,
) (*types.QueryEncryptionStatusResponse, error) {
	if req == nil {
		return nil, types.ErrRequestCannotBeNil
	}

	// Default response with minimal information
	response := &types.QueryEncryptionStatusResponse{
		CurrentKeyVersion:     0,
		ValidatorSet:          []string{},
		SingleNodeMode:        true,
		LastRotation:          0,
		NextRotation:          0,
		TotalEncryptedRecords: 0,
	}

	// If we have encryption subkeeper, get detailed status
	if k.encryptionSubkeeper != nil {
		// Get current key version
		response.CurrentKeyVersion = k.encryptionSubkeeper.GetCurrentKeyVersion(c)

		// Check if single node mode
		response.SingleNodeMode = k.encryptionSubkeeper.isSingleNodeMode(c)

		// Get validator set
		validators, err := k.encryptionSubkeeper.getActiveValidators(c)
		if err == nil {
			validatorAddrs := make([]string, len(validators))
			for i, v := range validators {
				validatorAddrs[i] = fmt.Sprintf("%v", v)
			}
			response.ValidatorSet = validatorAddrs
		}

		// Get stored key state for rotation timestamps
		keyState, keyErr := k.encryptionSubkeeper.getStoredKeyState(c)
		if keyErr == nil {
			response.LastRotation = keyState.LastRotation
			response.NextRotation = keyState.NextRotation
		}

		// Get encryption statistics from the encryption subkeeper
		stats, statsErr := k.encryptionSubkeeper.GetEncryptionStats(c)
		if statsErr == nil {
			// Safely convert int64 to uint64
			if stats.TotalEncryptedRecords < 0 {
				response.TotalEncryptedRecords = 0
			} else {
				response.TotalEncryptedRecords = uint64(stats.TotalEncryptedRecords)
			}
		}
	}

	k.Logger().Debug("Retrieved encryption status",
		"key_version", response.CurrentKeyVersion,
		"single_node_mode", response.SingleNodeMode,
		"validator_count", len(response.ValidatorSet),
	)

	return response, nil
}

// VRFContributions lists VRF contributions for current consensus round
func (k Querier) VRFContributions(
	c context.Context,
	req *types.QueryVRFContributionsRequest,
) (*types.QueryVRFContributionsResponse, error) {
	if req == nil {
		return nil, types.ErrRequestCannotBeNil
	}

	ctx := sdk.UnwrapSDKContext(c)

	// Query VRF contributions from the database using filters
	var indexKey apiv1.VRFContributionIndexKey

	switch {
	case req.ValidatorAddress != "" && req.BlockHeight > 0:
		// Filter by both validator address and block height
		indexKey = apiv1.VRFContributionValidatorAddressBlockHeightIndexKey{}.
			WithValidatorAddressBlockHeight(req.ValidatorAddress, req.BlockHeight)
	case req.ValidatorAddress != "":
		// Filter by validator address only
		indexKey = apiv1.VRFContributionValidatorAddressBlockHeightIndexKey{}.
			WithValidatorAddress(req.ValidatorAddress)
	case req.BlockHeight > 0:
		// Filter by block height only
		indexKey = apiv1.VRFContributionBlockHeightIndexKey{}.
			WithBlockHeight(req.BlockHeight)
	default:
		// No filters, list all contributions
		indexKey = apiv1.VRFContributionPrimaryKey{}
	}

	// Query with pagination
	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &query.PageRequest{Limit: 100}
	}

	contributions := []types.VRFContribution{}

	iter, err := k.OrmDB.VRFContributionTable().List(c, indexKey)
	if err != nil {
		return nil, fmt.Errorf("failed to list VRF contributions: %w", err)
	}
	defer iter.Close()

	count := uint64(0)
	offset := pageReq.Offset
	limit := pageReq.Limit

	for iter.Next() {
		contrib, iterErr := iter.Value()
		if iterErr != nil {
			continue
		}

		count++

		// Handle pagination
		if count <= offset {
			continue
		}

		if uint64(len(contributions)) >= limit {
			break
		}

		// Convert from API type to types
		contributions = append(contributions, types.VRFContribution{
			ValidatorAddress: contrib.ValidatorAddress,
			Randomness:       contrib.Randomness,
			Proof:            contrib.Proof,
			BlockHeight:      contrib.BlockHeight,
			Timestamp:        contrib.Timestamp,
		})
	}

	// Get current consensus round information from database
	blockHeight := ctx.BlockHeight()
	var roundNumber uint64
	if blockHeight > 0 {
		// Safe conversion: blockHeight is positive int64, division result fits in uint64
		roundNumber = uint64(blockHeight) / 100
	}

	var currentRound *types.VRFConsensusRound

	// Try to get stored consensus round from database
	storedRound, roundErr := k.OrmDB.VRFConsensusRoundTable().Get(c, roundNumber)
	if roundErr == nil {
		// Convert from API type
		currentRound = &types.VRFConsensusRound{
			RoundNumber:           storedRound.RoundNumber,
			RequiredContributions: storedRound.RequiredContributions,
			ReceivedContributions: storedRound.ReceivedContributions,
			Status:                storedRound.Status,
			ExpiryHeight:          storedRound.ExpiryHeight,
		}
	} else {
		// Create new round information if not found in database
		// Safely convert contribution count to uint32
		var receivedContributions uint32
		contributionCount := len(contributions)
		switch {
		case contributionCount > 4294967295: // Max uint32
			receivedContributions = 4294967295
		case contributionCount < 0:
			receivedContributions = 0
		default:
			receivedContributions = uint32(contributionCount)
		}

		currentRound = &types.VRFConsensusRound{
			RoundNumber:           roundNumber,
			RequiredContributions: 1,
			ReceivedContributions: receivedContributions,
			Status:                "waiting_for_contributions",
			ExpiryHeight:          ctx.BlockHeight() + 100,
		}

		// Calculate required contributions based on active validators
		if k.encryptionSubkeeper != nil {
			validators, validatorErr := k.encryptionSubkeeper.getActiveValidators(c)
			if validatorErr == nil {
				validatorCount := len(validators)
				if validatorCount > 1 {
					// Byzantine fault tolerance: need 2/3 + 1 contributions
					bftThreshold := (validatorCount * 2 / 3) + 1
					if bftThreshold > 0 && bftThreshold <= int(^uint32(0)) {
						currentRound.RequiredContributions = uint32(bftThreshold)
					}
				}

				// Update status based on single node mode
				if k.encryptionSubkeeper.isSingleNodeMode(c) {
					currentRound.Status = "single_node_mode"
					currentRound.ReceivedContributions = 1
				} else if currentRound.ReceivedContributions >= currentRound.RequiredContributions {
					currentRound.Status = "completed"
				}
			}
		}
	}

	pageRes := &query.PageResponse{
		Total: uint64(len(contributions)),
	}

	k.Logger().Debug("Retrieved VRF contributions",
		"contributions_count", len(contributions),
		"round_number", currentRound.RoundNumber,
		"required_contributions", currentRound.RequiredContributions,
	)

	return &types.QueryVRFContributionsResponse{
		Contributions: contributions,
		CurrentRound:  currentRound,
		Pagination:    pageRes,
	}, nil
}
