package crypto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/mr-tron/base58"
)

// AddressPrefix is prepended to derived addresses so they are self-describing.
const AddressPrefix = "ed25519:"

// Address derives a stable, human-manageable identifier from a public key.
// The address is a base58 encoding of the first 20 bytes of SHA-256(pub).
func Address(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	short := sum[:20]
	return AddressPrefix + base58.Encode(short)
}

// ParseAddress validates an address string and returns it unchanged.
func ParseAddress(s string) (string, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, AddressPrefix) {
		return "", fmt.Errorf("address %q must start with %q", s, AddressPrefix)
	}
	raw, err := base58.Decode(strings.TrimPrefix(s, AddressPrefix))
	if err != nil {
		return "", fmt.Errorf("invalid address %q: %w", s, err)
	}
	if len(raw) != 20 {
		return "", fmt.Errorf("invalid address %q: wrong length", s)
	}
	return s, nil
}

// EncodePrivate renders a private key seed as lowercase hex.
func EncodePrivate(priv ed25519.PrivateKey) string {
	return hex.EncodeToString(priv)
}

// DecodePrivate parses a hex-encoded private key seed.
func DecodePrivate(s string) (ed25519.PrivateKey, error) {
	b, err := hex.DecodeString(strings.TrimSpace(s))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	switch len(b) {
	case ed25519.SeedSize:
		return ed25519.NewKeyFromSeed(b), nil
	case ed25519.PrivateKeySize:
		return ed25519.PrivateKey(b), nil
	default:
		return nil, fmt.Errorf("invalid private key length: %d", len(b))
	}
}

// EncodePublic renders a public key as lowercase hex.
func EncodePublic(pub ed25519.PublicKey) string {
	return hex.EncodeToString(pub)
}

// DecodePublic parses a hex-encoded public key.
func DecodePublic(s string) (ed25519.PublicKey, error) {
	b, err := hex.DecodeString(strings.TrimSpace(s))
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}
	if len(b) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key length: %d", len(b))
	}
	return ed25519.PublicKey(b), nil
}
