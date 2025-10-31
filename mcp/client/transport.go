package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/x402-go"
	mcpx402 "github.com/mark3labs/x402-go/mcp"
)

// X402Transport wraps an MCP transport and adds x402 payment capabilities
type X402Transport struct {
	underlying     transport.Interface
	paymentHandler *PaymentHandler
	eventCallback  func(mcpx402.PaymentEvent)
	sessionID      string
}

// TransportOption configures an X402Transport
type TransportOption func(*X402Transport)

// WithSigner adds a signer to the transport
func WithSigner(signer x402.Signer) TransportOption {
	return func(t *X402Transport) {
		if t.paymentHandler == nil {
			t.paymentHandler = NewPaymentHandler([]x402.Signer{signer}, nil)
		} else {
			t.paymentHandler.signers = append(t.paymentHandler.signers, signer)
		}
	}
}

// WithPaymentCallback sets a callback for payment events
func WithPaymentCallback(callback func(mcpx402.PaymentEvent)) TransportOption {
	return func(t *X402Transport) {
		t.eventCallback = callback
	}
}

// NewX402Transport creates a new x402-enabled MCP transport
func NewX402Transport(underlying transport.Interface, signers []x402.Signer, opts ...TransportOption) (*X402Transport, error) {
	if underlying == nil {
		return nil, fmt.Errorf("underlying transport cannot be nil")
	}

	if len(signers) == 0 {
		return nil, fmt.Errorf("at least one signer is required")
	}

	// Validate all signers are non-nil
	for i, signer := range signers {
		if signer == nil {
			return nil, fmt.Errorf("signer at index %d is nil", i)
		}
	}

	t := &X402Transport{
		underlying:     underlying,
		paymentHandler: NewPaymentHandler(signers, nil),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t, nil
}

// Start implements transport.Interface
func (t *X402Transport) Start(ctx context.Context) error {
	return t.underlying.Start(ctx)
}

// SendRequest implements transport.Interface with x402 payment support
func (t *X402Transport) SendRequest(ctx context.Context, request transport.JSONRPCRequest) (*transport.JSONRPCResponse, error) {
	// Send the initial request
	resp, err := t.underlying.SendRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// Check if this is a 402 error requiring payment
	if resp.Error != nil && resp.Error.Code == 402 {
		// Extract payment requirements from error
		requirements, err := t.extractPaymentRequirements(resp.Error)
		if err != nil {
			return nil, fmt.Errorf("extract payment requirements: %w", err)
		}

		// Emit payment attempt event
		if t.eventCallback != nil {
			t.eventCallback(mcpx402.PaymentEvent{
				Type: mcpx402.PaymentEventAttempt,
			})
		}

		// Create payment
		payment, err := t.paymentHandler.CreatePayment(requirements)
		if err != nil {
			if t.eventCallback != nil {
				t.eventCallback(mcpx402.PaymentEvent{
					Type:  mcpx402.PaymentEventFailure,
					Error: err,
				})
			}
			return nil, fmt.Errorf("create payment: %w", err)
		}

		// Inject payment into request params
		requestWithPayment, err := t.injectPayment(request, payment)
		if err != nil {
			return nil, fmt.Errorf("inject payment: %w", err)
		}

		// Retry with payment
		resp, err = t.underlying.SendRequest(ctx, requestWithPayment)
		if err != nil {
			if t.eventCallback != nil {
				t.eventCallback(mcpx402.PaymentEvent{
					Type:  mcpx402.PaymentEventFailure,
					Error: err,
				})
			}
			return nil, err
		}

		// Emit success event
		if t.eventCallback != nil {
			t.eventCallback(mcpx402.PaymentEvent{
				Type:    mcpx402.PaymentEventSuccess,
				Network: payment.Network,
			})
		}
	}

	return resp, nil
}

// SendNotification implements transport.Interface
func (t *X402Transport) SendNotification(ctx context.Context, notification mcp.JSONRPCNotification) error {
	return t.underlying.SendNotification(ctx, notification)
}

// SetNotificationHandler implements transport.Interface
func (t *X402Transport) SetNotificationHandler(handler func(notification mcp.JSONRPCNotification)) {
	t.underlying.SetNotificationHandler(handler)
}

// Close implements transport.Interface
func (t *X402Transport) Close() error {
	return t.underlying.Close()
}

// GetSessionId implements transport.Interface
func (t *X402Transport) GetSessionId() string {
	if t.sessionID != "" {
		return t.sessionID
	}
	return t.underlying.GetSessionId()
}

// GetUnderlyingTransport implements TransportWrapper
func (t *X402Transport) GetUnderlyingTransport() transport.Interface {
	return t.underlying
}

// extractPaymentRequirements extracts payment requirements from a 402 error response
func (t *X402Transport) extractPaymentRequirements(jsonrpcError *mcp.JSONRPCErrorDetails) ([]x402.PaymentRequirement, error) {
	if jsonrpcError == nil || jsonrpcError.Data == nil {
		return nil, fmt.Errorf("no payment requirements in error")
	}

	// Parse the error data as PaymentRequirementsResponse
	dataBytes, err := json.Marshal(jsonrpcError.Data)
	if err != nil {
		return nil, fmt.Errorf("marshal error data: %w", err)
	}

	var paymentReqs x402.PaymentRequirementsResponse
	if err := json.Unmarshal(dataBytes, &paymentReqs); err != nil {
		return nil, fmt.Errorf("unmarshal payment requirements: %w", err)
	}

	if len(paymentReqs.Accepts) == 0 {
		return nil, fmt.Errorf("no payment options in requirements")
	}

	return paymentReqs.Accepts, nil
}

// injectPayment adds x402 payment to the request params._meta field
func (t *X402Transport) injectPayment(request transport.JSONRPCRequest, payment *x402.PaymentPayload) (transport.JSONRPCRequest, error) {
	// Parse params as map
	var params map[string]interface{}
	if request.Params != nil {
		paramsBytes, err := json.Marshal(request.Params)
		if err != nil {
			return request, fmt.Errorf("marshal params: %w", err)
		}
		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return request, fmt.Errorf("unmarshal params: %w", err)
		}
	} else {
		params = make(map[string]interface{})
	}

	// Ensure _meta exists
	meta, ok := params["_meta"].(map[string]interface{})
	if !ok {
		meta = make(map[string]interface{})
		params["_meta"] = meta
	}

	// Add payment to _meta
	meta[mcpx402.MetaKeyPayment] = payment

	// Create new request with updated params
	newRequest := request
	newRequest.Params = params

	return newRequest, nil
}
