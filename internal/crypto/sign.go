package crypto

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// ErrBadSignature is returned when a signature does not verify.
var ErrBadSignature = errors.New("invalid signature")

// Signer produces detached Ed25519 signatures over arbitrary messages.
type Signer interface {
	Sign(msg []byte) ([]byte, error)
	Public() ed25519.PublicKey
}

// Verifier checks detached Ed25519 signatures.
type Verifier interface {
	Verify(pub ed25519.PublicKey, msg, sig []byte) bool
}

// Ed25519Signer is the default Signer backed by a private key.
type Ed25519Signer struct {
	priv ed25519.PrivateKey
}

// NewSigner wraps a private key.
func NewSigner(priv ed25519.PrivateKey) *Ed25519Signer {
	return &Ed25519Signer{priv: priv}
}

// Sign returns the detached signature of msg.
func (s *Ed25519Signer) Sign(msg []byte) ([]byte, error) {
	return ed25519.Sign(s.priv, msg), nil
}

// Public returns the public key of the signer.
func (s *Ed25519Signer) Public() ed25519.PublicKey {
	return s.priv.Public().(ed25519.PublicKey)
}

// Ed25519Verifier is the default Verifier.
type Ed25519Verifier struct{}

// Verify reports whether sig is a valid signature of msg under pub.
func (Ed25519Verifier) Verify(pub ed25519.PublicKey, msg, sig []byte) bool {
	return ed25519.Verify(pub, msg, sig)
}

// EncodeSignature renders a signature as lowercase hex.
func EncodeSignature(sig []byte) string { return hex.EncodeToString(sig) }

// DecodeSignature parses a hex-encoded signature and checks its length.
func DecodeSignature(s string) ([]byte, error) {
	b, err := hex.DecodeString(strings.TrimSpace(s))
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding: %w", err)
	}
	if len(b) != ed25519.SignatureSize {
		return nil, fmt.Errorf("invalid signature length: %d", len(b))
	}
	return b, nil
}
