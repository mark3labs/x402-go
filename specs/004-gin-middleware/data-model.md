# Data Model: Gin Middleware for x402 Payment Protocol

**Date**: 2025-10-29  
**Feature**: Gin Middleware for x402 Payment Protocol

## Core Entities

### PaymentMiddleware

**Description**: The main Gin middleware function that enforces payment gating on protected routes. Configured with amount, address, and options; intercepts requests, validates payments, and controls handler execution.

**Fields**:
- `amount *big.Float` - Decimal USDC amount to charge (e.g., 0.01 for 1 cent)
- `address string` - Recipient wallet address for payments
- `options *PaymentMiddlewareOptions` - Configuration options for behavior customization

**Methods**:
- `gin.HandlerFunc` - Returns a Gin-compatible middleware function

**Relationships**:
- Uses `FacilitatorClient` for payment verification and settlement
- Creates `PaymentRequirement` instances for each request
- Stores `VerifyResponse` in Gin context after successful verification
- Wraps `gin.ResponseWriter` with custom `responseWriter` for settlement

### PaymentMiddlewareOptions

**Description**: Configuration structure containing all optional settings for the middleware behavior.

**Fields**:
- `Description string` - Human-readable payment description
- `MimeType string` - Expected response MIME type (default: "application/json")
- `MaxTimeoutSeconds int` - Payment validity timeout (default: 300)
- `Testnet bool` - Use testnet network (default: true)
- `CustomPaywallHTML string` - Custom HTML for browser paywalls
- `FacilitatorURL string` - Custom facilitator endpoint URL
- `VerifyOnly bool` - Skip settlement, only verify payments
- `OutputSchema *json.RawMessage` - JSON schema for expected response

**Validation Rules**:
- `MaxTimeoutSeconds` must be positive (>= 1)
- `FacilitatorURL` must be valid URL if provided
- `CustomPaywallHTML` must be valid HTML if provided

### ResponseWriter

**Description**: Custom wrapper around Gin's ResponseWriter that captures response body and status code. Required for settlement after handler execution but before response is sent to client.

**Fields**:
- `gin.ResponseWriter` - Embedded original response writer
- `body *strings.Builder` - Buffer to capture response content
- `statusCode int` - Captured HTTP status code
- `written bool` - Flag to track if headers have been written

**Methods**:
- `WriteHeader(code int)` - Intercepts status code setting
- `Write(b []byte) (int, error)` - Intercepts response body writing
- `WriteString(s string) (int, error)` - Intercepts string writing

**State Transitions**:
1. **Initial**: `written = false`, `statusCode = 200`
2. **After WriteHeader**: `written = true`, `statusCode` set
3. **After Write/WriteString**: Content appended to `body` buffer

## Configuration Patterns

### Functional Options

**Description**: Functions that modify `PaymentMiddlewareOptions` for clean, extensible configuration.

**Available Options**:
- `WithDescription(string)` - Set payment description
- `WithMimeType(string)` - Set expected MIME type
- `WithMaxTimeoutSeconds(int)` - Set payment timeout
- `WithTestnet(bool)` - Enable/disable testnet mode
- `WithCustomPaywallHTML(string)` - Set custom HTML paywall
- `WithFacilitatorURL(string)` - Set custom facilitator URL
- `WithVerifyOnly(bool)` - Enable verify-only mode
- `WithOutputSchema(*json.RawMessage)` - Set response schema

### Default Configuration

**Testnet Defaults**:
- Network: "base-sepolia"
- USDC Address: "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
- FacilitatorURL: "https://api.x402.coinbase.com"

**Mainnet Defaults**:
- Network: "base" 
- USDC Address: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
- FacilitatorURL: "https://api.x402.coinbase.com"

## Request Flow Data

### Payment Detection

**Input Headers**:
- `X-PAYMENT` - Base64-encoded payment payload
- `User-Agent` - Browser detection string
- `Accept` - Content type negotiation

**Detection Logic**:
```go
isWebBrowser := strings.Contains(acceptHeader, "text/html") && 
                strings.Contains(userAgent, "Mozilla")
```

### Payment Processing

**Parsed Payment**:
- `PaymentPayload` - Decoded from X-PAYMENT header
- `PaymentRequirement` - Generated from middleware config + request URL
- `VerifyResponse` - From facilitator verification
- `SettlementResponse` - From facilitator settlement

