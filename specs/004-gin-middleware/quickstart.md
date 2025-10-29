# Quickstart Guide: x402 Gin Middleware

**Date**: 2025-10-29  
**Feature**: Gin Middleware for x402 Payment Protocol

## Overview

The x402 Gin middleware provides payment gating for Gin web applications using the x402 payment protocol. This guide shows you how to integrate the middleware into your Gin application and protect endpoints with payment requirements.

## Prerequisites

- Go 1.25.1 or later
- Gin web framework (`github.com/gin-gonic/gin`)
- x402-go package (`github.com/mark3labs/x402-go`)

## Installation

Add the x402-go package to your project:

```bash
go get github.com/mark3labs/x402-go
```

Import the required packages in your application:

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/mark3labs/x402-go/http/gin"
    "math/big"
)
```

## Basic Usage

### Simple Payment Protection

Protect an endpoint with a basic USDC payment requirement:

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/mark3labs/x402-go/http/gin"
    "math/big"
)

func main() {
    r := gin.Default()
    
    // Protect endpoint with 0.01 USDC payment
    r.Use(gin.PaymentMiddleware(
        big.NewFloat(0.01), // $0.01 in USDC
        "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", // Recipient address
    ))
    
    r.GET("/protected", func(c *gin.Context) {
        // Access payment information from context
        if paymentInfo, exists := c.Get("x402_payment"); exists {
            // Payment was verified successfully
            c.JSON(200, gin.H{
                "message": "Access granted with valid payment",
                "status": "success",
            })
        } else {
            c.JSON(500, gin.H{"error": "Payment information not found"})
        }
    })
    
    r.Run(":8080")
}
```

### Accessing Payment Information

Retrieve payment details in your protected handlers:

```go
r.GET("/user-data", func(c *gin.Context) {
    // Get payment verification response from context
    paymentInfo, exists := c.Get("x402_payment")
    if !exists {
        c.JSON(500, gin.H{"error": "Payment verification failed"})
        return
    }
    
    // Type assertion to get verification response
    verifyResp, ok := paymentInfo.(*gin.VerifyResponse)
    if !ok {
        c.JSON(500, gin.H{"error": "Invalid payment information"})
        return
    }
    
    // Use payment information
    c.JSON(200, gin.H{
        "message": "Welcome!",
        "payer": verifyResp.Payer,
        "network": verifyResp.Network,
        "timestamp": time.Now(),
    })
})
```

## Configuration Options

### Functional Options Pattern

Configure the middleware using functional options:

```go
r.Use(gin.PaymentMiddleware(
    big.NewFloat(0.05), // $0.05 USDC
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    gin.WithDescription("Premium API access"),
    gin.WithMaxTimeoutSeconds(600), // 10 minutes
    gin.WithTestnet(false), // Use mainnet
    gin.WithFacilitatorURL("https://custom-facilitator.example.com"),
    gin.WithVerifyOnly(true), // Only verify, don't settle
))
```

### Available Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `WithDescription(string)` | string | "Payment required for {path}" | Human-readable payment description |
| `WithMimeType(string)` | string | "application/json" | Expected response MIME type |
| `WithMaxTimeoutSeconds(int)` | int | 300 | Payment validity timeout |
| `WithTestnet(bool)` | bool | true | Use testnet network |
| `WithCustomPaywallHTML(string)` | string | Default HTML | Custom HTML paywall for browsers |
| `WithFacilitatorURL(string)` | string | Coinbase facilitator | Custom facilitator endpoint |
| `WithVerifyOnly(bool)` | bool | false | Skip settlement, only verify payments |
| `WithOutputSchema(*json.RawMessage)` | *json.RawMessage | nil | JSON schema for expected response |

## Network Configuration

### Testnet (Default)

```go
// Uses base-sepolia network and testnet USDC
r.Use(gin.PaymentMiddleware(
    big.NewFloat(0.01),
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    gin.WithTestnet(true), // Explicit testnet
))
```

**Testnet Details**:
- Network: `base-sepolia`
- USDC Address: `0x036CbD53842c5426634e7929541eC2318f3dCF7e`
- Facilitator: `https://api.x402.coinbase.com`

### Mainnet

```go
// Uses base network and mainnet USDC
r.Use(gin.PaymentMiddleware(
    big.NewFloat(1.00), // $1.00 USDC
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    gin.WithTestnet(false), // Mainnet
))
```

**Mainnet Details**:
- Network: `base`
- USDC Address: `0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913`
- Facilitator: `https://api.x402.coinbase.com`

## Route Group Protection

Protect multiple endpoints with the same payment requirement:

