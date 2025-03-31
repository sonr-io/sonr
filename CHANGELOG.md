## v0.6.4 (2025-03-27)

### Feat

- enable GitHub Actions to trigger releases
- enhance development environment with ripgrep and gh-milestone (#1252)
- enhance release process with Doppler integration (#1251)

## v0.6.4 (2025-03-27)

### Feat

- enhance release process with dynamic version display and streamlined automation
- enhance release process with real-time version feedback
- improve release process with interactive feedback
- enhance Docker image publishing with piped credentials
- streamline release process with Taskfile and conventional commits
- streamline development and release workflow
- streamline release process with improved automation
- consolidate release notes
- enhance goreleaser output with title for better UX
- streamline development and release processes
- streamline development and release processes
- streamline release process and update project metadata
- upgrade to go 1.24 and align binary name
- streamline build process by removing release target
- streamline build process and configuration
- streamline release process and update project metadata
- enable commit-less version bumping
- streamline release process with Devbox
- integrate Doppler for secrets management
- introduce devbox for consistent development environment
- automate release process with conventional commits
- streamline app initialization and dependency management
- enable task management with Taskfile
- introduce Taskfile automation for build and deployment
- **deploy**: add hdPath and coinType to sonr-1 chain config
- add go dependency and enhance devbox environment variables
- implement HTMX for dynamic updates
- add CI/CD infrastructure for development and deployment

### Fix

- Refactor crypto
- Deploy config
- resolve minor formatting inconsistencies in Dockerfile and accumulator.go
- correct typo in workflow name

### Refactor

- streamline release process and simplify dev environment
- streamline release process by removing redundant version display
- streamline release process with Taskfile
- rename project from onsonr/sonr to sonr-io/snrd (#1249)
- remove evmos chain from deployment configuration
- simplify deployment infrastructure using starship
- remove unnecessary Caddyfile
- remove process-compose and individual start scripts
- remove devnet and testnet network configurations

## v0.6.3 (2025-01-06)

### Feat

- Add release date to workflow environment
- automate release creation and deployment on tag push

### Fix

- revert version number to 0.6.2
- checkout main branch before updating version
- Specify main branch and fetch tags in merge workflow

### Refactor

- Rename workflow file for clarity

## v0.6.2 (2025-01-06)

### Refactor

- simplify release process and remove unnecessary tasks
- remove devcontainer configuration

## v0.6.1 (2024-12-30)

### Feat

- convert highway sink schema to SQLite
- **scopes**: add Web Authentication API documentation
- add install script for Sonr binaries

### Fix

- Handle only 500 errors in ErrorHandler
- Return only on 500 errors in error handler
- update sqlc queries to match schema and use SQLite conventions
- Rename DecodeWasmContext to decodeWasmContext
- resolve lint errors in WASMMiddleware function
- prevent automatic version bumps from dependabot
- correct merge workflow to increment patch version

### Refactor

- remove outdated scopes
- remove unnecessary seed message
- Update sqlc queries to match schema and sqlite conventions
- move gateway and vault components to new locations
- **scopes**: simplify scopes.json structure
- **scopes**: rename and reorganize scopes
- **api**: remove unused Allocate RPC from Query service
- remove postgresql tasks
- remove unused AI PR body generator step
- simplify Taskfile and remove unused chain configurations

## v0.6.0 (2024-12-24)

### Feat

- Add option to create draft issues from the new-issue script

### Refactor

- rename DID, DWN, and SVC modules to core-dids, core-dwns, and core-svcs respectively
- rename MsgInitialize to MsgSpawn for clarity
- rename scopes in .github/scopes.json for better clarity

## v0.5.28 (2024-12-22)

### Fix

- Sink

### Refactor

- update testnet configuration
- optimize GitHub Actions workflow triggers (#1204)
- Move embed

## v0.1.6 (2024-12-16)

## v0.5.26 (2024-12-13)

## v0.5.25 (2024-12-11)

### Feat

- enable GoReleaser releases on tags and snapshots
- automate release on tag and workflow dispatch

### Fix

- Correct regular expression for version tags in release workflow

## v0.5.24 (2024-12-11)

### Feat

- prevent duplicate releases

## v0.5.23 (2024-12-11)

### Refactor

- rename scheduled release workflow to versioned release
- remove changelog from release artifacts

## v0.5.22 (2024-12-11)

### Feat

- Implement passkey-based authentication and registration flow
- allow manual triggering of deployment workflow
- add start-tui command for interactive mode
- add coin selection and update passkey input in registration form
- add hway command for Sonr DID gateway
- Conditionally install process-compose only if binary not found
- Add process-compose support with custom start and down commands
- implement passkey registration flow
- Improve createProfile form layout with wider max-width and enhanced spacing
- improve index page UI with new navigation buttons and remove redundant settings buttons
- Make input rows responsive with grid layout for mobile and desktop
- enhance index page with additional settings buttons and style adjustments
- implement passkey-based authentication
- add support for Cloudsmith releases
- add go dependency and enhance devbox environment variables
- update create profile form placeholders and handle
- add DID-based authentication middleware
- Add validation for human verification slider sum in CreateProfile form
- implement passkey registration flow
- Update WebAuthn credential handling with modern browser standards
- Streamline passkey registration with automatic form submission
- Add credential parsing and logging in register finish handler
- Add credential details row with icon after passkey creation
- Add form validation for passkey credential input
- implement passkey registration flow
- Add hidden input to store passkey credential data for form submission
- add CI workflow for deploying network
- add hway binary support and Homebrew formula
- remove username from passkey creation
- implement passkey registration flow
- add passkey creation functionality
- add CNAME for onsonr.dev domain

### Fix

- use Unix domain sockets for devnet processes
- correct workflow name and improve devnet deployment process
- correct title of profile creation page
- rename devbox start script to up and remove stop script
- Consolidate archive configuration and add LICENSE file
- Improve cross-browser passkey credential handling and encoding
- Remove commented-out code in passkey registration script
- remove line-clamp from tailwind config
- remove unnecessary background and restart settings from process-compose.yaml
- suppress process-compose server output and log to file

### Refactor

- remove unnecessary git fetch step in deploy workflow
- remove obsolete interchain test dependencies
- update index views to use new nebula components
- move Wasm related code to pkg/common/wasm
- migrate config package to pkg directory
- migrate to new configuration system and model definitions
- move session package to pkg directory
- Refactor registration forms to use UI components
- move gateway config to vault package
- improve command line flag descriptions and variable names
- refactor hway command to use echo framework for server
- Update root command to load EnvImpl from cobra flags
- Modify command flags and environment loading logic in cmds.go
- improve build process and move process-compose.yaml
- remove unused devbox.json and related configurations
- Improve mobile layout responsiveness for Rows and Columns components
- Remove max-w-fit from Rows component
- replace session package with context package
- rename database initialization function
- move session management to dedicated database module
- remove unused UI components related to wallet and index pages
- consolidate handlers into single files
- move gateway and vault packages to internal directory
- Move registration form components to dedicated directory
- remove unused devbox package
- remove devbox configuration
- move vault package to app directory
- improve code structure within gateway package
- move gateway package to app directory
- move vault package internal components to root
- migrate layout imports to common styles package
- Move form templates and styles to common directory
- consolidate authentication and DID handling logic
- Improve WebAuthn credential handling and validation in register finish route
- remove profile card component
- Simplify passkey registration UI and move profile component inline
- Update credential logging with transport and ID type
- Update register handler to use protocol.CredentialDescriptor struct
- Update credential handling to use protocol.CredentialDescriptor
- improve profile card styling and functionality
- Simplify session management and browser information extraction
- Update PeerInfo to extract and store comprehensive device information
- improve address display in property details
- remove unused documentation generation script
- replace sonr/pkg/styles/layout with nebula/ui/layout
- migrate UI components to nebula module
- improve scopes.json structure and update scripts for better usability

## v0.5.20 (2024-12-07)

### Refactor

- simplify CI workflow by removing redundant asset publishing steps

## v0.5.19 (2024-12-06)

### Feat

- add support for parent field and resources list in Capability message
- add fast reflection methods for Capability and Resource
- add gum package and update devbox configuration
- add new button components and layout improvements

### Fix

- adjust fullscreen modal close button margin
- update devbox lockfile
- resolve rendering issue in login modal

### Refactor

- rename accaddr package to address
- Update Credential table to match WebAuthn Credential Descriptor
- Deployment setup
- migrate build system from Taskfile to Makefile
- rename Assertion to Account and update related code
- remove unused TUI components
- Move IPFS interaction functions to common package
- remove dependency on DWN.pkl
- remove unused dependencies and simplify module imports
- Rename x/vault -> x/dwn and x/service -> x/svc
- move resolver formatter to services package
- remove web documentation
- update devbox configuration and scripts
- rename layout component to root
- refactor authentication pages into their own modules
- update templ version to v0.2.778 and remove unused air config
- move signer implementation to mpc package

## v0.5.18 (2024-11-06)

## v0.5.17 (2024-11-05)

### Feat

- add remote client constructor
- add avatar image components
- add SVG CDN Illustrations to marketing architecture
- **marketing**: refactor marketing page components
- Refactor intro video component to use a proper script template
- Move Alpine.js script initialization to separate component
- Add intro video modal component
- add homepage architecture section
- add Hero section component with stats and buttons
- **css**: add new utility classes for group hover
- implement authentication register finish endpoint
- add controller creation step to allocate
- Update service module README based on protobuf files
- Update x/macaroon/README.md with details from protobuf files
- update Vault README with details from proto files

### Fix

- update file paths in error messages
- update intro video modal script

### Refactor

- update marketing section architecture
- change verification table id
- **proto**: remove macaroon proto
- rename ValidateBasic to Validate
- rename session cookie key
- remove unused sync-initial endpoint
- remove formatter.go from service module

## v0.5.16 (2024-10-21)

### Fix

- include assets generation in wasm build

## v0.5.15 (2024-10-21)

## v0.5.14 (2024-10-21)

### Refactor

- remove StakingKeeper dependency from GlobalFeeDecorator

## v0.5.13 (2024-10-21)

### Feat

- add custom secp256k1 pubkey

### Refactor

- update gRPC client to use new request types
- use RawPublicKey instead of PublicKey in macaroon issuer
- improve error handling in DID module

## v0.5.12 (2024-10-18)

### Feat

- add User-Agent and Platform to session
- introduce AuthState enum for authentication state

### Fix

- **version**: revert version bump to 0.5.11
- **version**: update version to 0.5.12

### Refactor

- remove dependency on proto change detection
- update asset publishing configuration

## v0.5.11 (2024-10-10)

### Feat

- nebula assets served from CDN
- use CDN for nebula frontend assets
- add static hero section content to homepage
- add wrangler scripts for development, build, and deployment
- remove build configuration
- move gateway web code to dedicated directory
- add PubKey fast reflection
- **macaroon**: add transaction allowlist/denylist caveats
- add PR labeler
- **devbox**: remove hway start command
- add GitHub Actions workflow for running tests
- add workflow for deploying Hway to Cloudflare Workers
- Publish configs to R2
- integrate nebula UI with worker-assets-gen
- extract reusable layout components
- Implement service worker for IPFS vault
- implement CDN support for assets
- add payment method support
- add support for public key management
- add ModalForm component
- add LoginStart and RegisterStart routes
- implement authentication views
- add json tags to config structs
- implement templ forms for consent privacy, credential assert, credential register, and profile details
- **vault**: introduce assembly of the initial vault
- add client logos to homepage
- add tailwind utility classes
- implement new profile card component

### Fix

- Correct source directory for asset publishing
- install dependencies before nebula build
- update Schema service to use new API endpoint
- fix broken logo image path

### Refactor

- remove unnecessary branch configuration from scheduled release workflow
- update dwn configuration generation import path
- use nebula/routes instead of nebula/global
- move index template to routes package
- remove cdn package and move assets to global styles
- move nebula assets to hway build directory
- remove docker build and deployment
- rename internal/session package to internal/ctx
- remove unused fields from
- rename PR_TEMPLATE to PULL_REQUEST_TEMPLATE
- remove devbox.json init hook
- rename sonrd dockerfile to Dockerfile
- remove unused dependency
- rename 'global/cdn' to 'assets'
- move CDN assets to separate folder
- move Pkl module definitions to dedicated package
- move CDN assets to js/ folder
- remove unused component templates
- move ui components to global
- move view handlers to router package

## v0.5.10 (2024-10-07)

### Feat

- **blocks**: remove button component

## v0.5.9 (2024-10-06)

### Feat

- add Motr support
- update UIUX PKL to utilize optional fields

### Fix

- Update source directory for asset publishing

## v0.5.8 (2024-10-04)

### Refactor

- Remove unused logs configuration

## v0.5.7 (2024-10-04)

### Feat

- **devbox**: use process-compose for testnet services
- remove motr.mjs dependency
- add markdown rendering to issue templates
- update issue templates for better clarity
- add issue templates for tracking and task issues
- add issue templates for bug report and tracking

### Refactor

- update issue template headings
- rename bug-report issue template to bug

## v0.5.6 (2024-10-03)

### Feat

- introduce docker-compose based setup
- add hway and sonr processes to dev environment

## v0.5.5 (2024-10-03)

### Feat

- add rudimentary DidController table
- update home section with new features
- introduce Home model and refactor views
- **nebula**: create Home model for home page

### Refactor

- reorganize pkl files for better separation of concerns
- rename msg_server_test.go to rpc_test.go

## v0.5.4 (2024-10-02)

## v0.5.3 (2024-10-02)

### Fix

- remove unnecessary telegram message template

## v0.5.2 (2024-10-02)

### Feat

- **service**: integrate group module (#1104)

### Refactor

- revert version bump to 0.5.1

## v0.5.1 (2024-10-02)

### Refactor

- move Motr API to state package

## v0.5.0 (2024-10-02)

### Feat

- allow multiple macaroons with the same id

## v0.4.5 (2024-10-02)

### Feat

- **release**: add docker images for sonrd and motr
- update homepage with new visual design
- add DID to vault genesis schema
- add video component
- add video component
- add hx-get attribute to primary button in hero section

### Fix

- use correct secret for docker login
- **layout**: add missing favicon
- **hero**: Use hx-swap for primary button to prevent flicker

### Refactor

- use single GITHUB_TOKEN for release workflow
- update workflow variables

## v0.4.2 (2024-10-01)

### Refactor

- use single GITHUB_TOKEN for release workflow

## v0.4.1 (2024-10-01)

### Feat

- Implement session management
- allow manual release triggers
- add Input and RegistrationForm models
- add new utility classes
- add login and registration pages
- add tailwindcss utilities
- add support for ARM64 architecture
- add DWN resolver field
- add stats section to homepage
- implement hero section using Pkl

### Fix

- **version**: update version number to 0.4.0
- update release workflow to use latest tag
- **versioning**: revert version to 0.9.0
- **cta**: Fix typo in CTA title
- change bento section title to reflect security focus
- adjust hero image dimensions
- **Input**: Change type from to
- update hero image height in config.pkl

### Refactor

- move home page sections to home package
- rename motrd to motr
- update hero image dimensions
- move nebula configuration to static file

## v0.4.0 (2024-09-30)

### Feat

- add PKL schema for message formats
- add Homebrew tap for sonr
- update release workflow to use latest tag
- **dwn**: add wasm build for dwn
- add macaroon and oracle genesis states
- add scheduled binary release workflow
- introduce process-compose for process management
- add counter animation to hero section
- add registration page

### Fix

- Enable scheduled release workflow

### Refactor

- remove old changelog entries
- rename buf-publish.yml to publish-assets.yml
- remove unused field from
- remove unnecessary checkout in scheduled-release workflow
- rename build ID to sonr
- remove unnecessary release existence check
- move dwn wasm build to pkg directory

## v0.3.1 (2024-09-29)

### Refactor

- move nebula/pages to pkg/nebula/pages

## v0.3.0 (2024-09-29)

### Feat

- add buf.lock for proto definitions

### Fix

- remove unused linting rules
- update proto breaking check target to master branch

### Refactor

- remove unused lock files and configurations

## v0.2.0 (2024-09-29)

### Feat

- disable goreleaser workflow
- update workflows to include master branch
- remove global style declaration
- **oracle**: add oracle module
- optimize IPFS configuration for better performance
- add local IPFS bootstrap script and refactor devbox config
- add AllocateVault HTTP endpoint
- add WebAuthn credential management functionality
- remove unused coins interface
- remove global integrity proof from genesis state
- add vault module
- enable buf.build publishing on master and develop branches
- add Gitflow workflow for syncing branches
- add automated production release workflow
- **ui**: implement profile page
- add automated production release workflow
- **did**: remove unused proto files
- add enums.pulsar.go file for PermissionScope enum (#4)
- add initial DID implementation
- remove builder interface
- add basic UI for block explorer
- add Usage: pkl [OPTIONS] COMMAND [ARGS]...
- use SQLite embedded driver
- add DID method for each coin
- Expand KeyType enum and update KeyInfo message in genesis.proto
- Add whitelisted key types to genesis params
- Add DID grants protobuf definition
- Add fields to KeyInfo struct to distinguish CBOR and standard blockchain key types
- Add new message types for AssetInfo, ChainInfo, Endpoint, ExplorerInfo, FeeInfo, and KeyInfo
- run sonr-node container in testnet network and make network external
- Add docker-compose.yaml file to start a Sonr testnet node
- configure Sonr testnet environment
- Update Dockerfile to start and run a testnet
- add Equal methods for AssetInfo and ChainInfo types
- Add ProveWitness and SyncVault RPCs
- Add MsgRegisterService to handle service registration
- Add MsgRegisterService to handle service registration
- add enums.pulsar.go file for PermissionScope enum

### Fix

- ensure go version is up-to-date
- use GITHUB_TOKEN for version bump workflow
- update account table interface to use address, chain and network
- **ci**: update docker vm release workflow with new token
- use mnemonic phrases for test account keys
- reduce motr proxy shutdown timeout
- **nebula**: use bunx for tailwindcss build
- **proto**: update protobuf message index numbers
- **ante**: reduce POA rate floor and ceiling
- Update proc_list_width in mprocs.yaml
- Add service to database when registering
- pin added did documents to local ipfs node
- remove extra spaces in typeUrl
- **release**: remove unnecessary quotes in tag pattern
- remove unused imports and simplify KeyInfo message
- bind node ports to localhost
- Update docker-compose network name to dokploy-network
- Update network name to dokploy
- remove unused port mapping
- Update docker-compose.yaml to use correct volume path
- update docker-compose volume name
- Update docker-compose.yaml to use shell directly for sonrd command
- replace "sh" with "/bin/sh" in docker-compose.yaml command
- Update runner image dependencies for debian-11
- **deps**: update golang image to 1.21
- **chains**: update nomic chain build target
- Remove unused `Meta` message from `genesis.proto`
- Add ProveWitness and SyncVault RPCs

### Refactor

- adjust source directory for config files (#1102)
- Use actions/checkout@v4
- remove unused master branch from CI workflow
- rename github token secret
- remove unnecessary x-cloak styles
- optimize oracle genesis proto
- remove unused code related to whitelisted assets
- update buf publish source directory
- adjust devbox configuration to reflect nebula changes
- rename msg_server.go to rpc.go
- remove devbox integration
- move dwn package to app/config
- move configuration files to app directory
- extract root command creation to separate file
- move ipfs setup to function
- remove unnecessary proxy config
- rename script to
- move DWN proxy server logic to separate file
- use htmx instead of dwn for vault client
- remove unused environment variables
- simplify verification method structure
- use staking keeper in DID keeper
- remove unused dependencies
- remove unused image building workflow
- add field to
- Update KeyKind Enum to have proper naming conventions
- Update `DIDNamespace` to have proper naming convention
- expose ports directly in docker-compose
- remove unused port mappings
- streamline script execution
- use CMD instead of ENTRYPOINT in Dockerfile
- **deps**: Upgrade Debian base image to 11
- Simplify the types and properties to keep a consistent structure for the blockchain
- remove PERMISSION_SCOPE_IDENTIFIERS_ENS enum value
