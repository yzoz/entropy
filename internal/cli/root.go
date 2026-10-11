// Package cli wires the Entropy Money command tree to the domain packages.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/yzoz/entropy/internal/domain"
	"github.com/yzoz/entropy/internal/store"
)

// options holds global flags shared by every command.
type options struct {
	Home string
	JSON bool
}

// app carries resolved runtime state (open store, paths) between commands.
type app struct {
	opt  *options
	st   *store.Store
	keys string
}

// active is set by withStore so nested helpers can reach the open store.
var active *app

func (a *app) keysDir() string { return filepath.Join(a.opt.Home, "keys") }
func (a *app) dbPath() string  { return filepath.Join(a.opt.Home, "ledger.db") }

// NewRootCmd builds the root command with all subcommands attached.
func NewRootCmd() *cobra.Command {
	opt := &options{}
	root := &cobra.Command{
		Use:   "entropy",
		Short: "Entropy Money — signed mutual-credit contracts",
		Long: "Entropy Money is a CLI for partner mutual credit (IOU) built on\n" +
			"signed contracts. Participants issue, countersign and settle\n" +
			"obligations, exchanging them as portable signed bundles.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&opt.Home, "home", defaultHome(), "data directory (default ~/.entropy)")
	root.PersistentFlags().BoolVar(&opt.JSON, "json", false, "emit machine-readable JSON")

	root.AddCommand(
		newKeyCmd(opt),
		newIDCmd(opt),
		newIssueCmd(opt),
		newSendCmd(opt),
		newSignCmd(opt),
		newSettleCmd(opt),
		newViewCmd(opt),
		newBalanceCmd(opt),
		newVerifyCmd(opt),
		newGraphCmd(opt),
		newExportCmd(opt),
		newImportCmd(opt),
	)
	return root
}

func defaultHome() string {
	if h := os.Getenv("ENTROPY_HOME"); h != "" {
		return h
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".entropy"
	}
	return filepath.Join(home, ".entropy")
}

// withStore opens the ledger, runs fn, and closes it afterwards.
func (a *app) withStore(fn func() error) error {
	if err := os.MkdirAll(a.opt.Home, 0o700); err != nil {
		return err
	}
	st, err := store.Open(a.dbPath())
	if err != nil {
		return err
	}
	defer st.Close()
	a.st = st
	a.keys = a.keysDir()
	active = a
	return fn()
}

// contractSummary is the machine-readable shape printed for a single contract.
func (a *app) contractSummary(c *domain.Contract) map[string]any {
	return map[string]any{
		"id":         c.ID,
		"type":       c.Type,
		"from":       c.From,
		"to":         c.To,
		"amount":     domain.FormatAmount(c.Currency, c.Amount),
		"currency":   c.Currency,
		"status":     c.EffectiveStatus(time.Now()),
		"issued_at":  c.IssuedAt,
		"expires_at": c.ExpiresAt,
		"signatures": len(c.Signatures),
		"settles":    c.Settles,
	}
}

// shortLine renders a contract as a single human-readable line.
func (a *app) shortLine(c *domain.Contract) string {
	st := c.EffectiveStatus(time.Now())
	return fmt.Sprintf("%s  %-8s  %s -> %s  %s %s  [%s]",
		c.ID, c.Type, a.labelFor(c.From), a.labelFor(c.To),
		domain.FormatAmount(c.Currency, c.Amount), c.Currency, st)
}

// printJSON emits v as indented JSON when --json is set and reports whether it
// handled the output.
func (a *app) printJSON(v any) bool {
	if !a.opt.JSON {
		return false
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return true
	}
	fmt.Println(string(b))
	return true
}
