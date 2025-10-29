# Quickstart: Chi Middleware for x402 Payment Protocol

**Feature**: Chi Middleware Adapter  
**Date**: 2025-10-29  
**Target Audience**: Developers using Chi router who want to add x402 payment gating

## Overview

This guide shows you how to protect your Chi HTTP routes with x402 payment gating in under 5 minutes. The middleware verifies payments with the facilitator service and makes payment details available to your handlers.

## Prerequisites

- Go 1.25.1 or later
- Chi router installed: `go get -u github.com/go-chi/chi/v5`
- x402-go package: `go get -u github.com/mark3labs/x402-go`
- Basic familiarity with Chi routing

## Basic Setup (30 seconds)

### 1. Import packages

```go
package main

import (
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/mark3labs/x402-go"
    httpx402 "github.com/mark3labs/x402-go/http"
    chix402 "github.com/mark3labs/x402-go/http/chi"
)
```

### 2. Configure middleware

```go
func main() {
    // Configure payment requirements
    config := &httpx402.Config{
        FacilitatorURL: "https://api.x402.coinbase.com",
        PaymentRequirements: []x402.PaymentRequirement{{
            Scheme:            "exact",
            Network:           "base-sepolia",
            MaxAmountRequired: "10000",  // 0.01 USDC (6 decimals)
            Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",  // USDC on Base Sepolia
            PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",  // Your wallet
            MaxTimeoutSeconds: 300,
        }},
    }

    // Create router and apply middleware
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(chix402.NewChiX402Middleware(config))
    
    // Protected route
    r.Get("/protected", protectedHandler)
    
    http.ListenAndServe(":3000", r)
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
    // Access payment details from context
    payment := r.Context().Value(httpx402.PaymentContextKey).(*httpx402.VerifyResponse)
    w.Write([]byte("Access granted! Payer: " + payment.Payer))
}
```

### 3. Test it

```bash
# Start server
go run main.go

# Test without payment (gets 402)
curl http://localhost:3000/protected

# Test with payment (use x402 client library)
# See client examples in the x402-go repository
```

## Common Patterns

### Pattern 1: Apply to Route Group

Protect only specific routes:

```go
r := chi.NewRouter()

// Public routes (no middleware)
r.Get("/", homeHandler)
r.Get("/about", aboutHandler)

// Protected route group
r.Route("/premium", func(r chi.Router) {
    r.Use(chix402.NewChiX402Middleware(config))
    r.Get("/feature1", premiumFeature1)
    r.Get("/feature2", premiumFeature2)
})
```

### Pattern 2: Apply to Single Route

Protect just one route with inline middleware:

```go
r := chi.NewRouter()
r.Get("/free", freeHandler)
r.With(chix402.NewChiX402Middleware(config)).Get("/paid", paidHandler)
```

### Pattern 3: Access Payment Details

Get payer information in your handler:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Get payment from context
    payment, ok := r.Context().Value(httpx402.PaymentContextKey).(*httpx402.VerifyResponse)
    if !ok {
        http.Error(w, "Payment info not found", http.StatusInternalServerError)
        return
    }
    
    // Use payment details
    log.Printf("Request from payer: %s", payment.Payer)
    log.Printf("Payment valid: %v", payment.IsValid)
    
    // Your handler logic here
    w.Write([]byte("Success"))
}
```

### Pattern 4: Verify-Only Mode (Testing)

Skip settlement during development:

```go
config := &httpx402.Config{
    FacilitatorURL: "https://api.x402.coinbase.com",
    VerifyOnly:     true,  // Don't settle payments
    PaymentRequirements: []x402.PaymentRequirement{{
        // ... requirements
    }},
}
```

### Pattern 5: Fallback Facilitator

Add redundancy with a backup facilitator:

```go
config := &httpx402.Config{
    FacilitatorURL:         "https://api.x402.coinbase.com",
    FallbackFacilitatorURL: "https://backup.x402.example.com",
    PaymentRequirements: []x402.PaymentRequirement{{
        // ... requirements
    }},
}
```

## Complete Example

Here's a full working example with multiple routes and payment details:

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/mark3labs/x402-go"
    httpx402 "github.com/mark3labs/x402-go/http"
    chix402 "github.com/mark3labs/x402-go/http/chi"
)

func main() {
    // Payment config
    config := &httpx402.Config{
        FacilitatorURL: "https://api.x402.coinbase.com",
        PaymentRequirements: []x402.PaymentRequirement{{
            Scheme:            "exact",
            Network:           "base-sepolia",
            MaxAmountRequired: "10000",
            Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
            PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
            MaxTimeoutSeconds: 300,
            Description:       "Payment for premium API access",
        }},
    }

    r := chi.NewRouter()
    
    // Standard middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    // Public routes
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Welcome! Visit /premium for paid content"))
    })

    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    })

    // Protected routes
    r.Route("/premium", func(r chi.Router) {
        r.Use(chix402.NewChiX402Middleware(config))
        
        r.Get("/data", func(w http.ResponseWriter, r *http.Request) {
            payment := r.Context().Value(httpx402.PaymentContextKey).(*httpx402.VerifyResponse)
            
            response := map[string]interface{}{
                "message": "Premium data access granted",
                "payer":   payment.Payer,
                "data":    []string{"secret1", "secret2", "secret3"},
            }
            
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(response)
        })
        
        r.Get("/analytics", func(w http.ResponseWriter, r *http.Request) {
            payment := r.Context().Value(httpx402.PaymentContextKey).(*httpx402.VerifyResponse)
            
            response := map[string]interface{}{
                "message": "Premium analytics access granted",
                "payer":   payment.Payer,
                "metrics": map[string]int{"views": 1000, "clicks": 250},
            }
            
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(response)
        })
    })

    log.Println("Server starting on :3000")
    http.ListenAndServe(":3000", r)
}
```

