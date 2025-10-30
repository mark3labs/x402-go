package coinbase

import (
	"context"
	"fmt"

	"github.com/mark3labs/x402-go"
)

// CDPAccount represents a blockchain wallet account managed by the Coinbase Developer Platform.
// Each account corresponds to a unique address on a specific blockchain network (EVM or SVM).
//
// CDPAccount is immutable after creation - all fields are set during account creation
// via the CDP API and never modified.
//
// Example:
//
//	// EVM Account
//	account := &CDPAccount{
//	    ID:      "accounts/abc-123",
//	    Address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
//	    Network: "base-sepolia",
//	}
//
//	// SVM Account
//	account := &CDPAccount{
//	    ID:      "accounts/def-456",
//	    Address: "DYw8jCTfwHNRJhhmFcbXvVDTqWMEVFBX6ZKUmG5CNSKK",
//	    Network: "solana-devnet",
//	}
type CDPAccount struct {
	// ID is the CDP-internal account identifier (e.g., "accounts/abc-123")
	ID string `json:"id"`

	// Address is the blockchain address:
	//   - EVM: 0x-prefixed hex address (42 characters)
	//   - SVM: base58-encoded public key (32-44 characters)
	Address string `json:"address"`

	// Network is the CDP network identifier (e.g., "base-sepolia", "solana-devnet")
	Network string `json:"network"`
}

// CreateAccountRequest represents the request body for creating a new CDP account.
// The Name field is optional but recommended to avoid creating duplicate unnamed accounts.
type CreateAccountRequest struct {
	// Name is an optional identifier for the account (2-36 alphanumeric chars + hyphens)
	// Must be unique across all accounts in the CDP project
	Name string `json:"name,omitempty"`
}

// AccountResponse represents a single account in CDP API responses for list operations.
type AccountResponse struct {
	// ID is the CDP-internal account identifier
	ID string `json:"id"`

	// Address is the blockchain address
	Address string `json:"address"`

	// Network is the CDP network identifier
	Network string `json:"network"`

	// Name is the optional account identifier
	Name string `json:"name,omitempty"`
}

// CreateAccountResponse represents the response from creating a new account.
// Note: CDP API returns address, name (if provided), and timestamps, but no id or network.
type CreateAccountResponse struct {
	// Address is the blockchain address
	Address string `json:"address"`

	// Name is the account identifier (if provided during creation)
	Name string `json:"name,omitempty"`

	// CreatedAt is the timestamp when the account was created
	CreatedAt string `json:"createdAt"`

	// UpdatedAt is the timestamp when the account was last updated
	UpdatedAt string `json:"updatedAt"`
}

// ListAccountsResponse represents the response from listing existing CDP accounts.
// The Accounts field contains all accounts accessible with the current credentials.
type ListAccountsResponse struct {
	// Accounts is the list of existing accounts
	Accounts []AccountResponse `json:"accounts"`
}

// CreateOrGetAccount creates or retrieves a CDP account for the specified x402 network.
// This function implements a GET-then-POST pattern to ensure idempotency:
//
//  1. Attempts to retrieve existing accounts for the network type via GET request
//  2. If an account with the given name exists for the target network, returns it
//  3. If no matching account exists, creates a new one with the given name via POST request
//  4. Returns the created or retrieved account
//
// The function automatically maps x402 network names to CDP network identifiers and
// determines the appropriate API endpoints based on network type (EVM or SVM).
//
// Parameters:
//   - ctx: Request context for timeout and cancellation
//   - client: Configured CDP API client with authentication
//   - x402Network: x402 network identifier (e.g., "base", "base-sepolia", "solana", "solana-devnet")
//   - accountName: Unique identifier for the account (2-36 alphanumeric chars + hyphens, required)
//
// Returns:
//   - *CDPAccount on success with ID, Address, and Network populated
//   - x402.ErrInvalidNetwork if the network is not supported
//   - CDPError if the CDP API returns an error
//   - Standard error for network or serialization failures
//
// The function is idempotent - calling it multiple times with the same parameters
// returns the same account without creating duplicates.
//
// Example usage:
//
//	auth, _ := NewCDPAuth(apiKeyName, apiKeySecret, walletSecret)
//	client := NewCDPClient(auth)
//	account, err := CreateOrGetAccount(ctx, client, "base-sepolia", "my-payment-wallet")
//	if err != nil {
//	    log.Fatalf("Failed to create/get account: %v", err)
//	}
//	fmt.Printf("Account address: %s\n", account.Address)
//
// Supported networks:
//   - EVM: base, base-sepolia, ethereum, sepolia
//   - SVM: solana, mainnet-beta, solana-devnet, devnet
func CreateOrGetAccount(ctx context.Context, client *CDPClient, x402Network string, accountName string) (*CDPAccount, error) {
	// Validate account name
	if accountName == "" {
		return nil, fmt.Errorf("account name is required")
	}
	if len(accountName) < 2 || len(accountName) > 36 {
		return nil, fmt.Errorf("account name must be between 2 and 36 characters")
	}

	// Map x402 network to CDP network identifier
	cdpNetwork, err := getCDPNetwork(x402Network)
	if err != nil {
		return nil, err
	}

	// Determine network type (EVM or SVM)
	networkType := getNetworkType(x402Network)
	if networkType == NetworkTypeUnknown {
		return nil, fmt.Errorf("%w: %s", x402.ErrInvalidNetwork, x402Network)
	}

	// Determine API endpoint based on network type
	var listEndpoint, createEndpoint string
	switch networkType {
	case NetworkTypeEVM:
		listEndpoint = "/platform/v2/evm/accounts"
		createEndpoint = "/platform/v2/evm/accounts"
	case NetworkTypeSVM:
		listEndpoint = "/platform/v2/solana/accounts"
		createEndpoint = "/platform/v2/solana/accounts"
	default:
		return nil, fmt.Errorf("%w: %s", x402.ErrInvalidNetwork, x402Network)
	}

	// First, try to retrieve existing accounts
	// Note: Listing accounts does NOT require Wallet Auth (read-only operation)
	var listResp ListAccountsResponse
	err = client.doRequestWithRetry(ctx, "GET", listEndpoint, nil, &listResp, false)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	// Check if an account with this name already exists for this network
	for _, account := range listResp.Accounts {
		if account.Network == cdpNetwork && account.Name == accountName {
			// Account exists with matching name and network, return it
			return &CDPAccount{
				ID:      account.ID,
				Address: account.Address,
				Network: account.Network,
			}, nil
		}
	}

	// No existing account found, create a new one with the given name
	// Note: Creating accounts REQUIRES Wallet Auth (sensitive operation)
	createReq := CreateAccountRequest{
		Name: accountName,
	}
	var createResp CreateAccountResponse
	err = client.doRequestWithRetry(ctx, "POST", createEndpoint, createReq, &createResp, true)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	// Validate response
	if createResp.Address == "" {
		return nil, fmt.Errorf("CDP API returned empty account address")
	}

	// After creating the account, we need to fetch its ID by listing accounts again
	// The create response doesn't include the account ID, which is required for signing operations
	err = client.doRequestWithRetry(ctx, "GET", listEndpoint, nil, &listResp, false)
	if err != nil {
		return nil, fmt.Errorf("list accounts after create: %w", err)
	}

	// Find the newly created account by name and network
	for _, account := range listResp.Accounts {
		if account.Network == cdpNetwork && account.Name == accountName {
			return &CDPAccount{
				ID:      account.ID,
				Address: account.Address,
				Network: account.Network,
			}, nil
		}
	}

	// This should not happen - we just created the account
	return nil, fmt.Errorf("failed to retrieve account ID after creation")
}
