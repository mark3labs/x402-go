# Feature Specification: PocketBase Middleware for x402 Payment Protocol

**Feature Branch**: `005-pocketbase-middleware`  
**Created**: 2025-10-29  
**Status**: Draft  
**Input**: User description: "I want to add a new pocketbase compatible middleware in the spirit of the stdlib http and gin middleware. The middleware is simply a minimal implementation that reuses all the same parts of the main repo for x402 paywalling. Check the old specs"

## Clarifications

### Session 2025-10-29

- Q: When the middleware returns HTTP 402 for missing or invalid payment, what specific JSON structure should be returned to the client? → A: Follow x402 HTTP spec and match stdlib/Gin middleware implementation (PaymentRequirementsResponse with x402Version, error, accepts fields)
- Q: When implementing PocketBase-specific error responses, should the middleware use PocketBase's native error types (e.g., `e.BadRequestError()`, `apis.NewBadRequestError()`) or send standard JSON responses directly using `e.JSON()`? → A: Standard JSON using e.JSON() - maintains exact x402 protocol compliance
- Q: The Gin middleware duplicates helper functions (`parsePaymentHeaderFromRequest`, `sendPaymentRequiredGin`, `findMatchingRequirementGin`, `addPaymentResponseHeaderGin`) instead of reusing stdlib helpers. Should PocketBase middleware follow the same duplication pattern or attempt to reuse stdlib helpers? → A: Duplicate helpers - follows established Gin pattern, self-contained implementation
- Q: The stdlib and Gin middleware both hardcode facilitator client timeouts (VerifyTimeout: 5s, SettleTimeout: 60s). Should these same timeout values be used in PocketBase middleware? → A: Yes - Use same hardcoded timeouts (5s verify, 60s settle)
- Q: The Gin middleware calls `facilitator.EnrichRequirements()` at startup to fetch network-specific configuration (like `feePayer` for SVM chains) from the facilitator's `/supported` endpoint. Should PocketBase middleware implement this same enrichment step? → A: Yes - Implement enrichment step

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic Payment Gating with PocketBase (Priority: P1)

A developer building APIs with the PocketBase framework needs to protect their custom endpoints with x402 payment gating. They should be able to apply the middleware to specific routes or route groups without changing their existing PocketBase application structure.

**Why this priority**: This is the core functionality - without it, PocketBase users cannot use x402 payment gating at all. This represents the minimum viable product.

**Independent Test**: Can be fully tested by creating a simple PocketBase application with a protected endpoint, sending requests with and without valid X-PAYMENT headers, and verifying 402 responses and successful access with valid payment. Delivers immediate value as developers can gate any PocketBase endpoint.

**Acceptance Scenarios**:

1. **Given** a PocketBase route protected by x402 middleware, **When** a request arrives without X-PAYMENT header, **Then** the response is HTTP 402 with payment requirements in JSON format
2. **Given** a PocketBase route protected by x402 middleware, **When** a request arrives with valid X-PAYMENT header, **Then** the payment is verified, settled, and the protected handler is executed via e.Next()
3. **Given** a PocketBase route protected by x402 middleware, **When** a request arrives with invalid X-PAYMENT header, **Then** the response is HTTP 402 with payment requirements
4. **Given** a PocketBase route protected by x402 middleware, **When** payment verification fails at the facilitator, **Then** the response is HTTP 402 with appropriate error details
5. **Given** a PocketBase route protected by x402 middleware, **When** a request arrives with malformed X-PAYMENT header (invalid base64 or JSON), **Then** the response is HTTP 400 Bad Request with x402Version error response

---

### User Story 2 - PocketBase Context Integration (Priority: P2)

A developer needs access to payment details (payer address, verification status) within their PocketBase handler after successful payment verification. This information should be available through the PocketBase request store using e.Get("x402_payment").

**Why this priority**: Enables developers to build payment-aware features (logging, analytics, user tracking) but the core payment gating works without it.

**Independent Test**: Can be tested by creating a protected handler that accesses payment details from PocketBase request store and returns them in the response. Delivers value by enabling payment-aware application logic.

**Acceptance Scenarios**:

