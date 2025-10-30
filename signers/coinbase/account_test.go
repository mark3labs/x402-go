package coinbase

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/x402-go"
)

// TestCreateOrGetAccount_CreateNew tests creating a new account when none exists.
// This covers T034: Contract test for CreateOrGetAccount success case.
func TestCreateOrGetAccount_CreateNew(t *testing.T) {
	tests := []struct {
		name           string
		x402Network    string
		cdpNetwork     string
		listEndpoint   string
		createEndpoint string
		wantAddress    string
	}{
		{
			name:           "create new EVM account on Base Sepolia",
			x402Network:    "base-sepolia",
			cdpNetwork:     "base-sepolia",
			listEndpoint:   "/platform/v2/evm/accounts",
			createEndpoint: "/platform/v2/evm/accounts",
			wantAddress:    "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		},
		{
			name:           "create new EVM account on Base mainnet",
			x402Network:    "base",
			cdpNetwork:     "base-mainnet",
			listEndpoint:   "/platform/v2/evm/accounts",
			createEndpoint: "/platform/v2/evm/accounts",
			wantAddress:    "0x1234567890123456789012345678901234567890",
		},
		{
			name:           "create new EVM account on Ethereum mainnet",
			x402Network:    "ethereum",
			cdpNetwork:     "ethereum-mainnet",
			listEndpoint:   "/platform/v2/evm/accounts",
			createEndpoint: "/platform/v2/evm/accounts",
			wantAddress:    "0xabcdef0123456789abcdef0123456789abcdef01",
		},
		{
			name:           "create new SVM account on Solana devnet",
			x402Network:    "solana-devnet",
			cdpNetwork:     "solana-devnet",
			listEndpoint:   "/platform/v2/solana/accounts",
			createEndpoint: "/platform/v2/solana/accounts",
			wantAddress:    "DYw8jCTfwHNRJhhmFcbXvVDTqWMEVFBX6ZKUmG5CNSKK",
		},
		{
			name:           "create new SVM account on Solana mainnet",
			x402Network:    "solana",
			cdpNetwork:     "solana-mainnet",
			listEndpoint:   "/platform/v2/solana/accounts",
			createEndpoint: "/platform/v2/solana/accounts",
			wantAddress:    "9B5XszUGdMaxCZ7uSQhPzdks5ZQSmWxrmzCSvtJ6Ns6g",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track which endpoints were called
			var listCalled, createCalled bool

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == "GET" && r.URL.Path == tt.listEndpoint:
					// List accounts - return empty list
					listCalled = true
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(ListAccountsResponse{
						Accounts: []AccountResponse{},
					})

				case r.Method == "POST" && r.URL.Path == tt.createEndpoint:
					// Create account - verify request body and return new account
					createCalled = true

					var req CreateAccountRequest
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						t.Errorf("Failed to decode request body: %v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					if req.NetworkID != tt.cdpNetwork {
						t.Errorf("Expected network_id=%s, got %s", tt.cdpNetwork, req.NetworkID)
					}

					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(AccountResponse{
						ID:      "accounts/test-account-123",
						Address: tt.wantAddress,
						Network: tt.cdpNetwork,
					})

				default:
					t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			auth := &mockCDPAuth{}
			client := NewCDPClient(auth)
			client.baseURL = server.URL

			account, err := CreateOrGetAccount(context.Background(), client, tt.x402Network)
			if err != nil {
				t.Fatalf("CreateOrGetAccount failed: %v", err)
			}

			// Verify both list and create were called (GET-then-POST pattern)
			if !listCalled {
				t.Error("Expected list accounts endpoint to be called")
			}
			if !createCalled {
				t.Error("Expected create account endpoint to be called")
			}

			// Verify account details
			if account.ID != "accounts/test-account-123" {
				t.Errorf("Expected account ID=accounts/test-account-123, got %s", account.ID)
			}
			if account.Address != tt.wantAddress {
				t.Errorf("Expected address=%s, got %s", tt.wantAddress, account.Address)
			}
			if account.Network != tt.cdpNetwork {
				t.Errorf("Expected network=%s, got %s", tt.cdpNetwork, account.Network)
			}
		})
	}
}

