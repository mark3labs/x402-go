# Quickstart: PocketBase Middleware for x402 Payment Protocol

**Feature**: 005-pocketbase-middleware  
**Date**: 2025-10-29  
**Audience**: Developers integrating x402 payment gating into PocketBase applications

## Overview

This guide shows how to integrate x402 payment gating into your PocketBase application in under 5 minutes. The middleware protects your custom API endpoints by requiring cryptographic payment verification before granting access.

---

## Prerequisites

- Go 1.25.1 or later
- PocketBase application (github.com/pocketbase/pocketbase)
- Access to a facilitator service (default: Coinbase facilitator)
- Wallet addresses for receiving payments

---

## Installation

Add the x402-go package to your PocketBase application:

```bash
go get github.com/mark3labs/x402-go
```

---

## Basic Usage

### Step 1: Import Packages

```go
package main

import (
    "log"
    "net/http"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
    
    "github.com/mark3labs/x402-go"
    httpx402 "github.com/mark3labs/x402-go/http"
    pbx402 "github.com/mark3labs/x402-go/http/pocketbase"
)
```

### Step 2: Configure Middleware

```go
func main() {
    app := pocketbase.New()

    // Configure payment requirements
    config := &httpx402.Config{
        FacilitatorURL: "https://api.x402.coinbase.com",
        PaymentRequirements: []x402.PaymentRequirement{{
            Scheme:            "exact",
            Network:           "base-sepolia", // testnet
            MaxAmountRequired: "10000",        // 0.01 USDC (USDC has 6 decimals: 1 USDC = 1,000,000 atomic units)
            Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e", // USDC on base-sepolia
            PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", // your wallet
            MaxTimeoutSeconds: 300,
        }},
    }

    // Register middleware
    app.OnServe().BindFunc(func(se *core.ServeEvent) error {
        // Create middleware
        middleware := pbx402.NewPocketBaseX402Middleware(config)

        // Protect a single route
        se.Router.GET("/api/premium/data", func(e *core.RequestEvent) error {
            // Access payment details (optional)
            payment := e.Get("x402_payment").(*httpx402.VerifyResponse)
            
            return e.JSON(http.StatusOK, map[string]any{
                "data":  "Premium content here",
                "payer": payment.Payer,
            })
        }).BindFunc(middleware)

        return se.Next()
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### Step 3: Test the Endpoint

**Without payment**:
```bash
curl http://localhost:8090/api/premium/data
```

**Response (402 Payment Required)**:
```json
{
  "x402Version": 1,
  "error": "Payment required for this resource",
  "accepts": [{
    "scheme": "exact",
    "network": "base-sepolia",
    "maxAmountRequired": "10000",
    "asset": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
    "payTo": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    "maxTimeoutSeconds": 300,
    "resource": "http://localhost:8090/api/premium/data",
    "description": "Payment required for /api/premium/data"
  }]
}
```

**With valid payment**:
```bash
curl -H "X-PAYMENT: eyJ4NDAyVmVyc2lvbiI6MSwic2NoZW1lIjoiZXhhY3QiLC4uLn0=" \
     http://localhost:8090/api/premium/data
