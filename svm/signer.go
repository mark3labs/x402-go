package svm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/mark3labs/x402-go"
)

// Signer implements the x402.Signer interface for Solana (SVM).
type Signer struct {
	privateKey solana.PrivateKey
	publicKey  solana.PublicKey
	network    string
	tokens     []x402.TokenConfig
	priority   int
	maxAmount  *big.Int
}

// SignerOption configures a Signer.
type SignerOption func(*Signer) error

// NewSigner creates a new Solana signer with the given options.
func NewSigner(opts ...SignerOption) (*Signer, error) {
	s := &Signer{
		priority: 0,
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	// Validation
	if len(s.privateKey) == 0 {
		return nil, x402.ErrInvalidKey
	}
	if s.network == "" {
		return nil, x402.ErrInvalidNetwork
	}
	if len(s.tokens) == 0 {
		return nil, x402.ErrNoTokens
	}

	// Derive public key
	s.publicKey = s.privateKey.PublicKey()

	return s, nil
}

// WithPrivateKey sets the private key from a base58 string.
func WithPrivateKey(base58Key string) SignerOption {
	return func(s *Signer) error {
		privateKey, err := solana.PrivateKeyFromBase58(base58Key)
		if err != nil {
			return x402.ErrInvalidKey
		}
		s.privateKey = privateKey
		return nil
	}
}

// WithKeygenFile loads a private key from a Solana keygen JSON file.
func WithKeygenFile(path string) SignerOption {
	return func(s *Signer) error {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("%w: %v", x402.ErrInvalidKeystore, err)
		}

		// Parse JSON array format: [1, 2, 3, ...]
		var keyBytes []byte
		if err := json.Unmarshal(data, &keyBytes); err != nil {
			return fmt.Errorf("%w: invalid JSON format", x402.ErrInvalidKeystore)
		}

		if len(keyBytes) != 64 {
			return fmt.Errorf("%w: invalid key length", x402.ErrInvalidKeystore)
		}

		s.privateKey = solana.PrivateKey(keyBytes)
		return nil
	}
}

// WithNetwork sets the blockchain network.
func WithNetwork(network string) SignerOption {
	return func(s *Signer) error {
		s.network = network
		return nil
	}
}

// WithToken adds a token configuration.
func WithToken(mintAddress, symbol string, decimals int) SignerOption {
	return func(s *Signer) error {
		s.tokens = append(s.tokens, x402.TokenConfig{
			Address:  mintAddress,
			Symbol:   symbol,
			Decimals: decimals,
			Priority: 0,
		})
		return nil
	}
}

// WithTokenPriority adds a token configuration with a priority.
func WithTokenPriority(mintAddress, symbol string, decimals, priority int) SignerOption {
	return func(s *Signer) error {
		s.tokens = append(s.tokens, x402.TokenConfig{
			Address:  mintAddress,
			Symbol:   symbol,
			Decimals: decimals,
			Priority: priority,
		})
		return nil
	}
}

// WithPriority sets the signer priority.
func WithPriority(priority int) SignerOption {
	return func(s *Signer) error {
		s.priority = priority
		return nil
	}
}

// WithMaxAmountPerCall sets the maximum amount per payment call.
func WithMaxAmountPerCall(amount string) SignerOption {
	return func(s *Signer) error {
		maxAmount, ok := new(big.Int).SetString(amount, 10)
		if !ok {
			return x402.ErrInvalidAmount
		}
		s.maxAmount = maxAmount
		return nil
	}
}

// Network implements x402.Signer.
func (s *Signer) Network() string {
	return s.network
}

// Scheme implements x402.Signer.
func (s *Signer) Scheme() string {
	return "exact"
}

// CanSign implements x402.Signer.
func (s *Signer) CanSign(requirements *x402.PaymentRequirement) bool {
	// Check network match
	if requirements.Network != s.network {
		return false
	}

	// Check scheme match
	if requirements.Scheme != "exact" {
		return false
	}

	// Check if we have the required token
	for _, token := range s.tokens {
		if strings.EqualFold(token.Address, requirements.Asset) {
			return true
		}
	}

	return false
}