1. **Given** a protected PocketBase handler with payment middleware, **When** a valid payment is processed, **Then** payment verification details (VerifyResponse) are available via e.Get("x402_payment")
2. **Given** payment details stored in request store, **When** the handler accesses them, **Then** payer address, IsValid status, and InvalidReason fields are available

---

### User Story 3 - Verify-Only Mode (Priority: P2)

A developer needs to verify payments without settling them (for testing or when settlement is handled separately). They should be able to enable verify-only mode via Config.VerifyOnly flag matching stdlib middleware behavior.

**Why this priority**: Essential for testing and certain deployment scenarios, but not needed for basic payment gating.

**Independent Test**: Can be tested by enabling VerifyOnly flag and verifying that settlement is skipped after successful verification. Delivers value by supporting test environments.

**Acceptance Scenarios**:

1. **Given** middleware configured with VerifyOnly=true, **When** a valid payment is provided, **Then** verification succeeds but settlement is skipped
2. **Given** middleware configured with VerifyOnly=true, **When** a valid payment is processed, **Then** no X-PAYMENT-RESPONSE header is added to response
3. **Given** middleware configured with VerifyOnly=false (default), **When** a valid payment is provided, **Then** both verification and settlement are performed

---

### Edge Cases

- When facilitator service is unavailable or times out during verification, middleware returns HTTP 503 Service Unavailable
- Malformed X-PAYMENT headers (invalid base64 or JSON) return HTTP 400 Bad Request with x402Version error response
- When payment amount is insufficient but payment header is otherwise valid, middleware returns HTTP 402 with payment requirements (facilitator verification fails)
- Middleware behaves identically when applied to PocketBase route groups vs. individual routes (uses same hook.Handler signature)
- When settlement fails after successful verification, middleware returns HTTP 503 Service Unavailable with error details
- CORS preflight OPTIONS requests bypass payment verification (follow web standards)
- Replay attacks are handled by facilitator service and signature scheme, not middleware concern

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a PocketBase-compatible middleware function (hook.Handler[*core.RequestEvent]) that wraps handlers with x402 payment gating
- **FR-002**: Middleware MUST accept payment requirements via Config struct PaymentRequirements field (amount specified in atomic units as string using USDC's 6 decimal precision. USDC has 6 decimals, so 1 USDC = 1,000,000 atomic units. Example: "10000" = 0.01 USDC)
- **FR-003**: Middleware MUST accept recipient wallet address via PaymentRequirement.PayTo field in Config
- **FR-004**: Middleware MUST accept configuration through Config struct matching stdlib middleware (FacilitatorURL, FallbackFacilitatorURL, PaymentRequirements, VerifyOnly fields)
- **FR-005**: Middleware MUST check for X-PAYMENT header in incoming requests via e.Request.Header.Get("X-Payment")
- **FR-006**: Middleware MUST return HTTP 402 with PaymentRequirementsResponse JSON (x402Version=1, error message, accepts array) when X-PAYMENT header is missing; MUST return HTTP 400 Bad Request with x402Version error response for malformed headers (invalid base64/JSON)
- **FR-007**: Middleware MUST parse and decode base64-encoded X-PAYMENT header containing JSON payment payload
- **FR-008**: Middleware MUST verify payments by calling facilitator's /verify endpoint
- **FR-009**: Middleware MUST settle payments by calling facilitator's /settle endpoint (unless verify-only mode is enabled)
- **FR-010**: Middleware MUST store payment verification details in PocketBase request store using e.Set("x402_payment", verifyResp) with VerifyResponse struct for handler access
- **FR-011**: Middleware MUST add X-PAYMENT-RESPONSE header with base64-encoded settlement details using PocketBase-specific helper function (addPaymentResponseHeaderPocketBase) following Gin middleware duplication pattern
- **FR-012**: Middleware MUST support testnet mode via Config.PaymentRequirements network field (base-sepolia or base) - no default, user specifies explicitly
- **FR-013**: Middleware MUST use network-appropriate USDC contract address specified in PaymentRequirement.Asset field (user-configured; standard addresses: base = 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913, base-sepolia = 0x036CbD53842c5426634e7929541eC2318f3dCF7e)
- **FR-014**: Middleware MUST construct payment requirements with resource URL from incoming request using same logic as stdlib middleware (scheme + host + requestURI)
- **FR-015**: Middleware MUST call e.Next() to continue handler chain after successful payment verification
- **FR-016**: Middleware MUST NOT call e.Next() when payment verification fails (return error directly)
- **FR-017**: Middleware MUST work with both route-level binding (Route.Bind()) and group-level binding (RouterGroup.Bind())
- **FR-018**: Middleware MUST return JSON responses using e.JSON() method matching stdlib/Gin middleware behavior: HTTP 402 with PaymentRequirementsResponse for missing/invalid payments, HTTP 400 for malformed headers, HTTP 503 for facilitator failures. Error paths must return the result of e.JSON() without calling e.Next() to stop handler chain execution (equivalent to Gin's c.AbortWithStatusJSON pattern).
- **FR-019**: Middleware MUST return error responses with x402Version field and structured error messages matching x402 HTTP transport specification (not using PocketBase native error types like e.BadRequestError)
- **FR-020**: Middleware MUST create FacilitatorClient with hardcoded timeouts matching stdlib/Gin: VerifyTimeout=5s, SettleTimeout=60s
- **FR-021**: Middleware MUST call facilitator.EnrichRequirements() at initialization to fetch network-specific configuration (e.g., feePayer for SVM chains) from facilitator's /supported endpoint
- **FR-022**: Middleware MUST log warning and continue with original requirements if EnrichRequirements() fails (graceful degradation matching stdlib/Gin behavior). Enrichment failures include: network timeout, HTTP non-2xx response, invalid JSON response, or missing expected fields in /supported endpoint response.

### Key Entities

- **PocketBaseMiddleware**: The PocketBase middleware handler with signature `func(*core.RequestEvent) error` compatible with PocketBase's `hook.Handler[*core.RequestEvent]` type that enforces payment gating on protected routes. Configured via stdlib Config struct; translates core.RequestEvent to stdlib http patterns; reuses all logic from http/middleware.go.

- **Config**: Configuration structure from stdlib http package (http.Config) containing FacilitatorURL, FallbackFacilitatorURL, PaymentRequirements slice, and VerifyOnly flag. Shared between stdlib, Gin, and PocketBase middleware for consistency.

- **VerifyResponse**: Payment verification result from http/facilitator.go stored in request store with key "x402_payment". Contains Payer address, IsValid status, and InvalidReason if verification failed.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can protect a PocketBase endpoint with x402 payment gating by passing stdlib http.Config to PocketBase middleware function
- **SC-002**: Middleware correctly handles 100% of test scenarios matching stdlib middleware_test.go: missing payment (402 response), invalid payment (400/402 response), valid payment (handler executes via e.Next()), and verification failures (503 response)
- **SC-003**: Payment verification details (VerifyResponse with Payer, IsValid, InvalidReason fields) are accessible in protected handler via e.Get("x402_payment")
- **SC-004**: Middleware supports both EVM networks (base, base-sepolia) and SVM networks (solana-mainnet, solana-devnet) via PaymentRequirement.Network field with automatic enrichment from facilitator
- **SC-005**: Configuration uses same Config struct as stdlib middleware for consistency (fewer than 15 lines typical setup)
- **SC-006**: PocketBase middleware passes all test scenarios from stdlib middleware_test.go adapted to core.RequestEvent (3+ core tests)
- **SC-007**: Middleware integrates with PocketBase's Bind/BindFunc methods for group-level and route-level attachment

## Assumptions *(mandatory)*

- The PocketBase framework (github.com/pocketbase/pocketbase) is already installed and available as a dependency in the project
- Developers using this middleware are familiar with basic PocketBase concepts (middleware, request events, handlers)
- The existing stdlib http middleware (http/middleware.go) is fully functional and tested
- The Gin middleware helper duplication pattern (parsePaymentHeaderFromRequest, sendPaymentRequiredGin, etc.) is the established approach for framework adapters
- The facilitator service provides /verify and /settle endpoints with the expected API contracts
- USDC is the primary payment token (specified via PaymentRequirement.Asset field)
- Both EVM networks (base, base-sepolia) and SVM networks (solana-mainnet, solana-devnet) are supported via PaymentRequirement.Network field
- The facilitator's /supported endpoint provides network-specific configuration (like feePayer for SVM) that is enriched at middleware initialization
- Payments are stateless - the middleware does not track payment history or nonce values (delegated to facilitator)
- The PocketBase application is served over HTTP/HTTPS and request URLs can be constructed from core.RequestEvent request metadata
- PocketBase middleware should match stdlib middleware behavior exactly (verify-then-settle, not write-then-settle)

## Out of Scope *(mandatory)*

- Support for frameworks other than PocketBase (chi, echo, fiber, etc.) - those would be separate features
- Payment tracking or analytics - middleware is stateless, tracking must be implemented by developers in handlers
- Rate limiting or abuse prevention - should be handled at infrastructure layer or in custom handler logic
- Multi-token support beyond USDC - would require additional specification for token selection and configuration
- Dynamic pricing based on request parameters - developers must implement custom middleware if needed
- Payment refund or reversal functionality - not part of payment verification flow
- Webhook notifications for payment events - would be a separate feature
- Payment batching or aggregation - each request is independently gated
- Functional options pattern (WithTestnet, WithFacilitatorURL, etc.) - PocketBase middleware uses same Config struct as stdlib middleware for consistency
- HTML paywall for browser requests - PocketBase middleware matches stdlib behavior (JSON-only responses)
- Custom ResponseWriter for write-then-settle pattern - PocketBase middleware uses verify-then-settle like stdlib (no response buffering)
- Per-request timeout configuration - Timeouts are hardcoded in FacilitatorClient matching stdlib/Gin behavior (VerifyTimeout=5s, SettleTimeout=60s)
- Browser detection based on Accept/User-Agent headers - not implemented in stdlib, not needed for PocketBase adapter
- Integration with PocketBase's built-in auth system - developers can access e.Auth separately and combine if needed
- Automatic PocketBase collection/record creation for payment tracking - developers implement custom tracking if needed

## Dependencies *(mandatory)*

- PocketBase framework (github.com/pocketbase/pocketbase) must be available as a dependency. **Stdlib-first justification**: PocketBase applications use `core.RequestEvent` (not `http.Request`) and a hook-based middleware system (`hook.Handler[*core.RequestEvent]`) that is incompatible with stdlib `http.Handler` interface. A framework-specific adapter is required to bridge PocketBase's request handling to x402's stdlib-based payment logic.
- Existing x402-go core package must be available for type definitions
- Existing http/facilitator.go client implementation must be available for payment verification and settlement
- Facilitator service must be running and accessible at configured URL (defaults to Coinbase facilitator)
- USDC smart contracts must be deployed on target networks (base, base-sepolia)

## Notes

This specification is based on the existing stdlib and Gin middleware patterns but adapted for PocketBase's unique middleware system:
- Uses PocketBase's hook.Handler[*core.RequestEvent] type instead of http.HandlerFunc or gin.HandlerFunc
- Uses e.Set/e.Get for request store instead of context.Context or gin.Context
- Uses e.Next() to continue handler chain instead of calling next handler directly
- Uses e.JSON() for x402 protocol-compliant responses (not PocketBase native error types)

Key differences from stdlib/Gin implementations:
- Lives in http/pocketbase/ package for consistency with http/gin/
- Wraps core.RequestEvent instead of http.Request/gin.Context
- Uses e.JSON() for all error responses to maintain x402 protocol compliance (not PocketBase native error types)
- Uses PocketBase's request store (e.Set/e.Get) instead of context values
- Middleware signature is hook.Handler[*core.RequestEvent] for Bind/BindFunc compatibility

**Helper Function Pattern**: Following the Gin middleware pattern, this middleware duplicates four framework-specific helper functions (parsePaymentHeaderFromRequest, sendPaymentRequiredPocketBase, findMatchingRequirementPocketBase, addPaymentResponseHeaderPocketBase) rather than sharing helpers with stdlib. This duplication maintains self-contained adapters and avoids framework coupling.
