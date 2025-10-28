package x402

import (
	"errors"
	"testing"
)

func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"PaymentRequired", ErrPaymentRequired, "payment required"},
		{"InvalidPayment", ErrInvalidPayment, "invalid payment"},
		{"MalformedHeader", ErrMalformedHeader, "malformed payment header"},
		{"UnsupportedVersion", ErrUnsupportedVersion, "unsupported x402 version"},
		{"UnsupportedScheme", ErrUnsupportedScheme, "unsupported payment scheme"},
		{"UnsupportedNetwork", ErrUnsupportedNetwork, "unsupported network"},
		{"InvalidSignature", ErrInvalidSignature, "invalid signature"},
		{"InvalidAuthorization", ErrInvalidAuthorization, "invalid authorization"},
		{"ExpiredAuthorization", ErrExpiredAuthorization, "expired authorization"},
		{"InsufficientFunds", ErrInsufficientFunds, "insufficient funds"},
		{"InvalidNonce", ErrInvalidNonce, "invalid nonce"},
		{"RecipientMismatch", ErrRecipientMismatch, "recipient mismatch"},
		{"AmountMismatch", ErrAmountMismatch, "amount mismatch"},
		{"FacilitatorUnavailable", ErrFacilitatorUnavailable, "facilitator unavailable"},
		{"SettlementFailed", ErrSettlementFailed, "settlement failed"},
		{"VerificationFailed", ErrVerificationFailed, "verification failed"},
		{"Timeout", ErrTimeout, "operation timed out"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("Error message mismatch: got %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestErrorComparison(t *testing.T) {
	tests := []struct {
		name string
		err1 error
		err2 error
		want bool
	}{
		{
			name: "same error",
			err1: ErrPaymentRequired,
			err2: ErrPaymentRequired,
			want: true,
		},
		{
			name: "different errors",
			err1: ErrPaymentRequired,
			err2: ErrInvalidPayment,
			want: false,
		},
		{
			name: "wrapped error",
			err1: errors.New("wrapped: payment required"),
			err2: ErrPaymentRequired,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err1, tt.err2)
			if result != tt.want {
				t.Errorf("errors.Is() = %v, want %v", result, tt.want)
			}
		})
	}
}
