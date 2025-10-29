# Data Model: Chi Middleware for x402 Payment Protocol

**Feature**: Chi Middleware Adapter  
**Date**: 2025-10-29  
**Status**: Complete

## Overview

This document defines the data entities, relationships, and validation rules for the Chi middleware implementation. The Chi middleware is a thin adapter that reuses existing types from the stdlib http package and x402 core package.

## Core Entities

### 1. NewChiX402Middleware (Constructor Function)

**Type**: Function  
**Signature**: `func NewChiX402Middleware(config *httpx402.Config) func(http.Handler) http.Handler`

**Description**: Constructor function that creates and returns Chi-compatible middleware for x402 payment gating.

**Input Fields**:
- `config` (*httpx402.Config): Configuration for middleware behavior

**Returns**: Middleware handler function with signature `func(http.Handler) http.Handler`

**Validation Rules**:
- Config must not be nil
- Config.FacilitatorURL must be non-empty string
- Config.PaymentRequirements must contain at least one requirement
- Each PaymentRequirement must have valid Network, Asset, PayTo fields

**Side Effects**:
- Creates FacilitatorClient(s) with hardcoded timeouts (5s verify, 60s settle)
- Calls facilitator.EnrichRequirements() to fetch network-specific config
- Logs warning if enrichment fails (graceful degradation)
- Logs info if enrichment succeeds

**State Transitions**: N/A (stateless constructor)

---

### 2. Config (from stdlib http package)

**Type**: Struct  
**Package**: `github.com/mark3labs/x402-go/http`  
**Defined**: http/middleware.go:14-26

**Description**: Configuration structure shared across all middleware implementations (stdlib, Gin, PocketBase, Chi).

**Fields**:
```go
type Config struct {
    FacilitatorURL         string                     // Primary facilitator endpoint
    FallbackFacilitatorURL string                     // Optional backup facilitator
    PaymentRequirements    []x402.PaymentRequirement  // Accepted payment methods
    VerifyOnly             bool                       // Skip settlement if true
}
```

**Field Validation**:
- `FacilitatorURL`: Required, must be valid HTTP/HTTPS URL
- `FallbackFacilitatorURL`: Optional, if provided must be valid HTTP/HTTPS URL
- `PaymentRequirements`: Required, non-empty slice
- `VerifyOnly`: Optional, defaults to false

**Relationships**:
- Contains slice of `PaymentRequirement` (one-to-many)
- Used by all middleware constructors (stdlib, Gin, PocketBase, Chi)

---

### 3. PaymentRequirement (from x402 core)

**Type**: Struct  
**Package**: `github.com/mark3labs/x402-go`  
**Defined**: types.go

**Description**: Specifies accepted payment method for a resource.

**Fields**:
```go
type PaymentRequirement struct {
    Scheme            string  // "exact" or other payment scheme
    Network           string  // "base", "base-sepolia", etc.
    MaxAmountRequired string  // Amount in atomic units (e.g., "10000")
    Asset             string  // Token contract address
    PayTo             string  // Recipient wallet address
    MaxTimeoutSeconds int     // Payment timeout window
    Resource          string  // Populated by middleware from request URL
    Description       string  // Optional human-readable description
    FeePayer          string  // Optional, populated by EnrichRequirements() for SVM
}
```

**Validation Rules** (enforced by facilitator):
- `Scheme`: Must be valid scheme (e.g., "exact")
- `Network`: Must be supported network (base, base-sepolia, solana, solana-devnet)
- `MaxAmountRequired`: Must be valid numeric string
- `Asset`: Must be valid contract address for network
- `PayTo`: Must be valid wallet address for network
- `MaxTimeoutSeconds`: Must be positive integer

**Relationships**:
- Contained in `Config.PaymentRequirements` (many-to-one)
- Matched against `PaymentPayload.Scheme` and `PaymentPayload.Network`

---

### 4. PaymentPayload (from x402 core)

**Type**: Struct  
**Package**: `github.com/mark3labs/x402-go`  
**Defined**: types.go

**Description**: Payment information sent by client in X-PAYMENT header (base64-encoded JSON).

**Fields**:
```go
type PaymentPayload struct {
    X402Version int    // Protocol version (must be 1)
    Scheme      string // Payment scheme (must match requirement)
    Network     string // Blockchain network (must match requirement)
    Data        any    // Scheme-specific payment data
}
```

**Validation Rules**:
- `X402Version`: Must equal 1 (enforced by parsePaymentHeaderFromRequest)
- `Scheme`: Must match one of the configured PaymentRequirement schemes
- `Network`: Must match one of the configured PaymentRequirement networks
- Header must be valid base64-encoded JSON

