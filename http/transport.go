package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mark3labs/x402-go"
)

// X402Transport is a custom RoundTripper that handles x402 payment flows.
// It wraps an existing http.RoundTripper and automatically handles 402 Payment Required responses.
type X402Transport struct {
	// Base is the underlying RoundTripper (typically http.DefaultTransport).
	Base http.RoundTripper

	// Signers is the list of available payment signers.
	Signers []x402.Signer

	// Selector is used to choose the appropriate signer and create payments.
	Selector x402.PaymentSelector
}

// RoundTrip implements http.RoundTripper.
// It makes the initial request, and if a 402 Payment Required response is received,
// it automatically signs a payment and retries the request.
func (t *X402Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Ensure we have a base transport
	if t.Base == nil {
		t.Base = http.DefaultTransport
	}

	// Clone the request to avoid modifying the original
	reqCopy := req.Clone(req.Context())

	// Make the first attempt
	resp, err := t.Base.RoundTrip(reqCopy)
	if err != nil {
		return nil, err
	}

	// Check if payment is required
	if resp.StatusCode != http.StatusPaymentRequired {
		return resp, nil
	}

	// Parse payment requirements from 402 response
	requirements, err := parsePaymentRequirements(resp)
	if err != nil {
		resp.Body.Close()
		return nil, x402.NewPaymentError(x402.ErrCodeInvalidRequirements, "failed to parse payment requirements", err)
	}

	// Close the 402 response body
	resp.Body.Close()

	// Select signer and create payment
	payment, err := t.Selector.SelectAndSign(requirements, t.Signers)
	if err != nil {
		return nil, err
	}

	// Build payment header
	paymentHeader, err := buildPaymentHeader(payment)
	if err != nil {
		return nil, x402.NewPaymentError(x402.ErrCodeSigningFailed, "failed to build payment header", err)
	}

	// Clone the request again for the retry
	reqRetry := req.Clone(req.Context())

	// Add payment header
	reqRetry.Header.Set("X-PAYMENT", paymentHeader)

	// Retry the request with payment
	respRetry, err := t.Base.RoundTrip(reqRetry)
	if err != nil {
		return nil, err
	}

	return respRetry, nil
}

// parsePaymentRequirements extracts payment requirements from a 402 response.
func parsePaymentRequirements(resp *http.Response) (*x402.PaymentRequirement, error) {
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// The response body should be a PaymentRequirementsResponse with an accepts array
	var paymentReqResp struct {
		X402Version int    `json:"x402Version"`
		Error       string `json:"error"`
		Accepts     []struct {
			Scheme            string                 `json:"scheme"`
			Network           string                 `json:"network"`
			MaxAmountRequired string                 `json:"maxAmountRequired"`
			Asset             string                 `json:"asset"`
			PayTo             string                 `json:"payTo"`
			Resource          string                 `json:"resource"`
			Description       string                 `json:"description,omitempty"`
			MimeType          string                 `json:"mimeType,omitempty"`
			MaxTimeoutSeconds int                    `json:"maxTimeoutSeconds"`
			Extra             map[string]interface{} `json:"extra,omitempty"`
		} `json:"accepts"`
	}

	if err := json.Unmarshal(body, &paymentReqResp); err != nil {
		return nil, fmt.Errorf("failed to parse payment requirements JSON: %w", err)
	}

	// Validate we got at least one payment requirement
	if len(paymentReqResp.Accepts) == 0 {
		return nil, fmt.Errorf("no payment requirements in response")
	}

	// Use the first requirement (for now, client doesn't support selecting from multiple)
	req := paymentReqResp.Accepts[0]

	requirements := &x402.PaymentRequirement{
		Scheme:            req.Scheme,
		Network:           req.Network,
		MaxAmountRequired: req.MaxAmountRequired,
		Asset:             req.Asset,
		PayTo:             req.PayTo,
		Resource:          req.Resource,
		Description:       req.Description,
		MimeType:          req.MimeType,
		MaxTimeoutSeconds: req.MaxTimeoutSeconds,
		Extra:             req.Extra,
	}

	return requirements, nil
}

// buildPaymentHeader creates the X-PAYMENT header value from a payment payload.
func buildPaymentHeader(payment *x402.PaymentPayload) (string, error) {
	// Serialize payment to JSON
	paymentJSON, err := json.Marshal(payment)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payment: %w", err)
	}

	// Encode to base64
	paymentBase64 := base64.StdEncoding.EncodeToString(paymentJSON)

	return paymentBase64, nil
}

// parseSettlement extracts settlement information from the X-SETTLEMENT header.
func parseSettlement(headerValue string) (*x402.SettlementResponse, error) {
	// Decode base64
	settlementJSON, err := base64.StdEncoding.DecodeString(headerValue)
	if err != nil {
		return nil, fmt.Errorf("failed to decode settlement header: %w", err)
	}

	// Parse JSON
	var settlement x402.SettlementResponse
	if err := json.Unmarshal(settlementJSON, &settlement); err != nil {
		return nil, fmt.Errorf("failed to parse settlement JSON: %w", err)
	}

	return &settlement, nil
}

// RequestWithBody clones an HTTP request with a new body.
// This is needed because request bodies can only be read once.
func RequestWithBody(req *http.Request, body []byte) *http.Request {
	clone := req.Clone(req.Context())
	clone.Body = io.NopCloser(bytes.NewReader(body))
	clone.ContentLength = int64(len(body))
	return clone
}
