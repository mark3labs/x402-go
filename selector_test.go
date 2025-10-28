package x402

import (
	"math/big"
	"testing"
)

// mockSignerForSelector implements Signer for selector testing
type mockSignerForSelector struct {
	network      string
	scheme       string
	tokens       []TokenConfig
	priority     int
	maxAmount    *big.Int
	canSignValue bool
	signError    error
	signCalled   bool
}

func (m *mockSignerForSelector) Network() string { return m.network }
func (m *mockSignerForSelector) Scheme() string  { return m.scheme }
func (m *mockSignerForSelector) GetPriority() int {
	if m.priority == 0 {
		return 0
	}
	return m.priority
}
func (m *mockSignerForSelector) GetTokens() []TokenConfig { return m.tokens }
func (m *mockSignerForSelector) GetMaxAmount() *big.Int   { return m.maxAmount }
func (m *mockSignerForSelector) CanSign(req *PaymentRequirement) bool {
	// Check network match
	if m.network != req.Network {
		return false
	}
	// Check if we have the requested token
	for _, token := range m.tokens {
		if token.Address == req.Asset {
			return m.canSignValue
		}
	}
	return false
}

func (m *mockSignerForSelector) Sign(req *PaymentRequirement) (*PaymentPayload, error) {
	m.signCalled = true
	if m.signError != nil {
		return nil, m.signError
	}
	return &PaymentPayload{
		X402Version: 1,
		Scheme:      m.scheme,
		Network:     m.network,
		Payload:     map[string]interface{}{"mock": "payment"},
	}, nil
}

func TestDefaultPaymentSelector_SelectAndSign_NoSigners(t *testing.T) {
	selector := NewDefaultPaymentSelector()
	requirements := &PaymentRequirement{
		Scheme:            "exact",
		Network:           "base",
		MaxAmountRequired: "1000000",
		Asset:             "0xUSDC",
	}

	_, err := selector.SelectAndSign(requirements, []Signer{})
	if err == nil {
		t.Fatal("expected error with no signers, got nil")
	}

	paymentErr, ok := err.(*PaymentError)
	if !ok {
		t.Fatalf("expected PaymentError, got %T", err)
	}
	if paymentErr.Code != ErrCodeNoValidSigner {
		t.Errorf("expected error code %s, got %s", ErrCodeNoValidSigner, paymentErr.Code)
	}
}

func TestDefaultPaymentSelector_SelectAndSign_InvalidAmount(t *testing.T) {
	selector := NewDefaultPaymentSelector()
	signer := &mockSignerForSelector{
		network:      "base",
		scheme:       "exact",
		canSignValue: true,
		tokens: []TokenConfig{
			{Address: "0xUSDC", Symbol: "USDC", Decimals: 6},
		},
	}

	requirements := &PaymentRequirement{
		Scheme:            "exact",
		Network:           "base",
		MaxAmountRequired: "invalid",
		Asset:             "0xUSDC",
	}

	_, err := selector.SelectAndSign(requirements, []Signer{signer})
	if err == nil {
		t.Fatal("expected error with invalid amount, got nil")
	}

	paymentErr, ok := err.(*PaymentError)
	if !ok {
		t.Fatalf("expected PaymentError, got %T", err)
	}
	if paymentErr.Code != ErrCodeInvalidRequirements {
		t.Errorf("expected error code %s, got %s", ErrCodeInvalidRequirements, paymentErr.Code)
	}
}

