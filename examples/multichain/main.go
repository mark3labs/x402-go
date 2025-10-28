// Package main demonstrates multi-chain payment configuration.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mark3labs/x402-go"
	x402http "github.com/mark3labs/x402-go/http"
)

func main() {
	// Configure payment requirements for premium data (higher price)
	premiumConfig := &x402http.Config{
		FacilitatorURL: "https://facilitator.x402.com",
		PaymentRequirements: []x402.PaymentRequirement{
			// Option 1: Base (EVM chain)
			{
				Scheme:            "exact",
				Network:           "base",
				MaxAmountRequired: "1000000",                                    // 1 USDC
				Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // Base USDC
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/data",
				Description:       "Premium Data Access (Base)",
				MaxTimeoutSeconds: 60,
			},
			// Option 2: Solana (SVM chain)
			{
				Scheme:            "exact",
				Network:           "solana",
				MaxAmountRequired: "1000000",                                      // 1 USDC
				Asset:             "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // Solana USDC
				PayTo:             "YourSolanaWalletAddress",
				Resource:          "https://api.example.com/data",
				Description:       "Premium Data Access (Solana)",
				MaxTimeoutSeconds: 60,
				Extra: map[string]any{
					"feePayer": "FacilitatorSolanaAddress",
				},
			},
			// Option 3: Base Sepolia testnet (EVM chain)
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000", // 0.01 USDC (for testing)
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/data",
				Description:       "Premium Data Access (Base Sepolia Testnet)",
				MaxTimeoutSeconds: 60,
			},
		},
	}

	// Configure payment requirements for basic data (lower price)
	basicConfig := &x402http.Config{
		FacilitatorURL: "https://facilitator.x402.com",
		PaymentRequirements: []x402.PaymentRequirement{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "1000", // 0.001 USDC (cheaper for basic)
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/basic",
				Description:       "Basic Data Access",
				MaxTimeoutSeconds: 60,
			},
		},
	}

	// Create middleware instances with different pricing
	premiumMiddleware := x402http.NewX402Middleware(premiumConfig)
	basicMiddleware := x402http.NewX402Middleware(basicConfig)

	// Premium data handler (high price)
	premiumHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{
			"message": "Premium data access granted",
			"tier": "premium",
			"price": "1 USDC",
			"data": {
				"timestamp": "2025-10-28T12:00:00Z",
				"metric": "premium_metric",
				"value": 42.5,
				"details": "Full details with premium insights"
			}
		}`)); err != nil {
			log.Printf("Failed to write premium response: %v", err)
		}
	})

	// Basic data handler (low price)
	basicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{
			"message": "Basic data access granted",
			"tier": "basic",
			"price": "0.001 USDC",
			"data": {
				"timestamp": "2025-10-28T12:00:00Z",
				"metric": "basic_metric",
				"value": 42.5
			}
		}`)); err != nil {
			log.Printf("Failed to write basic response: %v", err)
		}
	})

	// Apply different middleware to different routes
	http.Handle("/premium", premiumMiddleware(premiumHandler))
	http.Handle("/basic", basicMiddleware(basicHandler))

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Failed to write health response: %v", err)
		}
	})

	// Info endpoint showing pricing tiers
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{
			"service": "Multi-Chain Payment API",
			"tiers": [
				{
					"endpoint": "/basic",
					"price": "0.001 USDC",
					"networks": ["base-sepolia"],
					"description": "Basic data access"
				},
				{
					"endpoint": "/premium",
					"price": "1 USDC",
					"networks": ["base", "solana", "base-sepolia"],
					"description": "Premium data with full details"
				}
			],
			"payment_options": [
				{"network": "base", "chain_type": "EVM"},
				{"network": "solana", "chain_type": "SVM"},
				{"network": "base-sepolia", "chain_type": "EVM (testnet)"}
			],
			"note": "Make a request without payment to any endpoint to see full requirements"
		}`)); err != nil {
			log.Printf("Failed to write info response: %v", err)
		}
	})

	// Start server
	port := ":8080"
	fmt.Printf("Multi-chain payment server with route-specific pricing starting on %s\n", port)
	fmt.Println("\nPricing tiers:")
	fmt.Println("  - /basic: 0.001 USDC (base-sepolia)")
	fmt.Println("  - /premium: 1 USDC (base, solana, base-sepolia)")
	fmt.Println("\nSupported networks:")
	fmt.Println("  - Base (EVM)")
	fmt.Println("  - Solana (SVM)")
	fmt.Println("  - Base Sepolia (EVM testnet)")
	fmt.Println("\nEndpoints:")
	fmt.Println("  - GET http://localhost:8080/basic (payment required - lower price)")
	fmt.Println("  - GET http://localhost:8080/premium (payment required - higher price)")
	fmt.Println("  - GET http://localhost:8080/info (info about pricing tiers)")
	fmt.Println("  - GET http://localhost:8080/health (health check)")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
