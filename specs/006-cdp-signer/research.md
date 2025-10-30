# Research: CDP Signer Implementation

**Date**: 2025-10-30  
**Feature**: 006-cdp-signer  
**Objective**: Resolve technical unknowns and document technology decisions

---

## Decision: JWT Library Selection

**Chosen**: `gopkg.in/square/go-jose.v2` v2.6.0

**Rationale**:
- CDP requires ES256 (ECDSA P-256) algorithm for JWT signing
- go-jose provides native ES256 support with simple API
- Mature, stable library (v2.6.0, latest v2.x release, production-ready since 2022)
- Supports custom JWT claims required by CDP (sub, iss, nbf, exp, uri, reqHash)
- Used by major projects (OAuth2 servers, authentication systems)
- Well-documented with clear examples for ECDSA JWT signing

**Alternatives Considered**:
1. **golang-jwt/jwt (v5)**: 
   - Rejected: Complex key handling for ECDSA
   - Limited examples for ES256 with custom claims
   - Less ergonomic API for JWT signing with PEM keys
   
2. **Standard library crypto package**:
   - Rejected: No high-level JWT support
   - Would require manual implementation of JWT encoding, signing, base64url encoding
   - Increases code complexity and maintenance burden

3. **lestrrat-go/jwx**:
   - Considered: Good ES256 support
   - Rejected: Larger dependency footprint, more complex API than needed

**Testing Evidence**: 
- Successfully generated ES256 JWT matching CDP requirements in scratch/cdp.md:23-107
- Token structure validated against CDP API documentation
- PEM key parsing works with Ed25519 and ECDSA keys

**Dependencies Added**: `gopkg.in/square/go-jose.v2` (MIT license)

---

## Decision: Account Creation Strategy

**Chosen**: Idempotent CreateOrGetAccount helper with GET-then-POST pattern

**Rationale**:
- CDP accounts cannot be created through portal (API-only creation per spec:242)
- Must handle first-time initialization (create) and subsequent initializations (retrieve)
- Idempotent operation prevents duplicate accounts during concurrent initialization
- Simple two-step flow: attempt GET (list accounts), if empty then POST (create)

**Implementation Pattern**:
```go
func CreateOrGetAccount(ctx context.Context, auth *CDPAuth, network string) (*CDPAccount, error) {
    // Step 1: GET /platform/v2/{chain_type}/accounts
    // Returns list of existing accounts for credentials + network
    
    // Step 2: If list non-empty, return first account
    // CDP API guarantees account uniqueness per credentials + network
    
    // Step 3: If list empty, POST /platform/v2/{chain_type}/accounts
    // Creates new account and returns account details
    
    // Result: Always returns valid account, create or retrieve
}
```

**Race Condition Handling**:
- If concurrent calls both see empty list and attempt create, CDP API handles deduplication
- Based on CDP documentation, API is idempotent for account creation with same credentials
- Worst case: Second create returns existing account or 409 Conflict (treat as success)

**Network-Specific Endpoints** (from CDP API docs):
- EVM: `POST /platform/v2/evm/accounts` with `{"network_id": "base-sepolia"}`
- SVM: `POST /platform/v2/solana/accounts` with `{"network_id": "solana-devnet"}`

**Alternatives Considered**:
1. **Create-only (error if exists)**:
   - Rejected: Requires manual account management, poor DX
   
2. **External account ID tracking**:
   - Rejected: Adds storage dependency, violates stateless design
   
3. **Try create, catch duplicate error**:
   - Rejected: Makes error case the happy path, poor performance

---

## Decision: Network Identifier Mapping

**Chosen**: Bidirectional map between x402 network names and CDP network IDs

**Rationale**:
- x402 uses short network names: "base", "base-sepolia", "ethereum", "sepolia", "solana", "solana-devnet"
- CDP uses longer descriptive IDs: "base-mainnet", "base-sepolia", "ethereum", "sepolia", "solana-mainnet", "solana-devnet"
- Some networks match exactly (ethereum, sepolia, base-sepolia, solana-devnet)
- Some need mapping: "base" → "base-mainnet", "solana" / "mainnet-beta" → "solana-mainnet"

