# `x/did`

The Decentralized Identifier (DID) module implements the W3C DID specification for the Sonr blockchain, providing a decentralized identity layer that enables self-sovereign identity management. This module allows users to create, manage, and resolve DIDs while supporting verifiable credentials and various authentication methods including WebAuthn.

## Overview

The DID module provides:

- **W3C DID Documents**: Full implementation of the W3C DID Core specification
- **Verifiable Credentials**: Issue and manage W3C Verifiable Credentials
- **Multiple Verification Methods**: Support for various key types and WebAuthn
- **Signature Verification**: Multi-algorithm cryptographic signature verification
- **Service Endpoints**: Associate services with DIDs
- **Decentralized Resolution**: On-chain DID resolution without external dependencies

## Core Concepts

### Decentralized Identifiers (DIDs)

DIDs are globally unique identifiers that are created and controlled by their owners without requiring a central authority. In Sonr, DIDs follow the format: `did:sonr:<unique-identifier>`

### DID Documents

DID Documents are the core data structures that contain:

- Verification methods (public keys, WebAuthn credentials)
- Service endpoints
- Controller relationships
- Authentication and other verification relationships

### Verifiable Credentials

Digital credentials that can be cryptographically verified, containing claims about subjects and issued by trusted authorities.

### WebAuthn Integration

Native support for WebAuthn allows users to control their DIDs using device biometrics and hardware security keys.

## State

### W3C DID Documents

```protobuf
message DIDDocument {
  string id = 1;                                    // DID identifier
  string context = 2;                               // JSON-LD context
  repeated string controller = 3;                   // Controller DIDs
  repeated VerificationMethod verification_method = 4;
  repeated string authentication = 5;              // References to verification methods
  repeated string assertion_method = 6;
  repeated string key_agreement = 7;
  repeated string capability_invocation = 8;
  repeated string capability_delegation = 9;
  repeated Service service = 10;               // Service endpoints
}
```

### W3C Verification Methods

```protobuf
message VerificationMethod {
  string id = 1;                                   // Method identifier
  string verification_method_kind = 2;             // Type of verification method
  string controller = 3;                           // Controller DID
  oneof verification_material {
    W3CJSONWebKey json_web_key = 4;              // JWK representation
    string public_key_multibase = 5;              // Multibase encoded key
    WebAuthnCredential webauthn_credential = 6; // WebAuthn credential
  }
  string blockchain_account_id = 11;               // CAIP-10 blockchain account ID
}
```

### W3C Verifiable Credentials

```protobuf
message VerifiableCredential {
  string id = 1;                                   // Credential identifier
  repeated string type = 2;                        // Credential types
  string issuer = 3;                              // Issuer DID
  google.protobuf.Timestamp issuance_date = 4;    // When issued
  google.protobuf.Timestamp expiration_date = 5;  // When expires
  google.protobuf.Any credential_subject = 6;     // Claims about subject
  google.protobuf.Any proof = 7;                  // Cryptographic proof
  CredentialStatus credential_status = 8;         // Revocation status
}
```

## Messages

### DID Management

#### MsgCreateDID

Creates a new DID document on-chain.

```protobuf
message MsgCreateDID {
  string creator = 1;
  string did = 2;
  DIDDocument document = 3;
  W3CCryptoProof proof = 4;
}
```

#### MsgUpdateDID

Updates an existing DID document.

```protobuf
message MsgUpdateDID {
  string creator = 1;
  string did = 2;
  DIDDocument document = 3;
  W3CCryptoProof proof = 4;
}
```

#### MsgDeactivateDID

Deactivates a DID, making it non-resolvable.

```protobuf
message MsgDeactivateDID {
  string creator = 1;
  string did = 2;
  W3CCryptoProof proof = 3;
}
```

### Verification Method Management

#### MsgAddVerificationMethod

Adds a new verification method to a DID document.

```protobuf
message MsgAddVerificationMethod {
  string creator = 1;
  string did = 2;
  VerificationMethod method = 3;
  W3CCryptoProof proof = 4;
}
```

#### MsgRemoveVerificationMethod

Removes a verification method from a DID document.

```protobuf
message MsgRemoveVerificationMethod {
  string creator = 1;
  string did = 2;
  string method_id = 3;
  W3CCryptoProof proof = 4;
}
```

#### MsgRegisterWebAuthnCredential

Creates a new DID with WebAuthn credential using gasless transaction processing. This message bypasses all fees and signature requirements, enabling users to create their first decentralized identity without existing tokens.

```protobuf
message MsgRegisterWebAuthnCredential {
  string username = 1;                          // Unique username for DID generation
  WebAuthnCredential webauthn_credential = 2; // Complete WebAuthn credential data
  string controller = 3;                         // Controller address (optional in enhanced mode)
  bool auto_create_vault = 4;                   // Whether to auto-create DWN vault
}
```

**Gasless Processing**: This is the only message type that qualifies for gasless transaction processing. The ante handler chain validates the WebAuthn credential and bypasses all fees, signature verification, and gas requirements.

#### MsgLinkExternalWallet

Links an external wallet (MetaMask, Keplr) to a DID as an assertion method.

