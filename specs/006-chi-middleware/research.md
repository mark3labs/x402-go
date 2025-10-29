# Research: Chi Middleware for x402 Payment Protocol

**Feature**: Chi Middleware Adapter  
**Date**: 2025-10-29  
**Status**: Complete

## Overview

This document consolidates research findings for implementing a Chi-compatible middleware adapter for the x402 payment protocol. All technical unknowns from the Technical Context have been resolved through analysis of existing implementations and Chi documentation.

## Key Decisions

### Decision 1: Chi Middleware Signature

**What was chosen**: `func NewChiX402Middleware(config *httpx402.Config) func(http.Handler) http.Handler`

**Rationale**:
- Chi uses standard net/http middleware pattern: `func(http.Handler) http.Handler`
- Identical to stdlib middleware signature in `http/middleware.go`
- Constructor function returns the middleware handler (following Gin pattern)
- Allows initialization of facilitator client and enrichment at construction time

**Alternatives considered**:
- Direct middleware function without constructor: Rejected because enrichment and facilitator client setup must happen before middleware executes
- Variadic options pattern (functional options): Rejected per spec (Out of Scope line 149) for consistency with stdlib Config struct approach

**Supporting evidence**:
- Chi documentation: "middlewares are just stdlib net/http middleware handlers"
- Chi router interface explicitly defines: `Use(middlewares ...func(http.Handler) http.Handler)`
- stdlib middleware.go:38 uses identical pattern

### Decision 2: Helper Function Sharing Strategy

**What was chosen**: Create `http/internal/helpers` package with shared functions

**Rationale**:
- Chi uses identical types to stdlib (http.Request, http.ResponseWriter)
- Eliminates code duplication across 4 middleware implementations (stdlib, Gin, PocketBase, Chi)
- Ensures consistent behavior for payment parsing, validation, and response handling
- Internal package prevents external usage while enabling cross-package sharing

**Alternatives considered**:
- Duplicate helpers in Chi package: Rejected due to maintenance burden and Constitution Principle V (Code Conciseness)
- Export helpers publicly: Rejected to keep internal implementation details private
- Keep framework-specific helpers: Rejected because Chi/stdlib use identical types (no adapter needed)

**Functions to share**:
1. `parsePaymentHeaderFromRequest(r *http.Request) (x402.PaymentPayload, error)` - Parse and validate X-PAYMENT header
2. `findMatchingRequirement(payment x402.PaymentPayload, requirements []x402.PaymentRequirement) (x402.PaymentRequirement, error)` - Match payment to requirement
3. `sendPaymentRequired(w http.ResponseWriter, requirements []x402.PaymentRequirement)` - Send 402 response
4. `addPaymentResponseHeader(w http.ResponseWriter, settlement *x402.SettlementResponse) error` - Add X-PAYMENT-RESPONSE header

### Decision 3: Context Storage Pattern

**What was chosen**: Use stdlib `context.WithValue(r.Context(), httpx402.PaymentContextKey, verifyResp)`

**Rationale**:
- Chi handlers receive standard `http.Request` with `Context()` method
- Consistent with stdlib middleware pattern (http/middleware.go:167)
- Uses existing `PaymentContextKey` constant from http package
- No Chi-specific context type to manage

**Alternatives considered**:
- Chi-specific context keys: Rejected because Chi uses stdlib context directly
- Chi middleware context (chi.Context): Chi doesn't have a special context type
- Store in request metadata: Not available in stdlib http.Request

### Decision 4: OPTIONS Request Bypass Strategy

**What was chosen**: Check `r.Method == "OPTIONS"` and call `next.ServeHTTP(w, r)` immediately

**Rationale**:
- CORS preflight requests are method-safe (no side effects) per RFC 7231
- Preflight requests don't carry payment information
- Browser sends OPTIONS before actual request with payment header
- Matches spec requirement (FR-022, Edge Cases line 75)

**Alternatives considered**:
- Verify OPTIONS requests: Rejected because CORS preflight has no payment data
- Configurable OPTIONS bypass: Rejected to keep implementation simple (not in spec)
- Check for CORS headers: Rejected as over-engineered (method check is sufficient)

**Implementation**:
```go
if r.Method == "OPTIONS" {
    next.ServeHTTP(w, r)
    return
}
```

### Decision 5: Logging Strategy

**What was chosen**: Use `slog.Default()` for all logging events

**Rationale**:
- Matches Gin middleware pattern (gin/middleware.go:85, 108, 116, etc.)
- Spec requirement FR-019 and FR-023 explicitly require slog.Default()
- Consistent log levels: Info (success), Warn (client errors), Error (service failures)
- No custom logger configuration needed

**Log events**:
- Info: Payment verified, payment settled, requirements enriched
- Warn: No payment header, invalid payment, verification failed, settlement unsuccessful
- Error: Facilitator unreachable, settlement failed

**Alternatives considered**:
- Custom logger injection: Rejected for simplicity (not in spec)
- Framework-specific logging (Chi middleware logger): Chi doesn't provide one
- Silent operation: Rejected for observability requirements

### Decision 6: Error Response Format

**What was chosen**: JSON responses with `x402Version` field matching x402 HTTP transport spec

**Rationale**:
- Consistent with stdlib middleware (http/middleware.go:102, parser.go error responses)
- Spec requirement FR-020 mandates x402Version field
- Standard HTTP status codes: 400 (Bad Request), 402 (Payment Required), 503 (Service Unavailable)

