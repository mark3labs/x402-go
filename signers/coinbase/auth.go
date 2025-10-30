package coinbase

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// CDPAuth manages CDP API authentication credentials and JWT token generation.
// It handles parsing of PEM-encoded private keys and generates both Bearer tokens
// for standard API authentication and Wallet Authentication tokens for sensitive
// signing operations.
//
// CDPAuth is immutable after construction and thread-safe for concurrent use.
// The parsed private key is cached internally to avoid repeated parsing overhead.
//
// Example usage:
//
//	auth, err := NewCDPAuth(
//	    "organizations/abc/apiKeys/xyz",
//	    "-----BEGIN EC PRIVATE KEY-----\n...\n-----END EC PRIVATE KEY-----",
//	    "wallet-secret-123",
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	token, err := auth.GenerateBearerToken("GET", "/platform/v2/evm/accounts")
//	if err != nil {
//	    log.Fatal(err)
//	}
type CDPAuth struct {
	// apiKeyName is the CDP API key identifier (e.g., "organizations/xxx/apiKeys/yyy")
	apiKeyName string

	// apiKeySecret is the PEM-encoded ECDSA or Ed25519 private key
	apiKeySecret string

	// walletSecret is the wallet-specific secret for signing operations (optional)
	walletSecret string

	// privateKey is the parsed private key, cached to avoid repeated parsing
	privateKey interface{}
}

// APIKeyClaims represents the JWT claims structure required by CDP API.
// It extends the standard JWT claims with CDP-specific fields for request
// authentication and integrity verification.
type APIKeyClaims struct {
	*jwt.Claims
	// URI is the full request URI in format: "{METHOD} api.cdp.coinbase.com{path}"
	URI string `json:"uri"`
	// ReqHash is the hex-encoded SHA-256 hash of the request body (optional)
	ReqHash string `json:"reqHash,omitempty"`
}

// NewCDPAuth creates a new CDPAuth instance with the provided credentials.
// It validates the API key name and parses the PEM-encoded private key,
// caching the parsed key for efficient token generation.
//
// Parameters:
//   - apiKeyName: CDP API key identifier (required, must not be empty)
//   - apiKeySecret: PEM-encoded ECDSA or Ed25519 private key (required)
//   - walletSecret: Wallet-specific secret for signing operations (optional)
//
// Returns:
//   - *CDPAuth: Configured authentication instance
//   - error: Validation or parsing errors
//
// Errors:
//   - Returns error if apiKeyName is empty
//   - Returns error if apiKeySecret is not valid PEM format
//   - Returns error if private key parsing fails
//
// Example:
//
//	auth, err := NewCDPAuth(
//	    os.Getenv("CDP_API_KEY_NAME"),
//	    os.Getenv("CDP_API_KEY_SECRET"),
//	    os.Getenv("CDP_WALLET_SECRET"),
//	)
func NewCDPAuth(apiKeyName, apiKeySecret, walletSecret string) (*CDPAuth, error) {
	// Validate API key name
	if apiKeyName == "" {
		return nil, fmt.Errorf("apiKeyName must not be empty")
	}

	// Parse PEM-encoded private key
	block, _ := pem.Decode([]byte(apiKeySecret))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block: invalid PEM format")
	}

	// Try parsing as ECDSA key first (most common for CDP)
	var privateKey interface{}
	var err error

	privateKey, err = x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		// Try parsing as PKCS8 format (supports both ECDSA and Ed25519)
		privateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	// Validate key type
	switch privateKey.(type) {
	case *ecdsa.PrivateKey:
		// ECDSA is valid
	case crypto.Signer:
		// Ed25519 implements crypto.Signer
	default:
		return nil, fmt.Errorf("unsupported private key type: must be ECDSA or Ed25519")
	}

	return &CDPAuth{
		apiKeyName:   apiKeyName,
		apiKeySecret: apiKeySecret,
		walletSecret: walletSecret,
		privateKey:   privateKey,
	}, nil
}

