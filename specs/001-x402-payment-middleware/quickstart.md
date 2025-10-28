# Quickstart: x402 Payment Middleware

## Installation

```bash
go get github.com/mark3labs/x402-go
```

## Basic Usage

### 1. Simple Payment-Protected Endpoint

```go
package main

import (
    "log"
    "net/http"
    
    "github.com/mark3labs/x402-go/x402"
    x402http "github.com/mark3labs/x402-go/http"
)

func main() {
    // Configure payment requirements
    config := &x402http.Config{
        FacilitatorURL: "https://facilitator.x402.com",
        PaymentRequirements: []x402.PaymentRequirement{
            {
                Scheme:             "exact",
                Network:            "base-sepolia",
                MaxAmountRequired:  "10000", // 0.01 USDC (6 decimals)
                Asset:              "0x036CbD53842c5426634e7929541eC2318f3dCF7e", // USDC
                PayTo:              "0xYourWalletAddress",
                Resource:           "https://api.example.com/premium",
                Description:        "Premium API Access",
                MaxTimeoutSeconds:  60,
            },
        },
    }
    
    // Create middleware
    paymentMiddleware := x402http.NewX402Middleware(config)
    
    // Protected handler
    premiumHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Premium content!"))
    })
    
    // Apply middleware
    http.Handle("/premium", paymentMiddleware(premiumHandler))
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 2. Multi-Chain Configuration

```go
// Accept payments on multiple chains
// Note: feePayer for SVM chains is automatically fetched from the facilitator
config := &x402http.Config{
    FacilitatorURL: "https://facilitator.x402.com",
    PaymentRequirements: []x402.PaymentRequirement{
        {
            Scheme:             "exact",
            Network:            "base",
            MaxAmountRequired:  "1000000", // 1 USDC
            Asset:              "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // Base USDC
            PayTo:              "0xYourWalletAddress",
            Resource:           "https://api.example.com/data",
            Description:        "Data Access - Base",
            MaxTimeoutSeconds:  60,
        },
        {
            Scheme:             "exact", 
            Network:            "solana",
            MaxAmountRequired:  "1000000", // 1 USDC
            Asset:              "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // Solana USDC
            PayTo:              "YourSolanaWalletAddress",
            Resource:           "https://api.example.com/data",
            Description:        "Data Access - Solana",
            MaxTimeoutSeconds:  60,
            // No need to specify Extra.feePayer - automatically populated from facilitator
        },
    },
}
```

### 3. Route-Specific Pricing

```go
// Different prices for different endpoints
basicConfig := &x402http.Config{
    FacilitatorURL: "https://facilitator.x402.com",
    PaymentRequirements: []x402.PaymentRequirement{
        {
            Scheme:             "exact",
            Network:            "base-sepolia", 
            MaxAmountRequired:  "1000", // 0.001 USDC
            Asset:              "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
            PayTo:              "0xYourWalletAddress",
            Resource:           "https://api.example.com/basic",
            Description:        "Basic API Access",
            MaxTimeoutSeconds:  60,
        },
    },
}

premiumConfig := &x402http.Config{
    FacilitatorURL: "https://facilitator.x402.com",
    PaymentRequirements: []x402.PaymentRequirement{
        {
            Scheme:             "exact",
            Network:            "base-sepolia",
            MaxAmountRequired:  "10000", // 0.01 USDC  
            Asset:              "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
            PayTo:              "0xYourWalletAddress",
            Resource:           "https://api.example.com/premium",
            Description:        "Premium API Access",
            MaxTimeoutSeconds:  60,
        },
    },
}

// Create different middleware instances
basicMiddleware := x402http.NewX402Middleware(basicConfig)
premiumMiddleware := x402http.NewX402Middleware(premiumConfig)

// Apply to different routes
http.Handle("/basic", basicMiddleware(basicHandler))
http.Handle("/premium", premiumMiddleware(premiumHandler))
```

### 4. Verification-Only Mode

```go
// Verify payments without settling
config := &x402http.Config{
    FacilitatorURL: "https://facilitator.x402.com",
    VerifyOnly:     true, // Only verify, don't settle
    PaymentRequirements: []x402.PaymentRequirement{
        // ... payment requirements
    },
}

middleware := x402http.NewX402Middleware(config)

handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Payment verified but not settled
    // You can access payment info from context if needed
    payment := r.Context().Value(x402http.PaymentContextKey)
    // Custom business logic here
    w.Write([]byte("Payment verified!"))
})
```

### 5. Fallback Facilitator

```go
// Configure fallback for high availability
config := &x402http.Config{
    FacilitatorURL:         "https://primary.facilitator.com",
    FallbackFacilitatorURL: "https://fallback.facilitator.com",
    PaymentRequirements: []x402.PaymentRequirement{
        // ... payment requirements
    },
}
```

## Client Examples

### Making a Payment Request (JavaScript)

```javascript
// 1. First request to get payment requirements
const response = await fetch('https://api.example.com/premium');
if (response.status === 402) {
    const requirements = await response.json();
    
    // 2. Create payment authorization (using web3 library)
    const payment = await createPayment(requirements.accepts[0]);
    
    // 3. Encode payment as base64
    const paymentHeader = btoa(JSON.stringify(payment));
    
    // 4. Retry with payment
    const paidResponse = await fetch('https://api.example.com/premium', {
        headers: {
            'X-PAYMENT': paymentHeader
        }
    });
    
    // 5. Check settlement response
    const settlementHeader = paidResponse.headers.get('X-PAYMENT-RESPONSE');
    if (settlementHeader) {
        const settlement = JSON.parse(atob(settlementHeader));
        console.log('Payment settled:', settlement.transaction);
    }
}
```

### Making a Payment Request (Go)

```go
client := &http.Client{}

// 1. Get payment requirements
resp, _ := client.Get("https://api.example.com/premium")
if resp.StatusCode == 402 {
    var requirements x402.PaymentRequirementsResponse
    json.NewDecoder(resp.Body).Decode(&requirements)
    
    // 2. Create payment (implementation depends on wallet)
    payment := createPayment(requirements.Accepts[0])
    
    // 3. Encode payment
    paymentJSON, _ := json.Marshal(payment)
    paymentHeader := base64.StdEncoding.EncodeToString(paymentJSON)
    
    // 4. Make request with payment
    req, _ := http.NewRequest("GET", "https://api.example.com/premium", nil)
    req.Header.Set("X-PAYMENT", paymentHeader)
    
    paidResp, _ := client.Do(req)
    
    // 5. Check settlement
    if settlement := paidResp.Header.Get("X-PAYMENT-RESPONSE"); settlement != "" {
        decoded, _ := base64.StdEncoding.DecodeString(settlement)
        var result x402.SettlementResponse
        json.Unmarshal(decoded, &result)
        fmt.Printf("Payment settled: %s\n", result.Transaction)
    }
}
```

## Testing

### Unit Testing Your Handlers

```go
func TestPaymentProtectedEndpoint(t *testing.T) {
    // Create test config
    config := &x402http.Config{
        FacilitatorURL: "http://mock-facilitator",
        PaymentRequirements: []x402.PaymentRequirement{
            // test requirements
        },
    }
    
    middleware := x402http.NewX402Middleware(config)
    handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    
    // Test without payment
    req := httptest.NewRequest("GET", "/test", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    
    if rec.Code != http.StatusPaymentRequired {
        t.Errorf("Expected 402, got %d", rec.Code)
    }
    
    // Verify requirements in response
    var requirements x402.PaymentRequirementsResponse
    json.NewDecoder(rec.Body).Decode(&requirements)
    if len(requirements.Accepts) == 0 {
        t.Error("No payment requirements returned")
    }
}
```

## Common Patterns

### Logging Payments

```go
middleware := x402http.NewX402Middleware(config)

loggingMiddleware := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Log after payment processing
        if payment := r.Context().Value(x402http.PaymentContextKey); payment != nil {
            log.Printf("Payment received from: %s", payment.Payer)
        }
        next.ServeHTTP(w, r)
    })
}

// Chain middleware
http.Handle("/api", middleware(loggingMiddleware(apiHandler)))
```

### Custom Error Handling

```go
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Check if payment failed
    if err := r.Context().Value(x402http.PaymentErrorKey); err != nil {
        // Custom error response
        http.Error(w, "Payment required for this resource", 402)
        return
    }
    // Normal handling
})
```

## Troubleshooting

### Payment Not Being Accepted

1. Check that the X-PAYMENT header is properly base64 encoded
2. Verify the payment amount meets the requirement
3. Ensure the network and asset match the requirements
4. Check facilitator service is reachable
5. Verify wallet has sufficient balance

### 503 Service Unavailable

- Primary facilitator is down
- Configure fallback facilitator URL
- Check network connectivity

### 400 Bad Request  

- Malformed X-PAYMENT header
- Invalid JSON in payment payload
- Missing required fields

## Next Steps

- Review the [full API documentation](https://pkg.go.dev/github.com/mark3labs/x402-go)
- Explore [example applications](https://github.com/mark3labs/x402-go/tree/main/examples)
- Learn about [facilitator setup](https://docs.x402.com/facilitator)
- Join the [x402 community](https://discord.gg/x402)