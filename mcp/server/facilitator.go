package server

import (
	"context"
	nethttp "net/http"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/http"
	"github.com/mark3labs/x402-go/mcp"
)

// FacilitatorWrapper wraps the x402-go facilitator client with MCP-specific timeouts
type FacilitatorWrapper struct {
	client *http.FacilitatorClient
}

// NewFacilitatorWrapper creates a new facilitator wrapper
func NewFacilitatorWrapper(facilitatorURL string) (*FacilitatorWrapper, error) {
	client := &http.FacilitatorClient{
		BaseURL:       facilitatorURL,
		Client:        &nethttp.Client{},
		VerifyTimeout: mcp.VerificationTimeout,
		SettleTimeout: mcp.SettlementTimeout,
	}
	return &FacilitatorWrapper{
		client: client,
	}, nil
}

// Verify verifies a payment with the configured timeout (FR-017)
func (f *FacilitatorWrapper) Verify(ctx context.Context, payment x402.PaymentPayload, requirement x402.PaymentRequirement) (*http.VerifyResponse, error) {
	// Create context with 5-second timeout
	verifyCtx, cancel := context.WithTimeout(ctx, mcp.VerificationTimeout)
	defer cancel()

	// TODO: Call facilitator verify endpoint
	_ = verifyCtx
	return nil, nil
}

// Settle settles a payment with the configured timeout (FR-018)
func (f *FacilitatorWrapper) Settle(ctx context.Context, payment x402.PaymentPayload, requirement x402.PaymentRequirement) (*x402.SettlementResponse, error) {
	// Create context with 60-second timeout
	settleCtx, cancel := context.WithTimeout(ctx, mcp.SettlementTimeout)
	defer cancel()

	// TODO: Call facilitator settle endpoint
	_ = settleCtx
	return nil, nil
}

// withVerifyTimeout wraps context with verification timeout
//
//nolint:unused // Reserved for future facilitator implementation
func withVerifyTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, mcp.VerificationTimeout)
}

// withSettleTimeout wraps context with settlement timeout
//
//nolint:unused // Reserved for future facilitator implementation
func withSettleTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, mcp.SettlementTimeout)
}

// Custom errors for timeout scenarios
var (
	ErrVerifyTimeout = mcp.ErrVerificationTimeout
	ErrSettleTimeout = mcp.ErrSettlementTimeout
)

// timeoutError wraps an error indicating which operation timed out
//
//nolint:unused // Reserved for future error handling
func timeoutError(err error, operation string) error {
	if err == context.DeadlineExceeded {
		switch operation {
		case "verify":
			return ErrVerifyTimeout
		case "settle":
			return ErrSettleTimeout
		}
	}
	return err
}

// verifyWithTimeout calls verify with automatic timeout handling
//
//nolint:unused // Reserved for future facilitator implementation
func (f *FacilitatorWrapper) verifyWithTimeout(payment x402.PaymentPayload, requirement x402.PaymentRequirement) (*http.VerifyResponse, error) {
	verifyCtx, cancel := context.WithTimeout(context.Background(), mcp.VerificationTimeout)
	defer cancel()

	// TODO: Implement actual verify call
	_ = verifyCtx
	_ = payment
	_ = requirement
	return nil, nil
}

// settleWithTimeout calls settle with automatic timeout handling
//
//nolint:unused // Reserved for future facilitator implementation
func (f *FacilitatorWrapper) settleWithTimeout(payment x402.PaymentPayload, requirement x402.PaymentRequirement) (*x402.SettlementResponse, error) {
	settleCtx, cancel := context.WithTimeout(context.Background(), mcp.SettlementTimeout)
	defer cancel()

	// TODO: Implement actual settle call
	_ = settleCtx
	_ = payment
	_ = requirement
	return nil, nil
}
