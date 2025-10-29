# Tasks: Chi Middleware for x402 Payment Protocol

**Input**: Design documents from `/specs/006-chi-middleware/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/chi-middleware-api.yaml

**Tests**: Test tasks included per Constitution Principle III (Test-First Development). Tests MUST be written and passing before implementation tasks.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

This is a Go library package. All paths relative to repository root:
- Middleware implementation: `http/chi/`
- Shared helpers: `http/internal/helpers/`
- Example usage: `examples/chi/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure for Chi middleware

- [ ] T001 Create http/chi/ package directory for Chi middleware adapter
- [ ] T002 Create http/internal/helpers/ package directory for shared helper functions
- [ ] T003 [P] Verify Chi dependency (github.com/go-chi/chi/v5) exists in go.mod
- [ ] T004 [P] Create examples/chi/ directory for Chi usage example

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared helper functions that ALL user stories depend on - MUST be complete before ANY user story implementation

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T005 [P] Extract and share parsePaymentHeaderFromRequest helper from http/middleware.go to http/internal/helpers/helpers.go (move existing function, preserve behavior)
- [ ] T006 [P] Extract and share findMatchingRequirement helper from http/middleware.go to http/internal/helpers/helpers.go (move existing function, preserve behavior)
- [ ] T007 [P] Extract and share sendPaymentRequired helper from http/middleware.go to http/internal/helpers/helpers.go (move existing function, preserve behavior)
- [ ] T008 [P] Extract and share addPaymentResponseHeader helper from http/middleware.go to http/internal/helpers/helpers.go (move existing function, preserve behavior)
- [ ] T009 Refactor http/middleware.go to use shared helpers from http/internal/helpers/ package (replace extracted functions with helper calls)
- [ ] T010 Refactor http/gin/middleware.go to use shared helpers from http/internal/helpers/ package (replace duplicated code with helper calls)

**Checkpoint**: Foundation ready - all helper functions extracted and tested via existing middleware, user story implementation can now begin

---

## Phase 2.5: Test Foundation (Constitution Requirement)

**Purpose**: Write ALL tests BEFORE implementation per Constitution Principle III (Test-First Development)

**⚠️ CONSTITUTIONAL REQUIREMENT**: These tests must be written first, verified to fail, then implementation proceeds

**Test Coverage Requirements**:
- All functional requirements (FR-001 through FR-023)
- All user stories (US1, US2, US3)
- All edge cases from spec
- Minimum coverage: match or exceed existing stdlib/Gin middleware test coverage

### Helper Function Tests (Foundation for all middleware)

- [ ] T-001 [P] Write TestParsePaymentHeaderFromRequest in http/internal/helpers/helpers_test.go (valid base64 JSON, invalid base64, invalid JSON, missing header)
- [ ] T-002 [P] Write TestFindMatchingRequirement in http/internal/helpers/helpers_test.go (scheme+network match, no match, multiple requirements)
- [ ] T-003 [P] Write TestSendPaymentRequired in http/internal/helpers/helpers_test.go (verify 402 status, JSON format, payment requirements in body)
- [ ] T-004 [P] Write TestAddPaymentResponseHeader in http/internal/helpers/helpers_test.go (verify base64 encoding, header presence, settlement data structure)

### Chi Middleware Core Tests (User Story 1 - Basic Payment Gating)

- [ ] T-005 Write TestNewChiX402Middleware in http/chi/middleware_test.go (constructor returns valid middleware, facilitator client created, EnrichRequirements called)
- [ ] T-006 Write TestChiMiddleware_MissingPayment in http/chi/middleware_test.go (no X-PAYMENT header → HTTP 402 with payment requirements JSON)
- [ ] T-007 Write TestChiMiddleware_InvalidPaymentHeader in http/chi/middleware_test.go (malformed base64 → HTTP 400, invalid JSON → HTTP 400, both with x402Version error response per FR-020)
- [ ] T-008 Write TestChiMiddleware_ValidPayment in http/chi/middleware_test.go (valid payment → verification called → settlement called → handler executes → X-PAYMENT-RESPONSE header added per FR-011)
- [ ] T-009 Write TestChiMiddleware_VerificationFailure in http/chi/middleware_test.go (facilitator returns error → HTTP 503 with error details)
- [ ] T-010 Write TestChiMiddleware_SettlementFailure in http/chi/middleware_test.go (verification succeeds but settlement fails → HTTP 503)
- [ ] T-011 Write TestChiMiddleware_OptionsRequestBypass in http/chi/middleware_test.go (r.Method == "OPTIONS" → skip all payment logic → call next handler immediately per FR-022)
- [ ] T-012 Write TestChiMiddleware_FacilitatorTimeouts in http/chi/middleware_test.go (verify VerifyTimeout=5s and SettleTimeout=60s per FR-017)

