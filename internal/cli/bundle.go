package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yzoz/entropy/internal/bundle"
)

func newExportCmd(opt *options) *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export the ledger into a portable bundle",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a := &app{opt: opt}
			return a.withStore(func() error {
				b, err := bundle.Export(a.st)
				if err != nil {
					return err
				}
				raw, err := b.Encode()
				if err != nil {
					return err
				}
				target := out
				if target == "" || target == "-" {
					_, err = os.Stdout.Write(raw)
					if err == nil {
						fmt.Println()
					}
					return err
				}
				if err := os.WriteFile(target, raw, 0o600); err != nil {
					return err
				}
				fmt.Printf("Exported %d contract(s) to %s\n", len(b.Contracts), target)
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&out, "out", "o", "-", "output file (default stdout)")
	return cmd
}

func newImportCmd(opt *options) *cobra.Command {
	return &cobra.Command{
		Use:   "import <bundle.json>",
		Short: "Verify and import a bundle into the ledger",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
