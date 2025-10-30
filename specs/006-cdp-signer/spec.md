# Feature Specification: Coinbase CDP Signer Integration

**Feature Branch**: `006-cdp-signer`  
**Created**: Thu Oct 30 2025  
**Status**: Draft  
**Input**: User description: "I want to add a new signer option which allows devs to use a server wallet from Coinbase CDP. Read scratch/cdp.md. The signer will support both EVM and SVM"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - EVM Transaction Signing with CDP Wallet (Priority: P1)

A developer needs to process x402 payments on Ethereum-compatible chains (Base, Ethereum mainnet) without managing private keys locally. They want to use Coinbase's managed wallet infrastructure to sign payment transactions securely.

**Why this priority**: EVM chain support is the primary use case for x402 payments, as most blockchain activity currently occurs on Ethereum and Layer 2 networks. This provides immediate value and enables developers to start using CDP signers for production payment processing.

**Independent Test**: Can be fully tested by configuring CDP credentials, initializing a CDP signer for Base Sepolia testnet, and successfully signing an x402 payment transaction. Delivers standalone value by enabling secure EVM transaction signing without any SVM support.

**Acceptance Scenarios**:

1. **Given** a developer has CDP API credentials and wallet secret, **When** they initialize a CDP signer for an EVM network, **Then** the signer successfully authenticates and returns a valid address
2. **Given** an initialized EVM CDP signer, **When** an x402 payment request is received, **Then** the signer signs the transaction and returns a valid signature that can be broadcast to the network
3. **Given** an EVM CDP signer on Base mainnet, **When** a payment transaction is signed, **Then** the transaction broadcasts successfully and is confirmed on-chain

---

### User Story 2 - Solana Transaction Signing with CDP Wallet (Priority: P2)

A developer wants to accept x402 payments on Solana using CDP-managed wallets. They need the same secure signing capabilities for SVM chains that they have for EVM chains.

**Why this priority**: Solana support extends the CDP signer to all x402-supported chains, providing feature parity with existing local signers. However, it's lower priority than EVM since most current x402 usage is on EVM chains.

**Independent Test**: Can be tested independently by configuring CDP for Solana devnet, initializing an SVM signer, and signing a test payment transaction. Delivers value even if EVM support doesn't exist by enabling Solana-only deployments.

**Acceptance Scenarios**:

1. **Given** a developer has CDP credentials, **When** they initialize a CDP signer for Solana, **Then** the signer returns a valid Solana address and can query account balance
2. **Given** an initialized SVM CDP signer, **When** a payment transaction needs signing, **Then** the signer produces a valid Solana transaction signature
3. **Given** a Solana payment request, **When** the CDP signer signs and broadcasts the transaction, **Then** the transaction confirms on-chain

---

### User Story 3 - Secure Credential Management (Priority: P1)

A developer needs to configure CDP credentials securely without hardcoding secrets in their application. They want to use environment variables or secrets management systems to protect their API keys and wallet secrets.

**Why this priority**: Security is critical for production deployments. Without proper credential management, developers risk exposing their CDP keys, which could lead to unauthorized access and financial loss.

**Independent Test**: Can be tested by configuring credentials via environment variables, initializing a signer, and verifying that credentials are never logged or exposed. Delivers immediate value by preventing credential leakage regardless of which chains are supported.

**Acceptance Scenarios**:

1. **Given** CDP credentials are set in environment variables, **When** a developer initializes a CDP signer, **Then** credentials are loaded securely without being exposed in logs or error messages
2. **Given** missing or invalid credentials, **When** initialization is attempted, **Then** a clear error message is returned without revealing partial credential information
3. **Given** CDP credentials with insufficient permissions, **When** a signing operation is attempted, **Then** the system returns a specific permission error without exposing the full credential
4. **Given** multiple environment configurations (dev, staging, prod), **When** credentials are rotated, **Then** each environment can use different CDP credentials without code changes

---

### User Story 4 - Error Handling and Retry Logic (Priority: P2)

