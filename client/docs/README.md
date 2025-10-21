# Sonr API Documentation

This directory contains the OpenAPI/Swagger documentation for the Sonr blockchain API.

## Structure

- `config.json` - Configuration for swagger-combine, defines which module swagger files to merge
- `swagger-ui/swagger.yaml` - The unified OpenAPI specification for all Sonr modules

## Generation Process

The OpenAPI documentation is generated following the Cosmos SDK standard approach:

1. **Proto to Swagger**: `buf generate` creates individual swagger JSON files for each module (DID, DWN, Service, DEX) in a temporary directory
2. **Combine**: `swagger-combine` merges all module specifications into a single `swagger.yaml` file
3. **Cleanup**: Temporary files are removed

### Generate Documentation

```bash
# From the repo root
make swagger-gen

# Or from the proto directory
cd proto && make swagger-gen
```

The generated `swagger.yaml` file can be:
- Viewed in Swagger UI
- Used to generate client SDKs
- Embedded in the binary for serving via the API server
- Published to API documentation platforms

## Module Specifications

The following Sonr modules are included:

- **DID Module** (`sonr/did/v1`) - Decentralized Identity management
- **DWN Module** (`sonr/dwn/v1`) - Decentralized Web Node operations
- **Service Module** (`sonr/svc/v1`) - Service registration and management
- **DEX Module** (`sonr/dex/v1`) - Decentralized exchange functionality

Each module contributes both query (read) and transaction (write) endpoints to the unified API specification.

## Operation ID Conflicts

To prevent naming conflicts when combining module specifications, the `config.json` file renames common operation IDs with module-specific prefixes:

- `Params` → `DIDParams`, `DWNParams`, `ServiceParams`, `DEXParams`

This ensures all operations have unique identifiers in the combined specification.

## Serving the API Documentation

The swagger.yaml can be embedded in the Sonr binary and served via the API server. To enable:

```toml
# ~/.snrd/config/app.toml
[api]
enable = true
swagger = true
```

Then access the swagger UI at: `http://localhost:1317/swagger/`
