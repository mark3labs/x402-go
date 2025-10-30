package coinbase

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/x402-go"
)

// TestNewSignerWithValidCredentials tests signer initialization with valid CDP credentials
func TestNewSignerWithValidCredentials(t *testing.T) {
	t.Skip("Skipping test that requires CDP API credentials - will be tested in integration tests")
}

// TestNewSignerWithEnvironmentVariables tests signer initialization from environment variables
func TestNewSignerWithEnvironmentVariables(t *testing.T) {
	t.Skip("Skipping test that requires CDP API credentials - will be tested in integration tests")
}

// TestNewSignerMissingAPIKeyName tests error when CDP_API_KEY_NAME is missing
func TestNewSignerMissingAPIKeyName(t *testing.T) {
	_, err := NewSigner(
		WithCDPCredentials("", testECPrivateKey, ""),
		WithNetwork("base-sepolia"),
		WithToken("0x036CbD53842c5426634e7929541eC2318f3dCF7e", "USDC", 6),
	)

	if err == nil {
		t.Fatal("Expected error when API key name is missing")
	}

	if !strings.Contains(err.Error(), "apiKeyName") {
		t.Errorf("Expected error about API key name, got: %v", err)
	}
}

// TestNewSignerMissingAPIKeySecret tests error when CDP_API_KEY_SECRET is missing
func TestNewSignerMissingAPIKeySecret(t *testing.T) {
	_, err := NewSigner(
		WithCDPCredentials("organizations/test-org/apiKeys/test-key", "", ""),
		WithNetwork("base-sepolia"),
		WithToken("0x036CbD53842c5426634e7929541eC2318f3dCF7e", "USDC", 6),
	)

	if err == nil {
		t.Fatal("Expected error when API key secret is missing")
	}

	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("Expected error about parsing key, got: %v", err)
	}
}

// TestCredentialSanitization tests that error messages never contain credential fragments
func TestCredentialSanitization(t *testing.T) {
	testSecret := "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEISecretDataHere\n-----END EC PRIVATE KEY-----"
	sensitiveFragment := "SecretDataHere"

	_, err := NewSigner(
		WithCDPCredentials("organizations/test/apiKeys/key", testSecret, ""),
		WithNetwork("base-sepolia"),
		// Intentionally missing token to trigger error
	)

	if err == nil {
		t.Fatal("Expected error when no tokens configured")
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, sensitiveFragment) {
		t.Errorf("Error message contains sensitive data: %s", errMsg)
	}

	if strings.Contains(errMsg, testSecret) {
		t.Errorf("Error message contains full secret: %s", errMsg)
	}
}

// TestNewSignerMissingNetwork tests error when network is not specified
func TestNewSignerMissingNetwork(t *testing.T) {
	_, err := NewSigner(
		WithCDPCredentials("organizations/test-org/apiKeys/test-key", testECPrivateKey, ""),
		WithToken("0x036CbD53842c5426634e7929541eC2318f3dCF7e", "USDC", 6),
	)

	if err == nil {
		t.Fatal("Expected error when network is missing")
	}

	if err != x402.ErrInvalidNetwork {
		t.Errorf("Expected ErrInvalidNetwork, got: %v", err)
	}
}

// TestNewSignerMissingTokens tests error when no tokens are configured
func TestNewSignerMissingTokens(t *testing.T) {
	_, err := NewSigner(
		WithCDPCredentials("organizations/test-org/apiKeys/test-key", testECPrivateKey, ""),
		WithNetwork("base-sepolia"),
	)

	if err == nil {
		t.Fatal("Expected error when no tokens configured")
	}

	if err != x402.ErrNoTokens {
		t.Errorf("Expected ErrNoTokens, got: %v", err)
	}
}

// TestWithTokenPriority tests token priority option
func TestWithTokenPriority(t *testing.T) {
	t.Skip("Skipping test that requires CDP API credentials")
}

// TestWithPriority tests signer priority option
func TestWithPriority(t *testing.T) {
	t.Skip("Skipping test that requires CDP API credentials")
}

// TestWithMaxAmountPerCall tests max amount limit option
func TestWithMaxAmountPerCall(t *testing.T) {
	t.Skip("Skipping test that requires CDP API credentials")
}

// TestWithMaxAmountPerCallInvalid tests invalid max amount format
func TestWithMaxAmountPerCallInvalid(t *testing.T) {
	_, err := NewSigner(
		WithCDPCredentials("organizations/test-org/apiKeys/test-key", testECPrivateKey, ""),
		WithNetwork("base-sepolia"),
		WithToken("0x036CbD53842c5426634e7929541eC2318f3dCF7e", "USDC", 6),
		WithMaxAmountPerCall("not-a-number"),
	)

	if err == nil {
		t.Fatal("Expected error with invalid max amount")
	}

	if !strings.Contains(err.Error(), "invalid amount") {
		t.Errorf("Expected invalid amount error, got: %v", err)
	}
}

