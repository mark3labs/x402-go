# Quickstart: CDP Signer

Get started with Coinbase Developer Platform (CDP) signers for x402 payments in minutes.

---

## Prerequisites

### 1. CDP Account Setup

1. **Create CDP Project**: Visit https://portal.cdp.coinbase.com and create a new project
2. **Generate Secret API Key**:
   - Navigate to "API Keys" tab in your project
   - Click "Create API Key"
   - Select "Ed25519" algorithm (recommended for performance)
   - Save the **API Key Name** and **API Key Secret** (PEM format) immediately
   - Note: You won't be able to retrieve the secret again

3. **Generate Wallet Secret**:
   - Navigate to "Server Wallets" dashboard
   - Generate a new wallet secret
   - Save the **Wallet Secret** securely

### 2. Environment Configuration

Set required environment variables:

```bash
export CDP_API_KEY_NAME="organizations/your-org-id/apiKeys/your-key-id"
export CDP_API_KEY_SECRET="-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIBwZlpPvXJjwwWVXD0xvZa9rJ2hZyNxvRKjHU7qPLEe1oAoGCCqGSM49
AwEHoUQDQgAEi8X8hLT0HHfU8tqGfRLKsQZCYLMxY7vQP0D1E1hFbTuIGF3f6c/y
QR1wPqHkZlLzJ7vF8wJdP1RQhGxKLwNqbA==
-----END EC PRIVATE KEY-----"
export CDP_WALLET_SECRET="your-wallet-secret-here"
```

**Security Note**: Never commit these values to version control. Use a `.env` file (add to `.gitignore`) or a secrets manager like HashiCorp Vault for production.

---

## Installation

```bash
go get github.com/mark3labs/x402-go
go get github.com/mark3labs/x402-go/signers/coinbase
go get gopkg.in/square/go-jose.v2
```

---

## Quick Example: EVM (Base Sepolia)

```go
package main

import (
    "context"
    "log"
    "math/big"
    "os"
    "time"

    "github.com/mark3labs/x402-go"
    "github.com/mark3labs/x402-go/http"
    coinbase "github.com/mark3labs/x402-go/signers/coinbase"
)

func main() {
    // Initialize CDP signer for Base Sepolia testnet
    signer, err := coinbase.NewSigner(
        coinbase.WithCDPCredentials(
            os.Getenv("CDP_API_KEY_NAME"),
            os.Getenv("CDP_API_KEY_SECRET"),
            os.Getenv("CDP_WALLET_SECRET"),
        ),
        coinbase.WithNetwork("base-sepolia"),
        coinbase.WithToken("eth", "0x0000000000000000000000000000000000000000"),
        coinbase.WithMaxAmountPerCall(big.NewInt(1000000000000000000)), // 1 ETH max
    )
    if err != nil {
        log.Fatalf("Failed to initialize CDP signer: %v", err)
    }

    log.Printf("✓ CDP Signer initialized for Base Sepolia")
    log.Printf("  Address: %s", signer.Address())
    log.Printf("  Network: %s", signer.Network())

    // Create x402 HTTP client with CDP signer
    client, err := http.NewClient(
        http.WithSigner(signer),
    )
    if err != nil {
        log.Fatalf("Failed to create HTTP client: %v", err)
    }

    // Make x402-enabled HTTP request
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)
    if err != nil {
        log.Fatalf("Failed to create request: %v", err)
    }

    resp, err := client.Do(req)
    if err != nil {
        log.Fatalf("Request failed: %v", err)
    }
    defer resp.Body.Close()

    log.Printf("✓ Request successful: %s", resp.Status)
}
```

**Run:**
```bash
go run main.go
```

**Expected Output:**
```
✓ CDP Signer initialized for Base Sepolia
  Address: 0x742d35Cc6634C0532925a3b844Bc454e4438f44e
  Network: base-sepolia
✓ Request successful: 200 OK
```

---

## Quick Example: SVM (Solana Devnet)