**Context Storage**:
- Key: `"x402_payment"`
- Value: `*VerifyResponse` (contains payer, validity, etc.)

### Response Generation

**Success Response**:
- Original handler response (buffered)
- `X-PAYMENT-RESPONSE` header with settlement details
- Status code from handler

**Error Responses**:
- **402 Payment Required**: JSON with payment requirements OR HTML paywall
- **400 Bad Request**: JSON with "Invalid payment header" message
- **503 Service Unavailable**: JSON with facilitator error details

## Validation Rules

### Input Validation

**X-PAYMENT Header**:
1. Must be present and non-empty
2. Must be valid base64 encoding
3. Must decode to valid JSON
4. Must contain `x402Version: 1`

**Payment Amount**:
- Must be positive decimal number
- Must convert to integer atomic units without precision loss
- Must meet or exceed required amount

**Recipient Address**:
- Must be non-empty string
- Must be valid blockchain address format (EVM or SVM)

### Configuration Validation

**Timeout Values**:
- `MaxTimeoutSeconds` must be >= 1 and <= 3600 (1 hour max)
- `VerifyTimeout` fixed at 5 seconds
- `SettleTimeout` fixed at 60 seconds

**Network Configuration**:
- Testnet mode uses base-sepolia network
- Mainnet mode uses base network
- USDC addresses are network-specific

## Error Handling

### Error Types

**Malformed Header Errors**:
- Empty X-PAYMENT header
- Invalid base64 encoding
- Invalid JSON structure
- Unsupported x402 version

**Payment Validation Errors**:
- Insufficient payment amount
- Expired payment authorization
- Invalid signature
- Unsupported network/scheme

**Facilitator Errors**:
- Service unavailable
- Network timeout
- Settlement failure

### Error Response Format

**JSON Error Response**:
```json
{
  "error": "Human-readable error message",
  "x402Version": 1,
  "accepts": [PaymentRequirement...]
}
```

**HTML Paywall Response**:
```html
<html><body>Payment Required</body></html>
```

## State Management

### Request-Scoped State

**Middleware State** (per request):
- `isWebBrowser bool` - Browser detection result
- `paymentPayload PaymentPayload` - Parsed payment data
- `paymentRequirement PaymentRequirement` - Generated requirements
- `verifyResponse VerifyResponse` - Verification result
- `settlementResponse SettlementResponse` - Settlement result

**Response Writer State** (per request):
- `bufferedResponse []byte` - Captured handler output
- `statusCode int` - Handler status code
- `headersWritten bool` - Response writer state

### Configuration State

**Static Configuration** (middleware instance):
- Facilitator client instances (primary + fallback)
- Enriched payment requirements
- Default timeout values
- Network-specific addresses

## Integration Points

### Gin Context Integration

**Storage Pattern**:
```go
c.Set("x402_payment", verifyResponse)
paymentInfo := c.Get("x402_payment")
```

**Handler Access Pattern**:
```go
if paymentInfo, exists := c.Get("x402_payment"); exists {
    if verifyResp, ok := paymentInfo.(*VerifyResponse); ok {
        payer := verifyResp.Payer
        // Use payment information
    }
}
```

### Facilitator Client Integration

**Client Creation** (at middleware initialization):
```go
facilitator := &FacilitatorClient{
    BaseURL:       options.FacilitatorURL,
    Client:        &http.Client{},
    VerifyTimeout: 5 * time.Second,
    SettleTimeout: 60 * time.Second,
}
```

**Verification Flow**:
1. Build `FacilitatorRequest` with payment and requirements
2. POST to `/verify` endpoint
3. Parse `VerifyResponse`
4. Check `IsValid` flag

**Settlement Flow**:
1. Build `FacilitatorRequest` with same data
2. POST to `/settle` endpoint  
3. Parse `SettlementResponse`
4. Add `X-PAYMENT-RESPONSE` header

## Performance Considerations

### Memory Usage

**Per-Request Allocation**:
- Response writer buffer (scales with response size)
- Payment requirement copies (small, fixed size)
- Context values (small, fixed size)

**Optimizations**:
- Reuse facilitator clients across requests
- Buffer only when settlement is required
- Clear response writer state after each request

### Latency Impact

**Additional Latency**:
- Payment verification: ~5 seconds max
- Settlement: ~60 seconds max
- Response buffering: minimal overhead

**Optimizations**:
- Parallel verification with fallback facilitator
- Early termination on invalid payments
- Timeout handling to prevent hanging requests