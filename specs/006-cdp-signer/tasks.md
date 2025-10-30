# Tasks: Coinbase CDP Signer Integration

**Input**: Design documents from `/specs/006-cdp-signer/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/signer-api.yaml, quickstart.md

**Tests**: Tests are included as this is a library package that requires comprehensive test coverage (>80% target per plan.md).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US5)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md structure:
- Source files: `signers/coinbase/`
- Test files: `signers/coinbase/*_test.go`
- Documentation: `specs/006-cdp-signer/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic package structure

- [X] T001 Create signers/coinbase/ package directory structure
- [X] T002 Add gopkg.in/square/go-jose.v2 dependency to go.mod
- [X] T003 [P] Create signers/coinbase/errors.go with CDPError type
- [X] T004 [P] Create signers/coinbase/networks.go with network mapping structures
- [X] T005 [P] Create signers/coinbase/networks_test.go test file skeleton

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Network Mapping Infrastructure

- [X] T006 [P] Implement network type enumeration (NetworkTypeEVM, NetworkTypeSVM) in signers/coinbase/networks.go
- [X] T007 [P] Implement x402-to-CDP network mapping (getCDPNetwork) in signers/coinbase/networks.go
- [X] T008 [P] Implement network type detection (getNetworkType) in signers/coinbase/networks.go
- [X] T009 [P] Implement EVM chain ID mapping (getChainID) in signers/coinbase/networks.go
- [X] T010 Test network mapping functions in signers/coinbase/networks_test.go

### Authentication Infrastructure

- [X] T011 Create signers/coinbase/auth.go with CDPAuth struct definition
- [X] T012 Implement CDPAuth PEM key parsing and validation in signers/coinbase/auth.go
- [X] T013 Implement GenerateBearerToken JWT generation (2min expiration) in signers/coinbase/auth.go
- [X] T014 Implement GenerateWalletAuthToken JWT generation (1min expiration, reqHash) in signers/coinbase/auth.go
- [X] T015 Create signers/coinbase/auth_test.go with test cases for JWT generation
- [X] T016 [P] Test CDPAuth constructor with valid credentials in signers/coinbase/auth_test.go
- [X] T017 [P] Test CDPAuth constructor with invalid PEM format in signers/coinbase/auth_test.go
- [X] T018 [P] Test GenerateBearerToken output structure and claims in signers/coinbase/auth_test.go
- [X] T019 [P] Test GenerateWalletAuthToken with body hash in signers/coinbase/auth_test.go

### HTTP Client Infrastructure

- [X] T020 Create signers/coinbase/client.go with CDPClient struct definition
- [X] T021 Implement CDPClient constructor with HTTP client configuration in signers/coinbase/client.go
- [X] T022 Implement doRequest method with header injection in signers/coinbase/client.go
- [X] T023 Implement error classification logic (retryable vs non-retryable) in signers/coinbase/client.go
- [X] T024 Implement exponential backoff calculation in signers/coinbase/client.go
- [X] T025 Implement doRequestWithRetry method with retry logic in signers/coinbase/client.go
- [X] T026 Create signers/coinbase/client_test.go with mock HTTP server setup
- [X] T027 [P] Test doRequest with successful response in signers/coinbase/client_test.go
- [X] T028 [P] Test doRequest with authentication headers in signers/coinbase/client_test.go
- [X] T029 [P] Test error classification for 4xx errors in signers/coinbase/client_test.go
- [X] T030 [P] Test error classification for 5xx errors in signers/coinbase/client_test.go
- [X] T031 [P] Test retry logic with rate limit (429) in signers/coinbase/client_test.go
- [X] T032 [P] Test retry logic exhaustion (max attempts) in signers/coinbase/client_test.go
- [X] T033 [P] Test exponential backoff calculation in signers/coinbase/client_test.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 5 - Account Creation and Retrieval (Priority: P1) 🎯 MVP Component

**Goal**: Enable developers to create or retrieve CDP accounts programmatically, handling both first-time setup and subsequent initialization automatically.

**Independent Test**: Call CreateOrGetAccount with valid credentials on fresh account (creates new), then call again (retrieves existing), verify both work correctly and no duplicates are created.

### Tests for User Story 5

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T034 [P] [US5] Contract test for CreateOrGetAccount success case in signers/coinbase/account_test.go
- [X] T035 [P] [US5] Contract test for CreateOrGetAccount with existing account in signers/coinbase/account_test.go
- [X] T036 [P] [US5] Test CreateOrGetAccount with invalid credentials in signers/coinbase/account_test.go
- [X] T037 [P] [US5] Test CreateOrGetAccount with unsupported network in signers/coinbase/account_test.go
- [X] T038 [P] [US5] Test CreateOrGetAccount idempotency (repeated sequential calls) in signers/coinbase/account_test.go

### Implementation for User Story 5

- [X] T039 [US5] Create signers/coinbase/account.go with CDPAccount struct definition
- [X] T040 [US5] Implement CreateOrGetAccount helper function with GET-then-POST pattern (query existing accounts via GET /accounts, create via POST /accounts only if none exist) in signers/coinbase/account.go
- [X] T041 [US5] Implement EVM account creation endpoint call in signers/coinbase/account.go
- [X] T042 [US5] Implement SVM account creation endpoint call in signers/coinbase/account.go
- [X] T043 [US5] Implement account retrieval logic (list existing accounts) in signers/coinbase/account.go
- [X] T044 [US5] Add validation for account response format in signers/coinbase/account.go

**Checkpoint**: At this point, User Story 5 should be fully functional - accounts can be created and retrieved programmatically

---

## Phase 4: User Story 3 - Secure Credential Management (Priority: P1)

**Goal**: Enable developers to configure CDP credentials securely via environment variables without exposing secrets in logs or error messages.

**Independent Test**: Set credentials via environment variables, initialize signer, verify credentials never appear in logs or errors, test with invalid credentials to ensure no credential leakage.

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T045 [P] [US3] Test signer initialization with valid environment variables in signers/coinbase/signer_test.go
- [ ] T046 [P] [US3] Test signer initialization with missing CDP_API_KEY_NAME in signers/coinbase/signer_test.go
- [ ] T047 [P] [US3] Test signer initialization with missing CDP_API_KEY_SECRET in signers/coinbase/signer_test.go
- [ ] T048 [P] [US3] Test error messages never contain credential fragments in signers/coinbase/signer_test.go
- [ ] T049 [P] [US3] Test logging sanitization for authorization headers in signers/coinbase/client_test.go

### Implementation for User Story 3

- [ ] T050 [US3] Create signers/coinbase/signer.go with Signer struct definition
- [ ] T051 [US3] Implement SignerOption functional option type in signers/coinbase/signer.go
- [ ] T052 [US3] Implement WithCDPCredentials option with credential validation in signers/coinbase/signer.go
- [ ] T053 [US3] Implement NewSigner constructor with option processing in signers/coinbase/signer.go
- [ ] T054 [US3] Add credential sanitization to error messages in signers/coinbase/signer.go
- [ ] T055 [US3] Add credential sanitization to logging in signers/coinbase/client.go
- [ ] T056 [US3] Implement WithNetwork option with validation in signers/coinbase/signer.go
- [ ] T057 [US3] Implement WithToken option in signers/coinbase/signer.go
- [ ] T058 [US3] Implement WithTokenPriority option in signers/coinbase/signer.go
- [ ] T059 [US3] Implement WithPriority option in signers/coinbase/signer.go
- [ ] T060 [US3] Implement WithMaxAmountPerCall option in signers/coinbase/signer.go

**Checkpoint**: At this point, User Story 3 is complete - credentials can be configured securely without leakage

---

## Phase 5: User Story 1 - EVM Transaction Signing with CDP Wallet (Priority: P1) 🎯 MVP Core

**Goal**: Enable developers to sign x402 payment transactions on EVM chains (Base, Ethereum) using CDP-managed wallets without managing private keys locally.

**Independent Test**: Initialize CDP signer for Base Sepolia, create payment requirement, sign transaction, verify signature is valid and transaction can be broadcast to network.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T061 [P] [US1] Test CanSign with matching EVM network and token in signers/coinbase/signer_test.go
- [ ] T062 [P] [US1] Test CanSign with mismatched network in signers/coinbase/signer_test.go
- [ ] T063 [P] [US1] Test CanSign with mismatched token in signers/coinbase/signer_test.go
- [ ] T064 [P] [US1] Test Sign with valid EVM payment requirement in signers/coinbase/signer_test.go
- [ ] T065 [P] [US1] Test Sign with amount exceeding maxAmount in signers/coinbase/signer_test.go
- [ ] T066 [P] [US1] Test Sign with invalid amount format in signers/coinbase/signer_test.go
- [ ] T067 [P] [US1] Test EVM signature format and structure in signers/coinbase/signer_test.go
- [ ] T068 [P] [US1] Integration test for Base Sepolia signing (skip if no credentials) in signers/coinbase/signer_test.go

### Implementation for User Story 1

- [ ] T069 [US1] Implement Network() interface method in signers/coinbase/signer.go
- [ ] T070 [US1] Implement Scheme() interface method (return "exact") in signers/coinbase/signer.go
- [ ] T071 [US1] Implement GetPriority() interface method in signers/coinbase/signer.go
- [ ] T072 [US1] Implement GetTokens() interface method in signers/coinbase/signer.go
- [ ] T073 [US1] Implement GetMaxAmount() interface method in signers/coinbase/signer.go
- [ ] T074 [US1] Implement Address() helper method in signers/coinbase/signer.go
- [ ] T075 [US1] Implement CanSign validation logic (network, scheme, token matching) in signers/coinbase/signer.go
- [ ] T076 [US1] Implement Sign method skeleton with validation in signers/coinbase/signer.go
- [ ] T077 [US1] Implement EVM payment amount parsing and validation in signers/coinbase/signer.go
- [ ] T078 [US1] Implement EIP-3009 authorization struct building in signers/coinbase/signer.go
- [ ] T079 [US1] Implement EIP-712 typed data construction for EVM in signers/coinbase/signer.go
- [ ] T080 [US1] Implement CDP API call for EVM typed data signing in signers/coinbase/signer.go
- [ ] T081 [US1] Implement EVM PaymentPayload construction from signature in signers/coinbase/signer.go
- [ ] T082 [US1] Add account creation call to NewSigner constructor for EVM networks in signers/coinbase/signer.go

**Checkpoint**: At this point, User Story 1 is complete - EVM signing works end-to-end on Base and Ethereum

---

## Phase 6: User Story 2 - Solana Transaction Signing with CDP Wallet (Priority: P2)

**Goal**: Enable developers to sign x402 payment transactions on Solana using CDP-managed wallets, providing feature parity with EVM support.

**Independent Test**: Initialize CDP signer for Solana devnet, create payment requirement, sign transaction, verify signature is valid Solana format and can be broadcast.

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T083 [P] [US2] Test CanSign with matching SVM network and token in signers/coinbase/signer_test.go
- [ ] T084 [P] [US2] Test Sign with valid SVM payment requirement in signers/coinbase/signer_test.go
- [ ] T085 [P] [US2] Test SVM signature format and structure in signers/coinbase/signer_test.go
- [ ] T086 [P] [US2] Integration test for Solana devnet signing (skip if no credentials) in signers/coinbase/signer_test.go

### Implementation for User Story 2

- [ ] T087 [US2] Implement Solana transaction message building in signers/coinbase/signer.go
- [ ] T088 [US2] Implement Solana TransferChecked instruction construction in signers/coinbase/signer.go
- [ ] T089 [US2] Implement Solana compute budget instruction in signers/coinbase/signer.go
- [ ] T090 [US2] Implement transaction message serialization to base64 in signers/coinbase/signer.go
- [ ] T091 [US2] Implement CDP API call for SVM transaction signing in signers/coinbase/signer.go
- [ ] T092 [US2] Implement SVM PaymentPayload construction from signature in signers/coinbase/signer.go
- [ ] T093 [US2] Add account creation call to NewSigner constructor for SVM networks in signers/coinbase/signer.go
- [ ] T094 [US2] Add SVM-specific validation in CanSign method in signers/coinbase/signer.go

**Checkpoint**: At this point, User Stories 1 AND 2 work independently - both EVM and SVM signing functional

---

## Phase 7: User Story 4 - Error Handling and Retry Logic (Priority: P2)

**Goal**: Ensure production reliability by handling CDP API errors gracefully with automatic retry for transient failures.

**Independent Test**: Simulate various CDP API failures (rate limits, timeouts, 5xx), verify signer retries appropriately with exponential backoff and eventually succeeds or fails gracefully.

### Tests for User Story 4

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T095 [P] [US4] Test retry behavior on rate limit (429) with mock server in signers/coinbase/client_test.go
- [ ] T096 [P] [US4] Test retry behavior on 5xx server error in signers/coinbase/client_test.go
- [ ] T097 [P] [US4] Test immediate failure on 4xx client error in signers/coinbase/client_test.go
- [ ] T098 [P] [US4] Test retry exhaustion returns clear error in signers/coinbase/client_test.go
- [ ] T099 [P] [US4] Test network timeout handling in signers/coinbase/client_test.go
- [ ] T100 [P] [US4] Test Retry-After header respect in signers/coinbase/client_test.go

### Implementation for User Story 4

- [ ] T101 [US4] Enhance CDPError with detailed error context in signers/coinbase/errors.go
- [ ] T102 [US4] Implement error wrapping for better error messages in signers/coinbase/client.go
- [ ] T103 [US4] Add context deadline propagation through retry attempts in signers/coinbase/client.go
- [ ] T104 [US4] Add Retry-After header parsing in signers/coinbase/client.go
- [ ] T105 [US4] Add retry attempt tracking in error messages in signers/coinbase/client.go
- [ ] T106 [US4] Implement request ID extraction from CDP responses in signers/coinbase/client.go

**Checkpoint**: Error handling is robust - transient failures are retried, permanent failures fail fast with clear errors

---

## Phase 8: User Story 6 - Multi-Chain Support (Priority: P3)

**Goal**: Enable developers to use both EVM and SVM signers simultaneously with unified credential management.

**Independent Test**: Initialize both Base and Solana signers with same credentials, send concurrent payment requests to both, verify both process successfully without interference.

### Tests for User Story 6

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T107 [P] [US6] Test multiple signer initialization with same credentials in signers/coinbase/signer_test.go
- [ ] T108 [P] [US6] Integration test with both Base and Solana (skip if no credentials) in signers/coinbase/signer_test.go

### Implementation for User Story 6

- [ ] T109 [US6] Validate multi-chain support in NewSigner (no blocking state) in signers/coinbase/signer.go
- [ ] T110 [US6] Add documentation example for multi-chain usage in signers/coinbase/signer.go
- [ ] T111 [US6] Verify credential reuse across multiple signer instances in signers/coinbase/auth_test.go

**Checkpoint**: Multi-chain support verified - developers can use EVM and SVM simultaneously

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T112 [P] Add GoDoc comments to all exported types in signers/coinbase/signer.go
- [ ] T113 [P] Add GoDoc comments to all exported functions in signers/coinbase/account.go
- [ ] T114 [P] Add GoDoc comments to CDPAuth and CDPClient in signers/coinbase/auth.go and signers/coinbase/client.go
- [ ] T115 [P] Add package-level documentation in signers/coinbase/doc.go
- [ ] T116 Run go fmt on all files in signers/coinbase/
- [ ] T117 Run go vet on signers/coinbase/ package
- [ ] T118 Run golangci-lint on signers/coinbase/ package
- [ ] T119 Run go test -race -cover on signers/coinbase/ package and verify >80% coverage
- [ ] T120 Validate quickstart.md examples compile and run
- [ ] T121 Update AGENTS.md via .specify/scripts/bash/update-agent-context.sh
- [ ] T122 Final integration test with x402 HTTP client using CDP signer

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 5 (Phase 3)**: Depends on Foundational - BLOCKS stories 1, 2, 3 (account creation required)
- **User Story 3 (Phase 4)**: Depends on Foundational - Can run parallel to US5
- **User Story 1 (Phase 5)**: Depends on Foundational + US5 + US3 - EVM signing needs account creation and credentials
- **User Story 2 (Phase 6)**: Depends on Foundational + US5 + US3 - SVM signing needs account creation and credentials
- **User Story 4 (Phase 7)**: Depends on Foundational - Enhances client error handling (can run parallel to US1/US2)
- **User Story 6 (Phase 8)**: Depends on US1 + US2 - Multi-chain requires both EVM and SVM working
- **Polish (Phase 9)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 5 (P1)**: Can start after Foundational - BLOCKS US1 and US2 (account creation prerequisite)
- **User Story 3 (P1)**: Can start after Foundational - BLOCKS US1 and US2 (credential management prerequisite)
- **User Story 1 (P1)**: Depends on US5 + US3 - Core EVM signing
- **User Story 2 (P2)**: Depends on US5 + US3 - Core SVM signing (can run parallel to US1)
- **User Story 4 (P2)**: Can start after Foundational - Enhances error handling (can run parallel to US1/US2)
- **User Story 6 (P3)**: Depends on US1 + US2 - Multi-chain testing

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Foundation components (network mapping, auth, client) before account management
- Account management before signing
- Core signing logic before advanced features
- Story complete and independently testable before moving to next priority

### Parallel Opportunities

**Phase 1 (Setup)**: All tasks marked [P] can run in parallel:
- T003 (errors.go) || T004 (networks.go) || T005 (networks_test.go)

**Phase 2 (Foundational)**: Within each subsection:
- Network mapping tests: T010 (after T006-T009 complete)
- Auth tests: T016 || T017 || T018 || T019 (after T011-T014 complete)
- Client tests: T027 || T028 || T029 || T030 || T031 || T032 || T033 (after T020-T025 complete)

**Phase 3 (US5)**: Test tasks can run in parallel:
- T034 || T035 || T036 || T037 || T038 (all tests before implementation)

**Phase 4 (US3)**: Test tasks can run in parallel:
- T046 || T047 || T048 || T049 || T050 (all tests before implementation)

**Phase 5 (US1)**: Test tasks can run in parallel:
- T062 || T063 || T064 || T065 || T066 || T067 || T068 || T069 || T070 (all tests before implementation)

**Phase 6 (US2)**: Test tasks can run in parallel:
- T085 || T086 || T087 || T088 || T089 (all tests before implementation)

**Phase 7 (US4)**: Test tasks can run in parallel:
- T098 || T099 || T100 || T101 || T102 || T103 (all tests before implementation)

**Phase 8 (US6)**: Test tasks can run in parallel:
- T111 || T112 || T113 (all tests before implementation)

**Phase 9 (Polish)**: Documentation tasks can run in parallel:
- T117 || T118 || T119 || T120 (all GoDoc additions)

**Once Foundational completes**:
- US5 + US3 can start in parallel (different concerns)
- US4 can start in parallel with US3/US5 (different files)

**After US5 + US3 complete**:
- US1 and US2 can run in parallel (EVM vs SVM, different code paths)

---

## Parallel Example: User Story 1 (EVM Signing)

```bash
# Launch all tests for User Story 1 together:
Task: "Test CanSign with matching EVM network" & 
Task: "Test CanSign with mismatched network" &
Task: "Test Sign with valid EVM payment" &
Task: "Test Sign with amount exceeding max" &
Task: "Test Sign with invalid amount" &
Task: "Test EVM signature format" &
Task: "Test concurrent EVM signing" &
Task: "Integration test Base Sepolia"
wait

# After tests fail, implement in sequence (due to dependencies):
Task: "Implement Network() method"
Task: "Implement Scheme() method"
Task: "Implement CanSign validation"
Task: "Implement Sign method skeleton"
Task: "Implement EVM amount parsing"
Task: "Implement EIP-3009 authorization"
Task: "Implement EIP-712 typed data"
Task: "Implement CDP API signing call"
Task: "Implement PaymentPayload construction"
```

---

## Implementation Strategy

### MVP First (User Stories 5, 3, 1 Only)

**Rationale**: Account creation (US5) + secure credentials (US3) + EVM signing (US1) = minimal viable CDP signer

1. Complete Phase 1: Setup (5 tasks, ~30 min)
2. Complete Phase 2: Foundational (28 tasks, ~4-6 hours) - CRITICAL blocker
3. Complete Phase 3: User Story 5 - Account Creation (12 tasks, ~2-3 hours)
4. Complete Phase 4: User Story 3 - Credential Management (11 tasks, ~2 hours)
5. Complete Phase 5: User Story 1 - EVM Signing (24 tasks, ~4-6 hours)
6. **STOP and VALIDATE**: 
   - Test account creation on Base Sepolia
   - Test EVM signing end-to-end
   - Verify credentials never leak
   - Run coverage report (should be >70% at this point)
7. **MVP COMPLETE**: Can ship with EVM-only support for Base and Ethereum

**MVP Delivers**:
- ✅ EVM transaction signing (Base, Ethereum mainnet/testnet)
- ✅ Secure credential management
- ✅ Automatic account creation/retrieval
- ✅ Integration with existing x402 middleware

### Incremental Delivery

1. **MVP (US5 + US3 + US1)** → Test → Deploy/Demo Base support
2. **Add User Story 2 (Solana)** → Test independently → Deploy/Demo multi-chain
3. **Add User Story 4 (Error Handling)** → Test under load → Deploy/Demo production-ready
4. **Add User Story 6 (Multi-Chain)** → Test concurrent usage → Deploy/Demo full feature set
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With 3 developers after Foundational phase:

1. **Team completes Setup + Foundational together** (required baseline)
2. **Once Foundational is done**:
   - Developer A: User Story 5 (Account Creation) - PRIORITY
   - Developer B: User Story 3 (Credential Management) - PRIORITY
   - Developer C: User Story 4 (Error Handling) - Can proceed in parallel
3. **After US5 + US3 complete**:
   - Developer A: User Story 1 (EVM Signing)
   - Developer B: User Story 2 (SVM Signing)
   - Developer C: Continues US4, then US6
4. Stories integrate independently, merge as complete

---

## Notes

- [P] tasks = different files, no dependencies, safe to parallelize
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- **TDD approach**: Write tests first, verify they fail, then implement
- Tests use table-driven pattern following existing x402-go conventions
- Commit after each logical group of tasks (per story phase)
- Integration tests skip if CDP_API_KEY_NAME not set (graceful degradation)
- Stop at any checkpoint to validate story independently
- Coverage target: >80% (measure with `go test -race -cover ./signers/coinbase/`)
- All code must pass `go vet` and `golangci-lint` before PR

---

## Quality Gates

Before considering feature complete:

- [ ] All unit tests pass with `-race` flag
- [ ] Test coverage >80%
- [ ] All linting passes (go vet + golangci-lint)
- [ ] Integration tests pass on Base Sepolia and Solana Devnet
- [ ] quickstart.md examples compile and run successfully
- [ ] No credential leakage in logs (security scan)
- [ ] Documentation complete (GoDoc on all exported items)
- [ ] Signer integrates with x402 HTTP client without code changes
