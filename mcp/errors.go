package mcp

import (
	"errors"
	"fmt"
)

// MCP-specific error types
var (
	// ErrPaymentRequired indicates payment is required for the resource
	ErrPaymentRequired = errors.New("payment required")

	// ErrNoMatchingSigner indicates no configured signer can fulfill requirements
	ErrNoMatchingSigner = errors.New("no matching signer for payment requirements")

	// ErrInsufficientBalance indicates signer has insufficient balance
	ErrInsufficientBalance = errors.New("insufficient balance")

	// ErrPaymentRejected indicates facilitator rejected the payment
	ErrPaymentRejected = errors.New("payment rejected by facilitator")

	// ErrSettlementFailed indicates blockchain transaction failed
	ErrSettlementFailed = errors.New("settlement failed")

	// ErrSessionTerminated indicates MCP session has ended
	ErrSessionTerminated = errors.New("session terminated")

	// ErrInvalidRequest indicates malformed MCP request
	ErrInvalidRequest = errors.New("invalid request")

	// ErrToolNotFound indicates unknown tool name
	ErrToolNotFound = errors.New("tool not found")

	// ErrToolExecutionFailed indicates tool handler error
	ErrToolExecutionFailed = errors.New("tool execution failed")

	// ErrVerificationTimeout indicates payment verification exceeded timeout
	ErrVerificationTimeout = errors.New("payment verification timeout")

	// ErrSettlementTimeout indicates payment settlement exceeded timeout
	ErrSettlementTimeout = errors.New("payment settlement timeout")
)

// PaymentError wraps payment-related errors with additional context
type PaymentError struct {
	Op      string // Operation that failed
	Err     error  // Underlying error
	Network string // Network where error occurred
	Amount  string // Payment amount
}

func (e *PaymentError) Error() string {
	if e.Network != "" && e.Amount != "" {
		return fmt.Sprintf("%s: %v (network: %s, amount: %s)", e.Op, e.Err, e.Network, e.Amount)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *PaymentError) Unwrap() error {
	return e.Err
}

// NewPaymentError creates a new PaymentError
func NewPaymentError(op string, err error, network, amount string) error {
	return &PaymentError{
		Op:      op,
		Err:     err,
		Network: network,
		Amount:  amount,
	}
}
