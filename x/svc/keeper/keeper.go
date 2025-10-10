package keeper

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/orm/model/ormdb"

	apiv1 "github.com/sonr-io/sonr/api/svc/v1"
	"github.com/sonr-io/crypto/keys"
	"github.com/sonr-io/crypto/ucan"
	"github.com/sonr-io/sonr/x/svc/types"
)

type Keeper struct {
	cdc codec.BinaryCodec

	logger log.Logger

	// state management
	Schema collections.Schema
	Params collections.Item[types.Params]
	OrmDB  apiv1.StateStore

	// dependencies
	didKeeper types.DIDKeeper

	// UCAN functionality
	ucanVerifier        *ucan.Verifier
	permissionValidator *PermissionValidator

	authority string
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	logger log.Logger,
	authority string,
	didKeeper types.DIDKeeper,
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

	// Create UCAN verifier with DID resolver
	didResolver := &DIDKeeperResolver{didKeeper: didKeeper}
	ucanVerifier := ucan.NewVerifier(didResolver)

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

		didKeeper:    didKeeper,
		ucanVerifier: ucanVerifier,
		authority:    authority,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	// Initialize UCAN permission validator (after keeper is fully constructed)
	k.permissionValidator = NewPermissionValidator(k)

	return k
}

// GetPermissionValidator returns the UCAN permission validator
func (k Keeper) GetPermissionValidator() *PermissionValidator {
	return k.permissionValidator
}

func (k Keeper) Logger() log.Logger {
	return k.logger
}

