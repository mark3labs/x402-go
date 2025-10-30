# Data Model: CDP Signer

**Feature**: 006-cdp-signer  
**Date**: 2025-10-30

---

## Entity Overview

```
┌─────────────┐       uses        ┌────────────┐       calls       ┌──────────────┐
│   Signer    │ ─────────────────> │ CDPClient  │ ───────────────> │  CDP API     │
│             │                    │            │                   │ (External)   │
│ - network   │                    │ - baseURL  │                   │              │
│ - tokens    │                    │ - auth     │                   │              │
│ - cdpClient │                    └────────────┘                   └──────────────┘
│             │                           │
└─────────────┘                           │ uses
                                          ▼
                                   ┌────────────┐
                                   │  CDPAuth   │
                                   │            │
                                   │ - apiKey   │
                                   │ - secret   │
                                   └────────────┘
                                          │
                                          │ generates
                                          ▼
                                   ┌────────────┐
                                   │ JWT Tokens │
                                   │            │
                                   │ - Bearer   │
                                   │ - WalletAuth│
                                   └────────────┘
```

---

## Entity: CDPAuth

**Purpose**: Manages CDP API authentication credentials and JWT token generation for API requests

**Fields**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiKeyName` | string | Yes | CDP API key identifier (e.g., "organizations/xxx/apiKeys/yyy") |
| `apiKeySecret` | string | Yes | PEM-encoded ECDSA or Ed25519 private key for signing JWTs |
| `walletSecret` | string | Conditional | Wallet-specific secret for signing operations (optional for account creation, required for signing) |

**Methods**:

### `GenerateBearerToken(method, path string) (string, error)`
Generates a 2-minute JWT Bearer token for standard CDP API authentication.

**Parameters**:
- `method`: HTTP method (GET, POST, PUT, DELETE)
- `path`: API path (e.g., "/platform/v2/evm/accounts")

**Returns**: JWT string with claims:
- `sub`: apiKeyName
- `iss`: "coinbase-cloud"
- `nbf`: current timestamp
- `exp`: current timestamp + 2 minutes
- `uri`: "{method} api.cdp.coinbase.com{path}"

**Errors**:
- Invalid PEM format in apiKeySecret
- Failed to parse private key
- JWT signing failure

### `GenerateWalletAuthToken(method, path string, bodyHash []byte) (string, error)`
Generates a 1-minute JWT Wallet Authentication token for transaction signing operations.

**Parameters**:
- `method`: HTTP method (typically POST)
- `path`: API path
- `bodyHash`: SHA-256 hash of request body bytes

**Returns**: JWT string with additional claim:
- `reqHash`: hex-encoded SHA-256 hash of request body

**Errors**: Same as GenerateBearerToken

**Validation Rules**:
- `apiKeyName` must not be empty
- `apiKeySecret` must be valid PEM-encoded ECDSA or Ed25519 key
- `walletSecret` validated only when GenerateWalletAuthToken() is called

**State Transitions**: Immutable (all fields set during construction, never modified)

**Example**:
```go
auth := &CDPAuth{
    apiKeyName:   "organizations/abc/apiKeys/xyz",
    apiKeySecret: "-----BEGIN EC PRIVATE KEY-----\n...\n-----END EC PRIVATE KEY-----",
    walletSecret: "wallet-secret-123",
}

token, err := auth.GenerateBearerToken("GET", "/platform/v2/evm/accounts")
// Returns: "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## Entity: CDPClient

**Purpose**: HTTP client wrapper for CDP REST API communication with automatic authentication, retry logic, and error handling

