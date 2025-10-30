# Implementation Plan: Coinbase CDP Signer Integration

**Branch**: `006-cdp-signer` | **Date**: 2025-10-30 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/006-cdp-signer/spec.md`

## Summary

This feature adds support for Coinbase Developer Platform (CDP) server wallets to x402-go, enabling developers to sign payment transactions without managing private keys locally. The CDP signer implements the existing `x402.Signer` interface and supports both EVM and Solana chains through CDP's REST API. Authentication uses JWT tokens generated per-request, with automatic account creation/retrieval on initialization. The implementation follows existing signer patterns (evm/signer.go, svm/signer.go) with comprehensive test coverage.

## Technical Context

**Language/Version**: Go 1.25.1  
**Primary Dependencies**: 
  - gopkg.in/square/go-jose.v2 (JWT/JWS for CDP authentication)
  - github.com/mark3labs/x402-go (existing signer interfaces)
  - Go stdlib: crypto/sha256, crypto/x509, encoding/pem, net/http, encoding/json

**Storage**: N/A (stateless signer, CDP manages wallet state)  
**Testing**: Go testing package (`go test -race -cover`)  
**Target Platform**: Linux/macOS/Windows servers (any platform supporting Go 1.25.1+)  
**Project Type**: Library package (signers/coinbase/)

**Constraints**: 
  - CDP API rate limit responses (429 status): 600 reads/500 writes per 10 seconds
  - JWT token expiration: 2min (Bearer), 1min (Wallet Auth) - no caching
  - Zero credential leakage in logs/errors

**Scale/Scope**: 
  - Support 8+ networks (4 EVM: base, base-sepolia, ethereum, sepolia; 2 SVM: solana-devnet, mainnet-beta)
  - 3 environment variables for configuration
  - 7 source files + 7 test files
  - Target test coverage: >80% (maintain existing project coverage)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**✅ Principle I: No Unnecessary Documentation**
- Only creating documentation explicitly required by user and development process
- research.md: Required for Phase 0 (resolve technical unknowns)
- data-model.md: Required for Phase 1 (entity definitions)
- contracts/: Required for Phase 1 (API interfaces)
- quickstart.md: Required for Phase 1 (usage examples)
- No README, CONTRIBUTING, or other unsolicited docs

**✅ Principle II: Test Coverage Preservation**
- Will maintain or exceed existing test coverage
- Table-driven tests for all public functions
- Unit tests for auth, client, account, signing logic
- Integration tests against CDP testnet (optional, skipped if no credentials)
- Coverage measured with `go test -race -cover ./...`

**✅ Principle III: Test-First Development**
- Tests written before implementation for each component
- TDD cycle: write test → verify fail → implement → verify pass
- Test files created alongside source files

**✅ Principle IV: Stdlib-First Approach**
- Using stdlib for: net/http (HTTP client), crypto/sha256 (hashing), encoding/json (serialization)
- gopkg.in/square/go-jose.v2: Required for CDP JWT - stdlib crypto/jwt insufficient (needs ES256 signing)
- Justified: CDP requires ECDSA ES256 signatures for JWT, stdlib jwt package doesn't support signing

**✅ Principle V: Code Conciseness**
- Following existing signer patterns (functional options, minimal abstractions)
- Comments explain why (e.g., "Generate fresh JWT for each request to avoid token expiration edge cases")
- No obvious comments restating code (avoid "// Set priority" for `s.priority = priority`)

**✅ Principle VI: Binary Cleanup**
- No binaries in signers/coinbase/ (library package, no executables)
- .gitignore already excludes binaries

**🔄 Re-check After Phase 1**: Verify no extra dependencies added, test coverage plan documented

## Project Structure

### Documentation (this feature)

```text
specs/006-cdp-signer/
├── plan.md              # This file
├── research.md          # Phase 0: Technology decisions, CDP API analysis
├── data-model.md        # Phase 1: Signer, Auth, Client, Account entities
├── quickstart.md        # Phase 1: Usage examples (EVM + SVM)
├── contracts/           # Phase 1: API interfaces
│   └── signer-api.yaml  # Public signer interface documentation
└── tasks.md             # Phase 2: Implementation task breakdown (created by /speckit.tasks)
```

### Source Code (repository root)

```text
signers/coinbase/        # New package location
├── signer.go           # Main Signer struct, interface implementation, functional options
├── signer_test.go      # Signer tests (constructor, CanSign, Sign, interface methods)
├── auth.go             # CDPAuth struct, JWT generation (Bearer + Wallet Auth)
├── auth_test.go        # Auth tests (JWT generation, credential validation)
├── client.go           # CDPClient struct, HTTP request handling, retry logic
├── client_test.go      # Client tests (request construction, error handling, retries)
├── account.go          # CreateOrGetAccount helper, account creation/retrieval logic
├── account_test.go     # Account tests (creation, retrieval, idempotency)
├── errors.go           # CDP-specific error types (optional, may inline if minimal)
├── networks.go         # Network mapping (x402 names → CDP network IDs)
└── networks_test.go    # Network mapping tests

