package main

import (
	"log"
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"

	"github.com/mark3labs/x402-go"
	httpx402 "github.com/mark3labs/x402-go/http"
	pbx402 "github.com/mark3labs/x402-go/http/pocketbase"
)

func main() {
	app := pocketbase.New()

	// Configure payment requirements
	config := &httpx402.Config{
		FacilitatorURL: "https://api.x402.coinbase.com", // Production facilitator - replace if using your own
		PaymentRequirements: []x402.PaymentRequirement{{
			Scheme:            "exact",
			Network:           "base-sepolia",                               // TESTNET - change to "base" for production
			MaxAmountRequired: "10000",                                      // 0.01 USDC (USDC has 6 decimals: 1 USDC = 1,000,000 atomic units)
			Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e", // USDC on base-sepolia - use production USDC for mainnet
			PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", // REPLACE WITH YOUR WALLET ADDRESS
			MaxTimeoutSeconds: 300,
		}},
	}

	// Register middleware
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Create middleware
		middleware := pbx402.NewPocketBaseX402Middleware(config)

		// Protect a single route
		se.Router.GET("/api/premium/data", func(e *core.RequestEvent) error {
			// Access payment details (optional)
			payment := e.Get("x402_payment").(*httpx402.VerifyResponse)

			return e.JSON(http.StatusOK, map[string]any{
				"data":  "Premium content here",
				"payer": payment.Payer,
			})
		}).BindFunc(middleware)

		// Example: Protect a group of routes
		premiumGroup := se.Router.Group("/api/premium")
		premiumGroup.BindFunc(middleware) // Apply to all routes in group

		// All these routes require payment
		premiumGroup.GET("/reports", func(e *core.RequestEvent) error {
			return e.JSON(http.StatusOK, map[string]any{
				"report": "Premium report data",
			})
		})

		premiumGroup.GET("/analytics", func(e *core.RequestEvent) error {
			return e.JSON(http.StatusOK, map[string]any{
				"analytics": "Premium analytics data",
			})
		})

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
