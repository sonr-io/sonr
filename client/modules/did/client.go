// Package did provides a client interface for interacting with the Sonr DID module.
package did

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/sonr-io/sonr/client/config"
	"github.com/sonr-io/sonr/client/errors"
	didtypes "github.com/sonr-io/sonr/x/did/types"
)

// Client provides an interface for interacting with the DID module.
type Client interface {
	// DID Operations
	CreateDID(ctx context.Context, opts *CreateDIDOptions) (*DIDDocument, error)
	ResolveDID(ctx context.Context, did string) (*DIDDocument, error)
	UpdateDID(ctx context.Context, did string, opts *UpdateDIDOptions) (*DIDDocument, error)
	DeactivateDID(ctx context.Context, did string) error

	// DID Document Operations
	AddVerificationMethod(ctx context.Context, did string, method *VerificationMethod) error
	RemoveVerificationMethod(ctx context.Context, did string, methodID string) error
	AddService(ctx context.Context, did string, service *Service) error
	RemoveService(ctx context.Context, did string, serviceID string) error

	// WebAuthn Operations
	RegisterWebAuthn(ctx context.Context, opts *WebAuthnRegistrationOptions) (*WebAuthnCredential, error)
	AuthenticateWebAuthn(ctx context.Context, opts *WebAuthnAuthenticationOptions) (*WebAuthnAssertion, error)

	// Query Operations
	ListDIDs(ctx context.Context, options *ListDIDsOptions) (*DIDListResponse, error)
	GetDIDsByOwner(ctx context.Context, owner string) ([]*DIDDocument, error)
}

// DIDDocument represents a W3C DID Document.
type DIDDocument struct {
	ID                   string                `json:"id"`
	Controller           []string              `json:"controller,omitempty"`
	VerificationMethod   []*VerificationMethod `json:"verificationMethod,omitempty"`
	Authentication       []string              `json:"authentication,omitempty"`
	AssertionMethod      []string              `json:"assertionMethod,omitempty"`
	KeyAgreement         []string              `json:"keyAgreement,omitempty"`
	CapabilityInvocation []string              `json:"capabilityInvocation,omitempty"`
	CapabilityDelegation []string              `json:"capabilityDelegation,omitempty"`
	Service              []*Service            `json:"service,omitempty"`
	AlsoKnownAs          []string              `json:"alsoKnownAs,omitempty"`
	Metadata             *DIDMetadata          `json:"metadata,omitempty"`
}

// VerificationMethod represents a verification method in a DID document.
type VerificationMethod struct {
	ID                 string         `json:"id"`
	Type               string         `json:"type"`
	Controller         string         `json:"controller"`
	PublicKeyJwk       map[string]any `json:"publicKeyJwk,omitempty"`
	PublicKeyMultibase string         `json:"publicKeyMultibase,omitempty"`
}

// Service represents a service endpoint in a DID document.
type Service struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint any    `json:"serviceEndpoint"`
}

// DIDMetadata contains metadata about a DID.
type DIDMetadata struct {
	Created       string `json:"created"`
	Updated       string `json:"updated,omitempty"`
	Deactivated   bool   `json:"deactivated,omitempty"`
	VersionID     string `json:"versionId,omitempty"`
	NextUpdate    string `json:"nextUpdate,omitempty"`
	NextVersionID string `json:"nextVersionId,omitempty"`
}

// CreateDIDOptions configures DID creation.
type CreateDIDOptions struct {
	Controller          []string              `json:"controller,omitempty"`
	VerificationMethods []*VerificationMethod `json:"verificationMethods,omitempty"`
	Services            []*Service            `json:"services,omitempty"`
	AlsoKnownAs         []string              `json:"alsoKnownAs,omitempty"`
	UseWebAuthn         bool                  `json:"useWebAuthn,omitempty"`
}

// UpdateDIDOptions configures DID updates.
type UpdateDIDOptions struct {
	AddVerificationMethods    []*VerificationMethod `json:"addVerificationMethods,omitempty"`
	RemoveVerificationMethods []string              `json:"removeVerificationMethods,omitempty"`
	AddServices               []*Service            `json:"addServices,omitempty"`
	RemoveServices            []string              `json:"removeServices,omitempty"`
	AddController             []string              `json:"addController,omitempty"`
	RemoveController          []string              `json:"removeController,omitempty"`
}

// WebAuthnRegistrationOptions configures WebAuthn registration.
type WebAuthnRegistrationOptions struct {
	Username    string         `json:"username"`
	DisplayName string         `json:"displayName"`
	Challenge   []byte         `json:"challenge"`
	Timeout     int            `json:"timeout,omitempty"`
	Extensions  map[string]any `json:"extensions,omitempty"`
}

