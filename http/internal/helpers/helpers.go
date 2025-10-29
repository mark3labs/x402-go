// Package helpers provides shared helper functions for x402 HTTP middleware implementations.
// These helpers are used by stdlib, Gin, PocketBase, and Chi middleware to ensure consistent behavior.
package helpers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mark3labs/x402-go"
)

// ParsePaymentHeaderFromRequest parses the X-PAYMENT header from an http.Request and returns the payment payload.
// It decodes the base64-encoded JSON and validates the x402 protocol version.
//
// Returns x402.ErrMalformedHeader if the header is missing, invalid base64, or invalid JSON.
// Returns x402.ErrUnsupportedVersion if X402Version != 1.
func ParsePaymentHeaderFromRequest(r *http.Request) (x402.PaymentPayload, error) {
	var payment x402.PaymentPayload

	headerValue := r.Header.Get("X-PAYMENT")
	if headerValue == "" {
		return payment, x402.ErrMalformedHeader
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(headerValue)
	if err != nil {
		return payment, fmt.Errorf("%w: invalid base64 encoding", x402.ErrMalformedHeader)
	}

	// Parse JSON
	if err := json.Unmarshal(decoded, &payment); err != nil {
		return payment, fmt.Errorf("%w: invalid JSON", x402.ErrMalformedHeader)
	}

	// Validate version
	if payment.X402Version != 1 {
		return payment, x402.ErrUnsupportedVersion
	}

	return payment, nil
}

// FindMatchingRequirement finds a payment requirement that matches the provided payment's scheme and network.
//
// Returns x402.ErrUnsupportedScheme if no matching requirement is found.
func FindMatchingRequirement(payment x402.PaymentPayload, requirements []x402.PaymentRequirement) (x402.PaymentRequirement, error) {
	for _, req := range requirements {
		if req.Scheme == payment.Scheme && req.Network == payment.Network {
			return req, nil
		}
	}
	return x402.PaymentRequirement{}, fmt.Errorf("%w: no matching requirement for scheme=%s, network=%s",
		x402.ErrUnsupportedScheme, payment.Scheme, payment.Network)
}

// SendPaymentRequired sends a 402 Payment Required response with payment requirements in JSON format.
// The response includes x402Version field and the list of accepted payment methods.
func SendPaymentRequired(w http.ResponseWriter, requirements []x402.PaymentRequirement) {
	response := x402.PaymentRequirementsResponse{
		X402Version: 1,
		Error:       "Payment required for this resource",
		Accepts:     requirements,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusPaymentRequired)
	// Ignore encoding errors - headers are already sent with 402 status
	// The response body may be incomplete, but the client will see the correct status code
	_ = json.NewEncoder(w).Encode(response)
}

// AddPaymentResponseHeader adds the X-PAYMENT-RESPONSE header with base64-encoded settlement information.
// The header contains JSON-encoded SettlementResponse data.
//
// Returns an error if JSON marshaling fails.
func AddPaymentResponseHeader(w http.ResponseWriter, settlement *x402.SettlementResponse) error {
	// Marshal settlement response to JSON
	data, err := json.Marshal(settlement)
	if err != nil {
		return fmt.Errorf("failed to marshal settlement response: %w", err)
	}

	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(data)

	// Set header
	w.Header().Set("X-PAYMENT-RESPONSE", encoded)
	return nil
}
