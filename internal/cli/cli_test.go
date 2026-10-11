package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yzoz/entropy/internal/domain"
	"github.com/yzoz/entropy/internal/ledger"
	"github.com/yzoz/entropy/internal/store"
)

// run executes a CLI invocation against a throwaway home directory.
func run(t *testing.T, home string, args ...string) error {
	t.Helper()
	full := append([]string{"--home", home}, args...)
	return ExecuteArgs(full...)
}

// mustRun fails the test if the command returns an error.
func mustRun(t *testing.T, home string, args ...string) {
	t.Helper()
	if err := run(t, home, args...); err != nil {
		t.Fatalf("command %v failed: %v", args, err)
	}
}

// withStore inspects the ledger written by the CLI and closes it before
// returning, since bbolt takes an exclusive file lock.
func withStore(t *testing.T, home string, fn func(*store.Store)) {
	t.Helper()
	st, err := store.Open(filepath.Join(home, "ledger.db"))
	if err != nil {
		t.Fatalf("cannot open store: %v", err)
	}
	defer st.Close()
	fn(st)
}

// netBalance returns the net balance of the given alias in a currency.
func netBalance(t *testing.T, home, alias, cur string) int64 {
	t.Helper()
	var got int64
	withStore(t, home, func(st *store.Store) {
		aliases, err := st.ListAliases()
		if err != nil {
			t.Fatal(err)
		}
		addr := aliases[alias]
		if addr == "" {
			t.Fatalf("alias %q is not registered", alias)
		}
		contracts, err := st.ListContracts()
		if err != nil {
			t.Fatal(err)
		}
		got = ledger.New(contracts).BalanceOf(addr, domain.Currency(cur))
	})
	return got
}

// contractsOf returns a snapshot of the stored contracts.
func contractsOf(t *testing.T, home string) []*domain.Contract {
	t.Helper()
	var out []*domain.Contract
	withStore(t, home, func(st *store.Store) {
		c, err := st.ListContracts()
		if err != nil {
			t.Fatal(err)
		}
		out = c
	})
	return out
}

func TestEndToEndLifecycle(t *testing.T) {
	home := t.TempDir()

	mustRun(t, home, "key", "gen", "-n", "alice", "--passphrase", "pw")
	mustRun(t, home, "key", "gen", "-n", "bob", "--passphrase", "pw")
	if _, err := os.Stat(filepath.Join(home, "keys", "alice.key")); err != nil {
		t.Fatalf("alice key not created: %v", err)
	}
	mustRun(t, home, "id", "set", "--name", "alice", "--key", "alice")
	mustRun(t, home, "id", "set", "--name", "bob", "--key", "bob")

	mustRun(t, home, "issue", "--to", "bob", "--amount", "12.50",
		"--currency", "USD", "--key", "alice", "--passphrase", "pw")

	contracts := contractsOf(t, home)
	if len(contracts) != 1 || contracts[0].Type != domain.IOU || contracts[0].Status != domain.Draft {
		t.Fatalf("unexpected contract set: %+v", contracts)
	}
	id := contracts[0].ID

	// A draft must not affect balances.
	if got := netBalance(t, home, "alice", "USD"); got != 0 {
		t.Fatalf("draft affected alice balance: %d", got)
	}

	// Counterparty signs; balances must move.
	mustRun(t, home, "sign", "--id", id, "--key", "bob", "--passphrase", "pw")
	if err := run(t, home, "verify"); err != nil {
		t.Fatalf("verify after signing failed: %v", err)
	}
	if got := netBalance(t, home, "alice", "USD"); got != -1250 {
		t.Fatalf("alice balance = %d, want -1250", got)
	}
	if got := netBalance(t, home, "bob", "USD"); got != 1250 {
		t.Fatalf("bob balance = %d, want 1250", got)
	}

	// Settle and countersign the settlement.
	mustRun(t, home, "settle", "--iou", id, "--key", "alice", "--passphrase", "pw")
	var settleID string
	for _, c := range contractsOf(t, home) {
		if c.Type == domain.SETTLE {
			settleID = c.ID
		}
	}
	if settleID == "" {
		t.Fatal("settle contract not created")
	}
	mustRun(t, home, "sign", "--id", settleID, "--key", "bob", "--passphrase", "pw")

	if err := run(t, home, "verify"); err != nil {
		t.Fatalf("verify after settle failed: %v", err)
	}
	if got := netBalance(t, home, "alice", "USD"); got != 0 {
		t.Fatalf("alice did not net to zero: %d", got)
	}
	if got := netBalance(t, home, "bob", "USD"); got != 0 {
		t.Fatalf("bob did not net to zero: %d", got)
	}
}

