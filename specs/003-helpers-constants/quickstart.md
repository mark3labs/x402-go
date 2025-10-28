# Quick Start: Helper Functions and Constants

## Overview

The x402-go library provides helper functions and constants to quickly set up x402 clients and middleware with USDC payments across multiple chains.

## Installation

```bash
go get github.com/mark3labs/x402-go
```

## Supported Chains

**Mainnet:**
- Solana (`solana`)
- Base (`base`)
- Polygon PoS (`polygon`)
- Avalanche C-Chain (`avalanche`)

**Testnet:**
- Solana Devnet (`solana-devnet`)
- Base Sepolia (`base-sepolia`)
- Polygon Amoy (`polygon-amoy`)
- Avalanche Fuji (`avalanche-fuji`)

## Common Use Cases

### 1. Quick Client Setup (Single Chain)

Configure an x402 client to pay for protected resources using USDC on Base:

```go
package main

import (
    "github.com/mark3labs/x402-go"
    "github.com/mark3labs/x402-go/http"
)

func main() {
    // Create token config for Base USDC
    token := x402.NewTokenConfig(x402.BaseMainnet, 1)
    
    // Create client with Base USDC support
    client := http.NewClient(http.ClientConfig{
        Tokens: []x402.TokenConfig{token},
        // ... other config
    })
    
    // Use client to make paid requests
    resp, err := client.Get("https://api.example.com/protected")
    // ...
}
```

### 2. Multi-Chain Client Setup

Support multiple chains and let the client automatically select the best token:

```go
package main

import (
    "github.com/mark3labs/x402-go"
    "github.com/mark3labs/x402-go/http"
)

func main() {
    // Create token configs for multiple chains
    tokens := []x402.TokenConfig{
        x402.NewTokenConfig(x402.BaseMainnet, 1),      // Priority 1
        x402.NewTokenConfig(x402.PolygonMainnet, 2),   // Priority 2
        x402.NewTokenConfig(x402.SolanaMainnet, 3),    // Priority 3
    }
    
    client := http.NewClient(http.ClientConfig{
        Tokens: tokens,
        // ... other config
    })
    
    // Client will use highest priority token that matches server requirements
    resp, err := client.Get("https://api.example.com/protected")
    // ...
}
```

### 3. Middleware Payment Requirements

Configure middleware to accept USDC payments on Base:

```go
package main

import (
    "net/http"
    "github.com/mark3labs/x402-go"
    x402http "github.com/mark3labs/x402-go/http"
)

func main() {
    // Create payment requirement for Base USDC
    req, err := x402.NewPaymentRequirement(x402.PaymentRequirementConfig{
        Chain:            x402.BaseMainnet,
        Amount:           "1.00",  // 1 USDC per request
        RecipientAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    })
    if err != nil {
        panic(err)
    }
    
    // Create middleware
    middleware := x402http.NewMiddleware(x402http.MiddlewareConfig{
        FacilitatorURL:       "https://facilitator.example.com",
        PaymentRequirements:  []x402.PaymentRequirement{req},
    })
    
    // Wrap your handler
    handler := middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Protected content"))
    }))
    
    http.ListenAndServe(":8080", handler)
}
```

### 4. Multi-Chain Middleware

Accept payments on multiple chains:

```go
package main

import (
    "net/http"
    "github.com/mark3labs/x402-go"
    x402http "github.com/mark3labs/x402-go/http"
)

func main() {
    recipientAddress := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"
    amount := "2.50"  // 2.5 USDC
    
    // Create payment requirements for multiple chains
    baseReq, _ := x402.NewPaymentRequirement(x402.PaymentRequirementConfig{
        Chain:            x402.BaseMainnet,
        Amount:           amount,
        RecipientAddress: recipientAddress,
    })
    
    polygonReq, _ := x402.NewPaymentRequirement(x402.PaymentRequirementConfig{
        Chain:            x402.PolygonMainnet,
        Amount:           amount,
        RecipientAddress: recipientAddress,
    })
    
    solanaReq, _ := x402.NewPaymentRequirement(x402.PaymentRequirementConfig{
        Chain:            x402.SolanaMainnet,
        Amount:           amount,
        RecipientAddress: "DYw8jCTfwHNRJhhmFcbXvVDTqWMEVFBX6ZKUmG5CNSKK",  // Solana address
    })
    
    // Create middleware accepting any of these chains
    middleware := x402http.NewMiddleware(x402http.MiddlewareConfig{
        FacilitatorURL: "https://facilitator.example.com",
        PaymentRequirements: []x402.PaymentRequirement{
            baseReq,
            polygonReq,
            solanaReq,
        },
    })
    
    handler := middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Protected content"))
    }))
    
    http.ListenAndServe(":8080", handler)
}
```

### 5. Custom Payment Configuration

Override default payment settings:

```go
package main

import (
    "github.com/mark3labs/x402-go"
)

func main() {
    // Create with custom settings
    req, err := x402.NewPaymentRequirement(x402.PaymentRequirementConfig{
        Chain:             x402.BaseMainnet,
        Amount:            "5.00",
        RecipientAddress:  "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
        Scheme:            "estimate",      // Use estimate instead of exact
        MaxTimeoutSeconds: 600,             // 10 minutes instead of default 5
        MimeType:          "application/json",
    })
    if err != nil {
        panic(err)
    }
    
    // Use req in middleware...
}
```

### 6. Testnet Development

Use testnet constants for development and testing:

