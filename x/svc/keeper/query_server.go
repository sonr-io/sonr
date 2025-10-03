package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	apiv1 "github.com/sonr-io/sonr/api/svc/v1"
	"github.com/sonr-io/sonr/x/svc/types"
)

// convertV1ServiceToTypes converts a v1.Service to types.Service
func convertV1ServiceToTypes(v1Service *apiv1.Service) *types.Service {
	return &types.Service{
		Id:                v1Service.Id,
		Domain:            v1Service.Domain,
		Owner:             v1Service.Owner,
		RootCapabilityCid: v1Service.RootCapabilityCid,
		Permissions:       v1Service.Permissions,
		Status:            types.ServiceStatus(v1Service.Status),
		CreatedAt:         v1Service.CreatedAt,
		UpdatedAt:         v1Service.UpdatedAt,
	}
}

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

// DomainVerification implements types.QueryServer.
func (k Querier) DomainVerification(
	goCtx context.Context,
	req *types.QueryDomainVerificationRequest,
) (*types.QueryDomainVerificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req.Domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	// Get domain verification from ORM
	verification, err := k.Keeper.OrmDB.DomainVerificationTable().Get(ctx, req.Domain)
	if err != nil {
		return nil, fmt.Errorf("domain verification not found: %w", err)
	}

	// Convert v1.DomainVerification to types.DomainVerification
	typesVerification := &types.DomainVerification{
		Domain:            verification.Domain,
		Owner:             verification.Owner,
		VerificationToken: verification.VerificationToken,
		Status:            types.DomainVerificationStatus(verification.Status),
		ExpiresAt:         verification.ExpiresAt,
	}

	return &types.QueryDomainVerificationResponse{
		DomainVerification: typesVerification,
	}, nil
}

// Service implements types.QueryServer.
func (k Querier) Service(
	goCtx context.Context,
	req *types.QueryServiceRequest,
) (*types.QueryServiceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req.ServiceId == "" {
		return nil, fmt.Errorf("service_id cannot be empty")
	}

	// Get service from ORM
	service, err := k.Keeper.OrmDB.ServiceTable().Get(ctx, req.ServiceId)
	if err != nil {
		return nil, fmt.Errorf("service not found: %w", err)
	}

	return &types.QueryServiceResponse{
		Service: convertV1ServiceToTypes(service),
	}, nil
}

// ServicesByOwner implements types.QueryServer.
func (k Querier) ServicesByOwner(
	goCtx context.Context,
	req *types.QueryServicesByOwnerRequest,
) (*types.QueryServicesByOwnerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req.Owner == "" {
		return nil, fmt.Errorf("owner cannot be empty")
	}

	// Create index key for owner
	ownerKey := apiv1.ServiceOwnerIndexKey{}.WithOwner(req.Owner)

	// List services by owner
	iter, err := k.Keeper.OrmDB.ServiceTable().List(ctx, ownerKey)
	if err != nil {
		return nil, fmt.Errorf("failed to list services by owner: %w", err)
	}
	defer iter.Close()

	var services []*types.Service
	for iter.Next() {
		service, err := iter.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to get service value: %w", err)
		}
		services = append(services, convertV1ServiceToTypes(service))
	}

	return &types.QueryServicesByOwnerResponse{
		Services: services,
	}, nil
}

// ServicesByDomain implements types.QueryServer.
func (k Querier) ServicesByDomain(
	goCtx context.Context,
	req *types.QueryServicesByDomainRequest,
) (*types.QueryServicesByDomainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req.Domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	// Create index key for domain
	domainKey := apiv1.ServiceDomainIndexKey{}.WithDomain(req.Domain)

	// List services by domain
	iter, err := k.Keeper.OrmDB.ServiceTable().List(ctx, domainKey)
	if err != nil {
		return nil, fmt.Errorf("failed to list services by domain: %w", err)
	}
	defer iter.Close()

	var services []*types.Service
	for iter.Next() {
		service, err := iter.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to get service value: %w", err)
		}
		services = append(services, convertV1ServiceToTypes(service))
	}

	return &types.QueryServicesByDomainResponse{
		Services: services,
	}, nil
}

