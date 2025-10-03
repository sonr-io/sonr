## Unreleased

### Feat

- integrate Dexie.js with vault plugin for account-based database persistence (#272)
- Motor WASM Service Worker - Payment Gateway & OIDC Authorization (#263)
- enable secure data storage with vault plugin
- enhance build process with gum logging for better visibility

### Fix

- **devops**: Merge flow
- **devops**: bump package version sync
- cz prehook
- CI and Docker Configuration Across Monorepo (#266)
- update go dependencies for compatibility and security

## v0.13.16 (2025-09-11)

### Feat

- implement end-to-end WebAuthn account registration with DID creation (#261)

## v0.13.15 (2025-09-10)

### Feat

- introduce vault WASM plugin for enhanced data security

## v0.13.14 (2025-09-09)

### Feat

- enable fine-grained, delegable authorization using UCAN

## v0.13.13 (2025-09-09)

## v0.13.12 (2025-09-09)

### Feat

- implement holistic UCAN authorization with OIDC integration (#258)

## v0.13.11 (2025-09-08)

## v0.13.10 (2025-09-08)

### Feat

- integrate PostgreSQL for enhanced data management

## v0.13.9 (2025-09-08)

### Feat

- implement decentralized identity DAO as CosmWasm smart contracts (#254)

## v0.13.8 (2025-09-07)

## v0.13.7 (2025-09-06)

### Feat

- enhance documentation navigation and API exploration

## v0.13.6 (2025-09-06)

## v0.13.5 (2025-09-06)

### Feat

- implement 'Sign in with Sonr' OAuth provider for crypto apps (#250)
- implement comprehensive developer dashboard for Sonr Services (#249)

## v0.13.4 (2025-09-04)

## v0.13.3 (2025-09-04)

### Feat

- consolidate deployment infrastructure with Docker Compose and enhance build targets (#248)

## v0.13.2 (2025-09-04)

### Feat

- implement OpenID Connect provider with WebAuthn in bridge handlers (#244)

## v0.13.1 (2025-09-04)

## v0.13.0 (2025-09-03)

### Fix

- Fix cryptographic implementations and enable disabled tests (#243)

## v0.12.11 (2025-09-03)

### Fix

- WebAuthn attestation and verification implementations (#242)

## v0.12.10 (2025-09-03)

### Feat

- Implement Service module capability system and parameters (#241)

## v0.12.9 (2025-09-03)

### Feat

- Complete DID module WebAuthn and parameter implementation (#240)

## v0.12.8 (2025-09-03)

### Feat

- introduce key rotation events and IPFS status endpoint (#238)

## v0.12.7 (2025-09-02)

### Feat

- implement module parameter validation and defaults (#237)

## v0.12.6 (2025-08-30)

### Feat

- complete remaining event emissions for DID and DWN modules (#235)

## v0.12.5 (2025-08-27)

### Feat

- implement typed Protobuf events for x/did, x/dwn, and x/svc modules (#233)

## v0.12.4 (2025-08-24)

### Feat

- Integrate Interchain Accounts for cross-chain DEX functionality

## v0.12.3 (2025-08-24)

### Feat

- implement ICA Controller system in x/dex module for cross-chain DEX operations (#221)

## v0.12.2 (2025-08-21)

## v0.12.1 (2025-08-20)

### Feat

- replace InterchainTest with Starship-based E2E testing framework (#217)

## v0.12.0 (2025-08-18)

### Feat

- add support for additional elliptic curves and JWK verification (#216)

## v0.11.4 (2025-08-18)

### Feat

- Complete WebAuthn/FIDO2 implementation for passwordless authentication (#215)

## v0.11.3 (2025-08-17)

### Feat

- complete DID keeper implementation with W3C compliance and WebAuthn authentication (#214)

## v0.11.2 (2025-08-16)

### Feat

- transform UI package to shadcn monorepo architecture for uniform styling (#213)

## v0.11.1 (2025-08-15)

### Refactor

- improve secure memory handling for enhanced security (#201)

## v0.11.0 (2025-08-15)

### Feat

- implement comprehensive cryptographic security enhancements (#200)

## v0.10.35 (2025-08-15)

### Feat

- implement comprehensive cryptographic security enhancements (#199)

## v0.10.34 (2025-08-15)

### Feat

- implement CLI commands for wallet module (#198)

## v0.10.33 (2025-08-14)

### Feat

- implement UCAN permission validation for wallet transactions (#197)

## v0.10.32 (2025-08-14)

### Feat

- implement Go client SDK with transaction signing and broadcasting (#196)

## v0.10.31 (2025-08-13)

### Feat

- enable Docker-based testnet execution

## v0.10.30 (2025-08-13)

### Feat

- streamline CI/CD pipeline by removing changeset dependency

## v0.10.29 (2025-08-13)

## v0.10.28 (2025-08-13)

### Feat

- implement monorepo structure with pnpm workspaces and changesets (#189)

## v0.10.27 (2025-08-12)

### Feat

- implement WebAuthn gasless transactions with comprehensive protocol integration (#186)

### Fix

- resolve chain ID validation and encryption key rotation issues (#188)

## v0.10.26 (2025-08-11)

## v0.10.25 (2025-08-11)

### Feat

- implement gasless WebAuthn registration with comprehensive security audit (#182)

## v0.10.24 (2025-08-09)

### Feat

- implement consensus-based encryption for DWN module (#181)

## v0.10.23 (2025-08-09)

### Feat

- enhance init command with VRF keypair generation and SonrContext system (#180)

## v0.10.22 (2025-08-09)

### Refactor

- move interchain tests to (#178)

## v0.10.21 (2025-08-09)

## v0.10.20 (2025-08-08)

### Feat

- refactor Motor WASM plugin as MPC-based UCAN source (#177)

### Refactor

- restructure documentation and navigation for clarity

## v0.10.19 (2025-08-08)

### Feat

- migrate documentation to Mintlify structure (#172)

### Refactor

- migrate x/ucan module to lightweight internal/ucan library (#174)

## v0.10.18 (2025-08-07)

## v0.10.17 (2025-08-06)

### Feat

- implement WebAuthn CLI registration with gasless transactions (#168)

## v0.10.16 (2025-08-06)

### Feat

- implement auto-create DWN vault with comprehensive security improvements (Fixes #153) (#161)

## v0.10.15 (2025-08-05)

### Feat

- migrate Highway service to Echo framework with WebSocket/SSE and JWT auth (#159)

## v0.10.14 (2025-08-05)

### Feat

- complete Highway proxy server implementation with asynq and proto.Actor (#157)

## v0.10.13 (2025-08-05)

### Feat

- refactor x/dwn vaults and introduce gasless transactions (#154)

## v0.10.12 (2025-08-03)

### Feat

- monorepo restructure with internal packages and enhanced CI/CD (#151)

## v0.10.11 (2025-08-03)

### Feat

- streamline testnet configuration

## v0.10.10 (2025-08-03)

### Feat

- enable faucet and explorer for improved testnet accessibility
- streamline deployment workflow and configuration
- streamline testnet configuration for faster iteration
- streamline testnet configuration

### Fix

- K8s deployment config simplified to single node

## v0.10.9 (2025-08-02)

### Feat

- streamline changelog management
- enhance testnet configuration with faucet and explorer settings (#145)

### Fix

- correct dependency for milestone closure (#149)
- ensure Docker containers are always pushed with latest tag
- deploy workflow (#142)

### Refactor

- streamline GitHub Actions workflows for efficiency (#148)

## v0.10.8 (2025-08-01)

### Feat

- optimize deployment workflows and update infrastructure configuration (#141)

## v0.10.7 (2025-08-01)

## v0.10.6 (2025-08-01)

### Feat

- introduce Starship network configurations for devnet and testnet (#139)

## v0.10.5 (2025-07-31)

### Feat

- enhance chain security with pod security context

## v0.10.4 (2025-07-31)

## v0.10.3 (2025-07-31)

### Refactor

- streamline Docker build for enhanced efficiency (#136)

## v0.10.2 (2025-07-31)

### Feat

- streamline starship configuration and local development (#135)

## v0.10.1 (2025-07-31)

### Feat

- optimize CI/CD workflows with smart testing and K8s deployment (#134)
- implement fee grant integration with BasicAllowance (#131)
- implement EVM transaction support in wallet module (#93) (#128)
- implement external wallet linking as DID assertion methods (#127)

### Fix

- update protobuf definitions to reflect wallet chain ID naming (#132)

## v0.10.0 (2025-07-21)

### Feat

- implement secure key management with WASM enclaves (#126)

## v0.9.22 (2025-07-21)

### Feat

- implement vault export/import with IPFS encryption (#125)

## v0.9.21 (2025-07-20)

### Feat

- enhance release automation with dedicated token (#124)

## v0.9.20 (2025-07-20)

### Feat

- Add transaction building framework and streamline release process (#123)

## v0.9.19 (2025-07-20)

### Feat

- Add cross-module keeper integration tests and optimize CI performance (#122)

## v0.9.18 (2025-07-20)

### Refactor

- improve service validation and UCAN integration (#121)

## v0.9.17 (2025-07-20)

### Feat

- Implement ServiceKeeper interface for x/dwn module (#120)

## v0.9.16 (2025-07-20)

### Feat

- Implement UCANKeeper interface for x/dwn module (#119)

## v0.9.15 (2025-07-18)

### Feat

- Implement UCANKeeper interface for x/svc module (#117)

## v0.9.14 (2025-07-18)

### Feat

- Implement DIDKeeper interface for x/svc module (#116)

## v0.9.13 (2025-07-18)

## v0.9.12 (2025-07-18)

## v0.9.11 (2025-07-18)

### Feat

- Implement ServiceKeeper interface methods in x/svc keeper (#113)

## v0.9.10 (2025-07-18)

### Feat

- Implement UCANKeeper interface methods and centralize error handling (#112)

## v0.9.9 (2025-07-18)

### Feat

- **x/did**: Implement VerifyDIDDocumentSignature method with multi-algorithm support (#111)

## v0.9.8 (2025-07-18)

### Feat

- Implement wallet derivation and keeper interface architecture (#110)

## v0.9.7 (2025-07-18)

## v0.9.6 (2025-07-16)

### Feat

- introduce VaultKeeper interface for enhanced modularity (#99)

## v0.9.5 (2025-07-15)

### Feat

- enable DWN vault spawning via query API and add comprehensive tests (#98)

## v0.9.4 (2025-07-14)

### Feat

- Update testnet configuration for DAO governance (#84)

## v0.9.3 (2025-07-09)

### Refactor

- Centralize vault actor system and optimize plugin management (#83)

## v0.9.2 (2025-07-07)

## v0.9.1 (2025-07-07)

### Feat

- automate issue triage with project board integration

### Fix

- use personal access token for project automation

## v0.9.0 (2025-07-06)

### Feat

- Refactor x/dwn module structure and integrate WebAssembly motor client (#81)
- enhance documentation accessibility for LLMs
- enhance site navigation and branding

### Refactor

- restructure app layout and navigation

## v0.8.12 (2025-07-05)

### Feat

- Implement Rybbit analytics and update documentation site styling (#80)

## v0.8.11 (2025-07-04)

### Feat

- Complete shadcn/TemplUI migration from NebulaUI (#79)

## v0.8.10 (2025-07-03)

### Feat

- rename  to

## v0.8.9 (2025-07-03)

## v0.8.8 (2025-07-03)

## v0.8.7 (2025-07-03)

### Feat

- remove TUI dashboard integration from main binary

## v0.8.6 (2025-07-03)

### Feat

- remove TUI dashboard feature

## v0.8.5 (2025-07-03)

## v0.8.4 (2025-07-02)

### Feat

- automate minor version bumps based on milestone completion

## v0.8.3 (2025-07-02)

## v0.8.2 (2025-07-02)

### Feat

- Enhanced documentation landing page with custom branding (#69)

## v0.8.1 (2025-07-02)

## v0.8.0 (2025-07-02)

### Feat

- integrate Trunk.io for code quality and linting (#65)

## v0.7.0 (2025-07-01)

### Feat

- Implement DWN module with enclave signing and DIF specification (#64)

## v0.6.1 (2025-07-01)

### Feat

- es-client protobuf generation (#63)

## v0.6.0 (2025-07-01)

### Feat

- Automated API Reference Generation for Cosmos Modules and Highway REST Service (#62)

### Fix

- Revert NTCharts integration and restore original TUI components (#60)

## v0.5.1 (2025-06-30)

### Feat

- Enhance TUI Dashboard with real-time data visualization and testnet support (#59)

## v0.5.0 (2025-06-30)

### Feat

- Implement x/ucan msgServer handlers with capability templates (#55)

## v0.4.1 (2025-06-30)

### Feat

- Integrate Nebula UI Component Library (#54)

## v0.4.0 (2025-06-29)

### Feat

- W3C DID Controller with WebAuthn support and comprehensive testing improvements (#52)

## v0.3.0 (2025-06-29)

### Feat

- Migrate UCAN capability definitions to x/ucan module (#49)

## v0.2.0 (2025-06-28)

### Feat

- Implement DNS record verification with UCAN delegation for x/svc (#41)

## v0.1.0 (2025-06-28)

### Feat

- Implement Highway Service API Handlers (#39)

## v0.0.23 (2025-06-26)

### Feat

- implement IPFS private network support and enhance CI/CD workflows (#38)
- introduce comprehensive tokenomics documentation (#37)
- enhance documentation with client integration guide (#36)
- introduce research section with whitepapers

## v0.0.22 (2025-06-25)

### Feat

- remove ajv dependency and related code

## v0.0.21 (2025-06-25)

## v0.0.20 (2025-06-25)

### Refactor

- streamline Docker release workflow for improved maintainability

## v0.0.19 (2025-06-25)

### Feat

- improve documentation and build process

## v0.0.18 (2025-06-25)

## v0.0.17 (2025-06-24)

### Feat

- remove auto-generated cosmos API reference pages

## v0.0.16 (2025-06-24)

### Feat

- enhance documentation generation with OpenAPI support
- migrate documentation to Fumadocs
- implement automated build and release process
- remove testnet workflows and data
- implement optimized Docker build workflow
- migrate to docker compose and Makefile

### Fix

- docs
- correct Dockerfile paths to match actual project structure

## v0.0.15 (2025-06-23)

### Feat

- enhance testnet data handling by excluding specific files
- introduce testnet and caddy docker release workflows
- introduce Caddy reverse proxy for enhanced network management

### Refactor

- streamline Docker configurations and builds

## v0.0.14 (2025-06-23)

### Feat

- introduce standalone hway and IPFS services with Docker Compose

## v0.0.13 (2025-06-23)

## v0.0.12 (2025-06-23)

## v0.0.11 (2025-06-23)

## v0.0.10 (2025-06-23)

### Feat

- introduce postgres docker image with extensions

## v0.0.9 (2025-06-23)

### Feat

- authenticate with GitHub Container Registry for image publishing

## v0.0.8 (2025-06-23)

## v0.0.7 (2025-06-23)

### Feat

- introduce multi-image build and push workflow
- enhance Sonr documentation with architecture and component details
- integrate task runner with fzf for improved command execution

### Refactor

- consolidate build tags for improved clarity
- streamline app initialization and service registration
- restructure project and update dependencies
- move server and middleware logic to pkg/server

## v0.0.6 (2025-06-22)

### Feat

- streamline deployment by removing hway proxy

## v0.0.5 (2025-06-22)

### Feat

- optimize build configurations for broader CPU compatibility

## v0.0.4 (2025-06-22)

## v0.0.3 (2025-06-22)

### Fix

- disable CGO for hway builds to improve portability

## v0.0.2 (2025-06-22)

## v0.0.1 (2025-06-22)

### Feat

- automate version bumping and changelog generation
- update go version to 1.24.2 across workflows
- enforce GITHUB_TOKEN and GITHUB_PAT_TOKEN for release
- integrate external APIs for market data retrieval and chain registry updates
- automate release process with goreleaser
- integrate secure enclave and decentralized web node runtime
- Streamline user experience with interactive TUI dashboard
- remove unused enclave
- introduce functional options for configurable vault spawning
- introduce vault options for flexible spawning
- integrate vault management with database
- enhance asset data with verification and market details
- add asset symbol linking to initial data refresh
- add asset symbol linking for improved data association
- enhance request logging with latency and status details
- integrate Claude Opus for enhanced blockchain development
- Update market data retrieval to use external API
- introduce highway service for enhanced user authentication
- add CoinPaprika global market data integration using http_get
- remove direct database access from application
- enable SQLC schema deployment to cloud
- enhance chain and asset data ingestion from Cosmos directory
- initialize market data upon deployment
- schedule and monitor cron jobs
- schedule market and cosmos data updates with pg_cron
- introduce commitizen for standardized commits and releases
- remove enclave wasm
- enable pg_net extension for enhanced network capabilities
- improve postgresql configuration for local development
- implement database migration using goose
- introduce database layer and session management
- introduce taskfile-based build and release processes
- introduce hway service for webauthn
- introduce authorize view
- Integrate CosmWasm VM for enhanced smart contract capabilities
- implement vault refresh functionality
- expose ActorSystem for external access
- remove in-memory cache implementation
- implement enclave actor with Extism plugin for secure key management
- implement enclave actor with Extism runtime for secure operations
- implement enclave actor for secure key management
- enhance .gitignore to exclude build outputs and temporary files
- streamline dev environment setup by removing default init hook
- add support for generating and unlocking enclaves
- introduce enclave build target for WASM
- Introduce enclave WASM runtime for secure operations
- integrate IPFS Kubo v0.35.0 and Boxo v0.31.0
- initialize database schema for core entities
- implement decentralized web node runtime with WASM enclave
- update chain registry for Sonr testnet

### Fix

- database migration issues for initial setup
- improve error message clarity during vault refresh

### Refactor

- centralize version bumping configuration
- streamline project dependencies and configurations for maintainability
- relocate vaults testing directory
- migrate enclave to motr for improved architecture
- improve readability of key string conversion
- move WebAuthn function queries to dedicated file
- Populate cosmos registry with http_get function
- consolidate task configurations for improved maintainability
- relocate deployment configuration to
- streamline enclave build and deployment process
- migrate configuration and metadata fields to  for improved flexibility
- migrate database schema definitions to standard location
- streamline dependencies by removing unused caching library
- simplify vault actor initialization
- enclave init
- replace ExampleData with comprehensive DID state management
- relocate devbox configuration for SQLite support
- replace task-based commands with direct wrangler scripts in devbox.json
- reorganize module imports for clarity
- rename migrations directory for clarity
- rename marketapi package for clarity
- move migrations to sqlite specific directory
- reorganize coins module for improved maintainability
- move proto definitions to separate packages for better organization
