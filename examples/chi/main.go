package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mark3labs/x402-go"
	x402http "github.com/mark3labs/x402-go/http"
	chix402 "github.com/mark3labs/x402-go/http/chi"
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8080", "Server port")
	network := flag.String("network", "base", "Network to accept payments on (base, base-sepolia, solana, solana-devnet)")
	payTo := flag.String("payTo", "", "Address to receive payments (required)")
	tokenAddr := flag.String("token", "", "Token address (auto-detected based on network if not specified)")
	amount := flag.String("amount", "", "Payment amount in atomic units (auto-detected based on network if not specified)")
	facilitatorURL := flag.String("facilitator", "https://facilitator.x402.rs", "Facilitator URL")

	flag.Parse()

	// Validate required flags
	if *payTo == "" {
		fmt.Println("Error: --payTo is required")
		fmt.Println()
		flag.PrintDefaults()
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

	fmt.Printf("Starting Chi server with x402 on port %s\n", *port)
	fmt.Printf("Network: %s\n", *network)
	fmt.Printf("Payment recipient: %s\n", *payTo)
	fmt.Printf("Payment amount: %s atomic units\n", *amount)
	fmt.Printf("Token: %s\n", *tokenAddr)
	fmt.Printf("Facilitator: %s\n", *facilitatorURL)
	fmt.Println()

	// Create Chi router
	r := chi.NewRouter()

	// Add standard Chi middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Create payment requirement
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
	r.Get("/public", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"message": "This is a free public endpoint",
			"info":    "Try /data endpoint to test x402 payments",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	// Paywalled endpoint group
	r.Route("/", func(r chi.Router) {
		// Apply x402 middleware to this group
		r.Use(chix402.NewChiX402Middleware(config))

		r.Get("/data", func(w http.ResponseWriter, r *http.Request) {
			// Access payment information from context
			response := map[string]interface{}{
				"message":   "Successfully accessed paywalled content!",
				"timestamp": time.Now().Format(time.RFC3339),
				"data": map[string]interface{}{
					"premium": true,
					"secret":  "This is premium data that requires payment",
				},
			}

			// Get payment info from context
			if paymentInfo := r.Context().Value(x402http.PaymentContextKey); paymentInfo != nil {
				verifyResp := paymentInfo.(*x402http.VerifyResponse)
				response["payer"] = verifyResp.Payer
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		})
	})

	// Info endpoint (public)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Chi x402 Demo Server\n\nEndpoints:\n  GET /data    - Paywalled endpoint (requires x402 payment)\n  GET /public  - Free public endpoint\n")
	})

	fmt.Println("Server endpoints:")
	fmt.Printf("  http://localhost:%s/       - Server info\n", *port)
	fmt.Printf("  http://localhost:%s/data   - Paywalled endpoint (requires payment)\n", *port)
	fmt.Printf("  http://localhost:%s/public - Free public endpoint\n", *port)
	fmt.Println()
	fmt.Println("Server is ready!")

	// Start server
	if err := http.ListenAndServe(":"+*port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
