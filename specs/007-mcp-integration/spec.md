# Feature Specification: MCP Integration

**Feature Branch**: `007-mcp-integration`  
**Created**: 2025-10-31  
**Status**: Draft  
**Input**: User description: "I want to take an existing library mcp-go-x402 and bring that functionality into this library.
- The library code is in scratch/mcp.txt
- The functionality should exist in the mcp subpackage and then mcp/client and mcp/server packages
- The new sub package will use OUR existing signers, facilitator, types and helpers code. We only want to use the code from mcp-go-x402 as a reference
- Create an example in examples/mcp it should have the same flow as examples/x402demo with client and server modes use the example in mcp-go-x402 for reference as well
- Add tests
- Do not worry about performance speed requirements
- reference previous specs for guidelines on this spec
- delegate your searching in parallel subagents"

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - MCP Client with x402 Payments (Priority: P1)

Developers need to create MCP (Model Context Protocol) clients that can automatically handle x402 payment requirements when interacting with paid MCP servers, enabling seamless access to premium AI tools and services.

**Why this priority**: Core functionality needed for any MCP client to interact with x402-protected servers; enables the primary use case of paying for premium AI services.

**Independent Test**: Can be fully tested by creating a client that connects to an x402 MCP server, receives payment requirements, signs payments using existing signers, and successfully accesses paid tools.

**Acceptance Scenarios**:

1. **Given** an MCP client with configured payment signers, **When** calling a paid tool on an x402 server, **Then** the client automatically handles the 402 payment flow and receives the tool response
2. **Given** an MCP client with multiple payment signers, **When** the primary signer fails, **Then** the client falls back to secondary signers in priority order
3. **Given** an MCP client accessing a free tool, **When** making the request, **Then** no payment is required and the response is received immediately
4. **Given** an MCP client making concurrent requests to the same paid tool, **When** both requests are sent simultaneously, **Then** each request includes and validates its own separate payment

---

### User Story 2 - MCP Server with x402 Protection (Priority: P1)

Developers need to create MCP servers that can protect their tools with x402 payment requirements, allowing them to monetize AI services while supporting both free and paid tool offerings.

**Why this priority**: Essential for service providers to monetize their MCP tools; enables the business model for premium AI services.

**Independent Test**: Can be tested by creating a server with both free and paid tools, verifying payment requirements are sent for paid tools, and validating payments through the facilitator.

**Acceptance Scenarios**:

1. **Given** an MCP server with a paid tool, **When** a client calls the tool without payment, **Then** the server returns a 402 error with payment requirements
2. **Given** an MCP server receiving a valid payment, **When** verifying with the facilitator, **Then** the tool executes and returns results
3. **Given** an MCP server with mixed tools, **When** clients access free tools, **Then** no payment is required

---

### User Story 3 - Multi-Chain Payment Support (Priority: P2)

Users need to make payments using different blockchain networks and assets based on their preferences and available balances, with automatic selection of the best payment option.

**Why this priority**: Provides flexibility for users with assets on different chains; reduces payment friction by supporting user preferences.

**Independent Test**: Can be tested by configuring multiple payment options across different networks and verifying the client selects the optimal option based on priority and availability.

**Acceptance Scenarios**:

1. **Given** a client with EVM and Solana signers, **When** the server accepts both, **Then** the client selects based on configured priority using the DefaultPaymentSelector algorithm (signer priority, then token priority, then configuration order)
2. **Given** a client with insufficient balance on the primary network, **When** attempting payment, **Then** it falls back to secondary payment options following the same priority rules
3. **Given** a server accepting multiple networks, **When** receiving payments, **Then** all supported networks are processed correctly

---

### User Story 4 - Example Implementation (Priority: P3)

Developers need working examples demonstrating both client and server implementations with x402 payment flows, similar to existing x402demo examples.

**Why this priority**: Accelerates developer adoption by providing reference implementations; reduces integration time.

**Independent Test**: Can be tested by running the example client against the example server, verifying successful payment flows for both EVM and Solana networks.

**Acceptance Scenarios**:

1. **Given** the example MCP implementation, **When** running in server mode, **Then** it serves both free and paid tools with proper x402 protection
2. **Given** the example MCP implementation, **When** running in client mode, **Then** it connects to the server and successfully accesses both tool types

---

### Edge Cases

