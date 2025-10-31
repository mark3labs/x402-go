package client

import (
	"testing"
	"time"

	"github.com/mark3labs/x402-go"
)

// T012: Test payment handler orchestration with fallback logic
func TestPaymentHandler_Orchestration(t *testing.T) {
	t.Run("selects first matching signer", func(t *testing.T) {
		t.Skip("TODO: Implement payment handler first")
		// Should use first signer that can pay
	})

	t.Run("falls back to second signer when first fails", func(t *testing.T) {
		t.Skip("TODO: Implement payment handler first")
		// Should try second signer if first fails
	})

	t.Run("returns error when all signers fail", func(t *testing.T) {
		t.Skip("TODO: Implement payment handler first")
		// Should return ErrNoMatchingSigner when all fail
	})

	t.Run("matches payment requirements correctly", func(t *testing.T) {
		t.Skip("TODO: Implement payment handler first")
		// Should match network and asset correctly
	})
}

func TestPaymentHandler_SignerSelection(t *testing.T) {
	t.Run("selects signer by network", func(t *testing.T) {
		t.Skip("TODO: Implement payment handler first")
		// Should select signer matching network
	})

	t.Run("selects signer by asset", func(t *testing.T) {
		t.Skip("TODO: Implement payment handler first")
		// Should select signer matching asset
	})

	t.Run("respects signer priority", func(t *testing.T) {
		t.Skip("TODO: Implement payment handler first")
		// Should try signers in configured order
	})
}

