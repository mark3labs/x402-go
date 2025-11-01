package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/facilitator"
)

func TestFacilitatorClient_Verify(t *testing.T) {
	// Create a mock facilitator server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/verify" {
			t.Errorf("Expected path /verify, got %s", r.URL.Path)
		}

		response := facilitator.VerifyResponse{
			IsValid: true,
			Payer:   "0x857b06519E91e3A54538791bDbb0E22373e36b66",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer mockServer.Close()

	client := &FacilitatorClient{
		BaseURL:       mockServer.URL,
		Client:        &http.Client{},
		VerifyTimeout: 5 * time.Second,
		SettleTimeout: 60 * time.Second,
	}

	payload := x402.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "base-sepolia",
	}

	requirement := x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "base-sepolia",
		MaxAmountRequired: "10000",
		Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
		Resource:          "https://api.example.com/test",
		Description:       "Test resource",
		MaxTimeoutSeconds: 60,
	}

	resp, err := client.Verify(context.Background(), payload, requirement)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if !resp.IsValid {
		t.Error("Expected IsValid to be true")
	}

	if resp.Payer != "0x857b06519E91e3A54538791bDbb0E22373e36b66" {
		t.Errorf("Expected payer address, got %s", resp.Payer)
	}
}

func TestFacilitatorClient_Settle(t *testing.T) {
	// Create a mock facilitator server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/settle" {
			t.Errorf("Expected path /settle, got %s", r.URL.Path)
		}

		response := x402.SettlementResponse{
			Success:     true,
			Transaction: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			Network:     "base-sepolia",
			Payer:       "0x857b06519E91e3A54538791bDbb0E22373e36b66",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer mockServer.Close()

	client := &FacilitatorClient{
		BaseURL:       mockServer.URL,
		Client:        &http.Client{},
		VerifyTimeout: 5 * time.Second,
		SettleTimeout: 60 * time.Second,
	}

	payload := x402.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "base-sepolia",
	}

	requirement := x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "base-sepolia",
		MaxAmountRequired: "10000",
		Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
		Resource:          "https://api.example.com/test",
		Description:       "Test resource",
		MaxTimeoutSeconds: 60,
	}

	resp, err := client.Settle(context.Background(), payload, requirement)
	if err != nil {
		t.Fatalf("Settle failed: %v", err)
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}

	if resp.Transaction == "" {
		t.Error("Expected transaction hash")
	}
}
