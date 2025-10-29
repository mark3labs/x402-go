# Implementation Requirements Checklist: PocketBase Middleware for x402 Payment Protocol

**Feature**: 005-pocketbase-middleware  
**Date**: 2025-10-29  
**Status**: Phase 1 Complete - Ready for Implementation

## Functional Requirements

### Core Middleware Functionality

- [ ] **FR-001**: System provides PocketBase-compatible middleware function (hook.Handler[*core.RequestEvent])
  - Implementation: `func NewPocketBaseX402Middleware(config *http.Config) func(*core.RequestEvent) error`
  - Verification: Check function signature and return type

- [ ] **FR-002**: Middleware accepts payment requirements via Config struct PaymentRequirements field
  - Implementation: Reuse `http.Config` struct from stdlib
  - Verification: Test with multiple payment requirements

- [ ] **FR-003**: Middleware accepts recipient wallet via PaymentRequirement.PayTo field
  - Implementation: Use existing `x402.PaymentRequirement` struct
  - Verification: Test with valid wallet addresses

- [ ] **FR-004**: Middleware accepts Config matching stdlib (FacilitatorURL, FallbackFacilitatorURL, PaymentRequirements, VerifyOnly)
  - Implementation: Import and use `http.Config`
  - Verification: Config compatibility test with stdlib

- [ ] **FR-005**: Middleware checks X-PAYMENT header via e.Request.Header.Get("X-Payment")
  - Implementation: `paymentHeader := e.Request.Header.Get("X-PAYMENT")`
  - Verification: Test with/without header

### Error Handling

- [ ] **FR-006**: Returns HTTP 402 with PaymentRequirementsResponse for missing payment; HTTP 400 for malformed headers
  - Implementation: `sendPaymentRequiredPocketBase()` and e.JSON() for 400
  - Verification: Test missing header → 402, invalid base64 → 400

- [ ] **FR-007**: Parses base64-encoded X-PAYMENT header containing JSON
  - Implementation: `parsePaymentHeaderFromRequest()` helper
  - Verification: Test base64 decode + JSON unmarshal

- [ ] **FR-008**: Verifies payments via facilitator /verify endpoint
  - Implementation: Reuse `facilitator.Verify()`
  - Verification: Mock facilitator test

- [ ] **FR-009**: Settles payments via facilitator /settle endpoint (unless VerifyOnly=true)
  - Implementation: Reuse `facilitator.Settle()`
  - Verification: Test both modes (settle and verify-only)

### Context Integration

- [ ] **FR-010**: Stores payment details in request store using e.Set("x402_payment", verifyResp)
  - Implementation: `e.Set("x402_payment", verifyResp)`
  - Verification: Handler retrieves via e.Get()

- [ ] **FR-011**: Adds X-PAYMENT-RESPONSE header with base64-encoded settlement
  - Implementation: `addPaymentResponseHeaderPocketBase()` helper
  - Verification: Check response header after settlement

### Network Support

- [ ] **FR-012**: Supports testnet mode via Config.PaymentRequirements network field
  - Implementation: User specifies network (base-sepolia, solana-devnet)
  - Verification: Test with testnet networks

- [ ] **FR-013**: Uses network-appropriate USDC contract address from PaymentRequirement.Asset
  - Implementation: User configures Asset field
  - Verification: Test with base and solana asset addresses

- [ ] **FR-014**: Constructs resource URL from request (scheme + host + requestURI)
  - Implementation: Match stdlib logic for URL construction
  - Verification: Test URL population in requirements

### Request Flow

- [ ] **FR-015**: Calls e.Next() after successful verification
  - Implementation: `return e.Next()` at end of middleware
  - Verification: Test handler execution after payment

- [ ] **FR-016**: Does NOT call e.Next() when verification fails
  - Implementation: Return error without calling e.Next()
  - Verification: Test handler not executed on failure

- [ ] **FR-017**: Works with route-level and group-level binding
  - Implementation: Compatible with Route.Bind() and RouterGroup.Bind()
  - Verification: Test both registration patterns

### Response Handling

- [ ] **FR-018**: Returns JSON responses using e.JSON() method
  - Implementation: Use e.JSON() for 402, 400, 503 responses
  - Verification: Check Content-Type and JSON structure

- [ ] **FR-019**: Returns errors with x402Version field and structured messages
  - Implementation: All errors include `{"x402Version": 1, "error": "..."}`
  - Verification: Validate error response structure

