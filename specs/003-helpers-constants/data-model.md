# Data Model: Helper Functions and Constants

## Entities

### ChainConfig

Represents chain-specific configuration for USDC tokens and payment requirements.

**Fields:**
- `NetworkID` (string): x402 protocol network identifier (e.g., "base", "solana", "polygon-amoy")
- `USDCAddress` (string): Official Circle USDC contract address or mint address
- `Decimals` (uint8): Token decimals (always 6 for USDC)
- `EIP3009Name` (string): EIP-3009 domain parameter `name` (empty for non-EVM chains like Solana)
- `EIP3009Version` (string): EIP-3009 domain parameter `version` (empty for non-EVM chains)

**Validation Rules:**
- NetworkID must not be empty
- USDCAddress must not be empty
- Decimals must be 6 (USDC standard)
- For EVM chains: EIP3009Name and EIP3009Version must not be empty
- For SVM chains: EIP3009Name and EIP3009Version must be empty

**Constants (instances):**
- `SolanaMainnet`: Solana mainnet config
- `SolanaDevnet`: Solana devnet config
- `BaseMainnet`: Base mainnet config
- `BaseSepolia`: Base Sepolia testnet config
- `PolygonMainnet`: Polygon PoS mainnet config
- `PolygonAmoy`: Polygon Amoy testnet config
- `AvalancheMainnet`: Avalanche C-Chain mainnet config
- `AvalancheFuji`: Avalanche Fuji testnet config

### PaymentRequirementConfig

Configuration struct for customizing PaymentRequirement creation.

**Fields:**
- `Chain` (ChainConfig): Chain configuration (required)
- `Amount` (string): Human-readable payment amount (e.g., "1.5" = 1.5 USDC)
- `RecipientAddress` (string): Payment recipient address
- `Scheme` (string): Payment scheme (optional, defaults to "exact")
- `MaxTimeoutSeconds` (uint32): Maximum payment timeout (optional, defaults to 300)
- `MimeType` (string): Response MIME type (optional, defaults to "application/json")

**Validation Rules:**
- Chain must be valid ChainConfig
- Amount must parse to valid float64 (zero and positive allowed; negative returns error)
- Amount precision beyond 6 decimals is rounded using standard float64 rounding (banker's rounding)
- Zero amounts ("0" or "0.0") are explicitly allowed for free-with-signature authorization flows
- RecipientAddress must not be empty
- RecipientAddress format depends on chain (EVM: 0x-prefixed hex, Solana: base58)
- MaxTimeoutSeconds must be positive if provided
- Scheme must be "exact", "estimate", or empty (defaults to "exact")
- All validation errors return structured format: "parameterName: reason"

### TokenConfig

Configuration for client signer token support (from existing x402 types).

**Fields (from existing type):**
- `Address` (string): Token contract/mint address
- `Symbol` (string): Token symbol (e.g., "USDC")
- `Decimals` (uint8): Token decimals
- `Priority` (int): Token priority for selection

**Helper Creation:**
Helper function creates TokenConfig from ChainConfig with sensible defaults.

### NetworkType

Enumeration for network categories.

**Values:**
- `NetworkTypeEVM`: Ethereum Virtual Machine chains (Base, Polygon, Avalanche, etc.)
- `NetworkTypeSVM`: Solana Virtual Machine chains (Solana mainnet, devnet)
- `NetworkTypeUnknown`: Unrecognized network

**Usage:**
Network validator helper returns NetworkType for a given network identifier string.

## Relationships

```text
ChainConfig
    ↓ (used by)
PaymentRequirementConfig
    ↓ (creates)
PaymentRequirement (existing x402 type)

ChainConfig
    ↓ (used by)
TokenConfigHelper
    ↓ (creates)
TokenConfig (existing x402 type)

NetworkID (string)
    ↓ (validated by)
NetworkValidator
    ↓ (returns)
NetworkType (enum)
```

## State Transitions

**ChainConfig**: Immutable constants, no state transitions

**PaymentRequirementConfig**: 
1. Created with required fields
2. Optional fields defaulted if not provided
3. Validated on helper function call
4. Converted to PaymentRequirement struct

## Data Flow

### PaymentRequirement Creation Flow
```text
Developer Input (chain, amount, recipient)
    ↓
PaymentRequirementConfig struct
    ↓
Validation (format, ranges, required fields)
    ↓
Amount conversion (decimal string → atomic uint64)
    ↓
EIP-3009 extra field population (for EVM chains)
    ↓
PaymentRequirement struct (existing x402 type)
```

### TokenConfig Creation Flow
```text
ChainConfig constant
    ↓
TokenConfigHelper function
    ↓
TokenConfig struct with:
  - Address from ChainConfig.USDCAddress
  - Symbol = "USDC"
  - Decimals = 6
  - Priority from optional parameter
```

### Network Validation Flow
```text
Network identifier string (e.g., "base")
    ↓
NetworkValidator function
    ↓
Lookup in known network list
    ↓
Return NetworkType enum or error
```

## Examples

### ChainConfig Usage
```go
// Use predefined constants
config := x402.BaseMainnet

fmt.Println(config.NetworkID)        // "base"
fmt.Println(config.USDCAddress)      // "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
fmt.Println(config.EIP3009Name)      // "USD Coin"
fmt.Println(config.EIP3009Version)   // "2"
```

### PaymentRequirement Creation
```go
// Create payment requirement for Base
req, err := x402.NewPaymentRequirement(x402.PaymentRequirementConfig{
    Chain:            x402.BaseMainnet,
    Amount:           "10.50",  // 10.5 USDC
    RecipientAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    // Scheme, MaxTimeoutSeconds, MimeType use defaults
})
// Result: PaymentRequirement with:
//   - Network: "base"
//   - Address: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
//   - Amount: 10500000 (atomic units)
//   - Scheme: "exact"
//   - Extra: {"name": "USD Coin", "version": "2"}
```

### TokenConfig Creation
```go
// Create token config for Solana mainnet
token := x402.NewTokenConfig(x402.SolanaMainnet, 1) // priority 1

// Result: TokenConfig with:
//   - Address: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
//   - Symbol: "USDC"
//   - Decimals: 6
//   - Priority: 1
```

### Network Validation
```go
// Validate network type
netType, err := x402.ValidateNetwork("polygon-amoy")
// Result: NetworkTypeEVM, nil

netType, err := x402.ValidateNetwork("solana")
// Result: NetworkTypeSVM, nil

netType, err := x402.ValidateNetwork("unknown-chain")
// Result: NetworkTypeUnknown, error
```
