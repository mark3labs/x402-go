package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/evm"
	x402http "github.com/mark3labs/x402-go/http"
)

func main() {
	// Subcommand handling
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		runServer(os.Args[2:])
	case "client":
		runClient(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("x402demo - Example x402 payment client and server")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  x402demo server [flags]  - Run a test server with paywalled endpoints")
	fmt.Println("  x402demo client [flags]  - Run client to access paywalled resources")
	fmt.Println()
	fmt.Println("Run 'x402demo server --help' or 'x402demo client --help' for more information.")
}

func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	port := fs.String("port", "8080", "Server port")
	network := fs.String("network", "base", "Network to accept payments on (base, base-sepolia)")
	payTo := fs.String("payTo", "", "Address to receive payments (required)")
	tokenAddr := fs.String("token", "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "Token address (default: USDC on Base)")
	amount := fs.String("amount", "100000", "Payment amount in atomic units (default: 0.1 USDC)")
	facilitatorURL := fs.String("facilitator", "https://facilitator.x402.rs", "Facilitator URL")
	verbose := fs.Bool("verbose", false, "Enable verbose debug output")

	fs.Parse(args)

	// Validate required flags
	if *payTo == "" {
		fmt.Println("Error: --payTo is required")
		fmt.Println()
		fs.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("Starting x402 demo server on port %s\n", *port)
	fmt.Printf("Network: %s\n", *network)
	fmt.Printf("Payment recipient: %s\n", *payTo)
	fmt.Printf("Payment amount: %s atomic units\n", *amount)
	fmt.Printf("Token: %s\n", *tokenAddr)
	fmt.Printf("Facilitator: %s\n", *facilitatorURL)
	if *verbose {
		fmt.Printf("Verbose mode: ENABLED\n")
	}
	fmt.Println()

	// Create payment requirements
	requirements := []x402.PaymentRequirement{
		{
			Scheme:            "exact",
			Network:           *network,
			MaxAmountRequired: *amount,
			Asset:             *tokenAddr,
			PayTo:             *payTo,
			MaxTimeoutSeconds: 60,
		},
	}

	// Create x402 middleware
	middleware := x402http.NewX402Middleware(&x402http.Config{
		FacilitatorURL:      *facilitatorURL,
		PaymentRequirements: requirements,
		VerifyOnly:          false,
	})

	// Create paywalled data handler
	dataHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This endpoint requires payment
		response := map[string]interface{}{
			"message":   "Successfully accessed paywalled content!",
			"timestamp": time.Now().Format(time.RFC3339),
			"data": map[string]interface{}{
				"premium": true,
				"secret":  "This is premium data that requires payment",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Create free public handler
	publicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"message": "This is a free public endpoint",
			"info":    "Try /data endpoint to test x402 payments",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Setup routes
	mux := http.NewServeMux()
	mux.Handle("/data", middleware(dataHandler)) // Paywalled endpoint
	mux.Handle("/public", publicHandler)         // Free endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "x402 Demo Server\n\n")
		fmt.Fprintf(w, "Endpoints:\n")
		fmt.Fprintf(w, "  GET /data    - Paywalled endpoint (requires x402 payment)\n")
		fmt.Fprintf(w, "  GET /public  - Free public endpoint\n")
	})

	fmt.Println("Server endpoints:")
	fmt.Printf("  http://localhost:%s/       - Server info\n", *port)
	fmt.Printf("  http://localhost:%s/data   - Paywalled endpoint (requires payment)\n", *port)
	fmt.Printf("  http://localhost:%s/public - Free public endpoint\n", *port)
	fmt.Println()
	fmt.Println("Server is ready!")

	// Start server
	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func runClient(args []string) {
	fs := flag.NewFlagSet("client", flag.ExitOnError)
	network := fs.String("network", "base", "Network to use (base, base-sepolia)")
	key := fs.String("key", "", "Private key (hex format, with or without 0x prefix)")
	url := fs.String("url", "", "URL to fetch (must be paywalled with x402)")
	tokenAddr := fs.String("token", "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "Token address (default: USDC on Base)")
	maxAmount := fs.String("max", "", "Maximum amount per call (optional)")
	verbose := fs.Bool("verbose", false, "Enable verbose debug output")

	fs.Parse(args)

	// Validate inputs
	if *key == "" {
		fmt.Println("Error: --key is required")
		fmt.Println()
		fs.PrintDefaults()
		os.Exit(1)
	}

	if *url == "" {
		fmt.Println("Error: --url is required")
		fmt.Println()
		fs.PrintDefaults()
		os.Exit(1)
	}

	// Create EVM signer
	signerOpts := []evm.SignerOption{
		evm.WithPrivateKey(*key),
		evm.WithNetwork(*network),
		evm.WithToken(*tokenAddr, "USDC", 6),
	}

	if *maxAmount != "" {
		signerOpts = append(signerOpts, evm.WithMaxAmountPerCall(*maxAmount))
	}

	signer, err := evm.NewSigner(signerOpts...)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	fmt.Printf("Created signer for address: %s\n", signer.Address().Hex())
	fmt.Printf("Network: %s\n", *network)
	fmt.Printf("Token: %s\n", *tokenAddr)

	// Create x402-enabled HTTP client
	client, err := x402http.NewClient(
		x402http.WithSigner(signer),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Printf("\nFetching: %s\n", *url)

	// Make the request
	resp, err := client.Get(*url)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Verbose output: show payment header if sent
	if *verbose && resp.Request.Header.Get("X-PAYMENT") != "" {
		fmt.Println("\n=== DEBUG: Payment Header ===")
		paymentHeader := resp.Request.Header.Get("X-PAYMENT")
		fmt.Printf("X-PAYMENT (base64): %s\n", paymentHeader)
		fmt.Printf("Length: %d bytes\n", len(paymentHeader))

		// Decode and show the actual payload
		if decoded, err := base64.StdEncoding.DecodeString(paymentHeader); err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(decoded, &payload); err == nil {
				prettyJSON, _ := json.MarshalIndent(payload, "", "  ")
				fmt.Printf("\nDecoded Payload:\n%s\n", string(prettyJSON))
			}
		}
		fmt.Println("=============================")
	}

	// Check for settlement info
	if settlement := x402http.GetSettlement(resp); settlement != nil {
		if settlement.Success {
			fmt.Printf("\n✓ Payment successful!\n")
			fmt.Printf("  Transaction: %s\n", settlement.Transaction)
			fmt.Printf("  Network: %s\n", settlement.Network)
			fmt.Printf("  Payer: %s\n", settlement.Payer)
		} else {
			fmt.Printf("\n✗ Payment failed: %s\n", settlement.ErrorReason)
		}
	}

	// Display response
	fmt.Printf("\nResponse Status: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))

	// Show X-PAYMENT-RESPONSE header if present
	if paymentResp := resp.Header.Get("X-PAYMENT-RESPONSE"); paymentResp != "" {
		fmt.Printf("X-PAYMENT-RESPONSE: %s\n", paymentResp)
	}

	fmt.Println()

	// Read and display body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Println("Response Body:")
	fmt.Println(string(body))

	// Verbose: Show raw payment details if available
	if *verbose {
		fmt.Println("\n=== DEBUG: Request Details ===")
		fmt.Printf("Final URL: %s\n", resp.Request.URL)
		fmt.Printf("Method: %s\n", resp.Request.Method)
		fmt.Println("Headers:")
		for k, v := range resp.Request.Header {
			if k == "X-PAYMENT" {
				fmt.Printf("  %s: [PRESENT - %d bytes]\n", k, len(v[0]))
			} else {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}
		fmt.Println("==============================")
	}
}
