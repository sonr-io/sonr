// Package keeper provides the DWN module keeper implementation.
package keeper

import (
	"context"
	"fmt"
	"slices"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/orm/model/ormdb"

	apiv1 "github.com/sonr-io/sonr/api/dwn/v1"
	sonrcontext "github.com/sonr-io/sonr/app/context"
	"github.com/sonr-io/sonr/crypto/mpc"
	"github.com/sonr-io/sonr/crypto/vrf"
	"github.com/sonr-io/sonr/types/ipfs"
	didtypes "github.com/sonr-io/sonr/x/did/types"
	"github.com/sonr-io/sonr/x/dwn/types"
)

type Keeper struct {
	cdc codec.BinaryCodec

	logger log.Logger

	// state management
	Schema collections.Schema
	Params collections.Item[types.Params]
	OrmDB  apiv1.StateStore

	// SDK keepers for wallet operations
	accountKeeper  authkeeper.AccountKeeper
	bankKeeper     bankkeeper.Keeper
	feegrantKeeper feegrantkeeper.Keeper
	stakingKeeper  *stakingkeeper.Keeper
	didKeeper      types.DIDKeeper
	serviceKeeper  types.ServiceKeeper

	// client context for transaction building
	clientCtx client.Context

	// vault client for enclave operations
	ipfsClient ipfs.IPFSClient
	// vaultClient vault.VaultClient

	// encryption subkeeper for consensus-based encryption
	encryptionSubkeeper *EncryptionSubkeeper

	// UCAN permission validator for DWN operations
	permissionValidator *PermissionValidator

	authority string

	vrfPrivateKey vrf.PrivateKey
	vrfPublicKey  vrf.PublicKey
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	logger log.Logger,
	authority string,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	feegrantKeeper feegrantkeeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper,
	didKeeper types.DIDKeeper,
	serviceKeeper types.ServiceKeeper,
	clientCtx client.Context,
) Keeper {
	logger = logger.With(log.ModuleKey, "x/"+types.ModuleName)

	sb := collections.NewSchemaBuilder(storeService)

	if authority == "" {
		authority = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	}

	db, err := ormdb.NewModuleDB(
		&types.ORMModuleSchema,
		ormdb.ModuleDBOptions{KVStoreService: storeService},
	)
	if err != nil {
		panic(err)
	}

	store, err := apiv1.NewStateStore(db)
	if err != nil {
		panic(err)
	}

	k := Keeper{
		cdc:    cdc,
		logger: logger,

		Params: collections.NewItem(
			sb,
			types.ParamsKey,
			"params",
			codec.CollValue[types.Params](cdc),
		),
		OrmDB: store,

		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
		feegrantKeeper: feegrantKeeper,
		didKeeper:      didKeeper,
		stakingKeeper:  stakingKeeper,
		serviceKeeper:  serviceKeeper,

		clientCtx: clientCtx,
		authority: authority,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	// Load VRF keys from global context if available
	if errB := k.loadVRFKeysFromContext(); errB != nil {
		logger.Warn("Failed to load VRF keys from context", "error", err)
		// Continue without VRF keys - they can be loaded later
	}

	// Initialize IPFS client
	ipfsClient, err := ipfs.GetClient()
	if err != nil {
		logger.Error(
			"Failed to initialize IPFS client",
			"error",
			types.ErrIPFSClientNotAvailable,
		)
		// Continue without IPFS client - this allows the keeper to still function
		// but IPFS operations will fail gracefully
	} else {
		k.ipfsClient = ipfsClient
	}

	// Initialize encryption subkeeper
	k.encryptionSubkeeper = NewEncryptionSubkeeper(&k)

	// Initialize UCAN permission validator
	k.permissionValidator = NewPermissionValidator(didKeeper)

	return k
}

func (k Keeper) Logger() log.Logger {
	return k.logger
}

// GetEncryptionSubkeeper returns the encryption subkeeper
func (k Keeper) GetEncryptionSubkeeper() *EncryptionSubkeeper {
	return k.encryptionSubkeeper
}

// GetPermissionValidator returns the UCAN permission validator
func (k Keeper) GetPermissionValidator() *PermissionValidator {
	return k.permissionValidator
}

// CheckAndPerformKeyRotation checks if key rotation is due and performs it if needed
func (k Keeper) CheckAndPerformKeyRotation(ctx context.Context) error {
	return k.encryptionSubkeeper.CheckAndPerformRotation(ctx)
}

// ShouldEncryptRecord determines if a record should be encrypted based on protocol/schema
func (k Keeper) ShouldEncryptRecord(ctx context.Context, protocol, schema string) (bool, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return false, err
	}

	// Check if encryption is globally enabled
	if !params.EncryptionEnabled {
		return false, nil
	}

	// Check if protocol requires encryption
	if slices.Contains(params.EncryptedProtocols, protocol) {
		return true, nil
	}

	// Check if schema requires encryption
	for _, encryptedSchema := range params.EncryptedSchemas {
		if schema == encryptedSchema ||
			(schema != "" && encryptedSchema != "" &&
				len(schema) >= len(encryptedSchema) &&
				schema[:len(encryptedSchema)] == encryptedSchema) {
			return true, nil
		}
	}

	return false, nil
}

