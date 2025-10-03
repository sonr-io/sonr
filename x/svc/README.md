# `x/svc`

The Service (SVC) module manages the registration and operation of decentralized services within the Sonr ecosystem. It provides a comprehensive framework for services to register with verified domains, define their permission requirements, and integrate with the broader Sonr authorization system through UCAN capabilities.

## Overview

The SVC module provides:

- **Domain Verification**: DNS-based domain ownership verification
- **Service Registration**: Register services with verified domains
- **Permission Management**: Define and request specific UCAN permissions
- **Service Discovery**: Query services by owner, domain, or ID
- **Capability Integration**: Seamless integration with the UCAN module

## Core Concepts

### Domain Verification

Services must verify ownership of their domain through DNS TXT records before registration. This ensures that only legitimate domain owners can register services.

### Service Registration

Once domain ownership is verified, services can be registered with:

- Unique service ID
- Verified domain binding
- Requested permissions (UCAN capabilities)
- Service metadata (name, description)

### Permission Model

Services request specific permissions during registration, which are granted as UCAN capabilities. These permissions define what actions the service can perform on behalf of users.

### Service Identity

Each service has a unique identity composed of:

- Service ID (chosen identifier)
- Domain (verified TLD)
- Owner (blockchain address)

## State

### Domain Verification

```protobuf
message DomainVerification {
  string domain = 1;                 // Domain being verified
  string owner = 2;                  // Address initiating verification
  string token = 3;                  // Verification token
  VerificationStatus status = 4;     // Pending, Verified, Failed
  int64 initiated_at = 5;           // Timestamp
  int64 expires_at = 6;             // Token expiration
}
```

### Service

```protobuf
message Service {
  string service_id = 1;            // Unique service identifier
  string domain = 2;                // Verified domain
  string owner = 3;                 // Service owner address
  string name = 4;                  // Human-readable name
  string description = 5;           // Service description
  repeated string permissions = 6;   // Requested permissions
  string ucan_delegation_chain = 7; // UCAN authorization
  int64 created_at = 8;            // Creation timestamp
  int64 updated_at = 9;            // Last update timestamp
}
```

## Messages

### Domain Verification

#### MsgInitiateDomainVerification

Initiates domain verification by generating a DNS TXT record token.

```protobuf
message MsgInitiateDomainVerification {
  string owner = 1;
  string domain = 2;
}
```

#### MsgVerifyDomain

Verifies domain ownership by checking DNS TXT records.

```protobuf
message MsgVerifyDomain {
  string owner = 1;
  string domain = 2;
}
```

### Service Management

#### MsgRegisterService

Registers a new service with a verified domain.

```protobuf
message MsgRegisterService {
  string owner = 1;
  string service_id = 2;
  string domain = 3;
  string name = 4;
  string description = 5;
  repeated string requested_permissions = 6;
  string ucan_delegation_chain = 7;  // Optional UCAN authorization
}
```

### Governance

#### MsgUpdateParams

Updates module parameters (governance only).

```protobuf
message MsgUpdateParams {
  string authority = 1;
  Params params = 2;
}
```

## Queries

### Domain Queries

- `DomainVerification`: Check domain verification status

### Service Queries

- `Service`: Get a specific service by ID
- `ServicesByOwner`: List all services owned by an address
- `ServicesByDomain`: List all services for a domain

### Module Queries

- `Params`: Get module parameters

## CLI Examples

### Domain Verification

```bash
# Initiate domain verification
snrd tx svc initiate-domain-verification example.com --from alice

# Check verification status
snrd query svc domain-verification example.com

# The system will provide a token like: sonr-verification=abc123xyz
# Add this as a TXT record to your domain's DNS

# Verify domain after DNS propagation
snrd tx svc verify-domain example.com --from alice
```

### Service Registration

```bash
# Register a service with basic permissions
snrd tx svc register-service my-app example.com \
  dwn:read,dwn:write,identity:read \
  --from alice

# Register with UCAN delegation
snrd tx svc register-service vault-service vault.example.com \
  vault:access,identity:read \
  --ucan-delegation-chain="<jwt-token>" \
  --from alice

# Query service
snrd query svc service my-app

# Query services by owner
snrd query svc services-by-owner $(snrd keys show alice -a)

# Query services by domain
snrd query svc services-by-domain example.com
```

