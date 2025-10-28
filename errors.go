package x402

import "errors"

// Standard x402 error definitions

var (
	// ErrPaymentRequired indicates that payment is required to access the resource.
	ErrPaymentRequired = errors.New("payment required")

	// ErrInvalidPayment indicates that the provided payment is invalid.
	ErrInvalidPayment = errors.New("invalid payment")

	// ErrMalformedHeader indicates that the X-PAYMENT header is malformed.
	ErrMalformedHeader = errors.New("malformed payment header")

	// ErrUnsupportedVersion indicates an unsupported x402 protocol version.
	ErrUnsupportedVersion = errors.New("unsupported x402 version")

	// ErrUnsupportedScheme indicates an unsupported payment scheme.
	ErrUnsupportedScheme = errors.New("unsupported payment scheme")

	// ErrUnsupportedNetwork indicates an unsupported blockchain network.
	ErrUnsupportedNetwork = errors.New("unsupported network")

	// ErrInvalidSignature indicates an invalid cryptographic signature.
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrInvalidAuthorization indicates invalid payment authorization data.
	ErrInvalidAuthorization = errors.New("invalid authorization")

	// ErrExpiredAuthorization indicates the payment authorization has expired.
	ErrExpiredAuthorization = errors.New("expired authorization")

	// ErrInsufficientFunds indicates the payer has insufficient funds.
	ErrInsufficientFunds = errors.New("insufficient funds")

	// ErrInvalidNonce indicates an invalid or reused nonce.
	ErrInvalidNonce = errors.New("invalid nonce")

	// ErrRecipientMismatch indicates payment recipient doesn't match requirements.
	ErrRecipientMismatch = errors.New("recipient mismatch")

	// ErrAmountMismatch indicates payment amount doesn't meet requirements.
	ErrAmountMismatch = errors.New("amount mismatch")

	// ErrFacilitatorUnavailable indicates the facilitator service is unavailable.
	ErrFacilitatorUnavailable = errors.New("facilitator unavailable")

	// ErrSettlementFailed indicates on-chain settlement failed.
	ErrSettlementFailed = errors.New("settlement failed")

	// ErrVerificationFailed indicates payment verification failed.
	ErrVerificationFailed = errors.New("verification failed")

	// ErrTimeout indicates the operation timed out.
	ErrTimeout = errors.New("operation timed out")
)