**Mapping Table**:
| x402 Network | CDP Network ID | Chain Type |
|--------------|----------------|------------|
| base | base-mainnet | EVM |
| base-sepolia | base-sepolia | EVM |
| ethereum | ethereum | EVM |
| sepolia | sepolia | EVM |
| solana | solana-mainnet | SVM |
| mainnet-beta | solana-mainnet | SVM |
| solana-devnet | solana-devnet | SVM |
| devnet | solana-devnet | SVM |

**Implementation**:
```go
// networks.go
var networkToCDP = map[string]string{
    "base":          "base-mainnet",
    "base-sepolia":  "base-sepolia",
    "ethereum":      "ethereum",
    "sepolia":       "sepolia",
    "solana":        "solana-mainnet",
    "mainnet-beta":  "solana-mainnet",
    "solana-devnet": "solana-devnet",
    "devnet":        "solana-devnet",
}

func getCDPNetwork(x402Network string) (string, error) {
    cdpNet, ok := networkToCDP[x402Network]
    if !ok {
        return "", x402.ErrInvalidNetwork
    }
    return cdpNet, nil
}

func getNetworkType(x402Network string) NetworkType {
    switch x402Network {
    case "base", "base-sepolia", "ethereum", "sepolia":
        return NetworkTypeEVM
    case "solana", "mainnet-beta", "solana-devnet", "devnet":
        return NetworkTypeSVM
    default:
        return NetworkTypeUnknown
    }
}
```

**Chain ID Mapping** (EVM only, reuse from evm/signer.go:249-263):
```go
func getChainID(network string) (*big.Int, error) {
    switch network {
    case "base":
        return big.NewInt(8453), nil
    case "base-sepolia":
        return big.NewInt(84532), nil
    case "ethereum":
        return big.NewInt(1), nil
    case "sepolia":
        return big.NewInt(11155111), nil
    default:
        return nil, x402.ErrInvalidNetwork
    }
}
```

---

## Decision: Error Classification and Retry Strategy

**Chosen**: Type-based error classification with exponential backoff for retryable errors

**Rationale**:
- CDP API returns standard HTTP status codes with structured error responses
- Some errors are transient (5xx, 429) and should be retried
- Some errors are permanent (4xx auth/validation) and should fail immediately
- Exponential backoff prevents thundering herd and respects rate limits

**Error Categories**:

### Retryable Errors (retry with backoff)
- **429 Too Many Requests**: Rate limit exceeded
  - Retry with exponential backoff
  - Respect Retry-After header if present
  - Continue until success or max attempts (5)
  
- **5xx Server Errors**: CDP service issues
  - 500 Internal Server Error
  - 502 Bad Gateway
  - 503 Service Unavailable
  - 504 Gateway Timeout
  - Retry with exponential backoff (5 attempts)

- **Network Errors**: Connection failures, timeouts
  - DNS resolution failures
  - Connection refused
  - Request timeout
  - Retry with exponential backoff (5 attempts)

### Non-Retryable Errors (fail immediately)
- **401 Unauthorized**: Invalid credentials
  - Return x402.ErrInvalidKey with context
  - Log error (sanitized, no credentials)
  - No retry
  
- **403 Forbidden**: Insufficient permissions
  - Return descriptive error
  - No retry
  
- **4xx Client Errors** (except 429):
  - 400 Bad Request: Invalid request format
  - 404 Not Found: Resource doesn't exist
  - 422 Unprocessable Entity: Validation failure
  - Return descriptive error with CDP error message
  - No retry

**Retry Configuration** (from spec:013):
```go
type RetryConfig struct {
    MaxAttempts  int           // 5 attempts
    InitialDelay time.Duration // 100ms
    MaxDelay     time.Duration // 10s total
    Multiplier   float64       // 2x
}
```

**Backoff Calculation**:
```go
func calculateBackoff(attempt int, cfg RetryConfig, retryAfter time.Duration) time.Duration {
    // Use Retry-After header if present (rate limit case)
    if retryAfter > 0 {
        return retryAfter
    }
    
    // Exponential backoff: initialDelay * (multiplier ^ attempt)
    delay := cfg.InitialDelay * time.Duration(math.Pow(cfg.Multiplier, float64(attempt)))
    
    // Cap at 10s total
    if delay > cfg.MaxDelay {
        delay = cfg.MaxDelay
    }
    
    // Add jitter (±25%) to prevent thundering herd
    jitter := time.Duration(rand.Int63n(int64(delay) / 2))
    return delay + jitter - (delay / 4)
}
```

