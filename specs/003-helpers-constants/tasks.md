# Tasks: Helper Functions and Constants

**Input**: Design documents from `/specs/003-helpers-constants/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/helpers-api.yaml

**Tests**: Tests are included per constitution requirement (Test-First Development, Test Coverage Preservation)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Go library with flat package structure at repository root:
- Core code: `/chains.go`, `/chains_test.go`
- Examples: `/examples/basic/main.go`
- Existing types: `/types.go`, `/errors.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and structure verification

- [X] T001 Verify existing project structure matches plan.md requirements (types.go, errors.go, http/, evm/, svm/ exist)
- [X] T002 [P] Run go mod tidy to ensure Go 1.25.1 environment is ready
- [X] T003 [P] Verify golangci-lint configuration exists and runs successfully

**Checkpoint**: Development environment ready for implementation

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core types and error handling that all user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Define ChainConfig struct in chains.go with fields: NetworkID, USDCAddress, Decimals, EIP3009Name, EIP3009Version
- [X] T005 Define NetworkType constants in chains.go (NetworkTypeEVM, NetworkTypeSVM, NetworkTypeUnknown)
- [X] T006 Define PaymentRequirementConfig struct in chains.go with fields: Chain, Amount, RecipientAddress, Scheme, MaxTimeoutSeconds, MimeType
- [X] T007 [P] Add structured error types in chains.go for parameter validation (use fmt.Errorf with format "parameterName: reason")
- [X] T008 Define all 8 ChainConfig constants in chains.go: SolanaMainnet, SolanaDevnet, BaseMainnet, BaseSepolia, PolygonMainnet, PolygonAmoy, AvalancheMainnet, AvalancheFuji per research.md verified addresses

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Quick Client Setup with Chain Constants (Priority: P1) 🎯 MVP

**Goal**: Developers can quickly configure x402 clients using chain constants and helper functions to create TokenConfig structs with correct USDC addresses and network identifiers.

**Independent Test**: Create a client using chain constants and NewTokenConfig helper, verify client has correct token address and network identifier. Test with Base mainnet, Solana mainnet, and Base Sepolia testnet.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T009 [P] [US1] Create chains_test.go with table-driven test for ChainConfig constants validation (verify all 8 constants have non-empty NetworkID, USDCAddress, Decimals=6)
- [X] T010 [P] [US1] Add table-driven test in chains_test.go for NewTokenConfig helper covering all 8 chain constants with various priority values
- [X] T011 [P] [US1] Add test in chains_test.go verifying TokenConfig has correct Address from ChainConfig.USDCAddress, Symbol="USDC", Decimals=6, Priority matches input

### Implementation for User Story 1

- [X] T012 [US1] Implement NewTokenConfig function in chains.go that accepts ChainConfig and priority int, returns TokenConfig with Address=chain.USDCAddress, Symbol="USDC", Decimals=6, Priority=priority
- [X] T013 [US1] Add GoDoc comments to ChainConfig struct and all 8 chain constant declarations explaining their purpose and verified date (2025-10-28)
- [X] T014 [US1] Add GoDoc comments to NewTokenConfig function explaining parameters and return value
- [X] T015 [US1] Run go test -race -cover ./... and verify all User Story 1 tests pass with no race conditions

**Checkpoint**: User Story 1 complete - developers can configure clients with chain constants

---

## Phase 4: User Story 2 - Quick Middleware Payment Requirements Setup (Priority: P1)

**Goal**: Developers can quickly create PaymentRequirement structs using helper functions with chain constants, amounts, and recipient addresses. Helper automatically handles atomic unit conversion and EIP-3009 domain parameters.