func TestDefaultPaymentSelector_SelectAndSign_SignerPriority(t *testing.T) {
	tests := []struct {
		name             string
		signers          []Signer
		requirements     *PaymentRequirement
		expectedPriority int // which signer should be selected (by priority)
	}{
		{
			name: "select lower priority number (higher priority)",
			signers: []Signer{
				&mockSignerForSelector{
					network:      "base",
					scheme:       "exact",
					priority:     2,
					canSignValue: true,
					tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
				},
				&mockSignerForSelector{
					network:      "base",
					scheme:       "exact",
					priority:     1,
					canSignValue: true,
					tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
				},
				&mockSignerForSelector{
					network:      "base",
					scheme:       "exact",
					priority:     3,
					canSignValue: true,
					tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
				},
			},
			requirements: &PaymentRequirement{
				Network:           "base",
				Asset:             "0xUSDC",
				MaxAmountRequired: "1000000",
			},
			expectedPriority: 1,
		},
		{
			name: "default priority (0) is higher than 1",
			signers: []Signer{
				&mockSignerForSelector{
					network:      "base",
					scheme:       "exact",
					priority:     1,
					canSignValue: true,
					tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
				},
				&mockSignerForSelector{
					network:      "base",
					scheme:       "exact",
					priority:     0,
					canSignValue: true,
					tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
				},
			},
			requirements: &PaymentRequirement{
				Network:           "base",
				Asset:             "0xUSDC",
				MaxAmountRequired: "1000000",
			},
			expectedPriority: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewDefaultPaymentSelector()
			payment, err := selector.SelectAndSign(tt.requirements, tt.signers)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if payment == nil {
				t.Fatal("expected payment, got nil")
			}

			// Verify the correct signer was used
			var selectedSigner *mockSignerForSelector
			for _, s := range tt.signers {
				mock := s.(*mockSignerForSelector)
				if mock.signCalled {
					selectedSigner = mock
					break
				}
			}
			if selectedSigner == nil {
				t.Fatal("no signer was called")
			}
			if selectedSigner.priority != tt.expectedPriority {
				t.Errorf("expected signer with priority %d, got %d", tt.expectedPriority, selectedSigner.priority)
			}
		})
	}
}

func TestDefaultPaymentSelector_SelectAndSign_TokenPriority(t *testing.T) {
	tests := []struct {
		name          string
		signers       []Signer
		requirements  *PaymentRequirement
		expectedToken string
	}{
		{
			name: "select token with lower priority number within same signer priority",
			signers: []Signer{
				&mockSignerForSelector{
					network:      "base",
					scheme:       "exact",
					priority:     1,
					canSignValue: true,
					tokens: []TokenConfig{
						{Address: "0xUSDT", Symbol: "USDT", Decimals: 6, Priority: 2},
					},
				},
				&mockSignerForSelector{
					network:      "base",
					scheme:       "exact",
					priority:     1, // Same signer priority
					canSignValue: true,
					tokens: []TokenConfig{
						{Address: "0xUSDC", Symbol: "USDC", Decimals: 6, Priority: 1},
					},
				},
			},
			requirements: &PaymentRequirement{
				Network:           "base",
				Asset:             "0xUSDC",
				MaxAmountRequired: "1000000",
			},
			expectedToken: "0xUSDC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewDefaultPaymentSelector()
			payment, err := selector.SelectAndSign(tt.requirements, tt.signers)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if payment == nil {
				t.Fatal("expected payment, got nil")
			}

			// Verify correct signer was selected based on token priority
			var selectedSigner *mockSignerForSelector
			for _, s := range tt.signers {
				mock := s.(*mockSignerForSelector)
				if mock.signCalled {
					selectedSigner = mock
					break
				}
			}
			if selectedSigner == nil {
				t.Fatal("no signer was called")
			}

			// Verify the selected signer has the expected token
			hasToken := false
			for _, token := range selectedSigner.tokens {
				if token.Address == tt.expectedToken {
					hasToken = true
					break
				}
			}
			if !hasToken {
				t.Errorf("selected signer does not have token %s", tt.expectedToken)
			}
		})
	}
}

