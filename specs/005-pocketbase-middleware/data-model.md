# Data Model: PocketBase Middleware for x402 Payment Protocol

**Feature**: 005-pocketbase-middleware  
**Date**: 2025-10-29  
**Status**: Phase 1 Design

## Overview

This document describes the data flow, entities, and state transitions for the PocketBase x402 payment middleware. The middleware is stateless and delegates all payment tracking to the facilitator service.

---

## Entities

### 1. Config (Reused from stdlib)

**Source**: `github.com/mark3labs/x402-go/http.Config`  
**Purpose**: Configuration for middleware initialization  
**Lifecycle**: Created by user, passed to middleware factory, read-only during request processing

**Fields**:
```go
type Config struct {
    // Primary facilitator endpoint (required)
    FacilitatorURL string
    
    // Optional backup facilitator endpoint
    FallbackFacilitatorURL string
    
    // Payment requirements (required, at least one)
    PaymentRequirements []x402.PaymentRequirement
    
    // If true, skip settlement (verification only)
    VerifyOnly bool
}
```

**Validation Rules**:
- `FacilitatorURL` must be non-empty valid URL
- `PaymentRequirements` must have at least one element
- Each PaymentRequirement must have: Scheme, Network, MaxAmountRequired, Asset, PayTo

**Relationships**:
- Contains multiple `PaymentRequirement` (1:N)
- Used by `FacilitatorClient` (1:1)

---

### 2. PaymentRequirement (Reused from core)

**Source**: `github.com/mark3labs/x402-go.PaymentRequirement`  
**Purpose**: Defines accepted payment terms  
**Lifecycle**: Created in Config, enriched by facilitator, used for matching

**Fields**:
```go
type PaymentRequirement struct {
    Scheme            string // "exact", "signature", etc.
    Network           string // "base", "base-sepolia", "solana-mainnet"
    MaxAmountRequired string // Atomic units as string (e.g., "10000" = 0.01 USDC with 6 decimals)
    Asset             string // Token contract address or mint
    PayTo             string // Recipient wallet address
    MaxTimeoutSeconds int    // Payment validity window
    Resource          string // Populated by middleware (request URL)
    Description       string // Human-readable description
    FeePayer          string // (SVM only) Fee payer address from facilitator
}
```

**State Transitions**:
1. **Created** → User defines in Config
2. **Enriched** → Middleware calls `facilitator.EnrichRequirements()` to populate network-specific fields (e.g., FeePayer for SVM)
3. **Populated** → Middleware sets `Resource` field to request URL
4. **Matched** → Middleware compares with incoming payment (Scheme + Network)

---

### 3. PaymentPayload (Reused from core)

**Source**: `github.com/mark3labs/x402-go.PaymentPayload`  
**Purpose**: Parsed payment data from X-PAYMENT header  
**Lifecycle**: Parsed from request, validated, sent to facilitator

**Fields**:
```go
type PaymentPayload struct {
    X402Version int         // Must be 1
    Scheme      string      // Must match PaymentRequirement.Scheme
    Network     string      // Must match PaymentRequirement.Network
    Payload     interface{} // Scheme-specific data (e.g., transaction string)
}
```

**Validation Rules**:
- `X402Version` must equal 1
- `Scheme` must match one of the PaymentRequirements
- `Network` must match the corresponding PaymentRequirement
- `Payload` structure depends on Scheme (e.g., `{"transaction": "<base64>"}` for EIP-3009)

---

### 4. VerifyResponse (Reused from http package)

**Source**: `github.com/mark3labs/x402-go/http.VerifyResponse` (internal struct from facilitator.go)  
**Purpose**: Payment verification result from facilitator  
**Lifecycle**: Returned by facilitator, stored in request store

**Fields**:
```go
type VerifyResponse struct {
    IsValid       bool   // True if payment is valid
    Payer         string // Payer wallet address
    InvalidReason string // Reason if IsValid=false
}
```

**Storage**: Stored in PocketBase request store with key `"x402_payment"`

**Access Pattern**:
```go
// Middleware stores
e.Set("x402_payment", verifyResp)

// Handler retrieves
verifyResp := e.Get("x402_payment").(*http.VerifyResponse)
```

