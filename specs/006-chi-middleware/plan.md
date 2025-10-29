# Implementation Plan: Chi Middleware for x402 Payment Protocol

**Branch**: `006-chi-middleware` | **Date**: 2025-10-29 | **Spec**: [spec.md](/home/space_cowboy/Workspace/x402-go/specs/006-chi-middleware/spec.md)
**Input**: Feature specification from `/specs/006-chi-middleware/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Create a Chi-compatible middleware adapter for x402 payment gating that follows the same pattern as the existing Gin and PocketBase middleware implementations. The middleware will use Chi's standard `func(http.Handler) http.Handler` signature, which is identical to stdlib, allowing maximum code reuse. Helper functions (parsePaymentHeaderFromRequest, sendPaymentRequired, findMatchingRequirement, addPaymentResponseHeader) will be shared with stdlib middleware via a new internal package to reduce duplication. The implementation will verify payments with the facilitator service, settle them (unless verify-only mode), and store payment details in request context for handler access.

## Technical Context

**Language/Version**: Go 1.25.1  
**Primary Dependencies**: 
- Chi router (github.com/go-chi/chi/v5) - HTTP router and middleware framework
- Existing x402-go core package (github.com/mark3labs/x402-go) - Type definitions and base types
- Existing http/facilitator.go - Payment verification and settlement client
- Go standard library (net/http, encoding/json, encoding/base64, context, log/slog)

**Storage**: N/A (stateless middleware, payment tracking delegated to facilitator)  
**Testing**: Go testing framework (`go test -race ./...`)  
**Target Platform**: Linux server (any platform supporting Go 1.25.1+)  
**Project Type**: Single (Go library package)  
**Performance Goals**: 
- Verification timeout: 5 seconds (hardcoded in FacilitatorClient)
- Settlement timeout: 60 seconds (hardcoded in FacilitatorClient)
- Middleware overhead: minimal (single context copy, header parsing)

**Constraints**: 
- Must match stdlib middleware behavior exactly (verify-then-settle pattern)
- Must use stdlib http.Request/http.ResponseWriter types (Chi compatibility)
- Must share helper functions with stdlib via internal package
- Stateless design - no payment history or nonce tracking

**Scale/Scope**: 
- Adapter package (~200-300 lines based on Gin middleware pattern)
- Reuses 80%+ logic from existing stdlib middleware via shared helpers
- Single middleware constructor function + minimal helper functions

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: No Unnecessary Documentation
- ✅ PASS - Documentation explicitly requested by user for feature implementation
- User requested: "Chi flavored middleware" with specific functionality requirements
- Generated docs serve clear purpose: research.md, data-model.md, contracts/, quickstart.md

### Principle II: Test Coverage Preservation  
- ✅ PASS - New tests required for Chi middleware adapter
- Existing coverage: stdlib middleware_test.go, Gin middleware_test.go
- Plan: Adapt stdlib test scenarios to Chi router (3+ core tests minimum per SC-007)
- Coverage will be measured with `go test -cover ./http/chi/...`

### Principle III: Test-First Development
- ✅ PASS - Tests will be written before implementation
- Development cycle: write test → verify fails → implement → verify passes
- Test scenarios defined in spec (User Stories 1-3, Edge Cases)

### Principle IV: Stdlib-First Approach
- ✅ PASS - Chi uses stdlib http.Handler interface exclusively
- Chi chosen because user requested "Chi flavored middleware"
- Chi dependency already exists in go.mod (user confirmed as assumption)
- Maximum code reuse from stdlib middleware via shared internal package

### Principle V: Code Conciseness
- ✅ PASS - Adapter pattern minimizes code duplication
- Estimated ~200-300 lines (matching Gin middleware size)
- Shares helpers with stdlib (parsePaymentHeaderFromRequest, sendPaymentRequired, findMatchingRequirement, addPaymentResponseHeader)
- No unnecessary abstractions - simple wrapper following Chi's func(http.Handler) http.Handler pattern

### Principle VI: Binary Cleanup
- ✅ PASS - No binaries will be committed
- Chi middleware is library code (no executables)
- Example usage may be added to examples/chi/ (binaries excluded via .gitignore)

### Code Quality Gates
- ✅ All tests must pass before merge
- ✅ Tests run with `-race` flag for race detection
- ✅ Code will pass `go fmt`, `go vet`, and `golangci-lint`
- ✅ Coverage reports reviewed (maintain or exceed existing levels)

### Summary
**ALL GATES PASSED** - Ready to proceed with Phase 0 research.

---

## Constitution Check (Post-Design Re-evaluation)

**Date**: 2025-10-29  
**Phase**: After Phase 1 Design Completion

### Principle I: No Unnecessary Documentation
- ✅ PASS - All generated docs serve explicit purpose:
  - research.md: Resolved technical unknowns, documented decisions
  - data-model.md: Defines entities and data flow for implementation
  - contracts/chi-middleware-api.yaml: API contract for developers
  - quickstart.md: User-facing guide for Chi middleware usage
  - All docs explicitly requested by user for "Chi flavored middleware" feature

### Principle II: Test Coverage Preservation
- ✅ PASS - Design preserves test coverage:
  - Shared helpers will have tests in http/internal/helpers/helpers_test.go
  - Chi middleware will have tests in http/chi/middleware_test.go
  - Test scenarios defined in spec cover all functional requirements
  - Coverage maintained through table-driven tests adapted from stdlib

### Principle III: Test-First Development
- ✅ PASS - Design supports test-first approach:
  - Test scenarios documented in spec and data-model.md
  - Clear input/output contracts in chi-middleware-api.yaml
  - Implementation will follow: test → verify fails → implement → verify passes
  - Tests can be written before implementation using mock facilitator

### Principle IV: Stdlib-First Approach
- ✅ PASS - Design maximizes stdlib usage:
  - Chi chosen by user, uses stdlib http.Handler interface
  - Shared helpers reduce code to stdlib types only
  - No additional external dependencies beyond Chi (already in go.mod)
  - Reuses existing stdlib middleware patterns

### Principle V: Code Conciseness
- ✅ PASS - Design ensures concise code:
  - Shared helpers eliminate duplication (4 helper functions shared across 4 middleware implementations)
  - Chi middleware estimated ~150-200 lines (constructor + minimal wrapper)
  - No unnecessary abstractions - simple adapter pattern
  - Comments will explain why, not what (per spec and constitution)

### Principle VI: Binary Cleanup
- ✅ PASS - Design prevents binary commits:
  - Chi middleware is library code (no executables generated)
  - Example in examples/chi/ will be covered by .gitignore
  - No build artifacts in implementation

### Code Quality Gates
- ✅ All design artifacts ready for implementation
- ✅ Test scenarios defined and ready for table-driven tests
- ✅ API contract specifies expected behavior for all scenarios
- ✅ Implementation will pass go fmt, go vet, golangci-lint

### Post-Design Summary
**ALL GATES STILL PASSED** - Design complete and ready for Phase 2 (tasks generation).

**Design Quality**:
- Clear separation of concerns (constructor, middleware handler, helpers)
- Consistent with existing middleware patterns (stdlib, Gin, PocketBase)
- Testable design (all functions pure, mock facilitator)
- No technical debt introduced (shared helpers reduce future maintenance)

**Constitution Compliance**:
- No violations introduced during design phase
- All documentation serves clear purpose
- Test coverage maintained/improved through shared tests
- Stdlib-first principle followed (Chi uses stdlib types)
- Code will be concise (shared helpers, minimal adapter)
- No binaries to manage (library code only)

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
http/
├── chi/
│   ├── middleware.go        # Chi middleware adapter (NEW - this feature)
│   └── middleware_test.go   # Chi middleware tests (NEW - this feature)
├── gin/
│   ├── middleware.go        # Gin middleware adapter (existing)
│   └── middleware_test.go   # Gin middleware tests (existing)
├── pocketbase/
│   ├── middleware.go        # PocketBase middleware adapter (existing)
│   └── middleware_test.go   # PocketBase middleware tests (existing)
├── internal/
│   └── helpers/
│       ├── helpers.go       # Shared helper functions (NEW - refactored from stdlib)
│       └── helpers_test.go  # Helper function tests (NEW)
├── middleware.go            # Stdlib middleware (existing)
├── middleware_test.go       # Stdlib middleware tests (existing)
├── facilitator.go           # Facilitator client (existing)
├── facilitator_test.go      # Facilitator client tests (existing)
├── handler.go               # HTTP handlers (existing)
└── transport.go             # HTTP transport types (existing)

examples/
├── chi/                     # Chi usage example (NEW - optional)
│   └── main.go
├── gin/                     # Gin usage example (existing)
│   └── main.go
└── basic/                   # Basic stdlib example (existing)
    └── main.go
```

**Structure Decision**: Single project with Go library packages. The Chi middleware will live in `http/chi/` following the established pattern from Gin (`http/gin/`) and PocketBase (`http/pocketbase/`) adapters. Shared helper functions will be extracted to `http/internal/helpers/` to reduce code duplication across all middleware implementations while maintaining the same behavior.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (No violations - table empty) | - | - |
