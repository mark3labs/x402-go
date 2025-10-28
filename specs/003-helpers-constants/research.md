# Research: Helper Functions and Constants

## EIP-3009 Domain Parameters for USDC Contracts

### Decision
Use chain-specific EIP-3009 domain parameters (name and version) fetched directly from deployed USDC contracts for each supported chain.

### Rationale
EIP-3009 `receiveWithAuthorization` requires signed transfers that include domain parameters in the signature. These parameters vary by chain and must match exactly for signature verification to succeed. Direct verification via `cast` ensures accuracy against production contracts.

### Verification Method
Used `cast call` to query `name()` and `version()` from official Circle USDC contracts on each chain:

```bash
# Example queries
cast call 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 "name()(string)" --rpc-url https://eth.llamarpc.com
cast call 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 "version()(string)" --rpc-url https://eth.llamarpc.com
```

### Verified EIP-3009 Parameters

**Mainnet Chains:**

| Chain | Network ID | USDC Address | Name | Version |
|-------|-----------|--------------|------|---------|
| Ethereum | `ethereum` | `0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48` | `USD Coin` | `2` |
| Arbitrum | `arbitrum` | `0xaf88d065e77c8cC2239327C5EDb3A432268e5831` | `USD Coin` | `2` |
| Base | `base` | `0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913` | `USD Coin` | `2` |
| Polygon | `polygon` | `0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359` | `USD Coin` | `2` |
| Avalanche | `avalanche` | `0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E` | `USD Coin` | `2` |

**Testnet Chains:**

| Chain | Network ID | USDC Address | Name | Version |
|-------|-----------|--------------|------|---------|
| Base Sepolia | `base-sepolia` | `0x036CbD53842c5426634e7929541eC2318f3dCF7e` | `USDC` | `2` |
| Polygon Amoy | `polygon-amoy` | `0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582` | `USDC` | `2` |
| Avalanche Fuji | `avalanche-fuji` | `0x5425890298aed601595a70AB815c96711a31Bc65` | `USD Coin` | `2` |

**Key Finding:** Testnet contracts on Base Sepolia and Polygon Amoy use short name `USDC` instead of `USD Coin`. All contracts use version `2`.

### Solana/SVM Considerations

#### Decision
Solana chains (mainnet and devnet) do NOT use EIP-3009. No domain parameters needed.

#### Rationale
EIP-3009 is an Ethereum-specific standard (EIP = Ethereum Improvement Proposal). Solana uses its own program-based authorization model with different signature schemes (ed25519 vs secp256k1).

#### Verified Solana USDC Addresses
- **Solana Mainnet**: `EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v`
- **Solana Devnet**: `4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU`

## Helper Function Architecture

### Decision
Provide three distinct helper functions: PaymentRequirement builder, TokenConfig builder, and network validator.

### Rationale
Separation of concerns keeps each helper focused on a single task while supporting various use cases:
1. PaymentRequirement builder: for middleware accepting payments
2. TokenConfig builder: for client signer configuration
3. Network validator: for payment matching logic

### Error Handling
All helper functions return structured errors with format: `"parameterName: reason"` (e.g., "amount: must be positive", "recipientAddress: cannot be empty"). This provides clear, actionable feedback for debugging invalid configurations.

### Alternatives Considered
- **Single uber-function**: Rejected due to complexity and unclear purpose
- **Method chaining builder**: Rejected to maintain Go stdlib-first approach and avoid unnecessary abstraction

## Amount Format Handling

### Decision
Accept amounts as human-readable decimal strings (e.g., "1.5") and convert to atomic units internally.

### Rationale
USDC uses 6 decimals on all chains. Converting "1.5" to 1,500,000 atomic units is straightforward with Go's `strconv` package and basic math. This matches developer expectations and reduces errors from manual atomic unit calculation.

### Implementation Approach
```go
// Pseudocode
amount, err := strconv.ParseFloat(humanAmount, 64)
atomicUnits := uint64(amount * 1e6)
```

### Edge Cases
- Precision beyond 6 decimals: round using standard float64 rounding (banker's rounding via Go's float64 arithmetic)
- Negative amounts: return structured error with parameter name and reason
- Zero amounts: explicitly allow (valid for free-with-signature authorization flows)

## Constants Structure

### Decision
Group related values (network ID, token address, decimals, EIP-3009 params) into exported structs per chain.

### Rationale
Grouping prevents mismatched values (e.g., using Base network ID with Polygon address). Struct fields are self-documenting and enable IDE autocomplete.

### Example Structure
```go
type ChainConfig struct {
    NetworkID string
    USDCAddress string
    Decimals uint8
    EIP3009Name string    // Empty for non-EVM chains
    EIP3009Version string // Empty for non-EVM chains
}

var BaseMainnet = ChainConfig{
    NetworkID: "base",
    USDCAddress: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
    Decimals: 6,
    EIP3009Name: "USD Coin",
    EIP3009Version: "2",
}
```

### Alternatives Considered
- **Separate constants**: Rejected due to high risk of mismatching values
- **Map-based lookup**: Rejected as less type-safe and harder to discover via IDE

## Default Values for PaymentRequirement

### Decision
Use these defaults:
- Scheme: `"exact"`
- MaxTimeoutSeconds: `300` (5 minutes)
- MimeType: `"application/json"`

### Rationale
Based on existing x402 protocol patterns and common use cases:
- Exact payment scheme is most common and predictable
- 5 minutes allows reasonable time for user action without excessive lock-up
- JSON is default API response format

### Customization Strategy
Developers can override by modifying returned struct fields directly or passing optional config struct.

## Package Organization

### Decision
Export all helpers and constants from root `x402` package.

### Rationale
Aligns with Go convention for small libraries. Avoids import path stuttering (`x402.constants.BaseMainnet` vs `x402.BaseMainnet`).

### Alternatives Considered
- **Separate subpackages**: Rejected as unnecessarily complex for ~10 constants and 3 helper functions

## Source Structure

### Decision
Add new file `chains.go` in project root for chain constants and helper functions.

### Rationale
- Keeps all chain-related code in one discoverable location
- Follows existing project pattern (types.go, errors.go, etc. at root)
- Stdlib-first approach: no need for additional directories

## Testing Strategy

### Decision
- Unit tests for each helper function with table-driven test cases
- Validate all USDC addresses via test that queries contracts (requires network access, may be skipped in CI)
- Test all edge cases: invalid amounts, empty addresses, unknown networks

### Rationale
Maintains test coverage requirement from constitution. Table-driven tests provide comprehensive coverage efficiently.

## Documentation Approach

### Decision
GoDoc comments on all exported types and functions. No markdown documentation beyond this research.md, data-model.md, and quickstart.md.

### Rationale
Follows constitution principle I: no unnecessary documentation. GoDoc is sufficient for API documentation and discoverable via `go doc` command.

## Verification Date

**Last Verified**: 2025-10-28

All USDC contract addresses and EIP-3009 parameters verified against:
- Circle's official documentation: https://developers.circle.com/stablecoins/usdc-contract-addresses
- Live contract queries via `cast call` on production and testnet RPCs

Developers should check for updated addresses when upgrading library versions.
