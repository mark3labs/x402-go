# Feature Specification: Gin Middleware for x402 Payment Protocol

**Feature Branch**: `004-gin-middleware`  
**Created**: 2025-10-29  
**Status**: Draft  
**Input**: User description: "I want to add a Gin flavored middleware along side our stdlib http middleware based on this https://raw.githubusercontent.com/coinbase/x402/refs/heads/main/go/pkg/gin/middleware.go should live in the package `http/gin`"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic Payment Gating with Gin (Priority: P1)

A developer building a REST API with the Gin framework needs to protect their endpoints with x402 payment gating. They should be able to apply the middleware to specific routes or route groups without changing their existing Gin application structure.

**Why this priority**: This is the core functionality - without it, Gin users cannot use x402 payment gating at all. This represents the minimum viable product.

**Independent Test**: Can be fully tested by creating a simple Gin application with a protected endpoint, sending requests with and without valid X-PAYMENT headers, and verifying 402 responses and successful access with valid payment. Delivers immediate value as developers can gate any Gin endpoint.

**Acceptance Scenarios**:

1. **Given** a Gin route protected by x402 middleware, **When** a request arrives without X-PAYMENT header, **Then** the response is HTTP 402 with payment requirements in JSON format
2. **Given** a Gin route protected by x402 middleware, **When** a request arrives with valid X-PAYMENT header, **Then** the payment is verified, settled, and the protected handler is executed
3. **Given** a Gin route protected by x402 middleware, **When** a request arrives with invalid X-PAYMENT header, **Then** the response is HTTP 402 with payment requirements
4. **Given** a Gin route protected by x402 middleware, **When** payment verification fails at the facilitator, **Then** the response is HTTP 402 with appropriate error details

---

### User Story 2 - Gin Context Integration (Priority: P2)

A developer needs access to payment details (payer address, verification status) within their Gin handler after successful payment verification. This information should be available through the standard Gin context using c.Get("x402_payment").

**Why this priority**: Enables developers to build payment-aware features (logging, analytics, user tracking) but the core payment gating works without it.

**Independent Test**: Can be tested by creating a protected handler that accesses payment details from Gin context and returns them in the response. Delivers value by enabling payment-aware application logic.

**Acceptance Scenarios**:

1. **Given** a protected Gin handler with payment middleware, **When** a valid payment is processed, **Then** payment verification details (VerifyResponse) are available via c.Get("x402_payment")
2. **Given** payment details stored in Gin context, **When** the handler accesses them, **Then** payer address, IsValid status, and InvalidReason fields are available

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

- When facilitator service is unavailable or times out during verification, middleware returns HTTP 503 Service Unavailable with retry-after header
- How does the system handle malformed X-PAYMENT headers with invalid base64 or JSON?
- What happens when payment amount is insufficient but payment header is otherwise valid?
- How does middleware behave when applied to Gin route groups vs. individual routes?
- What happens when settlement fails after successful verification?
- CORS preflight OPTIONS requests bypass payment verification (follow web standards)
- Replay attacks are handled by facilitator service and signature scheme, not middleware concern

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a Gin-compatible middleware function that wraps handlers with x402 payment gating
- **FR-002**: Middleware MUST accept payment requirements via Config struct PaymentRequirements field (amount specified in atomic units as string, e.g., "10000" for 0.01 USDC)
- **FR-003**: Middleware MUST accept recipient wallet address via PaymentRequirement.PayTo field in Config
- **FR-004**: Middleware MUST accept configuration through Config struct matching stdlib middleware (FacilitatorURL, FallbackFacilitatorURL, PaymentRequirements, VerifyOnly fields)
- **FR-005**: Middleware MUST check for X-PAYMENT header in incoming requests
- **FR-006**: Middleware MUST return HTTP 402 with payment requirements JSON when X-PAYMENT header is missing or invalid
- **FR-007**: Middleware MUST parse and decode base64-encoded X-PAYMENT header containing JSON payment payload
- **FR-008**: Middleware MUST verify payments by calling facilitator's /verify endpoint
- **FR-009**: Middleware MUST settle payments by calling facilitator's /settle endpoint (unless verify-only mode is enabled)
- **FR-010**: Middleware MUST store payment verification details in Gin context using c.Set("x402_payment", verifyResp) with VerifyResponse struct (from http/facilitator.go) for handler access
- **FR-011**: Middleware MUST add X-PAYMENT-RESPONSE header with base64-encoded settlement details by calling addPaymentResponseHeader helper from http/handler.go
- **FR-012**: ~~Middleware MUST detect web browser requests based on Accept and User-Agent headers~~ (OUT OF SCOPE - not in stdlib middleware)
- **FR-013**: ~~Middleware MUST return HTML paywall page for browser requests instead of JSON when payment is required~~ (OUT OF SCOPE - stdlib returns JSON only)
- **FR-014**: ~~Middleware MUST support custom paywall HTML through configuration option~~ (OUT OF SCOPE)
- **FR-015**: Middleware MUST support testnet mode via Config.PaymentRequirements network field (base-sepolia or base) - no default, user specifies explicitly
- **FR-016**: Middleware MUST use network-appropriate USDC contract address specified in PaymentRequirement.Asset field (user-configured)
- **FR-017**: Middleware MUST construct payment requirements with resource URL from incoming request using same logic as stdlib middleware (scheme + host + requestURI)
- **FR-018**: Middleware MUST abort Gin context using c.Abort() and return error when payment verification fails
- **FR-019**: ~~Middleware MUST convert decimal USDC amount to integer representation~~ (OUT OF SCOPE - amounts already in atomic units as strings in PaymentRequirement)
- **FR-020**: ~~Middleware MUST support configurable max timeout seconds~~ (OUT OF SCOPE - timeouts hardcoded in FacilitatorClient: 5s verify, 60s settle)

