## Overview
The DID module implements W3C Decentralized Identifiers (DIDs) with support for WebAuthn, verifiable credentials, and external wallet linking.

## Request: {{.RequestType.Name}}

| Field ID | Name | Type | Description |
| -------- | ---- | ---- | ----------- | {{range .RequestType.Fields}}
| {{.Number}} | {{.Name}} | {{if eq .Label.String "LABEL_REPEATED"}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## Response: {{.ResponseType.Name}}

| Field ID | Name | Type | Description |
| -------- | ---- | ---- | ----------- | {{range .ResponseType.Fields}}
| {{.Number}} | {{.Name}} | {{if eq .Label.String "LABEL_REPEATED"}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## Implementation Details

{{if eq .MethodDescriptorProto.Name "CreateDID"}}
- Creates a new W3C-compliant DID document
- Automatically generates unique DID identifier
- Can optionally create an associated vault for secure storage
- Supports multiple verification methods and service endpoints
{{else if eq .MethodDescriptorProto.Name "UpdateDID"}}
- Updates an existing DID document
- Maintains version history for auditability
- Only controller can perform updates
- Preserves immutable fields like creation timestamp
{{else if eq .MethodDescriptorProto.Name "DeactivateDID"}}
- Permanently deactivates a DID document
- Cannot be reversed once deactivated
- Maintains deactivated state for historical reference
- Revokes all associated credentials
{{else if eq .MethodDescriptorProto.Name "AddVerificationMethod"}}
- Adds new cryptographic verification methods
- Supports Ed25519, ECDSA, RSA, and WebAuthn
- Can specify verification relationships (authentication, assertion, etc.)
- Enables multi-signature capabilities
{{else if eq .MethodDescriptorProto.Name "RemoveVerificationMethod"}}
- Removes verification methods from DID document
- Cannot remove the last remaining verification method
- Updates all related verification relationships
{{else if eq .MethodDescriptorProto.Name "AddService"}}
- Adds service endpoints to DID document
- Enables discovery of services associated with DID
- Supports multiple service types and endpoints
{{else if eq .MethodDescriptorProto.Name "RemoveService"}}
- Removes service endpoints from DID document
- Cleans up service discovery metadata
{{else if eq .MethodDescriptorProto.Name "IssueVerifiableCredential"}}
- Issues W3C Verifiable Credentials
- Cryptographically signed by issuer's DID
- Supports custom credential schemas
- Includes issuance and expiration dates
{{else if eq .MethodDescriptorProto.Name "RevokeVerifiableCredential"}}
- Revokes previously issued credentials
- Maintains revocation registry on-chain
- Includes revocation reason for audit trail
{{else if eq .MethodDescriptorProto.Name "LinkExternalWallet"}}
- Links external blockchain wallets to DID
- Requires cryptographic proof of wallet ownership
- Supports Ethereum, Cosmos, and other chains
- Enables cross-chain identity bridging
{{else if eq .MethodDescriptorProto.Name "RegisterWebAuthnCredential"}}
- Registers WebAuthn credentials for passwordless authentication
- Creates new DID with WebAuthn as primary verification method
- Supports gasless onboarding for new users
- Can automatically create associated vault
- Enables biometric authentication (Face ID, Touch ID, etc.)
{{end}}

## Security Considerations

- All DID operations require controller authorization
- Cryptographic signatures validate all changes
- WebAuthn provides phishing-resistant authentication
- External wallet linking requires ownership proofs
- Verifiable credentials include tamper-evident signatures

## Standards Compliance

- W3C DID Core Specification v1.0
- W3C Verifiable Credentials Data Model v1.1
- WebAuthn Level 2 Specification
- DID Method Specification: did:sonr