- What happens when all configured signers fail to create a valid payment?
- How does the system handle network timeouts during payment verification?
- What occurs when a payment is valid but the facilitator is temporarily unavailable?
- How does the client handle malformed payment requirements from a server?
- What happens when a server's payment requirements exceed all client-configured limits?
- When a tool execution fails after payment verification succeeds, the payment is non-refundable (payment covers the execution attempt, not success)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide an MCP client transport that integrates x402 payment handling
- **FR-002**: System MUST provide an MCP server implementation that can protect tools with x402 requirements
- **FR-003**: Client MUST automatically detect 402 payment requirements and handle payment flow transparently
- **FR-004**: Server MUST support both free and paid tools in the same instance with per-tool payment configuration in code
- **FR-005**: Client MUST support fallback across multiple payment signers when primary options fail
- **FR-006**: Server MUST verify payments through the facilitator before executing paid tools
- **FR-015**: Server MUST NOT refund payments when tool execution fails after successful payment verification
- **FR-007**: System MUST reuse existing signer implementations without duplication
- **FR-008**: System MUST reuse existing facilitator client for payment verification
- **FR-009**: Client MUST support both HTTP header and JSON-RPC parameter payment transports
- **FR-010**: Server MUST extract payments from both X-PAYMENT headers and request parameters
- **FR-011**: System MUST provide comprehensive examples demonstrating client-server interaction
- **FR-012**: Examples MUST support both command-line client and server modes
- **FR-013**: System MUST include unit tests for core payment handling logic
- **FR-014**: System MUST support concurrent payment attempts without race conditions
- **FR-016**: Each concurrent request to a paid tool MUST include its own independent payment
- **FR-017**: Payment verification with facilitator MUST timeout after 5 seconds
- **FR-018**: Payment settlement with facilitator MUST timeout after 60 seconds

### Key Entities *(include if feature involves data)*

- **MCP Transport**: Client-side component handling x402 payment flows during MCP communication
- **MCP Server**: Server-side component protecting tools with x402 requirements and verifying payments
- **Payment Handler**: Orchestrates payment creation across multiple signers with fallback logic
- **Tool Configuration**: Associates MCP tools with optional payment requirements configured per-tool in server initialization code

## Clarifications

### Session 2025-10-31
- Q: How should MCP client select payment method when multiple options exist? → A: Use DefaultPaymentSelector priority algorithm
- Q: What should happen when an MCP server tool fails after successfully verifying payment? → A: Keep payment (tool execution was attempted)
- Q: How should the MCP integration handle concurrent requests to the same paid tool from a single client? → A: Each request requires separate payment
- Q: What timeout should be enforced for payment verification with the facilitator? → A: 5 seconds
- Q: When an MCP server exposes both free and paid tools, how should tool payment requirements be configured? → A: Per-tool configuration in code

## Success Criteria *(mandatory)*

<!--
  ACTION REQUIRED: Define measurable success criteria.
  These must be technology-agnostic and CONCRETELY measurable.
  
  FORBIDDEN - Do NOT include vague or unmeasurable criteria like:
  - "X% of users successfully..." (unless you have a way to measure this)
  - "Reduce/improve by X%" (percentage improvements are vague)
  - "Faster/better/easier" (subjective without concrete metrics)
  - "Developer satisfaction" or similar subjective measures
  
  REQUIRED - Only include criteria that are:
  - Directly observable and countable (lines of code, number of steps, time duration)
  - Verifiable through testing (can handle N concurrent requests, completes in X seconds)
  - Binary pass/fail (100% coverage, zero errors, all chains supported)
  - Specific numeric thresholds (under 10 lines, within 5 seconds, supports 8 chains)
-->

### Measurable Outcomes

- **SC-001**: Developers can protect MCP tools with x402 requirements in under 10 lines of per-tool configuration code
- **SC-002**: Client automatically handles payment flow with zero additional code beyond signer configuration
- **SC-003**: Payment fallback completes within 5 seconds when primary signer fails
- **SC-004**: Example implementation runs successfully with both EVM and Solana payment options
- **SC-005**: 100% of existing signer types work with the new MCP integration
- **SC-006**: Server processes mixed free/paid tool requests without payment overhead on free tools
- **SC-007**: All payment verification flows complete within 5 seconds (matching existing HTTP middleware timeout)
- **SC-008**: Integration adds fewer than 1000 lines of new code by reusing existing components
