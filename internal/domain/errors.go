package domain

import "errors"

// Sentinel errors returned by the domain layer. Callers can match them with
// errors.Is to produce user-friendly messages.
var (
	ErrUnknownCurrency = errors.New("unknown currency")
	ErrUnknownType     = errors.New("unknown contract type")
	ErrInvalidAmount   = errors.New("invalid amount")
	ErrMissingField    = errors.New("missing required field")
	ErrBadVersion      = errors.New("unsupported contract version")
)