### Key Entities

- **GinMiddleware**: The Gin handler function (gin.HandlerFunc) that enforces payment gating on protected routes. Configured via stdlib Config struct; translates gin.Context to stdlib http patterns; reuses all logic from http/middleware.go.

- **Config**: Configuration structure from stdlib http package (http.Config) containing FacilitatorURL, FallbackFacilitatorURL, PaymentRequirements slice, and VerifyOnly flag. Shared between stdlib and Gin middleware for consistency.

- **VerifyResponse**: Payment verification result from http/facilitator.go stored in gin.Context with key "x402_payment". Contains Payer address, IsValid status, and InvalidReason if verification failed.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can protect a Gin endpoint with x402 payment gating by passing stdlib http.Config to Gin middleware function
- **SC-002**: Middleware correctly handles 100% of test scenarios matching stdlib middleware_test.go: missing payment (402 response), invalid payment (400/402 response), valid payment (handler executes), and verification failures (503 response)
- **SC-003**: Payment verification and settlement use FacilitatorClient timeouts (5s verify, 60s settle) matching stdlib middleware behavior
- **SC-004**: Payment verification details (VerifyResponse with Payer, IsValid, InvalidReason fields) are accessible in protected handler via c.Get("x402_payment")
- **SC-005**: Middleware supports both mainnet (base) and testnet (base-sepolia) networks via PaymentRequirement.Network field in Config
- **SC-006**: Configuration uses same Config struct as stdlib middleware for consistency (fewer than 15 lines typical setup)
- **SC-007**: Gin middleware passes all test scenarios from stdlib middleware_test.go adapted to gin.Context (3+ core tests)

## Assumptions *(mandatory)*

- The Gin framework is already installed and available as a dependency in the project
- Developers using this middleware are familiar with basic Gin concepts (middleware, context, handlers)
- The existing stdlib http middleware (http/middleware.go) is fully functional and tested
- The existing helper functions (parsePaymentHeader, sendPaymentRequiredWithRequirements, addPaymentResponseHeader, findMatchingRequirement) work correctly
- The facilitator service provides /verify and /settle endpoints with the expected API contracts
- USDC is the primary payment token (specified via PaymentRequirement.Asset field)
- The base and base-sepolia networks are the primary target networks (configurable via PaymentRequirement.Network)
- Payments are stateless - the middleware does not track payment history or nonce values (delegated to facilitator)
- The Gin application is served over HTTP/HTTPS and request URLs can be constructed from gin.Context request metadata
- Gin middleware should match stdlib middleware behavior exactly (verify-then-settle, not write-then-settle)

## Out of Scope *(mandatory)*

- Support for frameworks other than Gin (chi, echo, fiber, etc.) - those would be separate features
- Payment tracking or analytics - middleware is stateless, tracking must be implemented by developers in handlers
- Rate limiting or abuse prevention - should be handled at infrastructure layer or in custom handler logic
- Multi-token support beyond USDC - would require additional specification for token selection and configuration
- Dynamic pricing based on request parameters - developers must implement custom middleware if needed
- Payment refund or reversal functionality - not part of payment verification flow
- Webhook notifications for payment events - would be a separate feature
- Payment batching or aggregation - each request is independently gated
- Functional options pattern (WithTestnet, WithFacilitatorURL, etc.) - Gin middleware uses same Config struct as stdlib middleware for consistency
- HTML paywall for browser requests - Gin middleware matches stdlib behavior (JSON-only responses)
- Custom ResponseWriter for write-then-settle pattern - Gin middleware uses verify-then-settle like stdlib (no response buffering)
- Per-request timeout configuration - Timeouts are hardcoded in FacilitatorClient (5s verify, 60s settle) matching stdlib behavior
- Browser detection based on Accept/User-Agent headers - not implemented in stdlib, not needed for Gin adapter

## Dependencies *(mandatory)*

- Gin web framework (github.com/gin-gonic/gin) must be available as a dependency
- Existing x402-go core package must be available for type definitions
- Existing http/facilitator.go client implementation must be available for payment verification and settlement
- Facilitator service must be running and accessible at configured URL (defaults to Coinbase facilitator)
- USDC smart contracts must be deployed on target networks (base, base-sepolia)

## Clarifications

### Session 2025-10-29

- Q: What are the acceptable latency and throughput targets for the Gin middleware? → A: No specific performance targets - prioritize correctness over speed
- Q: What should the middleware do when the facilitator service is unavailable? → A: Return HTTP 503 Service Unavailable with retry-after header
- Q: What specific data structure and context key should be used for storing payment details? → A: Context key "x402_payment" with VerificationResponse struct (matches stdlib)
- Q: Should CORS preflight OPTIONS requests bypass payment verification? → A: Bypass payment verification for OPTIONS requests
- Q: How should the middleware handle potential replay attacks? → A: Handled by facilitator and signature scheme, not middleware concern

## Notes

This specification is based on the Coinbase x402 reference implementation for Gin but adapted to use our project's existing patterns:
- Uses our facilitator client from http/facilitator.go instead of creating a new one
- Follows our configuration patterns established in http/middleware.go
- Maintains consistency with our stdlib middleware in terms of error handling and logging
- Leverages shared utilities like parsePaymentHeader and sendPaymentRequired where applicable

Key differences from Coinbase reference:
- Our implementation will live in http/gin/ package (Coinbase uses pkg/gin/)
- We will reuse existing facilitator client code rather than embedding it in middleware
- We will follow Go's functional options pattern more strictly
- We will provide better integration with existing x402-go types
