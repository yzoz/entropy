// Package ledger derives balances from a set of contracts. The set of
// contracts is the source of truth; balances are a computed projection.
package ledger

import (
	"sort"

	"github.com/yzoz/entropy/internal/domain"
)

// Entry is a signed balance contribution of a single contract toward one
// participant, expressed as a delta applied to that participant's balance.
type Entry struct {
	ContractID string
	Type       domain.Type
	From       string
	To         string
	Delta      int64 // change to the balance of From (negative) after settling
}

// Balance is a participant's net position in a single currency. A negative
// figure means the participant owes value to the network.
type Balance struct {
	Address  string
	Currency domain.Currency
	Net      int64
}

// Ledger is an immutable view over a set of contracts.
type Ledger struct {
	contracts []*domain.Contract
	active    map[string]bool
}

// New builds a ledger from contracts whose status counts toward balances.
// Only contracts in the Signed status are considered: IOU commits the debtor,
// TRANSFER moves credit immediately and SETTLE reduces a referenced IOU.
func New(contracts []*domain.Contract) *Ledger {
	l := &Ledger{active: map[string]bool{}}
	for _, c := range contracts {
		l.contracts = append(l.contracts, c)
		if c.Status == domain.Signed {
			l.active[c.ID] = true
		}
	}
	return l
}

// Contracts returns the contracts in the ledger.
func (l *Ledger) Contracts() []*domain.Contract { return l.contracts }

// Balance computes the net position per participant per currency.
func (l *Ledger) Balance() map[string]map[domain.Currency]int64 {
	bal := map[string]map[domain.Currency]int64{}
	add := func(addr string, cur domain.Currency, delta int64) {
		if bal[addr] == nil {
			bal[addr] = map[domain.Currency]int64{}
		}
		bal[addr][cur] += delta
	}

	// Multi-pass so that settlement of an IOU is applied after all IOUs are
	// known, regardless of the order contracts are stored in.
	for _, c := range l.contracts {
		if !l.active[c.ID] {
			continue
		}
		switch c.Type {
		case domain.IOU:
			// Debtor owes the creditor: debtor balance goes down, creditor up.
			add(c.From, c.Currency, -c.Amount)
			add(c.To, c.Currency, +c.Amount)
		case domain.TRANSFER:
			// Credit moves from sender to receiver immediately.
			add(c.From, c.Currency, -c.Amount)
			add(c.To, c.Currency, +c.Amount)
		case domain.SETTLE:
			// Settling repays the debtor: their balance improves.
			add(c.From, c.Currency, +c.Amount)
			add(c.To, c.Currency, -c.Amount)
		}
	}
	return bal
}

// BalanceOf returns the net position of a single address in a currency.
func (l *Ledger) BalanceOf(address string, cur domain.Currency) int64 {
	return l.Balance()[address][cur]
}

// Balances flattens the balance map into a sorted slice for display.
func (l *Ledger) Balances() []Balance {
	bal := l.Balance()
	var out []Balance
	for addr, byCur := range bal {
		for cur, net := range byCur {
			out = append(out, Balance{Address: addr, Currency: cur, Net: net})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Address != out[j].Address {
			return out[i].Address < out[j].Address
		}
		return out[i].Currency < out[j].Currency
	})
	return out
}

// Debt is a directed edge in the debt graph.
type Debt struct {
	From     string
	To       string
	Currency domain.Currency
	Amount   int64
}

// Debts aggregates outstanding IOU obligations, netting out settlements.
func (l *Ledger) Debts() []Debt {
	type key struct {
		from, to string
		cur      domain.Currency
	}
	agg := map[key]int64{}
	settled := map[string]int64{} // IOU id -> settled amount
	var ious []*domain.Contract

	for _, c := range l.contracts {
		if !l.active[c.ID] {
			continue
		}
		switch c.Type {
		case domain.IOU:
			ious = append(ious, c)
		case domain.SETTLE:
			if c.Settles != "" {
				settled[c.Settles] += c.Amount
			}
		}
	}
	for _, c := range ious {
		remaining := c.Amount - settled[c.ID]
		if remaining <= 0 {
			continue
		}
		agg[key{c.From, c.To, c.Currency}] += remaining
	}
	var out []Debt
	for k, amt := range agg {
		out = append(out, Debt{From: k.from, To: k.to, Currency: k.cur, Amount: amt})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}