```

**Response (200 OK)**:
```json
{
  "data": "Premium content here",
  "payer": "0x1234567890abcdef1234567890abcdef12345678"
}
```

---

## Advanced Usage

### Protecting Multiple Routes with a Group

```go
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    middleware := pbx402.NewPocketBaseX402Middleware(config)

    // Create a protected group
    premiumGroup := se.Router.Group("/api/premium")
    premiumGroup.BindFunc(middleware) // Apply to all routes in group

    // All these routes require payment
    premiumGroup.GET("/data", dataHandler)
    premiumGroup.GET("/reports", reportsHandler)
    premiumGroup.POST("/analytics", analyticsHandler)

    return se.Next()
})
```

### Verify-Only Mode (Testing)

For testing environments where you want to verify payments without executing on-chain settlement:

```go
config := &httpx402.Config{
    FacilitatorURL: "https://api.x402.coinbase.com",
    PaymentRequirements: []x402.PaymentRequirement{{
        Scheme:            "exact",
        Network:           "base-sepolia",
        MaxAmountRequired: "10000",
        Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
        PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
        MaxTimeoutSeconds: 300,
    }},
    VerifyOnly: true, // Skip settlement
}
```

### Fallback Facilitator

For high availability, configure a backup facilitator:

```go
config := &httpx402.Config{
    FacilitatorURL:         "https://api.x402.coinbase.com",
    FallbackFacilitatorURL: "https://backup.x402.example.com", // fallback
    PaymentRequirements: []x402.PaymentRequirement{{
        // ... requirements
    }},
}
```

### Mainnet Configuration

Switch to mainnet for production:

```go
config := &httpx402.Config{
    FacilitatorURL: "https://api.x402.coinbase.com",
    PaymentRequirements: []x402.PaymentRequirement{{
        Scheme:            "exact",
        Network:           "base",              // mainnet
        MaxAmountRequired: "1000000",           // 1 USDC (6 decimals)
        Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // USDC on base mainnet
        PayTo:             "0xYourProductionWallet",
        MaxTimeoutSeconds: 300,
    }},
}
```

### Multiple Payment Options

Support multiple networks and payment amounts:

```go
config := &httpx402.Config{
    FacilitatorURL: "https://api.x402.coinbase.com",
    PaymentRequirements: []x402.PaymentRequirement{
        {
            Scheme:            "exact",
            Network:           "base-sepolia",
            MaxAmountRequired: "10000", // 0.01 USDC on testnet
            Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
            PayTo:             "0xYourWallet",
            MaxTimeoutSeconds: 300,
            Description:       "Testnet payment",
        },
        {
            Scheme:            "exact",
            Network:           "solana-devnet",
            MaxAmountRequired: "10000", // 0.01 USDC on Solana devnet
            Asset:             "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU", // USDC mint
            PayTo:             "YourSolanaWallet",
            MaxTimeoutSeconds: 300,
            Description:       "Solana testnet payment",
        },
    },
}
```

---

## Accessing Payment Details in Handlers

After successful payment verification, access payment information:

```go
se.Router.GET("/api/premium/data", func(e *core.RequestEvent) error {
    // Retrieve payment details from request store
    paymentData := e.Get("x402_payment")
    if paymentData == nil {
        return e.InternalServerError("Payment data missing", nil)
    }

    verifyResp := paymentData.(*httpx402.VerifyResponse)
    
    // Access payment fields
    payer := verifyResp.Payer           // Wallet address
    isValid := verifyResp.IsValid       // Always true in handler
    reason := verifyResp.InvalidReason  // Empty if valid

    // Use payment info in your logic
    log.Printf("Payment from %s processed", payer)

    return e.JSON(http.StatusOK, map[string]any{
        "message": "Premium content",
        "payer":   payer,
    })
}).BindFunc(middleware)
```

---

## Error Handling

The middleware automatically handles errors and returns appropriate HTTP status codes:

| Error Scenario | Status Code | Response |
|----------------|-------------|----------|
| Missing X-PAYMENT | 402 | PaymentRequirementsResponse with accepts array |
| Invalid base64/JSON | 400 | `{"x402Version": 1, "error": "Invalid payment header"}` |
| Payment verification fails | 402 | PaymentRequirementsResponse |
| Facilitator unreachable | 503 | `{"x402Version": 1, "error": "Payment verification failed"}` |
| Settlement fails | 503 | `{"x402Version": 1, "error": "Payment settlement failed"}` |

---

## Integration with PocketBase Auth

Combine x402 payments with PocketBase authentication:

```go
se.Router.GET("/api/premium/data", func(e *core.RequestEvent) error {
    // Check PocketBase auth
    if e.Auth == nil {
        return e.UnauthorizedError("Authentication required", nil)
    }

    // Access both auth and payment
    user := e.Auth
    payment := e.Get("x402_payment").(*httpx402.VerifyResponse)

    // Log payment for user
    log.Printf("User %s paid from %s", user.Email(), payment.Payer)

    return e.JSON(http.StatusOK, map[string]any{
        "user":  user.Email(),
        "payer": payment.Payer,
        "data":  "Premium content",
    })
}).Bind(
    apis.RequireAuth(),                // Require auth first
    pbx402.NewPocketBaseX402Middleware(config), // Then require payment
)
```

---

## Testing

### Unit Tests

Test your protected handlers using PocketBase's test utilities:

```go
func TestPremiumEndpoint(t *testing.T) {
    app := pocketbase.NewWithConfig(...)
    
    // Setup test middleware
    config := &httpx402.Config{...}
    
    // Create test request with payment
    req := httptest.NewRequest("GET", "/api/premium/data", nil)
    req.Header.Set("X-PAYMENT", validPaymentHeader)
    
    w := httptest.NewRecorder()
    
    // Test endpoint
    // ... assertions
}
```

### Integration Tests

Test the full payment flow with a test facilitator:

```go
config := &httpx402.Config{
    FacilitatorURL: "http://localhost:8091", // test facilitator
    PaymentRequirements: []x402.PaymentRequirement{{
        Scheme:            "exact",
        Network:           "base-sepolia",
        MaxAmountRequired: "10000",
        Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
        PayTo:             "0xTestWallet",
        MaxTimeoutSeconds: 300,
    }},
}
```

---

## Production Checklist

Before deploying to production:

- [ ] Switch from testnet to mainnet networks
- [ ] Update USDC contract addresses for mainnet
- [ ] Configure production wallet addresses for `PayTo`
- [ ] Set up facilitator URL (or use Coinbase facilitator)
- [ ] Configure fallback facilitator for high availability
- [ ] Set appropriate `MaxAmountRequired` values
- [ ] Test payment flow end-to-end on testnet
- [ ] Monitor facilitator response times
- [ ] Set up logging for payment events
- [ ] Document payment requirements for API consumers

---

## Troubleshooting

### Problem: 402 Response Despite Valid Payment

**Cause**: Payment scheme/network mismatch  
**Solution**: Ensure client payment matches middleware configuration:
```go
// Middleware expects
Network: "base-sepolia"
Scheme: "exact"

// Client must send
{
  "network": "base-sepolia",
  "scheme": "exact",
  ...
}
```

### Problem: 503 Service Unavailable

**Cause**: Facilitator unreachable or timeout  
**Solution**: 
1. Check facilitator URL is correct
2. Verify network connectivity
3. Configure fallback facilitator
4. Check facilitator service status

### Problem: Payment Details Not in Handler

**Cause**: Accessing `e.Get("x402_payment")` returns nil  
**Solution**: Ensure middleware is registered BEFORE the handler:
```go
// Correct order
se.Router.GET("/path", handler).BindFunc(middleware)

// Or with Bind
se.Router.GET("/path", handler).Bind(&hook.Handler[*core.RequestEvent]{
    Func: middleware,
    Priority: -1, // Run before handler
})
```

---

## Next Steps

- Review the [spec.md](./spec.md) for detailed requirements
- Check [data-model.md](./data-model.md) for data flow details
- See [contracts/pocketbase-middleware-api.yaml](./contracts/pocketbase-middleware-api.yaml) for API reference
- Explore example code in `examples/pocketbase/`

---

## Support

For issues or questions:
- GitHub Issues: https://github.com/mark3labs/x402-go/issues
- Documentation: https://github.com/mark3labs/x402-go
