// Package domain defines the ledger's core types: contracts, currencies and
// their signing and validation rules.
package domain

import "time"

// Contract format versions.
const (
	// ContractVersion is the version written for new contracts.
	ContractVersion = 1
	// MinContractVersion is the oldest version this build can read.
	MinContractVersion = 1
)

// Type enumerates the supported contract kinds.
type Type string

const (
	// IOU is a debt note: From owes Amount to To until ExpiresAt.
	IOU Type = "IOU"
	// TRANSFER moves credit from From to To.
	TRANSFER Type = "TRANSFER"
	// SETTLE repays or nets out a debt between From and To.
	SETTLE Type = "SETTLE"
)

// Valid reports whether t is a known contract type.
func (t Type) Valid() bool {
	switch t {
	case IOU, TRANSFER, SETTLE:
		return true
	default:
		return false
	}
}

// Status enumerates the lifecycle states of a contract.
type Status string

const (
	// Draft means the contract exists but has not been countersigned.
	Draft Status = "draft"
	// Signed means both parties signed; the contract is in effect.
	Signed Status = "signed"
	// Settled means the obligation has been discharged.
	Settled Status = "settled"
	// Expired means ExpiresAt passed before settlement.
	Expired Status = "expired"
)

// Signature is one entry in a contract's signature chain.
type Signature struct {
	Signer   string `json:"signer"`
	Alg      string `json:"alg"`
	Value    string `json:"value"`
	SignedAt int64  `json:"signed_at"`
}

// Contract is a signed obligation between two participants. It is the single
// record stored in the ledger; balances are derived from the set of contracts.
type Contract struct {
	Version    int         `json:"version"`
	ID         string      `json:"id"`
	Type       Type        `json:"type"`
	From       string      `json:"from"`
	To         string      `json:"to"`
	Amount     int64       `json:"amount"` // minor units
	Currency   Currency    `json:"currency"`
	IssuedAt   int64       `json:"issued_at"`
	ExpiresAt  int64       `json:"expires_at"`
	Status     Status      `json:"status"`
	Signatures []Signature `json:"signatures"`
	Memo       string      `json:"memo,omitempty"`
	// Settles optionally references the ID of an IOU this contract repays.
	Settles string `json:"settles,omitempty"`
}

// HasSigner reports whether the contract already carries a signature by addr.
func (c *Contract) HasSigner(addr string) bool {
	for _, s := range c.Signatures {
		if s.Signer == addr {
			return true
		}
	}
	return false
}

// EffectiveStatus returns the status of the contract at time now, promoting
// Signed contracts whose ExpiresAt has passed to Expired.
func (c *Contract) EffectiveStatus(now time.Time) Status {
	if c.Status == Signed && c.ExpiresAt > 0 && now.Unix() > c.ExpiresAt {
		return Expired
	}
	return c.Status
}

// Validate checks structural invariants of the contract (not signatures).
func (c *Contract) Validate() error {
	if c.Version < MinContractVersion || c.Version > ContractVersion {
		return ErrBadVersion
	}
	if c.ID == "" || c.From == "" || c.To == "" {
		return ErrMissingField
	}
	if c.From == c.To {
		return ErrMissingField
	}
	if !c.Type.Valid() {
		return ErrUnknownType
	}
	if !c.Currency.Valid() {
		return ErrUnknownCurrency
	}
	if c.Amount <= 0 {
		return ErrInvalidAmount
	}
	if c.Status == "" {
		return ErrMissingField
	}
	return nil
}