func TestDefaultPaymentSelector_SelectAndSign_MaxAmountFiltering(t *testing.T) {
	tests := []struct {
		name             string
		signers          []Signer
		requirements     *PaymentRequirement
		expectError      bool
		expectedPriority int // which signer should be selected (if any)
	}{
		{
			name: "skip signer with insufficient max amount",
			signers: []Signer{
				&mockSignerForSelector{
					network:      "base",
					scheme:       "exact",
					priority:     1,
					maxAmount:    big.NewInt(500000), // 0.5 USDC
					canSignValue: true,
					tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
				},
				&mockSignerForSelector{
					network:      "base",
					scheme:       "exact",
					priority:     2,
					maxAmount:    big.NewInt(2000000), // 2 USDC (sufficient)
					canSignValue: true,
					tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
				},
			},
			requirements: &PaymentRequirement{
				Network:           "base",
				Asset:             "0xUSDC",
				MaxAmountRequired: "1000000", // 1 USDC
			},
			expectError:      false,
			expectedPriority: 2, // Should use second signer
		},
		{
			name: "all signers exceed max amount",
			signers: []Signer{
				&mockSignerForSelector{
					network:      "base",
					scheme:       "exact",
					priority:     1,
					maxAmount:    big.NewInt(500000),
					canSignValue: true,
					tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
				},
				&mockSignerForSelector{
					network:      "base",
					scheme:       "exact",
					priority:     2,
					maxAmount:    big.NewInt(800000),
					canSignValue: true,
					tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
				},
			},
			requirements: &PaymentRequirement{
				Network:           "base",
				Asset:             "0xUSDC",
				MaxAmountRequired: "1000000",
			},
			expectError: true,
		},
		{
			name: "signer with nil max amount (no limit) is used",
			signers: []Signer{
				&mockSignerForSelector{
					network:      "base",
					scheme:       "exact",
					priority:     1,
					maxAmount:    nil, // No limit
					canSignValue: true,
					tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
				},
			},
			requirements: &PaymentRequirement{
				Network:           "base",
				Asset:             "0xUSDC",
				MaxAmountRequired: "999999999999", // Very large amount
			},
			expectError:      false,
			expectedPriority: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewDefaultPaymentSelector()
			payment, err := selector.SelectAndSign(tt.requirements, tt.signers)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				paymentErr, ok := err.(*PaymentError)
				if !ok {
					t.Fatalf("expected PaymentError, got %T", err)
				}
				if paymentErr.Code != ErrCodeNoValidSigner {
					t.Errorf("expected error code %s, got %s", ErrCodeNoValidSigner, paymentErr.Code)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if payment == nil {
				t.Fatal("expected payment, got nil")
			}

			// Verify the correct signer was used
			var selectedSigner *mockSignerForSelector
			for _, s := range tt.signers {
				mock := s.(*mockSignerForSelector)
				if mock.signCalled {
					selectedSigner = mock
					break
				}
			}
			if selectedSigner == nil {
				t.Fatal("no signer was called")
			}
			if selectedSigner.priority != tt.expectedPriority {
				t.Errorf("expected signer with priority %d, got %d", tt.expectedPriority, selectedSigner.priority)
			}
		})
	}
}

func TestDefaultPaymentSelector_SelectAndSign_NetworkFiltering(t *testing.T) {
	signers := []Signer{
		&mockSignerForSelector{
			network:      "base",
			scheme:       "exact",
			priority:     1,
			canSignValue: true,
			tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
		},
		&mockSignerForSelector{
			network:      "solana",
			scheme:       "exact",
			priority:     2,
			canSignValue: true,
			tokens:       []TokenConfig{{Address: "USDC_MINT", Symbol: "USDC", Decimals: 6}},
		},
	}

	tests := []struct {
		name            string
		requirements    *PaymentRequirement
		expectedNetwork string
	}{
		{
			name: "select base network signer",
			requirements: &PaymentRequirement{
				Network:           "base",
				Asset:             "0xUSDC",
				MaxAmountRequired: "1000000",
			},
			expectedNetwork: "base",
		},
		{
			name: "select solana network signer",
			requirements: &PaymentRequirement{
				Network:           "solana",
				Asset:             "USDC_MINT",
				MaxAmountRequired: "1000000",
			},
			expectedNetwork: "solana",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewDefaultPaymentSelector()
			payment, err := selector.SelectAndSign(tt.requirements, signers)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if payment.Network != tt.expectedNetwork {
				t.Errorf("expected network %s, got %s", tt.expectedNetwork, payment.Network)
			}
		})
	}
}

func TestDefaultPaymentSelector_SelectAndSign_NoMatchingNetwork(t *testing.T) {
	signers := []Signer{
		&mockSignerForSelector{
			network:      "base",
			scheme:       "exact",
			priority:     1,
			canSignValue: true,
			tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
		},
	}

	requirements := &PaymentRequirement{
		Network:           "ethereum",
		Asset:             "0xUSDC",
		MaxAmountRequired: "1000000",
	}

	selector := NewDefaultPaymentSelector()
	_, err := selector.SelectAndSign(requirements, signers)
	if err == nil {
		t.Fatal("expected error with no matching network, got nil")
	}

	paymentErr, ok := err.(*PaymentError)
	if !ok {
		t.Fatalf("expected PaymentError, got %T", err)
	}
	if paymentErr.Code != ErrCodeNoValidSigner {
		t.Errorf("expected error code %s, got %s", ErrCodeNoValidSigner, paymentErr.Code)
	}
}