A developer's application needs to handle CDP API errors gracefully, including rate limits, network timeouts, and temporary service unavailability. They want automatic retry logic for transient failures without manual intervention.

**Why this priority**: Production systems require robust error handling to maintain reliability. While not as critical as core signing functionality, proper error handling prevents payment failures and improves user experience.

**Independent Test**: Can be tested by simulating various CDP API failure scenarios (rate limits, timeouts, 5xx errors) and verifying that the signer retries appropriately. Delivers value by improving reliability regardless of which specific signing operations are being performed.

**Acceptance Scenarios**:

1. **Given** a CDP API rate limit is exceeded, **When** a signing request is made, **Then** the signer retries with backoff until successful or maximum attempts reached
2. **Given** a network timeout occurs during signing, **When** the operation is retried, **Then** the signer retries the signing operation (duplicate prevention handled by x402 nonce mechanism)
3. **Given** CDP returns a 5xx server error, **When** the signing operation is attempted, **Then** the signer retries up to 3 times with backoff before returning an error
4. **Given** CDP returns a 4xx client error (invalid request), **When** the error is encountered, **Then** the signer immediately returns the error without retrying
5. **Given** all retry attempts fail, **When** the final attempt fails, **Then** a detailed error message is returned indicating the failure reason and retry count

---

### User Story 5 - Account Creation and Retrieval (Priority: P1)

A developer initializing a CDP signer for the first time needs to create a new CDP wallet account if one doesn't exist, or retrieve an existing account if it was previously created. They want a simple helper that automatically handles both scenarios without manual intervention.

**Why this priority**: CDP accounts cannot be created through the portal and must be created programmatically via the API. Without this capability, developers cannot use CDP signers at all. This is a critical prerequisite for all signing operations.

**Independent Test**: Can be tested by calling the account creation helper with CDP credentials on a fresh account (creates new), then calling again with same credentials (retrieves existing), verifying both scenarios work correctly. Delivers immediate value by enabling first-time setup.

**Acceptance Scenarios**:

1. **Given** valid CDP credentials and no existing account, **When** a developer initializes a CDP signer for an EVM network, **Then** a new account is created automatically and the address is returned
2. **Given** valid CDP credentials with an existing account, **When** a developer initializes a CDP signer with the same network, **Then** the existing account address is retrieved and returned without creating duplicates
3. **Given** CDP credentials and a specific network ID, **When** account creation is attempted, **Then** the account is created for that specific network (e.g., base-sepolia, solana-devnet)
4. **Given** an account creation request fails due to network error, **When** the operation is retried, **Then** the system handles retry appropriately without creating duplicate accounts
5. **Given** multiple concurrent initialization attempts, **When** both try to create accounts, **Then** the system prevents race conditions and ensures only one account is created

---

### User Story 6 - Multi-Chain Support Within Same Application (Priority: P3)

A developer wants to accept x402 payments on both EVM and SVM chains simultaneously using CDP signers. They need a consistent interface that works across both chain types without managing separate credential sets or configurations.

**Why this priority**: This enhances flexibility but is lower priority since most applications initially focus on one chain type. It's a nice-to-have that improves developer experience but isn't essential for initial deployments.

**Independent Test**: Can be tested by initializing both EVM and SVM CDP signers with the same credentials, sending concurrent payment requests to both, and verifying both process successfully. Delivers value by enabling true multi-chain applications with unified credential management.

**Acceptance Scenarios**:

1. **Given** a single set of CDP credentials, **When** a developer initializes signers for both Base and Solana, **Then** both signers authenticate successfully and can sign transactions
2. **Given** a multi-chain configuration, **When** credentials are rotated, **Then** both EVM and SVM signers update without requiring separate rotation procedures

---

### Edge Cases