**Parsing Flow**:
1. Extract X-PAYMENT header from http.Request
2. Base64 decode header value
3. JSON unmarshal into PaymentPayload struct
4. Validate X402Version == 1
5. Return error if any step fails

---

### 5. VerifyResponse (from stdlib http package)

**Type**: Struct  
**Package**: `github.com/mark3labs/x402-go/http`  
**Defined**: facilitator.go

**Description**: Payment verification result returned by facilitator and stored in request context.

**Fields**:
```go
type VerifyResponse struct {
    IsValid       bool   // Whether payment is valid
    Payer         string // Wallet address of payer
    InvalidReason string // Reason if IsValid is false
}
```

**Validation Rules**: N/A (response from facilitator)

**Context Storage**:
- Stored with key: `httpx402.PaymentContextKey` (const)
- Accessible in handler via: `r.Context().Value(httpx402.PaymentContextKey).(*httpx402.VerifyResponse)`

**State Transitions**:
- Initial: nil (no payment processed)
- After verification: IsValid=true/false with Payer and optional InvalidReason
- Remains in context for handler lifecycle

---

### 6. SettlementResponse (from x402 core)

**Type**: Struct  
**Package**: `github.com/mark3labs/x402-go`  
**Defined**: types.go

**Description**: Payment settlement result returned by facilitator.

**Fields**:
```go
type SettlementResponse struct {
    Success      bool   // Whether settlement succeeded
    Transaction  string // Transaction hash if successful
    ErrorReason  string // Error description if not successful
}
```

**Validation Rules**: N/A (response from facilitator)

**Response Header**:
- Encoded as base64 JSON in X-PAYMENT-RESPONSE header
- Only added when VerifyOnly=false and settlement succeeds
- Not added if VerifyOnly=true or settlement fails

---

### 7. FacilitatorClient (from stdlib http package)

**Type**: Struct  
**Package**: `github.com/mark3labs/x402-go/http`  
**Defined**: facilitator.go

**Description**: HTTP client for communicating with facilitator service.

**Fields**:
```go
type FacilitatorClient struct {
    BaseURL       string        // Facilitator service URL
    Client        *http.Client  // Underlying HTTP client
    VerifyTimeout time.Duration // Timeout for verification (5s)
    SettleTimeout time.Duration // Timeout for settlement (60s)
}
```

**Hardcoded Timeouts** (per spec FR-017):
- VerifyTimeout: 5 seconds (quick verification)
- SettleTimeout: 60 seconds (blockchain transaction execution)

**Methods Used by Middleware**:
- `EnrichRequirements(requirements []x402.PaymentRequirement) ([]x402.PaymentRequirement, error)` - Fetch network config
- `Verify(payment x402.PaymentPayload, requirement x402.PaymentRequirement) (*VerifyResponse, error)` - Verify payment
- `Settle(payment x402.PaymentPayload, requirement x402.PaymentRequirement) (*x402.SettlementResponse, error)` - Settle payment

**Lifecycle**: Created once in constructor, reused for all requests

---

## Data Flow

### Request Processing Flow

```
1. Request arrives → Chi router invokes middleware
                   ↓
2. Middleware checks r.Method == "OPTIONS"
   - If yes → next.ServeHTTP(w, r) [bypass]
   - If no → continue
                   ↓
3. Build resourceURL from request (scheme + host + requestURI)
                   ↓
4. Populate PaymentRequirement.Resource with resourceURL
                   ↓
5. Check X-PAYMENT header
   - Missing → sendPaymentRequired() [402 response]
   - Present → continue
                   ↓
6. parsePaymentHeaderFromRequest()
   - base64 decode
   - JSON unmarshal → PaymentPayload
   - validate X402Version == 1
   - Error → http.Error() [400 response]
                   ↓
7. findMatchingRequirement()
   - Match by Scheme + Network
   - Not found → sendPaymentRequired() [402 response]
                   ↓
8. facilitator.Verify()
   - Send to primary facilitator
   - Failure + fallback configured → try fallback
   - Error → http.Error() [503 response]
   - IsValid=false → sendPaymentRequired() [402 response]
                   ↓
9. If !config.VerifyOnly:
     facilitator.Settle()
     - Send to primary facilitator
     - Failure + fallback configured → try fallback
     - Error → http.Error() [503 response]
     - Success=false → sendPaymentRequired() [402 response]
     - Add X-PAYMENT-RESPONSE header
                   ↓
10. context.WithValue(r.Context(), PaymentContextKey, verifyResp)
                   ↓
11. next.ServeHTTP(w, r) [handler executes]
```

### Context Data Access in Handler

