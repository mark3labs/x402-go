// Package chi provides Chi-compatible middleware for x402 payment gating.
// This package is a thin adapter that uses stdlib http.Handler interface
// and delegates all payment verification and settlement logic to shared helpers.
package chi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/mark3labs/x402-go"
	httpx402 "github.com/mark3labs/x402-go/http"
	"github.com/mark3labs/x402-go/http/internal/helpers"
)

// NewChiX402Middleware creates a new x402 payment middleware for Chi.
// It returns a Chi-compatible middleware function that wraps handlers with payment gating.
//
// The middleware:
//   - Bypasses OPTIONS requests for CORS preflight support
//   - Checks for X-PAYMENT header in requests
//   - Returns 402 Payment Required if missing or invalid
//   - Verifies payments with the facilitator
//   - Settles payments (unless VerifyOnly=true)
//   - Stores payment information in request context via httpx402.PaymentContextKey
//   - Calls next handler on payment success
//
// Example usage:
//
//	config := &httpx402.Config{
//	    FacilitatorURL: "https://api.x402.coinbase.com",
//	    PaymentRequirements: []x402.PaymentRequirement{{
//	        Scheme:            "exact",
//	        Network:           "base-sepolia",
//	        MaxAmountRequired: "10000",
//	        Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
//	        PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
//	        MaxTimeoutSeconds: 300,
//	    }},
//	}
//	r := chi.NewRouter()
//	r.Use(NewChiX402Middleware(config))
//	r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
//	    payment := r.Context().Value(httpx402.PaymentContextKey).(*httpx402.VerifyResponse)
//	    w.Write([]byte("Access granted! Payer: " + payment.Payer))
//	})
func NewChiX402Middleware(config *httpx402.Config) func(http.Handler) http.Handler {
	// Create facilitator client with hardcoded timeouts per FR-017
	facilitator := &httpx402.FacilitatorClient{
		BaseURL:       config.FacilitatorURL,
		Client:        &http.Client{},
		VerifyTimeout: 5 * time.Second,  // Quick verification
		SettleTimeout: 60 * time.Second, // Longer for blockchain tx execution
	}

	// Create fallback facilitator client if configured
	var fallbackFacilitator *httpx402.FacilitatorClient
	if config.FallbackFacilitatorURL != "" {
		fallbackFacilitator = &httpx402.FacilitatorClient{
			BaseURL:       config.FallbackFacilitatorURL,
			Client:        &http.Client{},
			VerifyTimeout: 5 * time.Second,
			SettleTimeout: 60 * time.Second,
		}
	}

	// Enrich payment requirements with facilitator-specific data (like feePayer for SVM)
	enrichedRequirements, err := facilitator.EnrichRequirements(config.PaymentRequirements)
	if err != nil {
		// Log warning but continue with original requirements (graceful degradation per FR-019)
		slog.Default().Warn("failed to enrich payment requirements from facilitator", "error", err)
		enrichedRequirements = config.PaymentRequirements
	} else {
		slog.Default().Info("payment requirements enriched from facilitator", "count", len(enrichedRequirements))
	}

	// Return Chi middleware function with stdlib signature
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := slog.Default()

			// OPTIONS request bypass for CORS preflight support (FR-022)
			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// Build absolute URL for the resource (FR-014)
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			resourceURL := scheme + "://" + r.Host + r.RequestURI

			// Populate resource field in requirements with the actual request URL
			requirementsWithResource := make([]x402.PaymentRequirement, len(enrichedRequirements))
			for i, req := range enrichedRequirements {
				requirementsWithResource[i] = req
				requirementsWithResource[i].Resource = resourceURL
				if requirementsWithResource[i].Description == "" {
					requirementsWithResource[i].Description = "Payment required for " + r.URL.Path
				}
			}

			// Check for X-PAYMENT header (FR-005)
			paymentHeader := r.Header.Get("X-PAYMENT")
			if paymentHeader == "" {
				// No payment provided - return 402 with requirements (FR-007)
				logger.Warn("no payment header provided", "path", r.URL.Path)
				helpers.SendPaymentRequired(w, requirementsWithResource)
				return
			}

			// Parse payment header using shared helper (FR-007)
			payment, err := helpers.ParsePaymentHeaderFromRequest(r)
			if err != nil {
				logger.Warn("invalid payment header", "error", err)
				// Return 400 with x402Version error response (FR-020)
				sendErrorResponse(w, http.StatusBadRequest, "Invalid payment header")
				return
			}

			// Find matching requirement using shared helper
			requirement, err := helpers.FindMatchingRequirement(payment, requirementsWithResource)
			if err != nil {
				logger.Warn("no matching requirement", "error", err)
				helpers.SendPaymentRequired(w, requirementsWithResource)
				return
			}

			// Verify payment with facilitator (primary + fallback support)
			logger.Info("verifying payment", "scheme", payment.Scheme, "network", payment.Network)
			verifyResp, err := facilitator.Verify(payment, requirement)
			if err != nil && fallbackFacilitator != nil {
				logger.Warn("primary facilitator failed, trying fallback", "error", err)
				verifyResp, err = fallbackFacilitator.Verify(payment, requirement)
			}
			if err != nil {
				logger.Error("facilitator verification failed", "error", err)
				sendErrorResponse(w, http.StatusServiceUnavailable, "Payment verification failed")
				return
			}

			if !verifyResp.IsValid {
				logger.Warn("payment verification failed", "reason", verifyResp.InvalidReason)
				helpers.SendPaymentRequired(w, requirementsWithResource)
				return
			}

			// Payment verified successfully (FR-023 - Info level)
			logger.Info("payment verified", "payer", verifyResp.Payer)

			// Settle payment if not verify-only mode (FR-009)
			var settlementResp *x402.SettlementResponse
			if !config.VerifyOnly {
				logger.Info("settling payment", "payer", verifyResp.Payer)
				settlementResp, err = facilitator.Settle(payment, requirement)
				if err != nil && fallbackFacilitator != nil {
					logger.Warn("primary facilitator settlement failed, trying fallback", "error", err)
					settlementResp, err = fallbackFacilitator.Settle(payment, requirement)
				}
				if err != nil {
					logger.Error("settlement failed", "error", err)
					sendErrorResponse(w, http.StatusServiceUnavailable, "Payment settlement failed")
					return
				}

				if !settlementResp.Success {
					logger.Warn("settlement unsuccessful", "reason", settlementResp.ErrorReason)
					helpers.SendPaymentRequired(w, requirementsWithResource)
					return
				}

				logger.Info("payment settled", "transaction", settlementResp.Transaction)

				// Add X-PAYMENT-RESPONSE header with settlement info (FR-011)
				if err := helpers.AddPaymentResponseHeader(w, settlementResp); err != nil {
					logger.Warn("failed to add payment response header", "error", err)
					// Continue anyway - payment was successful
				}
			}

			// Store payment info in request context for handler access (FR-010)
			ctx := context.WithValue(r.Context(), httpx402.PaymentContextKey, verifyResp)
			r = r.WithContext(ctx)

			// Payment successful - call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// sendErrorResponse sends an error response with x402Version field (FR-020)
func sendErrorResponse(w http.ResponseWriter, statusCode int, errorMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	// Simple JSON response with x402Version field
	// Ignore encoding errors - status already sent
	_, _ = w.Write([]byte(`{"x402Version":1,"error":"` + errorMessage + `"}`))
}
