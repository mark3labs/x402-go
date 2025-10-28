# Implementation Plan: Helper Functions and Constants

**Branch**: `003-helpers-constants` | **Date**: 2025-10-28 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-helpers-constants/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement helper functions and constants to simplify x402 client and middleware configuration with USDC payments across 8 chains (4 mainnet + 4 testnet). Provides ChainConfig constants with verified USDC addresses and EIP-3009 domain parameters, plus helpers for creating PaymentRequirement and TokenConfig structs. Uses Go stdlib for decimal-to-atomic conversion and struct validation.

## Technical Context

**Language/Version**: Go 1.25.1  
**Primary Dependencies**: Go standard library (strconv, fmt, encoding/json)  
**Storage**: N/A (constants and pure functions)  
**Testing**: go test with table-driven tests, -race flag for race detection  
**Target Platform**: All Go-supported platforms (Linux, macOS, Windows)  
**Project Type**: Single library (extends existing x402-go package)  
**Performance Goals**: Sub-microsecond helper function execution (no I/O)  
**Constraints**: Stdlib-first approach (no external dependencies), maintain 100% test coverage  
**Scale/Scope**: 8 chain constants, 3 helper functions (~200 LOC), 8 testnet + 8 mainnet configs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. No Unnecessary Documentation | ✅ PASS | Only spec-required docs: research.md, data-model.md, quickstart.md, contracts/ |
| II. Test Coverage Preservation | ✅ PASS | New code will include comprehensive table-driven tests; no existing coverage reduction |
| III. Test-First Development | ✅ PASS | Plan includes test writing before implementation in tasks.md |
| IV. Stdlib-First Approach | ✅ PASS | Uses only stdlib: strconv for parsing, fmt for errors, no external deps |
| V. Code Conciseness | ✅ PASS | Simple constants and helper functions; no unnecessary abstractions |
| VI. Binary Cleanup | ✅ PASS | Library code only; no binaries produced |

## Project Structure

### Documentation (this feature)

```text
specs/003-helpers-constants/
├── plan.md              # This file
├── research.md          # Phase 0 output (COMPLETED)
├── data-model.md        # Phase 1 output (COMPLETED)
├── quickstart.md        # Phase 1 output (COMPLETED)
├── contracts/
│   └── helpers-api.yaml # Phase 1 output (COMPLETED)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
/
├── chains.go            # NEW: ChainConfig constants + helper functions
├── chains_test.go       # NEW: Tests for chains.go
├── types.go             # EXISTING: PaymentRequirement, TokenConfig types
├── errors.go            # EXISTING: Error types
├── evm/                 # EXISTING: EVM signer implementation
├── svm/                 # EXISTING: Solana signer implementation
├── http/                # EXISTING: Client and middleware
├── examples/
│   └── basic/           # NEW: Example using chain constants
└── testdata/            # EXISTING: Test fixtures
```

**Structure Decision**: Go library with flat package structure at root. All constants and helpers exported from main `x402` package (github.com/mark3labs/x402-go). This follows Go conventions for small libraries and avoids import stuttering. New `chains.go` file contains all chain-related code alongside existing types.go, errors.go, etc.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations. All constitution principles satisfied.
