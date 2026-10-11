package cli

import (
	"fmt"
	"os"
)

// Execute runs the root command and terminates the process on error.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
