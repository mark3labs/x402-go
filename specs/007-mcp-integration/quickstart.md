# Quickstart: MCP x402 Integration

**Branch**: `007-mcp-integration` | **Date**: 2025-10-31

## Overview

This guide shows how to quickly integrate x402 payments into your MCP applications using the x402-go library.

## Installation

```bash
go get github.com/mark3labs/x402-go
go get github.com/mark3labs/mcp-go@latest
```

## Client Setup (5 minutes)

### 1. Create MCP Client with x402 Payments

```go
package main

import (
    "context"
    "log"
    
    "github.com/mark3labs/x402-go/mcp/client"
    "github.com/mark3labs/x402-go/signers/evm"
    mcpclient "github.com/mark3labs/mcp-go/client"
    "github.com/mark3labs/mcp-go/mcp"
)

func main() {
    // Create EVM signer for Base network
    signer, err := evm.NewPrivateKeySigner(
        "YOUR_PRIVATE_KEY_HEX",
        evm.WithChain(evm.ChainBase),
        evm.WithToken(evm.TokenUSDC),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create x402 transport with payment support
    transport, err := client.NewTransport(
        "http://localhost:8080",
        client.WithSigner(signer),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create MCP client
    mcpClient := mcpclient.NewClient(transport)
    
    ctx := context.Background()
    if err := mcpClient.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer mcpClient.Close()

    // Initialize session
    _, err = mcpClient.Initialize(ctx, mcp.InitializeRequest{
        Params: mcp.InitializeParams{
            ProtocolVersion: "1.0.0",
            ClientInfo: mcp.Implementation{
                Name:    "my-app",
                Version: "1.0.0",
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    // Call a paid tool - payment handled automatically!
    result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
        Params: mcp.CallToolParams{
            Name: "premium-search",
            Arguments: map[string]any{
                "query": "blockchain",
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    // Process result
    log.Printf("Result: %v", result)
}
```

## Server Setup (5 minutes)

### 2. Create MCP Server with x402 Protection

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/mark3labs/x402-go/mcp/server"
    "github.com/mark3labs/mcp-go/mcp"
)

func main() {
    // Create x402 MCP server
    srv := server.NewX402Server(
        "my-mcp-server", 
        "1.0.0",
        &server.Config{
            FacilitatorURL: "https://facilitator.x402.rs",
            VerifyOnly:     false, // Set true for testing
        },
    )

    // Add a free tool
    srv.AddTool(
        mcp.NewTool("echo",
            mcp.WithDescription("Echo back the input"),
            mcp.WithString("message", mcp.Required()),
        ),
        echoHandler,
    )

    // Add a paid tool with USDC on Base
    srv.AddPayableTool(
        mcp.NewTool("premium-search",
            mcp.WithDescription("Premium search service"),
            mcp.WithString("query", mcp.Required()),
        ),
        searchHandler,
        server.RequireUSDCBase(
            "YOUR_WALLET_ADDRESS",
            "10000",  // 0.01 USDC in atomic units
            "Premium search - 0.01 USDC",
        ),
    )

    // Start server
    log.Println("Starting MCP server on :8080")
    if err := srv.Start(":8080"); err != nil {
        log.Fatal(err)
    }
}

func echoHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    message := req.GetString("message", "")
    return &mcp.CallToolResult{
        Content: []mcp.Content{
            mcp.NewTextContent(fmt.Sprintf("Echo: %s", message)),
        },
    }, nil
}

func searchHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    query := req.GetString("query", "")
    // Perform premium search logic
    results := fmt.Sprintf("Premium results for: %s", query)
    return &mcp.CallToolResult{
        Content: []mcp.Content{
            mcp.NewTextContent(results),
        },
    }, nil
}
```

## Multi-Chain Support

### EVM Networks (Ethereum, Base, Polygon, Avalanche)

```go
// Base mainnet
signer := evm.NewPrivateKeySigner(key, 
    evm.WithChain(evm.ChainBase),
    evm.WithToken(evm.TokenUSDC))

// Polygon mainnet
signer := evm.NewPrivateKeySigner(key,
    evm.WithChain(evm.ChainPolygon),
    evm.WithToken(evm.TokenUSDC))

// Base Sepolia testnet
signer := evm.NewPrivateKeySigner(key,
    evm.WithChain(evm.ChainBaseSepolia),
    evm.WithToken(evm.TokenUSDC))