// GenerateBearerToken generates a JWT Bearer token for standard CDP API authentication.
// The token is valid for 2 minutes and includes claims for request identification
// and authorization.
//
// Parameters:
//   - method: HTTP method (e.g., "GET", "POST", "PUT", "DELETE")
//   - path: API endpoint path (e.g., "/platform/v2/evm/accounts")
//
// Returns:
//   - string: Signed JWT token for Authorization header
//   - error: Token generation or signing errors
//
// The generated JWT includes the following claims:
//   - sub: API key name
//   - iss: "coinbase-cloud"
//   - nbf: Current timestamp (not before)
//   - exp: Current timestamp + 2 minutes (expiration)
//   - uri: "{method} api.cdp.coinbase.com{path}"
//
// Example:
//
//	token, err := auth.GenerateBearerToken("GET", "/platform/v2/evm/accounts")
//	if err != nil {
//	    return err
//	}
//	req.Header.Set("Authorization", "Bearer "+token)
func (a *CDPAuth) GenerateBearerToken(method, path string) (string, error) {
	return a.generateJWT(method, path, nil, 2*time.Minute)
}

// GenerateWalletAuthToken generates a JWT Wallet Authentication token for sensitive
// signing operations. The token is valid for 1 minute and includes a hash of the
// request body for integrity verification.
//
// Parameters:
//   - method: HTTP method (typically "POST" for signing operations)
//   - path: API endpoint path
//   - bodyHash: SHA-256 hash of the request body bytes
//
// Returns:
//   - string: Signed JWT token for X-Wallet-Auth header
//   - error: Token generation or signing errors
//
// The generated JWT includes all Bearer token claims plus:
//   - reqHash: Hex-encoded SHA-256 hash of request body
//
// Example:
//
//	bodyBytes, _ := json.Marshal(signRequest)
//	hash := sha256.Sum256(bodyBytes)
//	token, err := auth.GenerateWalletAuthToken("POST", "/platform/v2/evm/accounts/0x.../sign/typed-data", hash[:])
//	if err != nil {
//	    return err
//	}
//	req.Header.Set("X-Wallet-Auth", token)
func (a *CDPAuth) GenerateWalletAuthToken(method, path string, bodyHash []byte) (string, error) {
	return a.generateJWT(method, path, bodyHash, 1*time.Minute)
}

// generateJWT is the internal JWT generation implementation shared by both
// Bearer and Wallet Auth token generation methods.
func (a *CDPAuth) generateJWT(method, path string, bodyHash []byte, expiration time.Duration) (string, error) {
	// Determine signing algorithm based on key type
	var alg jose.SignatureAlgorithm
	switch a.privateKey.(type) {
	case *ecdsa.PrivateKey:
		alg = jose.ES256
	default:
		// Ed25519 or other crypto.Signer
		alg = jose.EdDSA
	}

	// Create signer with key ID header
	sig, err := jose.NewSigner(
		jose.SigningKey{Algorithm: alg, Key: a.privateKey},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", a.apiKeyName),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create JWT signer: %w", err)
	}

	// Build URI claim
	uri := fmt.Sprintf("%s api.cdp.coinbase.com%s", method, path)

	// Calculate request hash if body provided
	var reqHash string
	if len(bodyHash) > 0 {
		reqHash = hex.EncodeToString(bodyHash)
	}

	// Create claims
	now := time.Now()
	claims := &APIKeyClaims{
		Claims: &jwt.Claims{
			Subject:   a.apiKeyName,
			Issuer:    "coinbase-cloud",
			NotBefore: jwt.NewNumericDate(now),
			Expiry:    jwt.NewNumericDate(now.Add(expiration)),
		},
		URI:     uri,
		ReqHash: reqHash,
	}

	// Sign and serialize JWT
	token, err := jwt.Signed(sig).Claims(claims).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize JWT: %w", err)
	}

	return token, nil
}