**Independent Test**: Use NewPaymentRequirement helper to create payment requirements for Base (EVM), Solana (SVM), and Polygon Amoy (testnet). Verify: correct Network, Asset address, atomic amount conversion, EIP-3009 extra field populated for EVM chains only, defaults applied.

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T016 [P] [US2] Add table-driven test in chains_test.go for NewPaymentRequirement with valid inputs across all 8 chains, verify Network, Asset, MaxAmountRequired (atomic units), Scheme="exact", MaxTimeoutSeconds=300, MimeType="application/json"
- [X] T017 [P] [US2] Add test in chains_test.go for EVM chains (Base, Polygon, Avalanche, their testnets) verifying Extra field contains {"name": <chain-specific>, "version": "2"}
- [X] T018 [P] [US2] Add test in chains_test.go for SVM chains (Solana mainnet, devnet) verifying Extra field is empty or nil
- [X] T019 [P] [US2] Add test in chains_test.go for amount conversion: "1.5" → 1500000, "10.50" → 10500000, "0.123456" → 123456
- [X] T020 [P] [US2] Add test in chains_test.go for rounding: verify float64 banker's rounding (round-to-even) behavior with test cases: "1.1234567" → 1123457, "1.1234565" → 1123456 (rounds to even), "1.1234575" → 1123458 (rounds to even), "2.5555555" → 2555556 (rounds to even)
- [X] T021 [P] [US2] Add test in chains_test.go for zero amounts: "0" and "0.0" should both convert to "0" atomic units without error
- [X] T022 [P] [US2] Add error test in chains_test.go for invalid inputs: negative amount "-5", empty recipient address, invalid amount string "abc"
- [X] T023 [P] [US2] Add test in chains_test.go verifying custom config overrides: Scheme="estimate", MaxTimeoutSeconds=600, MimeType="text/plain"

### Implementation for User Story 2

- [X] T024 [US2] Implement amount parsing and validation in chains.go: parse string to float64 using strconv.ParseFloat, validate non-negative, return structured error "amount: <reason>" if invalid
- [X] T025 [US2] Implement atomic conversion in chains.go: multiply float64 by 1e6 and convert to uint64, then to string for MaxAmountRequired field
- [X] T026 [US2] Implement EIP-3009 extra field population in chains.go: for EVM chains (check if EIP3009Name is non-empty), create map[string]interface{}{"name": chain.EIP3009Name, "version": chain.EIP3009Version}
- [X] T027 [US2] Implement NewPaymentRequirement function in chains.go accepting PaymentRequirementConfig, returning PaymentRequirement with validated and converted fields
- [X] T028 [US2] Add validation in NewPaymentRequirement for RecipientAddress non-empty, return structured error "recipientAddress: cannot be empty" if empty
- [X] T029 [US2] Apply defaults in NewPaymentRequirement: Scheme defaults to "exact", MaxTimeoutSeconds defaults to 300, MimeType defaults to "application/json" if not provided
- [X] T030 [US2] Add GoDoc comments to PaymentRequirementConfig struct and NewPaymentRequirement function explaining parameters, defaults, error handling, and rounding behavior
- [X] T031 [US2] Run go test -race -cover ./... and verify all User Story 2 tests pass with no race conditions

**Checkpoint**: User Story 2 complete - developers can create payment requirements with helpers

---

## Phase 5: User Story 3 - Token Configuration Helper (Priority: P2)

**Goal**: Developers can configure client signers with multiple tokens across different chains using NewTokenConfig helper with priority support.

**Independent Test**: Use NewTokenConfig to create token configs for 3+ chains with different priorities, verify each TokenConfig has correct address, decimals, symbol, and priority.

**Note**: User Story 3 uses the same NewTokenConfig function as User Story 1 but emphasizes multi-chain and priority configuration use cases. Additional tests focus on priority handling and multi-chain scenarios.

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T032 [P] [US3] Add test in chains_test.go for multi-chain TokenConfig creation: create configs for BaseMainnet (priority 1), PolygonMainnet (priority 2), SolanaMainnet (priority 3) and verify priorities are correctly set
- [X] T033 [P] [US3] Add test in chains_test.go for testnet TokenConfig creation: BaseSepolia, PolygonAmoy, SolanaDevnet with same priority, verify correct testnet addresses used
- [X] T034 [P] [US3] Add test in chains_test.go verifying TokenConfig Symbol is always "USDC" and Decimals is always 6 for all chains

### Implementation for User Story 3

**Note**: Implementation already completed in User Story 1 (T012). This phase focuses on testing priority and multi-chain scenarios.

- [X] T035 [US3] Verify NewTokenConfig handles all edge cases by running go test -race -cover ./... and confirming all User Story 3 tests pass

**Checkpoint**: User Story 3 complete - developers can configure multi-token clients with priorities

---

## Phase 6: User Story 4 - Network Identifier Lookup (Priority: P3)

