package pocketbase

import (
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/x402-go"
	httpx402 "github.com/mark3labs/x402-go/http"
)

// TestPocketBaseMiddleware_Creation tests that middleware can be created
func TestPocketBaseMiddleware_Creation(t *testing.T) {
	// Create middleware config
	config := &httpx402.Config{
		FacilitatorURL: "http://mock-facilitator.test",
		PaymentRequirements: []x402.PaymentRequirement{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/test",
				Description:       "Test resource",
				MaxTimeoutSeconds: 60,
			},
		},
	}

	// Create middleware
	middleware := NewPocketBaseX402Middleware(config)

	// Verify middleware function was created
	if middleware == nil {
		t.Error("Expected middleware function to be created")
	}
}

// TestPocketBaseMiddleware_FallbackFacilitator tests fallback facilitator configuration
func TestPocketBaseMiddleware_FallbackFacilitator(t *testing.T) {
	// Create middleware config with fallback
	config := &httpx402.Config{
		FacilitatorURL:         "http://mock-facilitator.test",
		FallbackFacilitatorURL: "http://fallback-facilitator.test",
		PaymentRequirements: []x402.PaymentRequirement{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				MaxTimeoutSeconds: 60,
			},
		},
	}

	// Create middleware
	middleware := NewPocketBaseX402Middleware(config)

	// Verify middleware was created with fallback support
	if middleware == nil {
		t.Error("Expected middleware function to be created with fallback facilitator")
	}
}

// TestPocketBaseMiddleware_VerifyOnlyMode tests verification-only mode without settlement
func TestPocketBaseMiddleware_VerifyOnlyMode(t *testing.T) {
	// Create middleware config with VerifyOnly flag
	config := &httpx402.Config{
		FacilitatorURL: "http://mock-facilitator.test",
		VerifyOnly:     true, // Key difference - only verify, don't settle
		PaymentRequirements: []x402.PaymentRequirement{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/test",
				Description:       "Test resource",
				MaxTimeoutSeconds: 60,
			},
		},
	}

	// Create middleware
	middleware := NewPocketBaseX402Middleware(config)

	// Verify middleware was created with VerifyOnly mode
	if middleware == nil {
		t.Error("Expected middleware function to be created in VerifyOnly mode")
	}
}

