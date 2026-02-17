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

// FormatSyncResult formats a sync result for human-readable terminal output.
func (f *HumanFormatter) FormatSyncResult(result SyncResult) string {
	var b strings.Builder

	totalTargets := len(result.Links)

	if result.DryRun {
		b.WriteString("Dry run — no files will be modified.\n")
	} else {
		fmt.Fprintf(&b, "Syncing instructions to %d %s...\n", totalTargets, pluralize("target", totalTargets))
	}

	b.WriteString("\n")

	// Hub file status
	hubLabel := result.HubPath
	if result.DryRun {
		fmt.Fprintf(&b, "  %-40s %s\n", hubLabel, dryRunHubStatus(result.HubStatus))
	} else {
		fmt.Fprintf(&b, "  %-40s %s\n", hubLabel, result.HubStatus)
	}

	// Per-target status
	for _, link := range result.Links {
		label := link.LinkPath
		if link.Status == "error" {
			fmt.Fprintf(&b, "  %-40s error: %s\n", label, link.Error)
		} else if result.DryRun {
			fmt.Fprintf(&b, "  %-40s %s\n", label, dryRunLinkStatus(link.Status))
		} else {
			fmt.Fprintf(&b, "  %-40s %s\n", label, humanLinkStatus(link.Status))
		}
	}

	b.WriteString("\n")

	// Summary line
	var created, existing, errors int
	for _, link := range result.Links {
		switch link.Status {
		case "created", "replaced":
			created++
		case "exists":
			existing++
		case "error":
			errors++
		}
	}

	overlayCount := result.OverlayCount
	if errors > 0 {
		fmt.Fprintf(&b, "Synced %d of %d %s from %d %s (%d %s).\n",
			totalTargets-errors, totalTargets, pluralize("target", totalTargets),
			overlayCount, pluralize("overlay", overlayCount),
			errors, pluralize("error", errors))
	} else if existing == totalTargets {
		fmt.Fprintf(&b, "All %d %s up to date.\n", totalTargets, pluralize("target", totalTargets))
	} else if result.DryRun {
		fmt.Fprintf(&b, "Would sync %d %s from %d %s.\n",
			totalTargets, pluralize("target", totalTargets),
			overlayCount, pluralize("overlay", overlayCount))
	} else {
		fmt.Fprintf(&b, "Synced %d %s from %d %s.\n",
			totalTargets, pluralize("target", totalTargets),
			overlayCount, pluralize("overlay", overlayCount))
	}

	return b.String()
}

func dryRunHubStatus(status string) string {
	switch status {
	case "unchanged":
		return "unchanged"
	default:
		return "would be written"
	}
}

func dryRunLinkStatus(status string) string {
	switch status {
	case "exists":
		return "symlink ok"
	case "replaced":
		return "would replace symlink"
	default:
		return "would create symlink"
	}
}

func humanLinkStatus(status string) string {
	switch status {
	case "created":
		return "symlink created"
	case "exists":
		return "symlink ok"
	case "replaced":
		return "symlink replaced"
	default:
		return status
	}
}

// pluralize returns the singular or plural form of a word based on count.
func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}