```go
package main

import (
    "log"
    "math/big"
    "os"

    "github.com/mark3labs/x402-go/http"
    coinbase "github.com/mark3labs/x402-go/signers/coinbase"
)

func main() {
    // Initialize CDP signer for Solana Devnet
    signer, err := coinbase.NewSigner(
        coinbase.WithCDPCredentials(
            os.Getenv("CDP_API_KEY_NAME"),
            os.Getenv("CDP_API_KEY_SECRET"),
            os.Getenv("CDP_WALLET_SECRET"),
        ),
        coinbase.WithNetwork("solana-devnet"),
        coinbase.WithToken("sol", "So11111111111111111111111111111111111111112"),
        coinbase.WithMaxAmountPerCall(big.NewInt(1000000000)), // 1 SOL max
    )
    if err != nil {
        log.Fatalf("Failed to initialize CDP signer: %v", err)
    }

    log.Printf("✓ CDP Signer initialized for Solana Devnet")
    log.Printf("  Address: %s", signer.Address())
    log.Printf("  Network: %s", signer.Network())

    // Use with x402 HTTP client
    client, err := http.NewClient(
        http.WithSigner(signer),
    )
    if err != nil {
        log.Fatalf("Failed to create HTTP client: %v", err)
    }

    // Make requests as usual
    resp, err := client.Get("https://api.example.com/data")
    if err != nil {
        log.Fatalf("Request failed: %v", err)
    }
    defer resp.Body.Close()

    log.Printf("✓ Request successful: %s", resp.Status)
}
```

---

## Multi-Chain Example

Support both EVM and SVM payments in the same application:

```go
package main

import (
    "log"
    "math/big"
    "os"

    "github.com/mark3labs/x402-go/http"
    coinbase "github.com/mark3labs/x402-go/signers/coinbase"
)

func main() {
    // Load credentials once
    apiKeyName := os.Getenv("CDP_API_KEY_NAME")
    apiKeySecret := os.Getenv("CDP_API_KEY_SECRET")
    walletSecret := os.Getenv("CDP_WALLET_SECRET")

    // Create Base mainnet signer
    baseSigner, err := coinbase.NewSigner(
        coinbase.WithCDPCredentials(apiKeyName, apiKeySecret, walletSecret),
        coinbase.WithNetwork("base"),
        coinbase.WithToken("eth", "0x0000000000000000000000000000000000000000"),
        coinbase.WithToken("usdc", "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
        coinbase.WithPriority(0),
    )
    if err != nil {
        log.Fatalf("Failed to create Base signer: %v", err)
    }

    // Create Solana mainnet signer
    solanaSigner, err := coinbase.NewSigner(
        coinbase.WithCDPCredentials(apiKeyName, apiKeySecret, walletSecret),
        coinbase.WithNetwork("solana"),
        coinbase.WithToken("sol", "So11111111111111111111111111111111111111112"),
        coinbase.WithToken("usdc", "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"),
        coinbase.WithPriority(1),
    )
    if err != nil {
        log.Fatalf("Failed to create Solana signer: %v", err)
    }

    // Create HTTP client with both signers
    client, err := http.NewClient(
        http.WithSigner(baseSigner),
        http.WithSigner(solanaSigner),
    )
    if err != nil {
        log.Fatalf("Failed to create HTTP client: %v", err)
    }

    log.Printf("✓ Multi-chain client ready")
    log.Printf("  Base address: %s", baseSigner.Address())
    log.Printf("  Solana address: %s", solanaSigner.Address())

    // Client automatically selects appropriate signer based on
    // payment requirements from x402 server response
    resp, err := client.Get("https://api.example.com/data")
    if err != nil {
        log.Fatalf("Request failed: %v", err)
    }
    defer resp.Body.Close()

    log.Printf("✓ Request successful: %s", resp.Status)
}
```

---

## Advanced: Manual Account Creation

If you need to create accounts separately before initializing signers:

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    coinbase "github.com/mark3labs/x402-go/signers/coinbase"
)