```

### Solana Network

```go
import "github.com/mark3labs/x402-go/signers/svm"

// Solana mainnet
signer := svm.NewPrivateKeySigner(
    privateKeyBase58,
    svm.WithNetwork(svm.NetworkMainnet),
    svm.WithToken(svm.TokenUSDC),
)

// Solana devnet
signer := svm.NewPrivateKeySigner(
    privateKeyBase58,
    svm.WithNetwork(svm.NetworkDevnet),
    svm.WithToken(svm.TokenUSDC),
)
```

## Multiple Payment Options

### Server: Accept Multiple Networks

```go
srv.AddPayableTool(
    tool,
    handler,
    // Option 1: USDC on Base
    server.RequireUSDCBase(walletAddress, "10000", "0.01 USDC on Base"),
    // Option 2: USDC on Polygon (lower priority)
    server.RequireUSDCPolygon(walletAddress, "10000", "0.01 USDC on Polygon"),
    // Option 3: USDC on Solana
    server.RequireUSDCSolana(solanaAddress, "10000", "0.01 USDC on Solana"),
)
```

### Client: Multiple Signers with Fallback

```go
transport, err := client.NewTransport(
    serverURL,
    // Primary: Base (cheapest)
    client.WithSigner(baseSigner),
    // Fallback: Polygon
    client.WithSigner(polygonSigner),
    // Last resort: Solana
    client.WithSigner(solanaSigner),
)
```

## Testing with Verify-Only Mode

For development, enable verify-only mode to test payment flows without actual blockchain transactions:

```go
srv := server.NewX402Server(
    "test-server",
    "1.0.0", 
    &server.Config{
        FacilitatorURL: "https://facilitator.x402.rs",
        VerifyOnly:     true,  // No actual settlement
        Verbose:        true,  // Detailed logging
    },
)
```

## Payment Event Handling

### Monitor Payment Events

```go
transport, err := client.NewTransport(
    serverURL,
    client.WithSigner(signer),
    client.WithPaymentCallback(func(event client.PaymentEvent) {
        switch event.Type {
        case client.PaymentAttempt:
            log.Printf("Attempting payment: %s %s", event.Amount, event.Asset)
        case client.PaymentSuccess:
            log.Printf("Payment successful: tx=%s", event.Transaction)
        case client.PaymentFailure:
            log.Printf("Payment failed: %v", event.Error)
        }
    }),
)
```

## Complete Example

See `examples/mcp/` for a complete working example with both client and server:

```bash
# Terminal 1: Start server
cd examples/mcp
go run . -mode server -wallet YOUR_WALLET_ADDRESS

# Terminal 2: Run client
go run . -mode client -key YOUR_PRIVATE_KEY
```

## Common Patterns

### 1. Free and Paid Tools in Same Server

```go
// Free tools - no payment requirement
srv.AddTool(freeTool, freeHandler)

// Paid tools - require payment
srv.AddPayableTool(paidTool, paidHandler, paymentReqs...)
```

### 2. Dynamic Pricing

```go
func calculatePrice(params map[string]any) string {
    // Base price
    price := 10000  // 0.01 USDC
    
    // Adjust based on parameters
    if params["premium"] == true {
        price *= 10
    }
    
    return fmt.Sprintf("%d", price)
}

srv.AddPayableTool(
    tool,
    handler,
    server.RequireUSDCBase(wallet, calculatePrice(params), "Dynamic pricing"),
)
```

### 3. Budget Controls

```go
signer := evm.NewPrivateKeySigner(key,
    evm.WithChain(evm.ChainBase),
    evm.WithToken(evm.TokenUSDC),
    evm.WithMaxAmount("100000000"),  // Max 100 USDC per payment
    evm.WithDailyLimit("1000000000"), // Max 1000 USDC per day
)
```

## Troubleshooting

### Payment Rejected
- Check wallet balance
- Verify token approval (EVM)
- Check network selection matches server

### Facilitator Errors
- Ensure facilitator URL is reachable
- Check for network-specific issues
- Verify payment amounts match exactly

### Session Issues
- MCP session may timeout after inactivity
- Re-initialize if session terminated
- Check server logs for details

## Next Steps

1. Review the [API specification](contracts/mcp-x402-api.yaml)
2. Explore [examples](../../examples/mcp/)
3. Read the [data model](data-model.md) for advanced usage
4. Check [research notes](research.md) for design decisions