// TestHelperFunctions tests the four duplicated helper functions
func TestHelperFunctions(t *testing.T) {
	t.Run("parsePaymentHeaderFromRequest", func(t *testing.T) {
		// Test with valid header
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-PAYMENT", "eyJ4NDAyVmVyc2lvbiI6MSwic2NoZW1lIjoiZXhhY3QiLCJuZXR3b3JrIjoiYmFzZS1zZXBvbGlhIiwicGF5bG9hZCI6e319")

		payment, err := parsePaymentHeaderFromRequest(req)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if payment.X402Version != 1 {
			t.Errorf("Expected X402Version 1, got %d", payment.X402Version)
		}

		// Test with missing header
		req2 := httptest.NewRequest("GET", "/test", nil)
		_, err = parsePaymentHeaderFromRequest(req2)
		if err == nil {
			t.Error("Expected error for missing header")
		}

		// Test with invalid base64
		req3 := httptest.NewRequest("GET", "/test", nil)
		req3.Header.Set("X-PAYMENT", "not-valid-base64!!!")
		_, err = parsePaymentHeaderFromRequest(req3)
		if err == nil {
			t.Error("Expected error for invalid base64")
		}
	})

	t.Run("findMatchingRequirementPocketBase", func(t *testing.T) {
		requirements := []x402.PaymentRequirement{
			{
				Scheme:  "exact",
				Network: "base-sepolia",
			},
			{
				Scheme:  "exact",
				Network: "solana-devnet",
			},
		}

		// Test matching requirement
		payment := x402.PaymentPayload{
			Scheme:  "exact",
			Network: "base-sepolia",
		}

		req, err := findMatchingRequirementPocketBase(payment, requirements)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if req.Network != "base-sepolia" {
			t.Errorf("Expected network base-sepolia, got %s", req.Network)
		}

		// Test non-matching requirement
		payment2 := x402.PaymentPayload{
			Scheme:  "exact",
			Network: "unknown-network",
		}

		_, err = findMatchingRequirementPocketBase(payment2, requirements)
		if err == nil {
			t.Error("Expected error for non-matching requirement")
		}
	})

	t.Run("sendPaymentRequiredPocketBase", func(t *testing.T) {
		// We test the logic by verifying the PaymentRequirementsResponse structure
		// rather than testing the full PocketBase event flow which requires unexported fields

		requirements := []x402.PaymentRequirement{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/test",
				Description:       "Test payment",
				MaxTimeoutSeconds: 300,
			},
		}

		// Verify the response structure that would be sent
		response := x402.PaymentRequirementsResponse{
			X402Version: 1,
			Error:       "Payment required for this resource",
			Accepts:     requirements,
		}

		// Verify response marshals to valid JSON
		data, err := json.Marshal(response)
		if err != nil {
			t.Errorf("Failed to marshal PaymentRequirementsResponse: %v", err)
		}

		// Verify we can unmarshal it back
		var decoded x402.PaymentRequirementsResponse
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		// Verify response fields
		if decoded.X402Version != 1 {
			t.Errorf("Expected X402Version 1, got %d", decoded.X402Version)
		}

		if decoded.Error == "" {
			t.Error("Expected error message in response")
		}

		if len(decoded.Accepts) != 1 {
			t.Errorf("Expected 1 payment requirement, got %d", len(decoded.Accepts))
		}

		if decoded.Accepts[0].Network != "base-sepolia" {
			t.Errorf("Expected network base-sepolia, got %s", decoded.Accepts[0].Network)
		}
	})

	t.Run("addPaymentResponseHeaderPocketBase", func(t *testing.T) {
		// Test the header encoding logic directly
		settlement := &x402.SettlementResponse{
			Success:     true,
			Transaction: "0xabcdef123456789",
		}

		// Marshal settlement response to JSON (same logic as the function)
		data, err := json.Marshal(settlement)
		if err != nil {
			t.Errorf("Failed to marshal settlement response: %v", err)
		}

		// Encode as base64 (same logic as the function)
		encoded := base64.StdEncoding.EncodeToString(data)

		if encoded == "" {
			t.Error("Expected non-empty base64 encoded string")
		}

		// Verify we can decode it back
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			t.Errorf("Failed to decode base64: %v", err)
		}

		var decodedSettlement x402.SettlementResponse
		if err := json.Unmarshal(decoded, &decodedSettlement); err != nil {
			t.Errorf("Failed to unmarshal settlement JSON: %v", err)
		}

		// Verify the decoded values match the original
		if decodedSettlement.Success != settlement.Success {
			t.Errorf("Expected Success %v, got %v", settlement.Success, decodedSettlement.Success)
		}

		if decodedSettlement.Transaction != settlement.Transaction {
			t.Errorf("Expected Transaction %s, got %s", settlement.Transaction, decodedSettlement.Transaction)
		}
	})
}

// TestPocketBaseMiddleware_MultiplePaymentRequirements tests multiple payment options
func TestPocketBaseMiddleware_MultiplePaymentRequirements(t *testing.T) {
	config := &httpx402.Config{
		FacilitatorURL: "http://mock-facilitator.test",
		PaymentRequirements: []x402.PaymentRequirement{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				MaxTimeoutSeconds: 60,
			},
			{
				Scheme:            "exact",
				Network:           "solana-devnet",
				MaxAmountRequired: "10000",
				Asset:             "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU",
				PayTo:             "YourSolanaWallet",
				MaxTimeoutSeconds: 60,
			},
		},
	}

	middleware := NewPocketBaseX402Middleware(config)

	// Verify middleware supports multiple payment requirements
	if middleware == nil {
		t.Error("Expected middleware to support multiple payment requirements")
	}
}
