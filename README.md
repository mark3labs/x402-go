# x402-go

Go implementation of the x402 payment standard for paywalled HTTP endpoints.

## Overview

x402-go makes it simple to add crypto payments to HTTP APIs. This library provides:

- **USDC helpers** for easy payment setup across 8+ chains (Base, Polygon, Avalanche, Solana)
- **Middleware** for standard `net/http`, Chi, Gin, and PocketBase frameworks
- **HTTP client** with automatic payment handling
- **Multi-chain support** with automatic wallet selection

## Quick Start

### Server: Accept USDC Payments

Use the USDC helpers to create payment requirements in just a few lines:

```go
import (
    "net/http"
    "github.com/mark3labs/x402-go"
    x402http "github.com/mark3labs/x402-go/http"
)

func main() {
    // Create payment requirement using USDC helper
    requirement, _ := x402.NewUSDCPaymentRequirement(x402.USDCRequirementConfig{
        Chain:            x402.BaseMainnet,
        Amount:           "0.01",           // Human-readable USDC amount
        RecipientAddress: "0xYourAddress",
    })

    // Configure middleware
    config := &x402http.Config{
        FacilitatorURL: "https://facilitator.x402.rs",
        PaymentRequirements: []x402.PaymentRequirement{requirement},
    }

    // Protect your endpoint
    middleware := x402http.NewX402Middleware(config)
    http.Handle("/data", middleware(yourHandler))
    http.ListenAndServe(":8080", nil)
}
```

That's it! The helper automatically:
- Converts "0.01" to atomic units (10000)
- Sets the correct USDC contract address
- Configures EIP-3009 domain parameters
- Applies sensible defaults

### Client: Pay for API Access

```go
import (
    "github.com/mark3labs/x402-go/evm"
    x402http "github.com/mark3labs/x402-go/http"
)

// Create USDC token config using helper
token := x402.NewUSDCTokenConfig(x402.BaseMainnet, 1)

// Create signer with your wallet
signer, _ := evm.NewSigner(
    evm.WithPrivateKey("0xYourKey"),
    evm.WithNetwork("base"),
    evm.WithToken(token.Address, token.Symbol, token.Decimals),
)

// Create client - payments happen automatically
client, _ := x402http.NewClient(x402http.WithSigner(signer))
resp, _ := client.Get("https://api.example.com/data")
```

## USDC Chain Support

The library includes pre-configured USDC constants for these chains:

| Chain | Mainnet Constant | Testnet Constant |
|-------|-----------------|------------------|
| Base | `BaseMainnet` | `BaseSepolia` |
| Polygon | `PolygonMainnet` | `PolygonAmoy` |
| Avalanche | `AvalancheMainnet` | `AvalancheFuji` |
| Solana | `SolanaMainnet` | `SolanaDevnet` |

All USDC addresses are verified and include EIP-3009 parameters for EVM chains.

## Server Examples

### Accept Multiple Chains

Let clients pay with USDC on any supported chain:

```go
requirements := []x402.PaymentRequirement{}

// Accept Base USDC
baseReq, _ := x402.NewUSDCPaymentRequirement(x402.USDCRequirementConfig{
    Chain:            x402.BaseMainnet,
    Amount:           "0.50",
    RecipientAddress: "0xYourAddress",
})
requirements = append(requirements, baseReq)

// Or Polygon USDC
polygonReq, _ := x402.NewUSDCPaymentRequirement(x402.USDCRequirementConfig{
    Chain:            x402.PolygonMainnet,
    Amount:           "0.50",
    RecipientAddress: "0xYourAddress",
})
requirements = append(requirements, polygonReq)

// Or Solana USDC
solanaReq, _ := x402.NewUSDCPaymentRequirement(x402.USDCRequirementConfig{
    Chain:            x402.SolanaMainnet,
    Amount:           "0.50",
    RecipientAddress: "YourSolanaAddress",
})
requirements = append(requirements, solanaReq)

config := &x402http.Config{
    FacilitatorURL: "https://facilitator.x402.rs",
    PaymentRequirements: requirements,
}
```

### Using with Gin Framework

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/mark3labs/x402-go"
    ginx402 "github.com/mark3labs/x402-go/http/gin"
)

func main() {
    // Create payment requirement
    requirement, _ := x402.NewUSDCPaymentRequirement(x402.USDCRequirementConfig{
        Chain:            x402.BaseSepolia,
        Amount:           "0.01",
        RecipientAddress: "0xYourAddress",
    })

    config := &x402http.Config{
        FacilitatorURL: "https://facilitator.x402.rs",
        PaymentRequirements: []x402.PaymentRequirement{requirement},
    }

    // Setup Gin with x402 middleware
    r := gin.Default()
    r.Use(ginx402.NewGinX402Middleware(config))

    r.GET("/data", func(c *gin.Context) {
        // Access payment details from context
        if payment, exists := c.Get("x402_payment"); exists {
            verifyResp := payment.(*x402http.VerifyResponse)
            c.JSON(200, gin.H{
                "data": "your response",
                "payer": verifyResp.Payer,
            })
            return
        }
        c.JSON(402, gin.H{"error": "payment required"})
    })

    r.Run(":8080")
}
```

See `examples/gin/` for complete examples.

### Using with Chi Framework

Chi uses the standard `http.Handler` middleware interface, so you can use the base middleware directly:

```go
import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/mark3labs/x402-go"
    x402http "github.com/mark3labs/x402-go/http"
)