// TestCreateOrGetAccount_ExistingAccount tests retrieving an existing account.
// This covers T035: Contract test for CreateOrGetAccount with existing account.
func TestCreateOrGetAccount_ExistingAccount(t *testing.T) {
	tests := []struct {
		name             string
		x402Network      string
		cdpNetwork       string
		listEndpoint     string
		existingAccounts []AccountResponse
		wantAccountID    string
		wantAddress      string
	}{
		{
			name:         "retrieve existing EVM account",
			x402Network:  "base-sepolia",
			cdpNetwork:   "base-sepolia",
			listEndpoint: "/platform/v2/evm/accounts",
			existingAccounts: []AccountResponse{
				{
					ID:      "accounts/existing-base-123",
					Address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
					Network: "base-sepolia",
				},
			},
			wantAccountID: "accounts/existing-base-123",
			wantAddress:   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		},
		{
			name:         "retrieve existing SVM account",
			x402Network:  "solana-devnet",
			cdpNetwork:   "solana-devnet",
			listEndpoint: "/platform/v2/solana/accounts",
			existingAccounts: []AccountResponse{
				{
					ID:      "accounts/existing-solana-456",
					Address: "DYw8jCTfwHNRJhhmFcbXvVDTqWMEVFBX6ZKUmG5CNSKK",
					Network: "solana-devnet",
				},
			},
			wantAccountID: "accounts/existing-solana-456",
			wantAddress:   "DYw8jCTfwHNRJhhmFcbXvVDTqWMEVFBX6ZKUmG5CNSKK",
		},
		{
			name:         "select correct account from multiple",
			x402Network:  "base-sepolia",
			cdpNetwork:   "base-sepolia",
			listEndpoint: "/platform/v2/evm/accounts",
			existingAccounts: []AccountResponse{
				{
					ID:      "accounts/ethereum-mainnet-111",
					Address: "0x1111111111111111111111111111111111111111",
					Network: "ethereum-mainnet",
				},
				{
					ID:      "accounts/base-sepolia-222",
					Address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
					Network: "base-sepolia",
				},
				{
					ID:      "accounts/base-mainnet-333",
					Address: "0x3333333333333333333333333333333333333333",
					Network: "base-mainnet",
				},
			},
			wantAccountID: "accounts/base-sepolia-222",
			wantAddress:   "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var listCalled, createCalled bool

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == "GET" && r.URL.Path == tt.listEndpoint:
					// List accounts - return existing accounts
					listCalled = true
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(ListAccountsResponse{
						Accounts: tt.existingAccounts,
					})

				case r.Method == "POST":
					// Create should NOT be called when account exists
					createCalled = true
					t.Error("Create account endpoint should not be called when account exists")
					w.WriteHeader(http.StatusBadRequest)

				default:
					t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			auth := &mockCDPAuth{}
			client := NewCDPClient(auth)
			client.baseURL = server.URL

			account, err := CreateOrGetAccount(context.Background(), client, tt.x402Network)
			if err != nil {
				t.Fatalf("CreateOrGetAccount failed: %v", err)
			}

			// Verify only list was called (no create)
			if !listCalled {
				t.Error("Expected list accounts endpoint to be called")
			}
			if createCalled {
				t.Error("Create account endpoint should not be called when account exists")
			}

			// Verify correct account was returned
			if account.ID != tt.wantAccountID {
				t.Errorf("Expected account ID=%s, got %s", tt.wantAccountID, account.ID)
			}
			if account.Address != tt.wantAddress {
				t.Errorf("Expected address=%s, got %s", tt.wantAddress, account.Address)
			}
			if account.Network != tt.cdpNetwork {
				t.Errorf("Expected network=%s, got %s", tt.cdpNetwork, account.Network)
			}
		})
	}
}

// TestCreateOrGetAccount_InvalidCredentials tests handling of authentication failures.
// This covers T036: Test CreateOrGetAccount with invalid credentials.
func TestCreateOrGetAccount_InvalidCredentials(t *testing.T) {
	tests := []struct {
		name        string
		x402Network string
		statusCode  int
		wantError   string
	}{
		{
			name:        "401 unauthorized",
			x402Network: "base-sepolia",
			statusCode:  http.StatusUnauthorized,
			wantError:   "Authentication failed",
		},
		{
			name:        "403 forbidden",
			x402Network: "solana-devnet",
			statusCode:  http.StatusForbidden,
			wantError:   "Insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(map[string]string{
					"error": tt.wantError,
				})
			}))
			defer server.Close()

			auth := &mockCDPAuth{}
			client := NewCDPClient(auth)
			client.baseURL = server.URL

			account, err := CreateOrGetAccount(context.Background(), client, tt.x402Network)
			if err == nil {
				t.Fatal("Expected error for invalid credentials, got nil")
			}
			if account != nil {
				t.Errorf("Expected nil account on error, got %+v", account)
			}

			// Verify error contains a CDPError with correct status code
			if !strings.Contains(err.Error(), "CDP API error") {
				t.Fatalf("Expected error to contain CDP API error, got: %v", err)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("Expected error to contain %q, got: %v", tt.wantError, err)
			}
			// The error is wrapped by fmt.Errorf in account.go, so we check the message contains the status code
			errMsg := err.Error()
			expectedCode := ""
			if tt.statusCode == http.StatusUnauthorized {
				expectedCode = "[401]"
			} else if tt.statusCode == http.StatusForbidden {
				expectedCode = "[403]"
			}
			if !strings.Contains(errMsg, expectedCode) {
				t.Errorf("Expected status code %s in error message, got: %v", expectedCode, err)
			}
		})
	}
}

