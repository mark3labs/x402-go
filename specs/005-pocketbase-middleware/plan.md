# Implementation Plan: PocketBase Middleware for x402 Payment Protocol

**Branch**: `005-pocketbase-middleware` | **Date**: 2025-10-29 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/005-pocketbase-middleware/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Implement a PocketBase-compatible middleware for x402 payment gating by adapting the existing stdlib HTTP middleware pattern to PocketBase's `core.RequestEvent` and `hook.Handler` system. The middleware will reuse all core logic from `http/middleware.go` and `http/facilitator.go` while duplicating framework-specific helpers following the established Gin middleware pattern. Key features include payment verification, settlement, request store integration via e.Set/e.Get, and support for both route-level and group-level binding using PocketBase's Bind/BindFunc methods.

## Technical Context

**Language/Version**: Go 1.25.1  
**Primary Dependencies**: 
- PocketBase framework (github.com/pocketbase/pocketbase/core)
- Existing x402-go core package (types, errors)
- Existing http package (facilitator client, config)
- Go standard library (net/http, encoding/json, encoding/base64)

**Storage**: N/A (stateless middleware, payment tracking delegated to facilitator)  
**Testing**: Go standard testing package + table-driven tests matching stdlib middleware_test.go patterns  
**Target Platform**: Any platform running PocketBase (Linux, macOS, Windows)  
**Project Type**: Library package (middleware component)  
**Performance Goals**: <5ms middleware overhead for verification, <60s total for settlement  
**Constraints**: 
- Must match stdlib middleware behavior exactly (verify-then-settle)
- No response buffering or custom ResponseWriter
- Uses hardcoded timeouts (see spec.md FR-020)
- JSON-only responses using e.JSON() (not PocketBase native error types)

**Scale/Scope**: Single package (~300 lines), 4 helper functions, 3+ core tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ Principle I: No Unnecessary Documentation
**Status**: PASS - No documentation created beyond required spec/plan/research artifacts

### ✅ Principle II: Test Coverage Preservation
**Status**: PASS - New middleware will include comprehensive tests matching stdlib patterns. Adding new code with tests maintains overall coverage.

### ✅ Principle III: Test-First Development
**Status**: PASS - Plan includes test cases before implementation. Will adapt stdlib middleware_test.go patterns.

### ✅ Principle IV: Stdlib-First Approach
**Status**: PASS - Uses Go stdlib (net/http, encoding/json, encoding/base64). PocketBase framework is necessary external dependency for the adapter.

### ✅ Principle V: Code Conciseness
**Status**: PASS - Minimal adapter (~300 lines) with four duplicated framework-specific helpers (parsePaymentHeaderFromRequest, sendPaymentRequiredPocketBase, findMatchingRequirementPocketBase, addPaymentResponseHeaderPocketBase) following established Gin pattern. No unnecessary abstractions.

### ✅ Principle VI: Binary Cleanup
**Status**: PASS - No binaries generated for middleware library code. Tests do not produce committed artifacts.

**Overall Gate Status**: ✅ PASS - No violations. Proceed to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/005-pocketbase-middleware/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── pocketbase-middleware-api.yaml
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
http/
├── pocketbase/
│   ├── middleware.go        # Main PocketBase middleware implementation
│   └── middleware_test.go   # Tests adapted from stdlib middleware_test.go
├── middleware.go            # Existing stdlib middleware (reference)
├── gin/
│   └── middleware.go        # Existing Gin middleware (reference for duplication pattern)
├── facilitator.go           # Shared facilitator client (reused)
└── facilitator_test.go      # Facilitator tests

examples/
├── pocketbase/
│   ├── main.go              # Example PocketBase app with x402 middleware
│   └── README.md            # Usage example

testdata/
├── evm/
│   └── valid_payment.json   # Test fixtures (reused)
└── svm/
    └── valid_payment.json   # Test fixtures (reused)