```
Handler receives http.Request
         ↓
Call r.Context().Value(httpx402.PaymentContextKey)
         ↓
Type assert to *httpx402.VerifyResponse
         ↓
Access verifyResp.Payer, verifyResp.IsValid, etc.
```

## Shared Helper Functions (Internal Package)

### Package: `http/internal/helpers`

These functions will be extracted from stdlib middleware.go and shared across all middleware implementations:

#### parsePaymentHeaderFromRequest

**Signature**: `func parsePaymentHeaderFromRequest(r *http.Request) (x402.PaymentPayload, error)`

**Input**: `*http.Request` with X-PAYMENT header

**Output**: 
- `x402.PaymentPayload` on success
- `error` if header missing, invalid base64, invalid JSON, or unsupported version

**Logic**:
1. Extract header: `r.Header.Get("X-PAYMENT")`
2. Base64 decode: `base64.StdEncoding.DecodeString()`
3. JSON unmarshal: `json.Unmarshal()`
4. Validate: `payment.X402Version == 1`

---

#### findMatchingRequirement

**Signature**: `func findMatchingRequirement(payment x402.PaymentPayload, requirements []x402.PaymentRequirement) (x402.PaymentRequirement, error)`

**Input**: 
- `payment`: Payment from client
- `requirements`: Accepted payment methods

**Output**:
- Matching `PaymentRequirement` on success
- `x402.ErrUnsupportedScheme` if no match

**Logic**: Iterate requirements, match on `Scheme` and `Network`

---

#### sendPaymentRequired

**Signature**: `func sendPaymentRequired(w http.ResponseWriter, requirements []x402.PaymentRequirement)`

**Input**:
- `w`: Response writer
- `requirements`: Payment requirements to include

**Output**: Writes 402 response with JSON body

**Logic**:
1. Create `x402.PaymentRequirementsResponse{X402Version: 1, Error: "...", Accepts: requirements}`
2. Set Content-Type: application/json
3. Write status 402
4. JSON encode and write body

---

#### addPaymentResponseHeader

**Signature**: `func addPaymentResponseHeader(w http.ResponseWriter, settlement *x402.SettlementResponse) error`

**Input**:
- `w`: Response writer
- `settlement`: Settlement result

**Output**: Error if marshaling fails, nil on success

**Logic**:
1. JSON marshal settlement
2. Base64 encode
3. Set X-PAYMENT-RESPONSE header

---

## Validation Summary

### Constructor Validation (NewChiX402Middleware)
- Config non-nil ✓ (implicit panic if nil)
- FacilitatorURL non-empty ✓ (would fail on first facilitator call)
- PaymentRequirements non-empty ✓ (would fail matching)

### Runtime Validation (per request)
- X-PAYMENT header present → 402 if missing
- X-PAYMENT valid base64 → 400 if invalid
- X-PAYMENT valid JSON → 400 if invalid
- X402Version == 1 → 400 if unsupported
- Scheme+Network match → 402 if no match
- Facilitator reachable → 503 if unreachable
- Payment valid → 402 if invalid
- Settlement successful (if not VerifyOnly) → 503 if failed

### No State Validation
- Middleware is stateless
- No nonce tracking (delegated to facilitator per spec Assumptions line 135)
- No payment history (delegated to application handlers)

## Relationships Diagram

```
Config (1)
  ├─ contains ──> PaymentRequirement (many)
  ├─ creates ──> FacilitatorClient (1-2, primary + optional fallback)
  └─ used by ──> NewChiX402Middleware

NewChiX402Middleware
  ├─ returns ──> func(http.Handler) http.Handler
  └─ calls ──> facilitator.EnrichRequirements()

Middleware Handler (per request)
  ├─ receives ──> http.Request
  ├─ parses ──> PaymentPayload (from X-PAYMENT header)
  ├─ matches ──> PaymentRequirement
  ├─ calls ──> facilitator.Verify() → VerifyResponse
  ├─ calls ──> facilitator.Settle() → SettlementResponse (if !VerifyOnly)
  ├─ stores ──> VerifyResponse (in request context)
  └─ invokes ──> next handler

Handler (user code)
  ├─ receives ──> http.Request (with context)
  └─ accesses ──> VerifyResponse (via context.Value)
```

## Notes

- **No new types**: Chi middleware reuses 100% of existing types from http and x402 packages
- **No Chi-specific types**: Chi uses stdlib http.Request/ResponseWriter directly
- **Stateless design**: No persistent storage, session management, or nonce tracking
- **Shared logic**: Helper functions moved to internal package reduce duplication
- **Context immutability**: Request context is cloned (via WithContext) when storing payment info