// InitGenesis initializes the module's state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}

	// Import DWN records
	for _, record := range data.Records {
		// Convert to API type
		apiRecord := &apiv1.DWNRecord{
			RecordId:            record.RecordId,
			Target:              record.Target,
			Authorization:       record.Authorization,
			Data:                record.Data,
			Protocol:            record.Protocol,
			ProtocolPath:        record.ProtocolPath,
			Schema:              record.Schema,
			ParentId:            record.ParentId,
			Published:           record.Published,
			Attestation:         record.Attestation,
			Encryption:          record.Encryption,
			KeyDerivationScheme: record.KeyDerivationScheme,
			CreatedAt:           record.CreatedAt,
			UpdatedAt:           record.UpdatedAt,
			CreatedHeight:       record.CreatedHeight,
		}
		if record.Descriptor_ != nil {
			apiRecord.Descriptor_ = &apiv1.DWNMessageDescriptor{
				InterfaceName:    record.Descriptor_.InterfaceName,
				Method:           record.Descriptor_.Method,
				MessageTimestamp: record.Descriptor_.MessageTimestamp,
				DataCid:          record.Descriptor_.DataCid,
				DataSize:         record.Descriptor_.DataSize,
				DataFormat:       record.Descriptor_.DataFormat,
			}
		}
		if err := k.OrmDB.DWNRecordTable().Insert(ctx, apiRecord); err != nil {
			return err
		}
	}

	// Import DWN protocols
	for _, protocol := range data.Protocols {
		// Convert to API type
		apiProtocol := &apiv1.DWNProtocol{
			Target:        protocol.Target,
			ProtocolUri:   protocol.ProtocolUri,
			Definition:    protocol.Definition,
			Published:     protocol.Published,
			CreatedAt:     protocol.CreatedAt,
			CreatedHeight: protocol.CreatedHeight,
		}
		if err := k.OrmDB.DWNProtocolTable().Insert(ctx, apiProtocol); err != nil {
			return err
		}
	}

	// Import DWN permissions
	for _, permission := range data.Permissions {
		// Convert to API type
		apiPermission := &apiv1.DWNPermission{
			PermissionId:  permission.PermissionId,
			Grantor:       permission.Grantor,
			Grantee:       permission.Grantee,
			Target:        permission.Target,
			InterfaceName: permission.InterfaceName,
			Method:        permission.Method,
			Protocol:      permission.Protocol,
			RecordId:      permission.RecordId,
			Conditions:    permission.Conditions,
			ExpiresAt:     permission.ExpiresAt,
			CreatedAt:     permission.CreatedAt,
			Revoked:       permission.Revoked,
			CreatedHeight: permission.CreatedHeight,
		}
		if err := k.OrmDB.DWNPermissionTable().Insert(ctx, apiPermission); err != nil {
			return err
		}
	}

	// Import vault states
	for _, vault := range data.Vaults {
		// Convert to API type
		apiVault := &apiv1.VaultState{
			VaultId:       vault.VaultId,
			Owner:         vault.Owner,
			PublicKey:     vault.PublicKey,
			CreatedAt:     vault.CreatedAt,
			LastRefreshed: vault.LastRefreshed,
			CreatedHeight: vault.CreatedHeight,
		}
		if vault.EnclaveData != nil {
			apiVault.EnclaveData = &apiv1.EnclaveData{
				PrivateData: vault.EnclaveData.PrivateData,
				PublicKey:   vault.EnclaveData.PublicKey,
				EnclaveId:   vault.EnclaveData.EnclaveId,
				Version:     vault.EnclaveData.Version,
			}
		}
		if err := k.OrmDB.VaultStateTable().Insert(ctx, apiVault); err != nil {
			return err
		}
	}

	return nil
}

