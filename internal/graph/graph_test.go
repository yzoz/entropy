package graph

import (
	"strings"
	"testing"

	"github.com/yzoz/entropy/internal/domain"
	"github.com/yzoz/entropy/internal/ledger"
)

func TestRenderMermaidFormatsAmount(t *testing.T) {
	debts := []ledger.Debt{{From: "ed25519:a", To: "ed25519:b", Currency: domain.USD, Amount: 1250}}
	out := RenderMermaid(debts, func(s string) string { return strings.TrimPrefix(s, "ed25519:") })
	if !strings.Contains(out, "a\"] -->|12.50 USD|") {
		t.Errorf("mermaid output does not format the amount: %q", out)
	}
}

func TestRenderDOTFormatsAmount(t *testing.T) {
	debts := []ledger.Debt{{From: "ed25519:a", To: "ed25519:b", Currency: domain.GLD, Amount: 12345}}
	out := RenderDOT(debts, func(s string) string { return strings.TrimPrefix(s, "ed25519:") })
	if !strings.Contains(out, `label="1.2345 GLD"`) {
		t.Errorf("dot output does not format the amount: %q", out)
	}
}
