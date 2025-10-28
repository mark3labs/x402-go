# Data Model: x402 Payment Middleware

**Date**: 2025-10-28  
**Feature**: x402 Payment Middleware

## Core Entities

### PaymentRequirement

**Purpose**: Defines a single acceptable payment method for a protected resource

**Fields**:
- `Scheme` (string, required): Payment scheme identifier ("exact")
- `Network` (string, required): Blockchain network ("base-sepolia", "solana", etc.)
- `MaxAmountRequired` (string, required): Required payment amount in atomic units
- `Asset` (string, required): Token contract/mint address
- `PayTo` (string, required): Recipient wallet address
- `Resource` (string, required): URL of the protected resource
- `Description` (string, required): Human-readable resource description
- `MimeType` (string, optional): Expected response MIME type
- `OutputSchema` (object, optional): JSON schema for response format
- `MaxTimeoutSeconds` (int, required): Maximum time for payment completion
- `Extra` (map, optional): Scheme-specific additional data

**Validations**:
- Scheme must be supported ("exact")
- Network must be valid chain identifier
- Addresses must be valid for the network type
- Amount must be positive numeric string
- Timeout must be positive integer

**Relationships**:
- Many PaymentRequirements can exist in one PaymentRequirementsResponse
- Each requirement is independent (client chooses one)

---

### PaymentRequirementsResponse

**Purpose**: Complete response body for 402 Payment Required status

**Fields**:
- `X402Version` (int, required): Protocol version (currently 1)
- `Error` (string, required): Human-readable error message
- `Accepts` ([]PaymentRequirement, required): Array of payment options

**Validations**:
- Version must be 1
- At least one payment requirement must be present
- Error message must be non-empty

**State**: Stateless (generated per request)

---

### PaymentPayload

**Purpose**: Payment authorization data sent by client

**Fields**:
- `X402Version` (int, required): Protocol version (must be 1)
- `Scheme` (string, required): Payment scheme used
- `Network` (string, required): Blockchain network used
- `Payload` (SchemePayload, required): Scheme-specific payment data

**Validations**:
- Version must match supported version
- Scheme/network must match one of the offered requirements
- Payload structure must match scheme type

**Relationships**:
- Contains one SchemePayload (polymorphic based on scheme)

---

### SchemePayload

**Purpose**: Abstract type for scheme-specific payment data

#### EVMPayload (for scheme="exact", network=EVM chains)

**Fields**:
- `Signature` (string, required): EIP-712 signature hex string
- `Authorization` (object, required):
  - `From` (string, required): Payer address
  - `To` (string, required): Recipient address  
  - `Value` (string, required): Amount in atomic units
  - `ValidAfter` (string, required): Unix timestamp string
  - `ValidBefore` (string, required): Unix timestamp string
  - `Nonce` (string, required): 32-byte hex nonce

**Validations**:
- Signature must be valid hex with 0x prefix
- Addresses must be valid EVM addresses (0x + 40 hex chars)
- ValidBefore must be after ValidAfter
- Nonce must be 32 bytes (66 chars with 0x)

#### SVMPayload (for scheme="exact", network=SVM chains)

**Fields**:
- `Transaction` (string, required): Base64-encoded serialized transaction

**Validations**:
- Must be valid base64
- Must deserialize to valid Solana transaction format

---

### SettlementResponse

**Purpose**: Payment settlement result information

**Fields**:
- `Success` (bool, required): Settlement success indicator
- `ErrorReason` (string, optional): Error description if failed
- `Transaction` (string, required): Blockchain transaction hash (empty if failed)
- `Network` (string, required): Network where settlement occurred
- `Payer` (string, required): Payer's wallet address

**Validations**:
- If success=false, errorReason must be present
- If success=true, transaction must be non-empty
- Network must match request network

**State**: Immutable once created

---

### MiddlewareConfig

**Purpose**: Runtime configuration for middleware instance

**Fields**:
- `FacilitatorURL` (string, required): Primary facilitator endpoint
- `FallbackFacilitatorURL` (string, optional): Backup facilitator
- `PaymentRequirements` ([]PaymentRequirement, required): Accepted payments
- `VerifyOnly` (bool, optional): Skip settlement if true
- `Resource` (string, required): Resource URL being protected
- `Description` (string, required): Resource description

**Validations**:
- URLs must be valid HTTP(S) endpoints
- At least one payment requirement must be configured
- Resource URL must be valid

**Lifecycle**: Created at middleware initialization, immutable during runtime

---

### FacilitatorRequest

**Purpose**: Request payload for facilitator API calls

**Fields**:
- `PaymentPayload` (PaymentPayload, required): Client's payment data
- `PaymentRequirements` (PaymentRequirement, required): Expected requirements

**Usage**: Sent to /verify and /settle endpoints

---

### FacilitatorVerifyResponse

**Purpose**: Response from facilitator /verify endpoint

**Fields**:
- `IsValid` (bool, required): Validation result
- `InvalidReason` (string, optional): Reason if invalid
- `Payer` (string, required): Payer address

---

### FacilitatorSupportedResponse  

**Purpose**: Response from facilitator /supported endpoint

**Fields**:
- `Kinds` ([]SupportedKind, required): Supported payment types

**SupportedKind fields**:
- `X402Version` (int, required): Protocol version
- `Scheme` (string, required): Payment scheme
- `Network` (string, required): Blockchain network

---

## Entity State Transitions

### Payment Flow States

1. **No Payment** → Client request without X-PAYMENT header
2. **Payment Required** → Server returns 402 with requirements
3. **Payment Submitted** → Client sends X-PAYMENT header
4. **Payment Verified** → Facilitator validates authorization
5. **Payment Settled** → Facilitator executes blockchain transaction
6. **Request Completed** → Resource returned with X-PAYMENT-RESPONSE

### Error States

- **Invalid Payment** → Return 400 Bad Request
- **Facilitator Unavailable** → Return 503 or try fallback
- **Settlement Failed** → Return 402 with error details
- **Verification Failed** → Return 402 with requirements

## Data Serialization

### JSON Encoding
- All entities use standard JSON encoding
- Amounts represented as strings to avoid precision issues
- Timestamps as Unix epoch strings
- Addresses as hex strings with 0x prefix (EVM)

### Header Encoding
- X-PAYMENT: Base64(JSON(PaymentPayload))
- X-PAYMENT-RESPONSE: Base64(JSON(SettlementResponse))

## Validation Rules Summary

1. **Type Safety**: All fields strongly typed in Go structs
2. **Required Fields**: Enforced via JSON tags and validation functions
3. **Format Validation**: Chain-specific address and signature formats
4. **Business Rules**: Amount matching, timeout enforcement
5. **Security**: Nonce uniqueness (delegated to facilitator)

## Performance Considerations

- All entities are request-scoped (no persistent storage)
- Minimal allocations per request
- Efficient JSON encoding/decoding with struct tags
- Base64 operations on small payloads (<1KB typically)