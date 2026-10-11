package domain

import (
	"testing"
	"time"
)

func sample() *Contract {
	return &Contract{
		Version:   ContractVersion,
		ID:        "abc123",
		Type:      IOU,
		From:      "ed25519:alice",
		To:        "ed25519:bob",
		Amount:    1250,
		Currency:  USD,
		IssuedAt:  1000,
		ExpiresAt: 2000,
		Status:    Draft,
	}
}

func TestSigningPayloadStableAcrossSignatures(t *testing.T) {
	c := sample()
	before := SigningPayload(c)
	c.Signatures = append(c.Signatures, Signature{Signer: "ed25519:alice", Alg: "ed25519", Value: "aa"})
	after := SigningPayload(c)
	if string(before) != string(after) {
		t.Error("adding a signature must not change the signing payload")
	}
	c.Status = Signed
	if string(before) != string(SigningPayload(c)) {
		t.Error("changing status must not change the signing payload")
	}
}

func TestSigningPayloadSensitiveToFields(t *testing.T) {
	a := sample()
	b := sample()
	b.Amount = 1251
	if string(SigningPayload(a)) == string(SigningPayload(b)) {
		t.Error("changing the amount must change the signing payload")
	}
}

func TestEffectiveStatus(t *testing.T) {
	c := sample()
	c.Status = Signed
	now := time.Unix(1500, 0)
	if c.EffectiveStatus(now) != Signed {
		t.Error("contract should still be signed before expiry")
	}
	later := time.Unix(2500, 0)
	if c.EffectiveStatus(later) != Expired {
		t.Error("contract should be expired after expiry")
	}
}

func TestValidate(t *testing.T) {
	if err := sample().Validate(); err != nil {
		t.Fatalf("valid contract rejected: %v", err)
	}
	bad := sample()
	bad.Amount = 0
	if err := bad.Validate(); err == nil {
		t.Error("zero-amount contract should be invalid")
	}
	bad = sample()
	bad.From = bad.To
	if err := bad.Validate(); err == nil {
		t.Error("self-dealing contract should be invalid")
	}
}
