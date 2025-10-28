package http

import (
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mark3labs/x402-go"
)

// mockSigner implements x402.Signer for testing
type mockSigner struct {
	network      string
	scheme       string
	canSignValue bool
	signError    error
	priority     int
	maxAmount    *big.Int
}

func (m *mockSigner) Network() string                            { return m.network }
func (m *mockSigner) Scheme() string                             { return m.scheme }
func (m *mockSigner) CanSign(req *x402.PaymentRequirement) bool { return m.canSignValue }
func (m *mockSigner) GetPriority() int                           { return m.priority }
func (m *mockSigner) GetTokens() []x402.TokenConfig              { return nil }
func (m *mockSigner) GetMaxAmount() *big.Int                     { return m.maxAmount }

func (m *mockSigner) Sign(req *x402.PaymentRequirement) (*x402.PaymentPayload, error) {
	if m.signError != nil {
		return nil, m.signError
	}
	return &x402.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     m.network,
		Payload: map[string]interface{}{
			"test": "payload",
		},
	}, nil
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		opts    []ClientOption
		wantErr bool
	}{
		{
			name:    "default client",
			opts:    nil,
			wantErr: false,
		},
		{
			name: "client with custom HTTP client",
			opts: []ClientOption{
				WithHTTPClient(&http.Client{
					Timeout: 30 * time.Second,
				}),
			},
			wantErr: false,
		},
		{
			name: "client with signer",
			opts: []ClientOption{
				WithSigner(&mockSigner{
					network:      "base",
					scheme:       "exact",
					canSignValue: true,
				}),
			},
			wantErr: false,
		},
		{
			name: "client with multiple signers",
			opts: []ClientOption{
				WithSigner(&mockSigner{
					network:      "base",
					scheme:       "exact",
					canSignValue: true,
					priority:     1,
				}),
				WithSigner(&mockSigner{
					network:      "solana",
					scheme:       "exact",
					canSignValue: true,
					priority:     2,
				}),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestClient_WithSigner(t *testing.T) {
	client, err := NewClient(
		WithSigner(&mockSigner{
			network:      "base",
			scheme:       "exact",
			canSignValue: true,
		}),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Verify transport is wrapped
	transport, ok := client.Transport.(*X402Transport)
	if !ok {
		t.Fatal("expected X402Transport")
	}

	if len(transport.Signers) != 1 {
		t.Errorf("expected 1 signer, got %d", len(transport.Signers))
	}
}

func TestClient_WithMultipleSigners(t *testing.T) {
	signer1 := &mockSigner{network: "base", scheme: "exact", canSignValue: true, priority: 1}
	signer2 := &mockSigner{network: "solana", scheme: "exact", canSignValue: true, priority: 2}

	client, err := NewClient(
		WithSigner(signer1),
		WithSigner(signer2),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	transport, ok := client.Transport.(*X402Transport)
	if !ok {
		t.Fatal("expected X402Transport")
	}

	if len(transport.Signers) != 2 {
		t.Errorf("expected 2 signers, got %d", len(transport.Signers))
	}
}

func TestClient_WithCustomHTTPClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	client, err := NewClient(
		WithHTTPClient(customClient),
		WithSigner(&mockSigner{
			network:      "base",
			scheme:       "exact",
			canSignValue: true,
		}),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Verify the custom timeout is preserved
	if client.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", client.Timeout)
	}
}

func TestClient_WithSelector(t *testing.T) {
	customSelector := x402.NewDefaultPaymentSelector()

	client, err := NewClient(
		WithSelector(customSelector),
		WithSigner(&mockSigner{
			network:      "base",
			scheme:       "exact",
			canSignValue: true,
		}),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	transport, ok := client.Transport.(*X402Transport)
	if !ok {
		t.Fatal("expected X402Transport")
	}

	if transport.Selector == nil {
		t.Error("expected selector to be set")
	}
}

func TestClient_NonPaymentRequest(t *testing.T) {
	// Create a test server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	// Create client without signers
	client, err := NewClient()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_StdlibCompatibility(t *testing.T) {
	// Test that client works like a standard http.Client for non-402 requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that no payment headers are added for non-402 requests
		if r.Header.Get("X-PAYMENT") != "" {
			t.Error("unexpected X-PAYMENT header on non-402 request")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client, err := NewClient(
		WithSigner(&mockSigner{
			network:      "base",
			scheme:       "exact",
			canSignValue: true,
		}),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestGetSettlement_NoHeader(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{},
	}

	settlement := GetSettlement(resp)
	if settlement != nil {
		t.Error("expected nil settlement when header is missing")
	}
}

func TestGetSettlement_InvalidBase64(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{
			"X-Settlement": []string{"invalid base64!!!"},
		},
	}

	settlement := GetSettlement(resp)
	if settlement != nil {
		t.Error("expected nil settlement for invalid base64")
	}
}

func TestGetSettlement_InvalidJSON(t *testing.T) {
	// Valid base64 but invalid JSON
	resp := &http.Response{
		Header: http.Header{
			"X-Settlement": []string{"bm90IGpzb24="}, // "not json" in base64
		},
	}

	settlement := GetSettlement(resp)
	if settlement != nil {
		t.Error("expected nil settlement for invalid JSON")
	}
}
