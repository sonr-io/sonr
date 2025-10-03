// Package svc provides a client interface for interacting with the Sonr SVC (Service) module.
package svc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/sonr-io/sonr/client/config"
	"github.com/sonr-io/sonr/client/errors"
	svctypes "github.com/sonr-io/sonr/x/svc/types"
)

// Client provides an interface for interacting with the SVC module.
type Client interface {
	// Service Operations
	RegisterService(ctx context.Context, opts *RegisterServiceOptions) (*Service, error)
	UpdateService(ctx context.Context, serviceID string, opts *UpdateServiceOptions) (*Service, error)
	DeregisterService(ctx context.Context, serviceID string) error
	GetService(ctx context.Context, serviceID string) (*Service, error)

	// Service Discovery
	DiscoverServices(ctx context.Context, query *ServiceQuery) (*ServiceDiscoveryResponse, error)
	ListServices(ctx context.Context, opts *ListServicesOptions) (*ServiceListResponse, error)
	SearchServices(ctx context.Context, searchTerm string) (*ServiceSearchResponse, error)

	// Domain Operations
	RegisterDomain(ctx context.Context, opts *RegisterDomainOptions) (*Domain, error)
	VerifyDomain(ctx context.Context, domain string) (*DomainVerification, error)
	GetDomain(ctx context.Context, domain string) (*Domain, error)
	ListDomains(ctx context.Context, opts *ListDomainsOptions) (*DomainListResponse, error)

	// Service Capabilities
	AddCapability(ctx context.Context, serviceID string, capability *Capability) error
	RemoveCapability(ctx context.Context, serviceID string, capabilityID string) error
	ListCapabilities(ctx context.Context, serviceID string) ([]*Capability, error)

	// Service Endpoints
	AddEndpoint(ctx context.Context, serviceID string, endpoint *Endpoint) error
	UpdateEndpoint(ctx context.Context, serviceID string, endpointID string, opts *UpdateEndpointOptions) error
	RemoveEndpoint(ctx context.Context, serviceID string, endpointID string) error

	// Service Health
	CheckServiceHealth(ctx context.Context, serviceID string) (*HealthStatus, error)
	UpdateServiceHealth(ctx context.Context, serviceID string, status *HealthStatus) error
}

// Service represents a registered service.
type Service struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	Owner        string         `json:"owner"`
	Domain       string         `json:"domain,omitempty"`
	Version      string         `json:"version"`
	Type         string         `json:"type"`
	Endpoints    []*Endpoint    `json:"endpoints"`
	Capabilities []*Capability  `json:"capabilities"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	HealthStatus *HealthStatus  `json:"health_status,omitempty"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at,omitempty"`
	Tags         []string       `json:"tags,omitempty"`
}

// RegisterServiceOptions configures service registration.
type RegisterServiceOptions struct {
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	Domain       string         `json:"domain,omitempty"`
	Version      string         `json:"version"`
	Type         string         `json:"type"`
	Endpoints    []*Endpoint    `json:"endpoints"`
	Capabilities []*Capability  `json:"capabilities,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Tags         []string       `json:"tags,omitempty"`
}

// UpdateServiceOptions configures service updates.
type UpdateServiceOptions struct {
	Description string         `json:"description,omitempty"`
	Version     string         `json:"version,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
}

// Endpoint represents a service endpoint.
type Endpoint struct {
	ID       string            `json:"id"`
	URL      string            `json:"url"`
	Type     string            `json:"type"` // REST, GraphQL, gRPC, WebSocket, etc.
	Method   string            `json:"method,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
	Enabled  bool              `json:"enabled"`
}

// UpdateEndpointOptions configures endpoint updates.
type UpdateEndpointOptions struct {
	URL      string            `json:"url,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
	Enabled  *bool             `json:"enabled,omitempty"`
}

