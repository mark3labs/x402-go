package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/mcp/client"
	"github.com/mark3labs/x402-go/mcp/server"
	"github.com/mark3labs/x402-go/signers/evm"
)

var (
	mode        = flag.String("mode", "", "Mode: 'server' or 'client'")
	port        = flag.String("port", "8080", "Server port")
	serverURL   = flag.String("server", "http://localhost:8080", "Server URL (client mode)")
	privateKey  = flag.String("key", "", "Private key for payments (client mode)")
	payTo       = flag.String("pay-to", "", "Payment address (server mode)")
	facilitator = flag.String("facilitator", "https://facilitator.x402.rs", "Facilitator URL")
	verifyOnly  = flag.Bool("verify-only", false, "Verify only, don't settle payments")
	testnet     = flag.Bool("testnet", false, "Use testnet (Base Sepolia)")
	network     = flag.String("network", "base", "Network name")
	verbose     = flag.Bool("v", false, "Verbose logging")
)

func main() {
	flag.Parse()

	if *mode == "" {
		fmt.Println("Usage: mcp -mode [server|client] [options]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	switch *mode {
	case "server":
		runServer()
	case "client":
		runClient()
	default:
		log.Fatalf("Invalid mode: %s (must be 'server' or 'client')", *mode)
	}
}

func runServer() {
	if *payTo == "" {
		log.Fatal("Server mode requires -pay-to address")
	}

	// Determine network and payment requirements
	var (
		requirement func(string, string, string) x402.PaymentRequirement
		networkName string
	)

	switch {
	case *testnet:
		requirement = server.RequireUSDCBaseSepolia
		networkName = "base-sepolia"
	case *network == "base":
		requirement = server.RequireUSDCBase
		networkName = "base"
	case *network == "polygon":
		requirement = server.RequireUSDCPolygon
		networkName = "polygon"
	case *network == "solana":
		requirement = server.RequireUSDCSolana
		networkName = "solana"
	default:
		log.Fatalf("unsupported network: %s", *network)
	}

	// Create x402 MCP server
	config := &server.Config{
		FacilitatorURL: *facilitator,
		VerifyOnly:     *verifyOnly,
		Verbose:        *verbose,
	}

	srv := server.NewX402Server("x402-mcp-example", "1.0.0", config)

	// Add free tool: echo
	echoTool := mcp.NewTool(
		"echo",
		mcp.WithDescription("Echo back the input message"),
		mcp.WithString("message", mcp.Required(), mcp.Description("Message to echo")),
	)
	srv.AddTool(echoTool, echoHandler)

	// Add paid tool: search
	searchTool := mcp.NewTool(
		"search",
		mcp.WithDescription("Premium search service"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of results")),
	)
	err := srv.AddPayableTool(
		searchTool,
		searchHandler,
		requirement(*payTo, "10000", "Premium search - 0.01 USDC"),
	)
	if err != nil {
		log.Fatalf("Failed to add payable tool: %v", err)
	}

	// Start server
	addr := ":" + *port
	log.Printf("Starting x402 MCP server on %s", addr)
	log.Printf("Network: %s", networkName)
	log.Printf("Payment address: %s", *payTo)
	log.Printf("Facilitator: %s", *facilitator)
	log.Printf("Verify-only: %v", *verifyOnly)
	log.Printf("Tools: echo (free), search (0.01 USDC)")

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
		os.Exit(0)
	}()

	if err := srv.Start(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func echoHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	message, _ := args["message"].(string)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(fmt.Sprintf("Echo: %s", message)),
		},
	}, nil
}

func searchHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	query, _ := args["query"].(string)
	maxResults := 5
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// Simulate search results
	results := fmt.Sprintf("Premium search results for '%s' (max %d results):\n", query, maxResults)
	results += fmt.Sprintf("1. Result about %s\n", query)
	results += fmt.Sprintf("2. Another result for %s\n", query)
	results += fmt.Sprintf("3. More information on %s\n", query)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(results),
		},
	}, nil
}

func runClient() {
	if *privateKey == "" {
		log.Fatal("Client mode requires -key (private key)")
	}

	// Determine chain configuration
	var chain x402.ChainConfig
	switch {
	case *testnet:
		chain = x402.BaseSepolia
	case *network == "base":
		chain = x402.BaseMainnet
	case *network == "polygon":
		chain = x402.PolygonMainnet
	case *network == "solana":
		chain = x402.SolanaMainnet
	default:
		log.Fatalf("unsupported network: %s", *network)
	}

	// Create EVM signer with network and USDC token
	signer, err := evm.NewSigner(
		evm.WithPrivateKey(*privateKey),
		evm.WithNetwork(chain.NetworkID),
		evm.WithToken(chain.USDCAddress, "USDC", 6), // USDC has 6 decimals
	)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	// Create x402 transport
	transport, err := client.NewTransport(
		*serverURL,
		client.WithSigner(signer),
		client.WithPaymentCallback(paymentLogger),
	)
	if err != nil {
		log.Fatalf("Failed to create transport: %v", err)
	}

	// Create MCP client
	mcpClient := mcpclient.NewClient(transport)

	ctx := context.Background()

	// Start connection
	if err := mcpClient.Start(ctx); err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}
	defer mcpClient.Close()

	log.Printf("Connected to MCP server at %s", *serverURL)

	// Initialize session
	initResp, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "2024-11-05",
			ClientInfo: mcp.Implementation{
				Name:    "x402-example-client",
				Version: "1.0.0",
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	})
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	log.Printf("Session initialized: %s v%s", initResp.ServerInfo.Name, initResp.ServerInfo.Version)

	// List available tools
	toolsResp, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}

	log.Printf("Available tools:")
	for _, tool := range toolsResp.Tools {
		log.Printf("  - %s: %s", tool.Name, tool.Description)
	}

	// Call free tool (echo)
	log.Println("\n=== Calling free tool: echo ===")
	echoResult, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "echo",
			Arguments: map[string]interface{}{
				"message": "Hello from x402 MCP client!",
			},
		},
	})
	if err != nil {
		log.Fatalf("Echo call failed: %v", err)
	}
	log.Printf("Echo result: %v", echoResult.Content[0])

	// Call paid tool (search) - payment handled automatically
	log.Println("\n=== Calling paid tool: search (requires payment) ===")
	searchResult, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "search",
			Arguments: map[string]interface{}{
				"query":       "blockchain",
				"max_results": 3,
			},
		},
	})
	if err != nil {
		log.Fatalf("Search call failed: %v", err)
	}
	log.Printf("Search result: %v", searchResult.Content[0])

	log.Println("\n=== Example completed successfully ===")
}

func paymentLogger(event client.PaymentEvent) {
	switch event.Type {
	case client.PaymentAttempt:
		log.Printf("[PAYMENT] Attempting payment: %s %s on %s to %s",
			event.Amount, event.Asset, event.Network, event.Recipient)
	case client.PaymentSuccess:
		log.Printf("[PAYMENT] Payment successful on %s", event.Network)
		if event.Transaction != "" {
			log.Printf("[PAYMENT] Transaction: %s", event.Transaction)
		}
	case client.PaymentFailure:
		log.Printf("[PAYMENT] Payment failed: %v", event.Error)
	}
}
