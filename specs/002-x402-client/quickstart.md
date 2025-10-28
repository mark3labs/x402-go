# Quickstart: x402 Payment Client

## Installation

```bash
go get github.com/mark3labs/x402-go/x402
go get github.com/mark3labs/x402-go/x402/http
```

## Basic Usage

### 1. Create a Simple Client (Single EVM Signer)

```go
package main

import (
    "log"
    "net/http"
    
    "github.com/mark3labs/x402-go/x402"
    x402http "github.com/mark3labs/x402-go/x402/http"
    "github.com/mark3labs/x402-go/x402/evm"
)

func main() {
    // Create an EVM signer with private key
    signer, err := evm.NewSigner(
        evm.WithPrivateKey("0xYourPrivateKeyHex"),
        evm.WithNetwork("base"),
        evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Create x402-enabled HTTP client
    client := x402http.NewClient(
        x402http.WithSigner(signer),
    )
    
    // Make request to paywalled endpoint
    resp, err := client.Get("https://api.example.com/premium-data")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
    
    // Payment happens automatically!
    // Check if payment was made
    if settlement := x402http.GetSettlement(resp); settlement != nil {
        log.Printf("Payment successful: tx=%s", settlement.Transaction)
    }
}
```

### 2. Multi-Signer Setup (EVM + Solana)

```go
package main

import (
    "github.com/mark3labs/x402-go/x402/evm"
    "github.com/mark3labs/x402-go/x402/svm"
    x402http "github.com/mark3labs/x402-go/x402/http"
)

func main() {
    // Create EVM signer for Base network
    evmSigner, _ := evm.NewSigner(
        evm.WithPrivateKey("0xPrivateKey"),
        evm.WithNetwork("base"),
        evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
        evm.WithPriority(1), // Higher priority
    )
    
    // Create Solana signer
    solSigner, _ := svm.NewSigner(
        svm.WithPrivateKey("Base58PrivateKey"),
        svm.WithNetwork("solana"),
        svm.WithToken("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", "USDC", 6),
        svm.WithPriority(2), // Lower priority
    )
    
    // Create client with multiple signers
    client := x402http.NewClient(
        x402http.WithSigner(evmSigner),
        x402http.WithSigner(solSigner),
    )
    
    // Client automatically selects appropriate signer based on server requirements
    resp, _ := client.Get("https://api.example.com/data")
}
```

### 3. Per-Transaction Limits

```go
// Set maximum amount per payment call
signer, _ := evm.NewSigner(
    evm.WithPrivateKey("0xKey"),
    evm.WithNetwork("base"),
    evm.WithToken("0xUSDC", "USDC", 6),
    evm.WithMaxAmountPerCall("1000000"), // Max 1 USDC per call
)

client := x402http.NewClient(
    x402http.WithSigner(signer),
)

// Payments exceeding the max amount will be rejected
// The client will try other signers if available
```

### 4. Load Keys from Different Sources

```go
// From mnemonic
signer, _ := evm.NewSigner(
    evm.WithMnemonic("your twelve word mnemonic phrase here", 0), // account 0
    evm.WithNetwork("base"),
    evm.WithToken("0xUSDC", "USDC", 6),
)

// From keystore file
signer, _ := evm.NewSigner(
    evm.WithKeystore("/path/to/keystore.json", "password"),
    evm.WithNetwork("base"),
    evm.WithToken("0xUSDC", "USDC", 6),
)

// Solana from keygen file
solSigner, _ := svm.NewSigner(
    svm.WithKeygenFile("/path/to/id.json"),
    svm.WithNetwork("solana"),
    svm.WithToken("USDC_MINT", "USDC", 6),
)
```

### 5. Token Priority Configuration

```go
// Configure multiple tokens with priorities
signer, _ := evm.NewSigner(
    evm.WithPrivateKey("0xKey"),
    evm.WithNetwork("base"),
    evm.WithTokenPriority("0xUSDC", "USDC", 6, 1),  // Priority 1 (highest)
    evm.WithTokenPriority("0xUSDT", "USDT", 6, 2),  // Priority 2
    evm.WithTokenPriority("0xDAI", "DAI", 18, 3),   // Priority 3
)

// Client will prefer USDC, fall back to USDT, then DAI
```