---

### 5. SettlementResponse (Reused from core)

**Source**: `github.com/mark3labs/x402-go.SettlementResponse`  
**Purpose**: Payment settlement result from facilitator  
**Lifecycle**: Returned by facilitator (if VerifyOnly=false), added to X-PAYMENT-RESPONSE header

**Fields**:
```go
type SettlementResponse struct {
    Success      bool   // True if settlement succeeded
    Transaction  string // On-chain transaction hash
    ErrorReason  string // Reason if Success=false
}
```

**Encoding**: Base64-encoded JSON in X-PAYMENT-RESPONSE header

---

### 6. FacilitatorClient (Reused from http package)

**Source**: `github.com/mark3labs/x402-go/http.FacilitatorClient`  
**Purpose**: HTTP client for facilitator API  
**Lifecycle**: Created during middleware initialization, reused for all requests

**Fields**:
```go
type FacilitatorClient struct {
    BaseURL       string
    Client        *http.Client
    VerifyTimeout time.Duration // 5s
    SettleTimeout time.Duration // 60s
}
```

**Methods** (reused):
- `EnrichRequirements([]PaymentRequirement) ([]PaymentRequirement, error)` - Fetch network-specific config from `/supported`
- `Verify(PaymentPayload, PaymentRequirement) (*VerifyResponse, error)` - POST to `/verify`
- `Settle(PaymentPayload, PaymentRequirement) (*SettlementResponse, error)` - POST to `/settle`

---

## Data Flow

### Request Processing Flow

```
1. Request arrives → core.RequestEvent created by PocketBase
                    ↓
2. Middleware checks → e.Request.Header.Get("X-PAYMENT")
                    ↓
    ┌──────────────┴──────────────┐
    │ No header?                  │
    └──────────────┬──────────────┘
                   ↓
        ┌─────────┴─────────┐
        │ YES               │ NO
        ↓                   ↓
    Return 402           Parse header
    with requirements    (base64 + JSON)
                              ↓
                         ┌────┴────┐
                         │ Valid?  │
                         └────┬────┘
                              ↓
                    ┌─────────┴─────────┐
                    │ YES               │ NO
                    ↓                   ↓
                Find matching       Return 400
                requirement         Bad Request
                    ↓
            Verify with facilitator
                    ↓
            ┌───────┴───────┐
            │ Valid?        │
            └───────┬───────┘
                    ↓
        ┌───────────┴───────────┐
        │ YES                   │ NO
        ↓                       ↓
    Store in e.Set()        Return 402
    "x402_payment"          with requirements
        ↓
    Settle (if !VerifyOnly)
        ↓
    Add X-PAYMENT-RESPONSE header
        ↓
    Call e.Next()
        ↓
    Protected handler executes
```

### Data Transformations

#### 1. Config → Enriched Requirements
```
Config.PaymentRequirements
    ↓ facilitator.EnrichRequirements()
Enriched PaymentRequirements (with FeePayer for SVM)
```

#### 2. Enriched Requirements → Request-Specific Requirements
```
Enriched PaymentRequirements
    ↓ For each requirement:
      requirement.Resource = scheme + host + requestURI
      if requirement.Description == "":
          requirement.Description = "Payment required for " + path
Request-Specific PaymentRequirements
```

#### 3. X-PAYMENT Header → PaymentPayload
```
Header value (base64 string)
    ↓ base64.StdEncoding.DecodeString()
JSON bytes
    ↓ json.Unmarshal()
PaymentPayload struct
    ↓ Validate X402Version == 1
Valid PaymentPayload
```

#### 4. SettlementResponse → X-PAYMENT-RESPONSE Header
```
SettlementResponse struct
    ↓ json.Marshal()
JSON bytes
    ↓ base64.StdEncoding.EncodeToString()
Header value (base64 string)
    ↓ e.Response.Header().Set("X-PAYMENT-RESPONSE", ...)
Response header
```

---

## State Management

### Middleware State
**Type**: Stateless  
**Rationale**: Each request is independently verified. No session tracking, no nonce management.

