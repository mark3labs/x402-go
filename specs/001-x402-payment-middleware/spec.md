# Feature Specification: x402 Payment Middleware

**Feature Branch**: `001-x402-payment-middleware`  
**Created**: 2025-10-28  
**Status**: Draft  
**Input**: User description: "The first feature of this library is to allow devs to utilize a middleware in their Go stdlib http handlers that paywalls the route(s) using the x402 payment standard. We should be able to accept tokens on any EVM or SVM chain."

## Clarifications

### Session 2025-10-28

- Q: How should the middleware track used nonces for replay attack prevention? → A: Delegate to facilitator service
- Q: How should the middleware handle facilitator service unavailability? → A: Reject with 503 if no fallback, use fallback if configured
- Q: How should the middleware match routes for payment configuration? → A: Wrap specific handlers/mux, no pattern matching needed
- Q: How should the middleware handle blockchain network congestion? → A: Forward errors from facilitator service
- Q: How should the middleware handle concurrent request limits? → A: Out of scope, only payment gating

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic Middleware Integration (Priority: P1)

A developer wants to protect their HTTP endpoints with payment requirements using the x402 standard. They integrate the middleware into their existing Go HTTP server and specify payment requirements for protected routes.

**Why this priority**: This is the core functionality that enables developers to monetize their APIs with minimal code changes. Without this, the middleware has no value.

**Independent Test**: Can be fully tested by setting up a simple HTTP server with the middleware, making requests without payment (should return 402), and with valid payment (should return success).

**Acceptance Scenarios**:

1. **Given** a Go HTTP server with the x402 middleware protecting an endpoint, **When** a client makes a request without payment headers, **Then** the server returns a 402 Payment Required response with payment requirements in JSON format
2. **Given** a protected endpoint with EVM payment requirements, **When** a client sends a valid EIP-3009 payment authorization in the X-PAYMENT header, **Then** the server verifies the payment, settles it on-chain, and returns the protected resource with X-PAYMENT-RESPONSE header
3. **Given** a protected endpoint with SVM payment requirements, **When** a client sends a valid Solana transaction in the X-PAYMENT header, **Then** the server verifies the payment, settles it on-chain, and returns the protected resource

---

### User Story 2 - Multi-Chain Payment Configuration (Priority: P2)

A developer wants to accept payments on multiple blockchain networks (both EVM and SVM chains) for the same resource. They configure the middleware to accept different payment options and let clients choose their preferred payment method.

**Why this priority**: Flexibility in payment options increases the potential user base and allows developers to support users on different blockchain ecosystems.

**Independent Test**: Configure middleware with multiple payment options (e.g., USDC on Base and Solana), verify that payment requirements list all options, and that payments work on each chain independently.

**Acceptance Scenarios**:

1. **Given** an endpoint configured to accept payments on both Base and Solana, **When** a client requests the resource without payment, **Then** the payment requirements response includes both EVM and SVM payment options in the accepts array
2. **Given** multiple payment options available, **When** a client pays with USDC on Base (EVM), **Then** the payment is verified and settled on the Base network
3. **Given** multiple payment options available, **When** a client pays with USDC on Solana (SVM), **Then** the payment is verified and settled on the Solana network

---

### User Story 3 - Custom Payment Requirements per Route (Priority: P3)

A developer wants to set different payment amounts and configurations for different routes. They wrap individual handlers or route groups with middleware instances configured with specific payment requirements, allowing premium endpoints to cost more than basic ones.

**Why this priority**: Enables flexible pricing strategies where different resources can have different values, supporting tiered access models.

**Independent Test**: Set up multiple handlers wrapped with different payment middleware configurations, verify each handler returns its specific requirements, and that payments are validated against the correct amounts.

**Acceptance Scenarios**:

1. **Given** two endpoints with different payment amounts configured, **When** clients request each endpoint, **Then** each returns its specific payment requirements with the correct amount
2. **Given** a premium endpoint requiring 10000 units and a basic endpoint requiring 1000 units, **When** a client pays 1000 units to the premium endpoint, **Then** the payment is rejected as insufficient
3. **Given** route-specific payment configurations, **When** a valid payment is made to any route, **Then** the payment amount is validated against that specific route's requirements

---

### User Story 4 - Payment Verification without Settlement (Priority: P4)