// ServiceOIDCDiscovery implements types.QueryServer.
func (k Querier) ServiceOIDCDiscovery(goCtx context.Context, req *types.QueryServiceOIDCDiscoveryRequest) (*types.QueryServiceOIDCDiscoveryResponse, error) {
	if req == nil || req.ServiceId == "" {
		return nil, types.ErrInvalidServiceID
	}

	// Get service to verify it exists and is active
	service, err := k.Keeper.OrmDB.ServiceTable().Get(goCtx, req.ServiceId)
	if err != nil {
		return nil, types.ErrServiceNotFound
	}

	if service.Status != apiv1.ServiceStatus_SERVICE_STATUS_ACTIVE {
		return nil, types.ErrServiceNotActive
	}

	// Get OIDC config
	config, err := k.Keeper.GetServiceOIDCConfig(goCtx, req.ServiceId)
	if err != nil {
		// If no config exists, create default based on verified domain
		config, err = k.Keeper.CreateDefaultOIDCConfig(goCtx, req.ServiceId, service.Domain)
		if err != nil {
			return nil, err
		}
	}

	// Build OIDC discovery response according to spec
	return &types.QueryServiceOIDCDiscoveryResponse{
		Issuer:                            config.Issuer,
		AuthorizationEndpoint:             config.AuthorizationEndpoint,
		TokenEndpoint:                     config.TokenEndpoint,
		JwksUri:                           config.JwksUri,
		UserinfoEndpoint:                  config.UserinfoEndpoint,
		RegistrationEndpoint:              fmt.Sprintf("https://%s/oauth/register", service.Domain),
		ScopesSupported:                   config.ScopesSupported,
		ResponseTypesSupported:            config.ResponseTypesSupported,
		GrantTypesSupported:               config.GrantTypesSupported,
		IdTokenSigningAlgValuesSupported:  config.IdTokenSigningAlgValuesSupported,
		SubjectTypesSupported:             config.SubjectTypesSupported,
		TokenEndpointAuthMethodsSupported: config.TokenEndpointAuthMethodsSupported,
		ClaimsSupported:                   config.ClaimsSupported,
		ResponseModesSupported:            config.ResponseModesSupported,
		ServiceDocumentation:              fmt.Sprintf("https://%s/docs", service.Domain),
		UiLocalesSupported:                []string{"en-US"},
		ClaimsLocalesSupported:            []string{"en-US"},
		RequestParameterSupported:         true,
		RequestUriParameterSupported:      true,
		RequireRequestUriRegistration:     false,
		OpPolicyUri:                       fmt.Sprintf("https://%s/policy", service.Domain),
		OpTosUri:                          fmt.Sprintf("https://%s/terms", service.Domain),
	}, nil
}

// ServiceOIDCJWKS implements types.QueryServer.
func (k Querier) ServiceOIDCJWKS(goCtx context.Context, req *types.QueryServiceOIDCJWKSRequest) (*types.QueryServiceOIDCJWKSResponse, error) {
	if req == nil || req.ServiceId == "" {
		return nil, types.ErrInvalidServiceID
	}

	// Get service to verify it exists
	service, err := k.Keeper.OrmDB.ServiceTable().Get(goCtx, req.ServiceId)
	if err != nil {
		return nil, types.ErrServiceNotFound
	}

	if service.Status != apiv1.ServiceStatus_SERVICE_STATUS_ACTIVE {
		return nil, types.ErrServiceNotActive
	}

	// Get JWKS
	jwks, err := k.Keeper.GetServiceJWKS(goCtx, req.ServiceId)
	if err != nil {
		// If no JWKS exists, create default
		jwks, err = k.Keeper.CreateDefaultJWKS(goCtx, req.ServiceId)
		if err != nil {
			return nil, err
		}
	}

	// Return JWKS response
	return &types.QueryServiceOIDCJWKSResponse{
		Keys: jwks.Keys,
	}, nil
}

// ServiceOIDCMetadata implements types.QueryServer.
func (k Querier) ServiceOIDCMetadata(goCtx context.Context, req *types.QueryServiceOIDCMetadataRequest) (*types.QueryServiceOIDCMetadataResponse, error) {
	if req == nil || req.ServiceId == "" {
		return nil, types.ErrInvalidServiceID
	}

	// Get service
	service, err := k.Keeper.OrmDB.ServiceTable().Get(goCtx, req.ServiceId)
	if err != nil {
		return nil, types.ErrServiceNotFound
	}

	// Get OIDC config
	config, err := k.Keeper.GetServiceOIDCConfig(goCtx, req.ServiceId)
	if err != nil {
		// If no config exists, create default based on verified domain
		config, err = k.Keeper.CreateDefaultOIDCConfig(goCtx, req.ServiceId, service.Domain)
		if err != nil {
			return nil, err
		}
	}

	// Convert service status
	var serviceStatus types.ServiceStatus
	switch service.Status {
	case apiv1.ServiceStatus_SERVICE_STATUS_ACTIVE:
		serviceStatus = types.ServiceStatus_SERVICE_STATUS_ACTIVE
	case apiv1.ServiceStatus_SERVICE_STATUS_SUSPENDED:
		serviceStatus = types.ServiceStatus_SERVICE_STATUS_SUSPENDED
	case apiv1.ServiceStatus_SERVICE_STATUS_REVOKED:
		serviceStatus = types.ServiceStatus_SERVICE_STATUS_REVOKED
	default:
		serviceStatus = types.ServiceStatus_SERVICE_STATUS_ACTIVE
	}

	// Build metadata response
	return &types.QueryServiceOIDCMetadataResponse{
		Config:         config,
		VerifiedDomain: service.Domain,
		ServiceStatus:  serviceStatus,
		Metadata: map[string]string{
			"service_id":    service.Id,
			"owner":         service.Owner,
			"created_at":    fmt.Sprintf("%d", service.CreatedAt),
			"updated_at":    fmt.Sprintf("%d", service.UpdatedAt),
			"ucan_root_cid": service.RootCapabilityCid,
			"oidc_enabled":  "true",
		},
	}, nil
}