## Integration Guide

### For Service Developers

1. **Domain Setup**:
   - Register your domain with a DNS provider
   - Ensure you have access to manage DNS TXT records
   - Choose a unique service ID

2. **Verification Process**:

   ```bash
   # Step 1: Initiate verification
   snrd tx svc initiate-domain-verification your-domain.com --from your-key

   # Step 2: Add TXT record to DNS
   # Record: sonr-verification=<provided-token>

   # Step 3: Wait for DNS propagation (usually 5-30 minutes)

   # Step 4: Complete verification
   snrd tx svc verify-domain your-domain.com --from your-key
   ```

3. **Service Registration**:
   - Define required permissions carefully
   - Use descriptive service names and descriptions
   - Consider permission scope and user privacy

4. **Permission Planning**:
   Common permission patterns:
   - `dwn:read,dwn:write` - Basic data access
   - `identity:read` - Read user identity
   - `vault:access` - Vault operations
   - `credentials:verify` - Verify credentials

### For Application Integrators

1. **Service Discovery**:

   ```bash
   # Find services by domain
   snrd query svc services-by-domain app.example.com

   # Get service details
   snrd query svc service service-id
   ```

2. **Permission Verification**:
   - Check service permissions before integration
   - Validate UCAN delegation chains
   - Ensure permissions match your requirements

3. **User Authorization Flow**:
   - Service requests permissions from user
   - User reviews and approves via wallet
   - Service receives UCAN capability
   - Service can act on user's behalf

## Domain Verification Process

### DNS TXT Record Format

```
sonr-verification=<token>
```

### Verification Requirements

- Domain must be a valid TLD
- DNS TXT record must match the generated token
- Verification expires after 7 days if not completed
- Each domain can only be verified by one owner

### Example DNS Configuration

```
# For domain: app.example.com
# Add TXT record:
Type: TXT
Name: @ (or app if subdomain)
Value: sonr-verification=1234567890abcdef
TTL: 300 (5 minutes)
```

## Security Considerations

1. **Domain Ownership**: Only verified domain owners can register services
2. **Permission Scope**: Services can only request, not grant permissions
3. **UCAN Validation**: All capability chains are validated
4. **Unique Domains**: Each domain can only have one owner
5. **Service Isolation**: Services cannot access data from other services

## Module Parameters

- `verification_timeout`: Domain verification timeout (default: 7 days)
- `max_services_per_owner`: Maximum services per owner (default: 100)
- `allowed_permissions`: List of permissions services can request
- `service_registration_fee`: Fee for service registration (default: 1000usnr)

## Events

The module emits the following events:

- `domain_verification_initiated`: When verification starts
  - `domain`, `owner`, `token`, `expires_at`
- `domain_verified`: When domain is successfully verified
  - `domain`, `owner`, `verified_at`
- `service_registered`: When a new service is registered
  - `service_id`, `domain`, `owner`, `permissions`
- `service_updated`: When service is updated
  - `service_id`, `fields_updated`

## Building and Testing

### Running Tests

```bash
# Run unit tests
make -C x/svc test

# Run tests with race detection
make -C x/svc test-race

# Generate coverage report
make -C x/svc test-cover

# Run benchmarks
make -C x/svc benchmark
```

## Best Practices

### For Service Developers

1. **Choose Meaningful IDs**: Use descriptive service IDs that reflect your service
2. **Request Minimal Permissions**: Only request what you need
3. **Document Permissions**: Clearly explain why each permission is needed
4. **Plan for Updates**: Design your permission model for future growth
5. **Monitor Expiration**: Keep track of UCAN expiration times

### For Users

1. **Verify Services**: Check domain ownership before granting permissions
2. **Review Permissions**: Understand what each permission allows
3. **Regular Audits**: Review granted permissions periodically
4. **Revoke When Needed**: Remove permissions from unused services

## Future Enhancements

- **Service Categories**: Categorization for better discovery
- **Reputation System**: User ratings and reviews
- **Permission Templates**: Pre-defined permission sets
- **Multi-sig Ownership**: Support for team-owned services
- **Service Analytics**: Usage statistics and monitoring
- **Subdomain Support**: Hierarchical service structures
