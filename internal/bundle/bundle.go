// Package bundle implements the portable exchange format used by the
// `export` and `import` commands. A bundle is a single self-contained JSON file
// carrying contracts together with the public keys and aliases needed to verify
// them on the receiving side.
package bundle

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/yzoz/entropy/internal/crypto"
	"github.com/yzoz/entropy/internal/domain"
	"github.com/yzoz/entropy/internal/store"
)

// FormatVersion is the version of the bundle envelope itself.
const FormatVersion = 1

// Bundle is the on-disk exchange format.
type Bundle struct {
	Version    int                `json:"version"`
	ExportedAt int64              `json:"exported_at"`
	Aliases    map[string]string  `json:"aliases,omitempty"`
	Publics    map[string]string  `json:"publics,omitempty"`
	Contracts  []*domain.Contract `json:"contracts"`
}

// Export collects every stored contract into a bundle.
func Export(st *store.Store) (*Bundle, error) {
	contracts, err := st.ListContracts()
	if err != nil {
		return nil, err
	}
	aliases, err := st.ListAliases()
	if err != nil {
		return nil, err
	}
	publics, err := st.ListPublics()
	if err != nil {
		return nil, err
	}
	return &Bundle{
		Version:    FormatVersion,
		ExportedAt: time.Now().Unix(),
		Aliases:    aliases,
		Publics:    publics,
		Contracts:  contracts,
	}, nil
}

// Encode serializes a bundle to indented JSON.
func (b *Bundle) Encode() ([]byte, error) {
	return json.MarshalIndent(b, "", "  ")
}

// Decode parses a bundle and checks its format version.
func Decode(raw []byte) (*Bundle, error) {
	var b Bundle
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, fmt.Errorf("unreadable bundle: %w", err)
	}
	if b.Version != FormatVersion {
		return nil, fmt.Errorf("unsupported bundle version %d (this build supports %d)", b.Version, FormatVersion)
	}
	return &b, nil
}

// Result summarizes an import run.
type Result struct {
	Added   int
	Skipped int
}

// Import verifies every contract in the bundle against its signer public keys
// before adding it to the store. It is all-or-nothing: if any contract fails
// validation or signature checks, nothing is written and an error is returned.
// Contracts already present are skipped, which makes re-imports idempotent.
func Import(st *store.Store, b *Bundle) (*Result, error) {
	known := map[string]string{}
	for addr, pub := range b.Publics {
		known[addr] = pub
	}
	if stored, err := st.ListPublics(); err == nil {
		for addr, pub := range stored {
			known[addr] = pub
		}
	}

	res := &Result{}
	var toAdd []*domain.Contract
	for _, c := range b.Contracts {
		if err := c.Validate(); err != nil {
			return nil, fmt.Errorf("contract %s is invalid: %w", c.ID, err)
		}
		if err := verifySignatures(c, known); err != nil {
			return nil, err
		}
		exists, err := st.HasContract(c.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			res.Skipped++
			continue
		}
		toAdd = append(toAdd, c)
	}

	for addr, pub := range b.Publics {
		if err := st.PutPublic(addr, pub); err != nil {
			return nil, err
		}
	}
	aliases := make([]string, 0, len(b.Aliases))
	for a := range b.Aliases {
		aliases = append(aliases, a)
	}
	sort.Strings(aliases)
	for _, a := range aliases {
		if err := st.SetAlias(a, b.Aliases[a]); err != nil {
			return nil, err
		}
	}
	for _, c := range toAdd {
		if err := st.PutContract(c); err != nil {
			return nil, err
		}
		res.Added++
	}
	return res, nil
}

func verifySignatures(c *domain.Contract, known map[string]string) error {
	if len(c.Signatures) == 0 {
		return fmt.Errorf("contract %s has no signatures", c.ID)
	}
	payload := domain.SigningPayload(c)
	seen := map[string]bool{}
	for _, sig := range c.Signatures {
		if seen[sig.Signer] {
			return fmt.Errorf("contract %s has a duplicate signature by %s", c.ID, sig.Signer)
		}
		seen[sig.Signer] = true

		pubHex, ok := known[sig.Signer]
		if !ok {
			return fmt.Errorf("contract %s: unknown public key for signer %s", c.ID, sig.Signer)
		}
		pub, err := crypto.DecodePublic(pubHex)
		if err != nil {
			return fmt.Errorf("contract %s: bad public key for %s: %w", c.ID, sig.Signer, err)
		}
		if crypto.Address(pub) != sig.Signer {
			return fmt.Errorf("contract %s: public key does not match signer %s", c.ID, sig.Signer)
		}
		sigBytes, err := crypto.DecodeSignature(sig.Value)
		if err != nil {
			return fmt.Errorf("contract %s: malformed signature by %s", c.ID, sig.Signer)
		}
		if !(crypto.Ed25519Verifier{}).Verify(pub, payload, sigBytes) {
			return fmt.Errorf("contract %s: signature by %s does not verify", c.ID, sig.Signer)
		}
	}
	if c.Status == domain.Signed && (!c.HasSigner(c.From) || !c.HasSigner(c.To)) {
		return fmt.Errorf("contract %s: marked signed but a party signature is missing", c.ID)
	}
	return nil
}
