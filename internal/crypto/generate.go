package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
)

// GenerateKey creates a fresh Ed25519 key pair.
func GenerateKey() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}
