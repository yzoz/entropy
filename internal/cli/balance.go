package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/yzoz/entropy/internal/domain"
	"github.com/yzoz/entropy/internal/ledger"
)

func newBalanceCmd(opt *options) *cobra.Command {
	return &cobra.Command{
		Use:   "balance",
		Short: "Show net balances derived from the ledger",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := &app{opt: opt}
			return a.withStore(func() error {
				contracts, err := a.st.ListContracts()
				if err != nil {
					return err
				}
				l := ledger.New(contracts)
				rows := l.Balances()
				if opt.JSON {
					out := make([]map[string]string, 0, len(rows))
					for _, r := range rows {
						out = append(out, map[string]string{
							"address":  r.Address,
							"party":    a.labelFor(r.Address),
							"currency": string(r.Currency),
							"net":      domain.FormatAmount(r.Currency, r.Net),
						})
					}
					a.printJSON(out)
					return nil
				}
				if len(rows) == 0 {
					fmt.Println("No balances.")
					return nil
				}
				sort.Slice(rows, func(i, j int) bool { return rows[i].Address < rows[j].Address })
				for _, r := range rows {
					fmt.Printf("%-16s %10s %s\n", a.labelFor(r.Address),
						domain.FormatAmount(r.Currency, r.Net), r.Currency)
				}
				return nil
			})
		},
	}
}
