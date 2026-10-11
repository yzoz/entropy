package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yzoz/entropy/internal/bundle"
)

func newExportCmd(opt *options) *cobra.Command {
	var out, format string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export the ledger into a portable bundle or a CSV report",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			a := &app{opt: opt}
			return a.withStore(func() error {
				var raw []byte
				var count int
				switch format {
				case "json", "bundle":
					b, err := bundle.Export(a.st)
					if err != nil {
						return err
					}
					raw, err = b.Encode()
					if err != nil {
						return err
					}
					count = len(b.Contracts)
				case "csv":
					contracts, err := a.st.ListContracts()
					if err != nil {
						return err
					}
					raw, err = bundle.EncodeCSV(contracts)
					if err != nil {
						return err
					}
					count = len(contracts)
				default:
					return fmt.Errorf("unknown export format %q (want json or csv)", format)
				}
				target := out
				if target == "" || target == "-" {
					_, err := os.Stdout.Write(raw)
					return err
				}
				if err := os.WriteFile(target, raw, 0o600); err != nil {
					return err
				}
				fmt.Printf("Exported %d contract(s) to %s\n", count, target)
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&out, "out", "o", "-", "output file (default stdout)")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "output format {json|csv}")
	return cmd
}

func newImportCmd(opt *options) *cobra.Command {
	return &cobra.Command{
		Use:   "import <bundle.json>",
		Short: "Verify and import a bundle into the ledger",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			a := &app{opt: opt}
			raw, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			b, err := bundle.Decode(raw)
			if err != nil {
				return err
			}
			return a.withStore(func() error {
				res, err := bundle.Import(a.st, b)
				if err != nil {
					return err
				}
				if opt.JSON {
					a.printJSON(map[string]int{"added": res.Added, "skipped": res.Skipped})
					return nil
				}
				fmt.Printf("Imported %d contract(s), skipped %d already known.\n", res.Added, res.Skipped)
				return nil
			})
		},
	}
}
