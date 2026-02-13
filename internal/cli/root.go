package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ailign/cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	formatFlag string
	loadedCfg  *config.Config
)

// NewRootCommand creates the root ailign command with global flags.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ailign",
		Short: "Instruction governance & distribution for engineering organizations",
		Long:  "AIlign manages AI coding assistant instructions across tools and repositories.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip config loading for help and completion commands
			if cmd.Name() == "help" || cmd.Name() == "completion" {
				return nil
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}

			cfgPath := filepath.Join(cwd, ".ailign.yml")
			cfg, err := config.LoadFromFile(cfgPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(2)
			}

			loadedCfg = cfg
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "human",
		"Output format: human or json")

	return rootCmd
}

// GetConfig returns the loaded config. Must be called after PersistentPreRunE.
func GetConfig() *config.Config {
	return loadedCfg
}

// GetFormat returns the current output format flag value.
func GetFormat() string {
	return formatFlag
}