```protobuf
message MsgLinkExternalWallet {
  string controller = 1;                // DID controller address
  string did = 2;                       // Target DID
  string wallet_address = 3;            // External wallet address
  string chain_id = 4;                  // Blockchain chain ID
  string wallet_type = 5;               // "ethereum" or "cosmos"
  bytes ownership_proof = 6;            // Signature proving ownership
  bytes challenge = 7;                  // Challenge message that was signed
  string verification_method_id = 8;    // ID for new verification method
}
```

### Service Management

#### MsgAddService

Adds a service endpoint to a DID document.

```protobuf
message MsgAddService {
  string creator = 1;
  string did = 2;
  Service service = 3;
  W3CCryptoProof proof = 4;
}
```

#### MsgRemoveService

Removes a service endpoint from a DID document.

```protobuf
message MsgRemoveService {
  string creator = 1;
  string did = 2;
  string service_id = 3;
  W3CCryptoProof proof = 4;
}
```

### Verifiable Credentials

#### MsgIssueVerifiableCredential

Issues a new verifiable credential.

```protobuf
message MsgIssueVerifiableCredential {
  string issuer = 1;
  VerifiableCredential credential = 2;
  W3CCryptoProof proof = 3;
}
```

#### MsgRevokeVerifiableCredential

Revokes an existing verifiable credential.

```protobuf
message MsgRevokeVerifiableCredential {
  string issuer = 1;
  string credential_id = 2;
  string reason = 3;
  W3CCryptoProof proof = 4;
}
```

## Queries

### DID Queries

- `ResolveDID`: Resolve a DID to its document
- `GetDIDDocument`: Get a specific DID document
- `ListDIDDocuments`: List all DID documents
- `GetDIDDocumentsByController`: Get documents by controller

### Verification Method Queries

- `GetVerificationMethod`: Get a specific verification method
- `GetService`: Get a specific service endpoint
- `GetWebAuthnCredentials`: Get all WebAuthn credentials for a DID
- `ListWebAuthnCredentials`: List WebAuthn credentials with pagination support

### WebAuthn Ceremony Queries

- `RegisterStart`: Initiate WebAuthn registration ceremony for new assertion DID
- `LoginStart`: Initiate WebAuthn authentication ceremony for existing assertion DID

### Credential Queries

- `GetVerifiableCredential`: Get a specific credential
- `ListVerifiableCredentials`: List all credentials
- `GetVerifiableCredentialsByIssuer`: Get credentials by issuer
- `GetVerifiableCredentialsByHolder`: Get credentials by holder

## Signature Verification

The DID module includes comprehensive cryptographic signature verification capabilities for validating DID document authenticity and integrity.

### Supported Verification Methods

The `VerifyDIDDocumentSignature` keeper method supports multiple cryptographic signature types:

#### Ed25519VerificationKey2020
- **Algorithm**: Ed25519 elliptic curve signatures
- **Key Formats**: Base64, Hex encoding
- **Use Case**: High-performance, compact signatures
- **Security**: Quantum-resistant, 128-bit security level

#### JsonWebSignature2020
- **Algorithm**: JSON Web Signature (JWS) format
- **Key Formats**: JSON Web Key (JWK) with OKP, EC, RSA support
- **Use Case**: Web-compatible signature format
- **Features**: Flexible payload encoding, multiple signature algorithms

#### WebAuthn
- **Algorithm**: WebAuthn assertion verification
- **Key Formats**: CBOR-encoded public keys
- **Use Case**: Device-based authentication with biometrics
- **Features**: Platform/roaming authenticators, origin validation

#### ECDSASecp256k1VerificationKey2019
- **Algorithm**: ECDSA with secp256k1 curve
- **Key Formats**: PEM-encoded public keys
- **Use Case**: Bitcoin/Ethereum compatible signatures
- **Features**: Blockchain interoperability

#### RSAVerificationKey2018
- **Algorithm**: RSA with PKCS#1 v1.5 padding
- **Key Formats**: PEM-encoded public keys
- **Use Case**: Legacy system compatibility
- **Features**: Configurable key sizes (2048, 3072, 4096 bits)

### Verification Process

1. **DID Resolution**: Retrieve DID document from on-chain storage
2. **Status Check**: Verify DID document is not deactivated
3. **Method Iteration**: Test signature against all verification methods
4. **Cryptographic Verification**: Validate signature using appropriate algorithm
5. **Result**: Return verification status and detailed error information

### Security Features

- **Deactivation Protection**: Prevents verification of deactivated DIDs
- **Method Fallback**: Tries all verification methods until one succeeds
- **Comprehensive Logging**: Debug information for verification failures
- **Error Handling**: Detailed error messages for troubleshooting

### Usage Example

```go
// Verify a signature against a DID document
verified, err := keeper.VerifyDIDDocumentSignature(ctx, "did:sonr:123", signatureBytes)
if err != nil {
    // Handle verification error
    return fmt.Errorf("signature verification failed: %w", err)
}

if verified {
    // Signature is valid
    // Proceed with authenticated operation
} else {
    // Signature is invalid
    // Reject the request
}
```

## External Wallet Linking

The DID module supports linking external wallets (MetaMask, Keplr, etc.) as assertion methods in DID documents. This allows users to control their DID using existing wallets while maintaining security and compliance with W3C standards.

### Supported Wallet Types

#### Ethereum Wallets
- **Wallet Type**: `ethereum`
- **Verification Method**: `EcdsaSecp256k1RecoveryMethod2020`
- **Chain Namespace**: `eip155` (following CAIP-10 standard)
- **Signature Format**: Personal message signature with recovery parameter
- **Examples**: MetaMask, Coinbase Wallet, WalletConnect

