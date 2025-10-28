package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mark3labs/x402-go"
)

// sendPaymentRequired sends a 402 Payment Required response with payment requirements.
func sendPaymentRequired(w http.ResponseWriter, config *Config) {
	sendPaymentRequiredWithRequirements(w, config.PaymentRequirements)
}

// sendPaymentRequiredWithRequirements sends a 402 Payment Required response with specific payment requirements.
func sendPaymentRequiredWithRequirements(w http.ResponseWriter, requirements []x402.PaymentRequirement) {
	response := x402.PaymentRequirementsResponse{
		X402Version: 1,
		Error:       "Payment required for this resource",
		Accepts:     requirements,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusPaymentRequired)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// At this point headers are already sent, just log the error
		// Note: In production, consider using a structured logger
		http.Error(w, fmt.Sprintf("Failed to encode payment requirements: %v", err), http.StatusInternalServerError)
	}
}

// parsePaymentHeader parses the X-PAYMENT header and returns the payment payload.
func parsePaymentHeader(r *http.Request) (x402.PaymentPayload, error) {
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

// findMatchingRequirement finds a payment requirement that matches the provided payment.
func findMatchingRequirement(payment x402.PaymentPayload, requirements []x402.PaymentRequirement) (x402.PaymentRequirement, error) {
	for _, req := range requirements {
		if req.Scheme == payment.Scheme && req.Network == payment.Network {
			return req, nil
		}
	}
	return x402.PaymentRequirement{}, fmt.Errorf("%w: no matching requirement for scheme=%s, network=%s",
		x402.ErrUnsupportedScheme, payment.Scheme, payment.Network)
}

// addPaymentResponseHeader adds the X-PAYMENT-RESPONSE header with settlement information.
func addPaymentResponseHeader(w http.ResponseWriter, settlement *x402.SettlementResponse) error {
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
