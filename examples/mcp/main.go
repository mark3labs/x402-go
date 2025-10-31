package main

import (
	"fmt"
)

// This example demonstrates how to use the x402-go MCP integration.
//
// The MCP integration (mcp/client and mcp/server packages) provides:
// - X402Transport: MCP client transport with automatic payment handling
// - X402Server: MCP server with tool payment protection
// - Multi-chain payment support (Base, Polygon, Avalanche, Solana)
// - Automatic payment fallback and retry logic
//
// USAGE:
//
// Server Mode:
//   Create an MCP server with x402 payment requirements:
//
//     server := mcpserver.NewX402Server("my-server", "1.0.0", &mcpserver.Config{
//         FacilitatorURL: "https://facilitator.x402.rs",
//     })
//
//     // Add free tool (no payment required)
//     server.AddTool(tool, handler)
//
//     // Add paid tool (requires payment)
//     requirement := x402.PaymentRequirement{
//         Scheme:            "exact",
//         Network:           "base",
//         MaxAmountRequired: "10000",  // 0.01 USDC
//         Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
//         PayTo:             recipientAddress,
//         MaxTimeoutSeconds: 60,
//     }
//     server.AddPayableTool(tool, handler, requirement)
//
// Client Mode:
//   Create an MCP client with automatic x402 payment support:
//
//     // Create signer (EVM or Solana)
//     evmSigner, _ := evm.NewSigner(
//         evm.WithPrivateKey(privateKey),
//         evm.WithNetwork("base"),
//         evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
//     )
//
//     // Wrap base transport with x402 support
//     baseTransport := transport.NewStdio("", nil)
//     x402Transport, _ := mcpclient.NewX402Transport(baseTransport, []x402.Signer{evmSigner})
//
//     // Create MCP client
//     mcpClient := client.NewClient(x402Transport)
//
//     // Use client normally - payments are handled automatically
//     response, _ := mcpClient.CallTool(ctx, mcp.CallToolRequest{
//         Params: mcp.CallToolParams{
//             Name: "premium-tool",
//             Arguments: map[string]any{"query": "search term"},
//         },
//     })
//
// FEATURES:
// - Automatic 402 error detection and payment retry
// - Multi-chain support (Base, Polygon, Avalanche, Solana)
// - Payment fallback when primary signer fails
// - 5-second payment verification timeout
// - 60-second payment settlement timeout
// - Concurrent request support (each gets independent payment)
//
// For a complete working example, see:
// - github.com/mark3labs/x402-go/mcp/client/transport_test.go
// - github.com/mark3labs/x402-go/mcp/server/server_test.go
//
// NOTE: This is a documentation-only example. The actual server/client
// implementation requires the full mcp-go and x402-go packages to be
// properly integrated. See the test files for complete working examples.

func main() {
	fmt.Println("x402-go MCP Integration Example")
	fmt.Println()
	fmt.Println("This example demonstrates the x402-go MCP integration API.")
	fmt.Println("For working code examples, see:")
	fmt.Println("  - mcp/client/transport_test.go")
	fmt.Println("  - mcp/server/server_test.go")
	fmt.Println()
	fmt.Println("Key Features:")
	fmt.Println("  ✓ Automatic payment handling for MCP tools")
	fmt.Println("  ✓ Multi-chain support (Base, Polygon, Avalanche, Solana)")
	fmt.Println("  ✓ Payment fallback and retry logic")
	fmt.Println("  ✓ 5s verification and 60s settlement timeouts")
	fmt.Println()
	fmt.Println("See source code comments above for usage examples.")
}
