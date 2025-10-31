# Data Model: MCP Integration

**Branch**: `007-mcp-integration` | **Date**: 2025-10-31

## Core Entities

### 1. MCPTransport
**Purpose**: Client-side transport handling x402 payments for MCP communication

**Fields**:
- `serverURL` (url.URL): MCP server endpoint
- `httpClient` (*http.Client): HTTP client for requests
- `signers` ([]x402.Signer): Payment signers in priority order
- `selector` (x402.PaymentSelector): Payment method selector
- `sessionID` (string): Current MCP session identifier
- `protocolVersion` (string): Negotiated MCP protocol version

**Relationships**:
- Uses multiple `x402.Signer` instances for payment creation
- Integrates with `http.FacilitatorClient` for verification
- Implements `transport.Interface` from mcp-go

**Validation Rules**:
- At least one signer must be configured
- Server URL must be valid HTTP/HTTPS endpoint
- Session ID must be maintained across requests

**State Transitions**:
1. Uninitialized → Initialized (on first successful connection)
2. Active → PaymentRequired (on 402 error)
3. PaymentRequired → Active (on successful payment)
4. Active → Closed (on connection termination)

---

### 2. MCPServer
**Purpose**: Server wrapper providing x402 payment protection for MCP tools

**Fields**:
- `mcpServer` (*server.MCPServer): Underlying MCP server
- `paymentTools` (map[string][]PaymentRequirement): Tool payment configs
- `facilitatorClient` (*http.FacilitatorClient): Payment verifier
- `verifyOnly` (bool): Skip settlement if true

**Relationships**:
- Wraps `server.MCPServer` from mcp-go
- Uses `http.FacilitatorClient` for payment operations
- Maps tool names to `PaymentRequirement` arrays

**Validation Rules**:
- Tool names must be unique
- Payment requirements must have valid addresses
- Facilitator URL must be reachable

**State Transitions**:
- Tools are configured at initialization (immutable at runtime)
- Payment verification is stateless per request

---

### 3. PaymentContext
**Purpose**: Payment information passed through MCP request lifecycle

**Fields**:
- `payment` (*x402.PaymentPayload): Signed payment data
- `requirement` (*x402.PaymentRequirement): Matched requirement
- `verificationResult` (*VerifyResponse): Facilitator verification
- `settlementResult` (*x402.SettlementResponse): Settlement outcome

**Relationships**:
- Created from MCP request params._meta
- Validated against tool's PaymentRequirement
- Verified through FacilitatorClient

**Validation Rules**:
- Payment must match at least one requirement
- Verification must succeed before tool execution
- Settlement required unless verify-only mode

**State Transitions**:
1. Created (from request)
2. Validated (requirement matched)
3. Verified (facilitator confirmed)
4. Settled (blockchain executed)
5. Complete (response sent)

---

### 4. ToolPaymentConfig
**Purpose**: Associates MCP tools with payment requirements

**Fields**:
- `toolName` (string): MCP tool identifier
- `requirements` ([]x402.PaymentRequirement): Accepted payments
- `enabled` (bool): Payment requirement active flag

**Relationships**:
- References MCP tool by name
- Contains multiple PaymentRequirement options
- Checked on each tool invocation

**Validation Rules**:
- Tool must exist in MCP server
- At least one requirement if payment enabled
- All requirements must have same tool resource

---

## Request/Response Flow

### Client Payment Flow
```
1. Client sends MCP request
2. Server returns 402 with requirements
3. Client selects matching signer
4. Client creates signed payment
5. Client adds payment to params._meta
6. Client resends request with payment
7. Server verifies and executes tool
8. Client receives tool response
```

### Server Verification Flow
```
1. Receive MCP tool request
2. Check if tool requires payment
3. Extract payment from params._meta
4. Match payment to requirement
5. Verify with facilitator
6. Settle payment (if not verify-only)
7. Execute tool handler
8. Add settlement to response._meta
```

## Data Serialization

### Payment in Request (params._meta)
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "search",
    "arguments": {...},
    "_meta": {
      "x402/payment": {
        "x402Version": 1,
        "scheme": "exact",
        "network": "base",
        "payload": {
          "signature": "0x...",
          "authorization": {...}
        }
      }
    }
  }
}
```

### Payment Requirements (402 error)
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": 402,
    "message": "Payment required",
    "data": {
      "x402Version": 1,
      "error": "Payment required to access this resource",
      "accepts": [{
        "scheme": "exact",
        "network": "base",
        "maxAmountRequired": "10000",
        "asset": "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913",
        "payTo": "0x...",
        "resource": "mcp://tools/search"
      }]
    }
  }
}
```

### Settlement in Response (result._meta)
```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [...],
    "_meta": {
      "x402/payment-response": {
        "success": true,
        "transaction": "0x...",
        "network": "base",
        "payer": "0x..."
      }
    }
  }
}
```

## Concurrency Considerations

### Client-Side
- Each request gets independent payment (no payment reuse)
- Signers must be thread-safe for concurrent signing
- Session ID atomically updated
- Transport supports concurrent requests

### Server-Side
- Payment verification is stateless per request
- No shared state between concurrent tool calls
- Facilitator client handles concurrent verify/settle
- Tool handlers execute independently

## Error States

### Payment Errors
- `ErrPaymentRequired`: No payment provided
- `ErrNoMatchingSigner`: No signer can fulfill requirements
- `ErrInsufficientBalance`: Signer balance too low
- `ErrPaymentRejected`: Facilitator rejected payment
- `ErrSettlementFailed`: Blockchain transaction failed

### MCP Integration Errors
- `ErrSessionTerminated`: MCP session ended
- `ErrInvalidRequest`: Malformed MCP request
- `ErrToolNotFound`: Unknown tool name
- `ErrToolExecutionFailed`: Tool handler error

## Migration from mcp-go-x402

### Type Mappings
- `x402.PaymentSigner` → `x402.Signer`
- `x402.PaymentHandler` → (removed, logic in transport)
- `x402.PaymentRequirement` → `x402.PaymentRequirement` (reused)
- `x402.PaymentPayload` → `x402.PaymentPayload` (reused)

### Configuration Changes
- Signers now use x402-go signer types
- Facilitator client reused from http package
- Payment requirements use x402 types directly