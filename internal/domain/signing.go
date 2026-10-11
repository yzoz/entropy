package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

// SigningPayload exposes the deterministic bytes signed by contract parties.
func SigningPayload(c *Contract) []byte { return signingPayload(c) }

// signingPayload builds the deterministic byte string that is signed and
// verified. Only immutable, identity-defining fields are included: the mutable
// status and accumulated signature list are deliberately excluded so a
// countersignature does not invalidate the first signature.
func signingPayload(c *Contract) []byte {
	h := sha256.New()
	w := func(parts ...string) {
		for _, p := range parts {
			h.Write([]byte(p))
			h.Write([]byte{0})
		}
	}
	w(
		"entropy:v1",
		c.ID,
		string(c.Type),
		c.From,
		c.To,
		c.Currency.String(),
		FormatAmount(c.Currency, c.Amount),
		formatUnix(c.IssuedAt),
		formatUnix(c.ExpiresAt),
		c.Settles,
		c.Memo,
	)
	return h.Sum(nil)
}

func formatUnix(ts int64) string {
	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
}

// MarshalCanonical returns a deterministic JSON encoding of the contract.
// Because Contract contains no maps, encoding/json already emits fields in
// struct order, so the result is stable across runs and platforms.
func MarshalCanonical(c *Contract) ([]byte, error) {
	return json.Marshal(c)
}

// Hash returns a hex-encoded SHA-256 over the canonical JSON encoding.
func Hash(c *Contract) (string, error) {
	b, err := MarshalCanonical(c)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// DeriveID derives the default contract ID from its defining fields. The ID is
// independent of status and signatures, so it survives the whole lifecycle.
func DeriveID(c *Contract) (string, error) {
	if c.ID != "" {
		return c.ID, nil
	}
	h := sha256.New()
	h.Write(signingPayload(c))
	return hex.EncodeToString(h.Sum(nil))[:32], nil
}

// ErrNotSignable marks operations attempted on non-draft/signed contracts.
var ErrNotSignable = errors.New("contract is not in a signable state")
