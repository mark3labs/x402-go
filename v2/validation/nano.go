package validation

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/blake2b"
)

// nanoAlphabet is the base32 alphabet used by Nano account addresses.
const nanoAlphabet = "13456789abcdefghijkmnopqrstuwxyz"

// validateNanoAddress checks a Nano account address: the "nano_" (or legacy
// "xrb_") prefix, 60 characters from the Nano alphabet, and the 40-bit
// Blake2b checksum over the 32-byte public key. A mistyped payTo address
// fails here instead of sending funds to an account nobody controls.
func validateNanoAddress(address string) error {
	var body string
	switch {
	case strings.HasPrefix(address, "nano_"):
		body = strings.TrimPrefix(address, "nano_")
	case strings.HasPrefix(address, "xrb_"):
		body = strings.TrimPrefix(address, "xrb_")
	default:
		return fmt.Errorf("invalid Nano address format: %s (expected nano_ prefix)", address)
	}
	if len(body) != 60 {
		return fmt.Errorf("invalid Nano address format: %s (expected 60 characters after the prefix)", address)
	}

	// 52 characters encode 4 zero padding bits and the 256-bit public key;
	// the last 8 encode the 40-bit checksum.
	key, ok := decodeNanoBase32(body[:52])
	if !ok || key.BitLen() > 256 {
		return fmt.Errorf("invalid Nano address format: %s", address)
	}
	sum, ok := decodeNanoBase32(body[52:])
	if !ok {
		return fmt.Errorf("invalid Nano address format: %s", address)
	}

	pub := key.FillBytes(make([]byte, 32))
	h, err := blake2b.New(5, nil)
	if err != nil {
		return fmt.Errorf("nano checksum: %w", err)
	}
	h.Write(pub)
	digest := h.Sum(nil)
	// The checksum is encoded with its bytes in reverse order.
	for i, j := 0, len(digest)-1; i < j; i, j = i+1, j-1 {
		digest[i], digest[j] = digest[j], digest[i]
	}
	if !bytes.Equal(digest, sum.FillBytes(make([]byte, 5))) {
		return fmt.Errorf("invalid Nano address checksum: %s", address)
	}
	return nil
}

// decodeNanoBase32 decodes characters of the Nano alphabet, 5 bits each,
// most significant first.
func decodeNanoBase32(s string) (*big.Int, bool) {
	n := new(big.Int)
	for i := 0; i < len(s); i++ {
		v := strings.IndexByte(nanoAlphabet, s[i])
		if v < 0 {
			return nil, false
		}
		n.Lsh(n, 5)
		n.Or(n, big.NewInt(int64(v)))
	}
	return n, true
}
