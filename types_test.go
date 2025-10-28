package x402

import (
	"encoding/json"
	"testing"
)

func TestPaymentRequirement_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     PaymentRequirement
		wantErr bool
	}{
		{
			name: "valid requirement",
			req: PaymentRequirement{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/data",
				Description:       "Premium data access",
				MaxTimeoutSeconds: 60,
			},
			wantErr: false,
		},
		{
			name: "missing scheme",
			req: PaymentRequirement{
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/data",
				Description:       "Premium data access",
				MaxTimeoutSeconds: 60,
			},
			wantErr: true,
		},
		{
			name: "missing network",
			req: PaymentRequirement{
				Scheme:            "exact",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/data",
				Description:       "Premium data access",
				MaxTimeoutSeconds: 60,
			},
			wantErr: true,
		},
		{
			name: "invalid amount - zero",
			req: PaymentRequirement{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "0",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/data",
				Description:       "Premium data access",
				MaxTimeoutSeconds: 60,
			},
			wantErr: true,
		},
		{
			name: "invalid amount - negative",
			req: PaymentRequirement{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "-100",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/data",
				Description:       "Premium data access",
				MaxTimeoutSeconds: 60,
			},
			wantErr: true,
		},
		{
			name: "invalid timeout - zero",
			req: PaymentRequirement{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/data",
				Description:       "Premium data access",
				MaxTimeoutSeconds: 0,
			},
			wantErr: true,
		},
		{
			name: "missing asset",
			req: PaymentRequirement{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/data",
				Description:       "Premium data access",
				MaxTimeoutSeconds: 60,
			},
			wantErr: true,
		},
		{
			name: "missing payTo",
			req: PaymentRequirement{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				Resource:          "https://api.example.com/data",
				Description:       "Premium data access",
				MaxTimeoutSeconds: 60,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PaymentRequirement.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEVMPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload EVMPayload
		wantErr bool
	}{
		{
			name: "valid EVM payload",
			payload: EVMPayload{
				Signature: "0x2d6a7588d6acca505cbf0d9a4a227e0c52c6c34008c8e8986a1283259764173608a2ce6496642e377d6da8dbbf5836e9bd15092f9ecab05ded3d6293af148b571c",
				Authorization: Authorization{
					From:        "0x857b06519E91e3A54538791bDbb0E22373e36b66",
					To:          "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
					Value:       "10000",
					ValidAfter:  "1740672089",
					ValidBefore: "1740672154",
					Nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid signature format",
			payload: EVMPayload{
				Signature: "invalid-signature",
				Authorization: Authorization{
					From:        "0x857b06519E91e3A54538791bDbb0E22373e36b66",
					To:          "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
					Value:       "10000",
					ValidAfter:  "1740672089",
					ValidBefore: "1740672154",
					Nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid from address",
			payload: EVMPayload{
				Signature: "0x2d6a7588d6acca505cbf0d9a4a227e0c52c6c34008c8e8986a1283259764173608a2ce6496642e377d6da8dbbf5836e9bd15092f9ecab05ded3d6293af148b571c",
				Authorization: Authorization{
					From:        "invalid-address",
					To:          "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
					Value:       "10000",
					ValidAfter:  "1740672089",
					ValidBefore: "1740672154",
					Nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid to address",
			payload: EVMPayload{
				Signature: "0x2d6a7588d6acca505cbf0d9a4a227e0c52c6c34008c8e8986a1283259764173608a2ce6496642e377d6da8dbbf5836e9bd15092f9ecab05ded3d6293af148b571c",
				Authorization: Authorization{
					From:        "0x857b06519E91e3A54538791bDbb0E22373e36b66",
					To:          "invalid-address",
					Value:       "10000",
					ValidAfter:  "1740672089",
					ValidBefore: "1740672154",
					Nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid nonce - too short",
			payload: EVMPayload{
				Signature: "0x2d6a7588d6acca505cbf0d9a4a227e0c52c6c34008c8e8986a1283259764173608a2ce6496642e377d6da8dbbf5836e9bd15092f9ecab05ded3d6293af148b571c",
				Authorization: Authorization{
					From:        "0x857b06519E91e3A54538791bDbb0E22373e36b66",
					To:          "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
					Value:       "10000",
					ValidAfter:  "1740672089",
					ValidBefore: "1740672154",
					Nonce:       "0x1234",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid timestamps - validBefore before validAfter",
			payload: EVMPayload{
				Signature: "0x2d6a7588d6acca505cbf0d9a4a227e0c52c6c34008c8e8986a1283259764173608a2ce6496642e377d6da8dbbf5836e9bd15092f9ecab05ded3d6293af148b571c",
				Authorization: Authorization{
					From:        "0x857b06519E91e3A54538791bDbb0E22373e36b66",
					To:          "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
					Value:       "10000",
					ValidAfter:  "1740672154",
					ValidBefore: "1740672089",
					Nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid value - zero",
			payload: EVMPayload{
				Signature: "0x2d6a7588d6acca505cbf0d9a4a227e0c52c6c34008c8e8986a1283259764173608a2ce6496642e377d6da8dbbf5836e9bd15092f9ecab05ded3d6293af148b571c",
				Authorization: Authorization{
					From:        "0x857b06519E91e3A54538791bDbb0E22373e36b66",
					To:          "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
					Value:       "0",
					ValidAfter:  "1740672089",
					ValidBefore: "1740672154",
					Nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("EVMPayload.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSVMPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload SVMPayload
		wantErr bool
	}{
		{
			name: "valid SVM payload",
			payload: SVMPayload{
				Transaction: "base64encodedtransaction==",
			},
			wantErr: false,
		},
		{
			name: "empty transaction",
			payload: SVMPayload{
				Transaction: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SVMPayload.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEVMAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "valid address",
			address: "0x857b06519E91e3A54538791bDbb0E22373e36b66",
			wantErr: false,
		},
		{
			name:    "valid address lowercase",
			address: "0x857b06519e91e3a54538791bdbb0e22373e36b66",
			wantErr: false,
		},
		{
			name:    "invalid - missing 0x prefix",
			address: "857b06519E91e3A54538791bDbb0E22373e36b66",
			wantErr: true,
		},
		{
			name:    "invalid - too short",
			address: "0x857b06519E91e3A5453879",
			wantErr: true,
		},
		{
			name:    "invalid - too long",
			address: "0x857b06519E91e3A54538791bDbb0E22373e36b66FF",
			wantErr: true,
		},
		{
			name:    "invalid - non-hex characters",
			address: "0x857b06519E91e3A54538791bDbb0E22373e36bZZ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEVMAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEVMAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaymentRequirementsResponse_JSON(t *testing.T) {
	resp := PaymentRequirementsResponse{
		X402Version: 1,
		Error:       "Payment required for this resource",
		Accepts: []PaymentRequirement{
			{
				Scheme:            "exact",
				Network:           "base-sepolia",
				MaxAmountRequired: "10000",
				Asset:             "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
				PayTo:             "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
				Resource:          "https://api.example.com/data",
				Description:       "Premium data access",
				MaxTimeoutSeconds: 60,
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var decoded PaymentRequirementsResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify
	if decoded.X402Version != resp.X402Version {
		t.Errorf("X402Version mismatch: got %d, want %d", decoded.X402Version, resp.X402Version)
	}
	if decoded.Error != resp.Error {
		t.Errorf("Error mismatch: got %s, want %s", decoded.Error, resp.Error)
	}
	if len(decoded.Accepts) != len(resp.Accepts) {
		t.Errorf("Accepts length mismatch: got %d, want %d", len(decoded.Accepts), len(resp.Accepts))
	}
}

func TestPaymentPayload_JSON(t *testing.T) {
	evmPayload := EVMPayload{
		Signature: "0x2d6a7588d6acca505cbf0d9a4a227e0c52c6c34008c8e8986a1283259764173608a2ce6496642e377d6da8dbbf5836e9bd15092f9ecab05ded3d6293af148b571c",
		Authorization: Authorization{
			From:        "0x857b06519E91e3A54538791bDbb0E22373e36b66",
			To:          "0x209693Bc6afc0C5328bA36FaF03C514EF312287C",
			Value:       "10000",
			ValidAfter:  "1740672089",
			ValidBefore: "1740672154",
			Nonce:       "0xf3746613c2d920b5fdabc0856f2aeb2d4f88ee6037b8cc5d04a71a4462f13480",
		},
	}

	payloadData, err := json.Marshal(evmPayload)
	if err != nil {
		t.Fatalf("Failed to marshal EVM payload: %v", err)
	}

	payment := PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "base-sepolia",
		Payload:     payloadData,
	}

	// Marshal to JSON
	data, err := json.Marshal(payment)
	if err != nil {
		t.Fatalf("Failed to marshal payment: %v", err)
	}

	// Unmarshal back
	var decoded PaymentPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal payment: %v", err)
	}

	// Verify
	if decoded.X402Version != payment.X402Version {
		t.Errorf("X402Version mismatch")
	}
	if decoded.Scheme != payment.Scheme {
		t.Errorf("Scheme mismatch")
	}
	if decoded.Network != payment.Network {
		t.Errorf("Network mismatch")
	}

	// Unmarshal the payload
	var decodedEVM EVMPayload
	if err := json.Unmarshal(decoded.Payload, &decodedEVM); err != nil {
		t.Fatalf("Failed to unmarshal EVM payload: %v", err)
	}

	if decodedEVM.Signature != evmPayload.Signature {
		t.Errorf("Signature mismatch")
	}
}
