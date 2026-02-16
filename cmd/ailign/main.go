package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ailign/cli/internal/cli"
)

// version and commit are set by goreleaser via ldflags.
var (
	version = "dev"
	commit  = "none"
)

func main() {
	rootCmd := cli.NewRootCommand()
	rootCmd.Version = version + " (" + commit + ")"
	if err := rootCmd.Execute(); err != nil {
		if !errors.Is(err, cli.ErrAlreadyReported) {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(2)
	}
}
