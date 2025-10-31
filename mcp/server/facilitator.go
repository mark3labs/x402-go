package server

import (
	"context"
	"fmt"
	nethttp "net/http"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/http"
	"github.com/mark3labs/x402-go/mcp"
)

// Facilitator interface for payment verification and settlement
type Facilitator interface {
	// Verify verifies a payment without settling it
	Verify(ctx context.Context, payment *x402.PaymentPayload, requirement x402.PaymentRequirement) (*mcp.VerifyResponse, error)

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
		Client:        &nethttp.Client{Timeout: mcp.PaymentVerifyTimeout},
		VerifyTimeout: mcp.PaymentVerifyTimeout,
		SettleTimeout: mcp.PaymentSettleTimeout,
		MaxRetries:    2,
	}
	return &HTTPFacilitator{
		client: client,
	}
}

// Verify verifies a payment with the facilitator
func (f *HTTPFacilitator) Verify(ctx context.Context, payment *x402.PaymentPayload, requirement x402.PaymentRequirement) (*mcp.VerifyResponse, error) {
	// Note: The facilitator client doesn't accept context - it uses its own timeouts
	resp, err := f.client.Verify(*payment, requirement)
	if err != nil {
		return nil, fmt.Errorf("facilitator verify failed: %w", err)
	}

	// Convert facilitator response to MCP verify response
	verifyResp := &mcp.VerifyResponse{
		IsValid:       resp.IsValid,
		InvalidReason: resp.InvalidReason,
		Payer:         resp.Payer,
	}

	return verifyResp, nil
}

// Settle settles a payment through the facilitator
func (f *HTTPFacilitator) Settle(ctx context.Context, payment *x402.PaymentPayload, requirement x402.PaymentRequirement) (*x402.SettlementResponse, error) {
	// Note: The facilitator client doesn't accept context - it uses its own timeouts
	resp, err := f.client.Settle(*payment, requirement)
	if err != nil {
		return nil, fmt.Errorf("facilitator settle failed: %w", err)
	}

	// SettlementResponse from facilitator is already in the correct format
	return resp, nil
}
