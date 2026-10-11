package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yzoz/entropy/internal/crypto"
	"github.com/yzoz/entropy/internal/domain"
)

// publicForLocalKey reads the stored public key from a local key file without
// requiring the passphrase.
func publicForLocalKey(home, name string) ([]byte, error) {
	path := filepath.Join(home, "keys", name+".key")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read key %q: %w", name, err)
	}
	var k struct {
		Public string `json:"public"`
	}
	if err := json.Unmarshal(raw, &k); err != nil {
		return nil, fmt.Errorf("unreadable key file %q: %w", name, err)
	}
	return crypto.DecodePublic(k.Public)
}

// resolveAddress turns an alias or address into a canonical address using the
// open store's alias table.
func (a *app) resolveAddress(name string) (string, error) {
	addr := a.st.ResolveAlias(name)
	if _, err := crypto.ParseAddress(addr); err != nil {
		return "", fmt.Errorf("cannot resolve %q to an address: %w", name, err)
	}
	return addr, nil
}

// labelFor returns a display label for an address, preferring a known alias.
func (a *app) labelFor(addr string) string {
	aliases, err := a.st.ListAliases()
	if err != nil {
		return addr
	}
	for name, a2 := range aliases {
		if a2 == addr {
			return name
		}
	}
	return addr
}

// parseContractType validates a contract type flag value.
func parseContractType(s string) (domain.Type, error) {
	t := domain.Type(s)
	if !t.Valid() {
		return "", fmt.Errorf("unknown contract type %q (want IOU, TRANSFER or SETTLE)", s)
	}
	return t, nil
}