### 6. Custom HTTP Client

```go
// Use existing http.Client with custom settings
httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns: 100,
    },
}

client := x402http.NewClient(
    x402http.WithHTTPClient(httpClient),
    x402http.WithSigner(signer),
)
```

## Example CLI Application

### Running the Example

The repository includes a complete example that can run as both client and server:

```bash
# Build the example
cd examples/x402demo
go build -o x402demo

# Run as server (accepts USDC payments)
./x402demo server --network base --payTo 0xYourAddress

# Run as client (makes payment)
./x402demo client --network base --key 0xYourPrivateKey --url http://localhost:8080/data
```

### Server Mode

```bash
./x402demo server \
  --network base \
  --payTo 0x209693Bc6afc0C5328bA36FaF03C514EF312287C \
  --port 8080
```

The server:
- Requires x402 payments for `/data` endpoint
- Uses https://facilitator.x402.rs for payment verification
- Accepts USDC on the specified network

### Client Mode

```bash
./x402demo client \
  --network base \
  --key 0xYourPrivateKey \
  --url http://localhost:8080/data
```

The client:
- Automatically handles 402 Payment Required responses
- Signs payment with provided key
- Retries request with payment header

## Error Handling

```go
resp, err := client.Get(url)
if err != nil {
    var paymentErr *x402.PaymentError
    if errors.As(err, &paymentErr) {
        switch paymentErr.Code {
        case x402.ErrCodeAmountExceeded:
            log.Println("Payment exceeds max amount limit")
        case x402.ErrCodeNoValidSigner:
            log.Println("No signer can satisfy payment requirements")
        default:
            log.Printf("Payment error: %v", paymentErr)
        }
    }
}
```

## Testing

```go
// Use mock signers for testing
mockSigner := &MockSigner{
    NetworkID: "base",
    CanSignFunc: func(req *x402.PaymentRequirements) bool {
        return true
    },
    SignFunc: func(req *x402.PaymentRequirements) (*x402.PaymentPayload, error) {
        return &x402.PaymentPayload{
            X402Version: 1,
            Scheme:      "exact",
            Network:     "base",
            Payload:     mockPayload,
        }, nil
    },
}

client := x402http.NewClient(
    x402http.WithSigner(mockSigner),
)
```

## Advanced Configuration

### Concurrent Request Handling

The client handles concurrent requests safely:

```go
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        resp, err := client.Get("https://api.example.com/data")
        // Thread-safe operation, max amount limits enforced
    }()
}
wg.Wait()
```

### Custom Payment Selection

```go
// Implement custom selection logic
selector := &CustomSelector{
    SelectFunc: func(requirements *x402.PaymentRequirements, signers []x402.Signer) x402.Signer {
        // Your custom logic here
        return signers[0]
    },
}

client := x402http.NewClient(
    x402http.WithSelector(selector),
    x402http.WithSigner(signer),
)
```

## Troubleshooting

### Common Issues

1. **"No valid signer" error**
   - Check that signer network matches server requirements
   - Verify token addresses are correct
   - Ensure signer has the required token configured

2. **"Amount exceeded" error**
   - Check that payment amount is within max limit
   - Increase max amount per call in signer configuration
   - Or use a different signer with higher limits

3. **"Invalid signature" from server**
   - Verify private key is correct
   - Check network ID matches
   - Ensure token contract details (name, version) are correct

4. **Slow performance**
   - Reuse client instances (they're thread-safe)
   - Check network latency to blockchain RPC
   - Enable connection pooling in HTTP client

## Next Steps

- Review the [full API documentation](./contracts/client-api.yaml)
- See [examples](../../examples/x402demo) for complete working code
- Read the [x402 specification](https://github.com/coinbase/x402) for protocol details