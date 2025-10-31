package client

import (
	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/mcp"
)

// PaymentHandler manages payment creation and signer selection for MCP clients
type PaymentHandler struct {
	signers  []x402.Signer
	selector x402.PaymentSelector
}

// NewPaymentHandler creates a new payment handler with the given signers
func NewPaymentHandler(signers []x402.Signer, selector x402.PaymentSelector) *PaymentHandler {
	if selector == nil {
		selector = x402.NewDefaultPaymentSelector()
	}
	return &PaymentHandler{
		signers:  signers,
		selector: selector,
	}
}

// CreatePayment attempts to create a payment matching one of the given requirements
// It uses the configured selector to choose the best signer and create a payment
func (h *PaymentHandler) CreatePayment(requirements []x402.PaymentRequirement) (*x402.PaymentPayload, error) {
	if len(requirements) == 0 {
		return nil, mcp.ErrPaymentRequired
	}

	if len(h.signers) == 0 {
		return nil, mcp.ErrNoMatchingSigner
	}

	// Use selector to find best matching requirement and create payment
	payment, err := h.selector.SelectAndSign(requirements, h.signers)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

// CanFulfillAnyRequirement checks if any signer can fulfill any of the requirements
func (h *PaymentHandler) CanFulfillAnyRequirement(requirements []x402.PaymentRequirement) bool {
	for _, req := range requirements {
		for _, signer := range h.signers {
			if signer.CanSign(&req) {
				return true
			}
		}
	}
	return false
}

// GetSigners returns the configured signers
func (h *PaymentHandler) GetSigners() []x402.Signer {
	return h.signers
}
