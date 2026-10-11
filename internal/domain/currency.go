package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// Currency is a currency or asset code supported by the ledger.
type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	RUR Currency = "RUR"
	GLD Currency = "GLD" // gold
	SLV Currency = "SLV" // silver
	OIL Currency = "OIL" // oil
)

// currencyInfo describes how a currency is represented on the wire and on
// screen. Amounts are always stored as integers in minor units; Scale tells
// how many decimal places those minor units represent.
type currencyInfo struct {
	Scale  int
	Symbol string
}

var currencies = map[Currency]currencyInfo{
	USD: {Scale: 2, Symbol: "$"},
	EUR: {Scale: 2, Symbol: "€"},
	RUR: {Scale: 2, Symbol: "₽"},
	GLD: {Scale: 4, Symbol: "Au"},
	SLV: {Scale: 4, Symbol: "Ag"},
	OIL: {Scale: 3, Symbol: "bbl"},
}

// Valid reports whether c is a known currency code.
func (c Currency) Valid() bool {
	_, ok := currencies[c]
	return ok
}

// Scale returns the number of decimal places used by the currency. Unknown
// currencies fall back to 2.
func (c Currency) Scale() int {
	if info, ok := currencies[c]; ok {
		return info.Scale
	}
	return 2
}

// String returns the currency code.
func (c Currency) String() string { return string(c) }

// Symbol returns a short human-readable symbol for the currency.
func (c Currency) Symbol() string {
	if info, ok := currencies[c]; ok {
		return info.Symbol
	}
	return string(c)
}

// ParseCurrency validates and normalizes a currency code (case-insensitive).
func ParseCurrency(s string) (Currency, error) {
	c := Currency(strings.ToUpper(strings.TrimSpace(s)))
	if !c.Valid() {
		return "", fmt.Errorf("%w: %q", ErrUnknownCurrency, s)
	}
	return c, nil
}

// FormatAmount renders integer minor units as a decimal string honouring the
// currency scale, e.g. (USD, 1234) => "12.34".
func FormatAmount(c Currency, minor int64) string {
	scale := c.Scale()
	if scale == 0 {
		return strconv.FormatInt(minor, 10)
	}
	neg := minor < 0
	u := minor
	if neg {
		u = -u
	}
	units := strconv.FormatInt(u, 10)
	for len(units) <= scale {
		units = "0" + units
	}
	whole := units[:len(units)-scale]
	frac := units[len(units)-scale:]
	if neg {
		return "-" + whole + "." + frac
	}
	return whole + "." + frac
}

// ParseAmount parses a decimal string into integer minor units for the given
// currency. It rejects values with more decimal places than the currency allows.
func ParseAmount(c Currency, s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(strings.TrimPrefix(s, "-"), "+")

	whole := s
	frac := ""
	if dot := strings.IndexByte(s, '.'); dot >= 0 {
		whole = s[:dot]
		frac = s[dot+1:]
	}
	if whole == "" {
		whole = "0"
	}
	scale := c.Scale()
	if len(frac) > scale {
		return 0, fmt.Errorf("too many decimal places: %q allows %d", c, scale)
	}
	frac += strings.Repeat("0", scale-len(frac))

	digits := whole + frac
	if digits == "" {
		return 0, fmt.Errorf("invalid amount %q", s)
	}
	v, err := strconv.ParseInt(digits, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount %q: %w", s, err)
	}
	if neg {
		v = -v
	}
	return v, nil
}
