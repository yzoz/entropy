package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yzoz/entropy/internal/verify"
)

func newVerifyCmd(opt *options) *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Audit the ledger: signatures, statuses, references, balances",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := &app{opt: opt}
			return a.withStore(func() error {
				rep, err := verify.Ledger(a.st)
				if err != nil {
					return err
				}
				if opt.JSON {
					issues := make([]string, 0, len(rep.Issues))
					for _, i := range rep.Issues {
						issues = append(issues, i.String())
					}
					a.printJSON(map[string]any{
						"checked":     rep.Checked,
						"ok":          rep.OK(),
						"balances_ok": rep.BalancesOK,
						"issues":      issues,
					})
					return nil
				}
				fmt.Printf("Checked %d contracts.\n", rep.Checked)
				if rep.OK() {
					fmt.Println("OK — ledger is consistent.")
					return nil
				}
				for _, i := range rep.Issues {
					fmt.Println("  ✗ " + i.String())
				}
				return fmt.Errorf("%d issue(s) found", len(rep.Issues))
			})
		},
	}
}
