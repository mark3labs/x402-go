package coinbase

import (
	"crypto/sha256"
	"strings"
	"testing"
	"time"

	"gopkg.in/square/go-jose.v2/jwt"
)

// Test EC private key (ECDSA P-256) - DO NOT USE IN PRODUCTION
const testECPrivateKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIGlRFY0J0gbOFJbZqHRIhzgFjt6sMdVlvL+8zBcCIJmoAoGCCqGSM49
AwEHoUQDQgAEzXDFO5wEOHqMNLhFqn1NJl3vXqKLJJqL0YNn2R3DJCDm7fRXQzKt
YMJcQFMQKmC0BNm7hPpYPKJbZEcLQ9chMg==
-----END EC PRIVATE KEY-----`

// Test invalid PEM format
const testInvalidPEM = `-----BEGIN EC PRIVATE KEY-----
THIS IS NOT A VALID KEY
-----END EC PRIVATE KEY-----`

// Test non-PEM data
const testNonPEM = `this is not PEM encoded data at all`

func TestNewCDPAuth_ValidCredentials(t *testing.T) {
	tests := []struct {
		name         string
		apiKeyName   string
		apiKeySecret string
		walletSecret string
		wantErr      bool
	}{
		{
			name:         "valid ECDSA credentials with wallet secret",
			apiKeyName:   "organizations/test-org/apiKeys/test-key",
			apiKeySecret: testECPrivateKey,
			walletSecret: "wallet-secret-123",
			wantErr:      false,
		},
		{
			name:         "valid ECDSA credentials without wallet secret",
			apiKeyName:   "organizations/test-org/apiKeys/test-key",
			apiKeySecret: testECPrivateKey,
			walletSecret: "",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewCDPAuth(tt.apiKeyName, tt.apiKeySecret, tt.walletSecret)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if auth == nil {
				t.Fatal("expected auth to be non-nil")
			}

			// Verify fields are set
			if auth.apiKeyName != tt.apiKeyName {
				t.Errorf("expected apiKeyName %s, got %s", tt.apiKeyName, auth.apiKeyName)
			}

			if auth.apiKeySecret != tt.apiKeySecret {
				t.Errorf("expected apiKeySecret to be set")
			}

			if auth.walletSecret != tt.walletSecret {
				t.Errorf("expected walletSecret %s, got %s", tt.walletSecret, auth.walletSecret)
			}

			// Verify private key was parsed and cached
			if auth.privateKey == nil {
				t.Fatal("expected privateKey to be parsed and cached")
			}
		})
	}
}

func TestNewCDPAuth_InvalidCredentials(t *testing.T) {
	tests := []struct {
		name         string
		apiKeyName   string
		apiKeySecret string
		walletSecret string
		wantErrMsg   string
	}{
		{
			name:         "empty API key name",
			apiKeyName:   "",
			apiKeySecret: testECPrivateKey,
			walletSecret: "wallet-secret",
			wantErrMsg:   "apiKeyName must not be empty",
		},
		{
			name:         "invalid PEM format",
			apiKeyName:   "organizations/test-org/apiKeys/test-key",
			apiKeySecret: testInvalidPEM,
			walletSecret: "wallet-secret",
			wantErrMsg:   "failed to decode PEM block",
		},
		{
			name:         "non-PEM data",
			apiKeyName:   "organizations/test-org/apiKeys/test-key",
			apiKeySecret: testNonPEM,
			walletSecret: "wallet-secret",
			wantErrMsg:   "failed to decode PEM block",
		},
		{
			name:         "empty PEM data",
			apiKeyName:   "organizations/test-org/apiKeys/test-key",
			apiKeySecret: "",
			walletSecret: "wallet-secret",
			wantErrMsg:   "failed to decode PEM block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewCDPAuth(tt.apiKeyName, tt.apiKeySecret, tt.walletSecret)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if auth != nil {
				t.Errorf("expected auth to be nil, got %+v", auth)
			}

			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("expected error to contain %q, got %q", tt.wantErrMsg, err.Error())
			}
		})
	}
}

func TestGenerateBearerToken_ValidToken(t *testing.T) {
	auth, err := NewCDPAuth(
		"organizations/test-org/apiKeys/test-key",
		testECPrivateKey,
		"wallet-secret",
	)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "GET request",
			method: "GET",
			path:   "/platform/v2/evm/accounts",
		},
		{
			name:   "POST request",
			method: "POST",
			path:   "/platform/v2/evm/accounts",
		},
		{
			name:   "PUT request",
			method: "PUT",
			path:   "/platform/v2/evm/accounts/test-account",
		},
		{
			name:   "DELETE request",
			method: "DELETE",
			path:   "/platform/v2/evm/accounts/test-account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := auth.GenerateBearerToken(tt.method, tt.path)
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}

			if token == "" {
				t.Fatal("expected non-empty token")
			}

			// Parse and validate JWT structure
			parsedToken, err := jwt.ParseSigned(token)
			if err != nil {
				t.Fatalf("failed to parse JWT: %v", err)
			}

			// Verify token has headers
			if len(parsedToken.Headers) == 0 {
				t.Fatal("expected JWT to have headers")
			}

			// Verify algorithm is ES256
			if parsedToken.Headers[0].Algorithm != "ES256" {
				t.Errorf("expected algorithm ES256, got %s", parsedToken.Headers[0].Algorithm)
			}

			// Verify kid header (stored in KeyID field)
			kid := parsedToken.Headers[0].KeyID
			if kid == "" {
				t.Fatal("expected kid (KeyID) header to be set")
			}
			if kid != "organizations/test-org/apiKeys/test-key" {
				t.Errorf("expected kid %s, got %s", "organizations/test-org/apiKeys/test-key", kid)
			}

			// Verify type header
			typ, ok := parsedToken.Headers[0].ExtraHeaders["typ"]
			if !ok {
				t.Fatal("expected typ header to be set")
			}
			if typ != "JWT" {
				t.Errorf("expected typ JWT, got %s", typ)
			}
		})
	}
}

func TestGenerateBearerToken_Claims(t *testing.T) {
	auth, err := NewCDPAuth(
		"organizations/test-org/apiKeys/test-key",
		testECPrivateKey,
		"wallet-secret",
	)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	method := "GET"
	path := "/platform/v2/evm/accounts"

	token, err := auth.GenerateBearerToken(method, path)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Parse token (we can't verify signature without extracting public key, but we can inspect claims)
	parsedToken, err := jwt.ParseSigned(token)
	if err != nil {
		t.Fatalf("failed to parse JWT: %v", err)
	}

	// Extract claims without verification (for testing purposes)
	var claims APIKeyClaims
	if err := parsedToken.UnsafeClaimsWithoutVerification(&claims); err != nil {
		t.Fatalf("failed to extract claims: %v", err)
	}

	// Verify subject
	if claims.Subject != "organizations/test-org/apiKeys/test-key" {
		t.Errorf("expected subject %s, got %s", "organizations/test-org/apiKeys/test-key", claims.Subject)
	}

	// Verify issuer
	if claims.Issuer != "coinbase-cloud" {
		t.Errorf("expected issuer coinbase-cloud, got %s", claims.Issuer)
	}

	// Verify URI format
	expectedURI := "GET api.cdp.coinbase.com/platform/v2/evm/accounts"
	if claims.URI != expectedURI {
		t.Errorf("expected URI %s, got %s", expectedURI, claims.URI)
	}

	// Verify expiration is approximately 2 minutes from now
	now := time.Now()
	expTime := claims.Expiry.Time()
	expectedExp := now.Add(2 * time.Minute)
	timeDiff := expTime.Sub(expectedExp)
	if timeDiff < -5*time.Second || timeDiff > 5*time.Second {
		t.Errorf("expected expiration around %v, got %v (diff: %v)", expectedExp, expTime, timeDiff)
	}

	// Verify not-before is approximately now
	nbfTime := claims.NotBefore.Time()
	nbfDiff := nbfTime.Sub(now)
	if nbfDiff < -5*time.Second || nbfDiff > 5*time.Second {
		t.Errorf("expected not-before around %v, got %v (diff: %v)", now, nbfTime, nbfDiff)
	}

	// Verify reqHash is not set for Bearer token without body
	if claims.ReqHash != "" {
		t.Errorf("expected empty reqHash for Bearer token, got %s", claims.ReqHash)
	}
}

func TestGenerateWalletAuthToken_WithBodyHash(t *testing.T) {
	auth, err := NewCDPAuth(
		"organizations/test-org/apiKeys/test-key",
		testECPrivateKey,
		"wallet-secret",
	)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	method := "POST"
	path := "/platform/v2/evm/accounts/0x742d35Cc6634C0532925a3b844Bc454e4438f44e/sign/typed-data"

	// Create a test body and hash it
	testBody := []byte(`{"typedData": {"domain": {"name": "Test"}}}`)
	hash := sha256.Sum256(testBody)

	token, err := auth.GenerateWalletAuthToken(method, path, hash[:])
	if err != nil {
		t.Fatalf("failed to generate wallet auth token: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Parse token
	parsedToken, err := jwt.ParseSigned(token)
	if err != nil {
		t.Fatalf("failed to parse JWT: %v", err)
	}

	// Extract claims
	var claims APIKeyClaims
	if err := parsedToken.UnsafeClaimsWithoutVerification(&claims); err != nil {
		t.Fatalf("failed to extract claims: %v", err)
	}

	// Verify reqHash is set and correct
	expectedHash := "367fc472194508bcb21c0c67de93721bdc937da40cccf39b984929ee55cd6f32" // SHA-256 of testBody
	if claims.ReqHash != expectedHash {
		t.Errorf("expected reqHash %s, got %s", expectedHash, claims.ReqHash)
	}

	// Verify expiration is approximately 1 minute from now
	now := time.Now()
	expTime := claims.Expiry.Time()
	expectedExp := now.Add(1 * time.Minute)
	timeDiff := expTime.Sub(expectedExp)
	if timeDiff < -5*time.Second || timeDiff > 5*time.Second {
		t.Errorf("expected expiration around %v, got %v (diff: %v)", expectedExp, expTime, timeDiff)
	}

	// Verify URI format
	expectedURI := "POST api.cdp.coinbase.com/platform/v2/evm/accounts/0x742d35Cc6634C0532925a3b844Bc454e4438f44e/sign/typed-data"
	if claims.URI != expectedURI {
		t.Errorf("expected URI %s, got %s", expectedURI, claims.URI)
	}
}

func TestGenerateWalletAuthToken_EmptyBodyHash(t *testing.T) {
	auth, err := NewCDPAuth(
		"organizations/test-org/apiKeys/test-key",
		testECPrivateKey,
		"wallet-secret",
	)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	method := "POST"
	path := "/platform/v2/evm/accounts/test-account/sign"

	// Generate token with nil body hash
	token, err := auth.GenerateWalletAuthToken(method, path, nil)
	if err != nil {
		t.Fatalf("failed to generate wallet auth token: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Parse token
	parsedToken, err := jwt.ParseSigned(token)
	if err != nil {
		t.Fatalf("failed to parse JWT: %v", err)
	}

	// Extract claims
	var claims APIKeyClaims
	if err := parsedToken.UnsafeClaimsWithoutVerification(&claims); err != nil {
		t.Fatalf("failed to extract claims: %v", err)
	}

	// Verify reqHash is empty when no body hash provided
	if claims.ReqHash != "" {
		t.Errorf("expected empty reqHash, got %s", claims.ReqHash)
	}
}

func TestGenerateBearerToken_DifferentTokensForDifferentPaths(t *testing.T) {
	auth, err := NewCDPAuth(
		"organizations/test-org/apiKeys/test-key",
		testECPrivateKey,
		"wallet-secret",
	)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	token1, err := auth.GenerateBearerToken("GET", "/platform/v2/evm/accounts")
	if err != nil {
		t.Fatalf("failed to generate token1: %v", err)
	}

	token2, err := auth.GenerateBearerToken("GET", "/platform/v2/solana/accounts")
	if err != nil {
		t.Fatalf("failed to generate token2: %v", err)
	}

	// Tokens should be different for different paths
	if token1 == token2 {
		t.Error("expected different tokens for different paths")
	}
}

func TestGenerateBearerToken_DifferentTokensForDifferentMethods(t *testing.T) {
	auth, err := NewCDPAuth(
		"organizations/test-org/apiKeys/test-key",
		testECPrivateKey,
		"wallet-secret",
	)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	token1, err := auth.GenerateBearerToken("GET", "/platform/v2/evm/accounts")
	if err != nil {
		t.Fatalf("failed to generate token1: %v", err)
	}

	token2, err := auth.GenerateBearerToken("POST", "/platform/v2/evm/accounts")
	if err != nil {
		t.Fatalf("failed to generate token2: %v", err)
	}

	// Tokens should be different for different methods
	if token1 == token2 {
		t.Error("expected different tokens for different methods")
	}
}

func TestCDPAuth_ThreadSafe(t *testing.T) {
	auth, err := NewCDPAuth(
		"organizations/test-org/apiKeys/test-key",
		testECPrivateKey,
		"wallet-secret",
	)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	// Test concurrent token generation
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			token, err := auth.GenerateBearerToken("GET", "/platform/v2/evm/accounts")
			if err != nil {
				t.Errorf("goroutine %d: failed to generate token: %v", idx, err)
			}
			if token == "" {
				t.Errorf("goroutine %d: expected non-empty token", idx)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestAPIKeyClaims_Structure(t *testing.T) {
	auth, err := NewCDPAuth(
		"organizations/test-org/apiKeys/test-key",
		testECPrivateKey,
		"wallet-secret",
	)
	if err != nil {
		t.Fatalf("failed to create auth: %v", err)
	}

	tests := []struct {
		name         string
		method       string
		path         string
		bodyHash     []byte
		checkReqHash bool
	}{
		{
			name:         "Bearer token without body hash",
			method:       "GET",
			path:         "/platform/v2/evm/accounts",
			bodyHash:     nil,
			checkReqHash: false,
		},
		{
			name:         "Wallet auth token with body hash",
			method:       "POST",
			path:         "/platform/v2/evm/accounts/test/sign",
			bodyHash:     []byte("test-hash-content"),
			checkReqHash: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var token string
			var err error

			if tt.checkReqHash {
				hash := sha256.Sum256(tt.bodyHash)
				token, err = auth.GenerateWalletAuthToken(tt.method, tt.path, hash[:])
			} else {
				token, err = auth.GenerateBearerToken(tt.method, tt.path)
			}

			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}

			// Parse and verify claims structure
			parsedToken, err := jwt.ParseSigned(token)
			if err != nil {
				t.Fatalf("failed to parse JWT: %v", err)
			}

			var claims APIKeyClaims
			if err := parsedToken.UnsafeClaimsWithoutVerification(&claims); err != nil {
				t.Fatalf("failed to extract claims: %v", err)
			}

			// Verify all standard claims are present
			if claims.Subject == "" {
				t.Error("expected Subject to be set")
			}
			if claims.Issuer == "" {
				t.Error("expected Issuer to be set")
			}
			if claims.Expiry == nil {
				t.Error("expected Expiry to be set")
			}
			if claims.NotBefore == nil {
				t.Error("expected NotBefore to be set")
			}
			if claims.URI == "" {
				t.Error("expected URI to be set")
			}

			// Verify reqHash presence based on test case
			if tt.checkReqHash {
				if claims.ReqHash == "" {
					t.Error("expected ReqHash to be set for wallet auth token")
				}
			} else {
				if claims.ReqHash != "" {
					t.Error("expected ReqHash to be empty for bearer token")
				}
			}
		})
	}
}
