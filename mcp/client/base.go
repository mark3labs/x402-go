package client

import (
	"context"

	"github.com/mark3labs/mcp-go/client/transport"
)

// TransportWrapper provides a base interface for wrapping MCP transport implementations.
// This allows adding x402 payment capabilities to any transport type.
type TransportWrapper interface {
	transport.Interface

	// GetUnderlyingTransport returns the wrapped transport instance
	GetUnderlyingTransport() transport.Interface
}

// BidirectionalTransportWrapper extends TransportWrapper for bidirectional communication.
type BidirectionalTransportWrapper interface {
	transport.BidirectionalInterface
	TransportWrapper
}

// withTimeoutContext wraps a context with the specified timeout.
// Returns the new context and a cancel function that should be called when done.
func withTimeoutContext(ctx context.Context, timeout interface{}) (context.Context, context.CancelFunc) {
	// timeout can be a time.Duration or similar
	// For now, return the context as-is; this will be used in transport implementation
	return ctx, func() {}
}