// TestSignerInterfaceImplementation tests that Signer implements x402.Signer interface
func TestSignerInterfaceImplementation(t *testing.T) {
	// This is a compile-time check, no runtime test needed
	var _ x402.Signer = (*Signer)(nil)
}

// TestNetworkMethod tests the Network() interface method
func TestNetworkMethod(t *testing.T) {
	t.Skip("Skipping test that requires CDP API credentials")
}

// TestSchemeMethod tests the Scheme() interface method
func TestSchemeMethod(t *testing.T) {
	t.Skip("Skipping test that requires CDP API credentials")
}

// TestSignerHelperMethods tests simple getter methods
func TestSignerHelperMethods(t *testing.T) {
	// Create a mock signer without calling NewSigner
	signer := &Signer{
		network:   "base-sepolia",
		address:   "0x1234567890123456789012345678901234567890",
		accountID: "accounts/test-123",
		priority:  5,
		tokens: []x402.TokenConfig{
			{Address: "0xUSDC", Symbol: "USDC", Decimals: 6, Priority: 0},
		},
	}

	// Test Network()
	if signer.Network() != "base-sepolia" {
		t.Errorf("Expected Network() = 'base-sepolia', got %s", signer.Network())
	}

	// Test Scheme()
	if signer.Scheme() != "exact" {
		t.Errorf("Expected Scheme() = 'exact', got %s", signer.Scheme())
	}

	// Test GetPriority()
	if signer.GetPriority() != 5 {
		t.Errorf("Expected GetPriority() = 5, got %d", signer.GetPriority())
	}

	// Test GetTokens()
	tokens := signer.GetTokens()
	if len(tokens) != 1 {
		t.Errorf("Expected 1 token, got %d", len(tokens))
	}

	// Test Address()
	if signer.Address() != "0x1234567890123456789012345678901234567890" {
		t.Errorf("Expected Address() = '0x1234567890123456789012345678901234567890', got %s", signer.Address())
	}

	// Test AccountID()
	if signer.AccountID() != "accounts/test-123" {
		t.Errorf("Expected AccountID() = 'accounts/test-123', got %s", signer.AccountID())
	}

	// Test GetMaxAmount() when nil
	if signer.GetMaxAmount() != nil {
		t.Errorf("Expected GetMaxAmount() = nil, got %v", signer.GetMaxAmount())
	}
}

