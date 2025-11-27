package server

import (
	"context"
	"fmt"
	nethttp "net/http"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/facilitator"
	"github.com/mark3labs/x402-go/http"
)

// Facilitator interface for payment verification and settlement
type Facilitator interface {
	// Verify verifies a payment without settling it
	Verify(ctx context.Context, payment *x402.PaymentPayload, requirement x402.PaymentRequirement) (*facilitator.VerifyResponse, error)

	// Settle settles a payment on the blockchain
	Settle(ctx context.Context, payment *x402.PaymentPayload, requirement x402.PaymentRequirement) (*x402.SettlementResponse, error)
}

// HTTPFacilitator implements Facilitator using the http.FacilitatorClient
type HTTPFacilitator struct {
	client *http.FacilitatorClient
}

// HTTPFacilitatorOption configures an HTTPFacilitator
type HTTPFacilitatorOption func(*http.FacilitatorClient)

// WithAuthorization sets a static Authorization header value for the facilitator.
// Example: "Bearer your-api-key" or "Basic base64-encoded-credentials"
func WithAuthorization(authorization string) HTTPFacilitatorOption {
	return func(c *http.FacilitatorClient) {
		c.Authorization = authorization
	}
}

// WithAuthorizationProvider sets a dynamic Authorization header provider for the facilitator.
// This is useful for tokens that may need to be refreshed.
// If set, this takes precedence over the static Authorization value.
func WithAuthorizationProvider(provider http.AuthorizationProvider) HTTPFacilitatorOption {
	return func(c *http.FacilitatorClient) {
		c.AuthorizationProvider = provider
	}
}

// NewHTTPFacilitator creates a new HTTP facilitator client
func NewHTTPFacilitator(facilitatorURL string, opts ...HTTPFacilitatorOption) *HTTPFacilitator {
	timeouts := x402.DefaultTimeouts
	client := &http.FacilitatorClient{
		BaseURL:    facilitatorURL,
		Client:     &nethttp.Client{Timeout: timeouts.RequestTimeout},
		Timeouts:   timeouts,
		MaxRetries: 2,
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	return &HTTPFacilitator{
		client: client,
	}
}

// Verify verifies a payment with the facilitator
func (f *HTTPFacilitator) Verify(ctx context.Context, payment *x402.PaymentPayload, requirement x402.PaymentRequirement) (*facilitator.VerifyResponse, error) {
	resp, err := f.client.Verify(ctx, *payment, requirement)
	if err != nil {
		return nil, fmt.Errorf("facilitator verify failed: %w", err)
	}

	return resp, nil
}

// Settle settles a payment through the facilitator
func (f *HTTPFacilitator) Settle(ctx context.Context, payment *x402.PaymentPayload, requirement x402.PaymentRequirement) (*x402.SettlementResponse, error) {
	resp, err := f.client.Settle(ctx, *payment, requirement)
	if err != nil {
		return nil, fmt.Errorf("facilitator settle failed: %w", err)
	}

	return resp, nil
}
