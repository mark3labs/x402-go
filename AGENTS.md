# Agent Development Guidelines for x402-go

## Build & Test Commands
- **Build**: `go build ./...` - Builds all packages
- **Test All**: `go test ./...` - Runs all tests  
- **Test Single**: `go test -run TestName ./path/to/package` - Run specific test
- **Test Coverage**: `go test -cover ./...` - Run tests with coverage
- **Format**: `go fmt ./...` - Format all Go files
- **Lint**: `go vet ./...` - Run Go static analysis

## Code Style & Conventions
- **Package Structure**: Use `cmd/` for executables, `internal/` for private packages, `pkg/` for public libraries
- **Imports**: Group as stdlib, external deps, then internal packages with blank lines between
- **Naming**: Use camelCase for variables/functions, PascalCase for exported items, avoid abbreviations
- **Error Handling**: Always check errors; wrap with context using `fmt.Errorf("context: %w", err)`
- **Comments**: Start with function name for exported functions; use `//` for inline, `/* */` for blocks
- **Testing**: Test files end with `_test.go`; use table-driven tests; mock external dependencies
- **Concurrency**: Prefer channels over mutexes; always handle goroutine lifecycles properly
- **Dependencies**: Use go.mod; run `go mod tidy` after adding/removing deps
- **Project Scripts**: Use `.specify/scripts/bash/` for automation scripts (check-prerequisites.sh, create-new-feature.sh, etc.)

## Module: github.com/mark3labs/x402-go | Go Version: 1.25.1

## Active Technologies
- Go 1.25.1 + Go standard library (net/http, encoding/json, encoding/base64, context) (001-x402-payment-middleware)
- N/A (stateless middleware, nonce tracking delegated to facilitator) (001-x402-payment-middleware)

## Recent Changes
- 001-x402-payment-middleware: Added Go 1.25.1 + Go standard library (net/http, encoding/json, encoding/base64, context)
