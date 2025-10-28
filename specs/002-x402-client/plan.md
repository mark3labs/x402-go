# Implementation Plan: x402 Payment Client

**Branch**: `002-x402-client` | **Date**: 2025-10-28 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-x402-client/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Create a Go stdlib-compatible HTTP client that can automatically sign and send x402 payment responses to access paywalled endpoints. The client supports multiple payment signers (EVM via go-ethereum and SVM via gagliardetto/solana-go), intelligent payment selection based on server requirements and configured priorities, and optional per-transaction spending limits.

## Technical Context

**Language/Version**: Go 1.25.1  
**Primary Dependencies**: 
  - github.com/ethereum/go-ethereum (EVM signing)
  - github.com/gagliardetto/solana-go (SVM signing)
  - Standard library net/http (HTTP client)
**Storage**: N/A (stateless client, no persistence required)  
**Testing**: Go standard testing with table-driven tests and race detection  
**Target Platform**: Cross-platform (Linux/macOS/Windows)
**Project Type**: Go module/library with example CLI  
**Performance Goals**: 
  - Payment selection < 100ms with 10 signers
  - Concurrent request handling for 100+ requests
  - Memory overhead < 1MB per signer
**Constraints**: 
  - Must maintain stdlib http.Client compatibility
  - No on-chain balance checking (trust local state)
  - Payment authorization validity ~60 seconds
**Scale/Scope**: 
  - Support 10+ signers per client
  - Handle 100+ concurrent payment requests
  - Stateless operation (no persistence)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Initial Assessment (Phase 0)

1. **No Unnecessary Documentation** ✅
   - Only creating required specs, plan, and API documentation
   - No preemptive documentation

2. **Test Coverage Preservation** ✅
   - Will implement table-driven tests for all new code
   - Tests will be written before implementation (TDD)

3. **Test-First Development** ✅
   - Tests defined in plan before implementation
   - Each component will have tests written first

4. **Stdlib-First Approach** ✅
   - Using net/http for HTTP client (stdlib)
   - Using encoding/json for serialization (stdlib)
   - Using encoding/base64 for header encoding (stdlib)
   - External deps only for blockchain signing (justified - no stdlib support)

5. **Code Conciseness** ✅
   - Simple interface design following http.Client patterns
   - Minimal abstraction layers
   - Direct implementation without unnecessary indirection

6. **Binary Cleanup** ✅
   - Example binaries will be in gitignored directories
   - No compiled artifacts in repository

**INITIAL GATE STATUS: PASSED**

### Post-Design Assessment (Phase 1)

1. **No Unnecessary Documentation** ✅
   - Created only essential documentation (data model, API contracts, quickstart)
   - All documentation directly supports implementation

2. **Test Coverage Preservation** ✅
   - Test strategy defined for all components
   - Table-driven tests specified in data model
   - Race detection mandated for all tests

3. **Test-First Development** ✅
   - Test patterns documented in quickstart
   - Mock implementations provided for testing
   - Clear test boundaries defined

4. **Stdlib-First Approach** ✅
   - Confirmed stdlib usage for all non-blockchain operations
   - External dependencies justified:
     - go-ethereum: Required for EVM signing (no stdlib alternative)
     - gagliardetto/solana-go: Required for Solana (no stdlib alternative)


5. **Code Conciseness** ✅
   - Clean separation of concerns in package structure
   - Minimal interface definitions
   - No unnecessary abstractions in data model

6. **Binary Cleanup** ✅
   - Example binary location specified (examples/x402demo)
   - .gitignore patterns confirmed

**FINAL GATE STATUS: PASSED** - All constitution principles satisfied after design phase

## Project Structure

### Documentation (this feature)

```text
specs/002-x402-client/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
# Root package - shared x402 logic
x402/
├── types.go             # Shared types (PaymentRequirements, PaymentPayload, etc.)
├── errors.go            # Common error types
├── signer.go            # Signer interface definition
└── selector.go          # Payment selection logic

# HTTP client package
x402/http/
├── client.go            # Main HTTP client with x402 support
├── transport.go         # Custom RoundTripper for payment handling
├── parser.go            # Parse 402 responses and payment requirements
└── builder.go           # Build payment headers

# EVM signer implementation
x402/evm/
├── signer.go            # EVM signer implementation
├── keystore.go          # Keystore file support
└── eip3009.go           # EIP-3009 authorization signing

# SVM signer implementation  
x402/svm/
├── signer.go            # SVM signer implementation
├── keystore.go          # Keystore file support
└── transaction.go       # Solana transaction building

# Example CLI application
examples/x402demo/
└── main.go              # Combined client/server example

# Tests
x402/
└── *_test.go            # Unit tests for each component

x402/http/
└── *_test.go            # HTTP client tests

x402/evm/
└── *_test.go            # EVM signer tests

x402/svm/
└── *_test.go            # SVM signer tests

examples/x402demo/
└── main_test.go         # Integration tests
```

**Structure Decision**: Modular package structure with shared logic in root `x402` package, HTTP-specific code in `x402/http`, and blockchain-specific implementations in separate packages. This follows Go conventions and enables clean separation of concerns while maximizing code reuse.

## Complexity Tracking

> No violations - all constitution principles satisfied

## Phase 0: Research & Technical Decisions

### Research Tasks

1. **EIP-3009 Implementation Details**
   - How to properly sign transferWithAuthorization using go-ethereum
   - Domain separator and typed data construction
   - Nonce generation best practices

2. **Solana Transaction Building**
   - How to construct SPL token transfer instructions
   - Partial transaction signing with gagliardetto/solana-go
   - Fee payer coordination

3. **State Management Strategy**
   - Stateless client operation (no persistence required)
   - Per-transaction limits only (no cumulative budgets)
   - Thread-safe operation without shared mutable state

4. **HTTP Client Integration**
   - Best practices for extending http.Client
   - Custom RoundTripper implementation patterns
   - Header encoding standards

5. **Priority Selection Algorithm**
   - Efficient sorting with multiple criteria
   - Tie-breaking strategies
   - Performance optimization for large signer sets

### Next Steps

After research completion, Phase 1 will generate:
- Detailed data models
- API contracts
- Quickstart guide
- Update agent context

The plan will stop here. Run `/speckit.tasks` after this plan is complete to generate implementation tasks.