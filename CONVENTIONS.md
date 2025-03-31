# CLAUDE.md - Developer Guidelines for SNRD Codebase

## Build & Test Commands
- `make build` - Build application binary
- `make install` - Install application
- `make test` - Run unit tests
- `make test-unit` - Run all unit tests
- `make test-race` - Run tests with race detection
- `make test-cover` - Run tests with coverage
- `make lint` - Run golangci-lint and formatting checks
- `make format` - Format code using gofumpt
- `make proto-gen` - Generate Protobuf files
- `make testnet` - Setup and run local testnet with IBC

## Code Style Guidelines
- **Imports**: Group by standard library, third-party, then project-specific with blank lines between
- **Error Handling**: Use custom errors from `types/errors.go` with `sdkerrors.Register` and prefix errors with `Err`
- **Naming**: Use camelCase for variables, PascalCase for functions/exported types
- **Organization**: Follow Cosmos SDK module pattern (`x/{module}/`, `types/`, `keeper/`)
- **Documentation**: Use standard Go doc comments, document public APIs thoroughly
- **Types**: Prefer explicit types over interface{}, use pointers for mutable state
- **Design**: Follow dependency injection pattern for module components
- **Formatting**: Run `make format` before committing

Follows standard Go conventions and Cosmos SDK patterns with clear organization and consistent error handling.