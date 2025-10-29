package x402

import (
	"math/big"
	"sort"
	"strings"
)

// PaymentSelector selects the appropriate signer and creates a payment.
type PaymentSelector interface {
	// SelectAndSign chooses the best signer from the available signers
	// and creates a signed payment for the given requirements.
	SelectAndSign(requirements *PaymentRequirement, signers []Signer) (*PaymentPayload, error)
}

// DefaultPaymentSelector implements the standard payment selection algorithm.
// It selects signers based on:
// 1. Ability to satisfy requirements (network and token match)
// 2. Signer priority (lower number = higher priority)
// 3. Token priority within the signer
// 4. Configuration order (for ties)
type DefaultPaymentSelector struct{}

// NewDefaultPaymentSelector creates a new DefaultPaymentSelector.
func NewDefaultPaymentSelector() *DefaultPaymentSelector {
	return &DefaultPaymentSelector{}
}

// SelectAndSign implements PaymentSelector.
func (s *DefaultPaymentSelector) SelectAndSign(requirements *PaymentRequirement, signers []Signer) (*PaymentPayload, error) {
	if len(signers) == 0 {
		return nil, NewPaymentError(ErrCodeNoValidSigner, "no signers configured", ErrNoValidSigner)
	}

	// Parse required amount
	requiredAmount := new(big.Int)
	if _, ok := requiredAmount.SetString(requirements.MaxAmountRequired, 10); !ok {
		return nil, NewPaymentError(ErrCodeInvalidRequirements, "invalid amount in requirements", ErrInvalidRequirements)
	}

	// Find all signers that can satisfy the requirements
	var candidates []signerCandidate
	for _, signer := range signers {
		if !signer.CanSign(requirements) {
			continue
		}

		// Check max amount limit
		maxAmount := signer.GetMaxAmount()
		if maxAmount != nil && requiredAmount.Cmp(maxAmount) > 0 {
			continue
		}

		// Find matching token and its priority
		tokenPriority := 0
		for _, token := range signer.GetTokens() {
			if strings.EqualFold(token.Address, requirements.Asset) {
				tokenPriority = token.Priority
				break
			}
		}

		candidates = append(candidates, signerCandidate{
			signer:         signer,
			signerPriority: signer.GetPriority(),
			tokenPriority:  tokenPriority,
		})
	}

	if len(candidates) == 0 {
		return nil, NewPaymentError(ErrCodeNoValidSigner, "no signer can satisfy requirements", ErrNoValidSigner).
			WithDetails("network", requirements.Network).
			WithDetails("asset", requirements.Asset).
			WithDetails("amount", requirements.MaxAmountRequired)
	}

	// Sort by priority (signer first, then token)
	// Lower priority numbers come first (1 > 2 > 3)
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].signerPriority != candidates[j].signerPriority {
			return candidates[i].signerPriority < candidates[j].signerPriority
		}
		return candidates[i].tokenPriority < candidates[j].tokenPriority
	})

	// Use the highest priority signer
	selectedSigner := candidates[0].signer

	// Sign the payment
	payment, err := selectedSigner.Sign(requirements)
	if err != nil {
		return nil, NewPaymentError(ErrCodeSigningFailed, "failed to sign payment", err)
	}

	return payment, nil
}

// signerCandidate represents a signer that can satisfy the payment requirements.
type signerCandidate struct {
	signer         Signer
	signerPriority int
	tokenPriority  int
}
