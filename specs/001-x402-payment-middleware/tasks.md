---

description: "Task list for x402 Payment Middleware implementation"
---

# Tasks: x402 Payment Middleware

**Input**: Design documents from `/specs/001-x402-payment-middleware/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/facilitator.yaml

**Tests**: Tests are included as mandated by Constitution Principle III (Test-First Development) which requires all new features to have tests written before implementation, and Principle II (Test Coverage Preservation) which requires maintaining or improving test coverage.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Library structure**: Root package at repository root, subpackages in directories
- Core types in root `x402` package
- HTTP middleware in `http/` subpackage
- Examples in `examples/`
- Tests alongside source files with `_test.go` suffix

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create project structure per implementation plan in plan.md
- [X] T002 Initialize Go module github.com/mark3labs/x402-go with go mod init
- [X] T003 [P] Create directory structure: http/, examples/, testdata/
- [X] T004 [P] Setup .gitignore for Go project with common patterns

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core types and error definitions that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Create base types in types.go (PaymentRequirement, PaymentRequirementsResponse)
- [X] T006 [P] Create error definitions in errors.go (x402 standard error codes)
- [X] T007 [P] Write unit tests for types validation in types_test.go
- [X] T008 [P] Write unit tests for error handling in errors_test.go
- [X] T009 Create shared validation functions for addresses and amounts in types.go
- [X] T010 [P] Create test fixtures in testdata/evm/ and testdata/svm/ directories

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Basic Middleware Integration (Priority: P1) 🎯 MVP

**Goal**: Developers can protect HTTP endpoints with payment requirements using the x402 standard

**Independent Test**: Set up a simple HTTP server with the middleware, make requests without payment (returns 402), and with valid payment (returns success)

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T011 [P] [US1] Write middleware unit tests for 402 response in http/middleware_test.go
- [X] T012 [P] [US1] Write handler tests for payment validation in http/handler_test.go
- [X] T013 [P] [US1] Write facilitator client tests with mocks in http/facilitator_test.go
- [X] T014 [P] [US1] Create integration test for full payment flow in http/middleware_test.go

### Implementation for User Story 1

- [X] T015 [US1] Create PaymentPayload and SchemePayload types in types.go
- [X] T016 [P] [US1] Create SettlementResponse type in types.go
- [X] T017 [P] [US1] Create MiddlewareConfig type in http/middleware.go
- [X] T018 [US1] Implement NewX402Middleware function with closure pattern in http/middleware.go
- [X] T019 [US1] Implement payment requirement response handler in http/handler.go
- [X] T020 [US1] Implement X-PAYMENT header parsing and validation in http/handler.go
- [X] T021 [P] [US1] Create FacilitatorClient with verify method in http/facilitator.go
- [X] T022 [US1] Implement payment verification flow in http/handler.go
- [X] T023 [US1] Implement payment settlement flow in http/facilitator.go
- [X] T024 [US1] Add X-PAYMENT-RESPONSE header generation in http/handler.go
- [X] T025 [P] [US1] Create basic example in examples/basic/main.go
- [X] T026 [US1] Add error handling for malformed headers in http/handler.go
- [X] T027 [US1] Add logging with slog for payment operations in http/middleware.go

**Checkpoint**: User Story 1 is fully functional - basic middleware works with single chain payments

---

## Phase 4: User Story 2 - Multi-Chain Payment Configuration (Priority: P2)

**Goal**: Accept payments on multiple blockchain networks (both EVM and SVM chains) for the same resource

**Independent Test**: Configure middleware with multiple payment options (e.g., USDC on Base and Solana), verify that payment requirements list all options, and that payments work on each chain independently

### Tests for User Story 2 ⚠️

- [X] T028 [P] [US2] Write tests for multiple payment requirements in http/middleware_test.go
- [X] T029 [P] [US2] Write tests for EVM payload handling in http/handler_test.go
- [X] T030 [P] [US2] Write tests for SVM payload handling in http/handler_test.go
- [X] T031 [P] [US2] Write integration test for multi-chain selection in http/middleware_test.go

### Implementation for User Story 2

- [X] T032 [US2] Create EVMPayload type with EIP-3009 fields in types.go
- [X] T033 [P] [US2] Create SVMPayload type with transaction field in types.go
- [X] T034 [US2] Implement polymorphic payload unmarshaling in http/handler.go
- [X] T035 [US2] Update middleware to support multiple PaymentRequirements in accepts array
- [X] T036 [US2] Implement chain-specific validation logic in http/handler.go
- [X] T037 [P] [US2] Add support for Extra field in PaymentRequirement for SVM fee payer
- [X] T038 [P] [US2] Create multi-chain example in examples/multichain/main.go
- [X] T039 [US2] Update FacilitatorClient to handle both EVM and SVM payloads

**Checkpoint**: User Stories 1 AND 2 work - middleware supports multiple blockchain networks

---

## Phase 5: User Story 3 - Custom Payment Requirements per Route (Priority: P3)

**Goal**: Set different payment amounts and configurations for different routes

**Independent Test**: Set up multiple handlers wrapped with different payment middleware configurations, verify each handler returns its specific requirements, and that payments are validated against the correct amounts

### Tests for User Story 3 ⚠️

- [X] T040 [P] [US3] Write tests for route-specific configurations in http/middleware_test.go
- [X] T041 [P] [US3] Write tests for amount validation per route in http/handler_test.go
- [X] T042 [P] [US3] Write integration test for multiple middleware instances in http/middleware_test.go

### Implementation for User Story 3

- [X] T043 [US3] Refactor middleware to support per-instance configuration in http/middleware.go
- [X] T044 [US3] Update Config to include Resource and Description fields in http/middleware.go
- [X] T045 [US3] Implement route-specific requirement generation in http/handler.go
- [X] T046 [US3] Add validation for payment amount against route requirements
- [X] T047 [P] [US3] Create route-specific pricing example in examples/multichain/main.go
- [X] T048 [US3] Update handler to use route-specific config from middleware context

**Checkpoint**: All priority user stories (P1-P3) are functional - full flexibility in payment configuration

---

## Phase 6: User Story 4 - Payment Verification without Settlement (Priority: P4)

**Goal**: Verify payment authorizations before settling them on-chain for custom business logic

**Independent Test**: Configure middleware to use verification-only mode, send valid and invalid payment authorizations, verify that only validation occurs without on-chain settlement

### Tests for User Story 4 ⚠️

- [X] T049 [P] [US4] Write tests for verification-only mode in http/middleware_test.go
- [X] T050 [P] [US4] Write tests for payment context storage in http/handler_test.go
- [X] T051 [P] [US4] Write integration test for verify without settle flow in http/middleware_test.go

### Implementation for User Story 4

- [X] T052 [US4] Add VerifyOnly flag to MiddlewareConfig in http/middleware.go
- [X] T053 [US4] Implement verify-only flow bypassing settlement in http/handler.go
- [X] T054 [US4] Add PaymentContextKey for storing payment info in request context
- [X] T055 [US4] Create PaymentSession type for tracking payment state (VerifyResponse serves this purpose)
- [X] T056 [P] [US4] Create verification-only example in examples/verification/main.go
- [X] T057 [US4] Add programmatic settlement trigger method to FacilitatorClient (Settle method is already exported)
- [X] T058 [US4] Update handler to store verified payment in context for custom logic

**Checkpoint**: All user stories complete - full x402 payment middleware functionality

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T059 [P] Add facilitator /supported endpoint integration in http/facilitator.go
- [X] T060 [P] Implement optional fallback facilitator support per FR-008 in http/facilitator.go
- [X] T061 Add timeout handling with context for facilitator calls (2s timeout per SC-002)
- [X] T062 [P] Add comprehensive godoc documentation to all exported types and functions

- [X] T063 Run go fmt ./... to ensure consistent formatting
- [X] T064 Run go vet ./... to check for common issues
- [X] T065 Validate all examples compile and run correctly
- [X] T066 Verify quickstart.md examples match actual implementation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User stories can proceed in priority order (P1 → P2 → P3 → P4)
  - Or in parallel if team capacity allows
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Builds on US1 types but independently testable
- **User Story 3 (P3)**: Builds on US1 middleware but independently testable
- **User Story 4 (P4)**: Extends US1 with verification mode but independently testable

### Within Each User Story

- Tests MUST be written first and FAIL before implementation
- Types before handlers
- Handlers before middleware integration
- Core implementation before examples
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel (T003, T004)
- All Foundational type/error tasks marked [P] can run in parallel (T006, T007, T008, T010)
- All tests for each user story marked [P] can run in parallel
- Type definitions within a story marked [P] can run in parallel
- Examples marked [P] can be created in parallel with other tasks
- Polish tasks marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task T011: "Write middleware unit tests for 402 response in http/middleware_test.go"
Task T012: "Write handler tests for payment validation in http/handler_test.go"
Task T013: "Write facilitator client tests with mocks in http/facilitator_test.go"
Task T014: "Create integration test for full payment flow in http/middleware_test.go"

# Launch type definitions together:
Task T016: "Create SettlementResponse type in types.go"
Task T017: "Create MiddlewareConfig type in http/middleware.go"

# After core implementation, launch in parallel:
Task T021: "Create FacilitatorClient with verify method in http/facilitator.go"
Task T025: "Create basic example in examples/basic/main.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test basic middleware functionality
5. Release v0.1.0 with basic payment gating

### Incremental Delivery

1. Complete Setup + Foundational → Types and errors ready
2. Add User Story 1 → Test independently → Release v0.1.0 (MVP - single chain payments)
3. Add User Story 2 → Test independently → Release v0.2.0 (multi-chain support)
4. Add User Story 3 → Test independently → Release v0.3.0 (route-specific pricing)
5. Add User Story 4 → Test independently → Release v0.4.0 (verification mode)
6. Each release adds value without breaking previous functionality

### Test-Driven Development

Following the plan.md emphasis on TDD:
1. Write tests first for each user story
2. Ensure tests FAIL before implementation
3. Implement to make tests pass
4. Refactor while keeping tests green

---

## Notes

- [P] tasks = different files, no dependencies within same phase
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Tests use table-driven approach per Go best practices
- Use mock facilitator for tests to avoid blockchain dependencies
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: external dependencies, breaking Go stdlib patterns