// Package x402 provides types and utilities for implementing the x402 payment protocol.
package x402

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
)

// PaymentRequirement defines a single acceptable payment method for a protected resource.
type PaymentRequirement struct {
	Scheme            string         `json:"scheme"`
	Network           string         `json:"network"`
	MaxAmountRequired string         `json:"maxAmountRequired"`
	Asset             string         `json:"asset"`
	PayTo             string         `json:"payTo"`
	Resource          string         `json:"resource"`
	Description       string         `json:"description"`
	MimeType          string         `json:"mimeType,omitempty"`
	OutputSchema      map[string]any `json:"outputSchema,omitempty"`
	MaxTimeoutSeconds int            `json:"maxTimeoutSeconds"`
	Extra             map[string]any `json:"extra,omitempty"`
}

// PaymentRequirementsResponse is the complete response body for 402 Payment Required status.
type PaymentRequirementsResponse struct {
	X402Version int                  `json:"x402Version"`
	Error       string               `json:"error"`
	Accepts     []PaymentRequirement `json:"accepts"`
}

// PaymentPayload is the payment authorization data sent by the client.
type PaymentPayload struct {
	X402Version int             `json:"x402Version"`
	Scheme      string          `json:"scheme"`
	Network     string          `json:"network"`
	Payload     json.RawMessage `json:"payload"`
}

// SchemePayload is an interface for scheme-specific payment data.
type SchemePayload interface {
	Validate() error
}

// EVMPayload contains EIP-3009 authorization for EVM-based chains.
type EVMPayload struct {
	Signature     string        `json:"signature"`
	Authorization Authorization `json:"authorization"`
}

// Authorization contains the EIP-3009 authorization fields.
type Authorization struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	ValidAfter  string `json:"validAfter"`
	ValidBefore string `json:"validBefore"`
	Nonce       string `json:"nonce"`
}

// SVMPayload contains a serialized transaction for Solana-based chains.
type SVMPayload struct {
	Transaction string `json:"transaction"`
}

// SettlementResponse contains payment settlement result information.
type SettlementResponse struct {
	Success     bool   `json:"success"`
	ErrorReason string `json:"errorReason,omitempty"`
	Transaction string `json:"transaction"`
	Network     string `json:"network"`
	Payer       string `json:"payer"`
}

// EVM address pattern (0x + 40 hex characters)
var evmAddressPattern = regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)

// EVM signature pattern (0x + hex characters)
var evmSignaturePattern = regexp.MustCompile(`^0x[a-fA-F0-9]+$`)

// EVM nonce pattern (0x + 64 hex characters for 32 bytes)
var evmNoncePattern = regexp.MustCompile(`^0x[a-fA-F0-9]{64}$`)

// Validate validates a PaymentRequirement.
func (pr *PaymentRequirement) Validate() error {
	if pr.Scheme == "" {
		return fmt.Errorf("scheme is required")
	}
	if pr.Network == "" {
		return fmt.Errorf("network is required")
	}
	if pr.MaxAmountRequired == "" {
		return fmt.Errorf("maxAmountRequired is required")
	}
	if err := validateAmount(pr.MaxAmountRequired); err != nil {
		return fmt.Errorf("invalid maxAmountRequired: %w", err)
	}
	if pr.Asset == "" {
		return fmt.Errorf("asset is required")
	}
	if pr.PayTo == "" {
		return fmt.Errorf("payTo is required")
	}
	if pr.Resource == "" {
		return fmt.Errorf("resource is required")
	}
	if pr.Description == "" {
		return fmt.Errorf("description is required")
	}
	if pr.MaxTimeoutSeconds <= 0 {
		return fmt.Errorf("maxTimeoutSeconds must be positive")
	}
	return nil
}

// Validate validates an EVMPayload.
func (p *EVMPayload) Validate() error {
	if !evmSignaturePattern.MatchString(p.Signature) {
		return fmt.Errorf("invalid signature format")
	}
	if !evmAddressPattern.MatchString(p.Authorization.From) {
		return fmt.Errorf("invalid from address")
	}
	if !evmAddressPattern.MatchString(p.Authorization.To) {
		return fmt.Errorf("invalid to address")
	}
	if err := validateAmount(p.Authorization.Value); err != nil {
		return fmt.Errorf("invalid value: %w", err)
	}
	if !evmNoncePattern.MatchString(p.Authorization.Nonce) {
		return fmt.Errorf("invalid nonce format (must be 32 bytes)")
	}

	// Validate timestamps
	validAfter, err := strconv.ParseInt(p.Authorization.ValidAfter, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid validAfter timestamp: %w", err)
	}
	validBefore, err := strconv.ParseInt(p.Authorization.ValidBefore, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid validBefore timestamp: %w", err)
	}
	if validBefore <= validAfter {
		return fmt.Errorf("validBefore must be after validAfter")
	}

	return nil
}

// Validate validates an SVMPayload.
func (p *SVMPayload) Validate() error {
	if p.Transaction == "" {
		return fmt.Errorf("transaction is required")
	}
	// Note: Full base64 validation and Solana transaction deserialization
	// would be done by the facilitator
	return nil
}

// validateAmount validates that an amount string is a positive numeric value.
func validateAmount(amount string) error {
	if amount == "" {
		return fmt.Errorf("amount cannot be empty")
	}
	// Parse as integer to ensure it's a valid number
	val, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		return fmt.Errorf("amount must be a valid positive integer: %w", err)
	}
	if val == 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	return nil
}

// ValidateEVMAddress validates an EVM address format.
func ValidateEVMAddress(address string) error {
	if !evmAddressPattern.MatchString(address) {
		return fmt.Errorf("invalid EVM address format (must be 0x + 40 hex characters)")
	}
	return nil
}
