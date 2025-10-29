# Research: PocketBase Middleware for x402 Payment Protocol

**Feature**: 005-pocketbase-middleware  
**Date**: 2025-10-29  
**Status**: Phase 0 Complete

## Overview

This document resolves all technical unknowns identified in the implementation plan. Research focused on PocketBase middleware patterns, request/response handling, helper function requirements, and test adaptation strategies.

---

## Research Task 1: PocketBase Middleware Registration Patterns

**Question**: How do PocketBase middlewares register and what is the signature for `hook.Handler[*core.RequestEvent]`?

**Sources**:
- PocketBase documentation: https://pocketbase.io/docs/go-routing/#registering-middlewares
- Existing Gin middleware patterns in http/gin/middleware.go

**Findings**:

### Middleware Signature
PocketBase middlewares use the `hook.Handler[*core.RequestEvent]` type with three fields:
- `Id` (optional): Middleware name for unbinding
- `Priority` (optional): Execution order (default: registration order)
- `Func` (required): Middleware handler function with signature `func(e *core.RequestEvent) error`

### Registration Methods
Two approaches for registering middlewares:

1. **Full control** (when Id/Priority needed):
```go
se.Router.Bind(&hook.Handler[*core.RequestEvent]{
    Id: "x402-payment",
    Func: func(e *core.RequestEvent) error {
        // middleware logic
        return e.Next()
    },
    Priority: -1,
})
```

2. **Convenience** (most common):
```go
se.Router.BindFunc(func(e *core.RequestEvent) error {
    // middleware logic
    return e.Next()
})
```

### Key Patterns
- Middlewares must call `e.Next()` to continue the handler chain
- Returning an error WITHOUT calling `e.Next()` stops the chain
- Works on both routes (`se.Router.GET("/path", handler).BindFunc(...)`) and groups (`g.BindFunc(...)`)
- Global middlewares register on `se.Router` before defining routes

**Decision**: 
- **Rationale**: PocketBase expects factory functions to return `*hook.Handler[*core.RequestEvent]` for direct Bind() usage, OR users can call the factory and use BindFunc() with the returned function. We'll follow the stdlib/Gin pattern of returning a factory function that returns the handler function directly, letting users choose between `Route.BindFunc(NewPocketBaseX402Middleware(config))` or wrapping it themselves.
- **Alternatives Considered**: 
  - Returning `*hook.Handler[*core.RequestEvent]` struct → Rejected: Forces users into Bind() pattern, less flexible
  - Returning both function and struct → Rejected: Unnecessary complexity
- **Implementation**: `func NewPocketBaseX402Middleware(config *http.Config) func(*core.RequestEvent) error`

---

## Research Task 2: PocketBase Request/Response Handling

**Question**: What are the core.RequestEvent methods for headers, JSON responses, and request store?

**Sources**:
- PocketBase documentation: https://pocketbase.io/docs/go-routing/
- stdlib middleware.go patterns for comparison

**Findings**:

### Reading Headers
```go
paymentHeader := e.Request.Header.Get("X-PAYMENT")
```
Same as stdlib `r.Header.Get()` - core.RequestEvent wraps standard `*http.Request`.

### Writing Headers
```go
e.Response.Header().Set("X-PAYMENT-RESPONSE", encoded)
```
Access via `e.Response` (same as http.ResponseWriter).

### JSON Responses
```go
return e.JSON(http.StatusPaymentRequired, data)
```
Convenience method that:
- Sets `Content-Type: application/json`
- Marshals data to JSON
- Writes status code and body
- Returns nil (success) or error

### Request Store (Context Alternative)
```go
// Store value
e.Set("x402_payment", verifyResp)

// Retrieve value (requires type assertion)
verifyResp := e.Get("x402_payment").(*http.VerifyResponse)
```

### Calling Next Handler
```go
return e.Next() // Continue chain
```

### Request URL Construction
```go
scheme := "http"
if e.Request.TLS != nil {
    scheme = "https"
}
resourceURL := scheme + "://" + e.Request.Host + e.Request.RequestURI
```
Same pattern as stdlib middleware - core.RequestEvent.Request is standard `*http.Request`.