**Goal**: Developers can validate network identifiers from payment requirements and determine if they are EVM or SVM networks to route to appropriate signers.

**Independent Test**: Call ValidateNetwork with various network identifiers ("base", "solana", "polygon-amoy", "unknown-chain") and verify correct NetworkType returned or error for unknown networks.

### Tests for User Story 4

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T036 [P] [US4] Add table-driven test in chains_test.go for ValidateNetwork with EVM networks: "base", "base-sepolia", "polygon", "polygon-amoy", "avalanche", "avalanche-fuji" all return NetworkTypeEVM and nil error
- [X] T037 [P] [US4] Add test in chains_test.go for ValidateNetwork with SVM networks: "solana", "solana-devnet" return NetworkTypeSVM and nil error
- [X] T038 [P] [US4] Add test in chains_test.go for ValidateNetwork with unknown network: "ethereum", "arbitrum", "unknown" return NetworkTypeUnknown and structured error "networkID: unsupported network"
- [X] T039 [P] [US4] Add test in chains_test.go for ValidateNetwork with empty string: "" returns error "networkID: cannot be empty"

### Implementation for User Story 4

- [X] T040 [US4] Implement ValidateNetwork function in chains.go that accepts networkID string and returns (NetworkType, error)
- [X] T041 [US4] Create network type lookup map in ValidateNetwork mapping known network IDs to NetworkType (EVM: base, base-sepolia, polygon, polygon-amoy, avalanche, avalanche-fuji; SVM: solana, solana-devnet)
- [X] T042 [US4] Add validation in ValidateNetwork for empty networkID, return structured error "networkID: cannot be empty"
- [X] T043 [US4] Add validation in ValidateNetwork for unknown networkID, return NetworkTypeUnknown and structured error "networkID: unsupported network"
- [X] T044 [US4] Add GoDoc comments to NetworkType constants and ValidateNetwork function explaining purpose and supported networks
- [X] T045 [US4] Run go test -race -cover ./... and verify all User Story 4 tests pass with no race conditions

**Checkpoint**: All user stories complete - full helper function library ready

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Examples, documentation, and final validation

- [X] T046 [P] Create examples/basic/main.go demonstrating client setup with BaseMainnet using NewTokenConfig (User Story 1 example)
- [X] T047 [P] Add example in examples/basic/main.go demonstrating middleware setup with BaseMainnet PaymentRequirement using NewPaymentRequirement (User Story 2 example)
- [X] T048 [P] Add example in examples/basic/main.go demonstrating multi-chain client with Base, Polygon, Solana using different priorities (User Story 3 example)
- [X] T049 [P] Add example in examples/basic/main.go demonstrating network validation using ValidateNetwork (User Story 4 example)
- [X] T050 [P] Add package-level GoDoc comment to chains.go explaining the purpose of the package and linking to quickstart.md
- [X] T051 Run go test -race -cover ./... on full codebase and verify test coverage is maintained or improved (per constitution)
- [X] T052 Run golangci-lint run and fix any linting issues
- [X] T053 Run go build ./... to verify no build errors
- [X] T054 Run go fmt ./... to ensure code is properly formatted
- [X] T055 Verify quickstart.md examples can be executed: test example code matches actual implementation patterns
- [X] T056 [P] Update README.md usage section to reference chain constants and helpers (if README mentions usage)
- [X] T057 Commit all changes with descriptive message following project convention

**Checkpoint**: Feature complete, tested, documented, and ready for code review

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-6)**: All depend on Foundational phase completion
  - User stories can proceed in parallel (if staffed) or sequentially in priority order
  - US1 and US2 are both P1 priority - implement together for MVP
  - US3 (P2) extends US1 functionality
  - US4 (P3) is independent
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational - No dependencies on other stories  
- **User Story 3 (P2)**: Can start after Foundational - Extends US1 but independently testable
- **User Story 4 (P3)**: Can start after Foundational - No dependencies on other stories

### Within Each User Story

- Tests MUST be written and FAIL before implementation (per constitution principle III: Test-First Development)
- Test tasks can run in parallel (all marked [P])
- Implementation tasks follow TDD cycle: write test → verify fail → implement → verify pass
- Each story complete before moving to next priority

### Parallel Opportunities

