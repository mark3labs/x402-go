# Feature Specification: Chi Middleware for x402 Payment Protocol

**Feature Branch**: `006-chi-middleware`  
**Created**: 2025-10-29  
**Status**: Draft  
**Input**: User description: "I want to add a Chi flavored middleware. It should have all the same functionality as `http/middleware.go` and `http/gin/middleware.go` We don't care about performance or speed. Replay and TXs are not a concern of the middleware. They are a facilitator concern. Check previous middleware specs when making this spec. It should be a very minimal middleware"

## Clarifications

### Session 2025-10-29

- Q: The spec states that helper functions like `parsePaymentHeaderFromRequest`, `sendPaymentRequiredChi`, etc. should be duplicated rather than shared with stdlib (following Gin/PocketBase pattern). However, this creates maintenance burden and potential inconsistency. → A: Share helpers with stdlib (create shared internal package) since Chi uses identical http.Request/ResponseWriter types
- Q: The spec mentions that EnrichRequirements() should be called at initialization and log a warning on failure (FR-018, FR-019). However, "initialization" timing is ambiguous for Chi middleware. → A: During middleware constructor function
- Q: The spec mentions CORS preflight OPTIONS requests should bypass payment verification (Edge Cases, line 71). However, the implementation approach for detecting and bypassing OPTIONS requests is not specified. → A: Check r.Method == "OPTIONS" and call next handler immediately (skip all payment logic)
- Q: The spec states middleware should have a constructor function (implied by FR-018 "constructor function (NewChiX402Middleware)"). However, the exact function signature and return type need clarification for Chi's middleware pattern. → A: func NewChiX402Middleware(config *httpx402.Config) func(http.Handler) http.Handler
- Q: The spec requires logging via slog.Default() for enrichment warnings (FR-019) and references Gin's logging pattern. However, logger configuration strategy for other events (verification, settlement, errors) is not specified. → A: Use slog.Default() for all logging (Info, Warn, Error) matching Gin middleware pattern

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic Payment Gating with Chi (Priority: P1)

A developer building a REST API with the Chi router needs to protect their endpoints with x402 payment gating. They should be able to apply the middleware to specific routes or route groups without changing their existing Chi application structure.

**Why this priority**: This is the core functionality - without it, Chi users cannot use x402 payment gating at all. This represents the minimum viable product.

**Independent Test**: Can be fully tested by creating a simple Chi application with a protected endpoint, sending requests with and without valid X-PAYMENT headers, and verifying 402 responses and successful access with valid payment. Delivers immediate value as developers can gate any Chi endpoint.

**Acceptance Scenarios**:

1. **Given** a Chi route protected by x402 middleware, **When** a request arrives without X-PAYMENT header, **Then** the response is HTTP 402 with payment requirements in JSON format
2. **Given** a Chi route protected by x402 middleware, **When** a request arrives with valid X-PAYMENT header, **Then** the payment is verified, settled, and the protected handler is executed
3. **Given** a Chi route protected by x402 middleware, **When** a request arrives with invalid X-PAYMENT header, **Then** the response is HTTP 400 Bad Request with x402Version error response
4. **Given** a Chi route protected by x402 middleware, **When** payment verification fails at the facilitator, **Then** the response is HTTP 503 Service Unavailable with appropriate error details

---

### User Story 2 - Context Integration (Priority: P2)

A developer needs access to payment details (payer address, verification status) within their Chi handler after successful payment verification. This information should be available through the standard Chi request context.

**Why this priority**: Enables developers to build payment-aware features (logging, analytics, user tracking) but the core payment gating works without it.

**Independent Test**: Can be tested by creating a protected handler that accesses payment details from request context and returns them in the response. Delivers value by enabling payment-aware application logic.

**Acceptance Scenarios**:

