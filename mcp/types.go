package mcp

import (
	"time"

	"github.com/mark3labs/x402-go"
)

// Constants for payment timeouts as defined in spec

const (
	// PaymentVerifyTimeout is the maximum time to wait for payment verification (FR-017)
	PaymentVerifyTimeout = 5 * time.Second

	// PaymentSettleTimeout is the maximum time to wait for payment settlement (FR-018)
	PaymentSettleTimeout = 60 * time.Second
)

// PaymentRequirements represents the data structure returned in a 402 error response in MCP.
// This is MCP-specific and wraps the standard x402 payment requirements.
type PaymentRequirements struct {
	X402Version int                       `json:"x402Version"`
	Error       string                    `json:"error"`
	Accepts     []x402.PaymentRequirement `json:"accepts"`
}
