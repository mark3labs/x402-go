// Package http provides HTTP middleware for x402 payment gating.
package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/mark3labs/x402-go"
)

// Config holds the configuration for the x402 middleware.
type Config struct {
	// FacilitatorURL is the primary facilitator endpoint
	FacilitatorURL string

	// FallbackFacilitatorURL is the optional backup facilitator
	FallbackFacilitatorURL string

	// PaymentRequirements defines the accepted payment methods
	PaymentRequirements []x402.PaymentRequirement

	// VerifyOnly skips settlement if true (only verifies payments)
	VerifyOnly bool
}

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// PaymentContextKey is the context key for storing verified payment information.
const PaymentContextKey = contextKey("x402_payment")

// NewX402Middleware creates a new x402 payment middleware.
// It returns a middleware function that wraps HTTP handlers with payment gating.
func NewX402Middleware(config *Config) func(http.Handler) http.Handler {
	// Create facilitator client
	facilitator := &FacilitatorClient{
		BaseURL: config.FacilitatorURL,
		Client: &http.Client{
			Timeout: 2 * time.Second, // 2s timeout per SC-002
		},
	}

	// Create fallback facilitator client if configured
	var fallbackFacilitator *FacilitatorClient
	if config.FallbackFacilitatorURL != "" {
		fallbackFacilitator = &FacilitatorClient{
			BaseURL: config.FallbackFacilitatorURL,
			Client: &http.Client{
				Timeout: 2 * time.Second,
			},
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := slog.Default()

			// Check for X-PAYMENT header
			paymentHeader := r.Header.Get("X-PAYMENT")
			if paymentHeader == "" {
				// No payment provided - return 402 with requirements
				logger.Info("no payment header provided", "path", r.URL.Path)
				sendPaymentRequired(w, config)
				return
			}

			// Parse payment header
			payment, err := parsePaymentHeader(r)
			if err != nil {
				logger.Warn("invalid payment header", "error", err)
				http.Error(w, "Invalid payment header", http.StatusBadRequest)
				return
			}

			// Find matching requirement
			requirement, err := findMatchingRequirement(payment, config.PaymentRequirements)
			if err != nil {
				logger.Warn("no matching requirement", "error", err)
				sendPaymentRequired(w, config)
				return
			}

			// Verify payment with facilitator
			logger.Info("verifying payment", "scheme", payment.Scheme, "network", payment.Network)
			verifyResp, err := facilitator.Verify(payment, requirement)
			if err != nil && fallbackFacilitator != nil {
				logger.Warn("primary facilitator failed, trying fallback", "error", err)
				verifyResp, err = fallbackFacilitator.Verify(payment, requirement)
			}
			if err != nil {
				logger.Error("facilitator verification failed", "error", err)
				http.Error(w, "Payment verification failed", http.StatusServiceUnavailable)
				return
			}

			if !verifyResp.IsValid {
				logger.Warn("payment verification failed", "reason", verifyResp.InvalidReason)
				sendPaymentRequired(w, config)
				return
			}

			// Payment verified successfully
			logger.Info("payment verified", "payer", verifyResp.Payer)

			// Settle payment if not verify-only mode
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
					http.Error(w, "Payment settlement failed", http.StatusServiceUnavailable)
					return
				}

				if !settlementResp.Success {
					logger.Warn("settlement unsuccessful", "reason", settlementResp.ErrorReason)
					sendPaymentRequired(w, config)
					return
				}

				logger.Info("payment settled", "transaction", settlementResp.Transaction)

				// Add X-PAYMENT-RESPONSE header with settlement info
				if err := addPaymentResponseHeader(w, settlementResp); err != nil {
					logger.Warn("failed to add payment response header", "error", err)
					// Continue anyway - payment was successful
				}
			}

			// Store payment info in context for handler access
			ctx := context.WithValue(r.Context(), PaymentContextKey, verifyResp)
			r = r.WithContext(ctx)

			// Payment successful - call next handler
			next.ServeHTTP(w, r)
		})
	}
}
