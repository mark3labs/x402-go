package server

import (
	"github.com/mark3labs/x402-go"
)

// Payment requirement builder helpers for common networks and tokens

// RequireUSDCBase creates a payment requirement for USDC on Base network
func RequireUSDCBase(payTo, amount, description string) x402.PaymentRequirement {
	return x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "base",
		MaxAmountRequired: amount,
		Asset:             "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913", // USDC on Base
		PayTo:             payTo,
		Resource:          "", // Will be set by middleware
		Description:       description,
		MaxTimeoutSeconds: 300, // 5 minutes
	}
}

// RequireUSDCPolygon creates a payment requirement for USDC on Polygon network
func RequireUSDCPolygon(payTo, amount, description string) x402.PaymentRequirement {
	return x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "polygon",
		MaxAmountRequired: amount,
		Asset:             "0x3c499c542cef5e3811e1192ce70d8cc03d5c3359", // USDC on Polygon
		PayTo:             payTo,
		Resource:          "", // Will be set by middleware
		Description:       description,
		MaxTimeoutSeconds: 300,
	}
}

// RequireUSDCSolana creates a payment requirement for USDC on Solana network
func RequireUSDCSolana(payTo, amount, description string) x402.PaymentRequirement {
	return x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "solana",
		MaxAmountRequired: amount,
		Asset:             "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC on Solana
		PayTo:             payTo,
		Resource:          "", // Will be set by middleware
		Description:       description,
		MaxTimeoutSeconds: 300,
	}
}

// RequireUSDCBaseSepolia creates a payment requirement for USDC on Base Sepolia testnet
func RequireUSDCBaseSepolia(payTo, amount, description string) x402.PaymentRequirement {
	return x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "base-sepolia",
		MaxAmountRequired: amount,
		Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e", // USDC on Base Sepolia
		PayTo:             payTo,
		Resource:          "", // Will be set by middleware
		Description:       description,
		MaxTimeoutSeconds: 300,
	}
}

// RequirePayment creates a custom payment requirement
func RequirePayment(network, asset, payTo, amount, description string) x402.PaymentRequirement {
	return x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           network,
		MaxAmountRequired: amount,
		Asset:             asset,
		PayTo:             payTo,
		Resource:          "", // Will be set by middleware
		Description:       description,
		MaxTimeoutSeconds: 300,
	}
}
