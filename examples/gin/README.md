# Gin Middleware Example

This example demonstrates how to use the x402 Gin middleware to protect API endpoints with payment requirements.

## Running the Example

```bash
cd examples/gin
go run main.go
```

The server will start on `http://localhost:8080`.

## Available Endpoints

### Public Endpoints (No Payment Required)

**GET /public/status**
```bash
curl http://localhost:8080/public/status
```

Response:
```json
{
  "status": "healthy",
  "service": "x402 Gin Example"
}
```

### Protected Endpoints (Payment Required)

**GET /protected/data** - Requires 0.01 USDC payment
```bash
# Without payment - returns 402
curl http://localhost:8080/protected/data

# With payment (you need to generate a valid X-PAYMENT header)
curl -H "X-PAYMENT: <base64-encoded-payment>" http://localhost:8080/protected/data
```

Response without payment (402):
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
      "resource": "http://localhost:8080/protected/data",
      "description": "Payment for API access",
      "mimeType": "application/json",
      "maxTimeoutSeconds": 300
    }
  ]
}
```

Response with valid payment (200):
```json
{
  "message": "Access granted with valid payment",
  "payer": "0x1234...",
  "data": "This is protected data"
}
```

**GET /verify-only/check** - Verify-only mode (no settlement)

This endpoint verifies payments but doesn't execute blockchain settlement. Useful for testing or scenarios where settlement happens separately.

**GET /premium/analytics** - Premium tier requiring 0.05 USDC

Higher payment amount for premium features.

## Implementation Details

### Basic Usage

```go
// Create Gin router
r := gin.Default()

// Configure x402 middleware
config := &httpx402.Config{
    FacilitatorURL: "https://api.x402.coinbase.com",
    PaymentRequirements: []x402.PaymentRequirement{{
        Scheme:            "exact",
        Network:           "base-sepolia",
        MaxAmountRequired: "10000", // 0.01 USDC
        Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
        PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
        MaxTimeoutSeconds: 300,
    }},
}

// Apply middleware
r.Use(ginx402.NewGinX402Middleware(config))

// Protected handler
r.GET("/data", func(c *gin.Context) {
    if payment, exists := c.Get("x402_payment"); exists {
        verifyResp := payment.(*httpx402.VerifyResponse)
        c.JSON(200, gin.H{"payer": verifyResp.Payer})
    }
})
```

### Route Groups

```go
// Public routes (no payment)
public := r.Group("/public")
{
    public.GET("/status", statusHandler)
}

// Protected routes (payment required)
protected := r.Group("/protected")
protected.Use(ginx402.NewGinX402Middleware(config))
{
    protected.GET("/data", dataHandler)
}
```

### Accessing Payment Information

Inside your Gin handlers, access payment details via context:

```go
func handler(c *gin.Context) {
    // Get payment info from Gin context
    paymentInfo, exists := c.Get("x402_payment")
    if !exists {
        c.JSON(500, gin.H{"error": "No payment info"})
        return
    }
    
    // Type assert to VerifyResponse
    verifyResp := paymentInfo.(*httpx402.VerifyResponse)
    
    // Use payment information
    payer := verifyResp.Payer
    c.JSON(200, gin.H{"payer": payer})
}
```

## Configuration Options

### Testnet vs Mainnet

**Testnet (base-sepolia)**:
```go
PaymentRequirement{
    Network: "base-sepolia",
    Asset:   "0x036CbD53842c5426634e7929541eC2318f3dCF7e", // Testnet USDC
}
```

**Mainnet (base)**:
```go
PaymentRequirement{
    Network: "base",
    Asset:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // Mainnet USDC
}
```

### Verify-Only Mode

Skip settlement and only verify payment validity:

```go
Config{
    VerifyOnly: true,
    // ... other fields
}
```

### Fallback Facilitator

Configure backup facilitator for reliability:

```go
Config{
    FacilitatorURL:         "https://primary-facilitator.com",
    FallbackFacilitatorURL: "https://backup-facilitator.com",
}
```

## Testing

To test with actual payments, you'll need:

1. A wallet with testnet USDC on base-sepolia
2. An x402-compatible client to generate payment headers
3. The facilitator service running at the configured URL

For local testing without payments, use the public endpoints or mock the facilitator service.

## Production Considerations

1. **Use HTTPS**: Always use TLS in production
2. **Recipient Address**: Use your actual wallet address for `PayTo`
3. **Network**: Switch to mainnet (`base`) for production
4. **Amount**: Set appropriate payment amounts in atomic units (6 decimals for USDC)
5. **Timeouts**: Adjust `MaxTimeoutSeconds` based on your requirements
6. **Monitoring**: Log payment verification and settlement for debugging

## Learn More

- [x402 Protocol Specification](https://github.com/mark3labs/x402-go)
- [Gin Framework Documentation](https://gin-gonic.com/)
- [Base Network Documentation](https://docs.base.org/)