// TestCreateOrGetAccount_UnsupportedNetwork tests handling of unsupported networks.
// This covers T037: Test CreateOrGetAccount with unsupported network.
func TestCreateOrGetAccount_UnsupportedNetwork(t *testing.T) {
	tests := []struct {
		name        string
		x402Network string
	}{
		{
			name:        "completely invalid network",
			x402Network: "invalid-network",
		},
		{
			name:        "unsupported blockchain",
			x402Network: "polygon",
		},
		{
			name:        "empty network name",
			x402Network: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// No server needed - should fail before making API call
			auth := &mockCDPAuth{}
			client := NewCDPClient(auth)

			account, err := CreateOrGetAccount(context.Background(), client, tt.x402Network)
			if err == nil {
				t.Fatal("Expected error for unsupported network, got nil")
			}
			if account != nil {
				t.Errorf("Expected nil account on error, got %+v", account)
			}

			// Verify error is ErrInvalidNetwork
			if !strings.Contains(err.Error(), x402.ErrInvalidNetwork.Error()) {
				t.Errorf("Expected ErrInvalidNetwork, got: %v", err)
			}
		})
	}
}

// TestCreateOrGetAccount_Idempotency tests that repeated calls return the same account.
// This covers T038: Test CreateOrGetAccount idempotency (repeated sequential calls).
func TestCreateOrGetAccount_Idempotency(t *testing.T) {
	tests := []struct {
		name         string
		x402Network  string
		cdpNetwork   string
		listEndpoint string
		numCalls     int
	}{
		{
			name:         "repeated calls for EVM account",
			x402Network:  "base-sepolia",
			cdpNetwork:   "base-sepolia",
			listEndpoint: "/platform/v2/evm/accounts",
			numCalls:     3,
		},
		{
			name:         "repeated calls for SVM account",
			x402Network:  "solana-devnet",
			cdpNetwork:   "solana-devnet",
			listEndpoint: "/platform/v2/solana/accounts",
			numCalls:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				listCallCount   int
				createCallCount int
				accountID       = "accounts/idempotent-test-123"
				accountAddress  = "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
			)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == "GET" && r.URL.Path == tt.listEndpoint:
					listCallCount++

					// After first create, return the existing account
					if createCallCount > 0 {
						w.WriteHeader(http.StatusOK)
						json.NewEncoder(w).Encode(ListAccountsResponse{
							Accounts: []AccountResponse{
								{
									ID:      accountID,
									Address: accountAddress,
									Network: tt.cdpNetwork,
								},
							},
						})
					} else {
						// First call - no accounts exist yet
						w.WriteHeader(http.StatusOK)
						json.NewEncoder(w).Encode(ListAccountsResponse{
							Accounts: []AccountResponse{},
						})
					}

				case r.Method == "POST":
					createCallCount++

					// Should only be called once
					if createCallCount > 1 {
						t.Error("Create account called multiple times - idempotency violated")
					}

					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(AccountResponse{
						ID:      accountID,
						Address: accountAddress,
						Network: tt.cdpNetwork,
					})

				default:
					t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			auth := &mockCDPAuth{}
			client := NewCDPClient(auth)
			client.baseURL = server.URL

			var firstAccount *CDPAccount
			for i := 0; i < tt.numCalls; i++ {
				account, err := CreateOrGetAccount(context.Background(), client, tt.x402Network)
				if err != nil {
					t.Fatalf("Call %d failed: %v", i+1, err)
				}

				if i == 0 {
					firstAccount = account
				} else {
					// All subsequent calls should return the same account
					if account.ID != firstAccount.ID {
						t.Errorf("Call %d: Account ID changed from %s to %s", i+1, firstAccount.ID, account.ID)
					}
					if account.Address != firstAccount.Address {
						t.Errorf("Call %d: Address changed from %s to %s", i+1, firstAccount.Address, account.Address)
					}
					if account.Network != firstAccount.Network {
						t.Errorf("Call %d: Network changed from %s to %s", i+1, firstAccount.Network, account.Network)
					}
				}
			}

			// Verify create was only called once
			if createCallCount != 1 {
				t.Errorf("Expected create to be called exactly once, got %d calls", createCallCount)
			}

			// Verify list was called for each request
			if listCallCount != tt.numCalls {
				t.Errorf("Expected list to be called %d times, got %d calls", tt.numCalls, listCallCount)
			}
		})
	}
}

