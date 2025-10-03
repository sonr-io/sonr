package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	apiv1 "github.com/sonr-io/sonr/api/svc/v1"
	"github.com/sonr-io/sonr/x/svc/types"
)

// SetServiceOIDCConfig stores the OIDC configuration for a service
func (k Keeper) SetServiceOIDCConfig(ctx context.Context, config *types.ServiceOIDCConfig) error {
	// Validate service exists
	service, err := k.OrmDB.ServiceTable().Get(ctx, config.ServiceId)
	if err != nil {
		return fmt.Errorf("service not found: %s", config.ServiceId)
	}

	// Verify domain matches issuer
	if !k.validateIssuerDomain(config.Issuer, service.Domain) {
		return fmt.Errorf("issuer must match verified domain: %s", service.Domain)
	}

	// Convert to API type and store
	apiConfig := convertTypesOIDCConfigToAPI(config)
	return k.OrmDB.ServiceOIDCConfigTable().Save(ctx, apiConfig)
}

// GetServiceOIDCConfig retrieves the OIDC configuration for a service
func (k Keeper) GetServiceOIDCConfig(ctx context.Context, serviceID string) (*types.ServiceOIDCConfig, error) {
	apiConfig, err := k.OrmDB.ServiceOIDCConfigTable().Get(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	return convertAPIOIDCConfigToTypes(apiConfig), nil
}

// SetServiceJWKS stores the JWKS for a service
func (k Keeper) SetServiceJWKS(ctx context.Context, jwks *types.ServiceJWKS) error {
	// Validate service exists
	_, err := k.OrmDB.ServiceTable().Get(ctx, jwks.ServiceId)
	if err != nil {
		return fmt.Errorf("service not found: %s", jwks.ServiceId)
	}

	// Convert to API type and store
	apiJWKS := convertTypesJWKSToAPI(jwks)
	return k.OrmDB.ServiceJWKSTable().Save(ctx, apiJWKS)
}

// GetServiceJWKS retrieves the JWKS for a service
func (k Keeper) GetServiceJWKS(ctx context.Context, serviceID string) (*types.ServiceJWKS, error) {
	apiJWKS, err := k.OrmDB.ServiceJWKSTable().Get(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	return convertAPIJWKSToTypes(apiJWKS), nil
}

// validateIssuerDomain ensures the issuer URL matches the verified domain
func (k Keeper) validateIssuerDomain(issuer, domain string) bool {
	// Simple validation: issuer should contain the domain
	// In production, parse URL and validate hostname
	return true // Placeholder for now
}

// CreateDefaultOIDCConfig creates a default OIDC configuration for a service
func (k Keeper) CreateDefaultOIDCConfig(ctx context.Context, serviceID string, domain string) (*types.ServiceOIDCConfig, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	config := &types.ServiceOIDCConfig{
		ServiceId:             serviceID,
		Issuer:                fmt.Sprintf("https://%s", domain),
		AuthorizationEndpoint: fmt.Sprintf("https://%s/oauth/authorize", domain),
		TokenEndpoint:         fmt.Sprintf("https://%s/oauth/token", domain),
		JwksUri:               fmt.Sprintf("https://%s/.well-known/jwks.json", domain),
		UserinfoEndpoint:      fmt.Sprintf("https://%s/oauth/userinfo", domain),
		ScopesSupported: []string{
			"openid",
			"profile",
			"email",
			"offline_access",
		},
		ResponseTypesSupported: []string{
			"code",
			"token",
			"id_token",
			"code token",
			"code id_token",
			"token id_token",
			"code token id_token",
		},
		GrantTypesSupported: []string{
			"authorization_code",
			"implicit",
			"refresh_token",
			"client_credentials",
		},
		IdTokenSigningAlgValuesSupported: []string{
			"RS256",
			"ES256",
		},
		SubjectTypesSupported: []string{
			"public",
			"pairwise",
		},
		TokenEndpointAuthMethodsSupported: []string{
			"client_secret_basic",
			"client_secret_post",
			"none",
		},
		ClaimsSupported: []string{
			"sub",
			"iss",
			"aud",
			"exp",
			"iat",
			"nonce",
			"email",
			"email_verified",
			"name",
			"preferred_username",
			"picture",
			"did",
			"wallet_address",
		},
		ResponseModesSupported: []string{
			"query",
			"fragment",
			"form_post",
		},
		Metadata: map[string]string{
			"service_id": serviceID,
			"blockchain": "sonr",
			"chain_id":   sdkCtx.ChainID(),
		},
		CreatedAt: sdkCtx.BlockTime().Unix(),
		UpdatedAt: sdkCtx.BlockTime().Unix(),
	}

	// Store the config
	err := k.SetServiceOIDCConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// CreateDefaultJWKS creates a default JWKS for a service
func (k Keeper) CreateDefaultJWKS(ctx context.Context, serviceID string) (*types.ServiceJWKS, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// For now, create an empty JWKS
	// In production, this would generate or retrieve actual keys
	jwks := &types.ServiceJWKS{
		ServiceId: serviceID,
		Keys:      []*types.JWK{},
		RotatedAt: sdkCtx.BlockTime().Unix(),
	}

	// Store the JWKS
	err := k.SetServiceJWKS(ctx, jwks)
	if err != nil {
		return nil, err
	}

	return jwks, nil
}

// Conversion functions between API and Types

func convertTypesOIDCConfigToAPI(config *types.ServiceOIDCConfig) *apiv1.ServiceOIDCConfig {
	return &apiv1.ServiceOIDCConfig{
		ServiceId:                         config.ServiceId,
		Issuer:                            config.Issuer,
		AuthorizationEndpoint:             config.AuthorizationEndpoint,
		TokenEndpoint:                     config.TokenEndpoint,
		JwksUri:                           config.JwksUri,
		UserinfoEndpoint:                  config.UserinfoEndpoint,
		ScopesSupported:                   config.ScopesSupported,
		ResponseTypesSupported:            config.ResponseTypesSupported,
		GrantTypesSupported:               config.GrantTypesSupported,
		IdTokenSigningAlgValuesSupported:  config.IdTokenSigningAlgValuesSupported,
		SubjectTypesSupported:             config.SubjectTypesSupported,
		TokenEndpointAuthMethodsSupported: config.TokenEndpointAuthMethodsSupported,
		ClaimsSupported:                   config.ClaimsSupported,
		ResponseModesSupported:            config.ResponseModesSupported,
		Metadata:                          config.Metadata,
		CreatedAt:                         config.CreatedAt,
		UpdatedAt:                         config.UpdatedAt,
	}
}

func convertAPIOIDCConfigToTypes(config *apiv1.ServiceOIDCConfig) *types.ServiceOIDCConfig {
	return &types.ServiceOIDCConfig{
		ServiceId:                         config.ServiceId,
		Issuer:                            config.Issuer,
		AuthorizationEndpoint:             config.AuthorizationEndpoint,
		TokenEndpoint:                     config.TokenEndpoint,
		JwksUri:                           config.JwksUri,
		UserinfoEndpoint:                  config.UserinfoEndpoint,
		ScopesSupported:                   config.ScopesSupported,
		ResponseTypesSupported:            config.ResponseTypesSupported,
		GrantTypesSupported:               config.GrantTypesSupported,
		IdTokenSigningAlgValuesSupported:  config.IdTokenSigningAlgValuesSupported,
		SubjectTypesSupported:             config.SubjectTypesSupported,
		TokenEndpointAuthMethodsSupported: config.TokenEndpointAuthMethodsSupported,
		ClaimsSupported:                   config.ClaimsSupported,
		ResponseModesSupported:            config.ResponseModesSupported,
		Metadata:                          config.Metadata,
		CreatedAt:                         config.CreatedAt,
		UpdatedAt:                         config.UpdatedAt,
	}
}

func convertTypesJWKSToAPI(jwks *types.ServiceJWKS) *apiv1.ServiceJWKS {
	apiKeys := make([]*apiv1.JWK, len(jwks.Keys))
	for i, key := range jwks.Keys {
		apiKeys[i] = convertTypesJWKToAPI(key)
	}

	return &apiv1.ServiceJWKS{
		ServiceId: jwks.ServiceId,
		Keys:      apiKeys,
		RotatedAt: jwks.RotatedAt,
	}
}

func convertAPIJWKSToTypes(jwks *apiv1.ServiceJWKS) *types.ServiceJWKS {
	typesKeys := make([]*types.JWK, len(jwks.Keys))
	for i, key := range jwks.Keys {
		typesKeys[i] = convertAPIJWKToTypes(key)
	}

	return &types.ServiceJWKS{
		ServiceId: jwks.ServiceId,
		Keys:      typesKeys,
		RotatedAt: jwks.RotatedAt,
	}
}

func convertTypesJWKToAPI(key *types.JWK) *apiv1.JWK {
	return &apiv1.JWK{
		Kty: key.Kty,
		Use: key.Use,
		Kid: key.Kid,
		Alg: key.Alg,
		N:   key.N,
		E:   key.E,
		Crv: key.Crv,
		X:   key.X,
		Y:   key.Y,
	}
}

func convertAPIJWKToTypes(key *apiv1.JWK) *types.JWK {
	return &types.JWK{
		Kty: key.Kty,
		Use: key.Use,
		Kid: key.Kid,
		Alg: key.Alg,
		N:   key.N,
		E:   key.E,
		Crv: key.Crv,
		X:   key.X,
		Y:   key.Y,
	}
}
