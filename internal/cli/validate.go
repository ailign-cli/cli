package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the .ailign.yml configuration file",
		Long:  "Validates the .ailign.yml configuration file in the current working directory against the schema. Reports all errors and warnings. Does not trigger any other operations.",
		RunE:  runValidate,
	}
}

func runValidate(cmd *cobra.Command, args []string) error {
	result := loadAndValidateConfig(cmd)
	formatter := getFormatter(formatFlag)
	outResult := toOutputResult(result, ".ailign.yml")

	if !result.Valid {
		// Errors already printed by loadAndValidateConfig
		return fmt.Errorf("validation failed")
	}

	// Print success to stdout
	fmt.Fprint(cmd.OutOrStdout(), formatter.FormatSuccess(outResult))
	return nil
}
