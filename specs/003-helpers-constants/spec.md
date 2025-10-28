# Feature Specification: Helper Functions and Constants

**Feature Branch**: `003-helpers-constants`  
**Created**: 2025-10-28  
**Status**: Draft  
**Input**: User description: "Implement helper functions and constants the make creating clients and middleware easier for devs
- Include USDC address constants for the following chains solana, solana-devnet, base, base-sepolia, polygon, polygon-amoy, avalanche, avalanche-fuji
- Include helpers for quickly contructing payment requirements for a given chain"

**Supported Chains** (8 total):
- **Mainnet** (4): Solana, Base, Polygon, Avalanche
- **Testnet** (4): Solana Devnet, Base Sepolia, Polygon Amoy, Avalanche Fuji

*Note: Ethereum and Arbitrum mainnet were researched but excluded from the initial release to focus on the most requested chains. They may be added in future versions.*

## Clarifications

### Session 2025-10-28

- Q: Which USDC token standard should the constants represent for each chain (native Circle USDC, bridged USDC, or both)? → A: Native Circle USDC only (official Circle-deployed contracts)
- Q: What is the authoritative source for validating USDC addresses against official chain deployments? → A: https://developers.circle.com/stablecoins/usdc-contract-addresses
- Q: What format should error messages use when helper functions return errors for invalid parameters? → A: Structured with parameter name and reason (e.g., "amount: must be positive")
- Q: How should developers discover that their library version has outdated USDC addresses after contract upgrades/migrations? → A: Documentation clearly states version when addresses were last verified, no runtime checks
- Q: Should optional parameters for helper customization be passed as function arguments or via configuration struct? → A: Configuration struct with optional fields
- Q: Should EVM PaymentRequirement helpers include EIP-3009 extra fields (name, version) and how should they be provided? → A: Constants should include chain-specific EIP-3009 domain parameters (name varies by chain, e.g., "USD Coin" on Base vs "USDC" on Base Sepolia; version is typically 2); helper automatically populates extra field
- Q: How should the helper handle amounts with precision beyond 6 decimals (e.g., "1.1234567")? → A: Round using standard float64 rounding (banker's rounding)
- Q: Should the payment requirement helper allow zero amounts (e.g., "0" or "0.0")? → A: Allow zero amounts (valid for free-with-signature flows)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quick Client Setup with Chain Constants (Priority: P1)

A developer wants to quickly set up an x402 client to pay for protected resources using USDC on Base. They use provided chain constants and helper functions to configure their client in just a few lines of code, without needing to look up token addresses or network identifiers.

**Why this priority**: This is the core value proposition - reducing friction for developers getting started. Without this, developers must manually look up token addresses and network identifiers, which is time-consuming and error-prone.

**Independent Test**: Can be fully tested by creating a client using chain constants and helper functions, then verifying the client is properly configured with correct token addresses and network identifiers.

**Acceptance Scenarios**:

1. **Given** a developer wants to configure a client for Base USDC, **When** they use the provided Base USDC constant and helper function, **Then** the client is configured with the correct USDC contract address and "base" network identifier
2. **Given** a developer wants to configure a client for Solana USDC, **When** they use the provided Solana USDC constant and helper function, **Then** the client is configured with the correct USDC mint address and "solana" network identifier
3. **Given** a developer wants to test with testnet tokens, **When** they use testnet constants (base-sepolia, solana-devnet, polygon-amoy, avalanche-fuji), **Then** the client is configured with the correct testnet token addresses and network identifiers

---

### User Story 2 - Quick Middleware Payment Requirements Setup (Priority: P1)

A developer wants to configure their middleware to accept USDC payments on multiple chains. They use helper functions to quickly construct payment requirements for each supported chain without manually building the PaymentRequirement structs.

**Why this priority**: Essential for reducing middleware setup complexity. Manual construction of PaymentRequirement structs is verbose and error-prone, especially when supporting multiple chains.

**Independent Test**: Can be tested by using helper functions to create payment requirements, then verifying the middleware correctly accepts payments on all specified chains.

**Acceptance Scenarios**:

1. **Given** a developer wants to accept payments on Base, **When** they use the helper function with Base constants, **Then** a properly formatted PaymentRequirement is created with Base network details, USDC address, their recipient address, and EIP-3009 domain parameters in the extra field
2. **Given** a developer wants to accept payments on multiple chains, **When** they use helper functions for each chain, **Then** multiple PaymentRequirement structs are created, each with correct network-specific configuration
3. **Given** a developer wants to set a custom payment amount, **When** they use the helper function with an amount parameter, **Then** the PaymentRequirement is created with the specified amount in atomic units

---

### User Story 3 - Token Configuration Helper (Priority: P2)

A developer wants to configure their client signer to support multiple tokens across different chains. They use helper functions that create properly formatted TokenConfig structs with correct addresses, decimals, and symbols.

**Why this priority**: Simplifies multi-token configuration but is secondary to basic payment setup. Still important for production use cases.

**Independent Test**: Can be tested by using token configuration helpers, then verifying TokenConfig structs have correct addresses, decimals (6 for USDC), and symbols.

**Acceptance Scenarios**:

1. **Given** a developer wants to configure USDC tokens for multiple chains, **When** they use the token config helper with chain-specific USDC constants, **Then** TokenConfig structs are created with correct addresses, 6 decimals, and "USDC" symbol
2. **Given** a developer wants to set token priorities, **When** they use the helper function with a priority parameter, **Then** the TokenConfig includes the specified priority level
3. **Given** a developer needs mainnet and testnet configurations, **When** they use helpers with both mainnet and testnet constants, **Then** separate TokenConfig structs are created with appropriate network-specific addresses

---

### User Story 4 - Network Identifier Lookup (Priority: P3)

A developer receives a payment requirement from a server and needs to match it against their configured signers. They use helper functions to validate network identifiers and map them to their internal chain representations.

**Why this priority**: Nice-to-have convenience function that aids in payment matching logic but not essential for basic functionality.

**Independent Test**: Can be tested by calling network validation helpers with various network identifiers and verifying correct validation results.

**Acceptance Scenarios**:

1. **Given** a payment requirement with network "base", **When** the developer uses the network validator helper, **Then** the helper confirms it's a valid EVM network identifier
2. **Given** a payment requirement with network "solana", **When** the developer uses the network validator helper, **Then** the helper confirms it's a valid SVM network identifier
3. **Given** a payment requirement with an unknown network, **When** the developer uses the network validator helper, **Then** the helper returns an error indicating unsupported network

---

### Edge Cases

- What happens when a developer uses a testnet constant but points at mainnet infrastructure? → Constants only provide addresses; validation of network consistency is out of scope
- How does the system handle future USDC contract upgrades on supported chains? → Constants represent current addresses; documentation states verification version/date; developers must update to new library version for address changes
- What happens when a developer uses Base constants but the server requires Polygon? → Payment matching logic (existing functionality) will detect mismatch; helpers don't change this behavior
- How do helpers handle custom payment timeouts or extra parameters? → Helpers provide sensible defaults; developers can customize via configuration struct with optional fields or modify returned structs
- What happens when invalid parameters are provided to helpers? → Structured error returned with parameter name and reason (e.g., "recipientAddress: cannot be empty")
- What happens when an amount has precision beyond 6 decimals (e.g., "1.1234567")? → Amount is rounded using standard float64 rounding (banker's rounding / round-to-even) before conversion to atomic units. Example: "1.1234565" rounds to 1123456 (rounds to nearest even), "1.1234575" rounds to 1123458 (rounds to nearest even)
- What happens when a zero amount ("0" or "0.0") is provided? → Zero amounts are allowed; they are valid for free-with-signature authorization flows

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide constants for native Circle USDC token addresses on mainnet chains: Solana, Base, Polygon, Avalanche
- **FR-002**: System MUST provide constants for native Circle USDC token addresses on testnet chains: Solana Devnet, Base Sepolia, Polygon Amoy, Avalanche Fuji
- **FR-003**: Each chain constant MUST include: network identifier string (matching x402 protocol), USDC token address, token decimals (6 for USDC), and for EVM chains: EIP-3009 domain parameters (name and version)
- **FR-004**: System MUST provide a helper function that creates a PaymentRequirement struct from chain constants, payment amount, and recipient address
- **FR-005**: Payment requirement helper MUST accept amount as a human-readable decimal string (e.g., "1.5") and convert to atomic units; amounts with precision beyond 6 decimals MUST be rounded using standard float64 rounding
- **FR-006**: Payment requirement helper MUST set reasonable defaults for: scheme ("exact"), MaxTimeoutSeconds (300), and MimeType ("application/json"), and for EVM chains: automatically populate extra field with chain-specific EIP-3009 domain parameters (name, version)
- **FR-007**: System MUST provide a helper function that creates a TokenConfig struct from chain constants with optional fields via configuration struct
- **FR-008**: System MUST provide a helper function that validates network identifiers and returns network type (EVM or SVM)
- **FR-009**: All helper functions MUST return errors when provided with invalid parameters (negative amounts, empty addresses, etc.) with structured messages including parameter name and reason (e.g., "amount: must be positive"); zero amounts are allowed for free-with-signature flows
- **FR-010**: Constants MUST be exported and accessible from the root x402 package
- **FR-011**: Helper functions MUST be exported and accessible from the root x402 package
- **FR-012**: System MUST provide constants that group related values (network identifier, token address, decimals) for each supported chain

### Key Entities *(include if feature involves data)*

- **ChainConfig**: Grouped constants containing network identifier, USDC address, decimals, and for EVM chains: EIP-3009 domain parameters (name varies by chain such as "USD Coin" vs "USDC", version typically 2)
- **Network Type**: Enumeration or identifier indicating whether a network is EVM-based or SVM-based

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can configure a single-chain client in under 10 lines of code using provided helpers and constants
- **SC-002**: Developers can configure middleware to accept payments on 3+ chains in under 15 lines of code
- **SC-003**: 100% of provided USDC addresses are validated against Circle's official documentation (https://developers.circle.com/stablecoins/usdc-contract-addresses) before release
- **SC-004**: Zero runtime errors from helper functions when provided with valid inputs
- **SC-005**: Helper functions handle all supported chains (8 total: 4 mainnet + 4 testnet) with consistent interface

## Assumptions

- Native Circle USDC (not bridged variants) is the canonical token for x402 payments across all supported chains
- USDC maintains consistent 6 decimal places across all supported chains
- Network identifiers follow x402 protocol specification naming conventions
- Developers using testnet constants understand they're for testing and should switch to mainnet for production
- Token addresses remain stable on each chain (contract upgrades require library updates; documentation will indicate verification date/version)
- Default payment timeout of 300 seconds is reasonable for most use cases
- "exact" payment scheme is the most common and appropriate default
- Developers will override defaults when needed by directly modifying returned structs or passing configuration struct with optional fields
