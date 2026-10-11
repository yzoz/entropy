package cli

import (
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/yzoz/entropy/internal/domain"
)

func newViewCmd(opt *options) *cobra.Command {
	var f struct{ typeF, status, direction, currency string }
	cmd := &cobra.Command{
		Use:   "view [id]",
		Short: "View contracts and their status",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := &app{opt: opt}
			return a.withStore(func() error {
				if len(args) == 1 {
					c, err := a.st.GetContract(args[0])
					if err != nil {
						return fmt.Errorf("contract %q not found", args[0])
					}
					if opt.JSON {
						a.printJSON(c)
						return nil
					}
					a.printContract(c)
					return nil
				}
				contracts, err := a.st.ListContracts()
				if err != nil {
					return err
				}
				filtered := contracts[:0]
				for _, c := range contracts {
					if f.typeF != "" && string(c.Type) != f.typeF {
						continue
					}
					if f.status != "" && string(c.EffectiveStatus(time.Now())) != f.status {
						continue
					}
					if f.currency != "" && string(c.Currency) != f.currency {
						continue
					}
					if f.direction != "" {
						addr := a.st.ResolveAlias(f.direction)
						if addr != c.From && addr != c.To {
							continue
						}
					}
					filtered = append(filtered, c)
				}
				sort.Slice(filtered, func(i, j int) bool { return filtered[i].IssuedAt > filtered[j].IssuedAt })
				if opt.JSON {
					out := make([]map[string]any, 0, len(filtered))
					for _, c := range filtered {
						out = append(out, a.contractSummary(c))
					}
					a.printJSON(out)
					return nil
				}
				if len(filtered) == 0 {
					fmt.Println("No contracts.")
					return nil
				}
				for _, c := range filtered {
					fmt.Println(a.shortLine(c))
				}
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&f.typeF, "type", "t", "", "filter by type {IOU|TRANSFER|SETTLE}")
	cmd.Flags().StringVarP(&f.status, "status", "s", "", "filter by status {draft|signed|settled|expired}")
	cmd.Flags().StringVarP(&f.currency, "currency", "c", "", "filter by currency")
	cmd.Flags().StringVarP(&f.direction, "party", "p", "", "filter by party alias or address")
	return cmd
}

func (a *app) printContract(c *domain.Contract) {
	fmt.Printf("ID       : %s\n", c.ID)
	fmt.Printf("Type     : %s\n", c.Type)
	fmt.Printf("From     : %s (%s)\n", a.labelFor(c.From), c.From)
	fmt.Printf("To       : %s (%s)\n", a.labelFor(c.To), c.To)
	fmt.Printf("Amount   : %s %s\n", domain.FormatAmount(c.Currency, c.Amount), c.Currency)
	fmt.Printf("Status   : %s\n", c.EffectiveStatus(time.Now()))
	fmt.Printf("Issued   : %s\n", time.Unix(c.IssuedAt, 0).UTC().Format(time.RFC3339))
	if c.ExpiresAt > 0 {
		fmt.Printf("Expires  : %s\n", time.Unix(c.ExpiresAt, 0).UTC().Format(time.RFC3339))
	}
	if c.Settles != "" {
		fmt.Printf("Settles  : %s\n", c.Settles)
	}
	if c.Memo != "" {
		fmt.Printf("Memo     : %s\n", c.Memo)
	}
	fmt.Printf("Signed by: %d\n", len(c.Signatures))
	for _, s := range c.Signatures {
		fmt.Printf("  - %s\n", a.labelFor(s.Signer))
	}
}
