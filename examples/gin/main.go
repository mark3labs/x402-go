package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/x402-go"
	httpx402 "github.com/mark3labs/x402-go/http"
	ginx402 "github.com/mark3labs/x402-go/http/gin"
)

func main() {
	// Create Gin router
	r := gin.Default()

	// Example 1: Basic payment protection
	// Protect all routes with 0.01 USDC payment requirement
	basicConfig := &httpx402.Config{
		FacilitatorURL: "https://api.x402.coinbase.com",
		PaymentRequirements: []x402.PaymentRequirement{
			{
				Scheme:            "exact",
				Network:           "base-sepolia", // Testnet
				MaxAmountRequired: "10000",        // 0.01 USDC (6 decimals)
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
				Description:       "Payment for API access",
				MimeType:          "application/json",
				MaxTimeoutSeconds: 300,
			},
		},
	}

	// Apply middleware to all routes
	r.Use(ginx402.NewGinX402Middleware(basicConfig))

	// Example 2: Public endpoints (no payment required)
	public := r.Group("/public")
	{
		public.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "healthy",
				"service": "x402 Gin Example",
			})
		})
	}

	// Example 3: Protected endpoints with payment
	protected := r.Group("/protected")
	{
		// This endpoint requires payment (middleware applied globally)
		protected.GET("/data", func(c *gin.Context) {
			// Access payment information from context
			if paymentInfo, exists := c.Get("x402_payment"); exists {
				verifyResp := paymentInfo.(*httpx402.VerifyResponse)
				c.JSON(http.StatusOK, gin.H{
					"message": "Access granted with valid payment",
					"payer":   verifyResp.Payer,
					"data":    "This is protected data",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Payment information not found",
				})
			}
		})
	}

	// Example 4: Verify-only mode (skip settlement)
	verifyOnlyConfig := &httpx402.Config{
		FacilitatorURL: "https://api.x402.coinbase.com",
		VerifyOnly:     true, // Only verify, don't settle
		PaymentRequirements: []x402.PaymentRequirement{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "5000", // 0.005 USDC
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
				Description:       "Payment verification only",
				MaxTimeoutSeconds: 60,
			},
		},
	}

	verifyOnly := r.Group("/verify-only")
	verifyOnly.Use(ginx402.NewGinX402Middleware(verifyOnlyConfig))
	{
		verifyOnly.GET("/check", func(c *gin.Context) {
			if paymentInfo, exists := c.Get("x402_payment"); exists {
				verifyResp := paymentInfo.(*httpx402.VerifyResponse)
				c.JSON(http.StatusOK, gin.H{
					"message": "Payment verified (not settled)",
					"payer":   verifyResp.Payer,
				})
			}
		})
	}

	// Example 5: Route-specific payment requirements
	premium := r.Group("/premium")
	premiumConfig := &httpx402.Config{
		FacilitatorURL: "https://api.x402.coinbase.com",
		PaymentRequirements: []x402.PaymentRequirement{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "50000", // 0.05 USDC
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
				Description:       "Premium feature access",
				MaxTimeoutSeconds: 600,
			},
		},
	}
	premium.Use(ginx402.NewGinX402Middleware(premiumConfig))
	{
		premium.GET("/analytics", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Premium analytics data",
				"data":    []string{"metric1", "metric2", "metric3"},
			})
		})
	}

	// Start server
	log.Println("Starting Gin server with x402 payment middleware on :8080")
	log.Println("Try accessing:")
	log.Println("  - /public/status (no payment required)")
	log.Println("  - /protected/data (requires 0.01 USDC payment)")
	log.Println("  - /verify-only/check (verify-only mode)")
	log.Println("  - /premium/analytics (requires 0.05 USDC payment)")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
