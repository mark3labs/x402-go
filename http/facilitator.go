package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/http/internal/helpers"
)

// FacilitatorClient is a client for communicating with x402 facilitator services.
type FacilitatorClient struct {
	BaseURL       string
	Client        *http.Client
	VerifyTimeout time.Duration // Timeout for verify operations
	SettleTimeout time.Duration // Timeout for settle operations (longer due to blockchain tx)
	MaxRetries    int           // Maximum number of retry attempts for failed requests (default: 0)
	RetryDelay    time.Duration // Delay between retry attempts (default: 100ms)
}

// FacilitatorRequest is the request payload sent to the facilitator.
type FacilitatorRequest struct {
	X402Version         int                     `json:"x402Version"`
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
		X402Version:         1,
		PaymentPayload:      payment,
		PaymentRequirements: requirement,
	}

	// Marshal to JSON
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	return c.doWithRetry(func() (*VerifyResponse, error) {
		// Create request with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), c.VerifyTimeout)
		defer cancel()

		httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/verify", bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		// Send request
		resp, err := c.Client.Do(httpReq)
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

		if verifyResp.Payer != "" {
			return &verifyResp, nil
		}

		verifyResp.Payer = helpers.GetPayer(payment)

		return &verifyResp, nil
	})
}

// SupportedKind represents a supported payment type.
type SupportedKind struct {
	X402Version int            `json:"x402Version"`
	Scheme      string         `json:"scheme"`
	Network     string         `json:"network"`
	Extra       map[string]any `json:"extra,omitempty"`
}

// SupportedResponse is the response from the facilitator /supported endpoint.
type SupportedResponse struct {
	Kinds []SupportedKind `json:"kinds"`
}

// Supported queries the facilitator for supported payment types.
func (c *FacilitatorClient) Supported() (*SupportedResponse, error) {
	// Create request with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), c.VerifyTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/supported", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	resp, err := c.Client.Do(httpReq)
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
		X402Version:         1,
		PaymentPayload:      payment,
		PaymentRequirements: requirement,
	}

	// Marshal to JSON
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	return c.doSettleWithRetry(func() (*x402.SettlementResponse, error) {
		// Create request with timeout context (longer timeout for blockchain tx)
		ctx, cancel := context.WithTimeout(context.Background(), c.SettleTimeout)
		defer cancel()

		httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/settle", bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		// Send request
		resp, err := c.Client.Do(httpReq)
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
	})
}

// EnrichRequirements fetches supported payment types from the facilitator and
// enriches the provided payment requirements with network-specific data like feePayer.
// This is particularly useful for SVM chains where the feePayer must be specified.
func (c *FacilitatorClient) EnrichRequirements(requirements []x402.PaymentRequirement) ([]x402.PaymentRequirement, error) {
	// Fetch supported payment types
	supported, err := c.Supported()
	if err != nil {
		return requirements, fmt.Errorf("failed to fetch supported payment types: %w", err)
	}

	// Create a lookup map for supported kinds by network
	supportedMap := make(map[string]SupportedKind)
	for _, kind := range supported.Kinds {
		key := kind.Network + "-" + kind.Scheme
		supportedMap[key] = kind
	}

	// Enrich each requirement with extra data from the facilitator
	enriched := make([]x402.PaymentRequirement, len(requirements))
	for i, req := range requirements {
		enriched[i] = req
		key := req.Network + "-" + req.Scheme
		if kind, ok := supportedMap[key]; ok && kind.Extra != nil {
			// Initialize Extra map if it doesn't exist
			if enriched[i].Extra == nil {
				enriched[i].Extra = make(map[string]any)
			}
			// Merge facilitator's extra data into requirement
			for k, v := range kind.Extra {
				// Only set if not already present (user-specified values take precedence)
				if _, exists := enriched[i].Extra[k]; !exists {
					enriched[i].Extra[k] = v
				}
			}
		}
	}

	return enriched, nil
}

// doWithRetry executes a function with exponential backoff retry logic for transient failures.
// It automatically retries on facilitator unavailable errors up to MaxRetries times.
func (c *FacilitatorClient) doWithRetry(fn func() (*VerifyResponse, error)) (*VerifyResponse, error) {
	maxRetries := c.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	retryDelay := c.RetryDelay
	if retryDelay <= 0 {
		retryDelay = 100 * time.Millisecond
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Only retry on facilitator unavailable errors
		if !isFacilitatorUnavailableError(err) {
			return nil, err
		}
	}

	return nil, lastErr
}

// doSettleWithRetry executes a settle function with retry logic for transient failures.
func (c *FacilitatorClient) doSettleWithRetry(fn func() (*x402.SettlementResponse, error)) (*x402.SettlementResponse, error) {
	maxRetries := c.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	retryDelay := c.RetryDelay
	if retryDelay <= 0 {
		retryDelay = 100 * time.Millisecond
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Only retry on facilitator unavailable errors
		if !isFacilitatorUnavailableError(err) {
			return nil, err
		}
	}

	return nil, lastErr
}

// isFacilitatorUnavailableError checks if an error is a facilitator unavailable error.
// It uses errors.Is to properly detect wrapped errors.
func isFacilitatorUnavailableError(err error) bool {
	return errors.Is(err, x402.ErrFacilitatorUnavailable)
}