- **Phase 1**: T002 and T003 can run in parallel
- **Phase 2**: T007 can run in parallel with T004-T006 and T008 (different concerns)
- **After Phase 2 completes**: All user story phases (3-6) can start in parallel if team capacity allows
- **Within each user story**: All test tasks marked [P] can run together
- **Phase 7**: T046-T050 and T056 can all run in parallel (different files)

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together:
# All in chains_test.go but testing different aspects
Task T009: "Table-driven test for ChainConfig constants validation"
Task T010: "Table-driven test for NewTokenConfig helper"  
Task T011: "Test TokenConfig has correct fields"

# After tests fail, implement in chains.go:
Task T012: "Implement NewTokenConfig function"
```

## Parallel Example: User Story 2

```bash
# Launch all tests for User Story 2 together:
Task T016: "Test NewPaymentRequirement valid inputs"
Task T017: "Test EVM chains Extra field"
Task T018: "Test SVM chains Extra field"
Task T019: "Test amount conversion"
Task T020: "Test rounding behavior"
Task T021: "Test zero amounts"
Task T022: "Test error cases"
Task T023: "Test custom config overrides"
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T008) - **CRITICAL**
3. Complete Phase 3: User Story 1 (T009-T015) - Client setup with constants
4. Complete Phase 4: User Story 2 (T016-T031) - Middleware payment requirements
5. **STOP and VALIDATE**: Test US1 and US2 independently
6. Basic examples in Phase 7 (T046-T047)
7. Deploy/demo MVP

**MVP Scope**: Developers can configure clients and middleware with USDC payment support across 8 chains using simple helper functions.

### Incremental Delivery

1. **Foundation**: Setup + Foundational (T001-T008) → Environment ready
2. **MVP Release**: Add US1 + US2 (T009-T031) → Core functionality → Deploy/Demo
3. **Multi-chain Release**: Add US3 (T032-T035) → Priority-based multi-token support → Deploy/Demo
4. **Network Validation Release**: Add US4 (T036-T045) → Network routing helper → Deploy/Demo
5. **Polish Release**: Add Phase 7 (T046-T057) → Examples and documentation → Final Deploy

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T001-T008)
2. Once Foundational is done:
   - **Developer A**: User Story 1 (T009-T015) - Client setup
   - **Developer B**: User Story 2 (T016-T031) - Middleware setup
   - **Developer C**: User Story 3 (T032-T035) - Multi-token priorities
   - **Developer D**: User Story 4 (T036-T045) - Network validation
3. Stories complete independently, integrate via shared chains.go file
4. Team completes Polish together (T046-T057)

---

## Notes

- **[P] tasks**: Different files or independent test cases, can run in parallel
- **[Story] label**: Maps task to user story (US1, US2, US3, US4) for traceability
- **Test-First**: Per constitution, write tests first, verify they fail, then implement
- **Coverage**: Must maintain or improve existing test coverage (constitution principle II)
- **Stdlib-first**: Use strconv, fmt, encoding/json only - no external dependencies (constitution principle IV)
- **Conciseness**: Keep code simple and direct (constitution principle V)
- **Race detection**: Always test with -race flag (constitution testing standards)
- **File paths**: All code in /chains.go and /chains_test.go at repository root
- **Commit strategy**: Commit after each user story checkpoint or logical group
- **Avoid**: Same-file conflicts when parallelizing, unnecessary abstractions, verbose code

---

## Task Count Summary

- **Phase 1 (Setup)**: 3 tasks
- **Phase 2 (Foundational)**: 5 tasks
- **Phase 3 (User Story 1 - P1)**: 7 tasks (3 tests + 4 implementation)
- **Phase 4 (User Story 2 - P1)**: 16 tasks (8 tests + 8 implementation)
- **Phase 5 (User Story 3 - P2)**: 4 tasks (3 tests + 1 validation)
- **Phase 6 (User Story 4 - P3)**: 10 tasks (4 tests + 6 implementation)
- **Phase 7 (Polish)**: 12 tasks
- **Total**: 57 tasks

**Parallel opportunities**: 15+ tasks can run in parallel within phases
**MVP scope**: Phases 1-4 (31 tasks) deliver core value
**Independent stories**: Each user story (Phases 3-6) is independently testable
