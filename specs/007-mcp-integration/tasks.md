# Tasks: MCP Integration

**Input**: Design documents from `/specs/007-mcp-integration/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/mcp-x402-api.yaml

**Tests**: Tests are included as part of the feature implementation per spec requirements.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

**Context**: This implementation leverages existing x402-go components (signers, facilitator, types) and integrates with mark3labs/mcp-go for MCP protocol support. We are NOT building new MCP functionality - only adding x402 payment gating to MCP tools. Reference implementations from mcp-go-x402 will guide the patterns, but adapted to use our existing components.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- Root-level packages: `mcp/client/`, `mcp/server/`
- Examples: `examples/mcp/`
- Tests: `mcp/client/*_test.go`, `mcp/server/*_test.go`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic MCP integration structure

- [X] T001 Create mcp/ package directory structure with client/ and server/ subdirectories
- [X] T002 Add github.com/mark3labs/mcp-go dependency to go.mod (latest stable release)
- [X] T003 [P] Create mcp/errors.go for MCP-specific error types wrapping x402 errors
- [X] T004 [P] Create mcp/types.go for MCP-specific type aliases reusing x402.PaymentRequirement, x402.PaymentPayload, etc.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Define timeout constants in mcp/types.go: PaymentVerifyTimeout (5s per FR-017), PaymentSettleTimeout (60s per FR-018)
- [X] T006 [P] Create mcp/server/requirements.go with helper functions returning x402.PaymentRequirement: RequireUSDCBase(payTo, amount, desc), RequireUSDCBaseSepolia, RequireUSDCPolygon, RequireUSDCSolana with resource field set to "mcp://tools/{toolName}"
- [X] T007 [P] Create mcp/client/config.go defining Config struct with signers []x402.Signer, serverURL string, httpClient *http.Client, OnPaymentAttempt/Success/Failure callbacks, selector x402.PaymentSelector
- [X] T008 [P] Create mcp/server/config.go defining Config struct with FacilitatorURL string, VerifyOnly bool, Verbose bool, PaymentTools map[string][]x402.PaymentRequirement
- [X] T009 [P] Add validation helpers in mcp/server/requirements.go: validateAmount (amount > 0), validateEVMAddress (^0x[a-fA-F0-9]{40}$), validateNetwork (network in supported list), return descriptive errors for invalid requirements

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - MCP Client with x402 Payments (Priority: P1) 🎯 MVP

**Goal**: Enable MCP clients to automatically handle x402 payment flows when calling paid tools on MCP servers

**Independent Test**: Create a client with signers, connect to an x402 MCP server, receive 402 error with payment requirements, automatically sign payment, retry with payment in params._meta, and successfully access paid tool

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T010 [P] [US1] Test Transport initialization with multiple signers in mcp/client/transport_test.go
- [ ] T011 [P] [US1] Test 402 error detection and payment requirement extraction in mcp/client/transport_test.go
- [ ] T012 [P] [US1] Test payment creation using x402.Signer for EVM and Solana in mcp/client/transport_test.go
- [ ] T013 [P] [US1] Test payment injection into params._meta["x402/payment"] per MCP spec in mcp/client/transport_test.go
- [ ] T014 [P] [US1] Test free tool access without payment (no params._meta) in mcp/client/transport_test.go
- [ ] T015 [P] [US1] Test concurrent payment handling where 10 requests each generate unique payments (FR-014) in mcp/client/transport_test.go
- [ ] T016 [P] [US1] Test multi-signer fallback when primary signer fails in mcp/client/transport_test.go

### Implementation for User Story 1

- [X] T017 [US1] Create mcp/client/transport.go implementing Transport struct with fields: baseTransport transport.Interface (wraps transport.StreamableHTTP), config Config, selector x402.PaymentSelector; embed and delegate all transport.Interface methods to baseTransport
- [X] T018 [US1] Implement NewTransport(serverURL string, opts ...Option) in mcp/client/transport.go: create baseTransport using transport.NewStreamableHTTP(serverURL, httpOpts...), wrap with x402 Transport, apply config options for signers/callbacks, return (*Transport, error) implementing transport.Interface
- [X] T019 [US1] Implement Transport.SendRequest(ctx, transport.JSONRPCRequest) in mcp/client/transport.go: override baseTransport method to intercept requests/responses, call baseTransport.SendRequest, check for 402 error (resp.Error != nil && resp.Error.Code == 402), if 402 extract requirements from error.Data, create payment, inject into request.Params._meta["x402/payment"], retry request via baseTransport.SendRequest
- [X] T020 [US1] Add Transport.extractPaymentRequirements method in mcp/client/transport.go: cast error.Data to map[string]any, extract x402Version, error message, accepts []PaymentRequirement fields, unmarshal accepts array to []x402.PaymentRequirement, validate structure
- [X] T021 [US1] Add Transport.selectPaymentSigner method in mcp/client/transport.go: call config.selector.SelectAndSign(requirements, signers) using x402.DefaultPaymentSelector algorithm, return (*PaymentPayload, error)
- [X] T022 [US1] Add Transport.createPayment method in mcp/client/transport.go: call selector.SelectAndSign(requirements, signers) to generate x402.PaymentPayload, handle errors (no valid signer, signing failure), trigger config.OnPaymentAttempt callback with payment details
- [X] T023 [US1] Add Transport.injectPaymentMeta method in mcp/client/transport.go: cast request.Params to map[string]any, create/get "_meta" map, add "x402/payment" key with PaymentPayload value, update request.Params with modified map, preserve all existing params fields
- [X] T024 [US1] Add Transport.retryWithPayment method in mcp/client/transport.go: clone original request, inject payment via injectPaymentMeta, call baseTransport.SendRequest with modified request, check response success, trigger config.OnPaymentSuccess or OnPaymentFailure callbacks based on result
- [X] T025 [US1] Implement Transport.Start, SetNotificationHandler, SendNotification, Close, GetSessionId in mcp/client/transport.go: delegate directly to baseTransport methods (no x402 logic needed for these), preserve transparent wrapper behavior

**Checkpoint**: Client can connect to x402 MCP servers and automatically handle payment flows for paid tools

---

## Phase 4: User Story 2 - MCP Server with x402 Protection (Priority: P1)

**Goal**: Enable MCP servers to protect tools with x402 payment requirements, sending 402 errors to unpaid requests and verifying payments through facilitator before tool execution

**Independent Test**: Create server with free (echo) and paid (search) tools, send request without payment (receives 402 with payment requirements), send request with valid payment in params._meta (tool executes after facilitator verification), verify payment through facilitator

### Tests for User Story 2 ⚠️

- [ ] T026 [P] [US2] Test X402Server initialization and tool registration in mcp/server/server_test.go
- [ ] T027 [P] [US2] Test AddTool for free tools (no payment requirement) in mcp/server/server_test.go
- [ ] T028 [P] [US2] Test AddPayableTool with single payment requirement in mcp/server/server_test.go
- [ ] T029 [P] [US2] Test HTTP handler 402 error generation with payment requirements in error.data per MCP spec in mcp/server/handler_test.go
- [ ] T030 [P] [US2] Test HTTP handler payment extraction from params._meta["x402/payment"] in mcp/server/handler_test.go
- [ ] T031 [P] [US2] Test facilitator payment verification with 5s timeout (FR-015) in mcp/server/handler_test.go
- [ ] T032 [P] [US2] Test facilitator payment settlement with 60s timeout (FR-016) in mcp/server/handler_test.go
- [ ] T033 [P] [US2] Test settlement response injection in result._meta["x402/payment-response"] per MCP spec in mcp/server/handler_test.go
- [ ] T034 [P] [US2] Test verify-only mode (skips settlement) in mcp/server/handler_test.go
- [ ] T035 [P] [US2] Test non-refundable payment when tool execution fails after successful verification (FR-017) in mcp/server/handler_test.go
- [ ] T036 [P] [US2] Test settlement response extraction from error.data["x402/payment-response"] when payment settlement fails after verification succeeds per x402 MCP spec in mcp/server/handler_test.go

### Implementation for User Story 2

- [X] T037 [US2] Create mcp/server/server.go implementing X402Server struct with fields: mcpServer *server.MCPServer, config *Config (includes PaymentTools map)
- [X] T038 [US2] Implement NewX402Server(name, version string, config *Config) in mcp/server/server.go: call server.NewMCPServer(name, version) to create base MCP server, initialize config.PaymentTools map if nil, return *X402Server wrapping the MCPServer
- [X] T039 [US2] Add X402Server.AddTool(tool mcp.Tool, handler server.ToolHandlerFunc) in mcp/server/server.go: call s.mcpServer.AddTool(tool, handler) directly without payment requirements, do not add to PaymentTools map (free tool)
- [X] T040 [US2] Add X402Server.AddPayableTool(tool mcp.Tool, handler server.ToolHandlerFunc, requirements ...x402.PaymentRequirement) in mcp/server/server.go: validate len(requirements) > 0, add to config.PaymentTools[tool.Name], call s.mcpServer.AddTool(tool, handler) to register tool
- [X] T041 [US2] Add X402Server.Handler() http.Handler method in mcp/server/server.go: create httpServer := server.NewStreamableHTTPServer(s.mcpServer), wrap with NewX402Handler(httpServer, s.config), return wrapped handler
- [X] T042 [US2] Add X402Server.Start(addr string) in mcp/server/server.go: call http.ListenAndServe(addr, s.Handler()) to start HTTP server with x402-wrapped handler
- [X] T043 [US2] Create mcp/server/handler.go implementing X402Handler struct with fields: mcpHandler http.Handler (wraps MCPServer's HTTP handler), config *Config, facilitator Facilitator
- [X] T044 [US2] Implement NewX402Handler(mcpHandler http.Handler, config *Config) in mcp/server/handler.go: create facilitator wrapper, return *X402Handler with mcpHandler (from server.NewStreamableHTTPServer), config, facilitator
- [X] T045 [US2] Implement X402Handler.ServeHTTP(w http.ResponseWriter, r *http.Request) in mcp/server/handler.go: intercept POST requests, read body with io.ReadAll, parse as transport.JSONRPCRequest, check method == "tools/call", extract tool name from params, check if tool needs payment, handle payment flow or pass through to mcpHandler
- [X] T046 [US2] Add handler.checkPaymentRequired(toolName string) in mcp/server/handler.go: lookup config.PaymentTools[toolName], return (requirements, needsPayment bool), set resource field to "mcp://tools/{toolName}" on requirements
- [X] T047 [US2] Add handler.sendPaymentRequiredError(w, id, requirements) in mcp/server/handler.go: construct transport.JSONRPCResponse with Error.Code=402, Error.Data={x402Version:1, error:"Payment required", accepts:requirements}, write JSON response with HTTP 200 (JSON-RPC error, not HTTP error)
- [X] T048 [US2] Add handler.extractPayment(params mcp.CallToolParams) in mcp/server/handler.go: check params.Meta != nil && params.Meta.AdditionalFields != nil, extract params.Meta.AdditionalFields["x402/payment"], marshal to PaymentPayload struct, return payment or nil
- [X] T049 [US2] Add handler.findMatchingRequirement(payment, requirements) in mcp/server/handler.go: iterate requirements, match on network and scheme, return matched requirement or error
- [X] T050 [US2] Create mcp/server/facilitator.go with Facilitator interface (Verify, Settle methods) and HTTPFacilitator struct wrapping http.FacilitatorClient
- [X] T051 [US2] Add HTTPFacilitator.Verify(ctx, payment, requirement) in mcp/server/facilitator.go: create context.WithTimeout(ctx, 5*time.Second), call facilitatorClient.VerifyPayment(ctx, payment), return VerifyResponse{IsValid, InvalidReason, Payer} or error
- [X] T052 [US2] Add HTTPFacilitator.Settle(ctx, payment, requirement) in mcp/server/facilitator.go: create context.WithTimeout(ctx, 60*time.Second), call facilitatorClient.SettlePayment(ctx, payment), return SettleResponse{Success, Transaction, Network, Payer, ErrorReason} or error
- [X] T053 [US2] Add handler.forwardWithSettlementResponse(w, r, reqID, settleResp) in mcp/server/handler.go: create responseRecorder to capture MCP handler response, forward request to mcpHandler.ServeHTTP, parse JSON-RPC response, inject settleResp into result._meta["x402/payment-response"], write modified response

**Checkpoint**: Server can protect tools with x402 requirements, send 402 errors, verify payments via facilitator, and execute paid tools

---

## Phase 5: User Story 3 - Multi-Chain Payment Support (Priority: P2)

**Goal**: Support payments using different blockchain networks (EVM chains, Solana) with automatic selection based on DefaultPaymentSelector priority algorithm

**Independent Test**: Configure client with EVM (Base) and Solana signers, configure server to accept both networks, verify client selects based on signer priority → token priority → configuration order per DefaultPaymentSelector

### Tests for User Story 3 ⚠️

- [ ] T054 [P] [US3] Test DefaultPaymentSelector priority algorithm (signer priority → token priority → config order) in mcp/client/transport_test.go
- [ ] T055 [P] [US3] Test EVM payment creation for Base network in mcp/client/transport_test.go
- [ ] T056 [P] [US3] Test Solana payment creation in mcp/client/transport_test.go
- [ ] T057 [P] [US3] Test fallback from Base to Solana when primary fails in mcp/client/transport_test.go
- [ ] T058 [P] [US3] Test payment fallback completes within 5 seconds (SC-003) in mcp/client/transport_test.go
- [ ] T059 [P] [US3] Test server accepting multiple network payment requirements in mcp/server/server_test.go

### Implementation for User Story 3

- [ ] T060 [US3] Integrate x402.DefaultPaymentSelector in mcp/client/transport.go for signer selection logic
- [ ] T061 [US3] Add support for multiple payment signers with priority ordering in mcp/client/transport.go
- [ ] T062 [US3] Add EVM-specific payment handling in mcp/client/transport.go using existing evm.Signer
- [ ] T063 [US3] Add Solana-specific payment handling in mcp/client/transport.go using existing svm.Signer
- [ ] T064 [US3] Add multi-network requirement support in mcp/server/requirements.go (RequireUSDCBase, RequireUSDCPolygon, RequireUSDCSolana, RequireUSDCBaseSepolia)
- [ ] T065 [US3] Update handler.findMatchingRequirement in mcp/server/handler.go to handle multiple payment options per tool

**Checkpoint**: Multi-chain payment support working with automatic selection and fallback across EVM and Solana

---

## Phase 6: User Story 4 - Example Implementation (Priority: P3)

**Goal**: Provide working examples demonstrating both client and server implementations with x402 payment flows, similar to examples/x402demo

**Independent Test**: Run example in server mode, verify echo (free) and search (paid) tools are served; run example in client mode, verify successful connection and access to both tool types with automatic payment handling

### Implementation for User Story 4

- [X] T066 [P] [US4] Create examples/mcp/ directory with go.mod requiring github.com/mark3labs/x402-go (use replace directive pointing to ../.. for development)
- [X] T067 [P] [US4] Create examples/mcp/README.md with usage instructions: server mode (./mcp -mode server -pay-to ADDR), client mode (./mcp -mode client -key PRIVATE_KEY -server http://localhost:8080), testnet flags
- [X] T068 [US4] Implement examples/mcp/main.go with flag.String for -mode, switch on mode value to call runServer() or runClient()
- [X] T069 [US4] Add runServer() in examples/mcp/main.go: call server.NewX402Server(name, version, config), create echo tool with mcp.NewTool("echo", mcp.WithString("message", mcp.Required())), create search tool with mcp.NewTool("search", mcp.WithString("query", mcp.Required()), mcp.WithNumber("max_results")), call srv.AddTool(echoTool, echoHandler) and srv.AddPayableTool(searchTool, searchHandler, server.RequireUSDCBase(payTo, "10000", "0.01 USDC"))
- [X] T070 [US4] Add echoHandler(ctx context.Context, req mcp.CallToolRequest) in examples/mcp/main.go: extract message := req.GetString("message", ""), return mcp.NewToolResultText(fmt.Sprintf("Echo: %s", message))
- [X] T071 [US4] Add searchHandler(ctx context.Context, req mcp.CallToolRequest) in examples/mcp/main.go: extract query := req.GetString("query", ""), maxResults := req.GetFloat("max_results", 5), generate mock search results, return mcp.NewToolResultText(results)
- [X] T072 [US4] Add runClient() in examples/mcp/main.go: create evm.NewPrivateKeySigner(key, evm.WithChain(chain), evm.WithToken(token)), create transport := mcpclient.NewTransport(serverURL, client.WithSigner(signer), client.WithPaymentCallback(paymentLogger)), create mcpClient := client.NewClient(transport)
- [X] T073 [US4] Add MCP initialization in runClient() in examples/mcp/main.go: call mcpClient.Start(ctx), call mcpClient.Initialize(ctx, mcp.InitializeRequest{Params: {ProtocolVersion: "2025-06-18", ClientInfo: {Name: "x402-example", Version: "1.0.0"}}}), handle init response
- [X] T074 [US4] Add tool operations in runClient() in examples/mcp/main.go: call mcpClient.ListTools(ctx, mcp.ListToolsRequest{}), iterate tools, call mcpClient.CallTool(ctx, mcp.CallToolRequest{Params: {Name: "echo", Arguments: {"message": "test"}}}), call mcpClient.CallTool for search tool with payment
- [X] T075 [US4] Add paymentLogger callback in examples/mcp/main.go: implement function logging payment events with log.Printf("Attempting payment: %s %s to %s", event.Amount, event.Asset, event.Recipient) for attempt, "Payment successful: tx=%s" for success, "Payment failed: %v" for failure
- [X] T076 [US4] Add command-line flags in examples/mcp/main.go: flag.String("mode", "", "client or server"), flag.String("port", "8080", "server port"), flag.String("server", "http://localhost:8080", "server URL"), flag.String("key", "", "private key"), flag.String("pay-to", "", "payment address"), flag.String("facilitator", "https://facilitator.x402.rs", "facilitator URL"), flag.Bool("verify-only", false, "verify only"), flag.Bool("testnet", false, "use testnet"), flag.String("network", "base", "network name"), flag.Bool("v", false, "verbose")
- [X] T077 [US4] Add testnet support in examples/mcp/main.go: if testnet flag, use evm.ChainBaseSepolia and server.RequireUSDCBaseSepolia for payment requirements

**Checkpoint**: Working example demonstrates full x402 MCP integration for both client and server modes

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories, edge cases, and finalization

### Edge Case Tests

- [ ] T078 [P] Test behavior when all configured signers fail to create valid payment in mcp/client/transport_test.go
- [ ] T079 [P] Test network timeout handling during payment verification in mcp/server/handler_test.go
- [ ] T080 [P] Test behavior when facilitator is unavailable after valid payment in mcp/server/handler_test.go
- [ ] T081 [P] Test client handling of malformed payment requirements from server in mcp/client/transport_test.go
- [ ] T082 [P] Test behavior when server requirements exceed all client-configured limits in mcp/client/transport_test.go

### Protocol Error Tests (JSON-RPC Standard Errors)

- [ ] T083 [P] Test JSON-RPC parse error (-32700) when _meta["x402/payment"] contains invalid JSON in mcp/client/transport_test.go
- [ ] T084 [P] Test JSON-RPC invalid params error (-32602) when payment payload is malformed in mcp/server/handler_test.go
- [ ] T085 [P] Test JSON-RPC internal error (-32603) when facilitator is unreachable during verification in mcp/server/handler_test.go
- [ ] T086 [P] Test JSON-RPC method not found error (-32601) when server doesn't support x402 payments in mcp/client/transport_test.go

### Documentation and Polish

- [ ] T087 [P] Add comprehensive godoc comments to all exported types in mcp/client/ package
- [ ] T088 [P] Add comprehensive godoc comments to all exported types in mcp/server/ package
- [ ] T089 [P] Verify error messages are clear and actionable across mcp/client/ and mcp/server/
- [ ] T090 [P] Add verbose logging for payment lifecycle in mcp/client/transport.go
- [ ] T091 [P] Add verbose logging for payment verification in mcp/server/handler.go
- [X] T092 [P] Run go fmt ./mcp/... on all MCP package code
- [X] T093 [P] Run go vet ./mcp/... on all MCP package code
- [X] T094 [P] Run golangci-lint run ./mcp/... on all MCP package code
- [X] T095 [P] Validate examples/mcp/main.go builds without errors: go build ./examples/mcp
- [ ] T096 Verify quickstart.md code examples match actual implementation and update if needed
- [X] T097 Run full test suite with race detection: go test -race -cover ./mcp/...
- [ ] T098 Validate all success criteria from spec.md are met (SC-001 through SC-008)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User Story 1 (Client) and User Story 2 (Server) can proceed in parallel after Phase 2
  - User Story 3 (Multi-chain) depends on User Story 1 completion (extends client)
  - User Story 4 (Examples) depends on User Story 1 and 2 completion (demonstrates both)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1 - Client)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1 - Server)**: Can start after Foundational (Phase 2) - No dependencies on other stories (can develop in parallel with US1)
- **User Story 3 (P2 - Multi-chain)**: Depends on User Story 1 completion - Extends client signer selection
- **User Story 4 (P3 - Examples)**: Depends on User Story 1 and 2 completion - Demonstrates client + server integration

### Within Each User Story

- **User Story 1**: Tests first (T010-T016) → Config (T017-T018) → Transport implementation (T019-T024) → Delegation (T025)
- **User Story 2**: Tests first (T026-T036) → Server struct (T037-T042) → Handler (T043-T049) → Facilitator (T050-T052) → Integration (T053)
- **User Story 3**: Tests first (T054-T059) → Selector integration (T060-T061) → Network-specific handling (T062-T065)
- **User Story 4**: Structure (T066-T067) → Implementation (T068-T077) sequentially

### Parallel Opportunities

- **Phase 1**: T003 and T004 can run in parallel (different files)
- **Phase 2**: T006, T007, T008, T009 can run in parallel after T005 (different files)
- **User Story 1**: Tests T010-T016 can all run in parallel (independent test cases)
- **User Story 2**: Tests T026-T036 can all run in parallel (independent test cases)
- **User Story 3**: Tests T054-T059 can all run in parallel (independent test cases)
- **Phase 7**: All edge case tests T078-T082 can run in parallel; protocol error tests T083-T086 can run in parallel; all polish tasks T087-T095 can run in parallel
- **MAJOR PARALLELISM**: After Phase 2, User Story 1 and User Story 2 can be developed completely in parallel by different developers

---

## Parallel Example: User Story 1 (Client Implementation)

```bash
# After Phase 2 completes, launch tests in parallel:
Task: T010 "Test Transport initialization" 
Task: T011 "Test 402 error detection"
Task: T012 "Test payment creation"
Task: T013 "Test payment injection"
Task: T014 "Test free tool access"
Task: T015 "Test concurrent payments"
Task: T016 "Test multi-signer fallback"

# All 7 tests run simultaneously (different test cases)
```

## Parallel Example: User Story 1 + User Story 2

```bash
# After Phase 2 (T009) completes, launch both stories in parallel:

# Team A: User Story 1 (Client) - mcp/client/*
Task: T010-T016 "Client tests"
Task: T017-T025 "Client implementation"

# Team B: User Story 2 (Server) - mcp/server/* - runs CONCURRENTLY
Task: T026-T036 "Server tests"
Task: T037-T053 "Server implementation"

# These can proceed completely independently!
```

---

## Implementation Strategy

### MVP First (User Story 1 + User Story 2)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Client) - Can run parallel with US2
4. Complete Phase 4: User Story 2 (Server) - Can run parallel with US1
5. **STOP and VALIDATE**: Test client against server independently
6. Integration test: Client calling paid (search) and free (echo) tools on Server
7. Deploy/demo if ready (this is the MVP!)

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 + 2 (Client + Server) → Test together → Deploy/Demo (MVP!)
3. Add User Story 3 (Multi-chain) → Test independently → Deploy/Demo
4. Add User Story 4 (Examples) → Validate with examples → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (Phase 1-2)
2. Once Foundational is done:
   - Developer A: User Story 1 (Client) - mcp/client/
   - Developer B: User Story 2 (Server) - mcp/server/
   - Both work independently on separate packages
3. After US1 + US2 complete:
   - Developer A: User Story 3 (Multi-chain) - extends client
   - Developer B: User Story 4 (Examples)
4. Team does Phase 7 (Polish) together

---

## Notes

- [P] tasks = different files, no dependencies, can run concurrently
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- US1 and US2 are both P1 priority and can be developed in parallel (different packages: mcp/client/ vs mcp/server/)
- Commit after each task or logical group
- Stop at checkpoints to validate story independently
- **Reference implementations**: Use mcp-go-x402 examples as patterns but adapt to use existing x402-go components
- **Reuse existing components**: All signers (evm, svm, coinbase) work without modification; use http.FacilitatorClient for all payment operations
- **MCP integration**: Use mark3labs/mcp-go for all MCP protocol types and functionality (transport.Interface, server.MCPServer, mcp.Tool, etc.)
- **Testing**: Follow go test -race ./... throughout development
- **Timeouts**: Ensure all payment timeouts match spec: 5s verify (FR-015), 60s settle (FR-016)
- **Total tasks**: 98 (T001-T098) organized by user story for independent development

---

## Success Criteria Validation

After completing all tasks, verify these measurable outcomes from spec.md:

- **SC-001**: Developers can protect MCP tools with x402 in under 10 lines per tool (validate in examples/mcp/main.go)
- **SC-002**: Client handles payment flow with zero additional code beyond signer config (validate in examples/mcp/main.go client mode)
- **SC-003**: Payment fallback completes within 5 seconds when primary signer fails (T058 test)
- **SC-004**: Example runs successfully with both EVM and Solana payment options (examples/mcp/main.go)
- **SC-005**: 100% of existing signer types work with MCP integration (evm, svm, coinbase)
- **SC-006**: Server processes mixed free/paid tool requests without payment overhead on free tools (T027, T039 tests)
- **SC-007**: All payment verification flows complete within 5 seconds (T031 test)
- **SC-008**: Integration adds fewer than 1000 lines of new code by reusing existing components (measure after completion)