### Chi Middleware Context Tests (User Story 2 - Context Integration)

- [ ] T-013 Write TestChiMiddleware_ContextIntegration in http/chi/middleware_test.go (valid payment → VerifyResponse stored in context → handler can access via httpx402.PaymentContextKey → verify Payer, IsValid, InvalidReason fields per FR-010)

### Chi Middleware Verify-Only Tests (User Story 3 - Verify-Only Mode)

- [ ] T-014 Write TestChiMiddleware_VerifyOnlyMode in http/chi/middleware_test.go (config.VerifyOnly=true → verification succeeds → settlement NOT called → no X-PAYMENT-RESPONSE header per FR-009, FR-011)
- [ ] T-015 Write TestChiMiddleware_VerifyOnlyDisabled in http/chi/middleware_test.go (config.VerifyOnly=false → both verification AND settlement called)

### Chi Middleware Edge Cases & Additional Coverage

- [ ] T-016 Write TestChiMiddleware_NetworkSupport in http/chi/middleware_test.go (base-sepolia and base networks via PaymentRequirement.Network per FR-012)
- [ ] T-017 Write TestChiMiddleware_ResourceURLConstruction in http/chi/middleware_test.go (verify scheme + host + requestURI concatenation per FR-014)
- [ ] T-018 Write TestChiMiddleware_FallbackFacilitator in http/chi/middleware_test.go (primary fails → fallback used for verification and settlement)
- [ ] T-019 Write TestChiMiddleware_LoggingEvents in http/chi/middleware_test.go (verify slog.Default() calls for: missing payment (Warn), invalid payment (Warn), verification success (Info), verification failure (Error), settlement success (Info), settlement failure (Error) per FR-023)
- [ ] T-020 Write TestChiMiddleware_EnrichRequirementsWarning in http/chi/middleware_test.go (EnrichRequirements fails → log warning with slog.Warn() → continue with original requirements per FR-019)

### Test Coverage Validation

- [ ] T-021 Run `go test -race ./http/internal/helpers/...` and verify all helper tests pass
- [ ] T-022 Run `go test -race ./http/chi/...` and verify all Chi middleware tests pass (all should FAIL at this point - expected before implementation)
- [ ] T-023 Run `go test -race -cover ./http/chi/... ./http/internal/helpers/...` and document baseline coverage (target: ≥ stdlib/Gin middleware coverage per Constitution Principle II)

**Checkpoint**: All tests written and verified to fail (expected behavior before implementation). Ready to proceed with implementation that will make tests pass.

---

## Phase 3: User Story 1 - Basic Payment Gating with Chi (Priority: P1) 🎯 MVP

**Goal**: Enable developers to protect Chi routes with x402 payment gating using middleware that verifies and settles payments

**Independent Test**: Create a simple Chi application with a protected endpoint, send requests with/without valid X-PAYMENT headers, verify 402 responses for missing payments and successful access with valid payment

**Acceptance Scenarios**:
1. Request without X-PAYMENT header → HTTP 402 with payment requirements in JSON
2. Request with valid X-PAYMENT header → payment verified, settled, protected handler executes
3. Request with invalid X-PAYMENT header → HTTP 400 Bad Request with x402Version error
4. Payment verification fails at facilitator → HTTP 503 Service Unavailable

### Implementation for User Story 1