// InitGenesis initializes the module's state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *types.GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	// Set parameters
	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}

	// Import capabilities
	for _, capability := range data.Capabilities {
		// Convert to types.ServiceCapability for storage
		cap := &types.ServiceCapability{
			CapabilityId: capability.CapabilityId,
			ServiceId:    capability.ServiceId,
			Domain:       capability.Domain,
			Abilities:    capability.Abilities,
			Owner:        capability.Owner,
			CreatedAt:    capability.CreatedAt,
			ExpiresAt:    capability.ExpiresAt,
			Revoked:      capability.Revoked,
		}
		if err := k.StoreCapability(ctx, cap); err != nil {
			return fmt.Errorf("failed to import capability %s: %w", capability.CapabilityId, err)
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

	// Export all capabilities
	var capabilities []types.ServiceCapability

	// Iterate through all capabilities in the ORM
	iter, err := k.OrmDB.ServiceCapabilityTable().List(ctx, apiv1.ServiceCapabilityPrimaryKey{})
	if err != nil {
		panic(fmt.Errorf("failed to list capabilities for export: %w", err))
	}
	defer iter.Close()

	for iter.Next() {
		apiCap, err := iter.Value()
		if err != nil {
			panic(fmt.Errorf("failed to get capability during export: %w", err))
		}

		// Convert from API type to types
		cap := types.ServiceCapability{
			CapabilityId: apiCap.CapabilityId,
			ServiceId:    apiCap.ServiceId,
			Domain:       apiCap.Domain,
			Abilities:    apiCap.Abilities,
			Owner:        apiCap.Owner,
			CreatedAt:    apiCap.CreatedAt,
			ExpiresAt:    apiCap.ExpiresAt,
			Revoked:      apiCap.Revoked,
		}
		capabilities = append(capabilities, cap)
	}

	return &types.GenesisState{
		Params:       params,
		Capabilities: capabilities,
	}
}

// VerifyServiceRegistration verifies service registration and domain ownership
func (k Keeper) VerifyServiceRegistration(
	ctx context.Context,
	serviceID string,
	domain string,
) (bool, error) {
	if serviceID == "" {
		return false, types.ErrInvalidServiceID
	}

	if domain == "" {
		return false, types.ErrDomainNotVerified
	}

	// Check if the service exists
	service, err := k.OrmDB.ServiceTable().Get(ctx, serviceID)
	if err != nil {
		return false, types.ErrInvalidServiceID
	}

	// Verify the service belongs to the specified domain
	if service.Domain != domain {
		return false, types.ErrDomainNotVerified
	}

	// Check if the domain is verified
	if !k.IsVerifiedDomain(ctx, domain) {
		return false, types.ErrDomainNotVerified
	}

	// Check if the service is active
	if service.Status != apiv1.ServiceStatus_SERVICE_STATUS_ACTIVE {
		return false, types.ErrInvalidServiceID
	}

	return true, nil
}

// GetService gets service by ID
func (k Keeper) GetService(ctx context.Context, serviceID string) (*types.Service, error) {
	if serviceID == "" {
		return nil, types.ErrInvalidServiceID
	}

	// Get service from ORM
	service, err := k.OrmDB.ServiceTable().Get(ctx, serviceID)
	if err != nil {
		return nil, types.ErrInvalidServiceID
	}

	// Convert v1.Service to types.Service
	return &types.Service{
		Id:                service.Id,
		Domain:            service.Domain,
		Owner:             service.Owner,
		RootCapabilityCid: service.RootCapabilityCid,
		Permissions:       service.Permissions,
		Status:            types.ServiceStatus(service.Status),
		CreatedAt:         service.CreatedAt,
		UpdatedAt:         service.UpdatedAt,
	}, nil
}

// IsDomainVerified checks if domain is verified
func (k Keeper) IsDomainVerified(ctx context.Context, domain string, owner string) (bool, error) {
	if domain == "" {
		return false, types.ErrDomainNotVerified
	}

	// Get domain verification record
	verification, err := k.OrmDB.DomainVerificationTable().Get(ctx, domain)
	if err != nil {
		return false, types.ErrDomainNotVerified
	}

	// Check if the domain is verified
	if verification.Status != apiv1.DomainVerificationStatus_DOMAIN_VERIFICATION_STATUS_VERIFIED {
		return false, nil
	}

	// Check if the owner matches (if provided)
	if owner != "" && verification.Owner != owner {
		return false, nil
	}

	// Check if the verification hasn't expired
	if k.isDomainVerificationExpired(verification) {
		return false, nil
	}

	return true, nil
}

// GetServicesByDomain gets services by domain
func (k Keeper) GetServicesByDomain(ctx context.Context, domain string) ([]types.Service, error) {
	if domain == "" {
		return nil, types.ErrDomainNotVerified
	}

	// Create index key for domain
	domainKey := apiv1.ServiceDomainIndexKey{}.WithDomain(domain)

	// List services by domain
	iter, err := k.OrmDB.ServiceTable().List(ctx, domainKey)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var services []types.Service
	for iter.Next() {
		service, err := iter.Value()
		if err != nil {
			return nil, err
		}

		// Convert v1.Service to types.Service
		services = append(services, types.Service{
			Id:                service.Id,
			Domain:            service.Domain,
			Owner:             service.Owner,
			RootCapabilityCid: service.RootCapabilityCid,
			Permissions:       service.Permissions,
			Status:            types.ServiceStatus(service.Status),
			CreatedAt:         service.CreatedAt,
			UpdatedAt:         service.UpdatedAt,
		})
	}

	return services, nil
}

// VerifyOrigin validates a relying party origin for WebAuthn operations
func (k Keeper) VerifyOrigin(ctx context.Context, origin string) error {
	// Allow localhost origins for development
	if isLocalhostOrigin(origin) {
		return nil
	}

	// Extract domain from origin
	domain := extractDomainFromOrigin(origin)
	if domain == "" {
		return fmt.Errorf("could not extract domain from origin: %s", origin)
	}

	// Check if domain is verified
	if !k.IsVerifiedDomain(ctx, domain) {
		return fmt.Errorf("domain not verified: %s", domain)
	}

	// Check if there are active services for this domain
	services, err := k.GetServicesByDomain(ctx, domain)
	if err != nil {
		return fmt.Errorf("failed to get services for domain %s: %w", domain, err)
	}

	if len(services) == 0 {
		return fmt.Errorf("no services registered for domain: %s", domain)
	}

	// Check if at least one service is active
	hasActiveService := false
	for _, service := range services {
		if service.Status == types.ServiceStatus_SERVICE_STATUS_ACTIVE {
			hasActiveService = true
			break
		}
	}

	if !hasActiveService {
		return fmt.Errorf("no active services found for domain: %s", domain)
	}

	return nil
}

// isLocalhostOrigin checks if the origin is a localhost origin
func isLocalhostOrigin(origin string) bool {
	localhostPatterns := []string{
		"http://localhost",
		"https://localhost",
		"http://127.0.0.1",
		"https://127.0.0.1",
		"http://[::1]",
		"https://[::1]",
	}

	for _, pattern := range localhostPatterns {
		if strings.HasPrefix(origin, pattern) {
			return true
		}
	}
	return false
}

// extractDomainFromOrigin extracts the domain from an origin URL
func extractDomainFromOrigin(origin string) string {
	// Remove protocol
	domain := strings.TrimPrefix(origin, "https://")
	domain = strings.TrimPrefix(domain, "http://")

	// Remove port if present
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove path if present
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	return domain
}

// ValidateServiceOwnerDID verifies that the service owner has a valid DID document
func (k Keeper) ValidateServiceOwnerDID(ctx context.Context, ownerDID string) error {
	if ownerDID == "" {
		return types.ErrInvalidOwnerDID
	}

	// Get the DID document
	didDoc, err := k.didKeeper.GetDIDDocument(ctx, ownerDID)
	if err != nil {
		return types.ErrInvalidOwnerDID
	}

	// Check if the DID document exists
	if didDoc == nil {
		return types.ErrInvalidOwnerDID
	}

	// Check if the DID document is deactivated
	if didDoc.Deactivated {
		return types.ErrInvalidOwnerDID
	}

	// Additional validation: check if the DID document has valid verification methods
	if len(didDoc.VerificationMethod) == 0 {
		return types.ErrInvalidOwnerDID
	}

	return nil
}

// DIDKeeperResolver adapts the DID keeper to implement the UCAN DIDResolver interface
type DIDKeeperResolver struct {
	didKeeper types.DIDKeeper
}

// ResolveDIDKey resolves a DID string using the DID keeper
func (r *DIDKeeperResolver) ResolveDIDKey(ctx context.Context, did string) (keys.DID, error) {
	// Get the DID document from the keeper
	didDoc, err := r.didKeeper.GetDIDDocument(ctx, did)
	if err != nil {
		return keys.DID{}, err
	}

	if didDoc == nil {
		return keys.DID{}, types.ErrInvalidOwnerDID
	}

	// Parse the DID string into a keys.DID
	// This assumes the DID keeper can provide the public key information
	return keys.Parse(did)
}
