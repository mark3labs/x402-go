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
		{"NoValidSigner", ErrNoValidSigner, "x402: no signer can satisfy payment requirements"},
		{"AmountExceeded", ErrAmountExceeded, "x402: payment amount exceeds per-call limit"},
		{"InvalidRequirements", ErrInvalidRequirements, "x402: invalid payment requirements"},
		{"SigningFailed", ErrSigningFailed, "x402: payment signing failed"},
		{"NetworkError", ErrNetworkError, "x402: network error during payment"},
		{"InvalidAmount", ErrInvalidAmount, "x402: invalid amount"},
		{"InvalidKey", ErrInvalidKey, "x402: invalid private key"},
		{"InvalidNetwork", ErrInvalidNetwork, "x402: invalid or unsupported network"},
		{"InvalidToken", ErrInvalidToken, "x402: invalid token configuration"},
		{"InvalidKeystore", ErrInvalidKeystore, "x402: invalid keystore file"},
		{"InvalidMnemonic", ErrInvalidMnemonic, "x402: invalid mnemonic phrase"},
		{"NoTokens", ErrNoTokens, "x402: no tokens configured"},
		{"FacilitatorUnavailable", ErrFacilitatorUnavailable, "x402: facilitator service unavailable"},
		{"VerificationFailed", ErrVerificationFailed, "x402: payment verification failed"},
		{"MalformedHeader", ErrMalformedHeader, "x402: malformed payment header"},
		{"UnsupportedVersion", ErrUnsupportedVersion, "x402: unsupported protocol version"},
		{"UnsupportedScheme", ErrUnsupportedScheme, "x402: unsupported payment scheme"},
		{"SettlementFailed", ErrSettlementFailed, "x402: payment settlement failed"},
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
			err1: ErrNoValidSigner,
			err2: ErrNoValidSigner,
			want: true,
		},
		{
			name: "different errors",
			err1: ErrNoValidSigner,
			err2: ErrInvalidAmount,
			want: false,
		},
		{
			name: "wrapped error",
			err1: errors.New("wrapped: no valid signer"),
			err2: ErrNoValidSigner,
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
