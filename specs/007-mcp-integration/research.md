# Research: MCP Integration

**Branch**: `007-mcp-integration` | **Date**: 2025-10-31

## Overview

This document captures technical decisions and research findings for integrating MCP (Model Context Protocol) support into x402-go, enabling payment-gated access to AI tools and services.

## Technical Decisions

### 1. Architecture Approach

**Decision**: Adapter pattern with minimal new code
**Rationale**: 
- Maximizes reuse of existing x402-go components (signers, facilitator, types)
- Maintains consistency with current x402 implementation
- Reduces maintenance burden and potential bugs
**Alternatives considered**:
- Full port of mcp-go-x402: Rejected due to code duplication
- Wrapper library: Rejected as it would add unnecessary abstraction layer

### 2. Payment Flow Integration

**Decision**: JSON-RPC 402 error approach with params._meta injection
**Rationale**:
- Aligns with MCP protocol's JSON-RPC transport
- Transparent to MCP protocol layer
- Supports both HTTP header and JSON-RPC parameter transports
**Alternatives considered**:
- HTTP-only 402: Rejected as MCP uses JSON-RPC errors
- Custom protocol extension: Rejected as it would break MCP compatibility

### 3. Signer Reuse Strategy

**Decision**: Direct reuse of x402.Signer interface without modification
**Rationale**:
- Existing signers already support all required networks (EVM, Solana)
- No need to duplicate signing logic
- Maintains single source of truth for payment signing
**Alternatives considered**:
- MCP-specific signers: Rejected due to unnecessary duplication
- Signer adapters: Rejected as existing interface is sufficient

### 4. Facilitator Integration

**Decision**: Reuse http.FacilitatorClient directly
**Rationale**:
- Already implements verify/settle endpoints
- Handles retries and timeouts appropriately
- Well-tested in production use
**Alternatives considered**:
- New MCP facilitator client: Rejected as functionality is identical
- Mock facilitator for testing: Will use but not replace real client

### 5. Transport Implementation

**Decision**: Implement transport.Interface from mcp-go with x402 additions
**Rationale**:
- Required for MCP client compatibility
- Allows transparent payment handling
- Supports bidirectional communication and SSE
**Alternatives considered**:
- HTTP-only transport: Rejected as MCP requires full transport interface
- Custom transport protocol: Rejected for compatibility reasons

### 6. Server Middleware Pattern

**Decision**: Wrapper around MCP server handler
**Rationale**:
- Intercepts tool calls to check payment requirements
- Maintains clean separation of concerns
- Easy to enable/disable payment requirements per tool
**Alternatives considered**:
- Modified MCP server: Rejected as it would require forking mcp-go
- Post-processing hook: Rejected as payment must be verified before execution

### 7. Payment Requirements Configuration

**Decision**: Per-tool configuration in server initialization code
**Rationale**:
- Flexible pricing per tool
- Supports multiple payment options per tool
- No external configuration files needed
**Alternatives considered**:
- Configuration file: Rejected as it adds complexity
- Annotation-based: Rejected as Go doesn't support runtime annotations

### 8. Testing Strategy

**Decision**: Comprehensive unit tests with mock facilitator
**Rationale**:
- Tests payment flows without real blockchain interaction
- Fast test execution
- Deterministic test results
**Alternatives considered**:
- Integration tests only: Rejected as too slow for development
- No mocking: Rejected as it would require test wallets and funds

### 9. Example Implementation

**Decision**: Combined client/server example similar to x402demo
**Rationale**:
- Shows both sides of the payment flow
- Self-contained demonstration
- Easy for developers to understand and modify
**Alternatives considered**:
- Separate examples: Rejected as it complicates demonstration
- Multiple examples: Can be added later based on user needs

### 10. Error Handling

**Decision**: Use MCP JSON-RPC error codes with 402 for payment required
**Rationale**:
- Standard MCP error handling
- Clean integration with existing error flows
- Clear error messages for debugging
**Alternatives considered**:
- Custom error codes: Rejected for compatibility
- HTTP status codes: Rejected as MCP uses JSON-RPC

## Implementation Dependencies

### Required from x402-go
- `x402.Signer` interface and implementations (EVM, Solana)
- `http.FacilitatorClient` for payment verification/settlement
- Core types: `PaymentRequirement`, `PaymentPayload`, `SettlementResponse`
- Chain configurations and token addresses

### Required from mcp-go
- `transport.Interface` for client implementation
- `server.MCPServer` for server base
- JSON-RPC types and error codes
- SSE (Server-Sent Events) support

### New Components
- `mcp/client/transport.go`: MCP transport with x402 payment handling
- `mcp/server/middleware.go`: Payment verification middleware
- `mcp/types.go`: Minimal MCP-specific types (mostly aliases)

## Testing Considerations

### Unit Tests Required
1. Transport payment flow (attempt → verify → settle)
2. Multi-signer fallback logic
3. Payment requirement matching
4. Server middleware interception
5. Error handling for invalid payments
6. Concurrent payment handling

### Integration Tests Required
1. Full client-server payment flow
2. Free and paid tool mixing
3. Network-specific payments (EVM vs Solana)
4. Facilitator timeout handling

## Performance Requirements

### Targets
- Payment verification: < 5 seconds (facilitator timeout)
- Payment settlement: < 60 seconds (blockchain confirmation)
- Concurrent payments: Support at least 100 simultaneous
- Memory overhead: < 10MB per connection

### Optimization Opportunities
- Cache facilitator /supported responses
- Reuse HTTP connections
- Batch payment verifications (future)

## Security Considerations

1. **Payment Validation**: All payments verified through facilitator
2. **Replay Protection**: Nonce in each payment prevents reuse
3. **Amount Verification**: Exact amount matching required
4. **Network Verification**: Payment network must match requirement
5. **No Refunds**: Failed tool execution after payment is non-refundable

## Migration Path

For users of mcp-go-x402:
1. Update import paths to `github.com/mark3labs/x402-go/mcp`
2. Replace custom signers with x402-go signers
3. Update configuration to use x402-go types
4. No changes needed to MCP tool implementations

## Open Questions Resolved

1. **Q: Should we support payment batching?**
   A: No, each request requires separate payment per spec

2. **Q: How to handle partial payments?**
   A: Not supported, exact amount required

3. **Q: Should we cache payment verifications?**
   A: No, each payment must be verified independently

4. **Q: How to handle facilitator downtime?**
   A: Return 503 Service Unavailable with clear error

## Next Steps

1. Generate data model documentation
2. Create API contracts for MCP integration
3. Write quickstart guide for developers
4. Begin test-driven implementation