#### Cosmos Wallets
- **Wallet Type**: `cosmos`
- **Verification Method**: `Secp256k1VerificationKey2018`  
- **Chain Namespace**: `cosmos` (following CAIP-10 standard)
- **Signature Format**: Secp256k1 signature
- **Examples**: Keplr, Leap Wallet, Cosmostation

### Blockchain Account ID Format

External wallets are identified using the CAIP-10 blockchain account ID standard:

```
<namespace>:<chain_id>:<address>
```

**Examples:**
- Ethereum: `eip155:1:0x742d35Cc6635C0532925a3b8c17C6e583F4d6A42`
- Cosmos Hub: `cosmos:cosmoshub-4:cosmos1abc123def456ghi789`

### Wallet Linking Process

1. **DID Requirements**: DID must have an active DWN vault controller
2. **Challenge Generation**: System generates a unique challenge message
3. **Signature Creation**: User signs challenge with their external wallet
4. **Ownership Verification**: On-chain verification of signature authenticity
5. **Method Addition**: Wallet added as assertion method to DID document

### Security Features

- **Signature Verification**: Cryptographic proof of wallet ownership
- **Duplicate Prevention**: Prevents same wallet from being linked to multiple DIDs
- **Challenge-Response**: Prevents replay attacks with unique challenges
- **Controller Authorization**: Only authorized controllers can link wallets
- **DWN Vault Requirement**: Ensures DIDs have proper vault management

### Verification Method Structure

When a wallet is linked, a new verification method is added to the DID document:

```json
{
  "id": "did:sonr:123#wallet-1",
  "type": "EcdsaSecp256k1RecoveryMethod2020",
  "controller": "did:sonr:123",
  "blockchainAccountId": "eip155:1:0x742d35Cc6635C0532925a3b8c17C6e583F4d6A42"
}
```

### Linking Workflow

1. **Prepare Transaction**:
   ```go
   msg := &types.MsgLinkExternalWallet{
       Controller:           controllerAddr,
       Did:                  "did:sonr:123",
       WalletAddress:        "0x742d35Cc6635C0532925a3b8c17C6e583F4d6A42",
       ChainId:              "1",
       WalletType:           "ethereum",
       Challenge:            challengeBytes,
       OwnershipProof:       signatureBytes,
       VerificationMethodId: "did:sonr:123#wallet-1",
   }
   ```

2. **Generate Challenge**:
   ```
   "Link wallet 0x742d35Cc6635C0532925a3b8c17C6e583F4d6A42 to DID did:sonr:123 at block 12345. This proves ownership of the wallet."
   ```

3. **Create Signature**: User signs challenge with wallet
4. **Submit Transaction**: Broadcast MsgLinkExternalWallet
5. **Verification**: Chain verifies ownership and updates DID document

## WebAuthn Support

The DID module provides comprehensive WebAuthn integration for passwordless authentication, including a revolutionary **gasless registration** system that enables users to create their first decentralized identity without requiring existing tokens.

### WebAuthn Credential Structure

```protobuf
message WebAuthnCredential {
  string credential_id = 1;         // Base64url-encoded credential identifier
  string raw_id = 2;                // Base64url-encoded raw credential ID
  bytes public_key = 3;             // Credential public key (DER format)
  int32 algorithm = 4;              // Signing algorithm (e.g., -7 for ES256, -257 for RS256)
  string attestation_type = 5;      // Attestation type ("none", "basic", "attca")
  string origin = 6;                // Origin where credential was created
  int64 created_at = 7;             // Creation timestamp
  string client_data_json = 8;      // Base64url-encoded client data JSON
  string attestation_object = 9;    // Base64url-encoded attestation object
}
```

### Gasless WebAuthn Registration

The **gasless registration** system allows users to create their first DID without requiring any cryptocurrency. This removes the traditional Web3 barrier and enables true decentralized identity onboarding.

#### How Gasless Registration Works

1. **CLI Orchestration**: User runs `snrd auth register --username <username>`
2. **Browser Integration**: CLI spawns local server and opens browser for WebAuthn ceremony
3. **Credential Creation**: User completes biometric authentication in browser
4. **Gasless Transaction**: Transaction bypasses all fees and signature requirements
5. **DID Creation**: User receives complete DID document and optional vault
6. **Immediate Usability**: User can begin using Sonr ecosystem without tokens

#### Security Architecture

The gasless system maintains strong security through multiple layers:

- **Credential Validation**: Full cryptographic validation of WebAuthn credentials
- **Uniqueness Enforcement**: Prevents credential reuse across accounts
- **Origin Validation**: Ensures credentials created from legitimate origins
- **Anti-Replay Protection**: Prevents credential replay attacks
- **Limited Scope**: Only applies to single WebAuthn registration messages

#### Transaction Message

```protobuf
message MsgRegisterWebAuthnCredential {
  string username = 1;                          // Unique username for the DID
  WebAuthnCredential webauthn_credential = 2; // Complete WebAuthn credential data
  string controller = 3;                         // Optional controller address
  bool auto_create_vault = 4;                   // Whether to create DWN vault automatically
}
```

### WebAuthn Authentication Flow

#### Standard Registration
1. **Challenge Generation**: Server generates cryptographic challenge
2. **Browser Ceremony**: User completes WebAuthn registration in browser
3. **Credential Verification**: Server validates attestation and public key
4. **DID Storage**: Credential stored as verification method in DID document