// ExportGenesis exports the module's state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	genesis := &types.GenesisState{
		Params:      params,
		Records:     []types.DWNRecord{},
		Protocols:   []types.DWNProtocol{},
		Permissions: []types.DWNPermission{},
		Vaults:      []types.VaultState{},
	}

	// Export DWN records
	recordIter, err := k.OrmDB.DWNRecordTable().List(ctx, apiv1.DWNRecordPrimaryKey{})
	if err == nil {
		defer recordIter.Close()
		for recordIter.Next() {
			record, errB := recordIter.Value()
			if errB == nil {
				genesis.Records = append(genesis.Records, types.ConvertAPIRecordToType(record))
			}
		}
	}

	// Export DWN protocols
	protocolIter, err := k.OrmDB.DWNProtocolTable().List(ctx, apiv1.DWNProtocolPrimaryKey{})
	if err == nil {
		defer protocolIter.Close()
		for protocolIter.Next() {
			protocol, errB := protocolIter.Value()
			if errB == nil {
				genesis.Protocols = append(
					genesis.Protocols,
					types.ConvertAPIProtocolToType(protocol),
				)
			}
		}
	}

	// Export DWN permissions
	permissionIter, err := k.OrmDB.DWNPermissionTable().List(ctx, apiv1.DWNPermissionPrimaryKey{})
	if err == nil {
		defer permissionIter.Close()
		for permissionIter.Next() {
			permission, errB := permissionIter.Value()
			if errB == nil {
				genesis.Permissions = append(
					genesis.Permissions,
					types.ConvertAPIPermissionToType(permission),
				)
			}
		}
	}

	// Export vault states
	vaultIter, err := k.OrmDB.VaultStateTable().List(ctx, apiv1.VaultStatePrimaryKey{})
	if err == nil {
		defer vaultIter.Close()
		for vaultIter.Next() {
			vault, err := vaultIter.Value()
			if err == nil {
				genesis.Vaults = append(genesis.Vaults, types.ConvertAPIVaultToType(vault))
			}
		}
	}

	return genesis
}

// Vault operation methods that delegate to VaultKeeper

// ValidateServiceForProtocol validates that a service is registered for a protocol operation
func (k Keeper) ValidateServiceForProtocol(ctx context.Context, target, serviceID string) error {
	if serviceID == "" {
		// Allow operations without explicit service registration for backward compatibility
		return nil
	}

	// Extract domain from target (DID format: did:web:domain)
	var domain string
	if len(target) > 8 && target[:8] == "did:web:" {
		domain = target[8:]
	} else {
		k.Logger().Debug("Target is not a DID:web, skipping service verification", "target", target)
		return nil
	}

	// Verify service registration
	verified, err := k.serviceKeeper.VerifyServiceRegistration(ctx, serviceID, domain)
	if err != nil {
		return errors.Wrap(err, "failed to verify service registration")
	}

	if !verified {
		return errors.Wrapf(
			types.ErrServiceNotVerified,
			"service %s not verified for domain %s",
			serviceID,
			domain,
		)
	}

	return nil
}