**Error Struct**:
```go
type CDPError struct {
    StatusCode int
    ErrorType  string // "rate_limit", "server_error", "auth_error", "client_error"
    Message    string
    RequestID  string
    Retryable  bool
    RetryAfter time.Duration
}
```

**Alternatives Considered**:
1. **Simple retry without backoff**: Rejected, would hammer API during outages
2. **Linear backoff**: Rejected, grows too slowly for rate limiting
3. **Circuit breaker**: Rejected, adds complexity beyond requirements

---

## Decision: Signing Operation Flow

**Chosen**: Direct CDP API calls for transaction signing (no local signing)

**Rationale**:
- CDP manages private keys in TEE (Trusted Execution Environment)
- All signing happens server-side via CDP API
- Signer delegates to CDP for cryptographic operations
- Follows scratch/cdp.md architecture (lines 261-266, 458-570)

### EVM Signing Flow

**Endpoint**: `POST /platform/v2/evm/accounts/{address}/sign/typed-data`

**Input** (EIP-712 typed data from x402 EIP-3009 authorization):
```go
type SignTypedDataRequest struct {
    TypedData EIP712TypedData `json:"typedData"`
}

// Build from x402 payment requirements
// Reuse EIP-3009 structure from evm/eip3009.go:18-51
```

**Output**:
```go
type SignResponse struct {
    Signature string `json:"signature"` // Hex-encoded ECDSA signature
}
```

**Process**:
1. Validate payment requirements via CanSign()
2. Parse amount to *big.Int
3. Build EIP-3009 authorization struct (from/to/value/validAfter/validBefore/nonce)
4. Convert to EIP-712 typed data
5. Call CDP sign endpoint with Wallet Auth JWT
6. Construct x402.PaymentPayload with signature

### SVM Signing Flow

**Endpoint**: `POST /platform/v2/solana/accounts/{address}/sign/transaction`

**Input** (Solana transaction message):
```go
type SignTransactionRequest struct {
    Message string `json:"message"` // Base64-encoded transaction message
}

// Build from x402 payment requirements
// Reuse transaction building logic from svm/signer.go:322-396
```

**Output**:
```go
type SignResponse struct {
    Signature string `json:"signature"` // Base64-encoded Ed25519 signature
}
```

**Process**:
1. Validate payment requirements via CanSign()
2. Parse amount to *big.Int
3. Build Solana transaction (compute budget + transfer instructions)
4. Serialize transaction message to base64
5. Call CDP sign endpoint with Wallet Auth JWT
6. Construct x402.PaymentPayload with signed transaction

**JWT Requirements for Signing** (from spec:009-010):
- Bearer Token: Required (standard API authentication)
- Wallet Auth Token: Required (additional layer for signing operations)
- Both tokens generated fresh per request (no caching)
- Wallet Auth includes reqHash (SHA-256 of request body)

---

## Decision: Credential Management

**Chosen**: Environment variable loading with validation, no caching

**Rationale**:
- Environment variables are standard for 12-factor apps
- Secrets management systems (Vault, AWS SM) inject via env vars
- Keeps signer stateless and simple
- Validation at initialization prevents runtime errors

**Environment Variables** (from spec:195):
```bash
CDP_API_KEY_NAME="organizations/xxx/apiKeys/yyy"
CDP_API_KEY_SECRET="-----BEGIN EC PRIVATE KEY-----\n...\n-----END EC PRIVATE KEY-----"
CDP_WALLET_SECRET="your-wallet-secret-here"
```