1. **Given** a protected Chi handler with payment middleware, **When** a valid payment is processed, **Then** payment verification details (VerifyResponse) are available via context.Value(httpx402.PaymentContextKey)
2. **Given** payment details stored in request context, **When** the handler accesses them, **Then** payer address, IsValid status, and InvalidReason fields are available

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
- Middleware behaves identically when applied to Chi route groups vs. individual routes
- When settlement fails after successful verification, middleware returns HTTP 503 Service Unavailable with error details
- CORS preflight OPTIONS requests bypass payment verification by checking r.Method == "OPTIONS" and calling next handler immediately, skipping all payment logic (follows web standards for method-safe requests)
- Replay attacks are handled by facilitator service and signature scheme, not middleware concern

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a Chi-compatible middleware constructor function NewChiX402Middleware(config *httpx402.Config) that returns func(http.Handler) http.Handler for wrapping handlers with x402 payment gating
- **FR-002**: Middleware MUST accept payment requirements via Config struct PaymentRequirements field (amount specified in atomic units as string, e.g., "10000" for 0.01 USDC)
- **FR-003**: Middleware MUST accept recipient wallet address via PaymentRequirement.PayTo field in Config
- **FR-004**: Middleware MUST accept configuration through Config struct matching stdlib middleware (FacilitatorURL, FallbackFacilitatorURL, PaymentRequirements, VerifyOnly fields)
- **FR-005**: Middleware MUST check for X-PAYMENT header in incoming requests
- **FR-006**: Middleware MUST return HTTP 402 with payment requirements JSON when X-PAYMENT header is missing or invalid; MUST return HTTP 400 Bad Request with x402Version error response for malformed headers (invalid base64/JSON)
- **FR-007**: Middleware MUST parse and decode base64-encoded X-PAYMENT header containing JSON payment payload
- **FR-008**: Middleware MUST verify payments by calling facilitator's /verify endpoint
- **FR-009**: Middleware MUST settle payments by calling facilitator's /settle endpoint (unless verify-only mode is enabled)
- **FR-010**: Middleware MUST store VerifyResponse from payment verification in request context using stdlib context.WithValue with httpx402.PaymentContextKey constant for handler access
- **FR-011**: Middleware MUST add X-PAYMENT-RESPONSE header with base64-encoded settlement details following Gin/PocketBase middleware pattern
- **FR-012**: Middleware MUST support testnet mode via Config.PaymentRequirements network field (base-sepolia or base) - no default, user specifies explicitly
- **FR-013**: Middleware MUST use network-appropriate USDC contract address specified in PaymentRequirement.Asset field (user-configured)
- **FR-014**: Middleware MUST construct payment requirements with resource URL from incoming request using same logic as stdlib middleware (scheme + host + requestURI)
- **FR-015**: Middleware MUST call next handler in chain after successful payment verification
- **FR-016**: Middleware MUST NOT call next handler when payment verification fails (return immediately after writing error response)
- **FR-017**: Middleware MUST create FacilitatorClient with hardcoded timeouts matching stdlib/Gin: VerifyTimeout=5s, SettleTimeout=60s
- **FR-018**: Middleware constructor function (NewChiX402Middleware) MUST call facilitator.EnrichRequirements() before returning the middleware handler to fetch network-specific configuration (e.g., feePayer for SVM chains) from facilitator's /supported endpoint
- **FR-019**: Middleware MUST use slog.Default() for all logging (Info, Warn, Error levels) matching Gin middleware pattern; MUST log warning and continue with original requirements if EnrichRequirements() fails (graceful degradation matching stdlib/Gin/PocketBase behavior)
- **FR-020**: Middleware MUST return error responses with x402Version field and structured error messages matching x402 HTTP transport specification
- **FR-021**: Middleware MUST share helper functions (parsePaymentHeaderFromRequest, sendPaymentRequired, findMatchingRequirement, addPaymentResponseHeader) with stdlib middleware via internal package to reduce duplication and ensure consistency
- **FR-022**: Middleware MUST bypass all payment verification logic for OPTIONS requests (check r.Method == "OPTIONS" and call next handler immediately) to support CORS preflight requests per web standards
- **FR-023**: Middleware MUST log payment verification lifecycle events (missing payment, invalid payment, verification success/failure, settlement success/failure) using slog.Default() with appropriate log levels (Info for success paths, Warn for client errors, Error for service failures) matching Gin middleware logging behavior

### Key Entities

- **NewChiX402Middleware**: Constructor function with signature func(config *httpx402.Config) func(http.Handler) http.Handler that creates and returns Chi middleware. The returned middleware enforces payment gating on protected routes using Chi's standard middleware signature; configured via stdlib Config struct; reuses logic from http/middleware.go.

- **Config**: Configuration structure from stdlib http package (http.Config) containing FacilitatorURL, FallbackFacilitatorURL, PaymentRequirements slice, and VerifyOnly flag. Shared between stdlib, Gin, PocketBase, and Chi middleware for consistency.