func TestDefaultPaymentSelector_SelectAndSign_NoMatchingToken(t *testing.T) {
	signers := []Signer{
		&mockSignerForSelector{
			network:      "base",
			scheme:       "exact",
			priority:     1,
			canSignValue: true,
			tokens:       []TokenConfig{{Address: "0xUSDT", Symbol: "USDT", Decimals: 6}},
		},
	}

	requirements := &PaymentRequirement{
		Network:           "base",
		Asset:             "0xUSDC", // Different token
		MaxAmountRequired: "1000000",
	}

	selector := NewDefaultPaymentSelector()
	_, err := selector.SelectAndSign(requirements, signers)
	if err == nil {
		t.Fatal("expected error with no matching token, got nil")
	}

	paymentErr, ok := err.(*PaymentError)
	if !ok {
		t.Fatalf("expected PaymentError, got %T", err)
	}
	if paymentErr.Code != ErrCodeNoValidSigner {
		t.Errorf("expected error code %s, got %s", ErrCodeNoValidSigner, paymentErr.Code)
	}
}

func TestDefaultPaymentSelector_SelectAndSign_SigningError(t *testing.T) {
	signers := []Signer{
		&mockSignerForSelector{
			network:      "base",
			scheme:       "exact",
			priority:     1,
			canSignValue: true,
			signError:    ErrSigningFailed,
			tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
		},
	}

	requirements := &PaymentRequirement{
		Network:           "base",
		Asset:             "0xUSDC",
		MaxAmountRequired: "1000000",
	}

	selector := NewDefaultPaymentSelector()
	_, err := selector.SelectAndSign(requirements, signers)
	if err == nil {
		t.Fatal("expected signing error, got nil")
	}

	paymentErr, ok := err.(*PaymentError)
	if !ok {
		t.Fatalf("expected PaymentError, got %T", err)
	}
	if paymentErr.Code != ErrCodeSigningFailed {
		t.Errorf("expected error code %s, got %s", ErrCodeSigningFailed, paymentErr.Code)
	}
}

// T063 [P]: Benchmark for signer selection with 10 signers (SC-006: <100ms)
func BenchmarkDefaultPaymentSelector_SelectAndSign_10Signers(b *testing.B) {
	// Create 10 signers with different priorities
	signers := make([]Signer, 10)
	for i := 0; i < 10; i++ {
		signers[i] = &mockSignerForSelector{
			network:      "base",
			scheme:       "exact",
			priority:     i + 1,
			canSignValue: true,
			tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
		}
	}

	requirements := &PaymentRequirement{
		Network:           "base",
		Asset:             "0xUSDC",
		MaxAmountRequired: "1000000",
	}

	selector := NewDefaultPaymentSelector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := selector.SelectAndSign(requirements, signers)
		if err != nil {
			b.Fatalf("SelectAndSign failed: %v", err)
		}
	}
}

// T067 [P]: Test for priority ordering convention (1 > 2 > 3)
func TestDefaultPaymentSelector_PriorityOrderingConvention(t *testing.T) {
	tests := []struct {
		name             string
		signerPriorities []int
		expectedPriority int // lowest number = highest priority
	}{
		{
			name:             "priority 1 is highest (1 > 2 > 3)",
			signerPriorities: []int{3, 1, 2},
			expectedPriority: 1,
		},
		{
			name:             "priority 0 (default) is highest priority",
			signerPriorities: []int{1, 2, 0, 3},
			expectedPriority: 0,
		},
		{
			name:             "lower number always wins",
			signerPriorities: []int{10, 5, 1, 3},
			expectedPriority: 1,
		},
		{
			name:             "single digit priorities sorted correctly",
			signerPriorities: []int{9, 8, 7, 6, 5, 4, 3, 2, 1},
			expectedPriority: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signers := make([]Signer, len(tt.signerPriorities))
			for i, priority := range tt.signerPriorities {
				signers[i] = &mockSignerForSelector{
					network:      "base",
					scheme:       "exact",
					priority:     priority,
					canSignValue: true,
					tokens:       []TokenConfig{{Address: "0xUSDC", Symbol: "USDC", Decimals: 6}},
				}
			}

			requirements := &PaymentRequirement{
				Network:           "base",
				Asset:             "0xUSDC",
				MaxAmountRequired: "1000000",
			}

			selector := NewDefaultPaymentSelector()
			_, err := selector.SelectAndSign(requirements, signers)
			if err != nil {
				t.Fatalf("SelectAndSign failed: %v", err)
			}

			// Find which signer was called
			var selectedPriority int
			for _, s := range signers {
				mock := s.(*mockSignerForSelector)
				if mock.signCalled {
					selectedPriority = mock.priority
					break
				}
			}

			if selectedPriority != tt.expectedPriority {
				t.Errorf("expected priority %d to be selected, got %d", tt.expectedPriority, selectedPriority)
			}
		})
	}
}