#### Gasless Registration Flow
1. **CLI Command**: `snrd auth register --username alice --auto-vault`
2. **Server Startup**: CLI spawns HTTP server on random port
3. **Browser Launch**: System opens browser to WebAuthn registration page
4. **User Interaction**: User completes biometric authentication
5. **Credential Capture**: Browser sends credential data back to CLI
6. **Transaction Building**: CLI creates gasless MsgRegisterWebAuthnCredential
7. **Blockchain Processing**: Transaction processed without fees or signatures
8. **DID Creation**: Complete DID document created with WebAuthn verification method

### Supported WebAuthn Features

- **Passkeys**: Platform authenticators (Face ID, Touch ID, Windows Hello)
- **Security Keys**: Roaming authenticators (YubiKey, Titan Key)
- **Multiple Algorithms**: ES256 (ECDSA P-256), RS256 (RSA-2048)
- **Cross-Platform**: Works across all major browsers and operating systems
- **Gasless Onboarding**: Zero-cost registration for new users
- **Auto-Vault Creation**: Optional automatic DWN vault setup

### CLI Usage Examples

#### Basic Gasless Registration
```bash
# Create DID with WebAuthn credential (no tokens required)
snrd auth register --username alice

# Register with automatic vault creation
snrd auth register --username bob --auto-vault

# Register with specific controller address
snrd auth register --username carol --controller cosmos1abc...xyz
```

#### Querying WebAuthn Credentials
```bash
# Get all WebAuthn credentials for a DID
snrd query did webauthn-credentials did:sonr:alice

# List all DID documents with WebAuthn verification methods
snrd query did list-dids --filter webauthn
```

### WebAuthn Verification Process

The DID module includes sophisticated WebAuthn credential validation:

1. **Structure Validation**: Required fields, supported algorithms, key format
2. **Cryptographic Verification**: Public key parsing and algorithm validation  
3. **Attestation Validation**: Client data JSON, origin, and challenge verification
4. **Uniqueness Checking**: Ensures credential ID hasn't been used before
5. **Security Validation**: Prevents common WebAuthn security vulnerabilities

### Integration with External Systems

#### Browser Integration
```javascript
// Client-side WebAuthn registration
const credential = await navigator.credentials.create({
  publicKey: {
    challenge: new Uint8Array(32),
    rp: { name: "Sonr", id: "localhost" },
    user: { id: new TextEncoder().encode(username), name: username, displayName: username },
    pubKeyCredParams: [{ alg: -7, type: "public-key" }],
    authenticatorSelection: { authenticatorAttachment: "platform" }
  }
});

// Send credential to Sonr CLI server
await fetch('/register', {
  method: 'POST',
  body: JSON.stringify({
    username: username,
    credential: {
      id: credential.id,
      rawId: arrayBufferToBase64url(credential.rawId),
      type: credential.type,
      response: {
        clientDataJSON: arrayBufferToBase64url(credential.response.clientDataJSON),
        attestationObject: arrayBufferToBase64url(credential.response.attestationObject)
      }
    }
  })
});
```

### WebAuthn Query Endpoints

The DID module provides specialized query endpoints for initiating WebAuthn authentication ceremonies. These endpoints generate cryptographically secure challenges and prepare the client for WebAuthn operations.

#### QueryRegisterStart

Initiates a WebAuthn registration ceremony for a new assertion DID.

**Purpose**: Generate registration options and challenge for creating a new WebAuthn credential.

**Request**:
```protobuf
message QueryRegisterStartRequest {
  string assertion_did = 1; // Format: did:sonr:email:<blake3_hash> or did:sonr:phone:<blake3_hash>
}
```

**Response**:
```protobuf
message QueryRegisterStartResponse {
  bytes challenge = 1;               // Base64url-encoded 32-byte challenge
  string relying_party_id = 2;      // Relying party identifier (domain)
  map<string, string> user = 3;     // User info: id, name, displayName
}
```

**Behavior**:
1. Validates that assertion DID does NOT already exist
2. Generates deterministic challenge using BLAKE3(assertion_did || block_height || chain_id || nonce)
3. Extracts user information from assertion DID type (email, phone, github, etc.)
4. Returns WebAuthn-compliant registration options
5. Stores session state with expiration for later verification

**Example**:
```bash
# Start registration for new email-based DID
snrd query did register-start did:sonr:email:abc123def456

# Response
{
  "challenge": "R3VpZGVkIGRldGVybWluaXN0aWMgY2hhbGxlbmdl...",
  "relying_party_id": "sonr.io",
  "user": {
    "id": "did:sonr:email:abc123def456",
    "name": "Email User",
    "displayName": "Email (abc123de...)"
  }
}
```

**User Info Extraction**:
The endpoint automatically generates user-friendly information based on the assertion DID type:

| Assertion Type | Name | Display Name Format |
|----------------|------|---------------------|
| `email` | "Email User" | "Email (abc123...)" |
| `phone` | "Phone User" | "Phone (xyz789...)" |
| `github` | "GitHub User" | "GitHub (fedcba...)" |
| `google` | "Google User" | "Google (123abc...)" |
| Other | "{Type} User" | "{Type} (hash...)" |

**Error Cases**:
- `ErrInvalidRequest`: Nil request or empty assertion DID
- `ErrAssertionAlreadyExists`: Assertion DID already registered