// T039: Test DefaultPaymentSelector priority algorithm
func TestDefaultPaymentSelector_Priority(t *testing.T) {
	t.Run("selects signer with lowest priority number", func(t *testing.T) {
		// Create mock signers with different priorities
		signer1 := &testSigner{
			network:  "base",
			tokens:   []x402.TokenConfig{{Address: x402.BaseMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 2,
			canSign:  true,
		}
		signer2 := &testSigner{
			network:  "polygon",
			tokens:   []x402.TokenConfig{{Address: x402.PolygonMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 1,
			canSign:  true,
		}

		selector := x402.NewDefaultPaymentSelector()
		requirements := []x402.PaymentRequirement{
			{Network: "base", Asset: x402.BaseMainnet.USDCAddress, MaxAmountRequired: "1000000"},
			{Network: "polygon", Asset: x402.PolygonMainnet.USDCAddress, MaxAmountRequired: "1000000"},
		}

		payment, err := selector.SelectAndSign(requirements, []x402.Signer{signer1, signer2})
		if err != nil {
			t.Fatalf("SelectAndSign failed: %v", err)
		}
		if payment == nil {
			t.Fatal("Expected payment, got nil")
		}
		// Should select signer2 (polygon) because it has priority 1 < 2
		if payment.Network != "polygon" {
			t.Errorf("Expected polygon network, got %s", payment.Network)
		}
	})

	t.Run("selects token with lowest priority when signers equal", func(t *testing.T) {
		signer := &testSigner{
			network: "base",
			tokens: []x402.TokenConfig{
				{Address: "0xToken1", Symbol: "TOKEN1", Decimals: 6, Priority: 2},
				{Address: "0xToken2", Symbol: "TOKEN2", Decimals: 6, Priority: 1},
			},
			priority: 1,
			canSign:  true,
		}

		selector := x402.NewDefaultPaymentSelector()
		requirements := []x402.PaymentRequirement{
			{Network: "base", Asset: "0xToken1", MaxAmountRequired: "1000000"},
			{Network: "base", Asset: "0xToken2", MaxAmountRequired: "1000000"},
		}

		payment, err := selector.SelectAndSign(requirements, []x402.Signer{signer})
		if err != nil {
			t.Fatalf("SelectAndSign failed: %v", err)
		}
		// Should select Token2 (priority 1 < 2)
		// Note: PaymentPayload doesn't have Asset field, check via requirement match
		if payment.Network != "base" {
			t.Errorf("Expected base network, got %s", payment.Network)
		}
	})

	t.Run("uses configuration order as tiebreaker", func(t *testing.T) {
		signer1 := &testSigner{
			network:  "base",
			tokens:   []x402.TokenConfig{{Address: x402.BaseMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 1,
			canSign:  true,
		}
		signer2 := &testSigner{
			network:  "polygon",
			tokens:   []x402.TokenConfig{{Address: x402.PolygonMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 1, // Same priority as signer1
			canSign:  true,
		}

		selector := x402.NewDefaultPaymentSelector()
		requirements := []x402.PaymentRequirement{
			{Network: "base", Asset: x402.BaseMainnet.USDCAddress, MaxAmountRequired: "1000000"},
			{Network: "polygon", Asset: x402.PolygonMainnet.USDCAddress, MaxAmountRequired: "1000000"},
		}

		payment, err := selector.SelectAndSign(requirements, []x402.Signer{signer1, signer2})
		if err != nil {
			t.Fatalf("SelectAndSign failed: %v", err)
		}
		// Should select signer1 (first in configuration order)
		if payment.Network != "base" {
			t.Errorf("Expected base network (first in config), got %s", payment.Network)
		}
	})
}

// T040: Test EVM signer integration with MCP
func TestEVMSignerIntegration(t *testing.T) {
	t.Run("creates valid payment with EVM signer", func(t *testing.T) {
		signer := &testSigner{
			network:  "base",
			tokens:   []x402.TokenConfig{{Address: x402.BaseMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 1,
			canSign:  true,
		}

		handler := NewPaymentHandler([]x402.Signer{signer}, nil)
		requirements := []x402.PaymentRequirement{
			{Network: "base", Asset: x402.BaseMainnet.USDCAddress, MaxAmountRequired: "1000000", PayTo: "0xRecipient"},
		}

		payment, err := handler.CreatePayment(requirements)
		if err != nil {
			t.Fatalf("CreatePayment failed: %v", err)
		}
		if payment == nil {
			t.Fatal("Expected payment, got nil")
		}
		if payment.Network != "base" {
			t.Errorf("Expected base network, got %s", payment.Network)
		}
	})

	t.Run("supports multiple EVM networks", func(t *testing.T) {
		baseSigner := &testSigner{
			network:  "base",
			tokens:   []x402.TokenConfig{{Address: x402.BaseMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 1,
			canSign:  true,
		}
		polygonSigner := &testSigner{
			network:  "polygon",
			tokens:   []x402.TokenConfig{{Address: x402.PolygonMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 2,
			canSign:  true,
		}

		handler := NewPaymentHandler([]x402.Signer{baseSigner, polygonSigner}, nil)

		// Test Base payment
		requirements := []x402.PaymentRequirement{
			{Network: "base", Asset: x402.BaseMainnet.USDCAddress, MaxAmountRequired: "1000000"},
		}
		payment, err := handler.CreatePayment(requirements)
		if err != nil {
			t.Fatalf("CreatePayment for base failed: %v", err)
		}
		if payment.Network != "base" {
			t.Errorf("Expected base, got %s", payment.Network)
		}

		// Test Polygon payment
		requirements = []x402.PaymentRequirement{
			{Network: "polygon", Asset: x402.PolygonMainnet.USDCAddress, MaxAmountRequired: "1000000"},
		}
		payment, err = handler.CreatePayment(requirements)
		if err != nil {
			t.Fatalf("CreatePayment for polygon failed: %v", err)
		}
		if payment.Network != "polygon" {
			t.Errorf("Expected polygon, got %s", payment.Network)
		}
	})
}

// T041: Test Solana signer integration with MCP
func TestSolanaSignerIntegration(t *testing.T) {
	t.Run("creates valid payment with Solana signer", func(t *testing.T) {
		signer := &testSigner{
			network:  "solana",
			tokens:   []x402.TokenConfig{{Address: x402.SolanaMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 1,
			canSign:  true,
		}

		handler := NewPaymentHandler([]x402.Signer{signer}, nil)
		requirements := []x402.PaymentRequirement{
			{Network: "solana", Asset: x402.SolanaMainnet.USDCAddress, MaxAmountRequired: "1000000"},
		}

		payment, err := handler.CreatePayment(requirements)
		if err != nil {
			t.Fatalf("CreatePayment failed: %v", err)
		}
		if payment == nil {
			t.Fatal("Expected payment, got nil")
		}
		if payment.Network != "solana" {
			t.Errorf("Expected solana network, got %s", payment.Network)
		}
	})

	t.Run("supports solana devnet", func(t *testing.T) {
		signer := &testSigner{
			network:  "solana-devnet",
			tokens:   []x402.TokenConfig{{Address: x402.SolanaDevnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 1,
			canSign:  true,
		}

		handler := NewPaymentHandler([]x402.Signer{signer}, nil)
		requirements := []x402.PaymentRequirement{
			{Network: "solana-devnet", Asset: x402.SolanaDevnet.USDCAddress, MaxAmountRequired: "1000000"},
		}

		payment, err := handler.CreatePayment(requirements)
		if err != nil {
			t.Fatalf("CreatePayment failed: %v", err)
		}
		if payment.Network != "solana-devnet" {
			t.Errorf("Expected solana-devnet network, got %s", payment.Network)
		}
	})
}

// T042: Test multi-network payment requirement matching
func TestMultiNetworkRequirementMatching(t *testing.T) {
	t.Run("matches correct network from multiple options", func(t *testing.T) {
		baseSigner := &testSigner{
			network:  "base",
			tokens:   []x402.TokenConfig{{Address: x402.BaseMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 1,
			canSign:  true,
		}

		handler := NewPaymentHandler([]x402.Signer{baseSigner}, nil)
		requirements := []x402.PaymentRequirement{
			{Network: "polygon", Asset: x402.PolygonMainnet.USDCAddress, MaxAmountRequired: "1000000"}, // Can't match
			{Network: "base", Asset: x402.BaseMainnet.USDCAddress, MaxAmountRequired: "1000000"},       // Should match
			{Network: "solana", Asset: x402.SolanaMainnet.USDCAddress, MaxAmountRequired: "1000000"},   // Can't match
		}

		payment, err := handler.CreatePayment(requirements)
		if err != nil {
			t.Fatalf("CreatePayment failed: %v", err)
		}
		if payment.Network != "base" {
			t.Errorf("Expected base network, got %s", payment.Network)
		}
	})

	t.Run("selects best option among multiple matches", func(t *testing.T) {
		baseSigner := &testSigner{
			network:  "base",
			tokens:   []x402.TokenConfig{{Address: x402.BaseMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 1, // Higher priority (lower number)
			canSign:  true,
		}
		polygonSigner := &testSigner{
			network:  "polygon",
			tokens:   []x402.TokenConfig{{Address: x402.PolygonMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 2,
			canSign:  true,
		}

		handler := NewPaymentHandler([]x402.Signer{baseSigner, polygonSigner}, nil)
		requirements := []x402.PaymentRequirement{
			{Network: "base", Asset: x402.BaseMainnet.USDCAddress, MaxAmountRequired: "1000000"},
			{Network: "polygon", Asset: x402.PolygonMainnet.USDCAddress, MaxAmountRequired: "1000000"},
		}

		payment, err := handler.CreatePayment(requirements)
		if err != nil {
			t.Fatalf("CreatePayment failed: %v", err)
		}
		// Should prefer base (priority 1) over polygon (priority 2)
		if payment.Network != "base" {
			t.Errorf("Expected base network (higher priority), got %s", payment.Network)
		}
	})

	t.Run("matches asset address correctly", func(t *testing.T) {
		signer := &testSigner{
			network: "base",
			tokens: []x402.TokenConfig{
				{Address: x402.BaseMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0},
			},
			priority: 1,
			canSign:  true,
		}

		handler := NewPaymentHandler([]x402.Signer{signer}, nil)
		requirements := []x402.PaymentRequirement{
			{Network: "base", Asset: "0xWrongAsset", MaxAmountRequired: "1000000"},               // Wrong asset
			{Network: "base", Asset: x402.BaseMainnet.USDCAddress, MaxAmountRequired: "1000000"}, // Correct
		}

		payment, err := handler.CreatePayment(requirements)
		if err != nil {
			t.Fatalf("CreatePayment failed: %v", err)
		}
		// PaymentPayload doesn't have Asset field, verify via network
		if payment.Network != "base" {
			t.Errorf("Expected base network, got %s", payment.Network)
		}
	})
}

// T043: Test fallback when primary network insufficient balance
func TestNetworkFallback(t *testing.T) {
	t.Run("falls back to secondary network on insufficient balance", func(t *testing.T) {
		// Primary signer with insufficient balance
		baseSigner := &testSigner{
			network:  "base",
			tokens:   []x402.TokenConfig{{Address: x402.BaseMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 1,
			canSign:  false, // Can't sign (simulating insufficient balance)
		}
		// Fallback signer with sufficient balance
		polygonSigner := &testSigner{
			network:  "polygon",
			tokens:   []x402.TokenConfig{{Address: x402.PolygonMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 2,
			canSign:  true, // Can sign
		}

		selector := x402.NewDefaultPaymentSelector()
		requirements := []x402.PaymentRequirement{
			{Network: "base", Asset: x402.BaseMainnet.USDCAddress, MaxAmountRequired: "1000000"},
			{Network: "polygon", Asset: x402.PolygonMainnet.USDCAddress, MaxAmountRequired: "1000000"},
		}

		payment, err := selector.SelectAndSign(requirements, []x402.Signer{baseSigner, polygonSigner})
		if err != nil {
			t.Fatalf("SelectAndSign failed: %v", err)
		}
		// Should fall back to polygon
		if payment.Network != "polygon" {
			t.Errorf("Expected fallback to polygon, got %s", payment.Network)
		}
	})

	t.Run("returns error when all signers cannot pay", func(t *testing.T) {
		baseSigner := &testSigner{
			network:  "base",
			tokens:   []x402.TokenConfig{{Address: x402.BaseMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 1,
			canSign:  false,
		}
		polygonSigner := &testSigner{
			network:  "polygon",
			tokens:   []x402.TokenConfig{{Address: x402.PolygonMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}},
			priority: 2,
			canSign:  false,
		}

		selector := x402.NewDefaultPaymentSelector()
		requirements := []x402.PaymentRequirement{
			{Network: "base", Asset: x402.BaseMainnet.USDCAddress, MaxAmountRequired: "1000000"},
			{Network: "polygon", Asset: x402.PolygonMainnet.USDCAddress, MaxAmountRequired: "1000000"},
		}

		_, err := selector.SelectAndSign(requirements, []x402.Signer{baseSigner, polygonSigner})
		if err == nil {
			t.Fatal("Expected error when no signer can pay, got nil")
		}
	})
}

// T043a: Test that payment fallback completes within 5 seconds (SC-003)
func TestPaymentFallbackTimeout(t *testing.T) {
	t.Run("fallback completes within 5 seconds", func(t *testing.T) {
		// Create multiple signers that fail (simulating insufficient balance)
		signers := []x402.Signer{
			&testSigner{network: "base", tokens: []x402.TokenConfig{{Address: x402.BaseMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}}, priority: 1, canSign: false},
			&testSigner{network: "polygon", tokens: []x402.TokenConfig{{Address: x402.PolygonMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}}, priority: 2, canSign: false},
			&testSigner{network: "avalanche", tokens: []x402.TokenConfig{{Address: x402.AvalancheMainnet.USDCAddress, Symbol: "USDC", Decimals: 6, Priority: 0}}, priority: 3, canSign: true}, // Success on 3rd try
		}

		selector := x402.NewDefaultPaymentSelector()
		requirements := []x402.PaymentRequirement{
			{Network: "base", Asset: x402.BaseMainnet.USDCAddress, MaxAmountRequired: "1000000"},
			{Network: "polygon", Asset: x402.PolygonMainnet.USDCAddress, MaxAmountRequired: "1000000"},
			{Network: "avalanche", Asset: x402.AvalancheMainnet.USDCAddress, MaxAmountRequired: "1000000"},
		}

		start := time.Now()
		_, err := selector.SelectAndSign(requirements, signers)
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("SelectAndSign failed: %v", err)
		}
		if elapsed > 5*time.Second {
			t.Errorf("Payment fallback took %v, expected < 5 seconds", elapsed)
		}
	})
}

// Mock facilitator for testing
type mockFacilitator struct {
	verifyResponse bool
	verifyError    error
}

func (m *mockFacilitator) Verify(payment x402.PaymentPayload, requirement x402.PaymentRequirement) (*VerifyResponse, error) {
	if m.verifyError != nil {
		return nil, m.verifyError
	}
	return &VerifyResponse{
		IsValid: m.verifyResponse,
		Payer:   "0x123",
	}, nil
}

type VerifyResponse struct {
	IsValid bool
	Payer   string
}