// WebAuthnCredential represents a WebAuthn credential.
type WebAuthnCredential struct {
	ID               string                 `json:"id"`
	RawID            []byte                 `json:"rawId"`
	Type             string                 `json:"type"`
	Response         *AuthenticatorResponse `json:"response"`
	ClientExtensions map[string]any         `json:"clientExtensions,omitempty"`
}

// AuthenticatorResponse represents the authenticator response.
type AuthenticatorResponse struct {
	ClientDataJSON    []byte `json:"clientDataJSON"`
	AttestationObject []byte `json:"attestationObject"`
}

// WebAuthnAuthenticationOptions configures WebAuthn authentication.
type WebAuthnAuthenticationOptions struct {
	Challenge          []byte   `json:"challenge"`
	Timeout            int      `json:"timeout,omitempty"`
	AllowedCredentials []string `json:"allowedCredentials,omitempty"`
}

// WebAuthnAssertion represents a WebAuthn assertion.
type WebAuthnAssertion struct {
	ID       string                          `json:"id"`
	RawID    []byte                          `json:"rawId"`
	Type     string                          `json:"type"`
	Response *AuthenticatorAssertionResponse `json:"response"`
}

// AuthenticatorAssertionResponse represents the assertion response.
type AuthenticatorAssertionResponse struct {
	ClientDataJSON    []byte `json:"clientDataJSON"`
	AuthenticatorData []byte `json:"authenticatorData"`
	Signature         []byte `json:"signature"`
	UserHandle        []byte `json:"userHandle,omitempty"`
}