// Capability represents a service capability.
type Capability struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Type        string         `json:"type"`
	Schema      map[string]any `json:"schema,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Enabled     bool           `json:"enabled"`
}

// ServiceQuery defines query parameters for service discovery.
type ServiceQuery struct {
	Type         string            `json:"type,omitempty"`
	Domain       string            `json:"domain,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	HealthStatus string            `json:"health_status,omitempty"`
}

// ServiceDiscoveryResponse contains service discovery results.
type ServiceDiscoveryResponse struct {
	Services   []*Service    `json:"services"`
	TotalCount uint64        `json:"total_count"`
	Query      *ServiceQuery `json:"query"`
}

// ListServicesOptions configures service listing.
type ListServicesOptions struct {
	Owner  string `json:"owner,omitempty"`
	Type   string `json:"type,omitempty"`
	Limit  uint64 `json:"limit,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
}

// ServiceListResponse contains a list of services.
type ServiceListResponse struct {
	Services   []*Service `json:"services"`
	TotalCount uint64     `json:"total_count"`
	Limit      uint64     `json:"limit"`
	Offset     uint64     `json:"offset"`
}

// ServiceSearchResponse contains service search results.
type ServiceSearchResponse struct {
	Services   []*Service `json:"services"`
	TotalCount uint64     `json:"total_count"`
	SearchTerm string     `json:"search_term"`
}

// Domain represents a registered domain.
type Domain struct {
	Name         string              `json:"name"`
	Owner        string              `json:"owner"`
	Verified     bool                `json:"verified"`
	Verification *DomainVerification `json:"verification,omitempty"`
	Metadata     map[string]any      `json:"metadata,omitempty"`
	RegisteredAt string              `json:"registered_at"`
	ExpiresAt    string              `json:"expires_at,omitempty"`
}

// RegisterDomainOptions configures domain registration.
type RegisterDomainOptions struct {
	Name      string         `json:"name"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	ExpiresAt string         `json:"expires_at,omitempty"`
}

// DomainVerification contains domain verification information.
type DomainVerification struct {
	Method     string `json:"method"`    // DNS, HTTP, File
	Token      string `json:"token"`     // Verification token
	Challenge  string `json:"challenge"` // Challenge string
	Verified   bool   `json:"verified"`
	VerifiedAt string `json:"verified_at,omitempty"`
	ExpiresAt  string `json:"expires_at,omitempty"`
}

// ListDomainsOptions configures domain listing.
type ListDomainsOptions struct {
	Owner    string `json:"owner,omitempty"`
	Verified *bool  `json:"verified,omitempty"`
	Limit    uint64 `json:"limit,omitempty"`
	Offset   uint64 `json:"offset,omitempty"`
}

// DomainListResponse contains a list of domains.
type DomainListResponse struct {
	Domains    []*Domain `json:"domains"`
	TotalCount uint64    `json:"total_count"`
	Limit      uint64    `json:"limit"`
	Offset     uint64    `json:"offset"`
}

// HealthStatus represents service health status.
type HealthStatus struct {
	Status      string         `json:"status"` // healthy, unhealthy, degraded, unknown
	LastChecked string         `json:"last_checked"`
	Message     string         `json:"message,omitempty"`
	Metrics     map[string]any `json:"metrics,omitempty"`
	Uptime      string         `json:"uptime,omitempty"`
}

// client implements the SVC Client interface.
type client struct {
	grpcConn *grpc.ClientConn
	config   *config.NetworkConfig

	// Service clients for SVC module
	queryClient svctypes.QueryClient
	msgClient   svctypes.MsgClient
	txClient    tx.ServiceClient
}

// NewClient creates a new SVC module client.
func NewClient(grpcConn *grpc.ClientConn, cfg *config.NetworkConfig) Client {
	return &client{
		grpcConn:    grpcConn,
		config:      cfg,
		queryClient: svctypes.NewQueryClient(grpcConn),
		msgClient:   svctypes.NewMsgClient(grpcConn),
		txClient:    tx.NewServiceClient(grpcConn),
	}
}

