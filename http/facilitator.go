package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mark3labs/x402-go"
)

// FacilitatorClient is a client for communicating with x402 facilitator services.
type FacilitatorClient struct {
	BaseURL string
	Client  *http.Client
}

// FacilitatorRequest is the request payload sent to the facilitator.
type FacilitatorRequest struct {
	PaymentPayload      x402.PaymentPayload     `json:"paymentPayload"`
	PaymentRequirements x402.PaymentRequirement `json:"paymentRequirements"`
}

// VerifyResponse is the response from the facilitator /verify endpoint.
type VerifyResponse struct {
	IsValid       bool   `json:"isValid"`
	InvalidReason string `json:"invalidReason,omitempty"`
	Payer         string `json:"payer"`
}

// Verify verifies a payment authorization without executing the transaction.
func (c *FacilitatorClient) Verify(payment x402.PaymentPayload, requirement x402.PaymentRequirement) (*VerifyResponse, error) {
	// Create request payload
	req := FacilitatorRequest{
		PaymentPayload:      payment,
		PaymentRequirements: requirement,
	}

	// Marshal to JSON
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send POST request to /verify
	resp, err := c.Client.Post(c.BaseURL+"/verify", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", x402.ErrFacilitatorUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", x402.ErrVerificationFailed, resp.StatusCode)
	}

	// Parse response
	var verifyResp VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return nil, fmt.Errorf("failed to decode verify response: %w", err)
	}

	return &verifyResp, nil
}

// SupportedKind represents a supported payment type.
type SupportedKind struct {
	X402Version int    `json:"x402Version"`
	Scheme      string `json:"scheme"`
	Network     string `json:"network"`
}

// SupportedResponse is the response from the facilitator /supported endpoint.
type SupportedResponse struct {
	Kinds []SupportedKind `json:"kinds"`
}

// Supported queries the facilitator for supported payment types.
func (c *FacilitatorClient) Supported() (*SupportedResponse, error) {
	// Send GET request to /supported
	resp, err := c.Client.Get(c.BaseURL + "/supported")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", x402.ErrFacilitatorUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("supported endpoint failed: status %d", resp.StatusCode)
	}

	// Parse response
	var supportedResp SupportedResponse
	if err := json.NewDecoder(resp.Body).Decode(&supportedResp); err != nil {
		return nil, fmt.Errorf("failed to decode supported response: %w", err)
	}

	return &supportedResp, nil
}

// Settle executes a verified payment on the blockchain.
func (c *FacilitatorClient) Settle(payment x402.PaymentPayload, requirement x402.PaymentRequirement) (*x402.SettlementResponse, error) {
	// Create request payload
	req := FacilitatorRequest{
		PaymentPayload:      payment,
		PaymentRequirements: requirement,
	}

	// Marshal to JSON
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send POST request to /settle
	resp, err := c.Client.Post(c.BaseURL+"/settle", "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", x402.ErrFacilitatorUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", x402.ErrSettlementFailed, resp.StatusCode)
	}

	// Parse response
	var settlementResp x402.SettlementResponse
	if err := json.NewDecoder(resp.Body).Decode(&settlementResp); err != nil {
		return nil, fmt.Errorf("failed to decode settlement response: %w", err)
	}

	return &settlementResp, nil
}
