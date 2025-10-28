# Data Model: x402 Payment Client

**Date**: 2025-10-28
**Feature**: 002-x402-client

## Core Entities

### 1. PaymentClient
**Purpose**: Main HTTP client that handles x402 payment flows

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| httpClient | *http.Client | Yes | Underlying HTTP client |
| signers | []Signer | Yes | List of payment signers |
| selector | *PaymentSelector | Yes | Logic for signer/token selection |

**Relationships**:
- Owns multiple Signers
- Uses PaymentSelector for decision making

### 2. Signer (Interface)
**Purpose**: Abstract interface for blockchain-specific signing

| Method | Parameters | Returns | Description |
|--------|------------|---------|-------------|
| Network | - | string | Returns network identifier (e.g., "base", "solana") |
| Scheme | - | string | Returns scheme identifier (e.g., "exact") |
| CanSign | requirements | bool | Checks if can satisfy requirements |
| Sign | requirements, amount | PaymentPayload, error | Creates signed payment |
| GetPriority | - | int | Returns signer priority |
| GetTokens | - | []TokenConfig | Returns supported tokens |
| GetMaxAmount | - | *big.Int | Returns per-call limit |

### 3. EVMSigner
**Purpose**: EVM-specific implementation of Signer

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| privateKey | *ecdsa.PrivateKey | Yes | Signing key |
| network | string | Yes | Network identifier |
| tokens | []TokenConfig | Yes | Supported tokens |
| priority | int | No | Signer priority (default: 0) |
| maxAmount | *big.Int | No | Per-call limit |
| address | common.Address | Yes | Derived from private key |

### 4. SVMSigner  
**Purpose**: Solana-specific implementation of Signer

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| privateKey | solana.PrivateKey | Yes | Signing key |
| network | string | Yes | Network identifier |
| tokens | []TokenConfig | Yes | Supported tokens |
| priority | int | No | Signer priority (default: 0) |
| maxAmount | *big.Int | No | Per-call limit |
| publicKey | solana.PublicKey | Yes | Derived from private key |

### 5. TokenConfig
**Purpose**: Configuration for a specific token

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| address | string | Yes | Token contract/mint address |
| symbol | string | Yes | Token symbol (e.g., "USDC") |
| decimals | int | Yes | Token decimal places |
| priority | int | No | Token priority within signer |
| name | string | No | Human-readable name |

### 6. PaymentRequirements
**Purpose**: Server's payment requirements (from 402 response)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| scheme | string | Yes | Payment scheme |
| network | string | Yes | Blockchain network |
| maxAmountRequired | string | Yes | Amount in atomic units |
| asset | string | Yes | Token address |
| payTo | string | Yes | Recipient address |
| resource | string | Yes | Resource URL |
| maxTimeoutSeconds | int | Yes | Payment validity period |
| extra | map[string]interface{} | No | Scheme-specific data |

### 7. PaymentPayload
**Purpose**: Signed payment data to send to server

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| x402Version | int | Yes | Protocol version (1) |
| scheme | string | Yes | Payment scheme |
| network | string | Yes | Blockchain network |
| payload | interface{} | Yes | Scheme-specific signed data |



## State Transitions

### Signer Lifecycle
```
Created → Configured → Active
```

### Payment Flow
```
Request → 402 Response → Parse Requirements → Select Signer → 
Check Max Amount → Sign Payment → Retry Request → Success/Failure
```

## Validation Rules

### Signer Validation
- Private key must be valid for the blockchain
- Network must be supported ("base", "base-sepolia", "solana", etc.)
- At least one token must be configured
- Priority must be >= 0
- Max amount must be > 0 if set

### Payment Selection
- Signer network must match requirement network
- Signer must have token matching requirement asset
- Payment amount must not exceed max amount (if set)
- Payment must be within validity timeframe

## Indexes & Lookups

### Primary Keys
- Signer: Derived from private key address + network
- Token: Contract/mint address + network

### Lookup Patterns
1. Find signers by network → O(n) scan
2. Find signers by token → O(n*m) scan
3. Sort signers by priority → O(n log n)

## Concurrency Considerations

### Thread-Safe Components
- PaymentSelector: Stateless, inherently safe
- HTTP Transport: New instance per request

### Mutable State
- Signer configuration: Immutable after creation
- No shared mutable state between requests

## Serialization Formats

### Payment Header (Base64-encoded JSON)
```json
{
  "x402Version": 1,
  "scheme": "exact",
  "network": "base",
  "payload": {
    "signature": "0x...",
    "authorization": {...}
  }
}
```

## Performance Characteristics

| Operation | Complexity | Typical Time |
|-----------|------------|--------------|
| Signer selection | O(n log n) | < 1ms |
| Payment signing (EVM) | O(1) | ~2ms |
| Payment signing (SVM) | O(1) | ~5ms |
| Max amount check | O(1) | < 0.01ms |
| HTTP round-trip | O(1) | Network dependent |

## Storage Requirements

| Component | Size | Growth Rate |
|-----------|------|-------------|
| Signer config | ~500 bytes | None (static) |
| Token config | ~100 bytes/token | None (static) |
| Total per signer | < 1KB typical | None |