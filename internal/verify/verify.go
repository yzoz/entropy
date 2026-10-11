// Package verify performs an independent audit of a ledger: it re-checks every
// signature, cross-checks statuses and references, and confirms that balances
// are internally consistent.
package verify

import (
	"fmt"
	"sort"

	"github.com/yzoz/entropy/internal/crypto"
	"github.com/yzoz/entropy/internal/domain"
	"github.com/yzoz/entropy/internal/ledger"
	"github.com/yzoz/entropy/internal/store"
)

// Issue describes a single problem found during an audit.
type Issue struct {
	ContractID string
	Problem    string
}

func (i Issue) String() string {
	if i.ContractID == "" {
		return i.Problem
	}
	return fmt.Sprintf("[%s] %s", i.ContractID, i.Problem)
}

// Report is the outcome of an audit.
type Report struct {
	Checked    int
	Issues     []Issue
	BalancesOK bool
}

// OK reports whether the ledger is fully consistent.
func (r *Report) OK() bool { return len(r.Issues) == 0 }

// Ledger audits every contract stored in st and returns a report. It never
// modifies the store.
func Ledger(st *store.Store) (*Report, error) {
	contracts, err := st.ListContracts()
	if err != nil {
		return nil, err
	}
	rep := &Report{Checked: len(contracts)}

	byID := map[string]*domain.Contract{}
	for _, c := range contracts {
		byID[c.ID] = c
	}

	for _, c := range contracts {
		if err := c.Validate(); err != nil {
			rep.Issues = append(rep.Issues, Issue{c.ID, "invalid contract: " + err.Error()})
			continue
		}
		payload := domain.SigningPayload(c)
		for _, sig := range c.Signatures {
			pubHex, ok := st.GetPublic(sig.Signer)
			if !ok {
				rep.Issues = append(rep.Issues, Issue{c.ID, "unknown signer public key for " + sig.Signer})
				continue
			}
			pub, err := crypto.DecodePublic(pubHex)
			if err != nil {
				rep.Issues = append(rep.Issues, Issue{c.ID, "bad public key for " + sig.Signer + ": " + err.Error()})
				continue
			}
			if crypto.Address(pub) != sig.Signer {
				rep.Issues = append(rep.Issues, Issue{c.ID, "public key does not match signer " + sig.Signer})
				continue
			}
			sigBytes, err := crypto.DecodeSignature(sig.Value)
			if err != nil {
				rep.Issues = append(rep.Issues, Issue{c.ID, "malformed signature by " + sig.Signer})
				continue
			}
			if !(crypto.Ed25519Verifier{}).Verify(pub, payload, sigBytes) {
				rep.Issues = append(rep.Issues, Issue{c.ID, "signature by " + sig.Signer + " does not verify"})
			}
		}

		if c.Status == domain.Signed && (!c.HasSigner(c.From) || !c.HasSigner(c.To)) {
			rep.Issues = append(rep.Issues, Issue{c.ID, "status is signed but a counterparty signature is missing"})
		}
		if c.Status == domain.Settled && !c.HasSigner(c.From) {
			rep.Issues = append(rep.Issues, Issue{c.ID, "status is settled but the debtor signature is missing"})
		}
		if c.Type == domain.SETTLE {
			if c.Settles == "" {
				rep.Issues = append(rep.Issues, Issue{c.ID, "SETTLE does not reference an IOU"})
			} else if ref, ok := byID[c.Settles]; !ok {
				rep.Issues = append(rep.Issues, Issue{c.ID, "SETTLE references unknown contract " + c.Settles})
			} else if ref.Type != domain.IOU {
				rep.Issues = append(rep.Issues, Issue{c.ID, "SETTLE references a non-IOU contract"})
			}
		}
	}

	// Balances must net to zero per currency: every unit owed by someone is
	// owed to someone else.
	l := ledger.New(contracts)
	perCurrency := map[domain.Currency]int64{}
	for _, b := range l.Balances() {
		perCurrency[b.Currency] += b.Net
	}
	rep.BalancesOK = true
	var bad []domain.Currency
	for cur, sum := range perCurrency {
		if sum != 0 {
			rep.BalancesOK = false
			bad = append(bad, cur)
		}
	}
	sort.Slice(bad, func(i, j int) bool { return bad[i] < bad[j] })
	for _, cur := range bad {
		rep.Issues = append(rep.Issues, Issue{"", fmt.Sprintf("balances for %s do not net to zero", cur)})
	}
	return rep, nil
}