**Decision**: 
- **Rationale**: PocketBase provides convenient wrappers (e.JSON, e.Set/Get, e.Next) that simplify code while maintaining http.Request/ResponseWriter compatibility. Using these methods makes the middleware more idiomatic for PocketBase users.
- **Alternatives Considered**: 
  - Direct http.ResponseWriter usage → Rejected: Less idiomatic, requires manual JSON marshaling
  - stdlib context.Context for storage → Rejected: PocketBase convention is e.Set/Get
- **Implementation**: Use e.JSON() for all responses, e.Set/Get for storage, e.Next() for chain continuation

---

## Research Task 3: Helper Function Duplication Requirements

**Question**: Which helpers from stdlib need PocketBase-specific versions?

**Sources**:
- http/middleware.go (stdlib implementation)
- http/gin/middleware.go (Gin adapter showing duplication pattern)

**Findings**:

### Gin Middleware Duplication Pattern
The Gin middleware duplicates 4 helper functions:
1. `parsePaymentHeaderFromRequest(r *http.Request)` - 27 lines
2. `sendPaymentRequiredGin(c *gin.Context, requirements)` - 9 lines
3. `findMatchingRequirementGin(payment, requirements)` - 8 lines
4. `addPaymentResponseHeaderGin(c *gin.Context, settlement)` - 12 lines

### Why Duplication?
Looking at the implementations:
- stdlib helpers are **unexported** (lowercase function names)
- They use framework-specific types (http.ResponseWriter vs gin.Context vs core.RequestEvent)
- Gin chose duplication for **self-contained adapters** (no cross-dependencies)

### stdlib Helper Analysis
From http/middleware.go:
```go
// unexported - cannot be reused
func parsePaymentHeader(r *http.Request) (x402.PaymentPayload, error)
func sendPaymentRequiredWithRequirements(w http.ResponseWriter, requirements)
func findMatchingRequirement(payment, requirements)
func addPaymentResponseHeader(w http.ResponseWriter, settlement)
```

### Required PocketBase Helpers
Following Gin pattern, duplicate all 4 helpers:
1. `parsePaymentHeaderFromRequest` - Parse X-PAYMENT from e.Request
2. `sendPaymentRequiredPocketBase` - Use e.JSON() for 402 response
3. `findMatchingRequirementPocketBase` - Pure logic (identical to Gin)
4. `addPaymentResponseHeaderPocketBase` - Use e.Response.Header().Set()

**Decision**: 
- **Rationale**: Duplicate all 4 helpers following established Gin pattern. Self-contained adapters prevent cross-package dependencies and make each framework integration independent. The parsePaymentHeader logic (base64 decode + JSON unmarshal + version check) is only ~20 lines and worth duplicating for independence.
- **Alternatives Considered**: 
  - Export stdlib helpers for reuse → Rejected: Breaks encapsulation, mixes framework concerns
  - Shared helper package (http/internal/helpers) → Rejected: Over-engineering for 4 simple functions
  - Generic helpers with interface{} → Rejected: Loses type safety, unclear which method to call
- **Implementation**: Duplicate all 4 helpers with "PocketBase" suffix, adapting framework-specific calls (e.JSON, e.Response)

---

## Research Task 4: Test Adaptation Strategy

**Question**: How do we adapt http/middleware_test.go for core.RequestEvent?

**Sources**:
- http/middleware_test.go (stdlib tests)
- http/gin/middleware_test.go (if exists - check repo)
- PocketBase testing patterns

**Findings**:

### stdlib Middleware Test Structure
From http/middleware_test.go analysis:
- Uses `httptest.NewRequest()` and `httptest.NewRecorder()` for test requests/responses
- Table-driven tests with test cases: name, setupFunc, wantStatus, wantBodyContains
- Tests cover: missing payment (402), invalid payment (400), valid payment (200), facilitator errors (503)