- What happens when CDP API credentials expire mid-operation?
- How does the system handle CDP service downtime lasting longer than retry window?
- What happens if a transaction is signed successfully but broadcasting fails?
- How does the signer handle rate limiting when processing multiple concurrent payments?
- What happens when a CDP wallet has insufficient funds for gas/transaction fees?
- How does the system handle signature requests for unsupported networks?
- What happens when JWT token generation fails due to invalid private key format?
- How does the signer handle network partition during transaction signing?
- What happens when two signing requests arrive for the same nonce/sequence number?
- How does the system handle CDP API version changes or deprecations?
- What happens when multiple processes try to create accounts with the same credentials simultaneously?
- How does account retrieval behave when multiple accounts exist for the same credentials and network? (Handled by CDP API)
- What happens if account creation succeeds but the response is lost due to network failure?
- How does the system handle account creation when the network ID is invalid or unsupported?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a helper to create or retrieve CDP wallet accounts for a given network
- **FR-002**: System MUST create new CDP accounts when none exist for the specified credentials and network
- **FR-003**: System MUST retrieve existing CDP accounts when they already exist for the specified credentials and network
- **FR-004**: System MUST support account creation for EVM-compatible chains (Ethereum, Base, Polygon, etc.)
- **FR-005**: System MUST support account creation for Solana (SVM)
- **FR-006**: System MUST prevent duplicate account creation by querying existing accounts (GET) before attempting creation (POST)
- **FR-007**: System MUST support signing x402 payment transactions using Coinbase CDP server wallets
- **FR-008**: System MUST load CDP credentials securely from environment variables (CDP_API_KEY_NAME, CDP_API_KEY_SECRET, CDP_WALLET_SECRET)
- **FR-009**: System MUST generate fresh JWT Bearer tokens for each CDP API request (2-minute expiration, no caching)
- **FR-010**: System MUST generate fresh Wallet Authentication JWT for each transaction signing operation (1-minute expiration, no caching)
- **FR-011**: System MUST expose the same signer interface as existing EVM and SVM signers for compatibility with x402 middleware
- **FR-012**: System MUST handle CDP API errors (4xx, 5xx) with appropriate classification (retryable vs non-retryable)
- **FR-013**: System MUST implement exponential backoff retry logic for transient failures (rate limits, timeouts, 5xx errors)
- **FR-014**: System MUST prevent credential leakage in logs, error messages, and debug output
- **FR-015**: System MUST sanitize authorization headers and sensitive data before logging
- **FR-016**: System MUST handle CDP API rate limit responses (429 status) with exponential backoff retry logic (no client-side preventive rate limiting)
- **FR-017**: System MUST validate CDP credentials during signer initialization
- **FR-018**: System MUST return descriptive error messages for credential validation failures
- **FR-019**: System MUST support configurable network selection for both EVM and SVM signers
- **FR-020**: System MUST handle both mainnet and testnet networks for all supported chains
- **FR-021**: System MUST allow developers to retrieve the wallet address from an initialized CDP signer
- **FR-022**: System MUST maintain request context (30-second timeout) through retry attempts
- **FR-024**: System MUST comply with CDP's JWT claim requirements (sub, iss, nbf, exp, uri, reqHash)

### Key Entities

- **CDP Account**: Represents a blockchain wallet account managed by Coinbase CDP. Contains account address (EVM hex address or Solana base58 public key) and network identifier. Created programmatically via CDP API since portal creation is not supported.

- **CDP Signer (EVM)**: Represents a Coinbase CDP-backed signing service for Ethereum-compatible chains. Contains authentication credentials, wallet address, chain ID, and HTTP client for CDP API communication. Signs transaction payloads and returns signatures compatible with x402 payment protocol.

- **CDP Signer (SVM)**: Represents a Coinbase CDP-backed signing service for Solana. Contains authentication credentials, wallet address, network identifier, and HTTP client for CDP API communication. Signs Solana transaction messages and returns signatures compatible with x402 payment protocol.

- **CDP Authentication**: Manages JWT token generation using API key credentials. Includes API key name, private key secret, and helper methods for generating Bearer tokens and Wallet Authentication tokens. Generates fresh tokens for each request (no caching or refresh logic).

