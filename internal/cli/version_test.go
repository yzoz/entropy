package cli

import (
	"strings"
	"testing"
)

func TestVersionDefaults(t *testing.T) {
	if Version() == "" {
		t.Fatal("version must not be empty")
	}
	if err := run(t, t.TempDir(), "version"); err != nil {
		t.Fatalf("version command failed: %v", err)
	}
}

func TestRootHasVersionFlag(t *testing.T) {
	root := NewRootCmd()
	if !strings.Contains(root.Version, "commit") {
		t.Errorf("root version string looks wrong: %q", root.Version)
	}
	// cobra registers --version during InitDefaultVersionFlag; trigger it.
	root.InitDefaultVersionFlag()
	if flag := root.Flags().Lookup("version"); flag == nil {
		t.Fatal("root command is missing the --version flag")
	}
}