# Future refactoring (tracked in bd issue x402-go-15)
# evm/ → signers/evm/
# svm/ → signers/svm/
```

**Structure Decision**: Creating new `signers/coinbase/` package to house CDP signer implementation. This follows the pattern that will be used for future refactoring (evm/ and svm/ to signers/evm and signers/svm). The package contains all CDP-specific logic isolated from core x402 types.

## Complexity Tracking

> No constitution violations - all complexity justified

| Check | Status | Notes |
|-------|--------|-------|
| New dependency (go-jose) | ✅ Justified | CDP requires ES256 JWT signing, stdlib insufficient |
| Test coverage maintained | ✅ Planned | >80% coverage with table-driven tests |
| Stdlib-first | ✅ Followed | Only external dep is go-jose for JWT |
| No extra docs | ✅ Followed | Only required spec artifacts |

---

## Phase 0: Research & Decisions

**Objective**: Resolve all NEEDS CLARIFICATION items from Technical Context through research and experimentation.

### Research Tasks

1. **CDP API Endpoint Analysis**
   - Document exact REST API endpoints for EVM/SVM account creation
   - Document exact REST API endpoints for EVM/SVM transaction signing
   - Verify JWT claim requirements (sub, iss, nbf, exp, uri, reqHash)
   - Document response schemas for account creation and signing operations
   - Reference: https://docs.cdp.coinbase.com/api-reference/v2

2. **JWT Library Evaluation**
   - Confirm gopkg.in/square/go-jose.v2 supports ES256 algorithm
   - Test JWT generation with sample CDP credentials (testnet)
   - Verify token structure matches CDP requirements
   - Alternative considered: golang-jwt/jwt - rejected because limited ES256 support

3. **Account Creation Flow Research**
   - Determine if CDP API provides account listing/query capabilities
   - Document idempotency behavior (create vs retrieve existing)
   - Test account creation on testnet to understand response format
   - Verify network identifier mapping (x402 names vs CDP network IDs)

4. **Error Classification Research**
   - Document all CDP API error codes (4xx, 5xx)
   - Classify which errors are retryable vs non-retryable
   - Document rate limit response format (429 status)
   - Test retry behavior with exponential backoff

5. **Existing Signer Pattern Analysis**
   - Review evm/signer.go and svm/signer.go for interface patterns
   - Document functional options pattern usage
   - Identify validation patterns (CanSign logic)
   - Document signing flow patterns (amount parsing, payload construction)

### Research Output (research.md)

Document decisions in following format:

```markdown
# Research: CDP Signer Implementation

## Decision: JWT Library Selection

**Chosen**: gopkg.in/square/go-jose.v2 v2.6.3

**Rationale**: 
- Native ES256 (ECDSA P-256) support required by CDP
- Mature library (v2.6.3 stable since 2022)
- Used in production by major projects
- Simple API for JWT signing with custom claims

**Alternatives Considered**:
- golang-jwt/jwt: Limited ES256 support, complex key handling
- Standard library crypto: No high-level JWT support, would need manual implementation

**Testing**: Successfully generated JWT with ES256 signature matching CDP requirements

## Decision: Account Creation Strategy

