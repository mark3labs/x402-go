# Implementation Tasks: PocketBase Middleware for x402 Payment Protocol

**Feature**: 005-pocketbase-middleware  
**Branch**: `005-pocketbase-middleware`  
**Date**: 2025-10-29  
**Generated**: Via `/speckit.tasks` command

## Overview

This document contains the implementation tasks for the PocketBase middleware feature, organized by user story to enable independent implementation and testing. Each user story represents a complete, deliverable increment.

**Total Estimated Tasks**: 55  
**User Stories**: 3 (P1, P2, P2)  
**Suggested MVP**: User Story 1 only (core payment gating)

---

## Task Summary by Phase

| Phase | Description | Task Count | Parallelizable |
|-------|-------------|------------|----------------|
| Phase 1 | Setup & Project Structure | 5 | 3 |
| Phase 2 | Foundational (Blocking Prerequisites) | 8 | 4 |
| Phase 3 | User Story 1 - Basic Payment Gating (P1) | 23 | 15 |
| Phase 4 | User Story 2 - Context Integration (P2) | 8 | 4 |
| Phase 5 | User Story 3 - Verify-Only Mode (P2) | 5 | 2 |
| Phase 6 | Polish & Cross-Cutting Concerns | 6 | 2 |

---

## Dependencies & Execution Order

### User Story Dependencies

```
Setup (Phase 1)
    ↓
Foundational (Phase 2) 
    ↓
User Story 1 (P1) - Basic Payment Gating ← MVP COMPLETE
    ↓
User Story 2 (P2) - Context Integration ← Independent (can run parallel with US3)
    ↓
User Story 3 (P2) - Verify-Only Mode ← Independent (can run parallel with US2)
    ↓
Polish (Phase 6)
```

**Independent Stories**: US2 and US3 have NO dependencies on each other after US1 completes. They can be implemented in parallel by different developers.

### Parallel Execution Opportunities

**Phase 1 (Setup)**: Tasks T002, T003, T004 can run in parallel after T001  
**Phase 3 (US1)**: Tasks T009-T012 (helpers) can run in parallel; Tests T013-T016 can run in parallel  
**Phase 4 (US2)**: Tasks T027-T029 (tests) can run in parallel  
**Phase 5 (US3)**: Tasks T033-T034 (tests) can run in parallel  
**Phase 6 (Polish)**: Tasks T038, T039 can run in parallel

---

## Phase 1: Setup & Project Structure

**Goal**: Initialize the project structure for PocketBase middleware implementation

**Prerequisites**: None (starting from existing x402-go repository)

**Completion Criteria**: 
- `http/pocketbase/` directory exists with package structure
- Package documentation in place
- Dependencies verified in go.mod

### Tasks

- [X] T001 Create http/pocketbase/ package directory structure
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/`
  - **Action**: Create directory for PocketBase middleware package
  - **Verification**: Directory exists and is empty

- [X] T002 [P] Add PocketBase dependency to go.mod if not present
  - **Path**: `/home/space_cowboy/Workspace/x402-go/go.mod`
  - **Action**: Ensure `github.com/pocketbase/pocketbase` is in dependencies
  - **Command**: `go get github.com/pocketbase/pocketbase@latest && go mod tidy`
  - **Verification**: Dependency appears in go.mod

- [X] T003 [P] Create middleware.go file with package documentation
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Create file with package comment describing PocketBase adapter
  - **Content**: 
    ```go
    // Package pocketbase provides PocketBase-compatible middleware for x402 payment gating.
    // This package is a thin adapter that translates core.RequestEvent to stdlib http patterns
    // and delegates all payment verification and settlement logic to the http package.
    package pocketbase
    ```
  - **Verification**: File exists with package declaration

- [X] T004 [P] Create middleware_test.go file with package declaration
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Create test file with package declaration
  - **Content**: `package pocketbase`
  - **Verification**: File exists

- [X] T005 Create examples/pocketbase/ directory structure
  - **Path**: `/home/space_cowboy/Workspace/x402-go/examples/pocketbase/`
  - **Action**: Create directory for example PocketBase application
  - **Verification**: Directory exists

---

## Phase 2: Foundational (Blocking Prerequisites)

**Goal**: Implement four duplicated framework-specific helper functions that all user stories depend on

**Prerequisites**: Phase 1 complete

**Completion Criteria**: 
- All 4 duplicated framework-specific helpers implemented and compiling (parsePaymentHeaderFromRequest, sendPaymentRequiredPocketBase, findMatchingRequirementPocketBase, addPaymentResponseHeaderPocketBase)
- All 4 helper functions have passing unit tests
- Helpers follow Gin middleware duplication pattern (self-contained, no stdlib imports)

### Tasks

- [X] T006 Implement parsePaymentHeaderFromRequest helper function
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Duplicate from Gin middleware pattern - parse X-PAYMENT header (base64 decode → JSON unmarshal → version check)
  - **Signature**: `func parsePaymentHeaderFromRequest(r *http.Request) (x402.PaymentPayload, error)`
  - **Logic**: 
    - Get X-PAYMENT header
    - Base64 decode
    - JSON unmarshal into PaymentPayload
    - Validate X402Version == 1
    - Return payment or error
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/http/gin/middleware.go` lines 213-240
  - **Verification**: Function compiles, no tests yet

