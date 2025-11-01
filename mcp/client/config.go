package client

import (
	"net/http"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/mcp"
)

// Config holds configuration for the MCP client with x402 payment support
type Config struct {
	// Signers is the list of payment signers in priority order
	Signers []mcp.Signer

	// ServerURL is the MCP server endpoint
	ServerURL string

	// HTTPClient is the HTTP client for requests (optional, uses default if nil)
	HTTPClient *http.Client

	// OnPaymentAttempt is called when a payment attempt is made
	OnPaymentAttempt func(PaymentEvent)

	// OnPaymentSuccess is called when a payment succeeds
	OnPaymentSuccess func(PaymentEvent)

	// OnPaymentFailure is called when a payment fails
	OnPaymentFailure func(PaymentEvent)

	// Selector is the payment selector for choosing which signer to use (optional, uses default if nil)
	Selector mcp.PaymentSelector

	// Verbose enables detailed logging
	Verbose bool
}

// PaymentEvent represents a payment lifecycle event
type PaymentEvent struct {
	// Type is the event type
	Type PaymentEventType

	// Tool is the tool name that required payment
	Tool string

	// Amount is the payment amount in atomic units
	Amount string

	// Asset is the token address
	Asset string

	// Network is the blockchain network
	Network string

	// Recipient is the payment recipient address
	Recipient string

	// Transaction is the blockchain transaction hash (for success events)
	Transaction string

	// Error is the error details (for failure events)
	Error error

	// Payer is the address that made the payment
	Payer string
}

// PaymentEventType represents the type of payment event
type PaymentEventType string

const (
	// PaymentAttempt indicates a payment is being attempted
	PaymentAttempt PaymentEventType = "attempt"

	// PaymentSuccess indicates a payment succeeded
	PaymentSuccess PaymentEventType = "success"

	// PaymentFailure indicates a payment failed
	PaymentFailure PaymentEventType = "failure"
)

// Option is a functional option for configuring the Transport
type Option func(*Config)

// WithSigner adds a payment signer to the configuration
func WithSigner(signer mcp.Signer) Option {
	return func(c *Config) {
		c.Signers = append(c.Signers, signer)
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		c.HTTPClient = client
	}
}

// WithPaymentCallback sets a unified payment callback for all events
func WithPaymentCallback(callback func(PaymentEvent)) Option {
	return func(c *Config) {
		c.OnPaymentAttempt = callback
		c.OnPaymentSuccess = callback
		c.OnPaymentFailure = callback
	}
}

// WithPaymentAttemptCallback sets the payment attempt callback
func WithPaymentAttemptCallback(callback func(PaymentEvent)) Option {
	return func(c *Config) {
		c.OnPaymentAttempt = callback
	}
}

// WithPaymentSuccessCallback sets the payment success callback
func WithPaymentSuccessCallback(callback func(PaymentEvent)) Option {
	return func(c *Config) {
		c.OnPaymentSuccess = callback
	}
}

// WithPaymentFailureCallback sets the payment failure callback
func WithPaymentFailureCallback(callback func(PaymentEvent)) Option {
	return func(c *Config) {
		c.OnPaymentFailure = callback
	}
}

// WithSelector sets a custom payment selector
func WithSelector(selector mcp.PaymentSelector) Option {
	return func(c *Config) {
		c.Selector = selector
	}
}

// WithVerbose enables verbose logging
func WithVerbose() Option {
	return func(c *Config) {
		c.Verbose = true
	}
}

// DefaultConfig returns a Config with default settings
func DefaultConfig(serverURL string) *Config {
	return &Config{
		ServerURL:  serverURL,
		HTTPClient: http.DefaultClient,
		Selector:   &x402.DefaultPaymentSelector{},
		Signers:    make([]mcp.Signer, 0),
	}
}
