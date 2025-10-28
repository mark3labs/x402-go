package quickstart_test

// This file tests that all quickstart examples compile correctly

import (
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/evm"
	x402http "github.com/mark3labs/x402-go/http"
	"github.com/mark3labs/x402-go/svm"
)

// TestQuickstartExample1 - Basic single EVM signer
func TestQuickstartExample1(t *testing.T) {
	signer, err := evm.NewSigner(
		evm.WithPrivateKey("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		evm.WithNetwork("base"),
		evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
	)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	client, err := x402http.NewClient(
		x402http.WithSigner(signer),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

// TestQuickstartExample2 - Multi-signer setup
func TestQuickstartExample2(t *testing.T) {
	evmSigner, err := evm.NewSigner(
		evm.WithPrivateKey("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		evm.WithNetwork("base"),
		evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
		evm.WithPriority(1),
	)
	if err != nil {
		t.Fatalf("Failed to create EVM signer: %v", err)
	}

	client, err := x402http.NewClient(
		x402http.WithSigner(evmSigner),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

// TestQuickstartExample3 - Per-transaction limits
func TestQuickstartExample3(t *testing.T) {
	signer, err := evm.NewSigner(
		evm.WithPrivateKey("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		evm.WithNetwork("base"),
		evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
		evm.WithMaxAmountPerCall("1000000"),
	)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	client, err := x402http.NewClient(
		x402http.WithSigner(signer),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

// TestQuickstartExample4 - Load keys from different sources
func TestQuickstartExample4(t *testing.T) {
	// From mnemonic
	_, err := evm.NewSigner(
		evm.WithMnemonic("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about", 0),
		evm.WithNetwork("base"),
		evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
	)
	if err != nil {
		t.Fatalf("Failed to create signer from mnemonic: %v", err)
	}

	// From keystore file - API exists
	_, err = evm.NewSigner(
		evm.WithKeystore("/nonexistent/keystore.json", "password"),
		evm.WithNetwork("base"),
		evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
	)
	if err == nil {
		t.Fatal("Expected error for nonexistent keystore")
	}

	// Solana from keygen file - API exists
	_, err = svm.NewSigner(
		svm.WithKeygenFile("/nonexistent/id.json"),
		svm.WithNetwork("solana"),
		svm.WithToken("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", "USDC", 6),
	)
	if err == nil {
		t.Fatal("Expected error for nonexistent keygen file")
	}
}

// TestQuickstartExample5 - Token priority configuration
func TestQuickstartExample5(t *testing.T) {
	signer, err := evm.NewSigner(
		evm.WithPrivateKey("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		evm.WithNetwork("base"),
		evm.WithTokenPriority("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6, 1),
		evm.WithTokenPriority("0xdAC17F958D2ee523a2206206994597C13D831ec7", "USDT", 6, 2),
		evm.WithTokenPriority("0x6B175474E89094C44Da98b954EedeAC495271d0F", "DAI", 18, 3),
	)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	tokens := signer.GetTokens()
	if len(tokens) != 3 {
		t.Fatalf("Expected 3 tokens, got %d", len(tokens))
	}

	if tokens[0].Priority != 1 || tokens[1].Priority != 2 || tokens[2].Priority != 3 {
		t.Fatal("Token priorities not set correctly")
	}
}

// TestQuickstartExample6 - Custom HTTP client
func TestQuickstartExample6(t *testing.T) {
	signer, err := evm.NewSigner(
		evm.WithPrivateKey("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		evm.WithNetwork("base"),
		evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
	)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns: 100,
		},
	}

	client, err := x402http.NewClient(
		x402http.WithHTTPClient(httpClient),
		x402http.WithSigner(signer),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

// TestQuickstartExample7 - Error handling
func TestQuickstartExample7(t *testing.T) {
	signer, err := evm.NewSigner(
		evm.WithPrivateKey("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		evm.WithNetwork("base"),
		evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
	)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	_, err = x402http.NewClient(
		x402http.WithSigner(signer),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test error type checking compiles
	testErr := x402.NewPaymentError(x402.ErrCodeAmountExceeded, "test", x402.ErrAmountExceeded)
	var paymentErr *x402.PaymentError
	if !errors.As(testErr, &paymentErr) {
		t.Fatal("Error type checking should work")
	}

	if paymentErr.Code != x402.ErrCodeAmountExceeded {
		t.Fatalf("Expected ErrCodeAmountExceeded, got %s", paymentErr.Code)
	}
}

// TestQuickstartExample8 - Concurrent request handling
func TestQuickstartExample8(t *testing.T) {
	signer, err := evm.NewSigner(
		evm.WithPrivateKey("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		evm.WithNetwork("base"),
		evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
	)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	client, err := x402http.NewClient(
		x402http.WithSigner(signer),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test concurrent access doesn't panic
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.Client
		}()
	}
	wg.Wait()
}

// TestQuickstartExample9 - Custom payment selection
func TestQuickstartExample9(t *testing.T) {
	signer, err := evm.NewSigner(
		evm.WithPrivateKey("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
		evm.WithNetwork("base"),
		evm.WithToken("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", "USDC", 6),
	)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	selector := &customSelector{
		selectFunc: func(requirements *x402.PaymentRequirement, signers []x402.Signer) x402.Signer {
			if len(signers) > 0 {
				return signers[0]
			}
			return nil
		},
	}

	client, err := x402http.NewClient(
		x402http.WithSelector(selector),
		x402http.WithSigner(signer),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

type customSelector struct {
	selectFunc func(*x402.PaymentRequirement, []x402.Signer) x402.Signer
}

func (c *customSelector) SelectAndSign(requirements *x402.PaymentRequirement, signers []x402.Signer) (*x402.PaymentPayload, error) {
	signer := c.selectFunc(requirements, signers)
	if signer == nil {
		return nil, x402.ErrNoValidSigner
	}
	return signer.Sign(requirements)
}

// TestGetSettlement - Test GetSettlement API from quickstart
func TestGetSettlementAPI(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{},
	}

	settlement := x402http.GetSettlement(resp)
	if settlement != nil {
		t.Fatal("Expected nil settlement when no header present")
	}
}
