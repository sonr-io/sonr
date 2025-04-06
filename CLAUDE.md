# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands
- `make build` - Build application binary
- `make install` - Install application
- `make proto-gen` - Generate Protobuf files
- `make mod-tidy` - Run go mod tidy

## Test Commands
- `make test` - Run unit tests
- `make test-unit` - Run all unit tests 
- `make test-race` - Run tests with race detection
- For single test: `go test -v -run TestName ./path/to/package`

## Lint Commands
- `make lint` - Run golangci-lint and formatting checks
- `make format` - Format code using gofumpt, misspell and gci

## Code Style Guidelines
- **Imports**: Group by standard library, third-party, then project-specific with blank lines
- **Error Handling**: Use custom errors from `types/errors.go` with `sdkerrors.Register`
- **Naming**: camelCase for variables, PascalCase for functions/exported types, prefix errors with `Err`
- **Organization**: Follow Cosmos SDK module pattern (`x/{module}/`, `types/`, `keeper/`)
- **Types**: Prefer explicit types over interface{}, use pointers for mutable state
- **Documentation**: Document public APIs thoroughly with standard Go doc comments
- **Formatting**: Run `make format` before committing

This project follows standard Go conventions and Cosmos SDK patterns with clear organization and consistent error handling.