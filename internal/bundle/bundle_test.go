package bundle

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/yzoz/entropy/internal/crypto"
	"github.com/yzoz/entropy/internal/domain"
	"github.com/yzoz/entropy/internal/sign"
	"github.com/yzoz/entropy/internal/store"
)

// newStore opens a throwaway store in the test's temp directory.
func newStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "ledger.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

// identity creates a fresh party and registers its public key with st.
func identity(t *testing.T, st *store.Store) *sign.Identity {
	t.Helper()
	pub, priv, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	pubHex := crypto.EncodePublic(pub)
	addr := crypto.Address(pub)
	if err := st.PutPublic(addr, pubHex); err != nil {
		t.Fatal(err)
	}
	return &sign.Identity{Signer: crypto.NewSigner(priv), Address: addr, PubHex: pubHex}
}

func signedIOU(t *testing.T, st *store.Store, from, to *sign.Identity) *domain.Contract {
	t.Helper()
	svc := &sign.Service{Store: st}
	c := &domain.Contract{
		Version: domain.ContractVersion, Type: domain.IOU,
		From: from.Address, To: to.Address, Amount: 1250, Currency: domain.USD,
		IssuedAt: time.Now().Unix(), Status: domain.Draft,
	}
	id, _ := domain.DeriveID(c)
	c.ID = id
	if err := svc.Sign(c, from); err != nil {
		t.Fatal(err)
	}
	if err := svc.Sign(c, to); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestExportImportRoundTrip(t *testing.T) {
	src := newStore(t)
	alice := identity(t, src)
	bob := identity(t, src)
	c := signedIOU(t, src, alice, bob)
	if err := src.PutContract(c); err != nil {
		t.Fatal(err)
	}
	if err := src.SetAlias("alice", alice.Address); err != nil {
		t.Fatal(err)
	}

	b, err := Export(src)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := b.Encode()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}

	dst := newStore(t)
	res, err := Import(dst, decoded)
	if err != nil {
		t.Fatal(err)
	}
	if res.Added != 1 {
		t.Fatalf("added = %d, want 1", res.Added)
	}
	if _, err := dst.GetContract(c.ID); err != nil {
		t.Fatalf("contract missing after import: %v", err)
	}

	// Re-import must be idempotent.
	res2, err := Import(dst, decoded)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Added != 0 || res2.Skipped != 1 {
		t.Fatalf("re-import = %+v, want added 0 skipped 1", res2)
	}
}

func TestImportRejectsTamperedContract(t *testing.T) {
	src := newStore(t)
	alice := identity(t, src)
	bob := identity(t, src)
	c := signedIOU(t, src, alice, bob)

	// Tamper with the amount after signing.
	c.Amount = 999999

	b := &Bundle{Version: FormatVersion, Contracts: []*domain.Contract{c}, Publics: map[string]string{
		alice.Address: alice.PubHex,
		bob.Address:   bob.PubHex,
	}}
	if _, err := Import(newStore(t), b); err == nil {
		t.Fatal("import of a tampered contract must fail")
	}
}

func TestImportRejectsUnknownSigner(t *testing.T) {
	src := newStore(t)
	alice := identity(t, src)
	bob := identity(t, src)
	c := signedIOU(t, src, alice, bob)

	b := &Bundle{Version: FormatVersion, Contracts: []*domain.Contract{c}, Publics: map[string]string{
		alice.Address: alice.PubHex,
	}}
	if _, err := Import(newStore(t), b); err == nil {
		t.Fatal("import with a missing public key must fail")
	}
}

func TestDecodeRejectsFutureVersion(t *testing.T) {
	if _, err := Decode([]byte(`{"version":999,"contracts":[]}`)); err == nil {
		t.Fatal("unsupported bundle version must be rejected")
	}
}