// GetFeeGrantKeeper returns the underlying fee grant keeper for direct access if needed.
// This method provides access to the fee grant keeper for advanced operations.
func (k Keeper) GetFeeGrantKeeper() feegrantkeeper.Keeper {
	return k.feegrantKeeper
}

// loadVRFKeysFromContext loads VRF keys from the global SonrContext
func (k *Keeper) loadVRFKeysFromContext() error {
	ctx := sonrcontext.GetGlobalSonrContext()
	if ctx == nil {
		return fmt.Errorf("global SonrContext not available")
	}

	if !ctx.IsInitialized() {
		return fmt.Errorf("SonrContext not initialized")
	}

	privateKey, err := ctx.GetVRFPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to get VRF private key from context: %w", err)
	}

	publicKey, err := ctx.GetVRFPublicKey()
	if err != nil {
		return fmt.Errorf("failed to get VRF public key from context: %w", err)
	}

	k.vrfPrivateKey = privateKey
	k.vrfPublicKey = publicKey

	k.logger.Info("VRF keys loaded from SonrContext",
		"private_key_size", len(k.vrfPrivateKey),
		"public_key_size", len(k.vrfPublicKey),
	)
	return nil
}

// GetVRFKeys returns the loaded VRF keypair
func (k Keeper) GetVRFKeys() (vrf.PrivateKey, vrf.PublicKey, error) {
	if len(k.vrfPrivateKey) == 0 || len(k.vrfPublicKey) == 0 {
		// Try to load from context if not already loaded
		if err := k.loadVRFKeysFromContext(); err != nil {
			return nil, nil, fmt.Errorf("VRF keys not loaded: %w\n"+
				"To fix this issue:\n"+
				"  1. Run 'snrd init <moniker>' to generate VRF keys for your node\n"+
				"  2. Or disable encryption in DWN module params if not needed\n"+
				"  3. For existing nodes, VRF keys should be in ~/.sonr/vrf_secret.key", err)
		}
	}

	return k.vrfPrivateKey, k.vrfPublicKey, nil
}

// ComputeVRF generates VRF output using the keeper's loaded private key
func (k Keeper) ComputeVRF(input []byte) ([]byte, error) {
	privateKey, _, err := k.GetVRFKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to get VRF keys: %w", err)
	}

	if len(input) == 0 {
		return nil, fmt.Errorf("VRF input cannot be empty")
	}

	return privateKey.Compute(input), nil
}

// CreateVaultForDID creates a vault for a given DID using the WebAssembly enclave plugin
func (k Keeper) CreateVaultForDID(
	ctx context.Context,
	data *mpc.EnclaveData,
) (*didtypes.CreateVaultResponse, error) {
	// Input validation
	vaultState, err := k.AddEnclaveDataToIPFS(ctx, data)
	if err != nil {
		return nil, err
	}

	// Insert the vault state into the database
	if err := k.OrmDB.VaultStateTable().Insert(ctx, vaultState); err != nil {
		k.logger.Error("Failed to store vault state",
			"vault_id", vaultState.VaultId,
			"error", err,
		)
		return nil, fmt.Errorf("failed to store vault state: %w", err)
	}

	// Emit typed event
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	event := &types.EventVaultCreated{
		VaultId:     vaultState.VaultId,
		Owner:       vaultState.Owner,
		PublicKey:   string(vaultState.PublicKey),
		BlockHeight: uint64(sdkCtx.BlockHeight()),
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(event); err != nil {
		k.logger.With("error", err).Error("Failed to emit EventVaultCreated")
	}

	return &didtypes.CreateVaultResponse{}, nil
}