// RegisterService registers a new service.
func (c *client) RegisterService(ctx context.Context, opts *RegisterServiceOptions) (*Service, error) {
	// TODO: Implement service registration using SVC module
	// Should build MsgRegisterService with proper validation
	// Submit transaction to chain and wait for confirmation
	// Handle domain verification if domain is provided
	// Return complete service record with generated ID

	return nil, errors.NewModuleError("svc", "RegisterService",
		fmt.Errorf("service registration not yet implemented"))
}

// UpdateService updates an existing service.
func (c *client) UpdateService(ctx context.Context, serviceID string, opts *UpdateServiceOptions) (*Service, error) {
	// TODO: Implement service updates using SVC module
	// Should validate service ownership and permissions
	// Build MsgUpdateService with selective field updates
	// Preserve existing endpoints and capabilities unless modified
	// Return updated service record

	return nil, errors.NewModuleError("svc", "UpdateService",
		fmt.Errorf("service updates not yet implemented"))
}

// DeregisterService deregisters a service.
func (c *client) DeregisterService(ctx context.Context, serviceID string) error {
	// TODO: Implement service deregistration using SVC module
	// Should validate service ownership before deregistration
	// Build MsgDeregisterService and submit to chain
	// Clean up associated domain registrations and capabilities
	// Handle graceful shutdown of service endpoints

	return errors.NewModuleError("svc", "DeregisterService",
		fmt.Errorf("service deregistration not yet implemented"))
}

// GetService retrieves a service by ID.
func (c *client) GetService(ctx context.Context, serviceID string) (*Service, error) {
	// TODO: Implement service retrieval using SVC query client
	// Should query chain state for service record
	// Convert protobuf service to client Service type
	// Include current health status and endpoint information
	// Handle service not found errors gracefully

	return nil, errors.NewModuleError("svc", "GetService",
		fmt.Errorf("service retrieval not yet implemented"))
}

// DiscoverServices discovers services based on query criteria.
func (c *client) DiscoverServices(ctx context.Context, query *ServiceQuery) (*ServiceDiscoveryResponse, error) {
	// TODO: Implement service discovery using SVC module
	// Should support filtering by type, domain, tags, capabilities
	// Query chain state with proper pagination
	// Filter results by health status if specified
	// Return ranked results based on relevance

	return nil, errors.NewModuleError("svc", "DiscoverServices",
		fmt.Errorf("service discovery not yet implemented"))
}

// ListServices lists services with pagination.
func (c *client) ListServices(ctx context.Context, opts *ListServicesOptions) (*ServiceListResponse, error) {
	// TODO: Implement service listing using SVC module
	// Should support pagination with limit/offset
	// Filter by owner and service type if specified
	// Return services with basic metadata and status
	// Handle empty result sets gracefully

	return nil, errors.NewModuleError("svc", "ListServices",
		fmt.Errorf("service listing not yet implemented"))
}

// SearchServices searches for services by term.
func (c *client) SearchServices(ctx context.Context, searchTerm string) (*ServiceSearchResponse, error) {
	// TODO: Implement service search using SVC module
	// Should search service names, descriptions, and tags
	// Support fuzzy matching and relevance scoring
	// Query multiple fields with OR logic
	// Return results ranked by relevance

	return nil, errors.NewModuleError("svc", "SearchServices",
		fmt.Errorf("service search not yet implemented"))
}

// RegisterDomain registers a new domain.
func (c *client) RegisterDomain(ctx context.Context, opts *RegisterDomainOptions) (*Domain, error) {
	// TODO: Implement domain registration using SVC module
	// Should validate domain name format and availability
	// Build MsgInitiateDomainVerification and submit to chain
	// Generate verification challenge tokens
	// Return domain record with verification instructions

	return nil, errors.NewModuleError("svc", "RegisterDomain",
		fmt.Errorf("domain registration not yet implemented"))
}