**Loading Pattern**:
```go
func loadCredentialsFromEnv() (apiKeyName, apiKeySecret, walletSecret string, err error) {
    apiKeyName = os.Getenv("CDP_API_KEY_NAME")
    apiKeySecret = os.Getenv("CDP_API_KEY_SECRET")
    walletSecret = os.Getenv("CDP_WALLET_SECRET")
    
    if apiKeyName == "" {
        return "", "", "", fmt.Errorf("CDP_API_KEY_NAME environment variable not set")
    }
    if apiKeySecret == "" {
        return "", "", "", fmt.Errorf("CDP_API_KEY_SECRET environment variable not set")
    }
    // walletSecret optional for account creation, required for signing
    
    return apiKeyName, apiKeySecret, walletSecret, nil
}
```

**Validation**:
- PEM key parsing during CDPAuth initialization
- Rejects invalid PEM format immediately
- Returns x402.ErrInvalidKey with descriptive message

**Security**:
- Never log credentials or JWT tokens
- Sanitize authorization headers before logging
- No credential caching (stateless signer)

**Alternatives Considered**:
1. **Config file**: Rejected, insecure for credentials
2. **Direct parameter passing**: Rejected, risk of hardcoding
3. **Vault integration**: Out of scope, developers handle externally

---

## Decision: Concurrent Request Handling

**Chosen**: Stateless signer with per-request context, no shared mutable state

**Rationale**:
- Signer is immutable after initialization
- Each Sign() call operates independently
- CDP API handles concurrency (600 reads/500 writes per 10s)
- No need for internal synchronization or locking

**Thread Safety**:
- Signer struct is read-only after NewSigner() returns
- HTTP client is thread-safe (http.Client supports concurrent use)
- JWT generation is stateless (fresh token per request)
- No race conditions (verified with `go test -race`)

**Performance**:
- No mutex contention (no locks needed)
- Parallel requests limited only by CDP rate limits
- Context propagation for timeout/cancellation
- Request-scoped retries (no global state)

**Testing**:
```go
func TestConcurrentSigning(t *testing.T) {
    signer := createTestSigner(t)
    
    const numRequests = 100
    var wg sync.WaitGroup
    errors := make(chan error, numRequests)
    
    for i := 0; i < numRequests; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _, err := signer.Sign(testRequirement)
            if err != nil {
                errors <- err
            }
        }()
    }
    
    wg.Wait()
    close(errors)
    
    for err := range errors {
        t.Errorf("concurrent signing failed: %v", err)
    }
}
```

---

## Decision: Test Strategy

**Chosen**: Table-driven unit tests + optional integration tests with coverage >80%

**Rationale**:
- Table-driven tests provide comprehensive scenario coverage
- Mock CDP API responses for unit tests (fast, deterministic)
- Optional integration tests against CDP testnet (slow, requires credentials)
- Follows existing test patterns from evm/signer_test.go and svm/signer_test.go

**Test Categories**:

### 1. Unit Tests (mocked CDP API)
- **Constructor tests**: Valid configs, missing fields, invalid credentials
- **Interface method tests**: All getter methods return correct values
- **CanSign tests**: Network/token matching, case sensitivity
- **Sign tests**: Valid requests, amount limits, errors
- **JWT generation tests**: Token structure, expiration, claims
- **Error classification tests**: Retryable vs non-retryable
- **Retry logic tests**: Backoff calculation, max attempts
- **Network mapping tests**: x402 → CDP conversion

### 2. Integration Tests (real CDP API, optional)
- **Account creation**: First-time create, subsequent retrieve
- **EVM signing**: Base Sepolia testnet transaction
- **SVM signing**: Solana Devnet transaction
- **Rate limit handling**: Trigger 429, verify backoff

**Integration Test Skip Pattern** (from svm/signer_test.go:343):
```go
func TestCDPIntegration(t *testing.T) {
    if os.Getenv("CDP_API_KEY_NAME") == "" {
        t.Skip("Skipping integration test: CDP credentials not configured")
    }
    
    // Integration test using real CDP API
}
```

**Coverage Target**: >80% (maintain existing project coverage per constitution:030)

**Test Execution**:
```bash
# Unit tests (fast, always run)
go test -race -cover ./signers/coinbase/

# Integration tests (slow, opt-in)
CDP_API_KEY_NAME=xxx CDP_API_KEY_SECRET=yyy go test -race -v ./signers/coinbase/ -run Integration
```

---

## Best Practices Findings

### From CDP Documentation (scratch/cdp.md)

