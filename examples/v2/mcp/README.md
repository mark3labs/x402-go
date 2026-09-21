# x402 v2 MCP Client & Server Example

This example demonstrates how to implement an MCP (Model Context Protocol) server and client with built-in x402 v2 payments across EVM (Base) and SVM (Solana) chains using CAIP-2 network identifiers.

---

## 1. Running the Local Test Server

Start an MCP server with paywalled tools on Base Sepolia:

```bash
go run main.go server \
  --port 8080 \
  --network eip155:84532 \
  --pay-to 0xYourWalletAddress \
  --amount 10000 \
  --verbose
```

Flags:
- `--port`: Server listen port (default `8080`).
- `--network`: CAIP-2 network (`eip155:8453` for Base mainnet, `eip155:84532` for Base Sepolia, `solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp` for Solana mainnet).
- `--pay-to`: Receiving wallet address (required).
- `--amount`: Price per tool execution in atomic units (e.g. `1000` = $0.001 USDC, `10000` = $0.01 USDC).
- `--facilitator`: x402 facilitator endpoint (default `https://facilitator.x402.rs`).
- `--verify-only`: When `true`, verifies signatures without executing on-chain settlement.

---

## 2. Running the Client Against a Server

The client initializes an MCP session, lists available tools, calls a free tool, and then calls a paid tool where payment is signed and handled automatically.

### A. Calling Local Server (Base Sepolia)

```bash
go run main.go client \
  --server http://localhost:8080 \
  --network eip155:84532 \
  --key "0xYOUR_HEX_PRIVATE_KEY" \
  --verbose
```

### B. Calling an Independent Base Mainnet Endpoint (Integration Recipe)

To test client-side x402 v2 negotiation, EIP-3009 transfer authorization, and tool execution against an independent live mainnet endpoint:

* **Live Mainnet MCP Host:** `https://agentground.atlether.trade/mcp`
* **Live Direct REST Route:** `POST https://agentground.atlether.trade/api/v1/x402/validate-contract`
* **Network:** Base (`eip155:8453`)
* **Price:** $0.001 USDC (`1000` atomic units)
* **Pay-To Address:** `0xd782Be1135bD14569e829897D2f76F723F412b1a`

Run the client:

```bash
go run main.go client \
  --server https://agentground.atlether.trade/mcp \
  --network eip155:8453 \
  --key "$BASE_MAINNET_PRIVATE_KEY" \
  --verbose
```

What happens under the hood:
1. Client initializes session via `initialize` and lists tools via `tools/list`.
2. Invoking the tool triggers HTTP 402 with a Base64 `PAYMENT-REQUIRED` header.
3. Client extracts `accepts[]`, constructs an EIP-3009 `TransferWithAuthorization` using the local EVM signer on Base.
4. Client resubmits the request with `PAYMENT-SIGNATURE` header.
5. Server verifies signature, executes preflight tool contract validation, and returns the result with HTTP 200.
