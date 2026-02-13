package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ailign/cli/internal/config"
	"github.com/ailign/cli/internal/output"
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
			// Skip for validate command (it handles its own loading)
			if cmd.Name() == "validate" {
				return nil
			}

			result := loadAndValidateConfig(cmd)
			if !result.Valid {
				os.Exit(2)
			}

			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "human",
		"Output format: human or json")

	rootCmd.AddCommand(newValidateCommand())

	return rootCmd
}

// loadAndValidateConfig loads and validates the config, printing any
// errors or warnings to the appropriate output streams.
func loadAndValidateConfig(cmd *cobra.Command) *config.ValidationResult {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), "Error: getting working directory:", err)
		return &config.ValidationResult{Valid: false}
	}

	cfgPath := filepath.Join(cwd, ".ailign.yml")
	result := config.LoadAndValidate(cfgPath)
	formatter := getFormatter(formatFlag)
	outResult := toOutputResult(result, ".ailign.yml")

	if len(result.Warnings) > 0 {
		fmt.Fprint(cmd.ErrOrStderr(), formatter.FormatWarnings(outResult))
	}

	if !result.Valid {
		fmt.Fprint(cmd.ErrOrStderr(), formatter.FormatErrors(outResult))
		return result
	}

	loadedCfg = result.Config
	return result
}

func getFormatter(format string) output.Formatter {
	switch format {
	case "json":
		return &output.JSONFormatter{}
	default:
		return &output.HumanFormatter{}
	}
}

// toOutputResult converts a config.ValidationResult to an output.ValidationResult.
func toOutputResult(r *config.ValidationResult, file string) output.ValidationResult {
	out := output.ValidationResult{
		Valid: r.Valid,
		File:  file,
	}

	for _, e := range r.Errors {
		out.Errors = append(out.Errors, output.ValidationError{
			FieldPath:   e.FieldPath,
			Expected:    e.Expected,
			Actual:      e.Actual,
			Message:     e.Message,
			Remediation: e.Remediation,
			Severity:    e.Severity,
		})
	}

	for _, w := range r.Warnings {
		out.Warnings = append(out.Warnings, output.ValidationError{
			FieldPath:   w.FieldPath,
			Expected:    w.Expected,
			Actual:      w.Actual,
			Message:     w.Message,
			Remediation: w.Remediation,
			Severity:    w.Severity,
		})
	}

	return out
}

// GetConfig returns the loaded config. Must be called after PersistentPreRunE.
func GetConfig() *config.Config {
	return loadedCfg
}

// GetFormat returns the current output format flag value.
func GetFormat() string {
	return formatFlag
}