[...]
```

---

## Phase 1: Design & Contracts

**Prerequisites**: research.md complete

### 1. Data Model (data-model.md)

Extract entities from spec and define structure:

#### Entity: CDPAuth

**Purpose**: Manages CDP authentication credentials and JWT token generation

**Fields**:
- `apiKeyName` (string): CDP API key identifier (e.g., "organizations/xxx/apiKeys/yyy")
- `apiKeySecret` (string): PEM-encoded ECDSA private key for signing JWTs
- `walletSecret` (string): Optional wallet-specific secret for signing operations

**Methods**:
- `GenerateBearerToken(method, path string) (string, error)`: Creates 2-minute JWT for API requests
- `GenerateWalletAuthToken(method, path string, bodyHash []byte) (string, error)`: Creates 1-minute JWT for signing

**Validation**:
- `apiKeyName` must not be empty
- `apiKeySecret` must be valid PEM-encoded ECDSA key
- `walletSecret` optional (only required for signing operations)

#### Entity: CDPClient

**Purpose**: HTTP client wrapper for CDP REST API communication

**Fields**:
- `baseURL` (string): CDP API base URL (https://api.cdp.coinbase.com)
- `httpClient` (*http.Client): Configured HTTP client with timeouts
- `auth` (*CDPAuth): Authentication handler

**Methods**:
- `doRequest(ctx, method, path string, body, result interface{}, requireWalletAuth bool) error`: Core request handler
- `doRequestWithRetry(...)`: Request handler with exponential backoff retry logic

**State Transitions**: Stateless (no state changes)

#### Entity: CDPAccount

**Purpose**: Represents a blockchain wallet account managed by CDP

**Fields**:
- `ID` (string): CDP-internal account identifier
- `Address` (string): Blockchain address (EVM hex or Solana base58)
- `Network` (string): CDP network identifier

**Validation**:
- `Address` must be valid for network type (EVM: 0x prefix, SVM: base58)
- `Network` must be supported CDP network

#### Entity: Signer

**Purpose**: Implements x402.Signer interface using CDP for signing operations

**Fields**:
- `cdpClient` (*CDPClient): HTTP client for CDP API
- `auth` (*CDPAuth): Authentication credentials
- `accountID` (string): CDP account ID
- `address` (string): Blockchain address
- `network` (string): x402 network identifier
- `networkType` (NetworkType): EVM or SVM enum
- `chainID` (*big.Int): EVM chain ID (nil for SVM)
- `tokens` ([]x402.TokenConfig): Supported payment tokens
- `priority` (int): Signer selection priority
- `maxAmount` (*big.Int): Per-call spending limit

**Methods** (implement x402.Signer):
- `Network() string`
- `Scheme() string`
- `CanSign(*PaymentRequirement) bool`
- `Sign(*PaymentRequirement) (*PaymentPayload, error)`
- `GetPriority() int`
- `GetTokens() []TokenConfig`
- `GetMaxAmount() *big.Int`

**State Transitions**: Immutable after initialization

### 2. API Contracts (contracts/signer-api.yaml)

```yaml
# Public API contract for CDP signer

Signer:
  constructor:
    name: NewSigner
    options:
      - WithCDPCredentials(apiKeyName, apiKeySecret, walletSecret string)
      - WithNetwork(network string)
      - WithToken(symbol string, address string)
      - WithTokenPriority(symbol string, address string, priority int)
      - WithPriority(priority int)
      - WithMaxAmountPerCall(amount *big.Int)
    returns: (*Signer, error)
    errors:
      - ErrInvalidKey: Invalid CDP credentials
      - ErrInvalidNetwork: Network not specified or unsupported
      - ErrNoTokens: No tokens configured

  methods:
    Network:
      returns: string
      
    Scheme:
      returns: string # Always "exact"
      
    CanSign:
      params:
        - requirements: *PaymentRequirement
      returns: bool
      
    Sign:
      params:
        - requirements: *PaymentRequirement
      returns: (*PaymentPayload, error)
      errors:
        - ErrNoValidSigner: Cannot sign (network/token mismatch)
        - ErrInvalidAmount: Amount parsing failed
        - ErrAmountExceeded: Amount exceeds maxAmount limit
        - CDP API errors (auth, rate limit, server error)
        
    GetPriority:
      returns: int
      
    GetTokens:
      returns: []TokenConfig
      
    GetMaxAmount:
      returns: *big.Int # nil if no limit

Helper:
  CreateOrGetAccount:
    params:
      - ctx: context.Context
      - auth: *CDPAuth
      - network: string
    returns: (*CDPAccount, error)
    errors:
      - ErrInvalidKey: Authentication failed
      - ErrInvalidNetwork: Unsupported network
      - CDP API errors
    behavior: |
      Idempotent. Creates new account if none exists, 
      retrieves existing account otherwise.
```

### 3. Quickstart Guide (quickstart.md)

```markdown
# Quickstart: CDP Signer

## Prerequisites

1. Create CDP project at https://portal.cdp.coinbase.com
2. Generate Secret API Key (Ed25519 recommended)
3. Generate Wallet Secret from Server Wallet dashboard
4. Set environment variables:

```bash
export CDP_API_KEY_NAME="organizations/xxx/apiKeys/yyy"
export CDP_API_KEY_SECRET="-----BEGIN EC PRIVATE KEY-----
...
-----END EC PRIVATE KEY-----"
export CDP_WALLET_SECRET="your-wallet-secret"
```

## EVM Example (Base Sepolia)

