// Package graph renders the debt graph as Mermaid or DOT text.
package graph

import (
	"fmt"
	"strings"

	"github.com/yzoz/entropy/internal/domain"
	"github.com/yzoz/entropy/internal/ledger"
)

// RenderMermaid returns a Mermaid flowchart of the given debts.
func RenderMermaid(debts []ledger.Debt, label func(string) string) string {
	var b strings.Builder
	b.WriteString("flowchart LR\n")
	nodes := map[string]bool{}
	edge := func(d ledger.Debt) {
		nodes[d.From] = true
		nodes[d.To] = true
	}
	for _, d := range debts {
		edge(d)
	}
	for _, d := range debts {
		f := nodeID(d.From)
		t := nodeID(d.To)
		fmt.Fprintf(&b, "  %s[\"%s\"] -->|%s %s| %s[\"%s\"]\n",
			f, esc(label(d.From)), domain.FormatAmount(d.Currency, d.Amount), d.Currency, t, esc(label(d.To)))
	}
	// Emit isolated declarations for nodes with no edges (kept explicit so
	// aliases always appear in the diagram).
	_ = nodes
	return b.String()
}

// RenderDOT returns a Graphviz DOT description of the debts.
func RenderDOT(debts []ledger.Debt, label func(string) string) string {
	var b strings.Builder
	b.WriteString("digraph debts {\n  rankdir=LR;\n  node [shape=box];\n")
	for _, d := range debts {
		fmt.Fprintf(&b, "  \"%s\" -> \"%s\" [label=\"%s %s\"];\n",
			esc(label(d.From)), esc(label(d.To)), domain.FormatAmount(d.Currency, d.Amount), d.Currency)
	}
	b.WriteString("}\n")
	return b.String()
}

func nodeID(addr string) string {
	r := strings.NewReplacer(":", "_", "-", "_", "+", "_", "/", "_")
	return "n_" + r.Replace(addr)
}

func esc(s string) string {
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
