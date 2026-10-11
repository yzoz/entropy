package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd(opt *options) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if opt.JSON {
				(&app{opt: opt}).printJSON(map[string]string{
					"version": version,
					"commit":  commit,
					"date":    date,
				})
				return nil
			}
			fmt.Printf("entropy %s\n", version)
			fmt.Printf("  commit : %s\n", commit)
			fmt.Printf("  built  : %s\n", date)
			return nil
		},
	}
}