// VerifyDomain verifies domain ownership.
func (c *client) VerifyDomain(ctx context.Context, domain string) (*DomainVerification, error) {
	// TODO: Implement domain verification using SVC module
	// Should check DNS records, HTTP endpoints, or file verification
	// Build MsgVerifyDomain and submit proof to chain
	// Update domain status to verified upon success
	// Handle verification failures with clear error messages

	return nil, errors.NewModuleError("svc", "VerifyDomain",
		fmt.Errorf("domain verification not yet implemented"))
}

// GetDomain retrieves domain information.
func (c *client) GetDomain(ctx context.Context, domain string) (*Domain, error) {
	// TODO: Implement domain retrieval using SVC query client
	// Should query chain state for domain registration
	// Include verification status and expiration information
	// Convert protobuf domain to client Domain type
	// Handle domain not found cases

	return nil, errors.NewModuleError("svc", "GetDomain",
		fmt.Errorf("domain retrieval not yet implemented"))
}

// ListDomains lists domains with pagination.
func (c *client) ListDomains(ctx context.Context, opts *ListDomainsOptions) (*DomainListResponse, error) {
	// TODO: Implement domain listing using SVC module
	// Should support pagination and filtering by owner
	// Filter by verification status if specified
	// Return domains with metadata and expiration info
	// Handle empty result sets

	return nil, errors.NewModuleError("svc", "ListDomains",
		fmt.Errorf("domain listing not yet implemented"))
}

// AddCapability adds a capability to a service.
func (c *client) AddCapability(ctx context.Context, serviceID string, capability *Capability) error {
	// TODO: Implement capability addition using SVC module
	// Should validate service ownership and capability schema
	// Build MsgAddCapability and submit to chain
	// Update service record with new capability
	// Validate capability name uniqueness within service

	return errors.NewModuleError("svc", "AddCapability",
		fmt.Errorf("capability addition not yet implemented"))
}

// RemoveCapability removes a capability from a service.
func (c *client) RemoveCapability(ctx context.Context, serviceID string, capabilityID string) error {
	// TODO: Implement capability removal using SVC module
	// Should validate service ownership and capability existence
	// Build MsgRemoveCapability and submit to chain
	// Check for dependent services using this capability
	// Handle cascading capability removal safely

	return errors.NewModuleError("svc", "RemoveCapability",
		fmt.Errorf("capability removal not yet implemented"))
}

// ListCapabilities lists service capabilities.
func (c *client) ListCapabilities(ctx context.Context, serviceID string) ([]*Capability, error) {
	// TODO: Implement capability listing using SVC module
	// Should query service record for capabilities
	// Return capabilities with schemas and metadata
	// Include capability status (enabled/disabled)
	// Handle service not found errors

	return nil, errors.NewModuleError("svc", "ListCapabilities",
		fmt.Errorf("capability listing not yet implemented"))
}

// AddEndpoint adds an endpoint to a service.
func (c *client) AddEndpoint(ctx context.Context, serviceID string, endpoint *Endpoint) error {
	// TODO: Implement endpoint addition using SVC module
	// Should validate service ownership and endpoint URL format
	// Build MsgAddEndpoint and submit to chain
	// Validate endpoint accessibility if enabled
	// Update service record with new endpoint

	return errors.NewModuleError("svc", "AddEndpoint",
		fmt.Errorf("endpoint addition not yet implemented"))
}

// UpdateEndpoint updates a service endpoint.
func (c *client) UpdateEndpoint(ctx context.Context, serviceID string, endpointID string, opts *UpdateEndpointOptions) error {
	// TODO: Implement endpoint updates using SVC module
	// Should validate service ownership and endpoint existence
	// Build MsgUpdateEndpoint with selective field updates
	// Validate new URL format and accessibility
	// Preserve existing headers and metadata unless modified

	return errors.NewModuleError("svc", "UpdateEndpoint",
		fmt.Errorf("endpoint updates not yet implemented"))
}

