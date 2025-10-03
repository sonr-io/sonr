## Overview
The SVC (Service) module manages service registration with domain verification and UCAN-based authorization.

## Request: {{.RequestType.Name}}

| Field ID | Name | Type | Description |
| -------- | ---- | ---- | ----------- | {{range .RequestType.Fields}}
| {{.Number}} | {{.Name}} | {{if eq .Label.String "LABEL_REPEATED"}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## Response: {{.ResponseType.Name}}

| Field ID | Name | Type | Description |
| -------- | ---- | ---- | ----------- | {{range .ResponseType.Fields}}
| {{.Number}} | {{.Name}} | {{if eq .Label.String "LABEL_REPEATED"}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## Implementation Details

{{if eq .MethodDescriptorProto.Name "InitiateDomainVerification"}}
- Starts domain ownership verification process
- Generates unique verification token
- Requires DNS TXT record setup
- Provides clear instructions for domain setup
{{else if eq .MethodDescriptorProto.Name "VerifyDomain"}}
- Validates DNS TXT records for domain ownership
- Checks for correct verification token
- Establishes trust for service binding
- One-time verification persists on-chain
{{else if eq .MethodDescriptorProto.Name "RegisterService"}}
- Registers services with verified domains
- Binds service to specific domain endpoints
- Establishes permission boundaries
- Creates root capability with UCAN
{{else if eq .MethodDescriptorProto.Name "Params"}}
- Returns module configuration parameters
- Shows verification timeout settings
- Lists supported service types
{{else if eq .MethodDescriptorProto.Name "DomainVerification"}}
- Queries domain verification status
- Returns verification token if pending
- Shows verification timestamp if completed
{{else if eq .MethodDescriptorProto.Name "Service"}}
- Retrieves service details by ID
- Returns bound domain and permissions
- Shows UCAN capability chain
{{else if eq .MethodDescriptorProto.Name "ServicesByOwner"}}
- Lists all services owned by an address
- Useful for service management interfaces
- Returns service metadata and status
{{else if eq .MethodDescriptorProto.Name "ServicesByDomain"}}
- Finds services bound to a domain
- Supports service discovery by domain
- Returns active service endpoints
{{end}}

## Domain Verification Process

1. **Initiate**: Generate verification token
2. **Configure**: Add DNS TXT record with token
3. **Verify**: Check DNS records for ownership proof
4. **Register**: Bind services to verified domain

## UCAN Authorization

- **Delegation Chain**: Hierarchical permission delegation
- **Capability-based**: Fine-grained access control
- **JWT Format**: Standard token representation
- **Root Capabilities**: Stored on IPFS with CID reference

## Security Features

- Domain ownership verification prevents impersonation
- UCAN tokens enable secure delegation
- Permission scoping limits service capabilities
- On-chain verification audit trail

## Usage Examples

```bash
# Initiate domain verification
snrd tx svc initiate-domain-verification example.com

# Verify domain ownership
snrd tx svc verify-domain example.com

# Register service
snrd tx svc register-service service-1 example.com --permissions read,write

# Query verification status
snrd query svc domain example.com

# List services for domain
snrd query svc services-by-domain example.com
```