A developer wants to verify payment authorizations before settling them on-chain, allowing for pre-flight checks or custom business logic before committing the transaction.

**Why this priority**: Provides flexibility for developers to implement custom validation logic, rate limiting, or other checks before incurring blockchain transaction costs.

**Independent Test**: Configure middleware to use verification-only mode, send valid and invalid payment authorizations, verify that only validation occurs without on-chain settlement.

**Acceptance Scenarios**:

1. **Given** middleware configured for verification-only mode, **When** a valid payment authorization is received, **Then** the payment is verified but not settled on-chain
2. **Given** verification-only mode, **When** an invalid payment authorization is received, **Then** the middleware returns an error without attempting settlement
3. **Given** a verified but not settled payment, **When** the developer's custom logic approves it, **Then** the settlement can be triggered programmatically

---

### Edge Cases

- What happens when the facilitator service is unavailable? → Returns 503 Service Unavailable unless fallback facilitator is configured
- How does the system handle expired payment authorizations?
- What occurs when a payment is valid but the blockchain network is congested? → Facilitator handles blockchain concerns; middleware forwards facilitator errors
- How does the middleware handle malformed X-PAYMENT headers?
- What happens if the same payment nonce is used twice?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a Go middleware function compatible with standard net/http handlers
- **FR-002**: System MUST support payment requirements for both EVM chains (using EIP-3009) and SVM chains (using SPL token transfers)
- **FR-003**: Middleware MUST return HTTP 402 Payment Required status with proper payment requirements when no valid payment is provided
- **FR-004**: System MUST validate payment authorizations by verifying signatures, checking balances, and confirming amounts
- **FR-005**: System MUST support configuration of payment requirements including asset type, amount, recipient address, and network
- **FR-006**: Middleware MUST process X-PAYMENT headers containing base64-encoded payment payloads
- **FR-007**: System MUST return X-PAYMENT-RESPONSE headers with settlement information after successful payment
- **FR-008**: Middleware MUST support integration with facilitator services for payment verification and settlement, with optional fallback facilitator configuration; all blockchain operations are delegated to facilitator
- **FR-009**: System MUST prevent replay attacks by delegating nonce validation to the facilitator service
- **FR-010**: Middleware MUST support configuration of multiple payment options for the same resource
- **FR-011**: System MUST properly handle and return standardized x402 error codes
- **FR-012**: Middleware MUST allow handler-specific payment requirement configuration by wrapping individual handlers or muxes

### Key Entities *(include if feature involves data)*

- **PaymentRequirements**: Defines acceptable payment methods including scheme, network, amount, asset, recipient, and timeout
- **PaymentPayload**: Contains the payment authorization data including signature and transaction details
- **SettlementResponse**: Records the result of payment settlement including transaction hash and success status
- **MiddlewareConfig**: Stores middleware configuration including primary and optional fallback facilitator endpoints, payment requirements per route, and verification settings
- **PaymentSession**: Tracks payment state including verification status, settlement attempts, and nonce usage

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can integrate the middleware into an existing Go HTTP server in under 5 minutes
- **SC-002**: Payment verification completes in under 2 seconds for 95% of requests
- **SC-003**: The middleware successfully processes payments on at least 3 different EVM networks and 2 SVM networks
- **SC-004**: 99.9% of valid payment authorizations are successfully verified and settled
- **SC-005**: Zero duplicate payments are processed (100% replay attack prevention)
- **SC-006**: Middleware adds less than 50ms latency to requests that don't require payment
- **SC-007**: Documentation enables 90% of developers to successfully implement their first paid endpoint without external support
- **SC-008**: The middleware handles at least 100 concurrent payment requests without degradation (rate limiting is out of scope)
- **SC-009**: Invalid payment attempts are rejected within 500ms with clear error messages
- **SC-010**: Multi-chain payment configuration reduces failed payment attempts by 30% compared to single-chain setup

## Assumptions

- Developers have access to facilitator services (either third-party or self-hosted) for payment verification and settlement
- Payment will use USDC or other EIP-3009 compliant tokens on EVM chains
- SPL tokens are used for payments on Solana/SVM chains
- Developers understand basic blockchain concepts and have wallet addresses for receiving payments
- The middleware will integrate with existing x402 facilitator API specifications
- Network connectivity to blockchain nodes is available through facilitator services