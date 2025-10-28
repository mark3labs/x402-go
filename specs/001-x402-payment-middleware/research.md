# Research: x402 Payment Middleware

**Date**: 2025-10-28  
**Feature**: x402 Payment Middleware Implementation

## Executive Summary

Research conducted to resolve technical decisions for implementing x402 payment middleware in Go. Key findings support a stdlib-only approach using standard middleware patterns with clear separation between shared types and HTTP-specific logic.

## Research Findings

### 1. Middleware Pattern Implementation

**Decision**: Use closure-based middleware pattern with `func(http.Handler) http.Handler` signature

**Rationale**: 
- Standard Go pattern widely adopted in the community
- Allows easy composition and chaining of middleware
- No external dependencies required
- Supports both route-specific and global middleware application

**Alternatives considered**:
- Context-based middleware: More complex, not needed for our stateless design
- Interface-based middleware: Less flexible for composition
- Framework-specific patterns: Would introduce external dependencies

**Implementation pattern** (from Alex Edwards):
```go
func NewX402Middleware(config *Config) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Middleware logic here
            next.ServeHTTP(w, r)
        })
    }
}
```

### 2. x402 Protocol Integration

**Decision**: Direct implementation of x402 specification types and flows

**Rationale**:
- Protocol is well-defined with clear JSON schemas
- Simple HTTP header-based communication (X-PAYMENT, X-PAYMENT-RESPONSE)
- Facilitator API uses standard REST endpoints

**Key integration points**:
- PaymentRequirementsResponse: JSON response with 402 status
- PaymentPayload: Base64-encoded JSON in X-PAYMENT header
- SettlementResponse: Base64-encoded JSON in X-PAYMENT-RESPONSE header
- Facilitator endpoints: /verify, /settle, /supported

### 3. Multi-Chain Support Strategy

**Decision**: Chain-agnostic middleware with scheme-specific payload handling

**Rationale**:
- x402 protocol already abstracts chain differences
- Facilitator handles chain-specific logic
- Middleware only needs to route based on scheme/network fields

**Chain handling**:
- EVM: EIP-3009 authorization in payload.authorization
- SVM: Serialized transaction in payload.transaction
- Both use same PaymentPayload wrapper structure

### 4. Error Handling Strategy

**Decision**: Map x402 error codes to appropriate HTTP status codes

**Rationale**:
- Maintains HTTP semantics
- Clear client feedback
- Consistent with x402 specification

**Error mapping**:
```
Payment Required → 402 Payment Required
Invalid Payment → 400 Bad Request  
Facilitator Unavailable → 503 Service Unavailable
Settlement Failed → 402 Payment Required (with error details)
Malformed Headers → 400 Bad Request
```

### 5. Testing Strategy

**Decision**: Table-driven tests with mocked facilitator responses

**Rationale**:
- Follows Go testing best practices
- Enables comprehensive coverage without blockchain dependency
- Tests can run in isolation

**Test categories**:
- Unit tests: Type serialization, header parsing, error handling
- Integration tests: Full middleware flow with mock facilitator
- Example tests: Demonstrate usage patterns

### 6. Configuration Design

**Decision**: Functional options pattern for middleware configuration

**Rationale**:
- Flexible and extensible
- Maintains backward compatibility
- Clear API for developers

**Configuration options**:
```go
type Config struct {
    FacilitatorURL    string
    FallbackURL       string  // Optional
    PaymentRequirements []PaymentRequirement
    VerifyOnly        bool
}
```

### 7. Package Structure

**Decision**: Two-package design (x402 for types, http for middleware)

**Rationale**:
- Clear separation of concerns
- Reusable types for potential future transports
- Follows Go module conventions
- Aligns with user requirements

**Package responsibilities**:
- `x402/`: Core types, validation, serialization
- `http/`: Middleware, request handling, facilitator client

## Technical Constraints Resolved

1. **Stateless operation**: Confirmed - nonce tracking delegated to facilitator
2. **Performance targets**: Achievable with stdlib - minimal overhead expected
3. **Concurrent handling**: Go's http.Server handles concurrency naturally
4. **Memory usage**: Low footprint - only request-scoped allocations

## Dependencies Analysis

**Required**: None (stdlib only)

**Rationale for no external dependencies**:
- encoding/json: Native JSON handling
- encoding/base64: Header encoding/decoding  
- net/http: HTTP client and server
- context: Request cancellation and timeouts
- log/slog: Structured logging (Go 1.21+)

All requirements can be met with standard library, maintaining principle IV (Stdlib-First Approach).

## Security Considerations

1. **Replay attacks**: Mitigated by facilitator nonce tracking
2. **Header injection**: Validate base64 encoding and JSON structure
3. **Timeout attacks**: Use context with timeouts for facilitator calls
4. **Error disclosure**: Sanitize error messages to avoid information leakage

## Next Steps

All technical clarifications resolved. Ready to proceed with Phase 1 design:
- Data model definition
- API contract specification
- Implementation guide creation