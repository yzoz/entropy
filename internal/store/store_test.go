package store

import (
	"path/filepath"
	"testing"

	"github.com/yzoz/entropy/internal/domain"
)

func TestContractCRUD(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "l.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	c := &domain.Contract{
		Version: domain.ContractVersion, ID: "x", Type: domain.IOU,
		From: "a", To: "b", Amount: 1, Currency: domain.USD, Status: domain.Draft,
	}
	if err := st.PutContract(c); err != nil {
		t.Fatal(err)
	}
	got, err := st.GetContract("x")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "x" || got.Amount != 1 {
		t.Fatalf("unexpected contract %+v", got)
	}
	if _, err := st.GetContract("missing"); err != ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	list, err := st.ListContracts()
	if err != nil || len(list) != 1 {
		t.Fatalf("ListContracts = %v, %v", list, err)
	}
}

func TestAliasResolution(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "l.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	if err := st.SetAlias("alice", "ed25519:xyz"); err != nil {
		t.Fatal(err)
	}
	if got := st.ResolveAlias("alice"); got != "ed25519:xyz" {
		t.Fatalf("ResolveAlias = %q", got)
	}
	if got := st.ResolveAlias("unknown"); got != "unknown" {
		t.Fatalf("unknown alias should pass through, got %q", got)
	}
}
