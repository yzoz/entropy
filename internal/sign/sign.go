// Package sign.glues the crypto primitives to the ledger: it loads a
// passphrase-protected key from disk and applies/receives signatures on
// contracts, promoting their lifecycle status as counterparts sign.
package sign

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yzoz/entropy/internal/crypto"
	"github.com/yzoz/entropy/internal/domain"
	"github.com/yzoz/entropy/internal/store"
)

// Service signs contracts on behalf of locally held keys.
type Service struct {
	Store   *store.Store
	KeysDir string
}

// Identity is a loaded signer together with its derived address and public key.
type Identity struct {
	Signer  *crypto.Ed25519Signer
	Address string
	PubHex  string
}

// Load reads and decrypts the named key and registers its public key with the
// store so future verification succeeds.
func (s *Service) Load(keysName, passphrase string) (*Identity, error) {
	path := filepath.Join(s.KeysDir, keysName+".key")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read key %q: %w", keysName, err)
	}
	privHex, pubHex, err := crypto.DecryptPrivateKey(raw, passphrase)
	if err != nil {
		return nil, err
	}
	priv, err := crypto.DecodePrivate(privHex)
	if err != nil {
		return nil, err
	}
	pub, err := crypto.DecodePublic(pubHex)
	if err != nil {
		return nil, err
	}
	addr := crypto.Address(pub)
	if err := s.Store.PutPublic(addr, pubHex); err != nil {
		return nil, err
	}
	return &Identity{Signer: crypto.NewSigner(priv), Address: addr, PubHex: pubHex}, nil
}

// Sign applies id's signature to the contract, updating its status. The signer
// must be one of the two parties and must not have signed yet.
func (s *Service) Sign(c *domain.Contract, id *Identity) error {
	if c.Status == domain.Settled || c.Status == domain.Expired {
		return fmt.Errorf("contract %s is already %s", c.ID, c.Status)
	}
	if id.Address != c.From && id.Address != c.To {
		return fmt.Errorf("address %s is not a party to contract %s", id.Address, c.ID)
	}
	if c.HasSigner(id.Address) {
		return fmt.Errorf("contract %s is already signed by %s", c.ID, id.Address)
	}
	sig, err := id.Signer.Sign(domain.SigningPayload(c))
	if err != nil {
		return err
	}
	c.Signatures = append(c.Signatures, domain.Signature{
		Signer:   id.Address,
		Alg:      "ed25519",
		Value:    crypto.EncodeSignature(sig),
		SignedAt: time.Now().Unix(),
	})
	if c.HasSigner(c.From) && c.HasSigner(c.To) {
		c.Status = domain.Signed
	}
	return nil
}
