package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/client/transport"
	mcpproto "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/mcp"
)

// Transport wraps an MCP transport and adds x402 payment handling
type Transport struct {
	baseTransport transport.Interface
	config        *Config
}

// NewTransport creates a new x402-enabled MCP transport
func NewTransport(serverURL string, opts ...Option) (*Transport, error) {
	config := DefaultConfig(serverURL)
	for _, opt := range opts {
		opt(config)
	}

	// Create base HTTP transport
	baseTransport, err := transport.NewStreamableHTTP(serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create base transport: %w", err)
	}

	// Use default selector if none provided
	if config.Selector == nil {
		config.Selector = &x402.DefaultPaymentSelector{}
	}

	return &Transport{
		baseTransport: baseTransport,
		config:        config,
	}, nil
}

// Start starts the MCP connection
func (t *Transport) Start(ctx context.Context) error {
	return t.baseTransport.Start(ctx)
}

// SendRequest implements transport.Interface by intercepting requests and handling 402 errors
func (t *Transport) SendRequest(ctx context.Context, req transport.JSONRPCRequest) (*transport.JSONRPCResponse, error) {
	// Send initial request
	resp, err := t.baseTransport.SendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// Check if response is a 402 error
	if resp.Error != nil && resp.Error.Code == 402 {
		// Extract payment requirements
		var data json.RawMessage
		if resp.Error.Data != nil {
			dataBytes, err := json.Marshal(resp.Error.Data)
			if err != nil {
				return resp, fmt.Errorf("failed to marshal error data: %w", err)
			}
			data = dataBytes
		}

		requirements, err := t.extractPaymentRequirements(data)
		if err != nil {
			return resp, fmt.Errorf("failed to extract payment requirements: %w", err)
		}

		// Create payment
		payment, err := t.createPayment(ctx, requirements)
		if err != nil {
			return resp, mcp.WrapX402Error(err, req.Method)
		}

		// Inject payment and retry
		modifiedReq, err := t.injectPaymentMeta(req, payment)
		if err != nil {
			return resp, fmt.Errorf("failed to inject payment: %w", err)
		}

		// Retry with payment
		return t.retryWithPayment(ctx, modifiedReq, payment)
	}

	return resp, nil
}

// SendNotification sends a notification to the server
func (t *Transport) SendNotification(ctx context.Context, notif mcpproto.JSONRPCNotification) error {
	return t.baseTransport.SendNotification(ctx, notif)
}

// SetNotificationHandler sets the notification handler
func (t *Transport) SetNotificationHandler(handler func(mcpproto.JSONRPCNotification)) {
	t.baseTransport.SetNotificationHandler(handler)
}

// Close closes the transport
func (t *Transport) Close() error {
	return t.baseTransport.Close()
}

// GetSessionId returns the session ID
func (t *Transport) GetSessionId() string {
	return t.baseTransport.GetSessionId()
}

// extractPaymentRequirements extracts payment requirements from 402 error data
func (t *Transport) extractPaymentRequirements(data json.RawMessage) ([]x402.PaymentRequirement, error) {
	var reqData mcp.PaymentRequirements
	if err := json.Unmarshal(data, &reqData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment requirements: %w", err)
	}

	if reqData.X402Version != 1 {
		return nil, fmt.Errorf("unsupported x402 version: %d", reqData.X402Version)
	}

	if len(reqData.Accepts) == 0 {
		return nil, mcp.ErrNoPaymentRequirements
	}

	return reqData.Accepts, nil
}

// createPayment creates a payment using the configured signers
func (t *Transport) createPayment(ctx context.Context, requirements []x402.PaymentRequirement) (*x402.PaymentPayload, error) {
	if len(t.config.Signers) == 0 {
		return nil, mcp.ErrNoMatchingSigner
	}

	// Trigger payment attempt callback
	if t.config.OnPaymentAttempt != nil {
		if len(requirements) > 0 {
			req := requirements[0]
			t.config.OnPaymentAttempt(PaymentEvent{
				Type:      PaymentAttempt,
				Amount:    req.MaxAmountRequired,
				Asset:     req.Asset,
				Network:   req.Network,
				Recipient: req.PayTo,
			})
		}
	}

	// Use selector to choose signer and create payment
	payment, err := t.config.Selector.SelectAndSign(requirements, t.config.Signers)
	if err != nil {
		if t.config.OnPaymentFailure != nil {
			t.config.OnPaymentFailure(PaymentEvent{
				Type:  PaymentFailure,
				Error: err,
			})
		}
		return nil, err
	}

	return payment, nil
}

// injectPaymentMeta injects payment into request params._meta
func (t *Transport) injectPaymentMeta(req transport.JSONRPCRequest, payment *x402.PaymentPayload) (transport.JSONRPCRequest, error) {
	// Convert params to map
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		// If params is not a map, create one
		params = make(map[string]interface{})
		if req.Params != nil {
			// Try to marshal and unmarshal to convert
			data, err := json.Marshal(req.Params)
			if err != nil {
				return req, fmt.Errorf("failed to marshal params: %w", err)
			}
			if err := json.Unmarshal(data, &params); err != nil {
				return req, fmt.Errorf("failed to unmarshal params: %w", err)
			}
		}
	}

	// Get or create _meta
	meta, ok := params["_meta"].(map[string]interface{})
	if !ok {
		meta = make(map[string]interface{})
	}

	// Add payment to _meta
	meta["x402/payment"] = payment
	params["_meta"] = meta

	// Create modified request
	modifiedReq := req
	modifiedReq.Params = params

	return modifiedReq, nil
}

// retryWithPayment retries the request with payment
func (t *Transport) retryWithPayment(ctx context.Context, req transport.JSONRPCRequest, payment *x402.PaymentPayload) (*transport.JSONRPCResponse, error) {
	resp, err := t.baseTransport.SendRequest(ctx, req)
	if err != nil {
		if t.config.OnPaymentFailure != nil {
			t.config.OnPaymentFailure(PaymentEvent{
				Type:    PaymentFailure,
				Error:   err,
				Network: payment.Network,
			})
		}
		return resp, err
	}

	// Check if payment succeeded
	if resp.Error != nil {
		if t.config.OnPaymentFailure != nil {
			t.config.OnPaymentFailure(PaymentEvent{
				Type:    PaymentFailure,
				Error:   fmt.Errorf("payment rejected: %s", resp.Error.Message),
				Network: payment.Network,
			})
		}
		return resp, nil
	}

	// Payment succeeded
	if t.config.OnPaymentSuccess != nil {
		t.config.OnPaymentSuccess(PaymentEvent{
			Type:    PaymentSuccess,
			Network: payment.Network,
		})
	}

	return resp, nil
}
