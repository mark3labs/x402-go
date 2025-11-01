// Package mcp provides x402 payment integration for MCP (Model Context Protocol).
package mcp

import (
	"errors"
	"fmt"

	"github.com/mark3labs/x402-go"
)

// MCP-specific error types wrapping x402 errors

var (
	// ErrPaymentRequired indicates that a payment is required to access the resource
	ErrPaymentRequired = errors.New("payment required")

	// ErrNoMatchingSigner indicates that no configured signer can fulfill the payment requirements
	ErrNoMatchingSigner = errors.New("no matching signer for payment requirements")

	// ErrInsufficientBalance indicates that the signer's balance is too low
	ErrInsufficientBalance = errors.New("insufficient balance")

	// ErrPaymentRejected indicates that the facilitator rejected the payment
	ErrPaymentRejected = errors.New("payment rejected by facilitator")

	// ErrSettlementFailed indicates that the blockchain transaction failed
	ErrSettlementFailed = errors.New("payment settlement failed")

	// ErrInvalidPaymentPayload indicates that the payment payload is malformed
	ErrInvalidPaymentPayload = errors.New("invalid payment payload")

	// ErrNoPaymentRequirements indicates that no payment requirements were found in the 402 error
	ErrNoPaymentRequirements = errors.New("no payment requirements in 402 error")

	// ErrSessionTerminated indicates that the MCP session has ended
	ErrSessionTerminated = errors.New("mcp session terminated")

	// ErrInvalidRequest indicates that the MCP request is malformed
	ErrInvalidRequest = errors.New("invalid mcp request")

	// ErrToolNotFound indicates that the requested tool does not exist
	ErrToolNotFound = errors.New("tool not found")

	// ErrToolExecutionFailed indicates that the tool handler returned an error
	ErrToolExecutionFailed = errors.New("tool execution failed")

	// ErrFacilitatorUnavailable indicates that the facilitator is unreachable
	ErrFacilitatorUnavailable = errors.New("facilitator unavailable")

	// ErrVerificationTimeout indicates that payment verification took too long
	ErrVerificationTimeout = errors.New("payment verification timeout")

	// ErrSettlementTimeout indicates that payment settlement took too long
	ErrSettlementTimeout = errors.New("payment settlement timeout")
)

// PaymentError wraps an x402 error with MCP-specific context
type PaymentError struct {
	Err      error
	Tool     string
	Resource string
	Context  string
}

func (e *PaymentError) Error() string {
	if e.Tool != "" {
		return fmt.Sprintf("payment error for tool %s: %v", e.Tool, e.Err)
	}
	if e.Resource != "" {
		return fmt.Sprintf("payment error for resource %s: %v", e.Resource, e.Err)
	}
	if e.Context != "" {
		return fmt.Sprintf("payment error (%s): %v", e.Context, e.Err)
	}
	return fmt.Sprintf("payment error: %v", e.Err)
}

func (e *PaymentError) Unwrap() error {
	return e.Err
}

// WrapX402Error wraps an x402 error as a PaymentError
func WrapX402Error(err error, tool string) error {
	if err == nil {
		return nil
	}
	return &PaymentError{
		Err:  err,
		Tool: tool,
	}
}

// IsPaymentError checks if an error is payment-related
func IsPaymentError(err error) bool {
	if err == nil {
		return false
	}
	var paymentErr *PaymentError
	return errors.As(err, &paymentErr) ||
		errors.Is(err, ErrPaymentRequired) ||
		errors.Is(err, ErrNoMatchingSigner) ||
		errors.Is(err, ErrInsufficientBalance) ||
		errors.Is(err, ErrPaymentRejected) ||
		errors.Is(err, ErrSettlementFailed) ||
		errors.Is(err, ErrInvalidPaymentPayload) ||
		errors.Is(err, ErrNoPaymentRequirements) ||
		errors.Is(err, ErrFacilitatorUnavailable) ||
		errors.Is(err, ErrVerificationTimeout) ||
		errors.Is(err, ErrSettlementTimeout) ||
		errors.Is(err, x402.ErrNoValidSigner) ||
		errors.Is(err, x402.ErrSigningFailed) ||
		errors.Is(err, x402.ErrVerificationFailed)
}