#### QueryLoginStart

Initiates a WebAuthn authentication ceremony for an existing assertion DID.

**Purpose**: Generate authentication options and retrieve credential IDs for WebAuthn login.

**Request**:
```protobuf
message QueryLoginStartRequest {
  string assertion_did = 1; // Existing assertion DID
}
```

**Response**:
```protobuf
message QueryLoginStartResponse {
  repeated string credential_ids = 1; // WebAuthn credential IDs for allowCredentials
  bytes challenge = 2;                 // Base64url-encoded 32-byte challenge
  string relying_party_id = 3;        // Relying party identifier
}
```

**Behavior**:
1. Validates assertion DID exists in state
2. Retrieves controller DID from assertion relationship
3. Verifies controller DID is not deactivated
4. Extracts WebAuthn credential IDs from controller's authentication methods
5. Filters out non-WebAuthn verification methods (Ed25519, secp256k1, etc.)
6. Generates deterministic challenge
7. Returns credential allow list and challenge

**Example**:
```bash
# Start login for existing DID
snrd query did login-start did:sonr:email:abc123def456

# Response
{
  "credential_ids": [
    "credential_id_1",
    "credential_id_2"
  ],
  "challenge": "QXV0aGVudGljYXRpb24gY2hhbGxlbmdlIGhlcmU...",
  "relying_party_id": "sonr.io"
}
```

**Credential Extraction**:
The endpoint supports both embedded and referenced verification methods:

- **Embedded**: Credentials directly in the `authentication` array
- **Referenced**: Credentials in `verificationMethod` array, referenced by ID

Only WebAuthn-type verification methods (`WebAuthnAuthentication2021`, `WebAuthn2021`) are included in the credential allow list.

**Error Cases**:
- `ErrInvalidRequest`: Nil request or empty assertion DID
- `ErrAssertionNotFound`: Assertion DID doesn't exist
- `ErrInvalidAssertion`: Assertion has no controller
- `ErrDIDNotFound`: Controller DID doesn't exist
- `ErrDIDDeactivated`: Controller DID is deactivated
- `ErrNoCredentials`: No WebAuthn credentials found

#### Deterministic Challenge Generation

Both endpoints use the same deterministic challenge generation algorithm to ensure reproducibility and prevent replay attacks.

**Algorithm**:
```go
// Inputs
nonce := fmt.Sprintf("%d:%s:%d", blockHeight, assertionDid, timestamp)
input := []byte(assertionDid + nonce + chainID)

// BLAKE3 hash
hash := blake3.Sum256(input)

// Base64url encoding (RFC 4648 Section 5)
challenge := base64.RawURLEncoding.EncodeToString(hash[:])
```

**Properties**:
1. **Deterministic**: Same inputs always produce same challenge
2. **Unique**: Different DIDs produce different challenges
3. **Time-bound**: Block height and timestamp prevent replay
4. **Standard-compliant**: 32 bytes meets WebAuthn minimum requirements
5. **Secure encoding**: Base64url without padding per WebAuthn spec

**Session Storage**:
Both endpoints store session state for later verification:
- Session ID: `{blockHeight}:{assertionDid}:{timestamp}`
- Challenge: Base64url-encoded BLAKE3 hash
- Expiration: Configurable timeout (default 60 seconds)
- Nonce: Prevents replay attacks

#### Integration Example

**Complete WebAuthn Registration Flow**:

```javascript
// 1. Call RegisterStart endpoint
const registerStartResp = await fetch('/sonr/did/v1/register-start', {
  method: 'POST',
  body: JSON.stringify({
    assertion_did: 'did:sonr:email:abc123def456'
  })
});

const { challenge, relying_party_id, user } = await registerStartResp.json();

// 2. Perform WebAuthn registration in browser
const credential = await navigator.credentials.create({
  publicKey: {
    challenge: base64urlDecode(challenge),
    rp: { name: "Sonr", id: relying_party_id },
    user: {
      id: new TextEncoder().encode(user.id),
      name: user.name,
      displayName: user.displayName
    },
    pubKeyCredParams: [
      { alg: -7, type: "public-key" },  // ES256
      { alg: -257, type: "public-key" } // RS256
    ],
    authenticatorSelection: {
      authenticatorAttachment: "platform",
      requireResidentKey: false,
      userVerification: "required"
    },
    timeout: 60000,
    attestation: "none"
  }
});

// 3. Submit credential via MsgRegisterWebAuthnCredential
// (Handled by CLI or application)
```

**Complete WebAuthn Authentication Flow**:

```javascript
// 1. Call LoginStart endpoint
const loginStartResp = await fetch('/sonr/did/v1/login-start', {
  method: 'POST',
  body: JSON.stringify({
    assertion_did: 'did:sonr:email:abc123def456'
  })
});

const { credential_ids, challenge, relying_party_id } = await loginStartResp.json();

// 2. Perform WebAuthn authentication in browser
const assertion = await navigator.credentials.get({
  publicKey: {
    challenge: base64urlDecode(challenge),
    rpId: relying_party_id,
    allowCredentials: credential_ids.map(id => ({
      type: "public-key",
      id: base64urlDecode(id)
    })),
    userVerification: "required",
    timeout: 60000
  }
});

// 3. Submit assertion for verification
// (Handled by CLI or application)
```

#### Origin Validation