- [ ] T011 [US1] Create Chi middleware constructor NewChiX402Middleware with signature `func(config *httpx402.Config) func(http.Handler) http.Handler` in http/chi/middleware.go per FR-001
- [ ] T012 [US1] Implement facilitator client creation with hardcoded VerifyTimeout=5s and SettleTimeout=60s (exact values per FR-017) in http/chi/middleware.go NewChiX402Middleware constructor
- [ ] T013 [US1] Implement fallback facilitator client creation (if configured) in http/chi/middleware.go
- [ ] T014 [US1] Implement EnrichRequirements call at constructor time; on failure log warning with slog.Warn() and continue with original requirements (graceful degradation per FR-019) in http/chi/middleware.go
- [ ] T014a [US1] Implement x402Version error response formatting for all error scenarios (400 for malformed headers, 402 for missing/invalid payment, 503 for facilitator failures) matching x402 HTTP transport specification per FR-020 in http/chi/middleware.go
- [ ] T014b [US1] Validate Config struct fields (FacilitatorURL, FallbackFacilitatorURL, PaymentRequirements slice, VerifyOnly flag) match stdlib http.Config exactly per FR-002, FR-003, FR-004 in http/chi/middleware.go constructor
- [ ] T015 [US1] Implement OPTIONS request bypass - check r.Method == "OPTIONS" and skip all payment verification logic, call next handler immediately per FR-022 (CORS preflight support) in http/chi/middleware.go
- [ ] T016 [US1] Implement resource URL construction (scheme + host + requestURI) in http/chi/middleware.go
- [ ] T017 [US1] Implement PaymentRequirement.Resource population with request URL in http/chi/middleware.go
- [ ] T018 [US1] Implement X-PAYMENT header check by calling shared `parsePaymentHeaderFromRequest` helper from http/internal/helpers package in http/chi/middleware.go per FR-005, FR-007
- [ ] T019 [US1] Implement payment requirement matching using shared findMatchingRequirement helper in http/chi/middleware.go
- [ ] T020 [US1] Implement payment verification with facilitator (primary + fallback support) in http/chi/middleware.go
- [ ] T021 [US1] Implement payment settlement logic (unless VerifyOnly=true) in http/chi/middleware.go
- [ ] T022 [US1] Implement settlement response handling with fallback support in http/chi/middleware.go
- [ ] T023 [US1] Implement X-PAYMENT-RESPONSE header addition using shared addPaymentResponseHeader helper in http/chi/middleware.go
- [ ] T024 [US1] Implement error response handling (400 for malformed headers, 402 for missing/invalid payment, 503 for facilitator failures) in http/chi/middleware.go
- [ ] T025 [US1] Implement structured logging with slog.Default() for payment lifecycle events per FR-023: missing payment (Warn), invalid payment header (Warn), parse errors (Warn), verification success (Info), verification failure (Error), settlement success (Info), settlement failure (Error), facilitator errors (Error) in http/chi/middleware.go
- [ ] T025a [US1] Verify log level usage matches FR-023 specification: Info for success paths (verification success, settlement success), Warn for client errors (missing payment, invalid payment, parse errors), Error for service failures (verification failure, settlement failure, facilitator errors)

**Checkpoint**: At this point, User Story 1 should be fully functional - Chi middleware verifies and settles payments, returns appropriate error responses

---

## Phase 4: User Story 2 - Context Integration (Priority: P2)

**Goal**: Make payment details (payer address, verification status) available in Chi handlers via request context

**Independent Test**: Create a protected handler that accesses payment details from request context and returns them in the response, verify payer address and IsValid status are accessible

**Acceptance Scenarios**:
1. After valid payment processed → payment verification details (VerifyResponse) available via context.Value(httpx402.PaymentContextKey)
2. Handler can access payer address, IsValid status, and InvalidReason fields from context

### Implementation for User Story 2

- [ ] T026 [US2] Implement payment verification storage in request context using context.WithValue(r.Context(), httpx402.PaymentContextKey, verifyResp) in http/chi/middleware.go
- [ ] T027 [US2] Implement request context update with r.WithContext(ctx) before calling next handler in http/chi/middleware.go

**Checkpoint**: At this point, User Stories 1 AND 2 work - payment gating functional AND handlers can access payment details from context

---

## Phase 5: User Story 3 - Verify-Only Mode (Priority: P2)

**Goal**: Support verify-only mode where payments are verified but not settled (for testing or separate settlement handling)

**Independent Test**: Enable VerifyOnly flag in Config, verify that settlement is skipped after successful verification and no X-PAYMENT-RESPONSE header is added

**Acceptance Scenarios**:
1. VerifyOnly=true + valid payment → verification succeeds, settlement skipped
2. VerifyOnly=true + valid payment → no X-PAYMENT-RESPONSE header in response
3. VerifyOnly=false (default) + valid payment → both verification and settlement performed

### Implementation for User Story 3

- [ ] T028 [US3] Implement VerifyOnly flag check to skip settlement when config.VerifyOnly=true in http/chi/middleware.go
- [ ] T029 [US3] Implement conditional X-PAYMENT-RESPONSE header logic (only add when !config.VerifyOnly) in http/chi/middleware.go

**Checkpoint**: All user stories complete - Chi middleware supports full payment flow, context integration, and verify-only mode

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, examples, and cross-cutting improvements

