# Chi x402 Example

This example demonstrates how to use the x402 payment protocol with Chi router.

## Features

- Chi router with x402 payment gating middleware
- Public and paywalled endpoints
- Payment verification and settlement
- Context-based payment information access

## Prerequisites

- Go 1.25.1 or later
- Chi router (installed automatically via go.mod)
- x402-go package

## Usage

### Running the Server

```bash
# Basic usage (testnet)
go run main.go --payTo YOUR_WALLET_ADDRESS

# With custom settings
go run main.go \
  --port 8080 \
  --network base-sepolia \
  --payTo 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0 \
  --amount 10000 \
  --facilitator https://facilitator.x402.rs
```

### Command Line Options

- `--port` - Server port (default: 8080)
- `--network` - Blockchain network (base, base-sepolia, solana, solana-devnet)
- `--payTo` - Your wallet address to receive payments (required)
- `--token` - Token contract address (auto-detected if not specified)
- `--amount` - Payment amount in atomic units (default: 1000 = 0.001 USDC)
- `--facilitator` - Facilitator service URL (default: https://facilitator.x402.rs)

## Endpoints

### Public Endpoints
- `GET /` - Server information
- `GET /public` - Free public endpoint (no payment required)

### Paywalled Endpoints
- `GET /data` - Premium content (requires x402 payment)

## Testing with curl

### Access Public Endpoint
```bash
curl http://localhost:8080/public
```

### Access Paywalled Endpoint (will return 402)
```bash
curl http://localhost:8080/data
```

You'll receive a 402 Payment Required response with payment requirements in JSON format.

### Access Paywalled Endpoint with Payment

Use the x402 client library to make payments:

```bash
# See the x402-go client examples for payment setup
# Example: Using the Gin client example
go run examples/gin/main.go client \
  --network base-sepolia \
  --key YOUR_PRIVATE_KEY \
  --url http://localhost:8080/data
```

## Code Example

```go
package main

import (
	"net/http"
	
	"github.com/go-chi/chi/v5"
	"github.com/mark3labs/x402-go"
	httpx402 "github.com/mark3labs/x402-go/http"
	chix402 "github.com/mark3labs/x402-go/http/chi"
)

func main() {
	// Create Chi router
	r := chi.NewRouter()
	
	// Configure x402 middleware
	config := &httpx402.Config{
		FacilitatorURL: "https://facilitator.x402.rs",
		PaymentRequirements: []x402.PaymentRequirement{{
			Scheme:            "exact",
			Network:           "base-sepolia",
			MaxAmountRequired: "10000",
			Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
			PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			MaxTimeoutSeconds: 60,
		}},
	}
	
	// Apply middleware to routes
	r.Use(chix402.NewChiX402Middleware(config))
	
	// Protected endpoint
	r.Get("/data", func(w http.ResponseWriter, r *http.Request) {
		// Access payment info from context
		payment := r.Context().Value(httpx402.PaymentContextKey).(*httpx402.VerifyResponse)
		w.Write([]byte("Access granted! Payer: " + payment.Payer))
	})
	
	http.ListenAndServe(":8080", r)
}
```

## Middleware Features

The Chi x402 middleware provides:

1. **Automatic Payment Verification**: Validates X-PAYMENT headers
2. **Payment Settlement**: Processes payments through facilitator
3. **Context Integration**: Stores payment info in request context
4. **CORS Support**: Automatically bypasses OPTIONS requests
5. **Error Handling**: Returns proper 402/400/503 responses
6. **Logging**: Structured logging via slog.Default()
7. **Verify-Only Mode**: Optional mode for testing without settlement

## Advanced Usage

### Per-Route Protection

```go
r := chi.NewRouter()

// Public routes
r.Get("/public", publicHandler)

// Protected routes
r.With(chix402.NewChiX402Middleware(config)).Get("/premium", premiumHandler)
```

### Route Groups

```go
r := chi.NewRouter()

// Protected group
r.Route("/api", func(r chi.Router) {
	r.Use(chix402.NewChiX402Middleware(config))
	r.Get("/data", dataHandler)
	r.Get("/analytics", analyticsHandler)
})
```

### Verify-Only Mode

```go
config := &httpx402.Config{
	FacilitatorURL: "https://facilitator.x402.rs",
	VerifyOnly:     true, // Skip settlement, only verify
	PaymentRequirements: []x402.PaymentRequirement{{...}},
}
```

## Payment Information

After successful payment verification, payment details are available in the request context:

```go
func handler(w http.ResponseWriter, r *http.Request) {
	payment := r.Context().Value(httpx402.PaymentContextKey).(*httpx402.VerifyResponse)
	
	log.Printf("Payment from: %s", payment.Payer)
	log.Printf("Valid: %v", payment.IsValid)
	
	// Your handler logic here
}
```

## Network Configuration

### Base Sepolia (Testnet)
- Network: `base-sepolia`
- USDC Token: `0x036CbD53842c5426634e7929541eC2318f3dCF7e`

### Base (Mainnet)
- Network: `base`
- USDC Token: `0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913`

### Solana (Mainnet)
- Network: `solana`
- USDC Token: `EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v`

### Solana Devnet (Testnet)
- Network: `solana-devnet`
- USDC Token: `4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU`

## Troubleshooting

### Server returns 503
- Check that the facilitator URL is reachable
- Verify network configuration matches payment requirements

### Payment always returns 402
- Ensure X-PAYMENT header is properly formatted (base64 JSON)
- Verify payment scheme and network match server configuration
- Check that payment signature is valid

### Context value is nil
- Ensure middleware is applied before the handler in Chi middleware stack
- Check that payment was successfully verified

## Related Examples

- [Gin Example](../gin/) - Gin framework with x402
- [Basic Example](../basic/) - Stdlib HTTP with x402

## Documentation

- [Chi Middleware Spec](../../specs/006-chi-middleware/spec.md)
- [Chi Middleware API](../../specs/006-chi-middleware/contracts/chi-middleware-api.yaml)
- [x402 Protocol](https://x402.org)