// TestCreateOrGetAccount_ValidationErrors tests handling of invalid API responses.
func TestCreateOrGetAccount_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		x402Network string
		response    AccountResponse
		wantError   string
	}{
		{
			name:        "empty account ID",
			x402Network: "base-sepolia",
			response: AccountResponse{
				ID:      "",
				Address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
				Network: "base-sepolia",
			},
			wantError: "empty account ID",
		},
		{
			name:        "empty address",
			x402Network: "base-sepolia",
			response: AccountResponse{
				ID:      "accounts/test-123",
				Address: "",
				Network: "base-sepolia",
			},
			wantError: "empty account address",
		},
		{
			name:        "empty network",
			x402Network: "base-sepolia",
			response: AccountResponse{
				ID:      "accounts/test-123",
				Address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
				Network: "",
			},
			wantError: "empty account network",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					// Return empty list to trigger account creation
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(ListAccountsResponse{
						Accounts: []AccountResponse{},
					})
				} else if r.Method == "POST" {
					// Return invalid response
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			auth := &mockCDPAuth{}
			client := NewCDPClient(auth)
			client.baseURL = server.URL

			account, err := CreateOrGetAccount(context.Background(), client, tt.x402Network)
			if err == nil {
				t.Fatal("Expected validation error, got nil")
			}
			if account != nil {
				t.Errorf("Expected nil account on error, got %+v", account)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("Expected error containing %q, got: %v", tt.wantError, err)
			}
		})
	}
}

// TestCreateOrGetAccount_NetworkAliases tests that network aliases are properly mapped.
func TestCreateOrGetAccount_NetworkAliases(t *testing.T) {
	tests := []struct {
		name         string
		x402Network  string
		expectedCDP  string
		listEndpoint string
	}{
		{
			name:         "solana alias to mainnet",
			x402Network:  "solana",
			expectedCDP:  "solana-mainnet",
			listEndpoint: "/platform/v2/solana/accounts",
		},
		{
			name:         "mainnet-beta alias to solana-mainnet",
			x402Network:  "mainnet-beta",
			expectedCDP:  "solana-mainnet",
			listEndpoint: "/platform/v2/solana/accounts",
		},
		{
			name:         "devnet alias to solana-devnet",
			x402Network:  "devnet",
			expectedCDP:  "solana-devnet",
			listEndpoint: "/platform/v2/solana/accounts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedNetworkID string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(ListAccountsResponse{
						Accounts: []AccountResponse{},
					})
				} else if r.Method == "POST" {
					var req CreateAccountRequest
					json.NewDecoder(r.Body).Decode(&req)
					receivedNetworkID = req.NetworkID

					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(AccountResponse{
						ID:      "accounts/alias-test-123",
						Address: "DYw8jCTfwHNRJhhmFcbXvVDTqWMEVFBX6ZKUmG5CNSKK",
						Network: tt.expectedCDP,
					})
				}
			}))
			defer server.Close()

			auth := &mockCDPAuth{}
			client := NewCDPClient(auth)
			client.baseURL = server.URL

			account, err := CreateOrGetAccount(context.Background(), client, tt.x402Network)
			if err != nil {
				t.Fatalf("CreateOrGetAccount failed: %v", err)
			}

			if receivedNetworkID != tt.expectedCDP {
				t.Errorf("Expected network_id=%s, got %s", tt.expectedCDP, receivedNetworkID)
			}
			if account.Network != tt.expectedCDP {
				t.Errorf("Expected account network=%s, got %s", tt.expectedCDP, account.Network)
			}
		})
	}
}

// TestCreateOrGetAccount_RetryOnTransientError tests retry behavior for transient errors.
func TestCreateOrGetAccount_RetryOnTransientError(t *testing.T) {
	var attemptCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		if r.Method == "GET" {
			// Fail first 2 attempts with 500, succeed on 3rd
			if attemptCount < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ListAccountsResponse{
				Accounts: []AccountResponse{},
			})
		} else if r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(AccountResponse{
				ID:      "accounts/retry-test-123",
				Address: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
				Network: "base-sepolia",
			})
		}
	}))
	defer server.Close()

	auth := &mockCDPAuth{}
	client := NewCDPClient(auth)
	client.baseURL = server.URL

	account, err := CreateOrGetAccount(context.Background(), client, "base-sepolia")
	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}

	if account == nil {
		t.Fatal("Expected account to be returned after successful retry")
	}

	if attemptCount < 3 {
		t.Errorf("Expected at least 3 attempts (2 failures + 1 success), got %d", attemptCount)
	}
}
