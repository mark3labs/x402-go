# Implementation Plan: MCP Integration

**Branch**: `007-mcp-integration` | **Date**: 2025-10-31 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/007-mcp-integration/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Integrate MCP (Model Context Protocol) support into x402-go by creating client transport and server implementations that handle x402 payment flows transparently. The implementation will reuse existing x402-go signers, facilitator client, and types while adapting the payment flow patterns from mcp-go-x402 reference implementation.

## Technical Context

**Language/Version**: Go 1.25.1  
**Primary Dependencies**: 
  - github.com/mark3labs/mcp-go v0.42.0 (MCP protocol - client.Client, server.MCPServer, transport.Interface)
  - Existing x402-go components (signers, facilitator, types)
**Storage**: N/A (stateless middleware and transport)  
**Testing**: Go test with race detection, table-driven tests  
**Target Platform**: Linux/macOS servers running MCP services  
**Project Type**: single (library with subpackages)  
**Performance Goals**: Payment verification under 5 seconds, settlement under 60 seconds  
**Constraints**: No additional external dependencies beyond mcp-go, must reuse existing x402-go components  
**Scale/Scope**: ~1000 lines new code (leveraging existing components), support concurrent payments

**MCP Integration Points**:
  - Client: Implement transport.Interface to intercept JSON-RPC messages
  - Server: Wrap server.MCPServer's HTTP handler at the HTTP layer for x402 payment interception
  - Types: Use mcp.Tool, mcp.CallToolRequest, mcp.CallToolResult from mcp-go
  - Payment Flow: Inject payment in params._meta, extract from JSON-RPC error code 402

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- ✅ **No Unnecessary Documentation**: Only creating spec-related docs and API contracts
- ✅ **Test Coverage Preservation**: Will include comprehensive unit tests for all new code
- ✅ **Test-First Development**: Will write tests before implementing MCP transport and server
- ✅ **Stdlib-First Approach**: Using only mcp-go dependency (required for protocol), reusing x402-go
- ✅ **Code Conciseness**: Reusing existing signers, facilitator, types - minimal new code
- ✅ **Binary Cleanup**: Examples will be tested but binaries not committed

## Project Structure

### Documentation (this feature)

```text
specs/007-mcp-integration/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
mcp/
├── client/
│   ├── config.go            # Config struct with signers, callbacks, options
│   ├── transport.go         # Transport wraps transport.StreamableHTTP, overrides SendRequest
│   ├── transport_test.go    # Transport tests including payment flows
│   ├── handler.go           # Payment handler orchestration (reuses x402.Signer)
│   └── handler_test.go      # Handler unit tests
├── server/
│   ├── config.go            # Config struct with FacilitatorURL, VerifyOnly, Verbose, PaymentTools map
│   ├── server.go            # X402Server wraps server.MCPServer from mcp-go
│   ├── server_test.go       # Server tests
│   ├── handler.go           # X402Handler wraps http.Handler from server.NewStreamableHTTPServer
│   ├── handler_test.go      # Handler tests for HTTP interception
│   ├── facilitator.go       # Facilitator integration (reuses http.FacilitatorClient)
│   └── requirements.go      # Payment requirement builder helpers (RequireUSDCBase, etc.)
├── types.go                 # MCP-specific x402 types (minimal, reuses x402 types)
└── errors.go                # MCP-specific error types

examples/mcp/
├── main.go                  # Combined client/server example (like x402demo)
│                            # Client: wraps transport.StreamableHTTP with x402 Transport
│                            # Server: wraps server.NewStreamableHTTPServer with X402Handler
├── README.md                # Example documentation
└── go.mod                   # Example module file
```

**Wrapping Pattern**:
```
Client Side:
  mcp-go: transport.NewStreamableHTTP (implements transport.Interface)
     ↓ wrapped by
  x402-go: mcp/client.Transport (implements transport.Interface)
     ↓ overrides SendRequest to handle 402 errors
     ↓ passed to
  mcp-go: client.NewClient(transport.Interface)

Server Side:
  mcp-go: server.NewMCPServer (tool registry)
     ↓ passed to
  mcp-go: server.NewStreamableHTTPServer (returns http.Handler)
     ↓ wrapped by
  x402-go: mcp/server.X402Handler (implements http.Handler)
     ↓ intercepts HTTP POST requests, checks for tools/call
     ↓ serves via
  net/http: http.ListenAndServe(addr, handler)
```

**Key mcp-go Types Used**:

**Client Transport**:
- `transport.Interface`: Client transport interface with Start, SendRequest, SendNotification, Close, GetSessionId methods
- `transport.StreamableHTTP`: Concrete HTTP transport implementation that our x402 Transport wraps
- `transport.NewStreamableHTTP(serverURL, opts...)`: Creates StreamableHTTP transport
- `transport.JSONRPCRequest`: JSON-RPC request structure with Method, Params (any), ID fields
- `transport.JSONRPCResponse`: JSON-RPC response with Result (json.RawMessage), Error (*mcp.JSONRPCErrorDetails), ID fields

**Server**:
- `server.MCPServer`: Server base with AddTool, AddPrompt, AddResource methods
- `server.NewStreamableHTTPServer(mcpServer)`: Creates http.Handler from MCPServer for HTTP/SSE transport
- `server.ToolHandlerFunc`: Tool handler signature: func(ctx, mcp.CallToolRequest) (*mcp.CallToolResult, error)

**MCP Protocol Types**:
- `mcp.Tool`: Tool definition created via mcp.NewTool(name, opts...)
- `mcp.CallToolParams`: Tool params with Name (string), Arguments (map[string]any), Meta (*mcp.Meta) fields
- `mcp.CallToolRequest`: Tool invocation with Params.Arguments and Params.Meta fields
- `mcp.CallToolResult`: Tool response with Content ([]mcp.Content), StructuredContent, IsError fields
- `mcp.JSONRPCErrorDetails`: Error details with Code (int), Message (string), Data (any) fields
- `mcp.JSONRPCNotification`: Notification messages for server-to-client communication
- `mcp.Meta`: Metadata object with AdditionalFields map[string]any for storing x402/payment data

**Integration Pattern**:
- Client: Wrap `transport.StreamableHTTP` with x402 Transport, override SendRequest to intercept 402 errors
- Server: Wrap `server.NewStreamableHTTPServer` HTTP handler with X402Handler to intercept tool calls

**Structure Decision**: Single project structure with MCP subpackages. This follows Go conventions and allows clean separation between client and server components while maximizing code reuse from existing x402-go packages.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

*No violations - all constitution principles are satisfied.*