The DID module implements comprehensive origin validation for WebAuthn operations to prevent unauthorized relying parties from initiating authentication ceremonies.

**Validation Rules**:

1. **HTTPS Requirement**: Non-localhost origins must use HTTPS
2. **Localhost Exception**: `localhost`, `127.0.0.1`, and `[::1]` bypass HTTPS requirement
3. **x/svc Integration**: Origins registered in the x/svc module domain registry are automatically allowed
4. **Module Parameters**: Origins in `params.webauthn.allowed_origins` are permitted
5. **Wildcard Subdomains**: Supports patterns like `*.example.com` matching `app.example.com`, `api.example.com`, etc.

**Validation Process**:

```
1. Check origin format (must start with http:// or https://)
2. Extract domain from origin URL
3. If localhost → Allow (bypass further checks)
4. If HTTP (non-localhost) → Reject
5. Query x/svc module for domain verification
6. If found in x/svc → Allow
7. Check module params allowed_origins list
8. If no params configured → Reject
9. Match against allowed origins (exact or wildcard)
10. If match found → Allow
11. Otherwise → Reject
```

**Module Parameters**:

```protobuf
message WebauthnParams {
  repeated string allowed_origins = 2;  // Allowed WebAuthn origins
  // ...other fields
}
```

**Configuration Example**:

```json
{
  "webauthn": {
    "allowed_origins": [
      "https://sonr.io",
      "https://app.sonr.io",
      "https://*.example.com",  // Matches all subdomains
      "http://localhost:8080",
      "http://localhost:3000"
    ]
  }
}
```

**Wildcard Matching**:

| Pattern | Matches | Does Not Match |
|---------|---------|----------------|
| `https://*.example.com` | `https://app.example.com`<br>`https://api.example.com`<br>`https://example.com` | `https://example.org`<br>`https://malicious.com` |
| `https://sonr.io` | `https://sonr.io` (exact) | `https://app.sonr.io`<br>`https://sonr.com` |
| `http://localhost:*` | Not supported | Use exact ports |

**Security Features**:

- **IPv6 Support**: Properly handles `[::1]` and other IPv6 addresses
- **Port Handling**: Strips ports before domain matching
- **Path Stripping**: Removes paths and query parameters
- **Case Sensitive**: Domain matching is case-sensitive per RFC spec
- **Scheme Validation**: Enforces http:// or https:// prefix

**Helper Methods** (exported for testing):

- `ValidateServiceOrigin(ctx, origin)`: Main validation entry point
- `IsLocalhostOrigin(domain)`: Checks if domain is localhost
- `MatchesOrigin(fullOrigin, domain, allowedOrigin)`: Pattern matching with wildcard support
- `ExtractDomainFromOrigin(origin)`: Extracts domain from URL

**Integration with x/svc Module**:

The DID module can query the x/svc module's service registry for domain verification. If a service in x/svc has registered a domain, that domain is automatically trusted for WebAuthn operations:

```go
// x/svc integration (optional)
if k.serviceKeeper != nil {
    if err := k.serviceKeeper.VerifyOrigin(ctx, origin); err == nil {
        // Origin verified via x/svc module
        return nil
    }
}
// Fall back to module params check
```

### Security Considerations

#### Gasless Transaction Security
- **Limited Scope**: Only MsgRegisterWebAuthnCredential qualifies for gasless processing
- **Credential Validation**: Full WebAuthn cryptographic verification required
- **Anti-Abuse Measures**: Credential uniqueness prevents unlimited account creation
- **Origin Restrictions**: Localhost origins only during registration flow
- **Replay Prevention**: Each credential can only be used once

#### WebAuthn Security Best Practices
- **Origin Validation**: Ensures credentials created from expected domains
- **Challenge Verification**: Prevents replay attacks with unique challenges
- **Algorithm Support**: Only secure algorithms (ES256, RS256) supported
- **Attestation Verification**: Validates authenticator attestation when available
- **Device Binding**: Credentials bound to specific authenticator devices

## CLI Examples

### DID Operations

```bash
# Create a new DID
snrd tx did create-did did:sonr:123 '{"id":"did:sonr:123","verification_method":[...]}' \
  --proof '{"type":"JsonWebSignature2020","created":"2024-01-01T00:00:00Z","proof_purpose":"assertionMethod"}' \
  --from alice

# Resolve a DID
snrd query did resolve did:sonr:123

# Update DID document
snrd tx did update-did did:sonr:123 '{"id":"did:sonr:123","service":[...]}' \
  --proof '{"type":"JsonWebSignature2020"}' \
  --from alice

# Deactivate DID
snrd tx did deactivate-did did:sonr:123 \
  --proof '{"type":"JsonWebSignature2020"}' \
  --from alice
```

### Verification Method Management

```bash
# Add verification method
snrd tx did add-verification-method did:sonr:123 \
  '{"id":"did:sonr:123#key-2","verification_method_kind":"JsonWebKey2020","controller":"did:sonr:123","json_web_key":{...}}' \
  --proof '{"type":"JsonWebSignature2020"}' \
  --from alice

# Remove verification method
snrd tx did remove-verification-method did:sonr:123 did:sonr:123#key-2 \
  --proof '{"type":"JsonWebSignature2020"}' \
  --from alice
```

### External Wallet Linking

