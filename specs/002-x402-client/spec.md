# Feature Specification: x402 Payment Client

**Feature Branch**: `002-x402-client`  
**Created**: 2025-10-28  
**Status**: Draft  
**Input**: User description: "I want to be able to create a stdlib compatible http client that can sign and send payment responses to an x402 protected endpoint in order to access paywalled data. I should be able to instantiate either an EVMSigner or an SVMSigner or both. I should be able to specify which tokens I am will to pay for each signer. I should be able to set signer level max amount to pay per call (optional). I should be able to set signer level budget (optional). http client can attach multiple signers. signers can support multiple tokens for payment. I should be able to set priority levels for signers as well as tokens (optional). http client will intelligently decide how to pay based on my settings and the required/supported payments schemes provided by the server"

## Clarifications

### Session 2025-10-28

- Q: When multiple signers have the same priority level configured, how should the system resolve the tie? → A: Configuration order (first configured wins)
- Q: When a payment fails due to insufficient funds or rejected authorization, should the client automatically attempt with the next available signer/token? → A: Automatically try next signer/token
- Q: When a payment authorization is about to expire (approaching validBefore timestamp), should the client proactively regenerate it? → A: Never regenerate (let it fail)

### Additional Clarifications

- Signers can be instantiated from private keys, mnemonic phrases, or keystore files

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic Payment for Protected Resource (Priority: P1)

A developer wants to access a paywalled data endpoint using their payment credentials. They configure a client with a single payment signer (using their private key, mnemonic, or keystore file) and successfully retrieve protected content by automatically paying the required amount.

**Why this priority**: This is the core functionality that delivers immediate value - enabling programmatic access to paywalled resources. Without this, no other features matter.

**Independent Test**: Can be fully tested by configuring a client with one signer, making a request to a paywalled endpoint, and verifying successful payment and data retrieval.

**Acceptance Scenarios**:

1. **Given** a client configured with an EVM signer and USDC token settings, **When** accessing a resource that requires USDC payment on the same network, **Then** the client automatically creates, signs, and sends the payment to access the resource
2. **Given** a client configured with an SVM signer and SOL token settings, **When** accessing a resource that requires SOL payment on Solana, **Then** the client automatically creates, signs, and sends the payment to access the resource
3. **Given** a client with a configured signer, **When** the resource requires payment in an unsupported token or network, **Then** the client returns an appropriate error indicating no suitable payment method
4. **Given** signers instantiated from different credential sources (private key, mnemonic, or keystore), **When** making payments, **Then** all signers function identically regardless of credential source

---

### User Story 2 - Multi-Signer Payment Selection (Priority: P2)

A developer configures multiple payment signers (both EVM and SVM) with different tokens. The client intelligently selects the most appropriate signer based on the server's payment requirements and configured priorities.

**Why this priority**: Enables flexibility and cross-chain compatibility, allowing users to work with diverse payment requirements without manual intervention.

**Independent Test**: Can be tested by configuring multiple signers with different tokens/networks and verifying the client selects the appropriate one based on server requirements.

**Acceptance Scenarios**:

1. **Given** a client with both EVM and SVM signers configured, **When** the server accepts only EVM payments, **Then** the client uses the EVM signer automatically
2. **Given** multiple signers with overlapping supported tokens, **When** both could satisfy the payment requirement, **Then** the client selects based on configured priority levels (with ties resolved by configuration order)
3. **Given** a client with multiple signers each supporting different tokens, **When** accessing a resource that accepts multiple payment methods, **Then** the client selects the signer/token combination with the highest priority that meets the requirements

---

### User Story 3 - Payment Amount Controls (Priority: P3)

A developer sets maximum payment limits at the signer level to control costs per transaction. The client respects these limits when making payment decisions.

**Why this priority**: Provides transaction-level cost control and prevents unexpected large charges, important for production use but not essential for basic functionality.

**Independent Test**: Can be tested by setting max amounts on signers, then verifying payments are rejected when limits would be exceeded.

**Acceptance Scenarios**:

1. **Given** a signer with a max amount per call of 100 tokens, **When** a resource requires 150 tokens, **Then** the payment is rejected with an appropriate error
2. **Given** multiple signers where only one has sufficient per-call limits, **When** making a payment, **Then** the client selects the signer that can satisfy the payment within its limits

---

### User Story 4 - Token Priority Configuration (Priority: P4)

A developer configures priority levels for different tokens within each signer. The client uses these priorities to select the optimal token when multiple options are available.

**Why this priority**: Enables fine-grained control over payment preferences, useful for optimizing costs or preferring certain tokens over others.

**Independent Test**: Can be tested by configuring token priorities and verifying the client selects higher priority tokens when multiple options satisfy requirements.