### PocketBase Testing Challenge
**Problem**: PocketBase uses `core.RequestEvent` which wraps `*http.Request` and `core.ResponseWriter`. We need to:
1. Create mock `core.RequestEvent` instances
2. Mock `e.Next()` behavior
3. Capture `e.JSON()` responses

### Gin Test Adaptation (if exists)
Checking http/gin/middleware_test.go for reference patterns...
Based on the repo structure, Gin tests likely use:
```go
// Gin test recorder
w := httptest.NewRecorder()
c, _ := gin.CreateTestContext(w)
c.Request = httptest.NewRequest("GET", "/test", nil)
```

### PocketBase Test Strategy
Since PocketBase's `core.RequestEvent` wraps standard types, we can:

**Option 1**: Use httptest + manual RequestEvent construction
```go
// Create standard request/response
req := httptest.NewRequest("GET", "/test", nil)
w := httptest.NewRecorder()

// Wrap in PocketBase types
e := &core.RequestEvent{
    Request: req,
    Response: w,
}

// Call middleware
err := middleware(e)
```

**Option 2**: Mock core.ServeEvent in test
```go
app := core.NewBaseApp(...)
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    // Register test middleware
    se.Router.GET("/test", handler).Bind(middleware)
    return se.Next()
})
```

**Comparison**:
- Option 1: Simple, direct, mirrors stdlib tests
- Option 2: More realistic but requires full PocketBase app setup

### Test Cases to Adapt
From stdlib middleware_test.go, port these scenarios:
1. **No payment header** → 402 with PaymentRequirementsResponse
2. **Invalid base64 in X-PAYMENT** → 400 Bad Request
3. **Invalid JSON in X-PAYMENT** → 400 Bad Request
4. **Valid payment** → 200 OK, e.Next() called
5. **Facilitator verify failure** → 503 Service Unavailable
6. **Verify-only mode** → No settlement call

**Decision**: 
- **Rationale**: Use Option 1 (httptest + manual RequestEvent). It's simpler, mirrors stdlib tests, and avoids full PocketBase app setup overhead. We can construct minimal RequestEvent instances with just Request and Response fields, which is sufficient for middleware testing.
- **Alternatives Considered**: 
  - Full PocketBase app in tests → Rejected: Overkill, slow, unclear what we're testing
  - Interface for RequestEvent → Rejected: Over-engineering for test mocking
- **Implementation**: Create helper function `newTestRequestEvent(method, url, headers) *core.RequestEvent` that wraps httptest types. Adapt stdlib test table structure with PocketBase-specific assertions.

---

## Summary of Resolved Unknowns

| Unknown | Resolution |
|---------|-----------|
| Middleware signature | `func NewPocketBaseX402Middleware(config *http.Config) func(*core.RequestEvent) error` |
| Registration pattern | Users call `Route.BindFunc(NewPocketBaseX402Middleware(config))` or `Route.Bind(&hook.Handler{Func: ...})` |
| Request/response handling | Use e.JSON() for responses, e.Set/Get for storage, e.Request for headers |
| Helper duplication | Duplicate all 4 helpers with "PocketBase" suffix following Gin pattern |
| Test strategy | httptest + manual RequestEvent construction, adapt stdlib test table |

---

## Implementation Readiness Checklist

- [x] Middleware function signature defined
- [x] Registration pattern confirmed (BindFunc)
- [x] Request/response methods documented (e.JSON, e.Set/Get, e.Next)
- [x] Helper function list finalized (4 duplications)
- [x] Test adaptation strategy defined (httptest wrapper)
- [x] All NEEDS CLARIFICATION items resolved
- [x] No blocking unknowns remaining

**Status**: ✅ Ready for Phase 1 (Design & Contracts)

---

## References

1. PocketBase Routing Documentation: https://pocketbase.io/docs/go-routing/#registering-middlewares
2. stdlib middleware.go: /home/space_cowboy/Workspace/x402-go/http/middleware.go
3. Gin middleware.go: /home/space_cowboy/Workspace/x402-go/http/gin/middleware.go
4. PocketBase core package: github.com/pocketbase/pocketbase/core
