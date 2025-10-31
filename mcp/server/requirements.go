package server

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/mcp"
)

var (
	// evmAddressRegex matches valid EVM addresses
	evmAddressRegex = regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)

	// Supported networks for validation
	supportedNetworks = map[string]bool{
		"base":          true,
		"base-sepolia":  true,
		"polygon":       true,
		"solana":        true,
		"solana-devnet": true,
	}
)

// Helper functions returning x402.PaymentRequirement for common configurations

// RequireUSDCBase creates a payment requirement for USDC on Base mainnet
func RequireUSDCBase(payTo, amount, description string) x402.PaymentRequirement {
	return x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "base",
		MaxAmountRequired: amount,
		Asset:             "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913", // USDC on Base
		PayTo:             payTo,
		Description:       description,
		MaxTimeoutSeconds: 60,
	}
}

// RequireUSDCBaseSepolia creates a payment requirement for USDC on Base Sepolia testnet
func RequireUSDCBaseSepolia(payTo, amount, description string) x402.PaymentRequirement {
	return x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "base-sepolia",
		MaxAmountRequired: amount,
		Asset:             "0x036cbd53842c5426634e7929541ec2318f3dcf7e", // USDC on Base Sepolia
		PayTo:             payTo,
		Description:       description,
		MaxTimeoutSeconds: 60,
	}
}

// RequireUSDCPolygon creates a payment requirement for USDC on Polygon
func RequireUSDCPolygon(payTo, amount, description string) x402.PaymentRequirement {
	return x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "polygon",
		MaxAmountRequired: amount,
		Asset:             "0x3c499c542cef5e3811e1192ce70d8cc03d5c3359", // USDC on Polygon
		PayTo:             payTo,
		Description:       description,
		MaxTimeoutSeconds: 60,
	}
}

// RequireUSDCSolana creates a payment requirement for USDC on Solana
func RequireUSDCSolana(payTo, amount, description string) x402.PaymentRequirement {
	return x402.PaymentRequirement{
		Scheme:            "exact",
		Network:           "solana",
		MaxAmountRequired: amount,
		Asset:             "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC on Solana
		PayTo:             payTo,
		Description:       description,
		MaxTimeoutSeconds: 60,
	}
}

// Validation helpers

// validateAmount checks if the amount is valid (greater than 0)
func validateAmount(amount string) error {
	if amount == "" {
		return errors.New("amount cannot be empty")
	}

	// Parse as big.Int to handle large values
	amt := new(big.Int)
	amt, ok := amt.SetString(amount, 10)
	if !ok {
		return fmt.Errorf("invalid amount format: %s", amount)
	}

	if amt.Sign() <= 0 {
		return fmt.Errorf("amount must be greater than 0, got: %s", amount)
	}

	return nil
}

// validateEVMAddress validates an EVM address format
func validateEVMAddress(address string) error {
	if address == "" {
		return errors.New("address cannot be empty")
	}

	if !evmAddressRegex.MatchString(address) {
		return fmt.Errorf("invalid EVM address format: %s (expected 0x followed by 40 hex characters)", address)
	}

	return nil
}

// validateNetwork checks if the network is supported
func validateNetwork(network string) error {
	if network == "" {
		return errors.New("network cannot be empty")
	}

	if !supportedNetworks[network] {
		return fmt.Errorf("unsupported network: %s (supported: base, base-sepolia, polygon, solana, solana-devnet)", network)
	}

	return nil
}

// ValidateRequirement validates a complete payment requirement
func ValidateRequirement(req x402.PaymentRequirement) error {
	// Validate amount
	if err := validateAmount(req.MaxAmountRequired); err != nil {
		return fmt.Errorf("invalid requirement: %w", err)
	}

	// Validate network
	if err := validateNetwork(req.Network); err != nil {
		return fmt.Errorf("invalid requirement: %w", err)
	}

	// Validate recipient address based on network
	if req.Network == "solana" || req.Network == "solana-devnet" {
		// Solana addresses are base58 encoded, just check not empty
		if req.PayTo == "" {
			return errors.New("invalid requirement: payTo address cannot be empty")
		}
	} else {
		// EVM networks
		if err := validateEVMAddress(req.PayTo); err != nil {
			return fmt.Errorf("invalid requirement: %w", err)
		}
	}

	// Validate scheme
	if req.Scheme != "exact" {
		return fmt.Errorf("invalid requirement: unsupported scheme %s (only 'exact' is supported)", req.Scheme)
	}

	// Validate asset address
	if req.Asset == "" {
		return errors.New("invalid requirement: asset address cannot be empty")
	}

	return nil
}

// SetToolResource sets the resource field for a payment requirement based on tool name
func SetToolResource(req *mcp.PaymentRequirement, toolName string) {
	if req != nil && toolName != "" {
		req.Resource = fmt.Sprintf("mcp://tools/%s", toolName)
	}
}
