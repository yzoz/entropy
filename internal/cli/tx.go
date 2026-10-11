package cli

import (
	"time"

	"github.com/yzoz/entropy/internal/domain"
	"github.com/yzoz/entropy/internal/sign"
)

// signer loads a local key (prompting for the passphrase when necessary) and
// returns a ready-to-use identity.
func (a *app) signer(keyName, passphrase string) (*sign.Identity, error) {
	if passphrase == "" {
		var err error
		passphrase, err = promptPassphrase(false)
		if err != nil {
			return nil, err
		}
	}
	svc := &sign.Service{Store: a.st, KeysDir: a.keys}
	return svc.Load(keyName, passphrase)
}

// signService returns a signing service bound to the open store.
func (a *app) signService() *sign.Service {
	return &sign.Service{Store: a.st, KeysDir: a.keys}
}

type createParams struct {
	Type       domain.Type
	FromKey    string
	Passphrase string
	To         string
	Amount     string
	Currency   string
	Expires    time.Duration
	Memo       string
	Settles    string
}

// createContract builds, self-signs and stores a new draft contract.
func (a *app) createContract(p createParams) (*domain.Contract, error) {
	from, err := a.signer(p.FromKey, p.Passphrase)
	if err != nil {
		return nil, err
	}
	to, err := a.resolveAddress(p.To)
	if err != nil {
		return nil, err
	}
	cur, err := domain.ParseCurrency(p.Currency)
	if err != nil {
		return nil, err
	}
	amt, err := domain.ParseAmount(cur, p.Amount)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	exp := int64(0)
	if p.Expires > 0 {
		exp = now.Add(p.Expires).Unix()
	}
	c := &domain.Contract{
		Version:   domain.ContractVersion,
		Type:      p.Type,
		From:      from.Address,
		To:        to,
		Amount:    amt,
		Currency:  cur,
		IssuedAt:  now.Unix(),
		ExpiresAt: exp,
		Status:    domain.Draft,
		Memo:      p.Memo,
		Settles:   p.Settles,
	}
	id, err := domain.DeriveID(c)
	if err != nil {
		return nil, err
	}
	c.ID = id
	if err := a.signService().Sign(c, from); err != nil {
		return nil, err
	}
	if err := a.st.PutContract(c); err != nil {
		return nil, err
	}
	return c, nil
}
