# Tasks: MCP Integration

**Input**: Design documents from `/specs/007-mcp-integration/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are included as requested in the original specification ("Add tests").

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: Repository root with subpackages
- Main code: `mcp/client/`, `mcp/server/` at repository root
- Examples: `examples/mcp/`
- Tests: Alongside source files with `_test.go` suffix

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create MCP package directory structure at mcp/ in repository root
- [ ] T002 [P] Create mcp/types.go for MCP-specific type aliases and constants
- [ ] T003 [P] Create mcp/errors.go for MCP-specific error types
- [ ] T004 [P] Create mcp/client and mcp/server subdirectories
- [ ] T005 Update go.mod to include github.com/mark3labs/mcp-go dependency

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T006 Create base transport interface wrapper in mcp/client/base.go
- [ ] T007 Create base server wrapper interface in mcp/server/base.go
- [ ] T008 [P] Create payment context structure in mcp/types.go for request lifecycle
- [ ] T009 [P] Define MCP-specific constants for payment metadata keys in mcp/types.go
- [ ] T010 Create payment requirement builder helpers in mcp/server/requirements.go
- [ ] T010a [P] Implement 5-second verification timeout constant in mcp/types.go (FR-017)
- [ ] T010b [P] Implement 60-second settlement timeout constant in mcp/types.go (FR-018)
- [ ] T010c Create timeout context wrapper utilities in mcp/client/transport.go
- [ ] T010d Create timeout context wrapper utilities in mcp/server/facilitator.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - MCP Client with x402 Payments (Priority: P1) 🎯 MVP

**Goal**: Enable MCP clients to automatically handle x402 payment requirements when interacting with paid MCP servers

**Independent Test**: Client connects to x402 MCP server, receives payment requirements, signs payments, and accesses paid tools

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T011 [P] [US1] Test X402Transport initialization with multiple signers in mcp/client/transport_test.go
- [ ] T012 [P] [US1] Test payment handler orchestration with fallback logic in mcp/client/handler_test.go
- [ ] T013 [P] [US1] Test 402 error detection and payment flow in mcp/client/transport_test.go
- [ ] T014 [P] [US1] Test payment injection into params._meta["x402/payment"] field per MCP spec in mcp/client/transport_test.go
- [ ] T015 [P] [US1] Test concurrent payment handling with independent payments per request (FR-016) in mcp/client/transport_test.go
- [ ] T016 [P] [US1] Test free tool access without payment in mcp/client/transport_test.go
- [ ] T016a [P] [US1] Test that 10 concurrent requests each generate unique payment proofs (FR-016) in mcp/client/transport_test.go

### Implementation for User Story 1

- [ ] T017 [US1] Implement X402Transport type in mcp/client/transport.go implementing transport.Interface
- [ ] T018 [US1] Implement payment handler orchestration in mcp/client/handler.go with signer selection
- [ ] T019 [US1] Add JSON-RPC 402 error detection in mcp/client/transport.go
- [ ] T020 [US1] Implement payment requirement matching logic in mcp/client/handler.go
- [ ] T021 [US1] Add payment injection into params._meta["x402/payment"] field per MCP spec in mcp/client/transport.go
- [ ] T022 [US1] Implement payment event callbacks in mcp/client/transport.go
- [ ] T023 [US1] Add session management and protocol negotiation in mcp/client/transport.go
- [ ] T024 [US1] Implement MCP protocol negotiation and transport selection in mcp/client/transport.go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - MCP Server with x402 Protection (Priority: P1)

**Goal**: Enable MCP servers to protect their tools with x402 payment requirements and verify payments

**Independent Test**: Server exposes free and paid tools, sends payment requirements for paid tools, validates payments via facilitator

### Tests for User Story 2 ⚠️

- [ ] T025 [P] [US2] Test X402Server initialization with tool configuration in mcp/server/server_test.go
- [ ] T026 [P] [US2] Test middleware payment extraction from params._meta["x402/payment"] field per MCP spec in mcp/server/middleware_test.go
- [ ] T027 [P] [US2] Test 402 error generation with payment requirements in mcp/server/server_test.go
- [ ] T028 [P] [US2] Test facilitator payment verification in mcp/server/middleware_test.go
- [ ] T029 [P] [US2] Test mixed free/paid tool handling in mcp/server/server_test.go
- [ ] T030 [P] [US2] Test settlement response in result._meta in mcp/server/middleware_test.go
- [ ] T030a [P] [US2] Test non-refundable payment when tool execution fails after verification (FR-015) in mcp/server/server_test.go

### Implementation for User Story 2

- [ ] T031 [US2] Implement X402Server wrapper in mcp/server/server.go wrapping mcp.MCPServer
- [ ] T032 [US2] Implement payment middleware in mcp/server/middleware.go for tool interception
- [ ] T033 [US2] Add tool payment configuration methods in mcp/server/server.go (AddPayableTool)
- [ ] T034 [US2] Implement payment extraction from params._meta["x402/payment"] field per MCP spec in mcp/server/middleware.go
- [ ] T035 [US2] Add facilitator client integration with 5s verify and 60s settle timeouts (FR-017, FR-018) in mcp/server/facilitator.go
- [ ] T036 [US2] Implement 402 JSON-RPC error generation with code:402 and PaymentRequirementsResponse in error.data per MCP spec in mcp/server/server.go
- [ ] T037 [US2] Add settlement response injection in result._meta["x402/payment-response"] field per MCP spec in mcp/server/middleware.go
- [ ] T038 [US2] Implement verify-only mode support in mcp/server/server.go

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Multi-Chain Payment Support (Priority: P2)

**Goal**: Support payments using different blockchain networks with automatic selection

**Independent Test**: Client with multiple signers selects optimal payment option based on server requirements

### Tests for User Story 3 ⚠️

- [ ] T039 [P] [US3] Test DefaultPaymentSelector priority algorithm in mcp/client/handler_test.go
- [ ] T040 [P] [US3] Test EVM signer integration with MCP in mcp/client/handler_test.go
- [ ] T041 [P] [US3] Test Solana signer integration with MCP in mcp/client/handler_test.go
- [ ] T042 [P] [US3] Test multi-network payment requirement matching in mcp/client/handler_test.go
- [ ] T043 [P] [US3] Test fallback when primary network insufficient balance in mcp/client/handler_test.go
- [ ] T043a [P] [US3] Test that payment fallback completes within 5 seconds (SC-003) in mcp/client/handler_test.go

### Implementation for User Story 3

- [ ] T044 [US3] Integrate DefaultPaymentSelector from x402 in mcp/client/handler.go
- [ ] T045 [US3] Add EVM payment creation support in mcp/client/handler.go
- [ ] T046 [US3] Add Solana payment creation support in mcp/client/handler.go
- [ ] T047 [US3] Implement network-specific payment validation in mcp/server/middleware.go
- [ ] T048 [US3] Add multi-network requirement helpers in mcp/server/requirements.go (RequireUSDCBase, RequireUSDCPolygon, RequireUSDCSolana)

**Checkpoint**: Multi-chain payment support should now be fully functional

---

## Phase 6: User Story 4 - Example Implementation (Priority: P3)

**Goal**: Provide working examples demonstrating client and server implementations

**Independent Test**: Example runs in both client and server modes with successful payment flows

### Tests for User Story 4 ⚠️

- [ ] T049 [P] [US4] Test example server mode startup in examples/mcp/main_test.go
- [ ] T050 [P] [US4] Test example client mode connection in examples/mcp/main_test.go
- [ ] T051 [P] [US4] Test example payment flow end-to-end in examples/mcp/main_test.go

### Implementation for User Story 4

- [ ] T052 [US4] Create examples/mcp directory and go.mod
- [ ] T053 [US4] Implement main.go with client/server mode selection in examples/mcp/main.go
- [ ] T054 [US4] Add server mode with free and paid tools in examples/mcp/main.go
- [ ] T055 [US4] Add client mode with payment signers in examples/mcp/main.go
- [ ] T056 [US4] Implement echo tool handler (free) in examples/mcp/main.go
- [ ] T057 [US4] Implement premium search tool handler (paid) in examples/mcp/main.go
- [ ] T058 [P] [US4] Create README.md with usage instructions in examples/mcp/README.md
- [ ] T059 [US4] Add environment variable configuration support in examples/mcp/main.go

**Checkpoint**: Example implementation should demonstrate all features

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T060 [P] Add comprehensive error handling and logging across all components
- [ ] T061 [P] Implement request/response tracing for debugging
- [ ] T062 [P] Add metrics collection hooks for monitoring
- [ ] T063 [P] Create integration tests covering all user stories in mcp/integration_test.go
- [ ] T064 [P] Add benchmark tests for concurrent payment handling in mcp/benchmark_test.go
- [ ] T065 Run quickstart.md validation and update if needed
- [ ] T066 Ensure all tests pass with race detector enabled
- [ ] T067 Verify example builds and runs successfully
- [ ] T068 [P] Test behavior when all configured signers fail in mcp/client/handler_test.go
- [ ] T069 [P] Test network timeout handling during payment verification in mcp/server/middleware_test.go
- [ ] T070 [P] Test behavior when facilitator is unavailable after valid payment in mcp/server/middleware_test.go
- [ ] T071 [P] Test client handling of malformed payment requirements from server in mcp/client/transport_test.go
- [ ] T072 [P] Test behavior when server requirements exceed all client limits in mcp/client/handler_test.go

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - US1 and US2 are both P1 priority and can proceed in parallel
  - US3 (P2) can start after Foundational, integrates with US1
  - US4 (P3) depends on US1 and US2 being complete
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational - No dependencies on other stories
- **User Story 3 (P2)**: Can start after Foundational - Enhances US1 but independently testable
- **User Story 4 (P3)**: Depends on US1 and US2 - Demonstrates their functionality

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Base types before complex implementations
- Core logic before integration points
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- Foundational tasks T008, T009 can run in parallel
- US1 and US2 can be developed in parallel (both P1)
- All tests within a user story can run in parallel
- Polish phase tasks can all run in parallel

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
Task: "Test X402Transport initialization with multiple signers in mcp/client/transport_test.go"
Task: "Test payment handler orchestration with fallback logic in mcp/client/handler_test.go"
Task: "Test 402 error detection and payment flow in mcp/client/transport_test.go"
Task: "Test payment injection into params._meta in mcp/client/transport_test.go"
Task: "Test concurrent payment handling in mcp/client/transport_test.go"
Task: "Test free tool access without payment in mcp/client/transport_test.go"
```

---

## Parallel Example: User Story 2

```bash
# Launch all tests for User Story 2 together:
Task: "Test X402Server initialization with tool configuration in mcp/server/server_test.go"
Task: "Test middleware payment extraction from params._meta in mcp/server/middleware_test.go"
Task: "Test 402 error generation with payment requirements in mcp/server/server_test.go"
Task: "Test facilitator payment verification in mcp/server/middleware_test.go"
Task: "Test mixed free/paid tool handling in mcp/server/server_test.go"
Task: "Test settlement response in result._meta in mcp/server/middleware_test.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3 & 4 in parallel: User Story 1 (Client) and User Story 2 (Server)
4. **STOP and VALIDATE**: Test client-server interaction
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 & 2 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 3 → Test multi-chain → Deploy/Demo
4. Add User Story 4 → Provide examples → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Client)
   - Developer B: User Story 2 (Server)
   - Developer C: Can start User Story 3 tests
3. After US1 & US2 complete:
   - Any developer: User Story 4 (Examples)
   - All developers: Polish phase

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Tests included as requested in original specification
- US1 and US2 are both P1 priority and form the MVP together
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Total tasks: 78 (excluding parallel markers) - includes timeout, concurrent payment, edge case, and non-refundable payment tasks