func main() {
    // Create payment requirement
    requirement, _ := x402.NewUSDCPaymentRequirement(x402.USDCRequirementConfig{
        Chain:            x402.BaseSepolia,
        Amount:           "0.01",
        RecipientAddress: "0xYourAddress",
    })

    config := &x402http.Config{
        FacilitatorURL: "https://facilitator.x402.rs",
        PaymentRequirements: []x402.PaymentRequirement{requirement},
    }

    // Setup Chi with x402 middleware (uses standard http.Handler interface)
    r := chi.NewRouter()
    r.Use(x402http.NewX402Middleware(config))

    r.Get("/data", func(w http.ResponseWriter, r *http.Request) {
        // Access payment details from context
        if payment := r.Context().Value(x402http.PaymentContextKey); payment != nil {
            verifyResp := payment.(*x402http.VerifyResponse)
            w.Header().Set("Content-Type", "application/json")
            w.Write([]byte(`{"data": "your response", "payer": "` + verifyResp.Payer + `"}`))
            return
        }
        w.WriteHeader(http.StatusPaymentRequired)
        w.Write([]byte(`{"error": "payment required"}`))
    })

    http.ListenAndServe(":8080", r)
}
```

See `examples/chi/` for complete examples.

### Using with PocketBase Framework

```go
import (
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
    "github.com/mark3labs/x402-go"
    x402http "github.com/mark3labs/x402-go/http"
    pbx402 "github.com/mark3labs/x402-go/http/pocketbase"
)

func main() {
    app := pocketbase.New()

    // Create payment requirement
    requirement, _ := x402.NewUSDCPaymentRequirement(x402.USDCRequirementConfig{
        Chain:            x402.BaseSepolia,
        Amount:           "0.01",
        RecipientAddress: "0xYourAddress",
    })

    config := &x402http.Config{
        FacilitatorURL: "https://facilitator.x402.rs",
        PaymentRequirements: []x402.PaymentRequirement{requirement},
    }

    // Apply middleware to specific routes
    app.OnRecordBeforeCreateRequest("protected_collection").BindFunc(func(e *core.RequestEvent) error {
        middleware := pbx402.NewPocketBaseX402Middleware(config)
        return middleware(e)
    })

    app.Start()
}
```

See `examples/pocketbase/` for complete examples.

### Custom Configuration

Override defaults for specific use cases:

```go
requirement, _ := x402.NewUSDCPaymentRequirement(x402.USDCRequirementConfig{
    Chain:             x402.BaseMainnet,
    Amount:            "2.50",
    RecipientAddress:  "0xYourAddress",
    Scheme:            "estimate",        // Default: "exact"
    MaxTimeoutSeconds: 600,               // Default: 300
    MimeType:          "application/xml", // Default: "application/json"
})
```

## Client Examples

### Single Chain Client

```go
import (
    "github.com/mark3labs/x402-go"
    "github.com/mark3labs/x402-go/evm"
    x402http "github.com/mark3labs/x402-go/http"
)

// Use USDC helper for token config
token := x402.NewUSDCTokenConfig(x402.BaseMainnet, 1)

signer, _ := evm.NewSigner(
    evm.WithPrivateKey("0xYourPrivateKey"),
    evm.WithNetwork("base"),
    evm.WithToken(token.Address, token.Symbol, token.Decimals),
)

client, _ := x402http.NewClient(x402http.WithSigner(signer))
resp, _ := client.Get("https://api.example.com/data")
```

### Multi-Chain Client

Configure multiple wallets and the client will automatically choose the best one:

```go
// Setup Base wallet
baseToken := x402.NewUSDCTokenConfig(x402.BaseMainnet, 1)  // Priority 1 (highest)
baseSigner, _ := evm.NewSigner(
    evm.WithPrivateKey("0xYourKey"),
    evm.WithNetwork("base"),
    evm.WithToken(baseToken.Address, baseToken.Symbol, baseToken.Decimals),
)

// Setup Solana wallet
solanaToken := x402.NewUSDCTokenConfig(x402.SolanaMainnet, 2)  // Priority 2
solanaSigner, _ := svm.NewSigner(
    svm.WithPrivateKey("YourSolanaKey"),
    svm.WithNetwork("solana"),
    svm.WithToken(solanaToken.Address, solanaToken.Symbol, solanaToken.Decimals),
)

// Client automatically selects appropriate wallet
client, _ := x402http.NewClient(
    x402http.WithSigner(baseSigner),
    x402http.WithSigner(solanaSigner),
)

// Works with any paywalled endpoint
resp, _ := client.Get("https://api.example.com/data")
```

### Solana Client

```go
import (
    "github.com/mark3labs/x402-go"
    "github.com/mark3labs/x402-go/svm"
    x402http "github.com/mark3labs/x402-go/http"
)

// Use USDC helper for token config
token := x402.NewUSDCTokenConfig(x402.SolanaMainnet, 1)

signer, _ := svm.NewSigner(
    svm.WithPrivateKey("Base58PrivateKey"),
    // Or load from file: svm.WithKeygenFile("~/.config/solana/id.json")
    svm.WithNetwork("solana"),
    svm.WithToken(token.Address, token.Symbol, token.Decimals),
)

client, _ := x402http.NewClient(x402http.WithSigner(signer))
resp, _ := client.Get("https://api.example.com/data")
```
