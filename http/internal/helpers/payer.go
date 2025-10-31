package helpers

import "github.com/mark3labs/x402-go"

func GetPayer(payment x402.PaymentPayload) string {
	switch payment.Network {
	case x402.SolanaDevnet.NetworkID, x402.SolanaMainnet.NetworkID:
		return getPayerWithSolana(payment)
	default:
		return ""
	}
}
