// Package main demonstrates verification-only mode for custom payment logic.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/facilitator"
	x402http "github.com/mark3labs/x402-go/http"
)

func main() {
	// Configure middleware in verification-only mode
	config := &x402http.Config{
		FacilitatorURL: "https://facilitator.x402.com",
		VerifyOnly:     true, // Only verify, don't settle automatically
		PaymentRequirements: []x402.PaymentRequirement{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000", // 0.01 USDC
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/custom",
				Description:       "Custom Payment Logic Example",
				MaxTimeoutSeconds: 60,
			},
		},
	}

	// Create middleware
	middleware := x402http.NewX402Middleware(config)

	// Handler with custom payment logic
	customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Payment has been verified but NOT settled yet
		// Access payment info from context
		paymentInfo := r.Context().Value(x402http.PaymentContextKey)

		if paymentInfo == nil {
			http.Error(w, "No payment info in context", http.StatusInternalServerError)
			return
		}

		// Cast to facilitator.VerifyResponse to get payer info
		verifyResp, ok := paymentInfo.(*facilitator.VerifyResponse)
		if !ok {
			http.Error(w, "Invalid payment info", http.StatusInternalServerError)
			return
		}

		// Custom business logic here
		// For example: check user account, rate limiting, etc.
		payer := verifyResp.Payer

		// Simulate custom logic
		if shouldGrantAccess(payer) {
			w.Header().Set("Content-Type", "application/json")
			if _, err := fmt.Fprintf(w, `{
				"message": "Access granted after verification",
				"payer": "%s",
				"status": "verified_not_settled",
				"note": "Payment verified but not yet settled on-chain. You can settle it manually based on your business logic.",
				"data": "Custom business data here"
			}`, payer); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
			}
		} else {
			http.Error(w, "Access denied based on custom logic", http.StatusForbidden)
		}
	})

	// Apply middleware
	http.Handle("/custom", middleware(customHandler))

	// Info endpoint
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{
			"service": "Verification-Only Mode Example",
			"mode": "verify_only",
			"description": "Payment authorizations are verified but not automatically settled",
			"use_cases": [
				"Custom authorization logic before settlement",
				"Batch settlements",
				"Conditional access based on business rules",
				"Testing payment flows without on-chain transactions"
			],
			"endpoint": "/custom"
		}`)); err != nil {
			log.Printf("Failed to write info response: %v", err)
		}
	})

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Failed to write health response: %v", err)
		}
	})

	// Start server
	port := ":8080"
	fmt.Printf("Verification-only mode server starting on %s\n", port)
	fmt.Println("\nMode: VERIFY ONLY (no automatic settlement)")
	fmt.Println("Payments will be verified but NOT settled on-chain automatically.")
	fmt.Println("\nEndpoints:")
	fmt.Println("  - GET http://localhost:8080/custom (payment required - verify only)")
	fmt.Println("  - GET http://localhost:8080/info (info about verification mode)")
	fmt.Println("  - GET http://localhost:8080/health (health check)")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

// shouldGrantAccess is a placeholder for custom business logic
func shouldGrantAccess(payer string) bool {
	// Example custom logic:
	// - Check if payer is in allowlist/blocklist
	// - Check rate limits
	// - Verify account status
	// - Apply custom rules

	// For demo purposes, always grant access
	return true
}