```bash
# Link Ethereum wallet (MetaMask example)
snrd tx did link-external-wallet did:sonr:123 \
  --wallet-address 0x742d35Cc6635C0532925a3b8c17C6e583F4d6A42 \
  --chain-id 1 \
  --wallet-type ethereum \
  --challenge "Link wallet 0x742d35Cc6635C0532925a3b8c17C6e583F4d6A42 to DID did:sonr:123 at block 12345" \
  --ownership-proof 0x1234...5678 \
  --verification-method-id did:sonr:123#wallet-1 \
  --from alice

# Link Cosmos wallet (Keplr example)  
snrd tx did link-external-wallet did:sonr:123 \
  --wallet-address cosmos1abc123def456ghi789 \
  --chain-id cosmoshub-4 \
  --wallet-type cosmos \
  --challenge "Link wallet cosmos1abc123def456ghi789 to DID did:sonr:123 at block 12345" \
  --ownership-proof abcd...ef01 \
  --verification-method-id did:sonr:123#keplr-1 \
  --from alice

# Query DID document to verify linked wallet
snrd query did resolve did:sonr:123 | jq '.did_document.assertion_method'
```

### Service Management

```bash
# Add service endpoint
snrd tx did add-service did:sonr:123 \
  '{"id":"did:sonr:123#dwn","type":"DecentralizedWebNode","service_endpoint":"https://dwn.example.com"}' \
  --proof '{"type":"JsonWebSignature2020"}' \
  --from alice

# Remove service
snrd tx did remove-service did:sonr:123 did:sonr:123#dwn \
  --proof '{"type":"JsonWebSignature2020"}' \
  --from alice
```

### Verifiable Credentials

```bash
# Issue credential
snrd tx did issue-credential \
  '{"id":"cred-123","type":["VerifiableCredential"],"issuer":"did:sonr:issuer","credential_subject":{...}}' \
  --proof '{"type":"JsonWebSignature2020"}' \
  --from issuer

# Query credentials by holder
snrd query did credentials-by-holder did:sonr:holder

# Revoke credential
snrd tx did revoke-credential cred-123 \
  --reason "No longer valid" \
  --proof '{"type":"JsonWebSignature2020"}' \
  --from issuer
```

## Integration Guide

### For Application Developers

1. **DID Creation**: Generate DID and initial document structure
2. **Key Management**: Add appropriate verification methods
3. **Service Registration**: Register your application as a service
4. **Credential Issuance**: Issue credentials for user attributes
5. **Authentication**: Implement WebAuthn for passwordless login

### For Wallet Developers

1. **DID Resolution**: Implement DID resolution for identity verification
2. **Credential Management**: Build UI for credential viewing and sharing
3. **WebAuthn Integration**: Support device-based authentication
4. **Key Recovery**: Implement recovery methods for lost devices

## Technical Architecture

### Gasless Transaction Processing

The gasless WebAuthn registration system uses a sophisticated ante handler chain to selectively bypass fees and signature verification while maintaining security:

#### Ante Handler Chain Order
1. **WebAuthnBypassDecorator**: Identifies and validates WebAuthn registration transactions
2. **Standard Cosmos Decorators**: SetUpContext, ValidateBasic, TxTimeout, etc.
3. **WebAuthnGaslessDecorator**: Performs credential validation and account creation
4. **ConditionalFeeDecorator**: Skips fee deduction for gasless transactions
5. **ConditionalSignatureDecorator**: Bypasses signature verification when appropriate

#### Security Validation Pipeline
```go
// 1. Message Type Validation
if msg, ok := msgs[0].(*didtypes.MsgRegisterWebAuthnCredential); ok {
    // 2. Credential Structure Validation
    if err := msg.WebauthnCredential.ValidateStructure(); err != nil {
        return ctx, err
    }
    
    // 3. Credential Uniqueness Check
    if k.HasExistingCredential(ctx, msg.WebauthnCredential.CredentialId) {
        return ctx, errorsmod.Wrap(errortypes.ErrInvalidRequest, "credential already registered")
    }
    
    // 4. Account Creation/Verification
    account := k.accountKeeper.GetAccount(ctx, controllerAddr)
    if account == nil {
        account = k.accountKeeper.NewAccountWithAddress(ctx, controllerAddr)
        k.accountKeeper.SetAccount(ctx, account)
    }
    
    // 5. Context Marking for Downstream Decorators
    ctx = ctx.WithValue("webauthn_gasless", true)
}
```

### WebAuthn Credential Validation

The system performs multi-layer validation to ensure credential authenticity:

#### Layer 1: Structural Validation
- Required fields presence (credential_id, public_key, algorithm)
- Algorithm support verification (ES256: -7, RS256: -257)
- Base64url encoding validation
- Raw ID and credential ID consistency

#### Layer 2: Cryptographic Validation
- Public key format verification (DER encoding)
- Algorithm-specific key parsing (ECDSA/RSA)
- Key size and curve parameter validation

#### Layer 3: Attestation Validation (Optional)
- Client data JSON parsing and validation
- Challenge verification against provided value
- Origin validation for security context
- Attestation object size and format checks

## Events

The DID module emits comprehensive typed events for all state-changing operations. These events provide a detailed audit trail and enable efficient tracking of DID-related activities.

### Event Types

#### 1. EventDIDCreated
- **Emitted**: When a new DID is created
- **Fields**:
  - `did`: Unique DID identifier
  - `creator`: Address of the DID creator
  - `public_keys`: List of public keys added
  - `services`: List of services added
  - `created_at`: Timestamp of DID creation
  - `block_height`: Block number of creation

