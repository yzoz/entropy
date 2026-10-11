package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yzoz/entropy/internal/graph"
	"github.com/yzoz/entropy/internal/ledger"
)

func newGraphCmd(opt *options) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Render the debt graph as Mermaid or DOT",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			a := &app{opt: opt}
			return a.withStore(func() error {
				contracts, err := a.st.ListContracts()
				if err != nil {
					return err
				}
				debts := ledger.New(contracts).Debts()
				label := func(addr string) string { return a.labelFor(addr) }
				switch format {
				case "mermaid":
					fmt.Print(graph.RenderMermaid(debts, label))
				case "dot":
					fmt.Print(graph.RenderDOT(debts, label))
				default:
					return fmt.Errorf("unknown format %q (want mermaid or dot)", format)
				}
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "mermaid", "output format {mermaid|dot}")
	return cmd
}
