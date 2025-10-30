# Coinbase CDP Signer Implementation

## Summary

Successfully implemented Coinbase Developer Platform (CDP) signer for x402-go with full support for:
- EVM networks (Base, Ethereum, Polygon + testnets)
- Solana networks (mainnet-beta, devnet)
- Dual JWT authentication (Bearer + Wallet Auth)
- Account creation and management
- Transaction signing

## Key Implementation Details

### Authentication

CDP requires **two separate JWT tokens** for sensitive operations:

1. **Bearer Token** (`Authorization` header)
   - Algorithm: EdDSA (for Ed25519 keys) or ES256 (for ECDSA keys)
   - Claims: `sub`, `iss: "cdp"`, `aud: ["cdp_service"]`, `nbf`, `exp`, `uris` (array)
   - Headers: `alg`, `kid`, `nonce`, `typ: "JWT"`
   - Valid for 2 minutes

2. **Wallet Auth Token** (`X-Wallet-Auth` header)
   - Algorithm: ES256 (always ECDSA, even if API key is Ed25519)
   - Claims: `iat`, `nbf`, `jti`, `uris` (array), optional `reqHash`
   - Headers: `alg`, `typ: "JWT"` (NO `kid`)
   - Valid for 1 minute
   - `reqHash`: Only included when request body has data (empty `{}` = no hash)

### Critical Discovery: Empty Body Handling

The Python SDK's behavior with empty request bodies was key:
```python
if options.request_data:  # Empty dict {} is FALSY in Python!
    claims["reqHash"] = hashlib.sha256(json_bytes).hexdigest()
```

Our Go implementation matches this by checking if the parsed JSON is an empty object/array:
```go
switch val := v.(type) {
case map[string]interface{}:
    if len(val) == 0 {
        return nil, nil  // No hash for empty object
    }
}
```

### Account Creation API

**Endpoint**: `POST /platform/v2/{evm|solana}/accounts`

**Request Body** (optional):
```json
{
  "name": "optional-account-name",
  "accountPolicy": "uuid-policy-id"
}
```

For accounts without name/policy, send empty object: `{}`

**Response** (201 Created):
```json
{
  "address": "0x...",
  "name": "optional",
  "policies": ["uuid"],
  "createdAt": "2025-10-30T00:00:00Z",
  "updatedAt": "2025-10-30T00:00:00Z"
}
```

**Important**: No `id` or `network` field in creation response!

### Key Format Support

Both API Key Secret and Wallet Secret can be in multiple formats:
- Raw Ed25519 (64 bytes)
- Ed25519 seed (32 bytes)
- PKCS8 DER (supports both ECDSA and Ed25519)
- SEC1/EC format (ECDSA only)

All are base64-encoded in the `.env` file.

## Files Modified

### Core Implementation
- `signers/coinbase/auth.go` - Dual JWT generation with correct claims
- `signers/coinbase/client.go` - HTTP client with canonical JSON hashing
- `signers/coinbase/account.go` - Account creation with empty body handling
- `signers/coinbase/signer.go` - Main signer implementation
- `signers/coinbase/networks.go` - Network mapping
- `signers/coinbase/errors.go` - CDP error types

### Tests
- `signers/coinbase/*_test.go` - Comprehensive test coverage

### Example
- `examples/coinbase/main.go` - Complete client/server example
- `examples/coinbase/.env.example` - Template with correct key formats
- `examples/coinbase/README.md` - Usage documentation

## Testing

All tests pass:
```bash
go test -race ./signers/coinbase/
```

Live integration test:
```bash
cd examples/coinbase
go run main.go client --network base-sepolia --url <paywalled-url>
```

## Authentication Flow

1. Parse base64-encoded keys (API Key Secret + Wallet Secret)
2. Generate Bearer token with EdDSA/ES256
3. For POST/PUT/DELETE to `/accounts`: Generate Wallet Auth token with ES256
4. Include both tokens in request headers
5. Send request with `{}` body for account creation (or omit optional fields)

## Success Criteria

✅ Bearer token matches Python SDK format
✅ Wallet Auth token matches Python SDK format  
✅ Empty body `{}` generates NO `reqHash` claim
✅ Non-empty body generates correct `reqHash` with sorted JSON
✅ Account creation returns 201 with valid address
✅ Both EVM and Solana networks supported
✅ All unit tests passing
✅ Integration test with real CDP API succeeds

## References

- [CDP API Documentation](https://docs.cdp.coinbase.com/api-reference/v2/)
- [CDP Python SDK](https://github.com/coinbase/cdp-sdk/tree/main/python)
- [CDP Authentication](https://docs.cdp.coinbase.com/api-reference/v2/authentication)