**Response format**:
```go
// 402 Payment Required
{
    "x402Version": 1,
    "error": "Payment required for this resource",
    "accepts": [PaymentRequirement, ...]
}

// 400 Bad Request
{
    "x402Version": 1,
    "error": "Invalid payment header"
}

// 503 Service Unavailable  
{
    "x402Version": 1,
    "error": "Payment verification failed"
}
```

**Alternatives considered**:
- Plain text errors: Rejected for machine readability
- Custom error codes: Not in spec, rejected for simplicity
- HTML error pages: Out of scope (spec line 150)

## Technology Best Practices

### Chi Router Best Practices

**Middleware ordering** (from Chi documentation and examples):
```go
r := chi.NewRouter()
r.Use(middleware.Logger)      // Log all requests
r.Use(middleware.Recoverer)    // Recover from panics
r.Use(NewChiX402Middleware(config))  // Apply payment gating
r.Get("/protected", handler)   // Protected route
```

**Per-route application**:
```go
r.With(NewChiX402Middleware(config)).Get("/paid", handler)
```

**Route group application**:
```go
r.Route("/api", func(r chi.Router) {
    r.Use(NewChiX402Middleware(config))
    r.Get("/resource", handler)
})
```

### Go net/http Best Practices

**Middleware chaining** (Chi pattern matches stdlib):
```go
func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Pre-processing (payment verification)
        
        // Call next handler
        next.ServeHTTP(w, r)
        
        // Post-processing (if needed)
    })
}
```

**Context value access in handlers**:
```go
func handler(w http.ResponseWriter, r *http.Request) {
    payment, ok := r.Context().Value(httpx402.PaymentContextKey).(*httpx402.VerifyResponse)
    if !ok {
        // Handle missing payment info
    }
    // Use payment.Payer, payment.IsValid, etc.
}
```

### Testing Best Practices

**Test structure** (from stdlib middleware_test.go):
1. Table-driven tests for comprehensive scenario coverage
2. httptest.NewRecorder() for response capture
3. httptest.NewRequest() for request construction
4. Mock facilitator for isolated unit tests
5. Test with Chi router instance for integration tests

**Scenarios to test** (from spec User Stories):
- Missing X-PAYMENT header → 402 response
- Invalid X-PAYMENT header (bad base64/JSON) → 400 response
- Valid payment → handler executes
- Facilitator unreachable → 503 response
- Payment insufficient → 402 response
- VerifyOnly mode → no settlement call
- OPTIONS request → bypass payment verification
- Context storage → verify payment details accessible

## Integration Patterns

### Chi Router Integration

Chi middleware integrates at three levels:

1. **Global middleware** (all routes):
```go
r := chi.NewRouter()
r.Use(NewChiX402Middleware(config))
```

2. **Route groups** (subset of routes):
```go
r.Route("/paid", func(r chi.Router) {
    r.Use(NewChiX402Middleware(config))
    r.Get("/resource", handler)
})
```

3. **Inline middleware** (single route):
```go
r.With(NewChiX402Middleware(config)).Get("/paid", handler)
```

### Facilitator Service Integration

**Enrichment pattern** (from stdlib middleware.go:59-66):
```go
enrichedRequirements, err := facilitator.EnrichRequirements(config.PaymentRequirements)
if err != nil {
    slog.Default().Warn("failed to enrich payment requirements", "error", err)
    enrichedRequirements = config.PaymentRequirements  // Graceful degradation
}
```

**Fallback pattern** (from stdlib middleware.go:116-120):
```go
verifyResp, err := facilitator.Verify(payment, requirement)
if err != nil && fallbackFacilitator != nil {
    slog.Default().Warn("primary facilitator failed, trying fallback", "error", err)
    verifyResp, err = fallbackFacilitator.Verify(payment, requirement)
}
```

## Dependencies

### Required Dependencies

All dependencies already present in go.mod:

- **github.com/go-chi/chi/v5** (assumption from spec line 128) - Chi router framework
- **github.com/mark3labs/x402-go** (existing) - x402 core types and definitions
- Go stdlib packages:
  - `net/http` - HTTP server and types
  - `context` - Request context management
  - `encoding/json` - JSON marshaling
  - `encoding/base64` - Header encoding/decoding
  - `log/slog` - Structured logging
  - `time` - Timeout durations

### No New External Dependencies

- Chi is assumed already installed per spec Dependencies section (line 157)
- All other packages are Go stdlib or existing project code
- No additional go get commands needed

## Open Questions

**None remaining.** All technical unknowns resolved:

✅ Middleware signature → `func(http.Handler) http.Handler` with constructor  
✅ Helper function strategy → Share via internal package  
✅ Context storage → stdlib context.WithValue with httpx402.PaymentContextKey  
✅ OPTIONS bypass → Check r.Method == "OPTIONS"  
✅ Logging strategy → slog.Default() for all events  
✅ Error format → JSON with x402Version field  
✅ Testing approach → Table-driven tests adapted from stdlib  
✅ Chi integration → Uses standard net/http patterns

## References

- Chi documentation: https://github.com/go-chi/chi (README section on middleware)
- Stdlib middleware: http/middleware.go (reference implementation)
- Gin middleware: http/gin/middleware.go (constructor pattern)
- PocketBase middleware: http/pocketbase/middleware.go (adapter example)
- x402 spec: specs/006-chi-middleware/spec.md
- Go context package: https://golang.org/pkg/context
- Go slog package: https://golang.org/pkg/log/slog