- **CDP API Client**: HTTP client wrapper for Coinbase CDP REST API. Manages request construction, header injection (Authorization, X-Wallet-Auth), response parsing, and error handling. Implements rate limiting, retry logic, and timeout management.

- **Signing Request**: Contains transaction data that needs signing - transaction payload for EVM (to, value, data, gas parameters) or message for SVM. Includes network context (chain ID, network identifier) and request metadata (timeout, idempotency requirements).

- **Signing Response**: Contains the cryptographic signature produced by CDP wallet. Includes signature bytes, transaction hash (if applicable), and metadata about the signing operation (timestamp, CDP request ID for tracking).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: System successfully creates accounts on all supported EVM networks (minimum 4: Ethereum mainnet, Base, Polygon, Arbitrum)
- **SC-002**: System successfully creates accounts on both Solana mainnet and devnet
- **SC-003**: Account retrieval returns existing accounts without creating duplicates
- **SC-004**: System successfully signs transactions on all supported EVM networks after account creation
- **SC-005**: System successfully signs transactions on Solana networks after account creation
- **SC-006**: Zero credential leakage - all security scans and log reviews show no API keys, secrets, or tokens in output
- **SC-007**: Retry logic successfully recovers from transient failures (rate limits, timeouts, temporary service errors)
- **SC-008**: Error messages never contain sensitive data - validation confirms no tokens, keys, or secrets in error output
- **SC-009**: CDP signer interface matches existing signer interfaces - requires zero code changes to swap implementations
- **SC-010**: System correctly classifies CDP API errors as retryable or non-retryable
- **SC-011**: JWT token generation succeeds for valid credentials and fails with clear errors for invalid credentials
- **SC-012**: Configuration requires exactly 3 environment variables - no more, no less (CDP_API_KEY_NAME, CDP_API_KEY_SECRET, CDP_WALLET_SECRET)
- **SC-013**: System handles CDP rate limit (429) responses and retries until quota available

## Scope

### In Scope

- Helper function to create or retrieve CDP accounts (CreateOrGetAccount)
- CDP account creation for EVM chains via CDP API
- CDP account creation for Solana via CDP API
- Account retrieval to prevent duplicate creation
- CDP signer implementation for EVM chains
- CDP signer implementation for Solana
- Secure credential loading from environment variables
- HTTP client for CDP REST API communication
- Error handling and classification (retryable vs non-retryable)
- Exponential backoff retry logic with jitter (100ms initial delay, 2x multiplier, 10s max total time)
- Request timeout and context propagation
- Rate limit error handling (429 responses trigger exponential backoff)
- Log sanitization to prevent credential leakage
- Network selection and validation
- Interface compatibility with existing x402 signers
- Idempotent account creation to handle race conditions
- Unit tests for account creation, retrieval, signer initialization, signing operations, error handling, and retry logic
- Integration tests against CDP testnet environments

### Out of Scope

- Balance checking or transaction history queries (signer only handles signing and account setup)
- Gas price estimation or transaction broadcasting (handled by x402 client/middleware)
- Support for CDP's official Go SDK (using direct REST API instead due to SDK being pre-alpha)
- Faucet integration for testnet funding
- Secrets management systems integration (HashiCorp Vault, AWS Secrets Manager) - developers handle this separately
- Key rotation automation - manual rotation through CDP Portal
- Multi-sig or threshold signing - single CDP wallet per signer
- Transaction simulation or dry-run capabilities
- Custom JWT claim extensions beyond CDP requirements
- CDP webhook handling for transaction status updates
- Support for CDP's client-side API keys (server-side Secret API Keys only)
- Integration with other wallet providers (MetaMask, WalletConnect, etc.)
- Account listing or enumeration across all networks
- Account deletion or deactivation

## Assumptions