- [ ] T030 [P] Create complete Chi usage example in examples/chi/main.go (similar to examples/gin/main.go)
- [ ] T031 [P] Add package documentation comments to http/chi/middleware.go explaining usage patterns
- [ ] T032 [P] Add function documentation comments to NewChiX402Middleware with example in http/chi/middleware.go
- [ ] T033 Run go fmt on all modified files (http/chi/, http/internal/helpers/, http/, http/gin/)
- [ ] T034 Run go vet on all modified packages to check for common mistakes
- [ ] T035 Run golangci-lint on all modified files to ensure code quality
- [ ] T036 Verify quickstart.md example code matches implementation in examples/chi/main.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
  - Extracts shared helpers from stdlib/Gin middleware
  - Refactors existing middleware to use shared helpers
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - US1 (Phase 3): Basic payment gating - no dependencies on other stories
  - US2 (Phase 4): Depends on US1 T011-T025 (needs middleware handler to add context storage)
  - US3 (Phase 5): Depends on US1 T021-T023 (needs settlement logic to add conditional skip)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
  - Implements complete payment verification and settlement flow
  - Must be complete for US2 and US3 (they modify/extend US1 logic)

- **User Story 2 (P2)**: Depends on User Story 1 tasks T011-T025
  - Adds context storage to existing middleware handler
  - Can be implemented as modification to existing middleware code

- **User Story 3 (P2)**: Depends on User Story 1 tasks T021-T023
  - Adds conditional logic to skip settlement
  - Can be implemented as modification to existing settlement code

### Within Each User Story

**User Story 1**:
1. Constructor setup (T011-T014) → must complete before middleware handler logic
2. OPTIONS bypass (T015) → independent, can be early in handler
3. URL construction (T016-T017) → before payment verification
4. Payment parsing (T018-T019) → before verification
5. Verification (T020) → before settlement
6. Settlement (T021-T023) → after verification
7. Error handling (T024-T025) → throughout handler logic

**User Story 2**:
1. Context storage (T026) → after successful verification
2. Context update (T027) → before calling next handler

**User Story 3**:
1. VerifyOnly check (T028) → during settlement phase
2. Conditional header (T029) → during settlement phase

### Parallel Opportunities

**Phase 1 (Setup)**:
- T003 and T004 can run in parallel (different operations)

**Phase 2 (Foundational)**:
- T005, T006, T007, T008 can all run in parallel (extracting different helper functions to http/internal/helpers/helpers.go)
- T009 and T010 must run AFTER T005-T008 complete (they use the extracted helpers)

**User Story 1 (after Foundational complete)**:
- All of US1 must be sequential within the same file (http/chi/middleware.go)
- But if multiple developers: US1 can proceed in parallel with US2/US3 development (though US2/US3 depend on US1 completion)

**Phase 6 (Polish)**:
- T030, T031, T032 can run in parallel (different files)
- T033, T034, T035, T036 must run sequentially after code complete

---

## Parallel Example: Foundational Phase

```bash
# Launch all helper extractions in parallel (different functions to same file):
Task: "Extract parsePaymentHeaderFromRequest helper to http/internal/helpers/helpers.go"
Task: "Extract findMatchingRequirement helper to http/internal/helpers/helpers.go"
Task: "Extract sendPaymentRequired helper to http/internal/helpers/helpers.go"
Task: "Extract addPaymentResponseHeader helper to http/internal/helpers/helpers.go"

# After helpers complete, refactor existing middleware (sequential):
Task: "Refactor http/middleware.go to use shared helpers from http/internal/helpers/"
Task: "Refactor http/gin/middleware.go to use shared helpers from http/internal/helpers/"
```

---

## Parallel Example: Polish Phase

