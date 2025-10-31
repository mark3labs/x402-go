package server

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/http"
	"github.com/mark3labs/x402-go/mcp"
)

// PaymentMiddleware wraps tool handlers to enforce x402 payments
type PaymentMiddleware struct {
	facilitator  *http.FacilitatorClient
	requirements map[string][]x402.PaymentRequirement
	verifyOnly   bool
}

// NewPaymentMiddleware creates a new payment middleware
func NewPaymentMiddleware(facilitator *http.FacilitatorClient, verifyOnly bool) *PaymentMiddleware {
	return &PaymentMiddleware{
		facilitator:  facilitator,
		requirements: make(map[string][]x402.PaymentRequirement),
		verifyOnly:   verifyOnly,
	}
}

// extractPayment extracts x402 payment from params._meta["x402/payment"]
func (m *PaymentMiddleware) extractPayment(params map[string]interface{}) (*x402.PaymentPayload, error) {
	// TODO: Extract _meta from params
	meta, ok := params["_meta"].(map[string]interface{})
	if !ok {
		return nil, mcp.ErrPaymentRequired
	}

	// TODO: Extract payment from _meta["x402/payment"]
	paymentData, ok := meta[mcp.MetaKeyPayment]
	if !ok {
		return nil, mcp.ErrPaymentRequired
	}

	// TODO: Unmarshal payment data to PaymentPayload
	paymentBytes, err := json.Marshal(paymentData)
	if err != nil {
		return nil, err
	}

	var payment x402.PaymentPayload
	if err := json.Unmarshal(paymentBytes, &payment); err != nil {
		return nil, err
	}

	return &payment, nil
}

// verifyPayment verifies payment with facilitator
func (m *PaymentMiddleware) verifyPayment(ctx context.Context, payment *x402.PaymentPayload, requirement *x402.PaymentRequirement) (*http.VerifyResponse, error) {
	// TODO: Create context with 5-second timeout (FR-017)
	// TODO: Call facilitator.Verify()
	return nil, nil
}

// settlePayment settles payment with facilitator
func (m *PaymentMiddleware) settlePayment(ctx context.Context, payment *x402.PaymentPayload, requirement *x402.PaymentRequirement) (*x402.SettlementResponse, error) {
	if m.verifyOnly {
		return &x402.SettlementResponse{
			Success: true,
			Network: payment.Network,
		}, nil
	}

	// TODO: Create context with 60-second timeout (FR-018)
	// TODO: Call facilitator.Settle()
	return nil, nil
}

// injectSettlement adds settlement response to result._meta["x402/payment-response"]
func (m *PaymentMiddleware) injectSettlement(result map[string]interface{}, settlement *x402.SettlementResponse) error {
	// TODO: Ensure _meta exists in result
	// TODO: Add settlement to _meta["x402/payment-response"]
	return nil
}