func TestExportImportViaCLI(t *testing.T) {
	home := t.TempDir()
	mustRun(t, home, "key", "gen", "-n", "carol", "--passphrase", "pw")
	mustRun(t, home, "key", "gen", "-n", "dave", "--passphrase", "pw")
	mustRun(t, home, "id", "set", "--name", "carol", "--key", "carol")
	mustRun(t, home, "id", "set", "--name", "dave", "--key", "dave")
	mustRun(t, home, "send", "--to", "dave", "--amount", "5", "--currency", "GLD",
		"--key", "carol", "--passphrase", "pw")

	st := contractsOf(t, home)
	mustRun(t, home, "sign", "--id", st[0].ID, "--key", "dave", "--passphrase", "pw")

	bundlePath := filepath.Join(t.TempDir(), "b.json")
	mustRun(t, home, "export", "-o", bundlePath)
	if _, err := os.Stat(bundlePath); err != nil {
		t.Fatalf("bundle not written: %v", err)
	}

	fresh := t.TempDir()
	mustRun(t, fresh, "import", bundlePath)
	if err := run(t, fresh, "verify"); err != nil {
		t.Fatalf("verify in fresh home failed: %v", err)
	}
	if got := contractsOf(t, fresh); len(got) != 1 {
		t.Fatalf("expected 1 imported contract, got %d", len(got))
	}
}

func TestExportCSV(t *testing.T) {
	home := t.TempDir()
	mustRun(t, home, "key", "gen", "-n", "eve", "--passphrase", "pw")
	mustRun(t, home, "key", "gen", "-n", "mallory", "--passphrase", "pw")
	mustRun(t, home, "id", "set", "--name", "eve", "--key", "eve")
	mustRun(t, home, "id", "set", "--name", "mallory", "--key", "mallory")
	mustRun(t, home, "issue", "--to", "mallory", "--amount", "1.23",
		"--currency", "EUR", "--key", "eve", "--passphrase", "pw")

	csvPath := filepath.Join(t.TempDir(), "report.csv")
	mustRun(t, home, "export", "-f", "csv", "-o", csvPath)
	raw, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	header := "id,type,from,to,amount,currency"
	if !strings.HasPrefix(string(raw), header) {
		t.Fatalf("unexpected CSV header: %q", string(raw))
	}
	if !strings.Contains(string(raw), "1.23") || !strings.Contains(string(raw), "EUR") {
		t.Fatalf("CSV does not contain the contract: %q", string(raw))
	}
}

func TestErrorsRejected(t *testing.T) {
	home := t.TempDir()
	mustRun(t, home, "key", "gen", "-n", "k", "--passphrase", "pw")
	mustRun(t, home, "id", "set", "--name", "k", "--key", "k")

	if err := run(t, home, "verify"); err != nil {
		t.Errorf("empty verify should succeed, got %v", err)
	}

	cases := [][]string{
		{"issue", "--to", "k", "--amount", "1", "--currency", "XYZ", "--key", "k", "--passphrase", "pw"},
		{"issue", "--to", "k", "--amount", "1.234", "--currency", "USD", "--key", "k", "--passphrase", "pw"},
		{"issue", "--to", "nobody", "--amount", "1", "--currency", "USD", "--key", "k", "--passphrase", "pw"},
		{"sign", "--id", "does-not-exist", "--key", "k", "--passphrase", "pw"},
		{"issue", "--to", "k", "--amount", "1", "--currency", "USD", "--key", "k", "--passphrase", "wrong"},
	}
	for _, args := range cases {
		if err := run(t, home, args...); err == nil {
			t.Errorf("expected error for %v", args)
		}
	}
}

func TestSettleRejectsNonDebtor(t *testing.T) {
	home := t.TempDir()
	mustRun(t, home, "key", "gen", "-n", "alice", "--passphrase", "pw")
	mustRun(t, home, "key", "gen", "-n", "bob", "--passphrase", "pw")
	mustRun(t, home, "id", "set", "--name", "alice", "--key", "alice")
	mustRun(t, home, "id", "set", "--name", "bob", "--key", "bob")
	mustRun(t, home, "issue", "--to", "bob", "--amount", "10", "--currency", "USD",
		"--key", "alice", "--passphrase", "pw")

	st := contractsOf(t, home)
	// bob is the creditor, so he must not be allowed to settle the IOU.
	if err := run(t, home, "settle", "--iou", st[0].ID, "--key", "bob", "--passphrase", "pw"); err == nil {
		t.Fatal("creditor should not be able to settle the debtor's IOU")
	}
}