### Configuration

- [ ] **FR-020**: Creates FacilitatorClient with hardcoded timeouts (VerifyTimeout=5s, SettleTimeout=60s)
  - Implementation: Match stdlib/Gin timeout values
  - Verification: Check timeout values in client

- [ ] **FR-021**: Calls facilitator.EnrichRequirements() at initialization
  - Implementation: Call before returning middleware function
  - Verification: Test enrichment with mock facilitator

- [ ] **FR-022**: Logs warning and continues if EnrichRequirements() fails
  - Implementation: Graceful degradation with slog.Warn()
  - Verification: Test with unreachable facilitator during init

---

## Success Criteria

### Measurable Outcomes

- [ ] **SC-001**: Developers protect endpoint by passing stdlib http.Config
  - Verification: Create example with minimal Config
  - Acceptance: Example in quickstart.md works

- [ ] **SC-002**: 100% of stdlib test scenarios pass
  - Verification: Run adapted tests from middleware_test.go
  - Acceptance: All core tests pass (missing payment, invalid payment, valid payment, facilitator failures)

- [ ] **SC-003**: Payment details accessible via e.Get("x402_payment")
  - Verification: Handler test retrieves VerifyResponse
  - Acceptance: Payer, IsValid, InvalidReason fields available

- [ ] **SC-004**: Supports EVM (base, base-sepolia) and SVM (solana-*) networks
  - Verification: Test with both network types
  - Acceptance: Enrichment works for both EVM and SVM

- [ ] **SC-005**: Uses same Config struct as stdlib
  - Verification: Import http.Config, no duplication
  - Acceptance: Config definition identical to stdlib

- [ ] **SC-006**: 3+ core tests adapted from stdlib
  - Verification: Count test cases in middleware_test.go
  - Acceptance: At least 3 tests covering main scenarios

- [ ] **SC-007**: Works with Bind/BindFunc methods
  - Verification: Test both registration methods
  - Acceptance: Route.BindFunc() and RouterGroup.Bind() both work

---

## Implementation Checklist

### Phase 2.1: Setup

- [ ] Create `http/pocketbase/` directory
- [ ] Create `http/pocketbase/middleware.go` file
- [ ] Create `http/pocketbase/middleware_test.go` file
- [ ] Add package documentation

### Phase 2.2: Core Middleware

- [ ] Implement `NewPocketBaseX402Middleware(config *http.Config)` factory function
- [ ] Create FacilitatorClient instances (primary + fallback)
- [ ] Call `facilitator.EnrichRequirements()` with error handling
- [ ] Return middleware handler function

### Phase 2.3: Request Processing

- [ ] Check X-PAYMENT header
- [ ] Handle missing header → 402 response
- [ ] Parse payment header (call helper)
- [ ] Handle parse errors → 400 response
- [ ] Find matching requirement (call helper)
- [ ] Handle no match → 402 response

### Phase 2.4: Payment Verification

- [ ] Call facilitator.Verify() with payment and requirement
- [ ] Try fallback facilitator on primary failure
- [ ] Handle verification failure → 503 response
- [ ] Handle invalid payment → 402 response
- [ ] Store VerifyResponse in request store

### Phase 2.5: Payment Settlement

- [ ] Check VerifyOnly flag
- [ ] Call facilitator.Settle() if not verify-only
- [ ] Try fallback facilitator on primary failure
- [ ] Handle settlement failure → 503 response
- [ ] Add X-PAYMENT-RESPONSE header (call helper)
- [ ] Call e.Next() to continue handler chain

### Phase 2.6: Helper Functions

- [ ] Implement `parsePaymentHeaderFromRequest(r *http.Request)` (~27 lines)
  - Base64 decode
  - JSON unmarshal
  - Version validation
  
- [ ] Implement `sendPaymentRequiredPocketBase(e *core.RequestEvent, requirements)` (~9 lines)
  - Create PaymentRequirementsResponse
  - Use e.JSON(http.StatusPaymentRequired, response)
  
- [ ] Implement `findMatchingRequirementPocketBase(payment, requirements)` (~8 lines)
  - Match scheme and network
  - Return requirement or error
  
- [ ] Implement `addPaymentResponseHeaderPocketBase(e *core.RequestEvent, settlement)` (~12 lines)
  - JSON marshal settlement
  - Base64 encode
  - Set header via e.Response.Header().Set()

