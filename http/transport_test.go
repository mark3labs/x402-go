package http

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/x402-go"
)

// Helper function to create a proper PaymentRequirementsResponse as per x402 spec
func makePaymentRequirementsResponse(req x402.PaymentRequirement) []byte {
	response := struct {
		X402Version int                       `json:"x402Version"`
		Error       string                    `json:"error"`
		Accepts     []x402.PaymentRequirement `json:"accepts"`
	}{
		X402Version: 1,
		Error:       "Payment required",
		Accepts:     []x402.PaymentRequirement{req},
	}
	body, _ := json.Marshal(response)
	return body
}

func TestRoundTrip_NonPaymentRequest(t *testing.T) {
	// Server returns 200 OK without requiring payment
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	transport := &X402Transport{
		Base: http.DefaultTransport,
		Signers: []x402.Signer{
			&mockSigner{network: "base", scheme: "exact", canSignValue: true},
		},
		Selector: x402.NewDefaultPaymentSelector(),
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRoundTrip_PaymentRequired(t *testing.T) {
	// Track number of requests
	requestCount := 0

	// Server returns 402 on first request, 200 on retry with payment
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if r.Header.Get("X-PAYMENT") == "" {
			// First request without payment
			requirements := x402.PaymentRequirement{
				Scheme:            "exact",
				Network:           "base",
				Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				MaxAmountRequired: "100000",
				PayTo:             "0x1234567890123456789012345678901234567890",
				MaxTimeoutSeconds: 60,
			}

			body := makePaymentRequirementsResponse(requirements)
			w.WriteHeader(http.StatusPaymentRequired)
			w.Write(body)
		} else {
			// Retry with payment
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	transport := &X402Transport{
		Base: http.DefaultTransport,
		Signers: []x402.Signer{
			&mockSigner{network: "base", scheme: "exact", canSignValue: true},
		},
		Selector: x402.NewDefaultPaymentSelector(),
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	if requestCount != 2 {
		t.Errorf("expected 2 requests, got %d", requestCount)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestRoundTrip_NoValidSigner(t *testing.T) {
	// Server returns 402 requiring payment
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requirements := x402.PaymentRequirement{
			Scheme:            "exact",
			Network:           "ethereum", // Different network
			Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
			MaxAmountRequired: "100000",
			PayTo:             "0x1234567890123456789012345678901234567890",
			MaxTimeoutSeconds: 60,
		}

		body := makePaymentRequirementsResponse(requirements)
		w.WriteHeader(http.StatusPaymentRequired)
		w.Write(body)
	}))
	defer server.Close()

	transport := &X402Transport{
		Base: http.DefaultTransport,
		Signers: []x402.Signer{
			&mockSigner{network: "base", scheme: "exact", canSignValue: false}, // Can't sign
		},
		Selector: x402.NewDefaultPaymentSelector(),
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	_, err := transport.RoundTrip(req)

	if err == nil {
		t.Fatal("expected error for no valid signer")
	}

	// Should be a payment error
	var paymentErr *x402.PaymentError
	if !errors.As(err, &paymentErr) {
		t.Errorf("expected PaymentError, got %T", err)
	}
}

func TestParsePaymentRequirements(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name: "valid requirements",
			body: `{
				"x402Version": 1,
				"error": "Payment required",
				"accepts": [{
					"scheme": "exact",
					"network": "base",
					"asset": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
					"maxAmountRequired": "100000",
					"payTo": "0x1234567890123456789012345678901234567890",
					"maxTimeoutSeconds": 60
				}]
			}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			body:    "not json",
			wantErr: true,
		},
		{
			name:    "empty body",
			body:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Body: io.NopCloser(strings.NewReader(tt.body)),
			}

			requirements, err := parsePaymentRequirements(resp)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if requirements == nil {
				t.Error("expected non-nil requirements")
			}
		})
	}
}

func TestBuildPaymentHeader(t *testing.T) {
	payment := &x402.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "base",
		Payload: map[string]interface{}{
			"test": "data",
		},
	}

	header, err := buildPaymentHeader(payment)
	if err != nil {
		t.Fatalf("buildPaymentHeader failed: %v", err)
	}

	// Verify it's valid base64
	decoded, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		t.Fatalf("header is not valid base64: %v", err)
	}

	// Verify it's valid JSON
	var parsed x402.PaymentPayload
	if err := json.Unmarshal(decoded, &parsed); err != nil {
		t.Fatalf("decoded header is not valid JSON: %v", err)
	}

	// Verify content
	if parsed.X402Version != 1 {
		t.Errorf("expected version 1, got %d", parsed.X402Version)
	}
	if parsed.Scheme != "exact" {
		t.Errorf("expected scheme 'exact', got '%s'", parsed.Scheme)
	}
	if parsed.Network != "base" {
		t.Errorf("expected network 'base', got '%s'", parsed.Network)
	}
}

func TestParseSettlement(t *testing.T) {
	tests := []struct {
		name       string
		headerFunc func() string
		wantErr    bool
	}{
		{
			name: "valid settlement",
			headerFunc: func() string {
				settlement := x402.SettlementResponse{
					Success:     true,
					Transaction: "0x1234567890abcdef",
					Network:     "base",
					Payer:       "0x1234567890",
				}
				data, _ := json.Marshal(settlement)
				return base64.StdEncoding.EncodeToString(data)
			},
			wantErr: false,
		},
		{
			name: "invalid base64",
			headerFunc: func() string {
				return "invalid base64!!!"
			},
			wantErr: true,
		},
		{
			name: "invalid JSON",
			headerFunc: func() string {
				return base64.StdEncoding.EncodeToString([]byte("not json"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := tt.headerFunc()
			settlement, err := parseSettlement(header)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if settlement == nil {
				t.Error("expected non-nil settlement")
			}
		})
	}
}

func TestRoundTrip_WithSettlement(t *testing.T) {
	// Server returns settlement header on successful payment
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-PAYMENT") != "" {
			// Add settlement header
			settlement := x402.SettlementResponse{
				Success:     true,
				Transaction: "0xabcdef1234567890",
				Network:     "base",
				Payer:       "0x1234567890",
			}
			data, _ := json.Marshal(settlement)
			settlementHeader := base64.StdEncoding.EncodeToString(data)

			w.Header().Set("X-SETTLEMENT", settlementHeader)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		} else {
			requirements := x402.PaymentRequirement{
				Scheme:            "exact",
				Network:           "base",
				Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				MaxAmountRequired: "100000",
				PayTo:             "0x1234567890123456789012345678901234567890",
				MaxTimeoutSeconds: 60,
			}
			body := makePaymentRequirementsResponse(requirements)
			w.WriteHeader(http.StatusPaymentRequired)
			w.Write(body)
		}
	}))
	defer server.Close()

	transport := &X402Transport{
		Base: http.DefaultTransport,
		Signers: []x402.Signer{
			&mockSigner{network: "base", scheme: "exact", canSignValue: true},
		},
		Selector: x402.NewDefaultPaymentSelector(),
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	// Check settlement header
	settlement := GetSettlement(resp)
	if settlement == nil {
		t.Fatal("expected settlement header")
	}

	if settlement.Transaction != "0xabcdef1234567890" {
		t.Errorf("expected transaction 0xabcdef1234567890, got %s", settlement.Transaction)
	}

	if !settlement.Success {
		t.Error("expected success to be true")
	}
}

func TestRoundTrip_MultiSignerSelection_Priority(t *testing.T) {
	// Track which signer was used
	var selectedSignerPriority int

	// Server returns 402 requiring base network
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-PAYMENT") == "" {
			requirements := x402.PaymentRequirement{
				Scheme:            "exact",
				Network:           "base",
				Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				MaxAmountRequired: "100000",
				PayTo:             "0x1234567890123456789012345678901234567890",
				MaxTimeoutSeconds: 60,
			}
			body := makePaymentRequirementsResponse(requirements)
			w.WriteHeader(http.StatusPaymentRequired)
			w.Write(body)
		} else {
			// Parse payment to determine which signer was used
			paymentHeader := r.Header.Get("X-PAYMENT")
			decoded, _ := base64.StdEncoding.DecodeString(paymentHeader)
			var payment x402.PaymentPayload
			json.Unmarshal(decoded, &payment)

			// Mock payload includes priority for tracking
			if payloadMap, ok := payment.Payload.(map[string]interface{}); ok {
				if priority, ok := payloadMap["priority"].(float64); ok {
					selectedSignerPriority = int(priority)
				}
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	// Create multiple signers with different priorities
	// Lower number = higher priority
	transport := &X402Transport{
		Base: http.DefaultTransport,
		Signers: []x402.Signer{
			&mockSigner{
				network:      "base",
				scheme:       "exact",
				canSignValue: true,
				priority:     3, // Lowest priority
			},
			&mockSigner{
				network:      "base",
				scheme:       "exact",
				canSignValue: true,
				priority:     1, // Highest priority - should be selected
			},
			&mockSigner{
				network:      "base",
				scheme:       "exact",
				canSignValue: true,
				priority:     2, // Middle priority
			},
		},
		Selector: x402.NewDefaultPaymentSelector(),
	}

	// Modify mock signers to include priority in payload
	for i, s := range transport.Signers {
		mock := s.(*mockSigner)
		priority := mock.priority
		mock.signError = nil
		// Override Sign to include priority
		originalSigner := mock
		transport.Signers[i] = &mockSignerWithPayload{
			mockSigner: originalSigner,
			priority:   priority,
		}
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Verify the highest priority signer (priority 1) was selected
	if selectedSignerPriority != 1 {
		t.Errorf("expected signer with priority 1 to be selected, got priority %d", selectedSignerPriority)
	}
}

// mockSignerWithPayload extends mockSigner to include priority in payload
type mockSignerWithPayload struct {
	*mockSigner
	priority int
}

// mockSignerForNetworkTest is a mock signer that properly checks network matching
type mockSignerForNetworkTest struct {
	network  string
	scheme   string
	priority int
}

func (m *mockSignerForNetworkTest) Network() string               { return m.network }
func (m *mockSignerForNetworkTest) Scheme() string                { return m.scheme }
func (m *mockSignerForNetworkTest) GetPriority() int              { return m.priority }
func (m *mockSignerForNetworkTest) GetTokens() []x402.TokenConfig { return nil }
func (m *mockSignerForNetworkTest) GetMaxAmount() *big.Int        { return nil }
func (m *mockSignerForNetworkTest) CanSign(req *x402.PaymentRequirement) bool {
	return m.network == req.Network
}
func (m *mockSignerForNetworkTest) Sign(req *x402.PaymentRequirement) (*x402.PaymentPayload, error) {
	return &x402.PaymentPayload{
		X402Version: 1,
		Scheme:      m.scheme,
		Network:     m.network,
		Payload: map[string]interface{}{
			"network": m.network,
		},
	}, nil
}

func (m *mockSignerWithPayload) Sign(req *x402.PaymentRequirement) (*x402.PaymentPayload, error) {
	if m.signError != nil {
		return nil, m.signError
	}
	return &x402.PaymentPayload{
		X402Version: 1,
		Scheme:      m.scheme,
		Network:     m.network,
		Payload: map[string]interface{}{
			"priority": m.priority,
		},
	}, nil
}

func TestRoundTrip_MultiSignerSelection_NetworkMatch(t *testing.T) {
	tests := []struct {
		name            string
		requiredNetwork string
		expectedNetwork string
	}{
		{
			name:            "select base network signer",
			requiredNetwork: "base",
			expectedNetwork: "base",
		},
		{
			name:            "select solana network signer",
			requiredNetwork: "solana",
			expectedNetwork: "solana",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var selectedNetwork string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-PAYMENT") == "" {
					requirements := x402.PaymentRequirement{
						Scheme:            "exact",
						Network:           tt.requiredNetwork,
						Asset:             "0xUSDC",
						MaxAmountRequired: "100000",
						PayTo:             "0x1234567890123456789012345678901234567890",
						MaxTimeoutSeconds: 60,
					}
					body := makePaymentRequirementsResponse(requirements)
					w.WriteHeader(http.StatusPaymentRequired)
					w.Write(body)
				} else {
					paymentHeader := r.Header.Get("X-PAYMENT")
					decoded, _ := base64.StdEncoding.DecodeString(paymentHeader)
					var payment x402.PaymentPayload
					json.Unmarshal(decoded, &payment)
					selectedNetwork = payment.Network

					w.WriteHeader(http.StatusOK)
					w.Write([]byte("success"))
				}
			}))
			defer server.Close()

			// Create mock signers with proper CanSign implementation
			baseSigner := &mockSignerForNetworkTest{
				network:  "base",
				scheme:   "exact",
				priority: 1,
			}
			solanaSigner := &mockSignerForNetworkTest{
				network:  "solana",
				scheme:   "exact",
				priority: 2,
			}

			transport := &X402Transport{
				Base: http.DefaultTransport,
				Signers: []x402.Signer{
					baseSigner,
					solanaSigner,
				},
				Selector: x402.NewDefaultPaymentSelector(),
			}

			req, _ := http.NewRequest("GET", server.URL, nil)
			resp, err := transport.RoundTrip(req)
			if err != nil {
				t.Fatalf("RoundTrip failed: %v", err)
			}
			defer resp.Body.Close()

			if selectedNetwork != tt.expectedNetwork {
				t.Errorf("expected network %s, got %s", tt.expectedNetwork, selectedNetwork)
			}
		})
	}
}

func TestRoundTrip_MultiSignerSelection_MaxAmountFiltering(t *testing.T) {
	var selectedSignerPriority int

	// Server requires 1 USDC
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-PAYMENT") == "" {
			requirements := x402.PaymentRequirement{
				Scheme:            "exact",
				Network:           "base",
				Asset:             "0xUSDC",
				MaxAmountRequired: "1000000", // 1 USDC
				PayTo:             "0x1234567890123456789012345678901234567890",
				MaxTimeoutSeconds: 60,
			}
			body := makePaymentRequirementsResponse(requirements)
			w.WriteHeader(http.StatusPaymentRequired)
			w.Write(body)
		} else {
			paymentHeader := r.Header.Get("X-PAYMENT")
			decoded, _ := base64.StdEncoding.DecodeString(paymentHeader)
			var payment x402.PaymentPayload
			json.Unmarshal(decoded, &payment)

			if payloadMap, ok := payment.Payload.(map[string]interface{}); ok {
				if priority, ok := payloadMap["priority"].(float64); ok {
					selectedSignerPriority = int(priority)
				}
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	transport := &X402Transport{
		Base: http.DefaultTransport,
		Signers: []x402.Signer{
			&mockSignerWithPayload{
				mockSigner: &mockSigner{
					network:      "base",
					scheme:       "exact",
					canSignValue: true,
					priority:     1,
					maxAmount:    big.NewInt(500000), // 0.5 USDC - insufficient
				},
				priority: 1,
			},
			&mockSignerWithPayload{
				mockSigner: &mockSigner{
					network:      "base",
					scheme:       "exact",
					canSignValue: true,
					priority:     2,
					maxAmount:    big.NewInt(2000000), // 2 USDC - sufficient
				},
				priority: 2,
			},
		},
		Selector: x402.NewDefaultPaymentSelector(),
	}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	// Should use priority 2 signer (skipping priority 1 due to insufficient max amount)
	if selectedSignerPriority != 2 {
		t.Errorf("expected signer with priority 2 to be selected due to max amount, got priority %d", selectedSignerPriority)
	}
}