**Acceptance Scenarios**:

1. **Given** a signer configured with USDC (priority 1) and USDT (priority 2), **When** the server accepts both tokens, **Then** the client selects USDC for payment
2. **Given** token priorities are set, **When** the highest priority token cannot satisfy the payment (insufficient balance or amount limits), **Then** the client automatically falls back to the next priority token

---

### Edge Cases

- What happens when all configured signers lack sufficient funds for the required payment? (System tries all available options then returns comprehensive error)
- How does the system handle network errors during payment submission?
- What happens when a payment authorization expires before settlement? (Request fails with expiration error, user must retry)
- How does the client handle conflicting payment requirements (e.g., server requires exact amount but has multiple acceptable amounts)?
- How does the system handle concurrent requests with max amount limits?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a standard-library-compatible HTTP client interface for making requests to x402-protected endpoints
- **FR-002**: System MUST support configuration of multiple payment signers (EVM and/or SVM) on a single client instance, instantiated from private keys, mnemonic phrases, or keystore files
- **FR-003**: Each signer MUST support configuration of multiple tokens it is willing to pay with
- **FR-004**: System MUST automatically parse payment requirements from server responses (402 status with payment details)
- **FR-005**: System MUST automatically select an appropriate signer and token based on server requirements and client configuration
- **FR-006**: System MUST generate properly formatted payment authorizations according to the x402 protocol specification, without proactive regeneration for expiring authorizations
- **FR-007**: System MUST include signed payment data in the appropriate request header when retrying after receiving payment requirements
- **FR-008**: Each signer MUST support optional configuration of maximum amount per payment call
- **FR-009**: System MUST support optional priority levels for signers (to prefer one over another when multiple can satisfy requirements), with ties resolved by configuration order
- **FR-010**: System MUST support optional priority levels for tokens within each signer
- **FR-011**: System MUST validate payment requirements against signer capabilities before attempting payment
- **FR-012**: System MUST handle payment failures gracefully, automatically attempting the next available signer/token combination before returning error information
- **FR-013**: System MUST parse and make available payment settlement information from successful responses
- **FR-014**: Client MUST maintain compatibility with standard HTTP client operations (non-payment requests should work normally)
- **FR-015**: System MUST support concurrent requests while maintaining thread safety
- **FR-016**: System MUST validate that payment amounts don't exceed configured max amount limits before signing
- **FR-017**: Signers MUST support instantiation from multiple credential formats: raw private keys, BIP39 mnemonic phrases, and encrypted keystore files

### Key Entities *(include if feature involves data)*

- **Payment Client**: HTTP client wrapper that manages payment signers and handles x402 payment flows
- **Payment Signer**: Entity that can sign payment authorizations for a specific blockchain network (EVM or SVM), instantiated from private keys, mnemonic phrases, or keystore files
- **Token Configuration**: Settings for a specific token that a signer is willing to pay with, including priority and limits
- **Payment Requirements**: Server-specified payment terms including accepted tokens, amounts, and recipient addresses
- **Payment Authorization**: Signed payment data that authorizes a specific token transfer

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can successfully access paywalled resources with automatic payment in under 3 seconds for standard requests
- **SC-002**: System correctly selects appropriate signer/token combination in 100% of cases where valid configuration exists
- **SC-003**: Max amount limits are enforced with 100% accuracy, never allowing payments that exceed configured per-call limits
- **SC-004**: 95% of developers can configure and use basic payment functionality within 10 minutes of first use
- **SC-005**: System handles 100 concurrent payment requests without errors or race conditions
- **SC-006**: Payment selection with 10 configured signers completes in under 100 milliseconds
- **SC-007**: Error messages clearly identify the reason for payment failure in 100% of cases
- **SC-008**: System maintains full compatibility with standard HTTP client operations, with zero impact on non-payment requests
- **SC-009**: Memory overhead per configured signer remains under 1MB during normal operation
- **SC-010**: Developer can configure multi-signer, multi-token setup with priorities in under 5 minutes using clear documentation

## Assumptions

- Payment authorization signing mechanisms (EVM and SVM) follow standard blockchain practices
- Developers have access to private keys, mnemonic phrases, or keystore files for their payment accounts  
- Server payment requirements follow the x402 protocol specification
- Token balances are managed externally; the client does not check on-chain balances
- Network connectivity and blockchain availability are handled with standard timeout and retry patterns
- Payment settlement is handled by the server/facilitator; client only provides signed authorization
- Time-based payment authorization validity uses reasonable defaults (e.g., 60 seconds) without automatic regeneration
- Priority levels use numeric values where lower numbers indicate higher priority (1 = highest)