1. **JWT Token Freshness** (lines 89-94):
   - Generate fresh JWT for every request
   - No caching or refresh logic
   - 2-minute expiration for Bearer tokens
   - 1-minute expiration for Wallet Auth tokens
   - Prevents token expiration edge cases

2. **Request Hash for Signing** (lines 80-85):
   - POST/PUT requests include SHA-256 hash of body in JWT
   - Prevents request tampering
   - Required for wallet authentication

3. **Error Logging** (lines 786-818):
   - Sanitize authorization headers
   - Never log JWT tokens or credentials
   - Use [REDACTED] placeholder for sensitive data

4. **HTTP Client Configuration** (lines 171-178):
   - 30-second timeout
   - Connection pooling (100 max idle, 10 per host)
   - 90-second idle timeout
   - TLS verification enabled (never skip in production)

### From Existing Signers (evm/signer.go, svm/signer.go)

1. **Functional Options Pattern**:
   - Constructor takes variadic options
   - Each option is a function that modifies signer
   - Allows flexible configuration
   - Example: evm/signer.go:28-59

2. **Validation in Constructor**:
   - Check required fields immediately
   - Return descriptive errors
   - Fail fast (don't defer validation to Sign())
   - Example: evm/signer.go:40-48

3. **CanSign Logic**:
   - Network exact match
   - Scheme exact match (always "exact")
   - Token case-insensitive match
   - Example: evm/signer.go:142-161

4. **Amount Handling**:
   - Parse string amount to *big.Int
   - Check against maxAmount limit
   - Return ErrAmountExceeded if over limit
   - Example: evm/signer.go:171-179

5. **Chain ID Mapping**:
   - Simple switch statement
   - Returns ErrInvalidNetwork for unsupported
   - Example: evm/signer.go:249-263

---

## Integration Patterns

### With Existing x402 Middleware

CDP signer integrates seamlessly with no code changes:

```go
// http/client.go usage
client, _ := http.NewClient(
    http.WithSigner(evmSigner),    // Local EVM signer
    http.WithSigner(svmSigner),    // Local SVM signer
    http.WithSigner(cdpSigner),    // CDP signer
)
```

Selector logic (selector.go:68-133) works identically:
1. Calls `CanSign()` for each signer
2. Checks `GetMaxAmount()` limit
3. Accesses `GetTokens()` for priority
4. Gets `GetPriority()` for sorting
5. Calls `Sign()` on selected signer

No changes needed to selector, transport, or middleware layers.

### With Payment Requirements

Payment requirements flow from middleware → selector → signer:

```go
// From http/middleware.go
requirement := &x402.PaymentRequirement{
    Network: "base",
    Scheme:  "exact",
    Token:   "eth",
    Amount:  "1000000000000000", // 0.001 ETH
}

// CDP signer handles identically to local signers
payload, err := cdpSigner.Sign(requirement)
```

---

## Open Questions Resolved

✅ **JWT token caching**: Generate fresh per request (no caching)  
✅ **Account creation idempotency**: GET-then-POST pattern  
✅ **Rate limiting approach**: Rely on CDP 429 responses + exponential backoff  
✅ **Duplicate signature prevention**: Not needed (blockchain nonce prevents double-spend)  
✅ **Network identifier mapping**: Bidirectional map (x402 ↔ CDP)  
✅ **Error retry strategy**: Type-based classification (retryable vs non-retryable)  
✅ **Test approach**: Table-driven unit tests + optional integration tests  
✅ **Concurrent request handling**: Stateless signer, no synchronization needed

---

## Summary

All technical unknowns have been resolved. Key decisions:

1. **JWT Library**: gopkg.in/square/go-jose.v2 (ES256 support)
2. **Account Management**: Idempotent CreateOrGetAccount helper
3. **Network Mapping**: Bidirectional x402 ↔ CDP network map
4. **Error Handling**: Type-based classification with exponential backoff
5. **Signing Flow**: Direct CDP API calls (EVM: typed-data, SVM: transaction)
6. **Credentials**: Environment variables with validation
7. **Concurrency**: Stateless signer, thread-safe
8. **Testing**: Table-driven unit tests + optional integration tests

Ready for Phase 1 (Design & Contracts).
