package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ailign/cli/internal/cli"
)

func main() {
	rootCmd := cli.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		if !errors.Is(err, cli.ErrAlreadyReported) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(2)
	}
}
