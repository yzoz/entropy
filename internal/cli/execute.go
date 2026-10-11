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

// ExecuteArgs runs the CLI in-process with the given arguments (excluding the
// program name). It returns the command error instead of exiting the process,
// which makes end-to-end testing straightforward.
func ExecuteArgs(args ...string) error {
	root := NewRootCmd()
	root.SetArgs(args)
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	return root.Execute()
}