func main() {
    // Create authentication handler
    auth := &coinbase.CDPAuth{
        APIKeyName:   os.Getenv("CDP_API_KEY_NAME"),
        APIKeySecret: os.Getenv("CDP_API_KEY_SECRET"),
        WalletSecret: os.Getenv("CDP_WALLET_SECRET"),
    }

    // Create or retrieve account for Base Sepolia
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    account, err := coinbase.CreateOrGetAccount(ctx, auth, "base-sepolia")
    if err != nil {
        log.Fatalf("Failed to create/get account: %v", err)
    }

    log.Printf("✓ Account ready")
    log.Printf("  ID: %s", account.ID)
    log.Printf("  Address: %s", account.Address)
    log.Printf("  Network: %s", account.Network)

    // Use account address for logging, monitoring, etc.
    // The NewSigner constructor automatically calls CreateOrGetAccount
    // so this is optional - shown here for demonstration
}
```

---

## Configuration Options

### Network Selection

**Supported EVM Networks:**
- `"base"` - Base mainnet (Chain ID: 8453)
- `"base-sepolia"` - Base testnet (Chain ID: 84532)
- `"ethereum"` - Ethereum mainnet (Chain ID: 1)
- `"sepolia"` - Ethereum testnet (Chain ID: 11155111)

**Supported SVM Networks:**
- `"solana"` or `"mainnet-beta"` - Solana mainnet
- `"solana-devnet"` or `"devnet"` - Solana devnet

### Token Configuration

**Native Tokens:**
```go
// ETH (native token, zero address)
WithToken("eth", "0x0000000000000000000000000000000000000000")

// SOL (native token, wrapped SOL mint)
WithToken("sol", "So11111111111111111111111111111111111111112")
```

**ERC-20 Tokens (Base mainnet):**
```go
// USDC on Base
WithToken("usdc", "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")