- [X] T007 Implement sendPaymentRequiredPocketBase helper function
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Send 402 response with PaymentRequirementsResponse using e.JSON()
  - **Signature**: `func sendPaymentRequiredPocketBase(e *core.RequestEvent, requirements []x402.PaymentRequirement) error`
  - **Logic**: 
    - Create PaymentRequirementsResponse{x402Version: 1, error: "...", accepts: requirements}
    - Call e.JSON(http.StatusPaymentRequired, response) and return the error
    - This stops handler chain (equivalent to Gin's c.AbortWithStatusJSON)
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/http/gin/middleware.go` lines 242-250
  - **Pattern**: Return e.JSON() result - don't call e.Next() on errors
  - **Verification**: Function compiles with error return type

- [X] T008 Implement findMatchingRequirementPocketBase helper function
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Find requirement matching payment's scheme and network
  - **Signature**: `func findMatchingRequirementPocketBase(payment x402.PaymentPayload, requirements []x402.PaymentRequirement) (x402.PaymentRequirement, error)`
  - **Logic**: 
    - Iterate requirements
    - Match req.Scheme == payment.Scheme && req.Network == payment.Network
    - Return requirement or x402.ErrUnsupportedScheme
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/http/gin/middleware.go` lines 253-261
  - **Verification**: Function compiles

- [X] T009 Implement addPaymentResponseHeaderPocketBase helper function
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Add X-PAYMENT-RESPONSE header with base64-encoded settlement
  - **Signature**: `func addPaymentResponseHeaderPocketBase(e *core.RequestEvent, settlement *x402.SettlementResponse) error`
  - **Logic**: 
    - JSON marshal settlement
    - Base64 encode
    - Call e.Response.Header().Set("X-PAYMENT-RESPONSE", encoded)
    - Return error if marshal fails
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/http/gin/middleware.go` lines 264-277
  - **Verification**: Function compiles

### Foundational Tests

- [X] T006a [P] Write unit test for parsePaymentHeaderFromRequest helper
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test helper parses valid/invalid payment headers
  - **Test Logic**: 
    - Test valid base64 + JSON → returns PaymentPayload
    - Test invalid base64 → returns error
    - Test invalid JSON → returns error
    - Test wrong X402Version → returns error
  - **Verification**: Test compiles and passes

- [X] T007a [P] Write unit test for sendPaymentRequiredPocketBase helper
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test helper sends correct 402 response
  - **Test Logic**: 
    - Create mock RequestEvent
    - Call helper with requirements
    - Assert status = 402
    - Assert response contains x402Version, error, accepts fields
  - **Verification**: Test compiles and passes

- [X] T008a [P] Write unit test for findMatchingRequirementPocketBase helper
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test helper finds matching requirements
  - **Test Logic**: 
    - Test matching scheme + network → returns requirement
    - Test non-matching scheme → returns ErrUnsupportedScheme
    - Test non-matching network → returns error
  - **Verification**: Test compiles and passes

- [X] T009a [P] Write unit test for addPaymentResponseHeaderPocketBase helper
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test helper adds X-PAYMENT-RESPONSE header
  - **Test Logic**: 
    - Create mock RequestEvent and settlement response
    - Call helper
    - Assert X-PAYMENT-RESPONSE header present
    - Assert header value is valid base64
    - Decode and verify JSON structure
  - **Verification**: Test compiles and passes

---

## Phase 3: User Story 1 - Basic Payment Gating with PocketBase (P1)

**User Story**: A developer building APIs with the PocketBase framework needs to protect their custom endpoints with x402 payment gating. They should be able to apply the middleware to specific routes or route groups without changing their existing PocketBase application structure.

**Why P1**: This is the core functionality - without it, PocketBase users cannot use x402 payment gating at all. This represents the minimum viable product (MVP).

**Independent Test Criteria**: 
- Can create a PocketBase application with protected endpoint
- Can send request without X-PAYMENT → receives 402 with payment requirements
- Can send request with valid X-PAYMENT → payment verified, settled, handler executes
- Can send request with invalid X-PAYMENT → receives 402 with payment requirements
- Can send request when facilitator fails → receives 503 error

**Acceptance Scenarios** (from spec.md):
1. Request without X-PAYMENT header → HTTP 402 with PaymentRequirementsResponse
2. Request with valid X-PAYMENT header → Payment verified, settled, handler executes via e.Next()
3. Request with invalid X-PAYMENT header → HTTP 402 with payment requirements
4. Payment verification fails at facilitator → HTTP 402 with error details

### Tasks

- [X] T010 [P] [US1] Implement NewPocketBaseX402Middleware factory function skeleton
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Create factory function that accepts http.Config and returns middleware handler function
  - **Signature**: `func NewPocketBaseX402Middleware(config *http.Config) func(*core.RequestEvent) error`
  - **Logic**: 
    - Accept config parameter
    - Create FacilitatorClient (primary)
    - Create fallback FacilitatorClient if configured
    - Return anonymous middleware function with signature `func(e *core.RequestEvent) error`
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/http/gin/middleware.go` lines 53-83
  - **Verification**: Function signature compiles

- [X] T011 [P] [US1] Implement facilitator enrichment logic in factory
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Call facilitator.EnrichRequirements() at middleware initialization
  - **Logic**: 
    - Call `facilitator.EnrichRequirements(config.PaymentRequirements)`
    - If error: log warning with slog.Default().Warn("enrichment failed", "facilitator_url", facilitator.URL, "error", err, "requirement_count", len(originalRequirements)), use original requirements
    - If success: log info with slog.Default().Info("requirements enriched", "facilitator_url", facilitator.URL, "requirement_count", len(enrichedRequirements)), use enriched requirements
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/http/gin/middleware.go` lines 73-81
  - **Verification**: Enrichment logic compiles

- [X] T012 [P] [US1] Implement request URL construction and requirement population
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Inside middleware handler, construct resource URL and populate requirements
  - **Logic**: 
    - Build scheme: check e.Request.TLS (https if present, http otherwise)
    - Build resourceURL: `scheme + "://" + e.Request.Host + e.Request.RequestURI`
    - Copy enrichedRequirements to requirementsWithResource
    - For each requirement: set Resource field to resourceURL, default Description if empty
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/http/gin/middleware.go` lines 87-102
  - **Verification**: URL construction logic compiles

- [X] T012a [US1] Implement CORS OPTIONS request bypass
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Skip payment verification for OPTIONS requests
  - **Logic**: 
    - Check if e.Request.Method == "OPTIONS"
    - If true: log debug "bypassing OPTIONS request", call return e.Next()
  - **Reference**: Standard CORS preflight behavior (RFC 7231)
  - **Verification**: T025c test passes

- [X] T013 [P] [US1] Implement X-PAYMENT header check and missing payment response
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Check for X-PAYMENT header, return 402 if missing
  - **Logic**: 
    - Get header: `paymentHeader := e.Request.Header.Get("X-PAYMENT")`
    - If empty: log info, `return sendPaymentRequiredPocketBase(e, requirementsWithResource)`
    - **IMPORTANT**: Return the error from sendPaymentRequired - this stops handler chain (don't call e.Next())
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/http/gin/middleware.go` lines 105-111
  - **Pattern**: All error paths return without calling e.Next()
  - **Verification**: Missing payment path compiles

- [X] T014 [US1] Implement payment header parsing and error handling
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Parse payment header, handle malformed headers with 400 response
  - **Logic**: 
    - Call parsePaymentHeaderFromRequest(e.Request)
    - If error: log warning, return e.JSON(http.StatusBadRequest, map with x402Version and error)
    - Log debug info about parsed payment
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/http/gin/middleware.go` lines 114-133
  - **Verification**: Parsing logic compiles

- [X] T015 [US1] Implement requirement matching and mismatch response
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Find matching requirement, return 402 if no match
  - **Logic**: 
    - Call findMatchingRequirementPocketBase(payment, requirementsWithResource)
    - If error: log warning, call sendPaymentRequiredPocketBase(), return nil
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/http/gin/middleware.go` lines 136-141
  - **Verification**: Matching logic compiles

- [X] T016 [US1] Implement payment verification with facilitator
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Call facilitator.Verify() with fallback support
  - **Logic**: 
    - Log info "verifying payment"
    - Call facilitator.Verify(payment, requirement)
    - If error and fallbackFacilitator != nil: try fallback
    - If still error: log error, return e.JSON(503, error response)
    - Check verifyResp.IsValid: if false, log warning, send 402, return nil
    - If valid: log info "payment verified" with payer
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/http/gin/middleware.go` lines 144-163
  - **Verification**: Verification logic compiles

- [X] T017 [US1] Implement payment settlement with facilitator
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Call facilitator.Settle() if not verify-only mode
  - **Logic**: 
    - Check if config.VerifyOnly is false
    - If settling: log info, call facilitator.Settle(payment, requirement)
    - If error and fallbackFacilitator != nil: try fallback
    - If still error: log error, return e.JSON(503, error response)
    - Check settlementResp.Success: if false, log warning, send 402, return nil
    - If success: log info with transaction hash
    - Call addPaymentResponseHeaderPocketBase(e, settlementResp)
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/http/gin/middleware.go` lines 169-199
  - **Verification**: Settlement logic compiles

- [X] T018 [US1] Implement successful payment flow (store in context and call e.Next)
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Store payment info in request store and call next handler
  - **Logic**: 
    - Call e.Set("x402_payment", verifyResp)
    - Call return e.Next()
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/http/gin/middleware.go` lines 201-209
  - **Verification**: Success path compiles

- [X] T019 [US1] Write test for missing X-PAYMENT header scenario
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test request without X-PAYMENT returns 402 with PaymentRequirementsResponse
  - **Test Logic**: 
    - Create test config with payment requirements
    - Create test RequestEvent without X-PAYMENT header
    - Call middleware
    - Assert HTTP 402 status
    - Assert response body contains x402Version, error, accepts fields
  - **Verification**: Test compiles and runs

- [X] T020 [P] [US1] Write test for invalid base64 in X-PAYMENT header
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test malformed base64 returns 400 Bad Request with specific error
  - **Test Logic**: 
    - Create test RequestEvent with invalid base64 in X-PAYMENT header
    - Call middleware
    - Assert HTTP 400 status
    - Assert response contains x402Version and error message "invalid base64 encoding in X-PAYMENT header"
  - **Verification**: Test compiles and runs

- [X] T021 [P] [US1] Write test for invalid JSON in X-PAYMENT header
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test malformed JSON returns 400 Bad Request with specific error
  - **Test Logic**: 
    - Create test RequestEvent with valid base64 but invalid JSON
    - Call middleware
    - Assert HTTP 400 status
    - Assert response contains x402Version and error message "invalid JSON structure in payment payload"
  - **Verification**: Test compiles and runs

- [X] T022 [P] [US1] Write test for valid payment with successful verification and settlement
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test valid payment flow calls e.Next() and stores payment data
  - **Test Logic**: 
    - Mock facilitator with successful verify and settle responses
    - Create test RequestEvent with valid X-PAYMENT header
    - Call middleware
    - Assert e.Next() was called (handler executed)
    - Assert payment data stored in request store
    - Assert X-PAYMENT-RESPONSE header present
  - **Verification**: Test compiles and runs

- [X] T023 [P] [US1] Write test for facilitator verification failure
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test facilitator unreachable returns 503
  - **Test Logic**: 
    - Mock facilitator to return error on Verify()
    - Create test RequestEvent with valid payment
    - Call middleware
    - Assert HTTP 503 status
    - Assert e.Next() was NOT called
  - **Verification**: Test compiles and runs

- [X] T024 [P] [US1] Write test for route-level binding
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test middleware works with Route.BindFunc()
  - **Test Logic**: 
    - Create mock PocketBase router
    - Register route with middleware: `router.GET("/test", handler).BindFunc(middleware)`
    - Send request
    - Verify middleware executes
  - **Verification**: Test compiles and runs

- [X] T025 [P] [US1] Write test for group-level binding
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test middleware works with RouterGroup.BindFunc()
  - **Test Logic**: 
    - Create mock PocketBase router with group
    - Register middleware on group: `group.BindFunc(middleware)`
    - Register routes in group
    - Verify middleware applies to all group routes
  - **Verification**: Test compiles and runs

- [X] T025a [P] [US1] Write test for settlement failure after successful verification
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test settlement failure returns HTTP 503 after successful verification
  - **Test Logic**: 
    - Mock facilitator Verify() to succeed
    - Mock facilitator Settle() to return error
    - Create test RequestEvent with valid payment
    - Call middleware
    - Assert HTTP 503 status
    - Assert error response contains settlement failure message
    - Assert e.Next() was NOT called
  - **Verification**: Test compiles and runs

- [X] T025b [P] [US1] Write test for insufficient payment amount
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test payment with insufficient amount returns HTTP 402
  - **Test Logic**: 
    - Create payment with amount less than requirement
    - Mock facilitator Verify() to return IsValid=false with InvalidReason="insufficient amount"
    - Call middleware
    - Assert HTTP 402 status
    - Assert payment requirements returned
    - Assert e.Next() was NOT called
  - **Verification**: Test compiles and runs

- [X] T025c [P] [US1] Write test for CORS OPTIONS request bypass
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test OPTIONS requests bypass payment verification
  - **Test Logic**: 
    - Create test RequestEvent with method = "OPTIONS"
    - No X-PAYMENT header
    - Call middleware
    - Assert e.Next() WAS called (handler executes)
    - Assert no 402 response (bypassed verification)
  - **Verification**: Test compiles and runs

- [X] T025d [US1] Write test for facilitator timeout scenario
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test facilitator timeout returns HTTP 503
  - **Test Logic**: 
    - Mock facilitator Verify() to timeout (context.DeadlineExceeded)
    - Create test RequestEvent with valid payment
    - Call middleware
    - Assert HTTP 503 status
    - Assert error response mentions timeout
  - **Verification**: Test compiles and runs

- [X] T025e [P] [US1] Write test for EVM network verification (base-sepolia)
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test payment verification with base-sepolia network
  - **Test Logic**: 
    - Create payment with scheme=eip3009, network=base-sepolia
    - Mock facilitator Verify() to succeed
    - Verify requirement matching by network
    - Assert successful verification
  - **Verification**: Test compiles and runs

- [X] T025f [P] [US1] Write test for SVM network enrichment (solana-devnet)
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test SVM network enrichment adds feePayer field
  - **Test Logic**: 
    - Create requirement with scheme=svm, network=solana-devnet
    - Mock facilitator EnrichRequirements() to add feePayer field
    - Call middleware factory (triggers enrichment)
    - Assert enriched requirement contains feePayer
    - Log enrichment success
  - **Verification**: Test compiles and runs

- [X] T026 [US1] Run all User Story 1 tests and verify pass
  - **Command**: `go test -race -v ./http/pocketbase/...`
  - **Action**: Execute all tests for basic payment gating (including edge cases T025a-T025f)
  - **Verification**: All tests pass with no race conditions

---

## Phase 4: User Story 2 - PocketBase Context Integration (P2)

**User Story**: A developer needs access to payment details (payer address, verification status) within their PocketBase handler after successful payment verification. This information should be available through the PocketBase request store using e.Get("x402_payment").

**Why P2**: Enables developers to build payment-aware features (logging, analytics, user tracking) but the core payment gating works without it.

**Independent Test Criteria**: 
- Can create protected handler that accesses e.Get("x402_payment")
- VerifyResponse contains Payer, IsValid, InvalidReason fields
- Payment details are accessible after middleware execution
- Multiple handlers can access same payment data

**Prerequisites**: User Story 1 complete (T018 already implements payment storage in context). US2 tasks verify and document this existing functionality; they do not implement new storage logic.

**Acceptance Scenarios** (from spec.md):
1. Valid payment processed → VerifyResponse available via e.Get("x402_payment")
2. Handler accesses payment details → Payer, IsValid, InvalidReason available

### Tasks

- [X] T027 [P] [US2] Write test for accessing payment details in handler
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test handler can retrieve payment data from request store
  - **Test Logic**: 
    - Create test handler that calls e.Get("x402_payment")
    - Type assert to *VerifyResponse
    - Verify Payer, IsValid, InvalidReason fields accessible
    - Verify values match expected payment
  - **Verification**: Test compiles and runs

- [X] T028 [P] [US2] Write test for payment data structure
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test VerifyResponse contains all required fields
  - **Test Logic**: 
    - Create test with valid payment
    - After middleware execution, retrieve payment data
    - Assert verifyResp.Payer is valid address
    - Assert verifyResp.IsValid == true
    - Assert verifyResp.InvalidReason is empty for valid payment
  - **Verification**: Test compiles and runs

- [X] T029 [P] [US2] Write test for invalid payment data
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test VerifyResponse contains InvalidReason when payment invalid
  - **Test Logic**: 
    - Mock facilitator to return IsValid=false with reason
    - Verify InvalidReason field populated
    - Verify handler does NOT execute (no e.Next call)
  - **Verification**: Test compiles and runs

- [X] T030 [US2] Create example handler demonstrating payment data access
  - **Path**: `/home/space_cowboy/Workspace/x402-go/examples/pocketbase/main.go`
  - **Action**: Create example PocketBase app with handler accessing payment details
  - **Content**: 
    ```go
    se.Router.GET("/api/premium/data", func(e *core.RequestEvent) error {
        payment := e.Get("x402_payment").(*httpx402.VerifyResponse)
        return e.JSON(200, map[string]any{
            "data": "Premium content",
            "payer": payment.Payer,
        })
    }).BindFunc(middleware)
    ```
  - **Verification**: Example compiles

- [X] T031 [US2] Add payment data access section to README
  - **Path**: `/home/space_cowboy/Workspace/x402-go/examples/pocketbase/README.md`
  - **Action**: Document how to access payment details in handlers
  - **Content**: Code example showing e.Get("x402_payment") usage
  - **Verification**: README section exists

- [X] T032 [US2] Add godoc comments for request store key
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Document "x402_payment" key in package comment
  - **Content**: 
    ```go
    // After successful verification, payment details are stored in the request store
    // with key "x402_payment" as *VerifyResponse. Handlers can access via:
    //   verifyResp := e.Get("x402_payment").(*VerifyResponse)
    ```
  - **Verification**: Godoc renders correctly

- [X] T033 [US2] Write integration test for multiple handlers accessing payment data
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test multiple handlers in chain can all access payment data
  - **Test Logic**: 
    - Create handler chain: middleware → handler1 → handler2
    - Verify both handlers can access e.Get("x402_payment")
    - Verify payment data is same in both handlers
  - **Verification**: Test compiles and runs

- [X] T034 [US2] Run all User Story 2 tests and verify pass
  - **Command**: `go test -race -v ./http/pocketbase/... -run TestContext`
  - **Action**: Execute all tests for context integration
  - **Verification**: All US2 tests pass

---

## Phase 5: User Story 3 - Verify-Only Mode (P2)

**User Story**: A developer needs to verify payments without settling them (for testing or when settlement is handled separately). They should be able to enable verify-only mode via Config.VerifyOnly flag matching stdlib middleware behavior.

**Why P2**: Essential for testing and certain deployment scenarios, but not needed for basic payment gating.

**Independent Test Criteria**: 
- Can enable VerifyOnly=true in config
- Verification succeeds but settlement is skipped
- No X-PAYMENT-RESPONSE header added
- VerifyOnly=false performs both verification and settlement

**Prerequisites**: User Story 1 complete (T017 has settlement logic with VerifyOnly check)

**Acceptance Scenarios** (from spec.md):
1. VerifyOnly=true + valid payment → Verification succeeds, settlement skipped
2. VerifyOnly=true + valid payment → No X-PAYMENT-RESPONSE header
3. VerifyOnly=false + valid payment → Both verification and settlement performed

### Tasks

- [X] T035 [P] [US3] Write test for VerifyOnly=true skips settlement
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test VerifyOnly mode skips facilitator.Settle() call
  - **Test Logic**: 
    - Create config with VerifyOnly=true
    - Mock facilitator Verify() to succeed
    - Track if Settle() was called (should NOT be called)
    - Verify handler executes (e.Next() called)
    - Verify payment data stored
  - **Verification**: Test compiles and runs

- [X] T036 [P] [US3] Write test for VerifyOnly=true has no X-PAYMENT-RESPONSE header
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test no settlement header in verify-only mode
  - **Test Logic**: 
    - Create config with VerifyOnly=true
    - Process valid payment
    - Assert X-PAYMENT-RESPONSE header is NOT present
  - **Verification**: Test compiles and runs

- [X] T037 [US3] Write test for VerifyOnly=false performs settlement
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware_test.go`
  - **Action**: Test default mode performs both verification and settlement
  - **Test Logic**: 
    - Create config with VerifyOnly=false (or omit field)
    - Mock both Verify() and Settle() to succeed
    - Track that Settle() WAS called
    - Verify X-PAYMENT-RESPONSE header present
  - **Verification**: Test compiles and runs

- [X] T038 [P] [US3] Add VerifyOnly example to quickstart
  - **Path**: `/home/space_cowboy/Workspace/x402-go/specs/005-pocketbase-middleware/quickstart.md`
  - **Action**: Already exists in quickstart.md, verify it's correct
  - **Verification**: Example shows VerifyOnly configuration

- [X] T039 [US3] Run all User Story 3 tests and verify pass
  - **Command**: `go test -race -v ./http/pocketbase/... -run TestVerifyOnly`
  - **Action**: Execute all tests for verify-only mode
  - **Verification**: All US3 tests pass

---

## Phase 6: Polish & Cross-Cutting Concerns

**Goal**: Finalize implementation with documentation, examples, and validation

**Prerequisites**: All user stories complete

**Completion Criteria**: 
- All tests passing with race detection
- Code coverage maintained or improved
- Linter passing
- Example builds and runs
- Documentation complete
- No binaries committed

### Tasks

- [X] T040 [P] Add comprehensive godoc comments to exported functions
  - **Path**: `/home/space_cowboy/Workspace/x402-go/http/pocketbase/middleware.go`
  - **Action**: Add godoc comments following Go standards
  - **Content**: 
    - NewPocketBaseX402Middleware: Describe factory, parameters, return value, usage example
    - Reference stdlib and Gin middleware for consistency
  - **Verification**: `go doc` shows complete documentation

- [X] T041 [P] Create complete example PocketBase application
  - **Path**: `/home/space_cowboy/Workspace/x402-go/examples/pocketbase/main.go`
  - **Action**: Create working example showing all features
  - **Content**: 
    - Basic configuration
    - Protected endpoint
    - Handler accessing payment data
    - Group-level middleware
    - Comments explaining each part
  - **Reference**: `/home/space_cowboy/Workspace/x402-go/specs/005-pocketbase-middleware/quickstart.md`
  - **Verification**: Example compiles and runs

- [X] T042 Create examples/pocketbase/README.md with usage guide
  - **Path**: `/home/space_cowboy/Workspace/x402-go/examples/pocketbase/README.md`
  - **Action**: Create README with quickstart, configuration, examples
  - **Content**: 
    - Installation
    - Basic usage
    - Configuration options
    - Payment data access
    - Verify-only mode
    - Troubleshooting
  - **Reference**: Copy relevant sections from quickstart.md
  - **Verification**: README is comprehensive

- [X] T043 Run final validation suite
  - **Command Sequence**:
    ```bash
    # Run all tests with race detection
    go test -race ./http/pocketbase/...
    
    # Check coverage
    go test -race -cover ./http/pocketbase/...
    
    # Run linter
    golangci-lint run http/pocketbase/
    
    # Format code
    go fmt ./http/pocketbase/...
    
    # Vet code
    go vet ./http/pocketbase/...
    
    # Build example
    go build -o /tmp/pocketbase-example ./examples/pocketbase/
    
    # Clean binaries
    rm /tmp/pocketbase-example
    
    # Verify no binaries in repo
    git status | grep -E '\.(exe|out|test)$' && echo "ERROR: Binaries found" || echo "OK"
    ```
  - **Action**: Run all validation commands
  - **Verification**: All commands succeed, no binaries committed

- [X] T044 Update AGENTS.md with new technology stack
  - **Command**: `.specify/scripts/bash/update-agent-context.sh opencode`
  - **Action**: Run agent context update script to add PocketBase middleware to AGENTS.md
  - **Content**: 
    - Technology: PocketBase framework (github.com/pocketbase/pocketbase)
    - Recent change: "005-pocketbase-middleware: Added PocketBase middleware adapter for x402 payment gating"
  - **Verification**: AGENTS.md updated with PocketBase entry
  - **Note**: This task updates project documentation and can run independently after implementation is complete

---

## Implementation Strategy

### MVP Scope (Recommended)

**Minimum Viable Product**: Complete Phase 1, Phase 2, and Phase 3 only (User Story 1)

This delivers:
- ✅ Core payment gating functionality
- ✅ Request verification and settlement
- ✅ Error handling (402, 400, 503)
- ✅ Route and group-level binding
- ✅ Comprehensive tests

**Time Estimate**: 4-6 hours for experienced Go developer

**Value**: PocketBase users can protect endpoints with x402 payment gating immediately

### Incremental Delivery

After MVP, add features incrementally:

1. **Phase 4** (US2 - Context Integration): +2 hours
   - Enables payment-aware application logic
   - Independent of US3

2. **Phase 5** (US3 - Verify-Only Mode): +1 hour
   - Enables testing scenarios
   - Independent of US2

3. **Phase 6** (Polish): +1 hour
   - Documentation and examples
   - Can be done in parallel with US2/US3

### Parallel Development Opportunities

**After US1 Complete**:
- Developer A: Implement US2 (Context Integration)
- Developer B: Implement US3 (Verify-Only Mode)
- Developer C: Work on Phase 6 (Polish & Examples)

No conflicts - these are independent work streams.

---

## Validation Checklist

### Before Implementation
- [ ] All design documents reviewed (plan.md, spec.md, data-model.md, research.md)
- [ ] Dependencies verified (PocketBase, x402-go core, http package)
- [ ] Project structure understood

### After Each User Story
- [ ] All story tests passing (`go test -race ./http/pocketbase/...`)
- [ ] Code formatted (`go fmt ./http/pocketbase/...`)
- [ ] No linter errors (`golangci-lint run http/pocketbase/`)
- [ ] Story acceptance criteria met

### Final Validation (Phase 6)
- [ ] All 54 tasks completed
- [ ] All tests passing with race detection
- [ ] Code coverage maintained or improved
- [ ] Linter passing (golangci-lint)
- [ ] Example builds and runs
- [ ] No binaries committed to git
- [ ] Documentation complete (godoc + README)

---

## Success Criteria Verification

Map tasks to success criteria from spec.md:

| Success Criterion | Related Tasks | Verification |
|-------------------|---------------|--------------|
| SC-001: Protect endpoint with http.Config | T010, T041 | Example in main.go works |
| SC-002: 100% test scenarios pass | T019-T025 | All core tests pass |
| SC-003: Payment details via e.Get() | T027-T029 | US2 tests pass |
| SC-004: Support EVM & SVM networks | T011 | Enrichment tests pass |
| SC-005: Use stdlib Config | T010 | Import http.Config |
| SC-006: 3+ core tests | T019-T025 | 7 tests in US1 |
| SC-007: Bind/BindFunc support | T024, T025 | Binding tests pass |

---

## Notes

**Task Format**: All tasks follow the required checklist format:
- `- [ ]` checkbox
- `TXXX` task ID (sequential)
- `[P]` marker for parallelizable tasks
- `[USX]` label for user story tasks
- Clear description with file path

**Helper Functions**: Tasks T006-T009 (Foundational phase) implement four duplicated framework-specific helper functions (parsePaymentHeaderFromRequest, sendPaymentRequiredPocketBase, findMatchingRequirementPocketBase, addPaymentResponseHeaderPocketBase) following the established Gin middleware pattern. These maintain self-contained adapters and are blocking prerequisites for all user stories.

**Test Strategy**: Tests are integrated into each user story phase rather than separated. This ensures each story is independently testable and complete.

**File Paths**: All tasks include absolute file paths for clarity and LLM execution.

**References**: Many tasks include references to existing code (stdlib/Gin middleware) to guide implementation.

---

**Generated**: 2025-10-29 via `/speckit.tasks` command  
**Total Tasks**: 54  
**Estimated Time**: 10-12 hours total (5-7 hours for MVP)  
**Status**: Ready for implementation
