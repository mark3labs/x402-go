package client

import (
	"math/big"

	"github.com/mark3labs/x402-go"
)

// testSigner implements x402.Signer for testing
//
//nolint:unused // Reserved for future test implementations
type testSigner struct {
	network  string
	tokens   []x402.TokenConfig
	priority int
	canSign  bool
}

//nolint:unused // Reserved for future test implementations
func newTestSigner(network string, tokens []x402.TokenConfig, priority int, canSign bool) *testSigner {
	return &testSigner{
		network:  network,
		tokens:   tokens,
		priority: priority,
		canSign:  canSign,
	}
}

func (m *testSigner) Network() string {
	return m.network
}

func (m *testSigner) Scheme() string {
	return "exact"
}

func (m *testSigner) Sign(req *x402.PaymentRequirement) (*x402.PaymentPayload, error) {
	if !m.canSign {
		return nil, x402.ErrSigningFailed
	}
	return &x402.PaymentPayload{
		X402Version: 1,
		Scheme:      req.Scheme,
		Network:     req.Network,
		Payload: map[string]interface{}{
			"signature": "0xMockSignature",
		},
	}, nil
}

func (m *testSigner) CanSign(req *x402.PaymentRequirement) bool {
	if req.Network != m.network {
		return false
	}
	for _, token := range m.tokens {
		if token.Address == req.Asset {
			return m.canSign
		}
	}
	return false
}

func (m *testSigner) GetTokens() []x402.TokenConfig {
	return m.tokens
}

func (m *testSigner) GetPriority() int {
	return m.priority
}

func (m *testSigner) GetMaxAmount() *big.Int {
	return nil
}
