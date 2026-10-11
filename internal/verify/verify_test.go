package verify

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/yzoz/entropy/internal/crypto"
	"github.com/yzoz/entropy/internal/domain"
	"github.com/yzoz/entropy/internal/sign"
	"github.com/yzoz/entropy/internal/store"
)

func TestVerifyDetectsTamper(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "l.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	pub, priv, _ := crypto.GenerateKey()
	pubHex := crypto.EncodePublic(pub)
	alice := &sign.Identity{Signer: crypto.NewSigner(priv), Address: crypto.Address(pub), PubHex: pubHex}
	if err := st.PutPublic(alice.Address, pubHex); err != nil {
		t.Fatal(err)
	}

	pub2, priv2, _ := crypto.GenerateKey()
	pub2Hex := crypto.EncodePublic(pub2)
	bob := &sign.Identity{Signer: crypto.NewSigner(priv2), Address: crypto.Address(pub2), PubHex: pub2Hex}
	if err := st.PutPublic(bob.Address, pub2Hex); err != nil {
		t.Fatal(err)
	}

	svc := &sign.Service{Store: st}
	c := &domain.Contract{
		Version: domain.ContractVersion, Type: domain.IOU,
		From: alice.Address, To: bob.Address, Amount: 1000, Currency: domain.USD,
		IssuedAt: time.Now().Unix(), Status: domain.Draft,
	}
	c.ID, _ = domain.DeriveID(c)
	if err := svc.Sign(c, alice); err != nil {
		t.Fatal(err)
	}
	if err := svc.Sign(c, bob); err != nil {
		t.Fatal(err)
	}
	if err := st.PutContract(c); err != nil {
		t.Fatal(err)
	}

	rep, err := Ledger(st)
	if err != nil {
		t.Fatal(err)
	}
	if !rep.OK() {
		t.Fatalf("clean ledger reported issues: %v", rep.Issues)
	}

	c.Amount = 999999
	if err := st.PutContract(c); err != nil {
		t.Fatal(err)
	}
	rep, err = Ledger(st)
	if err != nil {
		t.Fatal(err)
	}
	if rep.OK() {
		t.Fatal("tampered ledger must not be OK")
	}
}
