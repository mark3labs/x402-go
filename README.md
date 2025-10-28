# x402 Payment Client for Go

A Go library for making HTTP requests to x402-protected endpoints with automatic payment handling.

## Status

**Current Implementation**: Core MVP components complete (Phase 3 - User Story 1)

### Completed Components

✅ **Core Infrastructure**
- `x402/types.go` - Payment types and data structures
- `x402/errors.go` - Error handling and error codes
- `x402/signer.go` - Signer interface definition
- `x402/selector.go` - Payment selection logic

✅ **EVM Support** 
- `x402/evm/signer.go` - EVM signer implementation
- `x402/evm/eip3009.go` - EIP-3009 signing logic
- `x402/evm/keystore.go` - Keystore and mnemonic support
- Supported networks: Base, Base Sepolia, Ethereum, Sepolia

✅ **Solana (SVM) Support**
- `x402/svm/signer.go` - Solana signer implementation (partial)
- Note: Transaction building needs completion

✅ **HTTP Client**
- `x402/http/client.go` - x402-enabled HTTP client
- `x402/http/transport.go` - Automatic 402 handling
- Payment header building and settlement parsing

✅ **Example Application**
- `examples/x402demo/main.go` - CLI demo application

### Remaining Work

⏳ **Testing** (Tasks T008-T018, T031)
- Unit tests for all components
- Integration tests
- End-to-end tests

⏳ **User Story 2** (Tasks T032-T039)
- Multi-signer selection already implemented
- Needs testing and validation

⏳ **User Story 3** (Tasks T040-T047)
- Max amount validation already implemented
- Needs testing

⏳ **User Story 4** (Tasks T048-T057)
- Token priority already implemented
- Needs testing

⏳ **Polish** (Tasks T058-T076)
- Edge case handling
- Performance optimizations
- Documentation

## Quick Start

### Creating a Paywalled Endpoint (Server)

```go
package main

import (
    "log"
    "net/http"
    
    "github.com/mark3labs/x402-go"
    x402http "github.com/mark3labs/x402-go/http"
)

func main() {
    // Define payment requirements
    config := &x402http.Config{
        FacilitatorURL: "https://facilitator.x402.rs",
        PaymentRequirements: []x402.PaymentRequirement{
            {
                Scheme:            "exact",
                Network:           "base",
                MaxAmountRequired: "100000", // 0.1 USDC (6 decimals)
                Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // USDC on Base
                PayTo:             "0xYourPaymentRecipientAddress",
                MaxTimeoutSeconds: 60,
            },
        },
    }

    // Create middleware
    middleware := x402http.NewX402Middleware(config)

    // Wrap your handler
    http.Handle("/premium-data", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // This code only runs after successful payment
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"data": "premium content"}`))
    })))

    log.Println("Server running on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Making Requests to Paywalled Endpoints (Client)

```go
package main

import (
    "log"
    "github.com/mark3labs/x402-go/evm"
    x402http "github.com/mark3labs/x402-go/http"
)

func main() {
    // Create a signer
    signer, err := evm.NewSigner(
        evm.WithPrivateKey("0xYourPrivateKey"),
        evm.WithNetwork("base"),
        evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create x402 client
    client, err := x402http.NewClient(
        x402http.WithSigner(signer),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Make request - payment happens automatically!
    resp, err := client.Get("https://api.example.com/premium-data")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    // Check settlement
    if settlement := x402http.GetSettlement(resp); settlement != nil {
        log.Printf("Payment successful: %s", settlement.Transaction)
    }
}
```

## Usage

### Using the Example CLI

```bash
# Build the example
go build -o x402demo ./examples/x402demo/

# Make a request to a paywalled endpoint
./x402demo \
  --network base \
  --key 0xYourPrivateKeyHex \
  --url https://api.example.com/data \
  --max 1000000  # Optional: max 1 USDC per call
```

## Building

```bash
# Build all packages
go build ./...

# Build example
go build -o examples/x402demo/x402demo ./examples/x402demo/

# Run tests
go test -race ./...
```

## Architecture

```
x402/               # Core package
├── types.go        # Data structures
├── errors.go       # Error handling
├── signer.go       # Signer interface
├── selector.go     # Payment selection
├── evm/           # EVM implementation
│   ├── signer.go
│   ├── eip3009.go
│   └── keystore.go
├── svm/           # Solana implementation
│   └── signer.go
└── http/          # HTTP client
    ├── client.go
    └── transport.go

examples/
└── x402demo/      # Example CLI
    └── main.go
```

## Next Steps

1. ✅ Complete User Story 1 (Basic Payment) - **DONE**
2. ⏳ Add comprehensive test coverage
3. ⏳ Complete Solana transaction building
4. ⏳ Validate multi-signer scenarios (User Story 2)
5. ⏳ Performance testing and optimization
6. ⏳ Final polish and edge case handling

See `specs/002-x402-client/tasks.md` for detailed task breakdown.