// DAI on Base
WithToken("dai", "0x50c5725949A6F0c72E6C4a641F24049A917DB0Cb")
```

**SPL Tokens (Solana):**
```go
// USDC on Solana
WithToken("usdc", "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
```

### Priority Configuration

**Signer Priority** (multiple signers):
```go
// Lower priority = selected first
WithPriority(0)  // Highest priority
WithPriority(1)  // Medium priority
WithPriority(2)  // Lower priority
```

**Token Priority** (within same signer):
```go
WithTokenPriority("usdc", "0x833...", 0)  // Prefer USDC
WithTokenPriority("eth", "0x000...", 1)   // Fallback to ETH
```

### Spending Limits

```go
// 1 ETH maximum per payment
WithMaxAmountPerCall(big.NewInt(1000000000000000000))

// 100 USDC maximum (6 decimals)
WithMaxAmountPerCall(big.NewInt(100000000))

// 10 SOL maximum (9 decimals)
WithMaxAmountPerCall(big.NewInt(10000000000))

// No limit
// (don't set WithMaxAmountPerCall)
```

---

## Error Handling

### Common Errors

**Invalid Credentials:**
```go
signer, err := coinbase.NewSigner(...)
if err != nil {
    if errors.Is(err, x402.ErrInvalidKey) {
        log.Fatal("Invalid CDP credentials - check API key and secret")
    }
}
```

**Unsupported Network:**
```go
signer, err := coinbase.NewSigner(
    coinbase.WithNetwork("polygon"), // Not yet supported
    ...
)
if err != nil {
    if errors.Is(err, x402.ErrInvalidNetwork) {
        log.Fatal("Network not supported by CDP signer")
    }
}
```

**Amount Exceeded:**
```go
payload, err := signer.Sign(requirement)
if err != nil {
    if errors.Is(err, x402.ErrAmountExceeded) {
        log.Printf("Payment amount exceeds configured maximum")
        // Handle gracefully (reject payment, prompt user, etc.)
    }
}
```

**CDP API Errors:**
```go
signer, err := coinbase.NewSigner(...)
if err != nil {
    if cdpErr, ok := err.(*coinbase.CDPError); ok {
        log.Printf("CDP API error: %s (status %d, request %s)",
            cdpErr.Message, cdpErr.StatusCode, cdpErr.RequestID)
        
        if cdpErr.Retryable {
            log.Printf("Error is retryable - will retry automatically")
        }
    }
}
```

---

## Testing

### Unit Tests with Mock Signer

```go
package myapp_test

import (
    "math/big"
    "testing"

    "github.com/mark3labs/x402-go"
)

// mockSigner implements x402.Signer for testing
type mockSigner struct {
    network   string
    canSign   bool
    signError error
}

func (m *mockSigner) Network() string { return m.network }
func (m *mockSigner) Scheme() string { return "exact" }
func (m *mockSigner) CanSign(*x402.PaymentRequirement) bool { return m.canSign }
func (m *mockSigner) Sign(*x402.PaymentRequirement) (*x402.PaymentPayload, error) {
    if m.signError != nil {
        return nil, m.signError
    }
    return &x402.PaymentPayload{X402Version: 1}, nil
}
func (m *mockSigner) GetPriority() int { return 0 }
func (m *mockSigner) GetTokens() []x402.TokenConfig { return nil }
func (m *mockSigner) GetMaxAmount() *big.Int { return nil }

func TestPaymentProcessing(t *testing.T) {
    mock := &mockSigner{network: "base", canSign: true}
    
    // Test your application logic with mock signer
    // ...
}
```

### Integration Tests with CDP Testnet

```go
package myapp_test

import (
    "os"
    "testing"

    coinbase "github.com/mark3labs/x402-go/signers/coinbase"
)

func TestCDPIntegration(t *testing.T) {
    // Skip if CDP credentials not configured
    if os.Getenv("CDP_API_KEY_NAME") == "" {
        t.Skip("Skipping integration test: CDP credentials not configured")
    }

    signer, err := coinbase.NewSigner(
        coinbase.WithCDPCredentials(
            os.Getenv("CDP_API_KEY_NAME"),
            os.Getenv("CDP_API_KEY_SECRET"),
            os.Getenv("CDP_WALLET_SECRET"),
        ),
        coinbase.WithNetwork("base-sepolia"),
        coinbase.WithToken("eth", "0x0000000000000000000000000000000000000000"),
    )
    if err != nil {
        t.Fatalf("Failed to create signer: %v", err)
    }

    // Test signing operation
    payload, err := signer.Sign(&x402.PaymentRequirement{
        Network: "base-sepolia",
        Scheme:  "exact",
        Token:   "eth",
        Amount:  "1000000000000000", // 0.001 ETH
    })
    if err != nil {
        t.Fatalf("Failed to sign: %v", err)
    }

    if payload.X402Version != 1 {
        t.Errorf("Expected version 1, got %d", payload.X402Version)
    }
}
```

**Run integration tests:**
```bash
# Set credentials
export CDP_API_KEY_NAME="..."
export CDP_API_KEY_SECRET="..."
export CDP_WALLET_SECRET="..."

# Run tests
go test -v -run Integration
```

---

## Troubleshooting

### Issue: "invalid CDP credentials"

**Cause**: API key or secret is invalid or incorrectly formatted

**Solution**:
1. Verify API key name matches format: `organizations/xxx/apiKeys/yyy`
2. Verify API key secret is PEM-encoded and includes headers:
   ```
   -----BEGIN EC PRIVATE KEY-----
   ...
   -----END EC PRIVATE KEY-----
   ```
3. Ensure no extra whitespace or line breaks
4. Regenerate credentials from CDP Portal if needed

### Issue: "rate limit exceeded"

**Cause**: Exceeded CDP API rate limits (600 reads/500 writes per 10 seconds)

**Solution**:
- CDP signer automatically retries with exponential backoff
- If persistent, reduce request rate in your application
- Consider implementing request batching
- Check for excessive concurrent requests

### Issue: "account creation fails"

**Cause**: Network error, invalid network, or insufficient permissions

**Solution**:
1. Check network connectivity to api.cdp.coinbase.com
2. Verify network identifier is supported
3. Ensure API key has account creation permissions
4. Check CDP Portal for service status

### Issue: "signature verification fails"

**Cause**: Clock skew or JWT token expiration

**Solution**:
1. Ensure system clock is synchronized (use NTP)
2. JWT tokens are generated fresh per request (no caching)
3. Verify network latency is reasonable (<2 minutes)
4. Check wallet secret is correct

---

## Next Steps

- **Production Deployment**: Set up secrets management (Vault, AWS Secrets Manager)
- **Monitoring**: Track CDP API errors and response times
- **Key Rotation**: Establish quarterly rotation schedule for API keys
- **Multi-Region**: Consider CDP availability across regions
- **Documentation**: Review full spec at [spec.md](./spec.md)

---

## Resources

- **CDP Portal**: https://portal.cdp.coinbase.com
- **CDP API Docs**: https://docs.cdp.coinbase.com/api-reference/v2
- **x402-go Docs**: https://github.com/mark3labs/x402-go
- **Support**: Open issue on GitHub for CDP signer questions

---

**Ready to build!** 🚀

Start with Base Sepolia or Solana Devnet for testing, then move to mainnet when ready.