```

**Structure Decision**: Following established pattern where framework-specific adapters live in `http/<framework>/` subdirectories. PocketBase middleware goes in `http/pocketbase/` alongside existing `http/gin/`. This maintains consistency with the codebase structure and makes framework adapters easy to discover.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations detected. N/A.

---

## Phase 0: Research & Unknowns Resolution

**Input**: spec.md Technical Context section, Constitution Check unknowns  
**Output**: research.md with resolved NEEDS CLARIFICATION items  
**Gate**: All technical unknowns resolved before Phase 1

### Research Tasks

1. **PocketBase middleware registration patterns**
   - Task: Analyze PocketBase documentation for Bind/BindFunc usage with hook.Handler[*core.RequestEvent]
   - Source: https://pocketbase.io/docs/go-routing/#registering-middlewares (already fetched)
   - Output: Confirm middleware signature and registration approach

2. **PocketBase request/response handling**
   - Task: Document core.RequestEvent methods for header parsing, JSON responses, request store
   - Source: PocketBase documentation + existing gin/middleware.go patterns
   - Output: Confirm e.JSON(), e.Set/e.Get, e.Next() usage patterns

3. **Helper function duplication requirements**
   - Task: Identify which helpers from stdlib need PocketBase-specific versions
   - Source: Compare http/middleware.go vs http/gin/middleware.go
   - Output: List of 4 helper functions to duplicate (parse, send402, findRequirement, addResponse)

4. **Test adaptation strategy**
   - Task: Review http/middleware_test.go to determine how tests adapt to core.RequestEvent
   - Source: stdlib middleware_test.go + PocketBase testing patterns
   - Output: Test structure and mock strategy for core.RequestEvent

### Known Technical Decisions (from spec.md)

- **Error responses**: Use e.JSON() for x402 protocol compliance (NOT PocketBase native error types)
- **Timeouts**: Hardcoded VerifyTimeout=5s, SettleTimeout=60s matching stdlib/Gin
- **Enrichment**: Call facilitator.EnrichRequirements() at initialization for network-specific config
- **Context storage**: Use e.Set("x402_payment", verifyResp) for request store
- **Helper duplication**: Follow Gin pattern - duplicate four framework-specific helpers (parsePaymentHeaderFromRequest, sendPaymentRequiredPocketBase, findMatchingRequirementPocketBase, addPaymentResponseHeaderPocketBase) for self-contained adapter

---

## Phase 1: Design & Contracts

**Input**: research.md, spec.md requirements  
**Output**: data-model.md, contracts/pocketbase-middleware-api.yaml, quickstart.md  
**Gate**: Design review + Constitution re-check

### Design Artifacts

1. **data-model.md**: Document data flow through middleware
   - PocketBase middleware config (uses stdlib http.Config)
   - Request flow: e.Request → parsePaymentHeader → facilitator verify/settle → e.Set → e.Next()
   - Response structures: PaymentRequirementsResponse, VerifyResponse, SettlementResponse

2. **contracts/pocketbase-middleware-api.yaml**: OpenAPI contract for middleware
   - Middleware function signature: `func NewPocketBaseX402Middleware(config *http.Config) *hook.Handler[*core.RequestEvent]`
   - Helper functions: parse, send402, findRequirement, addResponse
   - Request store key: "x402_payment" → VerifyResponse

3. **quickstart.md**: Example PocketBase integration
   - Import statement: `import pbx402 "github.com/mark3labs/x402-go/http/pocketbase"`
   - Config setup: same http.Config as stdlib
   - Middleware registration: `se.Router.Bind(pbx402.NewPocketBaseX402Middleware(config))`
   - Handler access: `verifyResp := e.Get("x402_payment").(*http.VerifyResponse)`

### Agent Context Update

After completing design artifacts, run:
```bash
.specify/scripts/bash/update-agent-context.sh opencode
```

This updates AGENTS.md with:
- Added technology: PocketBase framework (github.com/pocketbase/pocketbase)
- Recent change: "005-pocketbase-middleware: Added PocketBase middleware adapter"

---

## Phase 2: Implementation Tasks (Generated by /speckit.tasks)

**Note**: Phase 2 tasks are NOT generated by `/speckit.plan`. Run `/speckit.tasks` separately to generate `tasks.md` with implementation checklist.

Expected task categories:
- Setup: Create http/pocketbase/ package structure
- Core: Implement NewPocketBaseX402Middleware function
- Helpers: Duplicate 4 framework-specific helper functions
- Tests: Adapt stdlib middleware_test.go for core.RequestEvent
- Examples: Create examples/pocketbase/ with working demo
- Documentation: Update examples/pocketbase/README.md

---

## Phase 3: Validation (Post-Implementation)

**Success Criteria from spec.md**:
- ✅ SC-001: Developers can protect endpoint with stdlib http.Config
- ✅ SC-002: 100% of stdlib test scenarios pass (402, 400/402, e.Next(), 503)
- ✅ SC-003: VerifyResponse accessible via e.Get("x402_payment")
- ✅ SC-004: Supports EVM (base, base-sepolia) and SVM (solana-*) networks
- ✅ SC-005: Uses same Config struct as stdlib (consistency)
- ✅ SC-006: 3+ core tests adapted from stdlib
- ✅ SC-007: Works with Bind/BindFunc for groups and routes

**Validation Commands**:
```bash
# Run tests with race detection
go test -race ./http/pocketbase/...

# Check coverage
go test -race -cover ./http/pocketbase/...

# Lint
golangci-lint run http/pocketbase/

# Build example
go build -o /tmp/pocketbase-example ./examples/pocketbase/

# Verify no binaries committed
git status | grep -E '\.(exe|out|test)$' && echo "ERROR: Binaries found" || echo "OK"
```

---

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| PocketBase API changes | High - middleware breaks | Pin PocketBase version, document compatibility |
| Test adapter complexity | Medium - hard to mock core.RequestEvent | Study Gin patterns, use httptest.ResponseRecorder |
| Helper duplication maintenance | Low - code drift | Document why duplication needed (self-contained adapters) |

---

## Notes

- This middleware is a thin adapter - all business logic is in http/middleware.go and http/facilitator.go
- Follow Gin duplication pattern: duplicate four framework-specific helpers to maintain self-contained adapters without stdlib coupling
- Use e.JSON() for all responses to maintain x402 protocol compliance
- Store payment details in PocketBase request store with key "x402_payment"
- Middleware must work with both route-level and group-level Bind/BindFunc
