// Package main demonstrates basic usage of the x402 chain helpers and constants.
package main

import (
	"fmt"
	"log"

	"github.com/mark3labs/x402-go"
)

func main() {
	fmt.Println("=== x402 Chain Helpers Demo ===")

	// Example 1: User Story 1 - Client Setup with Chain Constants
	fmt.Println("1. Client Setup with Chain Constants (User Story 1)")
	fmt.Println("   Creating token config for Base mainnet...")
	baseToken := x402.NewUSDCTokenConfig(x402.BaseMainnet, 1)
	fmt.Printf("   ✓ Token Address: %s\n", baseToken.Address)
	fmt.Printf("   ✓ Symbol: %s\n", baseToken.Symbol)
	fmt.Printf("   ✓ Decimals: %d\n", baseToken.Decimals)
	fmt.Printf("   ✓ Priority: %d\n\n", baseToken.Priority)

	// Example 2: User Story 2 - Middleware Setup with Payment Requirements
	fmt.Println("2. Middleware Payment Requirement (User Story 2)")
	fmt.Println("   Creating payment requirement for Base mainnet: 1.50 USDC...")
	req, err := x402.NewUSDCPaymentRequirement(x402.USDCRequirementConfig{
		Chain:            x402.BaseMainnet,
		Amount:           "1.50",
		RecipientAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
	})
	if err != nil {
		log.Fatalf("Failed to create payment requirement: %v", err)
	}
	fmt.Printf("   ✓ Network: %s\n", req.Network)
	fmt.Printf("   ✓ Asset: %s\n", req.Asset)
	fmt.Printf("   ✓ Amount (atomic): %s (1.5 USDC = 1,500,000 atomic units)\n", req.MaxAmountRequired)
	fmt.Printf("   ✓ Scheme: %s\n", req.Scheme)
	fmt.Printf("   ✓ Timeout: %d seconds\n", req.MaxTimeoutSeconds)
	fmt.Printf("   ✓ MimeType: %s\n", req.MimeType)
	if req.Extra != nil {
		fmt.Printf("   ✓ EIP-3009 Domain: name=%s, version=%s\n\n", req.Extra["name"], req.Extra["version"])
	}

	// Example 3: User Story 3 - Multi-Chain Client with Priorities
	fmt.Println("3. Multi-Chain Token Configuration (User Story 3)")
	fmt.Println("   Creating token configs for 3 chains with different priorities...")
	tokens := []x402.TokenConfig{
		x402.NewUSDCTokenConfig(x402.BaseMainnet, 1),    // Priority 1 (highest)
		x402.NewUSDCTokenConfig(x402.PolygonMainnet, 2), // Priority 2
		x402.NewUSDCTokenConfig(x402.SolanaMainnet, 3),  // Priority 3
	}
	for _, token := range tokens {
		fmt.Printf("   ✓ %s: Address=%s, Priority=%d\n",
			getNetworkName(token.Address), token.Address[:10]+"...", token.Priority)
	}
	fmt.Println()

	// Example 4: User Story 4 - Network Validation
	fmt.Println("4. Network Validation (User Story 4)")
	networks := []string{"base", "solana", "polygon-amoy", "unknown-chain"}
	for _, netID := range networks {
		netType, err := x402.ValidateNetwork(netID)
		if err != nil {
			fmt.Printf("   ✗ %s: %v\n", netID, err)
		} else {
			typeName := "Unknown"
			switch netType {
			case x402.NetworkTypeEVM:
				typeName = "EVM"
			case x402.NetworkTypeSVM:
				typeName = "SVM"
			}
			fmt.Printf("   ✓ %s: %s chain\n", netID, typeName)
		}
	}
	fmt.Println()

	// Example 5: Zero Amount (Free-with-Signature)
	fmt.Println("5. Zero Amount Authorization (Free-with-Signature)")
	fmt.Println("   Creating payment requirement with zero amount...")
	freeReq, err := x402.NewUSDCPaymentRequirement(x402.USDCRequirementConfig{
		Chain:            x402.BaseSepolia,
		Amount:           "0",
		RecipientAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
	})
	if err != nil {
		log.Fatalf("Failed to create free payment requirement: %v", err)
	}
	fmt.Printf("   ✓ Network: %s\n", freeReq.Network)
	fmt.Printf("   ✓ Amount: %s (zero - signature required, no payment)\n\n", freeReq.MaxAmountRequired)

	// Example 6: Custom Configuration
	fmt.Println("6. Custom Payment Configuration")
	fmt.Println("   Creating payment requirement with custom settings...")
	customReq, err := x402.NewUSDCPaymentRequirement(x402.USDCRequirementConfig{
		Chain:             x402.PolygonMainnet,
		Amount:            "5.0",
		RecipientAddress:  "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		Scheme:            "estimate",
		MaxTimeoutSeconds: 600,
		MimeType:          "text/plain",
	})
	if err != nil {
		log.Fatalf("Failed to create custom payment requirement: %v", err)
	}
	fmt.Printf("   ✓ Scheme: %s (overridden from default 'exact')\n", customReq.Scheme)
	fmt.Printf("   ✓ Timeout: %d seconds (overridden from default 300)\n", customReq.MaxTimeoutSeconds)
	fmt.Printf("   ✓ MimeType: %s (overridden from default 'application/json')\n\n", customReq.MimeType)

	// Example 7: All Available Chain Constants
	fmt.Println("7. All Available Chain Constants")
	fmt.Println("   Mainnet Chains:")
	fmt.Printf("   - Solana:    %s (NetworkID: %s)\n", x402.SolanaMainnet.USDCAddress, x402.SolanaMainnet.NetworkID)
	fmt.Printf("   - Base:      %s (NetworkID: %s)\n", x402.BaseMainnet.USDCAddress, x402.BaseMainnet.NetworkID)
	fmt.Printf("   - Polygon:   %s (NetworkID: %s)\n", x402.PolygonMainnet.USDCAddress, x402.PolygonMainnet.NetworkID)
	fmt.Printf("   - Avalanche: %s (NetworkID: %s)\n", x402.AvalancheMainnet.USDCAddress, x402.AvalancheMainnet.NetworkID)
	fmt.Println()
	fmt.Println("   Testnet Chains:")
	fmt.Printf("   - Solana Devnet:   %s (NetworkID: %s)\n", x402.SolanaDevnet.USDCAddress, x402.SolanaDevnet.NetworkID)
	fmt.Printf("   - Base Sepolia:    %s (NetworkID: %s)\n", x402.BaseSepolia.USDCAddress, x402.BaseSepolia.NetworkID)
	fmt.Printf("   - Polygon Amoy:    %s (NetworkID: %s)\n", x402.PolygonAmoy.USDCAddress, x402.PolygonAmoy.NetworkID)
	fmt.Printf("   - Avalanche Fuji:  %s (NetworkID: %s)\n", x402.AvalancheFuji.USDCAddress, x402.AvalancheFuji.NetworkID)
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
	fmt.Println("\nAll USDC addresses and EIP-3009 parameters verified as of 2025-10-28")
	fmt.Println("For more examples, see quickstart.md in specs/003-helpers-constants/")
}

// getNetworkName returns a human-readable network name from a USDC address
func getNetworkName(address string) string {
	switch address {
	case x402.BaseMainnet.USDCAddress:
		return "Base"
	case x402.PolygonMainnet.USDCAddress:
		return "Polygon"
	case x402.SolanaMainnet.USDCAddress:
		return "Solana"
	default:
		return "Unknown"
	}
}