// RemoveEndpoint removes an endpoint from a service.
func (c *client) RemoveEndpoint(ctx context.Context, serviceID string, endpointID string) error {
	// TODO: Implement endpoint removal using SVC module
	// Should validate service ownership and endpoint existence
	// Build MsgRemoveEndpoint and submit to chain
	// Check if endpoint is primary before removal
	// Handle graceful endpoint shutdown

	return errors.NewModuleError("svc", "RemoveEndpoint",
		fmt.Errorf("endpoint removal not yet implemented"))
}

// CheckServiceHealth checks the health status of a service.
func (c *client) CheckServiceHealth(ctx context.Context, serviceID string) (*HealthStatus, error) {
	// TODO: Implement health checking using SVC module
	// Should query service endpoints for health status
	// Aggregate health across multiple endpoints
	// Check endpoint response times and error rates
	// Return comprehensive health report with metrics

	return nil, errors.NewModuleError("svc", "CheckServiceHealth",
		fmt.Errorf("health checking not yet implemented"))
}

// UpdateServiceHealth updates the health status of a service.
func (c *client) UpdateServiceHealth(ctx context.Context, serviceID string, status *HealthStatus) error {
	// TODO: Implement health status updates using SVC module
	// Should validate service ownership and status format
	// Build MsgUpdateServiceHealth and submit to chain
	// Update service discovery with new health status
	// Store health metrics and historical data

	return errors.NewModuleError("svc", "UpdateServiceHealth",
		fmt.Errorf("health status updates not yet implemented"))
}

// Utility functions

// GenerateServiceID generates a unique service ID.
func GenerateServiceID(name string) string {
	// TODO: Implement proper service ID generation
	return fmt.Sprintf("svc_%s_%d", name, time.Now().UnixNano())
}

// ValidateServiceName validates a service name.
func ValidateServiceName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("service name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("service name too long")
	}

	return nil
}

// CreateDefaultEndpoint creates a default HTTP endpoint.
func CreateDefaultEndpoint(url string) *Endpoint {
	return &Endpoint{
		ID:      fmt.Sprintf("endpoint_%d", time.Now().UnixNano()),
		URL:     url,
		Type:    "REST",
		Method:  "GET",
		Enabled: true,
	}
}

// CreateHealthyStatus creates a healthy status.
func CreateHealthyStatus() *HealthStatus {
	return &HealthStatus{
		Status:      "healthy",
		LastChecked: time.Now().UTC().Format(time.RFC3339),
		Message:     "Service is operating normally",
	}
}

// Message Builders - These create the actual transaction messages

// BuildMsgRegisterService builds a MsgRegisterService message.
func BuildMsgRegisterService(creator string, opts *RegisterServiceOptions) (*svctypes.MsgRegisterService, error) {
	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}

	if err := ValidateServiceName(opts.Name); err != nil {
		return nil, fmt.Errorf("invalid service name: %w", err)
	}

	// Generate service ID if not provided
	serviceID := GenerateServiceID(opts.Name)

	// Extract requested permissions from capabilities
	var requestedPermissions []string
	for _, cap := range opts.Capabilities {
		if cap != nil {
			requestedPermissions = append(requestedPermissions, cap.Name)
		}
	}

	return &svctypes.MsgRegisterService{
		Creator:              creator,
		ServiceId:            serviceID,
		Domain:               opts.Domain,
		RequestedPermissions: requestedPermissions,
		UcanDelegationChain:  "", // Will be set if UCAN is used
	}, nil
}

// BuildMsgInitiateDomainVerification builds a MsgInitiateDomainVerification message.
func BuildMsgInitiateDomainVerification(creator, domain string) (*svctypes.MsgInitiateDomainVerification, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	return &svctypes.MsgInitiateDomainVerification{
		Creator: creator,
		Domain:  domain,
	}, nil
}

// BuildMsgVerifyDomain builds a MsgVerifyDomain message.
func BuildMsgVerifyDomain(creator, domain string) (*svctypes.MsgVerifyDomain, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	return &svctypes.MsgVerifyDomain{
		Creator: creator,
		Domain:  domain,
	}, nil
}
