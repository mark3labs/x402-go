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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/evm"
	x402http "github.com/mark3labs/x402-go/http"
	ginx402 "github.com/mark3labs/x402-go/http/gin"
	"github.com/mark3labs/x402-go/svm"
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
	fmt.Println("gin-example - Gin-based x402 payment client and server")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gin-example server [flags]  - Run a Gin server with paywalled endpoints")
	fmt.Println("  gin-example client [flags]  - Run client to access paywalled resources")
	fmt.Println()
	fmt.Println("Run 'gin-example server --help' or 'gin-example client --help' for more information.")
}

func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	port := fs.String("port", "8080", "Server port")
	network := fs.String("network", "base", "Network to accept payments on (base, base-sepolia, solana, solana-devnet)")
	payTo := fs.String("payTo", "", "Address to receive payments (required)")
	tokenAddr := fs.String("token", "", "Token address (auto-detected based on network if not specified)")
	amount := fs.String("amount", "", "Payment amount in atomic units (auto-detected based on network if not specified)")
	facilitatorURL := fs.String("facilitator", "https://facilitator.x402.rs", "Facilitator URL")
	verbose := fs.Bool("verbose", false, "Enable verbose debug output")

	_ = fs.Parse(args)

	// Validate required flags
	if *payTo == "" {
		fmt.Println("Error: --payTo is required")
		fmt.Println()
		fs.PrintDefaults()
		os.Exit(1)
	}

	// Set defaults based on network if not specified
	if *tokenAddr == "" {
		switch strings.ToLower(*network) {
		case "solana":
			*tokenAddr = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" // USDC on Solana mainnet
		case "solana-devnet":
			*tokenAddr = "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU" // USDC on Solana devnet
		case "base", "base-sepolia":
			*tokenAddr = "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913" // USDC on Base
		default:
			*tokenAddr = "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913" // Default to Base USDC
		}
	}

	if *amount == "" {
		*amount = "1000" // Default: 0.001 USDC (6 decimals)
	}

	fmt.Printf("Starting Gin server with x402 on port %s\n", *port)
	fmt.Printf("Network: %s\n", *network)
	fmt.Printf("Payment recipient: %s\n", *payTo)
	fmt.Printf("Payment amount: %s atomic units\n", *amount)
	fmt.Printf("Token: %s\n", *tokenAddr)
	fmt.Printf("Facilitator: %s\n", *facilitatorURL)
	if *verbose {
		fmt.Printf("Verbose mode: ENABLED\n")
	}
	fmt.Println()

	// Set Gin mode
	if !*verbose {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.Default()

	// Create payment requirements
	requirement := x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           *network,
		MaxAmountRequired: *amount,
		Asset:             *tokenAddr,
		PayTo:             *payTo,
		MaxTimeoutSeconds: 60,
	}

	// Create x402 middleware config
	config := &x402http.Config{
		FacilitatorURL:      *facilitatorURL,
		PaymentRequirements: []x402.PaymentRequirement{requirement},
		VerifyOnly:          false,
	}

	// Public endpoint (no payment required)
	r.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "This is a free public endpoint",
			"info":    "Try /data endpoint to test x402 payments",
		})
	})

	// Paywalled endpoint
	r.GET("/data", ginx402.NewGinX402Middleware(config), func(c *gin.Context) {
		// Access payment information from context
		response := gin.H{
			"message":   "Successfully accessed paywalled content!",
			"timestamp": time.Now().Format(time.RFC3339),
			"data": gin.H{
				"premium": true,
				"secret":  "This is premium data that requires payment",
			},
		}

		if paymentInfo, exists := c.Get("x402_payment"); exists {
			verifyResp := paymentInfo.(*x402http.VerifyResponse)
			response["payer"] = verifyResp.Payer
		}

		c.JSON(http.StatusOK, response)
	})

	// Info endpoint
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Gin x402 Demo Server\n\nEndpoints:\n  GET /data    - Paywalled endpoint (requires x402 payment)\n  GET /public  - Free public endpoint\n")
	})

	fmt.Println("Server endpoints:")
	fmt.Printf("  http://localhost:%s/       - Server info\n", *port)
	fmt.Printf("  http://localhost:%s/data   - Paywalled endpoint (requires payment)\n", *port)
	fmt.Printf("  http://localhost:%s/public - Free public endpoint\n", *port)
	fmt.Println()
	fmt.Println("Server is ready!")

	// Start server
	if err := r.Run(":" + *port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func runClient(args []string) {
	fs := flag.NewFlagSet("client", flag.ExitOnError)
	network := fs.String("network", "base", "Network to use (base, base-sepolia, solana, solana-devnet)")
	key := fs.String("key", "", "Private key (hex for EVM, base58 for Solana)")
	keyFile := fs.String("keyfile", "", "Solana keygen JSON file (alternative to --key for Solana)")
	url := fs.String("url", "", "URL to fetch (must be paywalled with x402)")
	tokenAddr := fs.String("token", "", "Token address (auto-detected based on network if not specified)")
	maxAmount := fs.String("max", "", "Maximum amount per call (optional)")
	verbose := fs.Bool("verbose", false, "Enable verbose debug output")

	_ = fs.Parse(args)

	// Validate inputs
	if *key == "" && *keyFile == "" {
		fmt.Println("Error: --key or --keyfile is required")
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

	// Set defaults based on network if not specified
	if *tokenAddr == "" {
		switch strings.ToLower(*network) {
		case "solana":
			*tokenAddr = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" // USDC on Solana mainnet
		case "solana-devnet":
			*tokenAddr = "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU" // USDC on Solana devnet
		case "base", "base-sepolia":
			*tokenAddr = "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913" // USDC on Base
		default:
			*tokenAddr = "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913" // Default to Base USDC
		}
	}

	var client *x402http.Client
	var signerAddress string

	// Create appropriate signer based on network
	if strings.HasPrefix(strings.ToLower(*network), "solana") {
		// Create Solana signer
		var svmOpts []svm.SignerOption

		if *keyFile != "" {
			svmOpts = append(svmOpts, svm.WithKeygenFile(*keyFile))
		} else {
			svmOpts = append(svmOpts, svm.WithPrivateKey(*key))
		}

		svmOpts = append(svmOpts,
			svm.WithNetwork(*network),
			svm.WithToken(*tokenAddr, "USDC", 6),
		)

		if *maxAmount != "" {
			svmOpts = append(svmOpts, svm.WithMaxAmountPerCall(*maxAmount))
		}

		svmSigner, err := svm.NewSigner(svmOpts...)
		if err != nil {
			log.Fatalf("Failed to create Solana signer: %v", err)
		}

		signerAddress = svmSigner.Address()
		fmt.Printf("Created Solana signer for address: %s\n", signerAddress)

		// Create x402-enabled HTTP client with Solana signer
		client, err = x402http.NewClient(
			x402http.WithSigner(svmSigner),
		)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
	} else {
		// Create EVM signer
		signerOpts := []evm.SignerOption{
			evm.WithPrivateKey(*key),
			evm.WithNetwork(*network),
			evm.WithToken(*tokenAddr, "USDC", 6),
		}

		if *maxAmount != "" {
			signerOpts = append(signerOpts, evm.WithMaxAmountPerCall(*maxAmount))
		}

		evmSigner, err := evm.NewSigner(signerOpts...)
		if err != nil {
			log.Fatalf("Failed to create EVM signer: %v", err)
		}

		signerAddress = evmSigner.Address().Hex()
		fmt.Printf("Created EVM signer for address: %s\n", signerAddress)

		// Create x402-enabled HTTP client with EVM signer
		client, err = x402http.NewClient(
			x402http.WithSigner(evmSigner),
		)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
	}

	fmt.Printf("Network: %s\n", *network)
	fmt.Printf("Token: %s\n", *tokenAddr)

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
