// Package main demonstrates basic usage of the x402 payment middleware.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mark3labs/x402-go"
	x402http "github.com/mark3labs/x402-go/http"
)

func main() {
	// Configure payment requirements
	config := &x402http.Config{
		FacilitatorURL: "https://facilitator.x402.com",
		PaymentRequirements: []x402.PaymentRequirement{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",                                      // 0.01 USDC (6 decimals)
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e", // USDC on Base Sepolia
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C", // Your wallet address
				Resource:          "https://api.example.com/premium",
				Description:       "Premium API Access",
				MaxTimeoutSeconds: 60,
			},
		},
	}

	// Create middleware
	paymentMiddleware := x402http.NewX402Middleware(config)

	// Protected handler
	premiumHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Welcome to premium content!", "status": "success"}`))
	})

	// Public handler (no payment required)
	publicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "This is free content", "status": "success"}`))
	})

	// Apply middleware to protected route
	http.Handle("/premium", paymentMiddleware(premiumHandler))
	http.Handle("/public", publicHandler)

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	port := ":8080"
	fmt.Printf("Server starting on %s\n", port)
	fmt.Println("Try accessing:")
	fmt.Println("  - GET http://localhost:8080/public (no payment required)")
	fmt.Println("  - GET http://localhost:8080/premium (payment required)")
	fmt.Println("  - GET http://localhost:8080/health (health check)")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
