package server

import (
	"context"
	"time"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/mcp"
)

// ServerWrapper provides base functionality for wrapping MCP servers with x402 capabilities.
type ServerWrapper interface {
	// AddPayableTool adds a tool that requires payment
	AddPayableTool(tool ServerTool, requirements ...x402.PaymentRequirement) error

	// AddTool adds a free tool (no payment required)
	AddTool(tool ServerTool) error

	// SetVerifyOnly sets whether to only verify payments without settling
	SetVerifyOnly(verifyOnly bool)
}

// ServerTool combines an MCP tool with its handler function.
// This is used to register tools with the server.
type ServerTool struct {
	// Tool is the MCP tool definition
	Tool interface{}

	// Handler is the function that implements the tool's logic
	Handler interface{}
}

// withTimeoutContext creates a context with timeout for payment operations.
//
//nolint:unused // Reserved for future server implementation
func withTimeoutContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return ctx, func() {}
}

// extractPaymentFromMeta extracts x402 payment from MCP request metadata.
//
//nolint:unused // Reserved for future server implementation
func extractPaymentFromMeta(meta map[string]interface{}) (*x402.PaymentPayload, error) {
	if meta == nil {
		return nil, mcp.ErrPaymentRequired
	}

	paymentData, ok := meta[mcp.MetaKeyPayment]
	if !ok {
		return nil, mcp.ErrPaymentRequired
	}

	// Payment data should be a map that we can marshal/unmarshal
	// MCP spec requires payment in params._meta["x402/payment"]
	// TODO: Implement proper JSON marshaling of payment data
	_ = paymentData

	return nil, mcp.ErrPaymentRequired
}