#### 2. EventDIDUpdated
- **Emitted**: When a DID document is modified
- **Fields**:
  - `did`: Unique DID identifier
  - `updater`: Address performing the update
  - `fields_updated`: List of fields that were changed
  - `updated_at`: Timestamp of update
  - `block_height`: Block number of update

#### 3. EventDIDDeactivated
- **Emitted**: When a DID is deactivated
- **Fields**:
  - `did`: Unique DID identifier
  - `deactivator`: Address performing deactivation
  - `block_height`: Block number of deactivation

#### 4. EventVerificationMethodAdded
- **Emitted**: When a verification method is added to a DID
- **Fields**:
  - `did`: Target DID
  - `method_id`: Unique identifier for the method
  - `key_type`: Type of verification key
  - `public_key`: Base64 encoded public key
  - `block_height`: Block number of addition

#### 5. EventVerificationMethodRemoved
- **Emitted**: When a verification method is removed from a DID
- **Fields**:
  - `did`: Target DID
  - `method_id`: Identifier of removed method
  - `block_height`: Block number of removal

#### 6. EventServiceAdded
- **Emitted**: When a service is added to a DID
- **Fields**:
  - `did`: Target DID
  - `service_id`: Unique service identifier
  - `type`: Service type
  - `endpoint`: Service endpoint URL
  - `block_height`: Block number of addition

#### 7. EventServiceRemoved
- **Emitted**: When a service is removed from a DID
- **Fields**:
  - `did`: Target DID
  - `service_id`: Identifier of removed service
  - `block_height`: Block number of removal

#### 8. EventWebAuthnRegistered
- **Emitted**: When a WebAuthn credential is registered
- **Fields**:
  - `did`: DID associated with credential
  - `credential_id`: Unique credential identifier
  - `attestation_type`: Type of WebAuthn attestation
  - `block_height`: Block number of registration

#### 9. EventExternalWalletLinked
- **Emitted**: When an external wallet is linked to a DID
- **Fields**:
  - `did`: Target DID
  - `wallet_type`: Type of wallet (e.g., "ethereum", "cosmos")
  - `wallet_address`: Address of linked wallet
  - `block_height`: Block number of linking

#### 10. EventCredentialIssued
- **Emitted**: When a verifiable credential is issued
- **Fields**:
  - `credential_id`: Unique credential identifier
  - `issuer`: DID of credential issuer
  - `subject`: DID of credential subject
  - `type`: Credential type(s)
  - `issued_at`: Timestamp of issuance
  - `block_height`: Block number of issuance

#### 11. EventCredentialRevoked
- **Emitted**: When a credential is revoked
- **Fields**:
  - `credential_id`: Unique credential identifier
  - `revoker`: DID of credential revoker
  - `reason`: Reason for revocation
  - `revoked_at`: Timestamp of revocation
  - `block_height`: Block number of revocation

### Event Indexing and Querying

Events can be queried and filtered using CometBFT WebSocket or standard blockchain explorers. Example queries:

```bash
# Query all DID creation events
tm.event='Tx' AND did.v1.EventDIDCreated.did EXISTS

# Query events by creator
message.sender='idx1...' AND tx.height>=1000

# Subscribe to specific DID events
did.v1.EventDIDCreated.did='did:sonr:test123'
```

### Best Practices for Event Consumers

1. **Indexing**: Configure CometBFT event indexing for comprehensive tracking
2. **Performance**: Use efficient query strategies
3. **Replay Handling**: Implement mechanisms to handle event replay scenarios
4. **Error Resilience**: Design consumers to handle missing or out-of-order events

## Security Considerations

1. **Proof Requirements**: All DID operations require cryptographic proofs
2. **Controller Authority**: Only controllers can update DID documents
3. **Signature Verification**: Multi-algorithm signature verification with fallback protection
4. **Deactivation Security**: Deactivated DIDs cannot be used for signature verification
5. **WebAuthn Security**: Origin validation and challenge verification
6. **Credential Privacy**: Selective disclosure of credential attributes
7. **Key Rotation**: Regular rotation of verification methods
8. **Algorithm Diversity**: Support for multiple cryptographic algorithms prevents single points of failure
9. **Gasless Security**: Comprehensive validation prevents abuse of gasless registration
10. **Credential Uniqueness**: Global enforcement prevents credential reuse attacks
11. **Limited Transaction Scope**: Gasless processing only applies to specific message types

## Supported Standards

- **W3C DID Core**: Full compliance with DID specification
- **W3C Verifiable Credentials**: Standard credential format
- **WebAuthn**: W3C Web Authentication specification
- **JSON-LD**: Linked data format for DIDs
- **JWK/JWS**: JSON Web Key and Signature standards

## Building and Testing

### Running Tests

```bash
# Run unit tests
make -C x/did test

# Run tests with race detection
make -C x/did test-race

# Generate coverage report
make -C x/did test-cover

# Run benchmarks
make -C x/did benchmark
```

## Future Enhancements

- **DID Resolution Methods**: Support for off-chain resolution
- **Privacy Features**: Zero-knowledge proofs for credentials
- **Key Recovery**: Social recovery mechanisms
- **Cross-Chain DIDs**: Interoperability with other DID methods
- **Advanced Credentials**: Support for complex credential schemas