**Fields**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `baseURL` | string | Yes | CDP API base URL (https://api.cdp.coinbase.com) |
| `httpClient` | *http.Client | Yes | Configured HTTP client with timeouts and connection pooling |
| `auth` | *CDPAuth | Yes | Authentication handler for JWT generation |

**Methods**:

### `doRequest(ctx context.Context, method, path string, body, result interface{}, requireWalletAuth bool) error`
Executes a single HTTP request to CDP API with authentication headers.

**Parameters**:
- `ctx`: Request context for timeout/cancellation
- `method`: HTTP method
- `path`: API endpoint path
- `body`: Request body (marshaled to JSON), can be nil
- `result`: Response struct (unmarshaled from JSON), can be nil
- `requireWalletAuth`: Whether to include X-Wallet-Auth header

**Behavior**:
1. Marshal request body to JSON (if non-nil)
2. Generate Bearer JWT token
3. Generate Wallet Auth JWT token (if requireWalletAuth=true)
4. Create HTTP request with context
5. Set headers: Content-Type, Accept, Authorization, X-Wallet-Auth (conditional)
6. Execute request
7. Check status code (2xx = success)
8. Unmarshal response JSON (if result non-nil)

**Returns**: Error if request fails, nil on success

**Errors**:
- CDPError with status code and message
- Context deadline exceeded
- Network errors

### `doRequestWithRetry(ctx context.Context, method, path string, body, result interface{}, requireWalletAuth bool) error`
Wrapper around doRequest with exponential backoff retry logic for transient failures.

**Retry Behavior**:
- Max attempts: 5
- Initial delay: 100ms
- Multiplier: 2x
- Max total time: 10s
- Jitter: ±25%

**Retryable Errors**:
- 429 Too Many Requests (rate limit)
- 5xx Server Errors
- Network timeouts
- Connection failures

**Non-Retryable Errors** (fail immediately):
- 401 Unauthorized
- 403 Forbidden
- 4xx Client Errors (except 429)

**Example**:
```go
client := &CDPClient{
    baseURL:    "https://api.cdp.coinbase.com",
    httpClient: &http.Client{Timeout: 30 * time.Second},
    auth:       auth,
}

var account CDPAccount
err := client.doRequestWithRetry(
    ctx, 
    "POST", 
    "/platform/v2/evm/accounts",
    map[string]string{"network_id": "base-sepolia"},
    &account,
    false,
)
```

**State Transitions**: Stateless (no state changes, thread-safe)

---

## Entity: CDPAccount

**Purpose**: Represents a blockchain wallet account managed by CDP

**Fields**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ID` | string | Yes | CDP-internal account identifier |
| `Address` | string | Yes | Blockchain address (EVM: 0x-prefixed hex, SVM: base58 public key) |
| `Network` | string | Yes | CDP network identifier (e.g., "base-sepolia", "solana-devnet") |

**Validation Rules**:
- `ID` must not be empty
- `Address` must be valid for network type:
  - EVM: 40 hex characters with "0x" prefix (42 total)
  - SVM: 32-44 base58 characters
- `Network` must be supported CDP network

**Relationships**:
- Created by CDP API via `CreateOrGetAccount` helper
- Used by Signer during initialization
- One account per credentials + network combination

**State Transitions**: Immutable (created once, never modified)

**Example**:
```go
// EVM Account
account := CDPAccount{
    ID:      "accounts/abc-123",
    Address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    Network: "base-sepolia",
}

// SVM Account
account := CDPAccount{
    ID:      "accounts/def-456",
    Address: "DYw8jCTfwHNRJhhmFcbXvVDTqWMEVFBX6ZKUmG5CNSKK",
    Network: "solana-devnet",
}
```

---

## Entity: Signer

**Purpose**: Implements x402.Signer interface using CDP for transaction signing operations

**Fields**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cdpClient` | *CDPClient | Yes | HTTP client for CDP API communication |
| `auth` | *CDPAuth | Yes | Authentication credentials |
| `accountID` | string | Yes | CDP account identifier |
| `address` | string | Yes | Blockchain wallet address |
| `network` | string | Yes | x402 network identifier (e.g., "base", "solana") |
| `networkType` | NetworkType | Yes | EVM or SVM enum |
| `chainID` | *big.Int | Conditional | EVM chain ID (nil for SVM) |
| `tokens` | []x402.TokenConfig | Yes | Supported payment tokens (at least one required) |
| `priority` | int | Yes | Signer selection priority (lower = higher priority, default 0) |
| `maxAmount` | *big.Int | No | Per-call spending limit (nil = no limit) |

**Methods** (implements x402.Signer interface):

### `Network() string`
Returns x402 network identifier.

**Example**: "base", "base-sepolia", "solana", "solana-devnet"

### `Scheme() string`
Returns payment scheme (always "exact" for CDP signer).

### `CanSign(requirements *PaymentRequirement) bool`
Validates if signer can satisfy payment requirements.

**Logic**:
1. Check network match (exact string comparison)
2. Check scheme match (must be "exact")
3. Check token match (case-insensitive comparison)

**Returns**: true if all checks pass, false otherwise

### `Sign(requirements *PaymentRequirement) (*PaymentPayload, error)`
Signs payment transaction via CDP API and returns payment payload.

**Process**:
1. Validate with CanSign() (return error if false)
2. Parse amount string to *big.Int
3. Check maxAmount limit (return ErrAmountExceeded if exceeded)
4. Build chain-specific signing request:
   - **EVM**: EIP-712 typed data (EIP-3009 authorization)
   - **SVM**: Solana transaction message (TransferChecked instruction)
5. Call CDP API sign endpoint with Wallet Auth
6. Construct PaymentPayload with signature
7. Return payload

**Returns**: 
- `*PaymentPayload` on success
- Error on failure (ErrNoValidSigner, ErrInvalidAmount, ErrAmountExceeded, CDP API errors)

### `GetPriority() int`
Returns signer priority for selection.

### `GetTokens() []x402.TokenConfig`
Returns configured token list.

### `GetMaxAmount() *big.Int`
Returns per-call spending limit (nil if no limit).

**Additional Methods** (non-interface):

### `Address() string`
Returns blockchain wallet address for debugging/logging.

**Validation Rules** (enforced in constructor):
- `auth` must not be nil
- `network` must not be empty
- `tokens` must have at least one element
- `networkType` must be NetworkTypeEVM or NetworkTypeSVM
- `chainID` required for EVM, must be nil for SVM

**State Transitions**: Immutable (all fields set during construction via functional options)

**Relationships**:
- Uses CDPClient for API communication
- Uses CDPAuth for authentication
- Integrates with x402 middleware via Signer interface
- Used by DefaultPaymentSelector for payment processing

**Example**:
```go
signer := &Signer{
    cdpClient:   client,
    auth:        auth,
    accountID:   "accounts/abc-123",
    address:     "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
    network:     "base-sepolia",
    networkType: NetworkTypeEVM,
    chainID:     big.NewInt(84532),
    tokens: []x402.TokenConfig{
        {Symbol: "eth", Address: "0x0000000000000000000000000000000000000000", Priority: 0},
    },
    priority:  0,
    maxAmount: big.NewInt(1000000000000000000), // 1 ETH
}

payload, err := signer.Sign(&x402.PaymentRequirement{
    Network: "base-sepolia",
    Scheme:  "exact",
    Token:   "eth",
    Amount:  "500000000000000000", // 0.5 ETH
})
```

---

## Supporting Types

### NetworkType (enum)

```go
type NetworkType int

const (
    NetworkTypeUnknown NetworkType = iota
    NetworkTypeEVM
    NetworkTypeSVM
)
```

**Purpose**: Identifies blockchain type for network-specific logic

**Usage**:
- Determines which CDP API endpoints to use
- Controls signing logic (EIP-712 vs Solana transaction)
- Validates chain ID requirement (required for EVM, nil for SVM)

---

### CDPError

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

**Purpose**: Structured error type for CDP API failures with retry information

**Fields**:
- `StatusCode`: HTTP status code
- `ErrorType`: Error category for classification
- `Message`: Human-readable error message
- `RequestID`: CDP request ID for tracking
- `Retryable`: Whether error should be retried
- `RetryAfter`: Backoff duration (from Retry-After header or calculated)

**Usage**:
```go
if cdpErr, ok := err.(*CDPError); ok {
    if cdpErr.Retryable {
        time.Sleep(cdpErr.RetryAfter)
        // Retry request
    } else {
        return fmt.Errorf("non-retryable CDP error: %w", cdpErr)
    }
}
```

---

## Functional Options Pattern

### SignerOption

```go
type SignerOption func(*Signer) error
```

**Purpose**: Functional option for configuring Signer during construction

**Available Options**:

#### `WithCDPCredentials(apiKeyName, apiKeySecret, walletSecret string) SignerOption`
Sets CDP authentication credentials.

**Validation**: 
- apiKeyName must not be empty
- apiKeySecret must be valid PEM-encoded key
- walletSecret optional but recommended

**Error**: Returns x402.ErrInvalidKey if validation fails

#### `WithNetwork(network string) SignerOption`
Sets blockchain network.

**Validation**: network must be supported x402 network

**Error**: Returns x402.ErrInvalidNetwork if unsupported

#### `WithToken(symbol, address string) SignerOption`
Adds payment token with default priority 0.

**Parameters**:
- `symbol`: Token symbol (e.g., "eth", "usdc")
- `address`: Token contract address (native token uses zero address)

#### `WithTokenPriority(symbol, address string, priority int) SignerOption`
Adds payment token with specific priority.

**Priority**: Lower number = higher priority for selection

#### `WithPriority(priority int) SignerOption`
Sets signer priority for selection.

**Default**: 0

#### `WithMaxAmountPerCall(amount *big.Int) SignerOption`
Sets per-call spending limit.

**Default**: nil (no limit)

**Usage Example**:
```go
signer, err := NewSigner(
    WithCDPCredentials(
        os.Getenv("CDP_API_KEY_NAME"),
        os.Getenv("CDP_API_KEY_SECRET"),
        os.Getenv("CDP_WALLET_SECRET"),
    ),
    WithNetwork("base-sepolia"),
    WithToken("eth", "0x0000000000000000000000000000000000000000"),
    WithMaxAmountPerCall(big.NewInt(1000000000000000000)),
)
```

---

## Data Flow

### Account Creation Flow

```
User
  │
  ├─> NewSigner(WithCDPCredentials(...), WithNetwork("base"))
  │     │
  │     ├─> CreateOrGetAccount(ctx, auth, "base")
  │     │     │
  │     │     ├─> getCDPNetwork("base") → "base-mainnet"
  │     │     │
  │     │     ├─> GET /platform/v2/evm/accounts (list existing)
  │     │     │     │
  │     │     │     └─> Response: [] (empty, no accounts)
  │     │     │
  │     │     ├─> POST /platform/v2/evm/accounts {"network_id": "base-mainnet"}
  │     │     │     │
  │     │     │     └─> Response: {
  │     │     │           "id": "accounts/abc-123",
  │     │     │           "address": "0x742d...",
  │     │     │           "network": "base-mainnet"
  │     │     │         }
  │     │     │
  │     │     └─> Return CDPAccount{ID, Address, Network}
  │     │
  │     ├─> Set signer.accountID = account.ID
  │     ├─> Set signer.address = account.Address
  │     └─> Return signer
  │
  └─> signer.Sign(requirement)
```

### Signing Flow (EVM)

```
Middleware
  │
  ├─> PaymentRequirement{Network: "base", Token: "eth", Amount: "1000000"}
  │     │
  │     └─> selector.SelectSigner(requirement)
  │           │
  │           ├─> signer.CanSign(requirement)
  │           │     │
  │           │     ├─> Check network match: "base" == "base" ✓
  │           │     ├─> Check scheme: "exact" == "exact" ✓
  │           │     ├─> Check token: "eth" in tokens ✓
  │           │     └─> Return true
  │           │
  │           └─> signer.Sign(requirement)
  │                 │
  │                 ├─> Parse amount: "1000000" → big.Int(1000000)
  │                 ├─> Check maxAmount: 1000000 < 1000000000000000000 ✓
  │                 │
  │                 ├─> Build EIP-3009 authorization
  │                 │     {
  │                 │       from: signer.address,
  │                 │       to: requirement.recipient,
  │                 │       value: 1000000,
  │                 │       validAfter: now - 10s,
  │                 │       validBefore: now + 3600s,
  │                 │       nonce: cryptoRandom32Bytes()
  │                 │     }
  │                 │
  │                 ├─> Convert to EIP-712 TypedData
  │                 │
  │                 ├─> POST /platform/v2/evm/accounts/{address}/sign/typed-data
  │                 │     Headers: {
  │                 │       Authorization: "Bearer {jwt}",
  │                 │       X-Wallet-Auth: "{wallet-jwt}"
  │                 │     }
  │                 │     Body: {typedData: {...}}
  │                 │     │
  │                 │     └─> Response: {signature: "0xabc..."}
  │                 │
  │                 └─> Return PaymentPayload{
  │                       X402Version: 1,
  │                       Scheme: "exact",
  │                       Network: "base",
  │                       Payload: EVMPayload{
  │                         Signature: "0xabc...",
  │                         Authorization: {...}
  │                       }
  │                     }
  │
  └─> Middleware sends payload to facilitator
```

---

## Persistence

**Note**: CDP signer is stateless. All persistent state managed by CDP API:
- Account storage: CDP manages accounts in their database
- Private keys: CDP stores in TEE (Trusted Execution Environment)
- Transaction history: CDP maintains transaction logs

**Local State**: None persisted. All configuration loaded from environment variables on initialization.

---

## Summary

The data model consists of 4 core entities:

1. **CDPAuth**: Authentication credentials and JWT generation
2. **CDPClient**: HTTP client for CDP API communication
3. **CDPAccount**: Blockchain account representation
4. **Signer**: Main implementation of x402.Signer interface

Key characteristics:
- All entities immutable after construction
- Stateless design (thread-safe, no synchronization needed)
- Clear separation of concerns (auth, HTTP, account, signing)
- Follows existing x402 signer patterns
- Minimal abstraction (no unnecessary layers)
