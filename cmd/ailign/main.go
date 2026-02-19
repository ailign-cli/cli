package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/ailign/cli/internal/cli"
)

// version and commit are set by goreleaser via ldflags.
var (
	version = "dev"
	commit  = "none"
)

func main() {
	rootCmd := cli.NewRootCommand()
	rootCmd.Version = resolveVersion()
	if err := rootCmd.Execute(); err != nil {
		if !errors.Is(err, cli.ErrAlreadyReported) {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(2)
	}
}

// resolveVersion returns the version string. GoReleaser-built binaries use
// ldflags; go install-built binaries fall back to Go module build info.
func resolveVersion() string {
	if version != "dev" {
		return version + " (" + commit + ")"
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version + " (" + commit + ")"
}
