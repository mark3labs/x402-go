package chi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/x402-go"
	httpx402 "github.com/mark3labs/x402-go/http"
)

// TestNewChiX402Middleware_Constructor tests middleware constructor
func TestNewChiX402Middleware_Constructor(t *testing.T) {
	config := &httpx402.Config{
		FacilitatorURL: "http://mock-facilitator.test",
		PaymentRequirements: []x402.PaymentRequirement{{
			Scheme:            "exact",
			Network:           "base-sepolia",
			MaxAmountRequired: "10000",
			Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			MaxTimeoutSeconds: 300,
		}},
	}

	middleware := NewChiX402Middleware(config)
	if middleware == nil {
		t.Fatal("Expected non-nil middleware function")
	}

	// Verify it returns a valid http.Handler wrapper
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}
}

// TestChiMiddleware_MissingPayment tests 402 response when no payment header
func TestChiMiddleware_MissingPayment(t *testing.T) {
	config := &httpx402.Config{
		FacilitatorURL: "http://mock-facilitator.test",
		PaymentRequirements: []x402.PaymentRequirement{{
			Scheme:            "exact",
			Network:           "base-sepolia",
			MaxAmountRequired: "10000",
			Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			MaxTimeoutSeconds: 300,
		}},
	}

	middleware := NewChiX402Middleware(config)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called without payment")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusPaymentRequired {
		t.Errorf("Expected status %d, got %d", http.StatusPaymentRequired, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

// TestChiMiddleware_OptionsRequestBypass tests OPTIONS bypass for CORS
func TestChiMiddleware_OptionsRequestBypass(t *testing.T) {
	config := &httpx402.Config{
		FacilitatorURL: "http://mock-facilitator.test",
		PaymentRequirements: []x402.PaymentRequirement{{
			Scheme:            "exact",
			Network:           "base-sepolia",
			MaxAmountRequired: "10000",
			Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			MaxTimeoutSeconds: 300,
		}},
	}

	handlerCalled := false
	middleware := NewChiX402Middleware(config)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("Handler should be called for OPTIONS request (bypass payment check)")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d for OPTIONS, got %d", http.StatusOK, rec.Code)
	}
}

// TestChiMiddleware_InvalidPaymentHeader tests 400 response for malformed header
func TestChiMiddleware_InvalidPaymentHeader(t *testing.T) {
	config := &httpx402.Config{
		FacilitatorURL: "http://mock-facilitator.test",
		PaymentRequirements: []x402.PaymentRequirement{{
			Scheme:            "exact",
			Network:           "base-sepolia",
			MaxAmountRequired: "10000",
			Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			MaxTimeoutSeconds: 300,
		}},
	}

	middleware := NewChiX402Middleware(config)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with invalid payment")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-PAYMENT", "invalid-base64!@#")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid payment, got %d", http.StatusBadRequest, rec.Code)
	}
}

// TestChiMiddleware_VerifyOnlyMode tests verify-only configuration
func TestChiMiddleware_VerifyOnlyMode(t *testing.T) {
	config := &httpx402.Config{
		FacilitatorURL: "http://mock-facilitator.test",
		VerifyOnly:     true, // Skip settlement
		PaymentRequirements: []x402.PaymentRequirement{{
			Scheme:            "exact",
			Network:           "base-sepolia",
			MaxAmountRequired: "10000",
			Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
			PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			MaxTimeoutSeconds: 300,
		}},
	}

	middleware := NewChiX402Middleware(config)
	if middleware == nil {
		t.Fatal("Expected non-nil middleware with VerifyOnly=true")
	}

	// This test will require mock facilitator to fully verify
	// For now, just ensure constructor works with VerifyOnly flag
}