- **VerifyResponse**: Payment verification result structure from http/facilitator.go stored in request context with key httpx402.PaymentContextKey. Contains Payer address (string), IsValid status (bool), and InvalidReason (string) if verification failed.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can protect a Chi endpoint with x402 payment gating by passing stdlib http.Config to Chi middleware function
- **SC-002**: Middleware correctly handles 100% of test scenarios matching stdlib middleware_test.go: missing payment (402 response), invalid payment (400/402 response), valid payment (handler executes), and verification failures (503 response)
- **SC-003**: Payment verification and settlement use FacilitatorClient timeouts (5s verify, 60s settle) matching stdlib middleware behavior
- **SC-004**: Payment verification details (VerifyResponse with Payer, IsValid, InvalidReason fields) are accessible in protected handler via context.Value(httpx402.PaymentContextKey)
- **SC-005**: Middleware supports both mainnet (base) and testnet (base-sepolia) networks via PaymentRequirement.Network field in Config
- **SC-006**: Configuration uses same Config struct as stdlib middleware for consistency (fewer than 15 lines typical setup)
- **SC-007**: Chi middleware passes all test scenarios from stdlib middleware_test.go adapted to Chi router (3+ core tests)

## Assumptions *(mandatory)*

- The Chi router (github.com/go-chi/chi/v5) is already installed and available as a dependency in the project
- Developers using this middleware are familiar with basic Chi concepts (middleware, routing, handlers)
- The existing stdlib http middleware (http/middleware.go) is fully functional and tested
- Chi uses standard http.Handler interface, allowing maximum code reuse from stdlib middleware
- The facilitator service provides /verify and /settle endpoints with the expected API contracts
- USDC is the primary payment token (specified via PaymentRequirement.Asset field)
- The base and base-sepolia networks are the primary target networks (configurable via PaymentRequirement.Network)
- Payments are stateless - the middleware does not track payment history or nonce values (delegated to facilitator)
- The Chi application is served over HTTP/HTTPS and request URLs can be constructed from http.Request metadata
- Chi middleware should match stdlib middleware behavior exactly (verify-then-settle, not write-then-settle)

## Out of Scope *(mandatory)*

- Support for frameworks other than Chi (echo, fiber, etc.) - those would be separate features
- Payment tracking or analytics - middleware is stateless, tracking must be implemented by developers in handlers
- Rate limiting or abuse prevention - should be handled at infrastructure layer or in custom handler logic
- Multi-token support beyond USDC - would require additional specification for token selection and configuration
- Dynamic pricing based on request parameters - developers must implement custom middleware if needed
- Payment refund or reversal functionality - not part of payment verification flow
- Webhook notifications for payment events - would be a separate feature
- Payment batching or aggregation - each request is independently gated
- Functional options pattern (WithTestnet, WithFacilitatorURL, etc.) - Chi middleware uses same Config struct as stdlib middleware for consistency
- HTML paywall for browser requests - Chi middleware matches stdlib behavior (JSON-only responses)
- Custom ResponseWriter for write-then-settle pattern - Chi middleware uses verify-then-settle like stdlib (no response buffering)
- Per-request timeout configuration - Timeouts are hardcoded in FacilitatorClient (5s verify, 60s settle) matching stdlib behavior
- Browser detection based on Accept/User-Agent headers - not implemented in stdlib, not needed for Chi adapter

## Dependencies *(mandatory)*

- Chi router (github.com/go-chi/chi/v5) must be available as a dependency
- Existing x402-go core package must be available for type definitions
- Existing http/facilitator.go client implementation must be available for payment verification and settlement
- Facilitator service must be running and accessible at configured URL (defaults to Coinbase facilitator)
- USDC smart contracts must be deployed on target networks (base, base-sepolia)

## Notes

This specification is based on the existing stdlib, Gin, and PocketBase middleware patterns but adapted for Chi's standard http.Handler middleware system:
- Chi uses standard http.Handler interface with func(http.Handler) http.Handler middleware signature
- Chi middleware can maximally reuse stdlib http/middleware.go logic since both use http.Request/http.ResponseWriter
- Uses request context (context.WithValue) for storing payment details, accessible via httpx402.PaymentContextKey
- Uses standard http error responses matching stdlib middleware exactly

Key differences from other implementations:
- Lives in http/chi/ package for consistency with http/gin/ and http/pocketbase/
- Uses Chi's standard middleware signature func(http.Handler) http.Handler (identical to stdlib pattern)
- Since Chi uses stdlib http types, shares helper functions with stdlib middleware via http/internal/helpers package per FR-021
- No framework-specific context types - uses stdlib context.Context directly