// Sign implements x402.Signer.
func (s *Signer) Sign(requirements *x402.PaymentRequirement) (*x402.PaymentPayload, error) {
	// Verify we can sign
	if !s.CanSign(requirements) {
		return nil, x402.ErrNoValidSigner
	}

	// Parse amount
	amount := new(big.Int)
	if _, ok := amount.SetString(requirements.MaxAmountRequired, 10); !ok {
		return nil, x402.ErrInvalidAmount
	}

	// Check max amount limit
	if s.maxAmount != nil && amount.Cmp(s.maxAmount) > 0 {
		return nil, x402.ErrAmountExceeded
	}

	// Get mint address
	mintAddress, err := solana.PublicKeyFromBase58(requirements.Asset)
	if err != nil {
		return nil, fmt.Errorf("invalid mint address: %w", err)
	}

	// Get recipient address
	recipient, err := solana.PublicKeyFromBase58(requirements.PayTo)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient address: %w", err)
	}

	// Build the partially signed transaction
	txBase64, err := BuildPartiallySignedTransfer(
		s.privateKey,
		s.publicKey,
		mintAddress,
		recipient,
		amount.Uint64(),
	)
	if err != nil {
		return nil, x402.NewPaymentError(x402.ErrCodeSigningFailed, "failed to build transaction", err)
	}

	// Build payment payload
	payload := &x402.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     s.network,
		Payload: x402.SVMPayload{
			Transaction: txBase64,
		},
	}

	return payload, nil
}

// GetPriority implements x402.Signer.
func (s *Signer) GetPriority() int {
	return s.priority
}

// GetTokens implements x402.Signer.
func (s *Signer) GetTokens() []x402.TokenConfig {
	return s.tokens
}

// GetMaxAmount implements x402.Signer.
func (s *Signer) GetMaxAmount() *big.Int {
	return s.maxAmount
}

// Address returns the signer's public key as a base58 string.
func (s *Signer) Address() string {
	return s.publicKey.String()
}

// BuildPartiallySignedTransfer creates a partially signed SPL token transfer.
// The client signs with their private key, and the facilitator will add the fee payer signature.
func BuildPartiallySignedTransfer(
	clientPrivateKey solana.PrivateKey,
	clientPublicKey solana.PublicKey,
	mint solana.PublicKey,
	recipient solana.PublicKey,
	amount uint64,
) (string, error) {
	// Get associated token accounts
	sourceATA, _, err := solana.FindAssociatedTokenAddress(clientPublicKey, mint)
	if err != nil {
		return "", fmt.Errorf("failed to find source ATA: %w", err)
	}

	destATA, _, err := solana.FindAssociatedTokenAddress(recipient, mint)
	if err != nil {
		return "", fmt.Errorf("failed to find destination ATA: %w", err)
	}

	// Build SPL token transfer instruction
	// This uses the Token Program transfer instruction
	transferInstruction := solana.NewInstruction(
		solana.TokenProgramID, // program ID
		solana.AccountMetaSlice{
			solana.Meta(sourceATA).WRITE(),        // source account
			solana.Meta(destATA).WRITE(),          // destination account
			solana.Meta(clientPublicKey).SIGNER(), // authority (client)
		},
		buildTransferInstructionData(amount),
	)

	// Create transaction with placeholder blockhash
	// The facilitator will update this with a recent blockhash
	placeholderBlockhash := solana.Hash{}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{transferInstruction},
		placeholderBlockhash,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create transaction: %w", err)
	}

	// Sign with client private key (partial signature)
	// The facilitator will be the fee payer and will add their signature
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(clientPublicKey) {
			return &clientPrivateKey
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Serialize transaction to bytes
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("failed to marshal transaction: %w", err)
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(txBytes), nil
}

// buildTransferInstructionData builds the instruction data for an SPL token transfer.
// Format: [3, amount (u64 little-endian)]
// Instruction discriminator 3 = Transfer
func buildTransferInstructionData(amount uint64) []byte {
	data := make([]byte, 9)
	data[0] = 3 // Transfer instruction

	// Encode amount as little-endian u64
	data[1] = byte(amount)
	data[2] = byte(amount >> 8)
	data[3] = byte(amount >> 16)
	data[4] = byte(amount >> 24)
	data[5] = byte(amount >> 32)
	data[6] = byte(amount >> 40)
	data[7] = byte(amount >> 48)
	data[8] = byte(amount >> 56)

	return data
}
