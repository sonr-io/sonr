## Overview
Query endpoints for the DID module provide read access to decentralized identifiers, verifiable credentials, and identity metadata.

## Request: {{.RequestType.Name}}

| Field ID | Name | Type | Description |
| -------- | ---- | ---- | ----------- | {{range .RequestType.Fields}}
| {{.Number}} | {{.Name}} | {{if eq .Label.String "LABEL_REPEATED"}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## Response: {{.ResponseType.Name}}

| Field ID | Name | Type | Description |
| -------- | ---- | ---- | ----------- | {{range .ResponseType.Fields}}
| {{.Number}} | {{.Name}} | {{if eq .Label.String "LABEL_REPEATED"}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## Query Details

{{if eq .MethodDescriptorProto.Name "Params"}}
- Returns module parameters
- Includes WebAuthn configuration settings
- Shows supported signature algorithms
{{else if eq .MethodDescriptorProto.Name "ResolveDID"}}
- Resolves a DID to its complete document
- Returns verification methods and service endpoints
- Includes document metadata and version info
{{else if eq .MethodDescriptorProto.Name "GetDIDDocument"}}
- Retrieves a specific DID document by identifier
- Returns full W3C-compliant DID document
- Includes creation and update timestamps
{{else if eq .MethodDescriptorProto.Name "ListDIDDocuments"}}
- Lists all DID documents in the system
- Supports pagination for large result sets
- Returns documents in creation order
{{else if eq .MethodDescriptorProto.Name "GetDIDDocumentsByController"}}
- Finds all DIDs controlled by an address
- Useful for identity management interfaces
- Returns paginated results
{{else if eq .MethodDescriptorProto.Name "GetVerificationMethod"}}
- Retrieves specific verification method details
- Returns public key and algorithm information
- Shows associated verification relationships
{{else if eq .MethodDescriptorProto.Name "GetService"}}
- Gets specific service endpoint information
- Returns service type and endpoint URLs
- Includes service metadata
{{else if eq .MethodDescriptorProto.Name "GetVerifiableCredential"}}
- Retrieves a verifiable credential by ID
- Returns complete credential with proofs
- Shows issuance and expiration dates
{{else if eq .MethodDescriptorProto.Name "ListVerifiableCredentials"}}
- Lists verifiable credentials with filters
- Can filter by issuer or holder DID
- Optionally includes revoked credentials
{{else if eq .MethodDescriptorProto.Name "GetCredentialsByDID"}}
- Gets all credentials associated with a DID
- Includes both verifiable and WebAuthn credentials
- Shows vault storage status if applicable
{{end}}

## Usage Examples

```bash
# Query module parameters
snrd query did params

# Resolve a DID
snrd query did resolve did:sonr:123abc

# Get DID document
snrd query did document did:sonr:123abc

# List DIDs by controller
snrd query did documents-by-controller sonr1abc...

# Get verification method
snrd query did verification-method did:sonr:123abc key-1

# Query verifiable credential
snrd query did credential cred-123

# List credentials for a DID
snrd query did credentials-by-did did:sonr:123abc
```

## Response Formats

All responses follow W3C standards:
- DID Documents conform to W3C DID Core v1.0
- Verifiable Credentials follow W3C VC Data Model v1.1
- WebAuthn credentials comply with WebAuthn Level 2