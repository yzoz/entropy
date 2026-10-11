package ledger

import (
	"testing"

	"github.com/yzoz/entropy/internal/domain"
)

func ioU(id, from, to string, amt int64, status domain.Status) *domain.Contract {
	return &domain.Contract{
		Version: domain.ContractVersion, ID: id, Type: domain.IOU,
		From: from, To: to, Amount: amt, Currency: domain.USD, Status: status,
	}
}

func TestBalanceFromIOU(t *testing.T) {
	l := New([]*domain.Contract{
		ioU("1", "alice", "bob", 1250, domain.Signed),
	})
	if got := l.BalanceOf("alice", domain.USD); got != -1250 {
		t.Errorf("alice = %d, want -1250", got)
	}
	if got := l.BalanceOf("bob", domain.USD); got != 1250 {
		t.Errorf("bob = %d, want 1250", got)
	}
}

func TestDraftIgnored(t *testing.T) {
	l := New([]*domain.Contract{
		ioU("1", "alice", "bob", 1250, domain.Draft),
	})
	if got := l.BalanceOf("alice", domain.USD); got != 0 {
		t.Errorf("draft contract must not affect balances, got %d", got)
	}
}

func TestSettleNetsToZero(t *testing.T) {
	settle := &domain.Contract{
		Version: domain.ContractVersion, ID: "2", Type: domain.SETTLE,
		From: "alice", To: "bob", Amount: 1250, Currency: domain.USD,
		Status: domain.Signed, Settles: "1",
	}
	l := New([]*domain.Contract{
		ioU("1", "alice", "bob", 1250, domain.Signed),
		settle,
	})
	if got := l.BalanceOf("alice", domain.USD); got != 0 {
		t.Errorf("alice after settle = %d, want 0", got)
	}
	if got := l.BalanceOf("bob", domain.USD); got != 0 {
		t.Errorf("bob after settle = %d, want 0", got)
	}
}

func TestDebtsNetOutSettlement(t *testing.T) {
	settle := &domain.Contract{
		Version: domain.ContractVersion, ID: "2", Type: domain.SETTLE,
		From: "alice", To: "bob", Amount: 500, Currency: domain.USD,
		Status: domain.Signed, Settles: "1",
	}
	l := New([]*domain.Contract{
		ioU("1", "alice", "bob", 1250, domain.Signed),
		settle,
	})
	debts := l.Debts()
	if len(debts) != 1 || debts[0].Amount != 750 {
		t.Fatalf("expected one debt of 750, got %+v", debts)
	}
}
