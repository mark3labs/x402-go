package mcp

import (
	"time"

	"github.com/mark3labs/x402-go"
)

// MCP-specific type aliases reusing x402 types

// PaymentRequirement is an alias for x402.PaymentRequirement
type PaymentRequirement = x402.PaymentRequirement

// PaymentPayload is an alias for x402.PaymentPayload
type PaymentPayload = x402.PaymentPayload

// SettlementResponse is an alias for x402.SettlementResponse
type SettlementResponse = x402.SettlementResponse

// Signer is an alias for x402.Signer
type Signer = x402.Signer

// PaymentSelector is an alias for x402.PaymentSelector
type PaymentSelector = x402.PaymentSelector

// Constants for payment timeouts as defined in spec

const (
	// PaymentVerifyTimeout is the maximum time to wait for payment verification (FR-017)
	PaymentVerifyTimeout = 5 * time.Second

	// PaymentSettleTimeout is the maximum time to wait for payment settlement (FR-018)
	PaymentSettleTimeout = 60 * time.Second
)

// PaymentRequirements represents the data structure returned in a 402 error
type PaymentRequirements struct {
	X402Version int                  `json:"x402Version"`
	Error       string               `json:"error"`
	Accepts     []PaymentRequirement `json:"accepts"`
}
