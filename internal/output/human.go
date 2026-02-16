package output

import (
	"fmt"
	"strings"
)

// HumanFormatter formats validation results for human-readable terminal output.
type HumanFormatter struct{}

func (f *HumanFormatter) FormatSuccess(result ValidationResult) string {
	if len(result.Warnings) == 0 {
		return fmt.Sprintf("%s: valid\n", result.File)
	}
	n := len(result.Warnings)
	return fmt.Sprintf("%s: valid (%d %s)\n", result.File, n, pluralize("warning", n))
}

func (f *HumanFormatter) FormatErrors(result ValidationResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Error: %s validation failed\n", result.File)

	for _, e := range result.Errors {
		b.WriteString("\n")
		formatEntry(&b, e)
	}
	b.WriteString("\n")

	n := len(result.Errors)
	fmt.Fprintf(&b, "%d %s found\n", n, pluralize("error", n))
	return b.String()
}

func (f *HumanFormatter) FormatWarnings(result ValidationResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Warning: %s has warnings\n", result.File)

	for _, w := range result.Warnings {
		b.WriteString("\n")
		formatEntry(&b, w)
	}
	b.WriteString("\n")

	n := len(result.Warnings)
	fmt.Fprintf(&b, "%d %s found\n", n, pluralize("warning", n))
	return b.String()
}

// formatEntry writes a single error or warning entry to the builder.
func formatEntry(b *strings.Builder, e ValidationError) {
	fmt.Fprintf(b, "  %s: %s\n", e.FieldPath, e.Message)
	if e.Expected != "" {
		fmt.Fprintf(b, "    Expected: %s\n", e.Expected)
	}
	if e.Actual != "" {
		fmt.Fprintf(b, "    Found: %s\n", e.Actual)
	}
	if e.Remediation != "" {
		fmt.Fprintf(b, "    Fix: %s\n", e.Remediation)
	}
}

// pluralize returns the singular or plural form of a word based on count.
func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}