```go
package main

import (
    "github.com/mark3labs/x402-go"
    "github.com/mark3labs/x402-go/http"
)

func main() {
    // Use testnet tokens - NO REAL VALUE
    tokens := []x402.TokenConfig{
        x402.NewTokenConfig(x402.BaseSepolia, 1),
        x402.NewTokenConfig(x402.PolygonAmoy, 2),
        x402.NewTokenConfig(x402.SolanaDevnet, 3),
    }
    
    client := http.NewClient(http.ClientConfig{
        Tokens: tokens,
        // ... other config
    })
    
    // Test with testnet tokens
    resp, err := client.Get("https://api-dev.example.com/protected")
    // ...
}
```

### 7. Zero Amount (Free-with-Signature) Authorization

Create payment requirements for free resources that still require signature authorization:

```go
package main

import (
    "github.com/mark3labs/x402-go"
    x402http "github.com/mark3labs/x402-go/http"
)

func main() {
    // Create zero-amount payment requirement
    // Client still needs to sign, but no USDC transfer occurs
    req, err := x402.NewPaymentRequirement(x402.PaymentRequirementConfig{
        Chain:            x402.BaseMainnet,
        Amount:           "0",  // Zero amount: signature required, no payment
        RecipientAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
    })
    if err != nil {
        panic(err)
    }
    
    // Use in middleware for authenticated but free access
    middleware := x402http.NewMiddleware(x402http.MiddlewareConfig{
        FacilitatorURL:      "https://facilitator.example.com",
        PaymentRequirements: []x402.PaymentRequirement{req},
    })
    
    // Clients must still authorize, proving control of wallet
    // Useful for rate limiting, attribution, or compliance
}
```

### 8. Network Validation

Validate network identifiers from payment requirements:

```go
package main

import (
    "fmt"
    "github.com/mark3labs/x402-go"
)

func main() {
    // Validate network type
    netType, err := x402.ValidateNetwork("base")
    if err != nil {
        panic(err)
    }
    
    switch netType {
    case x402.NetworkTypeEVM:
        fmt.Println("EVM-based chain")
        // Use EVM signer
    case x402.NetworkTypeSVM:
        fmt.Println("Solana-based chain")
        // Use Solana signer
    default:
        fmt.Println("Unknown network")
    }
}
```

## Available Chain Constants

All constants are exported from the root `x402` package:

```go
// Mainnet
x402.SolanaMainnet
x402.BaseMainnet
x402.PolygonMainnet
x402.AvalancheMainnet

// Testnet
x402.SolanaDevnet
x402.BaseSepolia
x402.PolygonAmoy
x402.AvalancheFuji
```

Each constant provides:
- `NetworkID`: x402 protocol network identifier
- `USDCAddress`: Official Circle USDC token address
- `Decimals`: Token decimals (always 6)
- `EIP3009Name`: EIP-3009 domain name (EVM chains only)
- `EIP3009Version`: EIP-3009 version (EVM chains only)

## Helper Functions

### NewPaymentRequirement

Creates a `PaymentRequirement` struct for middleware configuration.

```go
func NewPaymentRequirement(config PaymentRequirementConfig) (PaymentRequirement, error)
```

**Parameters:**
- `Chain`: ChainConfig constant (e.g., `x402.BaseMainnet`)
- `Amount`: Human-readable amount string (e.g., "1.5")
- `RecipientAddress`: Payment recipient address
- `Scheme`: Optional, defaults to "exact"
- `MaxTimeoutSeconds`: Optional, defaults to 300
- `MimeType`: Optional, defaults to "application/json"

**Returns:**
- `PaymentRequirement`: Configured payment requirement
- `error`: Validation error with parameter name and reason

### NewTokenConfig

Creates a `TokenConfig` struct for client configuration.

```go
func NewTokenConfig(chain ChainConfig, priority int) TokenConfig
```

**Parameters:**
- `chain`: ChainConfig constant (e.g., `x402.SolanaMainnet`)
- `priority`: Token selection priority (lower = higher priority)

**Returns:**
- `TokenConfig`: Configured token with USDC details

### ValidateNetwork

Validates a network identifier and returns its type.

```go
func ValidateNetwork(networkID string) (NetworkType, error)
```

**Parameters:**
- `networkID`: Network identifier string (e.g., "base", "solana")

**Returns:**
- `NetworkType`: `NetworkTypeEVM`, `NetworkTypeSVM`, or `NetworkTypeUnknown`
- `error`: Error if network is unknown

## Error Handling

All helper functions return structured errors with parameter names:

```go
req, err := x402.NewPaymentRequirement(x402.PaymentRequirementConfig{
    Chain:            x402.BaseMainnet,
    Amount:           "-5.00",  // Invalid: negative
    RecipientAddress: "",       // Invalid: empty
})

if err != nil {
    // Error format: "amount: must be positive" or "recipientAddress: cannot be empty"
    fmt.Println(err)
}
```

## Important Notes

1. **Testnet tokens have NO financial value** - only use for testing
2. **USDC addresses verified 2025-10-28** - check for updates when upgrading library
3. **All chains use 6 decimals** - amount "1.0" becomes 1,000,000 atomic units
4. **Zero amounts allowed** - amount "0" or "0.0" is valid for free-with-signature authorization flows
5. **Precision rounding** - amounts with >6 decimals (e.g., "1.1234567") are rounded using standard float64 rounding
6. **EIP-3009 parameters auto-populated** - no manual domain setup needed for EVM chains; chain-specific names handled automatically
7. **Address formats differ by chain** - EVM uses 0x-prefixed hex, Solana uses base58

## Next Steps

- See [data-model.md](./data-model.md) for detailed type information
- See [contracts/helpers-api.yaml](./contracts/helpers-api.yaml) for complete API specification
- See examples in `examples/` directory for complete working code