1. Developers have already created CDP projects and generated Secret API Keys through CDP Portal
2. Developers have obtained Wallet Secrets via CDP Portal for their project
3. CDP accounts/wallets are created programmatically via API, not through portal
4. CDP API credentials use Ed25519 or ECDSA signature algorithms (Ed25519 recommended)
5. Developers manage their own secrets storage (environment variables, Vault, etc.)
6. Network connectivity to CDP API endpoints (api.cdp.coinbase.com) is reliable
7. TLS/SSL certificates are valid and trusted on the deployment environment
8. System clock is synchronized (required for JWT token timestamp validation)
9. Developers understand the distinction between testnet and mainnet and configure appropriately
10. CDP rate limits (600 reads/500 writes per 10 seconds) are sufficient for typical x402 payment volumes including account creation; rate limit (429) responses will be handled by retry logic
11. CDP's sub-200ms signing latency claim holds true for production workloads
12. Developers test thoroughly on testnets before deploying to mainnet
13. Go standard library HTTP client is sufficient (no need for third-party HTTP libraries)
14. CDP API is backward compatible within v2 endpoint family
15. Single CDP account per network per application is sufficient
16. Developers handle transaction fee funding for their CDP wallets
17. CDP's 99.9% availability SLA is acceptable for production use cases
18. Account creation is a one-time operation per network (infrequent compared to signing operations)
19. CDP API provides deterministic account creation or allows querying existing accounts

## Dependencies

- Coinbase CDP API v2 (https://api.cdp.coinbase.com)
- CDP API v2 account creation endpoints (/platform/v2/evm/accounts and /platform/v2/solana/accounts)
- CDP Secret API Keys with appropriate permissions for account creation and transaction signing
- CDP Wallet Secrets for wallet authentication
- gopkg.in/square/go-jose.v2 library for JWT generation
- Existing x402-go signer interfaces (defined in signer.go)
- Go standard library: crypto/sha256, crypto/x509, encoding/pem, net/http, encoding/json
- Access to CDP Portal for credential management (not account management)

## Clarifications

### Session 2025-10-30

- Q: What are the specific exponential backoff parameters (initial delay, multiplier, max total time)? → A: Balanced: 100ms initial, 2x multiplier, 10s max total time
- Q: How should rate limiting behave when CDP quotas are approached (reject, queue, sliding window, rely on API 429s)? → A: No active limiting, rely on CDP API 429 responses and exponential backoff
- Q: How does the signer prevent duplicate signatures during retries? → A: Not a concern - blockchain nonce mechanism (x402 payment protocol) prevents double-spending; signer can safely retry signing operations
- Q: How does account retrieval behave when multiple accounts exist for same credentials and network? → A: CDP API handles account management and retrieval logic
- Q: Should JWT tokens (Bearer 2min, Wallet Auth 1min) be cached with expiration tracking or generated fresh per request? → A: Generate fresh JWT for every request (matches CDP examples, simple, no edge cases)

## Open Questions

- None - all requirements are clear based on CDP documentation and existing x402 signer patterns

## Non-Functional Requirements

### Security

- All CDP credentials must be stored in environment variables or secure secrets management
- API keys and secrets must never appear in logs, error messages, or debug output
- All CDP API communication must use HTTPS with valid certificate verification
- JWT tokens must be generated fresh for each request with appropriate expiration (2 minutes for Bearer, 1 minute for Wallet Auth)
- Request body hashes (SHA-256) must be included in JWT tokens for POST/PUT operations
- Authorization headers must be sanitized before logging

### Reliability

- Transient failures must be retried with exponential backoff
- Maximum 5 retry attempts for retryable errors before final failure
- Clear error messages for non-retryable failures (authentication, permissions, invalid requests)
- Request context must be preserved through retry attempts
- System must handle partial failures gracefully (one signer fails, others continue)

### Maintainability

- Code must follow existing x402-go patterns and conventions
- Signer interface must match existing EVM/SVM signer interfaces exactly
- All public types and methods must have clear documentation
- Error types must be descriptive and include context for debugging
- Configuration must be simple (3 environment variables maximum)

### Compatibility

- Must work with Go 1.25.1 and later
- Must integrate seamlessly with existing x402 middleware (net/http, Gin, PocketBase)
- Must support all networks currently supported by x402 for EVM and SVM
- Must maintain backward compatibility with existing signer interface contracts
