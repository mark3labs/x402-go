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
	// It selects from multiple payment requirement options provided by the server.
	SelectAndSign(requirements []PaymentRequirement, signers []Signer) (*PaymentPayload, error)
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
func (s *DefaultPaymentSelector) SelectAndSign(requirements []PaymentRequirement, signers []Signer) (*PaymentPayload, error) {
	if len(signers) == 0 {
		return nil, NewPaymentError(ErrCodeNoValidSigner, "no signers configured", ErrNoValidSigner)
	}

	if len(requirements) == 0 {
		return nil, NewPaymentError(ErrCodeInvalidRequirements, "no payment requirements provided", ErrInvalidRequirements)
	}

	// Try each requirement option and find the best signer match
	type requirementCandidate struct {
		requirement    *PaymentRequirement
		signer         Signer
		signerPriority int
		tokenPriority  int
	}

	var allCandidates []requirementCandidate
	hasValidRequirement := false

	for i := range requirements {
		req := &requirements[i]

		// Parse required amount
		requiredAmount := new(big.Int)
		if _, ok := requiredAmount.SetString(req.MaxAmountRequired, 10); !ok {
			// If all requirements are invalid, we should return an error
			// But continue checking other requirements first
			continue
		}

		hasValidRequirement = true

		// Find all signers that can satisfy this requirement
		for _, signer := range signers {
			if !signer.CanSign(req) {
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
				if strings.EqualFold(token.Address, req.Asset) {
					tokenPriority = token.Priority
					break
				}
			}

			allCandidates = append(allCandidates, requirementCandidate{
				requirement:    req,
				signer:         signer,
				signerPriority: signer.GetPriority(),
				tokenPriority:  tokenPriority,
			})
		}
	}

	// If no valid requirements were found, return an error
	if !hasValidRequirement {
		return nil, NewPaymentError(ErrCodeInvalidRequirements, "invalid amount in requirements", ErrInvalidRequirements)
	}

	if len(allCandidates) == 0 {
		// Build error details from all requirements
		errorDetails := make([]string, 0, len(requirements))
		for _, req := range requirements {
			errorDetails = append(errorDetails, req.Network+":"+req.Asset)
		}
		return nil, NewPaymentError(ErrCodeNoValidSigner, "no signer can satisfy any payment requirement", ErrNoValidSigner).
			WithDetails("options", strings.Join(errorDetails, ", "))
	}

	// Sort by priority (signer first, then token)
	// Lower priority numbers come first (1 > 2 > 3)
	sort.Slice(allCandidates, func(i, j int) bool {
		if allCandidates[i].signerPriority != allCandidates[j].signerPriority {
			return allCandidates[i].signerPriority < allCandidates[j].signerPriority
		}
		return allCandidates[i].tokenPriority < allCandidates[j].tokenPriority
	})

	// Use the highest priority signer and requirement combination
	selectedCandidate := allCandidates[0]

	// Sign the payment
	payment, err := selectedCandidate.signer.Sign(selectedCandidate.requirement)
	if err != nil {
		return nil, NewPaymentError(ErrCodeSigningFailed, "failed to sign payment", err)
	}

	return payment, nil
}
