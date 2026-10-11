package cli

// Build metadata. The zero defaults are overwritten at build time via
// -ldflags "-X github.com/yzoz/entropy/internal/cli.version=...". GoReleaser
// and the Makefile set these automatically.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Version returns the human-readable version string.
func Version() string { return version }