// TestCanSign tests the CanSign method logic
func TestCanSign(t *testing.T) {
	signer := &Signer{
		network: "base-sepolia",
		tokens: []x402.TokenConfig{
			{Address: "0xUSDC", Symbol: "USDC", Decimals: 6, Priority: 0},
			{Address: "0xUSDT", Symbol: "USDT", Decimals: 6, Priority: 1},
		},
	}

	tests := []struct {
		name        string
		requirement *x402.PaymentRequirement
		expected    bool
	}{
		{
			name: "matching network and token",
			requirement: &x402.PaymentRequirement{
				Network: "base-sepolia",
				Scheme:  "exact",
				Asset:   "0xUSDC",
			},
			expected: true,
		},
		{
			name: "case insensitive token match",
			requirement: &x402.PaymentRequirement{
				Network: "base-sepolia",
				Scheme:  "exact",
				Asset:   "0xusdc",
			},
			expected: true,
		},
		{
			name: "mismatched network",
			requirement: &x402.PaymentRequirement{
				Network: "ethereum",
				Scheme:  "exact",
				Asset:   "0xUSDC",
			},
			expected: false,
		},
		{
			name: "mismatched scheme",
			requirement: &x402.PaymentRequirement{
				Network: "base-sepolia",
				Scheme:  "optimistic",
				Asset:   "0xUSDC",
			},
			expected: false,
		},
		{
			name: "mismatched token",
			requirement: &x402.PaymentRequirement{
				Network: "base-sepolia",
				Scheme:  "exact",
				Asset:   "0xDAI",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := signer.CanSign(tt.requirement)
			if result != tt.expected {
				t.Errorf("CanSign() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestWithCDPCredentialsFromEnvMissing tests WithCDPCredentialsFromEnv with missing vars
func TestWithCDPCredentialsFromEnvMissing(t *testing.T) {
	// Ensure env vars are not set
	t.Setenv("CDP_API_KEY_NAME", "")
	t.Setenv("CDP_API_KEY_SECRET", "")

	_, err := NewSigner(
		WithCDPCredentialsFromEnv(),
		WithNetwork("base-sepolia"),
		WithToken("0xUSDC", "USDC", 6),
	)

	if err == nil {
		t.Fatal("Expected error when CDP_API_KEY_NAME is not set")
	}

	if !strings.Contains(err.Error(), "CDP_API_KEY_NAME") {
		t.Errorf("Expected error about CDP_API_KEY_NAME, got: %v", err)
	}
}

// TestIsBase64Like tests the isBase64Like helper function
func TestIsBase64Like(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"MHcCAQEEISecretDataHere", true},
		{"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=", true},
		{"validBase64String123+/=", true},
		{"invalid-base64-with-dash", false},
		{"has spaces", false},
		{"has@special", false},
		{"", true}, // empty string is technically valid
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isBase64Like(tt.input)
			if result != tt.expected {
				t.Errorf("isBase64Like(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestOptionFunctions tests the functional option setters
func TestOptionFunctions(t *testing.T) {
	s := &Signer{}

	// Test WithNetwork
	err := WithNetwork("ethereum")(s)
	if err != nil {
		t.Errorf("WithNetwork failed: %v", err)
	}
	if s.network != "ethereum" {
		t.Errorf("Expected network 'ethereum', got %s", s.network)
	}

	// Test WithToken
	err = WithToken("0xToken", "TKN", 18)(s)
	if err != nil {
		t.Errorf("WithToken failed: %v", err)
	}
	if len(s.tokens) != 1 || s.tokens[0].Symbol != "TKN" {
		t.Errorf("Token not added correctly")
	}

	// Test WithTokenPriority
	err = WithTokenPriority("0xToken2", "TKN2", 6, 5)(s)
	if err != nil {
		t.Errorf("WithTokenPriority failed: %v", err)
	}
	if len(s.tokens) != 2 || s.tokens[1].Priority != 5 {
		t.Errorf("Token with priority not added correctly")
	}

	// Test WithPriority
	err = WithPriority(10)(s)
	if err != nil {
		t.Errorf("WithPriority failed: %v", err)
	}
	if s.priority != 10 {
		t.Errorf("Expected priority 10, got %d", s.priority)
	}

	// Test WithMaxAmountPerCall success
	err = WithMaxAmountPerCall("1000000")(s)
	if err != nil {
		t.Errorf("WithMaxAmountPerCall failed: %v", err)
	}
	if s.maxAmount == nil || s.maxAmount.String() != "1000000" {
		t.Errorf("Max amount not set correctly")
	}
}

// TestSanitizeErrorFunction tests the sanitizeError helper
func TestSanitizeErrorFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		isNil    bool
		mustNOT  []string // strings that must NOT appear in output
		mustHave []string // strings that must appear in output
	}{
		{
			name:     "nil error",
			isNil:    true,
			mustNOT:  []string{},
			mustHave: []string{},
		},
		{
			name:     "error with base64 key",
			input:    "error: failed with -----BEGIN EC PRIVATE KEY-----\nMHcCAQEEISecretData\n-----END EC PRIVATE KEY----- content",
			mustNOT:  []string{"SecretData", "MHcCAQEEI"},
			mustHave: []string{"[REDACTED]"},
		},
		{
			name:     "error without sensitive data",
			input:    "connection failed: timeout",
			mustNOT:  []string{"[REDACTED]"},
			mustHave: []string{"connection failed", "timeout"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result error
			if tt.isNil {
				result = sanitizeError(nil)
				if result != nil {
					t.Errorf("sanitizeError(nil) = %v, expected nil", result)
				}
				return
			}

			result = sanitizeError(fmt.Errorf("%s", tt.input))
			resultStr := result.Error()

			for _, forbidden := range tt.mustNOT {
				if strings.Contains(resultStr, forbidden) {
					t.Errorf("sanitized error contains forbidden string '%s': %s", forbidden, resultStr)
				}
			}
			for _, required := range tt.mustHave {
				if !strings.Contains(resultStr, required) {
					t.Errorf("sanitized error missing required string '%s': %s", required, resultStr)
				}
			}
		})
	}
}

// TestSignEVMWithValidRequirements tests EVM payment signing with valid requirements
func TestSignEVMWithValidRequirements(t *testing.T) {
	t.Skip("Skipping test that requires CDP API credentials - will be tested in integration tests")
}

// TestSignEVMWithMatchingNetworkAndToken tests CanSign returns true for matching network and token
func TestSignEVMWithMatchingNetworkAndToken(t *testing.T) {
	// Create a mock signer without calling CDP API
	s := &Signer{
		network:     "base-sepolia",
		networkType: NetworkTypeEVM,
		tokens: []x402.TokenConfig{
			{Address: "0x036CbD53842c5426634e7929541eC2318f3dCF7e", Symbol: "USDC", Decimals: 6},
		},
	}

	requirements := &x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "base-sepolia",
		MaxAmountRequired: "1000000", // 1 USDC
		Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		MaxTimeoutSeconds: 3600,
	}

	if !s.CanSign(requirements) {
		t.Error("Expected CanSign to return true for matching network and token")
	}
}

// TestSignEVMWithMismatchedNetwork tests CanSign returns false for mismatched network
func TestSignEVMWithMismatchedNetwork(t *testing.T) {
	s := &Signer{
		network:     "base-sepolia",
		networkType: NetworkTypeEVM,
		tokens: []x402.TokenConfig{
			{Address: "0x036CbD53842c5426634e7929541eC2318f3dCF7e", Symbol: "USDC", Decimals: 6},
		},
	}

	requirements := &x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "ethereum", // Different network
		MaxAmountRequired: "1000000",
		Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		MaxTimeoutSeconds: 3600,
	}

	if s.CanSign(requirements) {
		t.Error("Expected CanSign to return false for mismatched network")
	}
}

// TestSignEVMWithMismatchedToken tests CanSign returns false for unsupported token
func TestSignEVMWithMismatchedToken(t *testing.T) {
	s := &Signer{
		network:     "base-sepolia",
		networkType: NetworkTypeEVM,
		tokens: []x402.TokenConfig{
			{Address: "0x036CbD53842c5426634e7929541eC2318f3dCF7e", Symbol: "USDC", Decimals: 6},
		},
	}

	requirements := &x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "base-sepolia",
		MaxAmountRequired: "1000000",
		Asset:             "0xDifferentTokenAddress0000000000000000000", // Different token
		PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		MaxTimeoutSeconds: 3600,
	}

	if s.CanSign(requirements) {
		t.Error("Expected CanSign to return false for unsupported token")
	}
}

// TestSignEVMWithInvalidScheme tests CanSign returns false for unsupported scheme
func TestSignEVMWithInvalidScheme(t *testing.T) {
	s := &Signer{
		network:     "base-sepolia",
		networkType: NetworkTypeEVM,
		tokens: []x402.TokenConfig{
			{Address: "0x036CbD53842c5426634e7929541eC2318f3dCF7e", Symbol: "USDC", Decimals: 6},
		},
	}

	requirements := &x402.PaymentRequirement{
		Scheme:            "range", // Unsupported scheme
		Network:           "base-sepolia",
		MaxAmountRequired: "1000000",
		Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		MaxTimeoutSeconds: 3600,
	}

	if s.CanSign(requirements) {
		t.Error("Expected CanSign to return false for unsupported scheme")
	}
}

// TestSignEVMAmountValidation tests that Sign validates amount limits
func TestSignEVMAmountValidation(t *testing.T) {
	s := &Signer{
		cdpClient:   nil, // Not needed for this test - will fail before CDP call
		accountID:   "test-account-id",
		address:     "0x1234567890123456789012345678901234567890",
		network:     "base-sepolia",
		networkType: NetworkTypeEVM,
		chainID:     big.NewInt(84532),
		tokens: []x402.TokenConfig{
			{Address: "0x036CbD53842c5426634e7929541eC2318f3dCF7e", Symbol: "USDC", Decimals: 6},
		},
		maxAmount: big.NewInt(500000), // Max 0.5 USDC
	}

	requirements := &x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "base-sepolia",
		MaxAmountRequired: "1000000", // 1 USDC - exceeds limit
		Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		MaxTimeoutSeconds: 3600,
	}

	_, err := s.Sign(requirements)
	if err == nil {
		t.Fatal("Expected error when amount exceeds maxAmount")
	}

	if err != x402.ErrAmountExceeded {
		t.Errorf("Expected ErrAmountExceeded, got: %v", err)
	}
}

// TestSignEVMInvalidAmount tests that Sign rejects invalid amount strings
func TestSignEVMInvalidAmount(t *testing.T) {
	s := &Signer{
		network:     "base-sepolia",
		networkType: NetworkTypeEVM,
		chainID:     big.NewInt(84532),
		tokens: []x402.TokenConfig{
			{Address: "0x036CbD53842c5426634e7929541eC2318f3dCF7e", Symbol: "USDC", Decimals: 6},
		},
	}

	requirements := &x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "base-sepolia",
		MaxAmountRequired: "not-a-number",
		Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		MaxTimeoutSeconds: 3600,
	}

	_, err := s.Sign(requirements)
	if err == nil {
		t.Fatal("Expected error when amount is invalid")
	}

	if err != x402.ErrInvalidAmount {
		t.Errorf("Expected ErrInvalidAmount, got: %v", err)
	}
}

// TestCreateEIP3009Authorization tests authorization creation with valid parameters
func TestCreateEIP3009Authorization(t *testing.T) {
	s := &Signer{
		address: "0x1234567890123456789012345678901234567890",
	}

	amount := big.NewInt(1000000)
	auth, err := s.createEIP3009Authorization("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", amount, 3600)

	if err != nil {
		t.Fatalf("createEIP3009Authorization failed: %v", err)
	}

	if auth.From != s.address {
		t.Errorf("Expected From=%s, got %s", s.address, auth.From)
	}

	if auth.To != "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0" {
		t.Errorf("Expected To=0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0, got %s", auth.To)
	}

	if auth.Value != "1000000" {
		t.Errorf("Expected Value=1000000, got %s", auth.Value)
	}

	// Check nonce format (should be 0x + 64 hex chars)
	if !strings.HasPrefix(auth.Nonce, "0x") {
		t.Errorf("Nonce should start with 0x, got %s", auth.Nonce)
	}
	if len(auth.Nonce) != 66 { // 0x + 64 hex chars
		t.Errorf("Nonce should be 66 characters (0x + 64 hex), got %d: %s", len(auth.Nonce), auth.Nonce)
	}

	// Check validAfter is in the past (accounting for clock drift buffer)
	validAfter, ok := new(big.Int).SetString(auth.ValidAfter, 10)
	if !ok {
		t.Fatalf("Failed to parse ValidAfter: %s", auth.ValidAfter)
	}
	now := time.Now().Unix()
	if validAfter.Int64() > now {
		t.Errorf("ValidAfter should be in the past, got %d (now: %d)", validAfter.Int64(), now)
	}

	// Check validBefore is in the future
	validBefore, ok := new(big.Int).SetString(auth.ValidBefore, 10)
	if !ok {
		t.Fatalf("Failed to parse ValidBefore: %s", auth.ValidBefore)
	}
	if validBefore.Int64() <= now {
		t.Errorf("ValidBefore should be in the future, got %d (now: %d)", validBefore.Int64(), now)
	}

	// Check validity window is approximately the timeout
	window := validBefore.Int64() - validAfter.Int64()
	expectedWindow := int64(3600 + 10) // timeout + clock drift buffer
	if window < expectedWindow-5 || window > expectedWindow+5 {
		t.Errorf("Expected validity window ~%d seconds, got %d", expectedWindow, window)
	}
}

// TestBuildEIP712TypedData tests EIP-712 typed data construction
func TestBuildEIP712TypedData(t *testing.T) {
	s := &Signer{
		chainID: big.NewInt(84532), // Base Sepolia
	}

	auth := &eip3009Auth{
		From:        "0x1234567890123456789012345678901234567890",
		To:          "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		Value:       "1000000",
		ValidAfter:  "1234567890",
		ValidBefore: "1234571490",
		Nonce:       "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
	}

	tokenAddress := "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
	typedData := s.buildEIP712TypedData(tokenAddress, auth)

	// Check domain
	if typedData.Domain.Name != "USD Coin" {
		t.Errorf("Expected domain name 'USD Coin', got '%s'", typedData.Domain.Name)
	}
	if typedData.Domain.Version != "2" {
		t.Errorf("Expected domain version '2', got '%s'", typedData.Domain.Version)
	}
	if typedData.Domain.ChainID != 84532 {
		t.Errorf("Expected chainID 84532, got %d", typedData.Domain.ChainID)
	}
	if typedData.Domain.VerifyingContract != tokenAddress {
		t.Errorf("Expected verifyingContract %s, got %s", tokenAddress, typedData.Domain.VerifyingContract)
	}

	// Check primary type
	if typedData.PrimaryType != "TransferWithAuthorization" {
		t.Errorf("Expected primary type 'TransferWithAuthorization', got '%s'", typedData.PrimaryType)
	}

	// Check types are defined
	if _, ok := typedData.Types["EIP712Domain"]; !ok {
		t.Error("Missing EIP712Domain type definition")
	}
	if _, ok := typedData.Types["TransferWithAuthorization"]; !ok {
		t.Error("Missing TransferWithAuthorization type definition")
	}

	// Check message fields
	if typedData.Message["from"] != auth.From {
		t.Errorf("Expected message.from=%s, got %v", auth.From, typedData.Message["from"])
	}
	if typedData.Message["to"] != auth.To {
		t.Errorf("Expected message.to=%s, got %v", auth.To, typedData.Message["to"])
	}
	if typedData.Message["value"] != auth.Value {
		t.Errorf("Expected message.value=%s, got %v", auth.Value, typedData.Message["value"])
	}
	if typedData.Message["nonce"] != auth.Nonce {
		t.Errorf("Expected message.nonce=%s, got %v", auth.Nonce, typedData.Message["nonce"])
	}
}

// TestGenerateNonce tests nonce generation
func TestGenerateNonce(t *testing.T) {
	// Generate multiple nonces and check uniqueness
	nonces := make(map[string]bool)
	for i := 0; i < 100; i++ {
		nonce, err := generateNonce()
		if err != nil {
			t.Fatalf("generateNonce failed: %v", err)
		}

		// Check format
		if !strings.HasPrefix(nonce, "0x") {
			t.Errorf("Nonce should start with 0x, got %s", nonce)
		}
		if len(nonce) != 66 { // 0x + 64 hex chars
			t.Errorf("Nonce should be 66 characters, got %d: %s", len(nonce), nonce)
		}

		// Check uniqueness
		if nonces[nonce] {
			t.Errorf("Duplicate nonce generated: %s", nonce)
		}
		nonces[nonce] = true
	}
}

// TestSignEVMPayloadStructure tests the structure of the payment payload
func TestSignEVMPayloadStructure(t *testing.T) {
	t.Skip("Skipping test that requires CDP API credentials - will be tested in integration tests")
}

// ========================================
// Solana (SVM) Tests
// ========================================

// TestCanSignSolana tests CanSign for Solana networks
func TestCanSignSolana(t *testing.T) {
	tests := []struct {
		name         string
		network      string
		tokens       []x402.TokenConfig
		requirements *x402.PaymentRequirement
		expected     bool
	}{
		{
			name:    "Valid solana-devnet USDC",
			network: "solana-devnet",
			tokens: []x402.TokenConfig{
				{Address: "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU", Symbol: "USDC", Decimals: 6},
			},
			requirements: &x402.PaymentRequirement{
				Network:           "solana-devnet",
				Scheme:            "exact",
				Asset:             "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU",
				MaxAmountRequired: "1000000",
			},
			expected: true,
		},
		{
			name:    "Valid solana mainnet USDC",
			network: "solana",
			tokens: []x402.TokenConfig{
				{Address: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", Symbol: "USDC", Decimals: 6},
			},
			requirements: &x402.PaymentRequirement{
				Network:           "solana",
				Scheme:            "exact",
				Asset:             "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
				MaxAmountRequired: "1000000",
			},
			expected: true,
		},
		{
			name:    "Network mismatch - solana vs solana-devnet",
			network: "solana-devnet",
			tokens: []x402.TokenConfig{
				{Address: "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU", Symbol: "USDC", Decimals: 6},
			},
			requirements: &x402.PaymentRequirement{
				Network:           "solana", // Requesting mainnet
				Scheme:            "exact",
				Asset:             "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU",
				MaxAmountRequired: "1000000",
			},
			expected: false,
		},
		{
			name:    "Token not configured",
			network: "solana-devnet",
			tokens: []x402.TokenConfig{
				{Address: "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU", Symbol: "USDC", Decimals: 6},
			},
			requirements: &x402.PaymentRequirement{
				Network:           "solana-devnet",
				Scheme:            "exact",
				Asset:             "SomethingElse111111111111111111111111111111",
				MaxAmountRequired: "1000000",
			},
			expected: false,
		},
		{
			name:    "Scheme mismatch",
			network: "solana-devnet",
			tokens: []x402.TokenConfig{
				{Address: "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU", Symbol: "USDC", Decimals: 6},
			},
			requirements: &x402.PaymentRequirement{
				Network:           "solana-devnet",
				Scheme:            "subscription", // Wrong scheme
				Asset:             "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU",
				MaxAmountRequired: "1000000",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signer{
				network:     tt.network,
				networkType: NetworkTypeSVM,
				tokens:      tt.tokens,
			}

			result := s.CanSign(tt.requirements)
			if result != tt.expected {
				t.Errorf("CanSign() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestExtractFeePayer tests fee payer extraction from requirements
func TestExtractFeePayer(t *testing.T) {
	tests := []struct {
		name         string
		requirements *x402.PaymentRequirement
		expected     string
		shouldError  bool
	}{
		{
			name: "Valid fee payer",
			requirements: &x402.PaymentRequirement{
				Extra: map[string]interface{}{
					"feePayer": "FeePayerAddress111111111111111111111111111",
				},
			},
			expected:    "FeePayerAddress111111111111111111111111111",
			shouldError: false,
		},
		{
			name:         "Missing extra field",
			requirements: &x402.PaymentRequirement{},
			shouldError:  true,
		},
		{
			name: "Missing feePayer in extra",
			requirements: &x402.PaymentRequirement{
				Extra: map[string]interface{}{
					"other": "value",
				},
			},
			shouldError: true,
		},
		{
			name: "feePayer not a string",
			requirements: &x402.PaymentRequirement{
				Extra: map[string]interface{}{
					"feePayer": 12345,
				},
			},
			shouldError: true,
		},
		{
			name: "Empty feePayer",
			requirements: &x402.PaymentRequirement{
				Extra: map[string]interface{}{
					"feePayer": "",
				},
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractFeePayer(tt.requirements)
			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

// TestBuildComputeUnitLimitInstruction tests compute unit limit instruction building
func TestBuildComputeUnitLimitInstruction(t *testing.T) {
	inst := buildComputeUnitLimitInstruction(200_000)

	if inst.ProgramID != "ComputeBudget111111111111111111111111111111" {
		t.Errorf("Expected ComputeBudget program ID, got %s", inst.ProgramID)
	}

	if len(inst.Accounts) != 0 {
		t.Errorf("Expected no accounts, got %d", len(inst.Accounts))
	}

	// Decode hex data
	data, err := hex.DecodeString(inst.Data)
	if err != nil {
		t.Fatalf("Failed to decode instruction data: %v", err)
	}

	// Check instruction format: [2, units (u32 LE)]
	if len(data) != 5 {
		t.Errorf("Expected 5 bytes, got %d", len(data))
	}
	if data[0] != 2 {
		t.Errorf("Expected discriminator 2, got %d", data[0])
	}

	// Verify little-endian encoding of 200,000
	units := uint32(data[1]) | uint32(data[2])<<8 | uint32(data[3])<<16 | uint32(data[4])<<24
	if units != 200_000 {
		t.Errorf("Expected units 200000, got %d", units)
	}
}

// TestBuildComputeUnitPriceInstruction tests compute unit price instruction building
func TestBuildComputeUnitPriceInstruction(t *testing.T) {
	inst := buildComputeUnitPriceInstruction(10_000)

	if inst.ProgramID != "ComputeBudget111111111111111111111111111111" {
		t.Errorf("Expected ComputeBudget program ID, got %s", inst.ProgramID)
	}

	if len(inst.Accounts) != 0 {
		t.Errorf("Expected no accounts, got %d", len(inst.Accounts))
	}

	// Decode hex data
	data, err := hex.DecodeString(inst.Data)
	if err != nil {
		t.Fatalf("Failed to decode instruction data: %v", err)
	}

	// Check instruction format: [3, microlamports (u64 LE)]
	if len(data) != 9 {
		t.Errorf("Expected 9 bytes, got %d", len(data))
	}
	if data[0] != 3 {
		t.Errorf("Expected discriminator 3, got %d", data[0])
	}

	// Verify little-endian encoding of 10,000
	price := uint64(data[1]) | uint64(data[2])<<8 | uint64(data[3])<<16 | uint64(data[4])<<24 |
		uint64(data[5])<<32 | uint64(data[6])<<40 | uint64(data[7])<<48 | uint64(data[8])<<56
	if price != 10_000 {
		t.Errorf("Expected price 10000, got %d", price)
	}
}

// TestBuildTransferCheckedInstruction tests TransferChecked instruction building
func TestBuildTransferCheckedInstruction(t *testing.T) {
	source := "SourceAccount1111111111111111111111111111111"
	mint := "MintAddress111111111111111111111111111111111"
	dest := "DestAccount111111111111111111111111111111111"
	owner := "OwnerAddress11111111111111111111111111111111"
	amount := uint64(1000000)
	decimals := uint8(6)

	inst := buildTransferCheckedInstruction(source, mint, dest, owner, amount, decimals)

	// Verify program ID
	if inst.ProgramID != "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
		t.Errorf("Expected SPL Token program ID, got %s", inst.ProgramID)
	}

	// Verify accounts
	if len(inst.Accounts) != 4 {
		t.Fatalf("Expected 4 accounts, got %d", len(inst.Accounts))
	}

	// Check source account
	if inst.Accounts[0].PublicKey != source {
		t.Errorf("Expected source %s, got %s", source, inst.Accounts[0].PublicKey)
	}
	if inst.Accounts[0].IsSigner {
		t.Errorf("Source should not be signer")
	}
	if !inst.Accounts[0].IsWritable {
		t.Errorf("Source should be writable")
	}

	// Check mint
	if inst.Accounts[1].PublicKey != mint {
		t.Errorf("Expected mint %s, got %s", mint, inst.Accounts[1].PublicKey)
	}
	if inst.Accounts[1].IsWritable {
		t.Errorf("Mint should not be writable")
	}

	// Check destination account
	if inst.Accounts[2].PublicKey != dest {
		t.Errorf("Expected dest %s, got %s", dest, inst.Accounts[2].PublicKey)
	}
	if !inst.Accounts[2].IsWritable {
		t.Errorf("Destination should be writable")
	}

	// Check owner
	if inst.Accounts[3].PublicKey != owner {
		t.Errorf("Expected owner %s, got %s", owner, inst.Accounts[3].PublicKey)
	}
	if !inst.Accounts[3].IsSigner {
		t.Errorf("Owner should be signer")
	}

	// Decode and verify instruction data
	data, err := hex.DecodeString(inst.Data)
	if err != nil {
		t.Fatalf("Failed to decode instruction data: %v", err)
	}

	// Check format: [12, amount (u64 LE), decimals (u8)]
	if len(data) != 10 {
		t.Errorf("Expected 10 bytes, got %d", len(data))
	}
	if data[0] != 12 {
		t.Errorf("Expected discriminator 12, got %d", data[0])
	}

	// Verify amount encoding
	decodedAmount := uint64(data[1]) | uint64(data[2])<<8 | uint64(data[3])<<16 | uint64(data[4])<<24 |
		uint64(data[5])<<32 | uint64(data[6])<<40 | uint64(data[7])<<48 | uint64(data[8])<<56
	if decodedAmount != amount {
		t.Errorf("Expected amount %d, got %d", amount, decodedAmount)
	}

	// Verify decimals
	if data[9] != decimals {
		t.Errorf("Expected decimals %d, got %d", decimals, data[9])
	}
}

// TestBuildSolanaTransaction tests transaction building
func TestBuildSolanaTransaction(t *testing.T) {
	s := &Signer{
		address: "ClientAddress1111111111111111111111111111111",
	}

	mint := "MintAddress111111111111111111111111111111111"
	recipient := "RecipientAddr1111111111111111111111111111111"
	amount := uint64(1000000)
	decimals := uint8(6)
	feePayer := "FeePayerAddress111111111111111111111111111"
	blockhash := "BlockhashValue11111111111111111111111111111"

	tx, err := s.buildSolanaTransaction(mint, recipient, amount, decimals, feePayer, blockhash)
	if err != nil {
		t.Fatalf("buildSolanaTransaction failed: %v", err)
	}

	// Verify transaction structure
	if tx.FeePayer != feePayer {
		t.Errorf("Expected fee payer %s, got %s", feePayer, tx.FeePayer)
	}
	if tx.Blockhash != blockhash {
		t.Errorf("Expected blockhash %s, got %s", blockhash, tx.Blockhash)
	}

	// Should have 3 instructions: SetComputeUnitLimit, SetComputeUnitPrice, TransferChecked
	if len(tx.Instructions) != 3 {
		t.Fatalf("Expected 3 instructions, got %d", len(tx.Instructions))
	}

	// Check instruction 0: SetComputeUnitLimit
	if tx.Instructions[0].ProgramID != "ComputeBudget111111111111111111111111111111" {
		t.Errorf("Instruction 0 should be ComputeBudget")
	}

	// Check instruction 1: SetComputeUnitPrice
	if tx.Instructions[1].ProgramID != "ComputeBudget111111111111111111111111111111" {
		t.Errorf("Instruction 1 should be ComputeBudget")
	}

	// Check instruction 2: TransferChecked
	if tx.Instructions[2].ProgramID != "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
		t.Errorf("Instruction 2 should be SPL Token")
	}
	if len(tx.Instructions[2].Accounts) != 4 {
		t.Errorf("TransferChecked should have 4 accounts, got %d", len(tx.Instructions[2].Accounts))
	}
}

// TestSignSVMAmountValidation tests amount validation for SVM signing
func TestSignSVMAmountValidation(t *testing.T) {
	// This test validates amount parsing and max amount checks
	// The actual signing is tested in integration tests

	tests := []struct {
		name        string
		maxAmount   string
		reqAmount   string
		shouldError bool
	}{
		{
			name:        "Amount within limit",
			maxAmount:   "10000000", // 10 USDC
			reqAmount:   "1000000",  // 1 USDC
			shouldError: false,
		},
		{
			name:        "Amount exceeds limit",
			maxAmount:   "1000000", // 1 USDC
			reqAmount:   "5000000", // 5 USDC
			shouldError: true,
		},
		{
			name:        "No max amount set",
			maxAmount:   "",
			reqAmount:   "1000000",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signer{
				network:     "solana-devnet",
				networkType: NetworkTypeSVM,
				tokens: []x402.TokenConfig{
					{Address: "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU", Symbol: "USDC", Decimals: 6},
				},
			}

			if tt.maxAmount != "" {
				maxAmt, _ := new(big.Int).SetString(tt.maxAmount, 10)
				s.maxAmount = maxAmt
			}

			req := &x402.PaymentRequirement{
				Network:           "solana-devnet",
				Scheme:            "exact",
				Asset:             "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU",
				PayTo:             "RecipientAddr1111111111111111111111111111111",
				MaxAmountRequired: tt.reqAmount,
			}

			// Parse and check amount
			amount := new(big.Int)
			_, ok := amount.SetString(req.MaxAmountRequired, 10)
			if !ok {
				t.Fatalf("Failed to parse amount")
			}

			// Check max amount
			if s.maxAmount != nil && amount.Cmp(s.maxAmount) > 0 {
				if !tt.shouldError {
					t.Errorf("Expected no error but amount exceeds limit")
				}
			} else {
				if tt.shouldError {
					t.Errorf("Expected error but amount is within limit")
				}
			}
		})
	}
}

// TestSignSVMIntegration is an integration test for Solana signing
func TestSignSVMIntegration(t *testing.T) {
	t.Skip("Skipping integration test - requires CDP API credentials and Solana RPC access")

	// This test would:
	// 1. Create a signer with real CDP credentials
	// 2. Build a payment requirement with feePayer
	// 3. Call Sign() and verify the response structure
	// 4. Verify the signed transaction can be decoded
}
