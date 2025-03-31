# `x/svc`

The svc module is responsible for managing the registration and authorization of services within the Sonr ecosystem. It provides a secure and verifiable mechanism for registering and authorizing services using Decentralized Identifiers (DIDs) and now incorporates UCAN (User Controlled Authorization Networks) for enhanced authorization capabilities.

## Concepts

- **Service**: A decentralized svc on the Sonr Blockchain with properties such as ID, authority, origin, name, description, category, tags, and expiry height.
- **Profile**: Represents a DID alias with properties like ID, subject, origin, and controller.
- **Metadata**: Contains information about a svc, including name, description, category, icon, and tags.
- **UCAN Authorization**: The module utilizes UCANs for a decentralized and user-centric authorization mechanism.

### Dependencies

- [x/did](https://github.com/sonr-io/snrd/tree/master/x/did)
- [x/group](https://github.com/sonr-io/snrd/tree/master/x/group)
- [x/nft](https://github.com/sonr-io/snrd/tree/master/x/nft)

## State

The module uses the following state structures:

### Metadata

Stores information about services:

- Primary key: `id` (auto-increment)
- Unique index: `origin`
- Fields: id, origin, name, description, category, icon (URI), tags

### Profile

Stores DID alias information:

- Primary key: `id`
- Unique index: `subject,origin`
- Fields: id, subject, origin, controller

## Messages

### MsgUpdateParams

Updates the module parameters, including UCAN-related parameters. Can only be executed by the governance account.

### MsgRegisterService

Registers a new svc on the blockchain. Requires a valid TXT record in DNS for the origin and may be subject to UCAN authorization checks.

## Params

The module has the following parameters:

- `categories`: List of allowed svc categories
- `types`: List of allowed svc types
- `UcanPermissions`: Specifies the required UCAN permissions for various actions within the module, such as registering a service.

## Query

The module provides the following query:

### Params

Retrieves all parameters of the module, including UCAN-related parameters.

## Client

### gRPC

The module provides a gRPC Query svc with the following RPC:

- `Params`: Get all parameters of the module, including UCAN-related parameters.

### CLI

(TODO: Add CLI commands for interacting with the module)

## Events

(TODO: List and describe event tags used by the module, including those related to UCAN authorization)

## UCAN Authorization

This module utilizes UCAN (User Controlled Authorization Networks) to provide a decentralized and user-centric authorization mechanism. UCANs are self-contained authorization tokens that allow users to delegate specific capabilities to other entities without relying on a central authority.

### UCAN Integration

- The module parameters include a `UcanPermissions` field that defines the default UCAN permissions required for actions within the module.
- Message handlers in the `MsgServer` perform UCAN authorization checks by:
  - Retrieving the UCAN permissions from the context (injected by a middleware).
  - Retrieving the required UCAN permissions from the module parameters.
  - Verifying that the provided UCAN permissions satisfy the required permissions.
- A dedicated middleware is responsible for:
  - Parsing incoming requests for UCAN tokens.
  - Verifying UCAN token signatures and validity.
  - Extracting UCAN permissions.
  - Injecting UCAN permissions into the context.
- UCAN verification logic involves:
  - Checking UCAN token signatures against the issuer's public key (resolved via the `x/did` module).
  - Validating token expiration and other constraints.
  - Parsing token capabilities and extracting relevant permissions.

## Future Improvements

- Implement svc discovery mechanisms
- Add support for svc reputation and rating systems
- Enhance svc metadata with more detailed information
- Implement svc update and deactivation functionality

## Tests

(TODO: Add acceptance tests for the module)

## Appendix

This module is part of the Sonr blockchain project and interacts with other modules such as DID and NFT modules to provide a comprehensive decentralized svc ecosystem.