```bash
# Launch all documentation tasks in parallel (different files):
Task: "Create complete Chi usage example in examples/chi/main.go"
Task: "Add package documentation comments to http/chi/middleware.go"
Task: "Add function documentation comments to NewChiX402Middleware in http/chi/middleware.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup → Package structure ready
2. Complete Phase 2: Foundational → Shared helpers extracted and tested
3. Complete Phase 3: User Story 1 → Basic payment gating functional
4. **STOP and VALIDATE**: Test User Story 1 independently with Chi router
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready (~10 tasks)
2. Add User Story 1 → Test independently → **MVP COMPLETE** (~25 tasks total)
3. Add User Story 2 → Test context access → Enhanced functionality (~27 tasks total)
4. Add User Story 3 → Test verify-only mode → Full feature set (~29 tasks total)
5. Polish → Documentation and examples complete (~36 tasks total)

Each story adds value without breaking previous functionality.

### Parallel Team Strategy

With multiple developers:

1. **Team completes Setup + Foundational together** (Phase 1-2)
   - Critical path: Extract helpers, refactor existing middleware
   - ~10 tasks, establishes foundation

2. **Once Foundational is done, assign stories**:
   - Developer A: User Story 1 (T011-T025) in http/chi/middleware.go
   - Developer B: Can prepare User Story 2 code (T026-T027) - must wait for A to finish US1
   - Developer C: Can prepare User Story 3 code (T028-T029) - must wait for A to finish US1

3. **Best approach**: Single developer implements US1 completely, then US2/US3 can be added by same or different developers

**Note**: Since US2 and US3 modify the same file created in US1 (http/chi/middleware.go), sequential development is recommended. Parallel development would create merge conflicts.

---

## Implementation Notes

### Helper Function Strategy (Phase 2 - Critical)

The Foundational phase extracts 4 helper functions from stdlib middleware to `http/internal/helpers/`:

1. **parsePaymentHeaderFromRequest**: Parse and validate X-PAYMENT header
   - Currently duplicated in http/gin/middleware.go (lines 202-229)
   - Logic from http/middleware.go parsePaymentHeader function

2. **findMatchingRequirement**: Match payment to requirement by scheme+network
   - Currently duplicated in http/gin/middleware.go (lines 243-250)
   - Logic from http/middleware.go findMatchingRequirement function

3. **sendPaymentRequired**: Send 402 response with payment requirements
   - Currently duplicated in http/gin/middleware.go (lines 231-240)
   - Logic from http/middleware.go sendPaymentRequiredWithRequirements function

4. **addPaymentResponseHeader**: Add X-PAYMENT-RESPONSE header with settlement details
   - Currently duplicated in http/gin/middleware.go (lines 252-267)
   - Logic from http/middleware.go addPaymentResponseHeader function

After extraction, http/middleware.go and http/gin/middleware.go must be refactored to use the shared helpers. This ensures consistent behavior across all 4 middleware implementations (stdlib, Gin, PocketBase, Chi).

### Chi Middleware Specifics (Phase 3)

Chi middleware signature: `func(http.Handler) http.Handler`
- Identical to stdlib signature
- Maximum code reuse from shared helpers
- Uses standard http.Request/http.ResponseWriter types
- No Chi-specific types needed

Constructor pattern follows Gin:
- Constructor function returns middleware handler
- Facilitator client(s) created at construction time
- EnrichRequirements called at construction time
- Middleware handler is a closure over config and facilitator clients

### Context Storage (Phase 4)

Chi uses stdlib context directly:
- Store: `context.WithValue(r.Context(), httpx402.PaymentContextKey, verifyResp)`
- Update: `r = r.WithContext(ctx)`
- Access in handler: `r.Context().Value(httpx402.PaymentContextKey).(*httpx402.VerifyResponse)`

Same pattern as stdlib middleware - no Chi-specific context types.

### Verify-Only Mode (Phase 5)

Simple flag check:
```go
if !config.VerifyOnly {
    // Settle payment
    // Add X-PAYMENT-RESPONSE header
}
```

Matches stdlib and Gin middleware behavior exactly.

---

## Task Summary

**Total Tasks**: 36

**By Phase**:
- Phase 1 (Setup): 4 tasks
- Phase 2 (Foundational): 6 tasks ⚠️ BLOCKS all user stories
- Phase 3 (User Story 1 - P1): 15 tasks 🎯 MVP
- Phase 4 (User Story 2 - P2): 2 tasks
- Phase 5 (User Story 3 - P2): 2 tasks
- Phase 6 (Polish): 7 tasks

**By User Story**:
- User Story 1 (Basic Payment Gating): 15 tasks - MVP functionality
- User Story 2 (Context Integration): 2 tasks - Enhanced developer experience
- User Story 3 (Verify-Only Mode): 2 tasks - Testing/flexibility support

**Parallel Opportunities**:
- Phase 1: 2 tasks can run in parallel (T003, T004)
- Phase 2: 4 tasks can run in parallel (T005-T008), then 2 sequential (T009-T010)
- Phase 6: 3 tasks can run in parallel (T030-T032)

**MVP Scope (User Story 1 only)**: 25 tasks
- Setup (4) + Foundational (6) + User Story 1 (15) = 25 tasks
- Delivers functional Chi middleware with payment gating

**Full Feature Set**: 36 tasks
- All user stories + examples + documentation

---

## Notes

- [P] tasks = different files or independent operations, no dependencies
- [Story] label maps task to specific user story (US1, US2, US3) for traceability
- Each user story should be independently testable (validation checkpoints provided)
- Tests are NOT included per spec (not requested) - can be added separately if needed
- Helper extraction (Phase 2) is CRITICAL - establishes shared logic for all middleware
- Chi middleware reuses 80%+ code via shared helpers (Constitution Principle V: Code Conciseness)
- Avoid: duplicate code in Chi package, breaking existing middleware during refactoring
- Commit after each logical task group (e.g., after extracting all helpers, after US1 complete)