## Configuration Options

### Config Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `FacilitatorURL` | string | Yes | Primary facilitator endpoint URL |
| `FallbackFacilitatorURL` | string | No | Backup facilitator endpoint URL |
| `PaymentRequirements` | []PaymentRequirement | Yes | List of accepted payment methods |
| `VerifyOnly` | bool | No | Skip settlement (default: false) |

### PaymentRequirement Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Scheme` | string | Yes | Payment scheme (e.g., "exact") |
| `Network` | string | Yes | Blockchain network (e.g., "base-sepolia") |
| `MaxAmountRequired` | string | Yes | Amount in atomic units (e.g., "10000") |
| `Asset` | string | Yes | Token contract address |
| `PayTo` | string | Yes | Recipient wallet address |
| `MaxTimeoutSeconds` | int | Yes | Payment timeout in seconds |
| `Description` | string | No | Human-readable description |

## Network Configuration

### Testnet (Base Sepolia)

```go
x402.PaymentRequirement{
    Network: "base-sepolia",
    Asset:   "0x036CbD53842c5426634e7929541eC2318f3dCF7e",  // USDC testnet
    PayTo:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",  // Your testnet wallet
}
```

### Mainnet (Base)

```go
x402.PaymentRequirement{
    Network: "base",
    Asset:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",  // USDC mainnet
    PayTo:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",  // Your mainnet wallet
}
```

## Error Responses

The middleware returns these standard HTTP responses:

### 402 Payment Required

Returned when payment is missing, invalid, or insufficient:

```json
{
  "x402Version": 1,
  "error": "Payment required for this resource",
  "accepts": [{
    "Scheme": "exact",
    "Network": "base-sepolia",
    "MaxAmountRequired": "10000",
    "Asset": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
    "PayTo": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    "MaxTimeoutSeconds": 300,
    "Resource": "https://api.example.com/premium/data",
    "Description": "Payment for premium API access"
  }]
}
```

### 400 Bad Request

Returned when X-PAYMENT header is malformed:

```json
{
  "x402Version": 1,
  "error": "Invalid payment header"
}
```

### 503 Service Unavailable

Returned when facilitator is unreachable or processing fails:

```json
{
  "x402Version": 1,
  "error": "Payment verification failed"
}
```

## CORS Support

OPTIONS requests automatically bypass payment verification to support CORS preflight:

```go
// No configuration needed - automatic
// OPTIONS requests will skip middleware and proceed to next handler
```

## Logging

The middleware logs to `slog.Default()` at different levels:

- **Info**: Payment verified, payment settled, requirements enriched
- **Warn**: Missing payment, invalid payment, verification failed
- **Error**: Facilitator unreachable, settlement failed

Configure slog before starting your server:

```go
import "log/slog"

func main() {
    // Configure structured logging
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)
    
    // ... rest of your app
}
```

## Timeouts

The middleware uses hardcoded timeouts for facilitator calls:

- **Verification**: 5 seconds (quick check)
- **Settlement**: 60 seconds (blockchain transaction)

These values match the stdlib middleware and cannot be customized.

## Next Steps

1. **Review the spec**: See [spec.md](./spec.md) for complete requirements
2. **Check the contracts**: See [contracts/chi-middleware-api.yaml](./contracts/chi-middleware-api.yaml) for API details
3. **Understand the data model**: See [data-model.md](./data-model.md) for entity relationships
4. **Explore examples**: Check `examples/chi/` directory for more use cases
5. **Run tests**: Use `go test -race ./http/chi/...` to verify implementation

## Troubleshooting

### Problem: 503 responses for all requests

**Solution**: Check facilitator URL is reachable and correct. Enable debug logging to see facilitator responses.

### Problem: 402 responses even with valid payment

**Solution**: Verify payment header is base64-encoded JSON, X402Version is 1, and scheme/network match configuration.

### Problem: Payment context is nil in handler

**Solution**: Ensure middleware is applied before the route handler in Chi middleware stack.

### Problem: CORS preflight fails

**Solution**: This should work automatically. If failing, check that middleware is not interfering with OPTIONS responses.

## Support

- GitHub Issues: https://github.com/mark3labs/x402-go/issues
- Documentation: https://go-chi.io (Chi router)
- x402 Protocol: https://x402.org

## License

MIT License - See repository LICENSE file for details.
