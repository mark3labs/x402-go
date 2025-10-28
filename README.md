# x402-go

Go implementation of the x402 payment standard for paywalled HTTP endpoints.

## Quick Start with Chain Helpers

The library provides chain constants and helper functions to quickly set up x402 payments:

```go
import "github.com/mark3labs/x402-go"

// Create payment requirement using chain constants
req, err := x402.NewPaymentRequirement(x402.PaymentRequirementConfig{
    Chain:            x402.BaseMainnet,  // Built-in chain constant
    Amount:           "1.50",             // Human-readable amount
    RecipientAddress: "0xYourAddress",
})

// Create token config for client
token := x402.NewTokenConfig(x402.BaseMainnet, 1)  // Priority 1
```

**Available chain constants:** `BaseMainnet`, `BaseSepolia`, `PolygonMainnet`, `PolygonAmoy`, `AvalancheMainnet`, `AvalancheFuji`, `SolanaMainnet`, `SolanaDevnet`

See `examples/basic/main.go` for complete examples.

## Creating a Paywalled Server

```go
import (
    "github.com/mark3labs/x402-go"
    x402http "github.com/mark3labs/x402-go/http"
)

// Configure payment requirements
config := &x402http.Config{
    FacilitatorURL: "https://facilitator.x402.rs",
    PaymentRequirements: []x402.PaymentRequirement{
        {
            Scheme:            "exact",
            Network:           "base-sepolia",
            MaxAmountRequired: "10000",  // 0.01 USDC (6 decimals)
            Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
            PayTo:             "0xYourAddress",
            Description:       "API Access",
            MaxTimeoutSeconds: 60,
        },
    },
}

// Create middleware and protect endpoints
middleware := x402http.NewX402Middleware(config)
http.Handle("/protected", middleware(yourHandler))
```

### Multi-chain pricing

```go
PaymentRequirements: []x402.PaymentRequirement{
    // Accept Base USDC
    {
        Scheme:            "exact",
        Network:           "base",
        MaxAmountRequired: "1000000",  // 1 USDC
        Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
        PayTo:             "0xYourAddress",
    },
    // Or Solana USDC
    {
        Scheme:            "exact",
        Network:           "solana",
        MaxAmountRequired: "1000000",  // 1 USDC
        Asset:             "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
        PayTo:             "YourSolanaAddress",
    },
}
```

## Creating a Payment Client

### EVM Wallet

```go
import (
    "github.com/mark3labs/x402-go/evm"
    x402http "github.com/mark3labs/x402-go/http"
)

// Create EVM signer
signer, err := evm.NewSigner(
    evm.WithPrivateKey("0xYourPrivateKey"),
    evm.WithNetwork("base"),
    evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
    evm.WithMaxAmountPerCall("5000000"),  // Optional: 5 USDC max per call
)

// Create client
client, err := x402http.NewClient(x402http.WithSigner(signer))

// Make requests (payment handled automatically)
resp, err := client.Get("https://api.example.com/data")
```

**Supported EVM networks:** `base`, `base-sepolia`, `ethereum`, `sepolia`

### Solana Wallet

```go
import (
    "github.com/mark3labs/x402-go/svm"
    x402http "github.com/mark3labs/x402-go/http"
)

// Create Solana signer
signer, err := svm.NewSigner(
    svm.WithPrivateKey("Base58PrivateKey"),  // Or WithKeygenFile("~/.config/solana/id.json")
    svm.WithNetwork("solana"),
    svm.WithToken("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", "USDC", 6),
)

// Create client
client, err := x402http.NewClient(x402http.WithSigner(signer))

// Make requests
resp, err := client.Get("https://api.example.com/data")
```

**Supported Solana networks:** `solana`, `solana-devnet`, `testnet`, `mainnet-beta`

### Multi-wallet Client

```go
// Add multiple signers for automatic network selection
client, err := x402http.NewClient(
    x402http.WithSigner(evmSigner),
    x402http.WithSigner(svmSigner),
)

// Client automatically selects appropriate wallet based on server requirements
resp, err := client.Get(anyPaywalledURL)
```
