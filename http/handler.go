// Package http provides HTTP middleware for x402 payment gating in Go HTTP servers.
// This package enables servers to protect HTTP endpoints with cryptocurrency payment requirements
// and verify payments before granting access to protected resources.
//
// # Quick Start
//
// Create middleware to protect endpoints:
//
//	middleware := http.NewMiddleware(
//		http.WithFacilitator("https://facilitator.example.com"),
//		http.WithPaymentRequirement(x402.PaymentRequirement{
//			Scheme:            "exact",
//			Network:           "base",
//			MaxAmountRequired: "1000000", // 1 USDC (6 decimals)
//			Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
//			PayTo:             "0xYourAddress",
//			MimeType:          "application/json",
//			MaxTimeoutSeconds: 300,
//		}),
//	)
//
//	http.Handle("/protected", middleware(yourHandler))
//
// # Features
//
// - Automatic 402 Payment Required responses with payment requirements
// - Payment verification via facilitator
// - Support for multiple payment options (different networks/tokens)
// - Flexible payment requirement configuration
// - Standard net/http compatibility
//
// # Payment Flow
//
// 1. Client requests protected endpoint without payment
// 2. Middleware responds with 402 and payment requirements
// 3. Client signs payment and retries with X-PAYMENT header
// 4. Middleware verifies payment with facilitator
// 5. On success, request proceeds to handler
//
// See examples/basic/ for complete usage examples.
package http

import (
	"net/http"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/http/internal/helpers"
)

// sendPaymentRequired sends a 402 Payment Required response with payment requirements.
// It delegates to sendPaymentRequiredWithRequirements with the configured payment requirements.
func sendPaymentRequired(w http.ResponseWriter, config *Config) {
	sendPaymentRequiredWithRequirements(w, config.PaymentRequirements)
}

// sendPaymentRequiredWithRequirements sends a 402 Payment Required response with specific payment requirements.
func sendPaymentRequiredWithRequirements(w http.ResponseWriter, requirements []x402.PaymentRequirement) {
	helpers.SendPaymentRequired(w, requirements)
}

// parsePaymentHeader parses the X-PAYMENT header and returns the payment payload.
func parsePaymentHeader(r *http.Request) (x402.PaymentPayload, error) {
	return helpers.ParsePaymentHeaderFromRequest(r)
}

// findMatchingRequirement finds a payment requirement that matches the provided payment.
func findMatchingRequirement(payment x402.PaymentPayload, requirements []x402.PaymentRequirement) (x402.PaymentRequirement, error) {
	return helpers.FindMatchingRequirement(payment, requirements)
}

// addPaymentResponseHeader adds the X-PAYMENT-RESPONSE header with settlement information.
func addPaymentResponseHeader(w http.ResponseWriter, settlement *x402.SettlementResponse) error {
	return helpers.AddPaymentResponseHeader(w, settlement)
}
