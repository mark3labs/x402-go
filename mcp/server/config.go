package server

import (
	"github.com/mark3labs/x402-go/mcp"
)

// Config holds configuration for the MCP server with x402 payment support
type Config struct {
	// FacilitatorURL is the URL of the x402 facilitator service
	FacilitatorURL string

	// VerifyOnly when true, skips payment settlement (useful for testing)
	VerifyOnly bool

	// Verbose enables detailed logging
	Verbose bool

	// PaymentTools maps tool names to their payment requirements
	// Key: tool name, Value: list of acceptable payment options
	PaymentTools map[string][]mcp.PaymentRequirement
}

// DefaultConfig returns a Config with default settings
func DefaultConfig() *Config {
	return &Config{
		FacilitatorURL: "https://facilitator.x402.rs",
		VerifyOnly:     false,
		Verbose:        false,
		PaymentTools:   make(map[string][]mcp.PaymentRequirement),
	}
}

// AddPaymentTool adds payment requirements for a tool
func (c *Config) AddPaymentTool(toolName string, requirements ...mcp.PaymentRequirement) {
	if c.PaymentTools == nil {
		c.PaymentTools = make(map[string][]mcp.PaymentRequirement)
	}
	c.PaymentTools[toolName] = requirements
}

// RequiresPayment checks if a tool requires payment
func (c *Config) RequiresPayment(toolName string) bool {
	if c.PaymentTools == nil {
		return false
	}
	reqs, exists := c.PaymentTools[toolName]
	return exists && len(reqs) > 0
}

// GetPaymentRequirements returns the payment requirements for a tool
func (c *Config) GetPaymentRequirements(toolName string) []mcp.PaymentRequirement {
	if c.PaymentTools == nil {
		return nil
	}
	return c.PaymentTools[toolName]
}
