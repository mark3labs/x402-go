package mcp

import (
	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/http"
)

// MCP-specific constants for payment metadata keys
const (
	// MetaKeyPayment is the key for payment data in MCP request params._meta
	MetaKeyPayment = "x402/payment"

	// MetaKeyPaymentResponse is the key for settlement response in MCP result._meta
	MetaKeyPaymentResponse = "x402/payment-response"

	// MetaKeyProtocolVersion is the key for protocol version negotiation
	MetaKeyProtocolVersion = "protocol-version"
)

// Timeout constants for MCP payment operations.
// These reference the protocol-level defaults from the x402 package:
//   - VerificationTimeout (5s): Quick check if payment exists on-chain
//   - SettlementTimeout (60s): Wait for full on-chain confirmation
const (
	// VerificationTimeout is the maximum time to wait for payment verification
	VerificationTimeout = x402.DefaultVerifyTimeout

	// SettlementTimeout is the maximum time to wait for payment settlement
	SettlementTimeout = x402.DefaultSettleTimeout
)

// PaymentContext holds payment information during MCP request lifecycle
type PaymentContext struct {
	Payment            *x402.PaymentPayload     `json:"payment,omitempty"`
	Requirement        *x402.PaymentRequirement `json:"requirement,omitempty"`
	VerificationResult *http.VerifyResponse     `json:"verification_result,omitempty"`
	SettlementResult   *x402.SettlementResponse `json:"settlement_result,omitempty"`
}

// PaymentEvent represents different payment lifecycle events
type PaymentEvent struct {
	Type        PaymentEventType
	Amount      string
	Asset       string
	Network     string
	Transaction string
	Error       error
}

// PaymentEventType represents the type of payment event
type PaymentEventType string

const (
	PaymentEventAttempt PaymentEventType = "payment_attempt"
	PaymentEventSuccess PaymentEventType = "payment_success"
	PaymentEventFailure PaymentEventType = "payment_failure"
)
