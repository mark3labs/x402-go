package validation

import (
	"strings"
	"testing"

	v2 "github.com/mark3labs/x402-go/v2"
)

// nanoGenesis is the Nano genesis account; its public key is
// E89208DD038FBB269987689621D52292AE9C35941A7484756ECCED92A65093BA.
const nanoGenesis = "nano_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3"

func TestValidateNanoAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
		errMsg  string
	}{
		{name: "valid nano_ address", address: nanoGenesis},
		{name: "valid legacy xrb_ address", address: "xrb_" + strings.TrimPrefix(nanoGenesis, "nano_")},
		{
			name:    "wrong checksum",
			address: nanoGenesis[:len(nanoGenesis)-1] + "1",
			wantErr: true,
			errMsg:  "checksum",
		},
		{
			name:    "one key character changed",
			address: strings.Replace(nanoGenesis, "t6k35", "t6k36", 1),
			wantErr: true,
			errMsg:  "checksum",
		},
		{
			name:    "missing prefix",
			address: strings.TrimPrefix(nanoGenesis, "nano_"),
			wantErr: true,
			errMsg:  "expected nano_ prefix",
		},
		{
			name:    "too short",
			address: nanoGenesis[:40],
			wantErr: true,
			errMsg:  "expected 60 characters",
		},
		{
			name:    "character outside the Nano alphabet",
			address: "nano_" + "0" + nanoGenesis[6:],
			wantErr: true,
			errMsg:  "invalid Nano address format",
		},
		{
			name:    "key above 256 bits",
			address: "nano_" + "4" + nanoGenesis[6:],
			wantErr: true,
			errMsg:  "invalid Nano address format",
		},
		{
			name:    "EVM address on a Nano network",
			address: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
			wantErr: true,
			errMsg:  "expected nano_ prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddress(tt.address, v2.NetworkNanoMainnet)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateAddress() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateNanoRequirements(t *testing.T) {
	base := v2.PaymentRequirements{
		Scheme:            "exact",
		Network:           v2.NetworkNanoMainnet,
		Amount:            "1000000000000000000000000000", // 0.001 XNO in raw (10^30 raw = 1 XNO)
		Asset:             v2.NanoAsset,
		PayTo:             nanoGenesis,
		MaxTimeoutSeconds: 300,
	}

	if err := ValidatePaymentRequirements(base); err != nil {
		t.Fatalf("ValidatePaymentRequirements() error = %v for valid Nano requirements", err)
	}

	wrongAsset := base
	wrongAsset.Asset = "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
	if err := ValidatePaymentRequirements(wrongAsset); err == nil || !strings.Contains(err.Error(), "asset must be") {
		t.Errorf("expected asset error for non-XNO asset, got %v", err)
	}

	badPayTo := base
	badPayTo.PayTo = nanoGenesis[:len(nanoGenesis)-1] + "1"
	if err := ValidatePaymentRequirements(badPayTo); err == nil || !strings.Contains(err.Error(), "payTo") {
		t.Errorf("expected payTo checksum error, got %v", err)
	}

	maxAmount := base
	maxAmount.Amount = "340282366920938463463374607431768211455" // 2^128-1 raw
	if err := ValidatePaymentRequirements(maxAmount); err != nil {
		t.Errorf("ValidatePaymentRequirements() error = %v for amount 2^128-1", err)
	}

	tooLarge := base
	tooLarge.Amount = "340282366920938463463374607431768211456" // 2^128 raw
	if err := ValidatePaymentRequirements(tooLarge); err == nil || !strings.Contains(err.Error(), "Nano maximum") {
		t.Errorf("expected Nano maximum amount error for 2^128, got %v", err)
	}
}