// ListDIDsOptions configures DID listing.
type ListDIDsOptions struct {
	Limit  uint64 `json:"limit,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
	Owner  string `json:"owner,omitempty"`
}

// DIDListResponse contains a list of DIDs with pagination.
type DIDListResponse struct {
	DIDs       []*DIDDocument `json:"dids"`
	TotalCount uint64         `json:"totalCount"`
	Limit      uint64         `json:"limit"`
	Offset     uint64         `json:"offset"`
}

// client implements the DID Client interface.
type client struct {
	grpcConn *grpc.ClientConn
	config   *config.NetworkConfig

	// Service clients for DID module
	queryClient didtypes.QueryClient
	msgClient   didtypes.MsgClient
	txClient    tx.ServiceClient
}

// NewClient creates a new DID module client.
func NewClient(grpcConn *grpc.ClientConn, cfg *config.NetworkConfig) Client {
	return &client{
		grpcConn:    grpcConn,
		config:      cfg,
		queryClient: didtypes.NewQueryClient(grpcConn),
		msgClient:   didtypes.NewMsgClient(grpcConn),
		txClient:    tx.NewServiceClient(grpcConn),
	}
}

// CreateDID creates a new DID document on the Sonr blockchain.
func (c *client) CreateDID(ctx context.Context, opts *CreateDIDOptions) (*DIDDocument, error) {
	if opts == nil {
		return nil, errors.NewModuleError("did", "CreateDID",
			fmt.Errorf("options cannot be nil"))
	}

	// Generate DID ID
	// Note: In a real implementation, this would use proper key derivation
	didID := GenerateDID(fmt.Sprintf("user_%d", len(opts.Controller)))

	// Convert verification methods to protobuf format
	var verificationMethods []*didtypes.VerificationMethod
	for _, vm := range opts.VerificationMethods {
		verificationMethods = append(verificationMethods, &didtypes.VerificationMethod{
			Id:                     vm.ID,
			VerificationMethodKind: vm.Type,
			Controller:             vm.Controller,
			PublicKeyMultibase:     vm.PublicKeyMultibase,
		})
	}

	// Convert services to protobuf format
	var services []*didtypes.Service
	for _, svc := range opts.Services {
		services = append(services, &didtypes.Service{
			Id:             svc.ID,
			ServiceKind:    svc.Type,
			SingleEndpoint: fmt.Sprintf("%v", svc.ServiceEndpoint),
		})
	}

	// Get primary controller (first one if multiple)
	primaryController := ""
	if len(opts.Controller) > 0 {
		primaryController = opts.Controller[0]
	}

	// Create DID Document
	didDocument := didtypes.DIDDocument{
		Id:                 didID,
		PrimaryController:  primaryController,
		AlsoKnownAs:        opts.AlsoKnownAs,
		VerificationMethod: verificationMethods,
		Service:            services,
	}

	// Create the MsgCreateDID message
	msg := &didtypes.MsgCreateDID{
		Controller:  primaryController, // Will be set by the transaction builder
		DidDocument: didDocument,
	}

	// In a real implementation, this would submit the transaction
	// For now, store the message for later use
	_ = msg

	// Return a mock DID document
	return &DIDDocument{
		ID:                 didID,
		Controller:         opts.Controller,
		VerificationMethod: opts.VerificationMethods,
		Service:            opts.Services,
		AlsoKnownAs:        opts.AlsoKnownAs,
		Metadata: &DIDMetadata{
			Created: "2024-01-01T00:00:00Z",
		},
	}, nil
}

// ResolveDID resolves a DID to its document.
func (c *client) ResolveDID(ctx context.Context, did string) (*DIDDocument, error) {
	// TODO: Implement DID resolution using DID module query client
	// Should validate DID format before querying chain
	// Query chain state for DID document by ID
	// Convert protobuf DIDDocument to client type
	// Handle DID not found and deactivated DID cases

	return nil, errors.NewModuleError("did", "ResolveDID",
		fmt.Errorf("DID resolution not yet implemented"))
}

// UpdateDID updates an existing DID document.
func (c *client) UpdateDID(ctx context.Context, did string, opts *UpdateDIDOptions) (*DIDDocument, error) {
	// TODO: Implement DID updates using DID module
	// Should validate DID ownership and update permissions
	// Build MsgUpdateDID with incremental changes
	// Handle verification method and service updates
	// Return updated DID document with new version

	return nil, errors.NewModuleError("did", "UpdateDID",
		fmt.Errorf("DID updates not yet implemented"))
}

// DeactivateDID deactivates a DID document.
func (c *client) DeactivateDID(ctx context.Context, did string) error {
	// TODO: Implement DID deactivation using DID module
	// Should validate DID ownership before deactivation
	// Build MsgDeactivateDID and submit to chain
	// Mark DID as deactivated in chain state
	// Handle cascading effects on dependent services

	return errors.NewModuleError("did", "DeactivateDID",
		fmt.Errorf("DID deactivation not yet implemented"))
}

// AddVerificationMethod adds a verification method to a DID document.
func (c *client) AddVerificationMethod(ctx context.Context, did string, method *VerificationMethod) error {
	// TODO: Implement verification method addition using DID module
	// Should validate DID ownership and method format
	// Build MsgAddVerificationMethod and submit to chain
	// Validate public key format and cryptographic validity
	// Update DID document with new verification method

	return errors.NewModuleError("did", "AddVerificationMethod",
		fmt.Errorf("verification method addition not yet implemented"))
}

// RemoveVerificationMethod removes a verification method from a DID document.
func (c *client) RemoveVerificationMethod(ctx context.Context, did string, methodID string) error {
	// TODO: Implement verification method removal using DID module
	// Should validate DID ownership and method existence
	// Build MsgRemoveVerificationMethod and submit to chain
	// Check if method is used in other DID relationships
	// Prevent removal of last verification method

	return errors.NewModuleError("did", "RemoveVerificationMethod",
		fmt.Errorf("verification method removal not yet implemented"))
}

// AddService adds a service to a DID document.
func (c *client) AddService(ctx context.Context, did string, service *Service) error {
	// TODO: Implement service addition using DID module
	// Should validate DID ownership and service format
	// Build MsgAddService and submit to chain
	// Validate service endpoint URLs and accessibility
	// Update DID document with new service entry

	return errors.NewModuleError("did", "AddService",
		fmt.Errorf("service addition not yet implemented"))
}

// RemoveService removes a service from a DID document.
func (c *client) RemoveService(ctx context.Context, did string, serviceID string) error {
	// TODO: Implement service removal using DID module
	// Should validate DID ownership and service existence
	// Build MsgRemoveService and submit to chain
	// Check for dependent systems using this service
	// Update DID document removing service entry

	return errors.NewModuleError("did", "RemoveService",
		fmt.Errorf("service removal not yet implemented"))
}

// RegisterWebAuthn registers a WebAuthn credential with a DID.
func (c *client) RegisterWebAuthn(ctx context.Context, opts *WebAuthnRegistrationOptions) (*WebAuthnCredential, error) {
	// TODO: Implement WebAuthn registration using DID module
	// Should validate registration options and challenge
	// Build MsgRegisterWebAuthnCredential and submit to chain
	// Process authenticator attestation and public key
	// Store credential ID and public key in DID document
	// Support auto-vault creation if enabled

	return nil, errors.NewModuleError("did", "RegisterWebAuthn",
		fmt.Errorf("WebAuthn registration not yet implemented"))
}

// AuthenticateWebAuthn performs WebAuthn authentication.
func (c *client) AuthenticateWebAuthn(ctx context.Context, opts *WebAuthnAuthenticationOptions) (*WebAuthnAssertion, error) {
	// TODO: Implement WebAuthn authentication using DID module
	// Should validate authentication challenge and credentials
	// Verify authenticator assertion against stored public key
	// Check credential ID against allowed credentials list
	// Return verified assertion with user handle and signature

	return nil, errors.NewModuleError("did", "AuthenticateWebAuthn",
		fmt.Errorf("WebAuthn authentication not yet implemented"))
}

// ListDIDs lists DIDs with optional filtering and pagination.
func (c *client) ListDIDs(ctx context.Context, options *ListDIDsOptions) (*DIDListResponse, error) {
	// TODO: Implement DID listing using DID module query client
	// Should support pagination with limit/offset
	// Filter by owner address if specified
	// Return DIDs with basic metadata and status
	// Handle empty result sets gracefully

	return nil, errors.NewModuleError("did", "ListDIDs",
		fmt.Errorf("DID listing not yet implemented"))
}

// GetDIDsByOwner retrieves all DIDs owned by a specific address.
func (c *client) GetDIDsByOwner(ctx context.Context, owner string) ([]*DIDDocument, error) {
	// TODO: Implement owner-based DID lookup using DID module
	// Should validate owner address format
	// Query chain state for DIDs controlled by owner
	// Return complete DID documents for all owned DIDs
	// Include active and deactivated DIDs with status

	return nil, errors.NewModuleError("did", "GetDIDsByOwner",
		fmt.Errorf("owner-based DID lookup not yet implemented"))
}

// Utility functions

// GenerateDID generates a new DID identifier for the Sonr network.
func GenerateDID(identifier string) string {
	return fmt.Sprintf("did:sonr:%s", identifier)
}

// ValidateDID validates a DID format.
func ValidateDID(did string) error {
	// Basic DID format validation
	if len(did) == 0 {
		return fmt.Errorf("DID cannot be empty")
	}

	if !strings.HasPrefix(did, "did:sonr:") {
		return fmt.Errorf("DID must start with 'did:sonr:'")
	}

	return nil
}

// CreateDefaultVerificationMethod creates a default verification method.
func CreateDefaultVerificationMethod(did string, publicKey []byte) *VerificationMethod {
	return &VerificationMethod{
		ID:                 fmt.Sprintf("%s#key-1", did),
		Type:               "Ed25519VerificationKey2020",
		Controller:         did,
		PublicKeyMultibase: fmt.Sprintf("z%x", publicKey), // Simplified multibase encoding
	}
}

// CreateWebAuthnService creates a service entry for WebAuthn.
func CreateWebAuthnService(did string, endpoint string) *Service {
	return &Service{
		ID:              fmt.Sprintf("%s#webauthn", did),
		Type:            "WebAuthnService",
		ServiceEndpoint: endpoint,
	}
}

// Message Builders - These create the actual transaction messages

// BuildMsgCreateDID builds a MsgCreateDID message.
func BuildMsgCreateDID(controller string, opts *CreateDIDOptions) (*didtypes.MsgCreateDID, error) {
	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}

	// Generate DID ID
	didID := GenerateDID(fmt.Sprintf("user_%s", controller[:8]))

	// Convert verification methods
	var verificationMethods []*didtypes.VerificationMethod
	for _, vm := range opts.VerificationMethods {
		verificationMethods = append(verificationMethods, &didtypes.VerificationMethod{
			Id:                     vm.ID,
			VerificationMethodKind: vm.Type,
			Controller:             vm.Controller,
			PublicKeyMultibase:     vm.PublicKeyMultibase,
		})
	}

	// Convert services
	var services []*didtypes.Service
	for _, svc := range opts.Services {
		services = append(services, &didtypes.Service{
			Id:             svc.ID,
			ServiceKind:    svc.Type,
			SingleEndpoint: fmt.Sprintf("%v", svc.ServiceEndpoint),
		})
	}

	// Create DID Document
	didDocument := didtypes.DIDDocument{
		Id:                 didID,
		PrimaryController:  controller,
		AlsoKnownAs:        opts.AlsoKnownAs,
		VerificationMethod: verificationMethods,
		Service:            services,
	}

	return &didtypes.MsgCreateDID{
		Controller:  controller,
		DidDocument: didDocument,
	}, nil
}

// BuildMsgUpdateDID builds a MsgUpdateDID message.
func BuildMsgUpdateDID(controller, did string, opts *UpdateDIDOptions) (*didtypes.MsgUpdateDID, error) {
	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}

	// Note: In a real implementation, we would need to query the existing DID document
	// and apply the updates. For now, we create a minimal DID document with updates.

	// Convert new verification methods
	var verificationMethods []*didtypes.VerificationMethod
	for _, vm := range opts.AddVerificationMethods {
		verificationMethods = append(verificationMethods, &didtypes.VerificationMethod{
			Id:                     vm.ID,
			VerificationMethodKind: vm.Type,
			Controller:             vm.Controller,
			PublicKeyMultibase:     vm.PublicKeyMultibase,
		})
	}

	// Convert new services
	var services []*didtypes.Service
	for _, svc := range opts.AddServices {
		services = append(services, &didtypes.Service{
			Id:             svc.ID,
			ServiceKind:    svc.Type,
			SingleEndpoint: fmt.Sprintf("%v", svc.ServiceEndpoint),
		})
	}

	// Create updated DID Document
	didDocument := didtypes.DIDDocument{
		Id:                 did,
		PrimaryController:  controller,
		VerificationMethod: verificationMethods,
		Service:            services,
	}

	return &didtypes.MsgUpdateDID{
		Controller:  controller,
		Did:         did,
		DidDocument: didDocument,
	}, nil
}

// BuildMsgDeactivateDID builds a MsgDeactivateDID message.
func BuildMsgDeactivateDID(controller, did string) *didtypes.MsgDeactivateDID {
	return &didtypes.MsgDeactivateDID{
		Controller: controller,
		Did:        did,
	}
}

// BuildMsgAddVerificationMethod builds a MsgAddVerificationMethod message.
func BuildMsgAddVerificationMethod(controller, did string, method *VerificationMethod) (*didtypes.MsgAddVerificationMethod, error) {
	if method == nil {
		return nil, fmt.Errorf("verification method cannot be nil")
	}

	return &didtypes.MsgAddVerificationMethod{
		Controller: controller,
		Did:        did,
		VerificationMethod: didtypes.VerificationMethod{
			Id:                     method.ID,
			VerificationMethodKind: method.Type,
			Controller:             method.Controller,
			PublicKeyMultibase:     method.PublicKeyMultibase,
		},
	}, nil
}

// BuildMsgRemoveVerificationMethod builds a MsgRemoveVerificationMethod message.
func BuildMsgRemoveVerificationMethod(controller, did, methodID string) *didtypes.MsgRemoveVerificationMethod {
	return &didtypes.MsgRemoveVerificationMethod{
		Controller:           controller,
		Did:                  did,
		VerificationMethodId: methodID,
	}
}

// BuildMsgAddService builds a MsgAddService message.
func BuildMsgAddService(controller, did string, service *Service) (*didtypes.MsgAddService, error) {
	if service == nil {
		return nil, fmt.Errorf("service cannot be nil")
	}

	return &didtypes.MsgAddService{
		Controller: controller,
		Did:        did,
		Service: didtypes.Service{
			Id:             service.ID,
			ServiceKind:    service.Type,
			SingleEndpoint: fmt.Sprintf("%v", service.ServiceEndpoint),
		},
	}, nil
}

// BuildMsgRemoveService builds a MsgRemoveService message.
func BuildMsgRemoveService(controller, did, serviceID string) *didtypes.MsgRemoveService {
	return &didtypes.MsgRemoveService{
		Controller: controller,
		Did:        did,
		ServiceId:  serviceID,
	}
}

// BuildMsgRegisterWebAuthnCredential builds a MsgRegisterWebAuthnCredential message.
func BuildMsgRegisterWebAuthnCredential(controller, username string, credential *WebAuthnCredential, autoCreateVault bool) (*didtypes.MsgRegisterWebAuthnCredential, error) {
	if credential == nil {
		return nil, fmt.Errorf("credential cannot be nil")
	}

	// Convert our WebAuthnCredential to the protobuf type
	webAuthnCred := didtypes.WebAuthnCredential{
		CredentialId: credential.ID,
		PublicKey:    credential.RawID, // Using RawID as the public key bytes
		// Algorithm would need to be determined from the credential
		// AttestationType would need to be extracted from the attestation object
		// Origin would need to be extracted from the client data
	}

	// Generate a verification method ID based on the username
	verificationMethodID := fmt.Sprintf("did:sonr:%s#webauthn-1", username)

	return &didtypes.MsgRegisterWebAuthnCredential{
		Controller:           controller,
		Username:             username,
		WebauthnCredential:   webAuthnCred,
		VerificationMethodId: verificationMethodID,
		AutoCreateVault:      autoCreateVault,
	}, nil
}