### Phase 2.7: Testing

- [ ] Create test helper: `newTestRequestEvent(method, url, headers)`
- [ ] Test: Missing payment header → 402
- [ ] Test: Invalid base64 → 400
- [ ] Test: Invalid JSON → 400
- [ ] Test: Valid payment → 200 + e.Next() called
- [ ] Test: Facilitator verify failure → 503
- [ ] Test: VerifyOnly mode → no settlement
- [ ] Test: Payment details in request store
- [ ] Test: Route-level binding
- [ ] Test: Group-level binding
- [ ] Run tests with race detection: `go test -race ./http/pocketbase/...`
- [ ] Check coverage: `go test -cover ./http/pocketbase/...`

### Phase 2.8: Examples

- [ ] Create `examples/pocketbase/` directory
- [ ] Create `examples/pocketbase/main.go` example app
- [ ] Create `examples/pocketbase/README.md` usage guide
- [ ] Test example builds: `go build ./examples/pocketbase/`
- [ ] Verify no binaries committed: `git status`

### Phase 2.9: Documentation

- [ ] Update examples/pocketbase/README.md with usage instructions
- [ ] Add godoc comments to exported functions
- [ ] Document request store key "x402_payment"
- [ ] Document helper function purposes (internal comments)

### Phase 2.10: Validation

- [ ] Run all tests: `go test -race ./http/pocketbase/...`
- [ ] Check coverage: `go test -cover ./http/pocketbase/...`
- [ ] Run linter: `golangci-lint run http/pocketbase/`
- [ ] Format code: `go fmt ./http/pocketbase/...`
- [ ] Vet code: `go vet ./http/pocketbase/...`
- [ ] Build example: `go build -o /tmp/pb-example ./examples/pocketbase/`
- [ ] Clean binaries: `rm /tmp/pb-example`
- [ ] Verify no binaries in git: `git status | grep -E '\.(exe|out|test)$'`

---

## Constitution Compliance

### ✅ Principle I: No Unnecessary Documentation
- Only creating spec, plan, research, data-model, contracts, quickstart (all required)
- No additional markdown files

### ✅ Principle II: Test Coverage Preservation
- Adding tests for new code (middleware + helpers)
- Coverage maintained or improved

### ✅ Principle III: Test-First Development
- Tests planned before implementation (Phase 2.7)
- Test cases defined in requirements

### ✅ Principle IV: Stdlib-First Approach
- Uses Go stdlib (net/http, encoding/json, encoding/base64)
- PocketBase is necessary framework dependency

### ✅ Principle V: Code Conciseness
- Minimal adapter (~300 lines including helpers)
- No unnecessary abstractions
- Duplicates helpers only where needed (self-contained pattern)

### ✅ Principle VI: Binary Cleanup
- Build artifacts in /tmp or gitignored directories
- Cleanup step in validation checklist
- Git status verification before commit

---

## Dependencies Verification

- [ ] PocketBase framework available: `go list github.com/pocketbase/pocketbase/core`
- [ ] x402-go core available: `go list github.com/mark3labs/x402-go`
- [ ] http package available: `go list github.com/mark3labs/x402-go/http`
- [ ] All dependencies in go.mod: `go mod tidy`

---

## Pre-Implementation Review

Before starting implementation, verify:

- [ ] All Phase 0 research completed (research.md exists)
- [ ] All Phase 1 design completed (data-model.md, contracts/, quickstart.md exist)
- [ ] Agent context updated (AGENTS.md has PocketBase entry)
- [ ] Constitution check passed (no violations)
- [ ] Requirements understood (all FRs and SCs clear)

---

## Post-Implementation Review

After completing implementation, verify:

- [ ] All functional requirements implemented (FR-001 through FR-022)
- [ ] All success criteria met (SC-001 through SC-007)
- [ ] All tests passing with race detection
- [ ] Coverage maintained or improved
- [ ] Linter passing (golangci-lint)
- [ ] Example builds and runs
- [ ] No binaries committed
- [ ] Documentation complete

---

## Notes

This checklist is derived from:
- spec.md (functional requirements, success criteria)
- plan.md (implementation structure, constitution compliance)
- research.md (technical decisions, helper functions)
- data-model.md (data flow, entities, validation rules)

Estimated implementation time: 4-6 hours for an experienced Go developer

**Status**: ✅ Ready for Phase 2 (Implementation via /speckit.tasks)
