// Package helpers provides internal HTTP utilities for x402 v2 protocol handling.
package helpers

import (
	"encoding/json"
	"net/http"

	v2 "github.com/mark3labs/x402-go/v2"
	"github.com/mark3labs/x402-go/v2/encoding"
)

// ParsePaymentHeader extracts and decodes a PaymentPayload from the X-PAYMENT header.
// Returns ErrMalformedHeader if the header is missing or invalid.
func ParsePaymentHeader(r *http.Request) (*v2.PaymentPayload, error) {
	paymentHeader := r.Header.Get("X-PAYMENT")
	if paymentHeader == "" {
		return nil, v2.ErrMalformedHeader
	}

	payment, err := encoding.DecodePayment(paymentHeader)
	if err != nil {
		return nil, v2.NewPaymentError(v2.ErrCodeInvalidRequirements, "failed to decode payment header", err)
	}

	// Validate protocol version
	if payment.X402Version != v2.X402Version {
		return nil, v2.NewPaymentError(v2.ErrCodeUnsupportedScheme, "unsupported x402 version", v2.ErrUnsupportedVersion)
	}

	return &payment, nil
}

// SendPaymentRequired writes a 402 Payment Required response with the given requirements.
func SendPaymentRequired(w http.ResponseWriter, resource v2.ResourceInfo, requirements []v2.PaymentRequirements, errMsg string) {
	response := v2.PaymentRequired{
		X402Version: v2.X402Version,
		Error:       errMsg,
		Resource:    resource,
		Accepts:     requirements,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusPaymentRequired)
	json.NewEncoder(w).Encode(response)
}

// AddPaymentResponseHeader adds the X-PAYMENT-RESPONSE header with settlement information.
func AddPaymentResponseHeader(w http.ResponseWriter, settlement *v2.SettleResponse) error {
	encoded, err := encoding.EncodeSettlement(*settlement)
	if err != nil {
		return err
	}
	w.Header().Set("X-PAYMENT-RESPONSE", encoded)
	return nil
}

// ParsePaymentRequirements extracts PaymentRequired from a 402 response body.
func ParsePaymentRequirements(resp *http.Response) (*v2.PaymentRequired, error) {
	var paymentReq v2.PaymentRequired
	if err := json.NewDecoder(resp.Body).Decode(&paymentReq); err != nil {
		return nil, v2.NewPaymentError(v2.ErrCodeInvalidRequirements, "failed to decode payment requirements", err)
	}

	// Validate we have at least one requirement
	if len(paymentReq.Accepts) == 0 {
		return nil, v2.NewPaymentError(v2.ErrCodeInvalidRequirements, "no payment requirements in response", v2.ErrInvalidRequirements)
	}

	return &paymentReq, nil
}

// ParseSettlement extracts settlement information from the X-PAYMENT-RESPONSE header.
// Returns nil if the header is empty or cannot be parsed.
func ParseSettlement(headerValue string) *v2.SettleResponse {
	if headerValue == "" {
		return nil
	}

	settlement, err := encoding.DecodeSettlement(headerValue)
	if err != nil {
		return nil
	}

	return &settlement
}

// BuildPaymentHeader creates the X-PAYMENT header value from a PaymentPayload.
func BuildPaymentHeader(payment *v2.PaymentPayload) (string, error) {
	return encoding.EncodePayment(*payment)
}

// BuildResourceURL constructs the full URL for the protected resource from the request.
func BuildResourceURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host + r.RequestURI
}