**Initialized Once** (at app startup):
- FacilitatorClient instances (primary + fallback)
- Enriched PaymentRequirements (from facilitator `/supported`)

**Per-Request State** (stored in PocketBase request store):
- Key: `"x402_payment"`
- Value: `*VerifyResponse` (Payer, IsValid, InvalidReason)
- Lifetime: Single request duration
- Access: Middleware writes, handler reads (optional)

### External State (Facilitator)
- Nonce tracking (prevents replay attacks)
- Payment verification results (cached by facilitator)
- Settlement transaction status (on-chain)

**Middleware does NOT track**:
- Payment history
- User sessions
- Nonce values
- Transaction confirmations

---

## Error States

### Client Errors (4xx)

| Scenario | Status | Response Body | Data Impact |
|----------|--------|---------------|-------------|
| Missing X-PAYMENT | 402 | PaymentRequirementsResponse | None - request not processed |
| Invalid base64 | 400 | `{"x402Version": 1, "error": "Invalid payment header"}` | None - parsing failed |
| Invalid JSON | 400 | Same as above | None - parsing failed |
| Unsupported version | 400 | Same as above | None - version mismatch |
| No matching requirement | 402 | PaymentRequirementsResponse | None - payment rejected |
| Invalid payment | 402 | PaymentRequirementsResponse | None - verification failed |

### Server Errors (5xx)

| Scenario | Status | Response Body | Data Impact |
|----------|--------|---------------|-------------|
| Facilitator unreachable | 503 | `{"x402Version": 1, "error": "Payment verification failed"}` | Request blocked - retry possible |
| Facilitator timeout | 503 | Same as above | Request blocked - retry possible |
| Settlement failure | 503 | `{"x402Version": 1, "error": "Payment settlement failed"}` | Payment verified but not settled |

---

## Validation Rules

### Config Validation (at initialization)
- FacilitatorURL is non-empty: `config.FacilitatorURL != ""`
- At least one PaymentRequirement: `len(config.PaymentRequirements) > 0`
- Each PaymentRequirement has: Scheme, Network, MaxAmountRequired, Asset, PayTo

### PaymentPayload Validation (at request time)
- X402Version equals 1: `payment.X402Version == 1`
- Scheme matches requirement: `payment.Scheme == requirement.Scheme`
- Network matches requirement: `payment.Network == requirement.Network`
- Payload is non-nil: `payment.Payload != nil`

### Request Validation (at request time)
- X-PAYMENT header present: `e.Request.Header.Get("X-PAYMENT") != ""`
- Base64 decodes successfully: `base64.StdEncoding.DecodeString(headerValue)`
- JSON unmarshals successfully: `json.Unmarshal(decoded, &payment)`

---

## Concurrency Considerations

### Thread Safety
- **FacilitatorClient**: Thread-safe (uses standard http.Client with connection pooling)
- **Config**: Read-only after initialization (no mutations during requests)
- **PaymentRequirements**: Enriched once at startup, read-only thereafter
- **Request Store**: Request-scoped (no sharing between concurrent requests)

### Shared Resources
- HTTP connection pool (managed by http.Client)
- Logger (slog.Default() is thread-safe)

### No Synchronization Needed
- Each request is independent
- No shared mutable state
- No caching or session management

---

## Summary

**Key Characteristics**:
- ✅ Stateless: No session tracking, no in-memory state
- ✅ Reuses stdlib types: Config, PaymentRequirement, PaymentPayload, VerifyResponse, SettlementResponse, FacilitatorClient
- ✅ Request-scoped storage: Uses PocketBase request store with key "x402_payment"
- ✅ Fail-safe: Graceful degradation if enrichment fails (use original requirements)
- ✅ Thread-safe: No shared mutable state, read-only config

**Data Dependencies**:
- Upstream: Facilitator service for verification and settlement
- Downstream: Protected handlers access VerifyResponse via e.Get()
- External: Blockchain networks for transaction execution

**Validation Layers**:
1. Config validation at startup
2. Header parsing at request time
3. Payment matching against requirements
4. Facilitator verification (cryptographic)
5. Settlement confirmation (blockchain)