```go
package main

import (
    "log"
    "math/big"
    "os"
    
    "github.com/mark3labs/x402-go"
    "github.com/mark3labs/x402-go/http"
    cdp "github.com/mark3labs/x402-go/signers/coinbase"
)

func main() {
    // Initialize CDP signer for Base Sepolia
    signer, err := cdp.NewSigner(
        cdp.WithCDPCredentials(
            os.Getenv("CDP_API_KEY_NAME"),
            os.Getenv("CDP_API_KEY_SECRET"),
            os.Getenv("CDP_WALLET_SECRET"),
        ),
        cdp.WithNetwork("base-sepolia"),
        cdp.WithToken("eth", "0x0000000000000000000000000000000000000000"),
        cdp.WithMaxAmountPerCall(big.NewInt(1000000000000000000)), // 1 ETH
    )
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("CDP Signer initialized: %s", signer.Address())
    
    // Use with x402 HTTP client
    client, err := http.NewClient(
        http.WithSigner(signer),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Make x402-enabled HTTP request
    resp, err := client.Get("https://api.example.com/data")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
}
```

## SVM Example (Solana Devnet)

[Similar example for Solana...]
```

### 4. Update Agent Context

Run the agent context update script to add new technology information:

```bash
.specify/scripts/bash/update-agent-context.sh opencode
```

This will update `AGENTS.md` with:
- New dependency: gopkg.in/square/go-jose.v2
- New package: signers/coinbase
- CDP integration notes

---

## Phase 2: Implementation Tasks

**Note**: Detailed tasks created by `/speckit.tasks` command - not part of this plan.

High-level task categories:

1. **Setup & Dependencies**
   - Add go-jose dependency to go.mod
   - Create signers/coinbase/ package structure
   - Set up test infrastructure

2. **Authentication (auth.go + auth_test.go)**
   - Implement CDPAuth struct
   - Implement JWT generation (Bearer + Wallet Auth)
   - Test JWT structure and signing
   - Test credential validation

3. **HTTP Client (client.go + client_test.go)**
   - Implement CDPClient struct
   - Implement doRequest with header injection
   - Implement exponential backoff retry logic
   - Test error classification (retryable vs non-retryable)
   - Test rate limit handling (429)
   - Mock CDP API responses for testing

4. **Account Management (account.go + account_test.go)**
   - Implement CreateOrGetAccount helper
   - Implement account creation (EVM + SVM)
   - Implement account retrieval/listing
   - Test idempotency
   - Test concurrent creation race conditions

5. **Network Mapping (networks.go + networks_test.go)**
   - Implement x402 → CDP network mapping
   - Implement EVM chain ID mapping
   - Test all supported networks

6. **Signer Implementation (signer.go + signer_test.go)**
   - Implement Signer struct
   - Implement functional options (WithCDPCredentials, WithNetwork, etc.)
   - Implement interface methods (Network, Scheme, CanSign, Sign, etc.)
   - Implement EVM signing via CDP API
   - Implement SVM signing via CDP API
   - Test constructor with various option combinations
   - Test CanSign logic (network/token matching)
   - Test Sign logic (validation, amount checking, API calls)
   - Test error handling and retry behavior
   - Integration tests against CDP testnet (optional)

7. **Documentation**
   - Add GoDoc comments to all public types/methods
   - Update AGENTS.md via update-agent-context.sh

8. **Quality Gates**
   - Run `go test -race -cover ./signers/coinbase/`
   - Verify >80% test coverage
   - Run `go fmt ./signers/coinbase/`
   - Run `go vet ./signers/coinbase/`
   - Run `golangci-lint run ./signers/coinbase/`

---

## Constitution Re-Check (Post-Design)

**✅ Principle I**: Only created required docs (research, data-model, contracts, quickstart)
**✅ Principle II**: Test coverage plan documented (>80%, table-driven tests)
**✅ Principle III**: TDD approach documented in Phase 2 tasks
**✅ Principle IV**: go-jose dependency justified (ES256 requirement)
**✅ Principle V**: Following concise patterns from existing signers
**✅ Principle VI**: No binaries (library package)

**All gates passed** ✅

---

## Success Metrics

- Test coverage >80%
- Zero credential leakage in logs (verified by security scan)
- All tests pass with `-race` flag
- Integration with existing x402 middleware requires zero code changes

## Next Steps

1. Execute Phase 0 research (generate research.md)
2. Execute Phase 1 design (generate data-model.md, contracts/, quickstart.md)
3. Run update-agent-context.sh
4. Execute Phase 2 implementation (use `/speckit.tasks` command)
5. Create pull request on branch `006-cdp-signer`