```go
// Create payment-protected group
premium := r.Group("/premium")
premium.Use(gin.PaymentMiddleware(
    big.NewFloat(0.10), // $0.10 USDC
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    gin.WithDescription("Premium features access"),
))
{
    premium.GET("/analytics", analyticsHandler)
    premium.GET("/reports", reportsHandler)
    premium.POST("/export", exportHandler)
}

// Public endpoints (no payment required)
public := r.Group("/public")
{
    public.GET("/status", statusHandler)
    public.GET("/pricing", pricingHandler)
}
```

## Custom Paywall HTML

Provide a custom HTML paywall for browser clients:

```go
customHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>Payment Required</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .paywall { max-width: 400px; margin: 0 auto; }
        .amount { color: #007bff; font-size: 24px; font-weight: bold; }
    </style>
</head>
<body>
    <div class="paywall">
        <h1>🔒 Premium Content</h1>
        <p>This content requires a payment of <span class="amount">$0.05 USDC</span> to access.</p>
        <p>Please use your x402-enabled wallet to continue.</p>
    </div>
</body>
</html>
`

r.Use(gin.PaymentMiddleware(
    big.NewFloat(0.05),
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    gin.WithCustomPaywallHTML(customHTML),
))
```

## Error Handling

The middleware automatically handles different error scenarios:

### Missing Payment (402)

**API Clients** receive JSON:
```json
{
  "x402Version": 1,
  "error": "Payment required for this resource",
  "accepts": [
    {
      "scheme": "exact",
      "network": "base-sepolia",
      "maxAmountRequired": "10000",
      "asset": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
      "payTo": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
      "resource": "https://api.example.com/protected",
      "description": "Payment required for /protected",
      "mimeType": "application/json",
      "maxTimeoutSeconds": 300
    }
  ]
}
```

**Browser Clients** receive HTML:
```html
<html><body>Payment Required</body></html>
```

### Invalid Payment (400)

```json
{
  "x402Version": 1,
  "error": "Invalid payment header"
}
```

### Facilitator Error (503)

```json
{
  "x402Version": 1,
  "error": "Payment verification failed"
}
```

## Testing

### Test with Mock Payment

Create test requests with mock payment headers:

```go
func TestProtectedEndpoint(t *testing.T) {
    // Create Gin router with middleware
    r := gin.Default()
    r.Use(gin.PaymentMiddleware(
        big.NewFloat(0.01),
        "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
        gin.WithTestnet(true),
    ))
    
    r.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "success"})
    })
    
    // Test without payment (should return 402)
    req, _ := http.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    
    assert.Equal(t, 402, w.Code)
    
    // Test with valid payment (mock implementation needed)
    // This would require creating a valid X-PAYMENT header
}
```

## Production Considerations

### Security

1. **HTTPS**: Always use HTTPS in production
2. **Amount Validation**: Validate payment amounts server-side
3. **Address Verification**: Double-check recipient addresses
4. **Rate Limiting**: Implement rate limiting alongside payment middleware

### Performance

1. **Timeout Configuration**: Adjust timeouts based on your requirements
2. **Fallback Facilitator**: Configure backup facilitator for reliability
3. **Monitoring**: Monitor payment verification and settlement times

### Monitoring

Add logging and monitoring:

```go
// Custom middleware for monitoring
r.Use(func(c *gin.Context) {
    start := time.Now()
    c.Next()
    
    // Log payment attempts
    if paymentInfo, exists := c.Get("x402_payment"); exists {
        log.Printf("Payment verified for %s in %v", 
            paymentInfo.(*gin.VerifyResponse).Payer,
            time.Since(start))
    }
})

// Then apply payment middleware
r.Use(gin.PaymentMiddleware(
    big.NewFloat(0.01),
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
))
```

## Troubleshooting

### Common Issues

1. **"Invalid payment header"**: Check X-PAYMENT header format and encoding
2. **"Payment verification failed"**: Verify facilitator URL and network settings
3. **Settlement timeout**: Increase settlement timeout for slow networks
4. **Context access**: Ensure proper type assertion when accessing payment info

### Debug Mode

Enable debug logging:

```go
// Gin debug mode
gin.SetMode(gin.DebugMode)

// Add debug middleware
r.Use(func(c *gin.Context) {
    log.Printf("Request: %s %s", c.Request.Method, c.Request.URL.Path)
    log.Printf("X-PAYMENT header: %s", c.GetHeader("X-PAYMENT"))
    log.Printf("User-Agent: %s", c.GetHeader("User-Agent"))
    log.Printf("Accept: %s", c.GetHeader("Accept"))
    c.Next()
})
```

## Examples Repository

Complete working examples are available in the `examples/` directory:

- `examples/basic/` - Simple payment protection
- `examples/multichain/` - Multiple network support
- `examples/verification/` - Verify-only mode
- `examples/x402demo/` - Full-featured demonstration

## Support

- **Documentation**: [x402-go README](https://github.com/mark3labs/x402-go)
- **Issues**: [GitHub Issues](https://github.com/mark3labs/x402-go/issues)
- **Discord**: Community support channel
- **Email**: support@mark3labs.com