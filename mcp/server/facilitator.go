package server

import (
	"context"
	"fmt"
	nethttp "net/http"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/facilitator"
	"github.com/mark3labs/x402-go/http"
	"github.com/mark3labs/x402-go/mcp"
)

// Facilitator interface for payment verification and settlement
type Facilitator interface {
	// Verify verifies a payment without settling it
	Verify(ctx context.Context, payment *x402.PaymentPayload, requirement x402.PaymentRequirement) (*facilitator.VerifyResponse, error)

	// Settle settles a payment on the blockchain
	Settle(ctx context.Context, payment *x402.PaymentPayload, requirement x402.PaymentRequirement) (*x402.SettlementResponse, error)
}

// HTTPFacilitator implements Facilitator using the http.FacilitatorClient
type HTTPFacilitator struct {
	client *http.FacilitatorClient
}

// NewHTTPFacilitator creates a new HTTP facilitator client
func NewHTTPFacilitator(facilitatorURL string) *HTTPFacilitator {
	client := &http.FacilitatorClient{
		BaseURL:       facilitatorURL,
		Client:        &nethttp.Client{Timeout: mcp.PaymentSettleTimeout},
		VerifyTimeout: mcp.PaymentVerifyTimeout,
		SettleTimeout: mcp.PaymentSettleTimeout,
		MaxRetries:    2,
	}
	return &HTTPFacilitator{
		client: client,
	}
}

// Verify verifies a payment with the facilitator
func (f *HTTPFacilitator) Verify(ctx context.Context, payment *x402.PaymentPayload, requirement x402.PaymentRequirement) (*facilitator.VerifyResponse, error) {
	resp, err := f.client.Verify(ctx, *payment, requirement)
	if err != nil {
		return nil, fmt.Errorf("facilitator verify failed: %w", err)
	}

	return resp, nil
}

// Settle settles a payment through the facilitator
func (f *HTTPFacilitator) Settle(ctx context.Context, payment *x402.PaymentPayload, requirement x402.PaymentRequirement) (*x402.SettlementResponse, error) {
	resp, err := f.client.Settle(ctx, *payment, requirement)
	if err != nil {
		return nil, fmt.Errorf("facilitator settle failed: %w", err)
	}

	return resp, nil
}
