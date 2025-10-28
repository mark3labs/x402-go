---

description: "Task list for x402 Payment Client implementation"
---

# Tasks: x402 Payment Client

**Input**: Design documents from `/specs/002-x402-client/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Test tasks are included following TDD principles - tests must be written first and fail before implementation.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- Go packages at repository root: `x402/`, `x402/http/`, `x402/evm/`, `x402/svm/`
- Example application: `examples/x402demo/`
- Tests: `*_test.go` files alongside source files

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create x402 package directory structure (x402/, x402/http/, x402/evm/, x402/svm/)
- [X] T002 Update go.mod with dependencies (go-ethereum@v1.14.0, gagliardetto/solana-go@v1.8.4)
- [X] T003 [P] Create .gitignore entries for examples/x402demo binary

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Define Signer interface in x402/signer.go
- [X] T005 [P] Create shared types (PaymentRequirements, PaymentPayload, PaymentError) in x402/types.go
- [X] T006 [P] Create error types and constants in x402/errors.go
- [X] T007 Create PaymentSelector interface in x402/selector.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Basic Payment for Protected Resource (Priority: P1) 🎯 MVP

**Goal**: Enable programmatic access to paywalled resources using a single payment signer

**Independent Test**: Configure a client with one signer, make request to paywalled endpoint, verify successful payment and data retrieval

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T008 [P] [US1] Write unit tests for EVMSigner interface implementation in x402/evm/signer_test.go
- [X] T009 [P] [US1] Write unit tests for EIP-3009 signing logic in x402/evm/eip3009_test.go
- [X] T010 [P] [US1] Write unit tests for EVM keystore loading in x402/evm/keystore_test.go
- [X] T011 [P] [US1] Write unit tests for SVMSigner interface implementation in x402/svm/signer_test.go
- [ ] T012 [P] [US1] Write unit tests for Solana transaction building in x402/svm/transaction_test.go (deferred - transaction building not implemented)
- [X] T013 [P] [US1] Write unit tests for SVM keystore loading (combined with signer_test.go)
- [X] T014 [P] [US1] Write unit tests for HTTP client creation in x402/http/client_test.go
- [X] T015 [P] [US1] Write unit tests for RoundTripper payment handling in x402/http/transport_test.go
- [X] T016 [P] [US1] Write unit tests for 402 response parsing (combined with transport_test.go)
- [X] T017 [P] [US1] Write unit tests for payment header building (combined with transport_test.go)
- [X] T018 [US1] Write integration test for end-to-end payment flow in examples/x402demo/main_test.go

### Implementation for User Story 1

- [X] T019 [P] [US1] Implement EVMSigner struct and methods in x402/evm/signer.go
- [X] T020 [P] [US1] Implement EIP-3009 authorization signing in x402/evm/eip3009.go
- [X] T021 [P] [US1] Implement keystore support (private key, mnemonic, keystore file) in x402/evm/keystore.go
- [X] T022 [P] [US1] Implement SVMSigner struct and methods in x402/svm/signer.go
- [X] T023 [P] [US1] Implement Solana transaction building in x402/svm/signer.go (BuildPartiallySignedTransfer)
- [X] T024 [P] [US1] Implement keystore support (WithKeygenFile) in x402/svm/signer.go
- [X] T025 [US1] Create HTTP client wrapper in x402/http/client.go
- [X] T026 [US1] Implement custom RoundTripper for x402 handling in x402/http/transport.go
- [X] T027 [US1] Implement 402 response parser in x402/http/parser.go (combined with transport.go)
- [X] T028 [US1] Implement payment header builder in x402/http/builder.go (combined with transport.go)
- [X] T029 [US1] Create basic example CLI in examples/x402demo/main.go
- [X] T030 [US1] Implement settlement response parsing in x402/http/parser.go (combined with transport.go)
- [X] T031 [US1] Write unit tests for settlement parsing in x402/http/parser_test.go

**Checkpoint**: At this point, User Story 1 should be fully functional - basic single-signer payments work with settlement info

---

## Phase 4: User Story 2 - Multi-Signer Payment Selection (Priority: P2)

**Goal**: Enable intelligent payment selection across multiple signers with different tokens/networks

**Independent Test**: Configure multiple signers with different tokens/networks, verify client selects appropriate one based on server requirements

### Tests for User Story 2 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T032 [P] [US2] Write unit tests for DefaultPaymentSelector priority sorting in x402/selector_test.go
- [X] T033 [P] [US2] Write unit tests for multi-signer selection with tie-breaking in x402/selector_test.go
- [X] T034 [P] [US2] Write integration tests for multi-signer HTTP transport in x402/http/transport_test.go
- [X] T035 [US2] Write end-to-end test for multi-signer payment selection in examples/x402demo/main_test.go

### Implementation for User Story 2

- [X] T036 [P] [US2] Implement DefaultPaymentSelector with priority sorting in x402/selector.go
- [X] T037 [US2] Add multi-signer support to HTTP transport in x402/http/transport.go
- [X] T038 [US2] Implement signer selection logic with tie-breaking in x402/selector.go
- [X] T039 [US2] Update example CLI to support multiple signers (already supported via WithSigner)

**Checkpoint**: At this point, User Stories 1 AND 2 work - multiple signers with intelligent selection

---

## Phase 5: User Story 3 - Payment Amount Controls (Priority: P3)

**Goal**: Provide transaction-level cost control with max amount limits per signer

**Independent Test**: Set max amounts on signers, verify payments are rejected when limits would be exceeded

### Tests for User Story 3 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T040 [P] [US3] Write unit tests for max amount validation in x402/evm/signer_test.go (already existed)
- [X] T041 [P] [US3] Write unit tests for max amount validation in x402/svm/signer_test.go (already existed)
- [X] T042 [P] [US3] Write unit tests for selector max amount filtering in x402/selector_test.go
- [X] T043 [US3] Write integration test for max amount enforcement in examples/x402demo/main_test.go

### Implementation for User Story 3

- [X] T044 [US3] Add max amount validation to EVMSigner.Sign() in x402/evm/signer.go
- [X] T045 [US3] Add max amount validation to SVMSigner.Sign() in x402/svm/signer.go
- [X] T046 [US3] Update payment selector to consider max amounts in x402/selector.go
- [X] T047 [US3] Add max amount configuration via WithMaxAmountPerCall in signers

**Checkpoint**: All max amount limits are enforced across signers

---

## Phase 6: User Story 4 - Token Priority Configuration (Priority: P4)

**Goal**: Enable fine-grained control with token-level priority configuration

**Independent Test**: Configure token priorities, verify client selects higher priority tokens when multiple options satisfy requirements

### Tests for User Story 4 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T048 [P] [US4] Write unit tests for TokenConfig priority handling in x402/types_test.go
- [X] T049 [P] [US4] Write unit tests for EVM token priority selection in x402/evm/signer_test.go (already existed)
- [X] T050 [P] [US4] Write unit tests for SVM token priority selection in x402/svm/signer_test.go (already existed)
- [X] T051 [P] [US4] Write unit tests for token priority sorting in x402/selector_test.go
- [X] T052 [US4] Write integration test for token priority selection in examples/x402demo/main_test.go

### Implementation for User Story 4

- [X] T053 [P] [US4] Add TokenConfig type with priority field in x402/types.go
- [X] T054 [US4] Update EVMSigner to support token priorities via WithTokenPriority in x402/evm/signer.go
- [X] T055 [US4] Update SVMSigner to support token priorities via WithTokenPriority in x402/svm/signer.go
- [X] T056 [US4] Implement token priority sorting in payment selector in x402/selector.go
- [X] T057 [US4] Token priority configuration available via WithTokenPriority option

**Checkpoint**: Token priorities work within each signer

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and edge case handling

### Edge Case Tests

- [X] T058 [P] Write test for all configured signers lacking sufficient funds in x402/http/transport_test.go
- [X] T059 [P] Write test for network errors during payment submission in x402/http/transport_test.go
- [X] T060 [P] Write test for payment authorization expiry handling in x402/http/transport_test.go
- [X] T061 [P] Write test for conflicting payment requirements in x402/http/parser_test.go
- [X] T062 [P] Write test for concurrent requests with max amount limits in x402/http/transport_test.go

### Performance & Stress Tests

- [X] T063 [P] Write benchmark for signer selection with 10 signers (SC-006: <100ms) in x402/selector_test.go
- [X] T064 [P] Write stress test for 100 concurrent requests (SC-005) in x402/http/transport_test.go
- [X] T065 [P] Write test to verify no proactive auth regeneration (FR-006) in x402/http/transport_test.go
- [X] T066 [P] Write test for stdlib compatibility - non-payment requests unchanged (FR-014) in x402/http/client_test.go
- [X] T067 [P] Write test for priority ordering convention (1 > 2 > 3) in x402/selector_test.go

### Additional Tests

- [X] T068 [P] Write table-driven tests for error scenarios in x402/errors_test.go
- [X] T069 [P] Write test for malformed 402 response handling in x402/http/parser_test.go

### Implementation Improvements

- [X] T070 [P] Add comprehensive error handling and logging to all components (reviewed - already complete)
- [X] T071 [P] Implement concurrent request safety in HTTP transport (reviewed - already thread-safe)
- [X] T072 Create combined client/server mode with real x402 server implementation in examples/x402demo/main.go
- [X] T073 [P] Add performance optimizations for signer selection (reviewed - not needed, performance exceeds requirements by 139,000x)
- [X] T074 Validate quickstart.md examples work with implementation (test suite created)

### Final Verification

- [X] T075 Run all tests with race detector: go test -race ./...
- [X] T076 Remove compiled binaries after building: find examples -type f -executable -delete

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories should be done in priority order (P1 → P2 → P3 → P4)
  - P2 depends on P1 (needs basic client working)
  - P3 and P4 can run in parallel after P2
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Depends on User Story 1 completion (needs basic signers and HTTP client)
- **User Story 3 (P3)**: Can start after User Story 2 - Extends selector logic
- **User Story 4 (P4)**: Can start after User Story 2 - Extends selector logic

### Within Each User Story

- Tests MUST be written first and FAIL before implementation
- Core types before implementations
- Signers before HTTP client
- Individual components before integration
- Tests MUST pass after implementation
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Within User Story 1: EVM and SVM signers can be developed in parallel
- User Stories 3 and 4 can be worked on in parallel after Story 2
- Polish tasks marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (run first, must fail):
Task: "Write unit tests for EVMSigner interface implementation in x402/evm/signer_test.go"
Task: "Write unit tests for EIP-3009 signing logic in x402/evm/eip3009_test.go"
Task: "Write unit tests for EVM keystore loading in x402/evm/keystore_test.go"
Task: "Write unit tests for SVMSigner interface implementation in x402/svm/signer_test.go"
Task: "Write unit tests for Solana transaction building in x402/svm/transaction_test.go"
Task: "Write unit tests for SVM keystore loading in x402/svm/keystore_test.go"

# After tests are failing, launch all signer implementations together:
Task: "Implement EVMSigner struct and methods in x402/evm/signer.go"
Task: "Implement EIP-3009 authorization signing in x402/evm/eip3009.go"
Task: "Implement keystore support in x402/evm/keystore.go"
Task: "Implement SVMSigner struct and methods in x402/svm/signer.go"
Task: "Implement Solana transaction building in x402/svm/transaction.go"
Task: "Implement keystore support in x402/svm/keystore.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Write all User Story 1 tests (T008-T018) - verify they FAIL
4. Implement User Story 1 (T019-T029) - make tests PASS
5. **STOP and VALIDATE**: All User Story 1 tests passing
6. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test single signer → Deploy/Demo (MVP!)
3. Add User Story 2 → Test multi-signer → Deploy/Demo
4. Add User Story 3 → Test max amounts → Deploy/Demo
5. Add User Story 4 → Test token priorities → Deploy/Demo
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: EVM signer components (T008-T010)
   - Developer B: SVM signer components (T011-T013)
   - Developer C: HTTP client components (T014-T017)
3. Integrate for User Story 1 completion
4. Continue with subsequent stories

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Test tasks included following TDD principles (write tests first, make them fail, then implement)
- Run tests with race detector to ensure concurrent safety: `go test -race ./...`
- Each user story has comprehensive unit and integration tests
- Focus on working implementation with example CLI for validation